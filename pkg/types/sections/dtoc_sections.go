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
	
	// Handle zero-length VPD_R0 sections
	if len(data) == 0 {
		// VPD_R0 sections can have size 0 but still have CRC in ITOC entry
		// No header to parse in this case
		return nil
	}
	
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
	sectionJSON := &types.SectionJSON{
		Type:         s.Type(),
		TypeName:     s.TypeName(),
		Offset:       s.Offset(),
		Size:         s.Size(),
		CRCType:      s.CRCType().String(),
		IsEncrypted:  s.IsEncrypted(),
		IsDeviceData: s.IsDeviceData(),
		HasRawData:   true, // VPD_R0 needs binary data
	}
	
	if s.Header != nil {
		sectionJSON.VPD_R0 = &types.VPD_R0JSON{
			ID:       string(s.Header.ID[:]),
			Length:   s.Header.Length,
			DataSize: len(s.Data),
		}
	}
	
	return json.Marshal(sectionJSON)
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
	sectionJSON := &types.SectionJSON{
		Type:         s.Type(),
		TypeName:     s.TypeName(),
		Offset:       s.Offset(),
		Size:         s.Size(),
		CRCType:      s.CRCType().String(),
		IsEncrypted:  s.IsEncrypted(),
		IsDeviceData: s.IsDeviceData(),
		HasRawData:   true, // FW_NV_LOG needs binary data
	}
	
	if s.Header != nil {
		sectionJSON.FWNVLog = &types.FWNVLogJSON{
			LogVersion: s.Header.LogVersion,
			LogSize:    s.Header.LogSize,
			EntryCount: s.Header.EntryCount,
			DataSize:   len(s.Data),
		}
	}
	
	return json.Marshal(sectionJSON)
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
	sectionJSON := &types.SectionJSON{
		Type:         s.Type(),
		TypeName:     s.TypeName(),
		Offset:       s.Offset(),
		Size:         s.Size(),
		CRCType:      s.CRCType().String(),
		IsEncrypted:  s.IsEncrypted(),
		IsDeviceData: s.IsDeviceData(),
		HasRawData:   true, // NV_DATA needs binary data
	}
	
	if s.Header != nil {
		sectionJSON.NVData = &types.NVDataJSON{
			Version:        s.Header.Version,
			DataSize:       s.Header.DataSize,
			ActualDataSize: len(s.Data),
		}
	}
	
	return json.Marshal(sectionJSON)
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
	sectionJSON := &types.SectionJSON{
		Type:         s.Type(),
		TypeName:     s.TypeName(),
		Offset:       s.Offset(),
		Size:         s.Size(),
		CRCType:      s.CRCType().String(),
		IsEncrypted:  s.IsEncrypted(),
		IsDeviceData: s.IsDeviceData(),
		HasRawData:   true, // CRDUMP_MASK_DATA needs binary data
	}
	
	if s.Header != nil {
		sectionJSON.CRDumpMaskData = &types.CRDumpMaskDataJSON{
			Version:  s.Header.Version,
			MaskSize: s.Header.MaskSize,
			DataSize: len(s.Data),
		}
	}
	
	return json.Marshal(sectionJSON)
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
	sectionJSON := &types.SectionJSON{
		Type:         s.Type(),
		TypeName:     s.TypeName(),
		Offset:       s.Offset(),
		Size:         s.Size(),
		CRCType:      s.CRCType().String(),
		IsEncrypted:  s.IsEncrypted(),
		IsDeviceData: s.IsDeviceData(),
		HasRawData:   true, // FW_INTERNAL_USAGE needs binary data
	}
	
	if s.Header != nil {
		sectionJSON.FWInternalUsage = &types.FWInternalUsageJSON{
			Version:  s.Header.Version,
			Size:     s.Header.Size,
			Type:     s.Header.Type,
			DataSize: len(s.Data),
		}
	}
	
	return json.Marshal(sectionJSON)
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
	sectionJSON := &types.SectionJSON{
		Type:         s.Type(),
		TypeName:     s.TypeName(),
		Offset:       s.Offset(),
		Size:         s.Size(),
		CRCType:      s.CRCType().String(),
		IsEncrypted:  s.IsEncrypted(),
		IsDeviceData: s.IsDeviceData(),
		HasRawData:   true, // PROGRAMMABLE_HW_FW needs binary data
	}
	
	if s.Header != nil {
		sectionJSON.ProgrammableHWFW = &types.ProgrammableHWFWJSON{
			Version:     s.Header.Version,
			HWType:      s.Header.HWType,
			FWSize:      s.Header.FWSize,
			Checksum:    s.Header.Checksum,
			LoadAddress: fmt.Sprintf("0x%08X", s.Header.LoadAddress),
			EntryPoint:  fmt.Sprintf("0x%08X", s.Header.EntryPoint),
			DataSize:    len(s.Data),
		}
	}
	
	return json.Marshal(sectionJSON)
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
	sectionJSON := &types.SectionJSON{
		Type:         s.Type(),
		TypeName:     s.TypeName(),
		Offset:       s.Offset(),
		Size:         s.Size(),
		CRCType:      s.CRCType().String(),
		IsEncrypted:  s.IsEncrypted(),
		IsDeviceData: s.IsDeviceData(),
		HasRawData:   true, // DIGITAL_CERT_PTR needs binary data due to variable structure
	}
	
	if s.CertPtr != nil {
		sectionJSON.DigitalCertPtr = &types.DigitalCertPtrJSON{
			CertType:   s.CertPtr.CertType,
			CertOffset: fmt.Sprintf("0x%08X", s.CertPtr.CertOffset),
			CertSize:   s.CertPtr.CertSize,
		}
	}
	
	return json.Marshal(sectionJSON)
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
	sectionJSON := &types.SectionJSON{
		Type:         s.Type(),
		TypeName:     s.TypeName(),
		Offset:       s.Offset(),
		Size:         s.Size(),
		CRCType:      s.CRCType().String(),
		IsEncrypted:  s.IsEncrypted(),
		IsDeviceData: s.IsDeviceData(),
		HasRawData:   true, // DIGITAL_CERT_RW needs binary data for certificate
	}
	
	if s.Cert != nil {
		// Find actual certificate size
		certSize := 0
		for i := 0; i < len(s.Cert.Certificate); i++ {
			if s.Cert.Certificate[i] != 0 {
				certSize = i + 1
			}
		}
		
		digitalCertRW := &types.DigitalCertRWJSON{
			CertType:       s.Cert.CertType,
			CertSize:       s.Cert.CertSize,
			ValidFrom:      s.Cert.ValidFrom,
			ValidTo:        s.Cert.ValidTo,
			ActualCertSize: certSize,
		}
		
		if certSize > 0 && certSize <= 256 { // Include small cert preview
			digitalCertRW.CertPreview = hex.EncodeToString(s.Cert.Certificate[:certSize])
		}
		
		sectionJSON.DigitalCertRW = digitalCertRW
	}
	
	return json.Marshal(sectionJSON)
}