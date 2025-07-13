package main

import (
    "encoding/binary"
    "fmt"
    "os"
)

// CRC16 algorithm parameters
type CRC16Params struct {
    Name   string
    Poly   uint16
    Init   uint16
    RefIn  bool
    RefOut bool
    XorOut uint16
}

// Common CRC16 variants
var crc16Variants = []CRC16Params{
    {"CRC-16/CCITT-FALSE", 0x1021, 0xFFFF, false, false, 0x0000},
    {"CRC-16/ARC", 0x8005, 0x0000, true, true, 0x0000},
    {"CRC-16/AUG-CCITT", 0x1021, 0x1D0F, false, false, 0x0000},
    {"CRC-16/BUYPASS", 0x8005, 0x0000, false, false, 0x0000},
    {"CRC-16/CDMA2000", 0xC867, 0xFFFF, false, false, 0x0000},
    {"CRC-16/DDS-110", 0x8005, 0x800D, false, false, 0x0000},
    {"CRC-16/DECT-R", 0x0589, 0x0000, false, false, 0x0001},
    {"CRC-16/DECT-X", 0x0589, 0x0000, false, false, 0x0000},
    {"CRC-16/DNP", 0x3D65, 0x0000, true, true, 0xFFFF},
    {"CRC-16/EN-13757", 0x3D65, 0x0000, false, false, 0xFFFF},
    {"CRC-16/GENIBUS", 0x1021, 0xFFFF, false, false, 0xFFFF},
    {"CRC-16/MAXIM", 0x8005, 0x0000, true, true, 0xFFFF},
    {"CRC-16/MCRF4XX", 0x1021, 0xFFFF, true, true, 0x0000},
    {"CRC-16/RIELLO", 0x1021, 0xB2AA, true, true, 0x0000},
    {"CRC-16/T10-DIF", 0x8BB7, 0x0000, false, false, 0x0000},
    {"CRC-16/TELEDISK", 0xA097, 0x0000, false, false, 0x0000},
    {"CRC-16/TMS37157", 0x1021, 0x89EC, true, true, 0x0000},
    {"CRC-16/USB", 0x8005, 0xFFFF, true, true, 0xFFFF},
    {"CRC-A", 0x1021, 0xC6C6, true, true, 0x0000},
    {"CRC-16/KERMIT", 0x1021, 0x0000, true, true, 0x0000},
    {"CRC-16/MODBUS", 0x8005, 0xFFFF, true, true, 0x0000},
    {"CRC-16/X-25", 0x1021, 0xFFFF, true, true, 0xFFFF},
    {"CRC-16/XMODEM", 0x1021, 0x0000, false, false, 0x0000},
    {"MSTFLINT", 0x100B, 0xFFFF, false, false, 0xFFFF},
    {"MSTFLINT-0", 0x100B, 0x0000, false, false, 0x0000},
}

// Swap endianness of 16-bit words in data
func swapEndian16(data []byte) []byte {
    swapped := make([]byte, len(data))
    copy(swapped, data)
    
    // Swap every 2 bytes
    for i := 0; i < len(swapped)-1; i += 2 {
        swapped[i], swapped[i+1] = swapped[i+1], swapped[i]
    }
    
    return swapped
}

// Swap endianness of 32-bit words in data
func swapEndian32(data []byte) []byte {
    swapped := make([]byte, len(data))
    copy(swapped, data)
    
    // Pad to 4-byte alignment if needed
    for len(swapped)%4 != 0 {
        swapped = append(swapped, 0)
    }
    
    // Swap every 4 bytes
    for i := 0; i < len(swapped)-3; i += 4 {
        swapped[i], swapped[i+1], swapped[i+2], swapped[i+3] = 
            swapped[i+3], swapped[i+2], swapped[i+1], swapped[i]
    }
    
    return swapped[:len(data)] // Return original length
}

// Reverse bits in a byte
func reverseBits8(n uint8) uint8 {
    var result uint8
    for i := 0; i < 8; i++ {
        result = (result << 1) | (n & 1)
        n >>= 1
    }
    return result
}

// Reverse bits in a uint16
func reverseBits16(n uint16) uint16 {
    var result uint16
    for i := 0; i < 16; i++ {
        result = (result << 1) | (n & 1)
        n >>= 1
    }
    return result
}

// Generic CRC16 calculation
func calcCRC16(data []byte, params CRC16Params) uint16 {
    crc := params.Init
    
    for _, b := range data {
        if params.RefIn {
            b = reverseBits8(b)
        }
        
        crc ^= uint16(b) << 8
        for i := 0; i < 8; i++ {
            if crc&0x8000 != 0 {
                crc = (crc << 1) ^ params.Poly
            } else {
                crc <<= 1
            }
        }
    }
    
    if params.RefOut {
        crc = reverseBits16(crc)
    }
    
    return crc ^ params.XorOut
}

// mstflint Software CRC16 (32-bit word processing)
func mstflintSoftwareCRC(data []byte) uint16 {
    const poly = uint32(0x100b)
    crc := uint32(0xffff)
    
    // Pad to 4-byte alignment
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
    
    return uint16((crc >> 16) ^ 0xffff)
}

func main() {
    origFW, _ := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
    modFW, _ := os.ReadFile("sample_firmwares/franken_fw.bin")
    
    // Target CRCs
    origCRC := uint16(0x6885)
    modCRC := uint16(0x5220)
    
    fmt.Println("=== CRC Search with Endianness Conversion ===")
    fmt.Printf("Target CRCs: Original=0x%04X, Modified=0x%04X\n\n", origCRC, modCRC)
    
    // Find boundaries
    firstFF := 0x277
    boardIDStart := 0x4D8
    boardIDEndOrig := 0x4E6
    // boardIDEndMod := 0x4EA
    nextFF := 0x540
    crcLocation := 0x4FE
    
    fmt.Printf("Testing ranges from 0x%04X to 0x%04X (start) and 0x%04X to 0x%04X (end)\n", 
        firstFF, boardIDStart, boardIDEndOrig, nextFF)
    fmt.Println("With endianness conversions: none, 16-bit swap, 32-bit swap\n")
    
    totalTests := 0
    foundMatch := false
    
    // Test key ranges with different endianness conversions
    testRanges := []struct {
        name  string
        start int
        end   int
    }{
        // Most likely ranges first
        {"After padding", 0x28C, crcLocation},
        {"After pattern", 0x300, crcLocation},
        {"From 0x400", 0x400, crcLocation},
        {"From 0x450", 0x450, crcLocation},
        {"From 0x480", 0x480, crcLocation},
        {"Board ID area", 0x4D0, crcLocation},
        {"With length", 0x4D6, crcLocation},
        {"Just board ID", 0x4D8, crcLocation},
    }
    
    endianness := []struct {
        name string
        fn   func([]byte) []byte
    }{
        {"no conversion", func(d []byte) []byte { return d }},
        {"16-bit swap", swapEndian16},
        {"32-bit swap", swapEndian32},
    }
    
    for _, r := range testRanges {
        if r.start >= r.end || r.start < 0 || r.end > len(origFW) {
            continue
        }
        
        for _, endian := range endianness {
            origData := origFW[r.start:r.end]
            modData := modFW[r.start:r.end]
            
            // Apply endianness conversion
            origDataSwapped := endian.fn(origData)
            modDataSwapped := endian.fn(modData)
            
            // Test all CRC algorithms
            for _, params := range crc16Variants {
                totalTests++
                
                var origCalc, modCalc uint16
                
                if params.Name == "MSTFLINT" || params.Name == "MSTFLINT-0" {
                    origCalc = mstflintSoftwareCRC(origDataSwapped)
                    modCalc = mstflintSoftwareCRC(modDataSwapped)
                } else {
                    origCalc = calcCRC16(origDataSwapped, params)
                    modCalc = calcCRC16(modDataSwapped, params)
                }
                
                if origCalc == origCRC && modCalc == modCRC {
                    fmt.Printf("\n*** EXACT MATCH FOUND! ***\n")
                    fmt.Printf("Range: %s (0x%04X-0x%04X, %d bytes)\n", r.name, r.start, r.end, len(origData))
                    fmt.Printf("Endianness: %s\n", endian.name)
                    fmt.Printf("Algorithm: %s\n", params.Name)
                    fmt.Printf("Parameters: Poly=0x%04X, Init=0x%04X, RefIn=%v, RefOut=%v, XorOut=0x%04X\n",
                        params.Poly, params.Init, params.RefIn, params.RefOut, params.XorOut)
                    fmt.Printf("Original CRC: calculated=0x%04X, stored=0x%04X\n", origCalc, origCRC)
                    fmt.Printf("Modified CRC: calculated=0x%04X, stored=0x%04X\n", modCalc, modCRC)
                    
                    fmt.Println("\nFirst 64 bytes of data (after endian conversion):")
                    for i := 0; i < 64 && i < len(origDataSwapped); i += 16 {
                        fmt.Printf("%04X: ", r.start+i)
                        for j := 0; j < 16 && i+j < 64 && i+j < len(origDataSwapped); j++ {
                            fmt.Printf("%02X ", origDataSwapped[i+j])
                        }
                        fmt.Println()
                    }
                    
                    foundMatch = true
                    goto done
                }
            }
        }
    }
    
    // If no match with key ranges, do exhaustive search with endianness
    if !foundMatch {
        fmt.Println("\nNo match with key ranges. Starting exhaustive search with endianness...")
        
        for endianIdx, endian := range endianness {
            fmt.Printf("\nTesting with %s...\n", endian.name)
            
            // Limit exhaustive search to every 8th position to keep it manageable
            for start := firstFF; start < boardIDStart && start >= 0; start += 8 {
                for end := boardIDEndOrig; end <= nextFF && end <= len(origFW); end += 4 {
                    if start >= end {
                        continue
                    }
                    
                    origData := origFW[start:end]
                    modData := modFW[start:end]
                    
                    // Apply endianness conversion
                    origDataSwapped := endian.fn(origData)
                    modDataSwapped := endian.fn(modData)
                    
                    // Test subset of algorithms for speed
                    testAlgos := []CRC16Params{
                        crc16Variants[0],  // CCITT-FALSE
                        crc16Variants[1],  // ARC
                        crc16Variants[10], // GENIBUS
                        crc16Variants[23], // MSTFLINT
                    }
                    
                    for _, params := range testAlgos {
                        totalTests++
                        
                        origCalc := calcCRC16(origDataSwapped, params)
                        modCalc := calcCRC16(modDataSwapped, params)
                        
                        if origCalc == origCRC && modCalc == modCRC {
                            fmt.Printf("\n*** MATCH FOUND in exhaustive search! ***\n")
                            fmt.Printf("Start: 0x%04X, End: 0x%04X\n", start, end)
                            fmt.Printf("Endianness: %s\n", endian.name)
                            fmt.Printf("Algorithm: %s\n", params.Name)
                            foundMatch = true
                            goto done
                        }
                        
                        if totalTests % 5000 == 0 {
                            fmt.Printf("Tested %d combinations (endian %d/3)...\r", totalTests, endianIdx+1)
                        }
                    }
                }
            }
        }
    }
    
done:
    fmt.Printf("\nTotal combinations tested: %d\n", totalTests)
    
    if !foundMatch {
        fmt.Println("\nNo match found even with endianness conversions.")
        fmt.Println("The CRC uses a proprietary algorithm or non-standard data preprocessing.")
    }
}