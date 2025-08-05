# HASHES_TABLE CRC Verification in mstflint

## Overview

This document explains how mstflint handles CRC verification for HASHES_TABLE sections, particularly when they originate from HW pointers and may have special CRC values.

## HASHES_TABLE Structure

The HASHES_TABLE section contains cryptographic hashes and has the following structure:
- Header (IMAGE_LAYOUT_HASHES_TABLE_HEADER_SIZE)
- HTOC (Hash Table of Contents) with entries
- Tail including a 4-byte CRC field at the very end

## CRC Processing Steps

When mstflint verifies a HASHES_TABLE section, it follows these steps:

### 1. Reading the Section

For FS4/FS5 firmware, the HASHES_TABLE pointer comes from HW pointers:
- FS4: `hw_pointers.hashes_table_pointer.ptr`
- FS5: `hw_pointers.ncore_hashes_pointer.ptr`

### 2. CRC Calculation

```cpp
// Calculate CRC over all data except the last 4 bytes (CRC field)
u_int32_t hashes_table_calc_crc = CalcImageCRC((u_int32_t*)buff, (hashes_table_size / 4) - 1);

// Read the stored CRC from the last 4 bytes
u_int32_t hashes_table_crc = ((u_int32_t*)buff)[(hashes_table_size / 4) - 1];

// Convert from big-endian to CPU endianness
TOCPU1(hashes_table_crc)

// Mask to 16 bits (CRC16)
hashes_table_crc = hashes_table_crc & 0xFFFF;
```

### 3. CRC Type

HASHES_TABLE sections use **CRC16** (16-bit CRC), stored in the last 4 bytes of the section:
- The CRC is stored as a big-endian 32-bit value
- Only the lower 16 bits are significant
- The upper 16 bits are typically 0x0000

## Special Case: 0xFFFFFFFF CRC Field

When the last 4 bytes of a HASHES_TABLE section are 0xFFFFFFFF:

1. After TOCPU1 (byte swap on little-endian systems): still 0xFFFFFFFF
2. After masking with 0xFFFF: becomes 0xFFFF
3. This is **NOT** treated as a "blank CRC" for HASHES_TABLE sections

Unlike some other sections, HASHES_TABLE does not use the blank CRC mechanism. The DumpFs3CRCCheck call for HASHES_TABLE always passes `false` for the `ignore_crc` parameter.

## CRC Verification Behavior

The CRC check passes ("OK") when:
- The calculated CRC matches the stored CRC exactly
- This includes when both are 0xFFFF (e.g., when the section data results in a CRC of 0xFFFF and the stored CRC is 0xFFFFFFFF)

The CRC check fails when:
- The calculated CRC does not match the stored CRC
- This causes the verification to fail unless CRC checking is disabled

## Example: HW Pointer Handling

When reading HW pointers in FS4/FS5, mstflint includes special handling for invalid pointers:

```cpp
// Fix pointers that are 0xFFFFFFFF
for (unsigned int k = 0; k < size; k += 2) {
    if (buff[k] == 0xFFFFFFFF) {
        buff[k] = 0;     // Fix pointer
        buff[k + 1] = 0; // Fix CRC
    }
}
```

This prevents attempting to read from invalid addresses (0xFFFFFFFF).

## Implementation Details

### CalcImageCRC Function

The CalcImageCRC function:
1. Performs byte swapping on the input data (TOCPUn)
2. Calculates CRC16 over the specified number of words
3. Returns the 16-bit CRC value

### CRC Storage Format

In firmware files:
- Bytes 0-3 at end of section: CRC stored as big-endian u32
- Example: `00 00 f2 15` represents CRC 0xf215
- Example: `ff ff ff ff` represents CRC 0xffff (after masking)

## Summary

HASHES_TABLE CRC verification in mstflint:
- Uses CRC16 stored in the last 4 bytes
- Does not support "blank CRC" mechanism
- Properly handles 0xFFFFFFFF as a valid CRC value (0xFFFF after processing)
- Always validates CRC unless globally disabled
- Special handling exists for invalid HW pointers but not for CRC validation