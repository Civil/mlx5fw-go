package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/Civil/mlx5fw-go/pkg/parser"
)

func main() {
	// Original entry 22 (DBG_FW_PARAMS) from the firmware
	originalHex := "3200000800000000000000000000000000000000004fe92800007e0500006860"
	originalBytes, _ := hex.DecodeString(originalHex)
	
	// Modified entry 22 with updated offset
	modifiedHex := "3200000800000000000000000000000000000000004fe9c400007e050000fe86"
	modifiedBytes, _ := hex.DecodeString(modifiedHex)
	
	fmt.Println("Original entry:")
	fmt.Printf("  Hex: %s\n", originalHex)
	fmt.Printf("  Flash addr: 0x%x\n", binary.BigEndian.Uint32(originalBytes[20:24])>>3)
	
	fmt.Println("\nModified entry:")
	fmt.Printf("  Hex: %s\n", modifiedHex)
	fmt.Printf("  Flash addr: 0x%x\n", binary.BigEndian.Uint32(modifiedBytes[20:24])>>3)
	
	// Calculate CRC on first 30 bytes using different methods
	crcCalc := parser.NewCRCCalculator()
	
	// Try Hardware CRC
	hwCRC := crcCalc.CalculateHardwareCRC(originalBytes[:30])
	fmt.Printf("\nOriginal Hardware CRC: 0x%04x (expected: 0x6860)\n", hwCRC)
	
	hwCRCMod := crcCalc.CalculateHardwareCRC(modifiedBytes[:30])
	fmt.Printf("Modified Hardware CRC: 0x%04x (actual: 0x%04x)\n", hwCRCMod, binary.BigEndian.Uint16(modifiedBytes[30:32]))
	
	// Try Software CRC
	swCRC := crcCalc.CalculateSoftwareCRC16(originalBytes[:30])
	fmt.Printf("\nOriginal Software CRC16: 0x%04x\n", swCRC)
	
	swCRCMod := crcCalc.CalculateSoftwareCRC16(modifiedBytes[:30])
	fmt.Printf("Modified Software CRC16: 0x%04x\n", swCRCMod)
	
	// Try Image CRC (30 bytes = 7.5 dwords, so try 7 dwords)
	imgCRC := crcCalc.CalculateImageCRC(originalBytes[:28], 7)
	fmt.Printf("\nOriginal Image CRC (28 bytes): 0x%04x\n", imgCRC)
	
	// Check if the expected CRC matches any calculation
	if hwCRC == 0x6860 {
		fmt.Println("\n✓ Hardware CRC matches expected value!")
	}
	if swCRC == 0x6860 {
		fmt.Println("\n✓ Software CRC16 matches expected value!")
	}
	if imgCRC == 0x6860 {
		fmt.Println("\n✓ Image CRC matches expected value!")
	}
	
	// Let's also check what we're calculating vs what's in the file
	fmt.Printf("\nCRC in original entry: 0x%04x\n", binary.BigEndian.Uint16(originalBytes[30:32]))
	fmt.Printf("CRC in modified entry: 0x%04x\n", binary.BigEndian.Uint16(modifiedBytes[30:32]))
}