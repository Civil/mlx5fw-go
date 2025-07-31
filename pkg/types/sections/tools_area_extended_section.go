package sections

import (
	"encoding/json"
	
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// ToolsAreaExtendedSection represents a TOOLS_AREA section with extended parsing
type ToolsAreaExtendedSection struct {
	*interfaces.BaseSection
	ToolsArea *types.ToolsAreaExtended
}

// NewToolsAreaExtendedSection creates a new ToolsAreaExtended section
func NewToolsAreaExtendedSection(base *interfaces.BaseSection) *ToolsAreaExtendedSection {
	return &ToolsAreaExtendedSection{
		BaseSection: base,
	}
}

// Parse parses the TOOLS_AREA section data
func (s *ToolsAreaExtendedSection) Parse(data []byte) error {
	s.SetRawData(data)
	
	if len(data) < types.ToolsAreaSize {
		return merry.Errorf("TOOLS_AREA section too small: expected at least %d bytes, got %d", types.ToolsAreaSize, len(data))
	}
	
	s.ToolsArea = &types.ToolsAreaExtended{}
	if err := s.ToolsArea.Unmarshal(data[:types.ToolsAreaSize]); err != nil {
		return merry.Wrap(err)
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the TOOLS_AREA section
func (s *ToolsAreaExtendedSection) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type":      s.Type(),
		"type_name": s.TypeName(),
		"offset":    s.Offset(),
		"size":      s.Size(),
		"has_raw_data": true, // TOOLS_AREA needs binary data for TypeData and Reserved fields
	}
	
	if s.ToolsArea != nil {
		result["tools_area"] = map[string]interface{}{
			"tlvrc":        s.ToolsArea.TLVRC,
			"crc_flag":     s.ToolsArea.CRCFlag,
			"total_length": s.ToolsArea.TotalLength,
			"type_length":  s.ToolsArea.TypeLength,
			// type_data and reserved are arrays, can be included if needed
		}
	}
	
	return json.Marshal(result)
}