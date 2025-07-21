package main

import (
    "encoding/binary"
    "fmt"
    "os"
)

// CRC16 implementation matching mstflint's Crc16::add
func calculateMstflintCRC16(data []byte) uint16 {
    const poly = uint32(0x100b)
    crc := uint32(0xffff)
    
    // Pad data to 4-byte alignment
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
        
        // Process each bit
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
    
    // Final XOR and return lower 16 bits
    return uint16((crc >> 16) ^ 0xffff)
}

// Hardware CRC using lookup table approach
func calculateHardwareCRC(data []byte) uint16 {
    // This would need the CRC16Table2 from the codebase
    // For now, skip this implementation
    return 0
}

func main() {
    origFW, _ := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
    modFW, _ := os.ReadFile("sample_firmwares/franken_fw.bin")
    
    // CRCs from firmware
    origCRC := uint16(0x6885)
    modCRC := uint16(0x5220)
    
    fmt.Println("=== Testing mstflint CRC algorithms ===\n")
    
    // Test different data ranges
    testRanges := []struct {
        name   string
        start  int
        end    int
    }{
        {"Board ID only", 0x4D8, 0x4E6},                    // Just "MBF2M345A-HECO"
        {"Board ID extended", 0x4D8, 0x4EA},                // "MBF2M345A-VENOT_ES"
        {"With length prefix", 0x4D6, 0x4FE},               // 0x0249 + string + padding + value
        {"From 0x4D0", 0x4D0, 0x4FE},                      // Full structure
        {"From 0x4C0", 0x4C0, 0x4FE},                      // Extended range
        {"32-byte aligned before string", 0x4C0, 0x4E0},    // 32-byte chunk
        {"64-byte aligned", 0x4C0, 0x500},                  // 64-byte chunk
        {"String to value", 0x4D8, 0x4FC},                  // String + padding + 0x0109D8
        {"Fixed 32-byte string field", 0x4D8, 0x4F8},       // Exactly 32 bytes
        {"With value after string", 0x4D8, 0x4FC},          // 32 bytes + 4 byte value
        {"From after previous data", 0x4D0, 0x4FE},         // Standard range
    }
    
    for _, r := range testRanges {
        if r.start >= r.end || r.end > len(origFW) {
            continue
        }
        
        origData := origFW[r.start:r.end]
        modData := modFW[r.start:r.end]
        
        origCalc := calculateMstflintCRC16(origData)
        modCalc := calculateMstflintCRC16(modData)
        
        if origCalc == origCRC && modCalc == modCRC {
            fmt.Printf("*** EXACT MATCH FOUND! ***\n")
            fmt.Printf("Range: %s\n", r.name)
            fmt.Printf("Start: 0x%04X, End: 0x%04X\n", r.start, r.end)
            fmt.Printf("Length: %d bytes\n", len(origData))
            fmt.Printf("Original CRC: calc=0x%04X, stored=0x%04X ✓\n", origCalc, origCRC)
            fmt.Printf("Modified CRC: calc=0x%04X, stored=0x%04X ✓\n", modCalc, modCRC)
            
            fmt.Println("\nData that's being CRC'd:")
            for i := 0; i < len(origData) && i < 64; i += 16 {
                fmt.Printf("%04X: ", r.start+i)
                for j := 0; j < 16 && i+j < len(origData); j++ {
                    fmt.Printf("%02X ", origData[i+j])
                }
                fmt.Print(" |")
                for j := 0; j < 16 && i+j < len(origData); j++ {
                    if origData[i+j] >= 32 && origData[i+j] < 127 {
                        fmt.Printf("%c", origData[i+j])
                    } else {
                        fmt.Print(".")
                    }
                }
                fmt.Println("|")
            }
            return
        }
        
        // Show close matches for debugging
        if origCalc == origCRC || modCalc == modCRC {
            fmt.Printf("Partial match - %s: ", r.name)
            if origCalc == origCRC {
                fmt.Printf("orig matches (0x%04X) ", origCalc)
            }
            if modCalc == modCRC {
                fmt.Printf("mod matches (0x%04X) ", modCalc)
            }
            fmt.Println()
        }
    }
    
    fmt.Println("\nNo exact match found. Testing with inverted bytes...")
    
    // Try with first bytes inverted (like hardware CRC)
    for _, r := range testRanges {
        if r.start >= r.end || r.end > len(origFW) || r.end-r.start < 2 {
            continue
        }
        
        origData := make([]byte, r.end-r.start)
        modData := make([]byte, r.end-r.start)
        copy(origData, origFW[r.start:r.end])
        copy(modData, modFW[r.start:r.end])
        
        // Invert first 2 bytes
        origData[0] = ^origData[0]
        origData[1] = ^origData[1]
        modData[0] = ^modData[0]
        modData[1] = ^modData[1]
        
        origCalc := calculateMstflintCRC16(origData)
        modCalc := calculateMstflintCRC16(modData)
        
        if origCalc == origCRC && modCalc == modCRC {
            fmt.Printf("*** MATCH with inverted first bytes! Range: %s ***\n", r.name)
            return
        }
    }
}