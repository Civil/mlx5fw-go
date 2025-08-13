# Firmware Image Format and Parsing

## Image Format Overview

Mellanox firmware images use the FS4 (Fourth generation Flash Structure) format, which consists of:
1. Magic pattern for image detection
2. Image header with metadata
3. ITOC (Image Table of Contents) - code and configuration
4. DTOC (Data Table of Contents) - device-specific data
5. Binary sections referenced by TOC entries

## Image Detection and Validation

### 1. Finding Image Start
```c
#define MAGIC_PATTERN 0x4D544657  // "MTFW" in big-endian

int find_image_start(uint8_t *data, size_t size, uint32_t *offsets, int max) {
    int count = 0;
    
    // Search at standard offsets
    uint32_t search_offsets[] = {
        0x0,        // Start of file
        0x10000,    // 64KB
        0x20000,    // 128KB
        0x40000,    // 256KB
        0x80000,    // 512KB
        0x100000,   // 1MB
        0x200000,   // 2MB
        0x400000,   // 4MB
        0x800000    // 8MB
    };
    
    for (int i = 0; i < ARRAY_SIZE(search_offsets); i++) {
        if (search_offsets[i] >= size) {
            break;
        }
        
        uint32_t magic = be32_to_cpu(*(uint32_t*)(data + search_offsets[i]));
        if (magic == MAGIC_PATTERN) {
            offsets[count++] = search_offsets[i];
            if (count >= max) {
                break;
            }
        }
    }
    
    return count;
}
```

### 2. Image Header Parsing
```c
typedef struct {
    // FS4 Image Header Layout (from image start)
    uint32_t magic_pattern;      // 0x00: Must be 0x4D544657
    uint32_t reserved1[3];       // 0x04-0x0F
    uint32_t image_size;         // 0x10: Total image size
    uint32_t reserved2[3];       // 0x14-0x1F
    uint32_t itoc_offset;        // 0x20: Offset to ITOC
    uint32_t reserved3[7];       // 0x24-0x3F
    uint32_t boot_version;       // 0x40: Boot version
    uint32_t reserved4[15];      // 0x44-0x7F
    uint32_t dtoc_offset;        // 0x80: Offset to DTOC
    uint32_t reserved5[31];      // 0x84-0xFF
} __attribute__((packed)) fs4_header_raw_t;

int parse_image_header(uint8_t *data, size_t size, 
                      fs4_image_header_t *header) {
    if (size < sizeof(fs4_header_raw_t)) {
        return -EINVAL;
    }
    
    fs4_header_raw_t *raw = (fs4_header_raw_t *)data;
    
    // Convert from big-endian
    header->magic_pattern = be32_to_cpu(raw->magic_pattern);
    header->image_size = be32_to_cpu(raw->image_size);
    header->itoc_offset = be32_to_cpu(raw->itoc_offset);
    header->dtoc_offset = be32_to_cpu(raw->dtoc_offset);
    
    // Validate
    if (header->magic_pattern != MAGIC_PATTERN) {
        return -EINVAL;
    }
    
    if (header->image_size > size) {
        log_error("Image size %u exceeds file size %zu", 
                  header->image_size, size);
        return -EINVAL;
    }
    
    return 0;
}
```

## TOC (Table of Contents) Parsing

### 1. TOC Header Structure
```c
int parse_toc_header(uint8_t *data, size_t offset, 
                    toc_header_t *toc, bool is_itoc) {
    if (offset + sizeof(toc_header_t) > size) {
        return -EINVAL;
    }
    
    toc_header_t *raw = (toc_header_t *)(data + offset);
    
    // Parse header
    toc->signature = be32_to_cpu(raw->signature);
    toc->version = be32_to_cpu(raw->version);
    toc->num_entries = be32_to_cpu(raw->num_entries);
    toc->entry_size = be32_to_cpu(raw->entry_size);
    
    // Validate signature
    uint32_t expected_sig = is_itoc ? 0x4D544F43 : 0x44544F43;
    if (toc->signature != expected_sig) {
        log_error("Invalid %s signature: 0x%08x", 
                  is_itoc ? "ITOC" : "DTOC", toc->signature);
        return -EINVAL;
    }
    
    // Validate entry count
    if (toc->num_entries > 256) {  // Sanity check
        log_error("Too many TOC entries: %u", toc->num_entries);
        return -EINVAL;
    }
    
    return 0;
}
```

### 2. TOC Entry Parsing
```c
int parse_toc_entries(uint8_t *data, size_t toc_offset, 
                     toc_header_t *toc, toc_entry_t **entries) {
    size_t entries_offset = toc_offset + sizeof(toc_header_t);
    size_t entries_size = toc->num_entries * sizeof(toc_entry_t);
    
    *entries = malloc(entries_size);
    if (!*entries) {
        return -ENOMEM;
    }
    
    // Copy and convert entries
    for (int i = 0; i < toc->num_entries; i++) {
        toc_entry_t *raw = (toc_entry_t *)(data + entries_offset + 
                                          i * sizeof(toc_entry_t));
        toc_entry_t *entry = &(*entries)[i];
        
        entry->type = raw->type;
        entry->flags = raw->flags;
        entry->offset = be32_to_cpu(raw->offset);
        entry->size = be32_to_cpu(raw->size);
        entry->crc = be32_to_cpu(raw->crc);
        memcpy(entry->name, raw->name, sizeof(entry->name));
        
        // Validate entry
        if (entry->offset + entry->size > image_size) {
            log_error("TOC entry %d exceeds image bounds", i);
            free(*entries);
            return -EINVAL;
        }
    }
    
    return 0;
}
```

## Section Parsing

### 1. IMAGE_INFO Section
```c
int parse_image_info(uint8_t *data, toc_entry_t *entry, 
                    image_info_t *info) {
    if (entry->type != SECTION_IMAGE_INFO) {
        return -EINVAL;
    }
    
    // IMAGE_INFO raw format
    struct image_info_raw {
        uint32_t info_size;          // 0x00
        uint32_t info_version;       // 0x04
        uint8_t  psid[16];          // 0x08
        uint32_t fw_version[3];      // 0x18: major.minor.subminor
        uint32_t fw_build_time;      // 0x24
        uint32_t fw_build_date;      // 0x28
        uint32_t supported_hw_id[8]; // 0x2C
        uint32_t mic_version;        // 0x4C
        uint32_t pci_device_id;      // 0x50
        uint32_t reserved[32];       // 0x54
        uint8_t  vsd[208];          // 0xD4: Vendor Specific Data
        uint32_t security_version;   // 0x1A4
        // ... more fields ...
    } __attribute__((packed));
    
    struct image_info_raw *raw = (struct image_info_raw *)(data + entry->offset);
    
    // Verify size
    if (be32_to_cpu(raw->info_size) != sizeof(*raw)) {
        log_error("Unexpected IMAGE_INFO size: %u", 
                  be32_to_cpu(raw->info_size));
        return -EINVAL;
    }
    
    // Parse fields
    memcpy(info->psid, raw->psid, sizeof(info->psid));
    for (int i = 0; i < 3; i++) {
        info->fw_version[i] = be32_to_cpu(raw->fw_version[i]);
    }
    info->fw_rel_date = be32_to_cpu(raw->fw_build_date);
    info->fw_rel_time = be32_to_cpu(raw->fw_build_time);
    
    // Parse supported hardware IDs
    for (int i = 0; i < 8; i++) {
        info->supported_hw_id[i] = be32_to_cpu(raw->supported_hw_id[i]);
        if (info->supported_hw_id[i] == 0) {
            break;  // End of list
        }
    }
    
    // Parse VSD
    memcpy(info->vsd, raw->vsd, sizeof(info->vsd));
    
    // Security info
    info->security_version = be32_to_cpu(raw->security_version);
    
    return 0;
}
```

### 2. DEV_INFO Section (Device-specific data)
```c
typedef struct dev_info {
    uint32_t signature;          // "DEVI"
    uint32_t version;
    uint32_t size;
    
    // GUIDs and MACs
    uint64_t guids[4];
    uint64_t macs[4];
    
    // VPD data
    uint8_t  vpd_data[256];
    
} __attribute__((packed)) dev_info_t;

int parse_dev_info(uint8_t *data, toc_entry_t *entry, 
                  dev_info_t *dev_info) {
    if (entry->type != SECTION_DEV_INFO) {
        return -EINVAL;
    }
    
    dev_info_t *raw = (dev_info_t *)(data + entry->offset);
    
    // Verify signature
    if (be32_to_cpu(raw->signature) != 0x44455649) { // "DEVI"
        return -EINVAL;
    }
    
    // Parse GUIDs and MACs
    for (int i = 0; i < 4; i++) {
        dev_info->guids[i] = be64_to_cpu(raw->guids[i]);
        dev_info->macs[i] = be64_to_cpu(raw->macs[i]);
    }
    
    // Copy VPD
    memcpy(dev_info->vpd_data, raw->vpd_data, sizeof(dev_info->vpd_data));
    
    return 0;
}
```

### 3. Boot Code Sections
```c
int validate_boot_section(uint8_t *data, toc_entry_t *entry) {
    // Boot sections have special validation
    if (entry->type != SECTION_BOOT_CODE &&
        entry->type != SECTION_PCI_CODE) {
        return 0;  // Not a boot section
    }
    
    // Check boot signature at end of section
    uint32_t *boot_sig = (uint32_t *)(data + entry->offset + 
                                      entry->size - 4);
    if (be32_to_cpu(*boot_sig) != 0xBEEFCAFE) {
        log_error("Invalid boot signature in section type %d", 
                  entry->type);
        return -EINVAL;
    }
    
    // Verify CRC
    uint32_t calc_crc = crc32(data + entry->offset, entry->size - 4);
    if (calc_crc != entry->crc) {
        log_error("CRC mismatch in boot section: calc=0x%08x, expected=0x%08x",
                  calc_crc, entry->crc);
        return -EINVAL;
    }
    
    return 0;
}
```

## Complete Image Parsing Flow

```c
int parse_firmware_image(uint8_t *data, size_t size, 
                        firmware_image_t *fw) {
    int rc;
    
    // Step 1: Find image start
    uint32_t img_offsets[4];
    int num_images = find_image_start(data, size, img_offsets, 4);
    if (num_images == 0) {
        log_error("No valid firmware image found");
        return -EINVAL;
    }
    
    // Use first image
    fw->img_start = img_offsets[0];
    
    // Step 2: Parse header
    rc = parse_image_header(data + fw->img_start, 
                           size - fw->img_start, &fw->header);
    if (rc != 0) {
        return rc;
    }
    
    // Step 3: Parse ITOC
    rc = parse_toc_header(data, fw->img_start + fw->header.itoc_offset,
                         &fw->itoc, true);
    if (rc != 0) {
        return rc;
    }
    
    rc = parse_toc_entries(data, fw->img_start + fw->header.itoc_offset,
                          &fw->itoc, &fw->itoc_entries);
    if (rc != 0) {
        return rc;
    }
    
    // Step 4: Parse DTOC
    rc = parse_toc_header(data, fw->img_start + fw->header.dtoc_offset,
                         &fw->dtoc, false);
    if (rc != 0) {
        return rc;
    }
    
    rc = parse_toc_entries(data, fw->img_start + fw->header.dtoc_offset,
                          &fw->dtoc, &fw->dtoc_entries);
    if (rc != 0) {
        return rc;
    }
    
    // Step 5: Find and parse IMAGE_INFO
    toc_entry_t *info_entry = find_toc_entry(fw->itoc_entries, 
                                            fw->itoc.num_entries,
                                            SECTION_IMAGE_INFO);
    if (!info_entry) {
        log_error("IMAGE_INFO section not found");
        return -EINVAL;
    }
    
    rc = parse_image_info(data, info_entry, &fw->info);
    if (rc != 0) {
        return rc;
    }
    
    // Step 6: Validate all sections
    for (int i = 0; i < fw->itoc.num_entries; i++) {
        rc = validate_section(data, &fw->itoc_entries[i]);
        if (rc != 0) {
            log_error("Section %d validation failed", i);
            return rc;
        }
    }
    
    // Step 7: Check security
    rc = check_image_security(fw);
    if (rc != 0) {
        return rc;
    }
    
    log_info("Image parsed successfully:");
    log_info("  Version: %d.%d.%d", 
             fw->info.fw_version[0],
             fw->info.fw_version[1], 
             fw->info.fw_version[2]);
    log_info("  PSID: %.16s", fw->info.psid);
    log_info("  Size: %u bytes", fw->header.image_size);
    log_info("  ITOC entries: %u", fw->itoc.num_entries);
    log_info("  DTOC entries: %u", fw->dtoc.num_entries);
    
    return 0;
}
```

## Image Verification

### 1. Checksum Verification
```c
int verify_image_checksums(firmware_image_t *fw, uint8_t *data) {
    // Verify each section CRC
    for (int i = 0; i < fw->itoc.num_entries; i++) {
        toc_entry_t *entry = &fw->itoc_entries[i];
        
        if (entry->crc != 0) {  // CRC is optional
            uint32_t calc_crc = crc32(data + entry->offset, entry->size);
            if (calc_crc != entry->crc) {
                log_error("CRC mismatch in section %d (%s)",
                         i, section_type_to_string(entry->type));
                return -EINVAL;
            }
        }
    }
    
    return 0;
}
```

### 2. Compatibility Check
```c
int check_image_compatibility(firmware_image_t *fw, device_info_t *dev) {
    // Check hardware ID compatibility
    bool hw_match = false;
    for (int i = 0; i < 8 && fw->info.supported_hw_id[i]; i++) {
        if (fw->info.supported_hw_id[i] == dev->hw_id) {
            hw_match = true;
            break;
        }
    }
    
    if (!hw_match) {
        log_error("Hardware ID 0x%x not supported by image", dev->hw_id);
        return -EINVAL;
    }
    
    // Check PSID match (unless overridden)
    if (!params.allow_psid_change) {
        if (memcmp(fw->info.psid, dev->psid, 16) != 0) {
            log_error("PSID mismatch: image=%.16s, device=%.16s",
                     fw->info.psid, dev->psid);
            return MLX_IMG_PSID_MISMATCH;
        }
    }
    
    // Check version downgrade
    if (!params.force) {
        if (compare_versions(fw->info.fw_version, dev->fw_version) < 0) {
            log_error("Firmware downgrade not allowed without --force");
            return MLX_IMG_VERSION_ERR;
        }
    }
    
    return 0;
}
```

## Binary Layout Example

```
Offset      Content
----------  --------------------------------------------------
0x000000    MTFW magic (0x4D544657)
0x000010    Image size
0x000020    ITOC offset (e.g., 0x000100)
0x000080    DTOC offset (e.g., 0x008000)
...
0x000100    ITOC header (MTOC signature)
0x000110    ITOC entry 0: BOOT_CODE
0x000130    ITOC entry 1: PCI_CODE
0x000150    ITOC entry 2: MAIN_CODE
...
0x001000    BOOT_CODE section data
0x003000    PCI_CODE section data
0x010000    MAIN_CODE section data
...
0x008000    DTOC header (DTOC signature)
0x008010    DTOC entry 0: DEV_INFO
0x008030    DTOC entry 1: MFG_INFO
...
0x00A000    DEV_INFO section data
0x00A800    MFG_INFO section data
...
```

This comprehensive guide covers all aspects of parsing and validating Mellanox firmware images.