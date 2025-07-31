#!/bin/bash

# Script to convert structs to use annotated versions via type aliases
# Final version with enhanced safety and verification

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
BACKUP_DIR="backup_$(date +%Y%m%d_%H%M%S)"
DRY_RUN=false
VERBOSE=false
SUMMARY_FILE="conversion_summary.txt"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "This script automates the conversion of legacy struct definitions to use"
            echo "annotated versions via type aliases."
            echo ""
            echo "Options:"
            echo "  --dry-run    Show what would be done without making changes"
            echo "  --verbose    Show detailed output"
            echo "  --help       Show this help message"
            echo ""
            echo "What it does:"
            echo "  1. Creates type alias files (xxx_aliases.go)"
            echo "  2. Preserves helper methods from legacy structs"
            echo "  3. Removes old struct definition files"
            echo "  4. Removes ToAnnotated/FromAnnotated conversion methods"
            echo "  5. Creates backups of all modified files"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Function to log messages
log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

debug() {
    if [ "$VERBOSE" = true ]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

success() {
    echo -e "${CYAN}[SUCCESS]${NC} $1"
}

# Function to create backup
create_backup() {
    if [ "$DRY_RUN" = true ]; then
        log "Would create backup directory: $BACKUP_DIR"
        return
    fi
    
    if [ ! -d "$BACKUP_DIR" ]; then
        mkdir -p "$BACKUP_DIR"
        log "Created backup directory: $BACKUP_DIR"
    fi
}

# Function to backup a file
backup_file() {
    local file=$1
    if [ -f "$file" ]; then
        if [ "$DRY_RUN" = true ]; then
            log "Would backup: $file"
        else
            cp "$file" "$BACKUP_DIR/"
            debug "Backed up: $file"
        fi
    fi
}

# Function to extract struct names from annotated files
get_struct_names() {
    local annotated_file=$1
    grep -E "^type.*Annotated struct" "$annotated_file" | sed -E 's/type ([^ ]+) struct.*/\1/' | sed 's/Annotated$//'
}

# Function to check if file has ToAnnotated/FromAnnotated methods
has_conversion_methods() {
    local file=$1
    grep -q "ToAnnotated\|FromAnnotated" "$file"
}

# Function to extract methods for a struct (excluding Marshal/Unmarshal)
extract_struct_methods() {
    local file=$1
    local struct_name=$2
    local temp_file=$(mktemp)
    
    # Use awk to extract methods for the struct
    awk -v struct="$struct_name" '
    BEGIN { in_method = 0; }
    /^func \([^)]*\*?'"$struct_name"'\) [A-Za-z][A-Za-z0-9_]*\(/ {
        # Skip Marshal/Unmarshal methods as they are handled by annotations
        if ($0 ~ /Marshal\(/ || $0 ~ /Unmarshal\(/) {
            next
        }
        in_method = 1
        print ""
        print "// Method from " FILENAME
    }
    in_method {
        print
        if (/^}$/) {
            in_method = 0
            print ""
        }
    }
    ' "$file" > "$temp_file"
    
    # Only output if we found methods
    if [ -s "$temp_file" ]; then
        cat "$temp_file"
    fi
    rm -f "$temp_file"
}

# Function to verify the conversion was successful
verify_conversion() {
    local base_name=$1
    local alias_file="${base_name}_aliases.go"
    
    if [ "$DRY_RUN" = true ]; then
        return 0
    fi
    
    # Check if alias file was created
    if [ ! -f "$alias_file" ]; then
        error "Alias file not created: $alias_file"
        return 1
    fi
    
    # Check if it contains type aliases
    if ! grep -q "^type .* = .*Annotated$" "$alias_file"; then
        error "No type aliases found in: $alias_file"
        return 1
    fi
    
    return 0
}

# Function to process a struct conversion
process_struct() {
    local base_name=$1
    local annotated_file="${base_name}_annotated.go"
    local legacy_file="${base_name}.go"
    local marshal_file="${base_name}_marshal.go"
    local alias_file="${base_name}_aliases.go"
    
    log "Processing: $base_name"
    
    # Check if annotated file exists
    if [ ! -f "$annotated_file" ]; then
        warn "Annotated file not found: $annotated_file"
        return
    fi
    
    # Check if it has conversion methods
    if ! has_conversion_methods "$annotated_file"; then
        warn "No ToAnnotated/FromAnnotated methods found in: $annotated_file"
        return
    fi
    
    # Get struct names from annotated file
    local struct_names=$(get_struct_names "$annotated_file")
    debug "Found structs: $(echo $struct_names | tr '\n' ' ')"
    
    # Create the alias file content
    local alias_content="package types

// Type aliases to use annotated versions directly
"
    
    # Add type aliases for each struct
    for struct in $struct_names; do
        alias_content+="type ${struct} = ${struct}Annotated
"
    done
    
    # Collect methods from legacy files
    local methods_content=""
    local methods_found=false
    
    # Check legacy file for methods
    if [ -f "$legacy_file" ]; then
        debug "Checking legacy file for methods: $legacy_file"
        for struct in $struct_names; do
            local methods=$(extract_struct_methods "$legacy_file" "$struct")
            if [ ! -z "$methods" ]; then
                methods_found=true
                methods_content+="$methods"
                debug "Found methods for $struct in $legacy_file"
            fi
        done
    fi
    
    # Check marshal file for methods (excluding Marshal/Unmarshal)
    if [ -f "$marshal_file" ]; then
        debug "Checking marshal file for methods: $marshal_file"
        for struct in $struct_names; do
            local methods=$(extract_struct_methods "$marshal_file" "$struct")
            if [ ! -z "$methods" ]; then
                methods_found=true
                methods_content+="$methods"
                debug "Found methods for $struct in $marshal_file"
            fi
        done
    fi
    
    # Add methods to alias content if any were found
    if [ ! -z "$methods_content" ]; then
        alias_content+="
$methods_content"
    fi
    
    # Perform the conversion
    if [ "$DRY_RUN" = true ]; then
        log "Would create alias file: $alias_file"
        if [ "$VERBOSE" = true ]; then
            echo -e "${CYAN}=== Content of $alias_file ===${NC}"
            echo "$alias_content"
            echo -e "${CYAN}=== End of content ===${NC}"
        fi
        log "Would remove ToAnnotated/FromAnnotated methods from: $annotated_file"
        log "Would remove UnmarshalWithReserved/MarshalWithReserved if present"
        [ -f "$legacy_file" ] && log "Would remove: $legacy_file"
        [ -f "$marshal_file" ] && log "Would remove: $marshal_file"
        echo "[DRY RUN] $base_name: $(echo $struct_names | wc -w) structs, $methods_found methods preserved" >> "$SUMMARY_FILE"
    else
        # Backup files
        backup_file "$annotated_file"
        [ -f "$legacy_file" ] && backup_file "$legacy_file"
        [ -f "$marshal_file" ] && backup_file "$marshal_file"
        
        # Create alias file
        echo "$alias_content" > "$alias_file"
        log "Created alias file: $alias_file"
        
        # Remove conversion methods from annotated file
        # Create a temporary file for the cleaned content
        local temp_file=$(mktemp)
        
        # Remove ToAnnotated/FromAnnotated methods and their implementations
        awk '
            /^\/\/ (To|From)Annotated/ { skip=1; next }
            /^func.*\.(To|From)Annotated\(/ { skip=1; next }
            skip && /^}/ { skip=0; next }
            skip { next }
            /^\/\/ (Unmarshal|Marshal)WithReserved/ { skip2=1; next }
            /^func.*\.(Unmarshal|Marshal)WithReserved\(/ { skip2=1; next }
            skip2 && /^}/ { skip2=0; next }
            skip2 { next }
            { print }
        ' "$annotated_file" > "$temp_file"
        
        mv "$temp_file" "$annotated_file"
        log "Removed conversion methods from: $annotated_file"
        
        # Remove legacy files
        [ -f "$legacy_file" ] && rm "$legacy_file" && log "Removed: $legacy_file"
        [ -f "$marshal_file" ] && rm "$marshal_file" && log "Removed: $marshal_file"
        
        # Verify the conversion
        if verify_conversion "$base_name"; then
            success "Successfully converted: $base_name"
            echo "$base_name: $(echo $struct_names | wc -w) structs, $methods_found methods preserved" >> "$SUMMARY_FILE"
        else
            error "Conversion verification failed for: $base_name"
            echo "$base_name: FAILED" >> "$SUMMARY_FILE"
        fi
    fi
    
    echo
}

# Main execution
main() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}Struct to Alias Conversion Tool${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo
    
    # Initialize summary file
    if [ "$DRY_RUN" = true ]; then
        echo "DRY RUN - Conversion Summary" > "$SUMMARY_FILE"
    else
        echo "Conversion Summary - $(date)" > "$SUMMARY_FILE"
    fi
    echo "================================" >> "$SUMMARY_FILE"
    
    # Create backup directory
    create_backup
    
    # Find all annotated files that have ToAnnotated/FromAnnotated methods
    local annotated_files=$(grep -l "ToAnnotated\|FromAnnotated" *_annotated.go 2>/dev/null || true)
    
    if [ -z "$annotated_files" ]; then
        warn "No annotated files with conversion methods found"
        exit 0
    fi
    
    # Count files to process
    local total_files=$(echo "$annotated_files" | wc -w)
    log "Found $total_files files to process"
    echo
    
    # Process each annotated file
    local processed=0
    local skipped=0
    
    for annotated_file in $annotated_files; do
        # Extract base name
        local base_name=$(basename "$annotated_file" "_annotated.go")
        
        # Skip if alias file already exists
        if [ -f "${base_name}_aliases.go" ]; then
            warn "Alias file already exists for $base_name, skipping"
            ((skipped++))
            continue
        fi
        
        # Check if there's a corresponding legacy or marshal file
        if [ -f "${base_name}.go" ] || [ -f "${base_name}_marshal.go" ]; then
            process_struct "$base_name"
            ((processed++))
        else
            debug "No legacy files found for $base_name"
            ((skipped++))
        fi
    done
    
    # Print summary
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}Conversion Summary${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo "Total files found: $total_files"
    echo "Files processed: $processed"
    echo "Files skipped: $skipped"
    
    if [ "$DRY_RUN" = false ]; then
        echo
        log "Backup created in: $BACKUP_DIR"
        log "Summary written to: $SUMMARY_FILE"
        echo
        echo -e "${YELLOW}To restore from backup:${NC}"
        echo "  cp $BACKUP_DIR/* ."
        echo
        echo -e "${YELLOW}Next steps:${NC}"
        echo "  1. Run tests to ensure everything works"
        echo "  2. Check for any import errors"
        echo "  3. Review the generated alias files"
        echo "  4. Delete backup once satisfied"
    fi
    
    echo
    success "Conversion complete!"
}

# Run main function
main