//go:build ignore
// +build linux

package pcie

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// sysfsBackend implements Device using /sys/bus/pci/devices/.../config register access.
// NOTE: This is a minimal Phase-1 implementation. Address space selection encoding is TBD.
type sysfsBackend struct {
	logger *zap.Logger
	bdf    string
	cfg    *os.File
	vsec   int64 // Vendor Specific Capability base offset
}

const (
	// Offsets within VSEC (from mstflint mtcr_ul)
	pciCtrlOff      = 0x4
	pciCounterOff   = 0x8
	pciSemaphoreOff = 0xc
	pciAddrOff      = 0x10
	pciDataOff      = 0x14

	// Bit fields
	pciFlagBit       = 31
	pciStatusBitsOff = 29
	pciStatusBitsLen = 3
)

func openSysfs(spec string, logger *zap.Logger) (Device, error) {
	bdf := spec
	if !isBDF(bdf) {
		// allow direct path to /sys/bus/pci/devices/<BDF>
		if strings.Contains(spec, "/sys/bus/pci/devices/") {
			bdf = filepath.Base(spec)
		} else {
			return nil, fmt.Errorf("sysfs backend requires BDF or sysfs path: %s", spec)
		}
	}
	cfgPath := filepath.Join("/sys/bus/pci/devices", bdf, "config")
	f, err := os.OpenFile(cfgPath, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	sb := &sysfsBackend{logger: logger, bdf: bdf, cfg: f}
	vsec, err := sb.findVsec()
	if err != nil {
		f.Close()
		return nil, err
	}
	sb.vsec = int64(vsec)
	return sb, nil
}

func (d *sysfsBackend) Type() string { return "sysfs" }

func (d *sysfsBackend) Close() error { return d.cfg.Close() }

func (d *sysfsBackend) setAddrSpace(space uint16) error {
	// Read-modify-write CTRL lower 16 bits to space, then verify and status
	val, err := d.read32cfg(d.vsec + pciCtrlOff)
	if err != nil {
		return err
	}
	if d.logger != nil {
		d.logger.Debug("pciconf.set_space.read",
			zap.String("bdf", d.bdf),
			zap.Uint32("ctrl", val),
			zap.Uint16("current_space", uint16(val&0xFFFF)),
		)
	}
	val = (val & ^uint32(0xFFFF)) | uint32(space)
	if err := d.write32cfg(d.vsec+pciCtrlOff, val); err != nil {
		return err
	}
	rv, err := d.read32cfg(d.vsec + pciCtrlOff)
	if err != nil {
		return err
	}
	if d.logger != nil {
		d.logger.Debug("pciconf.set_space.after",
			zap.String("bdf", d.bdf),
			zap.Uint32("ctrl", rv),
			zap.Uint16("requested_space", space),
			zap.Uint16("accepted_space", uint16(rv&0xFFFF)),
			zap.Uint32("status_bits", (rv>>pciStatusBitsOff)&((1<<pciStatusBitsLen)-1)),
		)
	}
	if uint16(rv&0xFFFF) != space {
		return fmt.Errorf("space 0x%x not accepted (read back 0x%x)", space, rv&0xFFFF)
	}
	status := (rv >> pciStatusBitsOff) & ((1 << pciStatusBitsLen) - 1)
	if status == 0 {
		return fmt.Errorf("space 0x%x unsupported (status=0)", space)
	}
	return nil
}

func (d *sysfsBackend) Read32(space uint16, offset uint32) (uint32, error) {
	if err := d.acquireSem(true); err != nil {
		return 0, err
	}
	defer d.acquireSem(false)
	if err := d.setAddrSpace(space); err != nil {
		return 0, err
	}
	// Write address (aligned) with READ flag=0
	addr := offset &^ uint32(3)
	if err := d.write32cfg(d.vsec+pciAddrOff, addr); err != nil {
		return 0, fmt.Errorf("write addr: %w", err)
	}
	if err := d.waitOnFlag(1); err != nil {
		return 0, err
	}
	v, err := d.read32cfg(d.vsec + pciDataOff)
	if err != nil {
		return 0, fmt.Errorf("read data: %w", err)
	}
	return v, nil
}

func (d *sysfsBackend) Write32(space uint16, offset uint32, value uint32) error {
	if err := d.acquireSem(true); err != nil {
		return err
	}
	defer d.acquireSem(false)
	if err := d.setAddrSpace(space); err != nil {
		return err
	}
	if d.logger != nil {
		d.logger.Debug("pciconf.write32.begin", zap.String("bdf", d.bdf), zap.Uint16("space", space), zap.Uint32("offset", offset), zap.Uint32("value", value))
	}
	if err := d.write32cfg(d.vsec+pciDataOff, value); err != nil {
		return fmt.Errorf("write data: %w", err)
	}
	addr := (offset &^ uint32(3)) | (1 << pciFlagBit)
	if err := d.write32cfg(d.vsec+pciAddrOff, addr); err != nil {
		return fmt.Errorf("write addr: %w", err)
	}
	if err := d.waitOnFlag(0); err != nil {
		return err
	}
	if d.logger != nil {
		d.logger.Debug("pciconf.write32.ok", zap.String("bdf", d.bdf), zap.Uint16("space", space), zap.Uint32("offset", offset), zap.Uint32("value", value))
	}
	return nil
}

func (d *sysfsBackend) ReadBlock(space uint16, offset uint32, size int) ([]byte, error) {
	data := make([]byte, 0, size)
	for i := 0; i < size; i += 4 {
		v, err := d.Read32(space, offset+uint32(i))
		if err != nil {
			return nil, err
		}
		data = append(data, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
	}
	return data[:size], nil
}

func (d *sysfsBackend) WriteBlock(space uint16, offset uint32, b []byte) error {
	for i := 0; i < len(b); i += 4 {
		var v uint32
		chunk := b[i:min(i+4, len(b))]
		for j := 0; j < len(chunk); j++ {
			v |= uint32(chunk[j]) << (8 * (3 - j))
		}
		if err := d.Write32(space, offset+uint32(i), v); err != nil {
			return err
		}
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ReadPCIConfigSysfs reads a PCI config dword via sysfs config file.
func ReadPCIConfigSysfs(bdf string, offset uint32) (uint32, error) {
	if !isBDF(bdf) {
		return 0, fmt.Errorf("invalid BDF: %s", bdf)
	}
	cfgPath := filepath.Join("/sys/bus/pci/devices", bdf, "config")
	f, err := os.Open(cfgPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	var buf [4]byte
	if _, err := f.ReadAt(buf[:], int64(offset)); err != nil {
		return 0, err
	}
	v := uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24
	return v, nil
}

// Helpers for PCICONF via VSEC
func (d *sysfsBackend) read32cfg(off int64) (uint32, error) {
	var b [4]byte
	if _, err := d.cfg.ReadAt(b[:], off); err != nil {
		return 0, err
	}
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24, nil
}

func (d *sysfsBackend) write32cfg(off int64, v uint32) error {
	b := [4]byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)}
	if _, err := d.cfg.WriteAt(b[:], off); err != nil {
		return err
	}
	return nil
}

func (d *sysfsBackend) waitOnFlag(expected uint32) error {
	retries := 0
	var first uint32
	for {
		if retries > 2048 {
			if d.logger != nil {
				d.logger.Debug("pciconf.wait_flag.timeout", zap.String("bdf", d.bdf), zap.Uint32("expected", expected), zap.Uint32("first", first))
			}
			return fmt.Errorf("pciconf flag timeout")
		}
		v, err := d.read32cfg(d.vsec + pciAddrOff)
		if err != nil {
			return err
		}
		if retries == 0 {
			first = v
		}
		flag := (v >> pciFlagBit) & 0x1
		if flag == expected {
			if d.logger != nil {
				d.logger.Debug("pciconf.wait_flag.ok", zap.String("bdf", d.bdf), zap.Uint32("expected", expected), zap.Int("retries", retries))
			}
			return nil
		}
		retries++
		if (retries & 0xf) == 0 {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func (d *sysfsBackend) acquireSem(lock bool) error {
	if !lock {
		if d.logger != nil {
			d.logger.Debug("pciconf.sem.release", zap.String("bdf", d.bdf))
		}
		return d.write32cfg(d.vsec+pciSemaphoreOff, 0)
	}
	retries := 0
	for {
		if retries > 2048 {
			if d.logger != nil {
				d.logger.Debug("pciconf.sem.acquire.timeout", zap.String("bdf", d.bdf))
			}
			return fmt.Errorf("pciconf semaphore busy")
		}
		lockVal, err := d.read32cfg(d.vsec + pciSemaphoreOff)
		if err != nil {
			return err
		}
		if lockVal != 0 {
			retries++
			time.Sleep(1 * time.Millisecond)
			continue
		}
		counter, err := d.read32cfg(d.vsec + pciCounterOff)
		if err != nil {
			return err
		}
		if err := d.write32cfg(d.vsec+pciSemaphoreOff, counter); err != nil {
			return err
		}
		lv2, err := d.read32cfg(d.vsec + pciSemaphoreOff)
		if err != nil {
			return err
		}
		if lv2 == counter {
			if d.logger != nil {
				d.logger.Debug("pciconf.sem.acquire.ok", zap.String("bdf", d.bdf), zap.Int("retries", retries))
			}
			return nil
		}
		retries++
	}
}

// findVsec scans the capability list for Vendor Specific Capability (ID=0x09)
func (d *sysfsBackend) findVsec() (uint8, error) {
	// Capabilities pointer (byte) at 0x34
	var p [1]byte
	if _, err := d.cfg.ReadAt(p[:], 0x34); err != nil {
		return 0, fmt.Errorf("read cap ptr: %w", err)
	}
	off := int64(p[0])
	visited := 0
	for off != 0 && visited < 64 {
		var hdr [2]byte
		if _, err := d.cfg.ReadAt(hdr[:], off); err != nil {
			return 0, fmt.Errorf("read cap hdr: %w", err)
		}
		capID := hdr[0]
		next := hdr[1]
		if capID == 0x09 { // Vendor Specific Cap
			return uint8(off), nil
		}
		off = int64(next)
		visited++
	}
	return 0, fmt.Errorf("VSEC not found in PCI capabilities")
}
