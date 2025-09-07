# mlx5fw-go Project Index and Developer Guide

This document provides a comprehensive overview of the mlx5fw-go repository: its goals, structure, key packages, data flows, commands, testing harnesses, and the main extension points for implementing new functionality.

Use this as your entry point to understand where things live and how to add or modify behavior safely.

## Goals and Scope
- CLI-only tool; no APIs, metrics, or services.
- Output parity with `mstflint` by default; provide `--json` for `query` and `sections`.
- Focus on FS4 and FS5 firmware formats first (FS3 may come later).
- Simplicity over performance; single-threaded; firmwares are <128MB.
- No magic constants unless matching `mstflint` (include source references in code comments).
- Parse on-disk structures into Go structs via annotations in `pkg/annotations`.
- Logging via `zap`; errors via `merry` with rich context.
- Reassembly must recompute CRCs; never rely on saved CRCs.

## Repository Map
- `cmd/mlx5fw-go`: Cobra CLI; flags, logging init, and command wiring only.
- `pkg/annotations`: Core reflection/annotation machinery for struct (un)marshaling.
- `pkg/crc`: CRC helpers (image/hardware CRC, helpers used by parser and sections).
- `pkg/errors`: Domain error types and helpers.
- `pkg/interfaces`: Interfaces for parser, sections, CRC handlers, and options builder.
- `pkg/parser`: Firmware reading, TOC parsing, CRC verification; FS4-specific logic under `pkg/parser/fs4`.
- `pkg/reassemble`: Firmware reassembler; JSON+binary input, metadata; output writer.
- `pkg/section`: Section replacement utilities (size-preserving and relocation-aware flows).
- `pkg/types`: Structs for firmware parsing and marshaling; includes `types/sections` for concrete section models and `types/extracted` for extraction metadata.
- `pkg/utils`: Misc utilities.
- `docs/`: Design notes, investigations, and this guide.
- `scripts/`: Sample and integration test harnesses.
- `sample_firmwares/`: Local test images (do not commit proprietary binaries).

## Command-Line Interface (CLI)
Binary name: `mlx5fw-go`

Common commands (see `cmd/mlx5fw-go/*.go`):
- `sections`: Parse and list sections; supports `--json` output and verbose `-v` logging.
- `query`: Produce `mstflint`-like query output; supports `--json`.
- `reassemble`: Rebuild firmware from extracted JSON/BIN files; recomputes CRC as needed.
- `print-config`: Print decompressed contents of `DBG_FW_INI` when present.
- `diff`: Compare two firmware images (raw and section-aware) with optional side‑by‑side diffs.

Run examples:
- Build: `go build -o mlx5fw-go ./cmd/mlx5fw-go`
- Help: `go run ./cmd/mlx5fw-go --help`
- Sections: `./mlx5fw-go sections -f sample_firmwares/example.bin -v`

### Diff Firmware Command

Compare two firmware images at both raw and section granularity. Sections are aligned by type and ordinal (sorted by offset), so simple address drift between images does not produce false “missing” reports.

#### Usage

```bash
# Basic: raw + sections summary
mlx5fw-go diff --a A.bin --b B.bin

# JSON summary (suppresses hexdumps/log noise)
mlx5fw-go diff --a A.bin --b B.bin --json > diff.json

# Focus on specific sections (comma-separated, case-insensitive names)
mlx5fw-go diff --a A.bin --b B.bin --sections --filter-types=IMAGE_INFO,PUBLIC_KEYS_4096

# Limit to an address window (start:end, hex or decimal; start <= off < end)
mlx5fw-go diff --a A.bin --b B.bin --sections --filter-offset=0x00600000:0x00700000

# Side-by-side diffs around first difference (hex for most, text for DBG_FW_INI)
mlx5fw-go diff --a A.bin --b B.bin --hexdump

# Raw-only spans with hexdumps (up to 3 spans, expanded context)
mlx5fw-go diff --a A.bin --b B.bin --raw --sections=false --max-spans=3 \
  --hexdump --hexdump-context=32 --hexdump-max-bytes=128
```

#### Flags

- `--a <file>`: first firmware file (required)
- `--b <file>`: second firmware file (required)
- `--raw` (default true): perform raw byte diff (SHA256 + first N spans)
- `--sections` (default true): perform section-aware diff
- `--json`: machine-readable JSON summary
- `--max-spans <N>`: cap number of raw diff spans (default 40)
- `--filter-types <list>`: comma-separated section names to include (case-insensitive)
- `--filter-offset <start:end>`: include sections whose offsets satisfy `start <= off < end` (hex like `0x...` or decimal)

Hex/text diffs (enabled by `--hexdump`):

- `--hexdump`: enable side-by-side diffs
  - Most sections: hex + ASCII for A|B; differing bytes highlighted
  - Special case `DBG_FW_INI`: both sides are gunzipped and a side-by-side text diff (line-level) is shown
- `--hexdump-context <N>`: bytes of context around the first diff (default 16)
- `--hexdump-max-bytes <N>`: cap printed bytes per diff region (default 256)
- `--hexdump-width <N>`: bytes per hex row (default 16)
- `--no-color`: disable ANSI colors

#### Output semantics

- Raw: prints sizes, SHA256, and the first N differing spans (offsets/length). With `--hexdump`, each span includes a focused side-by-side hex dump.
- Sections: each aligned pair prints either `OK` or `DIFF`. For diffs, it shows the first differing offset within the section (`first+off=0x...`). If the same section type appears a different number of times, the extra occurrences are reported as `MISSING in A/B`.
- Offsets are shown for both images when different, e.g., `@A:0x... B:0x...`.
- `DBG_FW_INI`: when `--hexdump` is enabled, content is gunzipped and a side-by-side text diff of the INI file is printed.

#### JSON summary (shape)

`--json` prints a single object with `raw` and `sections`. Each section item contains:

```jsonc
{
  "name": "IMAGE_INFO",
  "type": 16,
  "offset_a": 6682880,
  "offset_b": 6454784,
  "size_a": 1024,
  "size_b": 1024,
  "crc_type_a": "IN_ITOC_ENTRY",
  "crc_type_b": "IN_ITOC_ENTRY",
  "crc_algo_a": "SOFTWARE",
  "crc_algo_b": "SOFTWARE",
  "encrypted_a": false,
  "encrypted_b": false,
  "identical": false,
  "first_diff_offset": 10
}
```

## Parsing Architecture
Key flow (FS4): `pkg/parser/fs4/parser.go`
- `NewParser(reader, logger)`: Constructs parser with CRC calculator, TOC reader, and section factory.
- `Parse()`: High-level pipeline
  - `FindMagicPattern()` via `pkg/parser.FirmwareReader`.
  - `parseHWPointers()`: Reads hardware pointers into `types.FS4HWPointers`.
  - `parseITOC()` and `parseDTOC()`: Read headers, verify CRC, and enumerate sections via `TOCReader.ReadTOCSectionsNew`.
  - `parseBoot2()`, `parseToolsArea()`, `parseHashesTable()`: Section-specific discovery using HW pointers and standard locations.
  - `buildMetadata()`: Populate `types.FirmwareMetadata`.
- CRC: Uses `pkg/parser.CRCCalculator` for image/hardware CRC verification.
- Encryption: Parser toggles `isEncrypted` and adapts verification rules accordingly.

Supporting components:
- `pkg/parser/firmware_reader.go`: Reader over firmware blob; `FindMagicPattern`, `ReadSection`, `ReadAt`, `Size`.
- `pkg/parser/toc_reader.go`: Generic TOC parser that creates sections via the factory.
- `pkg/types/sections/factory.go`: `SectionFactory` that instantiates typed sections with CRC behavior and metadata.

## Types and Annotations
- Annotation core: `pkg/annotations` parses struct tags like `offset:"0x.."`, `endian:be`, and bitfield tags to marshal/unmarshal binary layouts.
- Key annotated structs:
  - `types.FS4HWPointers`: HW pointers block.
  - `types.ITOCHeader`: TOC header (ITOC/DTOC share format).
  - Canonical names: no “Annotated” suffix — all binary-annotated structs live directly under `pkg/types/*` (e.g., `HashesTableHeader`, `DigitalCertPtr`).
  - Many section models in `pkg/types/sections/*.go` (e.g., `image_info_section.go`, `hashes_table_section.go`).
- Section enums and helpers: `types.SectionType*` and `types.GetSectionTypeName`, `GetDTOCSectionTypeName`.

## Sections Model
- `pkg/interfaces/section.go`: Central interfaces for sections:
  - Base and Complete section interfaces (type, offset, size, CRCType, encryption/device flags, optional ITOC entry, raw data access, CRC handlers).
- `pkg/types/sections/*`: Concrete implementations for common section types: IMAGE_INFO, DEV_INFO, MFG_INFO, BOOT2, TOOLS_AREA, signatures, hashes table, debug, etc.
- `pkg/types/extracted`: Metadata used during extraction/reassembly (e.g., `extracted.SectionMetadata`).

## Reassembly
Entry: `pkg/reassemble/reassembler.go`
- Consumes an extraction directory containing `firmware_metadata.json` plus per-section `.json` and/or `.bin` files.
- JSON-first reconstruction path (`reassembler_json.go`), falls back to binary when necessary or when `BinaryOnly`.
- Ensures sizes, padding, and CRC consistency per section type. The main writer reassembles and emits the final firmware.

Related: `pkg/section/` contains alternate replacement logic focused on in-place replacement and relocation. Prefer the `pkg/reassemble` path for full rebuilds from extracted artifacts.

## Interfaces
- `pkg/interfaces/parser.go`: Legacy-style `FirmwareParser` interface; newer code uses FS4 parser directly.
- `pkg/interfaces/section_interfaces.go` and `section_options.go`: Options pattern for building sections with CRC/encryption/device flags and raw data.
- `pkg/interfaces/crc.go`: Abstraction for CRC handlers.

## Testing and Scripts
Go tests:
- Run all: `go test ./... -v`
- Focus while iterating: `go test ./pkg/parser -v` or the specific packages you change.

Scripted tests:
- Sample tests (verbose, inspectable): `scripts/sample_tests/`
  - `sections.sh`, `query.sh`, `reassemble.sh`, `strict-reassemble.sh`.
- Integration tests: `scripts/integration_tests/`
- Both suites report a score from 0.0 to 1.0; do not modify these scripts; assume they’re correct.

mstflint tips:
- Run: `mstflint -i <FILENAME> <command>`.
- Debug: `export FW_COMPS_DEBUG=1 && mstflint -i <FILENAME> <command>`.
- Avoid stderr redirection with `mstflint`.

## Extension Points (Where to Implement Changes)
- New section type
  - Add struct(s) under `pkg/types/sections/` for the binary layout (use annotations).
  - Extend `SectionFactory` to create the new type and wire CRC behavior.
  - Update `types` helpers if a new `SectionType*` is introduced and ensure name lookup functions cover it.
  - Add parser hooks if discovery needs HW pointers or special handling (`pkg/parser/fs4/parser.go`).

- Modify parsing of existing structure
  - Update annotated struct in `pkg/types` and its marshal/unmarshal methods.
  - If bitfields/endianness are involved, verify `hex_as_dec` or bit numbering conventions as needed.
  - Re-run `sections`/`query` and strict reassemble scripts.

- Change CRC rules
  - Check `pkg/parser/crc.go` and any section-specific CRC handlers.
  - Ensure both verification and reassembly paths reflect the change.

- Reassembly behavior
  - Implement in `pkg/reassemble/` (prefer JSON-first reconstruction), ensuring size alignment and CRC recomputation.
  - If adding new metadata, extend `types/extracted` and update the writer accordingly.

- CLI changes
  - Wire new flags/subcommands in `cmd/mlx5fw-go` only; keep logic in `pkg`.

## Common Pitfalls
- Bit numbering (MSB=0 vs LSB=0) mismatches cause subtle bugs; validate with specific firmwares.
- BCD-encoded time/version fields require `hex_as_dec` handling.
- Field names can be misleading (e.g., `FlashAddrDwords` may represent bytes); check call sites and tests.
- Always recompute CRCs during reassembly (in-section and ITOC/DTOC where applicable).
- Compare hex dumps of original vs reassembled images when debugging.

## Format Notes
- MFG_INFO size variability: Some firmwares use 0x100 (256) total bytes for MFG_INFO, others use 0x140 (320). We model the trailing region from offset 0x40 as a dynamic slice to accept both. CRC/size come from TOC entries (no embedded trailer). Reference: mstflint symbol `image_layout_mfg_info` (image_layout_layouts.h).

## Quick File Index (selected)
- CLI: `cmd/mlx5fw-go/main.go`, `sections.go`, `query.go`, `reassemble.go`, `print_config.go`.
- Parser core: `pkg/parser/firmware_reader.go`, `pkg/parser/toc_reader.go`, `pkg/parser/crc.go`.
- FS4 parser: `pkg/parser/fs4/parser.go` (+ `verification.go`, `itoc_crc.go`).
- Sections factory: `pkg/types/sections/factory.go` and concrete sections in the same dir.
- Reassembler: `pkg/reassemble/reassembler.go`, `reassembler_json.go`.
- Replacement (in-place): `pkg/section/replacer.go`.
- Errors: `pkg/errors/errors.go`.
- Interfaces: `pkg/interfaces/*.go`.

## Workflow Cheatsheet
1) Build: `go build -o mlx5fw-go ./cmd/mlx5fw-go`
2) Inspect: `./mlx5fw-go sections -f <fw.bin> -v` (or `--json`)
3) Query: `./mlx5fw-go query -f <fw.bin> --json`
4) Extract/modify -> Reassemble: follow `scripts/sample_tests/reassemble.sh` and `strict-reassemble.sh` patterns
5) Validate against `mstflint` output for parity

## How To: Add a New Section Type
Goal: add support for a new firmware section (e.g., FOO_BAR) so it parses, prints, participates in CRC checks, and reassembles.

High-level steps:
- Define the section type constant and name mapping in `pkg/types`.
- Implement the binary structure with annotations in `pkg/types/sections` and provide `Parse`/`Marshal` as needed.
- Wire the factory in `pkg/types/sections/factory.go` to return a typed section with appropriate CRC handler.
- If the section requires special discovery, add a hook in `pkg/parser/fs4/parser.go`.
- Add minimal tests and validate via `sections`/`query` and strict reassemble scripts.

1) Add the section type constant and names
- File: `pkg/types/common.go`
  - Add `SectionTypeFooBar = 0xXYZ` to the section type block.
- File: `pkg/types/section_names.go`
  - Update `GetSectionTypeName` and `GetSectionTypeByName` to include the human-readable name.
  - If the section is DTOC-only, also consider `GetDTOCSectionTypeName`.

2) Define the annotated struct and section wrapper
- File: `pkg/types/sections/foo_bar_section.go`
  - Example annotated struct (adjust offsets/types/endian per mstflint/spec):
    
    type FooBar struct {
        Version   uint16 `offset:"0x0,endian:be"`
        Flags     uint16 `offset:"0x2,endian:be"`
        PayloadSz uint32 `offset:"0x4,endian:be"`
        // ... additional fields ...
    }
    
    // Provide Marshal/Unmarshal using annotations package if you need JSON-driven reconstruction:
    // func (s *FooBar) Unmarshal(data []byte) error { ... }
    // func (s *FooBar) Marshal() ([]byte, error) { ... }

  - Provide a section type that implements `interfaces.CompleteSectionInterface` by wrapping the base section and parsing raw data:
    
    func NewFooBarSection(base interfaces.SectionInterface) interfaces.CompleteSectionInterface {
        return &FooBarSection{Base: base}
    }
    
    type FooBarSection struct {
        interfaces.BaseSectionImpl // or keep a Base field; follow local patterns
        Model FooBar
    }
    
    func (s *FooBarSection) TypeName() string { return types.GetSectionTypeName(s.Type()) }
    
    func (s *FooBarSection) Parse(data []byte) error {
        // Use annotated Unmarshal or manual parse for now
        // Populate s.Model, keep raw with s.SetRawData(data) if needed
        return s.Model.Unmarshal(data)
    }

3) Wire in the factory
- File: `pkg/types/sections/factory.go`
  - Add a case mapping `SectionTypeFooBar` to `NewFooBarSection(base)`.
  - If CRC behavior differs, extend `types.GetSectionCRCAlgorithm` in `pkg/types/crc_algorithm.go` and adjust handler selection.

4) Parser hooks (only if special discovery is needed)
- File: `pkg/parser/fs4/parser.go`
  - If the section is not surfaced by ITOC/DTOC or needs HW pointer discovery, add a `parseXxx()` helper similar to `parseBoot2`/`parseToolsArea` and call it from `Parse()`.

5) Validate
- Build and run: `./mlx5fw-go sections -f <fw.bin> -v` and with `--json`.
- Reassemble through `scripts/sample_tests/strict-reassemble.sh`; ensure CRCs recompute and the output matches original (when unmodified).
- Add targeted tests under the appropriate package if feasible.

Notes:
- Avoid magic constants unless mirrored from `mstflint`; add a short source reference in comments.
- Use annotations for binary struct layouts; prefer updating annotated models over ad-hoc byte slicing.
- Ensure sizes and alignment (dword) and remember many section CRCs cover all but the last dword.

## Package Graph (internal packages)
Edges show package -> internal import. This reflects current state and helps locate extension points.

```
github.com/Civil/mlx5fw-go/cmd/mlx5fw-go -> github.com/Civil/mlx5fw-go/pkg/extract
github.com/Civil/mlx5fw-go/cmd/mlx5fw-go -> github.com/Civil/mlx5fw-go/pkg/interfaces
github.com/Civil/mlx5fw-go/cmd/mlx5fw-go -> github.com/Civil/mlx5fw-go/pkg/parser
github.com/Civil/mlx5fw-go/cmd/mlx5fw-go -> github.com/Civil/mlx5fw-go/pkg/parser/fs4
github.com/Civil/mlx5fw-go/cmd/mlx5fw-go -> github.com/Civil/mlx5fw-go/pkg/reassemble
github.com/Civil/mlx5fw-go/cmd/mlx5fw-go -> github.com/Civil/mlx5fw-go/pkg/section
github.com/Civil/mlx5fw-go/cmd/mlx5fw-go -> github.com/Civil/mlx5fw-go/pkg/types
github.com/Civil/mlx5fw-go/cmd/mlx5fw-go -> github.com/Civil/mlx5fw-go/pkg/utils
github.com/Civil/mlx5fw-go/pkg/crc -> github.com/Civil/mlx5fw-go/pkg/errors
github.com/Civil/mlx5fw-go/pkg/crc -> github.com/Civil/mlx5fw-go/pkg/interfaces
github.com/Civil/mlx5fw-go/pkg/crc -> github.com/Civil/mlx5fw-go/pkg/parser
github.com/Civil/mlx5fw-go/pkg/crc -> github.com/Civil/mlx5fw-go/pkg/types
github.com/Civil/mlx5fw-go/pkg/extract -> github.com/Civil/mlx5fw-go/pkg/interfaces
github.com/Civil/mlx5fw-go/pkg/extract -> github.com/Civil/mlx5fw-go/pkg/parser/fs4
github.com/Civil/mlx5fw-go/pkg/extract -> github.com/Civil/mlx5fw-go/pkg/types
github.com/Civil/mlx5fw-go/pkg/extract -> github.com/Civil/mlx5fw-go/pkg/types/extracted
github.com/Civil/mlx5fw-go/pkg/extract -> github.com/Civil/mlx5fw-go/pkg/types/sections
github.com/Civil/mlx5fw-go/pkg/interfaces -> github.com/Civil/mlx5fw-go/pkg/types
github.com/Civil/mlx5fw-go/pkg/parser/fs4 -> github.com/Civil/mlx5fw-go/pkg/errors
github.com/Civil/mlx5fw-go/pkg/parser/fs4 -> github.com/Civil/mlx5fw-go/pkg/interfaces
github.com/Civil/mlx5fw-go/pkg/parser/fs4 -> github.com/Civil/mlx5fw-go/pkg/parser
github.com/Civil/mlx5fw-go/pkg/parser/fs4 -> github.com/Civil/mlx5fw-go/pkg/types
github.com/Civil/mlx5fw-go/pkg/parser/fs4 -> github.com/Civil/mlx5fw-go/pkg/types/sections
github.com/Civil/mlx5fw-go/pkg/parser -> github.com/Civil/mlx5fw-go/pkg/errors
github.com/Civil/mlx5fw-go/pkg/parser -> github.com/Civil/mlx5fw-go/pkg/interfaces
github.com/Civil/mlx5fw-go/pkg/parser -> github.com/Civil/mlx5fw-go/pkg/types
github.com/Civil/mlx5fw-go/pkg/reassemble -> github.com/Civil/mlx5fw-go/pkg/annotations
github.com/Civil/mlx5fw-go/pkg/reassemble -> github.com/Civil/mlx5fw-go/pkg/parser
github.com/Civil/mlx5fw-go/pkg/reassemble -> github.com/Civil/mlx5fw-go/pkg/types
github.com/Civil/mlx5fw-go/pkg/reassemble -> github.com/Civil/mlx5fw-go/pkg/types/extracted
github.com/Civil/mlx5fw-go/pkg/section -> github.com/Civil/mlx5fw-go/pkg/errors
github.com/Civil/mlx5fw-go/pkg/section -> github.com/Civil/mlx5fw-go/pkg/interfaces
github.com/Civil/mlx5fw-go/pkg/section -> github.com/Civil/mlx5fw-go/pkg/parser
github.com/Civil/mlx5fw-go/pkg/section -> github.com/Civil/mlx5fw-go/pkg/parser/fs4
github.com/Civil/mlx5fw-go/pkg/section -> github.com/Civil/mlx5fw-go/pkg/types
github.com/Civil/mlx5fw-go/pkg/types/extracted -> github.com/Civil/mlx5fw-go/pkg/interfaces
github.com/Civil/mlx5fw-go/pkg/types/extracted -> github.com/Civil/mlx5fw-go/pkg/types
github.com/Civil/mlx5fw-go/pkg/types -> github.com/Civil/mlx5fw-go/pkg/annotations
github.com/Civil/mlx5fw-go/pkg/types -> github.com/Civil/mlx5fw-go/pkg/errors
github.com/Civil/mlx5fw-go/pkg/types/sections -> github.com/Civil/mlx5fw-go/pkg/crc
github.com/Civil/mlx5fw-go/pkg/types/sections -> github.com/Civil/mlx5fw-go/pkg/errors
github.com/Civil/mlx5fw-go/pkg/types/sections -> github.com/Civil/mlx5fw-go/pkg/interfaces
github.com/Civil/mlx5fw-go/pkg/types/sections -> github.com/Civil/mlx5fw-go/pkg/parser
github.com/Civil/mlx5fw-go/pkg/types/sections -> github.com/Civil/mlx5fw-go/pkg/types
github.com/Civil/mlx5fw-go/pkg/utils -> github.com/Civil/mlx5fw-go/pkg/interfaces
github.com/Civil/mlx5fw-go/pkg/utils -> github.com/Civil/mlx5fw-go/pkg/types
```

Layered view (simplified):
- CLI → parser, types, sections, reassemble, section (replace), utils, extract
- Parser (core + fs4) → interfaces, types, sections, errs/errors, crc (via factory)
- Sections → interfaces, types, crc, parser (helpers)
- Reassemble → annotations, parser, types, extracted
- Types → annotations, errs; Sections/Extracted under types reference back to interfaces/types
