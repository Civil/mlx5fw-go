# MLX5FW-GO Architecture

## Overview

The mlx5fw-go project follows a layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                    CLI Layer (cmd/mlx5fw-go)                │
├─────────────────────────────────────────────────────────────┤
│                    Business Logic Layer                      │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │   Parser    │  │   Extract    │  │   Reassemble     │  │
│  │   (pkg/     │  │   (pkg/      │  │   (pkg/          │  │
│  │   parser)   │  │   extract)   │  │   reassemble)    │  │
│  └─────────────┘  └──────────────┘  └──────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                    Core Components                           │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ Interfaces  │  │     CRC      │  │   Annotations    │  │
│  │   (pkg/     │  │   (pkg/crc)  │  │     (pkg/        │  │
│  │ interfaces) │  │              │  │   annotations)   │  │
│  └─────────────┘  └──────────────┘  └──────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                    Data Types Layer                          │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Types (pkg/types)                       │   │
│  │  - Base types and constants                         │   │
│  │  - Section definitions (pkg/types/sections)         │   │
│  │  - Annotated structures for binary parsing          │   │
│  └─────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│                    Utility Layer                             │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │   Utils     │  │    Errors    │  │      Errs        │  │
│  │   (pkg/     │  │    (pkg/     │  │     (pkg/errs)   │  │
│  │   utils)    │  │   errors)    │  │                  │  │
│  └─────────────┘  └──────────────┘  └──────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Layer Descriptions

### 1. CLI Layer (cmd/mlx5fw-go)

The command-line interface layer provides user-facing commands:

- **main.go**: Entry point, command setup, and global flag handling
- **query.go**: Firmware information query command
- **sections.go**: Section listing and extraction commands
- **extract.go**: Full firmware extraction functionality
- **reassemble.go**: Firmware reassembly from extracted sections
- **replace_section_*.go**: Section replacement implementations
- **json_output.go**: JSON formatting for machine-readable output

### 2. Business Logic Layer

Core business logic implementations:

#### Parser (pkg/parser)
- **FirmwareReader**: Low-level binary reading operations
- **FS4 Parser**: FS4 format-specific parsing logic
- **TOCReader**: Table of Contents (ITOC/DTOC) parsing
- **CRCCalculator**: CRC calculation and verification

#### Extract (pkg/extract)
- **Extractor**: Section extraction logic

#### Reassemble (pkg/reassemble)
- **Reassembler**: Firmware reconstruction from sections
- **Metadata handling**: Section metadata management

### 3. Core Components

#### Interfaces (pkg/interfaces)
Defines contracts for the entire system:
- **FirmwareParser**: Main parsing interface
- **SectionInterface**: Base interface for all sections
- **CRCHandler**: CRC calculation interface
- **SectionFactory**: Section creation factory pattern

#### CRC (pkg/crc)
Multiple CRC implementations:
- **BaseHandler**: Common CRC functionality
- **UnifiedHandler**: General-purpose CRC handler
- **Boot2Handler**: BOOT2 section-specific CRC
- **ToolsAreaHandler**: TOOLS_AREA section-specific CRC

#### Annotations (pkg/annotations)
Custom binary parsing framework:
- Tag-based struct field annotations
- Support for bitfields, arrays, and dynamic lists
- Endianness handling
- Marshal/Unmarshal operations

### 4. Data Types Layer (pkg/types)

All data structures and type definitions:

#### Base Types
- **FirmwareFormat**: FS4/FS5 format enumeration
- **CRCType**: CRC verification types
- **SectionType**: Section type constants
- **Binary structures**: ITOCEntry, ITOCHeader, HWPointers, etc.

#### Section Types (pkg/types/sections)
Specific section implementations:
- **GenericSection**: Default section implementation
- **DeviceInfoSection**: Device information parsing
- **ImageInfoSection**: Image metadata parsing
- **Boot2Section**: Boot loader section
- **HashesTableSection**: Security hashes
- **SignatureSection**: Digital signatures
- **ToolsAreaExtendedSection**: Tools area with extended data

### 5. Utility Layer

Supporting utilities:
- **Utils**: Common helper functions
- **Errors/Errs**: Error handling and custom error types

## Key Design Patterns

### 1. Factory Pattern
Used for creating section instances based on type:
```go
type SectionFactory interface {
    CreateSection(sectionType uint16, ...) (CompleteSectionInterface, error)
}
```

### 2. Strategy Pattern
CRC handlers implement different CRC calculation strategies:
```go
type SectionCRCHandler interface {
    CalculateCRC(data []byte, crcType types.CRCType) (uint32, error)
    VerifyCRC(data []byte, expectedCRC uint32, crcType types.CRCType) error
}
```

### 3. Template Method Pattern
BaseSection provides common implementation with hooks for specialization:
```go
type BaseSection struct {
    // Common fields and methods
}

func (b *BaseSection) Parse(data []byte) error {
    // Default implementation that can be overridden
}
```

### 4. Functional Options Pattern
Used for configurable object creation:
```go
type SectionOption func(*BaseSection)

func WithCRC(crcType types.CRCType, crc uint32) SectionOption {
    return func(s *BaseSection) {
        s.SectionCRCType = crcType
        s.SectionCRC = crc
    }
}
```

## Data Flow

1. **Parsing Flow**:
   ```
   Binary File → FirmwareReader → Parser → Sections → Query Output
   ```

2. **Extraction Flow**:
   ```
   Parsed Sections → Extractor → Individual Section Files + Metadata
   ```

3. **Reassembly Flow**:
   ```
   Section Files + Metadata → Reassembler → New Firmware Binary
   ```

## Extension Points

The architecture is designed to be extensible:

1. **New Section Types**: Implement SectionInterface in pkg/types/sections
2. **New CRC Algorithms**: Add handlers in pkg/crc
3. **New Firmware Formats**: Add parser implementations in pkg/parser
4. **New Commands**: Add command files in cmd/mlx5fw-go