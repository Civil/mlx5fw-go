package sections

import (
	"encoding/json"

	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// ResetInfoSection represents a RESET_INFO section
type ResetInfoSection struct {
	*interfaces.BaseSection
	ResetInfo *types.ResetInfo
}

// NewResetInfoSection creates a new ResetInfo section
func NewResetInfoSection(base *interfaces.BaseSection) *ResetInfoSection {
	return &ResetInfoSection{
		BaseSection: base,
	}
}

// Parse parses the RESET_INFO section data
func (s *ResetInfoSection) Parse(data []byte) error {
	s.SetRawData(data)

	s.ResetInfo = &types.ResetInfo{}
	if err := s.ResetInfo.Unmarshal(data); err != nil {
		return merry.Wrap(err)
	}

	return nil
}

// MarshalJSON returns JSON representation of the RESET_INFO section
func (s *ResetInfoSection) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type":         s.Type(),
		"type_name":    s.TypeName(),
		"offset":       s.Offset(),
		"size":         s.Size(),
		"has_raw_data": true, // RESET_INFO structure doesn't match firmware layout
	}

	if s.ResetInfo != nil {
		// Parse version vector from the first 52 bytes of the raw data
		vv := &types.VersionVector{}
		_ = vv.Unmarshal(s.GetRawData()[:52])
		versionVector := map[string]interface{}{
			"reset_capabilities": map[string]interface{}{
				"reset_ver_en":       vv.ResetCapabilities.ResetVerEn,
				"version_vector_ver": vv.ResetCapabilities.VersionVectorVer,
			},
			"scratchpad":  formatResetVersion(vv.Scratchpad),
			"icm_context": formatResetVersion(vv.ICMContext),
			"pci":         formatResetVersion(vv.PCI),
			"phy":         formatResetVersion(vv.PHY),
			"ini":         formatResetVersion(vv.INI),
		}
		result["reset_info"] = map[string]interface{}{
			"version_vector": versionVector,
		}
	}

	return json.Marshal(result)
}

func formatResetVersion(rv types.ResetVersion) map[string]interface{} {
	return map[string]interface{}{
		"major":  rv.Major,
		"branch": rv.Branch,
		"minor":  rv.Minor,
	}
}
