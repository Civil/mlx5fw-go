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

You are tasked to improve extract command and implement reassemble and later on custom diff commands. Here is the breakdown of the task into smaller tasks:
- [ ] Right now test doesn't pass for `scripts/sample_tests/strict-reassemble.sh` - for a few encrypted firmwares it doesn't reassemble back correctly (firmware validates but is not perfect). You need to fix that.
- [ ] With current extraction, there are several improvements that should be made for gaps extraction:
  - [ ] If gap consist only of 0xFF or 0x00 - that can be replaced by a special file that would contain size and the byte that fills. 
  - [ ] For the first "gap", the header can be actually skipped completely and coded as a constant in firmware. Please implement that while keeping it passing the tests.
  - [ ] For the gaps - you need to split them by 0xFFFF as well as those would be different sections within gaps. You can test that with `./sample_firmwares/fw-BlueField-2-rel-24_40_1000-MBF2M516A-EEEO_Ax_Bx-NVME-20.4.1-UEFI-21.4.13-UEFI-22.4.12-UEFI-14.33.10-FlexBoot-3.7.300.bin` that have such gaps.
  - [ ] Pure codestyle - it is better to keep all common logic in relevant structs and packages, while keeping command extremely slim - so please move the logic into relevant packages and keep the necessary boilerplate in separate file inside cmd/mlx5fw-go
  - [ ] Testing our new code against `sample_firmwares/fw-ConnectX7-rel-28_33_0751.bin`
    - [ ] Resulting firmware must be validated with `scripts/sample_tests/reassemble.sh` and `scripts/sample_tests/strict-reassemble.sh`
    - [ ] Resulting firmware must have exactly same sha256 as original one - that is ultimate criteria for success
