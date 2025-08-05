package types

import (
	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// ITOCHeaderAnnotated represents the ITOC header structure (32 bytes) using annotations
// Based on image_layout_itoc_header from mstflint
type ITOCHeaderAnnotated struct {
	Signature0     uint32 `offset:"byte:0,endian:be"`  // Should be ITOCSignature (offset 0x0)
	Signature1     uint32 `offset:"byte:4,endian:be"`  // offset 0x4
	Signature2     uint32 `offset:"byte:8,endian:be"`  // offset 0x8
	Signature3     uint32 `offset:"byte:12,endian:be"` // offset 0xc
	Version        uint32 `offset:"byte:16,endian:be"` // offset 0x10
	Reserved       uint32 `offset:"byte:20,endian:be,reserved:true"` // offset 0x14
	ITOCEntryCRC   uint32 `offset:"byte:24,endian:be"` // offset 0x18
	CRC            uint32 `offset:"byte:28,endian:be"` // offset 0x1c
}

// ITOCEntryAnnotated represents an ITOC entry structure (32 bytes) using annotations
// Based on image_layout_itoc_entry from mstflint
// This structure has complex bit packing
type ITOCEntryAnnotated struct {
	// Byte 0-3: Type (bits 0-7), Size (bits 8-29), Param0 bits (30-33)
	Type            uint8  `offset:"bit:0,len:8,endian:be"`          // bits 0-7
	SizeDwords      uint32 `offset:"bit:8,len:22,endian:be"`         // bits 8-29 (size in dwords)
	Param0Low       uint32 `offset:"bit:30,len:4,endian:be"`         // bits 30-33 (lower 4 bits of Param0)
	
	// Byte 4-7: Param0 bits (34-63)
	Param0High      uint32 `offset:"bit:34,len:30,endian:be"`        // bits 34-63 (upper 30 bits of Param0)
	
	// Byte 8-11: Param1 (bits 64-95)
	Param1          uint32 `offset:"byte:8,endian:be"`               // bits 64-95
	
	// Byte 12-15: Reserved
	Reserved1       uint32 `offset:"byte:12,endian:be,reserved:true"` // bits 96-127
	
	// Byte 16-19: Reserved  
	Reserved2       uint32 `offset:"byte:16,endian:be,reserved:true"` // bits 128-159
	
	// Byte 20-23: Flash address (in bytes, not dwords as the field name suggests)
	// Despite the field name, mstflint treats this as a byte address directly
	FlashAddrDwords uint32 `offset:"byte:20,endian:be"`                       // Flash address in bytes
	
	// Byte 24-25: Encrypted flag and CRC flags
	// Based on actual bit layout from legacy parser:
	// - bit 192: encrypted
	// - bits 193-204: reserved
	// - bits 205-207: CRC type field (3 bits)
	//   - 0: CRC in ITOC entry (section_crc field)
	//   - 1: No CRC validation
	//   - 2: CRC at end of section data
	Encrypted       bool   `offset:"bit:192,endian:be"`                       // bit 192
	Reserved5       uint8  `offset:"bit:193,len:12,endian:be,reserved:true"` // bits 193-204
	CRCField        uint8  `offset:"bit:205,len:3,endian:be"`                // bits 205-207 (3 bits)
	
	// Byte 26-27: Section CRC
	SectionCRC      uint16 `offset:"bit:208,endian:be"`                      // bits 208-223 (16 bits)
	
	// Byte 28-29: Reserved
	Reserved7       uint16 `offset:"byte:28,endian:be,reserved:true"`         // bits 224-239
	
	// Byte 30-31: ITOC Entry CRC
	ITOCEntryCRC    uint16 `offset:"byte:30,endian:be"`                       // bits 240-255
}

// Unmarshal methods using annotations

// Unmarshal unmarshals binary data into ITOCHeaderAnnotated
func (h *ITOCHeaderAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// Marshal marshals ITOCHeaderAnnotated into binary data
func (h *ITOCHeaderAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// Unmarshal unmarshals binary data into ITOCEntryAnnotated
func (e *ITOCEntryAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, e)
}

// Marshal marshals ITOCEntryAnnotated into binary data
func (e *ITOCEntryAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(e)
}


// Helper methods

// GetSize returns the size in bytes
func (e *ITOCEntryAnnotated) GetSize() uint32 {
	// The SizeDwords field stores the size in units of 4 bytes (dwords)
	// So we need to multiply by 4 to get bytes
	return e.SizeDwords * 4
}

// GetFlashAddr returns the flash address in bytes
func (e *ITOCEntryAnnotated) GetFlashAddr() uint32 {
	// Despite the field name, FlashAddrDwords actually contains the byte address directly
	// mstflint interprets this field as a byte address, not a dword address
	return e.FlashAddrDwords
}

// GetParam0 returns the combined Param0 value
func (e *ITOCEntryAnnotated) GetParam0() uint32 {
	return (e.Param0High << 4) | e.Param0Low
}

// GetNoCRC returns true if the CRC type is NONE
func (e *ITOCEntryAnnotated) GetNoCRC() bool {
	// Check if CRC type is CRCNone (value 1)
	return e.GetCRCType() == CRCNone
}

// GetDeviceData returns true if the device_data flag is set
func (e *ITOCEntryAnnotated) GetDeviceData() bool {
	// device_data is bit 1 of the 3-bit CRC field (bit 206)
	return (e.CRCField & 0x2) != 0
}

// GetType returns the type as uint16 for compatibility
func (e *ITOCEntryAnnotated) GetType() uint16 {
	return uint16(e.Type)
}

// GetCRCType returns the CRC type based on the CRCField value
func (e *ITOCEntryAnnotated) GetCRCType() CRCType {
	// The CRCField contains the CRC type value directly
	// The annotations have already extracted the 3-bit field for us
	return CRCType(e.CRCField)
}
