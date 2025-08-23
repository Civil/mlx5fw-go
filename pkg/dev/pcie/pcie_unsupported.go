//go:build !linux
// +build !linux

package pcie

import (
	"fmt"
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

// Open is not supported on non-Linux platforms yet.
func Open(spec string, logger *zap.Logger) (Device, error) {
	return nil, fmt.Errorf("pcie device access not supported on this OS")
}
