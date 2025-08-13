mstflint PCIe Access: Overview and Quickstart

This directory documents how mstflint (from NVIDIA/Mellanox MFT) talks to PCIe devices and how to enable useful debug output while investigating device interaction. It serves as preparation for adding native PCIe device support to mlx5fw-go.

What mstflint does (high level)
- Accepts a device selector (BDF like 0000:07:00.0, IB name, or /dev/mst path).
- Opens a PCICONF “gateway” for register access via an MST kernel driver device (preferred), or directly via sysfs resource/config files.
- Performs register-level queries using the Access Register mechanism (MGIR/MCQI/etc.) to collect FW component information (running/pending FW version, security attributes), and (when needed) accesses MSPI/MCIA to read flash contents for image operations.

Key components in the source tree
- reference/mstflint/mtcr_ul/mtcr_ul_com.c: user‑level abstraction for memory/config/PCICONF access.
- reference/mstflint/kernel/mst.h: IOCTLs for PCICONF access via /dev/mst/*.
- reference/mstflint/fw_comps_mgr/*: FW components (query) logic; honors FW_COMPS_DEBUG.
- reference/mstflint/reg_access/* and tools_layouts/*: register and layout definitions (e.g., MGIR/MCQI/MSPI/MCIA).
- reference/mstflint/flint/*: CLI wiring for mstflint.

Quickstart: Investigating a physical device
1) Ensure MST devices exist
   - sudo mst start
   - mst status

2) List devices and pick one (BDF or /dev/mst path)
   - mstflint -l
   - Example device: 0000:07:00.0

3) Run a full query (with extra FW components debug prints)
   - FW_COMPS_DEBUG=1 sudo -E mstflint -d 07:00.0 query full

4) Optional: trace with gdb
   - sudo -E gdb --args mstflint -d 07:00.0 query full
   - Suggested breakpoints:
     - fw_comps_mgr::Query (C++) or the exported query function in fw_comps_mgr
     - mtcr_pciconf_open, mtcr_pciconf_set_addr_space, mtcr_pciconf_mread4/mwrite4 (mtcr_ul_com.c)
     - AccessRegister send paths (search for “AccessRegister” in mtcr_ul_com.c)

5) Useful captures
   - strace: sudo -E strace -f -o /tmp/mstflint.strace mstflint -d 07:00.0 query full
   - Keep the FW_COMPS_DEBUG=1 output alongside to correlate register flows.

See pcie_access.md for a deeper description of the PCICONF gateway and register flows. See debugging.md for more debug knobs.

