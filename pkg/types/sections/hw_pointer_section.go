package sections

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// HWPointerSection represents a Hardware Pointer section
type HWPointerSection struct {
	*interfaces.BaseSection
	FS4Pointers *types.FS4HWPointers
	FS5Pointers *types.FS5HWPointers
	Format      types.FirmwareFormat
}

// NewHWPointerSection creates a new HW Pointer section
func NewHWPointerSection(base *interfaces.BaseSection) *HWPointerSection {
	return &HWPointerSection{
		BaseSection: base,
	}
}

// Parse parses the HW Pointer section data
func (s *HWPointerSection) Parse(data []byte) error {
	s.SetRawData(data)
	
	if len(data) < 4 {
		return merry.New("HW pointer section too small")
	}
	
	// Determine format based on data size and content
	// FS5 has larger pointer structure
	if len(data) >= 0x100 { // FS5 size threshold
		s.Format = types.FormatFS5
		s.FS5Pointers = &types.FS5HWPointers{}
		reader := bytes.NewReader(data)
		if err := binary.Read(reader, binary.BigEndian, s.FS5Pointers); err != nil {
			return merry.Wrap(err)
		}
	} else {
		s.Format = types.FormatFS4
		s.FS4Pointers = &types.FS4HWPointers{}
		reader := bytes.NewReader(data)
		if err := binary.Read(reader, binary.BigEndian, s.FS4Pointers); err != nil {
			return merry.Wrap(err)
		}
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the HW Pointer section
func (s *HWPointerSection) MarshalJSON() ([]byte, error) {
	baseInfo := map[string]interface{}{
		"type":         s.Type(),
		"type_name":    s.TypeName(),
		"offset":       s.Offset(),
		"size":         s.Size(),
		"format":       s.Format.String(),
	}
	
	if s.FS4Pointers != nil {
		baseInfo["hw_pointers"] = map[string]interface{}{
			"boot_record_ptr": s.FS4Pointers.BootRecordPtr,
			"boot2_ptr": s.FS4Pointers.Boot2Ptr,
			"toc_ptr": s.FS4Pointers.TOCPtr,
			"tools_ptr": s.FS4Pointers.ToolsPtr,
			"fw_window_start_ptr": s.FS4Pointers.FWWindowStartPtr,
			"fw_window_end_ptr": s.FS4Pointers.FWWindowEndPtr,
			"image_info_section_ptr": s.FS4Pointers.ImageInfoSectionPtr,
			"hashes_table_ptr": s.FS4Pointers.HashesTablePtr,
			"digest_recovery_key_ptr": s.FS4Pointers.DigestRecoveryKeyPtr,
			"digest_ptr": s.FS4Pointers.DigestPtr,
		}
	} else if s.FS5Pointers != nil {
		baseInfo["hw_pointers"] = map[string]interface{}{
			"boot2_ptr": s.FS5Pointers.Boot2Ptr,
			"toc_ptr": s.FS5Pointers.TOCPtr,
			"tools_ptr": s.FS5Pointers.ToolsPtr,
			"image_info_section_ptr": s.FS5Pointers.ImageInfoSectionPtr,
			"fw_public_key_ptr": s.FS5Pointers.FWPublicKeyPtr,
			"fw_signature_ptr": s.FS5Pointers.FWSignaturePtr,
			"public_key_ptr": s.FS5Pointers.PublicKeyPtr,
			"forbidden_versions_ptr": s.FS5Pointers.ForbiddenVersionsPtr,
			"psc_bl1_ptr": s.FS5Pointers.PSCBl1Ptr,
			"psc_hashes_table_ptr": s.FS5Pointers.PSCHashesTablePtr,
			"ncore_hashes_pointer": s.FS5Pointers.NCoreHashesPointer,
			"psc_fw_update_handle_ptr": s.FS5Pointers.PSCFWUpdateHandlePtr,
			"psc_bch_pointer": s.FS5Pointers.PSCBCHPointer,
			"ncore_bch_pointer": s.FS5Pointers.NCoreBCHPointer,
		}
	}
	
	return json.Marshal(baseInfo)
}