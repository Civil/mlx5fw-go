package types

import (
	"reflect"
	
	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// HWPointerEntryAnnotated represents a single hardware pointer entry (8 bytes) using annotations
// Based on image_layout_hw_pointer_entry from mstflint
// Note: In the actual firmware, the CRC is stored at offset 6, not offset 4
type HWPointerEntryAnnotated struct {
	Ptr      uint32 `offset:"byte:0,endian:be"`          // Pointer value (offset 0x0)
	Reserved uint16 `offset:"byte:4,endian:be,reserved:true"` // Reserved field (offset 0x4)
	CRC      uint16 `offset:"byte:6,endian:be"`          // CRC16 value (offset 0x6)
}

// FS4HWPointersAnnotated represents the Carmel hardware pointers structure (128 bytes) using annotations
// Based on image_layout_hw_pointers_carmel from mstflint
type FS4HWPointersAnnotated struct {
	BootRecordPtr              HWPointerEntryAnnotated `offset:"byte:0"`   // offset 0x0
	Boot2Ptr                   HWPointerEntryAnnotated `offset:"byte:8"`   // offset 0x8
	TOCPtr                     HWPointerEntryAnnotated `offset:"byte:16"`  // offset 0x10
	ToolsPtr                   HWPointerEntryAnnotated `offset:"byte:24"`  // offset 0x18
	AuthenticationStartPtr     HWPointerEntryAnnotated `offset:"byte:32"`  // offset 0x20
	AuthenticationEndPtr       HWPointerEntryAnnotated `offset:"byte:40"`  // offset 0x28
	DigestPtr                  HWPointerEntryAnnotated `offset:"byte:48"`  // offset 0x30
	DigestRecoveryKeyPtr       HWPointerEntryAnnotated `offset:"byte:56"`  // offset 0x38
	FWWindowStartPtr           HWPointerEntryAnnotated `offset:"byte:64"`  // offset 0x40
	FWWindowEndPtr             HWPointerEntryAnnotated `offset:"byte:72"`  // offset 0x48
	ImageInfoSectionPtr        HWPointerEntryAnnotated `offset:"byte:80"`  // offset 0x50
	ImageSignaturePtr          HWPointerEntryAnnotated `offset:"byte:88"`  // offset 0x58
	PublicKeyPtr               HWPointerEntryAnnotated `offset:"byte:96"`  // offset 0x60
	FWSecurityVersionPtr       HWPointerEntryAnnotated `offset:"byte:104"` // offset 0x68
	GCMIVDeltaPtr              HWPointerEntryAnnotated `offset:"byte:112"` // offset 0x70
	HashesTablePtr             HWPointerEntryAnnotated `offset:"byte:120"` // offset 0x78
}

// FS5HWPointersAnnotated represents the Gilboa hardware pointers structure (128 bytes) using annotations
// Based on fs5_image_layout_hw_pointers_gilboa from mstflint
type FS5HWPointersAnnotated struct {
	Boot2Ptr                   HWPointerEntryAnnotated `offset:"byte:0"`   // offset 0x0
	TOCPtr                     HWPointerEntryAnnotated `offset:"byte:8"`   // offset 0x8
	ToolsPtr                   HWPointerEntryAnnotated `offset:"byte:16"`  // offset 0x10
	ImageInfoSectionPtr        HWPointerEntryAnnotated `offset:"byte:24"`  // offset 0x18
	FWPublicKeyPtr             HWPointerEntryAnnotated `offset:"byte:32"`  // offset 0x20
	FWSignaturePtr             HWPointerEntryAnnotated `offset:"byte:40"`  // offset 0x28
	PublicKeyPtr               HWPointerEntryAnnotated `offset:"byte:48"`  // offset 0x30
	ForbiddenVersionsPtr       HWPointerEntryAnnotated `offset:"byte:56"`  // offset 0x38
	PSCBl1Ptr                  HWPointerEntryAnnotated `offset:"byte:64"`  // offset 0x40
	PSCHashesTablePtr          HWPointerEntryAnnotated `offset:"byte:72"`  // offset 0x48
	NCoreHashesPointer         HWPointerEntryAnnotated `offset:"byte:80"`  // offset 0x50
	PSCFWUpdateHandlePtr       HWPointerEntryAnnotated `offset:"byte:88"`  // offset 0x58
	PSCBCHPointer              HWPointerEntryAnnotated `offset:"byte:96"`  // offset 0x60
	ReservedPtr13              HWPointerEntryAnnotated `offset:"byte:104"` // offset 0x68
	ReservedPtr14              HWPointerEntryAnnotated `offset:"byte:112"` // offset 0x70
	NCoreBCHPointer            HWPointerEntryAnnotated `offset:"byte:120"` // offset 0x78
}

// Unmarshal methods using annotations


// Unmarshal unmarshals binary data into HWPointerEntryAnnotated
func (h *HWPointerEntryAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (h *HWPointerEntryAnnotated) UnmarshalWithReserved(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*h))
	if err != nil {
		return err
	}
	opts := &annotations.UnmarshalOptions{
		IncludeReserved: true,
	}
	return annotations.UnmarshalWithOptions(data, h, annot, opts)
}

// Marshal marshals HWPointerEntryAnnotated into binary data
func (h *HWPointerEntryAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// MarshalWithReserved marshals HWPointerEntryAnnotated including reserved fields
func (h *HWPointerEntryAnnotated) MarshalWithReserved() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*h))
	if err != nil {
		return nil, err
	}
	opts := &annotations.MarshalOptions{
		IncludeReserved: true,
	}
	return annotations.MarshalWithOptions(h, annot, opts)
}

// Unmarshal unmarshals binary data into FS4HWPointersAnnotated
func (h *FS4HWPointersAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// Marshal marshals FS4HWPointersAnnotated into binary data
func (h *FS4HWPointersAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// Unmarshal unmarshals binary data into FS5HWPointersAnnotated
func (h *FS5HWPointersAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// Marshal marshals FS5HWPointersAnnotated into binary data
func (h *FS5HWPointersAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

