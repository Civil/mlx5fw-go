mlx5fw-go

A fast, CLI-only tool for parsing, inspecting, and rebuilding Mellanox (NVIDIA ConnectX) firmware images (FS4 and FS5). It focuses on output parity with mstflint and provides JSON output for scripting.

Features
- Parse FS4 and FS5 firmware images; detect encrypted firmwares.
- List sections and verify CRCs; JSON output for sections and query.
- Extract sections (CRC automatically stripped) with metadata for reassembly.
- Reassemble full firmware images from extracted artifacts (JSON-first path).
- Replace a section in-place (size-preserving and relocation-aware paths).
- Clean, structured logs with verbose (-v) and quiet (-q) modes.

Install / Build
- From repository root: `go build -o mlx5fw-go ./cmd/mlx5fw-go`
- Show help: `./mlx5fw-go --help`

Quick Start
- List sections (human output): `./mlx5fw-go sections -f sample_firmwares/example.bin`
- Same, but JSON (for scripting): `./mlx5fw-go sections -f sample_firmwares/example.bin --json`
- Query firmware metadata (mstflint-like): `./mlx5fw-go query -f sample_firmwares/example.bin`
- Extract all sections (CRC removed; JSON metadata saved): `./mlx5fw-go extract -f sample_firmwares/example.bin -o out_dir`
- Reassemble from extracted artifacts: `./mlx5fw-go reassemble -i out_dir -o reassembled.bin --verify-crc`
- Replace a section: `./mlx5fw-go replace-section -f firmware.bin DBG_FW_INI -r new.ini -o firmware-new.bin`

Commands (summary)
- `sections`: List and verify sections; supports `--json`.
- `query`: Print firmware metadata; supports `--json`.
- `print-config`: Print DBG_FW_INI contents (decompressed when possible).
- `extract`: Extract sections and metadata to a directory (CRC removed by default).
- `reassemble`: Rebuild a firmware image from extracted JSON/BIN files.
- `replace-section`: Replace a section in an existing image and recalculate CRCs.

Global flags
- `-f, --file`: Firmware file path (required for most commands)
- `-v, --verbose`: Verbose (debug) logging
- `-q, --quiet`: Quiet mode (errors only)
- `--json`: JSON output where supported (sections, query)

Logging & Output
- Default log level is INFO; `-v` enables DEBUG; `-q` forces ERROR level.
- For encrypted firmwares, certain CRC checks are expected to fail and are logged at DEBUG.

Development
- Format and vet: `go fmt ./... && go vet ./...`
- Run tests: `go test ./... -v`
- Build the CLI: `go build -o mlx5fw-go ./cmd/mlx5fw-go`

Project structure highlights
- `cmd/mlx5fw-go`: Cobra CLI (wiring only; logic lives in `pkg`).
- `pkg/parser` + `pkg/parser/fs4`: Firmware reader, TOC parsing, CRC verification.
- `pkg/types`: Firmware structs and (un)marshaling via annotations.
- `pkg/reassemble`: JSON-first reassembly of firmware images.
- `pkg/section`: In-place replacement utilities.
- `pkg/cliutil`: Shared CLI helpers (logger, parser init, validation).
- `docs/mlx5fw-go/README.md`: Comprehensive developer guide and reference.

More Documentation
- See `docs/mlx5fw-go/README.md` for design decisions, parsing model, reassembly details, testing instructions, and tips.
