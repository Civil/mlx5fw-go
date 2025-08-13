package types

import (
	"encoding/hex"
	"strings"
)

// ImageInfoJSON is a JSON-serializable version of ImageInfo that can be directly marshaled/unmarshaled
type ImageInfoJSON struct {
	// Security and version info
	SecurityAndVersion uint32 `json:"security_and_version"`
	MinorVersion       uint8  `json:"minor_version"`
	SecurityMode       uint8  `json:"security_mode"`
	MajorVersion       uint8  `json:"major_version"`
	Reserved0          uint8  `json:"reserved0"`
	MCCEnabled         bool   `json:"mcc_enabled"`
	DebugFW            bool   `json:"debug_fw"`
	SignedFW           bool   `json:"signed_fw"`
	SecureFW           bool   `json:"secure_fw"`

	// FW version fields
	FWVerMajor    uint16 `json:"fw_ver_major"`
	Reserved2     uint16 `json:"reserved2"`
	FWVerMinor    uint16 `json:"fw_ver_minor"`
	FWVerSubminor uint16 `json:"fw_ver_subminor"`

	// Date/time fields
	Reserved3a uint8  `json:"reserved3a"`
	Hour       uint8  `json:"hour"`
	Minutes    uint8  `json:"minutes"`
	Seconds    uint8  `json:"seconds"`
	Day        uint8  `json:"day"`
	Month      uint8  `json:"month"`
	Year       uint16 `json:"year"`

	// MIC version fields
	MICVerMajor    uint16 `json:"mic_ver_major"`
	Reserved4      uint16 `json:"reserved4"`
	MICVerSubminor uint16 `json:"mic_ver_subminor"`
	MICVerMinor    uint16 `json:"mic_ver_minor"`

	// PCI IDs
	PCIDeviceID    uint16 `json:"pci_device_id"`
	PCIVendorID    uint16 `json:"pci_vendor_id"`
	PCISubsystemID uint16 `json:"pci_subsystem_id"`
	PCISubVendorID uint16 `json:"pci_subvendor_id"`

	// PSID
	PSID string `json:"psid"`

	// VSD info
	Reserved5a  uint16 `json:"reserved5a"`
	VSDVendorID uint16 `json:"vsd_vendor_id"`
	VSD         string `json:"vsd"`

	// Image size data
	ImageSizeData []uint8 `json:"image_size_data"`
	Reserved6     []uint8 `json:"reserved6"`

	// Supported HW IDs
	SupportedHWIDs []uint32 `json:"supported_hw_ids"`

	// INI file num
	INIFileNum uint32  `json:"ini_file_num"`
	Reserved7  []uint8 `json:"reserved7"`

	// Product version
	ProductVer    string  `json:"product_ver"`
	ProductVerRaw []uint8 `json:"product_ver_raw"`

	// Description
	Description string  `json:"description"`
	Reserved8   []uint8 `json:"reserved8"`

	// Module versions
	ModuleVersions []uint8 `json:"module_versions"`

	// Name fields
	Name    string `json:"name"`
	PRSName string `json:"prs_name"`

	// Human-readable versions for convenience
	FWVersion          string `json:"fw_version"`
	FWReleaseDate      string `json:"fw_release_date"`
	MICVersion         string `json:"mic_version"`
	SecurityAttributes string `json:"security_attributes"`
}

// FromImageInfo converts ImageInfo to ImageInfoJSON
func (j *ImageInfoJSON) FromImageInfo(info *ImageInfo) {
	// Security and version
	j.SecurityAndVersion = info.GetSecurityAndVersion()
	j.MinorVersion = info.MinorVersion
	j.SecurityMode = info.SecurityMode
	j.MajorVersion = info.MajorVersion
	j.Reserved0 = info.Reserved0
	j.MCCEnabled = info.MCCEnabled
	j.DebugFW = info.DebugFW
	j.SignedFW = info.SignedFW
	j.SecureFW = info.SecureFW

	// FW version fields
	j.FWVerMajor = info.FWVerMajor
	j.Reserved2 = info.Reserved2
	j.FWVerMinor = info.FWVerMinor
	j.FWVerSubminor = info.FWVerSubminor

	// Date/time fields
	j.Reserved3a = info.Reserved3a
	j.Hour = info.Hour
	j.Minutes = info.Minutes
	j.Seconds = info.Seconds
	j.Day = info.Day
	j.Month = info.Month
	j.Year = info.Year

	// MIC version fields
	j.MICVerMajor = info.MICVerMajor
	j.Reserved4 = info.Reserved4
	j.MICVerSubminor = info.MICVerSubminor
	j.MICVerMinor = info.MICVerMinor

	// PCI IDs
	j.PCIDeviceID = info.PCIDeviceID
	j.PCIVendorID = info.PCIVendorID
	j.PCISubsystemID = info.PCISubsystemID
	j.PCISubVendorID = info.PCISubVendorID

	// PSID
	j.PSID = info.GetPSIDString()

	// VSD info
	j.Reserved5a = info.Reserved5a
	j.VSDVendorID = info.VSDVendorID
	j.VSD = info.GetVSDString()

	// Image size data
	j.ImageSizeData = info.ImageSizeData[:]
	j.Reserved6 = info.Reserved6[:]

	// Supported HW IDs
	j.SupportedHWIDs = info.SupportedHWID[:]

	// INI file num
	j.INIFileNum = info.INIFileNum
	j.Reserved7 = info.Reserved7[:]

	// Product version
	j.ProductVer = info.GetProductVerString()
	j.ProductVerRaw = info.ProductVer[:]

	// Description
	j.Description = info.GetDescriptionString()
	j.Reserved8 = info.Reserved8[:]

	// Module versions
	j.ModuleVersions = info.ModuleVersions[:]

	// Name fields
	j.Name = info.GetNameString()
	j.PRSName = info.GetPRSNameString()

	// Human-readable versions
	j.FWVersion = info.GetFWVersionString()
	j.FWReleaseDate = info.GetFWReleaseDateString()
	j.MICVersion = info.GetMICVersionString()
	j.SecurityAttributes = info.GetSecurityAttributesString()
}

// ToImageInfo converts ImageInfoJSON back to ImageInfo
func (j *ImageInfoJSON) ToImageInfo() *ImageInfo {
	info := &ImageInfo{}

	// Security and version - reconstruct from SecurityAndVersion
	sv := j.SecurityAndVersion
	info.MinorVersion = uint8(sv & 0xFF)         // Bits 0-7
	info.SecurityMode = uint8((sv >> 8) & 0xFF)  // Bits 8-15
	info.MajorVersion = uint8((sv >> 16) & 0xFF) // Bits 16-23
	info.Reserved0 = uint8((sv >> 24) & 0xFF)    // Bits 24-31

	// Extract individual flags from bits 8-15 (SecurityMode)
	info.MCCEnabled = (sv & (1 << 8)) != 0 // Bit 8
	info.DebugFW = (sv & (1 << 13)) != 0   // Bit 13
	info.SignedFW = (sv & (1 << 14)) != 0  // Bit 14
	info.SecureFW = (sv & (1 << 15)) != 0  // Bit 15

	// FW version fields
	info.FWVerMajor = j.FWVerMajor
	info.Reserved2 = j.Reserved2
	info.FWVerMinor = j.FWVerMinor
	info.FWVerSubminor = j.FWVerSubminor

	// Date/time fields
	info.Reserved3a = j.Reserved3a
	info.Hour = j.Hour
	info.Minutes = j.Minutes
	info.Seconds = j.Seconds
	info.Day = j.Day
	info.Month = j.Month
	info.Year = j.Year

	// MIC version fields
	info.MICVerMajor = j.MICVerMajor
	info.Reserved4 = j.Reserved4
	info.MICVerSubminor = j.MICVerSubminor
	info.MICVerMinor = j.MICVerMinor

	// PCI IDs
	info.PCIDeviceID = j.PCIDeviceID
	info.PCIVendorID = j.PCIVendorID
	info.PCISubsystemID = j.PCISubsystemID
	info.PCISubVendorID = j.PCISubVendorID

	// PSID
	psidBytes := []byte(j.PSID)
	if len(psidBytes) > 16 {
		psidBytes = psidBytes[:16]
	}
	copy(info.PSID[:], psidBytes)

	// VSD info
	info.Reserved5a = j.Reserved5a
	info.VSDVendorID = j.VSDVendorID
	vsdBytes := []byte(j.VSD)
	if len(vsdBytes) > 208 {
		vsdBytes = vsdBytes[:208]
	}
	copy(info.VSD[:], vsdBytes)

	// Image size data
	copy(info.ImageSizeData[:], j.ImageSizeData)
	copy(info.Reserved6[:], j.Reserved6)

	// Supported HW IDs
	copy(info.SupportedHWID[:], j.SupportedHWIDs)

	// INI file num
	info.INIFileNum = j.INIFileNum
	copy(info.Reserved7[:], j.Reserved7)

	// Product version - prefer raw bytes if available
	if len(j.ProductVerRaw) > 0 {
		copy(info.ProductVer[:], j.ProductVerRaw)
	} else {
		prodVerBytes := []byte(j.ProductVer)
		if len(prodVerBytes) > 16 {
			prodVerBytes = prodVerBytes[:16]
		}
		copy(info.ProductVer[:], prodVerBytes)
	}

	// Description
	descBytes := []byte(j.Description)
	if len(descBytes) > 256 {
		descBytes = descBytes[:256]
	}
	copy(info.Description[:], descBytes)
	copy(info.Reserved8[:], j.Reserved8)

	// Module versions
	copy(info.ModuleVersions[:], j.ModuleVersions)

	// Name fields
	nameBytes := []byte(j.Name)
	if len(nameBytes) > 64 {
		nameBytes = nameBytes[:64]
	}
	copy(info.Name[:], nameBytes)

	prsBytes := []byte(j.PRSName)
	if len(prsBytes) > 128 {
		prsBytes = prsBytes[:128]
	}
	copy(info.PRSName[:], prsBytes)

	return info
}

// DevInfoJSON is a JSON-serializable version of DevInfo
type DevInfoJSON struct {
	Signature0   uint32    `json:"signature0"`
	Signature1   uint32    `json:"signature1"`
	Signature2   uint32    `json:"signature2"`
	Signature3   uint32    `json:"signature3"`
	MinorVersion uint8     `json:"minor_version"`
	MajorVersion uint16    `json:"major_version"`
	Reserved1    uint8     `json:"reserved1"`
	Reserved2    []byte    `json:"reserved2"`
	Guids        GuidsJSON `json:"guids"`
	Macs         MacsJSON  `json:"macs"`
	Reserved3    []byte    `json:"reserved3"`
	Reserved4    uint16    `json:"reserved4"`
	CRC          uint16    `json:"crc"`
	OriginalCRC  uint16    `json:"original_crc"`
}

// GuidsJSON represents GUID info in JSON
type GuidsJSON struct {
	Reserved1    uint16 `json:"reserved1"`
	Step         uint8  `json:"step"`
	NumAllocated uint8  `json:"num_allocated"`
	Reserved2    uint32 `json:"reserved2"`
	UID          uint64 `json:"uid"`
}

// MacsJSON represents MAC info in JSON
type MacsJSON struct {
	Reserved1    uint16 `json:"reserved1"`
	Step         uint8  `json:"step"`
	NumAllocated uint8  `json:"num_allocated"`
	Reserved2    uint32 `json:"reserved2"`
	UID          uint64 `json:"uid"`
}

// MfgInfoJSON is a JSON-serializable version of MfgInfo
type MfgInfoJSON struct {
	PSID        []byte `json:"psid"`
	PartNumber  []byte `json:"part_number"`
	Revision    []byte `json:"revision"`
	ProductName []byte `json:"product_name"`
	Reserved    []byte `json:"reserved"`
}

// HashesTableJSON represents a hashes table section in JSON
type HashesTableJSON struct {
	Header       HashesTableHeaderJSON `json:"header"`
	Entries      []HashTableEntryJSON  `json:"entries"`
	ReservedTail string                `json:"reserved_tail"` // Hex encoded
}

// HashesTableHeaderJSON represents the header in JSON
type HashesTableHeaderJSON struct {
	Magic      uint32 `json:"magic"`
	Version    uint32 `json:"version"`
	Reserved1  uint32 `json:"reserved1"`
	Reserved2  uint32 `json:"reserved2"`
	TableSize  uint32 `json:"table_size"`
	NumEntries uint32 `json:"num_entries"`
	Reserved3  uint32 `json:"reserved3"`
	CRC        uint16 `json:"crc"`
	Reserved4  uint16 `json:"reserved4"`
}

// HashTableEntryJSON represents a hash table entry in JSON
type HashTableEntryJSON struct {
	Hash   string `json:"hash"` // Hex encoded
	Type   uint32 `json:"type"`
	Offset uint32 `json:"offset"`
	Size   uint32 `json:"size"`
}

// ImageSignatureJSON represents image signature sections in JSON
type ImageSignatureJSON struct {
	SignatureType uint32 `json:"signature_type"`
	Signature     string `json:"signature"`         // Hex encoded
	Padding       string `json:"padding,omitempty"` // Hex encoded padding data
}

// PublicKeyJSON represents a public key entry in JSON
type PublicKeyJSON struct {
	Index    int    `json:"index,omitempty"`
	Reserved uint32 `json:"reserved"`
	UUID     string `json:"uuid"` // Hex encoded
	Key      string `json:"key"`  // Hex encoded
}

// ForbiddenVersionsJSON represents forbidden versions in JSON
type ForbiddenVersionsJSON struct {
	Count    uint32   `json:"count"`
	Reserved uint32   `json:"reserved"`
	Versions []uint32 `json:"versions"`
}

// HWPointerEntryJSON represents a hardware pointer entry in JSON
type HWPointerEntryJSON struct {
	Pointer uint32 `json:"pointer"`
	CRC     uint16 `json:"crc"`
}

// FS4HWPointersJSON represents FS4 hardware pointers in JSON
type FS4HWPointersJSON struct {
	BootRecordPtr        HWPointerEntryJSON `json:"boot_record_ptr"`
	Boot2Ptr             HWPointerEntryJSON `json:"boot2_ptr"`
	TOCPtr               HWPointerEntryJSON `json:"toc_ptr"`
	ToolsPtr             HWPointerEntryJSON `json:"tools_ptr"`
	FWWindowStartPtr     HWPointerEntryJSON `json:"fw_window_start_ptr"`
	FWWindowEndPtr       HWPointerEntryJSON `json:"fw_window_end_ptr"`
	ImageInfoSectionPtr  HWPointerEntryJSON `json:"image_info_section_ptr"`
	HashesTablePtr       HWPointerEntryJSON `json:"hashes_table_ptr"`
	DigestRecoveryKeyPtr HWPointerEntryJSON `json:"digest_recovery_key_ptr"`
	DigestPtr            HWPointerEntryJSON `json:"digest_ptr"`
}

// FS5HWPointersJSON represents FS5 hardware pointers in JSON
type FS5HWPointersJSON struct {
	Boot2Ptr             HWPointerEntryJSON `json:"boot2_ptr"`
	TOCPtr               HWPointerEntryJSON `json:"toc_ptr"`
	ToolsPtr             HWPointerEntryJSON `json:"tools_ptr"`
	ImageInfoSectionPtr  HWPointerEntryJSON `json:"image_info_section_ptr"`
	FWPublicKeyPtr       HWPointerEntryJSON `json:"fw_public_key_ptr"`
	FWSignaturePtr       HWPointerEntryJSON `json:"fw_signature_ptr"`
	PublicKeyPtr         HWPointerEntryJSON `json:"public_key_ptr"`
	ForbiddenVersionsPtr HWPointerEntryJSON `json:"forbidden_versions_ptr"`
	PSCBl1Ptr            HWPointerEntryJSON `json:"psc_bl1_ptr"`
	PSCHashesTablePtr    HWPointerEntryJSON `json:"psc_hashes_table_ptr"`
	NCoreHashesPointer   HWPointerEntryJSON `json:"ncore_hashes_pointer"`
	PSCFWUpdateHandlePtr HWPointerEntryJSON `json:"psc_fw_update_handle_ptr"`
	PSCBCHPointer        HWPointerEntryJSON `json:"psc_bch_pointer"`
	NCoreBCHPointer      HWPointerEntryJSON `json:"ncore_bch_pointer"`
}

// ITOCEntryJSON represents an ITOC entry in JSON
type ITOCEntryJSON struct {
	Type            uint16 `json:"type"`
	DeviceData      bool   `json:"device_data"`
	NoHashCheck     bool   `json:"no_hash_check"`
	RelativeAddress bool   `json:"relative_address"`
	SectionPointer  uint32 `json:"section_pointer"`
	SectionSize     uint32 `json:"section_size"`
	FlashAddress    uint32 `json:"flash_address"`
	CRC             uint16 `json:"crc"`
}

// ITOCHeaderJSON represents an ITOC header in JSON
type ITOCHeaderJSON struct {
	IdentifierLo      uint32 `json:"identifier_lo"`
	IdentifierHi      uint32 `json:"identifier_hi"`
	Version           uint8  `json:"version"`
	Reserved0         uint8  `json:"reserved0"`
	Reserved1         uint16 `json:"reserved1"`
	Reserved2         uint32 `json:"reserved2"`
	Reserved3         uint32 `json:"reserved3"`
	Reserved4         uint32 `json:"reserved4"`
	NextHeaderPointer uint32 `json:"next_header_pointer"`
	NextHeaderSize    uint32 `json:"next_header_size"`
	Padding           string `json:"padding,omitempty"`
}

// ITOCJSON represents an ITOC section in JSON
type ITOCJSON struct {
	Header  ITOCHeaderJSON  `json:"header"`
	Entries []ITOCEntryJSON `json:"entries"`
}

// DTOCEntryJSON represents a DTOC entry in JSON
type DTOCEntryJSON struct {
	Type             uint16 `json:"type"`
	Size             uint32 `json:"size"`
	ParamStartOffset uint32 `json:"param_start_offset"`
	ParamEndOffset   uint32 `json:"param_end_offset"`
	DataStartOffset  uint32 `json:"data_start_offset"`
	DataEndOffset    uint32 `json:"data_end_offset"`
	CRC              uint16 `json:"crc"`
}

// DTOCHeaderJSON represents a DTOC header in JSON
type DTOCHeaderJSON struct {
	IdentifierLo      uint32 `json:"identifier_lo"`
	IdentifierHi      uint32 `json:"identifier_hi"`
	Version           uint8  `json:"version"`
	Reserved0         uint8  `json:"reserved0"`
	Reserved1         uint16 `json:"reserved1"`
	Reserved2         uint32 `json:"reserved2"`
	Reserved3         uint32 `json:"reserved3"`
	Reserved4         uint32 `json:"reserved4"`
	NextHeaderPointer uint32 `json:"next_header_pointer"`
	NextHeaderSize    uint32 `json:"next_header_size"`
	CRC               uint32 `json:"crc"`
}

// DTOCJSON represents a DTOC section in JSON
type DTOCJSON struct {
	Header  DTOCHeaderJSON  `json:"header"`
	Entries []DTOCEntryJSON `json:"entries"`
}

// ResetInfoJSON represents reset information in JSON
type ResetInfoJSON struct {
	HardReset uint32 `json:"hard_reset"`
	SoftReset uint32 `json:"soft_reset"`
	Reserved  uint32 `json:"reserved"`
}

// ToolsAreaExtendedJSON represents extended tools area in JSON
type ToolsAreaExtendedJSON struct {
	Data string `json:"data"` // Hex encoded
}

// Generic section JSON wrapper that includes both metadata and type-specific data
type SectionJSON struct {
	// Metadata fields
	Type         uint16 `json:"type"`
	TypeName     string `json:"type_name"`
	Offset       uint64 `json:"offset"`
	Size         uint32 `json:"size"`
	CRCType      string `json:"crc_type"`
	IsEncrypted  bool   `json:"is_encrypted"`
	IsDeviceData bool   `json:"is_device_data"`

	// Indicates if this section has raw data that needs binary file
	HasRawData bool `json:"has_raw_data"`

	// Type-specific data (one of these will be populated)
	ImageInfo         *ImageInfoJSON         `json:"image_info,omitempty"`
	DeviceInfo        *DevInfoJSON           `json:"device_info,omitempty"`
	MfgInfo           *MfgInfoJSON           `json:"mfg_info,omitempty"`
	HashesTable       *HashesTableJSON       `json:"hashes_table,omitempty"`
	ImageSignature    *ImageSignatureJSON    `json:"signature,omitempty"`
	PublicKeys        []PublicKeyJSON        `json:"keys,omitempty"`
	ForbiddenVersions *ForbiddenVersionsJSON `json:"forbidden_versions,omitempty"`
	FS4HWPointers     *FS4HWPointersJSON     `json:"fs4_hw_pointers,omitempty"`
	FS5HWPointers     *FS5HWPointersJSON     `json:"fs5_hw_pointers,omitempty"`
	ITOC              *ITOCJSON              `json:"itoc,omitempty"`
	DTOC              *DTOCJSON              `json:"dtoc,omitempty"`
	ResetInfo         *ResetInfoJSON         `json:"reset_info,omitempty"`
	ToolsAreaExtended *ToolsAreaExtendedJSON `json:"tools_area_extended,omitempty"`

	// DTOC sections
	VPD_R0           *VPD_R0JSON           `json:"vpd_r0,omitempty"`
	FWNVLog          *FWNVLogJSON          `json:"fw_nv_log,omitempty"`
	NVData           *NVDataJSON           `json:"nv_data,omitempty"`
	CRDumpMaskData   *CRDumpMaskDataJSON   `json:"crdump_mask_data,omitempty"`
	FWInternalUsage  *FWInternalUsageJSON  `json:"fw_internal_usage,omitempty"`
	ProgrammableHWFW *ProgrammableHWFWJSON `json:"programmable_hw_fw,omitempty"`
	DigitalCertPtr   *DigitalCertPtrJSON   `json:"digital_cert_ptr,omitempty"`
	DigitalCertRW    *DigitalCertRWJSON    `json:"digital_cert_rw,omitempty"`

	// For sections with padding
	Padding string `json:"padding,omitempty"` // Hex encoded padding data
}

// Helper functions for hex encoding/decoding
func bytesToHex(b []byte) string {
	return hex.EncodeToString(b)
}

func hexToBytes(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

// Helper to clean string from byte array (remove null terminators)
func cleanString(b []byte) string {
	s := string(b)
	if idx := strings.IndexByte(s, 0); idx >= 0 {
		s = s[:idx]
	}
	return strings.TrimRight(s, "\x00")
}

// DTOC section JSON types

// VPD_R0JSON represents VPD_R0 section data in JSON
type VPD_R0JSON struct {
	ID       string `json:"id"`
	Length   uint16 `json:"length"`
	DataSize int    `json:"data_size"`
}

// FWNVLogJSON represents FW_NV_LOG section data in JSON
type FWNVLogJSON struct {
	LogVersion uint32 `json:"log_version"`
	LogSize    uint32 `json:"log_size"`
	EntryCount uint32 `json:"entry_count"`
	DataSize   int    `json:"data_size"`
}

// NVDataJSON represents NV_DATA section data in JSON
type NVDataJSON struct {
	Version        uint32 `json:"version"`
	DataSize       uint32 `json:"data_size"`
	ActualDataSize int    `json:"actual_data_size"`
}

// CRDumpMaskDataJSON represents CRDUMP_MASK_DATA section data in JSON
type CRDumpMaskDataJSON struct {
	Version  uint32 `json:"version"`
	MaskSize uint32 `json:"mask_size"`
	DataSize int    `json:"data_size"`
}

// FWInternalUsageJSON represents FW_INTERNAL_USAGE section data in JSON
type FWInternalUsageJSON struct {
	Version  uint32 `json:"version"`
	Size     uint32 `json:"size"`
	Type     uint32 `json:"type"`
	DataSize int    `json:"data_size"`
}

// ProgrammableHWFWJSON represents PROGRAMMABLE_HW_FW section data in JSON
type ProgrammableHWFWJSON struct {
	Version     uint32 `json:"version"`
	HWType      uint32 `json:"hw_type"`
	FWSize      uint32 `json:"fw_size"`
	Checksum    uint32 `json:"checksum"`
	LoadAddress string `json:"load_address"`
	EntryPoint  string `json:"entry_point"`
	DataSize    int    `json:"data_size"`
}

// DigitalCertPtrJSON represents DIGITAL_CERT_PTR section data in JSON
type DigitalCertPtrJSON struct {
	CertType   uint32 `json:"cert_type"`
	CertOffset string `json:"cert_offset"`
	CertSize   uint32 `json:"cert_size"`
}

// DigitalCertRWJSON represents DIGITAL_CERT_RW section data in JSON
type DigitalCertRWJSON struct {
	CertType       uint32 `json:"cert_type"`
	CertSize       uint32 `json:"cert_size"`
	ValidFrom      uint64 `json:"valid_from"`
	ValidTo        uint64 `json:"valid_to"`
	ActualCertSize int    `json:"actual_cert_size"`
	CertPreview    string `json:"cert_preview,omitempty"`
}
