package main

import (
    "fmt"
    "os"
)

const CRCPolynomial = 0x100b

// Exact copy of CalculateSoftwareCRC16 from the codebase
func calculateSoftwareCRC16(data []byte) uint16 {
    // Process data as 32-bit words (big-endian)
    // Pad data to align to 4 bytes if needed
    dataLen := len(data)
    paddedLen := (dataLen + 3) & ^3 // Round up to multiple of 4
    paddedData := make([]byte, paddedLen)
    copy(paddedData, data)
    
    crc := uint16(0xFFFF)
    
    // Process 32-bit words
    for i := 0; i < paddedLen; i += 4 {
        // Get 32-bit word in big-endian
        word := uint32(paddedData[i])<<24 | uint32(paddedData[i+1])<<16 | 
            uint32(paddedData[i+2])<<8 | uint32(paddedData[i+3])
        
        // Process each bit of the 32-bit word (matches mstflint's Crc16::add)
        for j := 0; j < 32; j++ {
            if crc & 0x8000 != 0 {
                crc = ((crc << 1) | uint16(word >> 31)) ^ CRCPolynomial
            } else {
                crc = (crc << 1) | uint16(word >> 31)
            }
            crc &= 0xFFFF
            word <<= 1
        }
    }
    
    // Finish step - process 16 more bits of zeros
    for i := 0; i < 16; i++ {
        if crc&0x8000 != 0 {
            crc = (crc << 1) ^ CRCPolynomial
        } else {
            crc = crc << 1
        }
        crc &= 0xFFFF
    }
    
    // Final XOR
    return crc ^ 0xFFFF
}

// Simple byte-by-byte CRC16 with 0x100b polynomial
func calculateByteCRC16(data []byte) uint16 {
    crc := uint16(0xFFFF)
    
    for _, b := range data {
        crc ^= uint16(b) << 8
        for i := 0; i < 8; i++ {
            if crc&0x8000 != 0 {
                crc = (crc << 1) ^ CRCPolynomial
            } else {
                crc <<= 1
            }
        }
    }
    
    return crc ^ 0xFFFF
}

func main() {
    origFW, _ := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
    modFW, _ := os.ReadFile("sample_firmwares/franken_fw.bin")
    
    origCRC := uint16(0x6885)
    modCRC := uint16(0x5220)
    
    fmt.Println("=== Testing with exact mstflint CRC implementation ===\n")
    
    // Key insight: The board ID field is exactly 32 bytes at 0x4D8-0x4F8
    // After that is a 4-byte value (0x000109D8) at 0x4F8-0x4FC
    // Then 2 bytes padding at 0x4FC-0x4FE
    // Then the CRC at 0x4FE-0x500
    
    testRanges := []struct {
        name   string
        start  int
        end    int
    }{
        // Test different starting points with the 32-byte string field
        {"32-byte string only", 0x4D8, 0x4F8},
        {"String + 4-byte value", 0x4D8, 0x4FC},
        {"String + value + padding", 0x4D8, 0x4FE},
        
        // With various prefixes
        {"With 2-byte length", 0x4D6, 0x4FC},
        {"With 8-byte header", 0x4D0, 0x4FC},
        {"From 0x4C0", 0x4C0, 0x4FC},
        
        // Different end points
        {"To CRC field", 0x4D0, 0x4FE},
        {"To end of value", 0x4D6, 0x4FC},
        {"Exact 40 bytes", 0x4D6, 0x4FE},
    }
    
    fmt.Println("Testing Software CRC16 (32-bit word processing):")
    for _, r := range testRanges {
        if r.start >= r.end || r.end > len(origFW) {
            continue
        }
        
        origData := origFW[r.start:r.end]
        modData := modFW[r.start:r.end]
        
        origCalc := calculateSoftwareCRC16(origData)
        modCalc := calculateSoftwareCRC16(modData)
        
        fmt.Printf("%-25s: orig=0x%04X, mod=0x%04X", r.name, origCalc, modCalc)
        
        if origCalc == origCRC && modCalc == modCRC {
            fmt.Printf(" *** EXACT MATCH! ***")
        } else if origCalc == origCRC {
            fmt.Printf(" (orig matches)")
        } else if modCalc == modCRC {
            fmt.Printf(" (mod matches)")
        }
        fmt.Println()
    }
    
    fmt.Println("\nTesting Byte CRC16 (byte-by-byte processing):")
    for _, r := range testRanges {
        if r.start >= r.end || r.end > len(origFW) {
            continue
        }
        
        origData := origFW[r.start:r.end]
        modData := modFW[r.start:r.end]
        
        origCalc := calculateByteCRC16(origData)
        modCalc := calculateByteCRC16(modData)
        
        fmt.Printf("%-25s: orig=0x%04X, mod=0x%04X", r.name, origCalc, modCalc)
        
        if origCalc == origCRC && modCalc == modCRC {
            fmt.Printf(" *** EXACT MATCH! ***")
        } else if origCalc == origCRC {
            fmt.Printf(" (orig matches)")
        } else if modCalc == modCRC {
            fmt.Printf(" (mod matches)")
        }
        fmt.Println()
    }
    
    // Let's also check with hardware CRC (inverted first bytes)
    fmt.Println("\nTesting with inverted first 2 bytes:")
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
        
        origCalc := calculateByteCRC16(origData)
        modCalc := calculateByteCRC16(modData)
        
        if origCalc == origCRC && modCalc == modCRC {
            fmt.Printf("*** MATCH with inverted bytes! %s ***\n", r.name)
            return
        }
    }
}