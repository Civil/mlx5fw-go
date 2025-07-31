package types

import (
	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// SignatureBlockAnnotated represents a firmware signature block using annotations
// Used for IMAGE_SIGNATURE_256 and IMAGE_SIGNATURE_512 sections
type SignatureBlockAnnotated struct {
	SignatureType   uint32 `offset:"byte:0,endian:be"`             // Type of signature (RSA-2048, RSA-4096, etc.)
	SignatureLength uint32 `offset:"byte:4,endian:be"`             // Length of signature data
	Reserved1       uint32 `offset:"byte:8,endian:be,reserved:true"`  // Reserved field
	Reserved2       uint32 `offset:"byte:12,endian:be,reserved:true"` // Reserved field
	// Signature data follows after the header (variable length)
}

// PublicKeyBlockAnnotated represents a public key block using annotations
// Used for PUBLIC_KEYS_2048 and PUBLIC_KEYS_4096 sections
type PublicKeyBlockAnnotated struct {
	KeyType    uint32 `offset:"byte:0,endian:be"`              // Type of key (RSA-2048, RSA-4096, etc.)
	KeyLength  uint32 `offset:"byte:4,endian:be"`              // Length of key data
	KeyID      uint32 `offset:"byte:8,endian:be"`              // Key identifier
	Reserved   uint32 `offset:"byte:12,endian:be,reserved:true"` // Reserved field
	// Public key data follows after the header (variable length)
}

// ForbiddenVersionsHeaderAnnotated represents the header for forbidden versions section using annotations
type ForbiddenVersionsHeaderAnnotated struct {
	Magic      uint32 `offset:"byte:0,endian:be"`              // Magic value
	Version    uint32 `offset:"byte:4,endian:be"`              // Format version
	NumEntries uint32 `offset:"byte:8,endian:be"`              // Number of forbidden version entries
	Reserved   uint32 `offset:"byte:12,endian:be,reserved:true"` // Reserved field
}

// ForbiddenVersionEntryAnnotated represents a single forbidden version entry using annotations
type ForbiddenVersionEntryAnnotated struct {
	MinVersion uint32 `offset:"byte:0,endian:be"`              // Minimum forbidden version
	MaxVersion uint32 `offset:"byte:4,endian:be"`              // Maximum forbidden version
	Flags      uint32 `offset:"byte:8,endian:be"`              // Flags for this entry
	Reserved   uint32 `offset:"byte:12,endian:be,reserved:true"` // Reserved field
}

// Unmarshal methods

// Unmarshal unmarshals binary data into SignatureBlockAnnotated
func (s *SignatureBlockAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, s)
}

// Marshal marshals SignatureBlockAnnotated into binary data
func (s *SignatureBlockAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(s)
}

// Unmarshal unmarshals binary data into PublicKeyBlockAnnotated
func (p *PublicKeyBlockAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, p)
}

// Marshal marshals PublicKeyBlockAnnotated into binary data
func (p *PublicKeyBlockAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(p)
}

// Unmarshal unmarshals binary data into ForbiddenVersionsHeaderAnnotated
func (f *ForbiddenVersionsHeaderAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, f)
}

// Marshal marshals ForbiddenVersionsHeaderAnnotated into binary data
func (f *ForbiddenVersionsHeaderAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(f)
}

// Unmarshal unmarshals binary data into ForbiddenVersionEntryAnnotated
func (f *ForbiddenVersionEntryAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, f)
}

// Marshal marshals ForbiddenVersionEntryAnnotated into binary data
func (f *ForbiddenVersionEntryAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(f)
}

// Conversion methods for compatibility

// ToAnnotated converts SignatureBlock to SignatureBlockAnnotated
func (s *SignatureBlock) ToAnnotated() *SignatureBlockAnnotated {
	return &SignatureBlockAnnotated{
		SignatureType:   s.SignatureType,
		SignatureLength: s.SignatureLength,
		Reserved1:       s.Reserved1,
		Reserved2:       s.Reserved2,
	}
}

// FromAnnotated converts SignatureBlockAnnotated to SignatureBlock
func (s *SignatureBlockAnnotated) FromAnnotated(signature []byte) *SignatureBlock {
	return &SignatureBlock{
		SignatureType:   s.SignatureType,
		SignatureLength: s.SignatureLength,
		Reserved1:       s.Reserved1,
		Reserved2:       s.Reserved2,
		Signature:       signature,
	}
}

// ToAnnotated converts PublicKeyBlock to PublicKeyBlockAnnotated
func (p *PublicKeyBlock) ToAnnotated() *PublicKeyBlockAnnotated {
	return &PublicKeyBlockAnnotated{
		KeyType:   p.KeyType,
		KeyLength: p.KeyLength,
		KeyID:     p.KeyID,
		Reserved:  p.Reserved,
	}
}

// FromAnnotated converts PublicKeyBlockAnnotated to PublicKeyBlock
func (p *PublicKeyBlockAnnotated) FromAnnotated(publicKey []byte) *PublicKeyBlock {
	return &PublicKeyBlock{
		KeyType:   p.KeyType,
		KeyLength: p.KeyLength,
		KeyID:     p.KeyID,
		Reserved:  p.Reserved,
		PublicKey: publicKey,
	}
}

// ToAnnotated converts ForbiddenVersionsHeader to ForbiddenVersionsHeaderAnnotated
func (f *ForbiddenVersionsHeader) ToAnnotated() *ForbiddenVersionsHeaderAnnotated {
	return &ForbiddenVersionsHeaderAnnotated{
		Magic:      f.Magic,
		Version:    f.Version,
		NumEntries: f.NumEntries,
		Reserved:   f.Reserved,
	}
}

// FromAnnotated converts ForbiddenVersionsHeaderAnnotated to ForbiddenVersionsHeader
func (f *ForbiddenVersionsHeaderAnnotated) FromAnnotated() *ForbiddenVersionsHeader {
	return &ForbiddenVersionsHeader{
		Magic:      f.Magic,
		Version:    f.Version,
		NumEntries: f.NumEntries,
		Reserved:   f.Reserved,
	}
}

// ToAnnotated converts ForbiddenVersionEntry to ForbiddenVersionEntryAnnotated
func (f *ForbiddenVersionEntry) ToAnnotated() *ForbiddenVersionEntryAnnotated {
	return &ForbiddenVersionEntryAnnotated{
		MinVersion: f.MinVersion,
		MaxVersion: f.MaxVersion,
		Flags:      f.Flags,
		Reserved:   f.Reserved,
	}
}

// FromAnnotated converts ForbiddenVersionEntryAnnotated to ForbiddenVersionEntry
func (f *ForbiddenVersionEntryAnnotated) FromAnnotated() *ForbiddenVersionEntry {
	return &ForbiddenVersionEntry{
		MinVersion: f.MinVersion,
		MaxVersion: f.MaxVersion,
		Flags:      f.Flags,
		Reserved:   f.Reserved,
	}
}