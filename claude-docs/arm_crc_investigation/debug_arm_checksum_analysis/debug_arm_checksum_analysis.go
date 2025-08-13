//go:build ignore
// +build ignore

// Archived experimental code; excluded from build and vet
package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"hash/adler32"
	"hash/crc32"
	"hash/fnv"
	"os"
)

func main() {
	origFW, _ := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
	modFW, _ := os.ReadFile("sample_firmwares/franken_fw.bin")

	fmt.Println("=== Comprehensive Checksum Analysis ===\n")

	// First, let's identify ALL differences between the files
	fmt.Println("All differences in the range 0x400-0x600:")
	var diffs []int
	for i := 0x400; i < 0x600 && i < len(origFW) && i < len(modFW); i++ {
		if origFW[i] != modFW[i] {
			diffs = append(diffs, i)
			fmt.Printf("0x%04X: 0x%02X -> 0x%02X", i, origFW[i], modFW[i])
			if origFW[i] >= 32 && origFW[i] < 127 && modFW[i] >= 32 && modFW[i] < 127 {
				fmt.Printf(" ('%c' -> '%c')", origFW[i], modFW[i])
			}
			fmt.Println()
		}
	}

	fmt.Printf("\nTotal differences: %d\n", len(diffs))

	// The changing bytes at 0x4FE-0x4FF could be:
	// 1. 16-bit checksum (CRC16, sum16, etc.)
	// 2. Part of a 32-bit checksum
	// 3. Part of a larger hash (truncated MD5, etc.)

	fmt.Println("\n=== Testing various checksum algorithms ===")

	// Test ranges
	ranges := []struct {
		name  string
		start int
		end   int
	}{
		{"Board ID only", 0x4D8, 0x4E6},
		{"Board ID field", 0x4D8, 0x4F8},
		{"With length", 0x4D6, 0x4FC},
		{"Full struct", 0x4D0, 0x4FC},
		{"From 0x300", 0x300, 0x4FC},
		{"From 0x28C", 0x28C, 0x4FC},
	}

	for _, r := range ranges {
		if r.start >= r.end || r.start < 0 || r.end > len(origFW) {
			continue
		}

		origData := origFW[r.start:r.end]
		modData := modFW[r.start:r.end]

		fmt.Printf("\nRange: %s (0x%04X-0x%04X, %d bytes)\n", r.name, r.start, r.end, len(origData))

		// Simple checksums
		origSum16 := uint16(0)
		modSum16 := uint16(0)
		origSum32 := uint32(0)
		modSum32 := uint32(0)
		origXor16 := uint16(0)
		modXor16 := uint16(0)

		for i := 0; i < len(origData); i++ {
			origSum16 += uint16(origData[i])
			origSum32 += uint32(origData[i])
			origXor16 ^= uint16(origData[i])

			if i < len(modData) {
				modSum16 += uint16(modData[i])
				modSum32 += uint32(modData[i])
				modXor16 ^= uint16(modData[i])
			}
		}

		// Also try 16-bit word sums
		origSum16Word := uint16(0)
		modSum16Word := uint16(0)
		for i := 0; i < len(origData)-1; i += 2 {
			origSum16Word += binary.BigEndian.Uint16(origData[i : i+2])
			if i < len(modData)-1 {
				modSum16Word += binary.BigEndian.Uint16(modData[i : i+2])
			}
		}

		fmt.Printf("  Sum16: orig=0x%04X, mod=0x%04X", origSum16, modSum16)
		if origSum16 == 0x6885 && modSum16 == 0x5220 {
			fmt.Printf(" *** MATCH! ***")
		}
		fmt.Println()

		fmt.Printf("  Sum16Word: orig=0x%04X, mod=0x%04X", origSum16Word, modSum16Word)
		if origSum16Word == 0x6885 && modSum16Word == 0x5220 {
			fmt.Printf(" *** MATCH! ***")
		}
		fmt.Println()

		fmt.Printf("  XOR16: orig=0x%04X, mod=0x%04X", origXor16, modXor16)
		if origXor16 == 0x6885 && modXor16 == 0x5220 {
			fmt.Printf(" *** MATCH! ***")
		}
		fmt.Println()

		// Standard Go checksums
		origCRC32 := crc32.ChecksumIEEE(origData)
		modCRC32 := crc32.ChecksumIEEE(modData)
		fmt.Printf("  CRC32-IEEE: orig=0x%08X, mod=0x%08X\n", origCRC32, modCRC32)

		origAdler := adler32.Checksum(origData)
		modAdler := adler32.Checksum(modData)
		fmt.Printf("  Adler32: orig=0x%08X, mod=0x%08X\n", origAdler, modAdler)

		// FNV hashes
		fnv32 := fnv.New32()
		fnv32.Write(origData)
		origFNV := fnv32.Sum32()
		fnv32.Reset()
		fnv32.Write(modData)
		modFNV := fnv32.Sum32()
		fmt.Printf("  FNV32: orig=0x%08X, mod=0x%08X\n", origFNV, modFNV)

		// Check if lower 16 bits match
		if uint16(origCRC32) == 0x6885 && uint16(modCRC32) == 0x5220 {
			fmt.Printf("  *** CRC32-IEEE lower 16 bits match! ***\n")
		}
		if uint16(origAdler) == 0x6885 && uint16(modAdler) == 0x5220 {
			fmt.Printf("  *** Adler32 lower 16 bits match! ***\n")
		}
		if uint16(origFNV) == 0x6885 && uint16(modFNV) == 0x5220 {
			fmt.Printf("  *** FNV32 lower 16 bits match! ***\n")
		}

		// MD5 (first 2 bytes)
		md5orig := md5.Sum(origData)
		md5mod := md5.Sum(modData)
		fmt.Printf("  MD5 first 2 bytes: orig=0x%02X%02X, mod=0x%02X%02X\n",
			md5orig[0], md5orig[1], md5mod[0], md5mod[1])
	}

	// Check if it might be a fletcher checksum
	fmt.Println("\n=== Testing Fletcher checksums ===")

	for _, r := range ranges {
		if r.start >= r.end || r.start < 0 || r.end > len(origFW) {
			continue
		}

		origData := origFW[r.start:r.end]
		modData := modFW[r.start:r.end]

		// Fletcher-16
		origF16 := fletcher16(origData)
		modF16 := fletcher16(modData)

		fmt.Printf("%s - Fletcher16: orig=0x%04X, mod=0x%04X", r.name, origF16, modF16)
		if origF16 == 0x6885 && modF16 == 0x5220 {
			fmt.Printf(" *** MATCH! ***")
		}
		fmt.Println()
	}

	// Maybe the "checksum" is actually just a counter or version that changes
	fmt.Println("\n=== Analyzing the changing value ===")
	fmt.Printf("Original: 0x6885 = %d (decimal)\n", 0x6885)
	fmt.Printf("Modified: 0x5220 = %d (decimal)\n", 0x5220)
	fmt.Printf("Difference: %d\n", 0x6885-0x5220)

	// Check if it could be related to string length
	origLen := 14 // "MBF2M345A-HECO"
	modLen := 18  // "MBF2M345A-VENOT_ES"
	fmt.Printf("\nString lengths: orig=%d, mod=%d\n", origLen, modLen)

	// Maybe it's a simple calculation based on the string
	origStrSum := 0
	for _, b := range []byte("MBF2M345A-HECO") {
		origStrSum += int(b)
	}
	modStrSum := 0
	for _, b := range []byte("MBF2M345A-VENOT_ES") {
		modStrSum += int(b)
	}
	fmt.Printf("String ASCII sums: orig=%d (0x%04X), mod=%d (0x%04X)\n",
		origStrSum, origStrSum, modStrSum, modStrSum)
}

func fletcher16(data []byte) uint16 {
	var sum1, sum2 uint16 = 0, 0
	for _, b := range data {
		sum1 = (sum1 + uint16(b)) % 255
		sum2 = (sum2 + sum1) % 255
	}
	return (sum2 << 8) | sum1
}
