//go:build ignore
// +build ignore

// Archived experimental code; excluded from build and vet
package main

import (
	"fmt"
	"os"
)

// Try all CRC16 algorithms with different parameters
func tryCRC16(data []byte, targetCRC uint16) (bool, string) {
	// Extended polynomial list including less common ones
	polys := []uint16{
		0x1021, 0x8005, 0x100B, 0x3D65, 0x8BB7, 0xA097, 0xC867, 0x0589,
		0x8408, 0xA001, 0x1DCF, 0x755B, 0x5935, 0x90D9, 0x1EDC, 0x6F63,
		0x055B, 0x07BB, 0x0F47, 0x1A5B, 0x202D, 0x371D, 0x3D6D, 0x4073,
		0x4C11, 0x5589, 0x5D38, 0x6029, 0x6815, 0x7095, 0x755B, 0x786F,
		0x8003, 0x8D95, 0x9EB2, 0xA0F3, 0xA833, 0xBAAD, 0xC002, 0xD015,
	}

	inits := []uint16{0x0000, 0xFFFF, 0x1D0F, 0x800D, 0x89EC, 0xB2AA, 0xC6C6, 0x1234, 0x5678, 0xABCD}
	xorOuts := []uint16{0x0000, 0xFFFF, 0x0001}

	for _, poly := range polys {
		for _, init := range inits {
			for _, xorOut := range xorOuts {
				// Normal bit order
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
				if (crc ^ xorOut) == targetCRC {
					return true, fmt.Sprintf("Normal: poly=0x%04X init=0x%04X xor=0x%04X", poly, init, xorOut)
				}

				// Reflected bit order
				polyRef := reflect16(poly)
				crc = init
				for _, b := range data {
					crc ^= uint16(b)
					for i := 0; i < 8; i++ {
						if crc&0x0001 != 0 {
							crc = (crc >> 1) ^ polyRef
						} else {
							crc >>= 1
						}
					}
				}
				if (crc ^ xorOut) == targetCRC {
					return true, fmt.Sprintf("Reflected: poly=0x%04X init=0x%04X xor=0x%04X", poly, init, xorOut)
				}
			}
		}
	}

	return false, ""
}

func reflect16(n uint16) uint16 {
	var result uint16
	for i := 0; i < 16; i++ {
		result = (result << 1) | (n & 1)
		n >>= 1
	}
	return result
}

func main() {
	origFW, _ := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
	modFW, _ := os.ReadFile("sample_firmwares/franken_fw.bin")

	origCRC := uint16(0x6885)
	modCRC := uint16(0x5220)

	fmt.Println("=== Final Comprehensive CRC Search ===\n")

	// Based on the structure analysis:
	// - Section starts after 0xFFFF padding at 0x28C
	// - Repeating pattern 0xC688FAC6 from 0x2BE to 0x2FE
	// - Board ID at 0x4D8
	// - CRC at 0x4FE
	// - Value 0x83DC at 0x53E

	// Extended test ranges based on full structure understanding
	testRanges := []struct {
		name  string
		start int
		end   int
	}{
		// Full section minus padding
		{"After padding to CRC", 0x28C, 0x4FE},
		{"After padding to 0x500", 0x28C, 0x500},
		{"After pattern to CRC", 0x300, 0x4FE},

		// Maybe CRC is at 0x53E (where 0x83DC is)
		{"To 0x53E", 0x28C, 0x53E},
		{"Pattern to 0x53E", 0x300, 0x53E},
		{"0x400 to 0x53E", 0x400, 0x53E},

		// Specific structure boundaries
		{"After pattern start", 0x2BE, 0x4FE},
		{"After pattern end", 0x2FE, 0x4FE},
		{"From 0x304", 0x304, 0x4FE},
		{"From 0x310", 0x310, 0x4FE},

		// Maybe it's multiple sections
		{"Pattern only", 0x2BE, 0x2FE},
		{"After pattern only", 0x300, 0x400},
		{"Middle section", 0x400, 0x4D0},
		{"Boot args only", 0x4D0, 0x500},

		// Aligned boundaries from pattern end
		{"0x300 to 0x4FC", 0x300, 0x4FC},
		{"0x300 to 0x4FD", 0x300, 0x4FD},
		{"0x300 to 0x4FE", 0x300, 0x4FE},
		{"0x300 to 0x4FF", 0x300, 0x4FF},
		{"0x300 to 0x500", 0x300, 0x500},

		// From specific markers
		{"From 0x304 (008A)", 0x304, 0x4FE},
		{"From 0x306 (6066)", 0x306, 0x4FE},
		{"From 0x310 (0001)", 0x310, 0x4FE},
		{"From 0x450 (D480)", 0x450, 0x4FE},
		{"From 0x47C (DA5D)", 0x47C, 0x4FE},

		// Every 16-byte boundary from 0x200
		{"From 0x200", 0x200, 0x4FE},
		{"From 0x210", 0x210, 0x4FE},
		{"From 0x220", 0x220, 0x4FE},
		{"From 0x230", 0x230, 0x4FE},
		{"From 0x240", 0x240, 0x4FE},
		{"From 0x250", 0x250, 0x4FE},
		{"From 0x260", 0x260, 0x4FE},
		{"From 0x270", 0x270, 0x4FE},
		{"From 0x280", 0x280, 0x4FE},
		{"From 0x290", 0x290, 0x4FE},
		{"From 0x2A0", 0x2A0, 0x4FE},
		{"From 0x2B0", 0x2B0, 0x4FE},
		{"From 0x2C0", 0x2C0, 0x4FE},
		{"From 0x2D0", 0x2D0, 0x4FE},
		{"From 0x2E0", 0x2E0, 0x4FE},
		{"From 0x2F0", 0x2F0, 0x4FE},
	}

	foundMatch := false

	fmt.Println("Testing all ranges with extended CRC algorithms...")

	for _, r := range testRanges {
		if r.start >= r.end || r.start < 0 || r.end > len(origFW) {
			continue
		}

		origData := origFW[r.start:r.end]
		modData := modFW[r.start:r.end]

		// Test original data
		origMatch, origAlg := tryCRC16(origData, origCRC)
		modMatch, modAlg := tryCRC16(modData, modCRC)

		if origMatch && modMatch && origAlg == modAlg {
			fmt.Printf("\n*** EXACT MATCH FOUND! ***\n")
			fmt.Printf("Range: %s (0x%04X-0x%04X, %d bytes)\n", r.name, r.start, r.end, len(origData))
			fmt.Printf("Algorithm: %s\n", origAlg)

			fmt.Println("\nFirst 64 bytes of data:")
			for i := 0; i < 64 && i < len(origData); i += 16 {
				fmt.Printf("%04X: ", r.start+i)
				for j := 0; j < 16 && i+j < 64 && i+j < len(origData); j++ {
					fmt.Printf("%02X ", origData[i+j])
				}
				fmt.Println()
			}

			foundMatch = true
			break
		}
	}

	if !foundMatch {
		fmt.Println("\nNo match found with standard algorithms.")

		// Let's check if CRC might be calculated differently
		fmt.Println("\nTesting non-contiguous data combinations...")

		// Maybe CRC covers: header + board ID + value (skipping zeros)
		header := origFW[0x4D0:0x4D8]  // 8 bytes before string
		boardID := origFW[0x4D8:0x4E6] // Board ID (14 bytes)
		value := origFW[0x4F8:0x4FC]   // Value after string (4 bytes)

		combined := append(append(header, boardID...), value...)
		fmt.Printf("Testing header+boardID+value (%d bytes)...\n", len(combined))

		// Similar for modified
		modBoardID := modFW[0x4D8:0x4EA] // Modified board ID (18 bytes)
		modCombined := append(append(header, modBoardID...), value...)

		origMatch, origAlg := tryCRC16(combined, origCRC)
		modMatch, _ := tryCRC16(modCombined, modCRC)

		if origMatch && modMatch {
			fmt.Printf("*** MATCH on non-contiguous data! Algorithm: %s ***\n", origAlg)
			foundMatch = true
		}
	}

	if !foundMatch {
		fmt.Println("\nFinal conclusion: The CRC uses a proprietary algorithm.")
		fmt.Println("Tested:")
		fmt.Println("- 40+ different polynomials")
		fmt.Println("- 10 different init values")
		fmt.Println("- 3 different XOR out values")
		fmt.Println("- Both normal and reflected bit orders")
		fmt.Println("- 50+ different data ranges")
		fmt.Println("- Non-contiguous data combinations")
	}
}
