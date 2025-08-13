package types

import (
	"bytes"
	"encoding/binary"
)

// Marshal/Unmarshal methods for FWVersion
func (f *FWVersion) Unmarshal(data []byte) error {
	return binary.Read(bytes.NewReader(data), binary.BigEndian, f)
}

func (f *FWVersion) Marshal() ([]byte, error) {
	return MarshalBE(f)
}

// Marshal/Unmarshal methods for TripleVersion
func (t *TripleVersion) Unmarshal(data []byte) error {
	return binary.Read(bytes.NewReader(data), binary.BigEndian, t)
}

func (t *TripleVersion) Marshal() ([]byte, error) {
	return MarshalBE(t)
}

// Marshal/Unmarshal methods for ModuleVersion
func (m *ModuleVersion) Unmarshal(data []byte) error {
	return binary.Read(bytes.NewReader(data), binary.BigEndian, m)
}

func (m *ModuleVersion) Marshal() ([]byte, error) {
	return MarshalBE(m)
}

// Marshal/Unmarshal methods for ImageSize
func (i *ImageSize) Unmarshal(data []byte) error {
	return binary.Read(bytes.NewReader(data), binary.BigEndian, i)
}

func (i *ImageSize) Marshal() ([]byte, error) {
	return MarshalBE(i)
}

// Marshal/Unmarshal methods for ModuleVersions
func (m *ModuleVersions) Unmarshal(data []byte) error {
	return binary.Read(bytes.NewReader(data), binary.BigEndian, m)
}

func (m *ModuleVersions) Marshal() ([]byte, error) {
	return MarshalBE(m)
}

// Marshal/Unmarshal methods for UIDEntry
func (u *UIDEntry) Unmarshal(data []byte) error {
	return binary.Read(bytes.NewReader(data), binary.BigEndian, u)
}

func (u *UIDEntry) Marshal() ([]byte, error) {
	return MarshalBE(u)
}

// Marshal/Unmarshal methods for Guids
func (g *Guids) Unmarshal(data []byte) error {
	return binary.Read(bytes.NewReader(data), binary.BigEndian, g)
}

func (g *Guids) Marshal() ([]byte, error) {
	return MarshalBE(g)
}

// Marshal/Unmarshal methods for OperationKey
func (o *OperationKey) Unmarshal(data []byte) error {
	return binary.Read(bytes.NewReader(data), binary.BigEndian, o)
}

func (o *OperationKey) Marshal() ([]byte, error) {
	return MarshalBE(o)
}

// Marshal/Unmarshal methods for ImageInfoExtended
func (i *ImageInfoExtended) Unmarshal(data []byte) error {
	// Extended image info uses little-endian layout; use binary.Read fallback
	return binary.Read(bytes.NewReader(data), binary.LittleEndian, i)
}

func (i *ImageInfoExtended) Marshal() ([]byte, error) {
	return MarshalLE(i)
}

// Marshal/Unmarshal methods for HashesTableHeaderExtended
func (h *HashesTableHeaderExtended) Unmarshal(data []byte) error {
	return binary.Read(bytes.NewReader(data), binary.BigEndian, h)
}

func (h *HashesTableHeaderExtended) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for HTOCEntry
func (h *HTOCEntry) Unmarshal(data []byte) error {
	return binary.Read(bytes.NewReader(data), binary.BigEndian, h)
}

func (h *HTOCEntry) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for HTOCHeader
func (h *HTOCHeader) Unmarshal(data []byte) error {
	return binary.Read(bytes.NewReader(data), binary.BigEndian, h)
}

func (h *HTOCHeader) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for HTOC
func (h *HTOC) Unmarshal(data []byte) error {
	return binary.Read(bytes.NewReader(data), binary.BigEndian, h)
}

func (h *HTOC) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for HTOCHash
func (h *HTOCHash) Unmarshal(data []byte) error {
	return binary.Read(bytes.NewReader(data), binary.BigEndian, h)
}

func (h *HTOCHash) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for HashesTableExtended
func (h *HashesTableExtended) Unmarshal(data []byte) error {
	return binary.Read(bytes.NewReader(data), binary.BigEndian, h)
}

func (h *HashesTableExtended) Marshal() ([]byte, error) {
	return MarshalBE(h)
}

// Marshal/Unmarshal methods for BootVersion
func (b *BootVersion) Unmarshal(data []byte) error {
	return binary.Read(bytes.NewReader(data), binary.BigEndian, b)
}

func (b *BootVersion) Marshal() ([]byte, error) {
	return MarshalBE(b)
}

// Marshal/Unmarshal methods for DeviceInfoExtended
func (d *DeviceInfoExtended) Unmarshal(data []byte) error {
	return binary.Read(bytes.NewReader(data), binary.BigEndian, d)
}

func (d *DeviceInfoExtended) Marshal() ([]byte, error) {
	return MarshalBE(d)
}
