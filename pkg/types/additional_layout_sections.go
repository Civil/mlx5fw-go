package types

// Additional section structures ported from mstflint's image_layout_layouts.h (FS4)
// and fs5_image_layout_layouts.h (FS5)

// FS4ComponentAuthenticationConfiguration represents authentication configuration for FS4
// Based on image_layout_component_authentication_configuration from mstflint
type FS4ComponentAuthenticationConfiguration struct {
	AuthType        uint8  `bin:"bitfield=8"`  // bits 0-7: 0-FW_BURN_NONE, 1-SHA256, 2-SHA512, 3-RSA2048, 4-RSA4096
	Reserved1       uint8  `bin:"bitfield=17"` // bits 8-24: reserved
	CRTokenEn       uint8  `bin:"bitfield=1"`  // bit 25: challenge-response token enable
	BTCTokenEn      uint8  `bin:"bitfield=1"`  // bit 26: back to commissioning token enable
	FRCEn           uint8  `bin:"bitfield=1"`  // bit 27: factory re-configuration enable
	MLNXNVConfigEn  uint8  `bin:"bitfield=1"`  // bit 28: MLNX nvconfig enable
	VendorNVConfigEn uint8 `bin:"bitfield=1"`  // bit 29: vendor nvconfig enable
	CSTokenEn       uint8  `bin:"bitfield=1"`  // bit 30: CS token enable
	FWEn            uint8  `bin:"bitfield=1"`  // bit 31: firmware enable
}

// FS4HtocEntry represents a single entry in the HTOC (Hash Table of Contents) for FS4
// Based on image_layout_htoc_entry from mstflint
type FS4HtocEntry struct {
	HashOffset  uint16 `bin:"BE,bitfield=16"` // bits 0-15: offset to hash
	SectionType uint8  `bin:"BE"`             // bits 16-23: section type
	Reserved    uint8  `bin:"BE"`             // bits 24-31: reserved
	_           uint32 // padding to 8 bytes
}

// FS4HtocHeader represents the HTOC header for FS4
// Based on image_layout_htoc_header from mstflint
type FS4HtocHeader struct {
	Version       uint32 `bin:"BE"`             // offset 0x0
	NumOfEntries  uint8  `bin:"BE"`             // offset 0x4, bits 0-7
	HashType      uint8  `bin:"BE"`             // offset 0x4, bits 8-15
	HashSize      uint16 `bin:"BE"`             // offset 0x4, bits 16-31
	Reserved      [8]uint8                      // padding to 16 bytes
}

// FS4HtocHash represents a hash value in the HTOC for FS4
// Based on image_layout_htoc_hash from mstflint
type FS4HtocHash struct {
	HashVal [16]uint32 `bin:"BE"` // 512-bit hash value (64 bytes)
}

// FS4Htoc represents the complete Hash Table of Contents structure for FS4
// Based on image_layout_htoc from mstflint
type FS4Htoc struct {
	Header FS4HtocHeader   // offset 0x0-0xf (16 bytes)
	Entry  [28]FS4HtocEntry // offset 0x10-0xef (28 entries * 8 bytes = 224 bytes)
}

// FS4HashesTableHeader represents the hashes table header for FS4
// Based on image_layout_hashes_table_header from mstflint
type FS4HashesTableHeader struct {
	LoadAddress uint32 `bin:"BE"` // offset 0x0: hard-coded to 0
	DwSize      uint32 `bin:"BE"` // offset 0x4: num of payload DWs + 1
	CRC         uint16 `bin:"BE"` // offset 0x8: calculated over first 2 DWs
	Reserved    uint16 `bin:"BE"` // padding to 12 bytes
}

// FS4HashesTable represents the complete hashes table structure for FS4
// Based on image_layout_hashes_table from mstflint
type FS4HashesTable struct {
	Header FS4HashesTableHeader // offset 0x0-0xb (12 bytes)
	Htoc   FS4Htoc              // offset 0xc-0xfb (240 bytes)
	Hash   [28]FS4HtocHash      // offset 0xfc-0x7fb (28 * 64 = 1792 bytes)
	CRC    uint16               // offset 0x800: CRC at end
	_      [2]uint8             // padding to align
}

// FS4HwPointerEntry represents a hardware pointer entry for FS4
// Based on image_layout_hw_pointer_entry from mstflint
type FS4HwPointerEntry struct {
	Ptr uint32 `bin:"BE"` // offset 0x0: pointer
	CRC uint16 `bin:"BE"` // offset 0x4: crc16 as calculated by HW
	_   uint16 // padding to 8 bytes
}

// FS4HwPointersCarmel represents the HW pointers for Carmel architecture (FS4)
// Based on image_layout_hw_pointers_carmel from mstflint
type FS4HwPointersCarmel struct {
	BootRecordPtr               FS4HwPointerEntry // offset 0x0
	Boot2Ptr                    FS4HwPointerEntry // offset 0x8
	TocPtr                      FS4HwPointerEntry // offset 0x10
	ToolsPtr                    FS4HwPointerEntry // offset 0x18
	AuthenticationStartPointer  FS4HwPointerEntry // offset 0x20
	AuthenticationEndPointer    FS4HwPointerEntry // offset 0x28
	DigestPointer               FS4HwPointerEntry // offset 0x30
	DigestRecoveryKeyPointer    FS4HwPointerEntry // offset 0x38
	FwWindowStartPointer        FS4HwPointerEntry // offset 0x40
	FwWindowEndPointer          FS4HwPointerEntry // offset 0x48
	ImageInfoSectionPointer     FS4HwPointerEntry // offset 0x50
	ImageSignaturePointer       FS4HwPointerEntry // offset 0x58
	PublicKeyPointer            FS4HwPointerEntry // offset 0x60
	FwSecurityVersionPointer    FS4HwPointerEntry // offset 0x68
	GcmIvDeltaPointer          FS4HwPointerEntry // offset 0x70
	HashesTablePointer         FS4HwPointerEntry // offset 0x78
}

// FS5HwPointerEntry represents a hardware pointer entry for FS5
// Based on fs5_image_layout_hw_pointer_entry from mstflint
type FS5HwPointerEntry struct {
	Ptr uint32 `bin:"BE"` // offset 0x0: pointer
	CRC uint16 `bin:"BE"` // offset 0x4: crc16 as calculated by HW
	_   uint16 // padding to 8 bytes
}

// FS5HwPointersGilboa represents the HW pointers for Gilboa architecture (FS5)
// Based on fs5_image_layout_hw_pointers_gilboa from mstflint
type FS5HwPointersGilboa struct {
	PSCBCTPointer              FS5HwPointerEntry // offset 0x0
	Boot2Ptr                   FS5HwPointerEntry // offset 0x8
	TOCPtr                     FS5HwPointerEntry // offset 0x10
	ToolsPtr                   FS5HwPointerEntry // offset 0x18
	PSCBL1BCHPointer          FS5HwPointerEntry // offset 0x20
	PSCBL1Pointer             FS5HwPointerEntry // offset 0x28
	NCoreBCHPointer           FS5HwPointerEntry // offset 0x30
	Reserved                  FS5HwPointerEntry // offset 0x38
	PSCFWBCHPointer           FS5HwPointerEntry // offset 0x40
	PSCFWPointer              FS5HwPointerEntry // offset 0x48
	ImageInfoSectionPointer   FS5HwPointerEntry // offset 0x50
	ImageSignaturePointer     FS5HwPointerEntry // offset 0x58
	PublicKeyPointer          FS5HwPointerEntry // offset 0x60
	ForbiddenVersionsPointer  FS5HwPointerEntry // offset 0x68
	PSCHashesTablePointer     FS5HwPointerEntry // offset 0x70
	NCoreHashesPointer        FS5HwPointerEntry // offset 0x78
}

// FS4FilePublicKeys represents the file public keys structure (2048-bit RSA) for FS4
// Based on image_layout_file_public_keys from mstflint
type FS4FilePublicKeys struct {
	ComponentAuthConfig FS4ComponentAuthenticationConfiguration // offset 0x0
	Reserved1           [8]uint8                                // padding
	KeypairExp          uint32                                  `bin:"BE"` // offset 0xc: usually 65537
	KeypairUUID         [4]uint32                               `bin:"BE"` // offset 0x10-0x1c: UUID
	Key                 [64]uint32                              `bin:"BE"` // offset 0x20-0x11c: 2048-bit key
}

// FS4FilePublicKeys2 represents the file public keys structure (4096-bit RSA) for FS4
// Based on image_layout_file_public_keys_2 from mstflint
type FS4FilePublicKeys2 struct {
	ComponentAuthConfig FS4ComponentAuthenticationConfiguration // offset 0x0
	Reserved1           [8]uint8                                // padding
	KeypairExp          uint32                                  `bin:"BE"` // offset 0xc: usually 65537
	KeypairUUID         [4]uint32                               `bin:"BE"` // offset 0x10-0x1c: UUID
	Key                 [128]uint32                             `bin:"BE"` // offset 0x20-0x21c: 4096-bit key
}

// FS4FilePublicKeys3 represents the file public keys structure (alternate layout) for FS4
// Based on image_layout_file_public_keys_3 from mstflint
type FS4FilePublicKeys3 struct {
	KeypairExp          uint32                                  `bin:"BE"` // offset 0x0: usually 65537
	KeypairUUID         [4]uint32                               `bin:"BE"` // offset 0x4-0x10: UUID
	Key                 [128]uint32                             `bin:"BE"` // offset 0x14-0x210: 4096-bit key
	ComponentAuthConfig FS4ComponentAuthenticationConfiguration // offset 0x214
	Reserved            [8]uint8                                // padding to end
}

// FS4PublicKeysStruct represents the PUBLIC_KEYS_2048 section for FS4
// Based on image_layout_public_keys from mstflint
type FS4PublicKeysStruct struct {
	FilePublicKeys [8]FS4FilePublicKeys // 8 public keys of 288 bytes each = 2304 bytes
}

// FS4PublicKeys2Struct represents the PUBLIC_KEYS_4096 section for FS4
// Based on image_layout_public_keys_2 from mstflint
type FS4PublicKeys2Struct struct {
	FilePublicKeys2 [8]FS4FilePublicKeys2 // 8 public keys of 544 bytes each = 4352 bytes
}

// FS4PublicKeys3Struct represents the PUBLIC_KEYS_4096 section (alternate layout) for FS4
// Based on image_layout_public_keys_3 from mstflint
type FS4PublicKeys3Struct struct {
	FilePublicKeys3 [8]FS4FilePublicKeys3 // 8 public keys of 544 bytes each = 4352 bytes
}

// FS4ImageSignatureStruct represents the IMAGE_SIGNATURE_256 section structure for FS4
// Based on image_layout_image_signature from mstflint
type FS4ImageSignatureStruct struct {
	SignatureUUID [4]uint32  `bin:"BE"` // offset 0x0-0xc: time-based UUID
	KeypairUUID   [4]uint32  `bin:"BE"` // offset 0x10-0x1c: keypair UUID
	Signature     [64]uint32 `bin:"BE"` // offset 0x20-0x11c: 2048-bit signature
}

// FS4ImageSignature2Struct represents the IMAGE_SIGNATURE_512 section structure for FS4
// Based on image_layout_image_signature_2 from mstflint
type FS4ImageSignature2Struct struct {
	SignatureUUID [4]uint32   `bin:"BE"` // offset 0x0-0xc: time-based UUID
	KeypairUUID   [4]uint32   `bin:"BE"` // offset 0x10-0x1c: keypair UUID
	Signature     [128]uint32 `bin:"BE"` // offset 0x20-0x21c: 4096-bit signature
}

// FS4SecureBootSignaturesStruct represents the secure boot signatures structure with proper layout for FS4
// Based on image_layout_secure_boot_signatures from mstflint
type FS4SecureBootSignaturesStruct struct {
	BootSignature        [128]uint32 `bin:"BE"` // offset 0x0-0x1fc: boot signature
	CriticalSignature    [128]uint32 `bin:"BE"` // offset 0x200-0x3fc: critical signature
	NonCriticalSignature [128]uint32 `bin:"BE"` // offset 0x400-0x5fc: non-critical signature
}

// FS4BootVersionStruct represents the boot version structure for FS4
// Based on image_layout_boot_version from mstflint
type FS4BootVersionStruct struct {
	MinorVersion       uint8 // offset 0x0, bits 0-7
	MajorVersion       uint8 // offset 0x0, bits 8-15
	Reserved           uint8 // offset 0x0, bits 16-23
	ImageFormatVersion uint8 // offset 0x0, bits 24-31
}