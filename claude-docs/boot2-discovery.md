# BOOT2 Section Discovery

This document explains how the BOOT2 section is discovered in mlx5fw-go, based on analysis of the mstflint source code.

## Overview

The BOOT2 section discovery method depends on the firmware format:

### FS3 Format (Older Firmware)
For FS3 format firmware, BOOT2 is located at a **hardcoded offset**:
- Offset: `0x38` (56 bytes) from the start of the image
- Defined as: `FS3_BOOT_START = FS2_BOOT_START = 0x38`
- Reference: `mstflint/flint/subcommands/fs3_ops.cpp` line ~839

### FS4 Format (ConnectX-5 and Later)
For FS4 format firmware, BOOT2 location is **dynamically discovered** from HW pointers:
- HW pointers are read from offset `0x18` (24 bytes) from the magic pattern
- The `boot2_ptr` field is at offset `0x8` within the HW pointers structure
- Reference: `mstflint/flint/subcommands/fs4_ops.cpp` lines ~906-910

### Important: BOOT2 is NOT an ITOC Section
BOOT2 is a special section that exists outside of the ITOC (Image Table of Contents). It is not listed in ITOC entries but is discovered through HW pointers and displayed separately by mstflint.

## Size Determination

The BOOT2 size is **never hardcoded**. It is always read dynamically from the BOOT2 header:
- Size is stored at offset 4 in the BOOT2 header
- Formula: `(4 + num_of_dwords) * 4`
  - 4 = header (2 dwords) + tail (2 dwords)
  - num_of_dwords = data size in dwords

## Implementation in mlx5fw-go

Our implementation correctly follows the mstflint approach for FS4:

1. **HW Pointer Reading** (`pkg/parser/fs4/parser.go`):
   ```go
   // Line 133
   p.boot2Addr = p.hwPointers.Boot2Ptr.Ptr
   ```

2. **Dynamic Size Reading** (`pkg/parser/fs4/parser.go`):
   ```go
   // Lines 742-747
   if p.boot2Addr != 0 && p.boot2Addr != 0xffffffff {
       headerData, err := p.reader.ReadSection(int64(p.boot2Addr), 16)
       if err == nil {
           // Boot2 size is at offset 4
           size := binary.BigEndian.Uint32(headerData[4:8])
       }
   }
   ```

## Key mstflint References

- HW pointer structure: `mstflint/tools_layouts/image_layout_layouts.h`
- Boot start definitions: `mstflint/mlxfwops/lib/flint_base.h` lines 273, 279
- FS4 HW pointer offset: `FS4_HW_PTR_START = 0x18`
- FS3 boot offset: `FS3_BOOT_START = 0x38`

## Implementation Details

### Section Type Assignment
BOOT2 uses a special section type value of `0x100` (256) in mlx5fw-go to distinguish it from regular ITOC sections. This is because:
- BOOT2 is not part of ITOC/DTOC
- Section type 0x03 is already used for MAIN_CODE in ITOC
- Using a value > 0xFF ensures no conflicts with regular section types

### Display in sections.go
BOOT2 is discovered dynamically by the parser and added to the sections map with its special type. The display logic in `cmd/mlx5fw-go/sections.go` handles it like any other parsed section, ensuring it appears in the output at the correct address.

## Conclusion

The mlx5fw-go implementation correctly uses dynamic BOOT2 discovery for FS4 firmware, matching mstflint's behavior. There are no hardcoded addresses or sizes for BOOT2 in the FS4 parser. The initial issue where BOOT2 was missing from some firmware outputs has been resolved by adding proper BOOT2 parsing to the main parse flow.