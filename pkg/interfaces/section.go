package interfaces

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/Civil/mlx5fw-go/pkg/types"
)

// CRC-related constants
const (
	MinCRCSectionSize = 4 // Minimum size for embedded CRC
	CRCByteSize       = 4 // Size of CRC field in bytes
)

// SectionInterface defines the interface all section types must implement
type SectionInterface interface {
	// Type returns the section type
	Type() uint16
	
	// TypeName returns the human-readable name for this section type
	TypeName() string
	
	// Offset returns the section offset in the firmware
	Offset() uint64
	
	// Size returns the section size
	Size() uint32
	
	// CRCType returns the CRC type for this section
	CRCType() types.CRCType
	
	// HasCRC returns whether this section has CRC verification enabled
	HasCRC() bool
	
	// GetCRC returns the expected CRC value for this section
	GetCRC() uint32
	
	// CalculateCRC calculates the CRC for this section
	CalculateCRC() (uint32, error)
	
	// VerifyCRC verifies the section's CRC
	VerifyCRC() error
	
	// IsEncrypted returns whether the section is encrypted
	IsEncrypted() bool
	
	// IsDeviceData returns whether this is device-specific data
	IsDeviceData() bool
	
	// Parse parses the raw data into section-specific structures
	Parse(data []byte) error
	
	// GetRawData returns the raw section data
	GetRawData() []byte
	
	// Write writes the section data to the writer
	Write(w io.Writer) error
	
	// GetITOCEntry returns the ITOC entry for this section (may be nil)
	GetITOCEntry() *types.ITOCEntry
	
	// IsFromHWPointer returns true if this section was referenced from HW pointers
	IsFromHWPointer() bool
}

// BaseSection provides common implementation for all sections
type BaseSection struct {
	SectionType       types.SectionType `json:"type" offset:"-"`
	SectionOffset     uint64            `json:"offset" offset:"-"`
	SectionSize       uint32            `json:"size" offset:"-"`
	SectionCRCType    types.CRCType     `json:"crc_type" offset:"-"`
	SectionCRC        uint32            `json:"-" offset:"-"` // Internal use only
	EncryptedFlag     bool              `json:"is_encrypted" offset:"-"`
	DeviceDataFlag    bool              `json:"is_device_data" offset:"-"`
	HasRawData        bool              `json:"has_raw_data" offset:"-"` // Set by sections based on parse status
	FromHWPointerFlag bool              `json:"is_from_hw_pointer" offset:"-"`
	rawData           []byte            `json:"-" offset:"-"` // Internal use only
	entry             *types.ITOCEntry  `json:"-" offset:"-"` // Internal use only
	crcHandler        SectionCRCHandler `json:"-" offset:"-"` // Internal use only
	hasCRC            bool              `json:"-" offset:"-"` // Internal use only
}

// Type returns the section type
func (b *BaseSection) Type() uint16 {
	return uint16(b.SectionType)
}

// TypeName returns the human-readable name for this section type
func (b *BaseSection) TypeName() string {
	return types.GetSectionTypeName(uint16(b.SectionType))
}

// Offset returns the section offset in the firmware
func (b *BaseSection) Offset() uint64 {
	return b.SectionOffset
}

// Size returns the section size
func (b *BaseSection) Size() uint32 {
	return b.SectionSize
}

// CRCType returns the CRC type for this section
func (b *BaseSection) CRCType() types.CRCType {
	return b.SectionCRCType
}

// HasCRC returns whether this section has CRC verification enabled
func (b *BaseSection) HasCRC() bool {
	return b.hasCRC
}

// GetCRC returns the expected CRC value for this section
func (b *BaseSection) GetCRC() uint32 {
	return b.SectionCRC
}

// IsEncrypted returns whether the section is encrypted
func (b *BaseSection) IsEncrypted() bool {
	return b.EncryptedFlag
}

// IsDeviceData returns whether this is device-specific data
func (b *BaseSection) IsDeviceData() bool {
	return b.DeviceDataFlag
}

// GetRawData returns the raw section data
func (b *BaseSection) GetRawData() []byte {
	return b.rawData
}

// SetRawData sets the raw section data
func (b *BaseSection) SetRawData(data []byte) {
	b.rawData = data
}

// GetEntry returns the ITOC entry for this section
func (b *BaseSection) GetEntry() *types.ITOCEntry {
	return b.entry
}

// GetITOCEntry returns the ITOC entry for this section (may be nil)
func (b *BaseSection) GetITOCEntry() *types.ITOCEntry {
	return b.entry
}

// IsFromHWPointer returns whether section was discovered from HW pointer
func (b *BaseSection) IsFromHWPointer() bool {
	return b.FromHWPointerFlag
}

// Write writes the section data to the writer
func (b *BaseSection) Write(w io.Writer) error {
	_, err := w.Write(b.rawData)
	return err
}

// Default implementations that can be overridden by specific section types

// CalculateCRC calculates the CRC for this section
func (b *BaseSection) CalculateCRC() (uint32, error) {
	if b.crcHandler == nil {
		return 0, nil
	}
	return b.crcHandler.CalculateCRC(b.rawData, b.SectionCRCType)
}

// VerifyCRC verifies the section's CRC
func (b *BaseSection) VerifyCRC() error {
	if b.crcHandler == nil {
		return nil
	}
	
	// Get the expected CRC based on the CRC type
	var expectedCRC uint32
	if b.SectionCRCType == types.CRCInSection && b.crcHandler.HasEmbeddedCRC() {
		// CRC is embedded in the section data
		if len(b.rawData) < MinCRCSectionSize {
			return fmt.Errorf("section size %d too small for embedded CRC (minimum %d bytes)", 
				len(b.rawData), MinCRCSectionSize)
		}
		// Extract CRC from the last 4 bytes
		// For IN_SECTION CRCs, the format is:
		// - 16-bit CRC in upper 16 bits (bytes 0-1) 
		// - Lower 16 bits are 0 (bytes 2-3)
		crcBytes := b.rawData[len(b.rawData)-CRCByteSize:]
		expectedCRC = binary.BigEndian.Uint32(crcBytes)
		// Pass full data - the handler will handle CRC extraction
		return b.crcHandler.VerifyCRC(b.rawData, expectedCRC, b.SectionCRCType)
	} else {
		// CRC is external (from ITOC entry)
		expectedCRC = b.SectionCRC
		return b.crcHandler.VerifyCRC(b.rawData, expectedCRC, b.SectionCRCType)
	}
}

// SetCRCHandler sets the CRC handler for this section
func (b *BaseSection) SetCRCHandler(handler SectionCRCHandler) {
	b.crcHandler = handler
}

// GetCRCHandler returns the CRC handler for this section
func (b *BaseSection) GetCRCHandler() SectionCRCHandler {
	return b.crcHandler
}

// Parse parses the raw data into section-specific structures
func (b *BaseSection) Parse(data []byte) error {
	// Default implementation - just store raw data
	b.rawData = data
	return nil
}


// NewBaseSection creates a new base section with the given parameters
// Deprecated: Use NewBaseSectionWithOptions for new code
func NewBaseSection(sectionType uint16, offset uint64, size uint32, crcType types.CRCType, 
	crc uint32, isEncrypted, isDeviceData bool, entry *types.ITOCEntry, isFromHWPointer bool) *BaseSection {
	
	// Build options list from parameters
	opts := []SectionOption{
		WithCRC(crcType, crc),
	}
	
	if isEncrypted {
		opts = append(opts, WithEncryption())
	}
	
	if isDeviceData {
		opts = append(opts, WithDeviceData())
	}
	
	if entry != nil {
		opts = append(opts, WithITOCEntry(entry))
	}
	
	if isFromHWPointer {
		opts = append(opts, WithFromHWPointer())
	}
	
	// Handle the special case where entry has no_crc flag
	if entry != nil && entry.GetNoCRC() {
		opts = append(opts, WithNoCRC())
	}
	
	return NewBaseSectionWithOptions(sectionType, offset, size, opts...)
}

// SectionFactory creates section instances based on type
type SectionFactory interface {
	// CreateSection creates a new section instance for the given type
	CreateSection(sectionType uint16, offset uint64, size uint32, crcType types.CRCType,
		crc uint32, isEncrypted, isDeviceData bool, entry *types.ITOCEntry, isFromHWPointer bool) (CompleteSectionInterface, error)
	
	// CreateSectionFromData creates a section and parses its data
	CreateSectionFromData(sectionType uint16, offset uint64, size uint32, crcType types.CRCType,
		crc uint32, isEncrypted, isDeviceData bool, entry *types.ITOCEntry, isFromHWPointer bool, data []byte) (CompleteSectionInterface, error)
}