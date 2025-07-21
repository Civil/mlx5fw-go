package sections

import (
	"encoding/json"
	
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
	"github.com/ghostiam/binstruct"
)

// DeviceInfoSection represents a Device Info section
type DeviceInfoSection struct {
	*interfaces.BaseSection
	Info *types.DeviceInfo
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
	
	s.Info = &types.DeviceInfo{}
	
	if err := binstruct.UnmarshalBE(data, s.Info); err != nil {
		return merry.Wrap(err)
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the Device Info section
func (s *DeviceInfoSection) MarshalJSON() ([]byte, error) {
	if s.Info == nil {
		return s.BaseSection.MarshalJSON()
	}
	
	return json.Marshal(map[string]interface{}{
		"type":         s.Type(),
		"type_name":    s.TypeName(),
		"offset":       s.Offset(),
		"size":         s.Size(),
		"device_info": map[string]interface{}{
			"device_id": s.Info.DeviceID,
			"vendor_id": s.Info.VendorID,
			"subsystem_id": s.Info.SubsystemID,
			"subsystem_vendor_id": s.Info.SubsystemVendorID,
			"hw_version": s.Info.HWVersion,
			"hw_revision": s.Info.HWRevision,
			"capabilities": s.Info.Capabilities,
			"mac_guid": s.Info.MACGUID,
			"num_macs": s.Info.NumMACs,
		},
	})
}