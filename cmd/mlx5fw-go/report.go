package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	cliutil "github.com/Civil/mlx5fw-go/pkg/cliutil"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/parser/fs4"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/spf13/cobra"
)

// CreateReportCommand creates the report command
func CreateReportCommand() *cobra.Command {
	reportCmd := &cobra.Command{
		Use:   "report",
		Short: "Generate a detailed report about firmware sections",
		Long: `Generate a comprehensive report about the sections in a firmware file.
	
The report includes:
- Section metadata (type, offset, size)
- Section attributes (encrypted, device-data, from-hw-pointer)
- CRC information
- Memory layout analysis
- Overlapping sections detection
- Summary statistics`,
		RunE: runReport,
	}

	reportCmd.Flags().StringP("output", "o", "", "Output file for the report (default: stdout)")
	reportCmd.Flags().BoolP("verbose", "v", false, "Include verbose section details")
	reportCmd.Flags().BoolP("summary", "s", false, "Show only summary information")
	reportCmd.Flags().StringP("type", "t", "", "Filter by section type (e.g., BOOT2, ITOC)")
	reportCmd.Flags().Bool("overlapping", false, "Show only overlapping sections")
	reportCmd.Flags().Bool("encrypted", false, "Show only encrypted sections")
	reportCmd.Flags().Bool("device-data", false, "Show only device-specific sections")

	return reportCmd
}

func runReport(cmd *cobra.Command, args []string) error {
	if err := cliutil.ValidateFirmwarePath(firmwarePath); err != nil {
		return err
	}

	// Get flags
	outputFile, _ := cmd.Flags().GetString("output")
	verbose, _ := cmd.Flags().GetBool("verbose")
	summaryOnly, _ := cmd.Flags().GetBool("summary")
	typeFilter, _ := cmd.Flags().GetString("type")
	showOverlapping, _ := cmd.Flags().GetBool("overlapping")
	encryptedOnly, _ := cmd.Flags().GetBool("encrypted")
	deviceDataOnly, _ := cmd.Flags().GetBool("device-data")

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

	// Get all sections
	sectionsMap := fwParser.GetSections()
	var allSections []interface{}
	for _, sectionList := range sectionsMap {
		for _, section := range sectionList {
			allSections = append(allSections, section)
		}
	}

	// Apply filters
	filteredSections := filterSections(allSections, typeFilter, encryptedOnly, deviceDataOnly)

	// Generate report
	var report strings.Builder

	if !summaryOnly {
		report.WriteString("Firmware Section Report\n")
		report.WriteString("======================\n")
		report.WriteString(fmt.Sprintf("File: %s\n", firmwarePath))
		report.WriteString(fmt.Sprintf("Total sections: %d\n", len(allSections)))
		if len(filteredSections) != len(allSections) {
			report.WriteString(fmt.Sprintf("Filtered sections: %d\n", len(filteredSections)))
		}
		report.WriteString("\n")
	}

	if showOverlapping {
		// Show overlapping sections
		if err := reportOverlappingSections(&report, filteredSections); err != nil {
			return err
		}
	} else if summaryOnly {
		// Show summary only
		if err := reportSummary(&report, filteredSections); err != nil {
			return err
		}
	} else {
		// Show detailed report
		if err := reportDetailed(&report, filteredSections, verbose); err != nil {
			return err
		}
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

func filterSections(sections []interface{}, typeFilter string, encryptedOnly, deviceDataOnly bool) []interface{} {
	var filtered []interface{}

	for _, s := range sections {
		section := s.(interface {
			TypeName() string
			IsEncrypted() bool
			IsDeviceData() bool
		})

		// Apply type filter
		if typeFilter != "" && !strings.Contains(strings.ToUpper(section.TypeName()), strings.ToUpper(typeFilter)) {
			continue
		}

		// Apply encrypted filter
		if encryptedOnly && !section.IsEncrypted() {
			continue
		}

		// Apply device data filter
		if deviceDataOnly && !section.IsDeviceData() {
			continue
		}

		filtered = append(filtered, s)
	}

	return filtered
}

func reportOverlappingSections(report *strings.Builder, sections []interface{}) error {
	// Convert to SectionMetadata interface
	var metadata []interface{}
	for _, s := range sections {
		metadata = append(metadata, s)
	}

	// Find overlapping sections
	overlapping := findOverlappingPairs(metadata)

	if len(overlapping) == 0 {
		report.WriteString("No overlapping sections found.\n")
		return nil
	}

	report.WriteString(fmt.Sprintf("Found %d overlapping section pairs:\n\n", len(overlapping)))

	for i, pair := range overlapping {
		s1 := pair[0].(interface {
			TypeName() string
			Offset() uint64
			Size() uint32
		})
		s2 := pair[1].(interface {
			TypeName() string
			Offset() uint64
			Size() uint32
		})

		report.WriteString(fmt.Sprintf("Overlap %d:\n", i+1))
		report.WriteString(fmt.Sprintf("  Section 1: %s at 0x%08x (size: 0x%x)\n",
			s1.TypeName(), s1.Offset(), s1.Size()))
		report.WriteString(fmt.Sprintf("  Section 2: %s at 0x%08x (size: 0x%x)\n",
			s2.TypeName(), s2.Offset(), s2.Size()))

		// Calculate overlap region
		start := s1.Offset()
		if s2.Offset() > start {
			start = s2.Offset()
		}
		end := s1.Offset() + uint64(s1.Size())
		if s2.Offset()+uint64(s2.Size()) < end {
			end = s2.Offset() + uint64(s2.Size())
		}
		report.WriteString(fmt.Sprintf("  Overlap region: 0x%08x - 0x%08x (size: 0x%x)\n\n",
			start, end, end-start))
	}

	return nil
}

func reportSummary(report *strings.Builder, sections []interface{}) error {
	// Group by type
	groups := make(map[string][]interface{})
	var totalSize uint64
	encryptedCount := 0
	deviceDataCount := 0
	hwPointerCount := 0

	for _, s := range sections {
		section := s.(interface {
			TypeName() string
			Size() uint32
			IsEncrypted() bool
			IsDeviceData() bool
			IsFromHWPointer() bool
		})

		typeName := section.TypeName()
		groups[typeName] = append(groups[typeName], s)
		totalSize += uint64(section.Size())

		if section.IsEncrypted() {
			encryptedCount++
		}
		if section.IsDeviceData() {
			deviceDataCount++
		}
		if section.IsFromHWPointer() {
			hwPointerCount++
		}
	}

	// Sort type names
	var typeNames []string
	for name := range groups {
		typeNames = append(typeNames, name)
	}
	sort.Strings(typeNames)

	report.WriteString("Section Type Summary:\n")
	report.WriteString("--------------------\n")
	for _, typeName := range typeNames {
		typeSections := groups[typeName]
		var typeSize uint64
		for _, s := range typeSections {
			section := s.(interface{ Size() uint32 })
			typeSize += uint64(section.Size())
		}
		report.WriteString(fmt.Sprintf("%-20s: %3d sections, %10d bytes (%.1f%%)\n",
			typeName, len(typeSections), typeSize, float64(typeSize)/float64(totalSize)*100))
	}

	report.WriteString("\nStatistics:\n")
	report.WriteString("-----------\n")
	report.WriteString(fmt.Sprintf("Total sections:      %d\n", len(sections)))
	report.WriteString(fmt.Sprintf("Total size:          %d bytes (0x%x)\n", totalSize, totalSize))
	report.WriteString(fmt.Sprintf("Encrypted sections:  %d\n", encryptedCount))
	report.WriteString(fmt.Sprintf("Device data sections: %d\n", deviceDataCount))
	report.WriteString(fmt.Sprintf("HW pointer sections: %d\n", hwPointerCount))

	return nil
}

func reportDetailed(report *strings.Builder, sections []interface{}, verbose bool) error {
	// Sort sections by offset
	sort.Slice(sections, func(i, j int) bool {
		si := sections[i].(interface{ Offset() uint64 })
		sj := sections[j].(interface{ Offset() uint64 })
		return si.Offset() < sj.Offset()
	})

	// Group by type for organized output
	groups := make(map[string][]interface{})
	for _, s := range sections {
		section := s.(interface{ TypeName() string })
		typeName := section.TypeName()
		groups[typeName] = append(groups[typeName], s)
	}

	// Sort type names
	var typeNames []string
	for name := range groups {
		typeNames = append(typeNames, name)
	}
	sort.Strings(typeNames)

	// Generate section listings by type
	for _, typeName := range typeNames {
		typeSections := groups[typeName]
		report.WriteString(fmt.Sprintf("\n%s Sections (%d):\n", typeName, len(typeSections)))
		report.WriteString(strings.Repeat("-", len(typeName)+20) + "\n")

		for i, s := range typeSections {
			section := s.(interface {
				TypeName() string
				Type() uint16
				Offset() uint64
				Size() uint32
				IsEncrypted() bool
				IsDeviceData() bool
				IsFromHWPointer() bool
				HasCRC() bool
				GetCRC() uint32
				CRCType() types.CRCType
			})

			report.WriteString(fmt.Sprintf("[%d] Type: 0x%04x, Offset: 0x%08x, Size: 0x%08x (%d bytes)\n",
				i, section.Type(), section.Offset(), section.Size(), section.Size()))

			if verbose {
				var attributes []string
				if section.IsEncrypted() {
					attributes = append(attributes, "Encrypted")
				}
				if section.IsDeviceData() {
					attributes = append(attributes, "Device Data")
				}
				if section.IsFromHWPointer() {
					attributes = append(attributes, "From HW Pointer")
				}

				if len(attributes) > 0 {
					report.WriteString(fmt.Sprintf("    Attributes: %s\n", strings.Join(attributes, ", ")))
				}

				if section.HasCRC() {
					report.WriteString(fmt.Sprintf("    CRC: 0x%08x (Type: %s)\n",
						section.GetCRC(), section.CRCType()))
				}
			}
		}
	}

	// Memory layout visualization
	if verbose {
		report.WriteString("\n\nMemory Layout:\n")
		report.WriteString("--------------\n")

		lastEnd := uint64(0)
		for _, s := range sections {
			section := s.(interface {
				TypeName() string
				Offset() uint64
				Size() uint32
			})

			// Show gap if any
			if section.Offset() > lastEnd {
				gapSize := section.Offset() - lastEnd
				report.WriteString(fmt.Sprintf("  [GAP: 0x%08x - 0x%08x, size: 0x%x]\n",
					lastEnd, section.Offset(), gapSize))
			}

			report.WriteString(fmt.Sprintf("  0x%08x - 0x%08x: %-20s (0x%x bytes)\n",
				section.Offset(), section.Offset()+uint64(section.Size()),
				section.TypeName(), section.Size()))

			lastEnd = section.Offset() + uint64(section.Size())
		}
	}

	return nil
}

func findOverlappingPairs(sections []interface{}) [][]interface{} {
	var overlapping [][]interface{}

	for i := 0; i < len(sections); i++ {
		for j := i + 1; j < len(sections); j++ {
			s1 := sections[i].(interface {
				Offset() uint64
				Size() uint32
			})
			s2 := sections[j].(interface {
				Offset() uint64
				Size() uint32
			})

			s1End := s1.Offset() + uint64(s1.Size())
			s2End := s2.Offset() + uint64(s2.Size())

			// Check if sections overlap
			if s1.Offset() < s2End && s2.Offset() < s1End {
				overlapping = append(overlapping, []interface{}{sections[i], sections[j]})
			}
		}
	}

	return overlapping
}
