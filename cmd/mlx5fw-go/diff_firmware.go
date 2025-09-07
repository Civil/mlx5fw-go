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
    "github.com/Civil/mlx5fw-go/pkg/interfaces"
    "github.com/Civil/mlx5fw-go/pkg/types"
)

// Diff command that mirrors cmd/devtools/diff_firmware with type-based matching

func sha256Hex(b []byte) string {
    h := sha256.Sum256(b)
    return hex.EncodeToString(h[:])
}

func readAll(path string) ([]byte, error) {
    f, err := os.Open(path)
    if err != nil { return nil, err }
    defer f.Close()
    return os.ReadFile(path)
}

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

type sectKey struct {
    Offset uint64
    Type   uint16
}

type sectInfo struct {
    Key      sectKey
    Name     string
    Size     uint32
    CRCType  types.CRCType
    Enc      bool
    Reader   interfaces.CompleteSectionInterface
}

func collectSections(pctx *cliutil.ParserContext) map[sectKey]sectInfo {
    m := make(map[sectKey]sectInfo)
    for t, list := range pctx.Parser.GetSections() {
        for _, s := range list {
            key := sectKey{Offset: s.Offset(), Type: uint16(t)}
            m[key] = sectInfo{Key: key, Name: s.TypeName(), Size: s.Size(), CRCType: s.CRCType(), Enc: s.IsEncrypted(), Reader: s}
        }
    }
    return m
}

func readSectionBytes(ctx *cliutil.ParserContext, s sectInfo) ([]byte, error) {
    readSize := s.Size
    if s.CRCType == types.CRCInSection && !s.Enc {
        readSize += 4
    }
    return ctx.Parser.ReadSectionData(s.Key.Type, s.Key.Offset, readSize)
}

func parseFilterTypes(s string) map[string]struct{} {
    if s == "" { return nil }
    parts := strings.Split(s, ",")
    m := make(map[string]struct{}, len(parts))
    for _, p := range parts {
        n := strings.TrimSpace(strings.ToUpper(p))
        if n != "" { m[n] = struct{}{} }
    }
    return m
}

func withinRange(off uint64, rng string) bool {
    if rng == "" { return true }
    parts := strings.Split(rng, ":")
    if len(parts) != 2 { return true }
    parse := func(x string) (uint64, bool) {
        x = strings.TrimSpace(x)
        if x == "" { return 0, false }
        var v uint64
        var err error
        if strings.HasPrefix(x, "0x") || strings.HasPrefix(x, "0X") {
            _, err = fmt.Sscanf(x, "%x", &v)
        } else {
            _, err = fmt.Sscanf(x, "%d", &v)
        }
        if err != nil { return 0, false }
        return v, true
    }
    var startOk, endOk bool
    var start, end uint64
    if v, ok := parse(parts[0]); ok { start = v; startOk = true }
    if v, ok := parse(parts[1]); ok { end = v; endOk = true }
    if startOk && off < start { return false }
    if endOk && off >= end { return false }
    return true
}

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

                ma := collectSections(ctxA)
                mb := collectSections(ctxB)

                group := func(m map[sectKey]sectInfo) map[uint16][]sectInfo {
                    g := make(map[uint16][]sectInfo)
                    for k, v := range m { g[k.Type] = append(g[k.Type], v) }
                    for t := range g {
                        sort.Slice(g[t], func(i,j int) bool { return g[t][i].Key.Offset < g[t][j].Key.Offset })
                    }
                    return g
                }
                ga := group(ma)
                gb := group(mb)
                seen := map[uint16]struct{}{}
                for t := range ga { seen[t] = struct{}{} }
                for t := range gb { seen[t] = struct{}{} }
                allTypes := make([]uint16, 0, len(seen))
                for t := range seen { allTypes = append(allTypes, t) }
                sort.Slice(allTypes, func(i,j int) bool { return allTypes[i] < allTypes[j] })

                tfilters := parseFilterTypes(filterTypes)
                filterOK := func(name string, off uint64) bool {
                    if tfilters != nil {
                        if _, ok := tfilters[strings.ToUpper(name)]; !ok { return false }
                    }
                    if !withinRange(off, filterRange) { return false }
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
                        var sa, sb sectInfo
                        var oka, okb bool
                        if i < len(la) { sa = la[i]; oka = true }
                        if i < len(lb) { sb = lb[i]; okb = true }
                        name := "UNKNOWN"
                        off := uint64(0)
                        if oka { name = sa.Name; off = sa.Key.Offset } else if okb { name = sb.Name; off = sb.Key.Offset }
                        if !filterOK(name, off) { continue }
                        if !oka || !okb {
                            if !jsonOut {
                                if !oka { fmt.Printf("MISSING in A: %-22s @0x%08x size=0x%x\n", sb.Name, sb.Key.Offset, sb.Size) } else { fmt.Printf("MISSING in B: %-22s @0x%08x size=0x%x\n", sa.Name, sa.Key.Offset, sa.Size) }
                            }
                            items = append(items, sectionDiffJSON{
                                Name: name, Type: t,
                                OffsetA: func() uint64 { if oka { return sa.Key.Offset }; return 0 }(),
                                OffsetB: func() uint64 { if okb { return sb.Key.Offset }; return 0 }(),
                                SizeA: func() uint32 { if oka { return sa.Size }; return 0 }(),
                                SizeB: func() uint32 { if okb { return sb.Size }; return 0 }(),
                                CRCA: func() string { if oka { return sa.CRCType.String() }; return "" }(),
                                CRCB: func() string { if okb { return sb.CRCType.String() }; return "" }(),
                                AlgoA: func() string { if oka { return types.GetSectionCRCAlgorithm(sa.Key.Type).String() }; return "" }(),
                                AlgoB: func() string { if okb { return types.GetSectionCRCAlgorithm(sb.Key.Type).String() }; return "" }(),
                                EncryptedA: oka && sa.Enc,
                                EncryptedB: okb && sb.Enc,
                                Identical: false,
                                FirstDiff: 0xFFFFFFFF,
                                MissingIn: func() string { if !oka { return "A" }; return "B" }(),
                            })
                            missing++
                            continue
                        }
                        // both present
                        ba, _ := readSectionBytes(ctxA, sa)
                        bb, _ := readSectionBytes(ctxB, sb)
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
                                if sa.Key.Offset == sb.Key.Offset {
                                    fmt.Printf("OK     %-22s @0x%08x sizeA=0x%x sizeB=0x%x\n", sa.Name, sa.Key.Offset, sa.Size, sb.Size)
                                } else {
                                    fmt.Printf("OK     %-22s @A:0x%08x B:0x%08x sizeA=0x%x sizeB=0x%x\n", sa.Name, sa.Key.Offset, sb.Key.Offset, sa.Size, sb.Size)
                                }
                            } else {
                                if sa.Key.Offset == sb.Key.Offset {
                                    fmt.Printf("DIFF   %-22s @0x%08x sizeA=0x%x sizeB=0x%x first+off=0x%x\n", sa.Name, sa.Key.Offset, sa.Size, sb.Size, fd)
                                } else {
                                    fmt.Printf("DIFF   %-22s @A:0x%08x B:0x%08x sizeA=0x%x sizeB=0x%x first+off=0x%x\n", sa.Name, sa.Key.Offset, sb.Key.Offset, sa.Size, sb.Size, fd)
                                }
                            }
                        }
                        items = append(items, sectionDiffJSON{
                            Name: sa.Name, Type: t,
                            OffsetA: sa.Key.Offset, OffsetB: sb.Key.Offset,
                            SizeA: sa.Size, SizeB: sb.Size,
                            CRCA: sa.CRCType.String(), CRCB: sb.CRCType.String(),
                            AlgoA: types.GetSectionCRCAlgorithm(sa.Key.Type).String(),
                            AlgoB: types.GetSectionCRCAlgorithm(sb.Key.Type).String(),
                            EncryptedA: sa.Enc, EncryptedB: sb.Enc,
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

    cmd.MarkFlagRequired("a")
    cmd.MarkFlagRequired("b")

    return cmd
}
