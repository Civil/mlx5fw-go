package sections

import (
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// MFGInfoSection represents a Manufacturing Info section
type MFGInfoSection struct {
	*interfaces.BaseSection
	MfgInfo *types.MFGInfo `json:"mfg_info,omitempty"`
}

// NewMFGInfoSection creates a new MFG Info section
func NewMFGInfoSection(base *interfaces.BaseSection) *MFGInfoSection {
	base.HasRawData = true // Default to true until successfully parsed
	return &MFGInfoSection{
		BaseSection: base,
	}
}

// Parse parses the MFG Info section data
func (s *MFGInfoSection) Parse(data []byte) error {
	s.SetRawData(data)
	
	s.MfgInfo = &types.MFGInfo{}
	
	if err := s.MfgInfo.Unmarshal(data); err != nil {
		return merry.Wrap(err)
	}
	
	s.HasRawData = false // Successfully parsed
	return nil
}

