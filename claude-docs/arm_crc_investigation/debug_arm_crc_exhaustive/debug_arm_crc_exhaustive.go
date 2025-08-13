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

// All known CRC16 variants including mstflint
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
	{"MSTFLINT", 0x100B, 0xFFFF, false, false, 0xFFFF},
	{"MSTFLINT-0", 0x100B, 0x0000, false, false, 0x0000},
	{"MSTFLINT-NOXOR", 0x100B, 0xFFFF, false, false, 0x0000},
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
	origFW, _ := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
	modFW, _ := os.ReadFile("sample_firmwares/franken_fw.bin")

	// Target CRCs
	origCRC := uint16(0x6885)
	modCRC := uint16(0x5220)

	fmt.Println("=== Exhaustive ARM CRC Search ===")
	fmt.Printf("Target CRCs: Original=0x%04X, Modified=0x%04X\n\n", origCRC, modCRC)

	// Find boundaries
	// First 0xFF byte (start of padding)
	firstFF := -1
	for i := 0x200; i < 0x400; i++ {
		if origFW[i] == 0xFF {
			firstFF = i
			fmt.Printf("First 0xFF found at: 0x%04X\n", firstFF)
			break
		}
	}

	// Board ID string location
	boardIDStart := 0x4D8
	boardIDEndOrig := 0x4E6 // End of "MBF2M345A-HECO"
	boardIDEndMod := 0x4EA  // End of "MBF2M345A-VENOT_ES"

	fmt.Printf("Board ID starts at: 0x%04X\n", boardIDStart)
	fmt.Printf("Original board ID ends at: 0x%04X\n", boardIDEndOrig)
	fmt.Printf("Modified board ID ends at: 0x%04X\n", boardIDEndMod)

	// Find next 0xFF after board ID
	nextFF := -1
	for i := 0x500; i < 0x600; i++ {
		if origFW[i] == 0xFF {
			nextFF = i
			fmt.Printf("Next 0xFF found at: 0x%04X\n", nextFF)
			break
		}
	}

	// CRC location
	crcLocation := 0x4FE
	fmt.Printf("CRC location: 0x%04X\n\n", crcLocation)

	fmt.Println("Starting exhaustive search...")
	fmt.Printf("Testing all combinations from 0x%04X to 0x%04X (start) and 0x%04X to 0x%04X (end)\n\n",
		firstFF, boardIDStart, boardIDEndOrig, nextFF)

	totalTests := 0
	foundMatch := false

	// Test all possible start positions (from first FF to board ID start)
	for start := firstFF; start < boardIDStart && start >= 0; start++ {
		// Test all possible end positions (from board ID end to next FF)
		for end := boardIDEndOrig; end <= nextFF && end <= len(origFW); end++ {
			if start >= end {
				continue
			}

			origData := origFW[start:end]
			modData := modFW[start:end]

			// For modified firmware, adjust end if it includes the longer string
			if end > boardIDEndOrig && end <= boardIDEndMod {
				// Use the modified string end position
				modEnd := boardIDEndMod + (end - boardIDEndOrig)
				if modEnd <= len(modFW) {
					modData = modFW[start:modEnd]
				}
			}

			// Test all CRC algorithms
			for _, params := range crcVariants {
				totalTests++

				var origCalc, modCalc uint16

				if params.Name == "MSTFLINT" || params.Name == "MSTFLINT-0" || params.Name == "MSTFLINT-NOXOR" {
					origCalc = mstflintSoftwareCRC(origData)
					modCalc = mstflintSoftwareCRC(modData)
				} else {
					origCalc = calcCRC16(origData, params)
					modCalc = calcCRC16(modData, params)
				}

				// Check for exact match
				if origCalc == origCRC && modCalc == modCRC {
					fmt.Printf("\n*** EXACT MATCH FOUND! ***\n")
					fmt.Printf("Start: 0x%04X, End: 0x%04X (Length: %d bytes)\n", start, end, len(origData))
					fmt.Printf("Algorithm: %s\n", params.Name)
					fmt.Printf("Parameters: Poly=0x%04X, Init=0x%04X, RefIn=%v, RefOut=%v, XorOut=0x%04X\n",
						params.Poly, params.Init, params.RefIn, params.RefOut, params.XorOut)
					fmt.Printf("Original CRC: calculated=0x%04X, stored=0x%04X\n", origCalc, origCRC)
					fmt.Printf("Modified CRC: calculated=0x%04X, stored=0x%04X\n", modCalc, modCRC)

					fmt.Println("\nData covered by CRC (first 64 bytes):")
					for i := 0; i < 64 && i < len(origData); i += 16 {
						fmt.Printf("%04X: ", start+i)
						for j := 0; j < 16 && i+j < 64 && i+j < len(origData); j++ {
							fmt.Printf("%02X ", origData[i+j])
						}
						fmt.Println()
					}

					foundMatch = true
					goto done
				}

				// Progress indicator every 10000 tests
				if totalTests%10000 == 0 {
					fmt.Printf("Tested %d combinations...\r", totalTests)
				}
			}
		}
	}

done:
	fmt.Printf("\nTotal combinations tested: %d\n", totalTests)

	if !foundMatch {
		fmt.Println("\nNo match found in exhaustive search.")
		fmt.Println("The CRC must use a proprietary algorithm not in our test set.")
	}
}
