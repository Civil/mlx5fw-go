package types

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"reflect"
	
	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// ImageSignatureAnnotated represents the image signature structure using annotations
// Based on image_layout_image_signature from mstflint
type ImageSignatureAnnotated struct {
	SignatureType uint32 `offset:"byte:0,endian:be" json:"signature_type"`  // offset 0x0
	Signature     [256]uint8 `offset:"byte:4" json:"-"`                    // offset 0x4 - Handled separately in MarshalJSON/UnmarshalJSON
}

// ImageSignature2Annotated represents the extended image signature structure using annotations
// Based on image_layout_image_signature_2 from mstflint
type ImageSignature2Annotated struct {
	SignatureType uint32 `offset:"byte:0,endian:be" json:"signature_type"`  // offset 0x0
	Signature     [512]uint8 `offset:"byte:4" json:"-"`                    // offset 0x4 - Handled separately in MarshalJSON/UnmarshalJSON
}

// PublicKeyAnnotated represents a public key structure using annotations
// Based on image_layout_file_public_keys from mstflint
type PublicKeyAnnotated struct {
	Reserved uint32    `offset:"byte:0,endian:be,reserved:true" json:"reserved"` // offset 0x0
	UUID     [16]uint8 `offset:"byte:4" json:"-"`                              // offset 0x4 - Handled separately in MarshalJSON/UnmarshalJSON
	Key      [256]uint8 `offset:"byte:20" json:"-"`                            // offset 0x14 (20 decimal) - Handled separately in MarshalJSON/UnmarshalJSON
}

// PublicKeysAnnotated represents an array of public keys using annotations
// Based on image_layout_public_keys from mstflint
type PublicKeysAnnotated struct {
	Keys [8]PublicKeyAnnotated `offset:"byte:0" json:"keys"` // offset 0x0
}

// PublicKey2Annotated represents an extended public key structure using annotations
// Based on image_layout_file_public_keys_2 from mstflint
type PublicKey2Annotated struct {
	Reserved uint32    `offset:"byte:0,endian:be,reserved:true" json:"reserved"` // offset 0x0
	UUID     [16]uint8 `offset:"byte:4" json:"-"`                              // offset 0x4 - Handled separately in MarshalJSON/UnmarshalJSON
	Key      [512]uint8 `offset:"byte:20" json:"-"`                            // offset 0x14 (20 decimal) - Handled separately in MarshalJSON/UnmarshalJSON
}

// PublicKeys2Annotated represents an array of extended public keys using annotations
// Based on image_layout_public_keys_2 from mstflint
type PublicKeys2Annotated struct {
	Keys [8]PublicKey2Annotated `offset:"byte:0" json:"keys"` // offset 0x0
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

// MarshalJSON implements json.Marshaler interface for ImageSignatureAnnotated
func (s *ImageSignatureAnnotated) MarshalJSON() ([]byte, error) {
	type Alias ImageSignatureAnnotated
	return json.Marshal(&struct {
		*Alias
		Signature string `json:"signature"`
	}{
		Alias:     (*Alias)(s),
		Signature: base64.StdEncoding.EncodeToString(s.Signature[:]),
	})
}

// UnmarshalJSON implements json.Unmarshaler interface for ImageSignatureAnnotated
func (s *ImageSignatureAnnotated) UnmarshalJSON(data []byte) error {
	type Alias ImageSignatureAnnotated
	aux := &struct {
		*Alias
		Signature string `json:"signature"`
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	
	sigData, err := base64.StdEncoding.DecodeString(aux.Signature)
	if err != nil {
		return err
	}
	copy(s.Signature[:], sigData)
	
	return nil
}

// MarshalJSON implements json.Marshaler interface for ImageSignature2Annotated
func (s *ImageSignature2Annotated) MarshalJSON() ([]byte, error) {
	type Alias ImageSignature2Annotated
	return json.Marshal(&struct {
		*Alias
		Signature string `json:"signature"`
	}{
		Alias:     (*Alias)(s),
		Signature: base64.StdEncoding.EncodeToString(s.Signature[:]),
	})
}

// UnmarshalJSON implements json.Unmarshaler interface for ImageSignature2Annotated
func (s *ImageSignature2Annotated) UnmarshalJSON(data []byte) error {
	type Alias ImageSignature2Annotated
	aux := &struct {
		*Alias
		Signature string `json:"signature"`
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	
	sigData, err := base64.StdEncoding.DecodeString(aux.Signature)
	if err != nil {
		return err
	}
	copy(s.Signature[:], sigData)
	
	return nil
}

// MarshalJSON implements json.Marshaler interface for PublicKeyAnnotated
func (p *PublicKeyAnnotated) MarshalJSON() ([]byte, error) {
	type Alias PublicKeyAnnotated
	return json.Marshal(&struct {
		*Alias
		UUID string `json:"uuid"`
		Key  string `json:"key"`
	}{
		Alias: (*Alias)(p),
		UUID:  hex.EncodeToString(p.UUID[:]),
		Key:   base64.StdEncoding.EncodeToString(p.Key[:]),
	})
}

// UnmarshalJSON implements json.Unmarshaler interface for PublicKeyAnnotated
func (p *PublicKeyAnnotated) UnmarshalJSON(data []byte) error {
	type Alias PublicKeyAnnotated
	aux := &struct {
		*Alias
		UUID string `json:"uuid"`
		Key  string `json:"key"`
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	
	uuidData, err := hex.DecodeString(aux.UUID)
	if err != nil {
		return err
	}
	copy(p.UUID[:], uuidData)
	
	keyData, err := base64.StdEncoding.DecodeString(aux.Key)
	if err != nil {
		return err
	}
	copy(p.Key[:], keyData)
	
	return nil
}

// MarshalJSON implements json.Marshaler interface for PublicKey2Annotated
func (p *PublicKey2Annotated) MarshalJSON() ([]byte, error) {
	type Alias PublicKey2Annotated
	return json.Marshal(&struct {
		*Alias
		UUID string `json:"uuid"`
		Key  string `json:"key"`
	}{
		Alias: (*Alias)(p),
		UUID:  hex.EncodeToString(p.UUID[:]),
		Key:   base64.StdEncoding.EncodeToString(p.Key[:]),
	})
}

// UnmarshalJSON implements json.Unmarshaler interface for PublicKey2Annotated
func (p *PublicKey2Annotated) UnmarshalJSON(data []byte) error {
	type Alias PublicKey2Annotated
	aux := &struct {
		*Alias
		UUID string `json:"uuid"`
		Key  string `json:"key"`
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	
	uuidData, err := hex.DecodeString(aux.UUID)
	if err != nil {
		return err
	}
	copy(p.UUID[:], uuidData)
	
	keyData, err := base64.StdEncoding.DecodeString(aux.Key)
	if err != nil {
		return err
	}
	copy(p.Key[:], keyData)
	
	return nil
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

// Unmarshal unmarshals binary data into PublicKeyAnnotated
func (p *PublicKeyAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, p)
}

// Marshal marshals PublicKeyAnnotated into binary data
func (p *PublicKeyAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(p)
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

// Unmarshal unmarshals binary data into PublicKey2Annotated
func (p *PublicKey2Annotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, p)
}

// Marshal marshals PublicKey2Annotated into binary data
func (p *PublicKey2Annotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(p)
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

