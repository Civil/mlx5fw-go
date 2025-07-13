package main

import (
	"fmt"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"encoding/binary"
)

func main() {
	// Entry 0 from original firmware
	originalHex := "05015614007100000071619800000000000000000000700000000e0070000cb76"
	
	// Entry 0 from modified firmware (same content, different CRC)
	modifiedHex := "05015614007100000071619800000000000000000000700000000e0070006860"
	
	// Let's parse the correct bytes
	// The entry in the file shows: 0501 5614 0071 0000 0071 61d8 0000 0000 0000 0000 0000 7000 0000 e007 0000 cb76
	// But I had wrong data above. Let me use the correct hex from xxd
	originalHex = "050156140071000000716d800000000000000000000070000000e0070000cb76"
	modifiedHex = "050156140071000000716d800000000000000000000070000000e0070006860"
	
	// Wait, let me get the exact bytes from xxd output
	// 0501 5614 0071 0000 0071 61d8 0000 0000  
	// 0000 0000 0000 7000 0000 e007 0000 cb76
	originalBytes := []byte{
		0x05, 0x01, 0x56, 0x14, 0x00, 0x71, 0x00, 0x00,
		0x00, 0x71, 0x61, 0xd8, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x70, 0x00,
		0x00, 0x00, 0xe0, 0x07, 0x00, 0x00, 0xcb, 0x76,
	}
	
	fmt.Printf("Entry 0 (IRON_PREP_CODE):\n")
	fmt.Printf("  Type: 0x%02x\n", originalBytes[0])
	fmt.Printf("  Flash addr bytes: %02x %02x %02x %02x\n", 
		originalBytes[20], originalBytes[21], originalBytes[22], originalBytes[23])
	fmt.Printf("  Original CRC: 0x%02x%02x\n", originalBytes[30], originalBytes[31])
	
	// Calculate CRC on first 30 bytes
	crcCalc := parser.NewCRCCalculator()
	
	hwCRC := crcCalc.CalculateHardwareCRC(originalBytes[:30])
	fmt.Printf("\nCalculated Hardware CRC: 0x%04x\n", hwCRC)
	fmt.Printf("Expected (from file): 0x%04x\n", binary.BigEndian.Uint16(originalBytes[30:32]))
	
	if hwCRC == binary.BigEndian.Uint16(originalBytes[30:32]) {
		fmt.Println("✓ CRC matches!")
	} else {
		fmt.Println("✗ CRC mismatch!")
		
		// Let's try other CRC methods
		swCRC := crcCalc.CalculateSoftwareCRC16(originalBytes[:30])
		fmt.Printf("\nSoftware CRC16: 0x%04x\n", swCRC)
		
		// Try different byte counts
		for i := 28; i <= 30; i++ {
			if i % 4 == 0 {
				dwords := i / 4
				imgCRC := crcCalc.CalculateImageCRC(originalBytes[:i], dwords)
				fmt.Printf("Image CRC (%d bytes, %d dwords): 0x%04x\n", i, dwords, imgCRC)
			}
		}
	}
}