# DTOC Section Type Mapping

## Overview

DTOC (Device Table of Contents) section types in FS3/FS4 firmware use a different encoding scheme than ITOC (Image Table of Contents) section types. The key findings about DTOC section type mapping are:

## DTOC Type Encoding

1. **Raw DTOC Types**: In the DTOC entry structure, section types are stored as 8-bit values (0x00-0xFF)
2. **Actual Section Types**: When processing DTOC entries, these raw types are OR'd with 0xE000 to create the actual 16-bit section type identifier
3. **0xE Prefix**: All DTOC section types start with 0xE when represented as 16-bit values

## Why 0xE Prefix?

The 0xE prefix serves to distinguish DTOC sections from ITOC sections in the firmware's section namespace:
- ITOC sections use types like 0x01-0xAA (BOOT_CODE, MAIN_CODE, etc.)
- DTOC sections use types like 0xE0-0xF3 (MFG_INFO, DEV_INFO, etc.)

This separation ensures that device-specific data sections don't conflict with image/code sections.

## DTOC Section Type Mapping

Based on the mstflint source code (flint_base.h), here are the actual DTOC section type values:

| Raw DTOC Type | Actual Section Type | Section Name |
|---------------|---------------------|--------------|
| 0x00 | 0xE000 | MFG_INFO |
| 0x01 | 0xE001 | DEV_INFO |
| 0x02 | 0xE002 | NV_DATA1 (deprecated) |
| 0x03 | 0xE003 | VPD_R0 |
| 0x04 | 0xE004 | NV_DATA2 |
| 0x05 | 0xE005 | FW_NV_LOG |
| 0x06 | 0xE006 | NV_DATA0 |
| 0x07 | 0xE007 | DEV_INFO1 |
| 0x08 | 0xE008 | DEV_INFO2 |
| 0x09 | 0xE009 | CRDUMP_MASK_DATA |
| 0x0A | 0xE00A | FW_INTERNAL_USAGE |
| 0x0B | 0xE00B | PROGRAMMABLE_HW_FW1 |
| 0x0C | 0xE00C | PROGRAMMABLE_HW_FW2 |
| 0x0D | 0xE00D | DIGITAL_CERT_PTR |
| 0x0E | 0xE00E | DIGITAL_CERT_RW |
| 0x0F | 0xE00F | LC_INI1_TABLE |
| 0x10 | 0xE010 | LC_INI2_TABLE |
| 0x11 | 0xE011 | LC_INI_NV_DATA |
| 0x12 | 0xE012 | CERT_CHAIN_0 |
| 0x13 | 0xE013 | DIGITAL_CACERT_RW |

Note: In the C++ code, these are defined with the 0xE prefix directly (e.g., FS3_MFG_INFO = 0xe0).

## Implementation in Go

In the Go implementation (pkg/parser/fs4/parser.go), the mapping is handled as follows:

```go
// In parseDTOC() method:
// DTOC sections use different type mapping
sectionType := entry.GetType() | 0xE000
```

This takes the raw 8-bit type from the DTOC entry and ORs it with 0xE000 to create the final section type.

## Key Differences from ITOC

1. **Type Range**: ITOC types range from 0x01-0xAA, while DTOC types use 0xE0-0xF3
2. **Data Type**: ITOC sections typically contain firmware code and configuration, while DTOC sections contain device-specific data like manufacturing info, VPD, certificates, etc.
3. **CRC Handling**: DTOC sections often have CRC_IGNORED flag set, as they contain writable/variable data

## References

- Original C++ definitions: `reference/mstflint/mlxfwops/lib/flint_base.h`
- Go implementation: `pkg/parser/fs4/parser.go` (parseDTOC method)
- Section name mapping: `pkg/types/section_names.go` (GetDTOCSectionTypeName function)