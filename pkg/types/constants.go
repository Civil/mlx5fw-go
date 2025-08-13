package types

// Firmware header constants based on mstflint definitions
const (

	// Standard firmware sizes
	FirmwareSize8MB  = 0x00800000 // 8MB
	FirmwareSize16MB = 0x01000000 // 16MB
	FirmwareSize32MB = 0x02000000 // 32MB
	FirmwareSize64MB = 0x04000000 // 64MB

	// DTOC location for different firmware sizes
	DTOCOffset8MB  = 0x007ff000
	DTOCOffset16MB = 0x00fff000
	DTOCOffset32MB = 0x01fff000
	DTOCOffset64MB = 0x03fff000

	// Hardware pointer offsets from magic
	HWPointersOffsetFromMagic = 0x18 // 24 bytes
	HWPointersSize            = 128  // 16 entries * 8 bytes

	// BOOT2 section constants
	Boot2MaxSize = 0x8000 // 32KB max size

	// Standard CRC positions
	CRCSize = 4 // CRC32 is 4 bytes

	// Section alignment
	SectionAlignmentDword  = 4      // Sections aligned to dwords
	SectionAlignmentSector = 0x1000 // 4KB sector alignment

	// Special addresses
	InvalidAddress = 0xFFFFFFFF
	NullAddress    = 0x00000000

	// Header sizes
	ITOCHeaderSize = 32 // 8 dwords
	ITOCEntrySize  = 32 // 8 dwords

	// Tools area constants
	ToolsAreaSize = 0x40 // 64 bytes

	// Image info constants
	ImageInfoSize = 1024 // Standard size

	// DEV_INFO and MFG_INFO constants
	DevInfoSize = 512 // 0x200 - image_layout_device_info size
	MfgInfoSize = 320 // 0x140 - image_layout_mfg_info size

	// Padding patterns
	PaddingByteFF    uint8  = 0xFF
	PaddingByte00    uint8  = 0x00
	PaddingPatternFF uint32 = 0xFFFFFFFF

	// Section type ranges for validation
	SectionTypeImageMin  = 0x10 // Image sections start at 0x10
	SectionTypeImageMax  = 0x5F // Image sections end at 0x5F
	SectionTypeDeviceMin = 0x80 // Device sections start at 0x80
	SectionTypeDeviceMax = 0xFF // Device sections end at 0xFF

	// Authentication window constants
	AuthWindowAlignment = 0x1000 // 4KB alignment

	// Secure boot constants
	SignatureSize256  = 256 // RSA-2048 signature
	SignatureSize512  = 512 // RSA-4096 signature
	PublicKeySize2048 = 256 // RSA-2048 public key
	PublicKeySize4096 = 512 // RSA-4096 public key

	// Hash table constants
	HashesTableSize    = 2052                 // 0x804
	HashesTableMagic   = 0x484153485F5441424C // "HASH_TABL"
	HashTableEntrySize = 64                   // 32-byte hash + metadata

	// Firmware format identifiers
	FS3Magic = 0x4D544657 // "MTFW" for FS3

	// Boot version offset from magic pattern
	// Based on mstflint's FS4_BOOT_VERSION_OFFSET in mlxfwops/lib/fw_ops.h:492
	BootVersionOffset = 0x10 // Offset to boot version structure from magic pattern

	// Image format version values
	// Based on mstflint's enum in mlxfwops/lib/fw_ops.h:523-526
	ImageFormatVersionFS2 = 0
	ImageFormatVersionFS3 = 3
	ImageFormatVersionFS4 = 1
	ImageFormatVersionFS5 = 2
)

// GetDTOCOffset returns the DTOC offset for a given firmware size
func GetDTOCOffset(firmwareSize uint32) uint32 {
	switch firmwareSize {
	case FirmwareSize8MB:
		return DTOCOffset8MB
	case FirmwareSize16MB:
		return DTOCOffset16MB
	case FirmwareSize32MB:
		return DTOCOffset32MB
	case FirmwareSize64MB:
		return DTOCOffset64MB
	default:
		// For non-standard sizes, DTOC is at (size - 0x1000)
		return firmwareSize - SectionAlignmentSector
	}
}

// IsValidSectionType checks if a section type is valid
func IsValidSectionType(sectionType uint16) bool {
	return (sectionType >= SectionTypeImageMin && sectionType <= SectionTypeImageMax) ||
		(sectionType >= SectionTypeDeviceMin && sectionType <= SectionTypeDeviceMax)
}

// IsImageSection checks if a section type is an image section
func IsImageSection(sectionType uint16) bool {
	return sectionType >= SectionTypeImageMin && sectionType <= SectionTypeImageMax
}

// IsDeviceSection checks if a section type is a device section
func IsDeviceSection(sectionType uint16) bool {
	return sectionType >= SectionTypeDeviceMin && sectionType <= SectionTypeDeviceMax
}

// AlignToDword aligns a size to dword boundary
func AlignToDword(size uint32) uint32 {
	return (size + 3) & ^uint32(3)
}

// AlignToSector aligns an address to sector boundary
func AlignToSector(addr uint32) uint32 {
	return (addr + SectionAlignmentSector - 1) & ^uint32(SectionAlignmentSector-1)
}
