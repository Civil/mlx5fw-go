# TOOLS_AREA CRC Calculation in mstflint

## Overview

The TOOLS_AREA section in Mellanox firmware images contains metadata about the firmware binary and includes a CRC-16 checksum for data integrity verification. This document details how mstflint calculates and verifies this CRC.

## TOOLS_AREA Structure

The TOOLS_AREA is a 64-byte (0x40) structure located at various offsets in the firmware image, commonly at 0x500. The structure layout is:

```c
struct image_layout_tools_area {
    // DWORD[0] (Offset 0x0)
    u_int8_t minor;                    // 0x0.0 - 0x0.7   - tools area minor version
    u_int8_t major;                    // 0x0.8 - 0x0.15  - tools area major version
    
    // DWORD[1] (Offset 0x4)
    u_int8_t bin_ver_minor;            // 0x4.0 - 0x4.7   - binary version minor
    u_int8_t bin_ver_major;            // 0x4.8 - 0x4.15  - binary version major
    u_int16_t log2_img_slot_size;      // 0x4.16 - 0x4.31 - log2 image slot size
    
    // DWORD[2-14] (Offset 0x8-0x38)
    // Reserved/unused bytes (all zeros in typical images)
    
    // DWORD[15] (Offset 0x3c)
    u_int16_t crc;                     // 0x3c.0 - 0x3c.15 - CRC-16 checksum
};
```

## CRC Algorithm Details

### Algorithm Type: CRC-16 with custom polynomial

- **Polynomial**: 0x100b
- **Initial value**: 0xffff
- **Final XOR**: 0xffff
- **Bit order**: MSB first
- **Result**: 16-bit value stored in big-endian format

### Implementation

The CRC calculation is performed by the `Crc16` class in mstflint:

```cpp
class Crc16 {
    u_int16_t _crc = 0xffff;
    
    void add(u_int32_t o) {
        for (int i = 0; i < 32; i++) {
            if (_crc & 0x8000) {
                _crc = (u_int16_t)((((_crc << 1) | (o >> 31)) ^ 0x100b) & 0xffff);
            } else {
                _crc = (u_int16_t)(((_crc << 1) | (o >> 31)) & 0xffff);
            }
            o = (o << 1) & 0xffffffff;
        }
    }
    
    void finish() {
        for (int i = 0; i < 16; i++) {
            if (_crc & 0x8000) {
                _crc = ((_crc << 1) ^ 0x100b) & 0xffff;
            } else {
                _crc = (_crc << 1) & 0xffff;
            }
        }
        _crc = _crc ^ 0xffff;  // Final XOR
    }
};
```

## TOOLS_AREA CRC Calculation Process

### 1. Data Reading (fs4_ops.cpp)

```cpp
u_int32_t buff[IMAGE_LAYOUT_TOOLS_AREA_SIZE / 4] = {0};  // 16 dwords
READBUF((*_ioAccess), physAddr, buff, IMAGE_LAYOUT_TOOLS_AREA_SIZE, "Tools Area");
image_layout_tools_area_unpack(&tools_area, (u_int8_t*)buff);
```

### 2. CRC Calculation

```cpp
calculatedToolsAreaCRC = CalcImageCRC((u_int32_t*)buff, IMAGE_LAYOUT_TOOLS_AREA_SIZE / 4 - 1);
```

The `CalcImageCRC` function:
1. Converts data from big-endian to host endian
2. Processes data through the CRC algorithm
3. Converts data back to big-endian
4. Returns the 16-bit CRC result

Key points:
- **Data size**: 15 dwords (60 bytes) - excludes the last 4 bytes containing the CRC itself
- **Endianness handling**: Data is converted to host endian before CRC calculation

### 3. Verification

```cpp
toolsAreaCRC = tools_area.crc;  // Read CRC from structure
if (!DumpFs3CRCCheck(FS4_TOOLS_AREA, physAddr, IMAGE_LAYOUT_TOOLS_AREA_SIZE, 
                     calculatedToolsAreaCRC, toolsAreaCRC, false, verifyCallBackFunc)) {
    return false;
}
```

## Example Calculation

For the TOOLS_AREA data at offset 0x500:
```
00000500: 0000 0000 0018 0100 0000 0000 0000 0000  ................
00000510: 0000 0000 0000 0000 0000 0000 0000 0000  ................
00000520: 0000 0000 0000 0000 0000 0000 0000 0000  ................
00000530: 0000 0000 0000 0000 0000 0000 0000 83dc  ................
```

Breaking down the data:
- **Offset 0x0-0x3**: 0x00000000 (versions: minor=0, major=0)
- **Offset 0x4-0x7**: 0x00180100 (bin_ver_minor=0, bin_ver_major=1, log2_img_slot_size=0x0018)
- **Offset 0x8-0x3b**: All zeros (reserved)
- **Offset 0x3c-0x3f**: 0x000083dc (CRC=0x83dc in last 2 bytes)

The CRC calculation processes the first 60 bytes (15 dwords) in host-endian format:
1. Convert each dword from big-endian to host-endian
2. Process through CRC-16 algorithm with polynomial 0x100b
3. Apply final XOR with 0xffff
4. Result: 0x83dc

## Implementation in Go

To implement TOOLS_AREA CRC calculation in Go:

```go
type CRC16 struct {
    crc uint16
}

func NewCRC16() *CRC16 {
    return &CRC16{crc: 0xffff}
}

func (c *CRC16) Add(val uint32) {
    for i := 0; i < 32; i++ {
        if c.crc&0x8000 != 0 {
            c.crc = uint16((((uint32(c.crc) << 1) | (val >> 31)) ^ 0x100b) & 0xffff)
        } else {
            c.crc = uint16(((uint32(c.crc) << 1) | (val >> 31)) & 0xffff)
        }
        val = (val << 1) & 0xffffffff
    }
}

func (c *CRC16) Finish() uint16 {
    for i := 0; i < 16; i++ {
        if c.crc&0x8000 != 0 {
            c.crc = ((c.crc << 1) ^ 0x100b) & 0xffff
        } else {
            c.crc = (c.crc << 1) & 0xffff
        }
    }
    c.crc = c.crc ^ 0xffff
    return c.crc
}

func CalculateToolsAreaCRC(data []byte) uint16 {
    if len(data) != 64 {
        panic("TOOLS_AREA must be 64 bytes")
    }
    
    crc := NewCRC16()
    
    // Process first 15 dwords (60 bytes), excluding CRC field
    for i := 0; i < 15; i++ {
        // Read big-endian dword
        dword := binary.BigEndian.Uint32(data[i*4:])
        crc.Add(dword)
    }
    
    return crc.Finish()
}
```

## Special Considerations

1. **Byte Swapping**: The TOOLS_AREA data is stored in big-endian format, but the CRC algorithm processes data in host-endian format. The `CalcImageCRC` function handles this conversion automatically.

2. **CRC Field Exclusion**: The CRC calculation covers only the first 60 bytes of the 64-byte structure, excluding the 4-byte field containing the CRC itself.

3. **No Special TOOLS_AREA Handling**: Unlike some other sections that may use hardware CRC (`calc_hw_crc`), TOOLS_AREA uses the standard software CRC-16 implementation.

## Debugging

To debug TOOLS_AREA CRC issues:

1. Enable debug output:
   ```bash
   export FW_COMPS_DEBUG=1
   mstflint -i fw.bin v
   ```

2. Check the debug output for CRC verification messages

3. Use gdb to step through the CRC calculation:
   ```bash
   gdb mstflint
   (gdb) break CalcImageCRC
   (gdb) run -i fw.bin v
   ```