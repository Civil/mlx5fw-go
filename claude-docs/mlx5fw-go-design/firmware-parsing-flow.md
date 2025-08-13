# Firmware Image Handling and Parsing

## Overview

This document describes how mlx5fw-go handles and parses Mellanox firmware images, including the parsing flow, data structures, and key algorithms.

## Firmware Format Overview

### FS4 Format Structure

```
┌─────────────────────────────────────┐
│      Magic Pattern (16 bytes)       │ ← Found by scanning
├─────────────────────────────────────┤
│         ...other data...            │
├─────────────────────────────────────┤
│    HW Pointers (at magic + 0x24)   │ ← Contains addresses
├─────────────────────────────────────┤
│         ...other data...            │
├─────────────────────────────────────┤
│      BOOT2 Section (optional)       │ ← Boot loader
├─────────────────────────────────────┤
│      ITOC (Image TOC) Section       │ ← Section directory
├─────────────────────────────────────┤
│      DTOC (Device TOC) Section      │ ← Device-specific
├─────────────────────────────────────┤
│         Various Sections            │ ← As listed in ITOC
│      (DEV_INFO, IMAGE_INFO, etc)   │
├─────────────────────────────────────┤
│      TOOLS_AREA (optional)          │ ← Manufacturing data
└─────────────────────────────────────┘
```

## Parsing Flow

### 1. Initial Discovery Phase

```go
// pkg/parser/firmware_reader.go
func (fr *FirmwareReader) FindMagicPattern() (uint32, error) {
    // Scan for magic pattern: 0x4D544657 ("MTFW")
    // Can be at various offsets (0x0, 0x10000, 0x20000, etc.)
}
```

### 2. Hardware Pointers Parsing

```go
// pkg/parser/fs4/parser.go
func (p *Parser) parseHWPointers() error {
    // Read at magic_offset + 0x24
    // Parse FS4HWPointersAnnotated structure
    // Extract critical addresses:
    //   - boot2Addr
    //   - itocAddr
    //   - dtocAddr
}
```

The HW Pointers structure (from `pkg/types/hw_pointers_annotated.go`):
```go
type FS4HWPointersAnnotated struct {
    AbsAddr       uint32 `offset:"byte:0,endian:be"`
    Boot2Addr     uint32 `offset:"byte:4,endian:be"`
    BootAddrNoJmp uint32 `offset:"byte:8,endian:be"`
    ITOCAddr      uint32 `offset:"byte:12,endian:be"`
    DTOCAddr      uint32 `offset:"byte:16,endian:be"`
    // ... more fields
}
```

### 3. ITOC (Image Table of Contents) Parsing

The ITOC contains the directory of all firmware sections:

```go
// pkg/parser/toc_reader.go
func (tr *TOCReader) ReadTOC(reader io.ReaderAt, tocAddr uint32) error {
    // 1. Read and validate ITOC header
    // 2. Parse ITOC entries
    // 3. For each entry, create appropriate section
}
```

ITOC Entry structure:
```go
type ITOCEntryAnnotated struct {
    Type           uint8  `offset:"byte:0"`           // Section type
    ParamLen       uint8  `offset:"byte:1"`           // Parameter length
    Param0         uint8  `offset:"byte:2"`           // Flags (device_data, no_crc, etc.)
    Param1         uint8  `offset:"byte:3"`           // CRC type
    FlashOffsetRaw uint32 `offset:"byte:4,endian:be"` // Section offset
    SectionLenDW   uint32 `offset:"byte:8,endian:be"` // Size in DWORDs
    SectionCRC     uint16 `offset:"byte:12,endian:be"`// CRC value
}
```

### 4. Section Creation and Parsing

```go
// pkg/types/sections/factory.go
func (f *DefaultSectionFactory) CreateSection(
    sectionType uint16, 
    offset uint64, 
    size uint32,
    // ... other parameters
) (interfaces.CompleteSectionInterface, error) {
    switch sectionType {
    case types.FS4_SECT_DEV_INFO:
        return NewDeviceInfoSection(/* params */), nil
    case types.FS4_SECT_IMAGE_INFO:
        return NewImageInfoSection(/* params */), nil
    // ... other section types
    default:
        return NewGenericSection(/* params */), nil
    }
}
```

### 5. Section-Specific Parsing

Each section type has its own parsing logic:

#### Device Info Section
```go
// pkg/types/sections/device_info_section.go
func (s *DeviceInfoSection) Parse(data []byte) error {
    // Parse DevInfoAnnotated structure
    // Extract GUIDs, MACs, VSD, etc.
}
```

#### Image Info Section
```go
// pkg/types/sections/image_info_section.go
func (s *ImageInfoSection) Parse(data []byte) error {
    // Parse FWImageInfoAnnotated structure
    // Extract version, PSID, description, etc.
}
```

## CRC Verification

### CRC Types

1. **IN_ITOC_ENTRY** (0): CRC stored in ITOC entry
2. **NONE** (1): No CRC verification
3. **IN_SECTION** (2): CRC embedded at end of section data

### CRC Calculation Flow

```go
// pkg/crc/unified_handler.go
func (h *UnifiedHandler) CalculateCRC(data []byte, crcType types.CRCType) (uint32, error) {
    switch crcType {
    case types.CRCInSection:
        // Calculate CRC excluding last 4 bytes (where CRC is stored)
        return h.calculateSoftwareCRC(data[:len(data)-4])
    case types.CRCInITOCEntry:
        // Calculate CRC of entire data
        return h.calculateSoftwareCRC(data)
    }
}
```

### Special CRC Handlers

Some sections require special CRC handling:

1. **BOOT2**: Uses hardware CRC algorithm
2. **TOOLS_AREA**: Has extended format with additional CRC
3. **ITOC Header**: Has both header CRC and invariant CRC

## Encrypted Firmware Handling

When standard ITOC parsing fails:

```go
func (p *Parser) parseEncryptedFirmware() error {
    // 1. Check for specific patterns indicating encryption
    // 2. Try to parse limited sections (BOOT2, TOOLS_AREA)
    // 3. Mark firmware as encrypted
    // 4. Extract what information is available
}
```

## Binary Parsing with Annotations

The project uses a custom annotation system for binary parsing:

```go
// Example: Parsing a structure
type ImageInfoAnnotated struct {
    FWVersion [3]uint16 `offset:"byte:16,endian:be"`
    FWPSID    [16]byte  `offset:"byte:48"`
    // Fields are parsed based on offset tags
}

// Usage
info := &ImageInfoAnnotated{}
err := info.Unmarshal(binaryData)
```

### Annotation Features

1. **Byte/Bit Offsets**: `offset:"byte:16"` or `offset:"bit:128"`
2. **Endianness**: `endian:be` (big), `endian:le` (little)
3. **Bitfields**: `offset:"bit:0,len:4"` for 4-bit field
4. **Arrays**: Automatically handled for array types
5. **Reserved Fields**: `reserved:true` marks padding
6. **Dynamic Lists**: `list_size:"count_field"` for runtime-sized lists

## Error Handling

The parser uses wrapped errors for context:

```go
if err := p.parseITOC(); err != nil {
    return merry.Wrap(err).WithMessage("failed to parse ITOC")
}
```

## Memory Efficiency

1. **Lazy Loading**: Sections are parsed on-demand
2. **Reader Interface**: Uses `io.ReaderAt` to avoid loading entire file
3. **Selective Parsing**: Only requested sections are fully parsed

## Validation Steps

1. **Magic Pattern**: Verify firmware starts with valid magic
2. **CRC Verification**: Check all section CRCs
3. **Offset Validation**: Ensure sections don't overlap
4. **Size Validation**: Verify sections fit within file bounds
5. **Structure Validation**: Check required fields are present

## Extension Points

### Adding New Section Types

1. Create new section struct in `pkg/types/sections/`
2. Implement `SectionInterface`
3. Add to factory in `pkg/types/sections/factory.go`
4. Define constants in `pkg/types/section_names.go`

### Adding New Firmware Format

1. Create new parser in `pkg/parser/fs5/` (for FS5)
2. Implement `FirmwareParser` interface
3. Add format detection logic
4. Register with main parser factory