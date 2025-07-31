package types

import (
	"reflect"
	
	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// ImageSignatureAnnotated represents the image signature structure using annotations
// Based on image_layout_image_signature from mstflint
type ImageSignatureAnnotated struct {
	SignatureType uint32 `offset:"byte:0,endian:be"`  // offset 0x0
	Signature     [256]uint8 `offset:"byte:4"`         // offset 0x4
}

// ImageSignature2Annotated represents the extended image signature structure using annotations
// Based on image_layout_image_signature_2 from mstflint
type ImageSignature2Annotated struct {
	SignatureType uint32 `offset:"byte:0,endian:be"`  // offset 0x0
	Signature     [512]uint8 `offset:"byte:4"`         // offset 0x4
}

// PublicKeyAnnotated represents a public key structure using annotations
// Based on image_layout_file_public_keys from mstflint
type PublicKeyAnnotated struct {
	Reserved uint32    `offset:"byte:0,endian:be,reserved:true"` // offset 0x0
	UUID     [16]uint8 `offset:"byte:4"`                         // offset 0x4
	Key      [256]uint8 `offset:"byte:20"`                       // offset 0x14 (20 decimal)
}

// PublicKeysAnnotated represents an array of public keys using annotations
// Based on image_layout_public_keys from mstflint
type PublicKeysAnnotated struct {
	Keys [8]PublicKeyAnnotated `offset:"byte:0"` // offset 0x0
}

// PublicKey2Annotated represents an extended public key structure using annotations
// Based on image_layout_file_public_keys_2 from mstflint
type PublicKey2Annotated struct {
	Reserved uint32    `offset:"byte:0,endian:be,reserved:true"` // offset 0x0
	UUID     [16]uint8 `offset:"byte:4"`                         // offset 0x4
	Key      [512]uint8 `offset:"byte:20"`                       // offset 0x14 (20 decimal)
}

// PublicKeys2Annotated represents an array of extended public keys using annotations
// Based on image_layout_public_keys_2 from mstflint
type PublicKeys2Annotated struct {
	Keys [8]PublicKey2Annotated `offset:"byte:0"` // offset 0x0
}

// ToolsAreaExtendedAnnotated represents the tools area structure using annotations
// Based on image_layout_tools_area from mstflint
type ToolsAreaExtendedAnnotated struct {
	TLVRC       uint32    `offset:"byte:0,endian:be"`              // offset 0x0
	CRCFlag     uint32    `offset:"byte:4,endian:be"`              // offset 0x4
	TotalLength uint32    `offset:"byte:8,endian:be"`              // offset 0x8
	TypeLength  uint32    `offset:"byte:12,endian:be"`             // offset 0xc
	TypeData    [16]uint8 `offset:"byte:16"`                       // offset 0x10
	Reserved    [32]uint8 `offset:"byte:32,reserved:true"`         // offset 0x20
}

// Unmarshal methods

// Unmarshal unmarshals binary data into ImageSignatureAnnotated
func (s *ImageSignatureAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, s)
}

// Marshal marshals ImageSignatureAnnotated into binary data
func (s *ImageSignatureAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(s)
}

// Unmarshal unmarshals binary data into ImageSignature2Annotated
func (s *ImageSignature2Annotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, s)
}

// Marshal marshals ImageSignature2Annotated into binary data
func (s *ImageSignature2Annotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(s)
}

// Unmarshal unmarshals binary data into PublicKeysAnnotated
func (p *PublicKeysAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, p)
}

// Marshal marshals PublicKeysAnnotated into binary data
func (p *PublicKeysAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(p)
}

// MarshalWithReserved marshals PublicKeysAnnotated including reserved fields
func (p *PublicKeysAnnotated) MarshalWithReserved() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*p))
	if err != nil {
		return nil, err
	}
	opts := &annotations.MarshalOptions{
		IncludeReserved: true,
	}
	return annotations.MarshalWithOptions(p, annot, opts)
}

// Unmarshal unmarshals binary data into PublicKeys2Annotated
func (p *PublicKeys2Annotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, p)
}

// Marshal marshals PublicKeys2Annotated into binary data
func (p *PublicKeys2Annotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(p)
}

// MarshalWithReserved marshals PublicKeys2Annotated including reserved fields
func (p *PublicKeys2Annotated) MarshalWithReserved() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*p))
	if err != nil {
		return nil, err
	}
	opts := &annotations.MarshalOptions{
		IncludeReserved: true,
	}
	return annotations.MarshalWithOptions(p, annot, opts)
}

// Unmarshal unmarshals binary data into ToolsAreaExtendedAnnotated
func (t *ToolsAreaExtendedAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, t)
}

// Marshal marshals ToolsAreaExtendedAnnotated into binary data
func (t *ToolsAreaExtendedAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(t)
}

// Conversion methods for compatibility

// ToAnnotated converts ImageSignature to ImageSignatureAnnotated
func (s *ImageSignature) ToAnnotated() *ImageSignatureAnnotated {
	return &ImageSignatureAnnotated{
		SignatureType: s.SignatureType,
		Signature:     s.Signature,
	}
}

// FromAnnotated converts ImageSignatureAnnotated to ImageSignature
func (s *ImageSignatureAnnotated) FromAnnotated() *ImageSignature {
	return &ImageSignature{
		SignatureType: s.SignatureType,
		Signature:     s.Signature,
	}
}

// ToAnnotated converts ImageSignature2 to ImageSignature2Annotated
func (s *ImageSignature2) ToAnnotated() *ImageSignature2Annotated {
	return &ImageSignature2Annotated{
		SignatureType: s.SignatureType,
		Signature:     s.Signature,
	}
}

// FromAnnotated converts ImageSignature2Annotated to ImageSignature2
func (s *ImageSignature2Annotated) FromAnnotated() *ImageSignature2 {
	return &ImageSignature2{
		SignatureType: s.SignatureType,
		Signature:     s.Signature,
	}
}

// ToAnnotated converts PublicKeys to PublicKeysAnnotated
func (p *PublicKeys) ToAnnotated() *PublicKeysAnnotated {
	annotated := &PublicKeysAnnotated{}
	for i := range p.Keys {
		annotated.Keys[i] = PublicKeyAnnotated{
			Reserved: p.Keys[i].Reserved,
			UUID:     p.Keys[i].UUID,
			Key:      p.Keys[i].Key,
		}
	}
	return annotated
}

// FromAnnotated converts PublicKeysAnnotated to PublicKeys
func (p *PublicKeysAnnotated) FromAnnotated() *PublicKeys {
	legacy := &PublicKeys{}
	for i := range p.Keys {
		legacy.Keys[i] = PublicKey{
			Reserved: p.Keys[i].Reserved,
			UUID:     p.Keys[i].UUID,
			Key:      p.Keys[i].Key,
		}
	}
	return legacy
}

// ToAnnotated converts PublicKeys2 to PublicKeys2Annotated
func (p *PublicKeys2) ToAnnotated() *PublicKeys2Annotated {
	annotated := &PublicKeys2Annotated{}
	for i := range p.Keys {
		annotated.Keys[i] = PublicKey2Annotated{
			Reserved: p.Keys[i].Reserved,
			UUID:     p.Keys[i].UUID,
			Key:      p.Keys[i].Key,
		}
	}
	return annotated
}

// FromAnnotated converts PublicKeys2Annotated to PublicKeys2
func (p *PublicKeys2Annotated) FromAnnotated() *PublicKeys2 {
	legacy := &PublicKeys2{}
	for i := range p.Keys {
		legacy.Keys[i] = PublicKey2{
			Reserved: p.Keys[i].Reserved,
			UUID:     p.Keys[i].UUID,
			Key:      p.Keys[i].Key,
		}
	}
	return legacy
}

// ToAnnotated converts ToolsAreaExtended to ToolsAreaExtendedAnnotated
func (t *ToolsAreaExtended) ToAnnotated() *ToolsAreaExtendedAnnotated {
	return &ToolsAreaExtendedAnnotated{
		TLVRC:       t.TLVRC,
		CRCFlag:     t.CRCFlag,
		TotalLength: t.TotalLength,
		TypeLength:  t.TypeLength,
		TypeData:    t.TypeData,
		Reserved:    t.Reserved,
	}
}

// FromAnnotated converts ToolsAreaExtendedAnnotated to ToolsAreaExtended
func (t *ToolsAreaExtendedAnnotated) FromAnnotated() *ToolsAreaExtended {
	return &ToolsAreaExtended{
		TLVRC:       t.TLVRC,
		CRCFlag:     t.CRCFlag,
		TotalLength: t.TotalLength,
		TypeLength:  t.TypeLength,
		TypeData:    t.TypeData,
		Reserved:    t.Reserved,
	}
}