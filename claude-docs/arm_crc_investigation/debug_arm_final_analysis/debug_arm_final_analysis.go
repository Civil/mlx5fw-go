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
	origFW, _ := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
	modFW, _ := os.ReadFile("sample_firmwares/franken_fw.bin")

	fmt.Println("=== Final ARM Boot Args Analysis ===\n")

	// Display the exact structure
	fmt.Println("Structure at 0x4D0-0x500:")
	data := origFW[0x4D0:0x500]
	for i := 0; i < len(data); i += 16 {
		fmt.Printf("%04X: ", 0x4D0+i)
		for j := 0; j < 16 && i+j < len(data); j++ {
			fmt.Printf("%02X ", data[i+j])
		}
		fmt.Print(" |")
		for j := 0; j < 16 && i+j < len(data); j++ {
			if data[i+j] >= 32 && data[i+j] < 127 {
				fmt.Printf("%c", data[i+j])
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println("|")
	}

	fmt.Println("\nStructure breakdown:")
	fmt.Printf("0x4D0-0x4D5: Padding/flags   = %X\n", origFW[0x4D0:0x4D6])
	fmt.Printf("0x4D6-0x4D7: Length field    = 0x%04X (%d)\n",
		binary.BigEndian.Uint16(origFW[0x4D6:0x4D8]),
		binary.BigEndian.Uint16(origFW[0x4D6:0x4D8]))
	fmt.Printf("0x4D8-0x4F7: Board ID field  = '%s' (32 bytes, null-padded)\n",
		string(origFW[0x4D8:0x4E6]))
	fmt.Printf("0x4F8-0x4FB: Unknown value   = 0x%08X\n",
		binary.BigEndian.Uint32(origFW[0x4F8:0x4FC]))
	fmt.Printf("0x4FC-0x4FD: Padding         = %X\n", origFW[0x4FC:0x4FE])
	fmt.Printf("0x4FE-0x4FF: CRC16           = 0x%04X\n",
		binary.BigEndian.Uint16(origFW[0x4FE:0x500]))

	fmt.Println("\nModified firmware differences:")
	fmt.Printf("Board ID: 'MBF2M345A-HECO' -> 'MBF2M345A-VENOT_ES'\n")
	fmt.Printf("CRC16: 0x%04X -> 0x%04X\n",
		binary.BigEndian.Uint16(origFW[0x4FE:0x500]),
		binary.BigEndian.Uint16(modFW[0x4FE:0x500]))

	// Summary of CRC findings
	fmt.Println("\n=== CRC Algorithm Analysis Summary ===")
	fmt.Println("Tested algorithms:")
	fmt.Println("1. Standard CRC16 variants (23+ algorithms)")
	fmt.Println("2. mstflint Software CRC16 (poly 0x100b, 32-bit word processing)")
	fmt.Println("3. mstflint Hardware CRC (with inverted first bytes)")
	fmt.Println("4. Byte-by-byte CRC16 with poly 0x100b")
	fmt.Println("5. Various data ranges from 0x4C0 to 0x4FE")

	fmt.Println("\nConclusion:")
	fmt.Println("- The CRC algorithm protecting this ARM boot args section is NOT")
	fmt.Println("  any of the standard CRC16 algorithms or mstflint's implementations")
	fmt.Println("- The CRC does change correctly when the board ID is modified")
	fmt.Println("- This suggests a proprietary CRC algorithm or non-standard parameters")
	fmt.Println("- The section is likely verified by ARM boot code using its own CRC method")

	// Check if this might be related to other sections
	fmt.Println("\n=== Context Analysis ===")
	fmt.Printf("Section location: Within larger region 0x28C-0x540 (bounded by 0xFFFF padding)\n")
	fmt.Printf("Not part of ITOC/DTOC sections (those start at 0x%X)\n", 0x1000)
	fmt.Printf("Appears to be a hardcoded boot configuration area\n")
}
