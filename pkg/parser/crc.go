package parser

import (
	"github.com/Civil/mlx5fw-go/pkg/types"
)

// CRCCalculator provides CRC calculation methods
type CRCCalculator struct {
	crcType types.CRCType
}

// NewCRCCalculator creates a new CRC calculator
func NewCRCCalculator() *CRCCalculator {
	return &CRCCalculator{
		crcType: types.CRCInITOCEntry, // Default to software CRC
	}
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
			if crc&0x8000 != 0 {
				crc = ((crc << 1) | uint16(word>>31)) ^ types.CRCPolynomial
			} else {
				crc = (crc << 1) | uint16(word>>31)
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
// This matches mstflint's calc_hw_crc implementation
func (c *CRCCalculator) CalculateHardwareCRC(data []byte) uint16 {
	if len(data) < 2 {
		return 0
	}

	crc := uint16(0xFFFF)
	for i := 0; i < len(data); i++ {
		var d byte
		if i > 1 {
			d = data[i]
		} else {
			d = ^data[i] // Invert first 2 bytes
		}
		tableIndex := (crc ^ uint16(d)) & 0xFF
		crc = (crc >> 8) ^ types.CRC16Table2[tableIndex]
	}

	// Swap bytes
	crc = ((crc << 8) & 0xFF00) | ((crc >> 8) & 0xFF)

	return crc
}

// CalculateImageCRC calculates CRC16 on image data (matches mstflint's CalcImageCRC)
// data: byte array containing the data
// sizeInDwords: number of 32-bit words to process
func (c *CRCCalculator) CalculateImageCRC(data []byte, sizeInDwords int) uint16 {
	// Special case for zero-length data
	if sizeInDwords == 0 {
		// Based on mstflint investigation, CalcImageCRC(NULL, 0) returns 0x955 (2389)
		return 0x955
	}

	// Create a copy of data to avoid modifying the original
	dataCopy := make([]byte, sizeInDwords*4)
	copy(dataCopy, data)

	// Convert from big-endian to host-endian (mstflint uses TOCPUn)
	for i := 0; i < sizeInDwords; i++ {
		offset := i * 4
		// Read big-endian dword
		word := uint32(dataCopy[offset])<<24 | uint32(dataCopy[offset+1])<<16 |
			uint32(dataCopy[offset+2])<<8 | uint32(dataCopy[offset+3])

		// Write back in host-endian (little-endian on x86)
		dataCopy[offset] = byte(word)
		dataCopy[offset+1] = byte(word >> 8)
		dataCopy[offset+2] = byte(word >> 16)
		dataCopy[offset+3] = byte(word >> 24)
	}

	crc := uint16(0xFFFF)

	// Process each dword in host-endian format
	for i := 0; i < sizeInDwords; i++ {
		offset := i * 4

		// Get 32-bit word in host-endian (little-endian)
		word := uint32(dataCopy[offset]) | uint32(dataCopy[offset+1])<<8 |
			uint32(dataCopy[offset+2])<<16 | uint32(dataCopy[offset+3])<<24

		// Process each bit of the 32-bit word (matches mstflint's Crc16::add)
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

// Calculate implements the CRCCalculator interface
func (c *CRCCalculator) Calculate(data []byte) uint32 {
	// Default to software CRC16
	return uint32(c.CalculateSoftwareCRC16(data))
}

// CalculateWithParams implements the CRCCalculator interface
func (c *CRCCalculator) CalculateWithParams(data []byte, polynomial, initial, xorOut uint32) uint32 {
	// For hardware CRC, we use the special hardware CRC calculation
	if polynomial == types.CRCPolynomial && initial == 0xFFFF && xorOut == 0xFFFF {
		return uint32(c.CalculateHardwareCRC(data))
	}
	// For other parameters, use software CRC
	return uint32(c.CalculateSoftwareCRC16(data))
}

// GetType implements the CRCCalculator interface
func (c *CRCCalculator) GetType() types.CRCType {
	return c.crcType
}
