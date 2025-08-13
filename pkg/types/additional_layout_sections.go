package types

import (
	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// FS4HashesTableHeader represents the hashes table header for FS4 with annotations
type FS4HashesTableHeader struct {
	LoadAddress uint32 `offset:"0x0,endian:be"` // offset 0x0: hard-coded to 0
	DwSize      uint32 `offset:"0x4,endian:be"` // offset 0x4: num of payload DWs + 1
	CRC         uint16 `offset:"0x8,endian:be"` // offset 0x8: calculated over first 2 DWs
	Reserved    uint16 `offset:"0xa,endian:be"` // padding to 12 bytes
}

// Unmarshal unmarshals binary data
func (h *FS4HashesTableHeader) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// Marshal marshals to binary data
func (h *FS4HashesTableHeader) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// FS4HtocEntry represents a single entry in the HTOC (Hash Table of Contents) for FS4 with annotations
type FS4HtocEntry struct {
	HashOffset  uint16 `offset:"0x0,endian:be"` // bits 0-15: offset to hash
	SectionType uint8  `offset:"0x2,endian:be"` // bits 16-23: section type
	Reserved    uint8  `offset:"0x3,endian:be"` // bits 24-31: reserved
	_           uint32 `offset:"0x4,endian:be"` // padding to 8 bytes
}

// Unmarshal unmarshals binary data
func (h *FS4HtocEntry) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// Marshal marshals to binary data
func (h *FS4HtocEntry) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// FS4HtocHeader represents the HTOC header for FS4 with annotations
type FS4HtocHeader struct {
	Version      uint32   `offset:"0x0,endian:be"` // offset 0x0
	NumOfEntries uint8    `offset:"0x4,endian:be"` // offset 0x4, bits 0-7
	HashType     uint8    `offset:"0x5,endian:be"` // offset 0x4, bits 8-15
	HashSize     uint16   `offset:"0x6,endian:be"` // offset 0x4, bits 16-31
	Reserved     [8]uint8 `offset:"0x8"`           // padding to 16 bytes
}

// Unmarshal unmarshals binary data
func (h *FS4HtocHeader) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// Marshal marshals to binary data
func (h *FS4HtocHeader) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// FS4HtocHash represents a hash value in the HTOC for FS4 with annotations
type FS4HtocHash struct {
	HashVal [16]uint32 `offset:"0x0,endian:be"` // 512-bit hash value (64 bytes)
}

// Unmarshal unmarshals binary data
func (h *FS4HtocHash) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// Marshal marshals to binary data
func (h *FS4HtocHash) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// FS4Htoc represents the complete Hash Table of Contents structure for FS4 with annotations
type FS4Htoc struct {
	Header FS4HtocHeader    `offset:"0x0"`  // offset 0x0-0xf (16 bytes)
	Entry  [28]FS4HtocEntry `offset:"0x10"` // offset 0x10-0xef (28 entries * 8 bytes = 224 bytes)
}

// Unmarshal unmarshals binary data
func (h *FS4Htoc) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// Marshal marshals to binary data
func (h *FS4Htoc) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// FS4HashesTable represents the complete hashes table structure for FS4 with annotations
type FS4HashesTable struct {
	Header FS4HashesTableHeader `offset:"0x0"`             // offset 0x0-0xb (12 bytes)
	Htoc   FS4Htoc              `offset:"0xc"`             // offset 0xc-0xfb (240 bytes)
	Hash   [28]FS4HtocHash      `offset:"0xfc"`            // offset 0xfc-0x7fb (28 * 64 = 1792 bytes)
	CRC    uint32               `offset:"0x7fc,endian:be"` // offset 0x7fc: CRC at end
}

// Unmarshal unmarshals binary data
func (h *FS4HashesTable) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// Marshal marshals to binary data
func (h *FS4HashesTable) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// FS4HwPointerEntry represents a hardware pointer entry for FS4 with annotations
type FS4HwPointerEntry struct {
	Ptr uint32 `offset:"0x0,endian:be"` // offset 0x0: pointer
	CRC uint16 `offset:"0x4,endian:be"` // offset 0x4: crc16 as calculated by HW
	_   uint16 `offset:"0x6,endian:be"` // padding to 8 bytes
}

// Unmarshal unmarshals binary data
func (h *FS4HwPointerEntry) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// Marshal marshals to binary data
func (h *FS4HwPointerEntry) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// FS4HwPointersCarmel represents the HW pointers for Carmel architecture (FS4) with annotations
type FS4HwPointersCarmel struct {
	BootRecordPtr              FS4HwPointerEntry `offset:"0x0"`  // offset 0x0
	Boot2Ptr                   FS4HwPointerEntry `offset:"0x8"`  // offset 0x8
	TocPtr                     FS4HwPointerEntry `offset:"0x10"` // offset 0x10
	ToolsPtr                   FS4HwPointerEntry `offset:"0x18"` // offset 0x18
	AuthenticationStartPointer FS4HwPointerEntry `offset:"0x20"` // offset 0x20
	AuthenticationEndPointer   FS4HwPointerEntry `offset:"0x28"` // offset 0x28
	DigestPointer              FS4HwPointerEntry `offset:"0x30"` // offset 0x30
	DigestRecoveryKeyPointer   FS4HwPointerEntry `offset:"0x38"` // offset 0x38
	FwWindowStartPointer       FS4HwPointerEntry `offset:"0x40"` // offset 0x40
	FwWindowEndPointer         FS4HwPointerEntry `offset:"0x48"` // offset 0x48
	ImageInfoSectionPointer    FS4HwPointerEntry `offset:"0x50"` // offset 0x50
	ImageSignaturePointer      FS4HwPointerEntry `offset:"0x58"` // offset 0x58
	PublicKeyPointer           FS4HwPointerEntry `offset:"0x60"` // offset 0x60
	FwSecurityVersionPointer   FS4HwPointerEntry `offset:"0x68"` // offset 0x68
	GcmIvDeltaPointer          FS4HwPointerEntry `offset:"0x70"` // offset 0x70
	HashesTablePointer         FS4HwPointerEntry `offset:"0x78"` // offset 0x78
}

// Unmarshal unmarshals binary data
func (h *FS4HwPointersCarmel) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// Marshal marshals to binary data
func (h *FS4HwPointersCarmel) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// FS5HwPointerEntry represents a hardware pointer entry for FS5 with annotations
type FS5HwPointerEntry struct {
	Ptr uint32 `offset:"0x0,endian:be"` // offset 0x0: pointer
	CRC uint16 `offset:"0x4,endian:be"` // offset 0x4: crc16 as calculated by HW
	_   uint16 `offset:"0x6,endian:be"` // padding to 8 bytes
}

// Unmarshal unmarshals binary data
func (h *FS5HwPointerEntry) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// Marshal marshals to binary data
func (h *FS5HwPointerEntry) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// FS5HwPointersGilboa represents the HW pointers for Gilboa architecture (FS5) with annotations
type FS5HwPointersGilboa struct {
	PSCBCTPointer            FS5HwPointerEntry `offset:"0x0"`  // offset 0x0
	Boot2Ptr                 FS5HwPointerEntry `offset:"0x8"`  // offset 0x8
	TOCPtr                   FS5HwPointerEntry `offset:"0x10"` // offset 0x10
	ToolsPtr                 FS5HwPointerEntry `offset:"0x18"` // offset 0x18
	PSCBL1BCHPointer         FS5HwPointerEntry `offset:"0x20"` // offset 0x20
	PSCBL1Pointer            FS5HwPointerEntry `offset:"0x28"` // offset 0x28
	NCoreBCHPointer          FS5HwPointerEntry `offset:"0x30"` // offset 0x30
	Reserved                 FS5HwPointerEntry `offset:"0x38"` // offset 0x38
	PSCFWBCHPointer          FS5HwPointerEntry `offset:"0x40"` // offset 0x40
	PSCFWPointer             FS5HwPointerEntry `offset:"0x48"` // offset 0x48
	ImageInfoSectionPointer  FS5HwPointerEntry `offset:"0x50"` // offset 0x50
	ImageSignaturePointer    FS5HwPointerEntry `offset:"0x58"` // offset 0x58
	PublicKeyPointer         FS5HwPointerEntry `offset:"0x60"` // offset 0x60
	ForbiddenVersionsPointer FS5HwPointerEntry `offset:"0x68"` // offset 0x68
	PSCHashesTablePointer    FS5HwPointerEntry `offset:"0x70"` // offset 0x70
	NCoreHashesPointer       FS5HwPointerEntry `offset:"0x78"` // offset 0x78
}

// Unmarshal unmarshals binary data
func (h *FS5HwPointersGilboa) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, h)
}

// Marshal marshals to binary data
func (h *FS5HwPointersGilboa) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(h)
}

// FS4ComponentAuthenticationConfiguration represents authentication configuration for FS4 with annotations
type FS4ComponentAuthenticationConfiguration struct {
	AuthType         uint8  `offset:"bit:0,len:8,endian:be"`  // bits 0-7: 0-FW_BURN_NONE, 1-SHA256, 2-SHA512, 3-RSA2048, 4-RSA4096
	Reserved1        uint32 `offset:"bit:8,len:17,endian:be"` // bits 8-24: reserved
	CRTokenEn        uint8  `offset:"bit:25,len:1,endian:be"` // bit 25: challenge-response token enable
	BTCTokenEn       uint8  `offset:"bit:26,len:1,endian:be"` // bit 26: back to commissioning token enable
	FRCEn            uint8  `offset:"bit:27,len:1,endian:be"` // bit 27: factory re-configuration enable
	MLNXNVConfigEn   uint8  `offset:"bit:28,len:1,endian:be"` // bit 28: MLNX nvconfig enable
	VendorNVConfigEn uint8  `offset:"bit:29,len:1,endian:be"` // bit 29: vendor nvconfig enable
	CSTokenEn        uint8  `offset:"bit:30,len:1,endian:be"` // bit 30: CS token enable
	FWEn             uint8  `offset:"bit:31,len:1,endian:be"` // bit 31: firmware enable
}

// Unmarshal unmarshals binary data
func (c *FS4ComponentAuthenticationConfiguration) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, c)
}

// Marshal marshals to binary data
func (c *FS4ComponentAuthenticationConfiguration) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(c)
}

// FS4FilePublicKeys represents the file public keys structure (2048-bit RSA) for FS4 with annotations
type FS4FilePublicKeys struct {
	ComponentAuthConfig FS4ComponentAuthenticationConfiguration `offset:"0x0"`            // offset 0x0
	Reserved1           [8]uint8                                `offset:"0x4"`            // padding
	KeypairExp          uint32                                  `offset:"0xc,endian:be"`  // offset 0xc: usually 65537
	KeypairUUID         [4]uint32                               `offset:"0x10,endian:be"` // offset 0x10-0x1c: UUID
	Key                 [64]uint32                              `offset:"0x20,endian:be"` // offset 0x20-0x11c: 2048-bit key
}

// Unmarshal unmarshals binary data
func (f *FS4FilePublicKeys) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, f)
}

// Marshal marshals to binary data
func (f *FS4FilePublicKeys) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(f)
}

// FS4FilePublicKeys2 represents the file public keys structure (4096-bit RSA) for FS4 with annotations
type FS4FilePublicKeys2 struct {
	ComponentAuthConfig FS4ComponentAuthenticationConfiguration `offset:"0x0"`            // offset 0x0
	Reserved1           [8]uint8                                `offset:"0x4"`            // padding
	KeypairExp          uint32                                  `offset:"0xc,endian:be"`  // offset 0xc: usually 65537
	KeypairUUID         [4]uint32                               `offset:"0x10,endian:be"` // offset 0x10-0x1c: UUID
	Key                 [128]uint32                             `offset:"0x20,endian:be"` // offset 0x20-0x21c: 4096-bit key
}

// Unmarshal unmarshals binary data
func (f *FS4FilePublicKeys2) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, f)
}

// Marshal marshals to binary data
func (f *FS4FilePublicKeys2) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(f)
}

// FS4FilePublicKeys3 represents the file public keys structure (alternate layout) for FS4 with annotations
type FS4FilePublicKeys3 struct {
	KeypairExp          uint32                                  `offset:"0x0,endian:be"`  // offset 0x0: usually 65537
	KeypairUUID         [4]uint32                               `offset:"0x4,endian:be"`  // offset 0x4-0x10: UUID
	Key                 [128]uint32                             `offset:"0x14,endian:be"` // offset 0x14-0x210: 4096-bit key
	ComponentAuthConfig FS4ComponentAuthenticationConfiguration `offset:"0x214"`          // offset 0x214
	Reserved            [8]uint8                                `offset:"0x218"`          // padding to end
}

// Unmarshal unmarshals binary data
func (f *FS4FilePublicKeys3) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, f)
}

// Marshal marshals to binary data
func (f *FS4FilePublicKeys3) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(f)
}

// FS4PublicKeysStruct represents the PUBLIC_KEYS_2048 section for FS4 with annotations
type FS4PublicKeysStruct struct {
	FilePublicKeys [8]FS4FilePublicKeys `offset:"0x0"` // 8 public keys of 288 bytes each = 2304 bytes
}

// Unmarshal unmarshals binary data
func (p *FS4PublicKeysStruct) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, p)
}

// Marshal marshals to binary data
func (p *FS4PublicKeysStruct) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(p)
}

// FS4PublicKeys2Struct represents the PUBLIC_KEYS_4096 section for FS4 with annotations
type FS4PublicKeys2Struct struct {
	FilePublicKeys2 [8]FS4FilePublicKeys2 `offset:"0x0"` // 8 public keys of 544 bytes each = 4352 bytes
}

// Unmarshal unmarshals binary data
func (p *FS4PublicKeys2Struct) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, p)
}

// Marshal marshals to binary data
func (p *FS4PublicKeys2Struct) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(p)
}

// FS4PublicKeys3Struct represents the PUBLIC_KEYS_4096 section (alternate layout) for FS4 with annotations
type FS4PublicKeys3Struct struct {
	FilePublicKeys3 [8]FS4FilePublicKeys3 `offset:"0x0"` // 8 public keys of 544 bytes each = 4352 bytes
}

// Unmarshal unmarshals binary data
func (p *FS4PublicKeys3Struct) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, p)
}

// Marshal marshals to binary data
func (p *FS4PublicKeys3Struct) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(p)
}

// FS4ImageSignatureStruct represents the IMAGE_SIGNATURE_256 section structure for FS4 with annotations
type FS4ImageSignatureStruct struct {
	SignatureUUID [4]uint32  `offset:"0x0,endian:be"`  // offset 0x0-0xc: time-based UUID
	KeypairUUID   [4]uint32  `offset:"0x10,endian:be"` // offset 0x10-0x1c: keypair UUID
	Signature     [64]uint32 `offset:"0x20,endian:be"` // offset 0x20-0x11c: 2048-bit signature
}

// Unmarshal unmarshals binary data
func (i *FS4ImageSignatureStruct) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, i)
}

// Marshal marshals to binary data
func (i *FS4ImageSignatureStruct) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(i)
}

// FS4ImageSignature2Struct represents the IMAGE_SIGNATURE_512 section structure for FS4 with annotations
type FS4ImageSignature2Struct struct {
	SignatureUUID [4]uint32   `offset:"0x0,endian:be"`  // offset 0x0-0xc: time-based UUID
	KeypairUUID   [4]uint32   `offset:"0x10,endian:be"` // offset 0x10-0x1c: keypair UUID
	Signature     [128]uint32 `offset:"0x20,endian:be"` // offset 0x20-0x21c: 4096-bit signature
}

// Unmarshal unmarshals binary data
func (i *FS4ImageSignature2Struct) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, i)
}

// Marshal marshals to binary data
func (i *FS4ImageSignature2Struct) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(i)
}

// FS4SecureBootSignaturesStruct represents the secure boot signatures structure with proper layout for FS4 with annotations
type FS4SecureBootSignaturesStruct struct {
	BootSignature        [128]uint32 `offset:"0x0,endian:be"`   // offset 0x0-0x1fc: boot signature
	CriticalSignature    [128]uint32 `offset:"0x200,endian:be"` // offset 0x200-0x3fc: critical signature
	NonCriticalSignature [128]uint32 `offset:"0x400,endian:be"` // offset 0x400-0x5fc: non-critical signature
}

// Unmarshal unmarshals binary data
func (s *FS4SecureBootSignaturesStruct) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, s)
}

// Marshal marshals to binary data
func (s *FS4SecureBootSignaturesStruct) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(s)
}

// FS4BootVersionStruct represents the boot version structure for FS4 with annotations
type FS4BootVersionStruct struct {
	MinorVersion       uint8 `offset:"bit:0,len:8,endian:be"`  // offset 0x0, bits 0-7
	MajorVersion       uint8 `offset:"bit:8,len:8,endian:be"`  // offset 0x0, bits 8-15
	Reserved           uint8 `offset:"bit:16,len:8,endian:be"` // offset 0x0, bits 16-23
	ImageFormatVersion uint8 `offset:"bit:24,len:8,endian:be"` // offset 0x0, bits 24-31
}

// Unmarshal unmarshals binary data
func (b *FS4BootVersionStruct) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, b)
}

// Marshal marshals to binary data
func (b *FS4BootVersionStruct) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(b)
}
