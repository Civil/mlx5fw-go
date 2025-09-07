package main

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "os"
    "sort"
    "strings"

    "github.com/spf13/cobra"
    "go.uber.org/zap"

    cliutil "github.com/Civil/mlx5fw-go/pkg/cliutil"
    "github.com/Civil/mlx5fw-go/pkg/diffutil"
    "github.com/Civil/mlx5fw-go/pkg/types"
)

// Diff command that mirrors cmd/devtools/diff_firmware with type-based matching

func sha256Hex(b []byte) string {
    h := sha256.Sum256(b)
    return hex.EncodeToString(h[:])
}

func readAll(path string) ([]byte, error) { return os.ReadFile(path) }

type rawDiffSpan struct {
    Off   int64  `json:"off"`
    Len   int64  `json:"len"`
    Afirst byte  `json:"a_first"`
    Bfirst byte  `json:"b_first"`
}

func diffRaw(a, b []byte, maxSpans int) []rawDiffSpan {
    var spans []rawDiffSpan
    la, lb := int64(len(a)), int64(len(b))
    i := int64(0)
    max := la
    if lb < max { max = lb }
    for i < max && len(spans) < maxSpans {
        if a[i] == b[i] { i++; continue }
        start := i
        afirst, bfirst := a[i], b[i]
        for i < max && a[i] != b[i] { i++ }
        spans = append(spans, rawDiffSpan{Off: start, Len: i - start, Afirst: afirst, Bfirst: bfirst})
    }
    if len(spans) < maxSpans && la != lb {
        spans = append(spans, rawDiffSpan{Off: max, Len: (la - lb)})
    }
    return spans
}

// filter helpers are in pkg/cliutil

type sectionDiffJSON struct {
    Name      string `json:"name"`
    Type      uint16 `json:"type"`
    OffsetA   uint64 `json:"offset_a"`
    OffsetB   uint64 `json:"offset_b"`
    SizeA     uint32 `json:"size_a"`
    SizeB     uint32 `json:"size_b"`
    CRCA      string `json:"crc_type_a"`
    CRCB      string `json:"crc_type_b"`
    AlgoA     string `json:"crc_algo_a"`
    AlgoB     string `json:"crc_algo_b"`
    EncryptedA bool  `json:"encrypted_a"`
    EncryptedB bool  `json:"encrypted_b"`
    Identical bool   `json:"identical"`
    FirstDiff uint32 `json:"first_diff_offset"`
    MissingIn string `json:"missing_in,omitempty"`
}

type jsonReport struct {
    Raw struct {
        Enabled   bool          `json:"enabled"`
        SizeA     int           `json:"size_a"`
        SizeB     int           `json:"size_b"`
        SHA256A   string        `json:"sha256_a"`
        SHA256B   string        `json:"sha256_b"`
        Identical bool          `json:"identical"`
        Spans     []rawDiffSpan `json:"spans,omitempty"`
    } `json:"raw"`
    Sections struct {
        Enabled bool              `json:"enabled"`
        Diffs   int               `json:"diff_count"`
        Missing int               `json:"missing_count"`
        Items   []sectionDiffJSON `json:"items"`
    } `json:"sections"`
}

func CreateDiffFirmwareCommand() *cobra.Command {
    var aPath, bPath string
    var doRaw, doSections bool
    var maxSpans int
    var jsonOut bool
    var filterTypes, filterRange string
    var hexDump bool
    var hexDumpContext int
    var hexDumpMax int
    var hexDumpWidth int
    var noColor bool

    cmd := &cobra.Command{
        Use:   "diff",
        Short: "Diff two firmware images (raw and section-aware)",
        RunE: func(cmd *cobra.Command, args []string) error {
            if aPath == "" || bPath == "" {
                return fmt.Errorf("--a and --b are required")
            }
            // Silence logger for JSON output
            if jsonOut {
                logger = zap.NewNop()
            }
            var rep jsonReport

            if doRaw {
                a, err := readAll(aPath)
                if err != nil { return err }
                b, err := readAll(bPath)
                if err != nil { return err }
                rep.Raw.Enabled = true
                rep.Raw.SizeA = len(a)
                rep.Raw.SizeB = len(b)
                rep.Raw.SHA256A = sha256Hex(a)
                rep.Raw.SHA256B = sha256Hex(b)
                rep.Raw.Identical = rep.Raw.SizeA == rep.Raw.SizeB && rep.Raw.SHA256A == rep.Raw.SHA256B
                if !rep.Raw.Identical {
                    rep.Raw.Spans = diffRaw(a, b, maxSpans)
                }
                if !jsonOut {
                    fmt.Printf("== RAW BYTES ==\n")
                    fmt.Printf("A: size=%d sha256=%s\n", rep.Raw.SizeA, rep.Raw.SHA256A)
                    fmt.Printf("B: size=%d sha256=%s\n", rep.Raw.SizeB, rep.Raw.SHA256B)
                    if rep.Raw.Identical {
                        fmt.Println("IDENTICAL\n")
                    } else {
                        for i, sp := range rep.Raw.Spans {
                            if sp.Len >= 0 {
                                fmt.Printf("%2d: off=0x%08x len=0x%x A=%02x B=%02x\n", i, sp.Off, sp.Len, sp.Afirst, sp.Bfirst)
                                if hexDump {
                                    start := sp.Off - int64(hexDumpContext)
                                    if start < 0 { start = 0 }
                                    end := sp.Off + sp.Len + int64(hexDumpContext)
                                    maxCommon := int64(len(a))
                                    if int64(len(b)) < maxCommon { maxCommon = int64(len(b)) }
                                    if end > maxCommon { end = maxCommon }
                                    if end-start > int64(hexDumpMax) { end = start + int64(hexDumpMax) }
                                    fmt.Printf("-- hexdump raw span %d (0x%08x..0x%08x) --\n", i, start, end)
                                    diffutil.HexDumpSideBySide(a, b, 0, 0, int(start), int(end-start), hexDumpWidth, !noColor)
                                    fmt.Println()
                                }
                            }
                        }
                        fmt.Println()
                    }
                }
            }

            if doSections {
                if !jsonOut { fmt.Printf("== SECTIONS ==\n") }
                ctxA, err := cliutil.InitializeFirmwareParser(aPath, logger)
                if err != nil { return err }
                defer ctxA.Close()
                ctxB, err := cliutil.InitializeFirmwareParser(bPath, logger)
                if err != nil { return err }
                defer ctxB.Close()

                ga := cliutil.CollectSectionsByType(ctxA)
                gb := cliutil.CollectSectionsByType(ctxB)
                seen := map[uint16]struct{}{}
                for t := range ga { seen[t] = struct{}{} }
                for t := range gb { seen[t] = struct{}{} }
                allTypes := make([]uint16, 0, len(seen))
                for t := range seen { allTypes = append(allTypes, t) }
                sort.Slice(allTypes, func(i,j int) bool { return allTypes[i] < allTypes[j] })

                tfilters := cliutil.ParseFilterTypes(filterTypes)
                filterOK := func(name string, off uint64) bool {
                    if tfilters != nil {
                        if _, ok := tfilters[strings.ToUpper(name)]; !ok { return false }
                    }
                    if !cliutil.WithinRange(off, filterRange) { return false }
                    return true
                }

                missing, diffcnt := 0, 0
                items := make([]sectionDiffJSON, 0, 128)
                for _, t := range allTypes {
                    la := ga[t]
                    lb := gb[t]
                    maxn := len(la)
                    if len(lb) > maxn { maxn = len(lb) }
                    for i := 0; i < maxn; i++ {
                        var sa, sb cliutil.SectionInfo
                        var oka, okb bool
                        if i < len(la) { sa = la[i]; oka = true }
                        if i < len(lb) { sb = lb[i]; okb = true }
                        name := "UNKNOWN"
                        off := uint64(0)
                        if oka { name = sa.Name; off = sa.Offset } else if okb { name = sb.Name; off = sb.Offset }
                        if !filterOK(name, off) { continue }
                        if !oka || !okb {
                            if !jsonOut {
                                if !oka { fmt.Printf("MISSING in A: %-22s @0x%08x size=0x%x\n", sb.Name, sb.Offset, sb.Size) } else { fmt.Printf("MISSING in B: %-22s @0x%08x size=0x%x\n", sa.Name, sa.Offset, sa.Size) }
                            }
                            items = append(items, sectionDiffJSON{
                                Name: name, Type: t,
                                OffsetA: func() uint64 { if oka { return sa.Offset }; return 0 }(),
                                OffsetB: func() uint64 { if okb { return sb.Offset }; return 0 }(),
                                SizeA: func() uint32 { if oka { return sa.Size }; return 0 }(),
                                SizeB: func() uint32 { if okb { return sb.Size }; return 0 }(),
                                CRCA: func() string { if oka { return sa.CRCType.String() }; return "" }(),
                                CRCB: func() string { if okb { return sb.CRCType.String() }; return "" }(),
                                AlgoA: func() string { if oka { return types.GetSectionCRCAlgorithm(sa.Type).String() }; return "" }(),
                                AlgoB: func() string { if okb { return types.GetSectionCRCAlgorithm(sb.Type).String() }; return "" }(),
                                EncryptedA: oka && sa.Encrypted,
                                EncryptedB: okb && sb.Encrypted,
                                Identical: false,
                                FirstDiff: 0xFFFFFFFF,
                                MissingIn: func() string { if !oka { return "A" }; return "B" }(),
                            })
                            missing++
                            continue
                        }
                        // both present
                        ba, _ := cliutil.ReadSectionBytes(ctxA, sa)
                        bb, _ := cliutil.ReadSectionBytes(ctxB, sb)
                        same := len(ba) == len(bb)
                        if same {
                            for i := range ba { if ba[i] != bb[i] { same = false; break } }
                        }
                        fd := uint32(0xFFFFFFFF)
                        if !same {
                            max := len(ba); if len(bb) < max { max = len(bb) }
                            idx := -1
                            for i:=0;i<max;i++{ if ba[i]!=bb[i]{ idx=i; break } }
                            if idx >= 0 { fd = uint32(idx) }
                        }
                        if !jsonOut {
                            if same {
                                if sa.Offset == sb.Offset {
                                    fmt.Printf("OK     %-22s @0x%08x sizeA=0x%x sizeB=0x%x\n", sa.Name, sa.Offset, sa.Size, sb.Size)
                                } else {
                                    fmt.Printf("OK     %-22s @A:0x%08x B:0x%08x sizeA=0x%x sizeB=0x%x\n", sa.Name, sa.Offset, sb.Offset, sa.Size, sb.Size)
                                }
                            } else {
                                if sa.Offset == sb.Offset {
                                    fmt.Printf("DIFF   %-22s @0x%08x sizeA=0x%x sizeB=0x%x first+off=0x%x\n", sa.Name, sa.Offset, sa.Size, sb.Size, fd)
                                } else {
                                    fmt.Printf("DIFF   %-22s @A:0x%08x B:0x%08x sizeA=0x%x sizeB=0x%x first+off=0x%x\n", sa.Name, sa.Offset, sb.Offset, sa.Size, sb.Size, fd)
                                }
                                if hexDump {
                                    if strings.EqualFold(sa.Name, "DBG_FW_INI") {
                                        if aText, errA := diffutil.TryGunzip(ba); errA == nil {
                                            if bText, errB := diffutil.TryGunzip(bb); errB == nil {
                                                fmt.Printf("-- text diff DBG_FW_INI --\n")
                                                diffutil.PrintSideBySideTextDiff(aText, bText, !noColor, 400, 64)
                                                fmt.Println()
                                            }
                                        }
                                    } else {
                                        start := int64(0)
                                        if fd != 0xFFFFFFFF {
                                            start = int64(fd) - int64(hexDumpContext)
                                            if start < 0 { start = 0 }
                                        }
                                        maxLen := int64(hexDumpMax)
                                        if int64(len(ba)) < start+maxLen { maxLen = int64(len(ba)) - start }
                                        if int64(len(bb)) < start+maxLen { tmp := int64(len(bb)) - start; if tmp < maxLen { maxLen = tmp } }
                                        if maxLen < 0 { maxLen = 0 }
                                        fmt.Printf("-- hexdump %s (A:0x%08x B:0x%08x +0x%x) --\n", sa.Name, sa.Offset+uint64(start), sb.Offset+uint64(start), fd)
                                        diffutil.HexDumpSideBySide(ba, bb, int64(sa.Offset), int64(sb.Offset), int(start), int(maxLen), hexDumpWidth, !noColor)
                                        fmt.Println()
                                    }
                                }
                            }
                        }
                        items = append(items, sectionDiffJSON{
                            Name: sa.Name, Type: t,
                            OffsetA: sa.Offset, OffsetB: sb.Offset,
                            SizeA: sa.Size, SizeB: sb.Size,
                            CRCA: sa.CRCType.String(), CRCB: sb.CRCType.String(),
                            AlgoA: types.GetSectionCRCAlgorithm(sa.Type).String(),
                            AlgoB: types.GetSectionCRCAlgorithm(sb.Type).String(),
                            EncryptedA: sa.Encrypted, EncryptedB: sb.Encrypted,
                            Identical: same, FirstDiff: fd,
                        })
                        if !same { diffcnt++ }
                    }
                }
                rep.Sections.Enabled = true
                rep.Sections.Diffs = diffcnt
                rep.Sections.Missing = missing
                rep.Sections.Items = items
                if !jsonOut {
                    fmt.Printf("\nSummary: diffs=%d missing=%d\n", diffcnt, missing)
                }
            }

            if jsonOut {
                enc := json.NewEncoder(os.Stdout)
                enc.SetIndent("", "  ")
                return enc.Encode(rep)
            }
            return nil
        },
    }

    cmd.Flags().StringVar(&aPath, "a", "", "first firmware file (required)")
    cmd.Flags().StringVar(&bPath, "b", "", "second firmware file (required)")
    cmd.Flags().BoolVar(&doRaw, "raw", true, "perform raw byte diff")
    cmd.Flags().BoolVar(&doSections, "sections", true, "perform section-aware diff")
    cmd.Flags().IntVar(&maxSpans, "max-spans", 40, "maximum raw diff spans to show")
    cmd.Flags().BoolVar(&jsonOut, "json", false, "output machine-readable JSON")
    cmd.Flags().StringVar(&filterTypes, "filter-types", "", "comma-separated section type names to include")
    cmd.Flags().StringVar(&filterRange, "filter-offset", "", "limit sections by offset range start:end (hex or dec)")
    cmd.Flags().BoolVar(&hexDump, "hexdump", false, "print side-by-side hex dumps for diffs")
    cmd.Flags().IntVar(&hexDumpContext, "hexdump-context", 16, "bytes of context around diffs in hex dumps")
    cmd.Flags().IntVar(&hexDumpMax, "hexdump-max-bytes", 256, "maximum bytes to show per diff region in hex dumps")
    cmd.Flags().IntVar(&hexDumpWidth, "hexdump-width", 16, "bytes per row in hex dump")
    cmd.Flags().BoolVar(&noColor, "no-color", false, "disable ANSI colors in hex dumps")

    cmd.MarkFlagRequired("a")
    cmd.MarkFlagRequired("b")

    return cmd
}
