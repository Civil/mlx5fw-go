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

## Usage with annotations

Structures are parsed via `pkg/annotations` using `offset:"..."` tags:
- `offset:"byte:N"` for explicit byte offsets within the structure
- `offset:"bit:P,len:L,endian:be|le"` for bitfields (absolute bit positions)
- `offset:"...,endian:be|le"` to specify endianness for multi-byte fields
- `offset:"...,hex_as_dec:true"` for BCD-encoded fields (e.g., years, timestamps)

Types provide `Unmarshal([]byte) error` and `Marshal() ([]byte, error)` for binary round-tripping; prefer these over ad-hoc byte slicing implemented in business logic.

## Section Types

ITOC sections (0x0 - 0xFF) and DTOC sections are defined in `section_names.go` with human-readable names.
