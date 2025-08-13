# mstflint CRC Calculation Investigation

## Overview

This document investigates how mstflint calculates CRC16 for firmware sections, particularly focusing on sections where the CRC is stored in the ITOC entry (CRC type `INITOCENTRY`).

## Key Findings

### 1. CRC Algorithm
mstflint uses a CRC16 algorithm with polynomial 0x100b. The implementation is in `mft_utils/crc16.cpp`.

### 2. Byte Order Handling
The critical difference in mstflint's CRC calculation is its handling of byte order:

1. **Data is converted from big-endian to host-endian** before CRC calculation (`TOCPUn` macro)
2. **CRC is calculated on host-endian data** 
3. **Data is converted back to big-endian** after CRC calculation (`CPUTOn` macro)

This is implemented in `FwOperations::CalcImageCRC`:
```cpp
u_int32_t FwOperations::CalcImageCRC(u_int32_t* buff, u_int32_t size)
{
    Crc16 crc;
    TOCPUn(buff, size);      // Convert BE to host
    CRCn(crc, buff, size);   // Calculate CRC
    CPUTOn(buff, size);      // Convert back to BE
    crc.finish();
    return crc.get();
}
```

### 3. Size Parameter
- The `size` parameter to `CalcImageCRC` is in **dwords (4-byte units)**, not bytes
- This matches the size field in ITOC entries which also stores size in dwords

### 4. Zero-Length Sections
For zero-length sections (like VPD_R0), mstflint:
- Passes NULL pointer and size 0 to `CalcImageCRC`
- The CRC16 algorithm with no input data produces a fixed value: **2389 (0x955)**
- This value is stored in the ITOC entry's section_crc field

### 5. Test Results

#### VPD_R0 (Zero-length section)
- Size: 0 bytes (0 dwords)
- Expected CRC: 2389
- Calculated CRC: 2389 ✓

#### RESET_INFO
- Size: 256 bytes (64 dwords)
- Expected CRC: 15622
- Investigation ongoing - byte order conversion confirmed but CRC still doesn't match

#### DIGITAL_CERT_PTR
- Size: 40 bytes (10 dwords)
- Expected CRC: 54764
- Calculated CRC with byte swap: 54764 ✓
- Note: This section contains pointer data, not zeros

## Critical Finding: Byte Order Processing

The key to understanding mstflint's CRC calculation is that **CRC is calculated on host-endian data, not on the raw big-endian data from the firmware file**.

### Correct Implementation Steps

For sections with CRC in ITOC entry (CRC type `INITOCENTRY`):

1. Read section data from firmware (data is stored in big-endian format)
2. **Convert each dword from big-endian to host-endian** (using `ntohl` on little-endian systems)
3. Calculate CRC16 on the host-endian data
4. The result matches the CRC stored in the ITOC entry

### Verified Test Results

#### VPD_R0 (Zero-length section)
- Size: 0 bytes (0 dwords)
- Expected CRC: 2389
- Calculated CRC: 2389 ✓

#### RESET_INFO
- Size: 256 bytes (64 dwords)
- Expected CRC: 15622
- Calculated CRC without byte swap: 61385 ✗
- Calculated CRC with byte swap: 15622 ✓

#### DBG_FW_PARAMS
- Size: 8 bytes (2 dwords)
- Expected CRC: 32261
- Calculated CRC without byte swap: 10511 ✗
- Calculated CRC with byte swap: 32261 ✓

## Implementation Notes

The mstflint implementation in `FwOperations::CalcImageCRC`:
1. Converts data from big-endian to host-endian (`TOCPUn`)
2. Calculates CRC on host-endian data
3. Converts data back to big-endian (`CPUTOn`) - this preserves the original buffer

This approach ensures consistent CRC calculation across different architectures.

## Complete CRC Algorithm Summary

### For Go Implementation

To match mstflint's CRC calculation for sections with CRC in ITOC entry:

```go
func CalculateImageCRC(data []byte) uint16 {
    // Ensure data is aligned to dwords
    if len(data) % 4 != 0 {
        panic("data must be dword-aligned")
    }
    
    crc := NewCRC16()
    
    // Process data as dwords
    for i := 0; i < len(data); i += 4 {
        // Read dword in big-endian
        dword := binary.BigEndian.Uint32(data[i:i+4])
        
        // Convert to host-endian (assuming little-endian host)
        // This is equivalent to mstflint's TOCPUn
        hostDword := dword // On big-endian host, this would need byte swapping
        
        // Add to CRC
        crc.Add(hostDword)
    }
    
    crc.Finish()
    return crc.Get()
}
```

### Key Points

1. **Input format**: Section data as raw bytes from firmware (big-endian)
2. **Processing unit**: Dwords (4-byte units)
3. **Byte order**: Convert each dword from big-endian to host-endian before CRC
4. **Size parameter**: When stored in ITOC, size is in dwords, not bytes
5. **Zero-length sections**: Return fixed CRC value 2389

## Code References

- CRC implementation: `/reference/mstflint/mft_utils/crc16.cpp`
- CRC calculation: `/reference/mstflint/mlxfwops/lib/fw_ops.cpp:1230`
- ITOC verification: `/reference/mstflint/mlxfwops/lib/fs4_ops.cpp:772`
- VPD_R0 handling: `/reference/mstflint/mlxfwops/lib/fs4_ops.cpp:2115`