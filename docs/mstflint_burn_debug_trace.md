# mstflint Burn Command Debug Trace Guide

## Overview

This document provides a practical guide for debugging and tracing the mstflint burn command execution. It includes debug commands, key breakpoints, and analysis techniques to understand the burn process in detail.

## 1. Environment Setup for Debugging

### 1.1 Debug Environment Variables

```bash
# Enable component debug output
export FW_COMPS_DEBUG=1

# Enable MTCR debug output
export MTCR_DEBUG_LEVEL=1

# Enable flash operations debug
export MFLASH_DEBUG=1

# Enable register access debug
export REG_ACCESS_DEBUG=1
```

### 1.2 Building mstflint with Debug Symbols

```bash
cd /home/civil/go/src/github.com/Civil/mlx5fw-go/reference/mstflint
./configure --enable-debug CFLAGS="-g -O0" CXXFLAGS="-g -O0"
make clean && make
```

## 2. Key Debug Commands

### 2.1 Basic Burn Command with Debug Output

```bash
# Burn with verbose output
export FW_COMPS_DEBUG=1
sudo mstflint -d /dev/mst/mt4125_pciconf0 -i fw-ConnectX6Dx-rel-22_31_1014-MCX653106A-HDA_Ax-UEFI-14.24.13-FlexBoot-3.6.502.bin burn

# Burn with specific debug flags
sudo mstflint -d /dev/mst/mt4125_pciconf0 -i image.bin burn --yes --vv
```

### 2.2 Query Before Burn

```bash
# Query device state
sudo mstflint -d /dev/mst/mt4125_pciconf0 query

# Hardware query with extended info
sudo mstflint -d /dev/mst/mt4125_pciconf0 hw query
```

## 3. GDB Debugging Session

### 3.1 Starting GDB

```bash
# Start mstflint under gdb
sudo gdb /usr/local/bin/mstflint

# Set arguments
(gdb) set args -d /dev/mst/mt4125_pciconf0 -i image.bin burn --yes

# Set environment variables
(gdb) set environment FW_COMPS_DEBUG=1
```

### 3.2 Key Breakpoints for PCIe Access

```gdb
# MTCR PCIe configuration space access
break mtcr_pciconf_mread4
break mtcr_pciconf_mwrite4
break mtcr_pciconf_cap9_sem
break mtcr_pciconf_rw
break mtcr_pciconf_set_addr_space

# VSC capability detection
break mtcr_pciconf_open
break get_space_support_status

# Flash access functions
break common_read_write_mfba
break common_erase_sector
break reg_access_mfba
break reg_access_mfbe
break reg_access_mcc

# High-level burn functions
break BurnSubCommand::executeCommand
break BurnSubCommand::burnFs3
break FwOperations::FwBurn
break Flash::write
break Flash::erase_sector
```

### 3.3 Tracing PCIe Transactions

```gdb
# Set breakpoint with commands to trace PCIe access
(gdb) break mtcr_pciconf_rw
(gdb) commands
> printf "PCIe %s: addr=0x%x offset=0x%x\n", (rw?"WRITE":"READ"), mf->address_space, offset
> if rw == 1
>   printf "  data=0x%08x\n", *data
> end
> continue
> end

# Trace address space changes
(gdb) break mtcr_pciconf_set_addr_space
(gdb) commands
> printf "Setting address space to 0x%x\n", space_encoding
> continue
> end
```

### 3.4 Monitoring Flash Operations

```gdb
# Monitor MFBA register access
(gdb) break reg_access_mfba
(gdb) commands
> printf "MFBA: method=%s addr=0x%x bank=%d size=%d\n", \
    (method==0?"GET":"SET"), mfba->address, mfba->fs, mfba->size
> continue
> end

# Monitor erase operations
(gdb) break common_erase_sector
(gdb) commands
> printf "Erasing sector at 0x%x, bank=%d, size=%s\n", \
    addr, flash_bank, (erase_size==0x10000?"64KB":"4KB")
> continue
> end
```

## 4. Analyzing Debug Output

### 4.1 PCIe Access Pattern

A typical PCIe access sequence looks like:

```
1. Acquire semaphore (mtcr_pciconf_cap9_sem)
2. Set address space (mtcr_pciconf_set_addr_space)
3. Perform read/write (mtcr_pciconf_rw)
4. Release semaphore (mtcr_pciconf_cap9_sem)
```

### 4.2 Flash Write Sequence

The flash write sequence follows this pattern:

```
1. Query flash parameters (MFPA)
2. Erase target sectors (MFBE)
3. Write data in blocks (MFBA)
4. Verify written data (optional MFBA reads)
```

## 5. Common Debug Scenarios

### 5.1 Debugging VSC Detection Failure

```gdb
# Check VSC capability detection
(gdb) break mtcr_pciconf_open
(gdb) run
(gdb) next
# When at capability detection
(gdb) print mf->vsec_addr
(gdb) print mf->vsec_cap_mask
```

### 5.2 Debugging Semaphore Timeout

```gdb
# Monitor semaphore operations
(gdb) break mtcr_pciconf_cap9_sem
(gdb) commands
> printf "Semaphore %s\n", (state?"LOCK":"UNLOCK")
> if state == 1
>   watch lock_val
>   watch counter
> end
> continue
> end
```

### 5.3 Debugging Flash Access Errors

```gdb
# Check MFBA register values
(gdb) break reg_access_mfba
(gdb) run
(gdb) print *mfba
(gdb) print/x mfba->data[0]@4
```

## 6. Advanced Debugging Techniques

### 6.1 Capturing Full Transaction Log

```bash
# Create a gdb script for logging
cat > burn_trace.gdb << 'EOF'
set pagination off
set logging file burn_trace.log
set logging on

# Log all PCIe transactions
break mtcr_pciconf_rw
commands
silent
printf "[%d] PCIe %s: space=0x%x offset=0x%x", $bpnum, (rw?"WR":"RD"), mf->address_space, offset
if rw == 1
  printf " data=0x%08x", *data
end
printf "\n"
continue
end

# Log all register access
break reg_access_func
commands
silent
printf "[%d] REG_ACCESS: id=0x%x method=%d\n", $bpnum, reg_id, method
continue
end

run
EOF

# Run with script
sudo gdb -x burn_trace.gdb --args mstflint -d /dev/mst/mt4125_pciconf0 -i image.bin burn --yes
```

### 6.2 Analyzing Flash Layout

```gdb
# During burn operation
(gdb) break Flash::write
(gdb) commands
> printf "Flash write: addr=0x%08x size=0x%x\n", addr, cnt
> x/4wx data
> continue
> end
```

### 6.3 Memory-Mapped Access Debugging

For devices using memory-mapped access:

```gdb
# Monitor memory-mapped reads/writes
(gdb) break mread4_ul
(gdb) commands
> printf "MMIO read: offset=0x%x\n", offset
> continue
> end

(gdb) break mwrite4_ul
(gdb) commands
> printf "MMIO write: offset=0x%x value=0x%08x\n", offset, value
> continue
> end
```

## 7. Performance Analysis

### 7.1 Timing Flash Operations

```gdb
# Time erase operations
(gdb) break common_erase_sector
(gdb) commands
> set $start = clock()
> continue
> end

(gdb) break common_erase_sector if $start != 0
(gdb) commands
> printf "Erase took %d ms\n", (clock() - $start) / 1000
> set $start = 0
> continue
> end
```

### 7.2 Analyzing Block Transfer Sizes

```gdb
# Monitor MFBA block sizes
(gdb) break common_read_write_mfba
(gdb) commands
> printf "MFBA transfer: size=%d bytes\n", size
> continue
> end
```

## 8. Error Condition Debugging

### 8.1 Syndrome Analysis

```gdb
# Monitor syndrome errors
(gdb) break get_syndrome_code
(gdb) commands
> finish
> if $eax == 0
>   printf "Syndrome detected: code=0x%x\n", *syndrome_code
> end
> continue
> end
```

### 8.2 Address Space Retry Logic

```gdb
# Monitor address space swapping
(gdb) break swap_pci_address_space
(gdb) commands
> printf "Swapping address space from 0x%x to ", mf->address_space
> finish
> printf "0x%x\n", mf->address_space
> continue
> end
```

## 9. Useful Debug Output Patterns

### 9.1 Successful Burn Pattern

```
1. Device query and validation
2. VSC capability detection
3. Flash parameter query (MFPA)
4. Image sections identification
5. For each section:
   - Erase target area (MFBE)
   - Write data blocks (MFBA)
   - Verify if enabled
6. Update boot records
7. Final verification
```

### 9.2 Common Error Patterns

```
# Semaphore timeout
"Flash gateway timeout"
"ME_SEM_LOCKED"

# Address space error
"syndrome is set and syndrome_code is ADDRESS_OUT_OF_RANGE"

# Flash access error
"MFE_REG_ACCESS_BAD_PARAM"
"MFPA: Flash not connected"
```

## 10. Debug Checklist

When debugging burn issues:

1. ✓ Check device is accessible (`mstflint -d <dev> query`)
2. ✓ Verify VSC capability is detected
3. ✓ Confirm flash parameters are read correctly
4. ✓ Monitor semaphore acquisition/release
5. ✓ Track address space usage
6. ✓ Verify erase operations complete
7. ✓ Check write block sizes and alignment
8. ✓ Monitor syndrome errors
9. ✓ Verify final image state

## Summary

This debug guide provides practical techniques for understanding and troubleshooting the mstflint burn process. Key areas to focus on:

- PCIe VSC access patterns
- Semaphore synchronization
- Flash operation sequencing
- Error detection and recovery
- Performance bottlenecks

By following these debugging techniques, you can gain deep insight into how mstflint interacts with Mellanox/NVIDIA devices at the hardware level.