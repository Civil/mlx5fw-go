package interfaces

import (
	"github.com/Civil/mlx5fw-go/pkg/types"
)

// CRCCalculator defines the interface for CRC calculation
type CRCCalculator interface {
	// Calculate calculates CRC for the given data
	Calculate(data []byte) uint32
	
	// CalculateWithParams calculates CRC with specific parameters
	CalculateWithParams(data []byte, polynomial, initial, xorOut uint32) uint32
	
	// CalculateImageCRC calculates CRC for firmware image data (handles endianness)
	CalculateImageCRC(data []byte, sizeInDwords int) uint16
	
	// GetType returns the CRC type
	GetType() types.CRCType
}

// SectionCRCHandler provides CRC handling for specific section types
type SectionCRCHandler interface {
	// CalculateCRC calculates the CRC for the section data
	CalculateCRC(data []byte, crcType types.CRCType) (uint32, error)
	
	// VerifyCRC verifies the CRC for the section
	VerifyCRC(data []byte, expectedCRC uint32, crcType types.CRCType) error
	
	// GetCRCOffset returns the offset of CRC within the section data (if CRC is embedded)
	GetCRCOffset() int
	
	// HasEmbeddedCRC returns true if CRC is stored within the section data
	HasEmbeddedCRC() bool
}

