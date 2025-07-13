package main

import (
	"encoding/hex"
	"fmt"
)

// getBits extracts bits from a byte array
func getBits(data []byte, bitOffset, bitCount int) uint32 {
	var result uint32
	for i := 0; i < bitCount; i++ {
		bitPos := bitOffset + i
		byteIdx := bitPos / 8
		bitIdx := 7 - (bitPos % 8)
		if data[byteIdx]&(1<<bitIdx) != 0 {
			result |= (1 << uint(bitCount-1-i))
		}
	}
	return result
}

func main() {
	// Look at the exact bytes where flash address is stored
	// Flash address is at bits 161-189 (29 bits)
	originalHex := "3200000800000000000000000000000000000000004fe92800007e0500006860"
	originalBytes, _ := hex.DecodeString(originalHex)
	
	// Let's examine byte by byte around the flash address
	fmt.Println("Original entry bytes 20-23 (where flash addr is):")
	for i := 20; i < 24; i++ {
		fmt.Printf("  Byte %d: 0x%02x\n", i, originalBytes[i])
	}
	
	// Flash address at bits 161-189
	// Bit 161 is in byte 20, bit 1
	// 161 / 8 = 20.125, so byte 20, bit 6 (7-1)
	flashAddr := getBits(originalBytes, 161, 29)
	fmt.Printf("\nFlash address (29 bits): 0x%x\n", flashAddr)
	fmt.Printf("Flash address in bytes: 0x%x\n", flashAddr << 2)
	
	// Let's also check the CRC at the end
	fmt.Printf("\nEntry CRC bytes 30-31: 0x%02x%02x\n", originalBytes[30], originalBytes[31])
	
	// Now let's look at what happens if we just change the flash address
	testBytes := make([]byte, 32)
	copy(testBytes, originalBytes)
	
	// Original flash addr is 0x4fe928
	// New flash addr should be 0x4fe9c4
	// Difference is 0x9c (156 bytes)
	
	fmt.Printf("\nOriginal flash addr: 0x4fe928\n")
	fmt.Printf("New flash addr: 0x4fe9c4\n")
	fmt.Printf("Difference: 0x%x bytes\n", 0x4fe9c4 - 0x4fe928)
	
	// The flash address in dwords
	origFlashDwords := uint32(0x4fe928 >> 2)
	newFlashDwords := uint32(0x4fe9c4 >> 2)
	fmt.Printf("\nOriginal flash addr in dwords: 0x%x\n", origFlashDwords)
	fmt.Printf("New flash addr in dwords: 0x%x\n", newFlashDwords)
	
	// Now let's see the exact bit pattern
	fmt.Printf("\nOriginal bytes 20-23: %02x %02x %02x %02x\n", 
		originalBytes[20], originalBytes[21], originalBytes[22], originalBytes[23])
	fmt.Printf("Expected for new addr: ?? ?? ?? ??\n")
}