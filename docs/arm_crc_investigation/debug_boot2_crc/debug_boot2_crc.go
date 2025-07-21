package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/parser/fs4"
	"go.uber.org/zap"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run debug_boot2_crc.go <firmware.bin>")
		os.Exit(1)
	}

	logger, _ := zap.NewDevelopment()
	
	// Open firmware file
	reader, err := parser.NewFirmwareReader(os.Args[1], logger)
	if err != nil {
		logger.Fatal("Failed to open firmware", zap.Error(err))
	}
	defer reader.Close()
	
	// Parse to get sections
	p := fs4.NewParser(reader, logger)
	if err := p.Parse(); err != nil {
		logger.Fatal("Failed to parse firmware", zap.Error(err))
	}
	
	// Get BOOT2 section
	sections := p.GetSections()
	boot2Sections, ok := sections[0x100] // BOOT2 type
	if !ok || len(boot2Sections) == 0 {
		fmt.Println("No BOOT2 section found")
		return
	}
	
	boot2 := boot2Sections[0]
	fmt.Printf("BOOT2 section:\n")
	fmt.Printf("  Offset: 0x%x\n", boot2.Offset)
	fmt.Printf("  Size: 0x%x (%d bytes)\n", boot2.Size, boot2.Size)
	fmt.Printf("  CRC Type: %v\n", boot2.CRCType)
	
	// Read the actual BOOT2 data
	boot2Data, err := reader.ReadSection(int64(boot2.Offset), boot2.Size)
	if err != nil {
		fmt.Printf("Failed to read BOOT2: %v\n", err)
		return
	}
	
	// Show the header
	fmt.Printf("\nBOOT2 header:\n")
	for i := 0; i < 16 && i < len(boot2Data); i += 4 {
		dword := binary.BigEndian.Uint32(boot2Data[i:i+4])
		fmt.Printf("  [0x%02x]: 0x%08x\n", i, dword)
	}
	
	// The size in header should match
	headerSize := binary.BigEndian.Uint32(boot2Data[4:8])
	fmt.Printf("\nSize from header: 0x%x\n", headerSize)
	fmt.Printf("Size used by section: 0x%x\n", boot2.Size)
	
	if headerSize != boot2.Size {
		fmt.Printf("WARNING: Size mismatch!\n")
	}
	
	// Now let's see what's happening with verification
	status, err := p.VerifySection(boot2)
	fmt.Printf("\nVerification result: %s\n", status)
	if err != nil {
		fmt.Printf("Verification error: %v\n", err)
	}
}