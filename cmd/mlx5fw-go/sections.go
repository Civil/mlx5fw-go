package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/parser/fs4"
	"github.com/Civil/mlx5fw-go/pkg/types"
)

// SectionDisplay represents a section for display
type SectionDisplay struct {
	StartAddr    uint64
	EndAddr      uint64
	Size         uint32
	Name         string
	Status       string
	Section      *interfaces.Section
	IsHWPointer  bool
	IsHeader     bool
}

func runSectionsCommand(cmd *cobra.Command, args []string, showContent bool, outputFormat string) error {
	logger.Info("Starting sections command")

	// Initialize firmware parser
	ctx, err := InitializeFirmwareParser(firmwarePath, logger)
	if err != nil {
		return err
	}
	defer ctx.Close()

	fs4Parser := ctx.Parser

	// Get sections
	sections := fs4Parser.GetSections()
	format := fs4Parser.GetFormat()
	
	// Display sections
	return displaySections(firmwarePath, format, sections, fs4Parser, showContent, outputFormat)
}

// JSONSection represents a section for JSON output
type JSONSection struct {
	Type               string `json:"Type"`
	StartAddress       uint64 `json:"StartAddress"`
	EndAddress         uint64 `json:"EndAddress"`
	Size               uint32 `json:"Size"`
	CRCType            string `json:"CRCType"`
	CRC                uint32 `json:"CRC"`
	IsEncrypted        bool   `json:"IsEncrypted"`
	IsDeviceData       bool   `json:"IsDeviceData"`
	VerificationStatus string `json:"VerificationStatus"`
}

// JSONOutput represents the complete JSON output structure
type JSONOutput struct {
	FirmwareFormat     string        `json:"FirmwareFormat"`
	OverallStatus      string        `json:"OverallStatus"`
	IsBootable         bool          `json:"IsBootable"`
	Sections           []JSONSection `json:"Sections"`
}

func displaySections(filePath string, format types.FirmwareFormat, sections map[uint16][]*interfaces.Section, parser *fs4.Parser, showContent bool, outputFormat string) error {
	// If JSON output is requested, we'll collect all sections first
	if outputFormat == "json" {
		// Collect all sections and convert to JSON format
		var jsonSections []JSONSection
		overallBootable := true
		
		for _, sectionList := range sections {
			for _, section := range sectionList {
				// Get section type name
				typeName := GetSectionTypeName(section)
				
				// Convert CRC type to string
				var crcTypeName string
				switch section.CRCType {
				case types.CRCNone:
					crcTypeName = "CRC_NONE"
				case types.CRCInITOCEntry:
					crcTypeName = "CRC_IN_ITOC_ENTRY"
				case types.CRCInSection:
					crcTypeName = "CRC_IN_SECTION"
				default:
					crcTypeName = fmt.Sprintf("UNKNOWN_%d", section.CRCType)
				}
				
				// Verify section and get status
				status, err := parser.VerifySection(section)
				if err != nil {
					logger.Warn("Failed to verify section", zap.Error(err))
					status = "ERROR"
					overallBootable = false
				}
				
				// Check if verification passed
				// SIZE NOT ALIGNED is still considered OK for bootability
				if status != "OK" && status != "CRC IGNORED" && status != "NO ENTRY" && 
				   status != "NOT FOUND" && status != "SIZE NOT ALIGNED" {
					overallBootable = false
				}
				
				jsonSection := JSONSection{
					Type:               typeName,
					StartAddress:       section.Offset,
					EndAddress:         section.Offset + uint64(section.Size) - 1,
					Size:               section.Size,
					CRCType:            crcTypeName,
					CRC:                section.CRC,
					IsEncrypted:        section.IsEncrypted,
					IsDeviceData:       section.IsDeviceData,
					VerificationStatus: status,
				}
				
				jsonSections = append(jsonSections, jsonSection)
			}
		}
		
		// Sort by start address for consistent output
		sort.Slice(jsonSections, func(i, j int) bool {
			return jsonSections[i].StartAddress < jsonSections[j].StartAddress
		})
		
		// Create overall output structure
		overallStatus := "FW image verification succeeded"
		if !overallBootable {
			overallStatus = "FW image verification failed"
		}
		
		output := JSONOutput{
			FirmwareFormat: format.String(),
			OverallStatus:  overallStatus,
			IsBootable:     overallBootable,
			Sections:       jsonSections,
		}
		
		// Output as JSON
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)
	}
	
	// Regular text output
	fmt.Printf("%s failsafe image\n\n", format.String())
	
	// Prepare sections for display
	var displaySections []SectionDisplay
	
	// Add HW pointers (always at beginning)
	for i := 0; i < 16; i++ {
		offset := uint64(0x18 + i*8)
		displaySections = append(displaySections, SectionDisplay{
			StartAddr:   offset,
			EndAddr:     offset + 7,
			Size:        8,
			Name:        "HW_POINTERS",
			Status:      "OK",
			IsHWPointer: true,
		})
	}
	
	// BOOT2 section is discovered dynamically from HW pointers in parser.go
	// It will be added from the parsed sections map below
	
	// Add ITOC header
	itocAddr := uint64(parser.GetITOCAddress())
	itocStatus := "OK"
	if !parser.IsITOCHeaderValid() {
		// For encrypted firmware, skip ITOC CRC verification as mstflint does
		if parser.IsEncrypted() {
			itocStatus = "ENCRYPTED"
		} else {
			itocStatus = "FAIL (Invalid CRC)"
		}
	}
	displaySections = append(displaySections, SectionDisplay{
		StartAddr: itocAddr,
		EndAddr:   itocAddr + 0x1F,
		Size:      0x20,
		Name:      "ITOC_HEADER",
		Status:    itocStatus,
		IsHeader:  true,
	})
	
	// Add DTOC header
	dtocAddr := uint64(parser.GetDTOCAddress())
	dtocStatus := "OK"
	if !parser.IsDTOCHeaderValid() {
		// For encrypted firmware, skip DTOC CRC verification as mstflint does
		if parser.IsEncrypted() {
			dtocStatus = "ENCRYPTED"
		} else {
			dtocStatus = "FAIL (Invalid CRC)"
		}
	}
	displaySections = append(displaySections, SectionDisplay{
		StartAddr: dtocAddr,
		EndAddr:   dtocAddr + 0x1F,
		Size:      0x20,
		Name:      "DTOC_HEADER",
		Status:    dtocStatus,
		IsHeader:  true,
	})
	
	// Add parsed sections
	for _, sectionList := range sections {
		for _, section := range sectionList {
			// Verify section
			status, err := parser.VerifySection(section)
			if err != nil {
				logger.Warn("Failed to verify section", zap.Error(err))
				status = "ERROR"
			}
			
			name := GetSectionTypeName(section)
			
			displaySections = append(displaySections, SectionDisplay{
				StartAddr: section.Offset,
				EndAddr:   section.Offset + uint64(section.Size) - 1,
				Size:      section.Size,
				Name:      name,
				Status:    status,
				Section:   section,
			})
		}
	}
	
	// Sort by start address
	sort.Slice(displaySections, func(i, j int) bool {
		return displaySections[i].StartAddr < displaySections[j].StartAddr
	})
	
	// Display sections and track if any failed
	hasFailures := false
	for _, ds := range displaySections {
		fmt.Printf("     /0x%08x-0x%08x (0x%06x)/ (%s) - %s\n",
			ds.StartAddr, ds.EndAddr, ds.Size, ds.Name, ds.Status)
		
		// Check if this section failed verification
		// Don't count ENCRYPTED status as a failure (matches mstflint behavior)
		if strings.Contains(ds.Status, "FAIL") || ds.Status == "ERROR" {
			hasFailures = true
		}
		
		if showContent && ds.Section != nil {
			// Show section content
			// Show basic section content - in real implementation would format
			content := fmt.Sprintf("      Section data: %d bytes", ds.Section.Size)
			fmt.Println(content)
		}
	}
	
	// Only report success if no failures were found
	if hasFailures {
		fmt.Println("\n-E- FW image verification failed. CRC errors detected.")
		return fmt.Errorf("firmware verification failed")
	} else {
		fmt.Println("\n-I- FW image verification succeeded. Image is bootable.")
	}
	
	return nil
}

