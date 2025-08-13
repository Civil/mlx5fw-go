package sections

import (
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// ForbiddenVersionsSection represents a FORBIDDEN_VERSIONS section
type ForbiddenVersionsSection struct {
	*interfaces.BaseSection
	ForbiddenVersions *types.ForbiddenVersions `json:"forbidden_versions,omitempty"`
}

// NewForbiddenVersionsSection creates a new ForbiddenVersions section
func NewForbiddenVersionsSection(base *interfaces.BaseSection) *ForbiddenVersionsSection {
	base.HasRawData = true // Default to true until successfully parsed
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

	s.HasRawData = false // Successfully parsed
	return nil
}
