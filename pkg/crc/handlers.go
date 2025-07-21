package crc

import (
	"github.com/ansel1/merry/v2"
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
)

// SoftwareCRC16Handler implements CRC handling for software CRC16
type SoftwareCRC16Handler struct {
	calculator interfaces.CRCCalculator
}

// NewSoftwareCRC16Handler creates a new software CRC16 handler
func NewSoftwareCRC16Handler(calculator interfaces.CRCCalculator) *SoftwareCRC16Handler {
	return &SoftwareCRC16Handler{
		calculator: calculator,
	}
}

// CalculateCRC calculates CRC16 for the given data
func (h *SoftwareCRC16Handler) CalculateCRC(data []byte, crcType types.CRCType) (uint32, error) {
	if crcType != types.CRCInITOCEntry && crcType != types.CRCInSection {
		return 0, merry.Errorf("unsupported CRC type for software CRC16: %v", crcType)
	}
	return h.calculator.Calculate(data), nil
}

// VerifyCRC verifies the CRC for the section
func (h *SoftwareCRC16Handler) VerifyCRC(data []byte, expectedCRC uint32, crcType types.CRCType) error {
	calculated, err := h.CalculateCRC(data, crcType)
	if err != nil {
		return err
	}
	
	if calculated != expectedCRC {
		return interfaces.NewCRCMismatchError(expectedCRC, calculated)
	}
	
	return nil
}

// GetCRCOffset returns -1 as software CRC is typically stored externally
func (h *SoftwareCRC16Handler) GetCRCOffset() int {
	return -1
}

// HasEmbeddedCRC returns false as software CRC is typically stored externally
func (h *SoftwareCRC16Handler) HasEmbeddedCRC() bool {
	return false
}

// HardwareCRC16Handler implements CRC handling for hardware CRC16
type HardwareCRC16Handler struct {
	calculator interfaces.CRCCalculator
}

// NewHardwareCRC16Handler creates a new hardware CRC16 handler
func NewHardwareCRC16Handler(calculator interfaces.CRCCalculator) *HardwareCRC16Handler {
	return &HardwareCRC16Handler{
		calculator: calculator,
	}
}

// CalculateCRC calculates hardware CRC16 for the given data
func (h *HardwareCRC16Handler) CalculateCRC(data []byte, crcType types.CRCType) (uint32, error) {
	if crcType != types.CRCInSection {
		return 0, merry.Errorf("hardware CRC16 only supports CRCInSection type")
	}
	return h.calculator.CalculateWithParams(data, types.CRCPolynomial, 0xFFFF, 0xFFFF), nil
}

// VerifyCRC verifies the CRC for the section
func (h *HardwareCRC16Handler) VerifyCRC(data []byte, expectedCRC uint32, crcType types.CRCType) error {
	calculated, err := h.CalculateCRC(data, crcType)
	if err != nil {
		return err
	}
	
	if calculated != expectedCRC {
		return interfaces.NewCRCMismatchError(expectedCRC, calculated)
	}
	
	return nil
}

// GetCRCOffset returns the offset where CRC is stored (typically at the end)
func (h *HardwareCRC16Handler) GetCRCOffset() int {
	return -4 // Last 4 bytes
}

// HasEmbeddedCRC returns true as hardware CRC is typically embedded in section
func (h *HardwareCRC16Handler) HasEmbeddedCRC() bool {
	return true
}

// NoCRCHandler implements a handler for sections without CRC
type NoCRCHandler struct{}

// NewNoCRCHandler creates a new no-CRC handler
func NewNoCRCHandler() *NoCRCHandler {
	return &NoCRCHandler{}
}

// CalculateCRC returns 0 as there's no CRC
func (h *NoCRCHandler) CalculateCRC(data []byte, crcType types.CRCType) (uint32, error) {
	return 0, nil
}

// VerifyCRC always returns nil as there's no CRC to verify
func (h *NoCRCHandler) VerifyCRC(data []byte, expectedCRC uint32, crcType types.CRCType) error {
	return nil
}

// GetCRCOffset returns -1 as there's no CRC
func (h *NoCRCHandler) GetCRCOffset() int {
	return -1
}

// HasEmbeddedCRC returns false as there's no CRC
func (h *NoCRCHandler) HasEmbeddedCRC() bool {
	return false
}

// InSectionCRC16Handler implements CRC handling for sections with embedded CRC16
type InSectionCRC16Handler struct {
	calculator interfaces.CRCCalculator
}

// NewInSectionCRC16Handler creates a new in-section CRC16 handler
func NewInSectionCRC16Handler(calculator interfaces.CRCCalculator) *InSectionCRC16Handler {
	return &InSectionCRC16Handler{
		calculator: calculator,
	}
}

// CalculateCRC calculates CRC16 for the given data (excluding the CRC itself)
func (h *InSectionCRC16Handler) CalculateCRC(data []byte, crcType types.CRCType) (uint32, error) {
	if len(data) < 4 {
		return 0, merry.New("data too small for in-section CRC")
	}
	// Calculate CRC on all data except the last 4 bytes (where CRC is stored)
	return h.calculator.Calculate(data[:len(data)-4]), nil
}

// VerifyCRC verifies the CRC for the section
func (h *InSectionCRC16Handler) VerifyCRC(data []byte, expectedCRC uint32, crcType types.CRCType) error {
	calculated, err := h.CalculateCRC(data, crcType)
	if err != nil {
		return err
	}
	
	if calculated != expectedCRC {
		return interfaces.NewCRCMismatchError(expectedCRC, calculated)
	}
	
	return nil
}

// GetCRCOffset returns the offset where CRC is stored (last 4 bytes)
func (h *InSectionCRC16Handler) GetCRCOffset() int {
	return -4 // Last 4 bytes
}

// HasEmbeddedCRC returns true as CRC is embedded in section
func (h *InSectionCRC16Handler) HasEmbeddedCRC() bool {
	return true
}