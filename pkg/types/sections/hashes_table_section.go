package sections

import (
	"bytes"
	"encoding/binary"
	
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// HashesTableSection represents a Hashes Table section
type HashesTableSection struct {
	*interfaces.BaseSection
	Header       *types.HashesTableHeader `json:"header,omitempty"`
	Entries      []*types.HashTableEntry  `json:"entries,omitempty"`
	ReservedTail types.FWByteSlice        `json:"reserved_tail,omitempty"`  // Reserved data at the end of the section after entries
}

// NewHashesTableSection creates a new Hashes Table section
func NewHashesTableSection(base *interfaces.BaseSection) *HashesTableSection {
	base.HasRawData = true // Default to true until successfully parsed
	return &HashesTableSection{
		BaseSection: base,
	}
}

// Parse parses the Hashes Table section data
func (s *HashesTableSection) Parse(data []byte) error {
	s.SetRawData(data)
	
	if len(data) < 32 { // Minimum size for header
		return merry.New("Hashes table section too small")
	}
	
	// Parse header
	s.Header = &types.HashesTableHeader{}
	reader := bytes.NewReader(data)
	if err := binary.Read(reader, binary.BigEndian, s.Header); err != nil {
		return merry.Wrap(err)
	}
	
	// Parse entries if header indicates there are any
	offset := 32 // After header
	if s.Header.NumEntries > 0 && s.Header.NumEntries < 1000 { // Sanity check
		entrySize := 64 // Typical entry size (32-byte hash + metadata)
		
		for i := uint32(0); i < s.Header.NumEntries && offset+entrySize <= len(data); i++ {
			entry := &types.HashTableEntry{}
			entryData := data[offset:offset+entrySize]
			
			// Parse entry data
			entryReader := bytes.NewReader(entryData)
			if err := binary.Read(entryReader, binary.BigEndian, entry); err != nil {
				// If binary read fails, just store raw data
				if len(entryData) >= 32 {
					copy(entry.Hash[:], entryData[:32])
				}
			}
			
			s.Entries = append(s.Entries, entry)
			offset += entrySize
		}
	}
	
	// Store any remaining data after entries (or all data after header if no entries)
	if offset < len(data) {
		s.ReservedTail = types.FWByteSlice(data[offset:])
	}
	
	s.HasRawData = false // Successfully parsed
	return nil
}

