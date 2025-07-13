package types

// BinaryVersion represents the binary version structure used in Tools Area
// Based on tools_binary_header from mstflint
type BinaryVersion struct {
	Length      uint16 `bin:"BE"` // Length of the structure
	Type        uint8  `bin:""`   // Type identifier
	Version     uint8  `bin:""`   // Version number
	Reserved    uint32 `bin:"BE"` // Reserved field
}

// ToolsArea represents the tools area header structure
type ToolsArea struct {
	BinaryHeader BinaryVersion
	Data         []byte // Variable length data follows
}

// MagicPatternStruct represents the firmware magic pattern structure
// This is found at various offsets in the firmware image
type MagicPatternStruct struct {
	Magic    uint64 `bin:"BE"` // Should be MagicPattern constant
	Reserved uint64 `bin:"BE"` // Reserved/padding
}

// Boot2Header represents the Boot2 section header
type Boot2Header struct {
	Magic       uint32 `bin:"BE"` // Magic value
	Version     uint32 `bin:"BE"` // Boot2 version
	Size        uint32 `bin:"BE"` // Size of Boot2 section
	Reserved    uint32 `bin:"BE"` // Reserved field
}

// HWIDRecord represents a hardware ID record
// Used for device identification
type HWIDRecord struct {
	HWID        uint32 `bin:"BE"` // Hardware ID
	ChipType    uint32 `bin:"BE"` // Chip type
	DeviceType  uint32 `bin:"BE"` // Device type
	Reserved    uint32 `bin:"BE"` // Reserved field
}