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
- [X] All current commands should be unified to their maximum
    - [X] Define interfaces and base structs for sections
    - [X] Make all known section types their own structs and for all available commands you need to utilize those structs. Struct must contain data of the section itself
    - [X] Ensure that `scripts/sample_tests/sections.sh` still have 100% success rate.
    - [X] Ensure that `scripts/sample_tests/psid.sh` still have 100% success rate.
    - [X] Create helper structs and interfaces for calculating CRC (software and hardware, as defined by mlx5fw-go right now)
    - [X] For each section struct define the correct type of CRC (In Section or other) and set correct flags, add methods to return that
    - [X] Use embedding to embed correct CRC algorithm for each section type
    - [X] Rewrite commands to use new methods instead of reimplementing that in each command
    - [X] Ensure that `scripts/sample_tests/sections.sh` still have 100% success rate.
    - [X] Ensure that `scripts/sample_tests/psid.sh` still have 100% success rate.
    - [X] Rewrite commands to unify their code base as much as possible and move all potentially usable functions to be method of their respective sections or some other shared places - idea is that commands themself should be way smaller.
    - [X] Extract command must save metadata that is useful for later reassembly
    - [X] Extract command must define useful headers for firmwares and have them as constants that can be later on reused
    - [X] Extract command must save also all data that doesn't belong to any other section
    - [X] Extract must not preseve data that can be reconstructed later, e.x. checksums should be cut
    - [X] Added extra Currently FirmwareReader seems to have a lot of code duplication - e.x. ReadHWPointers and ReadSection are almost the same code path and can be combined. ReadUint64BE seems like it can be inlined (there is only one usage in the code base) and ReadUint32BE is unused. Those are just examples, your task is to go over the codebase again and try to simplify it as much as possible and make it ideomatic Golang codebase.
    - [X] For well understood sections like ITOC, DTOC, HWPointers and HashTables, data should be saved not only as raw binary but as parsed json.
    - [X] Ensure that `scripts/sample_tests/sections.sh` still have 100% success rate.
    - [X] Ensure that `scripts/sample_tests/psid.sh` still have 100% success rate.
- [ ] Implement reassemble command
  - [X] It must perform check that we have possibly enough data to reassemble firmware back
  - [X] It must reconstruct firmware image from given folder that have all the data. Note that all CRC and other reconstructable data must be reconstructed on demand.
  - [ ] During reconstruction you needed to change checksum types, but sections command was passing fine, that means that it was not properly implemented before and there are places that uses different source of truth for how to calculate checksums. Please fix that.
  - [ ] With current extraction, there are three improvements can be made for gaps extraction: first - if gap consist only of 0xFF - that can be skipped completely as that is default "empty" for the firmware. Second - for the first section, the header can be actually skipped completely and coded as a constant in firmware. Please implement that while keeping it passing the tests. Third is pure codestyle - it is better to keep all common logic in relevant structs and packages, while keeping command extremely slim - so please move the logic into relevant packages and keep the necessary boilerplate in separate file inside cmd/mlx5fw-go
  - [ ] Testing our new code against `sample_firmwares/fw-ConnectX7-rel-28_33_0751.bin`
    - [ ] Resulting firmware must be validated and pass validation of mstflint. No CRC errors allowed. Firmware must be considered bootable
    - [ ] Resulting firmware must have exactly same sha256 as original one - that is ultimate criteria for success
  - [ ] Testing our new code against `fw-ConnectX5-rel-16_35_4030-MCX516A-CDA_Ax_Bx-UEFI-14.29.15-FlexBoot-3.6.902.bin`
    - [ ] Resulting firmware must be validated and pass validation of mstflint. No CRC errors allowed. Firmware must be considered bootable
    - [ ] Resulting firmware must have exactly same sha256 as original one - that is ultimate criteria for success
  - [ ] Test code against all other firmwares. sha256 of reconstructed firmware must match
- [ ] Implement diff command
  - [ ] Implement a command that accepts two binary firmwares as arguments and highlights the difference between them. Command must do that on per-segment basis, where segment is either section or the padding between sections. If it is a section and it can be parsed (e.x. Image Info) it must show the field-difference. If section have binary data - it must show a 5-column table: `address || hex1 | hex2 || printable characters1 | printable characters2` and do a color highlight of what has changed on each pair - hex and printable characters. Address must be in absolute format from the beginning of firmware in hex. If section have []byte field - it must be displayed in that format as well.
  - [ ] Ensure that `scripts/sample_tests/sections.sh` still have 100% success rate.
  - [ ] Ensure that `scripts/sample_tests/psid.sh` still have 100% success rate.
