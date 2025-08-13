package types

import (
	"encoding/hex"
	"encoding/json"

	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// HashesTableHeader represents the hashes table header structure using annotations
// This is specific to FS5 format
type HashesTableHeader struct {
	Magic      uint32 `offset:"byte:0,endian:be" json:"magic"`                    // Should be a specific magic value
	Version    uint32 `offset:"byte:4,endian:be" json:"version"`                  // Version number
	Reserved1  uint32 `offset:"byte:8,endian:be,reserved:true" json:"reserved1"`  // Reserved field
	Reserved2  uint32 `offset:"byte:12,endian:be,reserved:true" json:"reserved2"` // Reserved field
	TableSize  uint32 `offset:"byte:16,endian:be" json:"table_size"`              // Size of the hashes table
	NumEntries uint32 `offset:"byte:20,endian:be" json:"num_entries"`             // Number of hash entries
	Reserved3  uint32 `offset:"byte:24,endian:be,reserved:true" json:"reserved3"` // Reserved field
	CRC        uint16 `offset:"byte:28,endian:be" json:"crc"`                     // CRC16 of the header
	Reserved4  uint16 `offset:"byte:30,endian:be,reserved:true" json:"reserved4"` // Reserved field
}

// HashTableEntry represents a single hash entry in the hashes table using annotations
type HashTableEntry struct {
	Type     uint32   `offset:"byte:0,endian:be" json:"type"`                    // Hash type/identifier
	Offset   uint32   `offset:"byte:4,endian:be" json:"offset"`                  // Offset in the image
	Size     uint32   `offset:"byte:8,endian:be" json:"size"`                    // Size of the hashed region
	Reserved uint32   `offset:"byte:12,endian:be,reserved:true" json:"reserved"` // Reserved field
	Hash     [32]byte `offset:"byte:16" json:"-"`                                // SHA-256 hash value - handled separately
}

// Unmarshal methods using annotations

// Unmarshal unmarshals binary data into HashesTableHeader
func (h *HashesTableHeader) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (h *HashesTableHeader) UnmarshalWithReserved(data []byte) error {
	opts := &annotations.UnmarshalOptions{IncludeReserved: true}
	return annotations.UnmarshalWithOptionsStruct(data, h, opts)
}

// Marshal marshals HashesTableHeader into binary data
func (h *HashesTableHeader) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// MarshalWithReserved marshals HashesTableHeader including reserved fields
func (h *HashesTableHeader) MarshalWithReserved() ([]byte, error) {
	opts := &annotations.MarshalOptions{IncludeReserved: true}
	return annotations.MarshalWithOptionsStruct(h, opts)
}

// Unmarshal unmarshals binary data into HashTableEntry
func (h *HashTableEntry) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (h *HashTableEntry) UnmarshalWithReserved(data []byte) error {
	opts := &annotations.UnmarshalOptions{IncludeReserved: true}
	return annotations.UnmarshalWithOptionsStruct(data, h, opts)
}

// Marshal marshals HashTableEntry into binary data
func (h *HashTableEntry) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// MarshalWithReserved marshals HashTableEntry including reserved fields
func (h *HashTableEntry) MarshalWithReserved() ([]byte, error) {
	opts := &annotations.MarshalOptions{IncludeReserved: true}
	return annotations.MarshalWithOptionsStruct(h, opts)
}

// MarshalJSON implements json.Marshaler interface for HashTableEntry
func (h *HashTableEntry) MarshalJSON() ([]byte, error) {
	type Alias HashTableEntry
	return json.Marshal(&struct {
		*Alias
		Hash string `json:"hash"`
	}{
		Alias: (*Alias)(h),
		Hash:  hex.EncodeToString(h.Hash[:]),
	})
}

// UnmarshalJSON implements json.Unmarshaler interface for HashTableEntry
func (h *HashTableEntry) UnmarshalJSON(data []byte) error {
	type Alias HashTableEntry
	aux := &struct {
		*Alias
		Hash string `json:"hash"`
	}{
		Alias: (*Alias)(h),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Hash != "" {
		hashBytes, err := hex.DecodeString(aux.Hash)
		if err != nil {
			return err
		}
		if len(hashBytes) == 32 {
			copy(h.Hash[:], hashBytes)
		}
	}

	return nil
}
