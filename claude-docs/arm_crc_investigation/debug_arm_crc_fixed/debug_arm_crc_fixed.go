//go:build ignore
// +build ignore

// Archived experimental code; excluded from build and vet
package main

import (
	"fmt"
	"os"
)

// Simple CRC16 implementation
func crc16(data []byte, poly uint16) uint16 {
	crc := uint16(0xFFFF)

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

	return crc
}

func crc16WithInit(data []byte, poly, init uint16) uint16 {
	crc := init

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

	return crc
}

func main() {
	// Read firmware files
	origFW, _ := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
	modFW, _ := os.ReadFile("sample_firmwares/franken_fw.bin")

	// Target CRCs
	origCRC := uint16(0x6885)
	modCRC := uint16(0x5220)

	fmt.Println("=== Testing fixed-size structure hypothesis ===\n")

	// The structure seems to be:
	// 0x4d0: 6 bytes padding/flags
	// 0x4d6: 2 bytes length (0x0249)
	// 0x4d8: N bytes string (null-padded to fixed size)
	// 0x4f8: 4 bytes unknown value (0x000109d8)
	// 0x4fc: 2 bytes padding
	// 0x4fe: 2 bytes CRC16

	// Total size: 0x500 - 0x4d0 = 48 bytes

	// Let's assume the string field is fixed at 32 bytes (0x4d8 to 0x4f8)
	// And CRC covers everything except the CRC itself

	// Extract just the fixed structure
	origStruct := make([]byte, 46) // 48 - 2 (CRC)
	modStruct := make([]byte, 46)

	copy(origStruct, origFW[0x4d0:0x4fe])
	copy(modStruct, modFW[0x4d0:0x4fe])

	fmt.Printf("Original structure (46 bytes):\n")
	for i := 0; i < len(origStruct); i += 16 {
		fmt.Printf("%02x: ", i)
		for j := 0; j < 16 && i+j < len(origStruct); j++ {
			fmt.Printf("%02x ", origStruct[i+j])
		}
		fmt.Println()
	}

	fmt.Printf("\nModified structure (46 bytes):\n")
	for i := 0; i < len(modStruct); i += 16 {
		fmt.Printf("%02x: ", i)
		for j := 0; j < 16 && i+j < len(modStruct); j++ {
			fmt.Printf("%02x ", modStruct[i+j])
		}
		fmt.Println()
	}

	// Common CRC16 polynomials
	polys := []uint16{0x1021, 0x8005, 0x3D65, 0x8BB7, 0xA097}
	inits := []uint16{0x0000, 0xFFFF}

	fmt.Println("\n=== Testing CRC16 on fixed structure ===")

	for _, poly := range polys {
		for _, init := range inits {
			origCalc := crc16WithInit(origStruct, poly, init)
			modCalc := crc16WithInit(modStruct, poly, init)

			if origCalc == origCRC && modCalc == modCRC {
				fmt.Printf("*** EXACT MATCH! Poly=0x%04X, Init=0x%04X ***\n", poly, init)
				fmt.Printf("Original: calc=0x%04X, expected=0x%04X\n", origCalc, origCRC)
				fmt.Printf("Modified: calc=0x%04X, expected=0x%04X\n", modCalc, modCRC)
			}
		}
	}

	// Maybe the string portion has a different fixed size
	// Let's try with just the non-zero part
	fmt.Println("\n=== Testing variable string with fixed total size ===")

	// Create structure with normalized string field
	type BootArgs struct {
		Padding  [6]byte
		Length   uint16
		BoardID  [32]byte // Fixed size field
		Unknown  uint32
		Padding2 uint16
		CRC      uint16
	}

	// Try interpreting as a C-style structure where string is null-terminated
	// but field is fixed size

	for strFieldSize := 16; strFieldSize <= 36; strFieldSize += 2 {
		// Build test structure
		origTest := make([]byte, 0)
		modTest := make([]byte, 0)

		// Add header (8 bytes)
		origTest = append(origTest, origFW[0x4d0:0x4d8]...)
		modTest = append(modTest, modFW[0x4d0:0x4d8]...)

		// Add string field (fixed size)
		strField := make([]byte, strFieldSize)
		copy(strField, origFW[0x4d8:])
		origTest = append(origTest, strField...)

		strField = make([]byte, strFieldSize)
		copy(strField, modFW[0x4d8:])
		modTest = append(modTest, strField...)

		// Add the rest if within bounds
		nextOffset := 0x4d8 + strFieldSize
		if nextOffset < 0x4fe {
			origTest = append(origTest, origFW[nextOffset:0x4fe]...)
			modTest = append(modTest, modFW[nextOffset:0x4fe]...)
		}

		// Test CRC
		for _, poly := range polys {
			origCalc := crc16WithInit(origTest, poly, 0xFFFF)
			modCalc := crc16WithInit(modTest, poly, 0xFFFF)

			if origCalc == origCRC && modCalc == modCRC {
				fmt.Printf("*** MATCH with string field size %d! Poly=0x%04X ***\n", strFieldSize, poly)
			}
		}
	}
}
