// Package types contains all data structures for Mellanox firmware parsing
package types

import (
	"encoding/json"
	"fmt"
)

// FirmwareFormat represents the firmware format type
type FirmwareFormat int

const (
	// FormatUnknown indicates unknown firmware format
	FormatUnknown FirmwareFormat = iota
	// FormatFS4 indicates FS4 firmware format
	FormatFS4
	// FormatFS5 indicates FS5 firmware format
	FormatFS5
)

// String returns the string representation of the firmware format
func (f FirmwareFormat) String() string {
	switch f {
	case FormatFS4:
		return "FS4"
	case FormatFS5:
		return "FS5"
	default:
		return "Unknown"
	}
}

// MarshalJSON implements json.Marshaler interface
func (f FirmwareFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

// UnmarshalJSON implements json.Unmarshaler interface
func (f *FirmwareFormat) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "FS4":
		*f = FormatFS4
	case "FS5":
		*f = FormatFS5
	case "Unknown":
		*f = FormatUnknown
	default:
		return fmt.Errorf("unknown firmware format: %s", s)
	}

	return nil
}

// CRCType represents the type of CRC verification
type CRCType uint8

const (
	// CRCInITOCEntry means CRC is stored in ITOC entry
	CRCInITOCEntry CRCType = 0
	// CRCNone means no CRC verification
	CRCNone CRCType = 1
	// CRCInSection means CRC is stored at end of section
	CRCInSection CRCType = 2
)

// String returns the string representation of the CRC type
func (c CRCType) String() string {
	switch c {
	case CRCInITOCEntry:
		return "IN_ITOC_ENTRY"
	case CRCNone:
		return "NONE"
	case CRCInSection:
		return "IN_SECTION"
	default:
		return "UNKNOWN"
	}
}

// MarshalJSON implements json.Marshaler interface
func (c CRCType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// UnmarshalJSON implements json.Unmarshaler interface
func (c *CRCType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "IN_ITOC_ENTRY":
		*c = CRCInITOCEntry
	case "NONE":
		*c = CRCNone
	case "IN_SECTION":
		*c = CRCInSection
	case "UNKNOWN":
		// Handle unknown as NONE for compatibility
		*c = CRCNone
	default:
		return fmt.Errorf("unknown CRC type: %s", s)
	}

	return nil
}

// SectionType represents a firmware section type
type SectionType uint16

// MarshalJSON implements json.Marshaler interface
func (s SectionType) MarshalJSON() ([]byte, error) {
	// Return both hex ID and name for better readability
	return json.Marshal(struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{
		ID:   fmt.Sprintf("0x%x", uint16(s)),
		Name: GetSectionTypeName(uint16(s)),
	})
}

// UnmarshalJSON implements json.Unmarshaler interface
func (s *SectionType) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a struct first (new format)
	var st struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &st); err == nil && st.ID != "" {
		var val uint16
		if _, err := fmt.Sscanf(st.ID, "0x%x", &val); err == nil {
			*s = SectionType(val)
			return nil
		}
	}

	// Try to unmarshal as a number (backward compatibility)
	var val uint16
	if err := json.Unmarshal(data, &val); err == nil {
		*s = SectionType(val)
		return nil
	}

	return fmt.Errorf("cannot unmarshal section type from %s", string(data))
}

// FirmwareMetadata contains parsed firmware metadata
type FirmwareMetadata struct {
	Format      FirmwareFormat
	ImageStart  uint32
	ImageSize   uint64
	ChunkSize   uint64
	IsEncrypted bool
	HWPointers  interface{} // Either *FS4HWPointers or *FS5HWPointers
	ITOCHeader  *ITOCHeader
	DTOCHeader  *ITOCHeader // DTOC uses same header structure as ITOC
	ImageInfo   *ImageInfo
	DeviceInfo  *DevInfo
	MFGInfo     *MfgInfo
}
