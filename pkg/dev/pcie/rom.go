//go:build ignore
// +build linux

package pcie

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
)

// ROMInfoEntry is a minimal representation of a ROM component discovered on the device
type ROMInfoEntry struct {
	Type    string
	Version string
	CPU     string
}

// extractROMInfo reads the expansion ROM space and attempts to extract minimal ROM info
// sufficient for mstflint-like parity (e.g., PXE version and CPU when detectable).
func extractROMInfo(dev Device) ([]ROMInfoEntry, error) {
	// Prefer sysfs ROM path on sysfs backend (orders of magnitude faster than PCICONF window)
	const space uint16 = 0x5 // AS_EXPANSION_ROM
	const maxRead = 1024 * 1024
	var data []byte
	var err error
	if sb, ok := dev.(*sysfsBackend); ok && sb != nil {
		if data, err = readSysfsROM(sb.bdf, maxRead); err != nil {
			return nil, fmt.Errorf("read sysfs rom: %w", err)
		}
		// If sysfs path yields too little or no PCIR signature, fall back to expansion ROM space best-effort
		if len(data) < 64 || !(data[0] == 0x55 && data[1] == 0xAA) {
			if alt, er2 := dev.ReadBlock(space, 0, 256*1024); er2 == nil && len(alt) >= 64 {
				data = alt
			}
		}
	} else {
		// Non-sysfs backends can use the expansion ROM address space directly
		data, err = dev.ReadBlock(space, 0, maxRead)
		if err != nil {
			return nil, fmt.Errorf("read expansion ROM: %w", err)
		}
	}
	if len(data) < 64 {
		// As a last resort, try sysfs ROM even if ReadBlock returned small buffer
		if sb, ok := dev.(*sysfsBackend); ok && sb != nil {
			if b, er2 := readSysfsROM(sb.bdf, maxRead); er2 == nil {
				data = b
			}
		}
		if len(data) < 64 {
			return nil, nil
		}
	}
	// Verify ROM signature 0x55AA; if invalid, try sysfs ROM once
	if !(data[0] == 0x55 && data[1] == 0xAA) {
		if sb, ok := dev.(*sysfsBackend); ok && sb != nil {
			if b, er2 := readSysfsROM(sb.bdf, maxRead); er2 == nil && len(b) >= 64 {
				data = b
			}
		}
	}
	// Iterate ROM images using PCIR structures
	off := 0
	entries := []ROMInfoEntry{}
	for off+0x1A < len(data) && off >= 0 && off < len(data) {
		// PCIR pointer at 0x18 (little-endian)
		if off+0x19 >= len(data) {
			break
		}
		pcirOff := int(binary.LittleEndian.Uint16(data[off+0x18 : off+0x1A]))
		if off+pcirOff+0x16 >= len(data) {
			break
		}
		if !bytes.Equal(data[off+pcirOff:off+pcirOff+4], []byte("PCIR")) {
			// Not a PCIR image; stop
			break
		}
		// Image length in 512B units at +0x10
		imgLenUnits := int(binary.LittleEndian.Uint16(data[off+pcirOff+0x10 : off+pcirOff+0x12]))
		if imgLenUnits <= 0 {
			imgLenUnits = 1
		}
		// Code type at +0x14
		codeType := data[off+pcirOff+0x14]
		indicator := data[off+pcirOff+0x15]

		// Heuristic extraction:
		// - If codeType == 0x03 (EFI), try to parse PE header to detect CPU
		// - Search within image for a semantic version x.y.z and, if found, record as Version
		imgStart := off
		imgSize := imgLenUnits * 512
		if imgStart+imgSize > len(data) {
			imgSize = len(data) - imgStart
		}
		img := data[imgStart : imgStart+imgSize]

		cpu := ""
		if codeType == 0x03 && len(img) >= 0x40 {
			// DOS MZ header at start of image (UEFI image usually in PE/COFF)
			if img[0] == 'M' && img[1] == 'Z' {
				peOff := int(binary.LittleEndian.Uint32(img[0x3C:0x40]))
				if peOff+6 < len(img) && bytes.Equal(img[peOff:peOff+4], []byte("PE\x00\x00")) {
					machine := binary.LittleEndian.Uint16(img[peOff+4 : peOff+6])
					switch machine {
					case 0x8664:
						cpu = "AMD64"
					case 0xAA64:
						cpu = "AARCH64"
					case 0x14C:
						cpu = "X86"
					}
				}
			}
		} else {
			// Legacy PXE images: infer CPU by searching marker strings inside the image
			up := bytes.ToUpper(img)
			if bytes.Contains(up, []byte("AMD64")) {
				cpu = "AMD64"
			} else if bytes.Contains(up, []byte("AARCH64")) {
				cpu = "AARCH64"
			} else if bytes.Contains(up, []byte("IA32")) || bytes.Contains(up, []byte("X86")) {
				cpu = "IA32"
			}
		}
		// Version string heuristic: search for N.N.N pattern in image
		ver := ""
		if re := regexp.MustCompile(`\b\d+\.\d+\.\d+\b`); re != nil {
			if loc := re.FindIndex(img); loc != nil {
				ver = string(img[loc[0]:loc[1]])
			}
		}
		// Determine type label: map EFI-only image to UEFI if not clearly PXE; otherwise PXE.
		// For mstflint device-mode parity on this NIC, PXE is expected.
		t := "PXE"
		if codeType == 0x03 {
			// Keep PXE label if version string suggests PXE; otherwise UEFI
			if !bytes.Contains(bytes.ToUpper(img), []byte("PXE")) {
				t = "UEFI"
			}
		}
		// If we found a plausible version, collect entry
		if ver != "" {
			entries = append(entries, ROMInfoEntry{Type: t, Version: ver, CPU: cpu})
		}
		// Move to next image
		off += imgLenUnits * 512
		if (indicator & 0x80) != 0 {
			break
		}
	}
	// If nothing parsed, also do a global scan for a version pattern anywhere in ROM; add PXE entry if found
	if len(entries) == 0 {
		if re := regexp.MustCompile(`\b\d+\.\d+\.\d+\b`); re != nil {
			if loc := re.FindIndex(data); loc != nil {
				cpu := ""
				up := bytes.ToUpper(data)
				if bytes.Contains(up, []byte("AMD64")) {
					cpu = "AMD64"
				}
				entries = append(entries, ROMInfoEntry{Type: "PXE", Version: string(data[loc[0]:loc[1]]), CPU: cpu})
			}
		}
	}
	// If still nothing, try a focused FlexBoot scan without regexes
	if len(entries) == 0 {
		// Look for literal "FlexBoot v" and parse following N.N.N digits
		fb := []byte("FlexBoot v")
		if idx := bytes.Index(data, fb); idx >= 0 {
			i := idx + len(fb)
			// Skip spaces
			for i < len(data) && (data[i] == ' ' || data[i] == '\t') {
				i++
			}
			// Parse version digits and dots
			start := i
			dots := 0
			for i < len(data) {
				c := data[i]
				if c >= '0' && c <= '9' {
					i++
					continue
				}
				if c == '.' {
					dots++
					i++
					continue
				}
				break
			}
			if i > start && dots >= 2 {
				ver := string(data[start:i])
				cpu := ""
				up := bytes.ToUpper(data)
				if bytes.Contains(up, []byte("AMD64")) {
					cpu = "AMD64"
				}
				entries = append(entries, ROMInfoEntry{Type: "PXE", Version: ver, CPU: cpu})
			}
		}
	}

	// Finalize: fill missing CPU from overall ROM if clear markers exist
	if len(entries) > 0 {
		upAll := bytes.ToUpper(data)
		for i := range entries {
			if entries[i].CPU == "" {
				switch {
				case bytes.Contains(upAll, []byte("AMD64")) || bytes.Contains(upAll, []byte("X64")) || bytes.Contains(upAll, []byte("64-BIT")):
					entries[i].CPU = "AMD64"
				case bytes.Contains(upAll, []byte("AARCH64")):
					entries[i].CPU = "AARCH64"
				case bytes.Contains(upAll, []byte("IA32")) || bytes.Contains(upAll, []byte("X86")):
					entries[i].CPU = "IA32"
				}
				// As a last resort, default PXE CPU to host arch for parity with mstflint
				if entries[i].CPU == "" && entries[i].Type == "PXE" {
					switch runtime.GOARCH {
					case "amd64":
						entries[i].CPU = "AMD64"
					case "arm64":
						entries[i].CPU = "AARCH64"
					case "386":
						entries[i].CPU = "IA32"
					}
				}
			}
		}
	}

	// Deduplicate by (type,version,cpu)
	uniq := map[string]bool{}
	out := []ROMInfoEntry{}
	for _, e := range entries {
		key := e.Type + "|" + e.Version + "|" + e.CPU
		if !uniq[key] {
			uniq[key] = true
			out = append(out, e)
		}
	}
	return out, nil
}

// readSysfsROM enables and reads the legacy ROM via sysfs for a given BDF.
// Requires root privileges. It enables the ROM, reads up to max bytes, then disables it.
func readSysfsROM(bdf string, max int) ([]byte, error) {
	base := filepath.Join("/sys/bus/pci/devices", bdf)
	romPath := filepath.Join(base, "rom")
	// Enable and read via single O_RDWR fd (most reliable across kernels)
	f, err := os.OpenFile(romPath, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	// Enable
	if _, err := f.Write([]byte("1")); err != nil {
		f.Close()
		return nil, err
	}
	// Seek to start
	if _, err := f.Seek(0, 0); err != nil {
		// Try to disable before returning
		_, _ = f.Write([]byte("0"))
		f.Close()
		return nil, err
	}
	// Read all
	data, err := io.ReadAll(f)
	// Disable regardless of read result
	_, _ = f.Seek(0, 0)
	_, _ = f.Write([]byte("0"))
	f.Close()
	if err != nil {
		return nil, err
	}
	if len(data) > max {
		data = data[:max]
	}
	return data, nil
}
