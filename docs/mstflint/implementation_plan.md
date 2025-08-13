Implementation Plan: PCIe Device Support in mlx5fw-go

Goal
Add native PCIe device support (query and, later, image read/write) to mlx5fw-go, following mstflint’s PCICONF gateway approach.

Principles
- Do not depend on shelling out to mstflint in steady state.
- Reuse mstflint’s model: /dev/mst PCICONF preferred, sysfs fallback when possible.
- Start small (device query) and iterate towards full flash read.

Phases
1) Device abstraction and discovery
   - Add `pkg/dev/pcie` with a minimal interface:
     - Open(BDF or path) → Handle
     - Read32(addrSpace, offset) / Write32(addrSpace, offset)
     - ReadBlock/WriteBlock helpers
   - Discovery:
     - Accept `-d <BDF>` on CLI, resolve to `/sys/bus/pci/devices/<BDF>`.
     - Prefer mapping to a /dev/mst PCICONF device for that BDF (via `/dev/mst` symlinks or scanning MST status).
     - Fallback: sysfs `resource`/`config` pwrite/pread to the PCICONF gateway registers (ADDR/DATA offsets 0x58/0x5c).
   - Implementation details:
     - Implement an MST backend via cgo (ioctls from reference/mstflint/kernel/mst.h): PCICONF_READ4/WRITE4 and buffered ops.
     - Implement a (limited) sysfs backend for environments without MST.

2) Query via Access Register (read‑only)
   - Implement minimal Access Register client on top of the PCICONF backend to send MGIR/MCQI commands and parse responses.
   - Populate a Go struct mirroring the fields needed for `query` (running/pending FW versions, PSID, product version, security flags).
   - Wire a new CLI path: `mlx5fw-go query -d <BDF> [--json]` that uses the device backend when `-d` is present.

3) Flash read (image dump)
   - Implement MSPI/MCIA register sequences (as in mstflint mflash/ and mlxfwops/lib) to read the complete image from flash via PCICONF gateway.
   - Integrate with the existing parser to allow `sections` / `query` against a live device without an external dump.
   - Add `mlx5fw-go dump -d <BDF> -o image.bin` (optional) for explicit dumps.

4) Flash write (burn) – optional and gated
   - Only after robust read/verify exists.
   - Implement write/erase sequences; enforce security checks and explicit `--i-know-what-im-doing` gates.

Testing strategy
- Phase 1/2:
  - Compare `mlx5fw-go query -d` against `mstflint -d ... query full` output on multiple NICs.
  - Use FW_COMPS_DEBUG=1 to compare register flows if needed.
- Phase 3:
  - Compare `dump` output against mstflint dump (hex diff, CRCs) and strict‑reassemble tests.

Risks & mitigations
- Kernel dependencies: MST driver presence simplifies IOCTLs; document required modules and `mst start`.
- Sysfs fallback may not work universally (capabilities vary across devices/kernels).
- Timing and retry semantics for Access Register and SPI flows; mirror mstflint retry logic and errors.

Deliverables
- `pkg/dev/pcie`: MST + sysfs backends, with tests where practical.
- CLI flags: `-d <BDF>` for query/sections/extract; `dump` (optional).
- Documentation updates and troubleshooting guides.

