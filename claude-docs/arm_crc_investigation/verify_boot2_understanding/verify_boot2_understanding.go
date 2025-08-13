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
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run verify_boot2_understanding.go <firmware.bin>")
		os.Exit(1)
	}

	file, _ := os.Open(os.Args[1])
	defer file.Close()

	// Read size at 0x1004
	file.Seek(0x1004, 0)
	var sizeBytes [4]byte
	file.Read(sizeBytes[:])
	size := binary.BigEndian.Uint32(sizeBytes[:])

	fmt.Printf("Size from header: 0x%x dwords\n", size)
	fmt.Printf("Total bytes mstflint reads: %d * 4 + 16 = 0x%x\n", size, size*4+16)
	fmt.Printf("CRC at dword index: %d + 3 = %d (0x%x)\n", size, size+3, size+3)
	fmt.Printf("CRC byte offset: 0x%x * 4 = 0x%x\n", size+3, (size+3)*4)

	totalBytes := size*4 + 16
	crcByteOffset := (size + 3) * 4
	fmt.Printf("\nCRC position relative to end: 0x%x - 0x%x = %d bytes from end\n",
		totalBytes, crcByteOffset, totalBytes-crcByteOffset)

	// Read the last 4 bytes of BOOT2
	boot2Start := int64(0x1000)
	file.Seek(boot2Start+int64(totalBytes)-4, 0)
	var lastDword [4]byte
	file.Read(lastDword[:])

	fmt.Printf("\nLast 4 bytes of BOOT2: %02x %02x %02x %02x\n",
		lastDword[0], lastDword[1], lastDword[2], lastDword[3])
	fmt.Printf("As big-endian u32: 0x%08x\n", binary.BigEndian.Uint32(lastDword[:]))

	// Also check what's at the exact CRC position
	file.Seek(boot2Start+int64(crcByteOffset), 0)
	var crcDword [4]byte
	file.Read(crcDword[:])
	fmt.Printf("\nAt CRC position 0x%x: %02x %02x %02x %02x\n",
		boot2Start+int64(crcByteOffset),
		crcDword[0], crcDword[1], crcDword[2], crcDword[3])
	fmt.Printf("As big-endian u32: 0x%08x\n", binary.BigEndian.Uint32(crcDword[:]))
}
