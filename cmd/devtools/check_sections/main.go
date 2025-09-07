package main

import (
    "encoding/hex"
    "fmt"
    "os"
    "path/filepath"

    "go.uber.org/zap"

    cliutil "github.com/Civil/mlx5fw-go/pkg/cliutil"
    "github.com/Civil/mlx5fw-go/pkg/interfaces"
    "github.com/Civil/mlx5fw-go/pkg/types"
)

func hexDump(b []byte, n int) string {
    if len(b) > n {
        b = b[:n]
    }
    return hex.EncodeToString(b)
}

func cmpSectionBytes(name string, orig, marshaled []byte) {
    same := len(orig) == len(marshaled)
    if same {
        for i := range orig {
            if orig[i] != marshaled[i] {
                same = false
                break
            }
        }
    }
    if same {
        fmt.Printf("OK %s: bytes match (len=%d)\n", name, len(orig))
        return
    }
    // Show first difference
    max := len(orig)
    if len(marshaled) < max {
        max = len(marshaled)
    }
    idx := -1
    for i := 0; i < max; i++ {
        if orig[i] != marshaled[i] {
            idx = i
            break
        }
    }
    fmt.Printf("DIFF %s: len orig=%d marshaled=%d first_diff_at=%d\n", name, len(orig), len(marshaled), idx)
    fmt.Printf("  orig[0:32]=%s\n", hexDump(orig, 32))
    fmt.Printf("  marsh[0:32]=%s\n", hexDump(marshaled, 32))
}

func main() {
    logger, _ := zap.NewDevelopment()
    defer logger.Sync()

    if len(os.Args) < 2 {
        fmt.Fprintf(os.Stderr, "usage: %s <firmware.bin> [more...]\n", filepath.Base(os.Args[0]))
        os.Exit(2)
    }

    for _, fw := range os.Args[1:] {
        fmt.Printf("\n== %s ==\n", fw)
        ctx, err := cliutil.InitializeFirmwareParser(fw, logger)
        if err != nil {
            fmt.Printf("init parser: %v\n", err)
            continue
        }
        p := ctx.Parser

        // IMAGE_INFO
        if secs := p.GetSections()[types.SectionTypeImageInfo]; len(secs) > 0 {
            s := secs[0].(interfaces.CompleteSectionInterface)
            data := s.GetRawData()
            if data == nil || len(data) == 0 {
                // Read directly
                raw, _ := p.ReadSectionData(s.Type(), s.Offset(), s.Size())
                data = raw
            }
            // Trim CRC if present
            if s.CRCType() == types.CRCInSection && !p.IsEncrypted() && len(data) >= 4 {
                data = data[:len(data)-4]
            }
            var ii types.ImageInfo
            if err := ii.Unmarshal(data); err != nil {
                fmt.Printf("IMAGE_INFO unmarshal: %v\n", err)
            } else {
                marsh, err := ii.Marshal()
                if err != nil {
                    fmt.Printf("IMAGE_INFO marshal: %v\n", err)
                } else {
                    // Ensure 1024 bytes
                    if len(marsh) < 1024 {
                        tmp := make([]byte, 1024)
                        copy(tmp, marsh)
                        marsh = tmp
                    }
                    cmpSectionBytes("IMAGE_INFO", data[:1024], marsh[:1024])
                }
            }
        } else {
            fmt.Printf("no IMAGE_INFO\n")
        }

        // DEV_INFO
        if secs := p.GetSections()[types.SectionTypeDevInfo]; len(secs) > 0 {
            s := secs[0].(interfaces.CompleteSectionInterface)
            data := s.GetRawData()
            if data == nil || len(data) == 0 {
                raw, _ := p.ReadSectionData(s.Type(), s.Offset(), s.Size())
                data = raw
            }
            if s.CRCType() == types.CRCInSection && !p.IsEncrypted() && len(data) >= 4 {
                data = data[:len(data)-4]
            }
            var di types.DevInfo
            if err := di.Unmarshal(data); err != nil {
                fmt.Printf("DEV_INFO unmarshal: %v\n", err)
            } else {
                // CRC is stored in last dword, keep whatever is in data by setting it explicitly
                marsh, err := di.Marshal()
                if err != nil {
                    fmt.Printf("DEV_INFO marshal: %v\n", err)
                } else {
                    if len(marsh) < len(data) {
                        tmp := make([]byte, len(data))
                        copy(tmp, marsh)
                        marsh = tmp
                    }
                    cmpSectionBytes("DEV_INFO", data[:512], marsh[:512])
                }
            }
        } else {
            fmt.Printf("no DEV_INFO\n")
        }

        // MFG_INFO
        if secs := p.GetSections()[types.SectionTypeMfgInfo]; len(secs) > 0 {
            s := secs[0].(interfaces.CompleteSectionInterface)
            data := s.GetRawData()
            if data == nil || len(data) == 0 {
                raw, _ := p.ReadSectionData(s.Type(), s.Offset(), s.Size())
                data = raw
            }
            if s.CRCType() == types.CRCInSection && !p.IsEncrypted() && len(data) >= 4 {
                data = data[:len(data)-4]
            }
            var mi types.MfgInfo
            if err := mi.Unmarshal(data); err != nil {
                fmt.Printf("MFG_INFO unmarshal: %v\n", err)
            } else {
                marsh, err := mi.Marshal()
                if err != nil {
                    fmt.Printf("MFG_INFO marshal: %v\n", err)
                } else {
                    // MFG_INFO may be 256 or 320, compare first len(marsh) bytes
                    cmpLen := len(data)
                    if len(marsh) < cmpLen { cmpLen = len(marsh) }
                    cmpSectionBytes(fmt.Sprintf("MFG_INFO(%d)", cmpLen), data[:cmpLen], marsh[:cmpLen])
                }
            }
        } else {
            fmt.Printf("no MFG_INFO\n")
        }

        ctx.Close()
    }
}

