package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"go.uber.org/zap"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run debug_boot2_mstflint_way.go <firmware.bin>")
		os.Exit(1)
	}

	logger, _ := zap.NewDevelopment()
	
	// Open firmware file
	reader, err := parser.NewFirmwareReader(os.Args[1], logger)
	if err != nil {
		logger.Fatal("Failed to open firmware", zap.Error(err))
	}
	defer reader.Close()

	// BOOT2 is at 0x1000
	boot2Addr := uint32(0x1000)
	
	// Read boot2 header (specifically the size field at offset 4)
	sizeData, _ := reader.ReadSection(int64(boot2Addr + 4), 4)
	sizeFromHeader := binary.BigEndian.Uint32(sizeData)
	fmt.Printf("Size field from header (at 0x%x): 0x%08x\n", boot2Addr + 4, sizeFromHeader)
	
	// Let's check if the size is already in bytes or in dwords
	// If it's 0xee0 and we see BOOT2 ending at 0x1edf, then:
	// 0x1edf - 0x1000 + 1 = 0xee0, so it's already in bytes!
	
	// But mstflint code shows: _fwImgInfo.boot2Size = (size + 4) * 4;
	// This suggests size is in dwords minus 4 (for the header)
	// So if size field = 0xee0, that would give us (0xee0 + 4) * 4 = 0x3b90 bytes
	// That's way too big!
	
	// Let me check if size is actually (total_size_in_bytes / 4) - 4
	// Total size = 0xee0 bytes
	// In dwords = 0xee0 / 4 = 0x3b8
	// Minus header (4 dwords) = 0x3b8 - 4 = 0x3b4
	// Let's see if 0x3b4 is what's in the header
	
	fmt.Printf("Interpreting size field 0x%x:\n", sizeFromHeader)
	fmt.Printf("  As bytes: 0x%x\n", sizeFromHeader)
	fmt.Printf("  As dwords: 0x%x * 4 = 0x%x bytes\n", sizeFromHeader, sizeFromHeader * 4)
	
	// Let me recalculate - if mstflint output shows 0x1000-0x1edf
	// That's 0xee0 bytes total
	// The size field should be: (0xee0 / 4) - 4 = 0x3b8 - 4 = 0x3b4
	expectedSizeField := (0xee0 / 4) - 4
	fmt.Printf("  Expected size field if total is 0xee0: (0xee0 / 4) - 4 = 0x%x\n", expectedSizeField)
	
	// But we read 0xee0, so maybe it's already total size in bytes?
	// Let's assume it is and work backwards
	boot2SizeBytes := sizeFromHeader
	fmt.Printf("\nAssuming size field is total bytes: 0x%x\n", boot2SizeBytes)
	
	// Read the full BOOT2 section
	boot2Data, err := reader.ReadSection(int64(boot2Addr), boot2SizeBytes)
	if err != nil {
		fmt.Printf("Failed to read boot2: %v\n", err)
		return
	}
	
	// According to mstflint: u_int32_t crc_act = buff[size + 3];
	// This is at dword position size + 3
	crcDwordPos := sizeFromHeader + 3
	crcBytePos := crcDwordPos * 4
	
	fmt.Printf("\nCRC location:\n")
	fmt.Printf("  Dword position: size + 3 = 0x%x + 3 = 0x%x\n", sizeFromHeader, crcDwordPos)
	fmt.Printf("  Byte position: 0x%x * 4 = 0x%x\n", crcDwordPos, crcBytePos)
	
	if crcBytePos + 4 > uint32(len(boot2Data)) {
		fmt.Printf("  ERROR: CRC position 0x%x is beyond boot2 data size 0x%x\n", crcBytePos, len(boot2Data))
		return
	}
	
	// Get the CRC dword
	crcDword := binary.BigEndian.Uint32(boot2Data[crcBytePos:crcBytePos+4])
	fmt.Printf("  CRC dword at position 0x%x: 0x%08x\n", crcBytePos, crcDword)
	
	// Now calculate CRC on the data
	// According to mstflint, it calculates CRC on dwords 0 to size-1
	// (not including the header first 3 dwords and the CRC dword itself)
	crcCalc := parser.NewCRCCalculator()
	
	// Start from dword 3 (skip 3 header dwords) 
	// Calculate on 'size' dwords
	startOffset := uint32(3 * 4)  // Skip first 3 dwords
	crcDataSize := sizeFromHeader * 4  // size dwords in bytes
	
	fmt.Printf("\nCRC calculation:\n")
	fmt.Printf("  Start offset: %d (skip first 3 dwords)\n", startOffset)
	fmt.Printf("  Data size: %d bytes (%d dwords)\n", crcDataSize, sizeFromHeader)
	
	if startOffset + crcDataSize > uint32(len(boot2Data)) {
		fmt.Printf("  ERROR: Data range exceeds boot2 size\n")
		return
	}
	
	crcData := boot2Data[startOffset:startOffset + crcDataSize]
	calculatedCRC := crcCalc.CalculateImageCRC(crcData, int(sizeFromHeader))
	
	fmt.Printf("  Calculated CRC: 0x%04x\n", calculatedCRC)
	fmt.Printf("  Expected CRC: 0x%08x\n", crcDword)
	
	if uint32(calculatedCRC) == crcDword {
		fmt.Println("  CRC MATCH!")
	} else {
		fmt.Println("  CRC MISMATCH!")
	}
}