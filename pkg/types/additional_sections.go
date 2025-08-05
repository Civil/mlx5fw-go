package types

// Additional section structures ported from mstflint

// SecureBootSignatures represents the secure boot signatures structure
// Based on image_layout_secure_boot_signatures from mstflint
type SecureBootSignatures struct {
	BootSignature         [128]uint32 // offset 0x0 - boot signature of data
	CriticalSignature     [128]uint32 // offset 0x200 - fw critical signature of itcos
	NonCriticalSignature  [128]uint32 // offset 0x400 - fw non critical signatures
}

// ResetCapabilities represents reset capabilities
// Based on image_layout_reset_capabilities from mstflint
type ResetCapabilities struct {
	ResetVerEn        uint8 `bin:"bitfield=1"` // bit 0
	Reserved1         uint8 `bin:"bitfield=7"` // bits 1-7
	VersionVectorVer  uint8 // bits 8-15
	Reserved2         uint16 // bits 16-31
}

// ResetVersion represents reset version structure
// Based on image_layout_reset_version from mstflint
type ResetVersion struct {
	Major  uint16 `bin:"BE,bitfield=16"` // bits 0-15
	Branch uint8  `bin:"BE,bitfield=4"`  // bits 16-19
	Minor  uint8  `bin:"BE,bitfield=8"`  // bits 20-27
	Reserved uint8 `bin:"BE,bitfield=4"` // bits 28-31
}

// VersionVector represents the version vector structure
// Based on image_layout_version_vector from mstflint
type VersionVector struct {
	ResetCapabilities ResetCapabilities // offset 0x0
	Scratchpad        ResetVersion      // offset 0x4 - cores SP
	ICMContext        ResetVersion      // offset 0x8
	PCI               ResetVersion      // offset 0xc - PCI domain
	PHY               ResetVersion      // offset 0x10 - PHY domain
	INI               ResetVersion      // offset 0x14
	Reserved1         ResetVersion      // offset 0x18
	Reserved2         ResetVersion      // offset 0x1c
	Reserved3         ResetVersion      // offset 0x20
	Reserved4         ResetVersion      // offset 0x24
	Reserved5         ResetVersion      // offset 0x28
	Reserved6         ResetVersion      // offset 0x2c
	Reserved7         ResetVersion      // offset 0x30 - total 52 bytes
}

// ResetInfo represents the RESET_INFO section structure
// This is custom based on typical section patterns, as the exact structure
// isn't explicitly defined in the reference files
type ResetInfo struct {
	VersionVector VersionVector // Version vector information (52 bytes)
	Reserved      [204]uint8    // Padding to 256 bytes total
}


// HMACDigest represents the HMAC_DIGEST section
// Based on typical digest patterns
type HMACDigest struct {
	DigestType uint32    // Type of digest (SHA256, SHA512, etc)
	DigestSize uint32    // Size of the digest in bytes
	Digest     [64]uint8 // The actual digest (max 512 bits)
	Reserved   [56]uint8 // Padding
}

// RSAPublicKey represents the RSA_PUBLIC_KEY section
// Based on image_layout patterns
type RSAPublicKey struct {
	KeyType     uint32     // RSA key type (2048, 4096, etc)
	KeySize     uint32     // Key size in bytes
	Exponent    uint32     // Public exponent (usually 65537)
	Reserved1   uint32     // Reserved
	Modulus     [512]uint8 // RSA modulus (up to 4096 bits)
	Reserved2   [496]uint8 // Padding to 1024 bytes
}

// DBGFwIni represents the DBG_FW_INI section
// This section contains debug firmware initialization data
type DBGFwIni struct {
	CompressionMethod uint32 // 0=Uncompressed, 1=Zlib, 2=LZMA
	UncompressedSize  uint32 // Size when uncompressed
	CompressedSize    uint32 // Size of compressed data
	Reserved          uint32 // Reserved
	// Data follows - variable size
}

// DBGFwParams represents the DBG_FW_PARAMS section
// This section contains debug firmware parameters
type DBGFwParams struct {
	CompressionMethod uint32 // 0=Uncompressed, 1=Zlib, 2=LZMA
	UncompressedSize  uint32 // Size when uncompressed
	CompressedSize    uint32 // Size of compressed data
	Reserved          uint32 // Reserved
	// Data follows - variable size
}

// FWAdb represents the FW_ADB section
// This section contains firmware ADB (Adb Database) data
type FWAdb struct {
	Version    uint32 // ADB version
	Size       uint32 // Size of ADB data
	Reserved1  uint32 // Reserved
	Reserved2  uint32 // Reserved
	// ADB data follows - variable size
}

// CRDumpMaskData represents the CRDUMP_MASK_DATA section
type CRDumpMaskData struct {
	Version  uint32    // Version of the mask data format
	MaskSize uint32    // Size of the mask data
	Reserved [8]uint32 // Reserved
	// Mask data follows - variable size
}

// FWNVLog represents the FW_NV_LOG section
type FWNVLog struct {
	LogVersion uint32    // Log format version
	LogSize    uint32    // Size of log data
	EntryCount uint32    // Number of log entries
	Reserved   [13]uint32 // Reserved
	// Log entries follow - variable size
}

// VPD_R0 represents the VPD_R0 section (Vital Product Data)
type VPD_R0 struct {
	ID       [2]uint8  // Should be "VPD_R0"
	Length   uint16    // Length of VPD data
	Reserved [60]uint8 // Reserved/padding
	// VPD data follows - variable size
}

// NVData represents the NV_DATA sections (NV_DATA0, NV_DATA1, NV_DATA2)
type NVData struct {
	Version   uint32     // NV data version
	DataSize  uint32     // Size of NV data
	Reserved1 uint32     // Reserved
	Reserved2 uint32     // Reserved
	Reserved3 [48]uint8  // Reserved padding
	// NV data follows - variable size
}

// FWInternalUsage represents the FW_INTERNAL_USAGE section
type FWInternalUsage struct {
	Version  uint32     // Version
	Size     uint32     // Data size
	Type     uint32     // Usage type
	Reserved [52]uint8  // Reserved
	// Internal data follows - variable size
}

// ProgrammableHWFW represents PROGRAMMABLE_HW_FW sections
type ProgrammableHWFW struct {
	Version      uint32    // Firmware version
	HWType       uint32    // Hardware type
	FWSize       uint32    // Firmware size
	Checksum     uint32    // Firmware checksum
	LoadAddress  uint32    // Load address in hardware
	EntryPoint   uint32    // Entry point address
	Reserved     [40]uint8 // Reserved
	// Firmware data follows - variable size
}

// DigitalCertPtr represents DIGITAL_CERT_PTR section
type DigitalCertPtr struct {
	CertType     uint32    // Certificate type
	CertOffset   uint32    // Offset to certificate
	CertSize     uint32    // Certificate size
	Reserved     [28]uint8 // Reserved (28 bytes to make total 40)
}

// DigitalCertRW represents DIGITAL_CERT_RW section
type DigitalCertRW struct {
	CertType     uint32      // Certificate type
	CertSize     uint32      // Certificate size
	ValidFrom    uint64      // Validity start timestamp
	ValidTo      uint64      // Validity end timestamp
	Reserved     [32]uint8   // Reserved
	Certificate  [4096]uint8 // Certificate data (up to 4K)
}

// CertChain represents CERT_CHAIN_0 section
type CertChain struct {
	ChainLength  uint32      // Number of certificates in chain
	Reserved     [12]uint8   // Reserved
	// Certificate chain data follows - variable size
}

// RootCertificates represents ROOT_CERTIFICATES sections
type RootCertificates struct {
	Version      uint32      // Version
	CertCount    uint32      // Number of certificates
	Reserved     [8]uint8    // Reserved
	// Root certificates follow - variable size
}

// CertificateChains represents CERTIFICATE_CHAINS sections
type CertificateChains struct {
	Version      uint32      // Version
	ChainCount   uint32      // Number of chains
	Reserved     [8]uint8    // Reserved
	// Certificate chains follow - variable size
}