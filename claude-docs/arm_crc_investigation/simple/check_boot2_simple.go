//go:build ignore
// +build ignore

// Archived experimental code; excluded from build and vet
package main

import (
	"encoding/binary"
	"fmt"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"go.uber.org/zap"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run check_boot2_simple.go <firmware.bin>")
		os.Exit(1)
	}

	logger, _ := zap.NewDevelopment()
	reader, _ := parser.NewFirmwareReader(os.Args[1], logger)
	defer reader.Close()

	// Read BOOT2 at 0x1000
	boot2Data, _ := reader.ReadSection(0x1000, 0xee0)

	// The size field at offset 4
	sizeField := binary.BigEndian.Uint32(boot2Data[4:8])
	fmt.Printf("Size field: 0x%x\n", sizeField)

	// If mstflint interprets this as dwords, then:
	// Total size = (sizeField + 4) * 4 = (0xee0 + 4) * 4 = 0x3b90
	// But that's way larger than our actual BOOT2!

	// So the size field must NOT be 0xee0 dwords
	// Let's check if the size field is actually something else

	// What if the size is stored differently?
	// Let me check all fields in header
	fmt.Println("\nBOOT2 Header fields:")
	for i := 0; i < 16; i += 4 {
		val := binary.BigEndian.Uint32(boot2Data[i : i+4])
		fmt.Printf("  [0x%02x]: 0x%08x", i, val)
		if i == 4 {
			// Check if this could be dwords
			totalIfDwords := (val + 4) * 4
			fmt.Printf(" => If dwords: total size = (0x%x + 4) * 4 = 0x%x", val, totalIfDwords)
		}
		fmt.Println()
	}

	// Now let's think differently
	// What if 0xee0 IS the total size in bytes
	// And mstflint's code is generic but FS4 stores it differently?

	// Let's calculate what the size field SHOULD be according to mstflint
	// Total bytes = 0xee0
	// Total dwords = 0xee0 / 4 = 0x3b8
	// Size field = total dwords - 4 = 0x3b8 - 4 = 0x3b4

	expectedSizeField := uint32((0xee0 / 4) - 4)
	fmt.Printf("\nExpected size field (if total is 0xee0 bytes): 0x%x\n", expectedSizeField)

	// Now check if we can find 0x3b4 anywhere in the header
	for i := 0; i < 16; i++ {
		// Check as 32-bit BE
		if i <= 12 {
			val32 := binary.BigEndian.Uint32(boot2Data[i : i+4])
			if val32 == expectedSizeField {
				fmt.Printf("  Found 0x%x as 32-bit BE at offset %d!\n", expectedSizeField, i)
			}
		}
		// Check as 16-bit BE
		if i <= 14 {
			val16 := binary.BigEndian.Uint16(boot2Data[i : i+2])
			if uint32(val16) == expectedSizeField {
				fmt.Printf("  Found 0x%x as 16-bit BE at offset %d!\n", expectedSizeField, i)
			}
		}
	}

	// Actually, let's just check if this is FS3-style BOOT2
	// In FS3, the CRC might be at offset 8
	crcCalc := parser.NewCRCCalculator()

	// Calculate CRC on data from offset 12 onwards
	data := boot2Data[12 : 0xee0-4] // Skip header, exclude last dword
	crc := crcCalc.CalculateImageCRC(data, len(data)/4)

	fmt.Printf("\nCRC at offset 8: 0x%04x\n", boot2Data[8]<<8|boot2Data[9])
	fmt.Printf("Calculated CRC: 0x%04x\n", crc)

	// Wait! Let me check the actual FS4 firmware format
	// Maybe BOOT2 is stored as an ITOC entry?

	// Read ITOC at 0x5000
	itocData, _ := reader.ReadSection(0x5000, 0x1000)

	// Skip ITOC header (32 bytes)
	// Then look for BOOT2 entry
	fmt.Println("\nSearching for BOOT2 in ITOC entries...")
	for i := 32; i < 0x1000; i += 32 {
		if i+32 > len(itocData) {
			break
		}

		// First byte is type
		entryType := itocData[i]
		if entryType == 0xFF {
			break // End marker
		}

		// Get flash address (bits 161-189, 29 bits)
		// This is complex bit unpacking, let's just check if we find 0x1000
		flashAddrBytes := itocData[i+20 : i+24]
		// Quick check - if byte 20 has 0x10, might be our BOOT2
		if flashAddrBytes[0] == 0x00 && flashAddrBytes[1] == 0x10 {
			fmt.Printf("  Found potential BOOT2 entry at ITOC offset 0x%x\n", i)
			fmt.Printf("    Type: 0x%02x\n", entryType)
			fmt.Printf("    Entry bytes [20-23]: %02x %02x %02x %02x\n",
				flashAddrBytes[0], flashAddrBytes[1], flashAddrBytes[2], flashAddrBytes[3])
		}
	}
}
