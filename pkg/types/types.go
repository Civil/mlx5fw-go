// Package types contains all data structures for Mellanox firmware parsing
package types

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

// FirmwareMetadata contains parsed firmware metadata
type FirmwareMetadata struct {
	Format         FirmwareFormat
	ImageStart     uint32
	ImageSize      uint64
	ChunkSize      uint64
	IsEncrypted    bool
	HWPointers     interface{} // Either *FS4HWPointers or *FS5HWPointers
	ITOCHeader     *ITOCHeader
	DTOCHeader     *ITOCHeader // DTOC uses same header structure as ITOC
	ImageInfo      *ImageInfo
	DeviceInfo     *DeviceInfo
	MFGInfo        *MFGInfo
}