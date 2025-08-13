# Flash Programming Sequences - Detailed Implementation

## Overview
This document provides the exact flash programming sequences used by mstflint, including timing, alignment requirements, and error handling.

## Flash Layout

### Memory Map
```
┌─────────────────┐ 0x000000
│   Boot Record   │ (4KB)
├─────────────────┤ 0x001000
│   Primary Image │ 
│   (Variable)    │
├─────────────────┤ 0x800000 (8MB mark)
│ Secondary Image │
│   (Variable)    │
├─────────────────┤ 0xF00000
│   NVDATA/VPD    │
│   (Reserved)    │
└─────────────────┘ 0xFFFFFF (16MB)
```

### Sector Organization
- Standard sector size: 4KB (0x1000)
- Large erase block: 64KB (0x10000)
- Write block size: 256 bytes (0x100)
- All addresses must be aligned to operation size

## Programming Sequences

### 1. Flash Initialization
```c
int init_flash_access(device_t *dev) {
    struct mfpa_reg mfpa = {0};
    
    // Query flash parameters
    mfpa.fs = 0;  // Flash select 0
    int rc = exec_access_reg(dev, REG_MFPA, 0, &mfpa, sizeof(mfpa));
    if (rc != 0) {
        return rc;
    }
    
    // Store flash info
    dev->flash.jedec_id = mfpa.jedec_id;
    dev->flash.size = mfpa.flash_size;
    dev->flash.sector_size = mfpa.sector_size;
    dev->flash.block_align = 1 << mfpa.block_alignment;
    dev->flash.capabilities = mfpa.capability_mask;
    
    // Verify expected flash type
    if ((mfpa.jedec_id & 0xFFFFFF) != 0x1840EF) {  // W25QxxBV
        log_warn("Unexpected flash type: 0x%06x", mfpa.jedec_id);
    }
    
    return 0;
}
```

### 2. Sector Erase Sequence
```c
int erase_flash_sector(device_t *dev, uint32_t addr, uint8_t erase_size) {
    struct mfbe_reg mfbe = {0};
    
    // Validate alignment
    uint32_t align_mask = (erase_size == MFBE_ERASE_4K) ? 0xFFF : 0xFFFF;
    if (addr & align_mask) {
        return -EINVAL;  // Not aligned
    }
    
    // Prepare erase command
    mfbe.fs = 0;
    mfbe.address = addr;
    mfbe.erase_size = erase_size;
    
    // Send erase command
    int rc = exec_access_reg(dev, REG_MFBE, 1, &mfbe, sizeof(mfbe));
    if (rc != 0) {
        return rc;
    }
    
    // Wait for erase completion (flash busy)
    // Typical times: 4KB=45ms, 64KB=150ms
    usleep(erase_size == MFBE_ERASE_4K ? 50000 : 200000);
    
    return 0;
}

// Erase with retry logic
int erase_sector_with_retry(device_t *dev, uint32_t addr) {
    int retry;
    
    for (retry = 0; retry < 3; retry++) {
        int rc = erase_flash_sector(dev, addr, MFBE_ERASE_4K);
        if (rc == 0) {
            return 0;
        }
        
        log_warn("Erase failed at 0x%06x, retry %d", addr, retry + 1);
        usleep(10000);  // 10ms between retries
    }
    
    return -EIO;
}
```

### 3. Flash Write Sequence
```c
int write_flash_block(device_t *dev, uint32_t addr, 
                      uint8_t *data, size_t size) {
    struct mfba_reg mfba = {0};
    
    // Validate parameters
    if (size > 256 || size == 0) {
        return -EINVAL;
    }
    
    // Must be aligned to block size
    if (addr & (dev->flash.block_align - 1)) {
        return -EINVAL;
    }
    
    // Prepare write command
    mfba.fs = 0;
    mfba.address = addr;
    mfba.size = size;
    mfba.access_mode = 1;  // Write mode
    memcpy(mfba.data, data, size);
    
    // Send write command
    int rc = exec_access_reg(dev, REG_MFBA, 1, &mfba, 
                            sizeof(mfba) - sizeof(mfba.data) + size);
    if (rc != 0) {
        return rc;
    }
    
    // Write time: ~1ms per 256 bytes
    usleep(2000);
    
    return 0;
}
```

### 4. Flash Read Sequence (for verification)
```c
int read_flash_block(device_t *dev, uint32_t addr, 
                     uint8_t *data, size_t size) {
    struct mfba_reg mfba = {0};
    
    if (size > 256) {
        return -EINVAL;
    }
    
    // Prepare read command
    mfba.fs = 0;
    mfba.address = addr;
    mfba.size = size;
    mfba.access_mode = 0;  // Read mode
    
    // Send read command
    int rc = exec_access_reg(dev, REG_MFBA, 0, &mfba, sizeof(mfba));
    if (rc != 0) {
        return rc;
    }
    
    // Copy read data
    memcpy(data, mfba.data, size);
    
    return 0;
}
```

### 5. Complete Sector Programming
```c
int program_flash_sector(device_t *dev, uint32_t sector_addr,
                        uint8_t *data, progress_cb_t progress_cb) {
    uint32_t offset;
    int rc;
    
    // Erase sector first
    rc = erase_sector_with_retry(dev, sector_addr);
    if (rc != 0) {
        return rc;
    }
    
    // Write in 256-byte blocks
    for (offset = 0; offset < dev->flash.sector_size; offset += 256) {
        size_t block_size = MIN(256, dev->flash.sector_size - offset);
        
        rc = write_flash_block(dev, sector_addr + offset,
                              data + offset, block_size);
        if (rc != 0) {
            log_error("Write failed at 0x%06x", sector_addr + offset);
            return rc;
        }
        
        // Update progress
        if (progress_cb) {
            progress_cb(sector_addr + offset, block_size);
        }
    }
    
    // Verify if requested
    if (!dev->params.no_verify) {
        uint8_t verify_buf[256];
        
        for (offset = 0; offset < dev->flash.sector_size; offset += 256) {
            size_t block_size = MIN(256, dev->flash.sector_size - offset);
            
            rc = read_flash_block(dev, sector_addr + offset,
                                 verify_buf, block_size);
            if (rc != 0) {
                return rc;
            }
            
            if (memcmp(data + offset, verify_buf, block_size) != 0) {
                log_error("Verify mismatch at 0x%06x", 
                         sector_addr + offset);
                return -EIO;
            }
        }
    }
    
    return 0;
}
```

## Failsafe Burn Sequence

### Phase 1: Burn Secondary Image
```c
int burn_secondary_image(device_t *dev, firmware_image_t *fw) {
    uint32_t secondary_addr = SECONDARY_IMAGE_ADDR;
    uint32_t addr, end_addr;
    int rc;
    
    // Calculate range
    end_addr = secondary_addr + fw->size;
    
    // Erase all affected sectors
    log_info("Erasing secondary image area...");
    for (addr = secondary_addr; addr < end_addr; addr += 0x1000) {
        rc = erase_sector_with_retry(dev, addr);
        if (rc != 0) {
            return rc;
        }
        
        update_progress(PROG_ERASING, addr - secondary_addr, fw->size);
    }
    
    // Write image data
    log_info("Writing secondary image...");
    for (addr = secondary_addr; addr < end_addr; addr += 0x1000) {
        uint32_t offset = addr - secondary_addr;
        
        rc = program_flash_sector(dev, addr, 
                                 fw->data + offset, NULL);
        if (rc != 0) {
            return rc;
        }
        
        update_progress(PROG_WRITING, offset, fw->size);
    }
    
    return 0;
}
```

### Phase 2: Update Boot Record
```c
int update_boot_record(device_t *dev, uint32_t boot_addr) {
    uint8_t boot_record[4096];
    struct boot_record_header *hdr;
    int rc;
    
    // Read current boot record
    for (int i = 0; i < 16; i++) {
        rc = read_flash_block(dev, i * 256, 
                             boot_record + i * 256, 256);
        if (rc != 0) {
            return rc;
        }
    }
    
    // Modify boot address
    hdr = (struct boot_record_header *)boot_record;
    hdr->boot_address = boot_addr;
    hdr->boot_address_checksum = calc_checksum(&boot_addr, 4);
    
    // Write back boot record
    rc = program_flash_sector(dev, 0, boot_record, NULL);
    
    return rc;
}
```

### Phase 3: Burn Primary Image
```c
int burn_primary_image(device_t *dev, firmware_image_t *fw) {
    // Similar to burn_secondary_image but at PRIMARY_IMAGE_ADDR
    // ...
}
```

## Component-Based Update (MCC)

### MCC Update Sequence
```c
int burn_via_mcc(device_t *dev, uint32_t component_id, 
                 uint8_t *data, uint32_t size) {
    struct mcc_reg mcc = {0};
    struct mcda_reg mcda = {0};
    uint32_t offset;
    int rc;
    
    // Step 1: Lock component
    mcc.command = MCC_CMD_LOCK_UPDATE;
    mcc.component_index = component_id;
    mcc.component_size = size;
    
    rc = exec_access_reg(dev, REG_MCC, 1, &mcc, sizeof(mcc));
    if (rc != 0) {
        return rc;
    }
    
    // Step 2: Transfer data in chunks
    for (offset = 0; offset < size; offset += 1024) {
        uint32_t chunk_size = MIN(1024, size - offset);
        
        // Send data chunk
        mcda.offset = offset;
        mcda.size = chunk_size;
        memcpy(mcda.data, data + offset, chunk_size);
        
        rc = exec_access_reg(dev, REG_MCDA, 1, &mcda,
                            8 + chunk_size);  // header + data
        if (rc != 0) {
            goto error;
        }
        
        // Trigger update
        mcc.command = MCC_CMD_UPDATE;
        mcc.offset = offset;
        mcc.data_size = chunk_size;
        
        rc = exec_access_reg(dev, REG_MCC, 1, &mcc, sizeof(mcc));
        if (rc != 0) {
            goto error;
        }
        
        // Update progress
        update_progress(PROG_WRITING, offset, size);
        
        // MCC requires polling for completion
        rc = poll_mcc_status(dev);
        if (rc != 0) {
            goto error;
        }
    }
    
    // Step 3: Verify
    mcc.command = MCC_CMD_VERIFY;
    rc = exec_access_reg(dev, REG_MCC, 1, &mcc, sizeof(mcc));
    if (rc != 0) {
        goto error;
    }
    
    // Step 4: Activate
    mcc.command = MCC_CMD_ACTIVATE;
    rc = exec_access_reg(dev, REG_MCC, 1, &mcc, sizeof(mcc));
    
    return rc;
    
error:
    // Cancel update on error
    mcc.command = MCC_CMD_CANCEL;
    exec_access_reg(dev, REG_MCC, 1, &mcc, sizeof(mcc));
    return rc;
}
```

## Cache Replacement Handling

### OCR (Override Cache Replacement)
```c
int handle_cache_replacement(device_t *dev) {
    if (!dev->params.ocr && !dev->is_livefish) {
        return 0;  // Not needed
    }
    
    log_warn("Firmware flash cache access enabled. "
             "Device may hang if interrupted.");
    
    // Perform dummy write to trigger cache flush
    // Address from mstflint: 0x1c748
    uint32_t dummy = 0;
    vsc_write(dev, CR_SPACE, 0x1c748, dummy);
    
    // Wait for cache to flush
    usleep(100000);  // 100ms
    
    return 0;
}
```

## Progress Tracking

```c
void update_progress(prog_stage_t stage, uint32_t current, uint32_t total) {
    static uint32_t last_percent = -1;
    uint32_t percent = (current * 100) / total;
    
    if (percent != last_percent) {
        last_percent = percent;
        
        const char *stage_str[] = {
            [PROG_ERASING] = "Erasing",
            [PROG_WRITING] = "Writing",
            [PROG_VERIFYING] = "Verifying"
        };
        
        printf("\r%s: [", stage_str[stage]);
        
        // Progress bar
        int bars = percent / 2;  // 50 chars total
        for (int i = 0; i < 50; i++) {
            putchar(i < bars ? '=' : ' ');
        }
        
        printf("] %3d%%", percent);
        fflush(stdout);
        
        if (percent == 100) {
            printf("\n");
        }
    }
}
```

## Error Recovery

### Flash Operation Error Codes
```c
enum flash_error {
    FLASH_OK = 0,
    FLASH_TIMEOUT = -1,
    FLASH_ERASE_FAIL = -2,
    FLASH_WRITE_FAIL = -3,
    FLASH_VERIFY_FAIL = -4,
    FLASH_PROTECTED = -5,
    FLASH_ALIGN_ERROR = -6,
    FLASH_SIZE_ERROR = -7
};

const char *flash_error_string(int err) {
    switch (err) {
    case FLASH_TIMEOUT:     return "Flash operation timeout";
    case FLASH_ERASE_FAIL:  return "Flash erase failed";
    case FLASH_WRITE_FAIL:  return "Flash write failed";
    case FLASH_VERIFY_FAIL: return "Flash verify failed";
    case FLASH_PROTECTED:   return "Flash area protected";
    case FLASH_ALIGN_ERROR: return "Flash alignment error";
    case FLASH_SIZE_ERROR:  return "Flash size error";
    default:                return "Unknown flash error";
    }
}
```

This document provides all the flash programming sequences needed for implementation, including exact timing, alignment requirements, and error handling.