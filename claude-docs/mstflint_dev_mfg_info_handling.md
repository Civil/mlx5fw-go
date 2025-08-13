# mstflint DEV_INFO and MFG_INFO Section Handling

## Structure Sizes

Based on analysis of mstflint source code, the actual structure sizes are:

### DEV_INFO (Section Type 0xE1 in DTOC, 0x11 in ITOC)
- **Structure Size**: 512 bytes (0x200)
- Defined in `image_layout_layouts.h` as `IMAGE_LAYOUT_DEVICE_INFO_SIZE`
- Also defined in `cibfw_layouts.h` as `CIBFW_DEVICE_INFO_SIZE` (same size)

### MFG_INFO (Section Type 0xE0 in DTOC, 0x10 in ITOC)
- **Structure Size**: 
  - Standard: 320 bytes (0x140) - Defined in `image_layout_layouts.h` as `IMAGE_LAYOUT_MFG_INFO_SIZE`
  - Reduced: 256 bytes (0x100) - Found in some firmware versions
- The actual size varies by firmware and should be taken from the ITOC/DTOC entry

## CRC Handling

mstflint handles CRC for these sections based on the ITOC/DTOC entry flags:

1. **CRC Type Determination** (from `toc_reader.go` and ITOC entry structure):
   - If `no_crc` flag is set (bit 16 of offset 0x18): `CRCNone`
   - If `section_crc` field (offset 0x18) is non-zero: `CRCInITOCEntry`
   - Otherwise: `CRCInSection` (CRC appended after the section data)

2. **CRC Location**:
   - **CRCInITOCEntry**: CRC is stored in the ITOC/DTOC entry at offset 0x18
   - **CRCInSection**: CRC is appended as 4 bytes after the section data
     - For DEV_INFO: CRC would be at offset 0x200 (after the 512-byte structure)
     - For MFG_INFO: CRC would be at offset 0x140 (after the 320-byte structure)

## Size Field in ITOC/DTOC

The size field in ITOC/DTOC entries is stored in **dwords** (4-byte units):
- `entry.SizeDwords` contains the size divided by 4
- To get actual byte size: `size_bytes = entry.SizeDwords * 4`

## Version-Based Structure Handling

mstflint checks structure versions to handle different formats:

### DEV_INFO Versions
- **Version 2** (new format): Uses `image_layout_device_info` structure
- **Version 1** (old format): Uses `cibfw_device_info` structure
- Check: `major_version == 2` for new format

### MFG_INFO Versions
- **Version 1** (new format): Uses `image_layout_mfg_info` structure
- **Version 0** (old format): Uses `cibfw_mfg_info` structure
- Check: `major_version == 1` for new format

## Reading Section Data

When reading these sections:

1. **Determine Read Size**:
   ```
   readSize = entry.GetSize()  // This is already in bytes
   if entry.CRCType == CRCInSection && !isEncrypted {
       readSize += 4  // Add 4 bytes for CRC
   }
   ```

2. **Parse Data**:
   ```
   data = ReadSection(offset, readSize)
   if entry.CRCType == CRCInSection && !isEncrypted && len(data) >= 4 {
       parseData = data[:len(data)-4]  // Remove CRC bytes for parsing
   }
   ```

## Common Issues and Solutions

### Issue: "data is too short" errors
This typically happens when:
1. The ITOC/DTOC entry size is in dwords but treated as bytes
2. The CRC bytes are not accounted for when reading IN_SECTION CRC
3. The structure expects more data than available

### Solution:
1. Ensure `GetSize()` multiplies dwords by 4
2. Add 4 bytes to read size when `CRCType == CRCInSection`
3. Pass data without CRC bytes to the unmarshal function

## Example ITOC Entry for DEV_INFO

```
Type: 0xE1 (DTOC) or 0x11 (ITOC)
SizeDwords: 0x80 (128 dwords = 512 bytes)
FlashAddr: <offset in firmware>
SectionCRC: 0x0000 (if CRC is in section) or actual CRC value
NoCRC: false
```

## Implementation Notes

1. Both `image_layout_*` and `cibfw_*` structures have the same size
2. mstflint uses the version field to determine which structure format to use
3. The actual data parsing uses little-endian byte order for these sections
4. Some firmware (like encrypted ConnectX-7) may have all-FF values in DEV_INFO, requiring fallback to MFG_INFO

## mlx5fw-go Implementation Fix

The issue in mlx5fw-go is that the MfgInfoAnnotated structure expects 320 bytes, but some firmware only provides 256 bytes. The solution is:

1. **Check actual size from ITOC/DTOC entry** before parsing
2. **Handle variable-sized MFG_INFO structures**:
   - For 256-byte version: Parse only the available fields
   - For 320-byte version: Parse the full structure
3. **Update the MfgInfoAnnotated structure** to handle optional fields at the end

The DEV_INFO parsing works correctly because the firmware provides the expected 512 bytes.