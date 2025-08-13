//go:build ignore
// +build ignore

// Archived experimental code; excluded from build and vet
package main

import (
	"fmt"
	"os"
)

// Standard CRC16 calculation
func crc16_ccitt(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

// CRC16 with configurable parameters
func crc16_custom(data []byte, poly, init uint16, refIn, refOut bool, xorOut uint16) uint16 {
	crc := init

	for _, b := range data {
		if refIn {
			b = reverseBits8(b)
		}
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc <<= 1
			}
		}
	}

	if refOut {
		crc = reverseBits16(crc)
	}

	return crc ^ xorOut
}

func reverseBits8(b byte) byte {
	var result byte
	for i := 0; i < 8; i++ {
		result = (result << 1) | (b & 1)
		b >>= 1
	}
	return result
}

func reverseBits16(n uint16) uint16 {
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

	// Target CRCs
	origCRC := uint16(0x6885)
	modCRC := uint16(0x5220)

	fmt.Println("=== Comprehensive CRC Search ===\n")

	// The board ID structure appears to be:
	// 0x4D6: 02 49 (length or type)
	// 0x4D8: Board ID string
	// 0x4FE: CRC16

	// Let's test different start/end points
	startPoints := []struct {
		name   string
		offset int
	}{
		{"From 0x4D6 (0249)", 0x4D6},
		{"From 0x4D8 (string)", 0x4D8},
		{"From 0x4D0", 0x4D0},
		{"From 0x4C0", 0x4C0},
		{"From 0x400", 0x400},
		{"From 0x300", 0x300},
	}

	endPoint := 0x4FE // CRC location

	// Test various CRC algorithms
	algorithms := []struct {
		name   string
		poly   uint16
		init   uint16
		refIn  bool
		refOut bool
		xorOut uint16
	}{
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
	}

	foundMatch := false

	for _, start := range startPoints {
		if start.offset >= endPoint {
			continue
		}

		origData := origFW[start.offset:endPoint]
		modData := modFW[start.offset:endPoint]

		for _, alg := range algorithms {
			origCalc := crc16_custom(origData, alg.poly, alg.init, alg.refIn, alg.refOut, alg.xorOut)
			modCalc := crc16_custom(modData, alg.poly, alg.init, alg.refIn, alg.refOut, alg.xorOut)

			if origCalc == origCRC && modCalc == modCRC {
				fmt.Printf("*** EXACT MATCH FOUND! ***\n")
				fmt.Printf("Start: %s (0x%X)\n", start.name, start.offset)
				fmt.Printf("End: 0x%X (before CRC)\n", endPoint)
				fmt.Printf("Length: %d bytes\n", len(origData))
				fmt.Printf("Algorithm: %s\n", alg.name)
				fmt.Printf("  Polynomial: 0x%04X\n", alg.poly)
				fmt.Printf("  Init: 0x%04X\n", alg.init)
				fmt.Printf("  RefIn: %v, RefOut: %v\n", alg.refIn, alg.refOut)
				fmt.Printf("  XorOut: 0x%04X\n", alg.xorOut)
				fmt.Printf("Original CRC: calculated=0x%04X, stored=0x%04X\n", origCalc, origCRC)
				fmt.Printf("Modified CRC: calculated=0x%04X, stored=0x%04X\n", modCalc, modCRC)

				fmt.Println("\nData being CRC'd (first 64 bytes):")
				for i := 0; i < 64 && i < len(origData); i += 16 {
					fmt.Printf("%04X: ", start.offset+i)
					for j := 0; j < 16 && i+j < 64 && i+j < len(origData); j++ {
						fmt.Printf("%02X ", origData[i+j])
					}
					fmt.Println()
				}

				foundMatch = true
				return
			}
		}
	}

	if !foundMatch {
		fmt.Println("No exact match found with standard algorithms.")
		fmt.Println("\nTrying byte-swapped CRC storage...")

		// Try with byte-swapped CRC values
		origCRCSwap := (origCRC >> 8) | (origCRC << 8)
		modCRCSwap := (modCRC >> 8) | (modCRC << 8)

		for _, start := range startPoints {
			if start.offset >= endPoint {
				continue
			}

			origData := origFW[start.offset:endPoint]
			modData := modFW[start.offset:endPoint]

			for _, alg := range algorithms {
				origCalc := crc16_custom(origData, alg.poly, alg.init, alg.refIn, alg.refOut, alg.xorOut)
				modCalc := crc16_custom(modData, alg.poly, alg.init, alg.refIn, alg.refOut, alg.xorOut)

				if origCalc == origCRCSwap && modCalc == modCRCSwap {
					fmt.Printf("*** MATCH FOUND (byte-swapped CRC)! ***\n")
					fmt.Printf("Start: %s (0x%X)\n", start.name, start.offset)
					fmt.Printf("Algorithm: %s\n", alg.name)
					fmt.Printf("CRC stored as little-endian\n")
					return
				}
			}
		}
	}
}
