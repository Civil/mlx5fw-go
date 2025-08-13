//go:build ignore
// +build ignore

// Archived experimental code; excluded from build and vet
package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

func crc16_poly(data []byte, poly uint16) uint16 {
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

func main() {
	origFW, _ := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
	modFW, _ := os.ReadFile("sample_firmwares/franken_fw.bin")

	// CRCs as stored in firmware (big endian)
	origCRC := uint16(0x6885)
	modCRC := uint16(0x5220)

	fmt.Println("=== Testing specific data subsets ===\n")

	// Maybe the CRC only covers the board ID string itself
	origBoardID := []byte("MBF2M345A-HECO")
	modBoardID := []byte("MBF2M345A-VENOT_ES")

	// Test just the board ID strings
	polys := []uint16{0x1021, 0x8005, 0x3D65}

	for _, poly := range polys {
		origCalc := crc16_poly(origBoardID, poly)
		modCalc := crc16_poly(modBoardID, poly)

		fmt.Printf("Board ID only, poly 0x%04X: orig=0x%04X (want 0x%04X), mod=0x%04X (want 0x%04X)\n",
			poly, origCalc, origCRC, modCalc, modCRC)
	}

	// Maybe it includes the length field with the string
	fmt.Println("\n=== Testing with length prefix ===")

	origWithLen := append([]byte{0x02, 0x49}, origBoardID...)
	modWithLen := append([]byte{0x02, 0x49}, modBoardID...)

	for _, poly := range polys {
		origCalc := crc16_poly(origWithLen, poly)
		modCalc := crc16_poly(modWithLen, poly)

		fmt.Printf("With 0x0249 prefix, poly 0x%04X: orig=0x%04X (want 0x%04X), mod=0x%04X (want 0x%04X)\n",
			poly, origCalc, origCRC, modCalc, modCRC)
	}

	// Let's check if the actual section has a specific structure
	fmt.Println("\n=== Analyzing actual firmware structure ===")

	// Extract the region around board ID
	region := origFW[0x4D0:0x510]
	fmt.Println("Region 0x4D0-0x510:")
	for i := 0; i < len(region); i += 16 {
		fmt.Printf("%04X: ", 0x4D0+i)
		for j := 0; j < 16 && i+j < len(region); j++ {
			fmt.Printf("%02X ", region[i+j])
		}
		fmt.Print(" |")
		for j := 0; j < 16 && i+j < len(region); j++ {
			if region[i+j] >= 32 && region[i+j] < 127 {
				fmt.Printf("%c", region[i+j])
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println("|")
	}

	// Check if CRC is little endian
	fmt.Println("\n=== Checking endianness ===")
	crcBytesOrig := origFW[0x4FE:0x500]
	crcBytesMod := modFW[0x4FE:0x500]

	fmt.Printf("Original CRC bytes: %02X %02X (BE: 0x%04X, LE: 0x%04X)\n",
		crcBytesOrig[0], crcBytesOrig[1],
		binary.BigEndian.Uint16(crcBytesOrig),
		binary.LittleEndian.Uint16(crcBytesOrig))

	fmt.Printf("Modified CRC bytes: %02X %02X (BE: 0x%04X, LE: 0x%04X)\n",
		crcBytesMod[0], crcBytesMod[1],
		binary.BigEndian.Uint16(crcBytesMod),
		binary.LittleEndian.Uint16(crcBytesMod))

	// Try with fixed-size string field
	fmt.Println("\n=== Testing fixed-size string field ===")

	// Create fixed 32-byte field for board ID
	origFixed := make([]byte, 32)
	modFixed := make([]byte, 32)
	copy(origFixed, origBoardID)
	copy(modFixed, modBoardID)

	// Add the value after string (0x000109D8)
	origFixed = append(origFixed, 0x00, 0x01, 0x09, 0xD8, 0x00, 0x00)
	modFixed = append(modFixed, 0x00, 0x01, 0x09, 0xD8, 0x00, 0x00)

	for _, poly := range polys {
		origCalc := crc16_poly(origFixed, poly)
		modCalc := crc16_poly(modFixed, poly)

		fmt.Printf("Fixed 32-byte + value, poly 0x%04X: orig=0x%04X (want 0x%04X), mod=0x%04X (want 0x%04X)\n",
			poly, origCalc, origCRC, modCalc, modCRC)
	}
}
