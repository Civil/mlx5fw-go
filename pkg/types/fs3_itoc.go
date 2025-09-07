package types

// FS3 ITOC structures (CIB firmware layouts)
// Based on reference/mstflint/tools_layouts/cibfw_layouts.h

import (
    "github.com/Civil/mlx5fw-go/pkg/annotations"
)

// FS3ITOCHeader represents the FS3/CIB ITOC header (32 bytes)
// Fields and offsets mirror cibfw_itoc_header
type FS3ITOCHeader struct {
    Signature0   uint32 `offset:"byte:0,endian:be"`  // 0x0: should be ITOC signature ("ITOC")
    Signature1   uint32 `offset:"byte:4,endian:be"`  // 0x4: 0x04081516
    Signature2   uint32 `offset:"byte:8,endian:be"`  // 0x8: 0x2342cafa
    Signature3   uint32 `offset:"byte:12,endian:be"` // 0xC: 0xbacafe00
    Version      uint8  `offset:"bit:152,len:8,endian:be"` // per cibfw_itoc_header_unpack: offset=152 bits
    // 0x11..0x1B reserved
    ITOCEntryCRC uint16 `offset:"byte:28,endian:be"` // 0x1C: CRC over header (except itself)
}

func (h *FS3ITOCHeader) Unmarshal(data []byte) error { return annotations.UnmarshalStruct(data, h) }
func (h *FS3ITOCHeader) Marshal() ([]byte, error)   { return annotations.MarshalStruct(h) }

// FS3ITOCEntry represents the FS3/CIB ITOC entry (32 bytes)
// Matches cibfw_itoc_entry layout
type FS3ITOCEntry struct {
    // Per cibfw_itoc_entry_{pack,unpack} offsets
    // size: offset=8 bits, len=22; type: offset=0, len=8
    SizeDwords uint32 `offset:"bit:8,len:22,endian:be"`
    Type       uint8  `offset:"bit:0,len:8,endian:be"`

    // param0: offset=34,len=30; flags at 33 and 32
    Param0       uint32 `offset:"bit:34,len:30,endian:be"`
    CacheLineCRC bool   `offset:"bit:33,endian:be"`
    ZippedImage  bool   `offset:"bit:32,endian:be"`

    // param1: offset=64, 4 bytes
    Param1 uint32 `offset:"bit:64,len:32,endian:be"`

    // flash_addr: offset=161,len=29; relative_addr: offset=160,len=1
    FlashAddrDwords uint32 `offset:"bit:161,len:29,endian:be"`
    RelativeAddr    bool   `offset:"bit:160,endian:be"`

    // section_crc: offset=208,len=16; no_crc: offset=207,len=1; device_data: offset=206,len=1
    SectionCRC uint16 `offset:"bit:208,len:16,endian:be"`
    NoCRC      bool   `offset:"bit:207,endian:be"`
    DeviceData bool   `offset:"bit:206,endian:be"`

    // itoc_entry_crc: offset=240,len=16
    ITOCEntryCRC uint16 `offset:"bit:240,len:16,endian:be"`
}

func (e *FS3ITOCEntry) Unmarshal(data []byte) error { return annotations.UnmarshalStruct(data, e) }
func (e *FS3ITOCEntry) Marshal() ([]byte, error)   { return annotations.MarshalStruct(e) }
