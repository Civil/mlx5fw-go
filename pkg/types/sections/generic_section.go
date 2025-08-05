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

// VerifyCRC verifies the section's CRC using the BaseSection implementation
// which properly uses the CRC handlers set up by the factory
func (s *GenericSection) VerifyCRC() error {
	// Use the BaseSection implementation which handles CRC handlers properly
	return s.BaseSection.VerifyCRC()
}

// MarshalJSON returns JSON representation of the generic section
func (s *GenericSection) MarshalJSON() ([]byte, error) {
	sectionJSON := &types.SectionJSON{
		Type:         s.Type(),
		TypeName:     s.TypeName(),
		Offset:       s.Offset(),
		Size:         s.Size(),
		CRCType:      s.CRCType().String(),
		IsEncrypted:  s.IsEncrypted(),
		IsDeviceData: s.IsDeviceData(),
		HasRawData:   true,  // Flag indicating binary file is needed for reconstruction
	}
	
	return json.Marshal(sectionJSON)
}