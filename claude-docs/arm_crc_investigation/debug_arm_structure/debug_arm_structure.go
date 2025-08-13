//go:build ignore
// +build ignore

// Archived experimental code; excluded from build and vet
package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

func main() {
	// Read firmware
	fw, err := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
	if err != nil {
		panic(err)
	}

	// Let's look for patterns around our section
	fmt.Println("=== Searching for command line argument patterns ===\n")

	// The string at 0x4d8 looks like a board/model identifier
	// Let's see if there are other references to it or similar structures

	// Check what's before our section
	fmt.Println("Data before our section (0x490-0x4d0):")
	for i := 0x490; i < 0x4d0; i += 16 {
		fmt.Printf("%04x: ", i)
		for j := 0; j < 16 && i+j < 0x4d0; j++ {
			fmt.Printf("%02x ", fw[i+j])
		}
		fmt.Print(" |")
		for j := 0; j < 16 && i+j < 0x4d0; j++ {
			b := fw[i+j]
			if b >= 32 && b < 127 {
				fmt.Printf("%c", b)
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println("|")
	}

	// Check what's after our section
	fmt.Println("\nData after our section (0x500-0x540):")
	for i := 0x500; i < 0x540 && i < len(fw); i += 16 {
		fmt.Printf("%04x: ", i)
		for j := 0; j < 16 && i+j < 0x540 && i+j < len(fw); j++ {
			fmt.Printf("%02x ", fw[i+j])
		}
		fmt.Print(" |")
		for j := 0; j < 16 && i+j < 0x540 && i+j < len(fw); j++ {
			b := fw[i+j]
			if b >= 32 && b < 127 {
				fmt.Printf("%c", b)
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println("|")
	}

	// Look for the 0x0249 value elsewhere (might be a section size)
	fmt.Println("\n=== Looking for 0x0249 (585) value in firmware ===")
	target := []byte{0x02, 0x49}
	for i := 0; i < len(fw)-2; i++ {
		if fw[i] == target[0] && fw[i+1] == target[1] {
			fmt.Printf("Found at 0x%04x: ", i)
			// Show context
			start := i - 8
			if start < 0 {
				start = 0
			}
			end := i + 10
			if end > len(fw) {
				end = len(fw)
			}
			for j := start; j < end; j++ {
				if j == i {
					fmt.Print("[")
				}
				fmt.Printf("%02x", fw[j])
				if j == i+1 {
					fmt.Print("]")
				}
				fmt.Print(" ")
			}
			fmt.Println()
		}
	}

	// Check if this section might be part of a larger structure
	// The value 0x0109d8 at 0x4f8 might be significant
	fmt.Println("\n=== Checking structure alignment ===")

	// Let's see if our section starts at a specific alignment
	fmt.Printf("Section start 0x4d0 = %d (decimal)\n", 0x4d0)
	fmt.Printf("Section start modulo 16: %d\n", 0x4d0%16)
	fmt.Printf("Section start modulo 32: %d\n", 0x4d0%32)
	fmt.Printf("Section start modulo 64: %d\n", 0x4d0%64)

	// Check if the CRC might cover data from multiple locations
	fmt.Println("\n=== Analyzing potential ARM boot argument structure ===")

	// Based on the hint that this is for ARM core command line arguments,
	// let's look for typical boot argument patterns

	// Common patterns in boot args:
	// - Board/device identifier (which we have: MBF2M345A-HECO)
	// - Memory addresses
	// - Boot flags
	// - Serial/version numbers

	section := fw[0x4d0:0x500]

	fmt.Println("\nParsed structure:")
	fmt.Printf("0x4d0-0x4d5: Reserved/flags: %02x %02x %02x %02x %02x %02x\n",
		section[0], section[1], section[2], section[3], section[4], section[5])
	fmt.Printf("0x4d6-0x4d7: Length: 0x%04x (%d)\n",
		binary.BigEndian.Uint16(section[6:8]), binary.BigEndian.Uint16(section[6:8]))
	fmt.Printf("0x4d8-0x4e5: Board ID: %s\n", string(section[8:22]))
	fmt.Printf("0x4e6-0x4f7: Padding (zeros)\n")
	fmt.Printf("0x4f8-0x4fb: Unknown value: 0x%08x (%d)\n",
		binary.BigEndian.Uint32(section[0x28:0x2c]), binary.BigEndian.Uint32(section[0x28:0x2c]))
	fmt.Printf("0x4fc-0x4fd: Reserved: %02x %02x\n", section[0x2c], section[0x2d])
	fmt.Printf("0x4fe-0x4ff: CRC16: 0x%04x\n", binary.BigEndian.Uint16(section[0x2e:0x30]))
}
