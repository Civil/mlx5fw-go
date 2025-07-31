package reassemble

import "github.com/Civil/mlx5fw-go/pkg/types"

// FirmwareMetadata represents the metadata JSON structure
type FirmwareMetadata struct {
	Format      string                 `json:"format"`
	Firmware    FirmwareFileInfo       `json:"firmware_file"`
	Magic       MagicInfo              `json:"magic_pattern"`
	HWPointers  HWPointersInfo         `json:"hw_pointers"`
	ITOC        TOCInfo                `json:"itoc"`
	DTOC        TOCInfo                `json:"dtoc"`
	IsEncrypted bool                   `json:"is_encrypted"`
	Sections    []SectionMetadata      `json:"sections"`
	MemoryLayout []MemorySegment       `json:"memory_layout"`
	CRCInfo     CRCMetadata           `json:"crc_info"`
	Boundaries  BoundariesInfo        `json:"boundaries"`
}

type FirmwareFileInfo struct {
	OriginalSize int64  `json:"original_size"`
	SHA256Hash   string `json:"sha256_hash"`
}

type MagicInfo struct {
	Offset        uint32   `json:"offset"`
	Pattern       string   `json:"pattern"`
	SearchOffsets []uint32 `json:"search_offsets"`
}

type HWPointersInfo struct {
	Offset  uint32                 `json:"offset"`
	RawData string                 `json:"raw_data"`
	Parsed  *types.FS4HWPointers   `json:"parsed"`
}

type TOCInfo struct {
	Address     uint32 `json:"address"`
	HeaderValid bool   `json:"header_valid"`
	RawHeader   string `json:"raw_header"`
}

type SectionMetadata struct {
	Type             uint16                 `json:"type"`
	TypeName         string                 `json:"type_name"`
	Offset           uint64                 `json:"offset"`
	Size             uint32                 `json:"size"`
	OriginalSize     uint32                 `json:"original_size_with_crc"`
	CRCType          string                 `json:"crc_type"`
	CRCValue         uint32                 `json:"crc_value"`
	IsEncrypted      bool                   `json:"is_encrypted"`
	IsDeviceData     bool                   `json:"is_device_data"`
	IsFromHWPointer  bool                   `json:"is_from_hw_pointer"`
	Index            int                    `json:"index"`
	ITOCEntry        *ITOCEntryMetadata     `json:"itoc_entry,omitempty"`
	FileName         string                 `json:"file_name,omitempty"` // Actual filename used during extraction
}

type ITOCEntryMetadata struct {
	Type       uint16 `json:"type"`
	FlashAddr  uint32 `json:"flash_addr"`
	SectionCRC uint16 `json:"section_crc"`
	Encrypted  bool   `json:"encrypted"`
}

type MemorySegment struct {
	Type        string `json:"type"`
	SectionType string `json:"section_type,omitempty"`
	Start       uint64 `json:"start"`
	End         uint64 `json:"end"`
	Size        uint64 `json:"size"`
}

type CRCMetadata struct {
	HardwareCRCSections []string          `json:"hardware_crc_sections"`
	ImageCRCSections    []string          `json:"image_crc_sections"`
	Algorithms          map[string]string `json:"algorithms"`
}

type BoundariesInfo struct {
	ImageStart  int64 `json:"image_start"`
	ImageEnd    int64 `json:"image_end"`
	SectorSize  int   `json:"sector_size"`
}