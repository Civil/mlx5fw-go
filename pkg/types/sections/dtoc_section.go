package sections

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// DTOCSection represents a DTOC (Device Table of Contents) section
type DTOCSection struct {
	*interfaces.BaseSection
	Header  *types.ITOCHeader // DTOC uses same header structure as ITOC
	Entries []*types.ITOCEntry
	crcCalc *parser.CRCCalculator
}

// NewDTOCSection creates a new DTOC section
func NewDTOCSection(base *interfaces.BaseSection) *DTOCSection {
	return &DTOCSection{
		BaseSection: base,
		crcCalc:     parser.NewCRCCalculator(),
	}
}

// Parse parses the DTOC section data
func (s *DTOCSection) Parse(data []byte) error {
	s.SetRawData(data)
	
	// Parse header
	s.Header = &types.ITOCHeader{}
	reader := bytes.NewReader(data)
	if err := binary.Read(reader, binary.BigEndian, s.Header); err != nil {
		return merry.Wrap(err)
	}
	
	// Validate signature
	if s.Header.Signature0 != types.DTOCSignature {
		return merry.Errorf("invalid DTOC signature: 0x%08x", s.Header.Signature0)
	}
	
	// Parse entries - start after header (32 bytes)
	offset := 32
	for offset+32 <= len(data) {
		entry := &types.ITOCEntry{}
		if err := entry.Unmarshal(data[offset:offset+32]); err != nil {
			return err
		}
		
		// Check for end marker
		if entry.Type == types.SectionTypeEnd {
			break
		}
		
		s.Entries = append(s.Entries, entry)
		offset += 32
	}
	
	return nil
}

// CalculateCRC calculates the CRC for the DTOC section
func (s *DTOCSection) CalculateCRC() (uint32, error) {
	if s.GetRawData() == nil {
		return 0, merry.New("no data to calculate CRC")
	}
	
	// DTOC CRC is calculated the same way as ITOC
	data := s.GetRawData()
	if len(data) < 32 {
		return 0, merry.New("DTOC data too small")
	}
	
	// Create a copy with CRC field zeroed
	crcData := make([]byte, len(data))
	copy(crcData, data)
	// Zero out CRC field at offset 0x1c
	binary.BigEndian.PutUint32(crcData[0x1c:], 0)
	
	// Calculate CRC16
	crc := s.crcCalc.CalculateSoftwareCRC16(crcData)
	return uint32(crc), nil
}

// VerifyCRC verifies the DTOC section's CRC
func (s *DTOCSection) VerifyCRC() error {
	expectedCRC := s.Header.CRC & 0xffff
	calculatedCRC, err := s.CalculateCRC()
	if err != nil {
		return err
	}
	
	if uint32(expectedCRC) != calculatedCRC&0xffff {
		return merry.Errorf("CRC mismatch: expected 0x%04x, got 0x%04x", 
			expectedCRC, calculatedCRC&0xffff)
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the DTOC section
func (s *DTOCSection) MarshalJSON() ([]byte, error) {
	entries := make([]map[string]interface{}, len(s.Entries))
	for i, entry := range s.Entries {
		entries[i] = map[string]interface{}{
			"type":        entry.Type,
			"type_name":   types.GetSectionTypeName(uint16(entry.Type)),
			"size":        entry.GetSize(),
			"flash_addr":  entry.GetFlashAddr(),
			"crc":         entry.GetCRC(),
			"no_crc":      entry.GetNoCRC(),
			"encrypted":   entry.Encrypted,
		}
	}
	
	return json.Marshal(map[string]interface{}{
		"type":         s.Type(),
		"type_name":    s.TypeName(),
		"offset":       s.Offset(),
		"size":         s.Size(),
		"header": map[string]interface{}{
			"signature":     s.Header.Signature0,
			"version":       s.Header.Version,
			"itoc_entry_crc": s.Header.ITOCEntryCRC,
			"crc":           s.Header.CRC,
		},
		"entries": entries,
	})
}