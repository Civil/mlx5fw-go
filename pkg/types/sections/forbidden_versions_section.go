package sections

import (
	"encoding/json"
	
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// ForbiddenVersionsSection represents a FORBIDDEN_VERSIONS section
type ForbiddenVersionsSection struct {
	*interfaces.BaseSection
	ForbiddenVersions *types.ForbiddenVersions
}

// NewForbiddenVersionsSection creates a new ForbiddenVersions section
func NewForbiddenVersionsSection(base *interfaces.BaseSection) *ForbiddenVersionsSection {
	return &ForbiddenVersionsSection{
		BaseSection: base,
	}
}

// Parse parses the FORBIDDEN_VERSIONS section data
func (s *ForbiddenVersionsSection) Parse(data []byte) error {
	s.SetRawData(data)
	
	s.ForbiddenVersions = &types.ForbiddenVersions{}
	if err := s.ForbiddenVersions.Unmarshal(data); err != nil {
		return merry.Wrap(err)
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the FORBIDDEN_VERSIONS section
func (s *ForbiddenVersionsSection) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type":      s.Type(),
		"type_name": s.TypeName(),
		"offset":    s.Offset(),
		"size":      s.Size(),
	}
	
	if s.ForbiddenVersions != nil {
		// Calculate actual number of versions based on section size
		// Section size = 8 (header) + numVersions * 4
		numVersions := (s.Size() - 8) / 4
		
		// Include the actual number of versions based on section size
		// Make sure we don't go beyond the actual slice length
		actualVersions := uint32(len(s.ForbiddenVersions.Versions))
		copyCount := numVersions
		if copyCount > actualVersions {
			copyCount = actualVersions
		}
		if copyCount > 34 {
			copyCount = 34
		}
		
		versions := make([]uint32, copyCount)
		for i := uint32(0); i < copyCount; i++ {
			versions[i] = s.ForbiddenVersions.Versions[i]
		}
		
		result["forbidden_versions"] = map[string]interface{}{
			"count":     s.ForbiddenVersions.Count,
			"reserved":  s.ForbiddenVersions.Reserved,
			"versions":  versions,
		}
	}
	
	return json.Marshal(result)
}