//go:build ignore
// +build ignore

// Archived experimental code; excluded from build and vet
package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

// Test different CRC variations
func crcVariant1(data []byte) uint16 {
	// Standard mstflint but process as 16-bit words
	const poly = uint16(0x100b)
	crc := uint16(0xffff)

	// Process as 16-bit words
	for i := 0; i < len(data); i += 2 {
		var word uint16
		if i+1 < len(data) {
			word = binary.BigEndian.Uint16(data[i : i+2])
		} else {
			word = uint16(data[i]) << 8
		}

		for j := 0; j < 16; j++ {
			if ((crc ^ word) & 0x8000) != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc = crc << 1
			}
			word = word << 1
		}
	}

	return crc ^ 0xffff
}

func crcVariant2(data []byte) uint16 {
	// Process byte by byte with mstflint poly
	const poly = uint16(0x100b)
	crc := uint16(0xffff)

	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc <<= 1
			}
		}
	}

	return crc ^ 0xffff
}

func crcVariant3(data []byte) uint16 {
	// Little endian processing
	const poly = uint16(0x100b)
	crc := uint16(0xffff)

	for i := 0; i < len(data); i += 2 {
		var word uint16
		if i+1 < len(data) {
			word = binary.LittleEndian.Uint16(data[i : i+2])
		} else {
			word = uint16(data[i])
		}

		for j := 0; j < 16; j++ {
			if ((crc ^ word) & 0x8000) != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc = crc << 1
			}
			word = word << 1
		}
	}

	return crc ^ 0xffff
}

func main() {
	origFW, _ := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
	modFW, _ := os.ReadFile("sample_firmwares/franken_fw.bin")

	origCRC := uint16(0x6885)
	modCRC := uint16(0x5220)

	fmt.Println("=== Testing ARM-specific CRC variations ===\n")

	// Focus on the actual data that changes
	// String field appears to be exactly 32 bytes at 0x4D8-0x4F8

	// Test 1: Just the 32-byte string field
	origString := origFW[0x4D8:0x4F8]
	modString := modFW[0x4D8:0x4F8]

	fmt.Println("Test 1: 32-byte string field only")
	fmt.Printf("Variant1: orig=0x%04X, mod=0x%04X\n", crcVariant1(origString), crcVariant1(modString))
	fmt.Printf("Variant2: orig=0x%04X, mod=0x%04X\n", crcVariant2(origString), crcVariant2(modString))
	fmt.Printf("Variant3: orig=0x%04X, mod=0x%04X\n", crcVariant3(origString), crcVariant3(modString))

	// Test 2: String + following value (36 bytes)
	origData2 := origFW[0x4D8:0x4FC]
	modData2 := modFW[0x4D8:0x4FC]

	fmt.Println("\nTest 2: String + value (36 bytes)")
	fmt.Printf("Variant1: orig=0x%04X, mod=0x%04X\n", crcVariant1(origData2), crcVariant1(modData2))
	fmt.Printf("Variant2: orig=0x%04X, mod=0x%04X\n", crcVariant2(origData2), crcVariant2(modData2))
	v2orig := crcVariant2(origData2)
	v2mod := crcVariant2(modData2)
	if v2orig == origCRC && v2mod == modCRC {
		fmt.Println("*** MATCH FOUND with Variant2! ***")
	}

	// Test 3: With length prefix
	origData3 := origFW[0x4D6:0x4FC]
	modData3 := modFW[0x4D6:0x4FC]

	fmt.Println("\nTest 3: With length prefix (38 bytes)")
	fmt.Printf("Variant1: orig=0x%04X, mod=0x%04X\n", crcVariant1(origData3), crcVariant1(modData3))
	fmt.Printf("Variant2: orig=0x%04X, mod=0x%04X\n", crcVariant2(origData3), crcVariant2(modData3))

	// Test specific ranges that might work
	fmt.Println("\n=== Testing specific data patterns ===")

	// Maybe the board ID has a specific format
	fmt.Printf("\nAnalyzing board ID structure:\n")
	fmt.Printf("Original: %s\n", string(origFW[0x4D8:0x4E6]))
	fmt.Printf("Modified: %s\n", string(modFW[0x4D8:0x4EA]))

	// The pattern seems to be: MBF2M345A-XXXX
	// Let's check if CRC covers a specific portion

	ranges := []struct {
		name  string
		start int
		end   int
	}{
		{"From type prefix", 0x4D4, 0x4FC},
		{"Aligned 40 bytes", 0x4D6, 0x4FE},
		{"Just changed part", 0x4E2, 0x4FE}, // Where HECO starts
		{"Full fixed struct", 0x4D0, 0x4FC},
	}

	for _, r := range ranges {
		origData := origFW[r.start:r.end]
		modData := modFW[r.start:r.end]

		v1o := crcVariant1(origData)
		v1m := crcVariant1(modData)
		v2o := crcVariant2(origData)
		v2m := crcVariant2(modData)

		fmt.Printf("\n%s (0x%X-0x%X, %d bytes):\n", r.name, r.start, r.end, len(origData))
		fmt.Printf("  V1: orig=0x%04X, mod=0x%04X\n", v1o, v1m)
		fmt.Printf("  V2: orig=0x%04X, mod=0x%04X", v2o, v2m)

		if v2o == origCRC && v2m == modCRC {
			fmt.Printf(" *** MATCH! ***")
		}
		fmt.Println()
	}
}
