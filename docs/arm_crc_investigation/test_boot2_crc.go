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
		fmt.Println("Usage: go run test_boot2_crc.go <firmware.bin>")
		os.Exit(1)
	}

	logger, _ := zap.NewDevelopment()
	reader, _ := parser.NewFirmwareReader(os.Args[1], logger)
	defer reader.Close()

	// Read BOOT2 at 0x1000
	// According to mstflint: size * 4 + 16 bytes are read
	// For size = 0xee0, that's 0xee0 * 4 + 16 = 0x3b90 bytes
	boot2Data, _ := reader.ReadSection(0x1000, 0x3b90)
	
	// Extract size from offset 4
	size := binary.BigEndian.Uint32(boot2Data[4:8])
	fmt.Printf("Size field: 0x%x (%d dwords)\n", size, size)
	
	// mstflint does TOCPUn which converts to CPU endianness
	// Let's create a dword array like mstflint
	dwordCount := size + 4
	dwords := make([]uint32, dwordCount)
	
	// Convert bytes to dwords (like mstflint's buff)
	for i := uint32(0); i < dwordCount; i++ {
		offset := i * 4
		if offset + 4 <= uint32(len(boot2Data)) {
			// mstflint reads as big-endian then converts to CPU
			dwords[i] = binary.BigEndian.Uint32(boot2Data[offset:offset+4])
		}
	}
	
	// CRC location according to mstflint: buff[size + 3]
	crcIndex := size + 3
	fmt.Printf("\nCRC location: dword index %d (0x%x)\n", crcIndex, crcIndex)
	if crcIndex < uint32(len(dwords)) {
		fmt.Printf("CRC value at that location: 0x%08x\n", dwords[crcIndex])
	}
	
	// Calculate CRC like mstflint
	crc := uint16(0xffff)
	
	// Process size + 3 dwords (CRC1n processes n-1 dwords)
	for i := uint32(0); i < size + 3; i++ {
		if i >= uint32(len(dwords)) {
			break
		}
		
		val := dwords[i]
		// Process 32 bits
		for j := 0; j < 32; j++ {
			if crc & 0x8000 != 0 {
				crc = ((crc << 1) | uint16(val >> 31)) ^ 0x100b
			} else {
				crc = (crc << 1) | uint16(val >> 31)
			}
			crc &= 0xffff
			val <<= 1
		}
	}
	
	// Finish
	for i := 0; i < 16; i++ {
		if crc & 0x8000 != 0 {
			crc = (crc << 1) ^ 0x100b
		} else {
			crc = crc << 1
		}
		crc &= 0xffff
	}
	
	// Final XOR
	crc = crc ^ 0xffff
	
	fmt.Printf("\nCalculated CRC: 0x%04x\n", crc)
	
	// Also check with our standard function
	crcCalc := parser.NewCRCCalculator()
	stdCRC := crcCalc.CalculateImageCRC(boot2Data, int(size + 4))
	fmt.Printf("Standard CalculateImageCRC: 0x%04x\n", stdCRC)
}