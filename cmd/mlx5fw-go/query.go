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

	// Return error if ITOC was invalid and firmware is not encrypted
	// For encrypted firmware, invalid ITOC is expected
	if hasInvalidITOC && !fs4Parser.IsEncrypted() {
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
	
	// Display GUID/MAC information based on format
	if info.UseDualFormat {
		// Dual format for encrypted firmware (GUID1/GUID2, MAC1/MAC2)
		fmt.Printf("Description:           UID                GuidsNumber  Step\n")
		
		// GUID1
		if info.BaseGUID != 0 {
			fmt.Printf("Base GUID1:            %016x        %d       %d\n", info.BaseGUID, info.BaseGUIDNum, info.GUIDStep)
		} else {
			fmt.Printf("Base GUID1:            N/A                    N/A       N/A\n")
		}
		
		// GUID2
		if info.BaseGUID2 != 0 {
			fmt.Printf("Base GUID2:            %016x        %d       %d\n", info.BaseGUID2, info.BaseGUID2Num, info.GUIDStep)
		} else {
			fmt.Printf("Base GUID2:            N/A                    N/A       N/A\n")
		}
		
		// MAC1
		if info.BaseMAC != 0 {
			fmt.Printf("Base MAC1:             %012x            %d       %d\n", info.BaseMAC, info.BaseMACNum, info.MACStep)
		} else {
			fmt.Printf("Base MAC1:             N/A                    N/A       N/A\n")
		}
		
		// MAC2
		if info.BaseMAC2 != 0 {
			fmt.Printf("Base MAC2:             %012x            %d       %d\n", info.BaseMAC2, info.BaseMAC2Num, info.MACStep)
		} else {
			fmt.Printf("Base MAC2:             N/A                    N/A       N/A\n")
		}
	} else {
		// Single format (normal firmware)
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
	}
	fmt.Printf("Image VSD:             %s\n", FormatNA(info.ImageVSD))
	fmt.Printf("Device VSD:            %s\n", FormatNA(info.DeviceVSD))
	fmt.Printf("PSID:                  %s\n", info.PSID)
	
	// Show Orig PSID if it exists or if using dual format (encrypted firmware)
	if info.OrigPSID != "" && info.OrigPSID != info.PSID {
		fmt.Printf("Orig PSID:             %s\n", info.OrigPSID)
	} else if info.UseDualFormat {
		// For encrypted firmware, always show Orig PSID even if N/A
		fmt.Printf("Orig PSID:             N/A\n")
	}
	
	fmt.Printf("Security Attributes:   %s\n", FormatNA(info.SecurityAttrs))
	fmt.Printf("Security Ver:          %d\n", info.SecurityVer)
	
	fmt.Printf("Default Update Method: fw_ctrl\n")

	return nil
}

