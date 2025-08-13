# mstflint Burn Command Implementation Guide - Complete Reference

## Overview

This guide consolidates all the research and documentation needed to implement a burn command compatible with mstflint for PCIe devices. All technical details, sequences, and data structures are documented in the accompanying files.

## Documentation Structure

### 1. **PCIe Register Access** (`pcie-register-access-sequences.md`)
- VSC discovery and access protocol
- Semaphore acquisition with ticket-based locking
- ICMD command execution through mailbox
- Exact timing and retry parameters
- Error recovery sequences

### 2. **ICMD Register Reference** (`icmd-register-reference.md`)
- Complete ACCESS_REG command format
- All register definitions (MGIR, MCQS, MCQI, MFPA, MFBA, MFBE, MCC, etc.)
- Register layouts with exact byte offsets
- Command sequences for common operations
- Status codes and error handling

### 3. **Flash Programming** (`flash-programming-sequences.md`)
- Flash memory layout and organization
- Erase sequences with timing (4KB=45ms, 64KB=150ms)
- Write sequences with 256-byte blocks
- Failsafe burn process (secondary → primary)
- Component-based update via MCC
- Cache replacement handling for PCIe devices

### 4. **State Machine** (`burn-command-state-machine.md`)
- Complete state diagram with all transitions
- Detailed state definitions and handlers
- Error state recovery flows
- Progress tracking implementation
- User interaction points

### 5. **Error Handling** (`error-codes-and-recovery.md`)
- Comprehensive error code definitions
- Diagnostic procedures for each error type
- Recovery strategies (device, flash, partial burn)
- Retry logic and cleanup procedures
- Best practices for robust operation

### 6. **Data Structures** (`data-structures.md`)
- Device handle and context structures
- Flash parameter definitions
- Image format structures (FS4)
- Burn context and progress tracking
- Command option structures

### 7. **Image Format** (`image-format-parsing.md`)
- FS4 image format specification
- Magic pattern detection (0x4D544657)
- ITOC/DTOC parsing procedures
- Section types and validation
- Compatibility checking logic

### 8. **Implementation Plan** (`implementation-checklist.md`)
- 6-week phased implementation
- Dependency graph
- Testing strategy
- Risk mitigation
- Critical path analysis

## Quick Start Guide

### Step 1: Implement PCIe Access Layer
```go
// Start with basic PCI device discovery
dev := pci.FindDevice(0x15b3, deviceID)

// Implement VSC access
vsc := pci.NewVSC(dev)
ticket := vsc.AcquireSemaphore()
defer vsc.ReleaseSemaphore()

// Read/write through VSC
data := vsc.Read(space, address)
vsc.Write(space, address, value)
```

### Step 2: Build ICMD Interface
```go
// Implement ACCESS_REG command
func AccessReg(dev *Device, regID uint16, write bool, data []byte) error {
    cmd := PrepareAccessRegCmd(regID, write, data)
    resp := dev.ExecuteICMD(cmd)
    return CheckStatus(resp)
}
```

### Step 3: Create Flash Operations
```go
// Query flash parameters
mfpa := QueryFlashParams(dev)

// Erase and program sector
EraseSector(dev, addr, ERASE_4KB)
for offset := 0; offset < sectorSize; offset += 256 {
    WriteBlock(dev, addr+offset, data[offset:])
}
```

### Step 4: Implement Burn Logic
```go
// Parse and validate image
img := ParseFirmwareImage(imageData)
if err := ValidateCompatibility(img, dev); err != nil {
    return err
}

// Execute failsafe burn
BurnSecondaryImage(dev, img)
UpdateBootRecord(dev, SECONDARY)
BurnPrimaryImage(dev, img)
UpdateBootRecord(dev, PRIMARY)
```

## Key Implementation Notes

### 1. **PCIe Specifics**
- Always use VSC for register access on PCIe devices
- Implement proper semaphore handling to avoid conflicts
- Handle cache replacement for certain devices (-ocr flag)
- Detect livefish mode and adjust behavior

### 2. **Safety First**
- Default to failsafe burn mode
- Always validate image before burning
- Implement comprehensive pre-burn checks
- Provide clear error messages and recovery options

### 3. **Performance Considerations**
- Use 256-byte blocks for optimal write performance
- Implement parallel sector erase where possible
- Cache register reads when safe
- Minimize VSC transactions

### 4. **Compatibility**
- Support both legacy (MFBA/MFBE) and modern (MCC) methods
- Handle different flash types gracefully
- Maintain compatibility with mstflint behavior
- Follow exact command-line interface

## Testing Approach

### Phase 1: Unit Tests
- Test each component in isolation
- Mock hardware interfaces
- Verify protocol compliance

### Phase 2: Integration Tests
- Test complete flows with mock devices
- Verify error handling paths
- Test recovery scenarios

### Phase 3: Hardware Tests
1. Start with read-only operations (query, verify)
2. Test with --use_fw flag (simulation)
3. Test actual burn with known firmware
4. Verify recovery procedures

## Common Pitfalls to Avoid

1. **Semaphore Issues**: Always implement timeout and recovery
2. **Alignment Errors**: Ensure all addresses are properly aligned
3. **Endianness**: All multi-byte values are big-endian in firmware
4. **Progress Updates**: Update frequently but not too often (performance)
5. **Error Recovery**: Never leave device in inconsistent state

## Debug Features to Implement

### Environment Variables
- `FLASH_DEBUG=1` - Flash operation debugging
- `MFT_DEBUG=1` - General debug messages
- `ICMD_DEBUG=1` - ICMD protocol tracing

### Command Flags
- `-v` - Verbose output levels
- `--log <file>` - Log to file
- `--dry-run` - Simulate without writing

## Validation Checklist

Before considering implementation complete:

- [ ] All query commands work correctly
- [ ] Verify command validates images properly
- [ ] Burn simulation (--use_fw) completes successfully
- [ ] Failsafe burn works on test device
- [ ] Non-failsafe recovery tested
- [ ] All error paths handle gracefully
- [ ] Performance meets expectations (<5 min for 16MB)
- [ ] Memory usage is reasonable
- [ ] Works with all supported devices

## Resources

### Source Code References
- MTCR implementation: `mtcr_ul/`
- Flash operations: `flash/`
- Burn command: `flint/subcommands.cpp`
- Image parsing: `fw_ops/`

### Test Firmware
- Sample images in: `/sample_firmwares/mcx5/`
- Use older version for safe testing
- Always backup current firmware first

## Support and Troubleshooting

### Common Issues
1. **"VSC not supported"**: Device may need driver reload
2. **"Semaphore timeout"**: Another process may be accessing device
3. **"Flash timeout"**: May indicate hardware issue
4. **"CRC mismatch"**: Image may be corrupted

### Recovery Procedures
1. Use `--nofs` flag for recovery burn
2. Boot from secondary image if primary fails
3. Use recovery firmware if both images bad
4. Power cycle if device hangs

This guide, combined with the detailed documentation files, provides everything needed to implement a fully compatible burn command.