package extracted

import (
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
)

// SectionMetadata represents metadata for an extracted section
// This is used in firmware_metadata.json and includes extraction-specific fields
type SectionMetadata struct {
	*interfaces.BaseSection
	OriginalSize uint32 `json:"original_size_with_crc,omitempty"`
	CRCValue     uint32 `json:"crc_value,omitempty"`
	Index        int    `json:"index"`
	FileName     string `json:"file_name,omitempty"` // Actual filename used during extraction
}

// FirmwareMetadata represents the complete firmware metadata structure
// written to firmware_metadata.json during extraction
type FirmwareMetadata struct {
	Format       string             `json:"format"`
	Firmware     FirmwareFileInfo   `json:"firmware_file"`
	Magic        MagicInfo          `json:"magic_pattern"`
	HWPointers   HWPointersInfo     `json:"hw_pointers"`
	ITOC         TOCInfo            `json:"itoc"`
	DTOC         TOCInfo            `json:"dtoc"`
	Sections     []SectionMetadata  `json:"sections"`
	MemoryLayout []MemorySegment    `json:"memory_layout"`
	Gaps         []GapInfo          `json:"gaps"`
	IsEncrypted  bool               `json:"is_encrypted"`
	CRCInfo      CRCInfo            `json:"crc_info,omitempty"`
	Boundaries   FirmwareBoundaries `json:"boundaries,omitempty"`
}

type FirmwareFileInfo struct {
	OriginalSize uint64 `json:"original_size"`
	SHA256Hash   string `json:"sha256_hash"`
}

type MagicInfo struct {
	Offset uint32 `json:"offset"`
	Value  string `json:"value"`
}

// HWPointersInfo stores HW pointers with proper type handling
type HWPointersInfo struct {
	Offset uint32                `json:"offset"`
	FS4    *types.FS4HWPointers  `json:"fs4,omitempty"`
	FS5    *types.FS5HWPointers  `json:"fs5,omitempty"`
}

// GetParsed returns the non-nil HW pointers (either FS4 or FS5)
func (h *HWPointersInfo) GetParsed() any {
	if h.FS4 != nil {
		return h.FS4
	}
	return h.FS5
}

type TOCInfo struct {
	Address     uint32 `json:"address"`
	HeaderValid bool   `json:"header_valid"`
	RawHeader   string `json:"raw_header"`
}

type MemorySegment struct {
	StartOffset uint64 `json:"start_offset"`
	EndOffset   uint64 `json:"end_offset"`
	Size        uint64 `json:"size"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type GapInfo struct {
	Index       int    `json:"index"`
	StartOffset uint64 `json:"start_offset"`
	EndOffset   uint64 `json:"end_offset"`
	Size        uint64 `json:"size"`
	FillByte    uint8  `json:"fill_byte"`
	IsUniform   bool   `json:"is_uniform"`
}

// CRCInfo provides human-readable CRC algorithm information
type CRCInfo struct {
	Algorithms struct {
		Hardware string `json:"hardware"`
		Software string `json:"software"`
	} `json:"algorithms"`
	Note string `json:"note"`
}

type FirmwareBoundaries struct {
	ImageStart uint64 `json:"image_start"`
	ImageEnd   uint64 `json:"image_end"`
	SectorSize uint32 `json:"sector_size"`
}