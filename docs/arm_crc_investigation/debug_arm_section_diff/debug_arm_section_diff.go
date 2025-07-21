package main

import (
    "encoding/binary"
    "fmt"
    "os"
)

func main() {
    origFW, _ := os.ReadFile("sample_firmwares/MBF2M345A-HECO_Ax_MT_0000000716_rel-24_40.1000.bin")
    modFW, _ := os.ReadFile("sample_firmwares/franken_fw.bin")
    
    fmt.Println("=== Full Section Difference Analysis ===\n")
    
    // Show all differences in the extended range
    fmt.Println("All differences from 0x200 to 0x600:")
    diffCount := 0
    
    for i := 0x200; i < 0x600 && i < len(origFW) && i < len(modFW); i++ {
        if origFW[i] != modFW[i] {
            fmt.Printf("0x%04X: 0x%02X -> 0x%02X", i, origFW[i], modFW[i])
            if origFW[i] >= 32 && origFW[i] < 127 && modFW[i] >= 32 && modFW[i] < 127 {
                fmt.Printf(" ('%c' -> '%c')", origFW[i], modFW[i])
            }
            fmt.Println()
            diffCount++
        }
    }
    
    fmt.Printf("\nTotal differences: %d\n", diffCount)
    
    // Show context around each difference
    fmt.Println("\n=== Context around differences ===")
    
    // Board ID change
    fmt.Println("\nBoard ID change context (0x4D0-0x510):")
    for i := 0x4D0; i < 0x510; i += 16 {
        fmt.Printf("%04X: ", i)
        for j := 0; j < 16 && i+j < 0x510; j++ {
            if origFW[i+j] != modFW[i+j] {
                fmt.Printf("[%02X->%02X] ", origFW[i+j], modFW[i+j])
            } else {
                fmt.Printf("%02X ", origFW[i+j])
            }
        }
        fmt.Println()
    }
    
    // Let's also check if there are any other CRC-like values that change
    fmt.Println("\n=== Searching for other changing 16-bit values ===")
    
    for i := 0x200; i < 0x600-2; i += 2 {
        origVal := binary.BigEndian.Uint16(origFW[i : i+2])
        modVal := binary.BigEndian.Uint16(modFW[i : i+2])
        
        if origVal != modVal && origVal != 0x0000 && modVal != 0x0000 {
            fmt.Printf("0x%04X: 0x%04X -> 0x%04X", i, origVal, modVal)
            
            // Check if this could be our CRC
            if (i < 0x4D8 || i > 0x4F0) { // Outside the string area
                fmt.Printf(" (potential CRC location)")
            }
            fmt.Println()
        }
    }
    
    // Display the full section to understand structure
    fmt.Println("\n=== Full section display (0x250-0x550) ===")
    fmt.Println("Showing original firmware:")
    
    for i := 0x250; i < 0x550 && i < len(origFW); i += 16 {
        fmt.Printf("%04X: ", i)
        for j := 0; j < 16 && i+j < 0x550 && i+j < len(origFW); j++ {
            fmt.Printf("%02X ", origFW[i+j])
        }
        fmt.Print(" |")
        for j := 0; j < 16 && i+j < 0x550 && i+j < len(origFW); j++ {
            if origFW[i+j] >= 32 && origFW[i+j] < 127 {
                fmt.Printf("%c", origFW[i+j])
            } else {
                fmt.Print(".")
            }
        }
        fmt.Println("|")
        
        // Mark section boundaries
        if i == 0x270 {
            fmt.Println("      ^^^^^^^^ 0xFFFF padding starts")
        } else if i == 0x280 {
            fmt.Println("      ^^^^^^^^ End of 0xFFFF padding")
        } else if i == 0x4D0 {
            fmt.Println("      ^^^^^^^^ Boot args structure starts")
        } else if i == 0x4F0 {
            fmt.Println("      ^^^^^^^^ CRC location at 0x4FE")
        } else if i == 0x530 {
            fmt.Println("      ^^^^^^^^ Near end of section")
        }
    }
}