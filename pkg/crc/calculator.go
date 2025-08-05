package crc

import (
	"github.com/Civil/mlx5fw-go/pkg/errors"
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/types"
)

// CRCStrategy defines the strategy for CRC calculation
type CRCStrategy interface {
	Calculate(calc *parser.CRCCalculator, data []byte) uint32
	GetType() types.CRCType
}

// SoftwareCRCStrategy implements software CRC calculation strategy
type SoftwareCRCStrategy struct{}

func (s *SoftwareCRCStrategy) Calculate(calc *parser.CRCCalculator, data []byte) uint32 {
	return uint32(calc.CalculateSoftwareCRC16(data))
}

func (s *SoftwareCRCStrategy) GetType() types.CRCType {
	return types.CRCInITOCEntry
}

// HardwareCRCStrategy implements hardware CRC calculation strategy
type HardwareCRCStrategy struct{}

func (h *HardwareCRCStrategy) Calculate(calc *parser.CRCCalculator, data []byte) uint32 {
	return uint32(calc.CalculateHardwareCRC(data))
}

func (h *HardwareCRCStrategy) GetType() types.CRCType {
	return types.CRCInSection
}

// GenericCRCCalculator implements CRC calculation using a strategy
type GenericCRCCalculator struct {
	calc     *parser.CRCCalculator
	strategy CRCStrategy
}

// NewGenericCRCCalculator creates a new generic CRC calculator with a strategy
func NewGenericCRCCalculator(strategy CRCStrategy) *GenericCRCCalculator {
	return &GenericCRCCalculator{
		calc:     parser.NewCRCCalculator(),
		strategy: strategy,
	}
}

// Calculate calculates CRC for the given data using the strategy
func (c *GenericCRCCalculator) Calculate(data []byte) uint32 {
	return c.strategy.Calculate(c.calc, data)
}

// CalculateWithParams calculates CRC with specific parameters (uses default calculation)
func (c *GenericCRCCalculator) CalculateWithParams(data []byte, polynomial, initial, xorOut uint32) uint32 {
	// For now, strategies use fixed parameters
	return c.Calculate(data)
}

// CalculateImageCRC calculates CRC for firmware image data (handles endianness)
func (c *GenericCRCCalculator) CalculateImageCRC(data []byte, sizeInDwords int) uint16 {
	return c.calc.CalculateImageCRC(data, sizeInDwords)
}

// GetType returns the CRC type from the strategy
func (c *GenericCRCCalculator) GetType() types.CRCType {
	return c.strategy.GetType()
}

// Factory functions for backward compatibility

// NewSoftwareCRCCalculator creates a new software CRC calculator
func NewSoftwareCRCCalculator() interfaces.CRCCalculator {
	return NewGenericCRCCalculator(&SoftwareCRCStrategy{})
}

// NewHardwareCRCCalculator creates a new hardware CRC calculator
func NewHardwareCRCCalculator() interfaces.CRCCalculator {
	return NewGenericCRCCalculator(&HardwareCRCStrategy{})
}

// DefaultCRCHandler provides default CRC handling for sections
type DefaultCRCHandler struct {
	softwareCalc interfaces.CRCCalculator
	hardwareCalc interfaces.CRCCalculator
}

// NewDefaultCRCHandler creates a new default CRC handler
func NewDefaultCRCHandler() *DefaultCRCHandler {
	return &DefaultCRCHandler{
		softwareCalc: NewSoftwareCRCCalculator(),
		hardwareCalc: NewHardwareCRCCalculator(),
	}
}

// CalculateCRC calculates the CRC for the section data
func (h *DefaultCRCHandler) CalculateCRC(data []byte, crcType types.CRCType) (uint32, error) {
	switch crcType {
	case types.CRCInITOCEntry:
		return h.softwareCalc.Calculate(data), nil
	case types.CRCInSection:
		// For CRC in section, calculate over data excluding last 4 bytes
		if len(data) >= 4 {
			return h.softwareCalc.Calculate(data[:len(data)-4]), nil
		}
		return 0, nil
	case types.CRCNone:
		return 0, nil
	default:
		return h.softwareCalc.Calculate(data), nil
	}
}

// VerifyCRC verifies the CRC for the section
func (h *DefaultCRCHandler) VerifyCRC(data []byte, expectedCRC uint32, crcType types.CRCType) error {
	calculatedCRC, err := h.CalculateCRC(data, crcType)
	if err != nil {
		return err
	}
	
	// Only compare lower 16 bits for CRC16
	if (expectedCRC & 0xFFFF) != (calculatedCRC & 0xFFFF) {
		return errors.CRCMismatchError(expectedCRC&0xFFFF, calculatedCRC&0xFFFF, "section")
	}
	
	return nil
}

// GetCRCOffset returns the offset of CRC within the section data (if CRC is embedded)
func (h *DefaultCRCHandler) GetCRCOffset() int {
	return -4 // CRC is last 4 bytes when embedded
}

// HasEmbeddedCRC returns true if CRC is stored within the section data
func (h *DefaultCRCHandler) HasEmbeddedCRC() bool {
	return false // Default is CRC in ITOC entry
}