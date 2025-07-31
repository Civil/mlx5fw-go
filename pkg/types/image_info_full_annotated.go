package types

import (
	"reflect"
	
	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// ImageInfoAnnotated represents the full IMAGE_INFO section structure using annotations
// Based on mstflint's connectx4_image_info structure
type ImageInfoAnnotated struct {
	// DWORD 0 - offset 0x0 (32 bits total) - stored as big-endian
	// Individual bit fields within the 32-bit SecurityAndVersion field
	MinorVersion   uint8  `offset:"byte:0,bit:0,len:8,endian:be"`   // Bits 0-7
	SecurityMode   uint8  `offset:"byte:0,bit:8,len:8,endian:be"`   // Bits 8-15
	Reserved0      uint8  `offset:"byte:0,bit:16,len:8,endian:be"`  // Bits 16-23
	SignedFW       bool   `offset:"byte:0,bit:24,len:1,endian:be"`  // Bit 24
	SecureFW       bool   `offset:"byte:0,bit:25,len:1,endian:be"`  // Bit 25
	MajorVersion   uint8  `offset:"byte:0,bit:26,len:4,endian:be"`  // Bits 26-29 (4 bits only)
	MCCEnabled     bool   `offset:"byte:0,bit:30,len:1,endian:be"`  // Bit 30
	DebugFW        bool   `offset:"byte:0,bit:31,len:1,endian:be"`  // Bit 31
	
	// DWORD 1-4 - offset 0x4 - FW Version
	FWVerMajor     uint16 `offset:"byte:4,endian:be"`  // offset 0x4
	Reserved2      uint16 `offset:"byte:6,endian:be,reserved:true"`  // offset 0x6
	FWVerMinor     uint16 `offset:"byte:8,endian:be"`  // offset 0x8
	FWVerSubminor  uint16 `offset:"byte:10,endian:be"` // offset 0xa
	
	// Date/time fields - TIME at offset 0xc, DATE at offset 0x10
	Reserved3a     uint8  `offset:"byte:12,reserved:true"` // offset 0xc byte 0
	Hour           uint8  `offset:"byte:13,hex_as_dec:true"`  // offset 0xc byte 1
	Minutes        uint8  `offset:"byte:14,hex_as_dec:true"`  // offset 0xc byte 2
	Seconds        uint8  `offset:"byte:15,hex_as_dec:true"`  // offset 0xc byte 3
	Year           uint16 `offset:"byte:16,endian:be,hex_as_dec:true"` // offset 0x10 - BCD year (e.g. 0x2024 = 2024)
	Month          uint8  `offset:"byte:18,hex_as_dec:true"`  // offset 0x12 - BCD month (e.g. 0x06 = 6)
	Day            uint8  `offset:"byte:19,hex_as_dec:true"`  // offset 0x13 - BCD day (e.g. 0x27 = 27)
	
	// MIC version - offset 0x14
	MICVerMajor    uint16 `offset:"byte:20,endian:be"` // offset 0x14
	Reserved4      uint16 `offset:"byte:22,endian:be,reserved:true"` // offset 0x16
	MICVerSubminor uint16 `offset:"byte:24,endian:be"` // offset 0x18
	MICVerMinor    uint16 `offset:"byte:26,endian:be"` // offset 0x1a
	
	// PCI IDs - offset 0x1c
	PCIDeviceID    uint16 `offset:"byte:28,endian:be"` // offset 0x1c
	PCIVendorID    uint16 `offset:"byte:30,endian:be"` // offset 0x1e
	PCISubsystemID uint16 `offset:"byte:32,endian:be"` // offset 0x20
	PCISubVendorID uint16 `offset:"byte:34,endian:be"` // offset 0x22
	
	// PSID - offset 0x24
	PSID           [16]byte `offset:"byte:36"` // offset 0x24-0x33
	
	// Padding/alignment - offset 0x34
	Reserved5a     uint16   `offset:"byte:52,endian:be,reserved:true"`   // offset 0x34-0x35
	
	// VSD vendor ID - offset 0x36
	VSDVendorID    uint16   `offset:"byte:54,endian:be"`   // offset 0x36-0x37
	
	// VSD - offset 0x38
	VSD            [208]byte `offset:"byte:56"` // offset 0x38-0x107
	
	// Image size - offset 0x108
	ImageSizeData  [8]byte   `offset:"byte:264"` // offset 0x108-0x10f
	
	// Reserved - offset 0x110
	Reserved6      [8]byte   `offset:"byte:272,reserved:true"` // offset 0x110-0x117
	
	// Supported HW IDs - offset 0x118
	SupportedHWID  [4]uint32 `offset:"byte:280,endian:be"` // offset 0x118-0x127
	
	// INI File number and reserved - offset 0x128
	INIFileNum     uint32    `offset:"byte:296,endian:be"` // offset 0x128-0x12b
	
	// Big reserved section - offset 0x12c
	Reserved7      [148]byte `offset:"byte:300,reserved:true"` // offset 0x12c-0x1bf
	
	// Product version - offset 0x1c0
	ProductVer     [16]byte  `offset:"byte:448"` // offset 0x1c0-0x1cf
	
	// Description - offset 0x1d0
	Description    [256]byte `offset:"byte:464"` // offset 0x1d0-0x2cf
	
	// Reserved - offset 0x2d0
	Reserved8      [48]byte  `offset:"byte:720,reserved:true"` // offset 0x2d0-0x2ff
	
	// Module versions - offset 0x300
	ModuleVersions [64]byte  `offset:"byte:768"` // offset 0x300-0x33f
	
	// Name (part number) - offset 0x340
	Name           [64]byte  `offset:"byte:832"` // offset 0x340-0x37f
	
	// PRS name - offset 0x380
	PRSName        [128]byte `offset:"byte:896"` // offset 0x380-0x3ff
}

// Unmarshal unmarshals binary data into ImageInfoAnnotated
func (i *ImageInfoAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, i)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (i *ImageInfoAnnotated) UnmarshalWithReserved(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*i))
	if err != nil {
		return err
	}
	opts := &annotations.UnmarshalOptions{
		IncludeReserved: true,
	}
	return annotations.UnmarshalWithOptions(data, i, annot, opts)
}

// Marshal marshals ImageInfoAnnotated into binary data
func (i *ImageInfoAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(i)
}

// MarshalWithReserved marshals ImageInfoAnnotated including reserved fields
func (i *ImageInfoAnnotated) MarshalWithReserved() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*i))
	if err != nil {
		return nil, err
	}
	opts := &annotations.MarshalOptions{
		IncludeReserved: true,
	}
	return annotations.MarshalWithOptions(i, annot, opts)
}

