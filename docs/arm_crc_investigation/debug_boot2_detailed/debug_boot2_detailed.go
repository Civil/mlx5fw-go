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
		fmt.Println("Usage: go run debug_boot2_detailed.go <firmware.bin>")
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
	
	// Read boot2 header
	headerData, _ := reader.ReadSection(int64(boot2Addr), 16)
	boot2Size := binary.BigEndian.Uint32(headerData[4:8])
	fmt.Printf("BOOT2 at 0x%x, size from header: 0x%x (%d bytes)\n", boot2Addr, boot2Size, boot2Size)
	
	// Read full boot2
	boot2Data, _ := reader.ReadSection(int64(boot2Addr), boot2Size)
	
	// Let's examine the structure more carefully
	// BOOT2 size is 0xee0 = 3808 bytes
	// 3808 / 4 = 952 dwords
	fmt.Printf("\nBOOT2 structure analysis:\n")
	fmt.Printf("  Total size: %d bytes\n", len(boot2Data))
	fmt.Printf("  Size / 4: %d dwords (remainder: %d)\n", len(boot2Data)/4, len(boot2Data)%4)
	
	// Show last 32 bytes to understand CRC location
	fmt.Println("\nLast 32 bytes of BOOT2:")
	for i := len(boot2Data) - 32; i < len(boot2Data); i += 4 {
		dword := binary.BigEndian.Uint32(boot2Data[i:i+4])
		fmt.Printf("  [0x%04x]: 0x%08x", i, dword)
		if i == len(boot2Data) - 4 {
			fmt.Printf(" <- Last dword (expected CRC location)")
		}
		fmt.Println()
	}
	
	// Let's check what CRC we get with different calculations
	crcCalc := parser.NewCRCCalculator()
	
	// Try calculating CRC on all data except last 4 bytes
	fmt.Println("\nCRC calculations:")
	
	// Method 1: Calculate on all data except last dword
	dataSizeBytes := len(boot2Data) - 4
	dataSizeDwords := dataSizeBytes / 4
	if dataSizeBytes % 4 != 0 {
		fmt.Printf("  WARNING: Data size %d is not dword aligned!\n", dataSizeBytes)
	}
	
	crc1 := crcCalc.CalculateImageCRC(boot2Data, dataSizeDwords)
	fmt.Printf("  Method 1 (ImageCRC on %d dwords): 0x%04x\n", dataSizeDwords, crc1)
	
	// Method 2: Try with one more dword
	crc2 := crcCalc.CalculateImageCRC(boot2Data, dataSizeDwords + 1)
	fmt.Printf("  Method 2 (ImageCRC on %d dwords): 0x%04x\n", dataSizeDwords + 1, crc2)
	
	// Method 3: Hardware CRC
	crc3 := crcCalc.CalculateHardwareCRC(boot2Data[:dataSizeBytes])
	fmt.Printf("  Method 3 (HardwareCRC on %d bytes): 0x%04x\n", dataSizeBytes, crc3)
	
	// Extract expected CRC from last dword
	lastDword := binary.BigEndian.Uint32(boot2Data[len(boot2Data)-4:])
	fmt.Printf("\nExpected CRC (from last dword 0x%08x):\n", lastDword)
	fmt.Printf("  Lower 16 bits: 0x%04x\n", lastDword & 0xFFFF)
	fmt.Printf("  Upper 16 bits: 0x%04x\n", (lastDword >> 16) & 0xFFFF)
	
	// Check if we're looking at the wrong location
	// In mstflint, BOOT2 might have a special structure
	fmt.Println("\nChecking alternative CRC locations:")
	
	// Check at offset 8 (after magic and size)
	if len(boot2Data) >= 12 {
		crcAt8 := binary.BigEndian.Uint32(boot2Data[8:12])
		fmt.Printf("  Dword at offset 8: 0x%08x (lower: 0x%04x)\n", crcAt8, crcAt8 & 0xFFFF)
		
		// Try calculating CRC without header
		headerSize := 12 // Skip first 3 dwords
		dataForCRC := boot2Data[headerSize:dataSizeBytes]
		crc4 := crcCalc.CalculateImageCRC(dataForCRC, len(dataForCRC)/4)
		fmt.Printf("  CRC without header (%d dwords): 0x%04x\n", len(dataForCRC)/4, crc4)
	}
}