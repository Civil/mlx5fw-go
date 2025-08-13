# Architecture Diagrams

## System Architecture Overview

```mermaid
graph TB
    subgraph "CLI Layer"
        CLI[mlx5fw-go CLI]
        CLI --> CMD_QUERY[query command]
        CLI --> CMD_SECTIONS[sections command]
        CLI --> CMD_EXTRACT[extract command]
        CLI --> CMD_REASSEMBLE[reassemble command]
    end

    subgraph "Business Logic"
        PARSER[Parser<br/>pkg/parser]
        EXTRACTOR[Extractor<br/>pkg/extract]
        REASSEMBLER[Reassembler<br/>pkg/reassemble]
        SECTION_MGR[Section Manager<br/>pkg/section]
    end

    subgraph "Core Components"
        INTERFACES[Interfaces<br/>pkg/interfaces]
        CRC_CALC[CRC Calculator<br/>pkg/crc]
        ANNOTATIONS[Annotations<br/>pkg/annotations]
    end

    subgraph "Data Layer"
        TYPES[Types & Structures<br/>pkg/types]
        SECTIONS[Section Types<br/>pkg/types/sections]
    end

    subgraph "Utilities"
        UTILS[Utils<br/>pkg/utils]
        ERRORS[Error Handling<br/>pkg/errors]
    end

    CMD_QUERY --> PARSER
    CMD_SECTIONS --> PARSER
    CMD_EXTRACT --> EXTRACTOR
    CMD_REASSEMBLE --> REASSEMBLER

    PARSER --> INTERFACES
    EXTRACTOR --> INTERFACES
    REASSEMBLER --> INTERFACES
    SECTION_MGR --> INTERFACES

    PARSER --> CRC_CALC
    PARSER --> ANNOTATIONS
    EXTRACTOR --> SECTIONS
    REASSEMBLER --> SECTIONS

    INTERFACES --> TYPES
    SECTIONS --> TYPES
    CRC_CALC --> TYPES

    PARSER --> UTILS
    EXTRACTOR --> UTILS
    REASSEMBLER --> UTILS

    style CLI fill:#f9f,stroke:#333,stroke-width:4px
    style INTERFACES fill:#bbf,stroke:#333,stroke-width:2px
    style TYPES fill:#bfb,stroke:#333,stroke-width:2px
```

## Firmware Parsing Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Parser
    participant FirmwareReader
    participant TOCReader
    participant SectionFactory
    participant CRCHandler

    User->>CLI: mlx5fw-go query -f firmware.bin
    CLI->>Parser: Parse(firmware.bin)
    
    Parser->>FirmwareReader: FindMagicPattern()
    FirmwareReader-->>Parser: magic_offset
    
    Parser->>FirmwareReader: ReadSection(hw_pointers_offset)
    FirmwareReader-->>Parser: hw_pointers_data
    
    Parser->>Parser: parseHWPointers()
    Note over Parser: Extract ITOC/DTOC addresses
    
    Parser->>TOCReader: ReadTOC(itoc_addr)
    TOCReader->>FirmwareReader: ReadSection(itoc_header)
    FirmwareReader-->>TOCReader: header_data
    
    loop For each ITOC entry
        TOCReader->>SectionFactory: CreateSection(type, offset, size)
        SectionFactory-->>TOCReader: section_instance
        TOCReader->>FirmwareReader: ReadSection(section_offset)
        FirmwareReader-->>TOCReader: section_data
        TOCReader->>CRCHandler: VerifyCRC(section_data)
        CRCHandler-->>TOCReader: crc_valid
    end
    
    TOCReader-->>Parser: sections_map
    Parser->>Parser: Query()
    Parser-->>CLI: FirmwareInfo
    CLI-->>User: Display query results
```

## Section Creation and Management

```mermaid
classDiagram
    class SectionInterface {
        <<interface>>
        +Type() uint16
        +TypeName() string
        +Offset() uint64
        +Size() uint32
        +Parse(data []byte) error
        +VerifyCRC() error
        +GetRawData() []byte
    }

    class BaseSection {
        #SectionType uint16
        #SectionOffset uint64
        #SectionSize uint32
        #rawData []byte
        #crcHandler SectionCRCHandler
        +Parse(data []byte) error
        +VerifyCRC() error
    }

    class DeviceInfoSection {
        -devInfo DevInfoAnnotated
        +Parse(data []byte) error
        +GetGUIDs() []uint64
        +GetMACs() []uint64
    }

    class ImageInfoSection {
        -imageInfo FWImageInfoAnnotated
        +Parse(data []byte) error
        +GetVersion() string
        +GetPSID() string
    }

    class GenericSection {
        +Parse(data []byte) error
    }

    class SectionFactory {
        <<interface>>
        +CreateSection(...) SectionInterface
    }

    class DefaultSectionFactory {
        +CreateSection(...) SectionInterface
    }

    SectionInterface <|-- BaseSection
    BaseSection <|-- DeviceInfoSection
    BaseSection <|-- ImageInfoSection
    BaseSection <|-- GenericSection
    SectionFactory <|-- DefaultSectionFactory
    DefaultSectionFactory ..> SectionInterface : creates
```

## CRC Verification Architecture

```mermaid
graph LR
    subgraph "CRC Types"
        CRC_NONE[No CRC]
        CRC_ITOC[CRC in ITOC Entry]
        CRC_SECTION[CRC in Section]
    end

    subgraph "CRC Handlers"
        UNIFIED[UnifiedHandler]
        BOOT2[Boot2Handler]
        TOOLS[ToolsAreaHandler]
    end

    subgraph "CRC Algorithms"
        SW_CRC16[Software CRC16]
        HW_CRC16[Hardware CRC16]
        CRC32[CRC32]
    end

    CRC_ITOC --> UNIFIED
    CRC_SECTION --> UNIFIED
    CRC_SECTION --> BOOT2
    CRC_SECTION --> TOOLS

    UNIFIED --> SW_CRC16
    BOOT2 --> HW_CRC16
    TOOLS --> SW_CRC16
    TOOLS --> CRC32
```

## Data Structure Relationships

```mermaid
erDiagram
    FIRMWARE ||--o{ ITOC : contains
    FIRMWARE ||--o{ DTOC : "may contain"
    FIRMWARE ||--|| HW_POINTERS : has
    
    ITOC ||--o{ ITOC_ENTRY : contains
    DTOC ||--o{ ITOC_ENTRY : contains
    
    ITOC_ENTRY ||--|| SECTION : references
    
    SECTION ||--o| CRC : "may have"
    SECTION ||--|| SECTION_TYPE : has
    
    SECTION_TYPE {
        uint16 type_id
        string name
    }
    
    HW_POINTERS {
        uint32 boot2_addr
        uint32 itoc_addr
        uint32 dtoc_addr
    }
    
    ITOC {
        uint32 signature
        uint16 entry_count
        uint16 header_crc
    }
    
    ITOC_ENTRY {
        uint8 type
        uint32 offset
        uint32 size
        uint16 crc
    }
```

## Component Dependencies

```mermaid
graph TD
    subgraph "External Dependencies"
        COBRA[spf13/cobra]
        ZAP[uber/zap]
        BINSTRUCT[ghostiam/binstruct]
        MERRY[ansel1/merry]
        TESTIFY[stretchr/testify]
    end

    subgraph "Internal Packages"
        CMD[cmd/mlx5fw-go]
        PARSER_PKG[pkg/parser]
        INTERFACES_PKG[pkg/interfaces]
        TYPES_PKG[pkg/types]
        CRC_PKG[pkg/crc]
        ANNOTATIONS_PKG[pkg/annotations]
    end

    CMD --> COBRA
    CMD --> ZAP
    
    PARSER_PKG --> MERRY
    PARSER_PKG --> BINSTRUCT
    
    TYPES_PKG --> BINSTRUCT
    ANNOTATIONS_PKG --> |alternative to| BINSTRUCT
    
    CMD -.-> TESTIFY
    PARSER_PKG -.-> TESTIFY
    CRC_PKG -.-> TESTIFY

    style COBRA fill:#f96,stroke:#333
    style ZAP fill:#f96,stroke:#333
    style BINSTRUCT fill:#f96,stroke:#333
    style MERRY fill:#f96,stroke:#333
    style TESTIFY fill:#f96,stroke:#333
```

## Future Device Access Architecture

```mermaid
graph TB
    subgraph "Application Layer"
        BURN_CMD[burn command]
        VERIFY_CMD[verify command]
    end

    subgraph "Device Access Layer"
        DEV_MGR[Device Manager]
        REG_ACCESS[Register Accessor]
        FLASH_CTL[Flash Controller]
        ICMD[ICMD Interface]
    end

    subgraph "Platform Abstraction"
        LINUX_PCI[Linux PCI Access]
        WIN_PCI[Windows PCI Access]
        MOCK_PCI[Mock PCI Access]
    end

    subgraph "Hardware"
        PCI_DEV[PCI Device]
        FLASH_MEM[Flash Memory]
    end

    BURN_CMD --> DEV_MGR
    VERIFY_CMD --> DEV_MGR

    DEV_MGR --> REG_ACCESS
    DEV_MGR --> FLASH_CTL
    DEV_MGR --> ICMD

    REG_ACCESS --> LINUX_PCI
    REG_ACCESS --> WIN_PCI
    REG_ACCESS --> MOCK_PCI

    FLASH_CTL --> REG_ACCESS
    ICMD --> REG_ACCESS

    LINUX_PCI --> PCI_DEV
    WIN_PCI --> PCI_DEV

    PCI_DEV --> FLASH_MEM

    style PCI_DEV fill:#faa,stroke:#333,stroke-width:4px
    style FLASH_MEM fill:#faa,stroke:#333,stroke-width:4px
```