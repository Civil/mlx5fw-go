package sections

import (
	"encoding/json"
	"strings"
	
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
	"github.com/ghostiam/binstruct"
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
	
	if err := binstruct.UnmarshalBE(data, s.Info); err != nil {
		return merry.Wrap(err)
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the MFG Info section
func (s *MFGInfoSection) MarshalJSON() ([]byte, error) {
	if s.Info == nil {
		return s.BaseSection.MarshalJSON()
	}
	
	// Clean null-terminated strings
	psid := strings.TrimRight(string(s.Info.PSID[:]), "\x00")
	partNumber := strings.TrimRight(string(s.Info.PartNumber[:]), "\x00")
	revision := strings.TrimRight(string(s.Info.Revision[:]), "\x00")
	productName := strings.TrimRight(string(s.Info.ProductName[:]), "\x00")
	
	return json.Marshal(map[string]interface{}{
		"type":         s.Type(),
		"type_name":    s.TypeName(),
		"offset":       s.Offset(),
		"size":         s.Size(),
		"mfg_info": map[string]interface{}{
			"psid": psid,
			"part_number": partNumber,
			"revision": revision,
			"product_name": productName,
		},
	})
}