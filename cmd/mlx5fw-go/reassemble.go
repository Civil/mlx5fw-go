package main

import (
	"github.com/spf13/cobra"

	"github.com/Civil/mlx5fw-go/pkg/reassemble"
)

// ReassembleOptions contains options for the reassemble command
type ReassembleOptions struct {
	InputDir   string
	OutputFile string
	VerifyCRC  bool
	BinaryOnly bool
}

func runReassembleCommand(cmd *cobra.Command, args []string, opts ReassembleOptions) error {
	// Create reassembler options
	reassembleOpts := reassemble.Options{
		InputDir:   opts.InputDir,
		OutputFile: opts.OutputFile,
		VerifyCRC:  opts.VerifyCRC,
		BinaryOnly: opts.BinaryOnly,
	}

	// Create and run reassembler
	reassembler := reassemble.New(logger, reassembleOpts)
	return reassembler.Reassemble()
}
