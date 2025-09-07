package main

import (
    "fmt"

    "github.com/spf13/cobra"
    "go.uber.org/zap"

    cliutil "github.com/Civil/mlx5fw-go/pkg/cliutil"
    "github.com/Civil/mlx5fw-go/pkg/interfaces"
    "github.com/Civil/mlx5fw-go/pkg/types"
    "github.com/Civil/mlx5fw-go/pkg/compressutil"
)

func runPrintConfigCommand(cmd *cobra.Command, args []string) error {
	logger.Info("Starting print-config command")

	// Initialize firmware parser
	ctx, err := cliutil.InitializeFirmwareParser(firmwarePath, logger)
	if err != nil {
		return err
	}
	defer ctx.Close()

	fs4Parser := ctx.Parser

	// Check if firmware is encrypted
	if fs4Parser.IsEncrypted() {
		return fmt.Errorf("Operation not supported on an encrypted image")
	}

	// Get all sections
	sections := fs4Parser.GetSections()

	// Look for DBG_FW_INI section (type 0x30)
	var dbgIniSection interfaces.SectionInterface
	if sectionList, ok := sections[types.SectionTypeDbgFWINI]; ok && len(sectionList) > 0 {
		dbgIniSection = sectionList[0] // Take the first one if multiple exist
	}

	if dbgIniSection == nil {
		return fmt.Errorf("DBG_FW_INI section not found in firmware")
	}

	logger.Info("Found DBG_FW_INI section",
		zap.Uint64("offset", dbgIniSection.Offset()),
		zap.Uint32("size", dbgIniSection.Size()))

	// Read section data
	sectionData, err := ctx.Reader.ReadSection(int64(dbgIniSection.Offset()), dbgIniSection.Size())
	if err != nil {
		return fmt.Errorf("failed to read DBG_FW_INI section: %w", err)
	}

	// According to mstflint source code analysis:
	// DBG_FW_INI section is always compressed with zlib in practice,
	// but this is NOT indicated in the ITOC entry flags.
	// mstflint just attempts to decompress unconditionally.

	logger.Debug("Attempting to decompress DBG_FW_INI section")

	// First check if data looks like it's compressed (zlib header)
	if len(sectionData) < 2 {
		return fmt.Errorf("DBG_FW_INI section too small")
	}

	// Try to decompress - mstflint always attempts this for DBG_FW_INI
    decompressedData, err := compressutil.DecompressZlib(sectionData)
	if err != nil {
		// If decompression fails, maybe it's not compressed
		logger.Debug("Decompression failed, trying as uncompressed data")
		decompressedData = sectionData
	}

	// Print the decompressed INI data
	fmt.Print(string(decompressedData))

	return nil
}

// zlib decompression now lives in pkg/compressutil
