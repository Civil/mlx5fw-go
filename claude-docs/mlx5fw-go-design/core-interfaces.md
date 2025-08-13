# Core Interfaces and Data Structures

## Overview

This document describes the core interfaces and data structures that form the foundation of the mlx5fw-go project. These interfaces define the contracts between different components and ensure modularity and extensibility.

## Core Interfaces

### 1. FirmwareParser Interface

Located in `pkg/interfaces/parser.go`:

```go
type FirmwareParser interface {
    // Parse reads and parses the firmware from the provided reader
    Parse(reader io.ReaderAt) error
    
    // Query returns firmware information similar to mstflint query output
    Query() (*FirmwareInfo, error)
    
    // GetFormat returns the firmware format (FS4 or FS5)
    GetFormat() int
    
    // GetSections returns all parsed sections
    GetSections() map[uint16]*Section
    
    // GetSection returns a specific section by type
    GetSection(sectionType uint16) (*Section, error)
}
```

**Purpose**: Main interface for parsing firmware files. Implementations exist for FS4 format, with FS5 support planned.

### 2. SectionInterface

Located in `pkg/interfaces/section.go`:

```go
type SectionInterface interface {
    // Type returns the section type
    Type() uint16
    
    // TypeName returns the human-readable name for this section type
    TypeName() string
    
    // Offset returns the section offset in the firmware
    Offset() uint64
    
    // Size returns the section size
    Size() uint32
    
    // CRCType returns the CRC type for this section
    CRCType() types.CRCType
    
    // HasCRC returns whether this section has CRC verification enabled
    HasCRC() bool
    
    // GetCRC returns the expected CRC value for this section
    GetCRC() uint32
    
    // CalculateCRC calculates the CRC for this section
    CalculateCRC() (uint32, error)
    
    // VerifyCRC verifies the section's CRC
    VerifyCRC() error
    
    // IsEncrypted returns whether the section is encrypted
    IsEncrypted() bool
    
    // IsDeviceData returns whether this is device-specific data
    IsDeviceData() bool
    
    // Parse parses the raw data into section-specific structures
    Parse(data []byte) error
    
    // GetRawData returns the raw section data
    GetRawData() []byte
    
    // Write writes the section data to the writer
    Write(w io.Writer) error
    
    // GetITOCEntry returns the ITOC entry for this section (may be nil)
    GetITOCEntry() *types.ITOCEntry
    
    // IsFromHWPointer returns true if this section was referenced from HW pointers
    IsFromHWPointer() bool
}
```

**Purpose**: Defines the contract all firmware sections must implement. Provides a uniform interface for section manipulation.

### 3. SectionCRCHandler Interface

Located in `pkg/interfaces/section.go`:

```go
type SectionCRCHandler interface {
    // CalculateCRC calculates the CRC for the given data
    CalculateCRC(data []byte, crcType types.CRCType) (uint32, error)
    
    // VerifyCRC verifies the CRC for the given data
    VerifyCRC(data []byte, expectedCRC uint32, crcType types.CRCType) error
    
    // HasEmbeddedCRC returns true if this handler expects CRC embedded in data
    HasEmbeddedCRC() bool
    
    // GetCRCOffset returns the offset where CRC is stored (-1 if not embedded)
    GetCRCOffset(dataLen int) int
}
```

**Purpose**: Abstraction for different CRC calculation methods. Allows section-specific CRC handling.

### 4. SectionFactory Interface

Located in `pkg/interfaces/section.go`:

```go
type SectionFactory interface {
    // CreateSection creates a new section instance for the given type
    CreateSection(sectionType uint16, offset uint64, size uint32, crcType types.CRCType,
        crc uint32, isEncrypted, isDeviceData bool, entry *types.ITOCEntry, 
        isFromHWPointer bool) (CompleteSectionInterface, error)
    
    // CreateSectionFromData creates a section and parses its data
    CreateSectionFromData(sectionType uint16, offset uint64, size uint32, 
        crcType types.CRCType, crc uint32, isEncrypted, isDeviceData bool, 
        entry *types.ITOCEntry, isFromHWPointer bool, data []byte) 
        (CompleteSectionInterface, error)
}
```

**Purpose**: Factory pattern for creating section instances based on type.

## Core Data Structures

### 1. FirmwareInfo

Located in `pkg/interfaces/parser.go`:

```go
type FirmwareInfo struct {
    // Basic information
    Format          string
    FormatVersion   int
    
    // Version information
    FWVersion       string
    FWReleaseDate   string
    MICVersion      string
    ProductVersion  string
    
    // Product information
    PartNumber      string
    Description     string
    PSID            string
    OrigPSID        string  // Original PSID (shown when different from PSID)
    PRSName         string
    
    // ROM information
    RomInfo         []RomInfo
    
    // GUID/MAC information
    BaseGUID        uint64
    BaseGUIDNum     int
    BaseMAC         uint64
    BaseMACNum      int
    
    // Additional GUID/MAC for dual format (encrypted firmware)
    BaseGUID2       uint64
    BaseGUID2Num    int
    BaseMAC2        uint64
    BaseMAC2Num     int
    GUIDStep        uint8
    MACStep         uint8
    UseDualFormat   bool  // Whether to display GUID1/GUID2 format
    
    // VSD information
    ImageVSD        string
    DeviceVSD       string
    
    // Security information
    SecurityAttrs   string
    SecurityVer     int
    IsEncrypted     bool
    
    // Device information
    DeviceID        uint16
    VendorID        uint16
    
    // Size information
    ImageSize       uint64
    ChunkSize       uint64
    
    // Additional metadata
    Sections        []SectionInfo
}
```

**Purpose**: Contains all firmware metadata extracted during parsing. Used for query output.

### 2. Section

Located in `pkg/interfaces/parser.go`:

```go
type Section struct {
    Type            uint16
    Offset          uint64
    Size            uint32
    Data            []byte
    CRCType         types.CRCType
    CRC             uint32
    IsEncrypted     bool
    IsDeviceData    bool
    Entry           *types.ITOCEntry
    IsFromHWPointer bool  // True if section was discovered from HW pointer
}
```

**Purpose**: Simple section representation used in parser interface.

### 3. BaseSection

Located in `pkg/interfaces/section.go`:

```go
type BaseSection struct {
    SectionType       types.SectionType
    SectionOffset     uint64
    SectionSize       uint32
    SectionCRCType    types.CRCType
    SectionCRC        uint32
    EncryptedFlag     bool
    DeviceDataFlag    bool
    HasRawData        bool
    FromHWPointerFlag bool
    rawData           []byte
    entry             *types.ITOCEntry
    crcHandler        SectionCRCHandler
    hasCRC            bool
}
```

**Purpose**: Base implementation providing common functionality for all section types.

### 4. ITOCEntry

Located in `pkg/types/itoc_annotated.go`:

```go
type ITOCEntryAnnotated struct {
    Type           uint8  `json:"type" offset:"byte:0"`
    ParamLen       uint8  `json:"param_len" offset:"byte:1"`
    Param0         uint8  `json:"param0" offset:"byte:2"`
    Param1         uint8  `json:"param1" offset:"byte:3"`
    FlashOffsetRaw uint32 `json:"flash_offset_raw" offset:"byte:4,endian:be"`
    SectionLenDW   uint32 `json:"section_len_dw" offset:"byte:8,endian:be"`
    SectionCRC     uint16 `json:"section_crc" offset:"byte:12,endian:be"`
    Padding        uint16 `json:"padding" offset:"byte:14,reserved:true"`
}
```

**Purpose**: Represents an entry in the Image Table of Contents (ITOC).

### 5. ITOCHeader

Located in `pkg/types/itoc_annotated.go`:

```go
type ITOCHeaderAnnotated struct {
    Signature  uint32 `json:"signature" offset:"byte:0,endian:be"`
    Version    uint8  `json:"version" offset:"byte:4"`
    Reserved   uint8  `json:"reserved" offset:"byte:5,reserved:true"`
    EntryNum   uint16 `json:"entry_num" offset:"byte:6,endian:be"`
    NextPtrDW  uint32 `json:"next_ptr_dw" offset:"byte:8,endian:be"`
    HeaderCRC  uint16 `json:"header_crc" offset:"byte:12,endian:be"`
    InvarCRC   uint16 `json:"invar_crc" offset:"byte:14,endian:be"`
}
```

**Purpose**: Header structure for ITOC sections.

## Key Type Definitions

### 1. SectionType

Located in `pkg/types/section_names.go`:

Common section types include:
- `FS4_SECT_BOOT2` (0x7)
- `FS4_SECT_DEV_INFO` (0x8)
- `FS4_SECT_FW_INFO` (0x1)
- `FS4_SECT_IMAGE_INFO` (0xD)
- `FS4_SECT_MFG_INFO` (0x5)
- `FS4_SECT_TOOLS_AREA` (0x14)
- `FS4_SECT_HASHES_TABLE` (0x24)

### 2. CRCType

Located in `pkg/types/types.go`:

```go
type CRCType uint8

const (
    CRCInITOCEntry CRCType = 0  // CRC stored in ITOC entry
    CRCNone        CRCType = 1  // No CRC verification
    CRCInSection   CRCType = 2  // CRC stored at end of section
)
```

### 3. FirmwareFormat

Located in `pkg/types/types.go`:

```go
type FirmwareFormat int

const (
    FormatUnknown FirmwareFormat = iota
    FormatFS4
    FormatFS5
)
```

## Design Principles

1. **Interface Segregation**: Interfaces are focused and specific to their purpose
2. **Dependency Inversion**: High-level modules depend on interfaces, not concrete implementations
3. **Single Responsibility**: Each interface and data structure has a clear, single purpose
4. **Open/Closed**: The design is open for extension (new section types) but closed for modification

## Extension Guidelines

To add new functionality:

1. **New Section Type**: 
   - Create a new struct embedding `BaseSection`
   - Implement any specialized parsing in the `Parse` method
   - Register with the section factory

2. **New CRC Algorithm**:
   - Implement the `SectionCRCHandler` interface
   - Add to the appropriate handler registry

3. **New Firmware Format**:
   - Implement the `FirmwareParser` interface
   - Add format detection logic