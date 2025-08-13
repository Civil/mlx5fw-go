# mstflint Burn Command Implementation Guide for PCIe Devices

## Overview

This guide provides a comprehensive implementation roadmap for creating a burn command compatible with mstflint's PCIe device handling, based on extensive analysis of the source code and debug traces.

## Key Components for Implementation

### 1. PCIe Configuration Space Access

#### VSC (Vendor Specific Capability) Structure
```go
type VSCRegisters struct {
    Semaphore uint32 // Offset 0x1c - Ticket-based locking
    Counter   uint32 // Offset 0x18 - Ticket counter
    Address   uint32 // Offset 0x10 - Target address
    Data      uint32 // Offset 0x14 - Data register
}
```

#### Address Space Definitions
- `CR_SPACE` (0x2): Configuration/Control Register space
- `ICMD` (0x3): Internal command space
- `NODNIC_INIT_SEG` (0x4): NIC initialization segment
- `EXPANSION_ROM` (0x5): Expansion ROM space
- `TOOLS_HCR` (0x6): Tools Host Command Register

### 2. MTCR Interface Implementation

#### Key Functions to Implement:
1. **mtcr_open()**: Opens device handle
2. **mtcr_pciconf_mread4()**: Read 4 bytes via PCIe config
3. **mtcr_pciconf_mwrite4()**: Write 4 bytes via PCIe config
4. **Semaphore handling**: Ticket-based locking mechanism

#### Semaphore Protocol:
```go
func acquireSemaphore(dev *Device) error {
    // 1. Read current ticket
    myTicket := readVSC(dev, VSC_COUNTER_OFFSET)
    
    // 2. Write ticket to semaphore
    writeVSC(dev, VSC_SEMAPHORE_OFFSET, myTicket)
    
    // 3. Verify we got the lock
    if readVSC(dev, VSC_SEMAPHORE_OFFSET) != myTicket {
        return errors.New("failed to acquire semaphore")
    }
    
    return nil
}
```

### 3. Flash Access Methods

#### MFBA (Management Flash Burn Access)
- Direct flash read/write operations
- Maximum transfer size: device-specific (check max_reg_size)
- Requires block alignment

#### MCC (Management Component Control)
- Component-based update flow
- Used for newer devices with component support
- Provides progress tracking

### 4. Burn Command Flow

#### Step 1: Device Validation
```go
func validateDevice(dev *Device) error {
    // Open device
    handle := mtcr_open(dev.Path)
    
    // Query device capabilities
    caps := queryDeviceCaps(handle)
    
    // Check if device supports required features
    if !caps.HasFailsafe && !flags.NoFailsafe {
        return errors.New("device doesn't support failsafe burn")
    }
    
    return nil
}
```

#### Step 2: Image Parsing and Validation
```go
func parseAndValidateImage(imagePath string) (*FirmwareImage, error) {
    // Read image file
    data := readFile(imagePath)
    
    // Find magic pattern (0x4d544657)
    imgStart := findMagicPattern(data)
    
    // Parse ITOC (Image Table of Contents)
    itoc := parseITOC(data, imgStart)
    
    // Validate sections
    for _, section := range itoc.Sections {
        if err := validateSection(section); err != nil {
            return nil, err
        }
    }
    
    return &FirmwareImage{
        Data: data,
        ITOC: itoc,
        Start: imgStart,
    }, nil
}
```

#### Step 3: Failsafe Burn Process
```go
func performFailsafeBurn(dev *Device, img *FirmwareImage) error {
    // 1. Burn to secondary image location
    if err := burnToLocation(dev, img, SECONDARY_IMAGE); err != nil {
        return err
    }
    
    // 2. Verify secondary image
    if err := verifyImage(dev, SECONDARY_IMAGE); err != nil {
        return err
    }
    
    // 3. Update boot record to point to secondary
    if err := updateBootRecord(dev, SECONDARY_IMAGE); err != nil {
        return err
    }
    
    // 4. Burn to primary image location
    if err := burnToLocation(dev, img, PRIMARY_IMAGE); err != nil {
        return err
    }
    
    // 5. Update boot record to point to primary
    if err := updateBootRecord(dev, PRIMARY_IMAGE); err != nil {
        return err
    }
    
    return nil
}
```

#### Step 4: Flash Programming
```go
func burnToLocation(dev *Device, img *FirmwareImage, location uint32) error {
    // Calculate erase sectors
    sectors := calculateSectors(img.Size, dev.SectorSize)
    
    // Erase sectors
    for _, sector := range sectors {
        if err := eraseSector(dev, location + sector); err != nil {
            return err
        }
    }
    
    // Write data in blocks
    blockSize := dev.WriteBlockSize
    for offset := 0; offset < img.Size; offset += blockSize {
        block := img.Data[offset:min(offset+blockSize, img.Size)]
        
        if err := writeFlashBlock(dev, location+offset, block); err != nil {
            return err
        }
        
        // Update progress
        updateProgress(offset, img.Size)
    }
    
    return nil
}
```

### 5. PCIe-Specific Considerations

#### Cache Replacement Guard
```go
func handleCacheReplacement(dev *Device, flags *BurnFlags) error {
    if dev.RequiresCacheReplacement || flags.OverrideCacheReplacement {
        // Perform dummy write to trigger cache flush
        writeToAddress(dev, CACHE_REPLACEMENT_ADDR, 0x0)
    }
    return nil
}
```

#### Livefish Mode Detection
```go
func isLivefishMode(dev *Device) bool {
    // Check if device is in recovery mode
    vendorID := readPCIConfig(dev, PCI_VENDOR_ID)
    deviceID := readPCIConfig(dev, PCI_DEVICE_ID)
    
    return vendorID == MELLANOX_VENDOR_ID && 
           (deviceID & LIVEFISH_MASK) == LIVEFISH_PATTERN
}
```

### 6. Error Handling and Recovery

#### Common Error Patterns:
1. **Semaphore timeout**: Retry with exponential backoff
2. **VSC space not supported**: Fall back to alternative access method
3. **Flash erase failure**: Retry sector erase up to 3 times
4. **Verification mismatch**: Log detailed comparison for debugging

### 7. Debug and Logging

#### Essential Debug Points:
```go
func debugLog(format string, args ...interface{}) {
    if os.Getenv("FLASH_DEBUG") == "1" {
        fmt.Printf("[FLASH_DEBUG]: "+format+"\n", args...)
    }
}

// Use throughout implementation:
debugLog("Erasing sector at 0x%x", sectorAddr)
debugLog("Writing block: offset=0x%x, size=%d", offset, len(block))
debugLog("Semaphore acquired, ticket=%d", ticket)
```

## Implementation Checklist

- [ ] PCIe configuration space access (VSC read/write)
- [ ] Semaphore/locking mechanism
- [ ] ICMD interface for register access
- [ ] Flash access (MFBA/MCC)
- [ ] Image parsing (ITOC, sections)
- [ ] Failsafe burn logic
- [ ] Block-aligned flash operations
- [ ] Progress reporting
- [ ] Error handling and recovery
- [ ] Debug logging infrastructure
- [ ] Cache replacement handling
- [ ] Livefish mode support
- [ ] PSID validation
- [ ] Version checking

## Testing Strategy

1. **Unit Tests**: Test each component in isolation
2. **Integration Tests**: Test with mock PCIe device
3. **Hardware Tests**: 
   - Start with query/verify commands
   - Test with --use_fw flag (simulation)
   - Perform actual burn with known-good firmware
   - Test recovery scenarios

## References

- PCIe configuration space: `/mtcr_ul/mtcr_ul_com.c`
- Flash operations: `/flint/subcommands.cpp` (BurnSubCommand)
- Image parsing: `/fw_ops/fw_ops.cpp`, `/fs4_ops.cpp`
- Component access: `/fw_comps_mgr/fw_comps_mgr.cpp`

This guide provides the foundation for implementing a compatible burn command. The key is to follow mstflint's layered architecture and reuse its patterns for reliability and compatibility.
