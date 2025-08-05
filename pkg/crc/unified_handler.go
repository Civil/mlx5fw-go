package crc

import (
	"github.com/Civil/mlx5fw-go/pkg/errors"
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
)

// HandlerStrategy defines the strategy for CRC calculation in handlers
type HandlerStrategy interface {
	CalculateCRC(calculator interfaces.CRCCalculator, data []byte, crcType types.CRCType) (uint32, error)
	ValidCRCTypes() []types.CRCType
}

// SoftwareCRC16Strategy implements software CRC16 calculation
type SoftwareCRC16Strategy struct{}

func (s *SoftwareCRC16Strategy) CalculateCRC(calculator interfaces.CRCCalculator, data []byte, crcType types.CRCType) (uint32, error) {
	return calculator.Calculate(data), nil
}

func (s *SoftwareCRC16Strategy) ValidCRCTypes() []types.CRCType {
	return []types.CRCType{types.CRCInITOCEntry, types.CRCInSection}
}

// HardwareCRC16Strategy implements hardware CRC16 calculation
type HardwareCRC16Strategy struct{}

func (h *HardwareCRC16Strategy) CalculateCRC(calculator interfaces.CRCCalculator, data []byte, crcType types.CRCType) (uint32, error) {
	return calculator.CalculateWithParams(data, types.CRCPolynomial, 0xFFFF, 0xFFFF), nil
}

func (h *HardwareCRC16Strategy) ValidCRCTypes() []types.CRCType {
	return []types.CRCType{types.CRCInSection}
}

// InSectionCRC16Strategy implements in-section CRC16 calculation
type InSectionCRC16Strategy struct{}

func (i *InSectionCRC16Strategy) CalculateCRC(calculator interfaces.CRCCalculator, data []byte, crcType types.CRCType) (uint32, error) {
	if len(data) < 4 {
		return 0, errors.DataTooShortError(4, len(data), "in-section CRC")
	}
	// Calculate CRC on all data except the last 4 bytes (where CRC is stored)
	// Use CalculateImageCRC which handles endianness conversion (TOCPU)
	sizeInDwords := (len(data) - 4) / 4
	crc16 := calculator.CalculateImageCRC(data[:len(data)-4], sizeInDwords)
	return uint32(crc16), nil
}

func (i *InSectionCRC16Strategy) ValidCRCTypes() []types.CRCType {
	return []types.CRCType{types.CRCInSection}
}

// UnifiedCRCHandler implements a unified CRC handler using strategies
type UnifiedCRCHandler struct {
	*BaseCRCHandler
	strategy HandlerStrategy
}

// NewUnifiedCRCHandler creates a new unified CRC handler
func NewUnifiedCRCHandler(calculator interfaces.CRCCalculator, strategy HandlerStrategy, crcOffset int, hasEmbeddedCRC bool) *UnifiedCRCHandler {
	return &UnifiedCRCHandler{
		BaseCRCHandler: NewBaseCRCHandler(calculator, crcOffset, hasEmbeddedCRC),
		strategy:       strategy,
	}
}

// CalculateCRC calculates CRC using the strategy
func (h *UnifiedCRCHandler) CalculateCRC(data []byte, crcType types.CRCType) (uint32, error) {
	if err := h.ValidateCRCType(crcType, h.strategy.ValidCRCTypes()...); err != nil {
		return 0, err
	}
	return h.strategy.CalculateCRC(h.GetCalculator(), data, crcType)
}

// VerifyCRC verifies the CRC for the section
func (h *UnifiedCRCHandler) VerifyCRC(data []byte, expectedCRC uint32, crcType types.CRCType) error {
	// Use appropriate verification based on whether it's CRC16
	if h.HasEmbeddedCRC() {
		return h.BaseCRCHandler.VerifyCRC16(data, expectedCRC, crcType, h.CalculateCRC)
	}
	return h.BaseCRCHandler.VerifyCRC(data, expectedCRC, crcType, h.CalculateCRC)
}

// Factory functions for creating specific handlers

// NewSoftwareCRC16Handler creates a new software CRC16 handler
func NewSoftwareCRC16Handler(calculator interfaces.CRCCalculator) interfaces.SectionCRCHandler {
	return NewUnifiedCRCHandler(calculator, &SoftwareCRC16Strategy{}, -1, false)
}

// NewHardwareCRC16Handler creates a new hardware CRC16 handler
func NewHardwareCRC16Handler(calculator interfaces.CRCCalculator) interfaces.SectionCRCHandler {
	return NewUnifiedCRCHandler(calculator, &HardwareCRC16Strategy{}, -4, true)
}

// NewInSectionCRC16Handler creates a new in-section CRC16 handler
func NewInSectionCRC16Handler(calculator interfaces.CRCCalculator) interfaces.SectionCRCHandler {
	return NewUnifiedCRCHandler(calculator, &InSectionCRC16Strategy{}, -4, true)
}