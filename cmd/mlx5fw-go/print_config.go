package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
)

func runPrintConfigCommand(cmd *cobra.Command, args []string) error {
	logger.Info("Starting print-config command")

	// Initialize firmware parser
	ctx, err := InitializeFirmwareParser(firmwarePath, logger)
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
	var dbgIniSection *interfaces.Section
	if sectionList, ok := sections[types.SectionTypeDbgFWINI]; ok && len(sectionList) > 0 {
		dbgIniSection = sectionList[0] // Take the first one if multiple exist
	}
	
	if dbgIniSection == nil {
		return fmt.Errorf("DBG_FW_INI section not found in firmware")
	}

	logger.Info("Found DBG_FW_INI section",
		zap.Uint64("offset", dbgIniSection.Offset),
		zap.Uint32("size", dbgIniSection.Size))

	// Read section data
	sectionData, err := ctx.Reader.ReadSection(int64(dbgIniSection.Offset), dbgIniSection.Size)
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
	decompressedData, err := decompressZlib(sectionData)
	if err != nil {
		// If decompression fails, maybe it's not compressed
		logger.Debug("Decompression failed, trying as uncompressed data")
		decompressedData = sectionData
	}

	// Print the decompressed INI data
	fmt.Print(string(decompressedData))
	
	return nil
}

// decompressZlib decompresses zlib compressed data
func decompressZlib(data []byte) ([]byte, error) {
	// Try to decompress with increasing buffer sizes
	// This mimics mstflint's behavior of trying different buffer sizes
	initialSize := len(data) * 10 // Start with 10x compressed size
	maxSize := 50 * 1024 * 1024   // Max 50MB
	
	for bufSize := initialSize; bufSize <= maxSize; bufSize *= 2 {
		reader, err := zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to create zlib reader: %w", err)
		}
		defer reader.Close()
		
		// Read all data
		decompressed, err := io.ReadAll(reader)
		if err != nil {
			// If buffer too small, try larger size
			if err == io.ErrUnexpectedEOF {
				continue
			}
			return nil, fmt.Errorf("failed to decompress: %w", err)
		}
		
		// Successfully decompressed
		return decompressed, nil
	}
	
	return nil, fmt.Errorf("failed to decompress: buffer size exceeded maximum")
}