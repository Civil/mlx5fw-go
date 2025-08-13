package types

import (
	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// SecureBootSignatures represents the secure boot signatures structure with annotations
type SecureBootSignatures struct {
	BootSignature        [128]uint32 `offset:"0x0,endian:be"`   // offset 0x0 - boot signature of data
	CriticalSignature    [128]uint32 `offset:"0x200,endian:be"` // offset 0x200 - fw critical signature of itcos
	NonCriticalSignature [128]uint32 `offset:"0x400,endian:be"` // offset 0x400 - fw non critical signatures
}

// Unmarshal unmarshals binary data
func (s *SecureBootSignatures) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, s)
}

// Marshal marshals to binary data
func (s *SecureBootSignatures) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(s)
}

// ResetCapabilities represents reset capabilities with annotations
type ResetCapabilities struct {
	ResetVerEn       uint8  `offset:"bit:0,len:1,endian:be"` // bit 0
	Reserved1        uint8  `offset:"bit:1,len:7,endian:be"` // bits 1-7
	VersionVectorVer uint8  `offset:"byte:1,endian:be"`      // bits 8-15
	Reserved2        uint16 `offset:"byte:2,endian:be"`      // bits 16-31
}

// Unmarshal unmarshals binary data
func (r *ResetCapabilities) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, r)
}

// Marshal marshals to binary data
func (r *ResetCapabilities) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(r)
}

// ResetVersion represents reset version structure with annotations
type ResetVersion struct {
	Major    uint16 `offset:"bit:0,len:16,endian:be"` // bits 0-15
	Branch   uint8  `offset:"bit:16,len:4,endian:be"` // bits 16-19
	Minor    uint8  `offset:"bit:20,len:8,endian:be"` // bits 20-27
	Reserved uint8  `offset:"bit:28,len:4,endian:be"` // bits 28-31
}

// Unmarshal unmarshals binary data
func (r *ResetVersion) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, r)
}

// Marshal marshals to binary data
func (r *ResetVersion) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(r)
}

// VersionVector represents the version vector structure with annotations
type VersionVector struct {
	ResetCapabilities ResetCapabilities `offset:"0x0"`  // offset 0x0
	Scratchpad        ResetVersion      `offset:"0x4"`  // offset 0x4 - cores SP
	ICMContext        ResetVersion      `offset:"0x8"`  // offset 0x8
	PCI               ResetVersion      `offset:"0xc"`  // offset 0xc - PCI domain
	PHY               ResetVersion      `offset:"0x10"` // offset 0x10 - PHY domain
	INI               ResetVersion      `offset:"0x14"` // offset 0x14
	Reserved1         ResetVersion      `offset:"0x18"` // offset 0x18
	Reserved2         ResetVersion      `offset:"0x1c"` // offset 0x1c
	Reserved3         ResetVersion      `offset:"0x20"` // offset 0x20
	Reserved4         ResetVersion      `offset:"0x24"` // offset 0x24
	Reserved5         ResetVersion      `offset:"0x28"` // offset 0x28
	Reserved6         ResetVersion      `offset:"0x2c"` // offset 0x2c
	Reserved7         ResetVersion      `offset:"0x30"` // offset 0x30 - total 52 bytes
}

// Unmarshal unmarshals binary data
func (v *VersionVector) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, v)
}

// Marshal marshals to binary data
func (v *VersionVector) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(v)
}

// ResetInfo represents the RESET_INFO section structure with annotations
type ResetInfo struct {
	Data [256]uint8 `offset:"0x0"` // Full 256-byte section
}

// Unmarshal unmarshals binary data
func (r *ResetInfo) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, r)
}

// Marshal marshals to binary data
func (r *ResetInfo) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(r)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (r *ResetInfo) UnmarshalWithReserved(data []byte) error {
	opts := &annotations.UnmarshalOptions{IncludeReserved: true}
	return annotations.UnmarshalWithOptionsStruct(data, r, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (r *ResetInfo) MarshalWithReserved() ([]byte, error) {
	opts := &annotations.MarshalOptions{IncludeReserved: true}
	return annotations.MarshalWithOptionsStruct(r, opts)
}

// HMACDigest represents the HMAC_DIGEST section with annotations
type HMACDigest struct {
	DigestType uint32    `offset:"0x0,endian:be"` // Type of digest (SHA256, SHA512, etc)
	DigestSize uint32    `offset:"0x4,endian:be"` // Size of the digest in bytes
	Digest     [64]uint8 `offset:"0x8"`           // The actual digest (max 512 bits)
	Reserved   [56]uint8 `offset:"0x48"`          // Padding
}

// Unmarshal unmarshals binary data
func (h *HMACDigest) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// Marshal marshals to binary data
func (h *HMACDigest) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (h *HMACDigest) UnmarshalWithReserved(data []byte) error {
	opts := &annotations.UnmarshalOptions{IncludeReserved: true}
	return annotations.UnmarshalWithOptionsStruct(data, h, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (h *HMACDigest) MarshalWithReserved() ([]byte, error) {
	opts := &annotations.MarshalOptions{IncludeReserved: true}
	return annotations.MarshalWithOptionsStruct(h, opts)
}

// RSAPublicKey represents the RSA_PUBLIC_KEY section with annotations
type RSAPublicKey struct {
	KeyType   uint32     `offset:"0x0,endian:be"` // RSA key type (2048, 4096, etc)
	KeySize   uint32     `offset:"0x4,endian:be"` // Key size in bytes
	Exponent  uint32     `offset:"0x8,endian:be"` // Public exponent (usually 65537)
	Reserved1 uint32     `offset:"0xc,endian:be"` // Reserved
	Modulus   [512]uint8 `offset:"0x10"`          // RSA modulus (up to 4096 bits)
	Reserved2 [496]uint8 `offset:"0x210"`         // Padding to 1024 bytes
}

// Unmarshal unmarshals binary data
func (r *RSAPublicKey) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, r)
}

// Marshal marshals to binary data
func (r *RSAPublicKey) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(r)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (r *RSAPublicKey) UnmarshalWithReserved(data []byte) error {
	opts := &annotations.UnmarshalOptions{IncludeReserved: true}
	return annotations.UnmarshalWithOptionsStruct(data, r, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (r *RSAPublicKey) MarshalWithReserved() ([]byte, error) {
	opts := &annotations.MarshalOptions{IncludeReserved: true}
	return annotations.MarshalWithOptionsStruct(r, opts)
}

// DBGFwIni represents the DBG_FW_INI section with annotations
type DBGFwIni struct {
	CompressionMethod uint32 `offset:"0x0,endian:be"` // 0=Uncompressed, 1=Zlib, 2=LZMA
	UncompressedSize  uint32 `offset:"0x4,endian:be"` // Size when uncompressed
	CompressedSize    uint32 `offset:"0x8,endian:be"` // Size of compressed data
	Reserved          uint32 `offset:"0xc,endian:be"` // Reserved
	// Data follows - variable size
}

// Unmarshal unmarshals binary data
func (d *DBGFwIni) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, d)
}

// Marshal marshals to binary data
func (d *DBGFwIni) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(d)
}

// DBGFwParams represents the DBG_FW_PARAMS section with annotations
type DBGFwParams struct {
	CompressionMethod uint32 `offset:"0x0,endian:be"` // 0=Uncompressed, 1=Zlib, 2=LZMA
	UncompressedSize  uint32 `offset:"0x4,endian:be"` // Size when uncompressed
	CompressedSize    uint32 `offset:"0x8,endian:be"` // Size of compressed data
	Reserved          uint32 `offset:"0xc,endian:be"` // Reserved
	// Data follows - variable size
}

// Unmarshal unmarshals binary data
func (d *DBGFwParams) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, d)
}

// Marshal marshals to binary data
func (d *DBGFwParams) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(d)
}

// FWAdb represents the FW_ADB section with annotations
type FWAdb struct {
	Version   uint32 `offset:"0x0,endian:be"` // ADB version
	Size      uint32 `offset:"0x4,endian:be"` // Size of ADB data
	Reserved1 uint32 `offset:"0x8,endian:be"` // Reserved
	Reserved2 uint32 `offset:"0xc,endian:be"` // Reserved
	// ADB data follows - variable size
}

// Unmarshal unmarshals binary data
func (f *FWAdb) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, f)
}

// Marshal marshals to binary data
func (f *FWAdb) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(f)
}

// CRDumpMaskData represents the CRDUMP_MASK_DATA section with annotations
type CRDumpMaskData struct {
	Version  uint32    `offset:"0x0,endian:be"` // Version of the mask data format
	MaskSize uint32    `offset:"0x4,endian:be"` // Size of the mask data
	Reserved [8]uint32 `offset:"0x8,endian:be"` // Reserved
	// Mask data follows - variable size
}

// Unmarshal unmarshals binary data
func (c *CRDumpMaskData) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, c)
}

// Marshal marshals to binary data
func (c *CRDumpMaskData) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(c)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (c *CRDumpMaskData) UnmarshalWithReserved(data []byte) error {
	opts := &annotations.UnmarshalOptions{IncludeReserved: true}
	return annotations.UnmarshalWithOptionsStruct(data, c, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (c *CRDumpMaskData) MarshalWithReserved() ([]byte, error) {
	opts := &annotations.MarshalOptions{IncludeReserved: true}
	return annotations.MarshalWithOptionsStruct(c, opts)
}

// ForbiddenVersions represents the FORBIDDEN_VERSIONS section with annotations
type ForbiddenVersions struct {
	Count    uint32   `offset:"byte:0,endian:be" json:"count"`                              // Number of forbidden versions
	Reserved uint32   `offset:"byte:4,endian:be,reserved:true" json:"reserved"`             // Reserved for alignment
	Versions []uint32 `offset:"byte:8,endian:be,list_size:Count" json:"versions,omitempty"` // List of forbidden version numbers
}

// Unmarshal unmarshals binary data
func (f *ForbiddenVersions) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, f)
}

// Marshal marshals to binary data
func (f *ForbiddenVersions) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(f)
}

// FWNVLog represents the FW_NV_LOG section with annotations
type FWNVLog struct {
	LogVersion uint32     `offset:"0x0,endian:be"`               // Log format version
	LogSize    uint32     `offset:"0x4,endian:be"`               // Size of log data
	EntryCount uint32     `offset:"0x8,endian:be"`               // Number of log entries
	Reserved   [13]uint32 `offset:"0xc,endian:be,reserved:true"` // Reserved
}

// Unmarshal unmarshals binary data
func (f *FWNVLog) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, f)
}

// Marshal marshals to binary data
func (f *FWNVLog) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(f)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (f *FWNVLog) UnmarshalWithReserved(data []byte) error {
	opts := &annotations.UnmarshalOptions{IncludeReserved: true}
	return annotations.UnmarshalWithOptionsStruct(data, f, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (f *FWNVLog) MarshalWithReserved() ([]byte, error) {
	opts := &annotations.MarshalOptions{IncludeReserved: true}
	return annotations.MarshalWithOptionsStruct(f, opts)
}

// VPD_R0 represents the VPD_R0 section (Vital Product Data) with annotations
type VPD_R0 struct {
	ID       [2]uint8  `offset:"0x0"`               // Should be "VPD_R0"
	Length   uint16    `offset:"0x2,endian:be"`     // Length of VPD data
	Reserved [60]uint8 `offset:"0x4,reserved:true"` // Reserved/padding
}

// Unmarshal unmarshals binary data
func (v *VPD_R0) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, v)
}

// Marshal marshals to binary data
func (v *VPD_R0) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(v)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (v *VPD_R0) UnmarshalWithReserved(data []byte) error {
	opts := &annotations.UnmarshalOptions{IncludeReserved: true}
	return annotations.UnmarshalWithOptionsStruct(data, v, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (v *VPD_R0) MarshalWithReserved() ([]byte, error) {
	opts := &annotations.MarshalOptions{IncludeReserved: true}
	return annotations.MarshalWithOptionsStruct(v, opts)
}

// NVData represents the NV_DATA sections (NV_DATA0, NV_DATA1, NV_DATA2) with annotations
type NVData struct {
	Version   uint32    `offset:"0x0,endian:be"`               // NV data version
	DataSize  uint32    `offset:"0x4,endian:be"`               // Size of NV data
	Reserved1 uint32    `offset:"0x8,endian:be,reserved:true"` // Reserved
	Reserved2 uint32    `offset:"0xc,endian:be,reserved:true"` // Reserved
	Reserved3 [48]uint8 `offset:"0x10,reserved:true"`          // Reserved padding
}

// Unmarshal unmarshals binary data
func (n *NVData) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, n)
}

// Marshal marshals to binary data
func (n *NVData) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(n)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (n *NVData) UnmarshalWithReserved(data []byte) error {
	opts := &annotations.UnmarshalOptions{IncludeReserved: true}
	return annotations.UnmarshalWithOptionsStruct(data, n, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (n *NVData) MarshalWithReserved() ([]byte, error) {
	opts := &annotations.MarshalOptions{IncludeReserved: true}
	return annotations.MarshalWithOptionsStruct(n, opts)
}

// FWInternalUsage represents the FW_INTERNAL_USAGE section with annotations
type FWInternalUsage struct {
	Version  uint32    `offset:"0x0,endian:be"`     // Version
	Size     uint32    `offset:"0x4,endian:be"`     // Data size
	Type     uint32    `offset:"0x8,endian:be"`     // Usage type
	Reserved [52]uint8 `offset:"0xc,reserved:true"` // Reserved
}

// Unmarshal unmarshals binary data
func (f *FWInternalUsage) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, f)
}

// Marshal marshals to binary data
func (f *FWInternalUsage) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(f)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (f *FWInternalUsage) UnmarshalWithReserved(data []byte) error {
	opts := &annotations.UnmarshalOptions{IncludeReserved: true}
	return annotations.UnmarshalWithOptionsStruct(data, f, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (f *FWInternalUsage) MarshalWithReserved() ([]byte, error) {
	opts := &annotations.MarshalOptions{IncludeReserved: true}
	return annotations.MarshalWithOptionsStruct(f, opts)
}

// ProgrammableHWFW represents PROGRAMMABLE_HW_FW sections with annotations
type ProgrammableHWFW struct {
	Version     uint32    `offset:"0x0,endian:be"`      // Firmware version
	HWType      uint32    `offset:"0x4,endian:be"`      // Hardware type
	FWSize      uint32    `offset:"0x8,endian:be"`      // Firmware size
	Checksum    uint32    `offset:"0xc,endian:be"`      // Firmware checksum
	LoadAddress uint32    `offset:"0x10,endian:be"`     // Load address in hardware
	EntryPoint  uint32    `offset:"0x14,endian:be"`     // Entry point address
	Reserved    [40]uint8 `offset:"0x18,reserved:true"` // Reserved
}

// Unmarshal unmarshals binary data
func (p *ProgrammableHWFW) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, p)
}

// Marshal marshals to binary data
func (p *ProgrammableHWFW) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(p)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (p *ProgrammableHWFW) UnmarshalWithReserved(data []byte) error {
	opts := &annotations.UnmarshalOptions{IncludeReserved: true}
	return annotations.UnmarshalWithOptionsStruct(data, p, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (p *ProgrammableHWFW) MarshalWithReserved() ([]byte, error) {
	opts := &annotations.MarshalOptions{IncludeReserved: true}
	return annotations.MarshalWithOptionsStruct(p, opts)
}

// DigitalCertPtr represents DIGITAL_CERT_PTR section with annotations
type DigitalCertPtr struct {
	CertType   uint32    `offset:"0x0,endian:be"`     // Certificate type
	CertOffset uint32    `offset:"0x4,endian:be"`     // Offset to certificate
	CertSize   uint32    `offset:"0x8,endian:be"`     // Certificate size
	Reserved   [28]uint8 `offset:"0xc,reserved:true"` // Reserved (28 bytes to make total 40)
}

// Unmarshal unmarshals binary data
func (d *DigitalCertPtr) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, d)
}

// Marshal marshals to binary data
func (d *DigitalCertPtr) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(d)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (d *DigitalCertPtr) UnmarshalWithReserved(data []byte) error {
	opts := &annotations.UnmarshalOptions{IncludeReserved: true}
	return annotations.UnmarshalWithOptionsStruct(data, d, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (d *DigitalCertPtr) MarshalWithReserved() ([]byte, error) {
	opts := &annotations.MarshalOptions{IncludeReserved: true}
	return annotations.MarshalWithOptionsStruct(d, opts)
}

// DigitalCertRW represents DIGITAL_CERT_RW section with annotations
type DigitalCertRW struct {
	CertType    uint32      `offset:"0x0,endian:be"`      // Certificate type
	CertSize    uint32      `offset:"0x4,endian:be"`      // Certificate size
	ValidFrom   uint64      `offset:"0x8,endian:be"`      // Validity start timestamp
	ValidTo     uint64      `offset:"0x10,endian:be"`     // Validity end timestamp
	Reserved    [32]uint8   `offset:"0x18,reserved:true"` // Reserved
	Certificate [4096]uint8 `offset:"0x38"`               // Certificate data (up to 4K)
}

// Unmarshal unmarshals binary data
func (d *DigitalCertRW) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, d)
}

// Marshal marshals to binary data
func (d *DigitalCertRW) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(d)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (d *DigitalCertRW) UnmarshalWithReserved(data []byte) error {
	opts := &annotations.UnmarshalOptions{IncludeReserved: true}
	return annotations.UnmarshalWithOptionsStruct(data, d, opts)
}

// MarshalWithReserved marshals to binary data including reserved fields
func (d *DigitalCertRW) MarshalWithReserved() ([]byte, error) {
	opts := &annotations.MarshalOptions{IncludeReserved: true}
	return annotations.MarshalWithOptionsStruct(d, opts)
}

// CertChain represents CERT_CHAIN_0 section with annotations
type CertChain struct {
	ChainLength uint32    `offset:"0x0,endian:be"`     // Number of certificates in chain
	Reserved    [12]uint8 `offset:"0x4,reserved:true"` // Reserved
}

// Unmarshal unmarshals binary data
func (c *CertChain) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, c)
}

// Marshal marshals to binary data
func (c *CertChain) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(c)
}

// RootCertificates represents ROOT_CERTIFICATES sections with annotations
type RootCertificates struct {
	Version   uint32   `offset:"0x0,endian:be"`     // Version
	CertCount uint32   `offset:"0x4,endian:be"`     // Number of certificates
	Reserved  [8]uint8 `offset:"0x8,reserved:true"` // Reserved
}

// Unmarshal unmarshals binary data
func (r *RootCertificates) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, r)
}

// Marshal marshals to binary data
func (r *RootCertificates) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(r)
}

// CertificateChains represents CERTIFICATE_CHAINS sections with annotations
type CertificateChains struct {
	Version    uint32   `offset:"0x0,endian:be"`     // Version
	ChainCount uint32   `offset:"0x4,endian:be"`     // Number of chains
	Reserved   [8]uint8 `offset:"0x8,reserved:true"` // Reserved
}

// Unmarshal unmarshals binary data
func (c *CertificateChains) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, c)
}

// Marshal marshals to binary data
func (c *CertificateChains) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(c)
}
