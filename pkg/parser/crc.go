package parser

import (
	"github.com/Civil/mlx5fw-go/pkg/types"
)

// CRCCalculator provides CRC calculation methods
type CRCCalculator struct{}

// NewCRCCalculator creates a new CRC calculator
func NewCRCCalculator() *CRCCalculator {
	return &CRCCalculator{}
}

// CalculateSoftwareCRC16 calculates software CRC16 using polynomial 0x100b
// This matches mstflint's implementation for FS4 sections
func (c *CRCCalculator) CalculateSoftwareCRC16(data []byte) uint16 {
	// Process data as 32-bit words (big-endian)
	// Pad data to align to 4 bytes if needed
	dataLen := len(data)
	paddedLen := (dataLen + 3) & ^3 // Round up to multiple of 4
	paddedData := make([]byte, paddedLen)
	copy(paddedData, data)
	
	crc := uint16(0xFFFF)
	
	// Process 32-bit words
	for i := 0; i < paddedLen; i += 4 {
		// Get 32-bit word in big-endian
		word := uint32(paddedData[i])<<24 | uint32(paddedData[i+1])<<16 | 
			uint32(paddedData[i+2])<<8 | uint32(paddedData[i+3])
		
		// Process each bit of the 32-bit word (matches mstflint's Crc16::add)
		for j := 0; j < 32; j++ {
			if crc & 0x8000 != 0 {
				crc = ((crc << 1) | uint16(word >> 31)) ^ types.CRCPolynomial
			} else {
				crc = (crc << 1) | uint16(word >> 31)
			}
			crc &= 0xFFFF
			word <<= 1
		}
	}
	
	// Finish step - process 16 more bits of zeros
	for i := 0; i < 16; i++ {
		if crc&0x8000 != 0 {
			crc = (crc << 1) ^ types.CRCPolynomial
		} else {
			crc = crc << 1
		}
		crc &= 0xFFFF
	}
	
	// Final XOR
	return crc ^ 0xFFFF
}

// CalculateHardwareCRC calculates hardware CRC using the special table
func (c *CRCCalculator) CalculateHardwareCRC(data []byte) uint16 {
	if len(data) < 2 {
		return 0
	}
	
	// Hardware CRC inverts first 2 bytes
	modifiedData := make([]byte, len(data))
	copy(modifiedData, data)
	modifiedData[0] = ^modifiedData[0]
	modifiedData[1] = ^modifiedData[1]
	
	crc := uint16(0xFFFF)
	for _, b := range modifiedData {
		tbl_idx := ((crc >> 8) ^ uint16(b)) & 0xFF
		crc = ((crc << 8) ^ types.CRC16Table2[tbl_idx]) & 0xFFFF
	}
	
	return crc
}

// CalculateImageCRC calculates CRC16 on image data (matches mstflint's CalcImageCRC)
// data: byte array containing the data
// sizeInDwords: number of 32-bit words to process
func (c *CRCCalculator) CalculateImageCRC(data []byte, sizeInDwords int) uint16 {
	crc := uint16(0xFFFF)
	
	// Process each dword
	for i := 0; i < sizeInDwords; i++ {
		offset := i * 4
		if offset+4 > len(data) {
			break
		}
		
		// Get 32-bit word in big-endian
		word := uint32(data[offset])<<24 | uint32(data[offset+1])<<16 | 
			uint32(data[offset+2])<<8 | uint32(data[offset+3])
		
		// Process each bit of the 32-bit word (matches mstflint's Crc16::add)
		for j := 0; j < 32; j++ {
			if crc & 0x8000 != 0 {
				crc = ((crc << 1) | uint16(word >> 31)) ^ types.CRCPolynomial
			} else {
				crc = (crc << 1) | uint16(word >> 31)
			}
			crc &= 0xFFFF
			word <<= 1
		}
	}
	
	// Finish step - process 16 more bits of zeros
	for i := 0; i < 16; i++ {
		if crc&0x8000 != 0 {
			crc = (crc << 1) ^ types.CRCPolynomial
		} else {
			crc = crc << 1
		}
		crc &= 0xFFFF
	}
	
	// Final XOR
	return crc ^ 0xFFFF
}

// VerifyCRC verifies CRC for a data buffer
func (c *CRCCalculator) VerifyCRC(data []byte, expectedCRC uint16, useHardwareCRC bool) bool {
	var calculatedCRC uint16
	
	if useHardwareCRC {
		calculatedCRC = c.CalculateHardwareCRC(data)
	} else {
		calculatedCRC = c.CalculateSoftwareCRC16(data)
	}
	
	return calculatedCRC == expectedCRC
}