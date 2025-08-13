//go:build !linux
// +build !linux

package pcie

import (
	"fmt"
	"go.uber.org/zap"
)

// Open is not supported on non-Linux platforms yet.
func Open(spec string, logger *zap.Logger) (Device, error) {
	return nil, fmt.Errorf("pcie device access not supported on this OS")
}
