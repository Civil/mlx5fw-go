package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/parser/fs4"
)

func runQueryCommand(cmd *cobra.Command, args []string, fullOutput bool, jsonOutput bool) error {
	// Set verbose logging if requested
	if verboseLogging {
		config := zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		var err error
		logger, err = config.Build()
		if err != nil {
			return fmt.Errorf("failed to initialize verbose logger: %w", err)
		}
	}

	logger.Info("Starting query command", zap.String("firmware", firmwarePath))

	// Check if file exists
	if _, err := os.Stat(firmwarePath); err != nil {
		return fmt.Errorf("cannot access firmware file: %w", err)
	}

	// Open firmware file
	reader, err := parser.NewFirmwareReader(firmwarePath, logger)
	if err != nil {
		return fmt.Errorf("failed to open firmware: %w", err)
	}
	defer reader.Close()

	// Create FS4 parser (for now, we only support FS4)
	fs4Parser := fs4.NewParser(reader, logger)
	
	// Parse firmware
	if err := fs4Parser.Parse(); err != nil {
		return fmt.Errorf("failed to parse firmware: %w", err)
	}

	// Query firmware information
	info, err := fs4Parser.Query()
	if err != nil {
		return fmt.Errorf("failed to query firmware: %w", err)
	}

	// Display query output
	return displayQueryInfo(info, fullOutput, jsonOutput)
}

func displayQueryInfo(info *interfaces.FirmwareInfo, fullOutput bool, jsonOutput bool) error {
	if jsonOutput {
		// Convert to JSON output format
		jsonData := convertToQueryJSON(info)
		return outputJSON(jsonData)
	}
	
	// Match mstflint output format
	fmt.Printf("Image type:            %s\n", info.Format)
	fmt.Printf("FW Version:            %s\n", info.FWVersion)
	fmt.Printf("FW Release Date:       %s\n", info.FWReleaseDate)
	fmt.Printf("MIC Version:           %s\n", info.MICVersion)
	fmt.Printf("PRS Name:              %s\n", info.PRSName)
	fmt.Printf("Part Number:           %s\n", info.PartNumber)
	fmt.Printf("Description:           %s\n", info.Description)
	
	if info.ProductVersion != "" {
		fmt.Printf("Product Version:       %s\n", info.ProductVersion)
	}
	
	// Display ROM info if available
	if len(info.RomInfo) > 0 {
		fmt.Printf("Rom Info:              ")
		romInfoStrs := []string{}
		for _, rom := range info.RomInfo {
			romStr := fmt.Sprintf("type=%s version=%s", rom.Type, rom.Version)
			if rom.CPU != "" {
				romStr += fmt.Sprintf(" cpu=%s", rom.CPU)
			}
			romInfoStrs = append(romInfoStrs, romStr)
		}
		fmt.Println(strings.Join(romInfoStrs, "\n                       "))
	}
	
	fmt.Printf("Description:           UID                GuidsNumber\n")
	if info.BaseGUID != 0 {
		fmt.Printf("Base GUID:             %016x        %d\n", info.BaseGUID, info.BaseGUIDNum)
	} else {
		fmt.Printf("Base GUID:             N/A                     %d\n", info.BaseGUIDNum)
	}
	if info.BaseMAC != 0 {
		fmt.Printf("Base MAC:              %012x            %d\n", info.BaseMAC, info.BaseMACNum)
	} else {
		fmt.Printf("Base MAC:              N/A                     %d\n", info.BaseMACNum)
	}
	fmt.Printf("Image VSD:             %s\n", formatNA(info.ImageVSD))
	fmt.Printf("Device VSD:            %s\n", formatNA(info.DeviceVSD))
	fmt.Printf("PSID:                  %s\n", info.PSID)
	fmt.Printf("Security Attributes:   %s\n", formatNA(info.SecurityAttrs))
	fmt.Printf("Security Ver:          %d\n", info.SecurityVer)
	
	fmt.Printf("Default Update Method: fw_ctrl\n")

	return nil
}

func formatNA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}