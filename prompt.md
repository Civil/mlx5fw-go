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

We are now in development phase. Here you will need to write code and start testing it using actual firmware and compare output with mstflint on each step
 - [ ] For cmd/mlx5fw-go implement sections printing command.
   - [ ] Implement and print sections content of the firmware in human-readable format
   - [ ] Write unit tests for the functionality (comprehensive test suite exists)
   - [ ] Add optional json output (fully implemented)
 - [ ] For cmd/mlx5fw-go implement metadata parsing
   - [ ] Parse firmware metadata (equivalent to `mstflint -i <FIRMWARE NAME> query full`) for unencrypted firmwares
   - [ ] Parse firmware metadata for all firmwares in `sample_firmwares/` with 100% match with data that mstflint shows
   - [ ] Make sure that all `sample_firmwares/test_*.sh` tests runs without errors
   - [ ] Make sure that 5 runs of `integration_tests/*.sh` runs with 100% completion rate
   - [ ] Write unit tests for the functionality
   - [ ] Add optional json output
   - [ ] Ensure that tests still completes at 100%
 - [ ] For cmd/mlx5fw-go implement equivalent to `mstflint -i <FILENAME> dc` command. That command ungzips content for `DBG_FW_INI` section if it is present.
   - [ ] Implement printing of the content (if it is in firmware).
   - [ ] Write custom test scripts that would compare mstflint and mlx5fw-go and check that output matches 100%
   - [ ] Write unit tests for that functionality
 - [ ] for cmd/mlx5fw-go implement verification commmand, equivalent to `mstflint -i <FILENAME> v`.
   - [ ] Implement that functionality.
   - [ ] Write custom test scripts that would compare results to what mstflint do. You must create few invalid files and few files where only some of the sections damaged. and check that verification pass and doesn't pass on same amount of firmwares.
   - [ ] Write unit tests for that functionality.
 - [ ] For cmd/mlx5fw-go implement extract functionality. That option must create a folder and dump content of all sections into files named after each section there.
   - [ ] Implement that functionality
   - [ ] Write custom test scripts that would extract sections using shell scripts, dd and information that `mstflint v` provides and check that output matches 100%
   - [ ] Write unit tests for that functionality
 - [ ] For cmd/mlx5fw-go implement section replacement functionality.
   - [ ] Implement functionality that would take `firmware_file`, `output_firmware`, `section_id` and `new_content_filename` and replace `section_id` in the firmware with new content and output it to a new file. All checksums if needed must be recalculated and updated.
   - [ ] Implement test scripts that would extract two firmwares of different versions and swap sections from one to another, and then call `mstflint -i <FILENAME> v` and ensure that firmware is still treated as bootable.
   - [ ] Implement unit tests
 - [ ] For cmd/mlx5fw-go implement section removal functionality
   - [ ] Implement command line option that would remove specified section ID from firmware and save it as a new file.
   - [ ] Write custom shell scripts that would remove a scetion and check that `mstflint -i <FILENAME> v` still shows that everything is fine.
