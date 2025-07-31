# Comprehensive List of Section Types in mstflint

Based on analysis of the mstflint source code and our Go implementation, here is a complete list of all section types that can be parsed from Mellanox firmware files.

## ITOC Section Types (Image Table of Contents)

These sections contain firmware code, configuration, and security data:

### Core Firmware Code Sections
| Type | Name | Description |
|------|------|-------------|
| 0x01 | BOOT_CODE | Boot loader code |
| 0x02 | PCI_CODE | PCI initialization code |
| 0x03 | MAIN_CODE | Main firmware code |
| 0x04 | PCIE_LINK_CODE | PCIe link management code |
| 0x05 | IRON_PREP_CODE | Iron preparation code |
| 0x06 | POST_IRON_BOOT_CODE | Post-iron boot code |
| 0x07 | UPGRADE_CODE | Firmware upgrade code |
| 0x0F | BOOT3_CODE | Third stage boot code |
| 0x14 | APU_KERNEL | APU kernel code |
| 0x15 | ACE_CODE | ACE (Acceleration Engine) code |
| 0x18 | ROM_CODE | ROM code |
| 0x2A | PRE_LINK_CODE | Pre-link code |
| 0x2C | POST_LINK_CODE | Post-link code |
| 0x34 | GB_FW_CODE | GB firmware code |
| 0x35 | TILE_FW_CODE | Tile firmware code |

### Configuration Sections
| Type | Name | Description |
|------|------|-------------|
| 0x08 | HW_BOOT_CFG | Hardware boot configuration |
| 0x09 | HW_MAIN_CFG | Hardware main configuration |
| 0x11 | FW_BOOT_CFG | Firmware boot configuration |
| 0x12 | FW_MAIN_CFG | Firmware main configuration |
| 0x20 | RESET_INFO | Reset information |
| 0x30 | DBG_FW_INI | Debug firmware INI |
| 0x32 | DBG_FW_PARAMS | Debug firmware parameters |
| 0x33 | FW_ADB | Firmware ADB |
| 0x36 | FW_TILE_INI | Firmware tile INI |
| 0x37 | HW_TILE_INI | Hardware tile INI |
| 0x40 | SLOT_DEPENDENT_INI | Slot dependent INI |
| 0xAA | PXIR_INI | PXIR INI |
| 0xAB | PXIR_INI1 | PXIR INI1 |
| 0xB0 | EXCLKSYNC_INFO | External clock sync info |

### PHY and UC Code Sections
| Type | Name | Description |
|------|------|-------------|
| 0x0A | PHY_UC_CODE | PHY microcode |
| 0x0B | PHY_UC_CONSTS | PHY microcode constants |
| 0x0C | PCIE_PHY_UC_CODE | PCIe PHY microcode |
| 0x0D | CCIR_INFRA_CODE | CCIR infrastructure code |
| 0x0E | CCIR_ALGO_CODE | CCIR algorithm code |

### Security and Signature Sections
| Type | Name | Description |
|------|------|-------------|
| 0xA0 | IMAGE_SIGNATURE_256 | RSA-2048 image signature |
| 0xA1 | PUBLIC_KEYS_2048 | RSA-2048 public keys |
| 0xA2 | FORBIDDEN_VERSIONS | Forbidden firmware versions |
| 0xA3 | IMAGE_SIGNATURE_512 | RSA-4096 image signature |
| 0xA4 | PUBLIC_KEYS_4096 | RSA-4096 public keys |
| 0xA5 | HMAC_DIGEST | HMAC digest |
| 0xA6 | RSA_PUBLIC_KEY | RSA public key |
| 0xA7 | RSA_4096_SIGNATURES | RSA-4096 signatures |
| 0xA9 | ENCRYPTION_KEY_TRANSITION | Encryption key transition |
| 0xAD | NVDA_ROT_CERTIFICATES | NVIDIA root of trust certificates |
| 0xB1 | MAIN_PAGES_HASHES | Main pages hashes |
| 0xB2 | MAIN_PAGES_LOCKED_HASHES | Main pages locked hashes |

### Data Sections
| Type | Name | Description |
|------|------|-------------|
| 0x10 | IMAGE_INFO | Image information |
| 0x21 | PROG_FW_META | Programmable firmware metadata |
| 0x22 | PROG_FW_BIN | Programmable firmware binary |
| 0x2B | PRE_LINK_DATA | Pre-link data |
| 0x2D | POST_LINK_DATA | Post-link data |
| 0xB4 | STRN_MAIN | String table main |
| 0xB5 | STRN_IRON | String table iron |
| 0xB6 | STRN_TILE | String table tile |
| 0xD3 | MAIN_DATA | Main data |
| 0xD4 | FW_DEBUG_DUMP_2 | Firmware debug dump 2 |
| 0xFA | HASHES_TABLE | Hashes table |

### Special Sections
| Type | Name | Description |
|------|------|-------------|
| 0xF9 | TOOLS_AREA | Tools area |
| 0xFB | HW_PTR | Hardware pointers |
| 0xFC | FW_DEBUG_DUMP | Firmware debug dump |
| 0xFD | ITOC | Image table of contents |
| 0xFE | DTOC | Device table of contents |
| 0xFF | END | End marker |

## DTOC Section Types (Device Table of Contents)

These sections contain device-specific data, certificates, and writable areas. DTOC types are stored as 8-bit values but are OR'd with 0xE000 when processed:

### Device Information
| Raw | Actual | Name | Description |
|-----|--------|------|-------------|
| 0x00 | 0xE000 | MFG_INFO | Manufacturing information |
| 0x01 | 0xE001 | DEV_INFO | Device information |
| 0x07 | 0xE007 | DEV_INFO1 | Device information 1 |
| 0x08 | 0xE008 | DEV_INFO2 | Device information 2 |

### Non-Volatile Data
| Raw | Actual | Name | Description |
|-----|--------|------|-------------|
| 0x02 | 0xE002 | NV_DATA1 | Non-volatile data 1 (deprecated) |
| 0x03 | 0xE003 | VPD_R0 | Vital Product Data R0 |
| 0x04 | 0xE004 | NV_DATA2 | Non-volatile data 2 |
| 0x05 | 0xE005 | FW_NV_LOG | Firmware NV log |
| 0x06 | 0xE006 | NV_DATA0 | Non-volatile data 0 |
| 0x11 | 0xE011 | LC_INI_NV_DATA | LC INI NV data |

### Debug and Internal Usage
| Raw | Actual | Name | Description |
|-----|--------|------|-------------|
| 0x09 | 0xE009 | CRDUMP_MASK_DATA | Core dump mask data |
| 0x0A | 0xE00A | FW_INTERNAL_USAGE | Firmware internal usage |
| 0xD5 | 0xE0D5 | SECURITY_LOG | Security log |

### Programmable Hardware Firmware
| Raw | Actual | Name | Description |
|-----|--------|------|-------------|
| 0x0B | 0xE00B | PROGRAMMABLE_HW_FW1 | Programmable HW firmware 1 |
| 0x0C | 0xE00C | PROGRAMMABLE_HW_FW2 | Programmable HW firmware 2 |

### Digital Certificates
| Raw | Actual | Name | Description |
|-----|--------|------|-------------|
| 0x0D | 0xE00D | DIGITAL_CERT_PTR | Digital certificate pointer |
| 0x0E | 0xE00E | DIGITAL_CERT_RW | Digital certificate R/W |
| 0x12 | 0xE012 | CERT_CHAIN_0 | Certificate chain 0 |
| 0x13 | 0xE013 | DIGITAL_CACERT_RW | Digital CA certificate R/W |
| 0x14 | 0xE014 | CERTIFICATE_CHAINS_1 | Certificate chains 1 |
| 0x15 | 0xE015 | CERTIFICATE_CHAINS_2 | Certificate chains 2 |
| 0x16 | 0xE016 | ROOT_CERTIFICATES_1 | Root certificates 1 |
| 0x17 | 0xE017 | ROOT_CERTIFICATES_2 | Root certificates 2 |

### LC INI Tables
| Raw | Actual | Name | Description |
|-----|--------|------|-------------|
| 0x0F | 0xE00F | LC_INI1_TABLE | LC INI1 table |
| 0x10 | 0xE010 | LC_INI2_TABLE | LC INI2 table |

## Special Section Types (Not in ITOC/DTOC)

| Type | Name | Description |
|------|------|-------------|
| 0x100 | BOOT2 | Boot2 section (referenced by HW pointer) |

## Section Type Ranges

- **Image sections**: 0x10 - 0x5F
- **Device sections**: 0x80 - 0xFF
- **DTOC sections**: 0xE000 - 0xE0FF (when processed)

## CRC Handling

Different sections use different CRC algorithms:
- **Hardware CRC**: HW_PTR (0xFB) sections
- **Software CRC**: All other sections

CRC can be stored in three ways:
1. **CRCInITOCEntry**: CRC stored in ITOC entry
2. **CRCNone**: No CRC verification
3. **CRCInSection**: CRC stored at end of section data

## Implementation Status

Currently implemented parsers:
- ✅ Generic section parser
- ✅ ITOC section parser
- ✅ DTOC section parser
- ✅ Image Info section parser
- ✅ Device Info section parser
- ✅ MFG Info section parser
- ✅ HW Pointer section parser
- ✅ Hashes Table section parser

To be implemented:
- ⬜ Boot code sections
- ⬜ Configuration sections
- ⬜ Security/signature sections
- ⬜ Certificate sections
- ⬜ Debug sections
- ⬜ NV data sections