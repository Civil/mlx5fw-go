package sections

import (
	"encoding/json"
	
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// GenericSection represents a generic/unknown section type
type GenericSection struct {
	*interfaces.BaseSection
	crcCalc *parser.CRCCalculator
}

// NewGenericSection creates a new generic section
func NewGenericSection(base *interfaces.BaseSection) *GenericSection {
	return &GenericSection{
		BaseSection: base,
		crcCalc:     parser.NewCRCCalculator(),
	}
}

// Parse stores the raw data for generic sections
func (s *GenericSection) Parse(data []byte) error {
	s.SetRawData(data)
	return nil
}

// CalculateCRC calculates the CRC for the generic section
func (s *GenericSection) CalculateCRC() (uint32, error) {
	switch s.CRCType() {
	case types.CRCInSection:
		// CRC is last 4 bytes of section
		data := s.GetRawData()
		if len(data) < 4 {
			return 0, merry.New("section too small for CRC")
		}
		
		// Calculate CRC over data excluding last 4 bytes
		crcData := data[:len(data)-4]
		crc := s.crcCalc.CalculateSoftwareCRC16(crcData)
		return uint32(crc), nil
		
	case types.CRCInITOCEntry:
		// CRC is in ITOC entry, not in section data
		return 0, nil
		
	case types.CRCNone:
		// No CRC
		return 0, nil
		
	default:
		return 0, merry.Errorf("unknown CRC type: %d", s.CRCType())
	}
}

// VerifyCRC verifies the section's CRC based on its CRC type
func (s *GenericSection) VerifyCRC() error {
	entry := s.GetEntry()
	if entry == nil {
		return nil // No entry, no CRC to verify
	}
	
	switch s.CRCType() {
	case types.CRCInSection:
		// CRC is in the last 4 bytes of section
		data := s.GetRawData()
		if len(data) < 4 {
			return merry.New("section too small for CRC")
		}
		
		// For IN_SECTION CRCs, the format is:
		// - 16-bit CRC in upper 16 bits (bytes 0-1)  
		// - Lower 16 bits are 0 (bytes 2-3)
		// Stored as big-endian uint32
		expectedCRCFull := uint32(data[len(data)-4])<<24 | uint32(data[len(data)-3])<<16 |
			uint32(data[len(data)-2])<<8 | uint32(data[len(data)-1])
		expectedCRC := uint16(expectedCRCFull >> 16) // Extract upper 16 bits
		
		calculatedCRC, err := s.CalculateCRC()
		if err != nil {
			return err
		}
		
		if expectedCRC != uint16(calculatedCRC) {
			return &interfaces.CRCMismatchError{
				Expected: uint32(expectedCRC),
				Actual:   calculatedCRC,
			}
		}
		
	case types.CRCInITOCEntry:
		// CRC is in ITOC entry
		if entry.GetNoCRC() {
			return nil // No CRC to verify
		}
		
		expectedCRC := entry.GetCRC()
		data := s.GetRawData()
		
		// Calculate CRC over entire section
		crc := s.crcCalc.CalculateSoftwareCRC16(data)
		calculatedCRC := uint32(crc)
		
		if uint16(expectedCRC) != uint16(calculatedCRC) {
			return &interfaces.CRCMismatchError{
				Expected: uint32(expectedCRC),
				Actual:   calculatedCRC,
			}
		}
		
	case types.CRCNone:
		// No CRC to verify
		return nil
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the generic section
func (s *GenericSection) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":         s.Type(),
		"type_name":    s.TypeName(),
		"offset":       s.Offset(),
		"size":         s.Size(),
		"crc_type":     s.CRCType(),
		"is_encrypted": s.IsEncrypted(),
		"is_device_data": s.IsDeviceData(),
		"data_size":    len(s.GetRawData()),
		"has_raw_data": true,  // Flag indicating binary file is needed for reconstruction
	})
}