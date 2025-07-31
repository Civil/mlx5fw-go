package sections

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/Civil/mlx5fw-go/pkg/interfaces"
)

// Boot2Section represents the BOOT2 section
type Boot2Section struct {
	*interfaces.BaseSection
	
	// BOOT2 specific fields
	Magic      uint32 `json:"magic"`       // Offset 0x00
	SizeDwords uint32 `json:"size_dwords"` // Offset 0x04 - size in dwords
	Reserved   uint64 `json:"reserved"`    // Offset 0x08
	
	// The actual boot2 code data (without header and CRC)
	CodeData []byte `json:"-"` // Binary data not included in JSON
}

// NewBoot2Section creates a new BOOT2 section from base
func NewBoot2Section(base *interfaces.BaseSection) *Boot2Section {
	return &Boot2Section{
		BaseSection: base,
	}
}

// Parse parses the raw BOOT2 data
func (s *Boot2Section) Parse(data []byte) error {
	if len(data) < 16 {
		return fmt.Errorf("boot2 data too small: %d bytes", len(data))
	}
	
	// Parse header
	s.Magic = binary.BigEndian.Uint32(data[0:4])
	s.SizeDwords = binary.BigEndian.Uint32(data[4:8])
	s.Reserved = binary.BigEndian.Uint64(data[8:16])
	
	// Store raw data for now (will be needed for CRC calculation)
	s.SetRawData(data)
	
	// Calculate expected data size
	expectedSize := (s.SizeDwords + 4) * 4
	if uint32(len(data)) < expectedSize {
		return fmt.Errorf("boot2 data size mismatch: got %d, expected %d", len(data), expectedSize)
	}
	
	// Extract code data (between header and CRC)
	// The CRC is at position (size + 3) dwords from start
	crcOffset := (s.SizeDwords + 3) * 4
	if crcOffset >= uint32(len(data)) {
		// If CRC is outside data bounds, use all data after header
		s.CodeData = data[16:]
	} else {
		// Extract code between header and CRC
		s.CodeData = data[16:crcOffset]
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the BOOT2 section
func (s *Boot2Section) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":         s.Type(),
		"type_name":    s.TypeName(),
		"offset":       s.Offset(),
		"size":         s.Size(),
		"crc_type":     s.CRCType(),
		"is_encrypted": s.IsEncrypted(),
		"magic":        s.Magic,
		"size_dwords":  s.SizeDwords,
		"reserved":     s.Reserved,
		"code_size":    len(s.CodeData),
		"has_raw_data": true, // BOOT2 always needs binary file
	})
}

// GetReconstruction returns the reconstruction info for reassembly
func (s *Boot2Section) GetReconstruction() interface{} {
	return map[string]interface{}{
		"magic":        s.Magic,
		"size_dwords":  s.SizeDwords,
		"reserved":     s.Reserved,
		"has_raw_data": true, // BOOT2 still needs binary file for code data
	}
}

