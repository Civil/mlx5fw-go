package types

// HWPointerEntry represents a single hardware pointer entry (8 bytes)
// Based on image_layout_hw_pointer_entry from mstflint
type HWPointerEntry struct {
	Ptr uint32 // Pointer value (offset 0x0)
	CRC uint32 // CRC value (offset 0x4)
}

// FS4HWPointers represents the Carmel hardware pointers structure (128 bytes)
// Based on image_layout_hw_pointers_carmel from mstflint
type FS4HWPointers struct {
	BootRecordPtr              HWPointerEntry // offset 0x0 - position 0
	Boot2Ptr                   HWPointerEntry // offset 0x8 - position 1
	TOCPtr                     HWPointerEntry // offset 0x10 - position 2
	ToolsPtr                   HWPointerEntry // offset 0x18 - position 3
	AuthenticationStartPtr     HWPointerEntry // offset 0x20 - position 4
	AuthenticationEndPtr       HWPointerEntry // offset 0x28 - position 5
	DigestPtr                  HWPointerEntry // offset 0x30 - position 6
	DigestRecoveryKeyPtr       HWPointerEntry // offset 0x38 - position 7
	FWWindowStartPtr           HWPointerEntry // offset 0x40 - position 8
	FWWindowEndPtr             HWPointerEntry // offset 0x48 - position 9
	ImageInfoSectionPtr        HWPointerEntry // offset 0x50 - position 10
	ImageSignaturePtr          HWPointerEntry // offset 0x58 - position 11
	PublicKeyPtr               HWPointerEntry // offset 0x60 - position 12
	FWSecurityVersionPtr       HWPointerEntry // offset 0x68 - position 13
	GCMIVDeltaPtr              HWPointerEntry // offset 0x70 - position 14
	HashesTablePtr             HWPointerEntry // offset 0x78 - position 15
}

// ITOCHeader represents the ITOC header structure (32 bytes)
// Based on image_layout_itoc_header from mstflint
type ITOCHeader struct {
	Signature0     uint32 // Should be ITOCSignature (offset 0x0)
	Signature1     uint32 // offset 0x4
	Signature2     uint32 // offset 0x8
	Signature3     uint32 // offset 0xc
	Version        uint32 // offset 0x10
	Reserved       uint32 // offset 0x14
	ITOCEntryCRC   uint32 // offset 0x18
	CRC            uint32 // offset 0x1c
}

// ITOCEntry represents a single ITOC entry (32 bytes)
// Based on image_layout_itoc_entry from mstflint
// This structure is stored as raw bytes due to complex bit packing
type ITOCEntry struct {
	Data [32]byte
	
	// Parsed fields
	Type           uint8
	Size           uint32
	FlashAddr      uint32
	SectionCRC     uint16
	CRC            uint8
	Encrypted      bool
	Param0         uint32
	Param1         uint32
	ITOCEntryCRC   uint16
}

// ParseFields parses the bit-packed fields from the raw data
// Based on image_layout_itoc_entry_unpack from mstflint
func (e *ITOCEntry) ParseFields() {
	// Helper to extract bits from byte array (big-endian bit order)
	getBits := func(data []byte, bitOffset int, bitCount int) uint32 {
		byteOffset := bitOffset / 8
		bitShift := bitOffset % 8
		
		var result uint32
		for i := 0; i < bitCount; i++ {
			byteIdx := byteOffset + (bitShift + i) / 8
			bitIdx := (bitShift + i) % 8
			if byteIdx < len(data) {
				bit := (data[byteIdx] >> (7 - bitIdx)) & 1
				result = (result << 1) | uint32(bit)
			}
		}
		return result
	}
	
	// Type at bits 0-7 (8 bits)
	e.Type = uint8(getBits(e.Data[:], 0, 8))
	
	// Size at bits 8-29 (22 bits) - stored in dwords
	e.Size = getBits(e.Data[:], 8, 22)
	e.Size = e.Size << 2  // Convert dwords to bytes
	
	// Cache line CRC at bit 33
	cacheLine := getBits(e.Data[:], 33, 1)
	
	// Zipped image at bit 32
	zipped := getBits(e.Data[:], 32, 1)
	
	// Param0 at bits 34-63 (30 bits)
	e.Param0 = getBits(e.Data[:], 34, 30)
	
	// Param1 at bits 64-95 (32 bits) - stored as big-endian integer
	e.Param1 = uint32(e.Data[8])<<24 | uint32(e.Data[9])<<16 | uint32(e.Data[10])<<8 | uint32(e.Data[11])
	
	// Version at bits 144-159 (16 bits)
	version := uint16(getBits(e.Data[:], 144, 16))
	
	// Flash address at bits 161-189 (29 bits)
	e.FlashAddr = getBits(e.Data[:], 161, 29)
	// Flash addresses are stored in dwords, convert to bytes
	e.FlashAddr = e.FlashAddr << 2
	
	// Encrypted section at bit 192
	e.Encrypted = getBits(e.Data[:], 192, 1) != 0
	
	// CRC field at bits 205-207 (3 bits)
	e.CRC = uint8(getBits(e.Data[:], 205, 3))
	
	// Section CRC at bits 208-223 (16 bits)
	e.SectionCRC = uint16(getBits(e.Data[:], 208, 16))
	
	// ITOC entry CRC at bits 240-255 (16 bits)
	e.ITOCEntryCRC = uint16(getBits(e.Data[:], 240, 16))
	
	// Store optional fields we might need later
	_ = cacheLine
	_ = zipped
	_ = version
}

// Helper methods for ITOCEntry
func (e *ITOCEntry) GetType() uint16 {
	return uint16(e.Type)
}

func (e *ITOCEntry) GetNoCRC() bool {
	return e.CRC == 1 // NOCRC value
}

func (e *ITOCEntry) GetCRC() uint8 {
	return e.CRC
}