package section

import (
	"encoding/binary"
	"fmt"
	
	"github.com/ansel1/merry/v2"
	"go.uber.org/zap"
	
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/parser/fs4"
	"github.com/Civil/mlx5fw-go/pkg/types"
)

// ReplacerNew handles section replacement with new section interfaces
type ReplacerNew struct {
	parser   *fs4.Parser
	firmware []byte
	logger   *zap.Logger
}

// NewReplacerNew creates a new section replacer
func NewReplacerNew(parser *fs4.Parser, firmware []byte, logger *zap.Logger) *ReplacerNew {
	return &ReplacerNew{
		parser:   parser,
		firmware: firmware,
		logger:   logger,
	}
}

// ReplaceSection replaces a section with new data using the new interface
func (r *ReplacerNew) ReplaceSection(section interfaces.CompleteSectionInterface, newData []byte) ([]byte, error) {
	r.logger.Info("Replacing section",
		zap.String("type", section.TypeName()),
		zap.Uint64("offset", section.Offset()),
		zap.Uint32("oldSize", section.Size()),
		zap.Int("newSize", len(newData)))
	
	// Validate the new data size
	if len(newData) == 0 {
		return nil, merry.New("replacement data cannot be empty")
	}
	
	// Check if section has CRC that needs to be updated
	if section.CRCType() == types.CRCInSection {
		// For sections with embedded CRC, we need to calculate and append CRC
		if _, err := section.CalculateCRC(); err == nil {
			// The CRC handler knows how to calculate CRC for the specific section type
			// For in-section CRC, it's typically at the end
			newDataWithCRC := make([]byte, len(newData))
			copy(newDataWithCRC, newData)
			
			// Ensure data is properly padded for CRC calculation
			if len(newDataWithCRC) % 4 != 0 {
				padding := 4 - (len(newDataWithCRC) % 4)
				newDataWithCRC = append(newDataWithCRC, make([]byte, padding)...)
			}
			
			// Calculate CRC on the new data
			// Note: This is a simplified version - real implementation would use the section's CRC handler
			crc := r.calculateCRC16(newDataWithCRC)
			
			// Append CRC at the end (big-endian)
			crcBytes := make([]byte, 4)
			binary.BigEndian.PutUint32(crcBytes, uint32(crc))
			newData = append(newDataWithCRC, crcBytes...)
		}
	}
	
	// Create a copy of the firmware
	modifiedFirmware := make([]byte, len(r.firmware))
	copy(modifiedFirmware, r.firmware)
	
	// Replace the section data
	offset := section.Offset()
	oldSize := section.Size()
	newSize := uint32(len(newData))
	
	if newSize != oldSize {
		// Size changed - need to adjust firmware
		// This is more complex and would require updating ITOC/DTOC entries
		// For now, we'll only support same-size replacements
		return nil, merry.New(fmt.Sprintf("section size change not yet supported (old: %d, new: %d)", oldSize, newSize))
	}
	
	// Simple case: same size replacement
	copy(modifiedFirmware[offset:], newData)
	
	// If section has CRC in ITOC entry, update it
	if section.CRCType() == types.CRCInITOCEntry && section.GetITOCEntry() != nil {
		// Calculate new CRC
		crc := r.calculateCRC16(newData)
		
		// Find and update the ITOC entry
		// This would require parsing the ITOC and updating the specific entry
		// For now, we'll log a warning
		r.logger.Warn("ITOC entry CRC update not yet implemented",
			zap.String("section", section.TypeName()),
			zap.Uint16("newCRC", crc))
	}
	
	return modifiedFirmware, nil
}

// calculateCRC16 calculates CRC16 using the standard polynomial
func (r *ReplacerNew) calculateCRC16(data []byte) uint16 {
	// This is a simplified version - should use the parser's CRC calculator
	crc := uint16(0xFFFF)
	
	// Process data as 32-bit words
	paddedLen := (len(data) + 3) & ^3
	paddedData := make([]byte, paddedLen)
	copy(paddedData, data)
	
	for i := 0; i < paddedLen; i += 4 {
		word := binary.BigEndian.Uint32(paddedData[i:])
		
		for j := 0; j < 32; j++ {
			if crc&0x8000 != 0 {
				crc = ((crc << 1) | uint16(word>>31)) ^ types.CRCPolynomial
			} else {
				crc = (crc << 1) | uint16(word>>31)
			}
			crc &= 0xFFFF
			word <<= 1
		}
	}
	
	// Finish
	for i := 0; i < 16; i++ {
		if crc&0x8000 != 0 {
			crc = (crc << 1) ^ types.CRCPolynomial
		} else {
			crc = crc << 1
		}
		crc &= 0xFFFF
	}
	
	return crc ^ 0xFFFF
}