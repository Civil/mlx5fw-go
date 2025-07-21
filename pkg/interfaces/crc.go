package interfaces

import (
	"fmt"
	"github.com/Civil/mlx5fw-go/pkg/types"
)

// CRCCalculator defines the interface for CRC calculation
type CRCCalculator interface {
	// Calculate calculates CRC for the given data
	Calculate(data []byte) uint32
	
	// CalculateWithParams calculates CRC with specific parameters
	CalculateWithParams(data []byte, polynomial, initial, xorOut uint32) uint32
	
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

// CRCMismatchError represents a CRC verification failure
type CRCMismatchError struct {
	Expected uint32
	Actual   uint32
}

// Error returns the error message
func (e *CRCMismatchError) Error() string {
	return fmt.Sprintf("CRC mismatch: expected 0x%04x, got 0x%04x", e.Expected, e.Actual)
}

// NewCRCMismatchError creates a new CRC mismatch error
func NewCRCMismatchError(expected, actual uint32) error {
	return &CRCMismatchError{
		Expected: expected,
		Actual:   actual,
	}
}