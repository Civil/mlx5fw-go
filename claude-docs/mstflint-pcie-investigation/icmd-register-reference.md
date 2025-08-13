# ICMD Register Reference - Complete Command Set

## Overview
The ICMD (Internal Command) interface is used to access device registers through the PCIe configuration space. All register operations go through the ACCESS_REG command.

## ACCESS_REG Command Format

### Command Header
```c
struct access_reg_cmd {
    // DWord 0
    uint16_t opcode;        // 0x905 for ACCESS_REG
    uint8_t  reserved0;
    uint8_t  opcode_mod;    // 0 = Query (read), 1 = Write
    
    // DWord 1
    uint16_t register_id;   // Register ID (see list below)
    uint8_t  reserved1;
    uint8_t  status;        // Returned in response
    
    // DWord 2
    uint32_t reserved2;
    
    // DWord 3
    uint32_t argument;      // Register-specific argument
    
    // Register data follows...
} __attribute__((packed));
```

## Complete Register List

### 1. MGIR - Management General Information Register (0x907f)
```c
#define REG_MGIR 0x907f
#define MGIR_SIZE 72  // From logs: "Register Size: 72 bytes"

struct mgir_reg {
    uint8_t  hw_info[32];       // Hardware info string
    uint8_t  fw_info[32];       // Firmware info string
    uint32_t hw_dev_id;         // Hardware device ID
    uint16_t hw_rev_id;         // Hardware revision
    uint16_t reserved;
} __attribute__((packed));

// Usage: Query general device information
```

### 2. MCQS - Management Component Query Status (0x9060)
```c
#define REG_MCQS 0x9060
#define MCQS_SIZE 16

struct mcqs_reg {
    uint32_t component_index;   // Component to query
    uint32_t device_index;      // Device index (multi-device)
    uint32_t last_index_flag;   // 1 if last component
    uint32_t identifier;        // Component identifier returned
} __attribute__((packed));

// Usage: Enumerate available components
```

### 3. MCQI - Management Component Query Information (0x9061)
```c
#define REG_MCQI 0x9061
#define MCQI_SIZE 148  // From logs: "Register Size: 148 bytes"

struct mcqi_reg {
    uint32_t component_index;   // Input: component to query
    uint32_t device_index;      // Input: device index
    uint32_t read_pending;      // Input: 1=pending, 0=active
    uint32_t info_type;         // Input: type of info requested
    uint32_t offset;            // Input: offset for data
    uint32_t data_size;         // Input/Output: size of data
    
    // Output fields
    uint32_t max_component_size;
    uint32_t component_alignment;
    uint32_t component_identifier;
    uint32_t component_version;
    uint32_t supported_info_bitmask;
    uint8_t  component_name[16];
    uint8_t  description[64];
    uint8_t  data[32];          // Component-specific data
} __attribute__((packed));

// Usage: Get detailed component information
// info_type values:
#define MCQI_INFO_TYPE_CAPABILITIES  0
#define MCQI_INFO_TYPE_VERSION       1
#define MCQI_INFO_TYPE_ACTIVATION    2
```

### 4. MCC - Management Component Control (0x9062)
```c
#define REG_MCC 0x9062
#define MCC_SIZE 24

struct mcc_reg {
    uint16_t command;           // Command to execute
    uint16_t command_type;      // Type of command
    uint32_t component_index;   // Target component
    uint32_t device_index;      // Target device
    uint32_t component_size;    // Size for update
    uint32_t offset;            // Offset for data transfer
    uint32_t data_size;         // Size of this transfer
} __attribute__((packed));

// Commands:
#define MCC_CMD_LOCK_UPDATE     0x01
#define MCC_CMD_RELEASE_UPDATE  0x02
#define MCC_CMD_UPDATE          0x03
#define MCC_CMD_VERIFY          0x04
#define MCC_CMD_ACTIVATE        0x05
#define MCC_CMD_CANCEL          0x06

// Usage: Component-based firmware update
```

### 5. MCDA - Management Component Data Access (0x9063)
```c
#define REG_MCDA 0x9063
#define MCDA_SIZE 2048  // Max data transfer size

struct mcda_reg {
    uint32_t offset;            // Offset in component
    uint32_t size;              // Size of data
    uint8_t  data[2040];        // Actual data
} __attribute__((packed));

// Usage: Transfer component data (with MCC)
```

### 6. MFPA - Management Flash Parameters Access (0x9010)
```c
#define REG_MFPA 0x9010
#define MFPA_SIZE 32

struct mfpa_reg {
    uint8_t  fs;                // Flash select (0-3)
    uint8_t  reserved[3];
    
    // Output fields
    uint32_t boot_address;      // Boot address
    uint32_t flash_size;        // Total flash size
    uint32_t jedec_id;          // JEDEC manufacturer ID
    uint16_t sector_size;       // Sector size
    uint8_t  block_alignment;   // Write block alignment
    uint8_t  reserved2;
    uint32_t capability_mask;   // Flash capabilities
} __attribute__((packed));

// Capability bits:
#define MFPA_CAP_ERASE_4K   (1 << 0)
#define MFPA_CAP_ERASE_64K  (1 << 1)
#define MFPA_CAP_WRITE_256B (1 << 2)
#define MFPA_CAP_WRITE_4K   (1 << 3)

// From logs: "JEDEC_ID: 0x1840ef" (Winbond W25QxxBV)
```

### 7. MFBA - Management Flash Burn Access (0x9011)
```c
#define REG_MFBA 0x9011
#define MFBA_SIZE 288  // 32 header + 256 data

struct mfba_reg {
    uint8_t  fs;                // Flash select
    uint8_t  reserved[3];
    uint32_t address;           // Flash address
    uint32_t size;              // Data size (up to 256)
    uint32_t access_mode;       // 0=read, 1=write
    uint8_t  reserved2[16];
    uint8_t  data[256];         // Read/write data
} __attribute__((packed));

// Usage: Direct flash read/write operations
```

### 8. MFBE - Management Flash Block Erase (0x9012)
```c
#define REG_MFBE 0x9012
#define MFBE_SIZE 16

struct mfbe_reg {
    uint8_t  fs;                // Flash select
    uint8_t  reserved[3];
    uint32_t address;           // Erase address (must be aligned)
    uint8_t  erase_size;        // Erase size selector
    uint8_t  reserved2[7];
} __attribute__((packed));

// Erase sizes:
#define MFBE_ERASE_4K   0
#define MFBE_ERASE_32K  1
#define MFBE_ERASE_64K  2
#define MFBE_ERASE_CHIP 3

// Usage: Erase flash sectors before writing
```

### 9. MQIS - Management Queue Information Status (0x9067)
```c
#define REG_MQIS 0x9067
#define MQIS_SIZE 16

struct mqis_reg {
    uint32_t info_type;         // Type of info requested
    uint32_t info_size;         // Size of info data
    uint32_t offset;            // Offset for large data
    uint32_t reserved;
} __attribute__((packed));

// Usage: Query queue/buffer information
```

### 10. MGCR - Management General Configuration Register (0x903B)
```c
#define REG_MGCR 0x903B
#define MGCR_SIZE 32

struct mgcr_reg {
    uint32_t pci_rescan_required;
    uint32_t driver_unload_required;
    uint8_t  reserved[24];
} __attribute__((packed));

// Usage: Check if PCI rescan needed after update
```

### 11. MRSI - Management Reset and Init Register (0x9023)
```c
#define REG_MRSI 0x9023
#define MRSI_SIZE 8

struct mrsi_reg {
    uint8_t  reset_level;       // Reset level to perform
    uint8_t  reserved[7];
} __attribute__((packed));

// Reset levels:
#define MRSI_RESET_SOFT     0
#define MRSI_RESET_HARD     1
#define MRSI_RESET_PCI      2

// Usage: Reset device after firmware update
```

### 12. MDIR - Management Dump Information Register (0x911A)
```c
#define REG_MDIR 0x911A
#define MDIR_SIZE 64

struct mdir_reg {
    uint32_t dump_type;         // Type of dump
    uint32_t seq_num;           // Sequence number
    uint32_t num_of_obj;        // Number of objects
    uint8_t  reserved[52];
} __attribute__((packed));

// Usage: Debug dumps and diagnostics
```

## ICMD Execution Flow

### 1. Command Preparation
```c
int prepare_access_reg_cmd(uint8_t *buffer, uint16_t reg_id, 
                          uint8_t write, void *reg_data, size_t reg_size) {
    struct access_reg_cmd *cmd = (struct access_reg_cmd *)buffer;
    
    // Clear buffer
    memset(buffer, 0, ICMD_MAILBOX_SIZE);
    
    // Fill header
    cmd->opcode = cpu_to_be16(0x905);
    cmd->opcode_mod = write ? 1 : 0;
    cmd->register_id = cpu_to_be16(reg_id);
    
    // Copy register data if writing
    if (write && reg_data) {
        memcpy(buffer + sizeof(*cmd), reg_data, reg_size);
    }
    
    return sizeof(*cmd) + (write ? reg_size : 0);
}
```

### 2. Command Execution Sequence
```c
int exec_access_reg(device_t *dev, uint16_t reg_id, uint8_t write,
                    void *reg_data, size_t reg_size) {
    uint8_t cmd_buf[ICMD_MAILBOX_SIZE];
    uint8_t resp_buf[ICMD_MAILBOX_SIZE];
    int cmd_size;
    
    // Prepare command
    cmd_size = prepare_access_reg_cmd(cmd_buf, reg_id, write, 
                                     reg_data, reg_size);
    
    // Execute via ICMD
    int rc = execute_icmd(dev, cmd_buf, cmd_size, 
                         resp_buf, sizeof(resp_buf));
    if (rc != 0) {
        return rc;
    }
    
    // Check response status
    struct access_reg_cmd *resp = (struct access_reg_cmd *)resp_buf;
    if (resp->status != 0) {
        return -EIO;
    }
    
    // Copy response data if reading
    if (!write && reg_data) {
        memcpy(reg_data, resp_buf + sizeof(*resp), reg_size);
    }
    
    return 0;
}
```

## Common Command Sequences

### 1. Component Enumeration
```c
// Step 1: Query component count
struct mcqs_reg mcqs = {.component_index = 0};
exec_access_reg(dev, REG_MCQS, 0, &mcqs, sizeof(mcqs));

// Step 2: For each component, get details
for (int i = 0; i <= mcqs.last_index_flag; i++) {
    struct mcqi_reg mcqi = {
        .component_index = i,
        .info_type = MCQI_INFO_TYPE_CAPABILITIES
    };
    exec_access_reg(dev, REG_MCQI, 0, &mcqi, sizeof(mcqi));
    // Process component info...
}
```

### 2. Flash Parameters Query
```c
struct mfpa_reg mfpa = {.fs = 0};  // Flash select 0
exec_access_reg(dev, REG_MFPA, 0, &mfpa, sizeof(mfpa));

printf("Flash JEDEC ID: 0x%x\n", mfpa.jedec_id);
printf("Flash size: %u bytes\n", mfpa.flash_size);
printf("Sector size: %u bytes\n", mfpa.sector_size);
```

### 3. Component Update Flow
```c
// Lock component for update
struct mcc_reg mcc = {
    .command = MCC_CMD_LOCK_UPDATE,
    .component_index = comp_idx,
    .component_size = fw_size
};
exec_access_reg(dev, REG_MCC, 1, &mcc, sizeof(mcc));

// Transfer data in chunks
for (offset = 0; offset < fw_size; offset += chunk_size) {
    // Send data via MCDA
    struct mcda_reg mcda = {
        .offset = offset,
        .size = chunk_size
    };
    memcpy(mcda.data, fw_data + offset, chunk_size);
    exec_access_reg(dev, REG_MCDA, 1, &mcda, 
                   sizeof(mcda) - sizeof(mcda.data) + chunk_size);
    
    // Update component
    mcc.command = MCC_CMD_UPDATE;
    mcc.offset = offset;
    mcc.data_size = chunk_size;
    exec_access_reg(dev, REG_MCC, 1, &mcc, sizeof(mcc));
}

// Verify and activate
mcc.command = MCC_CMD_VERIFY;
exec_access_reg(dev, REG_MCC, 1, &mcc, sizeof(mcc));

mcc.command = MCC_CMD_ACTIVATE;
exec_access_reg(dev, REG_MCC, 1, &mcc, sizeof(mcc));
```

## Error Codes

Common ICMD/ACCESS_REG status codes:
- 0x00: Success
- 0x01: Invalid opcode
- 0x02: Invalid register ID
- 0x03: Invalid size/parameter
- 0x04: Resource busy
- 0x05: Access denied
- 0x06: Hardware error
- 0x07: Timeout

This reference provides all the register definitions and command sequences needed for implementation.