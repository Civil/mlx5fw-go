package types

import (
	"reflect"
	
	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// HashesTableHeaderAnnotated represents the hashes table header structure using annotations
// This is specific to FS5 format
type HashesTableHeaderAnnotated struct {
	Magic          uint32 `offset:"byte:0,endian:be"`  // Should be a specific magic value
	Version        uint32 `offset:"byte:4,endian:be"`  // Version number
	Reserved1      uint32 `offset:"byte:8,endian:be,reserved:true"`  // Reserved field
	Reserved2      uint32 `offset:"byte:12,endian:be,reserved:true"` // Reserved field
	TableSize      uint32 `offset:"byte:16,endian:be"` // Size of the hashes table
	NumEntries     uint32 `offset:"byte:20,endian:be"` // Number of hash entries
	Reserved3      uint32 `offset:"byte:24,endian:be,reserved:true"` // Reserved field
	CRC            uint16 `offset:"byte:28,endian:be"` // CRC16 of the header
	Reserved4      uint16 `offset:"byte:30,endian:be,reserved:true"` // Reserved field
}

// HashTableEntryAnnotated represents a single hash entry in the hashes table using annotations
type HashTableEntryAnnotated struct {
	Type           uint32   `offset:"byte:0,endian:be"`   // Hash type/identifier
	Offset         uint32   `offset:"byte:4,endian:be"`   // Offset in the image
	Size           uint32   `offset:"byte:8,endian:be"`   // Size of the hashed region
	Reserved       uint32   `offset:"byte:12,endian:be,reserved:true"` // Reserved field
	Hash           [32]byte `offset:"byte:16"`            // SHA-256 hash value
}

// Unmarshal methods using annotations

// Unmarshal unmarshals binary data into HashesTableHeaderAnnotated
func (h *HashesTableHeaderAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (h *HashesTableHeaderAnnotated) UnmarshalWithReserved(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*h))
	if err != nil {
		return err
	}
	opts := &annotations.UnmarshalOptions{
		IncludeReserved: true,
	}
	return annotations.UnmarshalWithOptions(data, h, annot, opts)
}

// Marshal marshals HashesTableHeaderAnnotated into binary data
func (h *HashesTableHeaderAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// MarshalWithReserved marshals HashesTableHeaderAnnotated including reserved fields
func (h *HashesTableHeaderAnnotated) MarshalWithReserved() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*h))
	if err != nil {
		return nil, err
	}
	opts := &annotations.MarshalOptions{
		IncludeReserved: true,
	}
	return annotations.MarshalWithOptions(h, annot, opts)
}

// Unmarshal unmarshals binary data into HashTableEntryAnnotated
func (h *HashTableEntryAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (h *HashTableEntryAnnotated) UnmarshalWithReserved(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*h))
	if err != nil {
		return err
	}
	opts := &annotations.UnmarshalOptions{
		IncludeReserved: true,
	}
	return annotations.UnmarshalWithOptions(data, h, annot, opts)
}

// Marshal marshals HashTableEntryAnnotated into binary data
func (h *HashTableEntryAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// MarshalWithReserved marshals HashTableEntryAnnotated including reserved fields
func (h *HashTableEntryAnnotated) MarshalWithReserved() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*h))
	if err != nil {
		return nil, err
	}
	opts := &annotations.MarshalOptions{
		IncludeReserved: true,
	}
	return annotations.MarshalWithOptions(h, annot, opts)
}

