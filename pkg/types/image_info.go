package types

import (
	"fmt"
	"strings"
)

// ImageInfo represents the IMAGE_INFO section structure
// Based on mstflint's connectx4_image_info structure
// Note: This structure uses bit fields, so we need to handle them carefully
type ImageInfo struct {
	// DWORD 0 - offset 0x0 (32 bits total)
	SecurityAndVersion uint32 `bin:"len:4"` // Contains multiple bit fields
	
	// DWORD 1-4 - offset 0x4 - FW Version
	FWVerMajor     uint16 `bin:"len:2"` // offset 0x4
	Reserved2      uint16 `bin:"len:2"` // offset 0x6
	FWVerSubminor  uint16 `bin:"len:2"` // offset 0x8
	FWVerMinor     uint16 `bin:"len:2"` // offset 0xa
	
	// Release date/time - offset 0xc
	DateTimeWord   uint32 `bin:"len:4"` // offset 0xc - contains seconds/minutes/hour
	YearMonth      uint32 `bin:"len:4"` // offset 0x10 - contains year/month/day
	
	// MIC version - offset 0x14
	MICVerMajor    uint16 `bin:"len:2"` // offset 0x14
	Reserved4      uint16 `bin:"len:2"` // offset 0x16
	MICVerSubminor uint16 `bin:"len:2"` // offset 0x18
	MICVerMinor    uint16 `bin:"len:2"` // offset 0x1a
	
	// PCI IDs - offset 0x1c
	PCIDeviceID    uint16 `bin:"len:2"` // offset 0x1c
	PCIVendorID    uint16 `bin:"len:2"` // offset 0x1e
	PCISubsystemID uint16 `bin:"len:2"` // offset 0x20
	PCISubVendorID uint16 `bin:"len:2"` // offset 0x22
	
	// PSID - offset 0x24
	PSID           [16]byte `bin:"len:16"` // offset 0x24-0x33
	
	// VSD vendor ID - offset 0x34
	VSDVendorID    uint16   `bin:"len:2"`   // offset 0x34
	Reserved5      uint16   `bin:"len:2"`   // offset 0x36
	
	// VSD - offset 0x38
	VSD            [208]byte `bin:"len:208"` // offset 0x38-0x107
	
	// Image size - offset 0x108
	ImageSizeData  [8]byte   `bin:"len:8"`   // offset 0x108-0x10f
	
	// Skip to supported HW IDs - offset 0x110
	Reserved6      [8]byte   `bin:"len:8"`   // offset 0x110-0x117
	
	// Supported HW IDs - offset 0x118
	SupportedHWID  [4]uint32 `bin:"len:16"`  // offset 0x118-0x127
	
	// INI file num - offset 0x128
	INIFileNum     uint32    `bin:"len:4"`   // offset 0x128
	
	// Skip to product version - offset 0x12c
	Reserved7      [148]byte `bin:"len:148"` // offset 0x12c-0x1bf
	
	// Product version - offset 0x1c0
	ProductVer     [16]byte  `bin:"len:16"`  // offset 0x1c0-0x1cf
	
	// Description - offset 0x1d0
	Description    [256]byte `bin:"len:256"` // offset 0x1d0-0x2cf
	
	// Skip to module versions - offset 0x2d0
	Reserved8      [48]byte  `bin:"len:48"`  // offset 0x2d0-0x2ff
	
	// Module versions - offset 0x300
	ModuleVersions [64]byte  `bin:"len:64"`  // offset 0x300-0x33f
	
	// Name (part number) - offset 0x340
	Name           [64]byte  `bin:"len:64"`  // offset 0x340-0x37f
	
	// PRS name - offset 0x380
	PRSName        [128]byte `bin:"len:128"` // offset 0x380-0x3ff
}

// Extract bit fields from SecurityAndVersion
func (i *ImageInfo) IsMCCEnabled() bool {
	return (i.SecurityAndVersion & (1 << 8)) != 0
}

func (i *ImageInfo) IsDebugFW() bool {
	return (i.SecurityAndVersion & (1 << 13)) != 0
}

func (i *ImageInfo) IsSignedFW() bool {
	return (i.SecurityAndVersion & (1 << 14)) != 0
}

func (i *ImageInfo) IsSecureFW() bool {
	return (i.SecurityAndVersion & (1 << 15)) != 0
}

func (i *ImageInfo) GetMinorVersion() uint8 {
	return uint8((i.SecurityAndVersion >> 16) & 0xFF)
}

func (i *ImageInfo) GetMajorVersion() uint8 {
	return uint8((i.SecurityAndVersion >> 24) & 0xFF)
}

// Extract date/time fields
func (i *ImageInfo) GetSeconds() uint8 {
	return uint8(i.DateTimeWord & 0xFF)
}

func (i *ImageInfo) GetMinutes() uint8 {
	return uint8((i.DateTimeWord >> 8) & 0xFF)
}

func (i *ImageInfo) GetHour() uint8 {
	return uint8((i.DateTimeWord >> 16) & 0xFF)
}

func (i *ImageInfo) GetDay() uint8 {
	return uint8(i.YearMonth & 0xFF)
}

func (i *ImageInfo) GetMonth() uint8 {
	return uint8((i.YearMonth >> 8) & 0xFF)
}

func (i *ImageInfo) GetYear() uint16 {
	return uint16((i.YearMonth >> 16) & 0xFFFF)
}

// GetFWVersionString returns the firmware version as a string
func (i *ImageInfo) GetFWVersionString() string {
	return fmt.Sprintf("%d.%d.%d", i.FWVerMajor, i.FWVerMinor, i.FWVerSubminor)
}

// GetFWReleaseDateString returns the release date as a string
func (i *ImageInfo) GetFWReleaseDateString() string {
	// Format: DD.M.YYYY (remove leading zero from month)
	return fmt.Sprintf("%d.%d.%d", i.GetDay(), i.GetMonth(), i.GetYear())
}

// GetMICVersionString returns the MIC version as a string
func (i *ImageInfo) GetMICVersionString() string {
	return fmt.Sprintf("%d.%d.%d", i.MICVerMajor, i.MICVerMinor, i.MICVerSubminor)
}

// GetPSIDString returns the PSID as a string
func (i *ImageInfo) GetPSIDString() string {
	return nullTerminatedString(i.PSID[:])
}

// GetVSDString returns the VSD as a string
func (i *ImageInfo) GetVSDString() string {
	return nullTerminatedString(i.VSD[:])
}

// GetProductVerString returns the product version as a string
func (i *ImageInfo) GetProductVerString() string {
	return nullTerminatedString(i.ProductVer[:])
}

// GetDescriptionString returns the description as a string
func (i *ImageInfo) GetDescriptionString() string {
	return nullTerminatedString(i.Description[:])
}

// GetNameString returns the name (part number) as a string
func (i *ImageInfo) GetNameString() string {
	return nullTerminatedString(i.Name[:])
}

// GetPRSNameString returns the PRS name as a string
func (i *ImageInfo) GetPRSNameString() string {
	return nullTerminatedString(i.PRSName[:])
}

// GetSecurityMode returns the security mode bitmap
func (i *ImageInfo) GetSecurityMode() uint32 {
	var mode uint32
	if i.IsMCCEnabled() {
		mode |= SMMFlags.MCC_EN
	}
	if i.IsDebugFW() {
		mode |= SMMFlags.DEBUG_FW
	}
	if i.IsSignedFW() {
		mode |= SMMFlags.SIGNED_FW
	}
	if i.IsSecureFW() {
		mode |= SMMFlags.SECURE_FW
	}
	return mode
}

// GetSecurityAttributesString returns the security attributes as a string
func (i *ImageInfo) GetSecurityAttributesString() string {
	mode := i.GetSecurityMode()
	
	attrs := []string{}
	
	if mode&SMMFlags.SECURE_FW != 0 {
		attrs = append(attrs, "secure-fw")
	} else if mode&SMMFlags.SIGNED_FW != 0 {
		attrs = append(attrs, "signed-fw")
	} else {
		return "N/A"
	}
	
	if mode&SMMFlags.DEBUG_FW != 0 {
		attrs = append(attrs, "debug")
	}
	
	// Additional flags can be added here if needed
	
	return strings.Join(attrs, ", ")
}

// nullTerminatedString converts a byte array to string, stopping at first null byte
func nullTerminatedString(data []byte) string {
	for i, b := range data {
		if b == 0 {
			return string(data[:i])
		}
	}
	return string(data)
}