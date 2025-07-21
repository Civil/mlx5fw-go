package sections

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"encoding/hex"
	
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// HashesTableSection represents a Hashes Table section
type HashesTableSection struct {
	*interfaces.BaseSection
	Header  *types.HashesTableHeader
	Entries []*types.HashTableEntry
}

// NewHashesTableSection creates a new Hashes Table section
func NewHashesTableSection(base *interfaces.BaseSection) *HashesTableSection {
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
	if s.Header.NumEntries > 0 && s.Header.NumEntries < 1000 { // Sanity check
		offset := 32 // After header
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
	
	return nil
}

// MarshalJSON returns JSON representation of the Hashes Table section
func (s *HashesTableSection) MarshalJSON() ([]byte, error) {
	entries := make([]map[string]interface{}, len(s.Entries))
	for i, entry := range s.Entries {
		entries[i] = map[string]interface{}{
			"index":     i,
			"type":      entry.Type,
			"offset":    entry.Offset,
			"size":      entry.Size,
			"hash":      hex.EncodeToString(entry.Hash[:]),
			"reserved":  entry.Reserved,
		}
	}
	
	result := map[string]interface{}{
		"type":         s.Type(),
		"type_name":    s.TypeName(),
		"offset":       s.Offset(),
		"size":         s.Size(),
	}
	
	if s.Header != nil {
		result["header"] = map[string]interface{}{
			"magic":       s.Header.Magic,
			"version":     s.Header.Version,
			"table_size":  s.Header.TableSize,
			"num_entries": s.Header.NumEntries,
			"crc":         s.Header.CRC,
		}
	}
	
	if len(entries) > 0 {
		result["entries"] = entries
	}
	
	return json.Marshal(result)
}