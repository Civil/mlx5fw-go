# Error Codes and Recovery Procedures

## Error Code Reference

### System-Level Error Codes
```c
enum mstflint_error {
    // Success
    MLX_OK                      = 0,
    
    // Device errors (1-20)
    MLX_DEV_NOT_FOUND          = 1,
    MLX_DEV_BUSY               = 2,
    MLX_DEV_PERMISSION_DENIED  = 3,
    MLX_DEV_NOT_SUPPORTED      = 4,
    MLX_DEV_COMMUNICATION_ERR  = 5,
    MLX_DEV_TIMEOUT            = 6,
    MLX_DEV_BAD_STATE          = 7,
    MLX_DEV_IN_USE             = 8,
    
    // Flash errors (21-40)
    MLX_FLASH_NOT_DETECTED     = 21,
    MLX_FLASH_ERASE_FAILED     = 22,
    MLX_FLASH_WRITE_FAILED     = 23,
    MLX_FLASH_READ_FAILED      = 24,
    MLX_FLASH_VERIFY_FAILED    = 25,
    MLX_FLASH_PROTECTED        = 26,
    MLX_FLASH_SIZE_MISMATCH    = 27,
    MLX_FLASH_ALIGN_ERROR      = 28,
    
    // Image errors (41-60)
    MLX_IMG_NOT_FOUND          = 41,
    MLX_IMG_CORRUPTED          = 42,
    MLX_IMG_BAD_MAGIC          = 43,
    MLX_IMG_BAD_CHECKSUM       = 44,
    MLX_IMG_INCOMPATIBLE       = 45,
    MLX_IMG_PSID_MISMATCH      = 46,
    MLX_IMG_VERSION_ERR        = 47,
    MLX_IMG_SECURITY_ERR       = 48,
    MLX_IMG_SIZE_EXCEEDED      = 49,
    
    // Register/ICMD errors (61-80)
    MLX_REG_ACCESS_FAILED      = 61,
    MLX_REG_NOT_SUPPORTED      = 62,
    MLX_REG_TIMEOUT            = 63,
    MLX_REG_BAD_PARAM          = 64,
    MLX_REG_SEMAPHORE_TO       = 65,
    MLX_REG_CMD_FAILED         = 66,
    MLX_REG_MAILBOX_FULL       = 67,
    
    // Component errors (81-100)
    MLX_COMP_NOT_FOUND         = 81,
    MLX_COMP_UPDATE_FAILED     = 82,
    MLX_COMP_LOCKED            = 83,
    MLX_COMP_VERIFY_FAILED     = 84,
    MLX_COMP_ACTIVATE_FAILED   = 85,
    
    // User errors (101-120)
    MLX_USER_ABORT             = 101,
    MLX_USER_INPUT_ERR         = 102,
    MLX_USER_PSID_MISMATCH     = 103,
    MLX_USER_VERSION_MISMATCH  = 104,
    
    // System errors (121+)
    MLX_MEMORY_ERROR           = 121,
    MLX_INTERNAL_ERROR         = 122,
    MLX_NOT_IMPLEMENTED        = 123,
    MLX_UNKNOWN_ERROR          = 124
};
```

### ICMD Status Codes
```c
enum icmd_status {
    ICMD_OK                    = 0x00,
    ICMD_INVALID_OPCODE        = 0x01,
    ICMD_INVALID_OP_MOD        = 0x02,
    ICMD_INVALID_PARAMETER     = 0x03,
    ICMD_RESOURCE_BUSY         = 0x04,
    ICMD_EXCEEDED_LIMIT        = 0x05,
    ICMD_ACCESS_DENIED         = 0x06,
    ICMD_HW_ERROR              = 0x07,
    ICMD_TIMEOUT               = 0x08,
    ICMD_BAD_SYSTEM_STATE      = 0x09,
    ICMD_VERSION_NOT_SUPPORTED = 0x0A,
    ICMD_UNKNOWN_TLV           = 0x0B,
    ICMD_REG_NOT_SUPPORTED     = 0x0C,
    ICMD_CLASS_NOT_SUPPORTED   = 0x0D,
    ICMD_METHOD_NOT_SUPPORTED  = 0x0E,
    ICMD_BAD_PACKET            = 0x0F,
    ICMD_RESOURCE_NOT_AVAIL    = 0x10,
    ICMD_MSG_RECEIPT_ACK       = 0x11,
    ICMD_INTERNAL_ERR          = 0x20
};
```

## Error Detection and Diagnosis

### 1. VSC Access Errors
```c
int diagnose_vsc_error(device_t *dev) {
    uint32_t vendor_id, device_id;
    
    // Check if device is accessible
    vendor_id = read_pci_config_word(dev, PCI_VENDOR_ID);
    if (vendor_id == 0xFFFF) {
        return MLX_DEV_NOT_FOUND;
    }
    
    // Check if VSC exists
    uint8_t cap = find_pci_capability(dev, PCI_CAP_ID_VNDR);
    if (cap == 0) {
        log_error("No VSC capability found");
        return MLX_DEV_NOT_SUPPORTED;
    }
    
    // Test VSC functionality
    uint32_t test_val = 0x12345678;
    write_pci_config_dword(dev, vsc_base + VSC_ADDRESS, test_val);
    uint32_t read_val = read_pci_config_dword(dev, vsc_base + VSC_ADDRESS);
    
    if (read_val != test_val) {
        log_error("VSC not functional (wrote 0x%x, read 0x%x)",
                  test_val, read_val);
        return MLX_DEV_COMMUNICATION_ERR;
    }
    
    return MLX_OK;
}
```

### 2. Semaphore Timeout Diagnosis
```c
int diagnose_semaphore_timeout(device_t *dev) {
    uint32_t sem_val, counter_val;
    
    // Read current semaphore state
    sem_val = read_pci_config_dword(dev, vsc_base + VSC_SEMAPHORE);
    counter_val = read_pci_config_dword(dev, vsc_base + VSC_COUNTER);
    
    log_error("Semaphore timeout: sem=0x%x, counter=0x%x", 
              sem_val, counter_val);
    
    // Check if semaphore is stuck
    if (sem_val != 0 && sem_val < counter_val - 1000) {
        log_error("Semaphore appears stuck (old ticket)");
        return MLX_REG_SEMAPHORE_TO;
    }
    
    // Try to force-clear semaphore (recovery)
    write_pci_config_dword(dev, vsc_base + VSC_SEMAPHORE, 0);
    usleep(1000);
    
    // Retry acquisition
    uint32_t new_ticket = read_pci_config_dword(dev, vsc_base + VSC_COUNTER);
    write_pci_config_dword(dev, vsc_base + VSC_SEMAPHORE, new_ticket);
    
    if (read_pci_config_dword(dev, vsc_base + VSC_SEMAPHORE) == new_ticket) {
        log_info("Semaphore recovered after force-clear");
        return MLX_OK;
    }
    
    return MLX_REG_SEMAPHORE_TO;
}
```

### 3. Flash Error Diagnosis
```c
int diagnose_flash_error(device_t *dev, uint32_t addr, int operation) {
    struct mfpa_reg mfpa = {0};
    int rc;
    
    // Re-query flash parameters
    rc = exec_access_reg(dev, REG_MFPA, 0, &mfpa, sizeof(mfpa));
    if (rc != 0) {
        log_error("Cannot query flash parameters");
        return MLX_FLASH_NOT_DETECTED;
    }
    
    // Check address range
    if (addr >= mfpa.flash_size) {
        log_error("Address 0x%x exceeds flash size 0x%x", 
                  addr, mfpa.flash_size);
        return MLX_FLASH_SIZE_MISMATCH;
    }
    
    // Check alignment
    if (operation == OP_ERASE && (addr & 0xFFF)) {
        log_error("Erase address 0x%x not sector-aligned", addr);
        return MLX_FLASH_ALIGN_ERROR;
    }
    
    if (operation == OP_WRITE && (addr & 0xFF)) {
        log_error("Write address 0x%x not block-aligned", addr);
        return MLX_FLASH_ALIGN_ERROR;
    }
    
    // Check if area is protected
    if (addr < 0x1000) {  // Boot area
        log_error("Address 0x%x in protected boot area", addr);
        return MLX_FLASH_PROTECTED;
    }
    
    return MLX_OK;
}
```

## Recovery Procedures

### 1. Device Recovery
```c
int recover_device_communication(device_t *dev) {
    int rc;
    
    log_info("Attempting device recovery...");
    
    // Step 1: Clear any stuck semaphore
    write_pci_config_dword(dev, vsc_base + VSC_SEMAPHORE, 0);
    usleep(10000);
    
    // Step 2: Reset ICMD interface
    vsc_write(dev, ICMD_CTRL_SPACE, ICMD_CTRL_ADDR, 0);
    usleep(10000);
    
    // Step 3: Test basic communication
    struct mgir_reg mgir = {0};
    rc = exec_access_reg(dev, REG_MGIR, 0, &mgir, sizeof(mgir));
    if (rc == 0) {
        log_info("Device communication recovered");
        return MLX_OK;
    }
    
    // Step 4: Try alternate access method
    if (dev->access_type == ACCESS_PCICONF) {
        log_info("Trying memory-mapped access...");
        // Switch to mmap if available
        rc = switch_to_mmap_access(dev);
        if (rc == 0) {
            return MLX_OK;
        }
    }
    
    log_error("Device recovery failed");
    return MLX_DEV_COMMUNICATION_ERR;
}
```

### 2. Flash Recovery (Livefish Mode)
```c
int recover_flash_livefish(device_t *dev, firmware_image_t *recovery_fw) {
    int rc;
    
    log_warn("Device in livefish mode - attempting recovery");
    
    // Enable override cache replacement
    dev->params.ocr = true;
    handle_cache_replacement(dev);
    
    // Force non-failsafe burn
    dev->params.nofs = true;
    
    // Erase entire flash
    log_info("Erasing entire flash...");
    for (uint32_t addr = 0; addr < dev->flash.size; addr += 0x10000) {
        rc = erase_flash_sector(dev, addr, MFBE_ERASE_64K);
        if (rc != 0 && addr >= 0x1000) {  // Ignore boot area errors
            log_error("Erase failed at 0x%x", addr);
            return rc;
        }
        
        printf("\rErasing: %d%%", (addr * 100) / dev->flash.size);
        fflush(stdout);
    }
    printf("\n");
    
    // Write recovery image
    log_info("Writing recovery image...");
    rc = burn_image_direct(dev, recovery_fw);
    if (rc != 0) {
        return rc;
    }
    
    // Reset device
    log_info("Resetting device...");
    struct mrsi_reg mrsi = {.reset_level = MRSI_RESET_PCI};
    exec_access_reg(dev, REG_MRSI, 1, &mrsi, sizeof(mrsi));
    
    sleep(5);  // Wait for reset
    
    return MLX_OK;
}
```

### 3. Partial Burn Recovery
```c
int recover_partial_burn(device_t *dev, burn_context_t *ctx) {
    int rc;
    
    log_error("Burn failed at address 0x%x", ctx->last_addr);
    
    // Determine recovery strategy based on failure point
    if (ctx->state == BURN_SECONDARY) {
        // Secondary burn failed - primary still intact
        log_info("Secondary burn failed, primary image intact");
        log_info("Device should still boot from primary");
        
        // Clean up partial secondary
        for (uint32_t addr = SECONDARY_IMAGE_ADDR; 
             addr < ctx->last_addr; addr += 0x1000) {
            erase_flash_sector(dev, addr, MFBE_ERASE_4K);
        }
        
        return MLX_OK;  // Device still functional
    }
    
    if (ctx->state == BURN_PRIMARY && ctx->secondary_valid) {
        // Primary burn failed but secondary is valid
        log_info("Primary burn failed, switching to secondary");
        
        // Update boot record to use secondary
        rc = update_boot_record(dev, SECONDARY_IMAGE_ADDR);
        if (rc == 0) {
            log_info("Device will boot from secondary image");
            return MLX_OK;
        }
    }
    
    // Both images corrupted - need recovery
    log_error("Both firmware images corrupted - manual recovery required");
    log_error("Use --nofs burn with recovery image");
    
    return MLX_DEV_BAD_STATE;
}
```

### 4. Component Update Recovery
```c
int recover_component_update(device_t *dev, uint32_t component_id) {
    struct mcc_reg mcc = {0};
    int rc;
    
    log_info("Canceling failed component update...");
    
    // Send cancel command
    mcc.command = MCC_CMD_CANCEL;
    mcc.component_index = component_id;
    
    rc = exec_access_reg(dev, REG_MCC, 1, &mcc, sizeof(mcc));
    if (rc != 0) {
        log_error("Failed to cancel update");
        return rc;
    }
    
    // Release update lock
    mcc.command = MCC_CMD_RELEASE_UPDATE;
    rc = exec_access_reg(dev, REG_MCC, 1, &mcc, sizeof(mcc));
    
    return rc;
}
```

## Error Handling Best Practices

### 1. Hierarchical Error Handling
```c
typedef struct {
    int code;
    const char *message;
    int (*recovery_func)(device_t *dev, void *context);
} error_handler_t;

static const error_handler_t error_handlers[] = {
    {MLX_REG_SEMAPHORE_TO, "Semaphore timeout", diagnose_semaphore_timeout},
    {MLX_FLASH_ERASE_FAILED, "Flash erase failed", recover_flash_error},
    {MLX_DEV_COMMUNICATION_ERR, "Device communication error", recover_device_communication},
    // ... more handlers ...
};

int handle_error(device_t *dev, int error_code, void *context) {
    for (int i = 0; i < ARRAY_SIZE(error_handlers); i++) {
        if (error_handlers[i].code == error_code) {
            log_error("%s", error_handlers[i].message);
            
            if (error_handlers[i].recovery_func) {
                int rc = error_handlers[i].recovery_func(dev, context);
                if (rc == MLX_OK) {
                    log_info("Recovery successful");
                    return MLX_OK;
                }
            }
            break;
        }
    }
    
    return error_code;
}
```

### 2. Retry Strategy
```c
#define MAX_RETRIES 3
#define RETRY_DELAY_MS 100

int execute_with_retry(device_t *dev, 
                      int (*func)(device_t *dev, void *arg),
                      void *arg,
                      const char *operation) {
    int rc;
    int retry;
    
    for (retry = 0; retry < MAX_RETRIES; retry++) {
        rc = func(dev, arg);
        
        if (rc == MLX_OK) {
            return MLX_OK;
        }
        
        log_warn("%s failed (attempt %d/%d): %s",
                 operation, retry + 1, MAX_RETRIES,
                 error_to_string(rc));
        
        // Don't retry certain errors
        if (rc == MLX_USER_ABORT || 
            rc == MLX_IMG_INCOMPATIBLE ||
            rc == MLX_DEV_NOT_SUPPORTED) {
            break;
        }
        
        // Exponential backoff
        usleep(RETRY_DELAY_MS * 1000 * (1 << retry));
    }
    
    return rc;
}
```

### 3. Cleanup on Error
```c
typedef void (*cleanup_func_t)(void *context);

typedef struct {
    cleanup_func_t func;
    void *context;
} cleanup_handler_t;

static cleanup_handler_t cleanup_stack[16];
static int cleanup_depth = 0;

void push_cleanup(cleanup_func_t func, void *context) {
    if (cleanup_depth < ARRAY_SIZE(cleanup_stack)) {
        cleanup_stack[cleanup_depth].func = func;
        cleanup_stack[cleanup_depth].context = context;
        cleanup_depth++;
    }
}

void run_cleanup(void) {
    while (cleanup_depth > 0) {
        cleanup_depth--;
        cleanup_handler_t *handler = &cleanup_stack[cleanup_depth];
        if (handler->func) {
            handler->func(handler->context);
        }
    }
}

// Usage example:
int burn_with_cleanup(device_t *dev, firmware_image_t *fw) {
    int rc;
    
    // Register cleanup handlers
    push_cleanup(release_semaphore, dev);
    push_cleanup(free_image_buffer, fw);
    
    rc = do_burn(dev, fw);
    
    if (rc != MLX_OK) {
        log_error("Burn failed, running cleanup");
        run_cleanup();
    }
    
    return rc;
}
```

This comprehensive error handling and recovery system ensures robust operation even in failure scenarios.