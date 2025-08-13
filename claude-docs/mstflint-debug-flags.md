# mstflint Debug Flags and Environment Variables

This document provides a comprehensive list of all debug flags and environment variables that can be used with mstflint for debugging purposes.

## Environment Variables for Debug Output

### Core Debug Variables

1. **`FW_COMPS_DEBUG`**
   - Location: `fw_comps_mgr/fw_comps_mgr.h`
   - Purpose: Enables debug output for firmware components manager operations
   - Usage: `export FW_COMPS_DEBUG=1`

2. **`MFT_DEBUG`**
   - Location: `mtcr_ul/mtcr_common.h`, `mtcr_ul/mtcr_ul_com.c`, `mtcr_ul/mtcr_ul_icmd_cif.c`
   - Purpose: Enables general MFT (Mellanox Firmware Tools) debug messages
   - Usage: `export MFT_DEBUG=1`

3. **`FLASH_DEBUG`**
   - Location: `mflash/mflash_dev_capability.h`
   - Purpose: Enables flash operations debug output
   - Usage: `export FLASH_DEBUG=1`

4. **`MFT_FLASH_DEBUG`**
   - Location: `mflash/mflash_dev_capability.h`
   - Purpose: More specific flash debug output for MFT operations
   - Usage: `export MFT_FLASH_DEBUG=1`

5. **`FLASH_ACCESS_DEBUG`**
   - Location: `mflash/mflash.h`
   - Purpose: Debug output for low-level flash access operations
   - Usage: `export FLASH_ACCESS_DEBUG=1`

### Component-Specific Debug Variables

6. **`MLX_DPA_DEBUG`**
   - Location: `mlxdpa/mlxdpa_utils.h`
   - Purpose: Debug output for DPA (Data Path Accelerator) operations
   - Usage: `export MLX_DPA_DEBUG=1`

7. **`MLX_TOKEN_DEBUG`**
   - Location: `mlxtokengenerator/mlxtkngenerator_utils.h`
   - Purpose: Debug output for token generation operations
   - Usage: `export MLX_TOKEN_DEBUG=1`

8. **`HEX64_DEBUG`**
   - Location: `mft_utils/hex64.cpp`
   - Purpose: Debug output for hex64 utility operations
   - Usage: `export HEX64_DEBUG=1`

9. **`MLXFWOPS_ERRMSG_DEBUG`**
   - Location: `mlxfwops/lib/flint_base.cpp`
   - Purpose: Detailed error message debugging for firmware operations
   - Usage: `export MLXFWOPS_ERRMSG_DEBUG=1`

10. **`MGET_TEMP_DEBUG`**
    - Location: `small_utils/mget_temp.h`
    - Purpose: Debug output for temperature reading operations
    - Usage: `export MGET_TEMP_DEBUG=1`

11. **`FWCTL_DEBUG`**
    - Location: `include/mtcr_ul/fwctrl_ioctl.h`
    - Purpose: Debug output for firmware control operations
    - Usage: `export FWCTL_DEBUG=1`

### Register Access Debug Variables

12. **`DUMP_DEBUG`**
    - Location: `reg_access/reg_access.c`
    - Purpose: Debug output for register dump operations
    - Usage: `export DUMP_DEBUG=1`

13. **`MCQS_DEBUG`**
    - Location: `reg_access/reg_access.c`
    - Purpose: Debug output for MCQS (Module Control Query Status) operations
    - Usage: `export MCQS_DEBUG=1`

14. **`MCQI_DEBUG`**
    - Location: `reg_access/reg_access.c`
    - Purpose: Debug output for MCQI (Module Control Query Information) operations
    - Usage: `export MCQI_DEBUG=1`

15. **`ADB_DUMP`**
    - Location: `small_utils/mcra.c`
    - Purpose: Enables ADB (Adapter Database) dump functionality
    - Usage: `export ADB_DUMP=1`

## Environment Variables for Control and Configuration

### Flash Operation Controls

16. **`MFLASH_BANKS`**
    - Location: `mflash/mflash.c`
    - Purpose: Specifies which flash banks to use
    - Usage: `export MFLASH_BANKS=<bank_config>`

17. **`MFLASH_BANK_DEBUG`**
    - Location: `mflash/mflash_pack_layer.c`
    - Purpose: Debug output for flash bank operations
    - Usage: `export MFLASH_BANK_DEBUG=1`

18. **`MFLASH_ERASE_VERIFY`**
    - Location: `mflash/mflash.c`
    - Purpose: Enables verification after flash erase operations
    - Usage: `export MFLASH_ERASE_VERIFY=1`

19. **`MFLASH_WRITE_RETRIES`**
    - Location: `mflash/mflash.c`
    - Purpose: Sets the number of write retry attempts
    - Usage: `export MFLASH_WRITE_RETRIES=<number>`

20. **`FLINT_ERASE_SIZE`**
    - Location: `mlxfwops/lib/flint_io.cpp`
    - Purpose: Controls the erase block size for flash operations
    - Usage: `export FLINT_ERASE_SIZE=<size>`

### Hardware Control Variables

21. **`FORCE_GPIO_TOGGLE`**
    - Location: `mflash/mflash.c`
    - Purpose: Forces GPIO toggle for certain operations
    - Usage: `export FORCE_GPIO_TOGGLE=1`

22. **`FORCE_RESET_QUAD_EN`**
    - Location: `mflash/mflash.c`
    - Purpose: Forces reset of quad enable bit
    - Usage: `export FORCE_RESET_QUAD_EN=1`

23. **`CONNECTX_FLUSH`**
    - Location: `mtcr_ul/mtcr_ul_com.c`
    - Purpose: Controls ConnectX flush behavior
    - Usage: `export CONNECTX_FLUSH=1`

### Timeout and Performance Controls

24. **`MFT_CMD_SLEEP`**
    - Location: `mtcr_ul/mtcr_ul_icmd_cif.c`
    - Purpose: Sets sleep time between commands (in microseconds)
    - Usage: `export MFT_CMD_SLEEP=<microseconds>`

25. **`MFT_ICMD_TIMEOUT`**
    - Location: `mtcr_ul/mtcr_ul_icmd_cif.c`
    - Purpose: Sets timeout for ICMD operations
    - Usage: `export MFT_ICMD_TIMEOUT=<timeout>`

26. **`MTCR_IB_TIMEOUT`**
    - Location: `fw_comps_mgr/fw_comps_mgr.cpp`
    - Purpose: Sets InfiniBand operation timeout
    - Usage: `export MTCR_IB_TIMEOUT=<timeout>`

27. **`MTCR_SWRESET_TIMER`**
    - Location: `mtcr_ul/mtcr_ib_ofed.c`
    - Purpose: Sets software reset timer value
    - Usage: `export MTCR_SWRESET_TIMER=<value>`

28. **`MTCR_IB_SL`**
    - Location: `mtcr_ul/mtcr_ib_ofed.c`
    - Purpose: Sets InfiniBand service level
    - Usage: `export MTCR_IB_SL=<service_level>`

### Security and Access Controls

29. **`FLINT_IGNORE_SECURITY_VERSION_CHECK`**
    - Location: `mlxfwops/lib/fs4_ops.cpp`, `mlxfwops/lib/fsctrl_ops.cpp`
    - Purpose: Bypasses security version checks (USE WITH CAUTION)
    - Usage: `export FLINT_IGNORE_SECURITY_VERSION_CHECK=1`

30. **`DISABLE_DMA_ACCESS`**
    - Location: `fw_comps_mgr/fw_comps_mgr_abstract_access.cpp`
    - Purpose: Disables DMA access methods
    - Usage: `export DISABLE_DMA_ACCESS=1`

31. **`ENABLE_DMA_ICMD`**
    - Location: `mtcr_ul/mtcr_ul_icmd_cif.c`
    - Purpose: Enables DMA ICMD operations
    - Usage: `export ENABLE_DMA_ICMD=1`

32. **`FW_CTRL`**
    - Location: `fw_comps_mgr/fw_comps_mgr.cpp`
    - Purpose: Controls firmware control method usage
    - Usage: `export FW_CTRL=1`

## Log File Configuration

33. **`FLINT_LOG_FILE`**
    - Location: `mlxfwops/lib/flint_base.h`, `flint/subcommands.cpp`
    - Purpose: Specifies log file for flint operations
    - Usage: `export FLINT_LOG_FILE=/path/to/logfile`

## Usage Examples

### Basic Debug Output
```bash
# Enable basic firmware components debug
export FW_COMPS_DEBUG=1
mstflint -d /dev/mst/mt4103_pci_cr0 query

# Enable multiple debug outputs
export FW_COMPS_DEBUG=1
export MFT_DEBUG=1
export FLASH_DEBUG=1
mstflint -d /dev/mst/mt4103_pci_cr0 burn firmware.bin
```

### Advanced Debugging
```bash
# Full debug with log file
export FW_COMPS_DEBUG=1
export MFT_DEBUG=1
export FLASH_DEBUG=1
export FLASH_ACCESS_DEBUG=1
export FLINT_LOG_FILE=/tmp/mstflint_debug.log
mstflint -d /dev/mst/mt4103_pci_cr0 query
```

### Performance Tuning
```bash
# Adjust timeouts and retries for slow systems
export MFT_ICMD_TIMEOUT=30000
export MFLASH_WRITE_RETRIES=5
export MFT_CMD_SLEEP=1000
mstflint -d /dev/mst/mt4103_pci_cr0 burn firmware.bin
```

## Important Notes

1. **Performance Impact**: Enabling debug flags significantly increases output verbosity and may slow down operations.

2. **Security Warnings**: Some flags like `FLINT_IGNORE_SECURITY_VERSION_CHECK` bypass important security checks. Use with extreme caution and only in controlled environments.

3. **Log Files**: When using `FLINT_LOG_FILE`, ensure the directory exists and is writable.

4. **Multiple Flags**: You can enable multiple debug flags simultaneously for comprehensive debugging.

5. **Production Use**: Debug flags should generally not be used in production environments due to performance impact and verbose output.

## Troubleshooting Tips

1. If you're not seeing expected debug output, verify the environment variable is exported:
   ```bash
   echo $FW_COMPS_DEBUG
   ```

2. For comprehensive debugging of a specific issue, enable related debug flags:
   - Flash issues: `FLASH_DEBUG`, `MFT_FLASH_DEBUG`, `FLASH_ACCESS_DEBUG`
   - Communication issues: `MFT_DEBUG`, `FWCTL_DEBUG`
   - Component updates: `FW_COMPS_DEBUG`

3. Use `FLINT_LOG_FILE` to capture debug output for later analysis:
   ```bash
   export FLINT_LOG_FILE=/tmp/mstflint_$(date +%Y%m%d_%H%M%S).log
   ```