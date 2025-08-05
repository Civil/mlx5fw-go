package sections

import (
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// HWPointerSection represents a Hardware Pointer section
type HWPointerSection struct {
	*interfaces.BaseSection
	FS4Pointers *types.FS4HWPointers `json:"fs4_pointers,omitempty"`
	FS5Pointers *types.FS5HWPointers `json:"fs5_pointers,omitempty"`
	Format      types.FirmwareFormat `json:"format,omitempty"`
}

// NewHWPointerSection creates a new HW Pointer section
func NewHWPointerSection(base *interfaces.BaseSection) *HWPointerSection {
	base.HasRawData = true // Default to true until successfully parsed
	return &HWPointerSection{
		BaseSection: base,
	}
}

// Parse parses the HW Pointer section data
func (s *HWPointerSection) Parse(data []byte) error {
	s.SetRawData(data)
	
	if len(data) < 4 {
		return merry.New("HW pointer section too small")
	}
	
	// Determine format based on data size and content
	// FS5 has larger pointer structure
	if len(data) >= 0x100 { // FS5 size threshold
		s.Format = types.FormatFS5
		// Use annotated version for parsing
		annotated := &types.FS5HWPointersAnnotated{}
		if err := annotated.Unmarshal(data); err != nil {
			return merry.Wrap(err)
		}
		// Use annotated format directly
		s.FS5Pointers = annotated
	} else {
		s.Format = types.FormatFS4
		// Use annotated version for parsing
		annotated := &types.FS4HWPointersAnnotated{}
		if err := annotated.Unmarshal(data); err != nil {
			return merry.Wrap(err)
		}
		// Use annotated format directly
		s.FS4Pointers = annotated
	}
	
	s.HasRawData = false // Successfully parsed
	return nil
}

