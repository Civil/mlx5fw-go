package crc

import (
	"github.com/Civil/mlx5fw-go/pkg/errors"
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// BaseCRCHandler provides common CRC handling functionality
// that can be embedded in specific CRC handlers
type BaseCRCHandler struct {
	calculator     interfaces.CRCCalculator
	crcOffset      int
	hasEmbeddedCRC bool
}

// NewBaseCRCHandler creates a new base CRC handler
func NewBaseCRCHandler(calculator interfaces.CRCCalculator, crcOffset int, hasEmbeddedCRC bool) *BaseCRCHandler {
	return &BaseCRCHandler{
		calculator:     calculator,
		crcOffset:      crcOffset,
		hasEmbeddedCRC: hasEmbeddedCRC,
	}
}

// VerifyCRC provides common CRC verification logic
func (h *BaseCRCHandler) VerifyCRC(data []byte, expectedCRC uint32, crcType types.CRCType, calculateFunc func([]byte, types.CRCType) (uint32, error)) error {
	calculated, err := calculateFunc(data, crcType)
	if err != nil {
		return err
	}

	if calculated != expectedCRC {
		return errors.CRCMismatchError(expectedCRC, calculated, "section")
	}

	return nil
}

// GetCRCOffset returns the offset where CRC is stored
func (h *BaseCRCHandler) GetCRCOffset() int {
	return h.crcOffset
}

// HasEmbeddedCRC returns whether CRC is embedded in the section
func (h *BaseCRCHandler) HasEmbeddedCRC() bool {
	return h.hasEmbeddedCRC
}

// GetCalculator returns the CRC calculator
func (h *BaseCRCHandler) GetCalculator() interfaces.CRCCalculator {
	return h.calculator
}

// VerifyCRC16 provides common CRC16 verification logic
// where only the lower 16 bits are significant
func (h *BaseCRCHandler) VerifyCRC16(data []byte, expectedCRC uint32, crcType types.CRCType, calculateFunc func([]byte, types.CRCType) (uint32, error)) error {
	calculated, err := calculateFunc(data, crcType)
	if err != nil {
		return err
	}

	// For CRC16, only the lower 16 bits are significant
	expectedCRC16 := expectedCRC & 0xFFFF
	calculated16 := calculated & 0xFFFF

	if calculated16 != expectedCRC16 {
		return errors.CRCMismatchError(expectedCRC16, calculated16, "section")
	}

	return nil
}

// ValidateCRCType checks if the CRC type is supported
func (h *BaseCRCHandler) ValidateCRCType(crcType types.CRCType, supportedTypes ...types.CRCType) error {
	for _, supported := range supportedTypes {
		if crcType == supported {
			return nil
		}
	}
	return merry.Errorf("unsupported CRC type: %v", crcType)
}
