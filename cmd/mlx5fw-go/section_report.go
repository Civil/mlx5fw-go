package main

import (
	"fmt"
	"os"
	"strings"

	cliutil "github.com/Civil/mlx5fw-go/pkg/cliutil"
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/parser/fs4"
	"github.com/Civil/mlx5fw-go/pkg/utils"
	"github.com/spf13/cobra"
)

// CreateSectionReportCommand creates the section-report command
func CreateSectionReportCommand() *cobra.Command {
	sectionReportCmd := &cobra.Command{
		Use:   "section-report",
		Short: "Generate a section report using section utilities",
		Long: `Generate a section report using the section utility functions.
	
This command demonstrates the use of the smaller, focused interfaces
for working with firmware sections.`,
		RunE: runSectionReport,
	}

	sectionReportCmd.Flags().StringP("output", "o", "", "Output file for the report (default: stdout)")
	sectionReportCmd.Flags().String("format", "detailed", "Report format: detailed, summary, list")

	return sectionReportCmd
}

func runSectionReport(cmd *cobra.Command, args []string) error {
	if err := cliutil.ValidateFirmwarePath(firmwarePath); err != nil {
		return err
	}
	outputFile, _ := cmd.Flags().GetString("output")
	format, _ := cmd.Flags().GetString("format")

	// Create firmware reader
	reader, err := parser.NewFirmwareReader(firmwarePath, logger)
	if err != nil {
		return fmt.Errorf("failed to create firmware reader: %w", err)
	}
	defer reader.Close()

	// Parse firmware
	fwParser := fs4.NewParser(reader, logger)

	if err := fwParser.Parse(); err != nil {
		return fmt.Errorf("failed to parse firmware: %w", err)
	}

	// Get all sections and convert to interfaces
	sectionsMap := fwParser.GetSections()
	var allSections []interfaces.CompleteSectionInterface
	var readers []interfaces.SectionReader
	var metadata []interfaces.SectionMetadata
	var attributes []interfaces.SectionAttributes

	for _, sectionList := range sectionsMap {
		allSections = append(allSections, sectionList...)
		for _, section := range sectionList {
			readers = append(readers, section)
			metadata = append(metadata, section)
			attributes = append(attributes, section)
		}
	}

	// Generate report based on format
	var report strings.Builder

	switch format {
	case "summary":
		report.WriteString(generateSummaryReport(metadata, attributes))
	case "list":
		report.WriteString(generateListReport(readers))
	default:
		report.WriteString(generateDetailedReport(readers, metadata))
	}

	// Output report
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(report.String()), 0644); err != nil {
			return fmt.Errorf("failed to write report file: %w", err)
		}
		fmt.Printf("Report written to %s\n", outputFile)
	} else {
		fmt.Print(report.String())
	}

	return nil
}

func generateSummaryReport(metadata []interfaces.SectionMetadata, attributes []interfaces.SectionAttributes) string {
	var report strings.Builder

	report.WriteString("Section Summary Report\n")
	report.WriteString("=====================\n\n")

	// Basic statistics
	report.WriteString(fmt.Sprintf("Total sections: %d\n", len(metadata)))
	report.WriteString(fmt.Sprintf("Total size: 0x%x bytes\n", utils.CalculateTotalSize(metadata)))

	// Group by type
	groups := utils.GroupSectionsByType(metadata)
	report.WriteString(fmt.Sprintf("\nSection types: %d\n", len(groups)))
	report.WriteString("\nSections by type:\n")
	for typeName, sections := range groups {
		report.WriteString(fmt.Sprintf("  %-20s: %d sections\n", typeName, len(sections)))
	}

	// Count encrypted sections
	encryptedSections := utils.FilterEncryptedSections(attributes)
	report.WriteString(fmt.Sprintf("\nEncrypted sections: %d\n", len(encryptedSections)))

	// Check for overlapping sections
	overlapping := utils.FindOverlappingSections(metadata)
	report.WriteString(fmt.Sprintf("Overlapping section pairs: %d\n", len(overlapping)))

	return report.String()
}

func generateListReport(readers []interfaces.SectionReader) string {
	var report strings.Builder

	report.WriteString("Section List\n")
	report.WriteString("============\n\n")

	// Get formatted list
	lines := utils.FormatSectionList(readers, "  ")
	for _, line := range lines {
		report.WriteString(line + "\n")
	}

	return report.String()
}

func generateDetailedReport(readers []interfaces.SectionReader, metadata []interfaces.SectionMetadata) string {
	var report strings.Builder

	// Use the utility function for the main report
	report.WriteString(utils.GenerateSectionReport(readers))

	// Add validation results
	report.WriteString("\nValidation Results\n")
	report.WriteString("==================\n")

	validCount := 0
	for _, reader := range readers {
		if utils.IsSectionValid(reader) {
			validCount++
		}

		// Check CRC info consistency
		if err := utils.ValidateCRCInfo(reader); err != nil {
			report.WriteString(fmt.Sprintf("CRC validation error for %s: %v\n",
				utils.GetSectionDescription(reader), err))
		}
	}

	report.WriteString(fmt.Sprintf("\nValid sections: %d/%d\n", validCount, len(readers)))

	// Check for overlaps
	overlapping := utils.FindOverlappingSections(metadata)
	if len(overlapping) > 0 {
		report.WriteString("\nOverlapping Sections\n")
		report.WriteString("====================\n")
		for i, pair := range overlapping {
			report.WriteString(fmt.Sprintf("\nOverlap %d:\n", i+1))
			report.WriteString(fmt.Sprintf("  Section 1: %s\n", utils.GetSectionDescription(pair[0])))
			report.WriteString(fmt.Sprintf("  Section 2: %s\n", utils.GetSectionDescription(pair[1])))
		}
	}

	return report.String()
}
