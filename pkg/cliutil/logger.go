package cliutil

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ConfigureLogger creates and configures a zap logger based on the output format and log level
func ConfigureLogger(jsonOutput bool, debugLevel bool, quiet bool) (*zap.Logger, error) {
	var config zap.Config
	if jsonOutput {
		// Use production config for JSON output
		config = zap.NewProductionConfig()
	} else {
		// Use development config for human-readable output
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	switch {
	case quiet:
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	case debugLevel:
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	config.OutputPaths = []string{"stderr"}
	// Disable stacktrace for normal errors (only show for panic level)
	config.DisableStacktrace = true

	lg, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	return lg, nil
}

// FormatNA formats empty strings as "N/A" for display consistency
func FormatNA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

// ValidateFirmwarePath validates that the firmware path is provided
func ValidateFirmwarePath(firmwarePath string) error {
	if firmwarePath == "" {
		return fmt.Errorf("firmware file path is required, use -f or --file flag")
	}
	return nil
}
