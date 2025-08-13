//go:build linux
// +build linux

package pcie

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

// Device is a minimal abstraction over a PCIe access backend (MST or sysfs).
type Device interface {
	// Read/Write a 32-bit value from the given address space and offset.
	Read32(space uint16, offset uint32) (uint32, error)
	Write32(space uint16, offset uint32, value uint32) error

	// Block I/O helper.
	ReadBlock(space uint16, offset uint32, size int) ([]byte, error)
	WriteBlock(space uint16, offset uint32, data []byte) error

	// Backend type identifier (e.g., "mst", "sysfs").
	Type() string

	// Close underlying handles.
	Close() error
}

// Open opens a PCIe device by BDF (e.g., 0000:07:00.0) or a direct path (/dev/mst/*).
// Preference order: explicit /dev/mst path -> resolve MST by BDF -> sysfs fallback.
func Open(spec string, logger *zap.Logger) (Device, error) {
	// Direct MST path
	if filepath.IsAbs(spec) && (filepath.Base(filepath.Dir(spec)) == "mst" || filepath.Base(spec) == "mst") {
		if _, err := os.Stat(spec); err == nil {
			return openMST(spec, logger)
		}
	}

	// Try BDF → MST mapping (best effort): look for /dev/mst/* entries
	if isBDF(spec) {
		if path, ok := findMSTForBDF(spec); ok {
			return openMST(path, logger)
		}
		// Fallback to sysfs backend
		return openSysfs(spec, logger)
	}

	// Last resort: attempt sysfs with given spec
	if _, err := os.Stat(spec); err == nil {
		return openSysfs(spec, logger)
	}

	return nil, fmt.Errorf("unable to resolve device: %s", spec)
}
