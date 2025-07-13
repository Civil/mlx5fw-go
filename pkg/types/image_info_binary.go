package types

import (
	"fmt"
)

// ImageInfoBinary represents the binary IMAGE_INFO section structure
// Based on analysis of actual firmware dumps and mstflint's image_layout_image_info structure
type ImageInfoBinary struct {
	// First 16 bytes contain various fields including version
	MajorVersion   uint16    // offset 0x0 - IMAGE_INFO major version (3)
	MinorVersion   uint16    // offset 0x2 - IMAGE_INFO minor version (352)
	FWVerMajor     uint16    // offset 0x4 - FW major version (e.g., 16)
	Reserved1      uint16    // offset 0x6
	FWVerMinor     uint16    // offset 0x8 - FW minor version (e.g., 35)
	FWVerSubminor  uint16    // offset 0xa - FW subminor version (e.g., 4030)
	Reserved2      uint32    // offset 0xc
	
	// Date and other fields - offset 0x10
	FWReleaseDate  [8]byte   // offset 0x10 - Release date (8 bytes)
	Reserved3      [12]byte  // offset 0x18 - padding
	PSID           [16]byte  // offset 0x24 - PSID string "MT_0000000013"
	Reserved4      [220]byte // offset 0x34 - padding to 0x110 (lots of zeros)
	
	// Binary data section starts around 0x110
	Reserved5      [16]byte  // offset 0x110 - padding
	// UIDs section at 0x120
	BaseGUIDNumAllocated    uint8    // offset 0x120 - num allocated
	BaseGUIDStep            uint8    // offset 0x121 - step (not used for CX4+)
	BaseGUIDNumAllocatedMSB uint8    // offset 0x122 - num allocated MSB
	BaseGUID                uint64   // offset 0x123 - base GUID
	BaseMACNumAllocated     uint8    // offset 0x12b - num allocated
	BaseMACStep             uint8    // offset 0x12c - step (not used for CX4+)
	BaseMACNumAllocatedMSB  uint8    // offset 0x12d - num allocated MSB
	BaseMAC                 uint64   // offset 0x12e - base MAC
	Reserved5b              [74]byte // offset 0x136 - padding to 0x180
	
	// Text fields appear later - offset 0x180+
	Reserved6      [48]byte  // offset 0x180 - padding
	Reserved7      [16]byte  // offset 0x1b0 - more padding
	ProductVer     [16]byte  // offset 0x1c0 - "rel-16_35_4030"
	Description    [256]byte // offset 0x1d0 - Description string (up to 0x2d0)
	Reserved8      [96]byte  // offset 0x2d0 - padding to 0x330
	Reserved9      [16]byte  // offset 0x330 - padding  
	PartNumber     [32]byte  // offset 0x340 - Part number "MCX516A-CDA_Ax_Bx"
	Reserved10     [32]byte  // offset 0x360 - padding
	PRSName        [64]byte  // offset 0x380 - PRS name "cx5_MCX516A_2p_x16_pci4.prs"
}

// GetFWVersionString returns the firmware version as a string
func (i *ImageInfoBinary) GetFWVersionString() string {
	return fmt.Sprintf("%d.%d.%04d", i.FWVerMajor, i.FWVerMinor, i.FWVerSubminor)
}

// GetFWReleaseDateString returns the release date as a string
func (i *ImageInfoBinary) GetFWReleaseDateString() string {
	// Parse the date from the binary format
	// The date appears to be stored as "20240627" in the first 8 bytes
	if len(i.FWReleaseDate) >= 8 {
		year := fmt.Sprintf("%02x%02x", i.FWReleaseDate[0], i.FWReleaseDate[1])
		month := fmt.Sprintf("%02x", i.FWReleaseDate[2])
		day := fmt.Sprintf("%02x", i.FWReleaseDate[3])
		// Remove leading zeros from day and month for mstflint compatibility
		var dayInt, monthInt int
		fmt.Sscanf(day, "%d", &dayInt)
		fmt.Sscanf(month, "%d", &monthInt)
		return fmt.Sprintf("%d.%d.%s", dayInt, monthInt, year)
	}
	return ""
}

// GetPSIDString returns the PSID as a string
func (i *ImageInfoBinary) GetPSIDString() string {
	return nullTerminatedString(i.PSID[:])
}

// GetPRSNameString returns the PRS name as a string
func (i *ImageInfoBinary) GetPRSNameString() string {
	return nullTerminatedString(i.PRSName[:])
}

// GetPartNumberString returns the part number as a string
func (i *ImageInfoBinary) GetPartNumberString() string {
	return nullTerminatedString(i.PartNumber[:])
}

// GetDescriptionString returns the description as a string
func (i *ImageInfoBinary) GetDescriptionString() string {
	return nullTerminatedString(i.Description[:])
}

// GetProductVerString returns the product version as a string
func (i *ImageInfoBinary) GetProductVerString() string {
	return nullTerminatedString(i.ProductVer[:])
}