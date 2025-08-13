package cliutil

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/parser/fs4"
	"github.com/Civil/mlx5fw-go/pkg/types"
)

// ParserContext provides shared resources for commands needing a parsed firmware
type ParserContext struct {
	Logger       *zap.Logger
	FirmwarePath string
	Reader       *parser.FirmwareReader
	Parser       *fs4.Parser
}

// InitializeFirmwareParser creates and initializes a firmware parser
// This consolidates the common pattern used across commands
func InitializeFirmwareParser(firmwarePath string, logger *zap.Logger) (*ParserContext, error) {
	// Open firmware file
	reader, err := parser.NewFirmwareReader(firmwarePath, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to open firmware: %w", err)
	}

	// Detect firmware format by reading hardware pointers
	format, err := detectFirmwareFormat(reader, logger)
	if err != nil {
		reader.Close()
		return nil, fmt.Errorf("failed to detect firmware format: %w", err)
	}

	// Use FS4 parser but mark FS5 format when detected
	fs4Parser := fs4.NewParser(reader, logger)
	if format == types.FormatFS5 {
		fs4Parser.SetFormat(types.FormatFS5)
	}

	if err := fs4Parser.Parse(); err != nil {
		reader.Close()
		return nil, fmt.Errorf("failed to parse firmware: %w", err)
	}

	return &ParserContext{
		Logger:       logger,
		FirmwarePath: firmwarePath,
		Reader:       reader,
		Parser:       fs4Parser,
	}, nil
}

// Close releases resources held by the context
func (ctx *ParserContext) Close() {
	if ctx.Reader != nil {
		_ = ctx.Reader.Close()
	}
}

// detectFirmwareFormat detects whether the firmware is FS4 or FS5 based on boot version
// Mirrors mstflint logic at a high level
func detectFirmwareFormat(reader *parser.FirmwareReader, logger *zap.Logger) (types.FirmwareFormat, error) {
	// Find magic pattern
	magicOffset, err := reader.FindMagicPattern()
	if err != nil {
		return types.FormatUnknown, err
	}

	// Read boot version structure at offset 0x10 from magic pattern
	bootVersionOffset := int64(magicOffset + types.BootVersionOffset)
	bootVersionData, err := reader.ReadSection(bootVersionOffset, 4) // Boot version is 4 bytes
	if err != nil {
		return types.FormatUnknown, fmt.Errorf("failed to read boot version: %w", err)
	}

	// Parse boot version structure
	var bootVersion types.FirmwareBootVersion
	if err := bootVersion.Unmarshal(bootVersionData); err != nil {
		return types.FormatUnknown, fmt.Errorf("failed to parse boot version: %w", err)
	}

	switch bootVersion.ImageFormatVersion {
	case types.ImageFormatVersionFS4:
		logger.Debug("Detected FS4 format from boot version")
		return types.FormatFS4, nil
	case types.ImageFormatVersionFS5:
		logger.Debug("Detected FS5 format from boot version")
		return types.FormatFS5, nil
	default:
		logger.Warn("Unknown image format version", zap.Uint8("version", bootVersion.ImageFormatVersion))
		return types.FormatFS4, nil // default for compatibility
	}
}
