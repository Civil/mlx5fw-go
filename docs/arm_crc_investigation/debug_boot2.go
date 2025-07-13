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
		fmt.Println("Usage: go run debug_boot2.go <firmware.bin>")
		os.Exit(1)
	}

	logger, _ := zap.NewDevelopment()
	
	// Open firmware file
	reader, err := parser.NewFirmwareReader(os.Args[1], logger)
	if err != nil {
		logger.Fatal("Failed to open firmware", zap.Error(err))
	}
	defer reader.Close()

	// Read magic and boot2 pointer
	magicData, _ := reader.ReadSection(0, 8)
	magic := binary.BigEndian.Uint64(magicData)
	fmt.Printf("Magic: 0x%016x\n", magic)

	// Read HW pointers - they start after magic (8 bytes)
	// Each pointer is 8 bytes (ptr + crc)
	hwData, _ := reader.ReadSection(0, 0x80)
	
	// Show all HW pointer entries
	fmt.Println("\nHW Pointers:")
	for i := 0; i < 16; i++ {
		offset := i * 8
		ptr := binary.BigEndian.Uint32(hwData[offset:offset+4])
		crc := binary.BigEndian.Uint32(hwData[offset+4:offset+8])
		fmt.Printf("  Position %d (0x%02x): ptr=0x%08x crc=0x%08x\n", i, offset, ptr, crc)
	}
	
	// Boot2 is typically at 0x1000
	boot2Ptr := uint32(0x1000)

	// Read boot2 header
	headerData, _ := reader.ReadSection(int64(boot2Ptr), 16)
	boot2Size := binary.BigEndian.Uint32(headerData[4:8])
	fmt.Printf("Boot2 size from header: 0x%x (%d bytes)\n", boot2Size, boot2Size)

	// Read full boot2
	boot2Data, _ := reader.ReadSection(int64(boot2Ptr), boot2Size)
	
	// Check last few dwords
	fmt.Println("\nLast 16 bytes of BOOT2:")
	for i := len(boot2Data) - 16; i < len(boot2Data); i += 4 {
		dword := binary.BigEndian.Uint32(boot2Data[i:i+4])
		fmt.Printf("  Offset 0x%x: 0x%08x\n", boot2Ptr+uint32(i), dword)
	}

	// Skip CRC calculations for now
	
	// Extract expected CRC from last dword
	lastDword := binary.BigEndian.Uint32(boot2Data[len(boot2Data)-4:])
	fmt.Printf("\nLast dword: 0x%08x\n", lastDword)
	fmt.Printf("  Lower 16 bits: 0x%04x\n", lastDword & 0xFFFF)
	fmt.Printf("  Upper 16 bits: 0x%04x\n", (lastDword >> 16) & 0xFFFF)

	// Check if BOOT2 end aligns with what mlx5fw-go thinks
	fmt.Printf("\nBOOT2 range: 0x%x - 0x%x\n", boot2Ptr, boot2Ptr+boot2Size-1)
	
	// Show reported size vs actual
	reportedEnd := boot2Ptr + 0xee0 - 1  // 0x1edf from the error
	fmt.Printf("mlx5fw-go reported end: 0x%x (size=0x%x)\n", reportedEnd, reportedEnd - boot2Ptr + 1)
	
	// Check the actual size being used by mlx5fw-go
	// The error shows 0x1000-0x1edf which is 0xee0 bytes
	fmt.Printf("\nSize comparison:\n")
	fmt.Printf("  Header size field: 0x%x (%d bytes)\n", boot2Size, boot2Size)
	fmt.Printf("  mlx5fw-go using size: 0x%x (%d bytes)\n", 0xee0, 0xee0)
	
	// The issue might be that BOOT2 needs special handling for signed firmware
	// Check if there's an image signature pointer
	imageSignaturePtr := binary.BigEndian.Uint32(hwData[0x58:0x5c])
	fmt.Printf("\nImage signature pointer: 0x%x\n", imageSignaturePtr)
	if imageSignaturePtr == 0x00000000 {
		fmt.Println("  => No signature pointer (unsigned firmware)")
	} else {
		fmt.Println("  => Has signature pointer (signed firmware)")
	}
}