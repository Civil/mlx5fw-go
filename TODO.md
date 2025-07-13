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


Development plan:
 - [X] "Documentation phase" Using provided sample firmware and by running `FW_COMPS_DEBUG=1 mstflint -i "$FIRMWARE" v` and `FW_COMPS_DEBUG=1 mstflint -i "$FIRMWARE" query full` understand and document the process of parsing different firmwares. Source code for that part is stored in `refrence/mstflint` folder. You may modify it and recompile if needed, you may use other debug variables if that would be helpful to get more information, you can run it under debugger to analyse step-by-step execution. You can use gdb mcp server to help with investigation and dlv mcp server to debug Go code. More detailed checklist:
   - [X] Gather high level understanding of what functions from what files are called using `FW_COMPS_DEBUG` method and save in mstflint.md for `sample_firmwares/fw-ConnectX5-rel-16_35_4030-MCX516A-CDA_Ax_Bx-UEFI-14.29.15-FlexBoot-3.6.902.bin`
   - [X] Update high level understanding of what functions from what files are called using `FW_COMPS_DEBUG` method and save in mstflint.md for `sample_firmwares/fw-ConnectX6Dx-rel-22_41_1000-MCX623106AN-CDA_Ax-UEFI-14.34.12-FlexBoot-3.7.400.bin`
   - [X] Update high level understanding of what functions from what files are called using `FW_COMPS_DEBUG` method and save in mstflint.md for `sample_firmwares/fw-BlueField-2-rel-24_41_1000-MBF2M516A-CENO_Ax_Bx-NVME-20.4.1-UEFI-21.4.13-UEFI-22.4.13-UEFI-14.34.12-FlexBoot-3.7.400.bin`
   - [X] Update high level understanding of what functions from what files are called using `FW_COMPS_DEBUG` method and save in mstflint.md for `sample_firmwares/900-9D3B6-00CN-A_Ax_MT_0000000883_rel-32_45.1020.bin`
   - [X] Update high level understanding of what functions from what files are called using `FW_COMPS_DEBUG` method and save in mstflint.md for `sample_firmwares/fw-ConnectX7-rel-28_33_0800.bin`
   - [X] Update high level understanding of what functions from what files are called using `FW_COMPS_DEBUG` method and save in mstflint.md for `sample_firmwares/fw-ConnectX8-rel-40_45_1200-900-9X81Q-00CN-ST0_Ax-UEFI-14.38.16-FlexBoot-3.7.500.signed.bin`
   - [X] Enchance the understanding using gdb mcp server and step-by-setp execution and update mstflint.md with new findings for `sample_firmwares/fw-ConnectX5-rel-16_35_4030-MCX516A-CDA_Ax_Bx-UEFI-14.29.15-FlexBoot-3.6.902.bin`
   - [X] Enchance the understanding using gdb mcp server and step-by-setp execution and update mstflint.md with new findings for `sample_firmwares/fw-ConnectX6Dx-rel-22_41_1000-MCX623106AN-CDA_Ax-UEFI-14.34.12-FlexBoot-3.7.400.bin`
   - [X] Enchance the understanding using gdb mcp server and step-by-setp execution and update mstflint.md with new findings for `sample_firmwares/fw-BlueField-2-rel-24_41_1000-MBF2M516A-CENO_Ax_Bx-NVME-20.4.1-UEFI-21.4.13-UEFI-22.4.13-UEFI-14.34.12-FlexBoot-3.7.400.bin`
   - [X] Enchance the understanding using gdb mcp server and step-by-setp execution and update mstflint.md with new findings for `sample_firmwares/900-9D3B6-00CN-A_Ax_MT_0000000883_rel-32_45.1020.bin`
   - [X] Enchance the understanding using gdb mcp server and step-by-setp execution and update mstflint.md with new findings for `sample_firmwares/fw-ConnectX7-rel-28_33_0800.bin`
   - [X] Enchance the understanding using gdb mcp server and step-by-setp execution and update mstflint.md with new findings for `sample_firmwares/fw-ConnectX8-rel-40_45_1200-900-9X81Q-00CN-ST0_Ax-UEFI-14.38.16-FlexBoot-3.7.500.signed.bin`
   - [X] Verify documentation and document diviations (if any) for other firmwares (vendor firmwares)
   - [X] Analyse source code (you can modify and recomplie it if needed to add more debug information), you can use gdb mcp server to analyse execution and document the path better and document exact firmware parsing path in mstflint.md.
   - [X] Document all data structures that are used by mstflint and needs to be ported in mstflint.md, ensure that correct endianness is mentioned correctly.
   - [X] Document what type of checksum verification was implemented for which section and document algorithms for them in mstflint.md.
 - [X] Outline path for porting subset of mstflint's functionality to Go
   - [X] "Preparation phase" Define clean extensible arhitecture for firmware parsing. At first steps this tool only need to be able to print metadata of any firmware correclty (description, version, etc, and full sections list) in format similar to mstflint, however later it would be needed to print text-based sections (FW\_DBG\_INI is a compressed ini file for example and can be printed), implement checksum verification (and only checksum, you must not give any recomendations on security), do a diff between two firmwares (diff = show which sections are different, and what fields in those sections are different, if there is a correct datastructure to hold the, if there is no structure defined - hex-diff would be required) and replace sections (that requires replacing the section and recalculating the checksums)
   - [X] Define and port data structures that were discovered in "documentation phase"
 - [X] For cmd/mlx5fw-go implement sections printing command.
   - [X] Implement and print sections content of the firmware in human-readable format
   - [X] Write unit tests for the functionality (comprehensive test suite exists)
   - [X] Add optional json output (fully implemented)
 - [X] For cmd/mlx5fw-go implement metadata parsing
   - [X] Parse firmware metadata (equivalent to `mstflint -i <FIRMWARE NAME> query full`) for unencrypted firmwares
   - [X] Parse firmware metadata for all firmwares in `sample_firmwares/` with 100% match with data that mstflint shows
   - [X] Make sure that all `sample_firmwares/test_*.sh` tests runs without errors
   - [X] Make sure that 5 runs of `integration_tests/*.sh` runs with 100% completion rate
   - [X] Write unit tests for the functionality
   - [X] Add optional json output
   - [X] Ensure that tests still completes at 100%
 - [X] For cmd/mlx5fw-go implement equivalent to `mstflint -i <FILENAME> dc` command. That command ungzips content for `DBG_FW_INI` section if it is present.
   - [X] Implement printing of the content (if it is in firmware).
   - [X] Write custom test scripts that would compare mstflint and mlx5fw-go and check that output matches 100%
   - [X] Write unit tests for that functionality
 - [X] for cmd/mlx5fw-go implement verification commmand, equivalent to `mstflint -i <FILENAME> v`.
   - [X] Implement that functionality.
   - [X] Write custom test scripts that would compare results to what mstflint do. You must create few invalid files and few files where only some of the sections damaged. and check that verification pass and doesn't pass on same amount of firmwares.
   - [X] Write unit tests for that functionality.
 - [X] For cmd/mlx5fw-go implement extract functionality. That option must create a folder and dump content of all sections into files named after each section there.
   - [X] Implement that functionality
   - [X] Write custom test scripts that would extract sections using shell scripts, dd and information that `mstflint v` provides and check that output matches 100%
   - [X] Write unit tests for that functionality
 - [X] For cmd/mlx5fw-go implement section replacement functionality.
   - [X] Implement functionality that would take `firmware_file`, `output_firmware`, `section_id` and `new_content_filename` and replace `section_id` in the firmware with new content and output it to a new file. All checksums if needed must be recalculated and updated.
   - [X] Implement test scripts that would extract two firmwares of different versions and swap sections from one to another, and then call `mstflint -i <FILENAME> v` and ensure that firmware is still treated as bootable.
   - [X] Implement unit tests
 - [ ] For cmd/mlx5fw-go implement section removal functionality
   - [ ] Implement command line option that would remove specified section ID from firmware and save it as a new file.
   - [ ] Write custom shell scripts that would remove a scetion and check that `mstflint -i <FILENAME> v` still shows that everything is fine.
