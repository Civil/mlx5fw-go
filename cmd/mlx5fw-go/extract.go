package main

import (
	"github.com/spf13/cobra"
)

func runExtractCommand(cmd *cobra.Command, args []string, outputDir string) error {
	// Get JSON export flag
	exportJSON, _ := cmd.Flags().GetBool("json")
	
	// Always use new implementation with CRC removal and metadata inclusion
	opts := ExtractOptions{
		OutputDir:       outputDir,
		ExportJSON:      exportJSON,
		IncludeMetadata: true,  // Always include metadata
		RemoveCRC:       true,  // Always remove CRC
	}
	return runExtractCommandNew(cmd, args, opts)
}

