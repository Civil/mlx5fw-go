# Burn Command State Machine

## State Diagram

```
                           ┌─────────────┐
                           │   START     │
                           └──────┬──────┘
                                  │
                                  ▼
                           ┌─────────────┐
                           │ PARSE_ARGS  │ ←─── Parse command line
                           └──────┬──────┘      flags and options
                                  │
                                  ▼
                           ┌─────────────┐
                           │ OPEN_DEVICE │ ←─── mtcr_open()
                           └──────┬──────┘
                                  │
                                  ▼
                           ┌─────────────┐
                    ┌──────│ CHECK_MODE  │ ←─── Livefish detection
                    │      └──────┬──────┘      Device capabilities
                    │             │
                    │             ▼
                    │      ┌─────────────┐
                    │      │ LOAD_IMAGE  │ ←─── Read and validate
                    │      └──────┬──────┘      firmware file
                    │             │
                    │             ▼
                    │      ┌─────────────┐
                    │ ┌────│ QUERY_DEV   │ ←─── Get current FW info
                    │ │    └──────┬──────┘      Check compatibility
                    │ │           │
                    │ │           ▼
                    │ │    ┌─────────────┐
                    │ │    │ PRE_BURN    │ ←─── Validate PSIDs
                    │ │    │   CHECK     │      Check versions
                    │ │    └──────┬──────┘      Security checks
                    │ │           │
                    │ │           ▼
                    │ │    ┌─────────────┐
                    │ │    │  FAILSAFE   │ ──┐
                    │ │    │   CHECK?    │   │
                    │ │    └─────┬───────┘   │
                    │ │          │ Yes        │ No
                    │ │          ▼            ▼
                    │ │   ┌──────────────┐ ┌──────────────┐
                    │ │   │   FAILSAFE   │ │ NON_FAILSAFE │
                    │ │   │    BURN      │ │    BURN      │
                    │ │   └──────┬───────┘ └──────┬───────┘
                    │ │          │                 │
                    │ │          └────────┬────────┘
                    │ │                   ▼
                    │ │           ┌─────────────┐
                    │ │           │   VERIFY    │ ←─── Verify burned
                    │ │           └──────┬──────┘      firmware
                    │ │                  │
                    │ │                  ▼
                    │ │           ┌─────────────┐
                    │ └──────────→│   ERROR     │
                    │             │  HANDLING   │
                    │             └──────┬──────┘
                    │                    │
                    │                    ▼
                    │             ┌─────────────┐
                    └────────────→│  CLEANUP    │ ←─── Close device
                                  └──────┬──────┘      Free resources
                                         │
                                         ▼
                                  ┌─────────────┐
                                  │    EXIT     │
                                  └─────────────┘
```

## State Definitions

### 1. START
- Entry point of burn command
- Initialize global state
- Set up signal handlers

### 2. PARSE_ARGS
```c
struct burn_params {
    char *device_path;      // -d flag
    char *image_path;       // -i flag
    bool use_fw;            // --use_fw (simulation mode)
    bool nofs;              // --nofs (no failsafe)
    bool allow_psid_change; // --allow_psid_change
    bool ocr;               // --override_cache_replacement
    bool force;             // -f/--force
    bool yes;               // -y/--yes (no prompts)
    bool silent;            // -s/--silent
    bool no_flash_verify;   // --no_flash_verify
};
```

### 3. OPEN_DEVICE
```c
enum device_open_result {
    OPEN_SUCCESS,
    OPEN_NOT_FOUND,
    OPEN_NO_PERMISSION,
    OPEN_IN_USE,
    OPEN_VSC_NOT_SUPPORTED
};

struct device_context {
    mtcr_handle_t handle;
    device_type_t type;     // PCIe, I2C, etc.
    uint32_t hw_dev_id;
    uint32_t hw_rev_id;
    bool is_livefish;
    bool has_failsafe;
    flash_params_t flash;
};
```

### 4. CHECK_MODE
- Detect if device is in livefish mode
- Query device capabilities (MCC support, failsafe, etc.)
- Auto-enable OCR if needed
- Determine access method (legacy/MCC)

### 5. LOAD_IMAGE
```c
struct firmware_image {
    uint8_t *data;
    size_t size;
    uint32_t magic_pattern_addr;
    struct {
        bool is_fs4;
        bool is_encrypted;
        bool is_signed;
    } format;
    itoc_header_t itoc;
    dtoc_header_t dtoc;
    image_info_t info;
};
```

### 6. QUERY_DEV
- Get current firmware version
- Read PSIDs
- Check component compatibility
- Gather security features

### 7. PRE_BURN_CHECK
Decision points:
- PSID match check (unless --allow_psid_change)
- Version downgrade check (unless --force)
- Security version check
- Device state validation

### 8. FAILSAFE_BURN
```
Sequence:
1. BACKUP_PRIMARY    → Save primary image info
2. ERASE_SECONDARY   → Erase secondary sectors
3. BURN_SECONDARY    → Write to secondary location
4. VERIFY_SECONDARY  → Verify secondary image
5. UPDATE_BOOT_SEC   → Point boot to secondary
6. ERASE_PRIMARY     → Erase primary sectors
7. BURN_PRIMARY      → Write to primary location
8. VERIFY_PRIMARY    → Verify primary image
9. UPDATE_BOOT_PRI   → Point boot to primary
10. RESTORE_NVDATA   → Restore configuration
```

### 9. NON_FAILSAFE_BURN
```
Sequence:
1. ERASE_ALL        → Erase entire flash
2. BURN_IMAGE       → Write complete image
3. VERIFY_IMAGE     → Verify written data
4. UPDATE_BOOT      → Update boot records
```

## Detailed Sub-States

### BURN_SECONDARY/PRIMARY States
```
                    ┌────────────────┐
                    │ CALC_SECTORS   │ ← Calculate affected sectors
                    └───────┬────────┘
                            │
                            ▼
                    ┌────────────────┐
                    │ ERASE_SECTORS  │ ← For each sector:
                    └───────┬────────┘   - Send MFBE command
                            │            - Wait for completion
                            ▼            - Retry on failure
                    ┌────────────────┐
                    │ WRITE_BLOCKS   │ ← For each block:
                    └───────┬────────┘   - Align to block size
                            │            - Send MFBA write
                            ▼            - Update progress
                    ┌────────────────┐
                    │ VERIFY_BLOCKS  │ ← Read back and compare
                    └────────────────┘   (unless --no_flash_verify)
```

### ERROR_HANDLING States
```c
enum burn_error_code {
    ERR_SUCCESS = 0,
    ERR_DEVICE_NOT_FOUND = 1,
    ERR_PERMISSION_DENIED = 2,
    ERR_IMAGE_NOT_FOUND = 3,
    ERR_IMAGE_CORRUPTED = 4,
    ERR_PSID_MISMATCH = 5,
    ERR_VERSION_MISMATCH = 6,
    ERR_FLASH_ERASE_FAILED = 7,
    ERR_FLASH_WRITE_FAILED = 8,
    ERR_VERIFY_FAILED = 9,
    ERR_SEMAPHORE_TIMEOUT = 10,
    ERR_ICMD_FAILED = 11,
    ERR_USER_ABORT = 12,
    ERR_DEVICE_BUSY = 13,
    ERR_NO_FAILSAFE = 14,
    ERR_SECURITY_VERSION = 15
};
```

## State Transition Conditions

### Success Transitions
- Each state proceeds to next on successful completion
- Progress callbacks update UI/logs

### Error Transitions
- Any state can transition to ERROR_HANDLING
- ERROR_HANDLING determines if retry is possible
- Critical errors go directly to CLEANUP

### User Abort
- SIGINT handler sets abort flag
- States check abort flag before major operations
- Graceful cleanup on abort

## Progress Tracking

```c
struct burn_progress {
    enum {
        PROG_IDLE,
        PROG_ERASING,
        PROG_WRITING,
        PROG_VERIFYING
    } stage;
    
    uint32_t current_addr;
    uint32_t total_size;
    uint8_t percentage;
    
    // Callback for UI updates
    void (*callback)(struct burn_progress *prog);
};
```

## State Machine Implementation

```c
typedef enum {
    STATE_START,
    STATE_PARSE_ARGS,
    STATE_OPEN_DEVICE,
    // ... all states ...
    STATE_EXIT
} burn_state_t;

typedef struct {
    burn_state_t state;
    burn_params_t params;
    device_context_t device;
    firmware_image_t image;
    burn_progress_t progress;
    burn_error_code_t error;
} burn_context_t;

// State handler function type
typedef burn_state_t (*state_handler_t)(burn_context_t *ctx);

// State handler table
static const state_handler_t state_handlers[] = {
    [STATE_START] = handle_start,
    [STATE_PARSE_ARGS] = handle_parse_args,
    [STATE_OPEN_DEVICE] = handle_open_device,
    // ... all handlers ...
};

// Main state machine loop
int burn_main(int argc, char *argv[]) {
    burn_context_t ctx = {0};
    ctx.state = STATE_START;
    
    while (ctx.state != STATE_EXIT) {
        burn_state_t next = state_handlers[ctx.state](&ctx);
        
        if (next == STATE_ERROR_HANDLING && ctx.state != STATE_ERROR_HANDLING) {
            // Log state transition error
            log_error("State %s failed, transitioning to error handler",
                      state_names[ctx.state]);
        }
        
        ctx.state = next;
    }
    
    return ctx.error;
}
```

This state machine provides a clear execution flow with proper error handling and recovery mechanisms.