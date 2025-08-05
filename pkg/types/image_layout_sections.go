package types

// This file contains section structure definitions ported from mstflint's image_layout_layouts.h

// FWVersion represents the firmware version structure
// Based on image_layout_FW_VERSION from mstflint
type FWVersion struct {
	Major    uint16 `bin:"BE"` // offset 0x0.16-0x0.31
	Reserved uint16 `bin:"BE"` // offset 0x0.0-0x0.15 (padding)
	Minor    uint16 `bin:"BE"` // offset 0x4.16-0x4.31
	Subminor uint16 `bin:"BE"` // offset 0x4.0-0x4.15
	Hour     uint8  `bin:"BE"` // offset 0x8.8-0x8.15
	Reserved2 uint8 `bin:"BE"` // offset 0x8.0-0x8.7 (padding)
	Minutes  uint8  `bin:"BE"` // offset 0x8.16-0x8.23
	Seconds  uint8  `bin:"BE"` // offset 0x8.24-0x8.31
	Day      uint8  `bin:"BE"` // offset 0xc.0-0xc.7
	Month    uint8  `bin:"BE"` // offset 0xc.8-0xc.15
	Year     uint16 `bin:"BE"` // offset 0xc.16-0xc.31
}

// TripleVersion represents a triple version structure
// Based on image_layout_TRIPPLE_VERSION from mstflint
type TripleVersion struct {
	Major    uint16 `bin:"BE"` // offset 0x0.16-0x0.31
	Reserved uint16 `bin:"BE"` // offset 0x0.0-0x0.15 (padding)
	Minor    uint16 `bin:"BE"` // offset 0x4.16-0x4.31
	Subminor uint16 `bin:"BE"` // offset 0x4.0-0x4.15
}

// ModuleVersion represents a module version structure
// Based on image_layout_module_version from mstflint
type ModuleVersion struct {
	Major  uint16 `bin:"BE,bitfield=12"` // offset 0x0.20-0x0.31 (12 bits)
	Minor  uint16 `bin:"BE,bitfield=12"` // offset 0x0.8-0x0.19 (12 bits)
	Branch uint8  `bin:"BE"`              // offset 0x0.0-0x0.7
}

// ImageSize represents the image size structure
// Based on image_layout_image_size from mstflint
type ImageSize struct {
	LogStep     uint8  // log of next address in bytes to search for an image
	Reserved    [3]uint8 // padding to make it 32-bit aligned
	RunFromAny  uint8  // bit 31: this image can run from any partition
	MaxSize     uint32 // Max possible size in bytes of image
}

// ModuleVersions represents module versions structure
// Based on image_layout_module_versions from mstflint
type ModuleVersions struct {
	Core           ModuleVersion // offset 0x0
	Phy            ModuleVersion // offset 0x4
	Kernel         ModuleVersion // offset 0x8
	IronImage      ModuleVersion // offset 0xc
	HostManagement ModuleVersion // offset 0x10
	Mad            ModuleVersion // offset 0x14
	Reserved       [40]uint8     // padding to 64 bytes
}

// UIDEntry represents a UID entry structure
// Based on image_layout_uid_entry from mstflint
type UIDEntry struct {
	NumAllocated     uint8    // offset 0x0.0-0x0.7
	Reserved1        uint8    // offset 0x0.8-0x0.15
	Step             uint16   // offset 0x0.16-0x0.31
	NumOfGuids       uint8    // offset 0x4.16-0x4.23
	AllocatedUIDMask uint8    // offset 0x4.24-0x4.31
	Reserved2        uint16   // offset 0x4.0-0x4.15
	GUID0H           uint32   // offset 0x8
	GUID0L           uint32   // offset 0xc
	Reserved3        [16]uint8 // offset 0x10-0x1f
}

// Guids represents the GUIDs structure
// Based on image_layout_guids from mstflint
type Guids struct {
	Guids    UIDEntry  // offset 0x0
	Macs     UIDEntry  // offset 0x20
	Reserved [192]uint8 // padding to 256 bytes
}

// OperationKey represents an operation key structure
// Based on image_layout_operation_key from mstflint
type OperationKey struct {
	KeyModifier uint64 // offset 0x0
	Key         uint64 // offset 0x8
}

// Extended ImageInfo structure with all fields from mstflint
// Based on image_layout_image_info from mstflint
type ImageInfoExtended struct {
	// Flags (bits in first dword)
	LongKeys                   uint8  `bin:"bitfield=1"` // bit 6
	DebugFWTokensSupported     uint8  `bin:"bitfield=1"` // bit 7
	MccEn                      uint8  `bin:"bitfield=1"` // bit 8
	SignedVendorNvconfigFiles  uint8  `bin:"bitfield=1"` // bit 9
	SignedMlnxNvconfigFiles    uint8  `bin:"bitfield=1"` // bit 10
	FrcSupported               uint8  `bin:"bitfield=1"` // bit 11
	CSTokensSupported          uint8  `bin:"bitfield=1"` // bit 12
	DebugFW                    uint8  `bin:"bitfield=1"` // bit 13
	SignedFW                   uint8  `bin:"bitfield=1"` // bit 14
	SecureFW                   uint8  `bin:"bitfield=1"` // bit 15
	
	// Version info
	MinorVersion uint8  // bits 16-23
	MajorVersion uint8  // bits 24-31
	
	// Firmware version
	FWVersion    FWVersion     // offset 0x4-0x14
	MicVersion   TripleVersion // offset 0x14-0x1c
	
	// PCI IDs
	PCIVendorID      uint16 // offset 0x1c.0-0x1c.15
	PCIDeviceID      uint16 // offset 0x1c.16-0x1c.31
	PCISubVendorID   uint16 // offset 0x20.0-0x20.15
	PCISubsystemID   uint16 // offset 0x20.16-0x20.31
	
	// PSID and VSD
	PSID         [17]byte  // offset 0x24-0x34 (null-terminated string)
	VSDVendorID  uint16    // offset 0x34
	Reserved1    uint16    // offset 0x36
	VSD          [209]byte // offset 0x38-0x108 (null-terminated string)
	
	// Image size
	ImageSizeStruct ImageSize // offset 0x108-0x110
	
	// Supported HW IDs
	SupportedHWID [4]uint32 // offset 0x118-0x128
	
	// INI file number
	INIFileNum uint32 // offset 0x128
	
	// Reserved space
	Reserved2 [148]uint8 // offset 0x12c-0x1c0
	
	// Product version
	ProdVer [17]byte // offset 0x1c0-0x1d0 (null-terminated string)
	
	// Description
	Description [257]byte // offset 0x1d0-0x2d0 (null-terminated string)
	
	// Reserved space
	Reserved3 [48]uint8 // offset 0x2d1-0x300
	
	// Module versions
	ModuleVersionsStruct ModuleVersions // offset 0x300-0x340
	
	// Name
	Name [65]byte // offset 0x340-0x380 (null-terminated string)
	
	// PRS name
	PRSName [129]byte // offset 0x380-0x400 (null-terminated string)
}

// HashesTableHeaderExtended represents the hashes table header structure
// Based on image_layout_hashes_table_header from mstflint
type HashesTableHeaderExtended struct {
	Magic    uint64 // offset 0x0 - should be HashesTableMagic
	Version  uint32 // offset 0x8
	Reserved uint32 // offset 0xc
}

// HTOCEntry represents an HTOC entry structure
// Based on image_layout_htoc_entry from mstflint
type HTOCEntry struct {
	BinHashType uint8  // offset 0x0.0-0x0.15 (bits 0-15)
	HashOffset  uint16 // offset 0x0.16-0x0.31 (bits 16-31)
	Reserved    uint32 // offset 0x4
}

// HTOCHeader represents the HTOC header structure
// Based on image_layout_htoc_header from mstflint
type HTOCHeader struct {
	FirstType    uint16 // offset 0x0.0-0x0.15
	Version      uint8  // offset 0x0.16-0x0.23
	HashLocation uint8  // offset 0x0.24-0x0.31
	Reserved     uint32 // offset 0x4
}

// HTOC represents the HTOC structure with header and entries
// Based on image_layout_htoc from mstflint
type HTOC struct {
	Header  HTOCHeader    // offset 0x0
	Entries [28]HTOCEntry // offset 0x8
}

// HTOCHash represents a hash entry in the hashes table
// Based on image_layout_htoc_hash from mstflint
type HTOCHash struct {
	HashData [32]uint8 // offset 0x0 - SHA256 hash
	Reserved [32]uint8 // offset 0x20 - padding to 64 bytes
}

// HashesTableExtended represents the complete hashes table structure
// Based on image_layout_hashes_table from mstflint
type HashesTableExtended struct {
	Header  HashesTableHeaderExtended // offset 0x0
	HTOC    HTOC                      // offset 0x10
	Hashes  [28]HTOCHash              // offset 0xf0
}

// BootVersion represents the boot version structure
// Based on image_layout_boot_version from mstflint
type BootVersion struct {
	MajorVersion uint16 // offset 0x0.16-0x0.31
	MinorVersion uint16 // offset 0x0.0-0x0.15
	SubVersion   uint16 // offset 0x4.16-0x4.31
	Reserved     uint16 // offset 0x4.0-0x4.15
}

// DeviceInfoExtended represents the extended device info structure
// Based on image_layout_device_info from mstflint
type DeviceInfoExtended struct {
	Signature        uint32         // offset 0x0
	MinorVersion     uint16         // offset 0x4.0-0x4.15
	MajorVersion     uint16         // offset 0x4.16-0x4.31
	Reserved1        [24]uint8      // offset 0x8-0x20
	GUIDs            Guids          // offset 0x20-0x120
	Reserved2        [48]uint8      // offset 0x120-0x150
	VSD              [208]uint8     // offset 0x150-0x220
	OperationKeys    [4]OperationKey // offset 0x220-0x260
	Reserved3        [416]uint8     // offset 0x260-0x400
}

