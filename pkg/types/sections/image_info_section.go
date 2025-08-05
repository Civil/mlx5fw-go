package sections

import (
	"encoding/json"
	
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// ImageInfoSection represents an Image Info section
type ImageInfoSection struct {
	*interfaces.BaseSection
	Info    *types.ImageInfo
	crcCalc *parser.CRCCalculator
}

// NewImageInfoSection creates a new Image Info section
func NewImageInfoSection(base *interfaces.BaseSection) *ImageInfoSection {
	return &ImageInfoSection{
		BaseSection: base,
		crcCalc:     parser.NewCRCCalculator(),
	}
}

// Parse parses the Image Info section data
func (s *ImageInfoSection) Parse(data []byte) error {
	s.SetRawData(data)
	
	// Check minimum size requirement
	if len(data) < 1024 {
		return merry.Errorf("ImageInfo section too small: got %d bytes, minimum 1024 required", len(data))
	}
	
	s.Info = &types.ImageInfo{}
	
	if err := s.Info.Unmarshal(data); err != nil {
		return merry.Wrap(err)
	}
	
	// Workaround for VSD vendor ID alignment issue
	// The VSD vendor ID appears to be at offset 0x36 instead of 0x34
	if len(data) >= 0x38 {
		vsdVendorBytes := data[0x36:0x38]
		actualVendor := uint16(vsdVendorBytes[0]) | uint16(vsdVendorBytes[1])<<8
		if actualVendor != 0 && s.Info.VSDVendorID == 0 {
			s.Info.VSDVendorID = actualVendor
		}
	}
	
	return nil
}

// CalculateCRC calculates the CRC for the Image Info section
func (s *ImageInfoSection) CalculateCRC() (uint32, error) {
	// Image Info sections typically have CRC in ITOC entry
	if s.CRCType() == types.CRCInSection {
		// CRC is last 4 bytes of section
		data := s.GetRawData()
		if len(data) < 4 {
			return 0, merry.New("section too small for CRC")
		}
		
		// Calculate CRC over data excluding last 4 bytes
		crcData := data[:len(data)-4]
		crc := s.crcCalc.CalculateSoftwareCRC16(crcData)
		return uint32(crc), nil
	}
	
	return 0, nil
}

// MarshalJSON returns JSON representation of the Image Info section
func (s *ImageInfoSection) MarshalJSON() ([]byte, error) {
	// Create a wrapper struct with section metadata and embedded ImageInfo
	type SectionWithImageInfo struct {
		Type         uint16           `json:"type"`
		TypeName     string           `json:"type_name"`
		Offset       uint64           `json:"offset"`
		Size         uint32           `json:"size"`
		CRCType      string           `json:"crc_type"`
		IsEncrypted  bool             `json:"is_encrypted"`
		IsDeviceData bool             `json:"is_device_data"`
		HasRawData   bool             `json:"has_raw_data"`
		ImageInfo    *types.ImageInfo `json:"image_info,omitempty"`
	}
	
	section := &SectionWithImageInfo{
		Type:         s.Type(),
		TypeName:     s.TypeName(),
		Offset:       s.Offset(),
		Size:         s.Size(),
		CRCType:      s.CRCType().String(),
		IsEncrypted:  s.IsEncrypted(),
		IsDeviceData: s.IsDeviceData(),
		HasRawData:   s.Info == nil,
		ImageInfo:    s.Info,
	}
	
	return json.Marshal(section)
}