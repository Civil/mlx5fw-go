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

You are tasked to improve extract command and implement reassemble and later on custom diff commands. Here is the breakdown of the task into smaller tasks:
- [ ] Refactor the code. Right now marshaling and unmarshaling are two different methods in different places (with potential duplication between query and extract commands). You must go over the code and make Marhsal & Unmarshal methods for each section in the `types` package based on how extract, query and reassemble works. This will allow to keep context as small as possible when modifying any of those methods in a future.
- [ ] mstflint have structs for more sections (more than IMAGE_INFO, DEV_INFO and MFG_INFO) defined in `reference/mstflint/tools_layouts/image_layout_layouts.h` and surrounding files, so they can be parsed. You need to look into mstflint and port those section definitions into mlx5fw-go and use them during extraction and reassembly process where possible.
- [ ] Refactor current code base by implementing a custom struct annotations.
  - [ ] Right now all data parsing is done by using either binstruct or custon binary parsing code. You must implement a package that would allow you to define custom annotations for structs, that should include 'byte offset', support 'bit offset' and 'bit length' for each field and allow to define source endianness.
  - [ ] You must implement generic unmarshaling and marshaling methods that would allow to use those annotations to parse and marshal data from/to structs.
  - [ ] You must convert HW Pointers to use your new method with annotations.
  - [ ] You must ensure that sections, extract and reassemble method uses your new method to marshal and unmarshal HW Pointers.
  - [ ] You must ensure that `scripts/sample_tests/sections.sh` and `scripts/sample_tests/strict-reassemble.sh` pass successfully after your changes.
  - [ ] You must port other data types to use same annotantion-based method
  - [ ] You must ensure that sections, extract and reassemble uses only annotation-based methods to marshal and unmarshal all structs.
  - [ ] You must ensure that `scripts/sample_tests/sections.sh` and `scripts/sample_tests/strict-reassemble.sh` pass successfully after your changes.


