package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Civil/mlx5fw-go/pkg/interfaces"
)

func runQueryCommand(cmd *cobra.Command, args []string, fullOutput bool, jsonOutput bool) error {
	logger.Info("Starting query command")

	// Initialize firmware parser
	ctx, err := InitializeFirmwareParser(firmwarePath, logger)
	if err != nil {
		return err
	}
	defer ctx.Close()

	fs4Parser := ctx.Parser

	// Check ITOC validity
	hasInvalidITOC := false
	if !fs4Parser.IsITOCValid() {
		logger.Warn("ITOC header CRC verification failed - firmware may be corrupted")
		hasInvalidITOC = true
	}

	// Query firmware information
	info, err := fs4Parser.Query()
	if err != nil {
		return fmt.Errorf("failed to query firmware: %w", err)
	}

	// Display query output
	if err := displayQueryInfo(info, fullOutput, jsonOutput); err != nil {
		return err
	}

	// Return error if ITOC was invalid
	if hasInvalidITOC {
		return fmt.Errorf("ITOC header is invalid")
	}

	return nil
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
	fmt.Printf("Image VSD:             %s\n", FormatNA(info.ImageVSD))
	fmt.Printf("Device VSD:            %s\n", FormatNA(info.DeviceVSD))
	fmt.Printf("PSID:                  %s\n", info.PSID)
	fmt.Printf("Security Attributes:   %s\n", FormatNA(info.SecurityAttrs))
	fmt.Printf("Security Ver:          %d\n", info.SecurityVer)
	
	fmt.Printf("Default Update Method: fw_ctrl\n")

	return nil
}

