//go:build ignore
// +build ignore

// Archived experimental code; excluded from build and vet
package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

// CRC16 algorithm parameters
type CRCParams struct {
	Name   string
	Poly   uint16
	Init   uint16
	RefIn  bool
	RefOut bool
	XorOut uint16
}

// All known CRC16 variants
var crcVariants = []CRCParams{
	{"CRC-16/CCITT-FALSE", 0x1021, 0xFFFF, false, false, 0x0000},
	{"CRC-16/ARC", 0x8005, 0x0000, true, true, 0x0000},
	{"CRC-16/AUG-CCITT", 0x1021, 0x1D0F, false, false, 0x0000},
	{"CRC-16/BUYPASS", 0x8005, 0x0000, false, false, 0x0000},
	{"CRC-16/CDMA2000", 0xC867, 0xFFFF, false, false, 0x0000},
	{"CRC-16/DDS-110", 0x8005, 0x800D, false, false, 0x0000},
	{"CRC-16/DECT-R", 0x0589, 0x0000, false, false, 0x0001},
	{"CRC-16/DECT-X", 0x0589, 0x0000, false, false, 0x0000},
	{"CRC-16/DNP", 0x3D65, 0x0000, true, true, 0xFFFF},
	{"CRC-16/EN-13757", 0x3D65, 0x0000, false, false, 0xFFFF},
	{"CRC-16/GENIBUS", 0x1021, 0xFFFF, false, false, 0xFFFF},
	{"CRC-16/MAXIM", 0x8005, 0x0000, true, true, 0xFFFF},
	{"CRC-16/MCRF4XX", 0x1021, 0xFFFF, true, true, 0x0000},
	{"CRC-16/RIELLO", 0x1021, 0xB2AA, true, true, 0x0000},
	{"CRC-16/T10-DIF", 0x8BB7, 0x0000, false, false, 0x0000},
	{"CRC-16/TELEDISK", 0xA097, 0x0000, false, false, 0x0000},
	{"CRC-16/TMS37157", 0x1021, 0x89EC, true, true, 0x0000},
	{"CRC-16/USB", 0x8005, 0xFFFF, true, true, 0xFFFF},
	{"CRC-A", 0x1021, 0xC6C6, true, true, 0x0000},
	{"CRC-16/KERMIT", 0x1021, 0x0000, true, true, 0x0000},
	{"CRC-16/MODBUS", 0x8005, 0xFFFF, true, true, 0x0000},
	{"CRC-16/X-25", 0x1021, 0xFFFF, true, true, 0xFFFF},
	{"CRC-16/XMODEM", 0x1021, 0x0000, false, false, 0x0000},
	{"MSTFLINT-POLY", 0x100B, 0xFFFF, false, false, 0xFFFF},
	{"MSTFLINT-POLY-0", 0x100B, 0x0000, false, false, 0x0000},
}

// Reverse bits in a byte
func reverseBits8(n uint8) uint8 {
	var result uint8
	for i := 0; i < 8; i++ {
		result = (result << 1) | (n & 1)
		n >>= 1
	}
	return result
}

// Reverse bits in a uint16
func reverseBits16(n uint16) uint16 {
	var result uint16
	for i := 0; i < 16; i++ {
		result = (result << 1) | (n & 1)
		n >>= 1
	}
	return result
}

// Generic CRC16 calculation
func calcCRC16(data []byte, params CRCParams) uint16 {
	crc := params.Init

	for _, b := range data {
		if params.RefIn {
			b = reverseBits8(b)
		}

		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ params.Poly
			} else {
				crc <<= 1
			}
		}
	}

	if params.RefOut {
		crc = reverseBits16(crc)
	}

	return crc ^ params.XorOut
}

// mstflint Software CRC16 (32-bit word processing)
func mstflintSoftwareCRC(data []byte) uint16 {
	const poly = uint32(0x100b)
	crc := uint32(0xffff)

	// Pad to 4-byte alignment
	padded := make([]byte, len(data))
	copy(padded, data)
	if len(padded)%4 != 0 {
		padding := 4 - (len(padded) % 4)
		for i := 0; i < padding; i++ {
			padded = append(padded, 0)
		}
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
			word = word << 1
		}
	}

	// Finish - process 16 more bits of zeros
	for i := 0; i < 16; i++ {
		if (crc & 0x80000000) != 0 {
			crc = (crc << 1) ^ poly
		} else {
			crc = crc << 1
		}
	}

	return uint16((crc >> 16) ^ 0xffff)
}

func main() {
	origFW, err := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
	if err != nil {
		panic(err)
	}

	modFW, err := os.ReadFile("sample_firmwares/franken_fw.bin")
	if err != nil {
		panic(err)
	}

	// Target CRCs
	origCRC := uint16(0x6885)
	modCRC := uint16(0x5220)

	fmt.Println("=== Comprehensive ARM Section CRC Analysis ===")
	fmt.Printf("Target CRCs: Original=0x%04X, Modified=0x%04X\n\n", origCRC, modCRC)

	// First, let's examine the section structure
	fmt.Println("Section structure analysis:")
	fmt.Printf("0x276-0x289: 0xFFFF padding\n")
	fmt.Printf("0x28C: Start of actual section data\n")
	fmt.Printf("0x4D8: Board ID string location\n")
	fmt.Printf("0x4FE: CRC16 location\n")
	fmt.Printf("0x540: Start of next 0xFFFF padding\n\n")

	// Define comprehensive test ranges covering the entire section
	testRanges := []struct {
		name  string
		start int
		end   int
	}{
		// Full section variations
		{"Full section after padding", 0x28C, 0x4FE},
		{"Full section to 0x4FC", 0x28C, 0x4FC},
		{"Full section to 0x500", 0x28C, 0x500},
		{"Including padding start", 0x276, 0x4FE},
		{"From 0x280", 0x280, 0x4FE},
		{"From 0x290", 0x290, 0x4FE},

		// Starting from different points in the 0x200s
		{"From 0x200", 0x200, 0x4FE},
		{"From 0x220", 0x220, 0x4FE},
		{"From 0x240", 0x240, 0x4FE},
		{"From 0x250", 0x250, 0x4FE},
		{"From 0x260", 0x260, 0x4FE},
		{"From 0x270", 0x270, 0x4FE},

		// Starting from 0x300s
		{"From 0x300", 0x300, 0x4FE},
		{"From 0x320", 0x320, 0x4FE},
		{"From 0x340", 0x340, 0x4FE},
		{"From 0x360", 0x360, 0x4FE},
		{"From 0x380", 0x380, 0x4FE},
		{"From 0x3A0", 0x3A0, 0x4FE},
		{"From 0x3C0", 0x3C0, 0x4FE},
		{"From 0x3E0", 0x3E0, 0x4FE},

		// Starting from 0x400s
		{"From 0x400", 0x400, 0x4FE},
		{"From 0x420", 0x420, 0x4FE},
		{"From 0x440", 0x440, 0x4FE},
		{"From 0x450", 0x450, 0x4FE},
		{"From 0x460", 0x460, 0x4FE},
		{"From 0x470", 0x470, 0x4FE},
		{"From 0x480", 0x480, 0x4FE},
		{"From 0x490", 0x490, 0x4FE},
		{"From 0x4A0", 0x4A0, 0x4FE},
		{"From 0x4B0", 0x4B0, 0x4FE},
		{"From 0x4C0", 0x4C0, 0x4FE},
		{"From 0x4D0", 0x4D0, 0x4FE},

		// Different end points with full section start
		{"0x28C to 0x4F0", 0x28C, 0x4F0},
		{"0x28C to 0x4F8", 0x28C, 0x4F8},
		{"0x28C to 0x4FA", 0x28C, 0x4FA},
		{"0x28C to 0x4FB", 0x28C, 0x4FB},
		{"0x28C to 0x4FC", 0x28C, 0x4FC},
		{"0x28C to 0x4FD", 0x28C, 0x4FD},
		{"0x28C to 0x4FE", 0x28C, 0x4FE},
		{"0x28C to 0x4FF", 0x28C, 0x4FF},
		{"0x28C to 0x500", 0x28C, 0x500},

		// Specific offsets that might be boundaries
		{"After FFFF pattern", 0x28A, 0x4FE},
		{"After zeros", 0x2A0, 0x4FE},
		{"From pattern start", 0x2BE, 0x4FE},
		{"After pattern", 0x2FE, 0x4FE},

		// Very specific byte counts from section start
		{"626 bytes", 0x28C, 0x4FE}, // 0x4FE - 0x28C = 626
		{"624 bytes", 0x28C, 0x4FC}, // 0x4FC - 0x28C = 624
		{"622 bytes", 0x28C, 0x4FA},
		{"620 bytes", 0x28C, 0x4F8},

		// Maybe CRC is elsewhere
		{"CRC at 0x538", 0x28C, 0x538},
		{"CRC at 0x53A", 0x28C, 0x53A},
		{"CRC at 0x53C", 0x28C, 0x53C},
		{"CRC at 0x53E", 0x28C, 0x53E},

		// Different alignments
		{"16-byte aligned", 0x280, 0x500},
		{"32-byte aligned", 0x280, 0x4E0},
		{"64-byte aligned", 0x280, 0x4C0},
		{"128-byte aligned", 0x280, 0x480},
		{"256-byte aligned", 0x200, 0x500},
	}

	matchFound := false

	// Test each range with each CRC variant
	for _, r := range testRanges {
		if r.start >= r.end || r.start < 0 || r.end > len(origFW) {
			continue
		}

		origData := origFW[r.start:r.end]
		modData := modFW[r.start:r.end]

		// Test standard CRC variants
		for _, params := range crcVariants {
			origCalc := calcCRC16(origData, params)
			modCalc := calcCRC16(modData, params)

			// Check for exact match
			if origCalc == origCRC && modCalc == modCRC {
				fmt.Printf("\n*** EXACT MATCH FOUND! ***\n")
				fmt.Printf("Range: %s (0x%04X-0x%04X, %d bytes)\n", r.name, r.start, r.end, len(origData))
				fmt.Printf("Algorithm: %s\n", params.Name)
				fmt.Printf("Parameters: Poly=0x%04X, Init=0x%04X, RefIn=%v, RefOut=%v, XorOut=0x%04X\n",
					params.Poly, params.Init, params.RefIn, params.RefOut, params.XorOut)
				fmt.Printf("CRC Results: Orig=0x%04X, Mod=0x%04X\n", origCalc, modCalc)

				fmt.Println("\nFirst 128 bytes of protected data:")
				for i := 0; i < 128 && i < len(origData); i += 16 {
					fmt.Printf("%04X: ", r.start+i)
					for j := 0; j < 16 && i+j < 128 && i+j < len(origData); j++ {
						fmt.Printf("%02X ", origData[i+j])
					}
					fmt.Print(" |")
					for j := 0; j < 16 && i+j < 128 && i+j < len(origData); j++ {
						if origData[i+j] >= 32 && origData[i+j] < 127 {
							fmt.Printf("%c", origData[i+j])
						} else {
							fmt.Print(".")
						}
					}
					fmt.Println("|")
				}

				matchFound = true
				return
			}
		}

		// Also test mstflint 32-bit word processing
		origCalc := mstflintSoftwareCRC(origData)
		modCalc := mstflintSoftwareCRC(modData)

		if origCalc == origCRC && modCalc == modCRC {
			fmt.Printf("\n*** MSTFLINT SOFTWARE CRC MATCH! ***\n")
			fmt.Printf("Range: %s (0x%04X-0x%04X, %d bytes)\n", r.name, r.start, r.end, len(origData))
			matchFound = true
			return
		}
	}

	if !matchFound {
		fmt.Println("\nNo exact match found even with full section coverage.")
		fmt.Println("\nPossibilities:")
		fmt.Println("1. CRC uses a proprietary polynomial not in our list")
		fmt.Println("2. Data is preprocessed before CRC (e.g., XOR with a key)")
		fmt.Println("3. CRC is calculated on non-contiguous data")
		fmt.Println("4. CRC location is not at 0x4FE")

		// Let's check if CRC might be at a different location
		fmt.Println("\nChecking for CRC at different locations...")
		crcLocations := []int{0x538, 0x53A, 0x53C, 0x53E, 0x500, 0x502}

		for _, crcLoc := range crcLocations {
			if crcLoc+2 <= len(origFW) {
				possibleCRC := binary.BigEndian.Uint16(origFW[crcLoc : crcLoc+2])
				fmt.Printf("Value at 0x%04X: 0x%04X\n", crcLoc, possibleCRC)
			}
		}
	}
}
