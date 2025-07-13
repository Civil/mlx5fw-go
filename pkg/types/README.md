# Firmware Data Structures

This package contains all the data structures used for parsing Mellanox firmware files (FS4 and FS5 formats).

## Structure Overview

All structures are designed to match the binary layout from mstflint, with Big Endian byte order.

### Common Structures

- **HWPointerEntry**: 8-byte structure containing pointer and CRC
- **ITOCHeader**: 32-byte Image Table of Contents header
- **ITOCEntry**: 32-byte ITOC entry with complex bit packing
- **ImageInfo**: 960-byte structure containing firmware metadata
- **DeviceInfo**: Device-specific information
- **MFGInfo**: Manufacturing information
- **VPDData**: Vital Product Data

### FS4-Specific Structures

- **FS4HWPointers**: 64-byte Carmel hardware pointers (8 entries)
- Uses standard ITOC/DTOC structures

### FS5-Specific Structures

- **FS5HWPointers**: 128-byte Gilboa hardware pointers (16 entries)
- **HashesTableHeader**: Header for hashes table
- **HashTableEntry**: Individual hash entries for secure boot

### Security Structures

- **SignatureBlock**: RSA signature blocks (256/512)
- **PublicKeyBlock**: Public key storage (2048/4096)
- **ForbiddenVersionsHeader/Entry**: Version blacklisting

### CRC Implementation

- Software CRC16 with polynomial 0x100b
- Hardware CRC with special table and first 2 bytes inverted
- Three CRC modes: INITOCENTRY, NOCRC, INSECTION

## Usage with binstruct

All structures use `github.com/ghostiam/binstruct` tags:
- `bin:"BE"` for big-endian fields
- `bin:"offset=X"` for specific offsets
- `bin:""` for byte arrays without endianness conversion

## Section Types

ITOC sections (0x0 - 0xFF) and DTOC sections are defined in `section_names.go` with human-readable names.