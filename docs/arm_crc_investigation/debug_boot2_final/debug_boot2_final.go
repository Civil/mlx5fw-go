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
		fmt.Println("Usage: go run debug_boot2_final.go <firmware.bin>")
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
	
	// Read the first 16 bytes of BOOT2 header
	headerData, _ := reader.ReadSection(int64(boot2Addr), 16)
	
	// Parse header fields
	magic := binary.BigEndian.Uint32(headerData[0:4])
	sizeField := binary.BigEndian.Uint32(headerData[4:8])
	field2 := binary.BigEndian.Uint32(headerData[8:12])
	field3 := binary.BigEndian.Uint32(headerData[12:16])
	
	fmt.Printf("BOOT2 Header at 0x%x:\n", boot2Addr)
	fmt.Printf("  [0x00]: 0x%08x (magic?)\n", magic)
	fmt.Printf("  [0x04]: 0x%08x (size field)\n", sizeField)
	fmt.Printf("  [0x08]: 0x%08x\n", field2)
	fmt.Printf("  [0x0c]: 0x%08x\n", field3)
	
	// mstflint shows BOOT2 as 0x1000-0x1edf (0xee0 bytes)
	// And our size field is 0xee0
	// So it seems the size field IS in bytes, not dwords
	
	// But mstflint code does: boot2Size = (size + 4) * 4
	// Let's work backwards:
	// If boot2Size should be 0xee0, then:
	// 0xee0 = (size + 4) * 4
	// 0xee0 / 4 = size + 4
	// 0x3b8 = size + 4
	// size = 0x3b4
	
	fmt.Printf("\nAnalysis:\n")
	fmt.Printf("  mstflint shows BOOT2 range: 0x1000-0x1edf (0xee0 bytes)\n")
	fmt.Printf("  Size field value: 0x%x\n", sizeField)
	
	if sizeField == 0xee0 {
		fmt.Println("  => Size field appears to be total size in BYTES")
		// This contradicts mstflint code which expects dwords
		// Let's check the actual data
		
		// Read full BOOT2
		boot2Data, _ := reader.ReadSection(int64(boot2Addr), 0xee0)
		
		// Check what's at the end
		fmt.Println("\nLast 16 bytes of BOOT2:")
		for i := 0xee0 - 16; i < 0xee0; i += 4 {
			dword := binary.BigEndian.Uint32(boot2Data[i:i+4])
			fmt.Printf("  [0x%04x]: 0x%08x\n", i, dword)
		}
		
		// Now let's try CRC calculation
		// According to mstflint source, CRC is calculated from dword 3 to size-1
		// And CRC is at position buff[size + 3]
		
		// But if size is in bytes, we need to convert
		sizeInDwords := sizeField / 4
		fmt.Printf("\nSize in dwords: 0x%x / 4 = 0x%x\n", sizeField, sizeInDwords)
		
		// Try method 1: Assume mstflint expects size in dwords but we have bytes
		// So recalculate what the size field SHOULD be
		expectedSizeFieldDwords := (sizeField / 4) - 4  // Total dwords minus 4 header dwords
		fmt.Printf("Expected size field (dwords): (0x%x / 4) - 4 = 0x%x\n", sizeField, expectedSizeFieldDwords)
		
		// Now check if that value exists in our header
		// Field at offset 8 is 0x9bc5 - that's not it
		
		// Let's just try different CRC calculations
		crcCalc := parser.NewCRCCalculator()
		
		fmt.Println("\nTrying different CRC calculations:")
		
		// Method 1: Skip first 12 bytes (3 dwords), calculate on rest except last 4 bytes
		data1 := boot2Data[12:0xee0-4]
		crc1 := crcCalc.CalculateImageCRC(data1, len(data1)/4)
		fmt.Printf("  Method 1 (skip 3 dwords, exclude last dword): 0x%04x\n", crc1)
		
		// Method 2: Calculate on all data except last 4 bytes
		data2 := boot2Data[:0xee0-4]
		crc2 := crcCalc.CalculateImageCRC(data2, len(data2)/4)
		fmt.Printf("  Method 2 (all except last dword): 0x%04x\n", crc2)
		
		// Check what the expected CRC is
		lastDword := binary.BigEndian.Uint32(boot2Data[0xee0-4:])
		fmt.Printf("\nLast dword (potential CRC location): 0x%08x\n", lastDword)
		fmt.Printf("  Lower 16 bits: 0x%04x\n", lastDword & 0xFFFF)
		
		// Also check field at offset 8 
		fmt.Printf("\nField at offset 8: 0x%04x (could this be CRC?)\n", field2 & 0xFFFF)
		
		// Try matching against field at offset 8
		if uint32(crc1) == (field2 & 0xFFFF) {
			fmt.Println("  => Method 1 matches field at offset 8!")
		}
		if uint32(crc2) == (field2 & 0xFFFF) {
			fmt.Println("  => Method 2 matches field at offset 8!")
		}
		
		// Actually, let's look at FS4 more carefully
		// In FS4, BOOT2 might be part of ITOC sections
		// Let me check if this is really a standalone BOOT2 or part of ITOC
		fmt.Println("\nChecking if this might be an ITOC section...")
		
		// The value 0x3b4 that we calculated earlier...
		// Let's see if CRC calculation with that assumption works
		
		// If size field was 0x3b4 dwords, CRC would be at buff[0x3b4 + 3] = buff[0x3b7]
		// That's dword 0x3b7, byte offset 0x3b7 * 4 = 0xedc
		
		if 0xedc < 0xee0 {
			crcPos := 0xedc
			potentialCRC := binary.BigEndian.Uint32(boot2Data[crcPos:crcPos+4])
			fmt.Printf("\nIf size was 0x3b4 dwords, CRC at 0x%x: 0x%08x\n", crcPos, potentialCRC)
			
			// Calculate CRC for that scenario
			// Skip first 3 dwords, calculate on 0x3b4 dwords
			data3 := boot2Data[12:12+0x3b4*4]
			crc3 := crcCalc.CalculateImageCRC(data3, 0x3b4)
			fmt.Printf("  Calculated CRC for that scenario: 0x%04x\n", crc3)
			if uint32(crc3) == potentialCRC {
				fmt.Println("  => MATCH! This might be the correct interpretation")
			}
		}
	}
}