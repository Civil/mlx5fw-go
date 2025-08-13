//go:build ignore
// +build ignore

// Archived experimental code; excluded from build and vet
package main

import (
	"fmt"
)

// Common CRC16 variants
var crcTables = map[string]struct {
	poly   uint16
	init   uint16
	xorOut uint16
	refIn  bool
	refOut bool
}{
	"CRC16-CCITT":       {0x1021, 0xFFFF, 0x0000, false, false},
	"CRC16-CCITT-FALSE": {0x1021, 0xFFFF, 0x0000, false, false},
	"CRC16-XMODEM":      {0x1021, 0x0000, 0x0000, false, false},
	"CRC16-KERMIT":      {0x1021, 0x0000, 0x0000, true, true},
	"CRC16-AUG-CCITT":   {0x1021, 0x1D0F, 0x0000, false, false},
	"CRC16-BUYPASS":     {0x8005, 0x0000, 0x0000, false, false},
	"CRC16-ARC":         {0x8005, 0x0000, 0x0000, true, true},
	"CRC16-MODBUS":      {0x8005, 0xFFFF, 0x0000, true, true},
	"CRC16-USB":         {0x8005, 0xFFFF, 0xFFFF, true, true},
	"CRC16-MAXIM":       {0x8005, 0x0000, 0xFFFF, true, true},
	"CRC16-DNP":         {0x3D65, 0x0000, 0xFFFF, true, true},
	"CRC16-EN-13757":    {0x3D65, 0x0000, 0xFFFF, false, false},
	"CRC16-GENIBUS":     {0x1021, 0xFFFF, 0xFFFF, false, false},
	"CRC16-T10-DIF":     {0x8BB7, 0x0000, 0x0000, false, false},
	"CRC16-TELEDISK":    {0xA097, 0x0000, 0x0000, false, false},
	"CRC16-CDMA2000":    {0xC867, 0xFFFF, 0x0000, false, false},
}

func reverseBits16(n uint16) uint16 {
	var result uint16
	for i := 0; i < 16; i++ {
		result = (result << 1) | (n & 1)
		n >>= 1
	}
	return result
}

func reverseBits8(n uint8) uint8 {
	var result uint8
	for i := 0; i < 8; i++ {
		result = (result << 1) | (n & 1)
		n >>= 1
	}
	return result
}

func calcCRC16(data []byte, poly, init, xorOut uint16, refIn, refOut bool) uint16 {
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

func main() {
	// Original data
	origData := []byte("MBF2M345A-HECO")
	// Modified data
	modData := []byte("MBF2M345A-VENOT_ES")

	origCRC := uint16(0x6885) // Big endian from firmware
	modCRC := uint16(0x5220)  // Big endian from firmware

	// Also test little endian interpretations
	origCRCLE := uint16(0x8568)
	modCRCLE := uint16(0x2052)

	fmt.Printf("Original data: %s\n", string(origData))
	fmt.Printf("Modified data: %s\n", string(modData))
	fmt.Printf("Original CRC (BE): 0x%04X, (LE): 0x%04X\n", origCRC, origCRCLE)
	fmt.Printf("Modified CRC (BE): 0x%04X, (LE): 0x%04X\n\n", modCRC, modCRCLE)

	// Test all CRC variants
	for name, params := range crcTables {
		origCalc := calcCRC16(origData, params.poly, params.init, params.xorOut, params.refIn, params.refOut)
		modCalc := calcCRC16(modData, params.poly, params.init, params.xorOut, params.refIn, params.refOut)

		if origCalc == origCRC || origCalc == origCRCLE {
			fmt.Printf("MATCH! %s: orig=0x%04X (matches %s endian)\n", name, origCalc,
				map[bool]string{true: "little", false: "big"}[origCalc == origCRCLE])
		}
		if modCalc == modCRC || modCalc == modCRCLE {
			fmt.Printf("MATCH! %s: mod=0x%04X (matches %s endian)\n", name, modCalc,
				map[bool]string{true: "little", false: "big"}[modCalc == modCRCLE])
		}
	}

	// Test with different data ranges (maybe it includes more than just the string)
	fmt.Println("\n--- Testing with extended data ranges ---")

	// Include the length field before the string (0x0249 = 585 bytes)
	extData1Orig := append([]byte{0x02, 0x49}, origData...)
	extData1Mod := append([]byte{0x02, 0x49}, modData...)

	for name, params := range crcTables {
		origCalc := calcCRC16(extData1Orig, params.poly, params.init, params.xorOut, params.refIn, params.refOut)
		modCalc := calcCRC16(extData1Mod, params.poly, params.init, params.xorOut, params.refIn, params.refOut)

		if origCalc == origCRC || origCalc == origCRCLE {
			fmt.Printf("MATCH with length! %s: orig=0x%04X\n", name, origCalc)
		}
		if modCalc == modCRC || modCalc == modCRCLE {
			fmt.Printf("MATCH with length! %s: mod=0x%04X\n", name, modCalc)
		}
	}
}
