package sections

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// VPD_R0Section represents a VPD_R0 section
type VPD_R0Section struct {
	*interfaces.BaseSection
	Header *types.VPD_R0
	Data   []byte
}

// NewVPD_R0Section creates a new VPD_R0 section
func NewVPD_R0Section(base *interfaces.BaseSection) *VPD_R0Section {
	return &VPD_R0Section{
		BaseSection: base,
	}
}

// Parse parses the VPD_R0 section data
func (s *VPD_R0Section) Parse(data []byte) error {
	s.SetRawData(data)
	
	if len(data) < 64 {
		return merry.New("VPD_R0 section too small")
	}
	
	s.Header = &types.VPD_R0{}
	if err := s.Header.Unmarshal(data[:64]); err != nil {
		return merry.Wrap(err)
	}
	
	if len(data) > 64 {
		s.Data = data[64:]
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the VPD_R0 section
func (s *VPD_R0Section) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type":      s.Type(),
		"type_name": s.TypeName(),
		"offset":    s.Offset(),
		"size":      s.Size(),
		"has_raw_data": true, // VPD_R0 needs binary data
	}
	
	if s.Header != nil {
		result["vpd_r0"] = map[string]interface{}{
			"id":        string(s.Header.ID[:]),
			"length":    s.Header.Length,
			"data_size": len(s.Data),
		}
	}
	
	return json.Marshal(result)
}

// FWNVLogSection represents a FW_NV_LOG section
type FWNVLogSection struct {
	*interfaces.BaseSection
	Header *types.FWNVLog
	Data   []byte
}

// NewFWNVLogSection creates a new FWNVLog section
func NewFWNVLogSection(base *interfaces.BaseSection) *FWNVLogSection {
	return &FWNVLogSection{
		BaseSection: base,
	}
}

// Parse parses the FW_NV_LOG section data
func (s *FWNVLogSection) Parse(data []byte) error {
	s.SetRawData(data)
	
	if len(data) < 64 {
		return merry.New("FW_NV_LOG section too small")
	}
	
	s.Header = &types.FWNVLog{}
	if err := s.Header.Unmarshal(data[:64]); err != nil {
		return merry.Wrap(err)
	}
	
	if len(data) > 64 {
		s.Data = data[64:]
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the FW_NV_LOG section
func (s *FWNVLogSection) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type":      s.Type(),
		"type_name": s.TypeName(),
		"offset":    s.Offset(),
		"size":      s.Size(),
		"has_raw_data": true, // FW_NV_LOG needs binary data
	}
	
	if s.Header != nil {
		result["fw_nv_log"] = map[string]interface{}{
			"log_version": s.Header.LogVersion,
			"log_size":    s.Header.LogSize,
			"entry_count": s.Header.EntryCount,
			"data_size":   len(s.Data),
		}
	}
	
	return json.Marshal(result)
}

// NVDataSection represents NV_DATA sections
type NVDataSection struct {
	*interfaces.BaseSection
	Header *types.NVData
	Data   []byte
}

// NewNVDataSection creates a new NVData section
func NewNVDataSection(base *interfaces.BaseSection) *NVDataSection {
	return &NVDataSection{
		BaseSection: base,
	}
}

// Parse parses the NV_DATA section data
func (s *NVDataSection) Parse(data []byte) error {
	s.SetRawData(data)
	
	if len(data) < 64 {
		return merry.New("NV_DATA section too small")
	}
	
	s.Header = &types.NVData{}
	if err := s.Header.Unmarshal(data[:64]); err != nil {
		return merry.Wrap(err)
	}
	
	if len(data) > 64 {
		s.Data = data[64:]
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the NV_DATA section
func (s *NVDataSection) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type":      s.Type(),
		"type_name": s.TypeName(),
		"offset":    s.Offset(),
		"size":      s.Size(),
		"has_raw_data": true, // NV_DATA needs binary data
	}
	
	if s.Header != nil {
		result["nv_data"] = map[string]interface{}{
			"version":   s.Header.Version,
			"data_size": s.Header.DataSize,
			"actual_data_size": len(s.Data),
		}
	}
	
	return json.Marshal(result)
}

// CRDumpMaskDataSection represents a CRDUMP_MASK_DATA section
type CRDumpMaskDataSection struct {
	*interfaces.BaseSection
	Header *types.CRDumpMaskData
	Data   []byte
}

// NewCRDumpMaskDataSection creates a new CRDumpMaskData section
func NewCRDumpMaskDataSection(base *interfaces.BaseSection) *CRDumpMaskDataSection {
	return &CRDumpMaskDataSection{
		BaseSection: base,
	}
}

// Parse parses the CRDUMP_MASK_DATA section data
func (s *CRDumpMaskDataSection) Parse(data []byte) error {
	s.SetRawData(data)
	
	if len(data) < 40 {
		return merry.New("CRDUMP_MASK_DATA section too small")
	}
	
	s.Header = &types.CRDumpMaskData{}
	if err := s.Header.Unmarshal(data[:40]); err != nil {
		return merry.Wrap(err)
	}
	
	if len(data) > 40 {
		s.Data = data[40:]
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the CRDUMP_MASK_DATA section
func (s *CRDumpMaskDataSection) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type":      s.Type(),
		"type_name": s.TypeName(),
		"offset":    s.Offset(),
		"size":      s.Size(),
		"has_raw_data": true, // CRDUMP_MASK_DATA needs binary data
	}
	
	if s.Header != nil {
		result["crdump_mask_data"] = map[string]interface{}{
			"version":   s.Header.Version,
			"mask_size": s.Header.MaskSize,
			"data_size": len(s.Data),
		}
	}
	
	return json.Marshal(result)
}

// FWInternalUsageSection represents a FW_INTERNAL_USAGE section
type FWInternalUsageSection struct {
	*interfaces.BaseSection
	Header *types.FWInternalUsage
	Data   []byte
}

// NewFWInternalUsageSection creates a new FWInternalUsage section
func NewFWInternalUsageSection(base *interfaces.BaseSection) *FWInternalUsageSection {
	return &FWInternalUsageSection{
		BaseSection: base,
	}
}

// Parse parses the FW_INTERNAL_USAGE section data
func (s *FWInternalUsageSection) Parse(data []byte) error {
	s.SetRawData(data)
	
	if len(data) < 64 {
		return merry.New("FW_INTERNAL_USAGE section too small")
	}
	
	s.Header = &types.FWInternalUsage{}
	if err := s.Header.Unmarshal(data[:64]); err != nil {
		return merry.Wrap(err)
	}
	
	if len(data) > 64 {
		s.Data = data[64:]
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the FW_INTERNAL_USAGE section
func (s *FWInternalUsageSection) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type":      s.Type(),
		"type_name": s.TypeName(),
		"offset":    s.Offset(),
		"size":      s.Size(),
		"has_raw_data": true, // FW_INTERNAL_USAGE needs binary data
	}
	
	if s.Header != nil {
		result["fw_internal_usage"] = map[string]interface{}{
			"version":   s.Header.Version,
			"size":      s.Header.Size,
			"type":      s.Header.Type,
			"data_size": len(s.Data),
		}
	}
	
	return json.Marshal(result)
}

// ProgrammableHWFWSection represents PROGRAMMABLE_HW_FW sections
type ProgrammableHWFWSection struct {
	*interfaces.BaseSection
	Header *types.ProgrammableHWFW
	Data   []byte
}

// NewProgrammableHWFWSection creates a new ProgrammableHWFW section
func NewProgrammableHWFWSection(base *interfaces.BaseSection) *ProgrammableHWFWSection {
	return &ProgrammableHWFWSection{
		BaseSection: base,
	}
}

// Parse parses the PROGRAMMABLE_HW_FW section data
func (s *ProgrammableHWFWSection) Parse(data []byte) error {
	s.SetRawData(data)
	
	if len(data) < 64 {
		return merry.New("PROGRAMMABLE_HW_FW section too small")
	}
	
	s.Header = &types.ProgrammableHWFW{}
	if err := s.Header.Unmarshal(data[:64]); err != nil {
		return merry.Wrap(err)
	}
	
	if len(data) > 64 {
		s.Data = data[64:]
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the PROGRAMMABLE_HW_FW section
func (s *ProgrammableHWFWSection) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type":      s.Type(),
		"type_name": s.TypeName(),
		"offset":    s.Offset(),
		"size":      s.Size(),
		"has_raw_data": true, // PROGRAMMABLE_HW_FW needs binary data
	}
	
	if s.Header != nil {
		result["programmable_hw_fw"] = map[string]interface{}{
			"version":      s.Header.Version,
			"hw_type":      s.Header.HWType,
			"fw_size":      s.Header.FWSize,
			"checksum":     s.Header.Checksum,
			"load_address": fmt.Sprintf("0x%08X", s.Header.LoadAddress),
			"entry_point":  fmt.Sprintf("0x%08X", s.Header.EntryPoint),
			"data_size":    len(s.Data),
		}
	}
	
	return json.Marshal(result)
}

// DigitalCertPtrSection represents DIGITAL_CERT_PTR section
type DigitalCertPtrSection struct {
	*interfaces.BaseSection
	CertPtr *types.DigitalCertPtr
}

// NewDigitalCertPtrSection creates a new DigitalCertPtr section
func NewDigitalCertPtrSection(base *interfaces.BaseSection) *DigitalCertPtrSection {
	return &DigitalCertPtrSection{
		BaseSection: base,
	}
}

// Parse parses the DIGITAL_CERT_PTR section data
func (s *DigitalCertPtrSection) Parse(data []byte) error {
	s.SetRawData(data)
	
	s.CertPtr = &types.DigitalCertPtr{}
	if err := s.CertPtr.Unmarshal(data); err != nil {
		return merry.Wrap(err)
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the DIGITAL_CERT_PTR section
func (s *DigitalCertPtrSection) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type":      s.Type(),
		"type_name": s.TypeName(),
		"offset":    s.Offset(),
		"size":      s.Size(),
		"has_raw_data": true, // DIGITAL_CERT_PTR needs binary data due to variable structure
	}
	
	if s.CertPtr != nil {
		result["digital_cert_ptr"] = map[string]interface{}{
			"cert_type":   s.CertPtr.CertType,
			"cert_offset": fmt.Sprintf("0x%08X", s.CertPtr.CertOffset),
			"cert_size":   s.CertPtr.CertSize,
		}
	}
	
	return json.Marshal(result)
}

// DigitalCertRWSection represents DIGITAL_CERT_RW section
type DigitalCertRWSection struct {
	*interfaces.BaseSection
	Cert *types.DigitalCertRW
}

// NewDigitalCertRWSection creates a new DigitalCertRW section
func NewDigitalCertRWSection(base *interfaces.BaseSection) *DigitalCertRWSection {
	return &DigitalCertRWSection{
		BaseSection: base,
	}
}

// Parse parses the DIGITAL_CERT_RW section data
func (s *DigitalCertRWSection) Parse(data []byte) error {
	s.SetRawData(data)
	
	s.Cert = &types.DigitalCertRW{}
	if err := s.Cert.Unmarshal(data); err != nil {
		return merry.Wrap(err)
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the DIGITAL_CERT_RW section
func (s *DigitalCertRWSection) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type":      s.Type(),
		"type_name": s.TypeName(),
		"offset":    s.Offset(),
		"size":      s.Size(),
		"has_raw_data": true, // DIGITAL_CERT_RW needs binary data for certificate
	}
	
	if s.Cert != nil {
		// Find actual certificate size
		certSize := 0
		for i := 0; i < len(s.Cert.Certificate); i++ {
			if s.Cert.Certificate[i] != 0 {
				certSize = i + 1
			}
		}
		
		result["digital_cert_rw"] = map[string]interface{}{
			"cert_type":   s.Cert.CertType,
			"cert_size":   s.Cert.CertSize,
			"valid_from":  s.Cert.ValidFrom,
			"valid_to":    s.Cert.ValidTo,
			"actual_cert_size": certSize,
		}
		
		if certSize > 0 && certSize <= 256 { // Include small cert preview
			result["digital_cert_rw"].(map[string]interface{})["cert_preview"] = hex.EncodeToString(s.Cert.Certificate[:certSize])
		}
	}
	
	return json.Marshal(result)
}