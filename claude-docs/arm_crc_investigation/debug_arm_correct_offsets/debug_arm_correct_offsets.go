//go:build ignore
// +build ignore

// Archived experimental code; excluded from build and vet
package main

import (
	"fmt"
	"os"
)

func main() {
	// Read both firmware files
	origFW, err := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
	if err != nil {
		panic(err)
	}

	modFW, err := os.ReadFile("sample_firmwares/franken_fw.bin")
	if err != nil {
		panic(err)
	}

	fmt.Println("=== Analyzing ARM Boot Args Section with Correct Boundaries ===\n")

	// The section appears to be from 0x28a to 0x53f (before the next 0xFFFF padding)
	// Let's examine the structure

	fmt.Println("Looking for section boundaries...")

	// Find start after 0xFFFF padding
	startOffset := 0x276
	for i := startOffset; i < len(origFW)-2; i++ {
		if origFW[i] == 0xFF && origFW[i+1] == 0xFF {
			continue
		}
		if origFW[i] == 0x00 && origFW[i+1] == 0x00 && i > startOffset+10 {
			startOffset = i
			break
		}
	}

	// Find end before next 0xFFFF padding
	endOffset := 0x540
	for i := 0x500; i < 0x560 && i < len(origFW)-2; i++ {
		if origFW[i] == 0xFF && origFW[i+1] == 0xFF {
			endOffset = i
			break
		}
	}

	fmt.Printf("Section boundaries: 0x%X to 0x%X\n", startOffset, endOffset)
	fmt.Printf("Section size: %d bytes\n\n", endOffset-startOffset)

	// Extract sections
	origSection := origFW[startOffset:endOffset]
	modSection := modFW[startOffset:endOffset]

	// Display the section
	fmt.Println("Original section:")
	for i := 0; i < len(origSection) && i < 512; i += 16 {
		fmt.Printf("%04X: ", startOffset+i)
		for j := 0; j < 16 && i+j < len(origSection); j++ {
			fmt.Printf("%02X ", origSection[i+j])
		}
		fmt.Print(" |")
		for j := 0; j < 16 && i+j < len(origSection); j++ {
			b := origSection[i+j]
			if b >= 32 && b < 127 {
				fmt.Printf("%c", b)
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println("|")
	}

	// Find where our board ID string is located
	fmt.Println("\n=== Searching for board ID string ===")

	// Search for "MBF2M345" in the section
	searchStr := []byte("MBF2M345")
	for i := 0; i < len(origSection)-len(searchStr); i++ {
		match := true
		for j := 0; j < len(searchStr); j++ {
			if origSection[i+j] != searchStr[j] {
				match = false
				break
			}
		}
		if match {
			fmt.Printf("Found board ID at offset 0x%X (relative: 0x%X)\n", startOffset+i, i)

			// Show context
			contextStart := i - 16
			if contextStart < 0 {
				contextStart = 0
			}
			contextEnd := i + 64
			if contextEnd > len(origSection) {
				contextEnd = len(origSection)
			}

			fmt.Println("\nContext around board ID:")
			for j := contextStart; j < contextEnd; j += 16 {
				fmt.Printf("%04X: ", startOffset+j)
				for k := 0; k < 16 && j+k < contextEnd; k++ {
					fmt.Printf("%02X ", origSection[j+k])
				}
				fmt.Println()
			}
		}
	}

	// Now let's find the differences
	fmt.Println("\n=== Differences between original and modified ===")
	diffCount := 0
	for i := 0; i < len(origSection) && i < len(modSection); i++ {
		if origSection[i] != modSection[i] {
			if diffCount < 20 { // Show first 20 differences
				fmt.Printf("0x%04X: 0x%02X -> 0x%02X", startOffset+i, origSection[i], modSection[i])
				if origSection[i] >= 32 && origSection[i] < 127 && modSection[i] >= 32 && modSection[i] < 127 {
					fmt.Printf(" ('%c' -> '%c')", origSection[i], modSection[i])
				}
				fmt.Println()
			}
			diffCount++
		}
	}
	fmt.Printf("\nTotal differences: %d bytes\n", diffCount)
}
