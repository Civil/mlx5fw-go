package main

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	cliutil "github.com/Civil/mlx5fw-go/pkg/cliutil"
	"github.com/Civil/mlx5fw-go/pkg/extract"
)

// ExtractOptions contains options for the extract command
type ExtractOptions struct {
	OutputDir       string
	ExportJSON      bool
	IncludeMetadata bool
	RemoveCRC       bool
	KeepBinary      bool
}

func runExtractCommand(cmd *cobra.Command, args []string, outputDir string) error {
	// Get JSON export flag (deprecated but kept for compatibility)
	exportJSON, _ := cmd.Flags().GetBool("json")

	// Get keep-binary flag
	keepBinary, _ := cmd.Flags().GetBool("keep-binary")

	// Always use new implementation with CRC removal and metadata inclusion
	opts := ExtractOptions{
		OutputDir:       outputDir,
		ExportJSON:      exportJSON,
		IncludeMetadata: true, // Always include metadata
		RemoveCRC:       true, // Always remove CRC
		KeepBinary:      keepBinary,
	}
	return runExtractCommandCore(cmd, args, opts)
}

// Core implementation that performs extraction with provided options
func runExtractCommandCore(cmd *cobra.Command, args []string, opts ExtractOptions) error {
	logger.Debug("Starting extract command",
		zap.String("outputDir", opts.OutputDir),
		zap.Bool("exportJSON", opts.ExportJSON))

	// Initialize firmware parser
	ctx, err := cliutil.InitializeFirmwareParser(firmwarePath, logger)
	if err != nil {
		return err
	}
	defer ctx.Close()

	// Create extractor options
	extractOpts := extract.Options{
		OutputDir:       opts.OutputDir,
		ExportJSON:      opts.ExportJSON,
		IncludeMetadata: opts.IncludeMetadata,
		RemoveCRC:       opts.RemoveCRC,
		KeepBinary:      opts.KeepBinary,
	}

	// Create and run extractor
	extractor := extract.New(ctx.Parser, logger, extractOpts)
	return extractor.Extract()
}
