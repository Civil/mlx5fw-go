//go:build ignore
// +build linux

package pcie

import (
	"fmt"
)

// ProbeSpaces attempts to access a few common address spaces and reports support.
func ProbeSpaces(dev Device) []string {
	candidates := []struct {
		name string
		val  uint16
	}{
		{"cr", 0x2},
		{"icmd", 0x3},
		{"icmd_ext", 0x1},
		{"semaphore", 0x0a},
		{"pci_cr", 0x102},
		{"pci_icmd", 0x101},
		{"pci_gsem", 0x10a},
	}
	out := []string{}
	for _, c := range candidates {
		// Try a benign read at offset 0
		if _, err := dev.Read32(c.val, 0); err != nil {
			out = append(out, fmt.Sprintf("space %-10s 0x%03x: error=%v", c.name, c.val, err))
		} else {
			out = append(out, fmt.Sprintf("space %-10s 0x%03x: OK", c.name, c.val))
		}
	}
	return out
}
