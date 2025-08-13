//go:build linux
// +build linux

package pcie

import (
	"fmt"
	"strings"
)

// Minimal decoders for MGIR/MCQI payloads to extract fields needed for query.

type MGIRInfo struct {
	FWVersion      string
	ProductVersion string
	SecurityVer    uint16
	SecurityAttrs  string
	PRSName        string
	PartNumber     string
	Description    string
	ImageVSD       string
	DeviceVSD      string
	FWReleaseDate  string
	ImageType      string
	PSID           string
}

// MCQIInfo holds a minimal subset we may want to surface from MCQI
type MCQIInfo struct {
	VersionMajor     uint8
	VersionMinor     uint8
	ActivationMethod string
}

// decodeMCQI parses a very small subset of MCQI payload.
// This is a placeholder; proper field mapping should follow reg_access_hca_layouts.h
func decodeMCQI(b []byte) MCQIInfo {
	var out MCQIInfo
	// Minimal, conservative parse:
	// Try to derive activation method flags from the first byte of the data union (offset 0x18).
	if len(b) >= 0x19 {
		flags := b[0x18]
		parts := []string{}
		// Bit mapping per reg_access_hca_mcqi_activation_method_ext (heuristic, minimal)
		if flags&0x01 != 0 {
			parts = append(parts, "all_hosts_sync")
		}
		if flags&0x02 != 0 {
			parts = append(parts, "auto_activate")
		}
		if flags&0x04 != 0 {
			parts = append(parts, "pending_fw_reset")
		}
		if flags&0x08 != 0 {
			parts = append(parts, "pending_server_reboot")
		}
		if flags&0x10 != 0 {
			parts = append(parts, "pending_server_dc_power_cycle")
		}
		if flags&0x20 != 0 {
			parts = append(parts, "pending_server_ac_power_cycle")
		}
		if flags&0x40 != 0 {
			parts = append(parts, "self_activation")
		}
		if len(parts) > 0 {
			out.ActivationMethod = strings.Join(parts, ",")
		}
	}
	return out
}

// decodeMGIRFWInfo parses a subset of MGIR FW info area from raw bytes.
// This is a placeholder stub; real implementation should follow reg_access_hca_layouts.h
func decodeMGIRFWInfo(b []byte) MGIRInfo {
	var out MGIRInfo
	if len(b) < 0xa0 {
		return out
	}
	// mgir_ext layout (size 160 bytes)
	// hw_info at 0x00, fw_info at 0x20, sw_info at 0x60, dev_info at 0x80
	fwOff := 0x20
	// Defensive bounds
	if fwOff+0x5c > len(b) {
		return out
	}
	// Prefer extended_{major,minor,sub_minor} at 0x24/0x28/0x2c (u32 each, LE)
	// Fallback to legacy bytes [0..2] only if extended values look zero.
	extMaj := int(uint32(b[fwOff+0x24]) | uint32(b[fwOff+0x25])<<8 | uint32(b[fwOff+0x26])<<16 | uint32(b[fwOff+0x27])<<24)
	extMin := int(uint32(b[fwOff+0x28]) | uint32(b[fwOff+0x29])<<8 | uint32(b[fwOff+0x2a])<<16 | uint32(b[fwOff+0x2b])<<24)
	extSub := int(uint32(b[fwOff+0x2c]) | uint32(b[fwOff+0x2d])<<8 | uint32(b[fwOff+0x2e])<<16 | uint32(b[fwOff+0x2f])<<24)
	if extMaj != 0 || extMin != 0 || extSub != 0 {
		out.FWVersion = fmt.Sprintf("%d.%d.%04d", extMaj, extMin, extSub)
		out.ProductVersion = out.FWVersion
	} else {
		// Legacy 8-bit fields
		sub := int(b[fwOff+0])
		min := int(b[fwOff+1])
		maj := int(b[fwOff+2])
		out.FWVersion = fmt.Sprintf("%d.%d.%04d", maj, min, sub)
		out.ProductVersion = out.FWVersion
	}
	// Security attributes (secured/signed/dev/debug/dev_sc)
	secured := b[fwOff+3]&0x1 == 1
	signed := b[fwOff+4]&0x1 == 1
	debug := b[fwOff+5]&0x1 == 1
	dev := b[fwOff+6]&0x1 == 1
	devsc := b[fwOff+8]&0x1 == 1
	attrs := []string{}
	if secured {
		attrs = append(attrs, "secured")
	}
	if signed {
		attrs = append(attrs, "signed")
	}
	if debug {
		attrs = append(attrs, "debug")
	}
	if dev {
		attrs = append(attrs, "dev")
	}
	if devsc {
		attrs = append(attrs, "dev_sc")
	}
	if len(attrs) > 0 {
		out.SecurityAttrs = strings.Join(attrs, ",")
	}
	// Release date fields per reg_access_hca_mgir_fw_info_ext (LE within 32b words):
	// year (u16) at fwOff+0x8, day (u8) at +0xa, month (u8) at +0xb
	if fwOff+0x0c <= len(b) {
		y := int(uint32(b[fwOff+0x8]) | uint32(b[fwOff+0x9])<<8)
		d := int(b[fwOff+0xa])
		m := int(b[fwOff+0xb])
		// Heuristic for 2-digit year encodings: treat <100 as 2000+yy
		if y < 100 {
			y += 2000
		}
		if d >= 1 && d <= 31 && m >= 1 && m <= 12 && y >= 2000 && y <= 2099 {
			out.FWReleaseDate = fmt.Sprintf("%02d.%02d.%04d", d, m, y)
		}
	}
	// PSID 16-byte array at fw_info offset 0x10
	if fwOff+0x10+16 <= len(b) {
		psidBytes := b[fwOff+0x10 : fwOff+0x10+16]
		// Convert to ASCII if printable; otherwise hex
		isPrint := true
		for _, c := range psidBytes {
			if c == 0 {
				continue
			}
			if c < 0x20 || c > 0x7e {
				isPrint = false
				break
			}
		}
		if isPrint {
			out.PSID = strings.TrimRight(string(psidBytes), "\x00")
		} else {
			out.PSID = fmt.Sprintf("%X", psidBytes)
		}
	}
	// Security version (image security version / efuse) not in fw_info; default 0
	out.SecurityVer = 0
	// Conservative defaults for other fields (to be populated as layout knowledge improves)
	out.ImageType = ""
	out.PRSName = ""
	out.PartNumber = ""
	out.Description = ""
	out.ImageVSD = ""
	out.DeviceVSD = ""
	return out
}
