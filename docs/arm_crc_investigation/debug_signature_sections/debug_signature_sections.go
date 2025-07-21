package main

import (
	"fmt"
	"os"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/parser/fs4"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"go.uber.org/zap"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run debug_signature_sections.go <firmware.bin>")
		os.Exit(1)
	}

	logger, _ := zap.NewDevelopment()
	
	// Open firmware file
	reader, err := parser.NewFirmwareReader(os.Args[1], logger)
	if err != nil {
		logger.Fatal("Failed to open firmware", zap.Error(err))
	}
	defer reader.Close()
	
	p := fs4.NewParser(reader, logger)

	if err := p.Parse(); err != nil {
		logger.Fatal("Failed to parse firmware", zap.Error(err))
	}

	// Check BOOT2 section
	checkSections := []uint16{
		types.SectionTypeBoot2,
	}

	sections := p.GetSections()
	for _, sectionType := range checkSections {
		if sectionList, ok := sections[sectionType]; ok {
			for _, section := range sectionList {
				fmt.Printf("Section %s (0x%x):\n", types.GetSectionTypeName(sectionType), sectionType)
				fmt.Printf("  Offset: 0x%x\n", section.Offset)
				fmt.Printf("  Size: 0x%x\n", section.Size)
				fmt.Printf("  CRCType: %v\n", section.CRCType)
				if section.Entry != nil {
					fmt.Printf("  Entry CRC field: %d\n", section.Entry.CRC)
					fmt.Printf("  Entry GetNoCRC(): %v\n", section.Entry.GetNoCRC())
				}
				fmt.Println()
			}
		}
	}
}