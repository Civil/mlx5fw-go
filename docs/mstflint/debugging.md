Debugging mstflint: Flags, Env, and GDB

Runtime debug knobs
- FW components manager debug prints:
  - Set FW_COMPS_DEBUG=1 to enable DPRINTF in fw_comps_mgr (reference/mstflint/fw_comps_mgr/fw_comps_mgr.h).
  - Example: `FW_COMPS_DEBUG=1 sudo -E mstflint -d 07:00.0 query full`
- Verbose CLI flags:
  - mstflint itself does not expose a global `-v` for internal driver/PCICONF traces; most low-level prints are compile-time or FW_COMPS_DEBUG controlled.

Kernel/MST device prep
- Create MST devices: `sudo mst start`, then `mst status`.
- Confirm /dev/mst/* entries exist for your device (e.g., /dev/mst/mtxxxx_pciconf0).

GDB usage
- Launch under gdb with elevated privileges:
  - `sudo -E gdb --args mstflint -d 07:00.0 query full`
- Suggested breakpoints:
  - `mtcr_pciconf_open`, `mtcr_pciconf_set_addr_space`, `mtcr_pciconf_mread4`, `mtcr_pciconf_mwrite4` (reference/mstflint/mtcr_ul/mtcr_ul_com.c)
  - Access Register send path: search for “AccessRegister” in mtcr_ul_com.c and break there.
  - FW components manager entry points in reference/mstflint/fw_comps_mgr/ (e.g., functions that gather MGIR/MCQI data).
- Handy gdb commands:
  - `set pagination off`
  - `catch syscall ioctl`
  - `bt full`
  - `finish` (step out) / `ni` (next instruction) for tight loops.

System call tracing
- `sudo -E strace -f -o /tmp/mstflint.strace mstflint -d 07:00.0 query full`
  - Search for PCICONF IOCTL codes from reference/mstflint/kernel/mst.h (PCICONF_*).

Build‑time debug (optional)
- Some DBG_PRINTF macros in mtcr_ul are compile-time gated; if you build a local mstflint with debug toggles enabled, you’ll see more low-level traces.
  - The reference tree here includes autotools config; a full local build is outside the scope of this doc but standard `./configure && make` applies.
