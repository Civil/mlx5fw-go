package types

import (
	"encoding/json"
	"reflect"
	
	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// ImageInfoAnnotated represents the full IMAGE_INFO section structure using annotations
// Based on mstflint's connectx4_image_info structure
type ImageInfoAnnotated struct {
	// DWORD 0 - offset 0x0 (32 bits total) - stored as big-endian in firmware
	// Original parseSecurityAttributes uses LSB=0 bit numbering on BE uint32:
	// - mccEn: bit 8, debugFW: bit 13, signedFW: bit 14, secureFW: bit 15
	// But annotation parser uses MSB=0 bit numbering, so we convert:
	// - mccEn: bit 8 (LSB=0) = bit 23 (MSB=0)
	// - debugFW: bit 13 (LSB=0) = bit 18 (MSB=0)
	// - signedFW: bit 14 (LSB=0) = bit 17 (MSB=0)
	// - secureFW: bit 15 (LSB=0) = bit 16 (MSB=0)
	Reserved0      uint8  `offset:"bit:0,len:8,endian:be" json:"reserved0"`   // Bits 0-7 (MSB=0) = bits 24-31 (LSB=0)
	MajorVersion   uint8  `offset:"bit:8,len:8,endian:be" json:"major_version"`   // Bits 8-15 (MSB=0) = bits 16-23 (LSB=0)
	SecurityMode   uint8  `offset:"bit:16,len:8,endian:be" json:"security_mode"`  // Bits 16-23 (MSB=0) = bits 8-15 (LSB=0)
	SecureFW       bool   `offset:"bit:16,len:1,endian:be" json:"secure_fw"`  // Bit 16 (MSB=0) = bit 15 (LSB=0)
	SignedFW       bool   `offset:"bit:17,len:1,endian:be" json:"signed_fw"`  // Bit 17 (MSB=0) = bit 14 (LSB=0)
	DebugFW        bool   `offset:"bit:18,len:1,endian:be" json:"debug_fw"`  // Bit 18 (MSB=0) = bit 13 (LSB=0)
	MCCEnabled     bool   `offset:"bit:23,len:1,endian:be" json:"mcc_enabled"`  // Bit 23 (MSB=0) = bit 8 (LSB=0)
	MinorVersion   uint8  `offset:"bit:24,len:8,endian:be" json:"minor_version"`  // Bits 24-31 (MSB=0) = bits 0-7 (LSB=0)
	
	// DWORD 1-4 - offset 0x4 - FW Version
	FWVerMajor     uint16 `offset:"byte:4,endian:be" json:"fw_ver_major"`  // offset 0x4
	Reserved2      uint16 `offset:"byte:6,endian:be,reserved:true" json:"reserved2"`  // offset 0x6
	FWVerMinor     uint16 `offset:"byte:8,endian:be" json:"fw_ver_minor"`  // offset 0x8
	FWVerSubminor  uint16 `offset:"byte:10,endian:be" json:"fw_ver_subminor"` // offset 0xa
	
	// Date/time fields - TIME at offset 0xc, DATE at offset 0x10
	Reserved3a     uint8  `offset:"byte:12,reserved:true" json:"reserved3a"` // offset 0xc byte 0
	Hour           uint8  `offset:"byte:13,hex_as_dec:true" json:"hour"`  // offset 0xc byte 1
	Minutes        uint8  `offset:"byte:14,hex_as_dec:true" json:"minutes"`  // offset 0xc byte 2
	Seconds        uint8  `offset:"byte:15,hex_as_dec:true" json:"seconds"`  // offset 0xc byte 3
	Year           uint16 `offset:"byte:16,endian:be,hex_as_dec:true" json:"year"` // offset 0x10 - BCD year (e.g. 0x2024 = 2024)
	Month          uint8  `offset:"byte:18,hex_as_dec:true" json:"month"`  // offset 0x12 - BCD month (e.g. 0x06 = 6)
	Day            uint8  `offset:"byte:19,hex_as_dec:true" json:"day"`  // offset 0x13 - BCD day (e.g. 0x27 = 27)
	
	// MIC version - offset 0x14
	MICVerMajor    uint16 `offset:"byte:20,endian:be" json:"mic_ver_major"` // offset 0x14
	Reserved4      uint16 `offset:"byte:22,endian:be,reserved:true" json:"reserved4"` // offset 0x16
	MICVerSubminor uint16 `offset:"byte:24,endian:be" json:"mic_ver_subminor"` // offset 0x18
	MICVerMinor    uint16 `offset:"byte:26,endian:be" json:"mic_ver_minor"` // offset 0x1a
	
	// PCI IDs - offset 0x1c
	PCIDeviceID    uint16 `offset:"byte:28,endian:be" json:"pci_device_id"` // offset 0x1c
	PCIVendorID    uint16 `offset:"byte:30,endian:be" json:"pci_vendor_id"` // offset 0x1e
	PCISubsystemID uint16 `offset:"byte:32,endian:be" json:"pci_subsystem_id"` // offset 0x20
	PCISubVendorID uint16 `offset:"byte:34,endian:be" json:"pci_subvendor_id"` // offset 0x22
	
	// PSID - offset 0x24
	PSID           [16]byte `offset:"byte:36" json:"-"` // offset 0x24-0x33 (will be handled separately in MarshalJSON)
	
	// Padding/alignment - offset 0x34
	Reserved5a     uint16   `offset:"byte:52,endian:be,reserved:true" json:"reserved5a"`   // offset 0x34-0x35
	
	// VSD vendor ID - offset 0x36
	VSDVendorID    uint16   `offset:"byte:54,endian:be" json:"vsd_vendor_id"`   // offset 0x36-0x37
	
	// VSD - offset 0x38
	VSD            [208]byte `offset:"byte:56" json:"-"` // offset 0x38-0x107 (will be handled separately in MarshalJSON)
	
	// Image size - offset 0x108
	ImageSizeData  [8]byte   `offset:"byte:264" json:"image_size_data"` // offset 0x108-0x10f
	
	// Reserved - offset 0x110
	Reserved6      [8]byte   `offset:"byte:272,reserved:true" json:"reserved6"` // offset 0x110-0x117
	
	// Supported HW IDs - offset 0x118
	SupportedHWID  [4]uint32 `offset:"byte:280,endian:be" json:"supported_hw_ids"` // offset 0x118-0x127
	
	// INI File number and reserved - offset 0x128
	INIFileNum     uint32    `offset:"byte:296,endian:be" json:"ini_file_num"` // offset 0x128-0x12b
	
	// Big reserved section - offset 0x12c
	Reserved7      [148]byte `offset:"byte:300,reserved:true" json:"reserved7"` // offset 0x12c-0x1bf
	
	// Product version - offset 0x1c0
	ProductVer     [16]byte  `offset:"byte:448" json:"-"` // offset 0x1c0-0x1cf (will be handled separately in MarshalJSON)
	
	// Description - offset 0x1d0
	Description    [256]byte `offset:"byte:464" json:"-"` // offset 0x1d0-0x2cf (will be handled separately in MarshalJSON)
	
	// Reserved - offset 0x2d0
	Reserved8      [48]byte  `offset:"byte:720,reserved:true" json:"reserved8"` // offset 0x2d0-0x2ff
	
	// Module versions - offset 0x300
	ModuleVersions [64]byte  `offset:"byte:768" json:"module_versions"` // offset 0x300-0x33f
	
	// Name (part number) - offset 0x340
	Name           [64]byte  `offset:"byte:832" json:"-"` // offset 0x340-0x37f (will be handled separately in MarshalJSON)
	
	// PRS name - offset 0x380
	PRSName        [128]byte `offset:"byte:896" json:"-"` // offset 0x380-0x3ff (will be handled separately in MarshalJSON)
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

// MarshalJSON implements json.Marshaler interface
func (i *ImageInfoAnnotated) MarshalJSON() ([]byte, error) {
	// Create an anonymous struct that includes all fields plus computed ones
	type Alias ImageInfoAnnotated
	return json.Marshal(&struct {
		*Alias
		SecurityAndVersion  uint32 `json:"security_and_version"`
		PSID               string `json:"psid"`
		VSD                string `json:"vsd"`
		ProductVer         string `json:"product_ver"`
		ProductVerRaw      []byte `json:"product_ver_raw"`
		Description        string `json:"description"`
		Name               string `json:"name"`
		PRSName            string `json:"prs_name"`
		FWVersion          string `json:"fw_version"`
		FWReleaseDate      string `json:"fw_release_date"`
		MICVersion         string `json:"mic_version"`
		SecurityAttributes string `json:"security_attributes"`
	}{
		Alias:              (*Alias)(i),
		SecurityAndVersion: i.GetSecurityAndVersion(),
		PSID:              i.GetPSIDString(),
		VSD:               i.GetVSDString(),
		ProductVer:        i.GetProductVerString(),
		ProductVerRaw:     i.ProductVer[:],
		Description:       i.GetDescriptionString(),
		Name:              i.GetNameString(),
		PRSName:           i.GetPRSNameString(),
		FWVersion:         i.GetFWVersionString(),
		FWReleaseDate:     i.GetFWReleaseDateString(),
		MICVersion:        i.GetMICVersionString(),
		SecurityAttributes: i.GetSecurityAttributesString(),
	})
}

// UnmarshalJSON implements json.Unmarshaler interface
func (i *ImageInfoAnnotated) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary struct with all fields
	var temp struct {
		SecurityAndVersion  uint32   `json:"security_and_version"`
		MinorVersion       uint8    `json:"minor_version"`
		SecurityMode       uint8    `json:"security_mode"`
		MajorVersion       uint8    `json:"major_version"`
		Reserved0          uint8    `json:"reserved0"`
		MCCEnabled         bool     `json:"mcc_enabled"`
		DebugFW            bool     `json:"debug_fw"`
		SignedFW           bool     `json:"signed_fw"`
		SecureFW           bool     `json:"secure_fw"`
		FWVerMajor         uint16   `json:"fw_ver_major"`
		Reserved2          uint16   `json:"reserved2"`
		FWVerMinor         uint16   `json:"fw_ver_minor"`
		FWVerSubminor      uint16   `json:"fw_ver_subminor"`
		Reserved3a         uint8    `json:"reserved3a"`
		Hour               uint8    `json:"hour"`
		Minutes            uint8    `json:"minutes"`
		Seconds            uint8    `json:"seconds"`
		Day                uint8    `json:"day"`
		Month              uint8    `json:"month"`
		Year               uint16   `json:"year"`
		MICVerMajor        uint16   `json:"mic_ver_major"`
		Reserved4          uint16   `json:"reserved4"`
		MICVerSubminor     uint16   `json:"mic_ver_subminor"`
		MICVerMinor        uint16   `json:"mic_ver_minor"`
		PCIDeviceID        uint16   `json:"pci_device_id"`
		PCIVendorID        uint16   `json:"pci_vendor_id"`
		PCISubsystemID     uint16   `json:"pci_subsystem_id"`
		PCISubVendorID     uint16   `json:"pci_subvendor_id"`
		PSID               string   `json:"psid"`
		Reserved5a         uint16   `json:"reserved5a"`
		VSDVendorID        uint16   `json:"vsd_vendor_id"`
		VSD                string   `json:"vsd"`
		ImageSizeData      []uint8  `json:"image_size_data"`
		Reserved6          []uint8  `json:"reserved6"`
		SupportedHWIDs     []uint32 `json:"supported_hw_ids"`
		INIFileNum         uint32   `json:"ini_file_num"`
		Reserved7          []uint8  `json:"reserved7"`
		ProductVer         string   `json:"product_ver"`
		ProductVerRaw      []uint8  `json:"product_ver_raw"`
		Description        string   `json:"description"`
		Reserved8          []uint8  `json:"reserved8"`
		ModuleVersions     []uint8  `json:"module_versions"`
		Name               string   `json:"name"`
		PRSName            string   `json:"prs_name"`
	}
	
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	
	// Reconstruct from SecurityAndVersion if provided
	if temp.SecurityAndVersion != 0 {
		sv := temp.SecurityAndVersion
		i.MinorVersion = uint8(sv & 0xFF)              // Bits 0-7
		i.SecurityMode = uint8((sv >> 8) & 0xFF)       // Bits 8-15
		i.MajorVersion = uint8((sv >> 16) & 0xFF)      // Bits 16-23
		i.Reserved0 = uint8((sv >> 24) & 0xFF)         // Bits 24-31
		
		// Extract individual flags from bits 8-15 (SecurityMode)
		i.MCCEnabled = (sv & (1 << 8)) != 0   // Bit 8
		i.DebugFW = (sv & (1 << 13)) != 0     // Bit 13
		i.SignedFW = (sv & (1 << 14)) != 0    // Bit 14
		i.SecureFW = (sv & (1 << 15)) != 0    // Bit 15
	} else {
		// Use individual fields
		i.MinorVersion = temp.MinorVersion
		i.SecurityMode = temp.SecurityMode
		i.MajorVersion = temp.MajorVersion
		i.Reserved0 = temp.Reserved0
		i.MCCEnabled = temp.MCCEnabled
		i.DebugFW = temp.DebugFW
		i.SignedFW = temp.SignedFW
		i.SecureFW = temp.SecureFW
	}
	
	// Copy other fields
	i.FWVerMajor = temp.FWVerMajor
	i.Reserved2 = temp.Reserved2
	i.FWVerMinor = temp.FWVerMinor
	i.FWVerSubminor = temp.FWVerSubminor
	i.Reserved3a = temp.Reserved3a
	i.Hour = temp.Hour
	i.Minutes = temp.Minutes
	i.Seconds = temp.Seconds
	i.Day = temp.Day
	i.Month = temp.Month
	i.Year = temp.Year
	i.MICVerMajor = temp.MICVerMajor
	i.Reserved4 = temp.Reserved4
	i.MICVerSubminor = temp.MICVerSubminor
	i.MICVerMinor = temp.MICVerMinor
	i.PCIDeviceID = temp.PCIDeviceID
	i.PCIVendorID = temp.PCIVendorID
	i.PCISubsystemID = temp.PCISubsystemID
	i.PCISubVendorID = temp.PCISubVendorID
	
	// PSID
	psidBytes := []byte(temp.PSID)
	if len(psidBytes) > 16 {
		psidBytes = psidBytes[:16]
	}
	copy(i.PSID[:], psidBytes)
	
	// VSD info
	i.Reserved5a = temp.Reserved5a
	i.VSDVendorID = temp.VSDVendorID
	vsdBytes := []byte(temp.VSD)
	if len(vsdBytes) > 208 {
		vsdBytes = vsdBytes[:208]
	}
	copy(i.VSD[:], vsdBytes)
	
	// Image size data
	copy(i.ImageSizeData[:], temp.ImageSizeData)
	copy(i.Reserved6[:], temp.Reserved6)
	
	// Supported HW IDs
	copy(i.SupportedHWID[:], temp.SupportedHWIDs)
	
	// INI file num
	i.INIFileNum = temp.INIFileNum
	copy(i.Reserved7[:], temp.Reserved7)
	
	// Product version - prefer raw bytes if available
	if len(temp.ProductVerRaw) > 0 {
		copy(i.ProductVer[:], temp.ProductVerRaw)
	} else {
		prodVerBytes := []byte(temp.ProductVer)
		if len(prodVerBytes) > 16 {
			prodVerBytes = prodVerBytes[:16]
		}
		copy(i.ProductVer[:], prodVerBytes)
	}
	
	// Description
	descBytes := []byte(temp.Description)
	if len(descBytes) > 256 {
		descBytes = descBytes[:256]
	}
	copy(i.Description[:], descBytes)
	copy(i.Reserved8[:], temp.Reserved8)
	
	// Module versions
	copy(i.ModuleVersions[:], temp.ModuleVersions)
	
	// Name fields
	nameBytes := []byte(temp.Name)
	if len(nameBytes) > 64 {
		nameBytes = nameBytes[:64]
	}
	copy(i.Name[:], nameBytes)
	
	prsBytes := []byte(temp.PRSName)
	if len(prsBytes) > 128 {
		prsBytes = prsBytes[:128]
	}
	copy(i.PRSName[:], prsBytes)
	
	return nil
}

