Main design decisions:
 * End goal is a CLI tool, it must not have any API or metrics, nor microservices
 * Data representation must match mstflint by default, json output will be added later where it make sense
 * Focus on FS4 and FS5 firmware formats for now. It might be extended to FS3 later.
 * Performance doesn't matter. RAM ussage is irrelevant. Single-threaded code is enough. Firmwares guranteed to be smaller than 128MB uncompressed.
 * Project must use ideomatic Go 1.24
 * Project must not use any magic constants unless mstflint did the same
 * Project must interpret structures and parse sections by converting them to structs. `ghostiam/binstruct` must be used for struct unmarshaling.
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

Use `reference/mstflint/flint/mstflint` for mstflint executable.
If you need to trace execution you need to set env variable FW_COMPS_DEBUG to 1 before running mstflint.

Previous work has been moved to docs/ folder and there you can find all that was found. Use that information from now on as guidelines on how to continue and what is the success criterias.

Right now tool produces good enough output for ConnectX-5 and Bluefield-1, but for some firmwares it is still wrong. For example `sample_firmwares/fw-ConnectX7-rel-28_33_0800.bin` mlx5fw-go doesn't specify UID and description is truncated (you must use full padding as potential description field). For `sample_firmwares/900-9D3B6-00CN-A_Ax_MT_0000000883_rel-32_45.1020.bin` security attributes are missing while they should be. Please investigate with gdb-mcp what mstflint is doing and fix discrepencies in parsing.
