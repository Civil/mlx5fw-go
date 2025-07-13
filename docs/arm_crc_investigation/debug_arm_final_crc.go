package main

import (
    "fmt"
    "os"
)

// CRC16 with different byte processing orders
func crc16MSB(data []byte, poly, init uint16) uint16 {
    crc := init
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

func crc16LSB(data []byte, poly, init uint16) uint16 {
    crc := init
    for _, b := range data {
        crc ^= uint16(b)
        for i := 0; i < 8; i++ {
            if crc&0x0001 != 0 {
                crc = (crc >> 1) ^ poly
            } else {
                crc >>= 1
            }
        }
    }
    return crc
}

func swapBytes(v uint16) uint16 {
    return (v << 8) | (v >> 8)
}

func main() {
    origFW, _ := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
    modFW, _ := os.ReadFile("sample_firmwares/franken_fw.bin")
    
    // Extract the boot args section
    origData := origFW[0x4d0:0x4fe] // 46 bytes, excluding CRC
    modData := modFW[0x4d0:0x4fe]
    
    origCRC := uint16(0x6885)
    modCRC := uint16(0x5220)
    
    fmt.Println("=== Final CRC Analysis ===\n")
    
    // Let's also check if this matches existing CRC algorithms from the codebase
    // Based on the code analysis, there are multiple CRC16 implementations
    
    // From pkg/fs4/crc16.go - two different algorithms are used
    polys := []struct {
        name string
        poly uint16
        init uint16
        swap bool
        lsb  bool
    }{
        {"CCITT", 0x1021, 0xFFFF, false, false},
        {"CCITT-0", 0x1021, 0x0000, false, false},
        {"XMODEM", 0x1021, 0x0000, false, false},
        {"CRC16", 0x8005, 0x0000, false, false},
        {"CRC16-IBM", 0x8005, 0xFFFF, false, false},
        {"CRC16-USB", 0x8005, 0xFFFF, false, true},
        {"CRC16-IBM-REV", 0xA001, 0xFFFF, false, true}, // Reversed poly
        {"CCITT-REV", 0x8408, 0xFFFF, false, true},      // Reversed poly
    }
    
    // Test each algorithm
    for _, p := range polys {
        var origCalc, modCalc uint16
        
        if p.lsb {
            origCalc = crc16LSB(origData, p.poly, p.init)
            modCalc = crc16LSB(modData, p.poly, p.init)
        } else {
            origCalc = crc16MSB(origData, p.poly, p.init)
            modCalc = crc16MSB(modData, p.poly, p.init)
        }
        
        // Check direct match
        if origCalc == origCRC && modCalc == modCRC {
            fmt.Printf("*** MATCH! %s: orig=0x%04X, mod=0x%04X ***\n", p.name, origCalc, modCalc)
        }
        
        // Check with byte swap
        origSwap := swapBytes(origCalc)
        modSwap := swapBytes(modCalc)
        if origSwap == origCRC && modSwap == modCRC {
            fmt.Printf("*** MATCH (swapped)! %s: orig=0x%04X, mod=0x%04X ***\n", p.name, origSwap, modSwap)
        }
        
        // Check with XOR out
        origXOR := origCalc ^ 0xFFFF
        modXOR := modCalc ^ 0xFFFF
        if origXOR == origCRC && modXOR == modCRC {
            fmt.Printf("*** MATCH (XOR 0xFFFF)! %s: orig=0x%04X, mod=0x%04X ***\n", p.name, origXOR, modXOR)
        }
    }
    
    // Let's also check if the data might be processed differently
    fmt.Println("\n=== Checking alternative data ranges ===")
    
    // Maybe it's calculated from the string start
    stringData1 := origFW[0x4d8:0x4fe] // From string to CRC
    stringData2 := modFW[0x4d8:0x4fe]
    
    origCalc := crc16MSB(stringData1, 0x1021, 0xFFFF)
    modCalc := crc16MSB(stringData2, 0x1021, 0xFFFF)
    
    fmt.Printf("String to CRC (CCITT): orig=0x%04X (want 0x%04X), mod=0x%04X (want 0x%04X)\n",
        origCalc, origCRC, modCalc, modCRC)
    
    // Print hex differences to spot patterns
    fmt.Println("\n=== Hex dump of differences ===")
    fmt.Println("Offset | Original | Modified | Diff")
    fmt.Println("-------|----------|----------|-----")
    
    for i := 0; i < len(origData) && i < len(modData); i++ {
        if origData[i] != modData[i] {
            fmt.Printf("0x%04X | 0x%02X      | 0x%02X      | %c -> %c\n", 
                0x4d0+i, origData[i], modData[i],
                map[bool]byte{true: origData[i], false: '.'}[origData[i] >= 32 && origData[i] < 127],
                map[bool]byte{true: modData[i], false: '.'}[modData[i] >= 32 && modData[i] < 127])
        }
    }
}