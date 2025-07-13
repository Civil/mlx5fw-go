# mlx5fw CLI Tool

The `mlx5fw` command-line tool provides functionality to parse and analyze Mellanox firmware files (FS4 and FS5 formats).

## Building

```bash
go build ./cmd/mlx5fw
```

## Commands

### sections

The `sections` command displays all sections found in a firmware file.

#### Usage

```bash
# List all sections
mlx5fw sections -f <firmware.bin>

# Show section contents in human-readable format
mlx5fw sections -f <firmware.bin> -c

# Enable verbose logging
mlx5fw sections -f <firmware.bin> -v
```

#### Options

- `-f, --file`: Firmware file path (required)
- `-c, --content`: Show section content in human-readable format
- `-v, --verbose`: Enable verbose logging

#### Output

The command displays:
- ITOC (Image Table of Contents) sections
- DTOC (Device Table of Contents) sections
- Section type, name, offset, size, CRC type, and encryption status

When using the `-c` flag, additional details are shown:
- IMAGE_INFO: Firmware version, release date, description, VSD
- Code sections: Size and hex dump preview
- TOOLS_AREA: Binary version header details
- Signature sections: Signature type and length
- Device data sections: Specific parsing for DEV_INFO, MFG_INFO, VPD_R0

#### Example Output

```
Firmware File: fw-ConnectX5-rel-16_35_4030.bin
================================================================================
Type  Name  Offset  Size  CRC Type  Encrypted
--------------------------------------------------------------------------------

ITOC Sections:
0x0000  IMAGE_INFO           0x00010000  0x000003C0  ITOC_ENTRY  No
0x0001  MAIN_CODE            0x00020000  0x00100000  IN_SECTION  No
0x0006  BOOT_CODE            0x00120000  0x00010000  ITOC_ENTRY  No

DTOC Sections:
0x0001  DEV_INFO  0x00200000  0x00001000  IN_SECTION  No
0x0002  MFG_INFO  0x00201000  0x00000100  IN_SECTION  No
================================================================================
Total Sections: 5 (ITOC: 3, DTOC: 2)
```

## Notes

- This is a placeholder implementation in the preparation phase
- Actual firmware parsing will be implemented in future phases
- Section data shown is example data for demonstration purposes