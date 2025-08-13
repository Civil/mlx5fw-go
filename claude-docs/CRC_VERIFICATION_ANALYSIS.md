# CRC Verification Implementations in mstflint

## Overview
This document analyzes the CRC verification implementations for various sections in the mstflint codebase, focusing on when hardware CRC vs software CRC is used.

## CRC Types in FS4 Operations

The codebase defines three CRC types in `fs4_ops.h`:
```cpp
enum CRCTYPE
{
    INITOCENTRY = 0,  // CRC is in the ITOC entry
    NOCRC = 1,        // No CRC verification
    INSECTION = 2     // CRC is in the section itself
};
```

## 1. ITOC/DTOC CRC Verification

### verifyTocHeader (fs4_ops.cpp:467)
- Verifies the TOC (Table of Contents) header CRC
- Uses **software CRC** calculation via `CalcImageCRC()`
- Compares calculated CRC with `itocHeader.itoc_entry_crc`
- Handles both ITOC and DTOC headers

### verifyTocEntries (fs4_ops.cpp:636)
- Verifies individual TOC entries
- Uses **software CRC** for entry verification via `CalcImageCRC()`
- Handles three CRC modes:
  - `INITOCENTRY`: CRC stored in TOC entry
  - `INSECTION`: CRC stored at end of section data
  - `NOCRC`: No CRC verification performed
- Special handling for encrypted sections (ignores CRC)

## 2. Hardware Pointers CRC Verification

### getExtendedHWAravaPtrs (fs4_ops.cpp:246)
- Parses hardware pointers with CRC verification
- **Dual CRC approach**:
  1. First tries **software CRC** via `CalcImageCRC()`
  2. If software CRC doesn't match, falls back to **hardware CRC** via `calc_hw_crc()`
- Some devices (e.g., QT3) use software CRC for hardware pointers
- Each pointer has its own CRC (8 bytes total: 4 bytes pointer + 4 bytes CRC)

### getExtendedHWPtrs (fs4_ops.cpp:356)
- Alternative hardware pointer parsing (for ConnectX-6)
- Uses **hardware CRC** exclusively via `calc_hw_crc()`
- No fallback to software CRC

## 3. Boot2 CRC Verification

### CheckBoot2 (fw_ops.cpp:282)
- Verifies Boot2 section integrity
- Uses **software CRC** (CRC16) via custom implementation
- CRC stored at end of Boot2 section
- Only performs full CRC check if `fullRead=true` or not accessing flash

## 4. Tools Area CRC Verification

### verifyToolsArea (fs4_ops.cpp:408)
- Verifies tools area section
- Uses **software CRC** via `CalcImageCRC()`
- CRC is last 4 bytes of tools area
- Also validates binary version compatibility

## 5. Hashes Table CRC Verification

### Hashes Table Header (fs4_ops.cpp:964)
- Verifies hashes table header integrity
- Uses **software CRC** via `CalcImageCRC()`
- CRC stored in last 2 bytes (16-bit) of header

### Hashes Table (fs4_ops.cpp:998)
- Verifies entire hashes table
- Uses **software CRC** via `CalcImageCRC()`
- CRC stored at end of table

## 6. Section Data CRC Verification

### General Section Verification (in verifyTocEntries)
- Most sections use **software CRC**
- CRC location depends on `tocEntry.crc` field:
  - `INITOCENTRY`: CRC in TOC entry itself
  - `INSECTION`: CRC at end of section data
  - `NOCRC`: No CRC check
- Special cases:
  - Encrypted sections: CRC ignored
  - Cache line CRC sections: May have auth-tag instead of CRC

## 7. NOCRC Flag Handling

The `NOCRC` flag (value 1) indicates sections that should not have CRC verification:
- When `tocEntry.crc == NOCRC`, CRC verification is skipped
- Used for sections that may be modified or don't require integrity checking
- The `DumpFs3CRCCheck()` function is called with `ignore_crc=true`

## Hardware vs Software CRC Usage Summary

### Hardware CRC (`calc_hw_crc`)
- Used for hardware pointers in newer devices
- More efficient for hardware-specific structures
- Implementation likely matches hardware's CRC calculation method

### Software CRC (`CalcImageCRC` / CRC16)
- Used for most sections and structures
- Default for:
  - TOC headers and entries
  - Boot2 sections
  - Tools area
  - Hashes table
  - General section data
- Provides flexibility and compatibility across devices

### Special Cases
- Hardware pointers may use either hardware or software CRC depending on device
- Encrypted sections skip CRC verification entirely
- Some devices (QT3) use software CRC even for hardware pointers

## Key Functions

1. `CalcImageCRC()`: Software CRC calculation (CRC16-based)
2. `calc_hw_crc()`: Hardware CRC calculation
3. `DumpFs3CRCCheck()`: Common CRC verification and reporting function
4. `CheckBoot2()`: Specialized Boot2 verification
5. `verifyTocHeader()`: TOC header verification
6. `verifyTocEntries()`: TOC entries verification