# MLX5FW-GO Design Documentation

This directory contains comprehensive design documentation for the mlx5fw-go project, created to aid in the implementation of firmware flashing functionality.

## Documentation Structure

1. **[Project Overview](project-overview.md)**
   - Introduction to the project
   - Key features and technology stack
   - Current status

2. **[Architecture](architecture.md)**
   - Layered architecture overview
   - Component descriptions
   - Design patterns used
   - Data flow diagrams

3. **[Architecture Diagrams](architecture-diagram.md)**
   - Visual representations using Mermaid diagrams
   - System architecture overview
   - Sequence diagrams for key flows
   - Class diagrams for main components

4. **[Core Interfaces and Data Structures](core-interfaces.md)**
   - Detailed interface definitions
   - Core data structures
   - Design principles
   - Extension guidelines

5. **[Device Access Design](device-access-design.md)**
   - Current state (file-based operations)
   - Requirements for device access
   - Proposed interfaces for flashing
   - Platform abstraction considerations

6. **[Firmware Parsing Flow](firmware-parsing-flow.md)**
   - FS4 format structure
   - Step-by-step parsing process
   - CRC verification mechanisms
   - Binary parsing with annotations

7. **[Utilities and Helpers](utilities-and-helpers.md)**
   - Annotation framework for binary parsing
   - CRC calculation utilities
   - Error handling patterns
   - Common helper functions

8. **[Testing Infrastructure](testing-infrastructure.md)**
   - Test structure and approaches
   - Unit and integration tests
   - Test data management
   - Best practices

## Quick Navigation

### For Understanding the Codebase
- Start with [Project Overview](project-overview.md)
- Review [Architecture](architecture.md) and [Architecture Diagrams](architecture-diagram.md)
- Study [Core Interfaces](core-interfaces.md)

### For Implementing New Features
- Check [Firmware Parsing Flow](firmware-parsing-flow.md) for parsing logic
- Review [Utilities and Helpers](utilities-and-helpers.md) for available tools
- Follow patterns in [Core Interfaces](core-interfaces.md)

### For Adding Device Access
- Read [Device Access Design](device-access-design.md)
- Review proposed interfaces and platform abstraction

### For Testing
- Consult [Testing Infrastructure](testing-infrastructure.md)
- Follow established testing patterns

## Key Concepts

### Section Types
Common firmware sections you'll encounter:
- `BOOT2` (0x7): Boot loader section
- `DEV_INFO` (0x8): Device information
- `IMAGE_INFO` (0xD): Firmware image metadata
- `MFG_INFO` (0x5): Manufacturing information
- `TOOLS_AREA` (0x14): Tools and manufacturing data
- `HASHES_TABLE` (0x24): Security hashes

### CRC Types
- `IN_ITOC_ENTRY`: CRC stored in ITOC entry
- `NONE`: No CRC verification
- `IN_SECTION`: CRC embedded at end of section

### Important Offsets
- Magic pattern offset: Variable (0x0, 0x10000, 0x20000, etc.)
- HW Pointers: Magic offset + 0x24
- ITOC/DTOC addresses: Found in HW Pointers

## Development Guidelines

1. **Interface First**: Define interfaces before implementation
2. **Error Context**: Always wrap errors with meaningful context
3. **Logging**: Use structured logging with zap
4. **Testing**: Write tests for all new functionality
5. **Documentation**: Update relevant docs when adding features

## Future Work

The documentation will be updated as the project evolves, particularly:
- Device access implementation details
- Flash programming sequences
- Error recovery mechanisms
- Performance optimization strategies