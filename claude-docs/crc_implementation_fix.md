# CRC Implementation Fix Documentation

## Issue
The initial CRC implementation was failing verification for ConnectX-5 firmware sections. The calculated CRC values did not match the expected values stored in the firmware.

## Root Cause
The CRC calculation algorithm was not correctly implementing the bit-by-bit processing used by mstflint. Specifically:

1. **Incorrect bit processing**: The original implementation was checking CRC and data bits separately using XOR logic
2. **Missing bit combination**: mstflint combines the CRC shift with the data bit using OR operation: `(crc << 1) | (word >> 31)`

## Solution
Updated the CRC calculation to exactly match mstflint's Crc16::add() implementation:

```go
// Process each bit of the 32-bit word (matches mstflint's Crc16::add)
for j := 0; j < 32; j++ {
    if crc & 0x8000 != 0 {
        crc = ((crc << 1) | uint16(word >> 31)) ^ types.CRCPolynomial
    } else {
        crc = (crc << 1) | uint16(word >> 31)
    }
    crc &= 0xFFFF
    word <<= 1
}
```

## Key Implementation Details

### mstflint CRC16 Algorithm
- **Polynomial**: 0x100b
- **Initial value**: 0xFFFF
- **Final XOR**: 0xFFFF
- **Processing**: 32-bit words in big-endian format
- **Bit operation**: Combines CRC shift with data bit before XOR with polynomial

### Critical Differences from Common CRC Implementations
1. **Bit combination**: Uses `(crc << 1) | (data_bit)` instead of separate XOR operations
2. **Word processing**: Processes 32-bit words as a unit, not bytes
3. **Finish step**: Processes 16 additional zero bits after data
4. **Masking**: Applies 0xFFFF mask after each operation to ensure 16-bit result

## Verification
After implementing the fix:
- All 46 sections with CRC verification in the ConnectX-5 sample firmware now pass
- No CRC failures reported
- Output matches mstflint verification results

## Implementation Files
- `/pkg/parser/crc.go`: Contains the fixed CRC implementation
- Functions updated:
  - `CalculateSoftwareCRC16()`: General software CRC calculation
  - `CalculateImageCRC()`: Image-specific CRC calculation (takes size in dwords)