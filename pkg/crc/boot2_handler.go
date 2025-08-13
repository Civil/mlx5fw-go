package crc

import (
	"encoding/binary"
	"github.com/Civil/mlx5fw-go/pkg/errors"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// Boot2CRCHandler implements CRC handling specifically for BOOT2 sections
// Based on mstflint's CheckBoot2 implementation
type Boot2CRCHandler struct {
	*BaseCRCHandler
	crcCalculator *parser.CRCCalculator
}

// NewBoot2CRCHandler creates a new BOOT2 CRC handler
func NewBoot2CRCHandler() *Boot2CRCHandler {
	calc := parser.NewCRCCalculator()
	return &Boot2CRCHandler{
		BaseCRCHandler: NewBaseCRCHandler(calc, -4, true),
		crcCalculator:  calc,
	}
}

// CalculateCRC calculates CRC16 for BOOT2 section
// BOOT2 format:
// - Header (16 bytes): magic(4) + size_dwords(4) + reserved(8)
// - Data (size_dwords * 4 bytes)
// - CRC is at offset (size_dwords + 3) * 4
// CRC is calculated over (size_dwords + 3) dwords, excluding the CRC dword
func (h *Boot2CRCHandler) CalculateCRC(data []byte, crcType types.CRCType) (uint32, error) {
	if len(data) < 16 {
		return 0, errors.DataTooShortError(16, len(data), "BOOT2 header")
	}

	// Parse size from header (at offset 4)
	sizeDwords := binary.BigEndian.Uint32(data[4:8])

	// Calculate expected total size
	expectedSize := (sizeDwords + 4) * 4
	if uint32(len(data)) < expectedSize {
		return 0, errors.DataTooShortError(int(expectedSize), len(data), "BOOT2 section")
	}

	// CRC is calculated over (size_dwords + 3) dwords
	// This excludes the last dword which contains the CRC
	crcDataSize := (sizeDwords + 3) * 4
	if crcDataSize > uint32(len(data)) {
		return 0, errors.InvalidParameterError("crcDataSize", merry.Errorf("size %d exceeds section size %d", crcDataSize, len(data)).Error())
	}

	// Calculate CRC using CalculateImageCRC which matches mstflint's implementation
	crc16 := h.crcCalculator.CalculateImageCRC(data[:crcDataSize], int(sizeDwords+3))
	return uint32(crc16), nil
}

// VerifyCRC verifies the CRC for BOOT2 section
func (h *Boot2CRCHandler) VerifyCRC(data []byte, expectedCRC uint32, crcType types.CRCType) error {
	if len(data) < 16 {
		return errors.DataTooShortError(16, len(data), "BOOT2 header")
	}

	// Parse size from header
	sizeDwords := binary.BigEndian.Uint32(data[4:8])

	// CRC is at offset (size_dwords + 3) * 4
	crcOffset := (sizeDwords + 3) * 4
	if crcOffset+4 > uint32(len(data)) {
		return errors.InvalidParameterError("crcOffset", merry.Errorf("offset %d out of bounds", crcOffset).Error())
	}

	// Extract CRC from the last dword
	// BOOT2 stores CRC in the lower 16 bits of the last dword
	crcDword := binary.BigEndian.Uint32(data[crcOffset : crcOffset+4])
	storedCRC := uint16(crcDword & 0xFFFF)

	// Calculate CRC
	calculated, err := h.CalculateCRC(data, crcType)
	if err != nil {
		return err
	}

	// Compare only the 16-bit values
	if uint16(calculated) != storedCRC {
		return errors.CRCMismatchError(uint32(storedCRC), calculated, "BOOT2")
	}

	return nil
}

// GetCRCOffset and HasEmbeddedCRC are inherited from BaseCRCHandler
// For BOOT2, the CRC offset is dynamic based on the size field
// but we still indicate it's in the last 4 bytes with -4
