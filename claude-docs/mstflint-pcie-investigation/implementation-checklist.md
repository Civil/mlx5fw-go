# Implementation Checklist with Dependencies

## Phase 1: Foundation (Week 1)

### 1.1 PCIe Access Layer
- [ ] **PCI device discovery**
  - [ ] Enumerate PCI devices by vendor ID (0x15b3)
  - [ ] Parse BDF (Bus:Device.Function) format
  - [ ] Handle multiple devices
  - Dependencies: None
  - Files: `pci/discovery.go`

- [ ] **PCI configuration space access**
  - [ ] Read/write config space via sysfs
  - [ ] Find capabilities (CAP_ID=0x09)
  - [ ] Parse VSC header
  - Dependencies: Device discovery
  - Files: `pci/config.go`

- [ ] **VSC implementation**
  - [ ] VSC register access (read/write)
  - [ ] Semaphore acquisition/release
  - [ ] Address space validation
  - [ ] Error handling and retry
  - Dependencies: Config space access
  - Files: `pci/vsc.go`
  - Reference: `pcie-register-access-sequences.md`

### 1.2 ICMD Interface
- [ ] **ICMD mailbox protocol**
  - [ ] Command structure encoding
  - [ ] Mailbox read/write via VSC
  - [ ] Status polling
  - [ ] Timeout handling
  - Dependencies: VSC implementation
  - Files: `icmd/protocol.go`
  - Reference: `icmd-register-reference.md`

- [ ] **ACCESS_REG implementation**
  - [ ] Command preparation
  - [ ] Response parsing
  - [ ] Error code handling
  - Dependencies: ICMD mailbox
  - Files: `icmd/access_reg.go`

### 1.3 Device Abstraction
- [ ] **MTCR interface**
  - [ ] Device open/close
  - [ ] Access method selection
  - [ ] Resource management
  - Dependencies: VSC, ICMD
  - Files: `mtcr/device.go`
  - Reference: `data-structures.md`

## Phase 2: Flash Operations (Week 2)

### 2.1 Flash Access
- [ ] **MFPA register (flash params)**
  - [ ] Query flash parameters
  - [ ] Parse JEDEC ID
  - [ ] Determine capabilities
  - Dependencies: ACCESS_REG
  - Files: `flash/params.go`

- [ ] **MFBE register (erase)**
  - [ ] Sector erase (4KB)
  - [ ] Block erase (64KB)
  - [ ] Erase verification
  - [ ] Retry logic
  - Dependencies: ACCESS_REG
  - Files: `flash/erase.go`

- [ ] **MFBA register (read/write)**
  - [ ] Block read (up to 256B)
  - [ ] Block write (aligned)
  - [ ] Data verification
  - Dependencies: ACCESS_REG
  - Files: `flash/io.go`
  - Reference: `flash-programming-sequences.md`

### 2.2 Flash Management
- [ ] **Sector operations**
  - [ ] Calculate affected sectors
  - [ ] Erase sector with retry
  - [ ] Program sector (erase + write)
  - [ ] Verify sector
  - Dependencies: MFBE, MFBA
  - Files: `flash/sector.go`

- [ ] **Progress tracking**
  - [ ] Progress callback interface
  - [ ] Byte-level tracking
  - [ ] ETA calculation
  - [ ] Speed calculation
  - Dependencies: None
  - Files: `flash/progress.go`

## Phase 3: Image Handling (Week 3)

### 3.1 Image Parsing
- [ ] **Image detection**
  - [ ] Find magic pattern
  - [ ] Multiple image support
  - [ ] Header validation
  - Dependencies: None
  - Files: `image/detect.go`
  - Reference: `image-format-parsing.md`

- [ ] **FS4 format parser**
  - [ ] Parse image header
  - [ ] Parse ITOC/DTOC
  - [ ] Extract sections
  - [ ] CRC validation
  - Dependencies: Image detection
  - Files: `image/fs4.go`

- [ ] **Section handlers**
  - [ ] IMAGE_INFO parser
  - [ ] DEV_INFO parser
  - [ ] Boot code validation
  - [ ] Security checks
  - Dependencies: FS4 parser
  - Files: `image/sections.go`

### 3.2 Image Validation
- [ ] **Compatibility checks**
  - [ ] Hardware ID matching
  - [ ] PSID verification
  - [ ] Version comparison
  - [ ] Security version check
  - Dependencies: Section handlers
  - Files: `image/validate.go`

## Phase 4: Burn Implementation (Week 4)

### 4.1 Core Burn Logic
- [ ] **State machine**
  - [ ] State definitions
  - [ ] Transition logic
  - [ ] Error state handling
  - [ ] Recovery states
  - Dependencies: All previous
  - Files: `burn/state.go`
  - Reference: `burn-command-state-machine.md`

- [ ] **Failsafe burn**
  - [ ] Secondary image burn
  - [ ] Boot record update
  - [ ] Primary image burn
  - [ ] Final boot update
  - Dependencies: Flash operations, State machine
  - Files: `burn/failsafe.go`

- [ ] **Non-failsafe burn**
  - [ ] Full erase
  - [ ] Direct write
  - [ ] Recovery mode
  - Dependencies: Flash operations, State machine
  - Files: `burn/direct.go`

### 4.2 Special Handling
- [ ] **Cache replacement**
  - [ ] OCR flag handling
  - [ ] Livefish detection
  - [ ] Cache flush trigger
  - Dependencies: Device access
  - Files: `burn/cache.go`

- [ ] **Component updates (MCC)**
  - [ ] MCC protocol
  - [ ] MCDA data transfer
  - [ ] Progress tracking
  - [ ] Activation
  - Dependencies: ICMD interface
  - Files: `burn/component.go`

## Phase 5: Command Interface (Week 5)

### 5.1 CLI Implementation
- [ ] **Command parsing**
  - [ ] Flag definitions
  - [ ] Validation
  - [ ] Help text
  - Dependencies: None
  - Files: `cmd/burn.go`

- [ ] **User interaction**
  - [ ] Confirmation prompts
  - [ ] Progress display
  - [ ] Error reporting
  - [ ] Verbose output
  - Dependencies: Burn logic
  - Files: `cmd/ui.go`

### 5.2 Error Handling
- [ ] **Error recovery**
  - [ ] Device recovery
  - [ ] Flash recovery
  - [ ] Partial burn recovery
  - [ ] Component recovery
  - Dependencies: All components
  - Files: `errors/recovery.go`
  - Reference: `error-codes-and-recovery.md`

- [ ] **Logging system**
  - [ ] Debug levels
  - [ ] Log file support
  - [ ] Structured logging
  - Dependencies: None
  - Files: `log/logger.go`

## Phase 6: Testing (Week 6)

### 6.1 Unit Tests
- [ ] **Low-level tests**
  - [ ] VSC operations
  - [ ] ICMD protocol
  - [ ] CRC calculations
  - [ ] Image parsing
  - Dependencies: Core implementations
  - Files: `*_test.go`

- [ ] **Mock implementations**
  - [ ] Mock PCI device
  - [ ] Mock flash
  - [ ] Mock firmware images
  - Dependencies: Interfaces
  - Files: `mocks/`

### 6.2 Integration Tests
- [ ] **End-to-end tests**
  - [ ] Query operations
  - [ ] Verify operations
  - [ ] Simulated burn
  - [ ] Error scenarios
  - Dependencies: All components
  - Files: `integration/`

### 6.3 Hardware Testing
- [ ] **Real device tests**
  - [ ] Query commands
  - [ ] Verify with real images
  - [ ] Burn with --use_fw
  - [ ] Recovery scenarios
  - Dependencies: Complete implementation
  - Files: Test scripts

## Implementation Order

### Critical Path:
1. PCI device discovery → VSC access → ICMD protocol → ACCESS_REG
2. Flash parameters → Flash operations → Sector management
3. Image parsing → Validation → Compatibility checks
4. State machine → Burn logic → Error recovery
5. CLI interface → Testing → Hardware validation

### Parallel Work:
- Progress tracking (anytime)
- Logging system (anytime)
- Data structures (as needed)
- Unit tests (with each component)

## Dependencies Summary

```
┌─────────────────┐
│ PCI Discovery   │
└────────┬────────┘
         ▼
┌─────────────────┐
│ VSC Access      │
└────────┬────────┘
         ▼
┌─────────────────┐     ┌─────────────────┐
│ ICMD Protocol   │ ←───│ Image Parser    │
└────────┬────────┘     └─────────────────┘
         ▼                       │
┌─────────────────┐              │
│ Flash Operations│              │
└────────┬────────┘              │
         ▼                       ▼
┌─────────────────────────────────────┐
│        Burn State Machine           │
└─────────────────┬───────────────────┘
                  ▼
         ┌─────────────────┐
         │ CLI Interface   │
         └─────────────────┘
```

## Testing Strategy

### Unit Test Coverage Target: 80%
- Core algorithms: 100%
- Error paths: 90%
- Hardware interfaces: Mocked

### Integration Test Scenarios:
1. Normal failsafe burn
2. Non-failsafe recovery
3. Interrupted burn recovery
4. Version mismatch handling
5. PSID mismatch handling
6. Component update flow

### Hardware Test Plan:
1. Start with query operations only
2. Test verify command extensively
3. Use --use_fw for burn simulation
4. Test with known-good firmware
5. Test recovery procedures
6. Performance benchmarking

## Risk Mitigation

### High Risk Areas:
1. **Semaphore deadlock**: Implement timeout and force-release
2. **Flash corruption**: Always verify after write
3. **Partial burn**: Implement robust recovery
4. **Wrong device**: Multiple confirmation checks

### Safety Measures:
1. Never auto-select device
2. Always show device info before burn
3. Default to failsafe mode
4. Require explicit force flags
5. Comprehensive pre-burn validation

## Documentation Requirements

### Code Documentation:
- [ ] Package documentation
- [ ] Interface documentation  
- [ ] Example usage
- [ ] Error descriptions

### User Documentation:
- [ ] Command reference
- [ ] Troubleshooting guide
- [ ] Recovery procedures
- [ ] FAQ

This checklist provides a complete roadmap for implementing the burn command with all dependencies clearly mapped.