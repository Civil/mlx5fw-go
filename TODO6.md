Main design decisions:
 * End goal is a CLI tool, it must not have any API or metrics, nor microservices
 * Data representation must match mstflint by default, json output will be added later where it make sense
 * Focus on FS4 and FS5 firmware formats for now. It might be extended to FS3 later.
 * Performance doesn't matter. RAM ussage is irrelevant. Single-threaded code is enough. Firmwares guranteed to be smaller than 128MB uncompressed.
 * Project must use ideomatic Go 1.24
 * Project must not use any magic constants unless mstflint did the same
 * Project must interpret structures and parse sections by converting them to structs.
 * Use uber-go/zap for structured logging
 * ansel1/merry for error handling

Project structure:
 * cmd/mlx5fw-go/ - folder where main application is. It must contain only command line flags parsing, logging initialization and initialization of main parsing structures
 * pkg/types/ - folder where all the structs for firmware parsing must be placed. Marshaling and unmarshaling must be contained in types.
 * pkg/interfaces/ - folder where all interfaces (if any) must be placed
 * pkg/errs/ - folder where all the errors are defined


Available testing infrastucture:
 - `scripts/integration_tests/` - runs on firmwares from `sample_firmwares/` folder and on a random set of external firmwares. Tests can't be modified by anyone (score from 0 to 1, where 1 means 100% tests completed successfully)
 - `scripts/sample_tests/` - verbose tests on smaller set of firmwares. Can be inspected (test score from 0 to 1 in the end, 1 means 100% complete successfully)
 
To build the project: go build ./cmd/mlx5fw-go
To run tests (from project root): `scripts/sample_tests/sections.sh` and `scripts/sample_tests/strict-reassemble.sh`
To run mlx5fw-go with verbose logging: ./mlx5fw-go -v <...>
To run mstflint: mstflint -i <FILENAME> <command>
To run mstflint with debug logging: export FW_COMPS_DEBUG=1 && mstflint -i <FILENAME> <command>

Do not use stderr redirection with mstflint as it doesn't work with it.

You are tasked with investigating and fixing discrepency for the query command. For that there was a new test created: `./scripts/sample_tests/query.sh` - right now it shows a lot of differences. Potentially not all differences are bad - if there is just addition - it is probably OK, but if there is a difference - it is extremely likey it is a bug.

There are a lot of documentation available in docs/ about overall how project must work. You also can use gdb-mcp or pure gdb to investigate how mstflint works (it is available). mstflint have a debug mode that prints some of the bigger milestones on what is going on if you set env variable FW_COMPS_DEBUG to anything (and you need to unset if you don't need debug output).
