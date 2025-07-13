package main

import (
    "encoding/binary"
    "fmt"
    "os"
)

func main() {
    // Read firmware files
    origFW, _ := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
    modFW, _ := os.ReadFile("sample_firmwares/franken_fw.bin")
    
    fmt.Println("=== Analyzing Full ARM Section Structure ===\n")
    
    // The actual section is from 0x28C to 0x540
    startOffset := 0x28C
    endOffset := 0x540
    
    origSection := origFW[startOffset:endOffset]
    modSection := modFW[startOffset:endOffset]
    
    fmt.Printf("Section range: 0x%X - 0x%X\n", startOffset, endOffset)
    fmt.Printf("Section size: %d bytes (0x%X)\n\n", len(origSection), len(origSection))
    
    // Look for patterns and structures
    fmt.Println("=== Identifying sub-structures ===")
    
    // The repeating pattern 0xC688FAC6 from 0x2BE onwards
    fmt.Printf("\nRepeating pattern at 0x2BE: ")
    for i := 0x32; i < 0x72 && i < len(origSection); i += 4 {
        val := binary.BigEndian.Uint32(origSection[i:i+4])
        fmt.Printf("0x%08X ", val)
        if (i-0x32)%16 == 12 {
            fmt.Println()
        }
    }
    fmt.Println()
    
    // Look at the structure around our board ID
    boardIDOffset := 0x4D8 - startOffset  // 0x24C
    fmt.Printf("\nBoard ID at offset 0x%X (absolute: 0x%X)\n", boardIDOffset, 0x4D8)
    
    // Check what's before the board ID
    fmt.Println("\nData before board ID (0x4C8-0x4D8):")
    for i := boardIDOffset - 16; i < boardIDOffset; i++ {
        fmt.Printf("%02X ", origSection[i])
    }
    fmt.Println()
    
    // The CRC is at 0x4FE-0x4FF
    crcOffset := 0x4FE - startOffset
    origCRC := binary.BigEndian.Uint16(origSection[crcOffset:crcOffset+2])
    modCRC := binary.BigEndian.Uint16(modSection[crcOffset:crcOffset+2])
    
    fmt.Printf("\nCRC location: offset 0x%X (absolute: 0x%X)\n", crcOffset, 0x4FE)
    fmt.Printf("Original CRC: 0x%04X\n", origCRC)
    fmt.Printf("Modified CRC: 0x%04X\n", modCRC)
    
    // Now let's test CRC on different ranges within the full section
    fmt.Println("\n=== Testing CRC calculations ===")
    
    // Test ranges
    testRanges := []struct {
        name   string
        start  int
        end    int
    }{
        {"Board ID to CRC", boardIDOffset, crcOffset},
        {"Length field to CRC", boardIDOffset - 2, crcOffset},
        {"Start of data to CRC", boardIDOffset - 8, crcOffset},
        {"After pattern to CRC", 0x74, crcOffset},
        {"Full section to CRC", 0, crcOffset},
        {"From 0x300 to CRC", 0x300 - startOffset, crcOffset},
        {"From 0x400 to CRC", 0x400 - startOffset, crcOffset},
        {"From 0x450 to CRC", 0x450 - startOffset, crcOffset},
        {"From 0x480 to CRC", 0x480 - startOffset, crcOffset},
    }
    
    // Common CRC16 polynomials
    polys := []struct {
        name string
        poly uint16
        init uint16
    }{
        {"CCITT", 0x1021, 0xFFFF},
        {"CCITT-0", 0x1021, 0x0000},
        {"CRC16", 0x8005, 0x0000},
        {"CRC16-IBM", 0x8005, 0xFFFF},
    }
    
    for _, r := range testRanges {
        if r.start < 0 || r.end > len(origSection) {
            continue
        }
        
        origData := origSection[r.start:r.end]
        modData := modSection[r.start:r.end]
        
        for _, p := range polys {
            origCalc := calcCRC16(origData, p.poly, p.init)
            modCalc := calcCRC16(modData, p.poly, p.init)
            
            if origCalc == origCRC && modCalc == modCRC {
                fmt.Printf("\n*** MATCH FOUND! ***\n")
                fmt.Printf("Range: %s (0x%X-0x%X)\n", r.name, startOffset+r.start, startOffset+r.end)
                fmt.Printf("Algorithm: %s (poly=0x%04X, init=0x%04X)\n", p.name, p.poly, p.init)
                fmt.Printf("Data length: %d bytes\n", r.end-r.start)
                
                // Show the data that's being CRC'd
                fmt.Println("\nFirst 32 bytes of CRC'd data:")
                for i := 0; i < 32 && i < len(origData); i += 16 {
                    fmt.Printf("%04X: ", startOffset+r.start+i)
                    for j := 0; j < 16 && i+j < 32 && i+j < len(origData); j++ {
                        fmt.Printf("%02X ", origData[i+j])
                    }
                    fmt.Println()
                }
                return
            }
        }
    }
    
    fmt.Println("\nNo exact match found. Let me check specific byte ranges...")
    
    // Maybe the CRC covers a specific structure starting at a key offset
    // Let's look for key markers
    for i := 0; i < len(origSection)-4; i++ {
        // Look for the 0x0249 length field
        if origSection[i] == 0x02 && origSection[i+1] == 0x49 {
            fmt.Printf("\nFound 0x0249 at offset 0x%X (absolute: 0x%X)\n", i, startOffset+i)
            
            // Test CRC from this point
            if i+2 < crcOffset {
                testData := origSection[i:crcOffset]
                testDataMod := modSection[i:crcOffset]
                
                for _, p := range polys {
                    origCalc := calcCRC16(testData, p.poly, p.init)
                    modCalc := calcCRC16(testDataMod, p.poly, p.init)
                    
                    if origCalc == origCRC && modCalc == modCRC {
                        fmt.Printf("*** CRC MATCH from 0x0249! Algorithm: %s ***\n", p.name)
                        fmt.Printf("Start offset: 0x%X, length: %d\n", startOffset+i, len(testData))
                    }
                }
            }
        }
    }
}

func calcCRC16(data []byte, poly, init uint16) uint16 {
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