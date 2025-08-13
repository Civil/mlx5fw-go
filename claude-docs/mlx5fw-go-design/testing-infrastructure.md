# Testing Infrastructure

## Overview

The mlx5fw-go project has a comprehensive testing infrastructure that includes unit tests, integration tests, and shell-based test scripts for validating functionality against real firmware files.

## Test Structure

```
mlx5fw-go/
├── pkg/                     # Package-level unit tests
│   ├── annotations/
│   │   └── annotations_test.go
│   ├── crc/
│   │   ├── base_handler_test.go
│   │   ├── calculator_test.go
│   │   ├── handlers_test.go
│   │   └── unified_handler_test.go
│   ├── parser/
│   │   ├── crc_test.go
│   │   ├── firmware_reader_test.go
│   │   └── fs4/
│   │       ├── itoc_crc_test.go
│   │       └── parser_test.go
│   ├── section/
│   │   └── replacer_test.go
│   └── types/
│       ├── binary_version_test.go
│       ├── endian_test.go
│       ├── fs4_test.go
│       ├── hw_pointers_test.go
│       ├── itoc_entry_test.go
│       ├── section_names_test.go
│       └── types_test.go
├── cmd/mlx5fw-go/
│   └── replace_section_v4_test.go
└── scripts/
    ├── integration_tests/   # Full integration tests
    └── sample_tests/        # Sample firmware tests
```

## Testing Approaches

### 1. Unit Tests

Unit tests focus on individual components and functions:

#### Example: CRC Calculator Test
```go
// pkg/crc/calculator_test.go
func TestCalculateSoftwareCRC(t *testing.T) {
    calc := NewCRCCalculator()
    
    testCases := []struct {
        name     string
        data     []byte
        expected uint16
    }{
        {
            name:     "empty data",
            data:     []byte{},
            expected: 0xFFFF,
        },
        {
            name:     "single byte",
            data:     []byte{0x01},
            expected: 0x1E0E,
        },
        // More test cases...
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := calc.CalculateSoftwareCRC(tc.data)
            assert.Equal(t, tc.expected, result)
        })
    }
}
```

#### Example: Annotation Parsing Test
```go
// pkg/annotations/annotations_test.go
func TestUnmarshal(t *testing.T) {
    type TestStruct struct {
        Magic    uint32 `offset:"byte:0,endian:be"`
        Version  uint8  `offset:"byte:4"`
        Flags    uint8  `offset:"bit:40,len:4"`
    }
    
    data := []byte{0x12, 0x34, 0x56, 0x78, 0x01, 0xF0}
    
    var s TestStruct
    err := Unmarshal(data, &s)
    
    assert.NoError(t, err)
    assert.Equal(t, uint32(0x12345678), s.Magic)
    assert.Equal(t, uint8(0x01), s.Version)
    assert.Equal(t, uint8(0x0F), s.Flags)
}
```

### 2. Integration Tests

Shell-based integration tests that test the complete CLI:

#### Test Structure
```bash
# scripts/integration_tests/lib.sh
#!/bin/bash

# Common test functions
run_test() {
    local test_name="$1"
    local cmd="$2"
    local expected="$3"
    
    echo -n "Testing $test_name... "
    result=$(eval "$cmd" 2>&1)
    
    if [[ "$result" == *"$expected"* ]]; then
        echo "PASSED"
        return 0
    else
        echo "FAILED"
        echo "Expected: $expected"
        echo "Got: $result"
        return 1
    fi
}
```

#### Example Test Script
```bash
# scripts/integration_tests/query.sh
#!/bin/bash

source "$(dirname "$0")/lib.sh"

# Test basic query functionality
for fw in sample_firmwares/*.bin; do
    run_test "query $fw" \
        "./mlx5fw-go query -f $fw" \
        "FW Version"
done

# Test JSON output
run_test "query with JSON" \
    "./mlx5fw-go query -f sample.bin --json" \
    '"success":true'
```

### 3. Sample Firmware Tests

Tests using real firmware samples:

```bash
# scripts/sample_tests/sections.sh
#!/bin/bash

# Test section extraction
test_sections() {
    local fw="$1"
    
    # List sections
    ./mlx5fw-go sections -f "$fw" > sections.txt
    
    # Verify expected sections exist
    grep -q "BOOT2" sections.txt || error "BOOT2 section missing"
    grep -q "DEV_INFO" sections.txt || error "DEV_INFO section missing"
    grep -q "IMAGE_INFO" sections.txt || error "IMAGE_INFO section missing"
}

# Run tests on all sample firmwares
for fw in sample_firmwares/*.bin; do
    echo "Testing sections for $fw"
    test_sections "$fw"
done
```

## Test Data Management

### 1. Mock Data Generation

Helper functions for creating test data:

```go
// pkg/types/test_helpers.go
func GenerateMockITOCEntry(sectionType uint8, offset uint32, size uint32) *ITOCEntry {
    return &ITOCEntry{
        Type:         sectionType,
        FlashOffset:  offset,
        SectionLenDW: size / 4,
        SectionCRC:   0x1234,
    }
}

func GenerateMockFirmwareHeader() []byte {
    header := make([]byte, 256)
    // Magic pattern
    binary.BigEndian.PutUint32(header[0:4], 0x4D544657)
    // Version
    header[4] = 0x01
    return header
}
```

### 2. Test Fixtures

Sample firmware files for testing:

```
sample_firmwares/
├── MCX516A-CDA_Ax_Bx_MT_0000000013_rel-16_35.2000.bin
├── fw-ConnectX6Dx-rel-22_41_1000.bin
├── fw-ConnectX7-rel-28_33_0751.bin
├── fw-ConnectX8-encrypted.bin
└── broken_fw.bin  # Intentionally corrupted for error testing
```

## Testing Best Practices

### 1. Table-Driven Tests

Use table-driven tests for comprehensive coverage:

```go
func TestSectionTypeName(t *testing.T) {
    tests := []struct {
        input    uint16
        expected string
    }{
        {0x1, "FW_INFO"},
        {0x5, "MFG_INFO"},
        {0x7, "BOOT2"},
        {0x8, "DEV_INFO"},
        {0xFFFF, "UNKNOWN(0xFFFF)"},
    }
    
    for _, tt := range tests {
        t.Run(fmt.Sprintf("type_%x", tt.input), func(t *testing.T) {
            result := GetSectionTypeName(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### 2. Error Testing

Always test error conditions:

```go
func TestParseInvalidFirmware(t *testing.T) {
    testCases := []struct {
        name        string
        data        []byte
        expectedErr string
    }{
        {
            name:        "empty file",
            data:        []byte{},
            expectedErr: "file too small",
        },
        {
            name:        "no magic pattern",
            data:        make([]byte, 1024),
            expectedErr: "magic pattern not found",
        },
        {
            name:        "corrupted header",
            data:        generateCorruptedHeader(),
            expectedErr: "invalid header CRC",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            parser := NewParser()
            err := parser.Parse(bytes.NewReader(tc.data))
            
            assert.Error(t, err)
            assert.Contains(t, err.Error(), tc.expectedErr)
        })
    }
}
```

### 3. Mock Interfaces

Use interfaces for testability:

```go
// Mock reader for testing
type mockReader struct {
    data   []byte
    errors map[int64]error
}

func (m *mockReader) ReadAt(p []byte, off int64) (n int, err error) {
    if err, exists := m.errors[off]; exists {
        return 0, err
    }
    
    n = copy(p, m.data[off:])
    if n < len(p) {
        err = io.EOF
    }
    return
}
```

## Running Tests

### 1. Unit Tests

```bash
# Run all unit tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/crc/...

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestCalculateSoftwareCRC ./pkg/crc
```

### 2. Integration Tests

```bash
# Run all integration tests
cd scripts/integration_tests
./run.sh

# Run specific test suite
./query.sh
./sections.sh
```

### 3. Benchmarks

Performance testing for critical paths:

```go
func BenchmarkCRCCalculation(b *testing.B) {
    calc := NewCRCCalculator()
    data := make([]byte, 4096)
    rand.Read(data)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        calc.CalculateSoftwareCRC(data)
    }
}
```

Run benchmarks:
```bash
go test -bench=. ./pkg/crc
```

## Continuous Integration

### GitHub Actions Workflow

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.22'
    
    - name: Run unit tests
      run: go test -v -cover ./...
    
    - name: Run integration tests
      run: |
        cd scripts/integration_tests
        ./run.sh
```

## Test Coverage Goals

1. **Unit Test Coverage**: Aim for >80% code coverage
2. **Integration Coverage**: Test all CLI commands
3. **Error Coverage**: Test all error paths
4. **Edge Cases**: Test boundary conditions
5. **Performance**: Benchmark critical operations

## Future Testing Improvements

1. **Fuzz Testing**: Add fuzzing for parser robustness
2. **Property-Based Testing**: Use gopter for invariant testing
3. **Mock Device Testing**: Prepare mocks for device access layer
4. **Performance Regression**: Track performance over time
5. **Security Testing**: Add tests for malicious firmware handling