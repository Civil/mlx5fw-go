package types

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"

	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// ImageSignature represents the image signature structure using annotations
// Based on image_layout_image_signature from mstflint
type ImageSignature struct {
	SignatureType uint32     `offset:"byte:0,endian:be" json:"signature_type"` // offset 0x0
	Signature     [256]uint8 `offset:"byte:4" json:"-"`                        // offset 0x4 - Handled separately in MarshalJSON/UnmarshalJSON
}

// ImageSignature2 represents the extended image signature structure using annotations
// Based on image_layout_image_signature_2 from mstflint
type ImageSignature2 struct {
	SignatureType uint32     `offset:"byte:0,endian:be" json:"signature_type"` // offset 0x0
	Signature     [512]uint8 `offset:"byte:4" json:"-"`                        // offset 0x4 - Handled separately in MarshalJSON/UnmarshalJSON
}

// PublicKey represents a public key structure using annotations
// Based on image_layout_file_public_keys from mstflint
type PublicKey struct {
	Reserved uint32     `offset:"byte:0,endian:be,reserved:true" json:"reserved"` // offset 0x0
	UUID     [16]uint8  `offset:"byte:4" json:"-"`                                // offset 0x4 - Handled separately in MarshalJSON/UnmarshalJSON
	Key      [256]uint8 `offset:"byte:20" json:"-"`                               // offset 0x14 (20 decimal) - Handled separately in MarshalJSON/UnmarshalJSON
}

// PublicKeys represents an array of public keys using annotations
// Based on image_layout_public_keys from mstflint
type PublicKeys struct {
	Keys [8]PublicKey `offset:"byte:0" json:"keys"` // offset 0x0
}

// PublicKey2 represents an extended public key structure using annotations
// Based on image_layout_file_public_keys_2 from mstflint
type PublicKey2 struct {
	Reserved uint32     `offset:"byte:0,endian:be,reserved:true" json:"reserved"` // offset 0x0
	UUID     [16]uint8  `offset:"byte:4" json:"-"`                                // offset 0x4 - Handled separately in MarshalJSON/UnmarshalJSON
	Key      [512]uint8 `offset:"byte:20" json:"-"`                               // offset 0x14 (20 decimal) - Handled separately in MarshalJSON/UnmarshalJSON
}

// PublicKeys2 represents an array of extended public keys using annotations
// Based on image_layout_public_keys_2 from mstflint
type PublicKeys2 struct {
	Keys [8]PublicKey2 `offset:"byte:0" json:"keys"` // offset 0x0
}

// ToolsAreaExtended represents the tools area structure using annotations
// Based on image_layout_tools_area from mstflint
type ToolsAreaExtended struct {
	TLVRC       uint32    `offset:"byte:0,endian:be"`      // offset 0x0
	CRCFlag     uint32    `offset:"byte:4,endian:be"`      // offset 0x4
	TotalLength uint32    `offset:"byte:8,endian:be"`      // offset 0x8
	TypeLength  uint32    `offset:"byte:12,endian:be"`     // offset 0xc
	TypeData    [16]uint8 `offset:"byte:16"`               // offset 0x10
	Reserved    [32]uint8 `offset:"byte:32,reserved:true"` // offset 0x20
}

// MarshalJSON implements json.Marshaler interface for ImageSignature
func (s *ImageSignature) MarshalJSON() ([]byte, error) {
	type Alias ImageSignature
	return json.Marshal(&struct {
		*Alias
		Signature string `json:"signature"`
	}{
		Alias:     (*Alias)(s),
		Signature: base64.StdEncoding.EncodeToString(s.Signature[:]),
	})
}

// UnmarshalJSON implements json.Unmarshaler interface for ImageSignature
func (s *ImageSignature) UnmarshalJSON(data []byte) error {
	type Alias ImageSignature
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

// MarshalJSON implements json.Marshaler interface for ImageSignature2
func (s *ImageSignature2) MarshalJSON() ([]byte, error) {
	type Alias ImageSignature2
	return json.Marshal(&struct {
		*Alias
		Signature string `json:"signature"`
	}{
		Alias:     (*Alias)(s),
		Signature: base64.StdEncoding.EncodeToString(s.Signature[:]),
	})
}

// UnmarshalJSON implements json.Unmarshaler interface for ImageSignature2
func (s *ImageSignature2) UnmarshalJSON(data []byte) error {
	type Alias ImageSignature2
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

// MarshalJSON implements json.Marshaler interface for PublicKey
func (p *PublicKey) MarshalJSON() ([]byte, error) {
	type Alias PublicKey
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

// UnmarshalJSON implements json.Unmarshaler interface for PublicKey
func (p *PublicKey) UnmarshalJSON(data []byte) error {
	type Alias PublicKey
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

// MarshalJSON implements json.Marshaler interface for PublicKey2
func (p *PublicKey2) MarshalJSON() ([]byte, error) {
	type Alias PublicKey2
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

// UnmarshalJSON implements json.Unmarshaler interface for PublicKey2
func (p *PublicKey2) UnmarshalJSON(data []byte) error {
	type Alias PublicKey2
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

// Unmarshal unmarshals binary data into ImageSignature
func (s *ImageSignature) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, s)
}

// Marshal marshals ImageSignature into binary data
func (s *ImageSignature) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(s)
}

// Unmarshal unmarshals binary data into ImageSignature2
func (s *ImageSignature2) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, s)
}

// Marshal marshals ImageSignature2 into binary data
func (s *ImageSignature2) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(s)
}

// Unmarshal unmarshals binary data into PublicKey
func (p *PublicKey) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, p)
}

// Marshal marshals PublicKey into binary data
func (p *PublicKey) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(p)
}

// Unmarshal unmarshals binary data into PublicKeys
func (p *PublicKeys) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, p)
}

// Marshal marshals PublicKeys into binary data
func (p *PublicKeys) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(p)
}

// MarshalWithReserved marshals PublicKeys including reserved fields
func (p *PublicKeys) MarshalWithReserved() ([]byte, error) {
	opts := &annotations.MarshalOptions{IncludeReserved: true}
	return annotations.MarshalWithOptionsStruct(p, opts)
}

// Unmarshal unmarshals binary data into PublicKey2
func (p *PublicKey2) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, p)
}

// Marshal marshals PublicKey2 into binary data
func (p *PublicKey2) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(p)
}

// Unmarshal unmarshals binary data into PublicKeys2
func (p *PublicKeys2) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, p)
}

// Marshal marshals PublicKeys2 into binary data
func (p *PublicKeys2) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(p)
}

// MarshalWithReserved marshals PublicKeys2 including reserved fields
func (p *PublicKeys2) MarshalWithReserved() ([]byte, error) {
	opts := &annotations.MarshalOptions{IncludeReserved: true}
	return annotations.MarshalWithOptionsStruct(p, opts)
}

// Unmarshal unmarshals binary data into ToolsAreaExtended
func (t *ToolsAreaExtended) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, t)
}

// Marshal marshals ToolsAreaExtended into binary data
func (t *ToolsAreaExtended) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(t)
}
