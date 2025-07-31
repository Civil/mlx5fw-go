package interfaces

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/Civil/mlx5fw-go/pkg/types"
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
	
	// MarshalJSON returns JSON representation of the parsed data
	MarshalJSON() ([]byte, error)
	
	// Write writes the section data to the writer
	Write(w io.Writer) error
	
	// GetITOCEntry returns the ITOC entry for this section (may be nil)
	GetITOCEntry() *types.ITOCEntry
	
	// IsFromHWPointer returns true if this section was referenced from HW pointers
	IsFromHWPointer() bool
}

// BaseSection provides common implementation for all sections
type BaseSection struct {
	sectionType     uint16
	offset          uint64
	size            uint32
	crcType         types.CRCType
	crc             uint32
	isEncrypted     bool
	isDeviceData    bool
	rawData         []byte
	entry           *types.ITOCEntry
	isFromHWPointer bool
	crcHandler      SectionCRCHandler
	hasCRC          bool // private field - not marshaled/unmarshaled
}

// Type returns the section type
func (b *BaseSection) Type() uint16 {
	return b.sectionType
}

// TypeName returns the human-readable name for this section type
func (b *BaseSection) TypeName() string {
	return types.GetSectionTypeName(b.sectionType)
}

// Offset returns the section offset in the firmware
func (b *BaseSection) Offset() uint64 {
	return b.offset
}

// Size returns the section size
func (b *BaseSection) Size() uint32 {
	return b.size
}

// CRCType returns the CRC type for this section
func (b *BaseSection) CRCType() types.CRCType {
	return b.crcType
}

// HasCRC returns whether this section has CRC verification enabled
func (b *BaseSection) HasCRC() bool {
	return b.hasCRC
}

// IsEncrypted returns whether the section is encrypted
func (b *BaseSection) IsEncrypted() bool {
	return b.isEncrypted
}

// IsDeviceData returns whether this is device-specific data
func (b *BaseSection) IsDeviceData() bool {
	return b.isDeviceData
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
	return b.isFromHWPointer
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
	return b.crcHandler.CalculateCRC(b.rawData, b.crcType)
}

// VerifyCRC verifies the section's CRC
func (b *BaseSection) VerifyCRC() error {
	if b.crcHandler == nil {
		return nil
	}
	
	// Get the expected CRC based on the CRC type
	var expectedCRC uint32
	if b.crcType == types.CRCInSection && b.crcHandler.HasEmbeddedCRC() {
		// CRC is embedded in the section data
		if len(b.rawData) < 4 {
			return fmt.Errorf("section too small for embedded CRC")
		}
		// Extract CRC from the last 4 bytes
		// For IN_SECTION CRCs, the format is:
		// - 16-bit CRC in upper 16 bits (bytes 0-1) 
		// - Lower 16 bits are 0 (bytes 2-3)
		expectedCRC = uint32(b.rawData[len(b.rawData)-4])<<8 |
			uint32(b.rawData[len(b.rawData)-3])
		// Verify CRC on data without the CRC bytes
		return b.crcHandler.VerifyCRC(b.rawData[:len(b.rawData)-4], expectedCRC, b.crcType)
	} else {
		// CRC is external (from ITOC entry)
		expectedCRC = b.crc
		return b.crcHandler.VerifyCRC(b.rawData, expectedCRC, b.crcType)
	}
}

// SetCRCHandler sets the CRC handler for this section
func (b *BaseSection) SetCRCHandler(handler SectionCRCHandler) {
	b.crcHandler = handler
}

// Parse parses the raw data into section-specific structures
func (b *BaseSection) Parse(data []byte) error {
	// Default implementation - just store raw data
	b.rawData = data
	return nil
}

// MarshalJSON returns JSON representation of the parsed data
func (b *BaseSection) MarshalJSON() ([]byte, error) {
	// Default implementation returns basic info
	return json.Marshal(map[string]interface{}{
		"type":         b.sectionType,
		"type_name":    b.TypeName(),
		"offset":       b.offset,
		"size":         b.size,
		"crc_type":     b.crcType,
		"is_encrypted": b.isEncrypted,
		"is_device_data": b.isDeviceData,
	})
}

// NewBaseSection creates a new base section with the given parameters
func NewBaseSection(sectionType uint16, offset uint64, size uint32, crcType types.CRCType, 
	crc uint32, isEncrypted, isDeviceData bool, entry *types.ITOCEntry, isFromHWPointer bool) *BaseSection {
	// Determine if section has CRC based on the entry's no_crc flag
	hasCRC := true
	if entry != nil && entry.GetNoCRC() {
		hasCRC = false
	}
	
	return &BaseSection{
		sectionType:     sectionType,
		offset:          offset,
		size:            size,
		crcType:         crcType,
		crc:             crc,
		isEncrypted:     isEncrypted,
		isDeviceData:    isDeviceData,
		entry:           entry,
		isFromHWPointer: isFromHWPointer,
		hasCRC:          hasCRC,
	}
}

// SectionFactory creates section instances based on type
type SectionFactory interface {
	// CreateSection creates a new section instance for the given type
	CreateSection(sectionType uint16, offset uint64, size uint32, crcType types.CRCType,
		crc uint32, isEncrypted, isDeviceData bool, entry *types.ITOCEntry, isFromHWPointer bool) (SectionInterface, error)
}