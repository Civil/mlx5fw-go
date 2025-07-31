# TOOLS_AREA Address Determination in FS4 Format

## Summary

Based on analysis of the mstflint codebase, the TOOLS_AREA address in FS4 format is determined through **hardware pointers** stored in the firmware image, not at a fixed offset.

## Key Findings

### 1. Hardware Pointers Location
- Hardware pointers start at a fixed offset: `FS4_HW_PTR_START = 0x18`
- This is defined in `/reference/mstflint/mlxfwops/lib/flint_base.h`

### 2. Hardware Pointers Structure
The hardware pointers structure (defined in `cx6fw_layouts.h`) contains four pointer entries:
```
Offset  | Pointer Name         | Size
--------|---------------------|------
0x00    | boot_record_ptr     | 8 bytes
0x08    | boot2_ptr           | 8 bytes  
0x10    | toc_ptr             | 8 bytes
0x18    | tools_ptr           | 8 bytes
```

Each pointer entry consists of:
- 4 bytes: pointer value (offset in firmware)
- 2 bytes: CRC16 (hardware-calculated)
- 2 bytes: padding/reserved

### 3. TOOLS_AREA Address Calculation
The TOOLS_AREA address is calculated as:
```
TOOLS_AREA_ADDRESS = firmware_start + tools_ptr_value
```

Where:
- `firmware_start` is typically 0 for firmware files
- `tools_ptr_value` is read from offset 0x30 (0x18 + 0x18) in the firmware

### 4. Specific Example: 900-9D3B6-00CV-AAB_4_PORTS_MH_INT_Ax_tree_splits_ptr.bin

For this firmware file:
- At offset 0x30: value is `0x00000600` (big-endian)
- Therefore, TOOLS_AREA is at offset `0x600` in the firmware

This was verified by:
1. Reading the hardware pointers directly from the firmware
2. Confirming the tools_ptr value is 0x00000600
3. Verifying data exists at offset 0x600

## Code Flow in mstflint

1. **fs4_ops.cpp**: `Fs4Operations::verifyToolsArea()`
   - Calculates physical address: `physAddr = _fwImgInfo.imgStart + _tools_ptr`
   - `_tools_ptr` is populated from hardware pointers

2. **fs4_ops.cpp**: Hardware pointers reading
   - Reads from offset `FS4_HW_PTR_START` (0x18)
   - Unpacks the cx6fw_hw_pointers structure
   - Assigns: `_tools_ptr = hw_pointers.tools_ptr.ptr`

## Conclusion

The TOOLS_AREA address in FS4 format is **not at a fixed offset**. Instead, it is determined by reading the hardware pointers structure at offset 0x18 in the firmware, specifically the tools_ptr field at offset 0x30. For the analyzed firmware file, this pointer contains the value 0x600, which explains why mstflint shows TOOLS_AREA at that address.