package types

import (
	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// BinaryVersion represents the binary version structure used in Tools Area
// Based on tools_binary_header from mstflint
type BinaryVersion struct {
	Length   uint16 `offset:"byte:0,endian:be"`               // Length of the structure
	Type     uint8  `offset:"byte:2"`                         // Type identifier
	Version  uint8  `offset:"byte:3"`                         // Version number
	Reserved uint32 `offset:"byte:4,endian:be,reserved:true"` // Reserved field
}

// ToolsArea represents the tools area header structure
type ToolsArea struct {
	BinaryHeader BinaryVersion `offset:"byte:0"`
	// Data follows - variable length
}

// MagicPatternStruct represents the firmware magic pattern structure
// This is found at various offsets in the firmware image
type MagicPatternStruct struct {
	Magic    uint64 `offset:"byte:0,endian:be"`               // Should be MagicPattern constant
	Reserved uint64 `offset:"byte:8,endian:be,reserved:true"` // Reserved/padding
}

// Boot2Header represents the Boot2 section header
type Boot2Header struct {
	Magic    uint32 `offset:"byte:0,endian:be"`                // Magic value
	Version  uint32 `offset:"byte:4,endian:be"`                // Boot2 version
	Size     uint32 `offset:"byte:8,endian:be"`                // Size of Boot2 section
	Reserved uint32 `offset:"byte:12,endian:be,reserved:true"` // Reserved field
}

// HWIDRecord represents a hardware ID record
// Used for device identification
type HWIDRecord struct {
	HWID       uint32 `offset:"byte:0,endian:be"`                // Hardware ID
	ChipType   uint32 `offset:"byte:4,endian:be"`                // Chip type
	DeviceType uint32 `offset:"byte:8,endian:be"`                // Device type
	Reserved   uint32 `offset:"byte:12,endian:be,reserved:true"` // Reserved field
}

// Unmarshal unmarshals binary data into the annotated structure
func (b *BinaryVersion) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, b)
}

// Marshal marshals the annotated structure into binary data
func (b *BinaryVersion) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(b)
}

// Unmarshal unmarshals binary data into the annotated structure
func (t *ToolsArea) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, t)
}

// Marshal marshals the annotated structure into binary data
func (t *ToolsArea) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(t)
}

// Unmarshal unmarshals binary data into the annotated structure
func (m *MagicPatternStruct) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, m)
}

// Marshal marshals the annotated structure into binary data
func (m *MagicPatternStruct) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(m)
}

// Unmarshal unmarshals binary data into the annotated structure
func (b *Boot2Header) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, b)
}

// Marshal marshals the annotated structure into binary data
func (b *Boot2Header) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(b)
}

// Unmarshal unmarshals binary data into the annotated structure
func (h *HWIDRecord) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// Marshal marshals the annotated structure into binary data
func (h *HWIDRecord) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}
