# HASHES_TABLE Size Calculation in mstflint

## Overview

The mstflint tool calculates the HASHES_TABLE size differently depending on the context. There are two main functions used for size determination:

1. `getHashesTableSize()` (lowercase) - Used during verification
2. `GetHashesTableSize()` (uppercase) - Used for other operations

## Size Calculation Methods

### Method 1: getHashesTableSize() - Used During Verification

This method reads the size directly from the HASHES_TABLE header and is implemented in `fs4_ops.h`:

```cpp
int getHashesTableSize(u_int32_t address)
{
    return getBoot2Size(address);
}
```

The `getBoot2Size()` function (in `fs4_ops.cpp:4538`) reads the size from the firmware:

```cpp
int Fs4Operations::getBoot2Size(u_int32_t address)
{
    u_int32_t num_of_dwords = 0;
    
    // Read the num of DWs from the second dword
    READBUF((*_ioAccess), address + 4, &num_of_dwords, 4, "num of DWs");
    
    TOCPU1(num_of_dwords)
    
    return (4 + num_of_dwords) * 4; // 2 dwords for header + 2 dwords for tail
}
```

**Formula**: `size = (4 + num_of_dwords) * 4`

Where:
- `num_of_dwords` is read from offset 4 of the HASHES_TABLE header (second DWORD)
- The 4 represents 4 DWORDs for header/tail overhead
- Result is multiplied by 4 to convert DWORDs to bytes

### Method 2: GetHashesTableSize() - Used for Calculations

This method calculates the size based on the HTOC (Hash Table of Contents) structure:

```cpp
bool Fs4Operations::GetHashesTableSize(u_int32_t& size)
{
    // ... encryption checks ...
    
    // Read HTOC header for hash size
    u_int32_t htoc_header_address = _hashes_table_ptr + IMAGE_LAYOUT_HASHES_TABLE_HEADER_SIZE;
    READALLOCBUF((*_ioAccess), htoc_header_address, buff, IMAGE_LAYOUT_HTOC_HEADER_SIZE, "HTOC header");
    image_layout_htoc_header header;
    image_layout_htoc_header_unpack(&header, buff);
    htoc_hash_size = header.hash_size;
    htoc_max_num_of_entries = header.version == 1 ? MAX_HTOC_ENTRIES_NUM_VERSION_1 : htoc_max_num_of_entries;
    
    u_int32_t htoc_size = IMAGE_LAYOUT_HTOC_HEADER_SIZE + htoc_max_num_of_entries * (IMAGE_LAYOUT_HTOC_ENTRY_SIZE + htoc_hash_size);
    size = IMAGE_LAYOUT_HASHES_TABLE_HEADER_SIZE + htoc_size + HASHES_TABLE_TAIL_SIZE;
    
    return true;
}
```

**Formula**: `size = HASHES_TABLE_HEADER_SIZE + htoc_size + HASHES_TABLE_TAIL_SIZE`

Where:
- `HASHES_TABLE_HEADER_SIZE = 12` (0xC)
- `htoc_size = HTOC_HEADER_SIZE + num_entries * (HTOC_ENTRY_SIZE + hash_size)`
- `HTOC_HEADER_SIZE = 16`
- `HTOC_ENTRY_SIZE = 8`
- `HASHES_TABLE_TAIL_SIZE = 8`

## Example: ConnectX-7 Firmware

For the firmware `fw-ConnectX7-rel-28_33_0751.bin`:

### HASHES_TABLE Header (at offset 0x7000):
```
00 00 00 00 00 00 01 fd 00 00 34 75
```

- Second DWORD (offset 4): `0x000001fd` = 509 DWORDs

### HTOC Header (at offset 0x700C):
```
00 00 00 00 00 40 00 19 00 00 00 00 00 00 00 00
```

- Version: 0
- Hash size: 0x40 (64 bytes)
- Hash type: 0
- Number of entries: 0x19 (25)

### Size Calculations:

**Method 1 (used by mstflint verify):**
- `size = (4 + 509) * 4 = 513 * 4 = 2052 bytes = 0x804`

**Method 2 (theoretical calculation):**
- `htoc_size = 16 + 25 * (8 + 64) = 16 + 25 * 72 = 1816`
- `size = 12 + 1816 + 8 = 1836 bytes = 0x72C`

## Why the Difference?

The verification output shows `0x804` because:

1. During verification, mstflint uses `getHashesTableSize()` which reads the actual size from the firmware header
2. This size includes the actual data present in the section, including any padding or additional data
3. The `GetHashesTableSize()` method calculates a theoretical size based on the HTOC structure, which may be smaller than the actual allocated size

## Implementation Recommendation

For compatibility with mstflint, implementations should:

1. **For reading/verification**: Use the size from the HASHES_TABLE header (second DWORD)
2. **For writing/creation**: Calculate based on actual content but ensure the header contains the correct size value

The constant size of 0x800 (2048 bytes) used in some implementations is close but not exact for all cases. The actual size should be read from the firmware header for accurate parsing.

## Constants Reference

```cpp
#define IMAGE_LAYOUT_HASHES_TABLE_HEADER_SIZE    (0xc)  // 12 bytes
#define IMAGE_LAYOUT_HTOC_HEADER_SIZE            (0x10) // 16 bytes  
#define IMAGE_LAYOUT_HTOC_ENTRY_SIZE             (0x8)  // 8 bytes
#define HASHES_TABLE_TAIL_SIZE                   8      // 8 bytes
```