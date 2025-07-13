package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/parser/fs4"
	"github.com/Civil/mlx5fw-go/pkg/types"
)

func runExtractCommand(cmd *cobra.Command, args []string, outputDir string) error {
	logger.Info("Starting extract command", 
		zap.String("outputDir", outputDir))

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Initialize firmware parser
	ctx, err := InitializeFirmwareParser(firmwarePath, logger)
	if err != nil {
		return err
	}
	defer ctx.Close()

	fs4Parser := ctx.Parser

	// Get sections
	sections := fs4Parser.GetSections()
	
	// Extract sections
	return extractSections(sections, fs4Parser, outputDir)
}

func extractSections(sections map[uint16][]*interfaces.Section, parser *fs4.Parser, outputDir string) error {
	extractedCount := 0
	
	for _, sectionList := range sections {
		for _, section := range sectionList {
			// Get section type name
			typeName := GetSectionTypeName(section)
			
			// Clean up the type name for filename
			fileName := strings.ReplaceAll(typeName, " ", "_")
			fileName = strings.ReplaceAll(fileName, "/", "_")
			fileName = fmt.Sprintf("%s_0x%08x.bin", fileName, section.Offset)
			
			filePath := filepath.Join(outputDir, fileName)
			
			logger.Info("Extracting section",
				zap.String("type", typeName),
				zap.Uint64("offset", section.Offset),
				zap.Uint32("size", section.Size),
				zap.String("file", fileName))
			
			// Verify section to ensure data is loaded
			_, err := parser.VerifySection(section)
			if err != nil {
				logger.Warn("Failed to verify/load section", 
					zap.String("type", typeName),
					zap.Error(err))
				continue
			}
			
			// Get section data
			data := section.Data
			if data == nil || len(data) == 0 {
				logger.Warn("Section has no data", zap.String("type", typeName))
				continue
			}
			
			// Check if CRC is at the end of section and should be removed
			dataToWrite := data
			if section.CRCType == types.CRCInSection && len(data) >= 4 {
				// Remove last 4 bytes (CRC32)
				dataToWrite = data[:len(data)-4]
				logger.Debug("Removed CRC from end of section",
					zap.String("type", typeName),
					zap.Int("originalSize", len(data)),
					zap.Int("newSize", len(dataToWrite)))
			}
			
			// Write section to file
			if err := os.WriteFile(filePath, dataToWrite, 0644); err != nil {
				return fmt.Errorf("failed to write section %s: %w", fileName, err)
			}
			
			extractedCount++
		}
	}
	
	fmt.Printf("Successfully extracted %d sections to %s\n", extractedCount, outputDir)
	return nil
}