# Porting mstflint to Go - Implementation Plan

## Overview
This document outlines the path for porting a subset of mstflint's firmware parsing functionality to Go, focusing on FS4 and FS5 formats for query operations.

## Project Goals
- Create a CLI tool for parsing Mellanox firmware files
- Support FS4 and FS5 firmware formats
- Implement query functionality similar to `mstflint -i firmware.bin query`
- Use idiomatic Go 1.24 with specified libraries

## Phase 1: Foundation (Week 1)

### 1.1 Project Setup
```
mlx5fw-go/
├── cmd/mlx5fw-go/
│   └── main.go              # CLI entry point
├── pkg/
│   ├── types/
│   │   ├── common.go        # Common types and constants
│   │   ├── fs4.go           # FS4 structures
│   │   ├── fs5.go           # FS5 structures
│   │   └── endian.go        # Endianness helpers
│   ├── interfaces/
│   │   └── parser.go        # Parser interfaces
│   └── errs/
│       └── errors.go        # Error definitions
├── go.mod
└── go.sum
```

### 1.2 Core Dependencies
```go
// go.mod
require (
    github.com/ghostiam/binstruct v1.3.2
    go.uber.org/zap v1.27.0
    github.com/ansel1/merry/v2 v2.1.1
    github.com/spf13/cobra v1.8.0  // For CLI
)
```

### 1.3 Basic Error Definitions
```go
// pkg/errs/errors.go
package errs

import "github.com/ansel1/merry/v2"

var (
    ErrInvalidMagic = merry.New("invalid magic pattern")
    ErrInvalidCRC = merry.New("CRC verification failed")
    ErrUnsupportedFormat = merry.New("unsupported firmware format")
    ErrEncryptedImage = merry.New("encrypted images not supported")
)
```

## Phase 2: Type Definitions (Week 1-2)

### 2.1 Common Types
```go
// pkg/types/common.go
package types

const (
    MagicPattern = 0x4D545F46575F5F00 // "MT_FW__"
    ITOCSignature = 0x49544F43         // "ITOC"
    DTOCSignature = 0x44544F43         // "DTOC"
    
    FS4Format = 1
    FS5Format = 2
    
    CRCPolynomial = 0x100b
    NOCRC = 1
)

// Search offsets for magic pattern
var MagicSearchOffsets = []uint32{
    0x0, 0x10000, 0x20000, 0x40000, 0x80000,
    0x100000, 0x200000, 0x400000, 0x800000,
    0x1000000, 0x2000000,
}
```

### 2.2 FS4 Structures
```go
// pkg/types/fs4.go
package types

import "github.com/ghostiam/binstruct"

type HWPointerEntry struct {
    Ptr uint32 `bin:"BE"`
    CRC uint32 `bin:"BE"`
}

type FS4HWPointers struct {
    Boot2Ptr             HWPointerEntry
    TOCPtr               HWPointerEntry
    ToolsPtr             HWPointerEntry
    ImageInfoSectionPtr  HWPointerEntry
    FWPublicKeyPtr       HWPointerEntry
    FWSignaturePtr       HWPointerEntry
    PublicKeyPtr         HWPointerEntry
    ForbiddenVersionsPtr HWPointerEntry
}

type TOCHeader struct {
    Signature0     uint32 `bin:"BE"`
    Signature1     uint32 `bin:"BE"`
    Signature2     uint32 `bin:"BE"`
    Signature3     uint32 `bin:"BE"`
    Version        uint32 `bin:"BE"`
    Reserved       uint32 `bin:"BE"`
    ITOCEntryCRC   uint32 `bin:"BE"`
    CRC            uint32 `bin:"BE"`
}

type TOCEntry struct {
    Type           uint16 `bin:"BE,offset=2"` // Bits 16-31
    Reserved0      uint16 `bin:"BE"`
    Size           uint32 `bin:"BE"`
    OffsetInImage  uint64 `bin:"BE"`
    CRC            uint32 `bin:"BE"`
    Reserved1      uint32 `bin:"BE"`
    Reserved2      uint64 `bin:"BE"`
}
```

### 2.3 FS5 Structures
```go
// pkg/types/fs5.go
package types

type FS5HWPointers struct {
    Boot2Ptr               HWPointerEntry
    TOCPtr                 HWPointerEntry
    ToolsPtr               HWPointerEntry
    ImageInfoSectionPtr    HWPointerEntry
    FWPublicKeyPtr         HWPointerEntry
    FWSignaturePtr         HWPointerEntry
    PublicKeyPtr           HWPointerEntry
    ForbiddenVersionsPtr   HWPointerEntry
    PSCBl1Ptr              HWPointerEntry
    PSCHashesTablePtr      HWPointerEntry
    NCoreHashesPointer     HWPointerEntry
    PSCFWUpdateHandlePtr   HWPointerEntry
    PSCBCHPointer          HWPointerEntry
    ReservedPtr13          HWPointerEntry
    ReservedPtr14          HWPointerEntry
    NCoreBCHPointer        HWPointerEntry
}
```

## Phase 3: Core Parsing Logic (Week 2-3)

### 3.1 Parser Interface
```go
// pkg/interfaces/parser.go
package interfaces

import "io"

type FirmwareParser interface {
    Parse(reader io.ReaderAt) error
    Query() (*FirmwareInfo, error)
    GetFormat() int
}

type FirmwareInfo struct {
    Format          string
    Version         string
    ReleaseDate     string
    PartNumber      string
    Description     string
    PSID            string
    SecurityAttrs   string
    // ... other fields from mstflint query output
}
```

### 3.2 Base Parser Implementation
```go
// pkg/parser/base.go
package parser

type BaseParser struct {
    reader      io.ReaderAt
    imageStart  uint32
    format      int
    logger      *zap.Logger
    
    // Parsed data
    hwPointers  interface{}
    itocHeader  *types.TOCHeader
    dtocHeader  *types.TOCHeader
    sections    map[uint16]*Section
}

type Section struct {
    Type   uint16
    Offset uint32
    Size   uint32
    Data   []byte
}

func (p *BaseParser) FindMagicPattern() error {
    for _, offset := range types.MagicSearchOffsets {
        var magic uint64
        if err := p.readAt(offset, &magic); err != nil {
            continue
        }
        if magic == types.MagicPattern {
            p.imageStart = offset
            return nil
        }
    }
    return errs.ErrInvalidMagic
}
```

### 3.3 CRC Implementation
```go
// pkg/crc/crc.go
package crc

// Software CRC16 implementation
type SoftwareCRC struct {
    crc uint16
}

func NewSoftwareCRC() *SoftwareCRC {
    return &SoftwareCRC{crc: 0xffff}
}

func (c *SoftwareCRC) Add(val uint32) {
    for i := 0; i < 32; i++ {
        if c.crc&0x8000 != 0 {
            c.crc = uint16(((c.crc<<1)|(val>>31))^0x100b) & 0xffff
        } else {
            c.crc = uint16((c.crc<<1)|(val>>31)) & 0xffff
        }
        val <<= 1
    }
}

func (c *SoftwareCRC) Finish() uint16 {
    // Final processing
    for i := 0; i < 16; i++ {
        if c.crc&0x8000 != 0 {
            c.crc = ((c.crc << 1) ^ 0x100b) & 0xffff
        } else {
            c.crc = (c.crc << 1) & 0xffff
        }
    }
    return c.crc ^ 0xffff
}

// Hardware CRC implementation
func CalcHardwareCRC(data []byte) uint16 {
    // Implementation with table lookup and first 2 bytes inverted
}
```

## Phase 4: FS4 Parser (Week 3-4)

### 4.1 FS4 Parser Implementation
```go
// pkg/parser/fs4.go
package parser

type FS4Parser struct {
    BaseParser
    hwPointers *types.FS4HWPointers
}

func (p *FS4Parser) ParseHWPointers() error {
    offset := p.imageStart + 0x18
    
    // Read HW pointers
    data := make([]byte, 64) // Initial read
    if err := p.readAt(offset, data); err != nil {
        return err
    }
    
    // Check for extended pointers
    // Verify CRC for each pointer
    // Use dual CRC approach (software first, then hardware)
    
    return binstruct.UnmarshalBE(data, &p.hwPointers)
}

func (p *FS4Parser) ParseITOC() error {
    if p.hwPointers.TOCPtr.Ptr == 0 {
        return errs.New("invalid ITOC pointer")
    }
    
    // Check for encryption
    header := &types.TOCHeader{}
    if err := p.readStruct(p.hwPointers.TOCPtr.Ptr, header); err != nil {
        return err
    }
    
    if header.Signature0 != types.ITOCSignature {
        return errs.ErrEncryptedImage
    }
    
    // Verify header CRC
    // Parse entries
    // Build section map
    
    return nil
}
```

## Phase 5: FS5 Parser (Week 4)

### 5.1 FS5 Parser Implementation
```go
// pkg/parser/fs5.go
package parser

type FS5Parser struct {
    BaseParser
    hwPointers *types.FS5HWPointers
}

func (p *FS5Parser) ParseHWPointers() error {
    offset := p.imageStart + 0x18
    
    // FS5 has fixed 128-byte HW pointer structure
    data := make([]byte, 128)
    if err := p.readAt(offset, data); err != nil {
        return err
    }
    
    // Fix 0xFFFFFFFF pointers
    // Verify CRC
    
    return binstruct.UnmarshalBE(data, &p.hwPointers)
}

func (p *FS5Parser) ParseHashesTable() error {
    if p.hwPointers.NCoreHashesPointer.Ptr == 0 {
        return nil // Not present
    }
    
    // Read hashes table header
    // Verify header CRC
    // Read table data
    // Verify table CRC
    
    return nil
}
```

## Phase 6: Query Implementation (Week 5)

### 6.1 Image Info Parser
```go
// pkg/parser/imageinfo.go
package parser

type ImageInfo struct {
    FWVersion     [16]byte  `bin:"offset=0x10"`
    FWReleaseDate [16]byte  `bin:"offset=0x20"`
    MICVersion    [16]byte  `bin:"offset=0x40"`
    PRSName       [128]byte `bin:"offset=0x50"`
    PartNumber    [32]byte  `bin:"offset=0xd0"`
    Description   [512]byte `bin:"offset=0x170"`
    PSID          [16]byte  `bin:"offset=0x390"`
    // ... other fields
}

func ParseImageInfo(data []byte) (*ImageInfo, error) {
    info := &ImageInfo{}
    if err := binstruct.UnmarshalBE(data, info); err != nil {
        return nil, err
    }
    return info, nil
}
```

### 6.2 Query Output Formatter
```go
// pkg/output/formatter.go
package output

func FormatQuery(info *interfaces.FirmwareInfo) string {
    var buf strings.Builder
    
    buf.WriteString(fmt.Sprintf("Image type:            %s\n", info.Format))
    buf.WriteString(fmt.Sprintf("FW Version:            %s\n", info.Version))
    buf.WriteString(fmt.Sprintf("FW Release Date:       %s\n", info.ReleaseDate))
    // ... format other fields matching mstflint output
    
    return buf.String()
}
```

## Phase 7: CLI Implementation (Week 5-6)

### 7.1 Main CLI
```go
// cmd/mlx5fw-go/main.go
package main

import (
    "github.com/spf13/cobra"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    
    rootCmd := &cobra.Command{
        Use:   "mlx5fw-go",
        Short: "Mellanox firmware parser",
    }
    
    queryCmd := &cobra.Command{
        Use:   "query",
        Short: "Query firmware information",
        RunE: func(cmd *cobra.Command, args []string) error {
            filename, _ := cmd.Flags().GetString("image")
            return runQuery(filename, logger)
        },
    }
    
    queryCmd.Flags().StringP("image", "i", "", "Firmware image file")
    queryCmd.MarkFlagRequired("image")
    
    rootCmd.AddCommand(queryCmd)
    rootCmd.Execute()
}
```

## Phase 8: Testing Strategy (Week 6)

### 8.1 Unit Tests
- CRC calculation tests (compare with known values)
- Structure parsing tests with mock data
- Endianness conversion tests

### 8.2 Integration Tests
```bash
# Compare output with mstflint
./mlx5fw-go query -i firmware.bin > our_output.txt
mstflint -i firmware.bin query > mstflint_output.txt
diff our_output.txt mstflint_output.txt
```

### 8.3 Test Coverage Goals
- All sample firmwares should parse correctly
- Output should match mstflint for supported features
- Graceful handling of encrypted images
- Proper error messages for unsupported features

## Implementation Priority

### Must Have (MVP)
1. FS4/FS5 format detection
2. Magic pattern search
3. HW pointer parsing
4. ITOC/DTOC header parsing
5. IMAGE_INFO extraction
6. Basic query output
7. CRC verification for critical sections

### Should Have
1. Complete section enumeration
2. Vendor firmware support
3. Detailed error messages
4. Debug logging mode
5. All CRC verifications

### Nice to Have
1. JSON output format
2. Section dumping
3. Performance optimizations
4. FS3 support (future)

## Key Challenges and Solutions

### 1. Bit Field Handling
- Use binstruct field tags for offset handling
- Manual bit manipulation for complex fields

### 2. Endianness
- All firmware data is Big Endian
- Use binstruct's BE tags
- Helper functions for manual conversions

### 3. Dynamic Structure Sizes
- Read incrementally
- Check for terminators (0xFFFFFFFF)
- Validate with CRC

### 4. CRC Verification
- Implement both software and hardware CRC
- Try software first for HW pointers
- Cache results where possible

## Success Metrics
1. **Sample Tests**: Achieve 100% pass rate on `scripts/sample_tests/` for each implemented feature
   - Each ported feature must pass all relevant sample tests before considered complete
   - Sample tests provide verbose output for debugging and validation
2. **Integration Tests**: Achieve score > 0.9 on `scripts/integration_tests/`
   - Tests run on both sample firmwares and external firmware set
   - Must maintain this score as new features are added
3. Parse all non-encrypted sample firmwares correctly
4. Query output matches mstflint for all supported fields
5. Clean error handling for unsupported features (encrypted images, etc.)
6. Idiomatic Go code with proper error wrapping using merry

## Timeline Summary
- Week 1: Foundation and type definitions
- Week 2-3: Core parsing logic and CRC
- Week 3-4: FS4 parser implementation
- Week 4: FS5 parser implementation  
- Week 5: Query implementation and CLI
- Week 6: Testing and refinement

This plan provides a structured approach to porting mstflint's query functionality to Go, focusing on the most essential features first while maintaining compatibility with the original tool's output format.