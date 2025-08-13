# Data Structure Definitions

## Core Device Structures

### Device Handle
```c
typedef struct mtcr_handle {
    int fd;                     // File descriptor (if using file access)
    void *mmap_base;            // Memory mapped base (if using mmap)
    uint32_t vsc_base;          // VSC capability offset
    uint32_t bar0_base;         // BAR0 base address
    enum {
        ACCESS_PCICONF,         // PCI config space access
        ACCESS_MEMORY,          // Memory mapped access
        ACCESS_INBAND,          // InBand access via driver
    } access_type;
    
    // VSC info
    struct {
        uint16_t vendor_id;
        uint16_t device_id;
        uint8_t  space_supported[32];  // Bitmask of supported spaces
        bool     has_semaphore;
        bool     has_icmd;
    } vsc_info;
    
    // Device info
    struct {
        uint32_t hw_id;
        uint32_t hw_rev;
        uint32_t chip_type;
        bool     is_livefish;
    } device_info;
    
    // Semaphore state
    struct {
        uint32_t my_ticket;
        bool     owned;
    } semaphore;
    
} *mtcr_handle_t;
```

### Flash Parameters
```c
typedef struct flash_params {
    uint32_t jedec_id;          // JEDEC manufacturer + device ID
    uint32_t size;              // Total flash size in bytes
    uint32_t sector_size;       // Sector size (typically 4KB)
    uint32_t num_sectors;       // Number of sectors
    uint32_t block_size;        // Write block size (typically 256B)
    uint32_t erase_64k_size;    // Large erase block size
    uint32_t capabilities;      // Capability bitmask
    
    // Derived values
    uint32_t sector_mask;       // Mask for sector alignment
    uint32_t block_mask;        // Mask for block alignment
    
    // Flash type info
    char     type_name[32];     // e.g., "W25QxxBV"
    uint8_t  manufacturer;      // JEDEC manufacturer code
    uint8_t  memory_type;       // Memory type code
    uint8_t  capacity;          // Capacity code
} flash_params_t;
```

## Firmware Image Structures

### Image Header (FS4 Format)
```c
typedef struct fs4_image_header {
    uint32_t magic_pattern;     // 0x4D544657 ("MTFW")
    uint32_t image_size;        // Total image size
    uint32_t crc;               // Image CRC
    uint32_t version;           // Format version
    uint32_t itoc_offset;       // Offset to ITOC
    uint32_t dtoc_offset;       // Offset to DTOC
    uint8_t  reserved[40];
} __attribute__((packed)) fs4_image_header_t;
```

### ITOC/DTOC Headers
```c
typedef struct toc_header {
    uint32_t signature;         // "MTOC" for ITOC, "DTOC" for DTOC
    uint32_t version;           // TOC version
    uint32_t num_entries;       // Number of TOC entries
    uint32_t entry_size;        // Size of each entry
    uint32_t reserved[4];
} __attribute__((packed)) toc_header_t;
```

### TOC Entry
```c
typedef struct toc_entry {
    uint8_t  type;              // Section type
    uint8_t  flags;             // Section flags
    uint16_t reserved;
    uint32_t offset;            // Offset from image start
    uint32_t size;              // Section size
    uint32_t crc;               // Section CRC
    uint8_t  name[16];          // Section name
} __attribute__((packed)) toc_entry_t;

// Section types
enum section_type {
    // Code sections
    SECTION_BOOT_CODE       = 0x1,
    SECTION_PCI_CODE        = 0x2,
    SECTION_MAIN_CODE       = 0x3,
    SECTION_PCIE_LINK_CODE  = 0x4,
    SECTION_IRON_PREP_CODE  = 0x5,
    SECTION_POST_IRON_BOOT  = 0x6,
    SECTION_UPGRADE_CODE    = 0x7,
    
    // Configuration sections
    SECTION_HW_BOOT_CFG     = 0x8,
    SECTION_HW_MAIN_CFG     = 0x9,
    SECTION_PHY_UC_CODE     = 0xA,
    SECTION_PHY_UC_CONSTS   = 0xB,
    SECTION_PCIE_PHY_UC     = 0xC,
    
    // Info sections
    SECTION_IMAGE_INFO      = 0x10,
    SECTION_FW_BOOT_CFG     = 0x11,
    SECTION_FW_MAIN_CFG     = 0x12,
    SECTION_ROM_CODE        = 0x18,
    SECTION_RESET_INFO      = 0x20,
    
    // Debug sections
    SECTION_DBG_FW_INI      = 0x30,
    SECTION_DBG_FW_PARAMS   = 0x32,
    
    // Security sections
    SECTION_IMAGE_SIG_256   = 0xA0,
    SECTION_PUBLIC_KEYS_2K  = 0xA1,
    SECTION_FORBIDDEN_VER   = 0xA2,
    SECTION_IMAGE_SIG_512   = 0xA3,
    SECTION_PUBLIC_KEYS_4K  = 0xA4,
    
    // Device sections
    SECTION_MFG_INFO        = 0xE0,
    SECTION_DEV_INFO        = 0xE1,
    SECTION_VPD_R0          = 0xE3,
    SECTION_NV_DATA         = 0xE4,
    SECTION_FW_NV_LOG       = 0xE5,
    SECTION_NV_DATA2        = 0xE6,
    SECTION_CRDUMP_MASK     = 0xE9,
    SECTION_PROG_HW_FW      = 0xEB,
};
```

### Image Info Section
```c
typedef struct image_info {
    uint32_t fw_version[3];     // Major.Minor.SubMinor
    uint32_t fw_rel_date;       // Release date
    uint16_t fw_rel_time;       // Release time
    uint16_t fw_ver_extended;   // Extended version
    uint8_t  psid[16];          // Product String ID
    uint8_t  vsd[208];          // Vendor Specific Data
    uint32_t supported_hw_id[8];// Supported hardware IDs
    uint32_t mic_version;       // MIC version
    
    // Security info
    uint32_t security_version;
    uint32_t security_flags;
    
    // Component info
    uint32_t component_count;
    struct {
        uint32_t id;
        uint32_t version;
    } components[16];
} __attribute__((packed)) image_info_t;
```

## Burn Operation Structures

### Burn Context
```c
typedef struct burn_context {
    // Device context
    mtcr_handle_t device;
    flash_params_t flash;
    
    // Image context
    struct {
        uint8_t *data;
        size_t size;
        fs4_image_header_t *header;
        toc_header_t *itoc;
        toc_header_t *dtoc;
        image_info_t *info;
        bool is_encrypted;
        bool is_signed;
    } image;
    
    // Burn parameters
    struct {
        bool use_fw;            // Simulation mode
        bool nofs;              // No failsafe
        bool allow_psid_change; // Allow PSID mismatch
        bool ocr;               // Override cache replacement
        bool force;             // Force burn
        bool silent;            // Silent mode
        bool no_verify;         // Skip verification
        bool use_mcc;           // Use MCC instead of direct
    } params;
    
    // Progress tracking
    struct {
        enum {
            BURN_IDLE,
            BURN_QUERY_DEV,
            BURN_ERASE_SEC,
            BURN_WRITE_SEC,
            BURN_VERIFY_SEC,
            BURN_UPDATE_BOOT_SEC,
            BURN_ERASE_PRI,
            BURN_WRITE_PRI,
            BURN_VERIFY_PRI,
            BURN_UPDATE_BOOT_PRI,
            BURN_COMPLETE
        } stage;
        
        uint32_t current_addr;
        uint32_t total_size;
        uint32_t bytes_done;
        time_t start_time;
        
        void (*callback)(struct burn_context *ctx);
    } progress;
    
    // Error context
    struct {
        int code;
        char message[256];
        uint32_t failed_addr;
        int retry_count;
    } error;
    
    // State for recovery
    struct {
        bool primary_valid;
        bool secondary_valid;
        uint32_t last_good_addr;
        uint32_t boot_backup[1024]; // Boot record backup
    } recovery;
    
} burn_context_t;
```

### Component Update Context
```c
typedef struct component_context {
    uint32_t index;             // Component index
    uint32_t identifier;        // Component ID
    char name[32];              // Component name
    uint32_t version;           // Current version
    uint32_t new_version;       // New version
    uint32_t size;              // Component size
    uint32_t offset;            // Current offset
    
    enum {
        COMP_IDLE,
        COMP_LOCKED,
        COMP_UPDATING,
        COMP_VERIFYING,
        COMP_ACTIVATING,
        COMP_COMPLETE,
        COMP_ERROR
    } state;
    
    uint8_t *data;              // Component data
    time_t start_time;
    
} component_context_t;
```

## Command Structures

### Burn Command Options
```c
typedef struct burn_options {
    // Input/Output
    char *device_path;          // Device to burn
    char *image_path;           // Image file
    char *log_file;             // Log file path
    
    // Burn flags
    bool use_fw;                // --use_fw
    bool nofs;                  // --nofs
    bool allow_rom_change;      // --allow_rom_change
    bool ocr;                   // --override_cache_replacement
    bool no_flash_verify;       // --no_flash_verify
    bool use_image_ps;          // --use_image_ps
    bool use_image_guids;       // --use_image_guids
    bool use_dev_rom;           // --use_dev_rom
    bool ignore_dev_data;       // --ignore_dev_data
    bool dual_image;            // --dual_image
    bool striped_image;         // --striped_image
    bool low_cpu;               // --low_cpu
    
    // User interaction
    bool yes;                   // -y/--yes
    bool no;                    // --no
    bool silent;                // -s/--silent
    int verbosity;              // -v level
    
    // Advanced options
    uint32_t banks;             // --banks
    struct {
        uint32_t type;
        uint32_t log2size;
        uint32_t num_flashes;
    } flash_params;             // --flash_params
    
    // Security options
    char *private_key;          // --private_key
    char *public_key;           // --public_key
    char *key_uuid;             // --key_uuid
    
} burn_options_t;
```

## Utility Structures

### Circular Buffer (for logging)
```c
typedef struct log_buffer {
    char *data;
    size_t size;
    size_t head;
    size_t tail;
    pthread_mutex_t mutex;
} log_buffer_t;
```

### Progress Bar
```c
typedef struct progress_bar {
    char label[64];
    int width;
    int current;
    int total;
    time_t start_time;
    bool show_eta;
    bool show_speed;
    
    // Callbacks
    void (*update)(struct progress_bar *bar);
    void (*complete)(struct progress_bar *bar);
} progress_bar_t;
```

### Memory Pool (for efficiency)
```c
typedef struct mem_pool {
    uint8_t *base;
    size_t size;
    size_t used;
    
    struct mem_block {
        size_t size;
        bool free;
        struct mem_block *next;
    } *blocks;
    
} mem_pool_t;
```

## Constants and Enumerations

### Magic Numbers
```c
#define MTFW_MAGIC      0x4D544657  // "MTFW"
#define ITOC_MAGIC      0x4D544F43  // "MTOC"
#define DTOC_MAGIC      0x44544F43  // "DTOC"
#define BOOT_MAGIC      0x4D424F54  // "MBOT"
```

### Address Constants
```c
#define BOOT_RECORD_ADDR      0x000000
#define PRIMARY_IMAGE_ADDR    0x001000
#define SECONDARY_IMAGE_ADDR  0x800000
#define NVDATA_ADDR          0xF00000
#define MAX_IMAGE_SIZE       0x7FF000  // ~8MB
```

### Timing Constants
```c
#define VSC_RETRY_DELAY_US        100
#define VSC_MAX_RETRIES          1000
#define ICMD_POLL_DELAY_US       1000
#define ICMD_MAX_ITERATIONS     10000
#define FLASH_ERASE_4K_US       50000
#define FLASH_ERASE_64K_US     200000
#define FLASH_WRITE_256B_US      2000
#define COMPONENT_UPDATE_TIMEOUT   300  // seconds
```

This comprehensive set of data structures provides everything needed for implementing the burn command with full compatibility.