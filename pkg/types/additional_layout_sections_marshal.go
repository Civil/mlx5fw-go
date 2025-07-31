package types

import (
	"github.com/ghostiam/binstruct"
)

// Marshal/Unmarshal methods for FS4ComponentAuthenticationConfiguration
func (c *FS4ComponentAuthenticationConfiguration) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, c)
}

func (c *FS4ComponentAuthenticationConfiguration) Marshal() ([]byte, error) {
	return MarshalBE(c)
}

// Marshal/Unmarshal methods for FS4HtocEntry
func (h *FS4HtocEntry) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, h)
}

func (h *FS4HtocEntry) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for FS4HtocHeader
func (h *FS4HtocHeader) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, h)
}

func (h *FS4HtocHeader) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for FS4HtocHash
func (h *FS4HtocHash) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, h)
}

func (h *FS4HtocHash) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for FS4Htoc
func (h *FS4Htoc) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, h)
}

func (h *FS4Htoc) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for FS4HashesTableHeader
func (h *FS4HashesTableHeader) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, h)
}

func (h *FS4HashesTableHeader) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for FS4HashesTable
func (h *FS4HashesTable) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, h)
}

func (h *FS4HashesTable) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for FS4HwPointerEntry
func (h *FS4HwPointerEntry) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, h)
}

func (h *FS4HwPointerEntry) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for FS4HwPointersCarmel
func (h *FS4HwPointersCarmel) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, h)
}

func (h *FS4HwPointersCarmel) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for FS5HwPointerEntry
func (h *FS5HwPointerEntry) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, h)
}

func (h *FS5HwPointerEntry) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for FS5HwPointersGilboa
func (h *FS5HwPointersGilboa) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, h)
}

func (h *FS5HwPointersGilboa) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for FS4FilePublicKeys
func (f *FS4FilePublicKeys) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, f)
}

func (f *FS4FilePublicKeys) Marshal() ([]byte, error) {
	return MarshalBE(f)
}

// Marshal/Unmarshal methods for FS4FilePublicKeys2
func (f *FS4FilePublicKeys2) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, f)
}

func (f *FS4FilePublicKeys2) Marshal() ([]byte, error) {
	return MarshalBE(f)
}

// Marshal/Unmarshal methods for FS4FilePublicKeys3
func (f *FS4FilePublicKeys3) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, f)
}

func (f *FS4FilePublicKeys3) Marshal() ([]byte, error) {
	return MarshalBE(f)
}

// Marshal/Unmarshal methods for FS4PublicKeysStruct
func (p *FS4PublicKeysStruct) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, p)
}

func (p *FS4PublicKeysStruct) Marshal() ([]byte, error) {
	return MarshalBE(p)
}

// Marshal/Unmarshal methods for FS4PublicKeys2Struct
func (p *FS4PublicKeys2Struct) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, p)
}

func (p *FS4PublicKeys2Struct) Marshal() ([]byte, error) {
	return MarshalBE(p)
}

// Marshal/Unmarshal methods for FS4PublicKeys3Struct
func (p *FS4PublicKeys3Struct) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, p)
}

func (p *FS4PublicKeys3Struct) Marshal() ([]byte, error) {
	return MarshalBE(p)
}

// Marshal/Unmarshal methods for FS4ImageSignatureStruct
func (i *FS4ImageSignatureStruct) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, i)
}

func (i *FS4ImageSignatureStruct) Marshal() ([]byte, error) {
	return MarshalBE(i)
}

// Marshal/Unmarshal methods for FS4ImageSignature2Struct
func (i *FS4ImageSignature2Struct) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, i)
}

func (i *FS4ImageSignature2Struct) Marshal() ([]byte, error) {
	return MarshalBE(i)
}

// Marshal/Unmarshal methods for FS4SecureBootSignaturesStruct
func (s *FS4SecureBootSignaturesStruct) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, s)
}

func (s *FS4SecureBootSignaturesStruct) Marshal() ([]byte, error) {
	return MarshalBE(s)
}

// Marshal/Unmarshal methods for FS4BootVersionStruct
func (b *FS4BootVersionStruct) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, b)
}

func (b *FS4BootVersionStruct) Marshal() ([]byte, error) {
	return MarshalBE(b)
}