package types

import (
	"reflect"

	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// SecureBootSignaturesAnnotated represents the secure boot signatures structure with annotations
type SecureBootSignaturesAnnotated struct {
	BootSignature         [128]uint32 `offset:"0x0,endian:be"`   // offset 0x0 - boot signature of data
	CriticalSignature     [128]uint32 `offset:"0x200,endian:be"` // offset 0x200 - fw critical signature of itcos
	NonCriticalSignature  [128]uint32 `offset:"0x400,endian:be"` // offset 0x400 - fw non critical signatures
}

// Unmarshal unmarshals binary data
func (s *SecureBootSignaturesAnnotated) Unmarshal(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*s))
	if err != nil {
		return err
	}
	return annotations.Unmarshal(data, s, annot)
}

// Marshal marshals to binary data
func (s *SecureBootSignaturesAnnotated) Marshal() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*s))
	if err != nil {
		return nil, err
	}
	return annotations.Marshal(s, annot)
}

// FromAnnotated converts from annotated to legacy format
func (s *SecureBootSignaturesAnnotated) FromAnnotated() *SecureBootSignatures {
	return &SecureBootSignatures{
		BootSignature:        s.BootSignature,
		CriticalSignature:    s.CriticalSignature,
		NonCriticalSignature: s.NonCriticalSignature,
	}
}

// ToAnnotated converts SecureBootSignatures to annotated format
func (s *SecureBootSignatures) ToAnnotated() *SecureBootSignaturesAnnotated {
	return &SecureBootSignaturesAnnotated{
		BootSignature:        s.BootSignature,
		CriticalSignature:    s.CriticalSignature,
		NonCriticalSignature: s.NonCriticalSignature,
	}
}

// ResetCapabilitiesAnnotated represents reset capabilities with annotations
type ResetCapabilitiesAnnotated struct {
	ResetVerEn        uint8  `offset:"bit:0,len:1,endian:be"`   // bit 0
	Reserved1         uint8  `offset:"bit:1,len:7,endian:be"`   // bits 1-7
	VersionVectorVer  uint8  `offset:"byte:1,endian:be"`        // bits 8-15
	Reserved2         uint16 `offset:"byte:2,endian:be"`        // bits 16-31
}

// Unmarshal unmarshals binary data
func (r *ResetCapabilitiesAnnotated) Unmarshal(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*r))
	if err != nil {
		return err
	}
	return annotations.Unmarshal(data, r, annot)
}

// Marshal marshals to binary data
func (r *ResetCapabilitiesAnnotated) Marshal() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*r))
	if err != nil {
		return nil, err
	}
	return annotations.Marshal(r, annot)
}

// FromAnnotated converts from annotated to legacy format
func (r *ResetCapabilitiesAnnotated) FromAnnotated() *ResetCapabilities {
	return &ResetCapabilities{
		ResetVerEn:       r.ResetVerEn,
		Reserved1:        r.Reserved1,
		VersionVectorVer: r.VersionVectorVer,
		Reserved2:        r.Reserved2,
	}
}

// ToAnnotated converts ResetCapabilities to annotated format
func (r *ResetCapabilities) ToAnnotated() *ResetCapabilitiesAnnotated {
	return &ResetCapabilitiesAnnotated{
		ResetVerEn:       r.ResetVerEn,
		Reserved1:        r.Reserved1,
		VersionVectorVer: r.VersionVectorVer,
		Reserved2:        r.Reserved2,
	}
}

// ResetVersionAnnotated represents reset version structure with annotations
type ResetVersionAnnotated struct {
	Major    uint16 `offset:"bit:0,len:16,endian:be"`  // bits 0-15
	Branch   uint8  `offset:"bit:16,len:4,endian:be"`  // bits 16-19
	Minor    uint8  `offset:"bit:20,len:8,endian:be"`  // bits 20-27
	Reserved uint8  `offset:"bit:28,len:4,endian:be"`  // bits 28-31
}

// Unmarshal unmarshals binary data
func (r *ResetVersionAnnotated) Unmarshal(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*r))
	if err != nil {
		return err
	}
	return annotations.Unmarshal(data, r, annot)
}

// Marshal marshals to binary data
func (r *ResetVersionAnnotated) Marshal() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*r))
	if err != nil {
		return nil, err
	}
	return annotations.Marshal(r, annot)
}

// FromAnnotated converts from annotated to legacy format
func (r *ResetVersionAnnotated) FromAnnotated() *ResetVersion {
	return &ResetVersion{
		Major:    r.Major,
		Branch:   r.Branch,
		Minor:    r.Minor,
		Reserved: r.Reserved,
	}
}

// ToAnnotated converts ResetVersion to annotated format
func (r *ResetVersion) ToAnnotated() *ResetVersionAnnotated {
	return &ResetVersionAnnotated{
		Major:    r.Major,
		Branch:   r.Branch,
		Minor:    r.Minor,
		Reserved: r.Reserved,
	}
}

// VersionVectorAnnotated represents the version vector structure with annotations
type VersionVectorAnnotated struct {
	ResetCapabilities ResetCapabilitiesAnnotated `offset:"0x0"`  // offset 0x0
	Scratchpad        ResetVersionAnnotated      `offset:"0x4"`  // offset 0x4 - cores SP
	ICMContext        ResetVersionAnnotated      `offset:"0x8"`  // offset 0x8
	PCI               ResetVersionAnnotated      `offset:"0xc"`  // offset 0xc - PCI domain
	PHY               ResetVersionAnnotated      `offset:"0x10"` // offset 0x10 - PHY domain
	INI               ResetVersionAnnotated      `offset:"0x14"` // offset 0x14
	Reserved1         ResetVersionAnnotated      `offset:"0x18"` // offset 0x18
	Reserved2         ResetVersionAnnotated      `offset:"0x1c"` // offset 0x1c
	Reserved3         ResetVersionAnnotated      `offset:"0x20"` // offset 0x20
	Reserved4         ResetVersionAnnotated      `offset:"0x24"` // offset 0x24
	Reserved5         ResetVersionAnnotated      `offset:"0x28"` // offset 0x28
	Reserved6         ResetVersionAnnotated      `offset:"0x2c"` // offset 0x2c
	Reserved7         ResetVersionAnnotated      `offset:"0x30"` // offset 0x30 - total 52 bytes
}

// Unmarshal unmarshals binary data
func (v *VersionVectorAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, v)
}

// Marshal marshals to binary data
func (v *VersionVectorAnnotated) Marshal() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*v))
	if err != nil {
		return nil, err
	}
	return annotations.Marshal(v, annot)
}

// FromAnnotated converts from annotated to legacy format
func (v *VersionVectorAnnotated) FromAnnotated() *VersionVector {
	return &VersionVector{
		ResetCapabilities: *v.ResetCapabilities.FromAnnotated(),
		Scratchpad:        *v.Scratchpad.FromAnnotated(),
		ICMContext:        *v.ICMContext.FromAnnotated(),
		PCI:               *v.PCI.FromAnnotated(),
		PHY:               *v.PHY.FromAnnotated(),
		INI:               *v.INI.FromAnnotated(),
		Reserved1:         *v.Reserved1.FromAnnotated(),
		Reserved2:         *v.Reserved2.FromAnnotated(),
		Reserved3:         *v.Reserved3.FromAnnotated(),
		Reserved4:         *v.Reserved4.FromAnnotated(),
		Reserved5:         *v.Reserved5.FromAnnotated(),
		Reserved6:         *v.Reserved6.FromAnnotated(),
		Reserved7:         *v.Reserved7.FromAnnotated(),
	}
}

// ToAnnotated converts VersionVector to annotated format
func (v *VersionVector) ToAnnotated() *VersionVectorAnnotated {
	return &VersionVectorAnnotated{
		ResetCapabilities: *v.ResetCapabilities.ToAnnotated(),
		Scratchpad:        *v.Scratchpad.ToAnnotated(),
		ICMContext:        *v.ICMContext.ToAnnotated(),
		PCI:               *v.PCI.ToAnnotated(),
		PHY:               *v.PHY.ToAnnotated(),
		INI:               *v.INI.ToAnnotated(),
		Reserved1:         *v.Reserved1.ToAnnotated(),
		Reserved2:         *v.Reserved2.ToAnnotated(),
		Reserved3:         *v.Reserved3.ToAnnotated(),
		Reserved4:         *v.Reserved4.ToAnnotated(),
		Reserved5:         *v.Reserved5.ToAnnotated(),
		Reserved6:         *v.Reserved6.ToAnnotated(),
		Reserved7:         *v.Reserved7.ToAnnotated(),
	}
}

// ResetInfoAnnotated represents the RESET_INFO section structure with annotations
type ResetInfoAnnotated struct {
	Data [256]uint8 `offset:"0x0"` // Full 256-byte section
}

// Unmarshal unmarshals binary data
func (r *ResetInfoAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, r)
}

// Marshal marshals to binary data
func (r *ResetInfoAnnotated) Marshal() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*r))
	if err != nil {
		return nil, err
	}
	return annotations.Marshal(r, annot)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (r *ResetInfoAnnotated) UnmarshalWithReserved(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*r))
	if err != nil {
		return err
	}
	opts := &annotations.UnmarshalOptions{
		IncludeReserved: true,
	}
	return annotations.UnmarshalWithOptions(data, r, annot, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (r *ResetInfoAnnotated) MarshalWithReserved() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*r))
	if err != nil {
		return nil, err
	}
	opts := &annotations.MarshalOptions{
		IncludeReserved: true,
	}
	return annotations.MarshalWithOptions(r, annot, opts)
}

// FromAnnotated converts from annotated to legacy format
func (r *ResetInfoAnnotated) FromAnnotated() *ResetInfo {
	// Parse the VersionVector from the first 52 bytes
	vv := &VersionVectorAnnotated{}
	_ = vv.Unmarshal(r.Data[:52])
	
	// Copy reserved bytes
	var reserved [204]uint8
	copy(reserved[:], r.Data[52:])
	
	return &ResetInfo{
		VersionVector: *vv.FromAnnotated(),
		Reserved:      reserved,
	}
}

// ToAnnotated converts ResetInfo to annotated format
func (r *ResetInfo) ToAnnotated() *ResetInfoAnnotated {
	result := &ResetInfoAnnotated{}
	
	// Marshal VersionVector to first 52 bytes
	vvData, _ := r.VersionVector.ToAnnotated().Marshal()
	copy(result.Data[:52], vvData)
	
	// Copy reserved bytes
	copy(result.Data[52:], r.Reserved[:])
	
	return result
}

// HMACDigestAnnotated represents the HMAC_DIGEST section with annotations
type HMACDigestAnnotated struct {
	DigestType uint32    `offset:"0x0,endian:be"`  // Type of digest (SHA256, SHA512, etc)
	DigestSize uint32    `offset:"0x4,endian:be"`  // Size of the digest in bytes
	Digest     [64]uint8 `offset:"0x8"`            // The actual digest (max 512 bits)
	Reserved   [56]uint8 `offset:"0x48"`           // Padding
}

// Unmarshal unmarshals binary data
func (h *HMACDigestAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// Marshal marshals to binary data
func (h *HMACDigestAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (h *HMACDigestAnnotated) UnmarshalWithReserved(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*h))
	if err != nil {
		return err
	}
	opts := &annotations.UnmarshalOptions{
		IncludeReserved: true,
	}
	return annotations.UnmarshalWithOptions(data, h, annot, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (h *HMACDigestAnnotated) MarshalWithReserved() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*h))
	if err != nil {
		return nil, err
	}
	opts := &annotations.MarshalOptions{
		IncludeReserved: true,
	}
	return annotations.MarshalWithOptions(h, annot, opts)
}

// FromAnnotated converts from annotated to legacy format
func (h *HMACDigestAnnotated) FromAnnotated() *HMACDigest {
	return &HMACDigest{
		DigestType: h.DigestType,
		DigestSize: h.DigestSize,
		Digest:     h.Digest,
		Reserved:   h.Reserved,
	}
}

// ToAnnotated converts HMACDigest to annotated format
func (h *HMACDigest) ToAnnotated() *HMACDigestAnnotated {
	return &HMACDigestAnnotated{
		DigestType: h.DigestType,
		DigestSize: h.DigestSize,
		Digest:     h.Digest,
		Reserved:   h.Reserved,
	}
}

// RSAPublicKeyAnnotated represents the RSA_PUBLIC_KEY section with annotations
type RSAPublicKeyAnnotated struct {
	KeyType     uint32     `offset:"0x0,endian:be"`   // RSA key type (2048, 4096, etc)
	KeySize     uint32     `offset:"0x4,endian:be"`   // Key size in bytes
	Exponent    uint32     `offset:"0x8,endian:be"`   // Public exponent (usually 65537)
	Reserved1   uint32     `offset:"0xc,endian:be"`   // Reserved
	Modulus     [512]uint8 `offset:"0x10"`            // RSA modulus (up to 4096 bits)
	Reserved2   [496]uint8 `offset:"0x210"`           // Padding to 1024 bytes
}

// Unmarshal unmarshals binary data
func (r *RSAPublicKeyAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, r)
}

// Marshal marshals to binary data
func (r *RSAPublicKeyAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(r)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (r *RSAPublicKeyAnnotated) UnmarshalWithReserved(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*r))
	if err != nil {
		return err
	}
	opts := &annotations.UnmarshalOptions{
		IncludeReserved: true,
	}
	return annotations.UnmarshalWithOptions(data, r, annot, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (r *RSAPublicKeyAnnotated) MarshalWithReserved() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*r))
	if err != nil {
		return nil, err
	}
	opts := &annotations.MarshalOptions{
		IncludeReserved: true,
	}
	return annotations.MarshalWithOptions(r, annot, opts)
}

// FromAnnotated converts from annotated to legacy format
func (r *RSAPublicKeyAnnotated) FromAnnotated() *RSAPublicKey {
	return &RSAPublicKey{
		KeyType:   r.KeyType,
		KeySize:   r.KeySize,
		Exponent:  r.Exponent,
		Reserved1: r.Reserved1,
		Modulus:   r.Modulus,
		Reserved2: r.Reserved2,
	}
}

// ToAnnotated converts RSAPublicKey to annotated format
func (r *RSAPublicKey) ToAnnotated() *RSAPublicKeyAnnotated {
	return &RSAPublicKeyAnnotated{
		KeyType:   r.KeyType,
		KeySize:   r.KeySize,
		Exponent:  r.Exponent,
		Reserved1: r.Reserved1,
		Modulus:   r.Modulus,
		Reserved2: r.Reserved2,
	}
}

// DBGFwIniAnnotated represents the DBG_FW_INI section with annotations
type DBGFwIniAnnotated struct {
	CompressionMethod uint32 `offset:"0x0,endian:be"` // 0=Uncompressed, 1=Zlib, 2=LZMA
	UncompressedSize  uint32 `offset:"0x4,endian:be"` // Size when uncompressed
	CompressedSize    uint32 `offset:"0x8,endian:be"` // Size of compressed data
	Reserved          uint32 `offset:"0xc,endian:be"` // Reserved
	// Data follows - variable size
}

// Unmarshal unmarshals binary data
func (d *DBGFwIniAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, d)
}

// Marshal marshals to binary data
func (d *DBGFwIniAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(d)
}

// FromAnnotated converts from annotated to legacy format
func (d *DBGFwIniAnnotated) FromAnnotated() *DBGFwIni {
	return &DBGFwIni{
		CompressionMethod: d.CompressionMethod,
		UncompressedSize:  d.UncompressedSize,
		CompressedSize:    d.CompressedSize,
		Reserved:          d.Reserved,
	}
}

// ToAnnotated converts DBGFwIni to annotated format
func (d *DBGFwIni) ToAnnotated() *DBGFwIniAnnotated {
	return &DBGFwIniAnnotated{
		CompressionMethod: d.CompressionMethod,
		UncompressedSize:  d.UncompressedSize,
		CompressedSize:    d.CompressedSize,
		Reserved:          d.Reserved,
	}
}

// DBGFwParamsAnnotated represents the DBG_FW_PARAMS section with annotations
type DBGFwParamsAnnotated struct {
	CompressionMethod uint32 `offset:"0x0,endian:be"` // 0=Uncompressed, 1=Zlib, 2=LZMA
	UncompressedSize  uint32 `offset:"0x4,endian:be"` // Size when uncompressed
	CompressedSize    uint32 `offset:"0x8,endian:be"` // Size of compressed data
	Reserved          uint32 `offset:"0xc,endian:be"` // Reserved
	// Data follows - variable size
}

// Unmarshal unmarshals binary data
func (d *DBGFwParamsAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, d)
}

// Marshal marshals to binary data
func (d *DBGFwParamsAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(d)
}

// FromAnnotated converts from annotated to legacy format
func (d *DBGFwParamsAnnotated) FromAnnotated() *DBGFwParams {
	return &DBGFwParams{
		CompressionMethod: d.CompressionMethod,
		UncompressedSize:  d.UncompressedSize,
		CompressedSize:    d.CompressedSize,
		Reserved:          d.Reserved,
	}
}

// ToAnnotated converts DBGFwParams to annotated format
func (d *DBGFwParams) ToAnnotated() *DBGFwParamsAnnotated {
	return &DBGFwParamsAnnotated{
		CompressionMethod: d.CompressionMethod,
		UncompressedSize:  d.UncompressedSize,
		CompressedSize:    d.CompressedSize,
		Reserved:          d.Reserved,
	}
}

// FWAdbAnnotated represents the FW_ADB section with annotations
type FWAdbAnnotated struct {
	Version    uint32 `offset:"0x0,endian:be"`  // ADB version
	Size       uint32 `offset:"0x4,endian:be"`  // Size of ADB data
	Reserved1  uint32 `offset:"0x8,endian:be"`  // Reserved
	Reserved2  uint32 `offset:"0xc,endian:be"`  // Reserved
	// ADB data follows - variable size
}

// Unmarshal unmarshals binary data
func (f *FWAdbAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, f)
}

// Marshal marshals to binary data
func (f *FWAdbAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(f)
}

// FromAnnotated converts from annotated to legacy format
func (f *FWAdbAnnotated) FromAnnotated() *FWAdb {
	return &FWAdb{
		Version:   f.Version,
		Size:      f.Size,
		Reserved1: f.Reserved1,
		Reserved2: f.Reserved2,
	}
}

// ToAnnotated converts FWAdb to annotated format
func (f *FWAdb) ToAnnotated() *FWAdbAnnotated {
	return &FWAdbAnnotated{
		Version:   f.Version,
		Size:      f.Size,
		Reserved1: f.Reserved1,
		Reserved2: f.Reserved2,
	}
}

// CRDumpMaskDataAnnotated represents the CRDUMP_MASK_DATA section with annotations
type CRDumpMaskDataAnnotated struct {
	Version  uint32    `offset:"0x0,endian:be"`  // Version of the mask data format
	MaskSize uint32    `offset:"0x4,endian:be"`  // Size of the mask data
	Reserved [8]uint32 `offset:"0x8,endian:be"`  // Reserved
	// Mask data follows - variable size
}

// Unmarshal unmarshals binary data
func (c *CRDumpMaskDataAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, c)
}

// Marshal marshals to binary data
func (c *CRDumpMaskDataAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(c)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (c *CRDumpMaskDataAnnotated) UnmarshalWithReserved(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*c))
	if err != nil {
		return err
	}
	opts := &annotations.UnmarshalOptions{
		IncludeReserved: true,
	}
	return annotations.UnmarshalWithOptions(data, c, annot, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (c *CRDumpMaskDataAnnotated) MarshalWithReserved() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*c))
	if err != nil {
		return nil, err
	}
	opts := &annotations.MarshalOptions{
		IncludeReserved: true,
	}
	return annotations.MarshalWithOptions(c, annot, opts)
}

// FromAnnotated converts from annotated to legacy format
func (c *CRDumpMaskDataAnnotated) FromAnnotated() *CRDumpMaskData {
	return &CRDumpMaskData{
		Version:  c.Version,
		MaskSize: c.MaskSize,
		Reserved: c.Reserved,
	}
}

// ToAnnotated converts CRDumpMaskData to annotated format
func (c *CRDumpMaskData) ToAnnotated() *CRDumpMaskDataAnnotated {
	return &CRDumpMaskDataAnnotated{
		Version:  c.Version,
		MaskSize: c.MaskSize,
		Reserved: c.Reserved,
	}
}

// ForbiddenVersionsAnnotated represents the FORBIDDEN_VERSIONS section with annotations
type ForbiddenVersionsAnnotated struct {
	Count    uint32   `offset:"byte:0,endian:be" json:"count"`                           // Number of forbidden versions
	Reserved uint32   `offset:"byte:4,endian:be,reserved:true" json:"reserved"`             // Reserved for alignment
	Versions []uint32 `offset:"byte:8,endian:be,list_size:Count" json:"versions,omitempty"`          // List of forbidden version numbers
}

// Unmarshal unmarshals binary data
func (f *ForbiddenVersionsAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, f)
}

// Marshal marshals to binary data
func (f *ForbiddenVersionsAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(f)
}


// FWNVLogAnnotated represents the FW_NV_LOG section with annotations
type FWNVLogAnnotated struct {
	LogVersion uint32     `offset:"0x0,endian:be"`            // Log format version
	LogSize    uint32     `offset:"0x4,endian:be"`            // Size of log data
	EntryCount uint32     `offset:"0x8,endian:be"`            // Number of log entries
	Reserved   [13]uint32 `offset:"0xc,endian:be,reserved:true"` // Reserved
}

// Unmarshal unmarshals binary data
func (f *FWNVLogAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, f)
}

// Marshal marshals to binary data
func (f *FWNVLogAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(f)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (f *FWNVLogAnnotated) UnmarshalWithReserved(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*f))
	if err != nil {
		return err
	}
	opts := &annotations.UnmarshalOptions{
		IncludeReserved: true,
	}
	return annotations.UnmarshalWithOptions(data, f, annot, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (f *FWNVLogAnnotated) MarshalWithReserved() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*f))
	if err != nil {
		return nil, err
	}
	opts := &annotations.MarshalOptions{
		IncludeReserved: true,
	}
	return annotations.MarshalWithOptions(f, annot, opts)
}

// FromAnnotated converts from annotated to legacy format
func (f *FWNVLogAnnotated) FromAnnotated() *FWNVLog {
	return &FWNVLog{
		LogVersion: f.LogVersion,
		LogSize:    f.LogSize,
		EntryCount: f.EntryCount,
		Reserved:   f.Reserved,
	}
}

// ToAnnotated converts FWNVLog to annotated format
func (f *FWNVLog) ToAnnotated() *FWNVLogAnnotated {
	return &FWNVLogAnnotated{
		LogVersion: f.LogVersion,
		LogSize:    f.LogSize,
		EntryCount: f.EntryCount,
		Reserved:   f.Reserved,
	}
}

// VPD_R0Annotated represents the VPD_R0 section (Vital Product Data) with annotations
type VPD_R0Annotated struct {
	ID       [2]uint8  `offset:"0x0"`                   // Should be "VPD_R0"
	Length   uint16    `offset:"0x2,endian:be"`         // Length of VPD data
	Reserved [60]uint8 `offset:"0x4,reserved:true"`     // Reserved/padding
}

// Unmarshal unmarshals binary data
func (v *VPD_R0Annotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, v)
}

// Marshal marshals to binary data
func (v *VPD_R0Annotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(v)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (v *VPD_R0Annotated) UnmarshalWithReserved(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*v))
	if err != nil {
		return err
	}
	opts := &annotations.UnmarshalOptions{
		IncludeReserved: true,
	}
	return annotations.UnmarshalWithOptions(data, v, annot, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (v *VPD_R0Annotated) MarshalWithReserved() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*v))
	if err != nil {
		return nil, err
	}
	opts := &annotations.MarshalOptions{
		IncludeReserved: true,
	}
	return annotations.MarshalWithOptions(v, annot, opts)
}

// FromAnnotated converts from annotated to legacy format
func (v *VPD_R0Annotated) FromAnnotated() *VPD_R0 {
	return &VPD_R0{
		ID:       v.ID,
		Length:   v.Length,
		Reserved: v.Reserved,
	}
}

// ToAnnotated converts VPD_R0 to annotated format
func (v *VPD_R0) ToAnnotated() *VPD_R0Annotated {
	return &VPD_R0Annotated{
		ID:       v.ID,
		Length:   v.Length,
		Reserved: v.Reserved,
	}
}

// NVDataAnnotated represents the NV_DATA sections (NV_DATA0, NV_DATA1, NV_DATA2) with annotations
type NVDataAnnotated struct {
	Version   uint32    `offset:"0x0,endian:be"`            // NV data version
	DataSize  uint32    `offset:"0x4,endian:be"`            // Size of NV data
	Reserved1 uint32    `offset:"0x8,endian:be,reserved:true"`  // Reserved
	Reserved2 uint32    `offset:"0xc,endian:be,reserved:true"` // Reserved
	Reserved3 [48]uint8 `offset:"0x10,reserved:true"`      // Reserved padding
}

// Unmarshal unmarshals binary data
func (n *NVDataAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, n)
}

// Marshal marshals to binary data
func (n *NVDataAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(n)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (n *NVDataAnnotated) UnmarshalWithReserved(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*n))
	if err != nil {
		return err
	}
	opts := &annotations.UnmarshalOptions{
		IncludeReserved: true,
	}
	return annotations.UnmarshalWithOptions(data, n, annot, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (n *NVDataAnnotated) MarshalWithReserved() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*n))
	if err != nil {
		return nil, err
	}
	opts := &annotations.MarshalOptions{
		IncludeReserved: true,
	}
	return annotations.MarshalWithOptions(n, annot, opts)
}

// FromAnnotated converts from annotated to legacy format
func (n *NVDataAnnotated) FromAnnotated() *NVData {
	return &NVData{
		Version:   n.Version,
		DataSize:  n.DataSize,
		Reserved1: n.Reserved1,
		Reserved2: n.Reserved2,
		Reserved3: n.Reserved3,
	}
}

// ToAnnotated converts NVData to annotated format
func (n *NVData) ToAnnotated() *NVDataAnnotated {
	return &NVDataAnnotated{
		Version:   n.Version,
		DataSize:  n.DataSize,
		Reserved1: n.Reserved1,
		Reserved2: n.Reserved2,
		Reserved3: n.Reserved3,
	}
}

// FWInternalUsageAnnotated represents the FW_INTERNAL_USAGE section with annotations
type FWInternalUsageAnnotated struct {
	Version  uint32    `offset:"0x0,endian:be"`         // Version
	Size     uint32    `offset:"0x4,endian:be"`         // Data size
	Type     uint32    `offset:"0x8,endian:be"`         // Usage type
	Reserved [52]uint8 `offset:"0xc,reserved:true"`     // Reserved
}

// Unmarshal unmarshals binary data
func (f *FWInternalUsageAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, f)
}

// Marshal marshals to binary data
func (f *FWInternalUsageAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(f)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (f *FWInternalUsageAnnotated) UnmarshalWithReserved(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*f))
	if err != nil {
		return err
	}
	opts := &annotations.UnmarshalOptions{
		IncludeReserved: true,
	}
	return annotations.UnmarshalWithOptions(data, f, annot, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (f *FWInternalUsageAnnotated) MarshalWithReserved() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*f))
	if err != nil {
		return nil, err
	}
	opts := &annotations.MarshalOptions{
		IncludeReserved: true,
	}
	return annotations.MarshalWithOptions(f, annot, opts)
}

// FromAnnotated converts from annotated to legacy format
func (f *FWInternalUsageAnnotated) FromAnnotated() *FWInternalUsage {
	return &FWInternalUsage{
		Version:  f.Version,
		Size:     f.Size,
		Type:     f.Type,
		Reserved: f.Reserved,
	}
}

// ToAnnotated converts FWInternalUsage to annotated format
func (f *FWInternalUsage) ToAnnotated() *FWInternalUsageAnnotated {
	return &FWInternalUsageAnnotated{
		Version:  f.Version,
		Size:     f.Size,
		Type:     f.Type,
		Reserved: f.Reserved,
	}
}

// ProgrammableHWFWAnnotated represents PROGRAMMABLE_HW_FW sections with annotations
type ProgrammableHWFWAnnotated struct {
	Version     uint32    `offset:"0x0,endian:be"`      // Firmware version
	HWType      uint32    `offset:"0x4,endian:be"`      // Hardware type
	FWSize      uint32    `offset:"0x8,endian:be"`      // Firmware size
	Checksum    uint32    `offset:"0xc,endian:be"`      // Firmware checksum
	LoadAddress uint32    `offset:"0x10,endian:be"`     // Load address in hardware
	EntryPoint  uint32    `offset:"0x14,endian:be"`     // Entry point address
	Reserved    [40]uint8 `offset:"0x18,reserved:true"` // Reserved
}

// Unmarshal unmarshals binary data
func (p *ProgrammableHWFWAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, p)
}

// Marshal marshals to binary data
func (p *ProgrammableHWFWAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(p)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (p *ProgrammableHWFWAnnotated) UnmarshalWithReserved(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*p))
	if err != nil {
		return err
	}
	opts := &annotations.UnmarshalOptions{
		IncludeReserved: true,
	}
	return annotations.UnmarshalWithOptions(data, p, annot, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (p *ProgrammableHWFWAnnotated) MarshalWithReserved() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*p))
	if err != nil {
		return nil, err
	}
	opts := &annotations.MarshalOptions{
		IncludeReserved: true,
	}
	return annotations.MarshalWithOptions(p, annot, opts)
}

// FromAnnotated converts from annotated to legacy format
func (p *ProgrammableHWFWAnnotated) FromAnnotated() *ProgrammableHWFW {
	return &ProgrammableHWFW{
		Version:     p.Version,
		HWType:      p.HWType,
		FWSize:      p.FWSize,
		Checksum:    p.Checksum,
		LoadAddress: p.LoadAddress,
		EntryPoint:  p.EntryPoint,
		Reserved:    p.Reserved,
	}
}

// ToAnnotated converts ProgrammableHWFW to annotated format
func (p *ProgrammableHWFW) ToAnnotated() *ProgrammableHWFWAnnotated {
	return &ProgrammableHWFWAnnotated{
		Version:     p.Version,
		HWType:      p.HWType,
		FWSize:      p.FWSize,
		Checksum:    p.Checksum,
		LoadAddress: p.LoadAddress,
		EntryPoint:  p.EntryPoint,
		Reserved:    p.Reserved,
	}
}

// DigitalCertPtrAnnotated represents DIGITAL_CERT_PTR section with annotations
type DigitalCertPtrAnnotated struct {
	CertType   uint32    `offset:"0x0,endian:be"`      // Certificate type
	CertOffset uint32    `offset:"0x4,endian:be"`      // Offset to certificate
	CertSize   uint32    `offset:"0x8,endian:be"`      // Certificate size
	Reserved   [28]uint8 `offset:"0xc,reserved:true"`  // Reserved (28 bytes to make total 40)
}

// Unmarshal unmarshals binary data
func (d *DigitalCertPtrAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, d)
}

// Marshal marshals to binary data
func (d *DigitalCertPtrAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(d)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (d *DigitalCertPtrAnnotated) UnmarshalWithReserved(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*d))
	if err != nil {
		return err
	}
	opts := &annotations.UnmarshalOptions{
		IncludeReserved: true,
	}
	return annotations.UnmarshalWithOptions(data, d, annot, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (d *DigitalCertPtrAnnotated) MarshalWithReserved() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*d))
	if err != nil {
		return nil, err
	}
	opts := &annotations.MarshalOptions{
		IncludeReserved: true,
	}
	return annotations.MarshalWithOptions(d, annot, opts)
}

// FromAnnotated converts from annotated to legacy format
func (d *DigitalCertPtrAnnotated) FromAnnotated() *DigitalCertPtr {
	return &DigitalCertPtr{
		CertType:   d.CertType,
		CertOffset: d.CertOffset,
		CertSize:   d.CertSize,
		Reserved:   d.Reserved,
	}
}

// ToAnnotated converts DigitalCertPtr to annotated format
func (d *DigitalCertPtr) ToAnnotated() *DigitalCertPtrAnnotated {
	return &DigitalCertPtrAnnotated{
		CertType:   d.CertType,
		CertOffset: d.CertOffset,
		CertSize:   d.CertSize,
		Reserved:   d.Reserved,
	}
}

// DigitalCertRWAnnotated represents DIGITAL_CERT_RW section with annotations
type DigitalCertRWAnnotated struct {
	CertType    uint32      `offset:"0x0,endian:be"`      // Certificate type
	CertSize    uint32      `offset:"0x4,endian:be"`      // Certificate size
	ValidFrom   uint64      `offset:"0x8,endian:be"`      // Validity start timestamp
	ValidTo     uint64      `offset:"0x10,endian:be"`     // Validity end timestamp
	Reserved    [32]uint8   `offset:"0x18,reserved:true"` // Reserved
	Certificate [4096]uint8 `offset:"0x38"`               // Certificate data (up to 4K)
}

// Unmarshal unmarshals binary data
func (d *DigitalCertRWAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, d)
}

// Marshal marshals to binary data
func (d *DigitalCertRWAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(d)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (d *DigitalCertRWAnnotated) UnmarshalWithReserved(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*d))
	if err != nil {
		return err
	}
	opts := &annotations.UnmarshalOptions{
		IncludeReserved: true,
	}
	return annotations.UnmarshalWithOptions(data, d, annot, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (d *DigitalCertRWAnnotated) MarshalWithReserved() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*d))
	if err != nil {
		return nil, err
	}
	opts := &annotations.MarshalOptions{
		IncludeReserved: true,
	}
	return annotations.MarshalWithOptions(d, annot, opts)
}

// FromAnnotated converts from annotated to legacy format
func (d *DigitalCertRWAnnotated) FromAnnotated() *DigitalCertRW {
	return &DigitalCertRW{
		CertType:    d.CertType,
		CertSize:    d.CertSize,
		ValidFrom:   d.ValidFrom,
		ValidTo:     d.ValidTo,
		Reserved:    d.Reserved,
		Certificate: d.Certificate,
	}
}

// ToAnnotated converts DigitalCertRW to annotated format
func (d *DigitalCertRW) ToAnnotated() *DigitalCertRWAnnotated {
	return &DigitalCertRWAnnotated{
		CertType:    d.CertType,
		CertSize:    d.CertSize,
		ValidFrom:   d.ValidFrom,
		ValidTo:     d.ValidTo,
		Reserved:    d.Reserved,
		Certificate: d.Certificate,
	}
}

// CertChainAnnotated represents CERT_CHAIN_0 section with annotations
type CertChainAnnotated struct {
	ChainLength uint32    `offset:"0x0,endian:be"`     // Number of certificates in chain
	Reserved    [12]uint8 `offset:"0x4,reserved:true"` // Reserved
}

// Unmarshal unmarshals binary data
func (c *CertChainAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, c)
}

// Marshal marshals to binary data
func (c *CertChainAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(c)
}

// FromAnnotated converts from annotated to legacy format
func (c *CertChainAnnotated) FromAnnotated() *CertChain {
	return &CertChain{
		ChainLength: c.ChainLength,
		Reserved:    c.Reserved,
	}
}

// ToAnnotated converts CertChain to annotated format
func (c *CertChain) ToAnnotated() *CertChainAnnotated {
	return &CertChainAnnotated{
		ChainLength: c.ChainLength,
		Reserved:    c.Reserved,
	}
}

// RootCertificatesAnnotated represents ROOT_CERTIFICATES sections with annotations
type RootCertificatesAnnotated struct {
	Version   uint32    `offset:"0x0,endian:be"`     // Version
	CertCount uint32    `offset:"0x4,endian:be"`     // Number of certificates
	Reserved  [8]uint8  `offset:"0x8,reserved:true"` // Reserved
}

// Unmarshal unmarshals binary data
func (r *RootCertificatesAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, r)
}

// Marshal marshals to binary data
func (r *RootCertificatesAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(r)
}

// FromAnnotated converts from annotated to legacy format
func (r *RootCertificatesAnnotated) FromAnnotated() *RootCertificates {
	return &RootCertificates{
		Version:   r.Version,
		CertCount: r.CertCount,
		Reserved:  r.Reserved,
	}
}

// ToAnnotated converts RootCertificates to annotated format
func (r *RootCertificates) ToAnnotated() *RootCertificatesAnnotated {
	return &RootCertificatesAnnotated{
		Version:   r.Version,
		CertCount: r.CertCount,
		Reserved:  r.Reserved,
	}
}

// CertificateChainsAnnotated represents CERTIFICATE_CHAINS sections with annotations
type CertificateChainsAnnotated struct {
	Version    uint32    `offset:"0x0,endian:be"`     // Version
	ChainCount uint32    `offset:"0x4,endian:be"`     // Number of chains
	Reserved   [8]uint8  `offset:"0x8,reserved:true"` // Reserved
}

// Unmarshal unmarshals binary data
func (c *CertificateChainsAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, c)
}

// Marshal marshals to binary data
func (c *CertificateChainsAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(c)
}

// FromAnnotated converts from annotated to legacy format
func (c *CertificateChainsAnnotated) FromAnnotated() *CertificateChains {
	return &CertificateChains{
		Version:    c.Version,
		ChainCount: c.ChainCount,
		Reserved:   c.Reserved,
	}
}

// ToAnnotated converts CertificateChains to annotated format
func (c *CertificateChains) ToAnnotated() *CertificateChainsAnnotated {
	return &CertificateChainsAnnotated{
		Version:    c.Version,
		ChainCount: c.ChainCount,
		Reserved:   c.Reserved,
	}
}