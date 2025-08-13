# Device Access Design

## Current State

The mlx5fw-go project currently operates purely on firmware files without any direct device access. All operations are performed on firmware binary files in the filesystem.

## Device Access Requirements for Firmware Flashing

Based on the mstflint investigation, implementing firmware flashing will require:

### 1. PCI Device Access
- **PCI Configuration Space**: Access to read/write PCI configuration registers
- **Memory-Mapped I/O (MMIO)**: Access to device BAR regions for register operations
- **DMA Operations**: For efficient data transfer during flashing

### 2. Register Access Methods

#### Direct Register Access
```go
// Proposed interface for register operations
type RegisterAccessor interface {
    // Read a 32-bit register
    ReadRegister(offset uint32) (uint32, error)
    
    // Write a 32-bit register
    WriteRegister(offset uint32, value uint32) error
    
    // Read a block of registers
    ReadBlock(offset uint32, size int) ([]byte, error)
    
    // Write a block of registers
    WriteBlock(offset uint32, data []byte) error
}
```

#### ICMD (Internal Command) Interface
```go
// Proposed interface for ICMD operations
type ICMDInterface interface {
    // Execute an ICMD command
    ExecuteCommand(opcode uint16, params []byte) ([]byte, error)
    
    // Check command status
    GetCommandStatus() (ICMDStatus, error)
    
    // Wait for command completion
    WaitForCompletion(timeout time.Duration) error
}
```

### 3. Flash Access Operations

Based on mstflint's implementation, the following operations are needed:

#### Flash Command Set
```go
type FlashCommands interface {
    // Read flash ID
    ReadFlashID() (uint32, error)
    
    // Erase flash sector
    EraseSector(addr uint32) error
    
    // Program flash page
    ProgramPage(addr uint32, data []byte) error
    
    // Read flash data
    ReadFlash(addr uint32, size int) ([]byte, error)
    
    // Set flash write protection
    SetWriteProtection(enable bool) error
}
```

### 4. Device Identification

```go
type DeviceInfo struct {
    VendorID    uint16
    DeviceID    uint16
    RevisionID  uint8
    SubsystemID uint16
    ClassCode   uint32
    BARSizes    []uint64
}

type DeviceIdentifier interface {
    // Enumerate PCI devices
    EnumerateDevices() ([]DeviceInfo, error)
    
    // Open specific device
    OpenDevice(bus, device, function int) (DeviceHandle, error)
    
    // Close device
    CloseDevice(handle DeviceHandle) error
}
```

## Implementation Approach

### 1. Platform Abstraction Layer

Create platform-specific implementations:

```
pkg/device/
├── device.go          # Common interfaces
├── linux/
│   ├── pci.go        # Linux PCI access via sysfs
│   └── mmap.go       # Memory mapping implementation
├── windows/
│   └── pci.go        # Windows PCI access via WinAPI
└── mock/
    └── device.go     # Mock implementation for testing
```

### 2. Safety and Error Handling

Device access requires careful error handling:

```go
type DeviceError struct {
    Operation string
    Device    string
    Err       error
}

func (e *DeviceError) Error() string {
    return fmt.Sprintf("device operation %s failed on %s: %v", 
        e.Operation, e.Device, e.Err)
}
```

### 3. Permission Requirements

- **Linux**: Requires root access or CAP_SYS_RAWIO capability
- **Windows**: Requires Administrator privileges
- **Testing**: Mock implementations for unprivileged testing

## Integration Points

### 1. With Existing Parser

The parser can be extended to validate firmware compatibility:

```go
type FirmwareValidator interface {
    // Validate firmware is compatible with device
    ValidateFirmware(fw FirmwareInfo, dev DeviceInfo) error
    
    // Check version compatibility
    CheckVersionCompatibility(fwVersion, devVersion string) error
}
```

### 2. With CRC Module

Device operations need CRC verification:

```go
type DeviceCRCHandler interface {
    // Calculate CRC using hardware engine
    CalculateHardwareCRC(data []byte) (uint32, error)
    
    // Verify data integrity during transfer
    VerifyTransferCRC(data []byte, expectedCRC uint32) error
}
```

## Future Considerations

### 1. Asynchronous Operations

For long-running operations like flashing:

```go
type FlashOperation interface {
    // Start asynchronous flash operation
    StartFlash(ctx context.Context, firmware []byte) (<-chan Progress, error)
    
    // Cancel ongoing operation
    Cancel() error
}

type Progress struct {
    Current   int
    Total     int
    Operation string
    Error     error
}
```

### 2. Multi-Device Support

Support for flashing multiple devices:

```go
type MultiDeviceFlasher interface {
    // Flash multiple devices in parallel
    FlashDevices(devices []DeviceHandle, firmware []byte) []error
    
    // Verify all devices have same firmware version
    VerifyConsistency(devices []DeviceHandle) error
}
```

### 3. Recovery Mechanisms

Handle failed flash operations:

```go
type RecoveryHandler interface {
    // Check if device is in recovery mode
    IsInRecoveryMode(dev DeviceHandle) bool
    
    // Attempt recovery flash
    RecoveryFlash(dev DeviceHandle, recoveryFW []byte) error
    
    // Reset device to normal mode
    ExitRecoveryMode(dev DeviceHandle) error
}
```

## Dependencies

To implement device access, consider these Go packages:

1. **syscall**: Low-level system calls for device access
2. **golang.org/x/sys**: Extended system interface
3. **github.com/google/gousb**: USB device access (if needed)
4. **github.com/jaypipes/pcidb**: PCI device database

## Security Considerations

1. **Privilege Escalation**: Never auto-escalate privileges
2. **Input Validation**: Validate all device addresses and offsets
3. **Resource Cleanup**: Ensure proper cleanup of device handles
4. **Atomic Operations**: Make flash operations atomic where possible
5. **Backup**: Consider automatic backup before flashing