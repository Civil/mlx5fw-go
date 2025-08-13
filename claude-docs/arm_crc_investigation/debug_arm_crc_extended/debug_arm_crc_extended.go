//go:build ignore
// +build ignore

// Archived experimental code; excluded from build and vet
package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

// Test different CRC algorithms with various preprocessing
func testAllCRCVariations(origData, modData []byte, origCRC, modCRC uint16) bool {
	// Common CRC16 polynomials
	polys := []uint16{
		0x1021, // CCITT
		0x8005, // CRC16
		0x100B, // mstflint
		0x3D65, // DNP
		0x8BB7, // T10-DIF
		0xA097, // TELEDISK
		0xC867, // CDMA2000
		0x0589, // DECT
	}

	inits := []uint16{0x0000, 0xFFFF, 0x1D0F, 0x800D, 0x89EC, 0xB2AA, 0xC6C6}
	xorOuts := []uint16{0x0000, 0xFFFF}

	// Test with different preprocessing
	preprocessors := []struct {
		name string
		fn   func([]byte) []byte
	}{
		{"normal", func(d []byte) []byte { return d }},
		{"inverted-first-2", func(d []byte) []byte {
			if len(d) < 2 {
				return d
			}
			result := make([]byte, len(d))
			copy(result, d)
			result[0] = ^result[0]
			result[1] = ^result[1]
			return result
		}},
		{"inverted-all", func(d []byte) []byte {
			result := make([]byte, len(d))
			for i, b := range d {
				result[i] = ^b
			}
			return result
		}},
		{"byte-swapped-16", func(d []byte) []byte {
			result := make([]byte, len(d))
			for i := 0; i < len(d)-1; i += 2 {
				result[i] = d[i+1]
				result[i+1] = d[i]
			}
			if len(d)%2 == 1 {
				result[len(d)-1] = d[len(d)-1]
			}
			return result
		}},
	}

	// Test each combination
	for _, prep := range preprocessors {
		origPrep := prep.fn(origData)
		modPrep := prep.fn(modData)

		for _, poly := range polys {
			for _, init := range inits {
				for _, xorOut := range xorOuts {
					// Test normal bit order
					origCalc := crc16Basic(origPrep, poly, init, xorOut)
					modCalc := crc16Basic(modPrep, poly, init, xorOut)

					if origCalc == origCRC && modCalc == modCRC {
						fmt.Printf("\n*** MATCH FOUND! ***\n")
						fmt.Printf("Preprocessing: %s\n", prep.name)
						fmt.Printf("Poly=0x%04X, Init=0x%04X, XorOut=0x%04X\n", poly, init, xorOut)
						fmt.Printf("Normal bit order\n")
						return true
					}

					// Test reflected bit order
					origCalc = crc16Reflected(origPrep, poly, init, xorOut)
					modCalc = crc16Reflected(modPrep, poly, init, xorOut)

					if origCalc == origCRC && modCalc == modCRC {
						fmt.Printf("\n*** MATCH FOUND! ***\n")
						fmt.Printf("Preprocessing: %s\n", prep.name)
						fmt.Printf("Poly=0x%04X, Init=0x%04X, XorOut=0x%04X\n", poly, init, xorOut)
						fmt.Printf("Reflected bit order\n")
						return true
					}
				}
			}
		}
	}

	return false
}

// Basic CRC16 (MSB first)
func crc16Basic(data []byte, poly, init, xorOut uint16) uint16 {
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
	return crc ^ xorOut
}

// Reflected CRC16 (LSB first)
func crc16Reflected(data []byte, poly, init, xorOut uint16) uint16 {
	// Reflect polynomial
	polyRef := reflect16(poly)
	crc := init

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
	return crc ^ xorOut
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

	// Also test with CRC as little endian
	origCRCLE := uint16(0x8568)
	modCRCLE := uint16(0x2052)

	fmt.Println("=== Extended ARM CRC Analysis ===")
	fmt.Printf("Target CRCs (BE): Original=0x%04X, Modified=0x%04X\n", origCRC, modCRC)
	fmt.Printf("Target CRCs (LE): Original=0x%04X, Modified=0x%04X\n\n", origCRCLE, modCRCLE)

	// Extended test ranges including overlapping and non-obvious ranges
	testRanges := []struct {
		name  string
		start int
		end   int
	}{
		// Core ranges
		{"Board ID exact", 0x4D8, 0x4E6},
		{"Board ID + nulls", 0x4D8, 0x4F8},
		{"With value", 0x4D8, 0x4FC},
		{"With length", 0x4D6, 0x4FC},
		{"Full struct", 0x4D0, 0x4FE},

		// Overlapping ranges
		{"Overlap 1", 0x4D4, 0x4FA},
		{"Overlap 2", 0x4D5, 0x4FB},
		{"Overlap 3", 0x4D7, 0x4FD},

		// Non-aligned starts
		{"From 0x4D1", 0x4D1, 0x4FE},
		{"From 0x4D2", 0x4D2, 0x4FE},
		{"From 0x4D3", 0x4D3, 0x4FE},
		{"From 0x4D4", 0x4D4, 0x4FE},
		{"From 0x4D5", 0x4D5, 0x4FE},
		{"From 0x4D7", 0x4D7, 0x4FE},
		{"From 0x4D9", 0x4D9, 0x4FE},

		// Different lengths from string start
		{"String+35", 0x4D8, 0x4FB},
		{"String+36", 0x4D8, 0x4FC},
		{"String+37", 0x4D8, 0x4FD},
		{"String+38", 0x4D8, 0x4FE},

		// Before the main structure
		{"Pre-struct 16", 0x4C0, 0x4FE},
		{"Pre-struct 8", 0x4C8, 0x4FE},
		{"Pre-struct 4", 0x4CC, 0x4FE},
		{"Pre-struct 2", 0x4CE, 0x4FE},

		// Very specific byte counts
		{"39 bytes", 0x4D7, 0x4FE},
		{"40 bytes", 0x4D6, 0x4FE},
		{"41 bytes", 0x4D5, 0x4FE},
		{"42 bytes", 0x4D4, 0x4FE},
		{"43 bytes", 0x4D3, 0x4FE},
		{"44 bytes", 0x4D2, 0x4FE},
		{"45 bytes", 0x4D1, 0x4FE},
		{"46 bytes", 0x4D0, 0x4FE},

		// Maybe CRC protects data after itself
		{"After CRC 16", 0x500, 0x510},
		{"After CRC 32", 0x500, 0x520},
		{"With CRC", 0x4FE, 0x510},
	}

	found := false

	// Test each range
	for _, r := range testRanges {
		if r.start >= r.end || r.start < 0 || r.end > len(origFW) {
			continue
		}

		origData := origFW[r.start:r.end]
		modData := modFW[r.start:r.end]

		// Test with big endian CRC
		if testAllCRCVariations(origData, modData, origCRC, modCRC) {
			fmt.Printf("Range: %s (0x%04X-0x%04X, %d bytes)\n", r.name, r.start, r.end, len(origData))
			found = true
			break
		}

		// Test with little endian CRC
		if testAllCRCVariations(origData, modData, origCRCLE, modCRCLE) {
			fmt.Printf("Range: %s (0x%04X-0x%04X, %d bytes)\n", r.name, r.start, r.end, len(origData))
			fmt.Println("CRC stored as LITTLE ENDIAN!")
			found = true
			break
		}
	}

	if !found {
		fmt.Println("\nNo match found with extended search.")
		fmt.Println("\nTrying mstflint-specific 32-bit word processing...")

		// Test mstflint's exact algorithm on key ranges
		for _, r := range testRanges[:20] { // Test first 20 ranges
			if r.start >= r.end || r.start < 0 || r.end > len(origFW) {
				continue
			}

			origData := origFW[r.start:r.end]
			modData := modFW[r.start:r.end]

			origCalc := mstflintCRC32(origData)
			modCalc := mstflintCRC32(modData)

			if origCalc == origCRC && modCalc == modCRC {
				fmt.Printf("\n*** MSTFLINT 32-bit MATCH! ***\n")
				fmt.Printf("Range: %s\n", r.name)
				found = true
				break
			}
		}
	}

	if !found {
		fmt.Println("\nConclusion: The CRC algorithm is proprietary and not any standard variant.")
	}
}

// mstflint's 32-bit word CRC
func mstflintCRC32(data []byte) uint16 {
	const poly = uint32(0x100b)
	crc := uint32(0xffff)

	// Pad to 4-byte alignment
	padded := make([]byte, len(data))
	copy(padded, data)
	for len(padded)%4 != 0 {
		padded = append(padded, 0)
	}

	// Process 32-bit words
	for i := 0; i < len(padded); i += 4 {
		word := binary.BigEndian.Uint32(padded[i : i+4])

		for j := 0; j < 32; j++ {
			if ((crc ^ word) & 0x80000000) != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc = crc << 1
			}
			word <<= 1
		}
	}

	// Finish
	for i := 0; i < 16; i++ {
		if (crc & 0x80000000) != 0 {
			crc = (crc << 1) ^ poly
		} else {
			crc = crc << 1
		}
	}

	return uint16((crc >> 16) ^ 0xffff)
}
