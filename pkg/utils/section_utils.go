package utils

import (
	"fmt"
	"strings"

	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
)

// GetSectionDescription returns a formatted description of a section using only metadata
func GetSectionDescription(section interfaces.SectionMetadata) string {
	return fmt.Sprintf("%s at offset 0x%x (size: 0x%x)",
		section.TypeName(), section.Offset(), section.Size())
}

// IsSectionValid checks if a section has valid basic properties
// Note: This is a basic check. For full CRC validation, use parser.VerifySectionNew
func IsSectionValid(section interfaces.SectionReader) bool {
	// Basic validity checks
	if section.Size() == 0 && section.Type() != types.SectionTypeVpdR0 {
		// VPD_R0 can have size 0
		return false
	}
	if section.Type() == 0 {
		return false
	}

	// For sections without CRC, they're considered valid if basic checks pass
	if !section.HasCRC() {
		return true
	}

	// For sections with CRC, we can't fully validate without loading data
	// This is just a basic sanity check
	return true
}

// FilterEncryptedSections returns only encrypted sections
func FilterEncryptedSections(sections []interfaces.SectionAttributes) []interfaces.SectionAttributes {
	var encrypted []interfaces.SectionAttributes
	for _, section := range sections {
		if section.IsEncrypted() {
			encrypted = append(encrypted, section)
		}
	}
	return encrypted
}

// GetSectionSummary returns a summary of section metadata
func GetSectionSummary(sections []interfaces.SectionReader) []string {
	var summaries []string
	for _, section := range sections {
		summary := fmt.Sprintf("Type: %s, Offset: 0x%08x, Size: 0x%x, Encrypted: %v",
			section.TypeName(),
			section.Offset(),
			section.Size(),
			section.IsEncrypted())
		summaries = append(summaries, summary)
	}
	return summaries
}

// GroupSectionsByType groups sections by their type name
func GroupSectionsByType(sections []interfaces.SectionMetadata) map[string][]interfaces.SectionMetadata {
	groups := make(map[string][]interfaces.SectionMetadata)
	for _, section := range sections {
		typeName := section.TypeName()
		groups[typeName] = append(groups[typeName], section)
	}
	return groups
}

// ValidateCRCInfo provides basic CRC info validation
// Note: This does NOT perform actual CRC verification. For that, use parser.VerifySectionNew
func ValidateCRCInfo(section interfaces.SectionReader) error {
	// If section doesn't have CRC, no validation needed
	if !section.HasCRC() {
		return nil
	}

	// For sections with CRC, we can't validate the actual CRC without data
	// Some sections (like BOOT2, TOOLS_AREA, DEV_INFO) may have CRC value of 0
	// which is computed differently. This is not an error.

	// Basic consistency check
	if section.CRCType() == types.CRCNone && section.HasCRC() {
		return fmt.Errorf("section reports HasCRC() but CRCType is NONE")
	}

	return nil
}

// FormatSectionList formats a list of sections for display
func FormatSectionList(sections []interfaces.SectionReader, indent string) []string {
	var lines []string
	for i, section := range sections {
		line := fmt.Sprintf("%s[%d] %s", indent, i, GetSectionDescription(section))
		if section.IsEncrypted() {
			line += " [ENCRYPTED]"
		}
		if section.IsDeviceData() {
			line += " [DEVICE_DATA]"
		}
		if section.IsFromHWPointer() {
			line += " [FROM_HW_PTR]"
		}
		lines = append(lines, line)
	}
	return lines
}

// CalculateTotalSize calculates the total size of all sections
func CalculateTotalSize(sections []interfaces.SectionMetadata) uint64 {
	var total uint64
	for _, section := range sections {
		total += uint64(section.Size())
	}
	return total
}

// FindOverlappingSections finds sections that overlap in memory
func FindOverlappingSections(sections []interfaces.SectionMetadata) [][]interfaces.SectionMetadata {
	var overlapping [][]interfaces.SectionMetadata

	for i := 0; i < len(sections); i++ {
		for j := i + 1; j < len(sections); j++ {
			s1, s2 := sections[i], sections[j]
			s1End := s1.Offset() + uint64(s1.Size())
			s2End := s2.Offset() + uint64(s2.Size())

			// Check if sections overlap
			if s1.Offset() < s2End && s2.Offset() < s1End {
				overlapping = append(overlapping, []interfaces.SectionMetadata{s1, s2})
			}
		}
	}

	return overlapping
}

// GenerateSectionReport generates a detailed report about sections
func GenerateSectionReport(sections []interfaces.SectionReader) string {
	var report strings.Builder

	report.WriteString("Section Report\n")
	report.WriteString("==============\n\n")

	// Group by type
	groups := make(map[string][]interfaces.SectionReader)
	for _, section := range sections {
		typeName := section.TypeName()
		groups[typeName] = append(groups[typeName], section)
	}

	// Report by type
	for typeName, typeSections := range groups {
		report.WriteString(fmt.Sprintf("%s (%d sections):\n", typeName, len(typeSections)))
		for _, section := range typeSections {
			report.WriteString(fmt.Sprintf("  - Offset: 0x%08x, Size: 0x%x",
				section.Offset(), section.Size()))

			if section.HasCRC() {
				report.WriteString(fmt.Sprintf(", CRC: 0x%08x", section.GetCRC()))
			}

			var flags []string
			if section.IsEncrypted() {
				flags = append(flags, "encrypted")
			}
			if section.IsDeviceData() {
				flags = append(flags, "device-data")
			}
			if section.IsFromHWPointer() {
				flags = append(flags, "hw-pointer")
			}

			if len(flags) > 0 {
				report.WriteString(fmt.Sprintf(" [%s]", strings.Join(flags, ", ")))
			}

			report.WriteString("\n")
		}
		report.WriteString("\n")
	}

	// Summary
	report.WriteString(fmt.Sprintf("Total sections: %d\n", len(sections)))
	report.WriteString(fmt.Sprintf("Total size: 0x%x bytes\n", CalculateTotalSize(toMetadataSlice(sections))))

	return report.String()
}

// Helper function to convert SectionReader slice to SectionMetadata slice
func toMetadataSlice(readers []interfaces.SectionReader) []interfaces.SectionMetadata {
	metadata := make([]interfaces.SectionMetadata, len(readers))
	for i, r := range readers {
		metadata[i] = r
	}
	return metadata
}
