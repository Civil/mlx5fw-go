package main

import (
    "encoding/binary"
    "fmt"
    "os"
)

// CRC16 implementation with table-based approach
func makeCRCTable(poly uint16) [256]uint16 {
    var table [256]uint16
    for i := 0; i < 256; i++ {
        crc := uint16(i) << 8
        for j := 0; j < 8; j++ {
            if crc&0x8000 != 0 {
                crc = (crc << 1) ^ poly
            } else {
                crc <<= 1
            }
        }
        table[i] = crc
    }
    return table
}

func calcCRC16Table(data []byte, poly, init uint16) uint16 {
    table := makeCRCTable(poly)
    crc := init
    
    for _, b := range data {
        crc = (crc << 8) ^ table[((crc>>8)^uint16(b))&0xFF]
    }
    
    return crc
}

func reverseBits16(n uint16) uint16 {
    var result uint16
    for i := 0; i < 16; i++ {
        result = (result << 1) | (n & 1)
        n >>= 1
    }
    return result
}

func main() {
    // Read firmware files
    origFW, err := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
    if err != nil {
        panic(err)
    }
    
    modFW, err := os.ReadFile("sample_firmwares/franken_fw.bin")
    if err != nil {
        panic(err)
    }
    
    // Extract sections
    origSection := origFW[0x4d0:0x500]
    modSection := modFW[0x4d0:0x500]
    
    // Target CRCs
    origCRC := uint16(0x6885)
    modCRC := uint16(0x5220)
    
    // Common polynomials
    polynomials := []struct {
        name string
        poly uint16
        init uint16
    }{
        {"CCITT", 0x1021, 0xFFFF},
        {"CCITT-0", 0x1021, 0x0000},
        {"CRC16", 0x8005, 0x0000},
        {"CRC16-IBM", 0x8005, 0xFFFF},
        {"DNP", 0x3D65, 0x0000},
        {"DNP-FFFF", 0x3D65, 0xFFFF},
    }
    
    // Test different data ranges
    ranges := []struct {
        name   string
        offset int
        length int
    }{
        {"String only (14 bytes)", 8, 14},
        {"String only (18 bytes)", 8, 18},
        {"String + padding (32 bytes)", 8, 32},
        {"From 0x4d4 (42 bytes)", 4, 42},
        {"From 0x4d6 (40 bytes)", 6, 40},
        {"Full section (46 bytes)", 0, 46},
        {"Up to 0x4f8 (40 bytes from start)", 0, 40},
        {"0x4d8 to 0x4fc (36 bytes)", 8, 36},
    }
    
    fmt.Println("=== Testing CRC16 algorithms with different data ranges ===\n")
    
    foundMatch := false
    
    for _, r := range ranges {
        if r.offset+r.length > len(origSection) {
            continue
        }
        
        origData := origSection[r.offset : r.offset+r.length]
        modData := modSection[r.offset : r.offset+r.length]
        
        for _, p := range polynomials {
            // Test normal
            origCalc := calcCRC16Table(origData, p.poly, p.init)
            modCalc := calcCRC16Table(modData, p.poly, p.init)
            
            if origCalc == origCRC && modCalc == modCRC {
                fmt.Printf("*** MATCH FOUND! ***\n")
                fmt.Printf("Range: %s\n", r.name)
                fmt.Printf("Algorithm: %s (poly=0x%04X, init=0x%04X)\n", p.name, p.poly, p.init)
                fmt.Printf("Original CRC: 0x%04X (calculated: 0x%04X)\n", origCRC, origCalc)
                fmt.Printf("Modified CRC: 0x%04X (calculated: 0x%04X)\n", modCRC, modCalc)
                fmt.Printf("Data offset: 0x%X, length: %d\n\n", 0x4d0+r.offset, r.length)
                foundMatch = true
            }
            
            // Test with XOR out 0xFFFF
            origCalcXOR := origCalc ^ 0xFFFF
            modCalcXOR := modCalc ^ 0xFFFF
            
            if origCalcXOR == origCRC && modCalcXOR == modCRC {
                fmt.Printf("*** MATCH FOUND (with XOR 0xFFFF)! ***\n")
                fmt.Printf("Range: %s\n", r.name)
                fmt.Printf("Algorithm: %s (poly=0x%04X, init=0x%04X, xor=0xFFFF)\n", p.name, p.poly, p.init)
                fmt.Printf("Original CRC: 0x%04X (calculated: 0x%04X)\n", origCRC, origCalcXOR)
                fmt.Printf("Modified CRC: 0x%04X (calculated: 0x%04X)\n", modCRC, modCalcXOR)
                fmt.Printf("Data offset: 0x%X, length: %d\n\n", 0x4d0+r.offset, r.length)
                foundMatch = true
            }
            
            // Test reversed (for reflected algorithms)
            origCalcRev := reverseBits16(origCalc)
            modCalcRev := reverseBits16(modCalc)
            
            if origCalcRev == origCRC && modCalcRev == modCRC {
                fmt.Printf("*** MATCH FOUND (reversed)! ***\n")
                fmt.Printf("Range: %s\n", r.name)
                fmt.Printf("Algorithm: %s (poly=0x%04X, init=0x%04X, reflected)\n", p.name, p.poly, p.init)
                fmt.Printf("Original CRC: 0x%04X (calculated: 0x%04X)\n", origCRC, origCalcRev)
                fmt.Printf("Modified CRC: 0x%04X (calculated: 0x%04X)\n", modCRC, modCalcRev)
                fmt.Printf("Data offset: 0x%X, length: %d\n\n", 0x4d0+r.offset, r.length)
                foundMatch = true
            }
        }
    }
    
    if !foundMatch {
        fmt.Println("No exact CRC match found. Testing partial matches...")
        
        // Test if only one of them matches
        for _, r := range ranges {
            if r.offset+r.length > len(origSection) {
                continue
            }
            
            origData := origSection[r.offset : r.offset+r.length]
            modData := modSection[r.offset : r.offset+r.length]
            
            for _, p := range polynomials {
                origCalc := calcCRC16Table(origData, p.poly, p.init)
                modCalc := calcCRC16Table(modData, p.poly, p.init)
                
                if origCalc == origCRC {
                    fmt.Printf("Original matches: %s with %s (0x%04X)\n", r.name, p.name, origCalc)
                }
                if modCalc == modCRC {
                    fmt.Printf("Modified matches: %s with %s (0x%04X)\n", r.name, p.name, modCalc)
                }
            }
        }
    }
    
    // Also check if the CRC is stored in little endian
    fmt.Println("\n=== Checking little-endian CRC storage ===")
    origCRCLE := binary.LittleEndian.Uint16(origSection[0x2e:0x30])
    modCRCLE := binary.LittleEndian.Uint16(modSection[0x2e:0x30])
    fmt.Printf("Original CRC (LE): 0x%04X\n", origCRCLE)
    fmt.Printf("Modified CRC (LE): 0x%04X\n", modCRCLE)
}