# PCIe Register Access Sequences - Detailed Implementation

## 1. PCIe Configuration Space Access

### VSC (Vendor Specific Capability) Discovery
```c
// Step 1: Find VSC capability in PCI config space
uint8_t cap_offset = read_pci_config_byte(dev, PCI_CAPABILITY_LIST);
while (cap_offset != 0) {
    uint8_t cap_id = read_pci_config_byte(dev, cap_offset);
    if (cap_id == PCI_CAP_ID_VNDR) {  // 0x09 = Vendor Specific
        // Found VSC capability
        break;
    }
    cap_offset = read_pci_config_byte(dev, cap_offset + 1);
}

// Step 2: Read VSC header
uint32_t vsc_header = read_pci_config_dword(dev, cap_offset + 4);
if ((vsc_header & 0xFFFF) == MELLANOX_VSC_ID) {
    // Valid Mellanox VSC found
}
```

### VSC Register Layout
```
Offset  | Register      | Description
--------|---------------|------------------------------------------
0x00    | Cap ID        | 0x09 (Vendor Specific)
0x01    | Next Cap      | Pointer to next capability
0x02    | VSC Length    | Length of VSC structure
0x04    | Vendor ID     | 0x15b3 (Mellanox)
0x08    | Reserved      | -
0x0C    | Space         | Address space selector
0x10    | Address       | Target address within selected space
0x14    | Data          | Data register for read/write
0x18    | Counter       | Semaphore ticket counter
0x1C    | Semaphore     | Semaphore for exclusive access
```

### Basic VSC Read/Write Operations

#### VSC Read Sequence:
```c
uint32_t vsc_read(device_t *dev, uint8_t space, uint32_t addr) {
    // 1. Acquire semaphore
    uint32_t ticket = read_pci_config_dword(dev, vsc_base + VSC_COUNTER);
    write_pci_config_dword(dev, vsc_base + VSC_SEMAPHORE, ticket);
    
    // 2. Verify semaphore ownership
    if (read_pci_config_dword(dev, vsc_base + VSC_SEMAPHORE) != ticket) {
        // Retry or error
    }
    
    // 3. Set address space
    write_pci_config_dword(dev, vsc_base + VSC_SPACE, space);
    
    // 4. Set address
    write_pci_config_dword(dev, vsc_base + VSC_ADDRESS, addr);
    
    // 5. Read data
    uint32_t data = read_pci_config_dword(dev, vsc_base + VSC_DATA);
    
    // 6. Release semaphore (write 0)
    write_pci_config_dword(dev, vsc_base + VSC_SEMAPHORE, 0);
    
    return data;
}
```

#### VSC Write Sequence:
```c
void vsc_write(device_t *dev, uint8_t space, uint32_t addr, uint32_t data) {
    // 1. Acquire semaphore
    uint32_t ticket = read_pci_config_dword(dev, vsc_base + VSC_COUNTER);
    write_pci_config_dword(dev, vsc_base + VSC_SEMAPHORE, ticket);
    
    // 2. Verify semaphore ownership
    if (read_pci_config_dword(dev, vsc_base + VSC_SEMAPHORE) != ticket) {
        // Retry or error
    }
    
    // 3. Set address space
    write_pci_config_dword(dev, vsc_base + VSC_SPACE, space);
    
    // 4. Set address
    write_pci_config_dword(dev, vsc_base + VSC_ADDRESS, addr);
    
    // 5. Write data
    write_pci_config_dword(dev, vsc_base + VSC_DATA, data);
    
    // 6. Release semaphore
    write_pci_config_dword(dev, vsc_base + VSC_SEMAPHORE, 0);
}
```

## 2. ICMD (Internal Command) Interface

### ICMD Initialization Sequence
```c
// From debug logs:
// -D- Getting VCR_CMD_SIZE_ADDR
// -D- MREAD4_ICMD: off: 1000, addr_space: 3
// -D- iCMD command addr: 0x100000
// -D- iCMD ctrl addr: 0x0
// -D- iCMD max mailbox size: 0x340  size 832

struct icmd_info {
    uint32_t cmd_addr;      // 0x100000 (mailbox address)
    uint32_t ctrl_addr;     // 0x0 (control register)
    uint32_t mailbox_size;  // 0x340 (832 bytes)
    uint8_t  cmd_space;     // 2 (mailbox space)
    uint8_t  ctrl_space;    // 3 (control space)
};
```

### ICMD Command Execution Sequence
```c
int execute_icmd(device_t *dev, void *cmd_buf, size_t cmd_size, 
                 void *resp_buf, size_t resp_size) {
    // 1. Check if previous command completed
    uint32_t ctrl = vsc_read(dev, ICMD_CTRL_SPACE, ICMD_CTRL_ADDR);
    if (ctrl & ICMD_BUSY_BIT) {
        return -EBUSY;
    }
    
    // 2. Write command to mailbox
    for (int i = 0; i < cmd_size; i += 4) {
        uint32_t data = *(uint32_t*)(cmd_buf + i);
        vsc_write(dev, ICMD_CMD_SPACE, ICMD_CMD_ADDR + i, data);
    }
    
    // 3. Set GO bit to start command
    ctrl |= ICMD_GO_BIT;
    vsc_write(dev, ICMD_CTRL_SPACE, ICMD_CTRL_ADDR, ctrl);
    
    // 4. Wait for command completion
    int iterations = 0;
    do {
        usleep(1000); // 1ms delay
        ctrl = vsc_read(dev, ICMD_CTRL_SPACE, ICMD_CTRL_ADDR);
        iterations++;
        // Debug: "Waiting for busy-bit to clear (iteration #%d)..."
    } while ((ctrl & ICMD_BUSY_BIT) && iterations < MAX_ITERATIONS);
    
    // 5. Check status
    if (ctrl & ICMD_STATUS_MASK) {
        return -EIO; // Command failed
    }
    
    // 6. Read response from mailbox
    for (int i = 0; i < resp_size; i += 4) {
        uint32_t data = vsc_read(dev, ICMD_CMD_SPACE, ICMD_CMD_ADDR + i);
        *(uint32_t*)(resp_buf + i) = data;
    }
    
    return 0;
}
```

## 3. Register Access via ICMD

### Access Register Command Structure
```c
// From debug: "Sending Access Register:"
struct access_register_cmd {
    uint32_t opcode;        // 0x905 for ACCESS_REG
    uint32_t opcode_mod;    // 0 for read, 1 for write
    uint32_t register_id;   // e.g., 0x907f, 0x9061, etc.
    uint32_t argument;      // Register-specific
    uint8_t  data[0];       // Register data follows
};

// Common register IDs observed:
#define REG_MGIR  0x907f  // Management General Info Register
#define REG_MCQS  0x9060  // Management Component Query Status
#define REG_MCQI  0x9061  // Management Component Query Info
#define REG_MFPA  0x9010  // Management Flash Parameters Access
#define REG_MFBA  0x9011  // Management Flash Burn Access
#define REG_MFBE  0x9012  // Management Flash Block Erase
#define REG_MCC   0x9062  // Management Component Control
```

### Example: Reading MGIR Register
```c
int read_mgir(device_t *dev, struct mgir_reg *mgir) {
    struct {
        struct access_register_cmd cmd;
        uint8_t padding[64];
    } cmd_buf = {0};
    
    struct {
        struct access_register_cmd resp;
        struct mgir_reg mgir;
    } resp_buf;
    
    // Setup command
    cmd_buf.cmd.opcode = 0x905;
    cmd_buf.cmd.opcode_mod = 0; // Read
    cmd_buf.cmd.register_id = REG_MGIR;
    
    // Execute via ICMD
    int rc = execute_icmd(dev, &cmd_buf, sizeof(cmd_buf),
                          &resp_buf, sizeof(resp_buf));
    if (rc == 0) {
        *mgir = resp_buf.mgir;
    }
    
    return rc;
}
```

## 4. Flash Access Sequences

### MFPA - Query Flash Parameters
```c
struct mfpa_reg {
    uint32_t flash_num;     // Flash number (0-based)
    uint32_t jedec_id;      // JEDEC manufacturer and device ID
    uint32_t sector_size;   // Sector size in bytes
    uint32_t block_size;    // Write block alignment
    uint32_t capability;    // Flash capabilities bitmask
};

int query_flash_params(device_t *dev, struct mfpa_reg *mfpa) {
    // Use ACCESS_REG command with REG_MFPA
}
```

### MFBE - Erase Flash Sector
```c
int erase_flash_sector(device_t *dev, uint32_t addr) {
    struct mfbe_reg {
        uint32_t flash_num;
        uint32_t erase_addr;
        uint32_t erase_size;  // 0 = 4KB, 1 = 32KB, 2 = 64KB
    } mfbe = {
        .flash_num = 0,
        .erase_addr = addr,
        .erase_size = ERASE_SIZE_4KB
    };
    
    // Send via ACCESS_REG command
    return send_access_reg_write(dev, REG_MFBE, &mfbe, sizeof(mfbe));
}
```

### MFBA - Read/Write Flash Data
```c
int write_flash_block(device_t *dev, uint32_t addr, 
                      void *data, size_t size) {
    struct {
        uint32_t flash_num;
        uint32_t address;
        uint32_t size;
        uint32_t write_mode;  // 1 = write
        uint8_t  data[256];   // Max block size
    } mfba = {
        .flash_num = 0,
        .address = addr,
        .size = size,
        .write_mode = 1
    };
    
    memcpy(mfba.data, data, size);
    
    return send_access_reg_write(dev, REG_MFBA, &mfba, 
                                 sizeof(mfba) - sizeof(mfba.data) + size);
}
```

## 5. Timing and Retry Logic

### Semaphore Acquisition with Retry
```c
#define MAX_SEMAPHORE_RETRIES 1000
#define SEMAPHORE_RETRY_DELAY_US 100

int acquire_vsc_semaphore(device_t *dev) {
    for (int retry = 0; retry < MAX_SEMAPHORE_RETRIES; retry++) {
        uint32_t ticket = read_pci_config_dword(dev, vsc_base + VSC_COUNTER);
        write_pci_config_dword(dev, vsc_base + VSC_SEMAPHORE, ticket);
        
        // Verify ownership
        if (read_pci_config_dword(dev, vsc_base + VSC_SEMAPHORE) == ticket) {
            return ticket; // Success
        }
        
        usleep(SEMAPHORE_RETRY_DELAY_US);
    }
    
    return -ETIMEDOUT;
}
```

### ICMD Polling Parameters
```c
#define ICMD_POLL_DELAY_US      1000    // 1ms initial delay
#define ICMD_POLL_MAX_ITER      10000   // Max 10 seconds
#define ICMD_POLL_BACKOFF       1.5     // Exponential backoff factor

// Control register bits
#define ICMD_BUSY_BIT           (1 << 0)
#define ICMD_GO_BIT             (1 << 1)
#define ICMD_STATUS_SHIFT       8
#define ICMD_STATUS_MASK        (0xFF << ICMD_STATUS_SHIFT)
```

## 6. Error Handling Sequences

### VSC Space Validation
```c
// From logs: "actual_space_value != expected_space_value"
int validate_vsc_space(device_t *dev, uint8_t space) {
    write_pci_config_dword(dev, vsc_base + VSC_SPACE, space);
    uint32_t read_space = read_pci_config_dword(dev, vsc_base + VSC_SPACE);
    
    if (read_space != space) {
        // Space not supported, try alternate access method
        return -ENOTSUP;
    }
    
    return 0;
}
```

### ICMD Error Recovery
```c
int recover_icmd_error(device_t *dev) {
    // 1. Clear any pending semaphore
    write_pci_config_dword(dev, vsc_base + VSC_SEMAPHORE, 0);
    
    // 2. Reset ICMD control register
    vsc_write(dev, ICMD_CTRL_SPACE, ICMD_CTRL_ADDR, 0);
    
    // 3. Wait for hardware to settle
    usleep(10000); // 10ms
    
    // 4. Verify ICMD is idle
    uint32_t ctrl = vsc_read(dev, ICMD_CTRL_SPACE, ICMD_CTRL_ADDR);
    if (ctrl & ICMD_BUSY_BIT) {
        return -EIO; // Still busy, deeper reset needed
    }
    
    return 0;
}
```

## Implementation Notes

1. **Byte Order**: All multi-byte values are in little-endian format
2. **Alignment**: Flash addresses must be aligned to sector boundaries
3. **Caching**: Some operations may require cache flushes (see -ocr flag)
4. **Locking**: Always acquire semaphore before VSC operations
5. **Progress**: Track progress using bytes written vs total size

This document provides the exact register access sequences needed for implementation.