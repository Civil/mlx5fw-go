# Repository Guidelines

See also: comprehensive project index and developer guide in `docs/mlx5fw-go/README.md`.

## Design Decisions
- CLI-only: no APIs, metrics, or microservices.
- Output parity: match `mstflint` by default; provide `--json` for `query` and `sections` commands.
- Target formats: focus on FS4 and FS5 first; FS3 may come later.
- Simplicity over performance: single-threaded; RAM usage not a concern for <128MB firmwares.
- No magic constants: if unavoidable, add a reference to equivalent `mstflint` source.
- Parsing model: represent on-disk structures as Go structs using annotations implemented in `pkg`.
- Logging and errors: use `zap` and `merry` for structured logs and rich errors.
- Reassembly rule: always recompute CRCs during reassemble; never rely on previously saved CRCs.

## Project Structure & Module Organization
- `cmd/mlx5fw-go`: CLI entrypoint and commands (Cobra). Primary binary name: `mlx5fw-go`.
- `pkg/`: Core libraries (parsing, CRC, reassembly, types, utils, errors, interfaces).
- `pkg/types/`: All firmware structs and (un)marshaling via annotations.
- `pkg/interfaces/`: Shared interfaces across packages.
- `pkg/errors/`: Centralized error definitions.
- `docs/`: Design notes, investigations, and reference analyses.
- `scripts/`: Helper scripts and sample/integration tests.
- `sample_firmwares/`: Example firmware for local testing. Avoid committing proprietary images.

## Build, Test, and Development Commands
- Build CLI (repo default): `go build -o mlx5fw-go ./cmd/mlx5fw-go`
- Alt build (common local): `go build ./cmd/mlx5fw-go`
- Run CLI: `go run ./cmd/mlx5fw-go --help`
- Example: `./mlx5fw-go sections -f sample_firmwares/example.bin -v`
- All tests: `go test ./... -v`
- Coverage: `go test ./... -cover`
- Lint-ish checks: `go fmt ./... && go vet ./...`

### Test Scripts & Integration Harness
- Sample tests: `scripts/sample_tests/` (verbose; human-inspectable). Key scripts:
  - `scripts/sample_tests/sections.sh`, `strict-reassemble.sh`, `query.sh`, `reassemble.sh`.
- Integration tests: `scripts/integration_tests/` run on local and external firmwares.
- Scoring: both suites emit a score from 0 to 1; 1.0 means 100% passing.
- Do not modify these scripts; assume they’re correct. If a flag isn’t used, it’s intentional.

## Coding Style & Naming Conventions
- Go 1.22 modules; write idiomatic Go and run `go fmt` before pushing.
- Packages: lowercase, no underscores (`pkg/parser`, `pkg/section`).
- Exported API: PascalCase (`ParseFirmware`), unexported: camelCase (`parseHeader`).
- Files: descriptive lowercase; tests end with `_test.go`.
- Logging: use `zap` via provided helpers; prefer structured fields.
- Errors: prefer `merry` for wrapping/context and typed errors from `pkg/errors`.
- Constants: avoid magic numbers; when mirroring firmware specs or `mstflint`, add a short source reference in code comments.

## Firmware Parsing & Reassembly
- Struct annotations: implement binary layout with annotations in `pkg`; avoid ad-hoc byte slicing in business logic.
- Data parity: default text output should match `mstflint`; ensure `--json` is available for `query` and `sections`.
- Bit fields: verify definitions do not overlap; be explicit about MSB=0 vs LSB=0 conventions.
- BCD fields: some time/version fields are BCD-encoded; use `hex_as_dec` when appropriate.
- Reassembly: always recompute CRCs; strict-reassemble tests validate this behavior.

## Testing Guidelines
- Frameworks: standard `testing` + `testify` for assertions.
- Place unit tests alongside source as `*_test.go`; name tests `TestXxx` and keep them independent.
- Aim to add tests for new parsing/reassembly logic; no strict coverage threshold, but keep or raise current coverage.
- Run targeted packages while iterating: `go test ./pkg/parser -v`.
- Use sample/integration scripts for end-to-end checks; strict-reassemble is the ultimate validation.

## Commit & Pull Request Guidelines
- History shows informal “wip” commits; please use Conventional Commits going forward: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`.
- Commits: imperative mood, concise scope, reference issues when relevant.
- PRs: include a clear description, usage examples (commands/output), linked issues, and notes on testing/risks. Screenshots not required for CLI, but sample command output helps.

## Security & Configuration Tips
- Do not commit proprietary firmware binaries; use `sample_firmwares/` for sharable artifacts.
- Prefer `--json` for machine-readable logs when scripting; avoid leaking paths or sensitive data in logs.
- Large files: add to `.gitignore` if generated (extraction outputs, temporary binaries).

## Tooling & Reference (mstflint)
- Run `mlx5fw-go` with verbose logs: `./mlx5fw-go -v <...>`.
- Run `mstflint`: `mstflint -i <FILENAME> <command>`.
- Debug `mstflint`: `export FW_COMPS_DEBUG=1 && mstflint -i <FILENAME> <command>`.
- Note: avoid stderr redirection with `mstflint` as it can misbehave.

## Developer Tips
1. Compare hex dumps of original vs reassembled files when debugging parsing.
2. Treat strict-reassemble as the ultimate validation; it catches subtle issues.
3. Double-check bit field definitions in packed structs; avoid overlaps.
4. Use `hex_as_dec` for BCD-encoded fields where required.
5. Rebuild after annotation changes: `go build ./cmd/mlx5fw-go/`.
6. Start with specific firmware samples before full suites.
7. Test both query and reassembly; they hit different paths.
8. Be explicit about bit numbering (MSB=0 vs LSB=0).
9. Exercise multiple firmware families (FS4, FS5) to reveal edge cases.
10. Enable debug logging to trace bit/field parsing.

## Lessons Learned
1. Bit numbering conventions are critical; mismatches cause subtle bugs.
2. BCD time fields appear in several structures (e.g., IMAGE_INFO hour/min/sec).
3. Field names can be misleading (e.g., `FlashAddrDwords` may contain byte addresses).
4. Systematic, case-driven debugging beats broad changes.
5. Field-by-field validation surfaces firmware-specific issues (e.g., Product Version).
6. Comprehensive testing (incl. strict-reassemble) reveals reconstruction bugs.
7. Prefer fixing root-cause parsing issues over reverting changes.
