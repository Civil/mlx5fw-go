# mstflint PCIe Device Investigation Report

## Overview

This document provides a comprehensive analysis of mstflint's behavior when working with PCIe devices, including debug information, command analysis, and burn operation details.

## Investigation Contents

### 1. Debug Flags Analysis
- Complete list of all debug environment variables found: `/docs/mstflint-debug-flags.md`
- 33 debug and control environment variables documented
- Key flags for PCIe device debugging:
  - `FW_COMPS_DEBUG` - Firmware components debugging
  - `MFT_DEBUG` - General Mellanox Firmware Tools debug
  - `FLASH_DEBUG` / `MFT_FLASH_DEBUG` - Flash operations
  - `FLASH_ACCESS_DEBUG` - Low-level flash access

### 2. Non-Destructive Command Analysis

#### A. `query full` Command (Script: `01-query-full.sh`)
- **Log**: `logs/01-query-full.log`
- **Key Observations**:
  - Uses PCIe configuration space access (`mtcr_pciconf_*` functions)
  - Multiple VSC (Vendor Specific Capability) space warnings about unsupported spaces
  - Extensive use of iCMD (internal command) interface
  - Component enumeration through MCQI/MCQS registers
  - Identifies components: BOOT_IMG, DBG_TOKEN, CS_TOKEN, USER_NVCONFIG, MLNX_NVCONFIG, OEM_NVCONFIG, RUNTIME_IMG
  - Shows detailed firmware information including versions, PSIDs, and component details

#### B. `hw query` Command (Script: `02-hw-query.sh`)
- **Log**: `logs/02-hw-query.log`
- **Key Observations**:
  - Simpler than `query full` - focuses on hardware info
  - Flash initialization uses MFBA (Mellanox Flash Burn Access)
  - Identifies flash chip: W25QxxBV with 16MB capacity
  - Shows hardware device ID (525) and revision

### 3. Burn Command Analysis

Complete analysis available in: `/docs/mstflint_burn_command_analysis.md`

#### Key Flags and Their Behavior:

##### `-ocr` / `--override_cache_replacement`
- **Purpose**: Bypasses cache replacement guard on certain devices
- **PCIe Specific**: Automatically enabled in livefish mode for PCIe devices
- **Risk**: May cause firmware to hang if cache access timing is wrong
- **Implementation**: Performs dummy write to address 0x1c748 to trigger cache flush

##### `-nofs` / `--nofs` (No Fail Safe)
- **Purpose**: Bypasses all safety checks and burns directly
- **PCIe Impact**: Overwrites entire flash including invariant sectors
- **Use Cases**:
  - Device in bad state where normal query fails
  - Factory programming
  - Recovery scenarios
- **Risk**: No recovery if burn fails

##### `--allow_psid_change` / `--apc`
- **Purpose**: Allows changing device PSID (changes device capabilities/features)
- **PCIe Consideration**: May change PCIe device behavior/capabilities
- **Risk**: Device may malfunction with incompatible PSID

### 4. PCIe-Specific Behavior

#### Access Methods
1. **Configuration Space**: Primary method for PCIe devices
2. **iCMD Interface**: Used for register access and commands
3. **MFBA**: Flash access method for burn operations

#### Special Handling
1. **Livefish Mode Detection**: Automatic for PCIe devices in certain states
2. **Cache Replacement**: Critical for PCIe devices to prevent system hangs
3. **Semaphore Management**: Ensures exclusive access during operations
4. **Space Validation**: Checks for supported VSC spaces

#### Safety Mechanisms
1. **Query Before Burn**: Validates device state
2. **PSID Verification**: Ensures firmware compatibility
3. **Version Checks**: Prevents downgrades unless forced
4. **Failsafe Burns**: Default behavior with backup sectors

## Debug Output Analysis

### Key Debug Patterns Observed:

1. **VSC Space Warnings**: 
   ```
   actual_space_value != expected_space_value. expected_space_value: 0x101 actual_space_value: 0x1
   ```
   Indicates certain extended spaces not supported on this device.

2. **iCMD Operations**:
   ```
   -D- MWRITE4_ICMD: off: 0, addr_space: 3
   Busy-bit raised. Waiting for command to exec...
   ```
   Shows command execution flow through internal command interface.

3. **Component Discovery**:
   ```
   [2Kfw_comps_mgr.cpp:1249: -D- Found component with identifier=0x1 index=0 name=COMPID_BOOT_IMG
   ```
   Enumerates firmware components available on device.

## Recommendations for Firmware Updates

### Safe Update Process:
1. Always run `query full` first to verify device state
2. Use default failsafe burn unless recovery scenario
3. Monitor debug output for any errors or warnings
4. Have recovery firmware ready (like samples in `/sample_firmwares/mcx5/`)

### When to Use Special Flags:
- **`-ocr`**: Only if burn fails with cache-related errors
- **`-nofs`**: Only for recovery when device won't respond to queries
- **`--allow_psid_change`**: Only when intentionally changing device configuration

## PCIe Configuration Implementation Details

### Key Technical Findings:

1. **VSC (Vendor Specific Capability) Access**:
   - Located at PCI capability offset 0x09
   - Provides semaphore (0x1c), counter (0x18), address (0x10), data (0x14) registers
   - Uses ticket-based locking for exclusive access
   - Supports 22 different address spaces

2. **ICMD (Internal Command) Interface**:
   - Command mailbox at 0x100000 (address space 2)
   - Control register at offset 0x0 (address space 3)
   - Maximum mailbox size: 832 bytes (0x340)
   - Busy-wait polling for command completion

3. **Flash Access Architecture**:
   - MFBA (direct read/write) for legacy devices
   - MCC (component control) for newer devices
   - Block-aligned operations (typically 256 bytes)
   - Sector erase sizes: 4KB or 64KB

4. **Burn Process Flow**:
   - Failsafe: Secondary → Primary image updates
   - Non-failsafe: Direct overwrite (recovery mode)
   - Cache replacement required for some PCIe devices
   - Progress tracking via component percentage

## Scripts and Logs

All investigation scripts and logs are stored in this directory:
- `01-query-full.sh` - Query full command with debug
- `02-hw-query.sh` - Hardware query with debug
- `03-burn-simulation.sh` - Burn simulation with --use_fw flag
- `04-verify-burn.sh` - Firmware verification script
- `logs/` - Contains detailed debug output from each command

## Implementation Resources

- `burn-implementation-guide.md` - Step-by-step implementation guide
- Code examples with Go-like pseudocode
- PCIe-specific considerations and error handling
- Testing strategy and validation approach

## Related Documentation
- `/docs/mstflint-debug-flags.md` - Complete debug flag reference
- `/docs/mstflint_burn_command_analysis.md` - Detailed burn command analysis
- `/docs/mstflint_pcie_config_space_burn_analysis.md` - Deep technical analysis
- `/docs/mstflint_burn_debug_trace.md` - Debug and tracing guide