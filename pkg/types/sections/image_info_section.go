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
			"product_ver_raw": s.Info.ProductVer,
			"security_mode": s.Info.GetSecurityMode(),
			"security_attributes": s.Info.GetSecurityAttributesString(),
			"pci_device_id": s.Info.PCIDeviceID,
			"pci_vendor_id": s.Info.PCIVendorID,
			"pci_subsystem_id": s.Info.PCISubsystemID,
			"pci_subvendor_id": s.Info.PCISubVendorID,
			"fw_ver_major": s.Info.FWVerMajor,
			"reserved2": s.Info.Reserved2,
			"fw_ver_minor": s.Info.FWVerMinor,
			"fw_ver_subminor": s.Info.FWVerSubminor,
			"mic_ver_major": s.Info.MICVerMajor,
			"reserved4": s.Info.Reserved4,
			"mic_ver_minor": s.Info.MICVerMinor,
			"mic_ver_subminor": s.Info.MICVerSubminor,
			"reserved3a": s.Info.Reserved3a,
			"hour": s.Info.Hour,
			"minutes": s.Info.Minutes,
			"seconds": s.Info.Seconds,
			"day": s.Info.Day,
			"month": s.Info.Month,
			"year": s.Info.Year,
			"security_and_version": s.Info.GetSecurityAndVersion(),
			"reserved5a": s.Info.Reserved5a,
			"vsd_vendor_id": s.Info.VSDVendorID,
			"image_size_data": s.Info.ImageSizeData,
			"reserved6": s.Info.Reserved6,
			"supported_hw_ids": s.Info.SupportedHWID,
			"ini_file_num": s.Info.INIFileNum,
			"reserved7": s.Info.Reserved7,
			"reserved8": s.Info.Reserved8,
			"module_versions": s.Info.ModuleVersions,
			"name": s.Info.GetNameString(),
		},
	})
}