package main

import (
    "encoding/binary"
    "fmt"
    "os"
)

func main() {
    // Read both firmware files
    origData, err := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
    if err != nil {
        panic(err)
    }
    
    modData, err := os.ReadFile("sample_firmwares/franken_fw.bin")
    if err != nil {
        panic(err)
    }
    
    // Extract the section from 0x4d0 to 0x500
    origSection := origData[0x4d0:0x500]
    modSection := modData[0x4d0:0x500]
    
    fmt.Println("=== Analyzing section structure ===")
    fmt.Printf("Section offset: 0x4d0 - 0x500 (48 bytes)\n\n")
    
    // Parse the structure
    // Appears to be:
    // 0x4d0: 00 00 00 00 00 00 02 49  - possibly flags/reserved + length (0x249 = 585)
    // 0x4d8: String data starts "MBF2M345A-HECO" / "MBF2M345A-VENOT_ES"
    // 0x4f8: 00 01 09 d8 - possibly version or type (0x0109d8 = 68056)
    // 0x4fc: 00 00 XX XX - CRC16
    
    fmt.Println("Original section:")
    for i := 0; i < len(origSection); i += 16 {
        fmt.Printf("%04x: ", 0x4d0+i)
        for j := 0; j < 16 && i+j < len(origSection); j++ {
            fmt.Printf("%02x ", origSection[i+j])
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
    
    fmt.Println("\nModified section:")
    for i := 0; i < len(modSection); i += 16 {
        fmt.Printf("%04x: ", 0x4d0+i)
        for j := 0; j < 16 && i+j < len(modSection); j++ {
            fmt.Printf("%02x ", modSection[i+j])
        }
        fmt.Print(" |")
        for j := 0; j < 16 && i+j < len(modSection); j++ {
            b := modSection[i+j]
            if b >= 32 && b < 127 {
                fmt.Printf("%c", b)
            } else {
                fmt.Print(".")
            }
        }
        fmt.Println("|")
    }
    
    // Extract components
    length := binary.BigEndian.Uint16(origSection[6:8])
    fmt.Printf("\nLength field: 0x%04x (%d)\n", length, length)
    
    origStr := string(origSection[8:22])  // "MBF2M345A-HECO"
    modStr := string(modSection[8:26])    // "MBF2M345A-VENOT_ES"
    
    fmt.Printf("Original string: '%s' (len=%d)\n", origStr, len(origStr))
    fmt.Printf("Modified string: '%s' (len=%d)\n", modStr, len(modStr))
    
    // The value at 0x4f8
    val1 := binary.BigEndian.Uint32(origSection[0x28:0x2c])
    fmt.Printf("\nValue at 0x4f8: 0x%08x (%d)\n", val1, val1)
    
    // CRC values
    origCRC := binary.BigEndian.Uint16(origSection[0x2e:0x30])
    modCRC := binary.BigEndian.Uint16(modSection[0x2e:0x30])
    
    fmt.Printf("\nOriginal CRC: 0x%04x\n", origCRC)
    fmt.Printf("Modified CRC: 0x%04x\n", modCRC)
    
    // Let's look at what data might be covered by CRC
    // Option 1: Just the string
    // Option 2: String + padding
    // Option 3: Entire section minus CRC
    // Option 4: Some specific range
    
    fmt.Println("\n=== Testing CRC coverage ===")
    
    // Test different ranges
    testRanges := []struct {
        name  string
        start int
        end   int
    }{
        {"Just string (orig)", 8, 22},
        {"Just string (mod)", 8, 26},
        {"String with padding to 0x4f8", 8, 0x28},
        {"From length field", 6, 0x2e},
        {"From start to CRC", 0, 0x2e},
        {"Just before string to CRC", 4, 0x2e},
    }
    
    for _, r := range testRanges {
        fmt.Printf("\nTesting range %s (0x%x-0x%x):\n", r.name, 0x4d0+r.start, 0x4d0+r.end)
        if r.end <= len(origSection) {
            origData := origSection[r.start:r.end]
            fmt.Printf("  Original data length: %d bytes\n", len(origData))
        }
        if r.end <= len(modSection) {
            modData := modSection[r.start:r.end]
            fmt.Printf("  Modified data length: %d bytes\n", len(modData))
        }
    }
}