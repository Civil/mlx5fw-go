package main

import (
    "encoding/binary"
    "fmt"
    "os"
)

// CRC32 algorithm parameters
type CRC32Params struct {
    Name   string
    Poly   uint32
    Init   uint32
    RefIn  bool
    RefOut bool
    XorOut uint32
}

// All known CRC32 variants
var crc32Variants = []CRC32Params{
    {"CRC-32", 0x04C11DB7, 0xFFFFFFFF, true, true, 0xFFFFFFFF},
    {"CRC-32/BZIP2", 0x04C11DB7, 0xFFFFFFFF, false, false, 0xFFFFFFFF},
    {"CRC-32C", 0x1EDC6F41, 0xFFFFFFFF, true, true, 0xFFFFFFFF},
    {"CRC-32D", 0xA833982B, 0xFFFFFFFF, true, true, 0xFFFFFFFF},
    {"CRC-32/MPEG-2", 0x04C11DB7, 0xFFFFFFFF, false, false, 0x00000000},
    {"CRC-32/POSIX", 0x04C11DB7, 0x00000000, false, false, 0xFFFFFFFF},
    {"CRC-32Q", 0x814141AB, 0x00000000, false, false, 0x00000000},
    {"CRC-32/JAMCRC", 0x04C11DB7, 0xFFFFFFFF, true, true, 0x00000000},
    {"CRC-32/AUTOSAR", 0xF4ACFB13, 0xFFFFFFFF, true, true, 0xFFFFFFFF},
    {"CRC-32/XFER", 0x000000AF, 0x00000000, false, false, 0x00000000},
    {"CRC-32/ISO-HDLC", 0x04C11DB7, 0xFFFFFFFF, true, true, 0xFFFFFFFF},
    {"CRC-32/MEF", 0x741B8CD7, 0xFFFFFFFF, true, true, 0x00000000},
    {"CRC-32/CD-ROM-EDC", 0x8001801B, 0x00000000, true, true, 0x00000000},
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

// Reverse bits in a uint32
func reverseBits32(n uint32) uint32 {
    var result uint32
    for i := 0; i < 32; i++ {
        result = (result << 1) | (n & 1)
        n >>= 1
    }
    return result
}

// Generic CRC32 calculation
func calcCRC32(data []byte, params CRC32Params) uint32 {
    crc := params.Init
    
    for _, b := range data {
        if params.RefIn {
            b = reverseBits8(b)
        }
        
        crc ^= uint32(b) << 24
        for i := 0; i < 8; i++ {
            if crc&0x80000000 != 0 {
                crc = (crc << 1) ^ params.Poly
            } else {
                crc <<= 1
            }
        }
    }
    
    if params.RefOut {
        crc = reverseBits32(crc)
    }
    
    return crc ^ params.XorOut
}

func main() {
    origFW, _ := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
    modFW, _ := os.ReadFile("sample_firmwares/franken_fw.bin")
    
    fmt.Println("=== CRC32 Search for ARM Boot Args ===\n")
    
    // Let's first check what 32-bit values are at potential CRC locations
    fmt.Println("Checking potential CRC32 locations:")
    
    // Check around 0x4FE (where we thought CRC16 was)
    locations := []int{
        0x4FC, 0x4FD, 0x4FE, 0x4FF, 0x500,  // Around the known change
        0x4F8, 0x4F9, 0x4FA, 0x4FB,         // Before the change
        0x538, 0x539, 0x53A, 0x53B, 0x53C,  // Near 0x83DC
    }
    
    for _, loc := range locations {
        if loc+4 <= len(origFW) && loc+4 <= len(modFW) {
            origVal := binary.BigEndian.Uint32(origFW[loc : loc+4])
            modVal := binary.BigEndian.Uint32(modFW[loc : loc+4])
            origValLE := binary.LittleEndian.Uint32(origFW[loc : loc+4])
            modValLE := binary.LittleEndian.Uint32(modFW[loc : loc+4])
            
            if origVal != modVal {
                fmt.Printf("0x%04X: BE: 0x%08X -> 0x%08X (CHANGED)\n", loc, origVal, modVal)
                fmt.Printf("        LE: 0x%08X -> 0x%08X\n", origValLE, modValLE)
            }
        }
    }
    
    // The value that changes is at 0x4FC-0x4FF (4 bytes)
    // Original: 0x00006885 (BE) or 0x85680000 (LE)
    // Modified: 0x00005220 (BE) or 0x20520000 (LE)
    
    origCRC32_BE := uint32(0x00006885)
    modCRC32_BE := uint32(0x00005220)
    origCRC32_LE := uint32(0x85680000)
    modCRC32_LE := uint32(0x20520000)
    
    // Also check if it might be at 0x4FE as full 32-bit
    if len(origFW) > 0x502 {
        origCRC32_4FE_BE := binary.BigEndian.Uint32(origFW[0x4FE:0x502])
        modCRC32_4FE_BE := binary.BigEndian.Uint32(modFW[0x4FE:0x502])
        origCRC32_4FE_LE := binary.LittleEndian.Uint32(origFW[0x4FE:0x502])
        modCRC32_4FE_LE := binary.LittleEndian.Uint32(modFW[0x4FE:0x502])
        
        fmt.Printf("\nCRC32 at 0x4FE-0x501:\n")
        fmt.Printf("BE: 0x%08X -> 0x%08X\n", origCRC32_4FE_BE, modCRC32_4FE_BE)
        fmt.Printf("LE: 0x%08X -> 0x%08X\n", origCRC32_4FE_LE, modCRC32_4FE_LE)
    }
    
    fmt.Println("\nStarting CRC32 search...")
    
    // Test ranges
    testRanges := []struct {
        name  string
        start int
        end   int
    }{
        // From end of padding to potential CRC locations
        {"After padding to 0x4FC", 0x28C, 0x4FC},
        {"After padding to 0x4FE", 0x28C, 0x4FE},
        {"From 0x300 to 0x4FC", 0x300, 0x4FC},
        {"From 0x300 to 0x4FE", 0x300, 0x4FE},
        {"From 0x400 to 0x4FC", 0x400, 0x4FC},
        {"From 0x400 to 0x4FE", 0x400, 0x4FE},
        {"Board ID to 0x4FC", 0x4D8, 0x4FC},
        {"With length to 0x4FC", 0x4D6, 0x4FC},
        {"Full struct to 0x4FC", 0x4D0, 0x4FC},
        {"From 0x4C0 to 0x4FC", 0x4C0, 0x4FC},
    }
    
    // Test with different CRC32 target values
    targetCRCs := []struct {
        name string
        orig uint32
        mod  uint32
    }{
        {"BE at 0x4FC", origCRC32_BE, modCRC32_BE},
        {"LE at 0x4FC", origCRC32_LE, modCRC32_LE},
        {"BE at 0x4FE", 0x68850000, 0x52200000},  // If CRC32 is at 0x4FE
        {"LE at 0x4FE", 0x00008568, 0x00002052},  // If CRC32 is at 0x4FE
    }
    
    foundMatch := false
    
    for _, target := range targetCRCs {
        fmt.Printf("\nTesting with CRC32 target: %s (0x%08X -> 0x%08X)\n", 
            target.name, target.orig, target.mod)
        
        for _, r := range testRanges {
            if r.start >= r.end || r.start < 0 || r.end > len(origFW) {
                continue
            }
            
            origData := origFW[r.start:r.end]
            modData := modFW[r.start:r.end]
            
            // Test all CRC32 algorithms
            for _, params := range crc32Variants {
                origCalc := calcCRC32(origData, params)
                modCalc := calcCRC32(modData, params)
                
                if origCalc == target.orig && modCalc == target.mod {
                    fmt.Printf("\n*** CRC32 MATCH FOUND! ***\n")
                    fmt.Printf("Range: %s (0x%04X-0x%04X, %d bytes)\n", 
                        r.name, r.start, r.end, len(origData))
                    fmt.Printf("Algorithm: %s\n", params.Name)
                    fmt.Printf("Parameters: Poly=0x%08X, Init=0x%08X, RefIn=%v, RefOut=%v, XorOut=0x%08X\n",
                        params.Poly, params.Init, params.RefIn, params.RefOut, params.XorOut)
                    fmt.Printf("Original CRC32: calculated=0x%08X, stored=0x%08X\n", origCalc, target.orig)
                    fmt.Printf("Modified CRC32: calculated=0x%08X, stored=0x%08X\n", modCalc, target.mod)
                    
                    fmt.Println("\nData covered by CRC (first 64 bytes):")
                    for i := 0; i < 64 && i < len(origData); i += 16 {
                        fmt.Printf("%04X: ", r.start+i)
                        for j := 0; j < 16 && i+j < 64 && i+j < len(origData); j++ {
                            fmt.Printf("%02X ", origData[i+j])
                        }
                        fmt.Println()
                    }
                    
                    foundMatch = true
                    goto done
                }
            }
        }
    }
    
    // If no match with standard ranges, try exhaustive search on smaller range
    if !foundMatch {
        fmt.Println("\nNo match with standard ranges. Trying focused exhaustive search...")
        
        firstFF := 0x277
        boardIDStart := 0x4D8
        boardIDEnd := 0x4E6
        
        tested := 0
        // Test more focused ranges
        for start := firstFF; start < boardIDStart && start >= 0; start += 16 { // Step by 16 for speed
            for end := boardIDEnd; end <= 0x500 && end <= len(origFW); end += 4 { // Step by 4
                if start >= end {
                    continue
                }
                
                origData := origFW[start:end]
                modData := modFW[start:end]
                
                for _, target := range targetCRCs {
                    for _, params := range crc32Variants {
                        tested++
                        origCalc := calcCRC32(origData, params)
                        modCalc := calcCRC32(modData, params)
                        
                        if origCalc == target.orig && modCalc == target.mod {
                            fmt.Printf("\n*** CRC32 MATCH FOUND! ***\n")
                            fmt.Printf("Start: 0x%04X, End: 0x%04X (Length: %d bytes)\n", 
                                start, end, len(origData))
                            fmt.Printf("Algorithm: %s\n", params.Name)
                            fmt.Printf("Target: %s\n", target.name)
                            foundMatch = true
                            goto done
                        }
                        
                        if tested % 10000 == 0 {
                            fmt.Printf("Tested %d combinations...\r", tested)
                        }
                    }
                }
            }
        }
    }
    
done:
    if !foundMatch {
        fmt.Println("\nNo CRC32 match found.")
        fmt.Println("The checksum is either:")
        fmt.Println("1. A proprietary algorithm (neither standard CRC16 nor CRC32)")
        fmt.Println("2. Uses a polynomial not in our test set")
        fmt.Println("3. Is not actually at 0x4FC-0x4FF")
    }
}