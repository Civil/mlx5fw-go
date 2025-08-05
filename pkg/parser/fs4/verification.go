package fs4

import (
	"errors"
	"fmt"

	pkgerrors "github.com/Civil/mlx5fw-go/pkg/errors"
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"go.uber.org/zap"
)

// VerifySectionNew verifies a section using the new section interface
func (p *Parser) VerifySectionNew(section interfaces.SectionVerifier) (string, error) {
	// For encrypted firmware, skip CRC verification for sections from HW pointers
	// as their CRCs may be invalid due to encryption
	if p.isEncrypted && section.IsFromHWPointer() {
		return "OK", nil
	}

	// Check if section has CRC verification enabled
	if !section.HasCRC() {
		return "CRC IGNORED", nil
	}
	
	// Debug logging for sections that should have no CRC
	if entry := section.GetITOCEntry(); entry != nil && entry.GetNoCRC() {
		p.logger.Warn("Section has no_crc flag but HasCRC() returned true",
			zap.String("type", section.TypeName()),
			zap.Uint16("type_id", section.Type()),
			zap.Bool("has_crc", section.HasCRC()),
			zap.Bool("no_crc", entry.GetNoCRC()),
			zap.Uint8("crc_field", entry.GetCRC()))
	}
	
	// Load section data if not already loaded
	if section.GetRawData() == nil || len(section.GetRawData()) == 0 {
		// Special case: zero-length sections don't need data loading
		if section.Size() == 0 {
			// Set empty data so verification can proceed
			if err := section.Parse([]byte{}); err != nil {
				return "ERROR", fmt.Errorf("failed to parse zero-length section: %w", err)
			}
		} else {
			// Read size is just the section size
			// The CRC is already included in the section size for CRCInSection types
			readSize := section.Size()
			
			// Read section data
			data, err := p.reader.ReadSection(int64(section.Offset()), readSize)
			if err != nil {
				return "ERROR", fmt.Errorf("failed to read section data: %w", err)
			}
			
			// Parse the section data
			if err := section.Parse(data); err != nil {
				// Debug for specific sections
				if section.Type() == types.SectionTypeResetInfo || 
				   section.Type() == types.SectionTypeDigitalCertPtr {
					p.logger.Error("Section parse failed",
						zap.String("type", section.TypeName()),
						zap.Error(err),
						zap.Uint32("size", section.Size()),
						zap.Uint64("offset", section.Offset()),
						zap.Int("data_len", len(data)))
				}
				return "ERROR", fmt.Errorf("failed to parse section data: %w", err)
			}
		}
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
	
	// Debug: Log sections with CRC type mismatch
	if crcType != types.CRCNone {
		if entry := section.GetITOCEntry(); entry != nil && entry.GetNoCRC() {
			p.logger.Error("CRC type mismatch - section has no_crc but CRC type is not NONE",
				zap.String("type", section.TypeName()),
				zap.String("crc_type", crcType.String()),
				zap.Bool("no_crc", entry.GetNoCRC()),
				zap.Uint8("crc_field", entry.GetCRC()))
		}
	}
	
	switch crcType {
	case types.CRCNone:
		return "CRC IGNORED", nil
		
	case types.CRCInITOCEntry:
		// For sections with CRC in ITOC entry, we need to verify against the entry CRC
		if entry := section.GetITOCEntry(); entry != nil {
			// Debug logging for specific section types
			if section.Type() == types.SectionTypeResetInfo || 
			   section.Type() == types.SectionTypeDigitalCertPtr ||
			   section.Type() == types.SectionTypeHashesTable {
				p.logger.Info("Verifying CRC for section",
					zap.String("type", section.TypeName()),
					zap.Uint32("expected_crc", section.GetCRC()),
					zap.Uint32("size", section.Size()),
					zap.String("crc_type", crcType.String()))
			}
			
			// Use the section's VerifyCRC method which handles the correct CRC calculation
			if err := section.VerifyCRC(); err != nil {
				if errors.Is(err, pkgerrors.ErrCRCMismatch) {
					if crcData, ok := pkgerrors.GetCRCMismatchData(err); ok {
						return fmt.Sprintf("FAIL (0x%04X != 0x%04X)", crcData.Actual, crcData.Expected), nil
					}
					return "FAIL", nil
				}
				return "ERROR", err
			}
			return "OK", nil
		}
		return "NO ENTRY", nil
		
	case types.CRCInSection:
		// For sections with embedded CRC, use the section's VerifyCRC method
		if err := section.VerifyCRC(); err != nil {
			if errors.Is(err, pkgerrors.ErrCRCMismatch) {
				if crcData, ok := pkgerrors.GetCRCMismatchData(err); ok {
					return fmt.Sprintf("FAIL (0x%04X != 0x%04X)", crcData.Actual, crcData.Expected), nil
				}
				return "FAIL", nil
			}
			return "ERROR", err
		}
		return "OK", nil
		
	default:
		return "UNKNOWN CRC TYPE", nil
	}
}

// GetSectionByType retrieves sections of a specific type using the new interface
func (p *Parser) GetSectionByType(sectionType uint16) []interfaces.CompleteSectionInterface {
	return p.sections[sectionType]
}

// GetAllSectionsNew returns all sections using the new interface
func (p *Parser) GetAllSectionsNew() map[uint16][]interfaces.CompleteSectionInterface {
	return p.sections
}

// FindSectionNew finds a specific section by name and optional ID
func (p *Parser) FindSectionNew(sectionName string, sectionID int) interfaces.CompleteSectionInterface {
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
func (p *Parser) LoadSectionData(section interfaces.SectionParser) error {
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