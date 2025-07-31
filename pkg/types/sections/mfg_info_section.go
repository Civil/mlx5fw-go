package sections

import (
	"encoding/json"
	
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// MFGInfoSection represents a Manufacturing Info section
type MFGInfoSection struct {
	*interfaces.BaseSection
	Info *types.MFGInfo
}

// NewMFGInfoSection creates a new MFG Info section
func NewMFGInfoSection(base *interfaces.BaseSection) *MFGInfoSection {
	return &MFGInfoSection{
		BaseSection: base,
	}
}

// Parse parses the MFG Info section data
func (s *MFGInfoSection) Parse(data []byte) error {
	s.SetRawData(data)
	
	s.Info = &types.MFGInfo{}
	
	if err := s.Info.Unmarshal(data); err != nil {
		return merry.Wrap(err)
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the MFG Info section
func (s *MFGInfoSection) MarshalJSON() ([]byte, error) {
	if s.Info == nil {
		return s.BaseSection.MarshalJSON()
	}
	
	return json.Marshal(map[string]interface{}{
		"type":         s.Type(),
		"type_name":    s.TypeName(),
		"offset":       s.Offset(),
		"size":         s.Size(),
		"mfg_info": map[string]interface{}{
			"psid": s.Info.PSID,
			"part_number": s.Info.PartNumber,
			"revision": s.Info.Revision,
			"product_name": s.Info.ProductName,
			"reserved": s.Info.Reserved,
		},
	})
}