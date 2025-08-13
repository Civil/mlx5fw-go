package crc

import (
	"github.com/Civil/mlx5fw-go/pkg/types"
)

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
