package main

import (
    "encoding/binary"
    "fmt"
    "os"
)

// CRC16 algorithm parameters
type CRCParams struct {
    Name   string
    Poly   uint16
    Init   uint16
    RefIn  bool
    RefOut bool
    XorOut uint16
}

// All known CRC16 variants
var crcVariants = []CRCParams{
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
    {"MSTFLINT-SOFT", 0x100B, 0xFFFF, false, false, 0xFFFF},
    {"MSTFLINT-BYTE", 0x100B, 0xFFFF, false, false, 0xFFFF},
    {"MSTFLINT-HARD", 0x100B, 0xFFFF, false, false, 0x0000},
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
func calcCRC16(data []byte, params CRCParams) uint16 {
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

// mstflint Hardware CRC (with inverted first bytes)
func mstflintHardwareCRC(data []byte) uint16 {
    if len(data) < 2 {
        return 0
    }
    
    modData := make([]byte, len(data))
    copy(modData, data)
    modData[0] = ^modData[0]
    modData[1] = ^modData[1]
    
    params := CRCParams{"", 0x100B, 0xFFFF, false, false, 0x0000}
    return calcCRC16(modData, params)
}

func main() {
    origFW, err := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
    if err != nil {
        panic(err)
    }
    
    modFW, err := os.ReadFile("sample_firmwares/franken_fw.bin")
    if err != nil {
        panic(err)
    }
    
    // Target CRCs
    origCRC := uint16(0x6885)
    modCRC := uint16(0x5220)
    
    fmt.Println("=== Comprehensive ARM Boot Args CRC Analysis ===")
    fmt.Printf("Target CRCs: Original=0x%04X, Modified=0x%04X\n\n", origCRC, modCRC)
    
    // Define all possible data ranges to test
    testRanges := []struct {
        name  string
        start int
        end   int
    }{
        // Board ID string variations
        {"Board ID only (14 bytes)", 0x4D8, 0x4E6},
        {"Board ID only (18 bytes)", 0x4D8, 0x4EA},
        {"Board ID field (32 bytes)", 0x4D8, 0x4F8},
        
        // With value after string
        {"Board ID + value", 0x4D8, 0x4FC},
        {"Board ID + value + pad", 0x4D8, 0x4FE},
        
        // With length prefix
        {"Length + Board ID", 0x4D6, 0x4F8},
        {"Length + Board ID + value", 0x4D6, 0x4FC},
        {"Length + Board ID + all", 0x4D6, 0x4FE},
        
        // Full structure variations
        {"Full struct from 0x4D0", 0x4D0, 0x4FC},
        {"Full struct to CRC", 0x4D0, 0x4FE},
        
        // Extended ranges
        {"From 0x4C0", 0x4C0, 0x4FC},
        {"From 0x4C0 to CRC", 0x4C0, 0x4FE},
        {"From 0x4C8", 0x4C8, 0x4FC},
        {"From 0x4C8 to CRC", 0x4C8, 0x4FE},
        
        // Aligned boundaries
        {"16-byte aligned", 0x4D0, 0x4E0},
        {"32-byte aligned", 0x4C0, 0x4E0},
        {"64-byte aligned", 0x4C0, 0x500},
        
        // From various starting points
        {"From 0x480", 0x480, 0x4FE},
        {"From 0x490", 0x490, 0x4FE},
        {"From 0x4A0", 0x4A0, 0x4FE},
        {"From 0x4B0", 0x4B0, 0x4FE},
        
        // Specific offsets that might be significant
        {"After pattern", 0x300, 0x4FE},
        {"From 0x400", 0x400, 0x4FE},
        {"From 0x450", 0x450, 0x4FE},
        
        // Very specific ranges
        {"40 bytes exact", 0x4D6, 0x4FE},
        {"42 bytes", 0x4D4, 0x4FE},
        {"44 bytes", 0x4D2, 0x4FE},
        {"46 bytes", 0x4D0, 0x4FE},
        
        // Reverse direction (maybe CRC is before data)
        {"After CRC", 0x500, 0x520},
        {"Including CRC", 0x4FE, 0x520},
    }
    
    matchFound := false
    
    // Test each range with each CRC variant
    for _, r := range testRanges {
        if r.start >= r.end || r.start < 0 || r.end > len(origFW) {
            continue
        }
        
        origData := origFW[r.start:r.end]
        modData := modFW[r.start:r.end]
        
        // Test standard CRC variants
        for _, params := range crcVariants {
            var origCalc, modCalc uint16
            
            if params.Name == "MSTFLINT-SOFT" {
                origCalc = mstflintSoftwareCRC(origData)
                modCalc = mstflintSoftwareCRC(modData)
            } else if params.Name == "MSTFLINT-HARD" {
                origCalc = mstflintHardwareCRC(origData)
                modCalc = mstflintHardwareCRC(modData)
            } else {
                origCalc = calcCRC16(origData, params)
                modCalc = calcCRC16(modData, params)
            }
            
            // Check for exact match
            if origCalc == origCRC && modCalc == modCRC {
                fmt.Printf("\n*** EXACT MATCH FOUND! ***\n")
                fmt.Printf("Range: %s (0x%04X-0x%04X, %d bytes)\n", r.name, r.start, r.end, len(origData))
                fmt.Printf("Algorithm: %s\n", params.Name)
                fmt.Printf("Parameters: Poly=0x%04X, Init=0x%04X, RefIn=%v, RefOut=%v, XorOut=0x%04X\n",
                    params.Poly, params.Init, params.RefIn, params.RefOut, params.XorOut)
                fmt.Printf("CRC Results: Orig=0x%04X, Mod=0x%04X\n", origCalc, modCalc)
                
                fmt.Println("\nData being protected (first 64 bytes):")
                for i := 0; i < 64 && i < len(origData); i += 16 {
                    fmt.Printf("%04X: ", r.start+i)
                    for j := 0; j < 16 && i+j < 64 && i+j < len(origData); j++ {
                        fmt.Printf("%02X ", origData[i+j])
                    }
                    fmt.Println()
                }
                
                matchFound = true
                return
            }
            
            // Also check with byte-swapped CRC values
            origCRCSwap := (origCRC >> 8) | (origCRC << 8)
            modCRCSwap := (modCRC >> 8) | (modCRC << 8)
            
            if origCalc == origCRCSwap && modCalc == modCRCSwap {
                fmt.Printf("\n*** MATCH with byte-swapped CRC! ***\n")
                fmt.Printf("Range: %s, Algorithm: %s\n", r.name, params.Name)
                fmt.Printf("CRC stored as little-endian!\n")
                matchFound = true
                return
            }
        }
    }
    
    if !matchFound {
        fmt.Println("\nNo exact match found. Showing partial matches:\n")
        
        // Show partial matches for debugging
        for _, r := range testRanges {
            if r.start >= r.end || r.start < 0 || r.end > len(origFW) {
                continue
            }
            
            origData := origFW[r.start:r.end]
            modData := modFW[r.start:r.end]
            
            for _, params := range crcVariants {
                var origCalc, modCalc uint16
                
                if params.Name == "MSTFLINT-SOFT" {
                    origCalc = mstflintSoftwareCRC(origData)
                    modCalc = mstflintSoftwareCRC(modData)
                } else if params.Name == "MSTFLINT-HARD" {
                    origCalc = mstflintHardwareCRC(origData)
                    modCalc = mstflintHardwareCRC(modData)
                } else {
                    origCalc = calcCRC16(origData, params)
                    modCalc = calcCRC16(modData, params)
                }
                
                if origCalc == origCRC || modCalc == modCRC {
                    fmt.Printf("Partial: %s + %s: ", r.name, params.Name)
                    if origCalc == origCRC {
                        fmt.Printf("orig matches (0x%04X) ", origCalc)
                    }
                    if modCalc == modCRC {
                        fmt.Printf("mod matches (0x%04X) ", modCalc)
                    }
                    fmt.Println()
                }
            }
        }
    }
}