PCIe Access in mstflint (Deep Dive)

This document maps the key code paths used by mstflint to communicate with Mellanox/NVIDIA devices over PCIe, with references to the local source tree under reference/mstflint.

Device open and selection
- mstflint accepts device selectors such as a BDF (e.g., 0000:07:00.0) or an MST device path (e.g., /dev/mst/mt4123_pciconf0).
- MST kernel driver (reference/mstflint/kernel):
  - Creates PCICONF devices (/dev/mst/*_mstconf) per PCI function.
  - Provides IOCTLs to access the “PCICONF gateway” for register read/write and block transfers (mst.h: PCICONF_* definitions).
- User-land open logic (reference/mstflint/mtcr_ul/mtcr_ul_com.c):
  - Parses names like /sys/bus/pci/devices/%04x:%02x:%02x.%d/resource (or config) to resolve BDFs.
  - Prefers MST PCICONF when available (mf->tp == MST_PCICONF) and falls back to sysfs config/resource operations.

PCICONF gateway (VSEC based)
- The driver exposes a vendor-specific “PCICONF” access method via IOCTLs that proxy reads/writes of internal address spaces through a pair of registers (ADDR/DATA), typically located at offsets 0x58/0x5c.
- Relevant constants (reference/mstflint/mtcr_ul/mtcr_ul_com.c):
  - PCICONF_ADDR_OFF = 0x58, PCICONF_DATA_OFF = 0x5c
  - mtcr_pciconf_set_addr_space(...) selects the current address space (CAP9 VSEC).
  - mtcr_pciconf_rw(...) performs read/write of 32-bit words via IOCTLs.
  - Block ops (mread4_block_pciconf / mwrite4_block_pciconf) wrap buffered IOCTLs.
- Kernel IOCTLs (reference/mstflint/kernel/mst.h):
  - PCICONF_READ4 / PCICONF_WRITE4: 32-bit access via gateway.
  - PCICONF_READ4_BUFFER(_EX/_BC) / PCICONF_WRITE4_BUFFER: buffered operations.
  - PCICONF_VPD_READ4/WRITE4: PCI VPD access helpers.

Register access flows
- The Access Register mechanism is used to query/command on-device firmware subsystems (MGIR/MCQI/etc.).
- mtcr_ul implements multiple transports:
  - PCICONF (local gateway) – preferred for NICs.
  - MADs (SMP/GMP) – for IB/Ethernet out-of-band when applicable.
  - I2C – for certain EEPROM/board paths.
- FW components manager (reference/mstflint/fw_comps_mgr/*) issues a sequence of Access Register operations to gather:
  - Running/pending firmware versions (MCQI family).
  - Security and lifecycle state.
  - Component inventory and ROMs info (MGIR/MCQI variants).

Flash access
- For operations that require reading/writing the SPI flash, mstflint uses register sequences (MSPI/MCIA) to read blocks into host memory and reconstruct a complete image.
- The logic lives in mflash/ and mlxfwops/lib; flash operations are orchestrated above the raw register access layer.

Sysfs fallback
- When MST PCICONF is unavailable, mtcr_ul can operate via sysfs files (resource/config) using pwrite/pread to the vendor VSEC gateway registers, but this is less common in production setups.

Pointers for implementers
- Minimal viable subset for “query” over PCIe in Go:
  - Open device by BDF → locate /dev/mst/* for that BDF (via /sys/bus/pci/devices/... and /dev/mst symlinks), else resolve sysfs resource.
  - Implement PCICONF gateway: set address space; issue 32-bit reads/writes.
  - Implement a tiny Access Register client for MGIR/MCQI to fetch FW version, PSID, security bits.
- Full flash read support requires MSPI/MCIA command sequences and careful timing/error handling.

