package interfaces

import (
	"io"

	"github.com/Civil/mlx5fw-go/pkg/types"
)

// SectionMetadata provides basic metadata about a section
type SectionMetadata interface {
	// Type returns the section type
	Type() uint16
	
	// TypeName returns the human-readable name for this section type
	TypeName() string
	
	// Offset returns the section offset in the firmware
	Offset() uint64
	
	// Size returns the section size
	Size() uint32
}

// SectionAttributes provides access to section attributes
type SectionAttributes interface {
	// IsEncrypted returns whether the section is encrypted
	IsEncrypted() bool
	
	// IsDeviceData returns whether this is device-specific data
	IsDeviceData() bool
	
	// IsFromHWPointer returns true if this section was referenced from HW pointers
	IsFromHWPointer() bool
}

// SectionCRCInfo provides CRC-related information
type SectionCRCInfo interface {
	// CRCType returns the CRC type for this section
	CRCType() types.CRCType
	
	// HasCRC returns whether this section has CRC verification enabled
	HasCRC() bool
	
	// GetCRC returns the expected CRC value for this section
	GetCRC() uint32
}

// SectionCRCOperations provides CRC calculation and verification
type SectionCRCOperations interface {
	// CalculateCRC calculates the CRC for this section
	CalculateCRC() (uint32, error)
	
	// VerifyCRC verifies the section's CRC
	VerifyCRC() error
}

// SectionData provides access to section data
type SectionData interface {
	// Parse parses the raw data into section-specific structures
	Parse(data []byte) error
	
	// GetRawData returns the raw section data
	GetRawData() []byte
	
	// Write writes the section data to the writer
	Write(w io.Writer) error
}

// SectionExtras provides additional section information
type SectionExtras interface {
	// GetITOCEntry returns the ITOC entry for this section (may be nil)
	GetITOCEntry() *types.ITOCEntry
}

// SectionReader combines interfaces for reading section information
type SectionReader interface {
	SectionMetadata
	SectionAttributes
	SectionCRCInfo
	SectionExtras
}

// SectionParser combines interfaces for parsing sections
type SectionParser interface {
	SectionReader
	SectionData
}

// SectionVerifier combines interfaces for verifying sections
type SectionVerifier interface {
	SectionReader
	SectionCRCOperations
	SectionData
}

// CompleteSectionInterface combines all section interfaces
// This is equivalent to SectionInterface but built from smaller interfaces
type CompleteSectionInterface interface {
	SectionMetadata
	SectionAttributes
	SectionCRCInfo
	SectionCRCOperations
	SectionData
	SectionExtras
}