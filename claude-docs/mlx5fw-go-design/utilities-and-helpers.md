# Utilities and Helper Functions

## Overview

This document describes the utility functions and helper modules that support the core functionality of mlx5fw-go.

## Package Structure

```
pkg/
├── utils/          # General utilities
├── errors/         # Error handling
├── errs/          # Additional error types
├── annotations/    # Binary parsing utilities
└── crc/           # CRC calculation utilities
```

## Key Utility Modules

### 1. Binary Parsing Utilities (pkg/annotations)

The annotations package provides a powerful framework for parsing binary structures:

#### Marshal/Unmarshal Functions

```go
// Unmarshal binary data into annotated struct
func Unmarshal(data []byte, v interface{}) error

// Marshal annotated struct to binary data  
func Marshal(v interface{}) ([]byte, error)
```

#### Key Features

1. **Offset Annotations**: Specify exact byte/bit positions
2. **Endianness Control**: Per-field endianness specification
3. **Bitfield Support**: Extract specific bits from bytes
4. **Dynamic Lists**: Runtime-sized arrays based on count fields

Example usage:
```go
type Header struct {
    Magic    uint32 `offset:"byte:0,endian:be"`
    Version  uint8  `offset:"byte:4"`
    Count    uint16 `offset:"byte:6,endian:be"`
    Entries  []Entry `offset:"byte:8,list_size:Count"`
}
```

### 2. CRC Utilities (pkg/crc)

Various CRC calculation implementations:

#### CRC Calculator
```go
// pkg/crc/calculator.go
type CRCCalculator struct {
    // Contains CRC tables for different algorithms
}

func (c *CRCCalculator) CalculateSoftwareCRC(data []byte) uint16
func (c *CRCCalculator) CalculateHardwareCRC(data []byte) uint16
func (c *CRCCalculator) CalculateCRC32(data []byte) uint32
```

#### Specialized Handlers
- **UnifiedHandler**: General-purpose CRC handling
- **Boot2Handler**: BOOT2 section-specific CRC
- **ToolsAreaHandler**: TOOLS_AREA section CRC

### 3. Section Utilities (pkg/utils)

Helper functions for section manipulation:

```go
// pkg/utils/section_utils.go

// GetSectionTypeName returns human-readable section name
func GetSectionTypeName(sectionType uint16) string

// ValidateSectionBounds checks if section fits in firmware
func ValidateSectionBounds(offset uint64, size uint32, fileSize int64) error

// AlignToDBoundary aligns offset to 4-byte boundary
func AlignToDWBoundary(offset uint32) uint32
```

### 4. Error Handling (pkg/errors, pkg/errs)

Custom error types for better error context:

```go
// pkg/errors/errors.go
type ParseError struct {
    Section string
    Offset  uint64
    Err     error
}

type ValidationError struct {
    Field    string
    Expected interface{}
    Actual   interface{}
}

type CRCError struct {
    Section     string
    Expected    uint32
    Calculated  uint32
}
```

### 5. Type Conversion Utilities

Located throughout the types package:

```go
// pkg/types/common.go

// ConvertDWOffsetToBytes converts DWORD offset to byte offset
func ConvertDWOffsetToBytes(dwOffset uint32) uint64

// ConvertBytesToDW converts byte count to DWORD count
func ConvertBytesToDW(bytes uint32) uint32

// FormatGUID formats a GUID for display
func FormatGUID(guid uint64) string

// FormatMAC formats a MAC address for display  
func FormatMAC(mac uint64) string
```

### 6. Hex/Binary Formatting

```go
// pkg/types/fw_byte_slice.go
type FWByteSlice []byte

// String returns hex representation
func (f FWByteSlice) String() string

// MarshalJSON formats as hex string for JSON
func (f FWByteSlice) MarshalJSON() ([]byte, error)
```

### 7. Endianness Utilities

```go
// pkg/types/endian.go

// ReadBE16 reads big-endian uint16
func ReadBE16(data []byte, offset int) uint16

// WriteBE16 writes big-endian uint16
func WriteBE16(data []byte, offset int, value uint16)

// Similar for BE32, LE16, LE32...
```

## Common Patterns

### 1. Option Pattern for Configuration

```go
// pkg/interfaces/section_options.go
type SectionOption func(*BaseSection)

func WithCRC(crcType types.CRCType, crc uint32) SectionOption {
    return func(s *BaseSection) {
        s.SectionCRCType = crcType
        s.SectionCRC = crc
    }
}

// Usage
section := NewBaseSectionWithOptions(
    sectionType, offset, size,
    WithCRC(types.CRCInSection, 0x1234),
    WithEncryption(),
)
```

### 2. Validation Helpers

```go
// Common validation pattern
func ValidateFirmwarePath(path string) error {
    if path == "" {
        return fmt.Errorf("firmware path cannot be empty")
    }
    
    info, err := os.Stat(path)
    if err != nil {
        return fmt.Errorf("cannot access firmware file: %w", err)
    }
    
    if info.IsDir() {
        return fmt.Errorf("path is a directory, not a file")
    }
    
    return nil
}
```

### 3. Logging Utilities

The project uses zap for structured logging:

```go
// cmd/mlx5fw-go/common.go
func ConfigureLogger(jsonOutput bool, verbose bool) (*zap.Logger, error) {
    var cfg zap.Config
    
    if jsonOutput {
        cfg = zap.NewProductionConfig()
    } else {
        cfg = zap.NewDevelopmentConfig()
    }
    
    if !verbose {
        cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
    }
    
    return cfg.Build()
}
```

### 4. JSON Output Formatting

```go
// cmd/mlx5fw-go/json_output.go
type JSONOutput struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

func OutputJSON(data interface{}, err error) {
    output := JSONOutput{
        Success: err == nil,
        Data:    data,
    }
    
    if err != nil {
        output.Error = err.Error()
    }
    
    json.NewEncoder(os.Stdout).Encode(output)
}
```

## Best Practices

### 1. Error Wrapping

Always wrap errors with context:
```go
if err := operation(); err != nil {
    return merry.Wrap(err).WithMessage("failed to perform operation")
}
```

### 2. Resource Cleanup

Use defer for cleanup:
```go
file, err := os.Open(path)
if err != nil {
    return err
}
defer file.Close()
```

### 3. Validation First

Validate inputs before processing:
```go
func ProcessSection(data []byte, offset uint32) error {
    if len(data) < MinSectionSize {
        return fmt.Errorf("section too small: %d bytes", len(data))
    }
    
    if offset > MaxOffset {
        return fmt.Errorf("offset out of range: 0x%x", offset)
    }
    
    // Process section...
}
```

### 4. Constants Over Magic Numbers

Define constants for magic values:
```go
const (
    MagicPattern      = 0x4D544657 // "MTFW"
    MinSectionSize    = 16
    CRCSize          = 4
    DWordSize        = 4
)
```

## Testing Utilities

### Test Helpers

```go
// Common test pattern
func TestSectionParsing(t *testing.T) {
    testCases := []struct {
        name     string
        data     []byte
        expected SectionInfo
        wantErr  bool
    }{
        // Test cases...
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := ParseSection(tc.data)
            
            if tc.wantErr {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tc.expected, result)
        })
    }
}
```

### Mock Data Generation

```go
func GenerateMockFirmware() []byte {
    // Generate test firmware data
}

func GenerateMockSection(sectionType uint16, size int) []byte {
    // Generate test section data
}
```