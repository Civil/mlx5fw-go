package sections

import (
	"encoding/json"
	
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// DeviceInfoSection represents a Device Info section
type DeviceInfoSection struct {
	*interfaces.BaseSection
	Info *types.DevInfo
}

// NewDeviceInfoSection creates a new Device Info section
func NewDeviceInfoSection(base *interfaces.BaseSection) *DeviceInfoSection {
	return &DeviceInfoSection{
		BaseSection: base,
	}
}

// Parse parses the Device Info section data
func (s *DeviceInfoSection) Parse(data []byte) error {
	s.SetRawData(data)
	
	s.Info = &types.DevInfo{}
	
	if err := s.Info.Unmarshal(data); err != nil {
		return merry.Wrap(err)
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the Device Info section
func (s *DeviceInfoSection) MarshalJSON() ([]byte, error) {
	if s.Info == nil {
		return s.BaseSection.MarshalJSON()
	}
	
	result := map[string]interface{}{
		"type":         s.Type(),
		"type_name":    s.TypeName(),
		"offset":       s.Offset(),
		"size":         s.Size(),
	}
	
	if s.Info != nil {
		result["device_info"] = map[string]interface{}{
			"signature0": s.Info.Signature0,
			"signature1": s.Info.Signature1,
			"signature2": s.Info.Signature2,
			"signature3": s.Info.Signature3,
			"minor_version": s.Info.MinorVersion,
			"major_version": s.Info.MajorVersion,
			"reserved1": s.Info.Reserved1,
			"reserved2": s.Info.Reserved2,
			"guids": map[string]interface{}{
				"reserved1": s.Info.Guids.Reserved1,
				"step": s.Info.Guids.Step,
				"num_allocated": s.Info.Guids.NumAllocated,
				"reserved2": s.Info.Guids.Reserved2,
				"uid": s.Info.Guids.UID,
			},
			"macs": map[string]interface{}{
				"reserved1": s.Info.Macs.Reserved1,
				"step": s.Info.Macs.Step,
				"num_allocated": s.Info.Macs.NumAllocated,
				"reserved2": s.Info.Macs.Reserved2,
				"uid": s.Info.Macs.UID,
			},
			"reserved3": s.Info.Reserved3,
			"reserved4": s.Info.Reserved4,
			"original_crc": s.Info.CRC,
		}
	}
	
	return json.Marshal(result)
}