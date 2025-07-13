# Mstflint Firmware Parsing Process Documentation

## Overview
This document describes the firmware parsing process for ConnectX-5 and ConnectX-6 Dx firmware based on analyzing mstflint with `FW_COMPS_DEBUG=1` output.

## Sample Firmware Analysis

### ConnectX-5
Firmware: `fw-ConnectX5-rel-16_35_4030-MCX516A-CDA_Ax_Bx-UEFI-14.29.15-FlexBoot-3.6.902.bin`
- HW ID: 0x20d
- Chunk size: 2^23 (8MB)
- Image size: ~17MB

### ConnectX-6 Dx
Firmware: `fw-ConnectX6Dx-rel-22_41_1000-MCX623106AN-CDA_Ax-UEFI-14.34.12-FlexBoot-3.7.400.bin`
- HW ID: 0x212
- Chunk size: 2^24 (16MB)
- Image size: ~33MB
- Notable differences:
  - Searches additional offset 0x1000000 for magic pattern
  - Contains additional ITOC sections: ACE_CODE, CCIR_INFRA_CODE, CCIR_ALGO_CODE, RSA_PUBLIC_KEY, RSA_4096_SIGNATURES
  - Contains additional DTOC sections for certificate management: ROOT_CERTIFICATES_1/2, CERTIFICATE_CHAINS_1/2, DIGITAL_CERT_RW, DIGITAL_CACERT_RW, CERT_CHAIN_0, DIGITAL_CERT_PTR
  - Larger DTOC at 0x01fff000 (vs 0x00fff000 for CX5)

### BlueField-2
Firmware: `fw-BlueField-2-rel-24_41_1000-MBF2M516A-CENO_Ax_Bx-NVME-20.4.1-UEFI-21.4.13-UEFI-22.4.13-UEFI-14.34.12-FlexBoot-3.7.400.bin`
- HW ID: 0x214
- Chunk size: 2^24 (16MB)
- Image size: ~33MB
- Notable characteristics:
  - Same parsing flow as ConnectX-6 Dx
  - Searches up to 0x1000000 offset for magic pattern (but found at 0x0)
  - Contains all the same advanced security sections as CX6Dx (ACE_CODE, CCIR_*, RSA_*, certificate sections)
  - DTOC at 0x01fff000
  - Supports multiple ROM types: NVME, UEFI (multiple versions), FlexBoot
  - BlueField-specific: designed for DPU (Data Processing Unit) applications
  - **ITOC sections** (29 total):
    - Security-related: IMAGE_SIGNATURE_256/512, PUBLIC_KEYS_2048/4096, RSA_PUBLIC_KEY, RSA_4096_SIGNATURES, FORBIDDEN_VERSIONS
    - Code sections: IRON_PREP_CODE, MAIN_CODE, PCI_CODE, PCIE_LINK_CODE, POST_IRON_BOOT_CODE, UPGRADE_CODE
    - Advanced features: ACE_CODE, CCIR_INFRA_CODE, CCIR_ALGO_CODE (crypto/compression infrastructure)
    - PHY/PCIe: PHY_UC_CODE, PCIE_PHY_UC_CODE, PHY_UC_CONSTS
    - Configuration: IMAGE_INFO, FW_MAIN_CFG, FW_BOOT_CFG, HW_MAIN_CFG, HW_BOOT_CFG
    - Debug/diagnostic: DBG_FW_INI, DBG_FW_PARAMS, CRDUMP_MASK_DATA
    - ROM/Boot: ROM_CODE (contains NVME + multiple UEFI versions + FlexBoot)
    - Other: RESET_INFO
  - **DTOC sections** (15 total):
    - Certificate management: ROOT_CERTIFICATES_1/2, CERTIFICATE_CHAINS_1/2, CERT_CHAIN_0, DIGITAL_CERT_RW, DIGITAL_CACERT_RW, DIGITAL_CERT_PTR
    - Device data: DEV_INFO, MFG_INFO, VPD_R0
    - Non-volatile storage: FW_NV_LOG, NV_DATA (2 instances)
    - Programmable sections: PROGRAMMABLE_HW_FW (2 instances), FW_INTERNAL_USAGE
  - **Key differences from ConnectX-5**:
    - 16 additional HW_POINTERS entries (0x18-0x97 vs 0x18-0x5f for CX5)
    - Extensive certificate/security infrastructure for DPU trusted computing
    - ACE_CODE and CCIR_* sections for hardware acceleration
    - Multiple ROM images for different boot scenarios

### ConnectX-7
Firmware: `fw-ConnectX7-rel-28_33_0800.bin`
- HW ID: 0x218
- Firmware version: 28.33.0800
- Security: NOT ENCRYPTED (is_encrypted = FALSE)
- Image format: FS4
- Chunk size: 2^24 (16MB)
- Image size: ~6.7MB
- Notable characteristics:
  - Same parsing flow as ConnectX-6 Dx
  - DTOC at 0x01fff000
  - Contains BOOT3_CODE (0xf) section - new for CX7
  - Contains APU_KERNEL (0x14) section - new for CX7
  - Has all the same advanced security sections as CX6Dx
  - **ITOC sections** (31 total):
    - New sections: BOOT3_CODE, APU_KERNEL
    - Security-related: IMAGE_SIGNATURE_256/512, PUBLIC_KEYS_2048/4096, RSA_PUBLIC_KEY, RSA_4096_SIGNATURES, FORBIDDEN_VERSIONS
    - Code sections: IRON_PREP_CODE, MAIN_CODE, PCI_CODE, PCIE_LINK_CODE, POST_IRON_BOOT_CODE, UPGRADE_CODE
    - Advanced features: CCIR_INFRA_CODE, CCIR_ALGO_CODE
    - PHY/PCIe: PHY_UC_CODE, PCIE_PHY_UC_CODE, PHY_UC_CONSTS
    - Configuration: IMAGE_INFO, FW_MAIN_CFG, FW_BOOT_CFG, HW_MAIN_CFG, HW_BOOT_CFG
    - Debug/diagnostic: DBG_FW_INI, DBG_FW_PARAMS, CRDUMP_MASK_DATA
  - **DTOC sections** (13 total):
    - Certificate management: DIGITAL_CERT_RW, DIGITAL_CACERT_RW, CERT_CHAIN_0, DIGITAL_CERT_PTR
    - Device data: DEV_INFO, MFG_INFO, VPD_R0
    - Non-volatile storage: FW_NV_LOG, NV_DATA (2 instances)
    - Programmable sections: PROGRAMMABLE_HW_FW (2 instances), FW_INTERNAL_USAGE

### Encrypted Firmware (BlueField-3)
Firmware: `900-9D3B6-00CN-A_Ax_MT_0000000883_rel-32_45.1020.bin`
- **Device Type**: BlueField-3 (B3240 P-Series DPU)
- **HW ID**: 0x21c
- Firmware version: 32.45.1020
- Security: ENCRYPTED (is_encrypted = TRUE)
- Image format: FS4
- Chip type: 20 (BlueField-3)
- Part Number: 900-9D3B6-00CN-A_Ax
- PSID: MT_0000000883
- Security attribute: "secure-fw"

#### Encrypted Firmware Parsing Flow:
1. **Magic pattern detection**: Still found at 0x0 (same as non-encrypted)
2. **Encryption detection**:
   - `IsEncryptedImage()` checks ITOC header at 0x5000
   - Invalid signature (not 0x49544f43 "ITOC") indicates encryption
3. **Limited parsing for encrypted images**:
   - IMAGE_INFO at 0x786d80 (from HW pointer)
   - RSA_PUBLIC_KEY at 0x78d480 (from HW pointer)
   - DTOC still parsed normally (not encrypted)
   - **ITOC parsing skipped** - encrypted content
4. **Available information**:
   - Basic firmware info from IMAGE_INFO
   - Device capabilities from DTOC sections
   - Limited verification possible

### ConnectX-8
Firmware: `fw-ConnectX8-rel-40_45_1200.bin`
- **Device Type**: ConnectX-8 C8240 HHHL SuperNIC
- **HW ID**: 0x21e
- Firmware version: 40.45.1200
- Security: NOT ENCRYPTED
- Image format: FS5
- Chip type: 16
- PSID: MT_0000001222
- Security attribute: "secure-fw"
- Notable characteristics:
  - Uses FS5 format with enhanced hardware pointers
  - NCORE BCH (Boot Component Header) support
  - Hashes table for enhanced security
  - Same DTOC location as other modern devices (0x01fff000)
  - Advanced security but not encrypted

## Key Observations

### Magic Pattern Search
Both devices search for magic pattern (`MT_FW__`) at these offsets:
- 0x0 (found in all samples)
- 0x10000, 0x20000, 0x40000, 0x80000, 0x100000, 0x200000, 0x400000, 0x800000
- 0x1000000 (only CX6Dx)

### Hardware Pointers (HW_PTR)
After magic pattern found, reads hardware pointers from offset 0x18 (relative to image start):
- Boot2 pointer
- ITOC pointer  
- Tools pointer
- Image info pointer
- For FS4 extended: public key, forbidden versions (encrypted images)
- For FS5: additional pointers for hashes table, NCORE BCH

### Image Format Detection
- FS4: image_format_version = 1
- FS5: image_format_version = 2 (ConnectX-8 and newer)

### Common ITOC Sections
Present in both CX5 and CX6Dx:
- IMAGE_INFO (0x10) - Basic image metadata
- IMAGE_SIGNATURE_* - Security signatures
- PUBLIC_KEYS_* - RSA public keys
- Main firmware sections: IRON_PREP_CODE, MAIN_CODE, PCI_CODE, etc.
- Configuration: FW_MAIN_CFG, FW_BOOT_CFG, HW_MAIN_CFG, HW_BOOT_CFG
- PHY/PCIe code: PHY_UC_CODE, PCIE_PHY_UC_CODE, PHY_UC_CONSTS
- Debug: DBG_FW_INI, DBG_FW_PARAMS
- ROM_CODE - Boot ROM

### Security/Certificate Sections (CX6Dx only)
- ROOT_CERTIFICATES_1/2 - Root certificate storage
- CERTIFICATE_CHAINS_1/2 - Certificate chain storage
- DIGITAL_CERT_RW, DIGITAL_CACERT_RW - Writable certificate areas
- CERT_CHAIN_0, DIGITAL_CERT_PTR - Additional certificate management

### Size Differences
- CX5: ~17MB total image
- CX6Dx: ~33MB total image (nearly 2x)
- The increase is due to additional security features and certificate management

## Parsing Process Flow

### 1. Find Magic Pattern
```
for addr in [0x0, 0x10000, ..., 0x1000000]:
    if read_dword(addr) == "MT_FW__":
        image_start = addr
        break
```

### 2. Parse Hardware Pointers
- Read from image_start + 0x18
- Extract boot2_ptr, itoc_ptr, tools_ptr, image_info_ptr
- For FS5: Also extract hashes_table_ptr, ncore_bch_ptr

### 3. Verify Tools Area
- Check CRC at tools_ptr
- Validate tools area structure

### 4. Parse ITOC
- Read ITOC header at itoc_ptr (typically 0x5000)
- Verify header signature and CRC
- Parse ITOC entries to get section offsets

### 5. Parse DTOC  
- CX5: Read from 0x00fff000
- CX6Dx/BF2/CX7: Read from 0x01fff000
- Same structure as ITOC but for device-specific data

### Key Differences in Parsing
1. **Search range**: CX6Dx searches up to 0x1000000 vs 0x800000 for CX5
2. **DTOC location**: Different offset for newer devices
3. **Section count**: CX6Dx has significantly more sections
4. **Certificate support**: Extensive in CX6Dx, minimal in CX5

## TOC Entry Structure
Each TOC entry (32 bytes) contains:
- Type (2 bytes)
- Size (4 bytes)  
- Offset (8 bytes) - Contains flags in upper bits
- CRC (4 bytes)
- Additional metadata

## Important Constants
- FS4_BOOT_SIGNATURE: 0x4D544657 ("MTFW")
- ITOC_SIGNATURE: 0x49544f43 ("ITOC")
- CRC polynomial: 0x100b
- Default sector size: 0x1000 (4KB)

## Debug Output Format
When FW_COMPS_DEBUG=1, mstflint shows:
1. Function entry/exit traces
2. Magic pattern search results
3. Pointer values (boot2_ptr, itoc_ptr, etc.)
4. Section discovery and parsing
5. CRC verification status

## Implementation Notes
1. All multi-byte values are big-endian in firmware
2. CRC calculations use a specific polynomial (0x100b)
3. Section offsets in TOC entries are in dwords, must multiply by 4
4. Some sections have NOCRC flag (crc=1 means no CRC check)
5. Encrypted images have limited parsing capability

## Firmware Compatibility

### By Hardware ID
| HW ID | Device | Format | Notes |
|-------|--------|--------|-------|
| 0x20d | ConnectX-5 | FS4 | Base security features |
| 0x212 | ConnectX-6 Dx | FS4 | Enhanced security, certificates |
| 0x214 | BlueField-2 | FS4 | DPU features, same as CX6Dx |
| 0x218 | ConnectX-7 | FS4 | BOOT3_CODE, APU_KERNEL sections |
| 0x21c | BlueField-3 | FS4 | Encrypted firmware support |
| 0x21e | ConnectX-8 | FS5 | New format, NCORE BCH |

### By Chunk Size
- 2^23 (8MB): ConnectX-5
- 2^24 (16MB): ConnectX-6 Dx, BlueField-2, ConnectX-7, BlueField-3, ConnectX-8

## Detailed Section Types

### ITOC Sections
| ID | Name | Description | Devices |
|----|------|-------------|---------|
| 0x1 | BOOT2_CODE | Second stage bootloader | All |
| 0x2 | UPGRADE_CODE | Firmware upgrade handler | All |
| 0x4 | IRON_PREP_CODE | Hardware preparation | All |
| 0x5 | POST_IRON_BOOT_CODE | Post-boot initialization | All |
| 0x6 | MAIN_CODE | Main firmware code | All |
| 0x8 | PCIE_LINK_CODE | PCIe link management | All |
| 0x9 | PCI_CODE | PCI configuration | All |
| 0xa | ROM_CODE | Boot ROM (UEFI/FlexBoot) | All |
| 0xb | DBG_FW_INI | Debug initialization | All |
| 0xc | PHY_UC_CODE | PHY microcode | All |
| 0xd | PCIE_PHY_UC_CODE | PCIe PHY microcode | All |
| 0xe | PHY_UC_CONSTS | PHY constants | All |
| 0xf | BOOT3_CODE | Third stage bootloader | CX7+ |
| 0x10 | IMAGE_INFO | Image metadata | All |
| 0x11 | FW_BOOT_CFG | Boot configuration | All |
| 0x12 | FW_MAIN_CFG | Main configuration | All |
| 0x13 | CRDUMP_MASK_DATA | Crash dump masks | All |
| 0x14 | APU_KERNEL | APU kernel code | CX7+ |
| 0x18 | IMAGE_SIGNATURE_256 | 256-bit signature | All |
| 0x19 | IMAGE_SIGNATURE_512 | 512-bit signature | All |
| 0x1a | PUBLIC_KEYS_2048 | 2048-bit RSA keys | All |
| 0x1b | PUBLIC_KEYS_4096 | 4096-bit RSA keys | All |
| 0x1c | FORBIDDEN_VERSIONS | Blacklisted versions | All |
| 0x1d | RESET_INFO | Reset information | BF2+ |
| 0x20 | CRYPTO_NVSTORAGE | Crypto NV storage | CX6Dx+ |
| 0x21 | RSA_PUBLIC_KEY | RSA public key | CX6Dx+ |
| 0x22 | RSA_4096_SIGNATURES | 4096-bit signatures | CX6Dx+ |
| 0x23 | CCIR_INFRA_CODE | Crypto/compression infra | CX6Dx+ |
| 0x24 | CCIR_ALGO_CODE | Crypto/compression algos | CX6Dx+ |
| 0x25 | ACE_CODE | Acceleration engine | CX6Dx+ |
| 0x30 | DBG_FW_PARAMS | Debug parameters | All |
| 0x50 | HW_BOOT_CFG | Hardware boot config | All |
| 0x51 | HW_MAIN_CFG | Hardware main config | All |

### DTOC Sections
| ID | Name | Description | Devices |
|----|------|-------------|---------|
| 0xd5 | SECURITY_LOG | Security event log | All |
| 0xe0 | MFG_INFO | Manufacturing info | All |
| 0xe1 | DEV_INFO | Device information | All |
| 0xe3 | VPD_R0 | Vital Product Data | All |
| 0xe4 | NV_DATA | Non-volatile data 1 | All |
| 0xe5 | FW_NV_LOG | Firmware NV log | All |
| 0xe6 | NV_DATA | Non-volatile data 2 | All |
| 0xea | FW_INTERNAL_USAGE | Internal firmware data | All |
| 0xeb | PROGRAMMABLE_HW_FW | Programmable HW FW 1 | All |
| 0xec | PROGRAMMABLE_HW_FW | Programmable HW FW 2 | All |
| 0xed | DIGITAL_CERT_PTR | Certificate pointer | CX6Dx+ |
| 0xee | DIGITAL_CERT_RW | Digital certificate RW | CX6Dx+ |
| 0xf2 | CERT_CHAIN_0 | Certificate chain 0 | CX6Dx+ |
| 0xf3 | DIGITAL_CACERT_RW | CA certificate RW | CX6Dx+ |
| 0xf4 | CERTIFICATE_CHAINS_1 | Certificate chains 1 | CX6Dx+ |
| 0xf5 | CERTIFICATE_CHAINS_2 | Certificate chains 2 | CX6Dx+ |
| 0xf6 | ROOT_CERTIFICATES_1 | Root certificates 1 | CX6Dx+ |
| 0xf7 | ROOT_CERTIFICATES_2 | Root certificates 2 | CX6Dx+ |

## Parsing Algorithm

### Basic Flow
1. **Find Image Start**
   - Search for magic pattern "MT_FW__" (0x4D545F46575F5F00)
   - Try offsets: 0x0, 0x10000, 0x20000, ..., 0x1000000
   - First match becomes image_start

2. **Read Hardware Pointers**
   - FS4: Read from image_start + 0x18
   - FS5: Parse extended structure with CRC validation

3. **Determine Format**
   - Read image_format_version from IMAGE_INFO
   - 1 = FS4, 2 = FS5

4. **Check Encryption**
   - Try to read ITOC header
   - If signature invalid, image is encrypted

5. **Parse TOCs**
   - ITOC: Image sections at _itoc_ptr
   - DTOC: Device data at calculated offset

### FS4 Non-Encrypted Parsing
1. Magic pattern search → image_start
2. Parse HW pointers (boot2, itoc, tools, image_info)
3. Verify tools area CRC
4. Parse ITOC:
   - Read header, verify signature and CRC
   - Read entries, build section map
5. Parse DTOC:
   - Calculate offset based on device
   - Same parsing as ITOC

### FS4 Encrypted Firmware Parsing
1. Same magic pattern search and HW pointer initialization
2. **Check Encryption** (`IsEncryptedImage`)
   - Attempts to read ITOC header at _itoc_ptr
   - If invalid signature, firmware is encrypted
3. **Parse Encrypted Header**
   - IMAGE_INFO from _image_info_section_ptr
   - RSA keys from _public_key_ptr
4. **Skip ITOC parsing** (encrypted)
5. **Parse DTOC normally** (not encrypted)

### FS5 Firmware Parsing
1. Same magic pattern search (includes 0x2000000)
2. **FS5 HW Pointers** (`ParseHwPointers`)
   - Uses `fs5_image_layout_hw_pointers_gilboa` structure
   - Different structure with more fields (PSC, NCore, etc.)
3. Rest follows FS4 encrypted flow

## Vendor Firmware Analysis

### Vendor-Specific PSIDs
After analyzing various vendor firmware files, the following patterns were observed:

| Vendor | PSID Pattern | Example | Device Type | Notes |
|--------|--------------|---------|-------------|-------|
| NVIDIA | NVD0000000020 | 699140280000_Ax_NVD0000000020.bin | BlueField-2 | Standard NVIDIA firmware |
| OVH | OVH0000000001 | SSN4MELX100200_Ax_OVH0000000001.bin | BlueField-2 | OVH cloud provider custom |
| BD | BD_0000000005 | 900-9D3B6-F2SC-EA0_Ax_BD_0000000005.bin | BlueField-3 | Custom BD firmware |
| MT | MT_XXXXXXXXX | Various | Various | Standard Mellanox/NVIDIA format |

### Key Findings from Vendor Firmware
1. **No Parsing Deviations**: All vendor firmware files follow the standard FS4/FS5 parsing flow
2. **Same Magic Pattern**: All use the same magic pattern at offset 0x0
3. **Standard TOC Processing**: ITOC at 0x5000, DTOC locations match device generation
4. **Consistent CRC Algorithm**: Same CRC16 with polynomial 0x100b
5. **Device-Specific Features**: Features match the underlying hardware (e.g., BlueField vs ConnectX)

### Additional Device Types Found
| Device | HW ID | Chip Type | Description |
|--------|-------|-----------|-------------|
| BlueField G-Series | 0x211 | 10 | First-gen BlueField SmartNIC |
| ConnectX-6 Lx | 0x20f | 14 | ConnectX-6 Lx variant |

### Vendor Firmware Characteristics
- **Part Number Format**: Often includes vendor-specific prefixes (e.g., SSN4MELX for OVH)
- **Description Field**: May contain vendor-specific product names
- **ROM Info**: Standard across vendors, matches device capabilities
- **Security Attributes**: Consistent with device type (encrypted BlueField-3, etc.)
- **Update Method**: Always "fw_ctrl" for vendor firmware

### Implementation Notes for Vendor Firmware
1. No special handling required - standard parsing logic applies
2. PSID field helps identify vendor customizations but doesn't affect parsing
3. Part numbers and descriptions are metadata only - don't affect binary parsing
4. All vendor firmware follows the same structural format as reference firmware

## Documentation Verification Summary

After comprehensive analysis of multiple firmware files including vendor-specific variants, the following conclusions can be drawn:

### Parsing Consistency
1. **Universal Format Adherence**: All analyzed firmware files (NVIDIA reference, OVH, NVD, BD, etc.) follow the exact same FS4/FS5 parsing structure
2. **No Vendor Deviations**: Vendor customizations are limited to metadata (PSID, descriptions) and don't affect binary parsing
3. **Standard Magic Pattern**: All firmware starts with the same magic pattern at offset 0x0
4. **Consistent TOC Structure**: ITOC always at 0x5000, DTOC location based on device generation

### Key Implementation Requirements Confirmed
1. **Dynamic Section Offsets**: All section offsets must be read from TOC entries - no hardcoding
2. **CRC16 Algorithm**: Universal use of polynomial 0x100b across all firmware types
3. **Format Version Detection**: image_format_version field reliably indicates FS4 (1) vs FS5 (2)
4. **Encryption Handling**: Encrypted firmware detected via ITOC header check, limited parsing applies

### Device Coverage
The documentation now covers all major device families:
- ConnectX-5 through ConnectX-8
- BlueField G-Series, BlueField-2, and BlueField-3
- ConnectX-6 Lx variant
- Both encrypted and non-encrypted variants

### Vendor Firmware Findings
- All vendor firmware follows standard Mellanox/NVIDIA binary format
- Vendor-specific identifiers (PSID, part numbers) are metadata only
- No special parsing logic required for vendor firmware
- Security features match the underlying hardware platform

This documentation provides a complete reference for implementing FS4 and FS5 firmware parsing that will work with both reference and vendor firmware files.

## Data Structures for Firmware Parsing

### Endianness
**IMPORTANT**: All firmware data structures are stored in **Big Endian** format on disk/flash. The mstflint code uses dedicated pack/unpack functions that handle endianness conversion:
- `adb2c_push_integer_to_buff()` - Writes integers in Big Endian
- `adb2c_pop_integer_from_buff()` - Reads integers from Big Endian
- `adb2c_push_bits_to_buff()` - Writes bit fields in Big Endian
- `adb2c_pop_bits_from_buff()` - Reads bit fields from Big Endian

### Core Structures

#### 1. Hardware Pointer Entry (8 bytes)
```c
struct image_layout_hw_pointer_entry {
    u_int32_t ptr;    /* 0x0.0 - Pointer value */
    u_int32_t crc;    /* 0x4.0 - CRC of pointer */
};
```

#### 2. FS4 Hardware Pointers (Carmel) - Variable size
```c
struct image_layout_hw_pointers_carmel {
    struct image_layout_hw_pointer_entry boot2_ptr;              /* 0x0 */
    struct image_layout_hw_pointer_entry toc_ptr;                /* 0x8 */
    struct image_layout_hw_pointer_entry tools_ptr;              /* 0x10 */
    struct image_layout_hw_pointer_entry image_info_section_ptr; /* 0x18 */
    struct image_layout_hw_pointer_entry fw_public_key_ptr;      /* 0x20 */
    struct image_layout_hw_pointer_entry fw_signature_ptr;       /* 0x28 */
    struct image_layout_hw_pointer_entry public_key_ptr;         /* 0x30 */
    struct image_layout_hw_pointer_entry forbidden_versions_ptr; /* 0x38 */
    /* Additional entries may follow */
};
```

#### 3. FS5 Hardware Pointers (Gilboa) - 128 bytes fixed
```c
struct fs5_image_layout_hw_pointers_gilboa {
    struct image_layout_hw_pointer_entry boot2_ptr;               /* 0x0 */
    struct image_layout_hw_pointer_entry toc_ptr;                 /* 0x8 */
    struct image_layout_hw_pointer_entry tools_ptr;               /* 0x10 */
    struct image_layout_hw_pointer_entry image_info_section_ptr;  /* 0x18 */
    struct image_layout_hw_pointer_entry fw_public_key_ptr;       /* 0x20 */
    struct image_layout_hw_pointer_entry fw_signature_ptr;        /* 0x28 */
    struct image_layout_hw_pointer_entry public_key_ptr;          /* 0x30 */
    struct image_layout_hw_pointer_entry forbidden_versions_ptr;  /* 0x38 */
    struct image_layout_hw_pointer_entry psc_bl1_ptr;            /* 0x40 */
    struct image_layout_hw_pointer_entry psc_hashes_table_ptr;   /* 0x48 */
    struct image_layout_hw_pointer_entry ncore_hashes_pointer;   /* 0x50 */
    struct image_layout_hw_pointer_entry psc_fw_update_handle_ptr; /* 0x58 */
    struct image_layout_hw_pointer_entry psc_bch_pointer;        /* 0x60 */
    struct image_layout_hw_pointer_entry reserved_ptr13;         /* 0x68 */
    struct image_layout_hw_pointer_entry reserved_ptr14;         /* 0x70 */
    struct image_layout_hw_pointer_entry ncore_bch_pointer;      /* 0x78 */
};
```

#### 4. ITOC/DTOC Header (32 bytes)
```c
struct image_layout_itoc_header {
    u_int32_t signature0;      /* 0x0.0 - "ITOC" or "DTOC" */
    u_int32_t signature1;      /* 0x4.0 - 0 */
    u_int32_t signature2;      /* 0x8.0 - 0 */
    u_int32_t signature3;      /* 0xc.0 - 0 */
    u_int32_t version;         /* 0x10.0 - Format version */
    u_int32_t reserved;        /* 0x14.0 */
    u_int32_t itoc_entry_crc;  /* 0x18.0 - CRC of entries */
    u_int32_t crc;            /* 0x1c.0 - Header CRC */
};
```

#### 5. ITOC/DTOC Entry (32 bytes)
```c
struct image_layout_itoc_entry {
    /* 0x0.16 - 0x0.31 */ u_int16_t type;              /* Section type */
    /* 0x0.0 - 0x0.15 */  u_int16_t reserved0;
    /* 0x4.0 - 0x7.31 */  u_int32_t size;              /* Size in bytes */
    /* 0x8.0 - 0xf.31 */  u_int64_t offset_in_image;   /* Offset (includes flags in upper bits) */
    /* 0x10.0 - 0x13.31 */ u_int32_t crc;              /* Section CRC or NOCRC (1) */
    /* 0x14.0 - 0x17.31 */ u_int32_t reserved1;
    /* 0x18.0 - 0x1f.31 */ u_int64_t reserved2;
};
```

**Note**: The offset_in_image field contains both the offset and flags:
- Bits 0-31: Offset in dwords (multiply by 4 for byte offset)
- Bits 32-63: Flags and reserved bits

#### 6. Tools Area (64 bytes)
```c
struct image_layout_tools_area {
    u_int32_t image_layout_tools_area_Bytees;  /* 0x0.0 */
    u_int32_t image_layout_tools_area_mirror;  /* 0x4.0 */
    u_int32_t reserved0[14];                   /* 0x8.0 - 0x3c.31 */
};
```

#### 7. IMAGE_INFO Overview (1024 bytes total)
The IMAGE_INFO section contains multiple subsections:
- FW_VERSION (offset 0x10)
- FW_RELEASE_DATE (offset 0x20)
- MIC_VERSION (offset 0x40)
- PRS_NAME (offset 0x50)
- PART_NUMBER (offset 0xd0)
- DESCRIPTION (offset 0x170)
- BRANCH_VER (offset 0x370)
- PSID (offset 0x390)
- VSD (offset 0x3b0)
- PRODUCT_VER (offset 0x3f0)
- Additional metadata

### Implementation Notes

#### Bit-level Field Packing
ITOC entries use non-standard bit packing. For example:
```c
/* Type field is bits 16-31 of first dword */
type = (first_dword >> 16) & 0xFFFF;
```

#### CRC Calculation
- Software CRC: Standard CRC16 with polynomial 0x100b
- Hardware CRC: Special variant where first 2 bytes are inverted
- CRC location: Last 2-4 bytes of structure

#### Important Constants
```c
#define FS4_BOOT_SIGNATURE        0x4D544657  /* "MTFW" */
#define FS4_BOOT_SIGNATURE_SIZE   4
#define ITOC_SIGNATURE            0x49544f43  /* "ITOC" */
#define DTOC_SIGNATURE            0x44544f43  /* "DTOC" */
#define FS4_HW_PTR_START          0x18
#define FS4_DEFAULT_SECTOR_SIZE   0x1000      /* 4KB */
#define NOCRC                     1           /* CRC check disabled */
```

#### Field Alignment
- All structures are dword-aligned (4 bytes)
- Multi-byte fields are big-endian
- Bit fields within dwords follow big-endian bit ordering

#### Dynamic Structure Sizing
Some structures like HW pointers have variable size:
- Check for 0xFFFFFFFF terminators
- Use CRC validation to determine valid entries

#### Address/Size Conversions
TOC entry offsets are in dwords:
```c
byte_offset = toc_entry.offset * 4;
```

## Checksum Verification Documentation

### CRC Algorithms

#### 1. Software CRC16 (`CalcImageCRC`)
- **Polynomial**: 0x100b
- **Initial Value**: 0xffff
- **Final XOR**: 0xffff
- **Implementation**: `/mft_utils/crc16.cpp`
- **Algorithm**:
  ```c
  for each 32-bit word:
      for i = 0 to 31:
          if (crc & 0x8000):
              crc = ((crc << 1) | (word >> 31)) ^ 0x100b) & 0xffff
          else:
              crc = ((crc << 1) | (word >> 31)) & 0xffff
          word = (word << 1)
  // Final processing
  for i = 0 to 15:
      if (crc & 0x8000):
          crc = ((crc << 1) ^ 0x100b) & 0xffff
      else:
          crc = (crc << 1) & 0xffff
  crc = crc ^ 0xffff  // Final XOR
  ```

#### 2. Hardware CRC (`calc_hw_crc`)
- **Table-based CRC16**: Pre-computed table with 256 entries
- **Special Feature**: First 2 bytes of data are inverted (~data)
- **Implementation**: `/mft_utils/calc_hw_crc.c`
- **Algorithm**:
  ```c
  crc = 0xffff
  for i = 0 to size-1:
      data = (i > 1) ? d[i] : ~d[i]  // Invert first 2 bytes
      table_index = (crc ^ data) & 0xff
      crc = (crc >> 8) ^ crc16table2[table_index]
  crc = ((crc << 8) & 0xff00) | ((crc >> 8) & 0xff)  // Byte swap
  ```

### CRC Verification by Section Type

#### 1. Hardware Pointers
- **Location**: Image start + 0x18
- **CRC Type**: Dual approach (tries both software and hardware CRC)
- **FS4 Implementation**:
  ```c
  // Try software CRC first
  calcPtrSWCRC = CalcImageCRC(ptr_data, 1);
  if (ptrCRC == calcPtrSWCRC) {
      calcPtrCRC = calcPtrSWCRC;  // Some devices use SW CRC
  } else {
      calcPtrCRC = calc_hw_crc(ptr_data, 6);  // Hardware CRC
  }
  ```
- **FS5 Implementation**: Similar dual approach for compatibility

#### 2. ITOC/DTOC Headers
- **CRC Type**: Software CRC (`CalcImageCRC`)
- **Location**: Last 4 bytes of 32-byte header
- **Verification**:
  ```c
  header_crc_calc = CalcImageCRC(header_data, 7);  // First 28 bytes
  header_crc_stored = header_data[7];  // Last 4 bytes
  if (header_crc_calc != header_crc_stored) {
      return error("Bad CRC");
  }
  ```

#### 3. ITOC/DTOC Entries
- **CRC Type**: Software CRC (`CalcImageCRC`)
- **CRC Modes**:
  - **INITOCENTRY**: CRC within entry structure
  - **INSECTION**: CRC at end of section data
  - **NOCRC**: CRC = 1, verification skipped
- **Entry CRC**: Stored in ITOC header `itoc_entry_crc` field
- **Section CRC**: Stored in entry's `crc` field

#### 4. Boot2 Section
- **CRC Type**: Software CRC (`CalcImageCRC`)
- **Location**: End of Boot2 data
- **Special Note**: FS5 reports "CRC IGNORED" for Boot2

#### 5. Tools Area
- **CRC Type**: Software CRC (`CalcImageCRC`)  
- **Location**: Within tools area structure
- **Verification**: Standard CRC check

#### 6. Hashes Table (FS5 only)
- **Header CRC**:
  - Type: Software CRC
  - Location: Last 2 bytes of header
  - Size: 16-bit CRC
- **Table CRC**:
  - Type: Software CRC
  - Location: Last 2 bytes of table
  - Size: 16-bit CRC

#### 7. Image Sections (General)
- **CRC Type**: Software CRC (`CalcImageCRC`)
- **Location**: Determined by TOC entry `crc` field:
  - If `crc == NOCRC` (1): No CRC verification
  - If `crc == 0xFFFF`: Blank/uninitialized CRC
  - Otherwise: CRC location varies by section type

### CRC Error Handling

#### 1. Ignorable CRCs
- **Flag**: `_fwParams.ignoreCrcCheck`
- **Behavior**: Reports CRC mismatch but continues parsing
- **Use Case**: Development/debugging

#### 2. Critical CRCs
- **Headers**: Always critical (ITOC/DTOC headers)
- **Boot sections**: Critical unless ignore flag set
- **Security sections**: Always critical

#### 3. NOCRC Sections
- **Value**: `crc = 1` in TOC entry
- **Sections**: Typically configuration or non-critical data
- **Behavior**: CRC verification completely skipped

### Special Cases

#### 1. Encrypted Firmware
- **ITOC CRC**: Not verified (encrypted content)
- **DTOC CRC**: Normal verification (unencrypted)
- **Section CRCs**: Skipped for encrypted sections

#### 2. Blank CRC (0xFFFF)
- **Meaning**: Uninitialized/unprogrammed CRC
- **Behavior**: Treated as CRC mismatch
- **Common in**: Development firmware

#### 3. Boot Record CRC Masking
- **Function**: `recalcSectionCrc()`
- **Feature**: Can mask specific bytes before CRC calculation
- **Use Case**: Boot records with runtime-modified fields

### Implementation Notes

#### 1. Endianness Handling
```c
// CRC values are stored big-endian
TOCPUn(&crc_value, 1);  // Convert to CPU endianness
// After calculation, convert back
CPUTOn(&calc_crc, 1);   // Convert to big-endian
```

#### 2. CRC Calculation Size
- **Headers**: Size - 4 bytes (exclude CRC field)
- **Sections**: Full section size or size - 2/4 bytes
- **Entries**: Depends on structure definition

#### 3. Performance Optimization
- Hardware CRC uses lookup table (faster)
- Software CRC processes 32-bit words
- Cache CRC results when possible

### Summary Table

| Section Type | CRC Algorithm | CRC Location | Size | Special Notes |
|-------------|---------------|--------------|------|---------------|
| HW Pointers | HW/SW Dual | Entry + 4 | 16-bit | Try SW first, then HW |
| ITOC Header | Software | Header + 28 | 32-bit | Always verified |
| DTOC Header | Software | Header + 28 | 32-bit | Always verified |
| TOC Entries | Software | In header | 32-bit | All entries together |
| Boot2 | Software | End of data | 16-bit | May be ignored in FS5 |
| Tools Area | Software | In structure | 16-bit | Standard check |
| Hashes Table | Software | End of data | 16-bit | Header and data separate |
| Sections | Software | Per TOC entry | 16/32-bit | Unless NOCRC flag |

This comprehensive checksum verification system ensures firmware integrity while providing flexibility for development and special use cases.