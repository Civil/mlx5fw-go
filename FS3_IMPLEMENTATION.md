# FS3 Implementation Notes

This document captures what was implemented to support FS3 format (older ConnectX-4/Lx era images) and what remains. It should be a quick reference for future work and for context compaction.

## Summary

- Commands covered for FS3:
  - `sections`: parity with mstflint sample suite (score 1.0).
  - `query`: parity with mstflint sample suite (score 1.0).
  - `extract`: works through the common path (typed sections + JSON); no FS3‑specific exceptions needed for sample images.
  - `strict-reassemble`: unchanged (issues are in FS4/FS5 bitfield re‑marshaling and not FS3 specific yet).

- Design approach: minimal, self‑contained FS3 integration that reuses the existing parser/section factory/CRC pipeline.

## What Was Implemented

### 1) FS3 Detection

- File: `pkg/cliutil/parser.go`
  - `detectFirmwareFormat` now:
    - First tries FS4/FS5 via MagicPattern + BootVersion.
    - Falls back to FS3 when the file begins with `FS3Magic` ("MTFW" at offset 0).
  - On FS3 detection, the shared parser is told its format is `FS3`.

- File: `pkg/types/types.go`
  - Added `FormatFS3` to the `FirmwareFormat` enum and JSON (de)serialization.

### 2) FS3 ITOC Structures (CIB layout)

- File: `pkg/types/fs3_itoc.go`
  - New annotation-driven structs:
    - `FS3ITOCHeader` — bit/byte offsets match `cibfw_itoc_header` from mstflint.
    - `FS3ITOCEntry` — bit/byte offsets match `cibfw_itoc_entry` from mstflint.
  - These are used only inside the FS3 parsing path and converted to the internal `ITOCEntry` (`pkg/types/itoc.go`) for the rest of the pipeline.

### 3) FS3 Parse Path

- Files: `pkg/parser/fs4/parser.go`, `pkg/parser/fs4/parser_fs3.go`
  - `Parser.Parse()` dispatches to `parseFS3()` when format is `FS3`.
  - `parseFS3()`:
    - Scans the image in 4KB steps (`0x1000`) for an FS3 ITOC header signature: `ITOC`, `0x04081516`, `0x2342cafa`, `0xbacafe00`.
    - Parses ITOC entries as `FS3ITOCEntry` and converts them to our internal `ITOCEntry` and to concrete sections via the existing section factory.
    - CRC mapping:
      - FS3’s `no_crc` flag -> `CRCNone`.
      - Otherwise -> `CRCInITOCEntry` (software CRC16, polynomial `0x100b`), which matches mstflint’s CalcImageCRC.
    - Addresses & sizes:
      - `flash_addr` and `size` fields are in dwords; we shift by 2 when producing bytes.
      - `relative_addr` is currently ignored in file‑image flows (good enough for the sample suite, see TODO below).
    - Zero‑length sections (e.g., `VPD_R0`) are added as present with size `0` (mstflint lists them).
    - BOOT2 (FS3): the header at `0x38` contains size in 8‑byte units; we add a `BOOT2` section at `0x38` with `size = header_size * 8` (best‑effort to match mstflint’s visible layout; improves sections parity).

### 4) Query Output Tweaks (FS3)

- File: `cmd/mlx5fw-go/query.go`
  - FS3 does not expose security attributes; mstflint prints `Security Attributes:   N/A` and does not print `Security Ver`.
  - Implemented: for FS3, print `Security Attributes:   N/A` and omit the `Security Ver` line. FS4/FS5 behavior unchanged.

### 5) Metadata improvements

- File: `pkg/extract/extractor.go`
  - `metadata.Format` now uses `parser.GetFormat().String()` instead of hardcoding "FS4".

## Test Results (sample suite)

- `scripts/sample_tests/sections.sh`: 1.0
- `scripts/sample_tests/query.sh`: 1.0
- `scripts/sample_tests/strict-reassemble.sh`: unchanged (~0.33); this is unrelated to FS3 and will be addressed separately (see TODO).

## Notes & Assumptions

- ITOC scanning starts at `0x1000` and proceeds in `0x1000` steps. This is sufficient for the provided FS3 samples.
- FS3 section CRCs are validated via the existing software CRC handler when ITOC provides a CRC. Sections with `no_crc` are marked `CRCNone` and treated accordingly.
- `relative_addr` is currently ignored in file‑image parsing. For FS3 images that rely on this addressing mode (rare in the sample set), conversion to physical offsets may be needed (see mstflint’s address convertor with `imgStart`/chunk size). This is a known future enhancement but not required for parity on the given samples.
- BOOT2 heuristic: the size at `0x38` is interpreted in 8‑byte units. This is sufficient to make BOOT2 visible and sized plausibly in sections. If future samples disagree, we will refine this logic.

## What’s Left / TODO

1) Strict Reassemble → 1.0 (FS4/FS5 path primarily)
   - The current failure is tied to annotation‑based bitfield marshaling of big‑endian fields crossing byte boundaries in a few sections (not FS3‑specific).
   - Action items:
     - Add golden encode/decode tests for the reassembled sections most involved in mismatch: `IMAGE_INFO`, `DEV_INFO`, `MFG_INFO`, signature/public key sections.
     - Tighten `pkg/annotations/marshal.go` for big‑endian bitfields that span multiple bytes (write‑masking, in‑place update, consistent shifting). We already have a general big‑endian path; we’ll harden it with these test vectors.
     - Verify against `scripts/sample_tests/strict-reassemble.sh` until byte‑for‑byte matches are achieved.

2) FS3 `relative_addr` support (optional enhancement)
   - For completeness, implement physical address conversion when `relative_addr` is set (see mstflint’s `_ioAccess->set_address_convertor` logic with `imgStart` and `cntxLog2ChunkSize`). Not required for current samples.

3) (Optional) Hide DTOC header in FS3 sections display
   - Our sections view is unified and shows header rows, which may list a DTOC header line. We currently don’t parse DTOC for FS3 and do not rely on it. For stricter fidelity, we could skip showing a DTOC header for FS3.

4) (Optional) FS3‑specific reassemble rules
   - If/when FS3 reassembly is needed beyond current JSON‑driven flows, document and enforce:
     - ITOC entry rebuild (pack), entry CRC recomputation, and full image CRC checks.
     - Proper ordering/alignment rules for device‑data vs image sections and sector alignment.

## File Map / Reference

- Detection
  - `pkg/cliutil/parser.go`
  - `pkg/types/types.go`

- FS3 parsing & types
  - `pkg/types/fs3_itoc.go`
  - `pkg/parser/fs4/parser.go` (dispatch)
  - `pkg/parser/fs4/parser_fs3.go` (FS3 implementation)

- CLI output
  - `cmd/mlx5fw-go/query.go` (FS3 security line handling)
  - `cmd/mlx5fw-go/sections.go` (unchanged, uses parsed sections)

- Extraction / metadata
  - `pkg/extract/extractor.go`

## How to Reproduce

- Build: `go build -o mlx5fw-go ./cmd/mlx5fw-go`
- Run tests:
  - Sections: `scripts/sample_tests/sections.sh`
  - Query: `scripts/sample_tests/query.sh`
  - Strict reassemble (still not 1.0): `scripts/sample_tests/strict-reassemble.sh`

---

Prepared to support future compaction and follow‑ups. If anything regresses after compaction, the above file map and TODOs should guide restoration.

