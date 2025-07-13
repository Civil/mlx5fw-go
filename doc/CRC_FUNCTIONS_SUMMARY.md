# CRC Calculation Functions and Algorithms in mstflint

## CRC Algorithms

### 1. CRC16 (Software Implementation)
- **Location**: `/mft_utils/crc16.h` and `/mft_utils/crc16.cpp`
- **Class**: `Crc16`
- **Polynomial**: 0x100b
- **Initial Value**: 0xffff
- **Final XOR**: 0xffff
- **Key Methods**:
  - `add(u_int32_t val)`: Processes 32-bit values bit by bit
  - `finish()`: Finalizes CRC calculation with padding and XOR
  - `get()`: Returns the calculated CRC value

### 2. Hardware CRC (CRC16 with custom table)
- **Location**: `/mft_utils/calc_hw_crc.h` and `/mft_utils/calc_hw_crc.c`
- **Function**: `calc_hw_crc(u_int8_t* data, int size)`
- **Table**: `crc16table2[256]` - pre-calculated CRC table
- **Algorithm**: Table-based CRC16 with special handling for first 2 bytes (inverted)
- **Initial Value**: 0xffff
- **Returns**: 16-bit CRC with byte-swapped result

### 3. CRC32 (XZ Utils)
- **Location**: `/ext_libs/minixz/xz_crc32.c`
- **Algorithm**: IEEE-802.3 polynomial CRC32
- **Usage**: Used for XZ compression/decompression

## Main CRC Functions

### High-Level CRC Calculation
1. **`CalcImageCRC(u_int32_t* buff, u_int32_t size)`**
   - Location: `fw_ops.cpp`
   - Uses Crc16 class
   - Handles endianness conversion (TOCPUn/CPUTOn macros)
   - Returns 32-bit CRC value (extended from 16-bit)

2. **`recalcSectionCrc(u_int8_t* buf, u_int32_t data_size)`**
   - Location: `fw_ops.cpp`
   - Calculates CRC for a section and appends it to the end
   - Uses Crc16 class with big-endian conversion

### ITOC/DTOC CRC Functions
1. **`CalcItocEntryCRC(struct toc_info* curr_toc)`**
   - Location: `fs3_ops.cpp`
   - Calculates CRC for ITOC entries
   - Uses CalcImageCRC internally

2. **`DumpFs3CRCCheck(...)`**
   - Location: `fs3_ops.cpp`, `fs4_ops.cpp`, `fs5_ops.cpp`
   - Verifies CRC for various sections (ITOC, DTOC, HW pointers, etc.)
   - Calls CheckAndPrintCrcRes for result handling

### CRC Verification Functions
1. **`CheckAndPrintCrcRes(...)`**
   - Location: `fw_ops.cpp`
   - Central CRC verification function
   - Handles:
     - Blank CRC (0xffff)
     - CRC ignore flag
     - CRC mismatch reporting
   - Returns error on CRC mismatch unless ignoreCrcCheck is set

## CRC Usage by Section Type

### 1. ITOC (Image Table of Contents)
- **CRC Location**: Last 4 bytes of ITOC header
- **Coverage**: TOC_HEADER_SIZE - 4 bytes
- **Function**: `CalcImageCRC` on ITOC header data

### 2. DTOC (Data Table of Contents)
- **Similar to ITOC**
- **Used for**: Data sections in FS3/FS4 images

### 3. HW Pointers (FS4/FS5)
- **CRC Type**: Hardware CRC (`calc_hw_crc`)
- **Size**: 6 bytes of data per pointer
- **Location**: Last 2 bytes of each HW pointer entry

### 4. Section Data
- **CRC Calculation**: `CalcImageCRC` on section data
- **Storage**: In ITOC/DTOC entry's crc field
- **Special Cases**:
  - NOCRC flag (value = 1) indicates no CRC check
  - Some sections (NV_LOG, NV_DATA) may use NOCRC

### 5. Boot Record (FS4)
- **Function**: `MaskBootRecordCRC`
- **Special handling**: CRC/auth-tag is masked (set to 0xff) during certain operations

### 6. Tools Area
- **Used in**: FS4 operations
- **Size**: IMAGE_LAYOUT_TOOLS_AREA_SIZE

### 7. Hashes Table
- **CRC verification for secure boot hashes**
- **Location**: End of hashes table

## Endianness Handling

### Conversion Macros
- **`TOCPUn(s, n)`**: Convert n 32-bit words from big-endian to CPU endianness
- **`CPUTOn(s, n)`**: Convert n 32-bit words from CPU to big-endian
- **`CRCn(c, s, n)`**: Feed n 32-bit words to CRC object

## Error Handling

### CRC Check Results
1. **OK**: CRC matches expected value
2. **BLANK CRC**: CRC is 0xffff (uninitialized)
3. **CRC IGNORED**: Check skipped due to ignore flag
4. **Wrong CRC**: Mismatch between expected and actual
   - Error message: "Bad CRC."
   - Can be overridden with `ignoreCrcCheck` parameter

### Common CRC Error Patterns
- **BAD_CRC_MSG**: Standard error message for CRC failures
- **0xffff**: Indicates blank/uninitialized CRC
- **NOCRC (1)**: Special value indicating CRC check should be skipped

## Usage Examples

### Calculating Section CRC
```cpp
// Calculate CRC for image section
u_int32_t crc = CalcImageCRC((u_int32_t*)buffer, size_in_dwords);

// Recalculate and append CRC to section
recalcSectionCrc(section_data, section_size);
```

### Verifying CRC
```cpp
// Verify ITOC CRC
DumpFs3CRCCheck(FS3_ITOC, address, size, expected_crc, actual_crc, 
                ignore_crc, verifyCallBackFunc);

// Check and print CRC result
CheckAndPrintCrcRes(description, blank_crc, offset, actual_crc, 
                    expected_crc, ignore_crc, verifyCallBackFunc);
```

### Hardware CRC Usage
```cpp
// Calculate hardware CRC for HW pointer
u_int16_t hw_crc = calc_hw_crc(pointer_data, 6);
```