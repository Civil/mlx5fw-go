package interfaces

import "github.com/Civil/mlx5fw-go/pkg/types"

// Example usage of NewBaseSectionWithOptions:
//
//   // Simple section with minimal configuration
//   section := NewBaseSectionWithOptions(
//       types.SectionTypeBoot2,
//       0x1000,     // offset
//       0x2000,     // size
//   )
//
//   // Section with CRC configuration
//   section := NewBaseSectionWithOptions(
//       types.SectionTypeMainCode,
//       0x5000,
//       0x10000,
//       WithCRC(types.CRCInSection, 0x1234),
//   )
//
//   // Encrypted device data section with ITOC entry
//   section := NewBaseSectionWithOptions(
//       types.SectionTypeDevInfo,
//       0x8000,
//       0x200,
//       WithCRC(types.CRCInITOCEntry, 0x5678),
//       WithEncryption(),
//       WithDeviceData(),
//       WithITOCEntry(entry),
//   )
//
//   // Section from hardware pointer with custom CRC handler
//   section := NewBaseSectionWithOptions(
//       types.SectionTypeHashesTable,
//       0x10000,
//       0x800,
//       WithFromHWPointer(),
//       WithCRCHandler(customHandler),
//   )

// SectionOption is a functional option for configuring a BaseSection
type SectionOption func(*BaseSection)

// WithCRC sets the CRC type and value for the section
func WithCRC(crcType types.CRCType, crc uint32) SectionOption {
	return func(s *BaseSection) {
		s.SectionCRCType = crcType
		s.SectionCRC = crc
		// If CRC type is not NONE, assume section has CRC unless overridden
		if crcType != types.CRCNone {
			s.hasCRC = true
		}
	}
}

// WithEncryption sets whether the section is encrypted
func WithEncryption() SectionOption {
	return func(s *BaseSection) {
		s.EncryptedFlag = true
	}
}

// WithDeviceData sets whether the section contains device-specific data
func WithDeviceData() SectionOption {
	return func(s *BaseSection) {
		s.DeviceDataFlag = true
	}
}

// WithITOCEntry sets the ITOC entry for the section
func WithITOCEntry(entry *types.ITOCEntry) SectionOption {
	return func(s *BaseSection) {
		s.entry = entry
		// Update hasCRC based on entry's no_crc flag
		if entry != nil && entry.GetNoCRC() {
			s.hasCRC = false
		}
	}
}

// WithFromHWPointer sets whether the section is from a hardware pointer
func WithFromHWPointer() SectionOption {
	return func(s *BaseSection) {
		s.FromHWPointerFlag = true
	}
}

// WithNoCRC explicitly disables CRC for the section
func WithNoCRC() SectionOption {
	return func(s *BaseSection) {
		s.hasCRC = false
		s.SectionCRCType = types.CRCNone
	}
}

// WithCRCHandler sets a specific CRC handler for the section
func WithCRCHandler(handler SectionCRCHandler) SectionOption {
	return func(s *BaseSection) {
		s.crcHandler = handler
	}
}

// WithRawData sets pre-loaded raw data for the section
func WithRawData(data []byte) SectionOption {
	return func(s *BaseSection) {
		s.rawData = data
	}
}

// NewBaseSectionWithOptions creates a new base section using functional options
func NewBaseSectionWithOptions(sectionType uint16, offset uint64, size uint32, opts ...SectionOption) *BaseSection {
	// Create base section with required fields
	bs := &BaseSection{
		SectionType:   types.SectionType(sectionType),
		SectionOffset: offset,
		SectionSize:   size,
		// Default values
		SectionCRCType: types.CRCNone,
		hasCRC:         false,
	}
	
	// Apply options
	for _, opt := range opts {
		opt(bs)
	}
	
	return bs
}