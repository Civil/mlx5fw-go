package types

import (
	"github.com/Civil/mlx5fw-go/pkg/annotations"
)


// FirmwareBootVersionAnnotated represents the firmware boot version structure using annotations
// Based on mstflint's image_layout_boot_version_unpack which shows the actual bit layout:
// - image_format_version at bits 0-7 (offset 0)
// - major_version at bits 16-23 (offset 16)
// - minor_version at bits 24-31 (offset 24)
type FirmwareBootVersionAnnotated struct {
	ImageFormatVersion  uint8 `offset:"byte:0"`                        // Bits 0-7
	Reserved            uint8 `offset:"byte:1,reserved:true"`          // Bits 8-15
	MajorVersion        uint8 `offset:"byte:2"`                        // Bits 16-23
	MinorVersion        uint8 `offset:"byte:3"`                        // Bits 24-31
}

// Unmarshal unmarshals binary data into FirmwareBootVersionAnnotated structure
func (b *FirmwareBootVersionAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, b)
}

// Marshal marshals FirmwareBootVersionAnnotated structure into binary data
func (b *FirmwareBootVersionAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(b)
}

