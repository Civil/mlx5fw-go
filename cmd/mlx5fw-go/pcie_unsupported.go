//go:build !linux

package main

import (
	"github.com/Civil/mlx5fw-go/pkg/errors"

	"github.com/spf13/cobra"
)

const (
	DevSupported = false
)

func runQueryDeviceCommand(_ *cobra.Command, _ []string, _ bool, _ bool) error {
	return errors.ErrNotSupported
}

func createDebugARCommands() *cobra.Command {
	return nil
}

func CreateDebugCommand() *cobra.Command {
	return nil
}
