Main design decisions:
 * End goal is a CLI tool, it must not have any API or metrics, nor microservices
 * Data representation must match mstflint by default, json output must be present for query and sections command.
 * Focus on FS4 and FS5 firmware formats for now. It might be extended to FS3 later.
 * Performance doesn't matter. RAM ussage is irrelevant. Single-threaded code is enough. Firmwares guranteed to be smaller than 128MB uncompressed.
 * Project must use ideomatic Go 1.24
 * Project must not use any magic constants unless mstflint did the same, if you encounter such thing you must put a reference to mstflint source code that do the same.
 * Project must interpret structures and parse sections by converting them to structs using annotations that are implemented in pkg
 * Use uber-go/zap for structured logging
 * ansel1/merry for error handling
 * For all reassemble methods, you must always recompute CRCs and never rely on saving them before hands, as while strict-reassembly test doesn't modify files, but one of the main use cases for the software is to be able to easiy extract firmware, modify it and reassmble back.

Project structure:
 * cmd/mlx5fw-go/ - folder where main application is. It must contain only command line flags parsing, logging initialization and initialization of main parsing structures
 * pkg/types/ - folder where all the structs for firmware parsing must be placed. Marshaling and unmarshaling must be contained in types.
 * pkg/interfaces/ - folder where all interfaces (if any) must be placed
 * pkg/errs/ - folder where all the errors are defined

Available testing infrastucture:
 - `scripts/integration_tests/` - runs on firmwares from `sample_firmwares/` folder and on a random set of external firmwares. Tests can't be modified by anyone (score from 0 to 1, where 1 means 100% tests completed successfully)
 - `scripts/sample_tests/` - verbose tests on smaller set of firmwares. Can be inspected (test score from 0 to 1 in the end, 1 means 100% complete successfully)
 
 For all the tests, you must assume they are implemented correctly and never modify them. If it is doesn't use some flag - it is for a reason.
 
To build the project: go build ./cmd/mlx5fw-go
To run tests (from project root): `scripts/sample_tests/sections.sh` and `scripts/sample_tests/strict-reassemble.sh`
To run mlx5fw-go with verbose logging: ./mlx5fw-go -v <...>
To run mstflint: mstflint -i <FILENAME> <command>
To run mstflint with debug logging: export FW_COMPS_DEBUG=1 && mstflint -i <FILENAME> <command>

Do not use stderr redirection with mstflint as it doesn't work with it.

Previous work and some notes were logged in .claude/sessions/2025-07-28-0000-Complete-Annotation-Migration.md - please read that and especially tips and additional notes.

After the last attempt to improve the code qualtiy, something is broken with `mlx5fw-go sections` command, and therefore `scripts/sample_tests/strict-reassemble.sh` doesn't work anyore. Please identify what the problem is and fix it. Previous attempts successfuly fixed sections.sh (there was a problem in pkg/annotations) and likely problem with reassemble is somewhere near (annotations doesn't correctly marshal bit data that cross byte boundry)

Previously working code is stored in `reference/mlx5fw-go.wrk/` and the binary is available as `reference/mlx5fw-go.wrk/mlx5fw-go`
