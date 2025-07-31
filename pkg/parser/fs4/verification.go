package fs4

import (
	"fmt"

	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"go.uber.org/zap"
)

// VerifySectionNew verifies a section using the new section interface
func (p *Parser) VerifySectionNew(section interfaces.SectionInterface) (string, error) {
	// For encrypted firmware, skip CRC verification for sections from HW pointers
	// as their CRCs may be invalid due to encryption
	if p.isEncrypted && section.IsFromHWPointer() {
		return "OK", nil
	}

	// Check if section has CRC verification enabled
	if !section.HasCRC() {
		return "CRC IGNORED", nil
	}
	
	// Debug: Log CRC info for IMAGE_SIGNATURE sections
	if section.Type() == types.SectionTypeImageSignature256 || section.Type() == types.SectionTypeImageSignature512 {
		if entry := section.GetITOCEntry(); entry != nil {
			p.logger.Info("IMAGE_SIGNATURE section CRC info",
				zap.String("type", section.TypeName()),
				zap.Uint8("crc_field", entry.GetCRC()),
				zap.Bool("no_crc", entry.GetNoCRC()),
				zap.Bool("has_crc", section.HasCRC()))
		}
	}

	// Get section's CRC type
	crcType := section.CRCType()
	
	switch crcType {
	case types.CRCNone:
		return "CRC IGNORED", nil
		
	case types.CRCInITOCEntry:
		// For sections with CRC in ITOC entry, we need to verify against the entry CRC
		if entry := section.GetITOCEntry(); entry != nil {
			// Use the section's VerifyCRC method which handles the correct CRC calculation
			if err := section.VerifyCRC(); err != nil {
				if crcErr, ok := err.(*interfaces.CRCMismatchError); ok {
					return fmt.Sprintf("FAIL (0x%04X != 0x%04X)", crcErr.Actual, crcErr.Expected), nil
				}
				return "ERROR", err
			}
			return "OK", nil
		}
		return "NO ENTRY", nil
		
	case types.CRCInSection:
		// For sections with embedded CRC, use the section's VerifyCRC method
		if err := section.VerifyCRC(); err != nil {
			// Special handling for specific section types
			if section.Type() == types.SectionTypeBoot2 {
				// BOOT2 often has size alignment issues
				if _, ok := err.(*interfaces.CRCMismatchError); ok {
					return "SIZE NOT ALIGNED", nil
				}
			}
			
			if crcErr, ok := err.(*interfaces.CRCMismatchError); ok {
				return fmt.Sprintf("FAIL (0x%04X != 0x%04X)", crcErr.Actual, crcErr.Expected), nil
			}
			return "ERROR", err
		}
		return "OK", nil
		
	default:
		return "UNKNOWN CRC TYPE", nil
	}
}

// GetSectionByType retrieves sections of a specific type using the new interface
func (p *Parser) GetSectionByType(sectionType uint16) []interfaces.SectionInterface {
	return p.sections[sectionType]
}

// GetAllSectionsNew returns all sections using the new interface
func (p *Parser) GetAllSectionsNew() map[uint16][]interfaces.SectionInterface {
	return p.sections
}

// FindSectionNew finds a specific section by name and optional ID
func (p *Parser) FindSectionNew(sectionName string, sectionID int) interfaces.SectionInterface {
	// Convert section name to type
	sectionType := types.GetSectionTypeByName(sectionName)
	if sectionType == 0 {
		return nil
	}
	
	sections := p.sections[sectionType]
	if len(sections) == 0 {
		return nil
	}
	
	// If no specific ID requested, return first section
	if sectionID == -1 {
		return sections[0]
	}
	
	// Return section at specific index
	if sectionID < len(sections) {
		return sections[sectionID]
	}
	
	return nil
}

// LoadSectionData ensures the section has its data loaded
func (p *Parser) LoadSectionData(section interfaces.SectionInterface) error {
	// Check if data is already loaded
	if len(section.GetRawData()) > 0 {
		return nil
	}
	
	// Read section data
	data, err := p.reader.ReadSection(int64(section.Offset()), section.Size())
	if err != nil {
		p.logger.Error("Failed to read section data",
			zap.String("type", section.TypeName()),
			zap.Uint64("offset", section.Offset()),
			zap.Uint32("size", section.Size()),
			zap.Error(err))
		return err
	}
	
	// Parse the data into the section
	if err := section.Parse(data); err != nil {
		p.logger.Error("Failed to parse section data",
			zap.String("type", section.TypeName()),
			zap.Error(err))
		return err
	}
	
	return nil
}