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

// mstflint-style CRC32 (if it exists)
func mstflintCRC32(data []byte) uint32 {
    const poly = uint32(0x04C11DB7) // Standard CRC32 poly, but mstflint style processing
    crc := uint32(0xffffffff)
    
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
    
    return crc ^ 0xffffffff
}

func main() {
    origFW, _ := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
    modFW, _ := os.ReadFile("sample_firmwares/franken_fw.bin")
    
    fmt.Println("=== Exhaustive CRC32 Search ===")
    
    // Find boundaries
    firstFF := -1
    for i := 0x200; i < 0x400; i++ {
        if origFW[i] == 0xFF {
            firstFF = i
            fmt.Printf("First 0xFF found at: 0x%04X\n", firstFF)
            break
        }
    }
    
    boardIDStart := 0x4D8
    boardIDEndOrig := 0x4E6
    boardIDEndMod := 0x4EA
    
    nextFF := -1
    for i := 0x500; i < 0x600; i++ {
        if origFW[i] == 0xFF {
            nextFF = i
            fmt.Printf("Next 0xFF found at: 0x%04X\n", nextFF)
            break
        }
    }
    
    // Identify all possible CRC32 locations and values
    fmt.Println("\nSearching for potential CRC32 values that change between files...")
    
    var crcCandidates []struct {
        offset int
        origBE uint32
        modBE  uint32
        origLE uint32
        modLE  uint32
    }
    
    // Check every possible 4-byte location for changes
    for i := boardIDEndOrig; i < nextFF-3 && i < len(origFW)-3; i++ {
        origBE := binary.BigEndian.Uint32(origFW[i : i+4])
        modBE := binary.BigEndian.Uint32(modFW[i : i+4])
        origLE := binary.LittleEndian.Uint32(origFW[i : i+4])
        modLE := binary.LittleEndian.Uint32(modFW[i : i+4])
        
        if origBE != modBE {
            crcCandidates = append(crcCandidates, struct {
                offset int
                origBE uint32
                modBE  uint32
                origLE uint32
                modLE  uint32
            }{i, origBE, modBE, origLE, modLE})
            
            fmt.Printf("Potential CRC32 at 0x%04X: BE: 0x%08X->0x%08X, LE: 0x%08X->0x%08X\n",
                i, origBE, modBE, origLE, modLE)
        }
    }
    
    fmt.Printf("\nFound %d potential CRC32 locations\n", len(crcCandidates))
    
    fmt.Println("\nStarting exhaustive search...")
    totalTests := 0
    foundMatch := false
    
    // For each potential CRC location
    for _, crcCandidate := range crcCandidates {
        fmt.Printf("\nTesting CRC32 at offset 0x%04X...\n", crcCandidate.offset)
        
        // Test all possible start positions (from first FF to board ID start)
        for start := firstFF; start < boardIDStart && start >= 0; start++ {
            // Test all possible end positions (from board ID end to the CRC location)
            for end := boardIDEndOrig; end <= crcCandidate.offset && end <= len(origFW); end++ {
                if start >= end {
                    continue
                }
                
                origData := origFW[start:end]
                
                // For modified firmware, handle the longer string
                modEnd := end
                if end > boardIDEndOrig && end <= boardIDEndMod {
                    modEnd = boardIDEndMod + (end - boardIDEndOrig)
                }
                if modEnd > len(modFW) {
                    continue
                }
                modData := modFW[start:modEnd]
                
                // Test all CRC32 algorithms
                for _, params := range crc32Variants {
                    totalTests++
                    
                    origCalc := calcCRC32(origData, params)
                    modCalc := calcCRC32(modData, params)
                    
                    // Check against both BE and LE values
                    if (origCalc == crcCandidate.origBE && modCalc == crcCandidate.modBE) ||
                       (origCalc == crcCandidate.origLE && modCalc == crcCandidate.modLE) {
                        fmt.Printf("\n*** CRC32 MATCH FOUND! ***\n")
                        fmt.Printf("CRC Location: 0x%04X\n", crcCandidate.offset)
                        fmt.Printf("Data Range: 0x%04X-0x%04X (Length: %d bytes)\n", start, end, len(origData))
                        fmt.Printf("Algorithm: %s\n", params.Name)
                        fmt.Printf("Parameters: Poly=0x%08X, Init=0x%08X, RefIn=%v, RefOut=%v, XorOut=0x%08X\n",
                            params.Poly, params.Init, params.RefIn, params.RefOut, params.XorOut)
                        
                        if origCalc == crcCandidate.origBE {
                            fmt.Println("CRC stored as: BIG ENDIAN")
                            fmt.Printf("Original CRC32: calculated=0x%08X, stored=0x%08X\n", origCalc, crcCandidate.origBE)
                            fmt.Printf("Modified CRC32: calculated=0x%08X, stored=0x%08X\n", modCalc, crcCandidate.modBE)
                        } else {
                            fmt.Println("CRC stored as: LITTLE ENDIAN")
                            fmt.Printf("Original CRC32: calculated=0x%08X, stored=0x%08X\n", origCalc, crcCandidate.origLE)
                            fmt.Printf("Modified CRC32: calculated=0x%08X, stored=0x%08X\n", modCalc, crcCandidate.modLE)
                        }
                        
                        fmt.Println("\nData covered by CRC (first 64 bytes):")
                        for i := 0; i < 64 && i < len(origData); i += 16 {
                            fmt.Printf("%04X: ", start+i)
                            for j := 0; j < 16 && i+j < 64 && i+j < len(origData); j++ {
                                fmt.Printf("%02X ", origData[i+j])
                            }
                            fmt.Println()
                        }
                        
                        foundMatch = true
                        goto done
                    }
                    
                    // Also test mstflint-style CRC32
                    origCalcMst := mstflintCRC32(origData)
                    modCalcMst := mstflintCRC32(modData)
                    
                    if (origCalcMst == crcCandidate.origBE && modCalcMst == crcCandidate.modBE) ||
                       (origCalcMst == crcCandidate.origLE && modCalcMst == crcCandidate.modLE) {
                        fmt.Printf("\n*** MSTFLINT CRC32 MATCH! ***\n")
                        fmt.Printf("CRC Location: 0x%04X\n", crcCandidate.offset)
                        fmt.Printf("Data Range: 0x%04X-0x%04X\n", start, end)
                        foundMatch = true
                        goto done
                    }
                    
                    if totalTests % 10000 == 0 {
                        fmt.Printf("Tested %d combinations...\r", totalTests)
                    }
                }
            }
        }
    }
    
done:
    fmt.Printf("\nTotal combinations tested: %d\n", totalTests)
    
    if !foundMatch {
        fmt.Println("\nNo CRC32 match found in exhaustive search.")
        fmt.Println("The checksum is not a standard CRC32 algorithm.")
    }
}