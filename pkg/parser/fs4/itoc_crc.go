package fs4

import (
	"encoding/binary"
	"fmt"
	
	"github.com/Civil/mlx5fw-go/pkg/parser"
)

// VerifyITOCHeaderCRC verifies the CRC of an ITOC header
// Based on mstflint's verifyTocHeader implementation
func VerifyITOCHeaderCRC(headerData []byte, crcCalculator *parser.CRCCalculator) error {
	if len(headerData) < 32 {
		return fmt.Errorf("ITOC header too small: %d bytes (expected 32)", len(headerData))
	}
	
	// Extract the stored CRC from the header (last 4 bytes)
	storedCRCDword := binary.BigEndian.Uint32(headerData[28:32])
	// Only the lower 16 bits are used for CRC
	storedCRC := uint16(storedCRCDword & 0xFFFF)
	
	// Calculate CRC on the header data excluding the CRC field
	// Coverage: TOC_HEADER_SIZE - 4 bytes = 28 bytes = 7 dwords
	sizeInDwords := 28 / 4
	calculatedCRC := crcCalculator.CalculateImageCRC(headerData[:28], sizeInDwords)
	
	// Check for blank CRC (0xFFFF indicates uninitialized)
	if storedCRC == 0xFFFF {
		return fmt.Errorf("blank ITOC header CRC (0xFFFF)")
	}
	
	// Compare calculated and stored CRC
	if calculatedCRC != storedCRC {
		return fmt.Errorf("ITOC header CRC mismatch: calculated=0x%04X, stored=0x%04X", 
			calculatedCRC, storedCRC)
	}
	
	return nil
}

// CalculateITOCHeaderCRC calculates the CRC for an ITOC header
// The CRC field (last 4 bytes) should be set to 0 before calling this
func CalculateITOCHeaderCRC(headerData []byte, crcCalculator *parser.CRCCalculator) uint16 {
	if len(headerData) < 32 {
		return 0
	}
	
	// Calculate CRC on first 28 bytes (excluding CRC field)
	sizeInDwords := 28 / 4
	return crcCalculator.CalculateImageCRC(headerData[:28], sizeInDwords)
}

// UpdateITOCHeaderCRC updates the CRC field in an ITOC header
func UpdateITOCHeaderCRC(headerData []byte, crcCalculator *parser.CRCCalculator) {
	if len(headerData) < 32 {
		return
	}
	
	// Clear the CRC field first
	binary.BigEndian.PutUint32(headerData[28:32], 0)
	
	// Calculate new CRC
	crc := CalculateITOCHeaderCRC(headerData, crcCalculator)
	
	// Store CRC in the header (as 32-bit value with CRC in lower 16 bits)
	binary.BigEndian.PutUint32(headerData[28:32], uint32(crc))
}