# ITOC/DTOC CRC Verification Algorithm

This document describes how mstflint verifies checksums for ITOC (Image Table of Contents) and DTOC (Device Table of Contents) structures in Mellanox firmware images.

## Overview

ITOC/DTOC structures use two types of CRCs:
1. **Header CRC** - Protects the 32-byte TOC header
2. **Entry CRC** - Protects each 32-byte TOC entry

## CRC Algorithms

mstflint uses two different CRC algorithms:

### 1. Software CRC16 (CalculateImageCRC)
- **Polynomial**: 0x100b
- **Initial value**: 0xFFFF
- **Final XOR**: 0xFFFF
- **Used for**: TOC headers, section data
- **Implementation**: Bit-by-bit processing

### 2. Hardware CRC16 (CalculateHardwareCRC)
- **Polynomial**: 0x100b
- **Initial value**: 0xFFFF
- **Final XOR**: 0xFFFF
- **Used for**: TOC entries
- **Implementation**: Optimized for hardware

## TOC Header CRC

### Location
- **Offset**: 28 bytes (0x1C)
- **Size**: 32-bit field, but only lower 16 bits contain CRC
- **Coverage**: First 28 bytes of header (excluding CRC field itself)

### Calculation
```
1. Read first 28 bytes of TOC header
2. Calculate Software CRC16 over these bytes (7 dwords)
3. Store result in lower 16 bits of the 32-bit field at offset 28
4. Upper 16 bits of the field are preserved unchanged
```

### Verification
```
1. Extract stored CRC from bits [15:0] of the 32-bit field at offset 28
2. Calculate Software CRC16 over first 28 bytes
3. Compare calculated vs stored CRC
4. Special case: CRC value 0xFFFF indicates uninitialized/blank CRC
```

## TOC Entry CRC

### Location
- **Offset**: 30 bytes (0x1E) within each 32-byte entry
- **Size**: 16-bit field
- **Coverage**: First 28 bytes of entry (excluding last 4 bytes which contain CRC and padding)

### Calculation
```
1. Read first 28 bytes of TOC entry (7 dwords)
2. Calculate Software CRC16 over these bytes
3. Store result as 16-bit big-endian value at offset 30
```

### Verification
```
1. Extract stored CRC from bytes 30-31 of entry
2. Calculate Software CRC16 over first 28 bytes (7 dwords)
3. Compare calculated vs stored CRC
```

## Important Notes

1. **Byte Order**: All multi-byte values use big-endian format
2. **Data Preparation**: Before CRC calculation, data is converted to big-endian using TOCPUn() macro
3. **Section CRC**: Some sections store their CRC in the ITOC entry's SectionCRC field (bits 208-223)
4. **CRC Types**:
   - `CRCInITOCEntry`: CRC stored in ITOC entry's SectionCRC field
   - `CRCInSection`: CRC stored at the end of section data
   - `CRCNone`: No CRC for the section

## Implementation Differences Found

Our current implementation incorrectly uses Hardware CRC for ITOC entries. The correct approach is:
- **TOC Headers**: Software CRC16 (CalculateImageCRC) ✓ Correct
- **TOC Entries**: Software CRC16 (CalculateImageCRC) ✗ Currently using Hardware CRC

This mismatch causes mstflint verification to fail with "Bad IToc Entry CRC" errors.

## References

- mstflint source: `mlxfwops/lib/fs3_ops.cpp` - CalcItocEntryCRC()
- mstflint source: `mft_utils/crc16.cpp` - Crc16 implementation
- Error seen: "Bad IToc Entry CRC. Expected: 0xf24c, Actual: 0x498d"