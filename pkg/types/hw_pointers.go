package types

import (
	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// HWPointerEntry represents a single hardware pointer entry (8 bytes) using annotations
// Based on image_layout_hw_pointer_entry from mstflint
// Note: In the actual firmware, the CRC is stored at offset 6, not offset 4
type HWPointerEntry struct {
	Ptr      uint32 `offset:"byte:0,endian:be" json:"ptr"`                    // Pointer value (offset 0x0)
	Reserved uint16 `offset:"byte:4,endian:be,reserved:true" json:"reserved"` // Reserved field (offset 0x4)
	CRC      uint16 `offset:"byte:6,endian:be" json:"crc"`                    // CRC16 value (offset 0x6)
}

// FS4HWPointers represents the Carmel hardware pointers structure (128 bytes) using annotations
// Based on image_layout_hw_pointers_carmel from mstflint
type FS4HWPointers struct {
	BootRecordPtr          HWPointerEntry `offset:"byte:0" json:"boot_record_ptr"`           // offset 0x0
	Boot2Ptr               HWPointerEntry `offset:"byte:8" json:"boot2_ptr"`                 // offset 0x8
	TOCPtr                 HWPointerEntry `offset:"byte:16" json:"toc_ptr"`                  // offset 0x10
	ToolsPtr               HWPointerEntry `offset:"byte:24" json:"tools_ptr"`                // offset 0x18
	AuthenticationStartPtr HWPointerEntry `offset:"byte:32" json:"authentication_start_ptr"` // offset 0x20
	AuthenticationEndPtr   HWPointerEntry `offset:"byte:40" json:"authentication_end_ptr"`   // offset 0x28
	DigestPtr              HWPointerEntry `offset:"byte:48" json:"digest_ptr"`               // offset 0x30
	DigestRecoveryKeyPtr   HWPointerEntry `offset:"byte:56" json:"digest_recovery_key_ptr"`  // offset 0x38
	FWWindowStartPtr       HWPointerEntry `offset:"byte:64" json:"fw_window_start_ptr"`      // offset 0x40
	FWWindowEndPtr         HWPointerEntry `offset:"byte:72" json:"fw_window_end_ptr"`        // offset 0x48
	ImageInfoSectionPtr    HWPointerEntry `offset:"byte:80" json:"image_info_section_ptr"`   // offset 0x50
	ImageSignaturePtr      HWPointerEntry `offset:"byte:88" json:"image_signature_ptr"`      // offset 0x58
	PublicKeyPtr           HWPointerEntry `offset:"byte:96" json:"public_key_ptr"`           // offset 0x60
	FWSecurityVersionPtr   HWPointerEntry `offset:"byte:104" json:"fw_security_version_ptr"` // offset 0x68
	GCMIVDeltaPtr          HWPointerEntry `offset:"byte:112" json:"gcm_iv_delta_ptr"`        // offset 0x70
	HashesTablePtr         HWPointerEntry `offset:"byte:120" json:"hashes_table_ptr"`        // offset 0x78
}

// FS5HWPointers represents the Gilboa hardware pointers structure (128 bytes) using annotations
// Based on fs5_image_layout_hw_pointers_gilboa from mstflint
type FS5HWPointers struct {
	Boot2Ptr             HWPointerEntry `offset:"byte:0" json:"boot2_ptr"`                 // offset 0x0
	TOCPtr               HWPointerEntry `offset:"byte:8" json:"toc_ptr"`                   // offset 0x8
	ToolsPtr             HWPointerEntry `offset:"byte:16" json:"tools_ptr"`                // offset 0x10
	ImageInfoSectionPtr  HWPointerEntry `offset:"byte:24" json:"image_info_section_ptr"`   // offset 0x18
	FWPublicKeyPtr       HWPointerEntry `offset:"byte:32" json:"fw_public_key_ptr"`        // offset 0x20
	FWSignaturePtr       HWPointerEntry `offset:"byte:40" json:"fw_signature_ptr"`         // offset 0x28
	PublicKeyPtr         HWPointerEntry `offset:"byte:48" json:"public_key_ptr"`           // offset 0x30
	ForbiddenVersionsPtr HWPointerEntry `offset:"byte:56" json:"forbidden_versions_ptr"`   // offset 0x38
	PSCBl1Ptr            HWPointerEntry `offset:"byte:64" json:"psc_bl1_ptr"`              // offset 0x40
	PSCHashesTablePtr    HWPointerEntry `offset:"byte:72" json:"psc_hashes_table_ptr"`     // offset 0x48
	NCoreHashesPointer   HWPointerEntry `offset:"byte:80" json:"ncore_hashes_pointer"`     // offset 0x50
	PSCFWUpdateHandlePtr HWPointerEntry `offset:"byte:88" json:"psc_fw_update_handle_ptr"` // offset 0x58
	PSCBCHPointer        HWPointerEntry `offset:"byte:96" json:"psc_bch_pointer"`          // offset 0x60
	ReservedPtr13        HWPointerEntry `offset:"byte:104" json:"reserved_ptr13"`          // offset 0x68
	ReservedPtr14        HWPointerEntry `offset:"byte:112" json:"reserved_ptr14"`          // offset 0x70
	NCoreBCHPointer      HWPointerEntry `offset:"byte:120" json:"ncore_bch_pointer"`       // offset 0x78
}

// Unmarshal methods using annotations

// Unmarshal unmarshals binary data into HWPointerEntry
func (h *HWPointerEntry) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (h *HWPointerEntry) UnmarshalWithReserved(data []byte) error {
	opts := &annotations.UnmarshalOptions{IncludeReserved: true}
	return annotations.UnmarshalWithOptionsStruct(data, h, opts)
}

// Marshal marshals HWPointerEntry into binary data
func (h *HWPointerEntry) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// MarshalWithReserved marshals HWPointerEntry including reserved fields
func (h *HWPointerEntry) MarshalWithReserved() ([]byte, error) {
	opts := &annotations.MarshalOptions{IncludeReserved: true}
	return annotations.MarshalWithOptionsStruct(h, opts)
}

// Unmarshal unmarshals binary data into FS4HWPointers
func (h *FS4HWPointers) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// Marshal marshals FS4HWPointers into binary data
func (h *FS4HWPointers) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// Unmarshal unmarshals binary data into FS5HWPointers
func (h *FS5HWPointers) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// Marshal marshals FS5HWPointers into binary data
func (h *FS5HWPointers) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}
