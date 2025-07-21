package sections

import (
	"encoding/json"
	
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
	"github.com/ghostiam/binstruct"
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
	
	if err := binstruct.UnmarshalBE(data, s.Info); err != nil {
		return merry.Wrap(err)
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
	if s.Info == nil {
		return s.BaseSection.MarshalJSON()
	}
	
	return json.Marshal(map[string]interface{}{
		"type":         s.Type(),
		"type_name":    s.TypeName(),
		"offset":       s.Offset(),
		"size":         s.Size(),
		"image_info": map[string]interface{}{
			"fw_version": s.Info.GetFWVersionString(),
			"fw_release_date": s.Info.GetFWReleaseDateString(),
			"mic_version": s.Info.GetMICVersionString(),
			"prs_name": s.Info.GetPRSNameString(),
			"psid": s.Info.GetPSIDString(),
			"description": s.Info.GetDescriptionString(),
			"vsd": s.Info.GetVSDString(),
			"product_ver": s.Info.GetProductVerString(),
			"security_mode": s.Info.GetSecurityMode(),
			"security_attributes": s.Info.GetSecurityAttributesString(),
		},
	})
}