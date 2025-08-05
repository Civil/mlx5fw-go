package crc

import (
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// ToolsAreaCRCHandler implements CRC handling specifically for TOOLS_AREA sections
type ToolsAreaCRCHandler struct {
	*BaseCRCHandler
	crcCalculator *parser.CRCCalculator
}

// NewToolsAreaCRCHandler creates a new TOOLS_AREA CRC handler
func NewToolsAreaCRCHandler() *ToolsAreaCRCHandler {
	calc := parser.NewCRCCalculator()
	return &ToolsAreaCRCHandler{
		BaseCRCHandler: NewBaseCRCHandler(calc, 62, true),
		crcCalculator:  calc,
	}
}

// CalculateCRC calculates CRC16 for TOOLS_AREA (first 60 bytes)
func (h *ToolsAreaCRCHandler) CalculateCRC(data []byte, crcType types.CRCType) (uint32, error) {
	if len(data) < 64 {
		return 0, merry.New("TOOLS_AREA data too small")
	}
	// Calculate CRC on first 60 bytes (15 dwords) using CalculateImageCRC
	// which handles endianness conversion
	crc := h.crcCalculator.CalculateImageCRC(data[:60], 15)
	return uint32(crc), nil
}

// VerifyCRC verifies the CRC for TOOLS_AREA section
func (h *ToolsAreaCRCHandler) VerifyCRC(data []byte, expectedCRC uint32, crcType types.CRCType) error {
	// Use the CRC16 verification that only compares lower 16 bits
	return h.BaseCRCHandler.VerifyCRC16(data, expectedCRC, crcType, h.CalculateCRC)
}

// GetCRCOffset and HasEmbeddedCRC are inherited from BaseCRCHandler