package crc

import (
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/types"
)

// SoftwareCRCCalculator implements CRC calculation using software algorithm
type SoftwareCRCCalculator struct {
	calc *parser.CRCCalculator
}

// NewSoftwareCRCCalculator creates a new software CRC calculator
func NewSoftwareCRCCalculator() *SoftwareCRCCalculator {
	return &SoftwareCRCCalculator{
		calc: parser.NewCRCCalculator(),
	}
}

// Calculate calculates CRC for the given data
func (c *SoftwareCRCCalculator) Calculate(data []byte) uint32 {
	return uint32(c.calc.CalculateSoftwareCRC16(data))
}

// CalculateWithParams calculates CRC with specific parameters (not used for software CRC)
func (c *SoftwareCRCCalculator) CalculateWithParams(data []byte, polynomial, initial, xorOut uint32) uint32 {
	// Software CRC uses fixed parameters
	return c.Calculate(data)
}

// GetType returns the CRC type
func (c *SoftwareCRCCalculator) GetType() types.CRCType {
	return types.CRCInITOCEntry
}

// HardwareCRCCalculator implements CRC calculation using hardware algorithm
type HardwareCRCCalculator struct {
	calc *parser.CRCCalculator
}

// NewHardwareCRCCalculator creates a new hardware CRC calculator
func NewHardwareCRCCalculator() *HardwareCRCCalculator {
	return &HardwareCRCCalculator{
		calc: parser.NewCRCCalculator(),
	}
}

// Calculate calculates CRC for the given data
func (c *HardwareCRCCalculator) Calculate(data []byte) uint32 {
	return uint32(c.calc.CalculateHardwareCRC(data))
}

// CalculateWithParams calculates CRC with specific parameters (not used for hardware CRC)
func (c *HardwareCRCCalculator) CalculateWithParams(data []byte, polynomial, initial, xorOut uint32) uint32 {
	// Hardware CRC uses fixed parameters
	return c.Calculate(data)
}

// GetType returns the CRC type
func (c *HardwareCRCCalculator) GetType() types.CRCType {
	return types.CRCInSection
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
		return interfaces.NewCRCMismatchError(expectedCRC&0xFFFF, calculatedCRC&0xFFFF)
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