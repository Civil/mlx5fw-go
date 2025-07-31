package types

import (
	"github.com/ghostiam/binstruct"
)

// Marshal/Unmarshal methods for FWVersion
func (f *FWVersion) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, f)
}

func (f *FWVersion) Marshal() ([]byte, error) {
	return MarshalBE(f)
}

// Marshal/Unmarshal methods for TripleVersion
func (t *TripleVersion) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, t)
}

func (t *TripleVersion) Marshal() ([]byte, error) {
	return MarshalBE(t)
}

// Marshal/Unmarshal methods for ModuleVersion
func (m *ModuleVersion) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, m)
}

func (m *ModuleVersion) Marshal() ([]byte, error) {
	return MarshalBE(m)
}

// Marshal/Unmarshal methods for ImageSize
func (i *ImageSize) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, i)
}

func (i *ImageSize) Marshal() ([]byte, error) {
	return MarshalBE(i)
}

// Marshal/Unmarshal methods for ModuleVersions
func (m *ModuleVersions) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, m)
}

func (m *ModuleVersions) Marshal() ([]byte, error) {
	return MarshalBE(m)
}

// Marshal/Unmarshal methods for UIDEntry
func (u *UIDEntry) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, u)
}

func (u *UIDEntry) Marshal() ([]byte, error) {
	return MarshalBE(u)
}

// Marshal/Unmarshal methods for Guids
func (g *Guids) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, g)
}

func (g *Guids) Marshal() ([]byte, error) {
	return MarshalBE(g)
}

// Marshal/Unmarshal methods for OperationKey
func (o *OperationKey) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, o)
}

func (o *OperationKey) Marshal() ([]byte, error) {
	return MarshalBE(o)
}

// Marshal/Unmarshal methods for ImageInfoExtended
func (i *ImageInfoExtended) Unmarshal(data []byte) error {
	return binstruct.UnmarshalLE(data, i)
}

func (i *ImageInfoExtended) Marshal() ([]byte, error) {
	return MarshalLE(i)
}

// Marshal/Unmarshal methods for HashesTableHeaderExtended
func (h *HashesTableHeaderExtended) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, h)
}

func (h *HashesTableHeaderExtended) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for HTOCEntry
func (h *HTOCEntry) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, h)
}

func (h *HTOCEntry) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for HTOCHeader
func (h *HTOCHeader) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, h)
}

func (h *HTOCHeader) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for HTOC
func (h *HTOC) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, h)
}

func (h *HTOC) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for HTOCHash
func (h *HTOCHash) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, h)
}

func (h *HTOCHash) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for HashesTableExtended
func (h *HashesTableExtended) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, h)
}

func (h *HashesTableExtended) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for BootVersion
func (b *BootVersion) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, b)
}

func (b *BootVersion) Marshal() ([]byte, error) {
	return MarshalBE(b)
}

// Marshal/Unmarshal methods for DeviceInfoExtended
func (d *DeviceInfoExtended) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, d)
}

func (d *DeviceInfoExtended) Marshal() ([]byte, error) {
	return MarshalBE(d)
}

// Marshal/Unmarshal methods for ImageSignature
func (i *ImageSignature) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &ImageSignatureAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	*i = *annotated.FromAnnotated()
	return nil
}

func (i *ImageSignature) Marshal() ([]byte, error) {
	// Use annotated version for marshaling
	annotated := i.ToAnnotated()
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for ImageSignature2
func (i *ImageSignature2) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &ImageSignature2Annotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	*i = *annotated.FromAnnotated()
	return nil
}

func (i *ImageSignature2) Marshal() ([]byte, error) {
	// Use annotated version for marshaling
	annotated := i.ToAnnotated()
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for PublicKey
func (p *PublicKey) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, p)
}

func (p *PublicKey) Marshal() ([]byte, error) {
	return MarshalBE(p)
}

// Marshal/Unmarshal methods for PublicKeys
func (p *PublicKeys) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &PublicKeysAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	*p = *annotated.FromAnnotated()
	return nil
}

func (p *PublicKeys) Marshal() ([]byte, error) {
	// Use annotated version for marshaling
	annotated := p.ToAnnotated()
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for PublicKey2
func (p *PublicKey2) Unmarshal(data []byte) error {
	return binstruct.UnmarshalBE(data, p)
}

func (p *PublicKey2) Marshal() ([]byte, error) {
	return MarshalBE(p)
}

// Marshal/Unmarshal methods for PublicKeys2
func (p *PublicKeys2) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &PublicKeys2Annotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	*p = *annotated.FromAnnotated()
	return nil
}

func (p *PublicKeys2) Marshal() ([]byte, error) {
	// Use annotated version for marshaling
	annotated := p.ToAnnotated()
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for ToolsAreaExtended
func (t *ToolsAreaExtended) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &ToolsAreaExtendedAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	*t = *annotated.FromAnnotated()
	return nil
}

func (t *ToolsAreaExtended) Marshal() ([]byte, error) {
	// Use annotated version for marshaling
	annotated := t.ToAnnotated()
	return annotated.Marshal()
}