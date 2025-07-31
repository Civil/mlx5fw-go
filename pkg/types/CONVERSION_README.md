# Struct to Alias Conversion Script

This script automates the conversion of legacy struct definitions to use annotated versions via type aliases.

## What it does

For each struct that has `ToAnnotated`/`FromAnnotated` methods:

1. **Creates a type alias file** (e.g., `xxx_aliases.go`) that makes the legacy struct name point to the annotated version
2. **Preserves any methods** that are on the legacy struct (like `GetNumAllocated`, `GetPSIDString`, etc.)
3. **Removes the old struct definition file** (both `.go` and `_marshal.go` files)
4. **Removes conversion methods** from the annotated file:
   - `ToAnnotated`/`FromAnnotated` methods
   - `UnmarshalWithReserved`/`MarshalWithReserved` methods

## Usage

```bash
# Preview what will be done (dry run)
./convert_structs_to_aliases.sh --dry-run

# Run with verbose output
./convert_structs_to_aliases.sh --verbose

# Run the conversion
./convert_structs_to_aliases.sh
```

## Options

- `--dry-run`: Show what would be done without making changes
- `--verbose`: Show detailed output including method extraction
- `--help`: Show help message

## Safety Features

1. **Automatic Backup**: Creates timestamped backup directory with all modified/deleted files
2. **Skip Existing**: Won't process structs that already have alias files
3. **Method Preservation**: Automatically extracts and preserves all non-Marshal methods
4. **Validation**: Checks for conversion methods before processing

## Example Conversion

Before:
```
device_info.go          # Legacy struct with Marshal/Unmarshal delegating to annotated
device_info_annotated.go # Annotated struct with ToAnnotated/FromAnnotated methods
```

After:
```
device_info_aliases.go   # Type aliases and preserved methods
device_info_annotated.go # Cleaned annotated struct (no conversion methods)
# device_info.go - REMOVED
```

## Features

- **Automatic Detection**: Finds all structs with ToAnnotated/FromAnnotated methods
- **Method Preservation**: Extracts and preserves all non-Marshal methods
- **Safety First**: Creates timestamped backups before any modifications
- **Verification**: Checks that alias files are created correctly
- **Summary Report**: Generates a summary of all conversions
- **Color-coded Output**: Clear visual feedback during execution

## Restoration

If something goes wrong, restore from backup:
```bash
cp backup_YYYYMMDD_HHMMSS/* .
```

## Patterns Handled

The script handles these common patterns:

1. **Legacy struct files** with Marshal/Unmarshal that delegate to annotated
2. **Annotated files** with ToAnnotated/FromAnnotated conversion methods
3. **Marshal files** with separate Marshal/Unmarshal implementations
4. **Helper methods** that need to be preserved (Get*, Set*, String, etc.)

## Manual Steps After Conversion

1. Run tests to ensure everything still works
2. Check for any import statements that might need updating
3. Review the generated alias files to ensure all methods were preserved correctly
4. Remove the backup directory once you're satisfied with the conversion