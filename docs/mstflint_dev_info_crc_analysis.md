# mstflint DEV_INFO CRC Calculation Analysis

## Overview

This document details how mstflint calculates CRC for DEV_INFO sections in Mellanox firmware images, based on source code analysis and debugging.

## Key Findings

### 1. DEV_INFO Section Structure
- Total size: 512 bytes (0x200)
- CRC calculation range: First 508 bytes (127 32-bit words)
- CRC storage location: Bytes 508-511 (last 4 bytes)
- CRC format: 32-bit field with 16-bit CRC in lower 16 bits (big-endian)

### 2. CRC Algorithm Details

From `fs4_ops.cpp`:
```cpp
// Calculate CRC on first 127 words (508 bytes)
u_int32_t newSectionCRC = CalcImageCRC((u_int32_t*)DevInfoBuffer, IMAGE_LAYOUT_DEVICE_INFO_SIZE / 4 - 1);
u_int32_t newCRC = TOCPU1(newSectionCRC);
((u_int32_t*)DevInfoBuffer)[IMAGE_LAYOUT_DEVICE_INFO_SIZE / 4 - 1] = newCRC;
```

The CRC calculation process:
1. Uses CRC16 with polynomial 0x100b
2. Processes data as 32-bit words in big-endian format
3. Initial value: 0xFFFF
4. Final XOR: 0xFFFF

### 3. Data Processing

From `fw_ops.cpp`:
```cpp
u_int32_t FwOperations::CalcImageCRC(u_int32_t* buff, u_int32_t size)
{
    Crc16 crc;
    TOCPUn(buff, size);     // Convert to CPU endianness
    CRCn(crc, buff, size);  // Calculate CRC
    CPUTOn(buff, size);     // Convert back to big-endian
    
    crc.finish();
    return crc.get();
}
```

Key steps:
1. Data is converted from big-endian to host endianness
2. CRC is calculated on host-endian data
3. Data is converted back to big-endian
4. CRC result is stored as big-endian

### 4. CRC Storage Format

The CRC is stored differently than initially expected:
- **Location**: Offset 508-511 (4 bytes)
- **Format**: 32-bit big-endian value
- **Actual CRC**: Lower 16 bits of the 32-bit value
- **Upper 16 bits**: Always 0x0000 in observed firmware

Example:
```
Last 8 bytes: 00 00 00 00 00 00 05 a9
              ^-----------^  ^-------^
              Reserved       CRC32 field
                            Upper | Lower
                            0000  | 05a9
```

### 5. mlx5fw-go Implementation Issue

The original mlx5fw-go implementation reads CRC from bytes 510-511:
```go
expectedCRC := uint32(data[510])<<8 | uint32(data[511])
```

This works when upper 16 bits are zero but is technically incorrect. The correct implementation should read the full 32-bit field:
```go
crc32 := binary.BigEndian.Uint32(data[508:512])
expectedCRC := uint16(crc32 & 0xFFFF)
```

### 6. CRC Mismatch Analysis

When mlx5fw-go reports "FAIL (0x8AEB != 0x05A9)", it likely means:
- Calculated CRC: 0x8AEB 
- Expected CRC: 0x05A9

Possible causes:
1. Incorrect data range for CRC calculation
2. Endianness issues in data processing
3. Incorrect CRC algorithm implementation
4. Modified or corrupted DEV_INFO data

### 7. Verification Test Code

```go
// Correct CRC calculation matching mstflint
func calculateDevInfoCRC(devInfo []byte) uint16 {
    crc := NewCRC16() // Initialize with 0xFFFF
    
    // Process 127 32-bit words (508 bytes)
    for i := 0; i < 127; i++ {
        word := binary.BigEndian.Uint32(devInfo[i*4 : i*4+4])
        crc.Add(word)
    }
    
    return crc.Finish() // Includes final XOR with 0xFFFF
}
```

## Recommendations

1. Update mlx5fw-go to read CRC from the correct location (508-511 as 32-bit)
2. Ensure CRC calculation uses exactly 508 bytes (not 512)
3. Verify endianness handling matches mstflint
4. Add debug logging to show:
   - Data range used for CRC
   - Calculated vs expected CRC values
   - Raw bytes at CRC location

## References

- `mstflint/mlxfwops/lib/fs4_ops.cpp`: DEV_INFO section creation
- `mstflint/mlxfwops/lib/fw_ops.cpp`: CalcImageCRC implementation
- `mstflint/mft_utils/crc16.cpp`: CRC16 algorithm implementation
- `IMAGE_LAYOUT_DEVICE_INFO_SIZE`: 0x200 (512 bytes)