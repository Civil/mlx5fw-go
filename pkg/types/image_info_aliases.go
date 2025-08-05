package types

import (
	"fmt"
	"strings"
)

// Type aliases to use annotated versions directly
type ImageInfo = ImageInfoAnnotated
type ImageInfoBinary = ImageInfoAnnotated

// Preserved methods from original ImageInfo

func (i *ImageInfo) IsMCCEnabled() bool {
	return i.MCCEnabled
}

func (i *ImageInfo) IsDebugFW() bool {
	return i.DebugFW
}

func (i *ImageInfo) IsSignedFW() bool {
	return i.SignedFW
}

func (i *ImageInfo) IsSecureFW() bool {
	return i.SecureFW
}

func (i *ImageInfo) GetMinorVersion() uint8 {
	return i.MinorVersion
}

func (i *ImageInfo) GetMajorVersion() uint8 {
	return i.MajorVersion
}

// GetSeconds returns the seconds
func (i *ImageInfo) GetSeconds() uint8 {
	return i.Seconds
}

func (i *ImageInfo) GetMinutes() uint8 {
	return i.Minutes
}

func (i *ImageInfo) GetHour() uint8 {
	return i.Hour
}

func (i *ImageInfo) GetDay() uint8 {
	return i.Day
}

func (i *ImageInfo) GetMonth() uint8 {
	return i.Month
}

func (i *ImageInfo) GetYear() uint16 {
	return i.Year
}

// GetFWVersionString returns the firmware version as a string
func (i *ImageInfo) GetFWVersionString() string {
	return fmt.Sprintf("%d.%d.%04d", i.FWVerMajor, i.FWVerMinor, i.FWVerSubminor)
}

// GetFWReleaseDateString returns the firmware release date as a string
func (i *ImageInfo) GetFWReleaseDateString() string {
	return fmt.Sprintf("%d.%d.%d", i.GetDay(), i.GetMonth(), i.GetYear())
}

// GetMICVersionString returns the MIC version as a string
func (i *ImageInfo) GetMICVersionString() string {
	return fmt.Sprintf("%d.%d.%04d", i.MICVerMajor, i.MICVerMinor, i.MICVerSubminor)
}

// GetPSIDString returns the PSID as a string
func (i *ImageInfo) GetPSIDString() string {
	return nullTerminatedString(i.PSID[:])
}

// GetVSDString returns the VSD as a string
func (i *ImageInfo) GetVSDString() string {
	return nullTerminatedString(i.VSD[:])
}

// GetProductVerString returns the product version string
func (i *ImageInfo) GetProductVerString() string {
	if i.ProductVer != [16]byte{} {
		return nullTerminatedString(i.ProductVer[:])
	}
	return ""
}

// GetDescriptionString returns a description string
func (i *ImageInfo) GetDescriptionString() string {
	if i.Description != [256]byte{} {
		return nullTerminatedString(i.Description[:])
	}
	return fmt.Sprintf("FW %s", i.GetFWVersionString())
}

// GetNameString returns the firmware name
func (i *ImageInfo) GetNameString() string {
	if i.Name != [64]byte{} {
		return nullTerminatedString(i.Name[:])
	}
	return "mlx5_fw"
}

// GetPRSNameString returns the PRS name
func (i *ImageInfo) GetPRSNameString() string {
	if i.PRSName != [128]byte{} {
		return nullTerminatedString(i.PRSName[:])
	}
	return ""
}

// GetSecurityMode returns the security mode bits
func (i *ImageInfo) GetSecurityMode() uint32 {
	return uint32(i.SecurityMode)
}

// GetSecurityAndVersion reconstructs the SecurityAndVersion field from individual bitfields
func (i *ImageInfo) GetSecurityAndVersion() uint32 {
	var result uint32
	result |= uint32(i.MinorVersion) & 0xFF           // Bits 0-7
	result |= (uint32(i.SecurityMode) & 0xFF) << 8    // Bits 8-15 (includes MCCEnabled)
	result |= (uint32(i.MajorVersion) & 0xFF) << 16   // Bits 16-23
	result |= (uint32(i.Reserved0) & 0xFF) << 24      // Bits 24-31
	
	// Individual flags are already included in SecurityMode (bits 8-15)
	// bit 8: MCCEnabled, bit 13: DebugFW, bit 14: SignedFW, bit 15: SecureFW
	
	return result
}

// SecurityAttribute constants
const (
	SecurityAttributeNone         = 0x0
	SecurityAttributeEncrypted    = 0x1
	SecurityAttributeSigned       = 0x2
	SecurityAttributeCsCrypto     = 0x4
	SecurityAttributeCsDevelopment = 0x8
	SecurityAttributeDebugFW      = 0x10
)

// GetSecurityAttributesString returns a string describing security attributes
func (i *ImageInfo) GetSecurityAttributesString() string {
	secMode := i.GetSecurityMode()
	
	if secMode == SecurityAttributeNone {
		return "None"
	}
	
	var attrs []string
	
	if secMode&SecurityAttributeEncrypted != 0 {
		attrs = append(attrs, "Encrypted")
	}
	if secMode&SecurityAttributeSigned != 0 {
		attrs = append(attrs, "Signed")
	}
	if secMode&SecurityAttributeCsCrypto != 0 {
		attrs = append(attrs, "CS_CRYPTO")
	}
	if secMode&SecurityAttributeCsDevelopment != 0 {
		attrs = append(attrs, "CS_DEVELOPMENT")
	}
	if secMode&SecurityAttributeDebugFW != 0 {
		attrs = append(attrs, "DEBUG_FW")
	}
	
	return strings.Join(attrs, ", ")
}

// GetPartNumberString returns the part number as a string
func (i *ImageInfo) GetPartNumberString() string {
	// In ImageInfoAnnotated, the part number is in the Name field
	if i.Name != [64]byte{} {
		return nullTerminatedString(i.Name[:])
	}
	return ""
}

