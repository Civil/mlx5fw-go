package types

// FS5HWPointers represents the Gilboa hardware pointers structure (128 bytes)
// Based on fs5_image_layout_hw_pointers_gilboa from mstflint
// This structure has exactly 16 entries (fixed size)
type FS5HWPointers struct {
	Boot2Ptr               HWPointerEntry // offset 0x0
	TOCPtr                 HWPointerEntry // offset 0x8
	ToolsPtr               HWPointerEntry // offset 0x10
	ImageInfoSectionPtr    HWPointerEntry // offset 0x18
	FWPublicKeyPtr         HWPointerEntry // offset 0x20
	FWSignaturePtr         HWPointerEntry // offset 0x28
	PublicKeyPtr           HWPointerEntry // offset 0x30
	ForbiddenVersionsPtr   HWPointerEntry // offset 0x38
	PSCBl1Ptr              HWPointerEntry // offset 0x40
	PSCHashesTablePtr      HWPointerEntry // offset 0x48
	NCoreHashesPointer     HWPointerEntry // offset 0x50
	PSCFWUpdateHandlePtr   HWPointerEntry // offset 0x58
	PSCBCHPointer          HWPointerEntry // offset 0x60
	ReservedPtr13          HWPointerEntry // offset 0x68
	ReservedPtr14          HWPointerEntry // offset 0x70
	NCoreBCHPointer        HWPointerEntry // offset 0x78
}

// HashesTableHeader represents the hashes table header structure
// This is specific to FS5 format
type HashesTableHeader struct {
	Magic          uint32 // Should be a specific magic value
	Version        uint32 // Version number
	Reserved1      uint32 // Reserved field
	Reserved2      uint32 // Reserved field
	TableSize      uint32 // Size of the hashes table
	NumEntries     uint32 // Number of hash entries
	Reserved3      uint32 // Reserved field
	CRC            uint16 // CRC16 of the header
	Reserved4      uint16 // Reserved field
}

// HashTableEntry represents a single hash entry in the hashes table
type HashTableEntry struct {
	Type           uint32      // Hash type/identifier
	Offset         uint32      // Offset in the image
	Size           uint32      // Size of the hashed region
	Reserved       uint32      // Reserved field
	Hash           [32]byte `bin:""`      // SHA-256 hash value
}