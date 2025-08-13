# mstflint Burn Command Analysis

## Overview

The `burn` command in mstflint is used to flash firmware images to Mellanox/NVIDIA network adapters. This document provides a comprehensive analysis of the burn command behavior, its flags, and special handling for different device types.

## Command Syntax

```bash
mstflint -d <device> -i <image.bin> burn [options]
```

## Key Flags Analysis

### 1. `-ocr` / `--override_cache_replacement` Flag

**Purpose**: Allows accessing the flash even if the cache replacement mode is enabled on SwitchX/ConnectIB devices.

**When it's needed**:
- When the device has cache replacement enabled (indicated by non-zero values in cache replacement offset/command registers)
- When burning firmware in livefish mode (automatically enabled)
- When direct flash access is required but cache replacement guard is active

**Technical Details**:
- The flag sets `ignoreCacheRep` parameter to 1 in the FwOperations params
- Bypasses the cache replacement guard check in `check_cache_replacement_guard()`
- On devices with cache replacement enabled, a dummy write operation is performed to trigger any needed cache replacement
- Warning message displayed: "Firmware flash cache access is enabled. Running in this mode may cause the firmware to hang."

**Source Code References**:
```cpp
// From subcommands.cpp
if (_flintParams.override_cache_replacement)
{
    printf(FLINT_OCR_WARRNING);
}

// From mflash.c
int check_cache_replacement_guard(mflash* mfl, u_int8_t* needs_cache_replacement)
{
    if (mfl->opts[MFO_IGNORE_CASHE_REP_GUARD] == 0) {
        // Check cache replacement registers
        if ((cmd != 0) || (off != 0)) {
            *needs_cache_replacement = 1;
        }
    } else {
        // Ignore cache replacement check
        // Execute dummy write to trigger cache replacement
        MWRITE4(HW_DEV_ID, 0);
    }
}
```

### 2. `-nofs` / `--nofs` Flag

**Purpose**: Burns image in a non-failsafe manner, bypassing safety checks and redundancy mechanisms.

**When it's needed**:
- When device/image information cannot be extracted for verification
- When burning non-failsafe images
- When intentionally bypassing all safety checks
- When PSID/VSD information cannot be extracted from flash

**Technical Details**:
- Sets `burnFailsafe` to false in ExtBurnParams
- Disables all verification checks (PSID, GUIDs, VSD, etc.)
- Burns directly without maintaining backup image
- Overwrites all flash sectors including Invariant Sector
- No rollback capability if burn fails

**Warning Messages**:
```
Burn process will not be failsafe. No checks will be performed.
ALL flash, including the Invariant Sector will be overwritten.
If this process fails, computer may remain in an inoperable state.
```

**Source Code References**:
```cpp
// From subcommands.cpp
_burnParams.burnFailsafe = !_flintParams.nofs;

if (!_burnParams.burnFailsafe)
{
    printf("Burn process will not be failsafe. No checks will be performed.\n");
    printf("ALL flash, including the Invariant Sector will be overwritten.\n");
}
```

### 3. `--allow_psid_change` / `--apc` Flag

**Purpose**: Allows burning a firmware image with a different PSID (Parameter Set ID) than the one currently on flash.

**When it's needed**:
- When changing firmware variant/configuration
- When recovering devices with corrupted PSID
- During OEM/vendor transitions
- When explicitly changing device behavior/features

**Technical Details**:
- Sets `allowPsidChange` to true in ExtBurnParams
- Bypasses PSID verification in `checkPSID()` function
- Warning: Changing PSID may cause device malfunction

**Source Code References**:
```cpp
// From subcommands.cpp
bool BurnSubCommand::checkPSID()
{
    if (!_burnParams.allowPsidChange && 
        strcmp(_imgInfo.fw_info.psid, _devInfo.fw_info.psid))
    {
        // PSID mismatch error
        return false;
    }
}
```

## PCIe Device Specific Behavior

### Livefish Mode Detection

For PCIe devices, mstflint automatically detects "livefish mode" and adjusts behavior:

```cpp
int is_livefish_mode = dm_is_livefish_mode(mf);
if (is_livefish_mode == 1)
{
    _flintParams.override_cache_replacement = true;
}
```

**Livefish mode characteristics**:
- Device is in recovery/minimal firmware mode
- Automatic enabling of override_cache_replacement
- Special handling for MFA2 packages
- May require PSID to be explicitly provided

### PCIe vs Other Device Types

1. **Access Methods**:
   - PCIe devices: Direct PCI configuration space access or via driver
   - Other devices: May use InfiniBand MADs, Ethernet, or other protocols

2. **Cache Replacement**:
   - More critical for PCIe devices due to potential system hangs
   - PCIe devices often have active cache replacement during normal operation

3. **Firmware Control**:
   - PCIe devices may support MCC (Mellanox Configuration Channel) for safer burns
   - Fallback to legacy flow when certain flags are used

## Safety Checks and Validation

### Standard Burn Flow Safety Checks:

1. **Device Query**: Verify device is accessible and get current firmware info
2. **PSID Verification**: Ensure image PSID matches device (unless overridden)
3. **Version Check**: Compare firmware versions
4. **GUID/MAC Validation**: Verify network identifiers
5. **VSD Check**: Validate Vendor Specific Data
6. **Failsafe Verification**: Ensure both image and device support failsafe
7. **Flash Verify**: Verify each sector after write (unless disabled)

### Bypass Conditions:

When using `-nofs`, `-ocr`, or `--allow_psid_change`, various safety mechanisms are bypassed:

```cpp
if (_flintParams.nofs || _flintParams.allow_psid_change || _flintParams.use_dev_rom)
{
    // attempt to fallback to legacy flow (direct flash access via FW)
    _mccSupported = false;
}
```

## Best Practices and Recommendations

1. **Normal Operation**: Use standard burn without special flags when possible
2. **Recovery Scenarios**: Use `-nofs` when device information cannot be read
3. **Cache Replacement Issues**: Use `-ocr` only when necessary and ensure device is idle
4. **PSID Changes**: Use `--allow_psid_change` only with vendor guidance
5. **Always backup current firmware** before using bypass flags

## Implementation Details

### Device Access Flow

1. **Device Initialization**:
```cpp
// From subcommands.cpp
void SubCommand::initDeviceFwParams(char* errBuff, FwOperations::fw_ops_params_t& fwParams)
{
    fwParams.ignoreCacheRep = _flintParams.override_cache_replacement ? 1 : 0;
    fwParams.hndlType = FHT_MST_DEV;
    fwParams.mstHndl = _flintParams.device.c_str();
    fwParams.forceLock = _flintParams.clear_semaphore;
    fwParams.numOfBanks = _flintParams.banks;
}
```

2. **Livefish Mode Auto-Detection**:
```cpp
// Automatic override_cache_replacement for livefish devices
mfile* mf = _fwOps->getMfileObj();
int is_livefish_mode = dm_is_livefish_mode(mf);
if (is_livefish_mode == 1)
{
    _flintParams.override_cache_replacement = true;
}
```

### Burn Process Flow

1. **Pre-burn Checks** (skipped with -nofs):
   - Device accessibility check
   - Firmware query on both device and image
   - PSID verification
   - Version compatibility check
   - GUID/MAC validation
   - VSD verification
   - Failsafe capability check

2. **Burn Execution**:
   - **Failsafe burn**: Maintains backup, verifies each write
   - **Non-failsafe burn**: Direct write, no backup, no verification

3. **Post-burn Actions**:
   - Flash verification (unless --no_flash_verify)
   - Cache image request (for certain devices)
   - Progress reporting via callbacks

### Cache Replacement Mechanism

The cache replacement mechanism is critical for devices that use caching for flash access optimization:

1. **Detection**: Checks cache replacement offset/command registers
2. **Guard Check**: Validates if cache replacement is active
3. **Override**: When -ocr is used, performs dummy write to trigger cache flush

```cpp
// From mflash.c
if (mfl->opts[MFO_IGNORE_CASHE_REP_GUARD]) {
    // Execute dummy write to trigger cache replacement
    MWRITE4(HW_DEV_ID, 0);
}
```

## Error Handling

Common error scenarios and solutions:

1. **Cache Replacement Active**:
   - Error: "Firmware flash cache access is enabled"
   - Solution: Use `-ocr` flag or wait for device idle state

2. **PSID Mismatch**:
   - Error: "PSID mismatch"
   - Solution: Verify correct image or use `--allow_psid_change`

3. **Failsafe Incompatibility**:
   - Error: "Can not burn in a failsafe mode"
   - Solution: Use `-nofs` flag for non-failsafe burn

4. **Livefish Mode**:
   - Automatically handled by enabling override_cache_replacement
   - May require explicit PSID with `--use_psid` flag

5. **MCC Not Supported**:
   - When using -nofs, --allow_psid_change, or --use_dev_rom
   - Falls back to legacy direct flash access

## Flag Interactions and Dependencies

### Flag Combinations

1. **Livefish Mode**: Automatically enables `-ocr`
2. **-nofs**: Often used with `-ocr` for recovery scenarios
3. **--allow_psid_change**: May require `-nofs` if verification fails
4. **Image Reactivation**: Incompatible with `-ocr`

### Decision Flow for Flag Usage

```
Device accessible?
├─ No → Use -nofs (non-failsafe burn)
└─ Yes → Query device
         ├─ Cache replacement active?
         │  └─ Yes → Use -ocr
         ├─ PSID mismatch?
         │  └─ Yes → Use --allow_psid_change
         └─ Failsafe not supported?
            └─ Yes → Use -nofs
```

## Summary

The mstflint burn command provides multiple flags to handle various firmware update scenarios:

- **Normal operation**: No special flags needed for standard firmware updates
- **-ocr**: Override cache replacement for devices with active caching
- **-nofs**: Non-failsafe burn for recovery or special scenarios
- **--allow_psid_change**: Change device configuration/variant

These flags should be used with caution as they bypass important safety mechanisms. Always ensure you have a recovery plan before using these advanced options.