package main

import (
    "bytes"
    "compress/gzip"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "os"
    "sort"
    "strings"

    "go.uber.org/zap"

    cliutil "github.com/Civil/mlx5fw-go/pkg/cliutil"
    "github.com/Civil/mlx5fw-go/pkg/interfaces"
    "github.com/Civil/mlx5fw-go/pkg/types"
)

type rawDiffSpan struct {
    Off   int64
    Len   int64
    Afirst byte
    Bfirst byte
}

// JSON structures for --json output
type jsonRawReport struct {
    Enabled   bool          `json:"enabled"`
    SizeA     int           `json:"size_a"`
    SizeB     int           `json:"size_b"`
    SHA256A   string        `json:"sha256_a"`
    SHA256B   string        `json:"sha256_b"`
    Identical bool          `json:"identical"`
    Spans     []rawDiffSpan `json:"spans,omitempty"`
}

func sha256Hex(b []byte) string {
    h := sha256.Sum256(b)
    return hex.EncodeToString(h[:])
}

func readAll(path string) ([]byte, error) {
    f, err := os.Open(path)
    if err != nil { return nil, err }
    defer f.Close()
    return io.ReadAll(f)
}

func diffRaw(a, b []byte, maxSpans int) []rawDiffSpan {
    var spans []rawDiffSpan
    la, lb := int64(len(a)), int64(len(b))
    i := int64(0)
    max := la
    if lb < max { max = lb }
    for i < max && len(spans) < maxSpans {
        if a[i] == b[i] { i++; continue }
        // start of span
        start := i
        afirst, bfirst := a[i], b[i]
        for i < max && a[i] != b[i] { i++ }
        spans = append(spans, rawDiffSpan{Off: start, Len: i - start, Afirst: afirst, Bfirst: bfirst})
    }
    // tail size mismatch span
    if len(spans) < maxSpans {
        if la != lb {
            spans = append(spans, rawDiffSpan{Off: max, Len: (la - lb)})
        }
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
    MissingIn string `json:"missing_in,omitempty"` // "A" or "B"
}

type jsonSectionsReport struct {
    Enabled bool               `json:"enabled"`
    Diffs   int                `json:"diff_count"`
    Missing int                `json:"missing_count"`
    Items   []sectionDiffJSON  `json:"items"`
}

type jsonReport struct {
    Raw      jsonRawReport      `json:"raw"`
    Sections jsonSectionsReport `json:"sections"`
}

// ===== Hex dump helpers =====
const (
    ansiReset = "\x1b[0m"
    ansiRed   = "\x1b[31m"
    ansiCyan  = "\x1b[36m"
)

func isPrintable(b byte) bool { return b >= 0x20 && b <= 0x7e }

func maybeColor(s string, colorCode string, enable bool) string {
    if !enable { return s }
    return colorCode + s + ansiReset
}

// hexDumpSideBySide prints a side-by-side hex dump for a and b.
// baseA/baseB are absolute base offsets; start is relative index into both slices.
func hexDumpSideBySide(a, b []byte, baseA, baseB int64, start, length, width int, color bool) {
    end := start + length
    if start < 0 { start = 0 }
    if end < start { return }
    for row := start; row < end; row += width {
        rEnd := row + width
        if rEnd > end { rEnd = end }
        // address columns
        addrA := baseA + int64(row)
        addrB := baseB + int64(row)
        fmt.Printf("%s %08x%s  ", maybeColor("A:", ansiCyan, color), uint32(addrA), maybeColor("", ansiCyan, false))

        // bytes A
        for i := row; i < row+width; i++ {
            if i < rEnd && i < len(a) {
                diff := i < len(b) && a[i] != b[i]
                hx := fmt.Sprintf("%02x", a[i])
                if diff { hx = maybeColor(hx, ansiRed, color) }
                fmt.Printf("%s", hx)
            } else {
                fmt.Printf("  ")
            }
            if (i-row)%8 == 7 { fmt.Printf(" ") }
            fmt.Printf(" ")
        }
        // ascii A
        fmt.Printf(" ")
        for i := row; i < row+width; i++ {
            var ch string = " "
            if i < rEnd && i < len(a) {
                c := a[i]
                if isPrintable(c) { ch = string(rune(c)) } else { ch = "." }
                if i < len(b) && a[i] != b[i] { ch = maybeColor(ch, ansiRed, color) }
            }
            fmt.Printf("%s", ch)
        }

        // separator
        fmt.Printf("  |  ")

        // address B
        fmt.Printf("%s %08x%s  ", maybeColor("B:", ansiCyan, color), uint32(addrB), maybeColor("", ansiCyan, false))
        // bytes B
        for i := row; i < row+width; i++ {
            if i < rEnd && i < len(b) {
                diff := i < len(a) && a[i] != b[i]
                hx := fmt.Sprintf("%02x", b[i])
                if diff { hx = maybeColor(hx, ansiRed, color) }
                fmt.Printf("%s", hx)
            } else {
                fmt.Printf("  ")
            }
            if (i-row)%8 == 7 { fmt.Printf(" ") }
            fmt.Printf(" ")
        }
        // ascii B
        fmt.Printf(" ")
        for i := row; i < row+width; i++ {
            var ch string = " "
            if i < rEnd && i < len(b) {
                c := b[i]
                if isPrintable(c) { ch = string(rune(c)) } else { ch = "." }
                if i < len(a) && a[i] != b[i] { ch = maybeColor(ch, ansiRed, color) }
            }
            fmt.Printf("%s", ch)
        }

        fmt.Println()
    }
}

// ===== Text diff for DBG_FW_INI =====
func tryGunzip(data []byte) (string, error) {
    r, err := gzip.NewReader(bytes.NewReader(data))
    if err != nil { return "", err }
    defer r.Close()
    out, err := io.ReadAll(r)
    if err != nil { return "", err }
    return string(out), nil
}

// simple LCS for lines
type opKind int
const (
    opEq opKind = iota
    opDel
    opIns
)
type diffOp struct { kind opKind; a, b string }

func lcsLines(a, b []string) []diffOp {
    n, m := len(a), len(b)
    dp := make([][]int, n+1)
    for i := range dp { dp[i] = make([]int, m+1) }
    for i:=n-1;i>=0;i--{
        for j:=m-1;j>=0;j--{
            if a[i]==b[j] { dp[i][j]=dp[i+1][j+1]+1 } else if dp[i+1][j]>=dp[i][j+1] { dp[i][j]=dp[i+1][j] } else { dp[i][j]=dp[i][j+1] }
        }
    }
    i,j:=0,0
    ops := []diffOp{}
    for i<n && j<m {
        if a[i]==b[j] { ops=append(ops,diffOp{opEq,a[i],b[j]}); i++; j++; } else if dp[i+1][j]>=dp[i][j+1] { ops=append(ops,diffOp{opDel,a[i],""}); i++; } else { ops=append(ops,diffOp{opIns,"",b[j]}); j++; }
    }
    for ; i<n; i++ { ops=append(ops,diffOp{opDel,a[i],""}) }
    for ; j<m; j++ { ops=append(ops,diffOp{opIns,"",b[j]}) }
    return ops
}

func printSideBySideTextDiff(aText, bText string, color bool, maxLines int, width int) {
    al := strings.Split(aText, "\n")
    bl := strings.Split(bText, "\n")
    ops := lcsLines(al, bl)
    if width <= 0 { width = 60 }
    printed := 0
    for _, op := range ops {
        if printed >= maxLines { fmt.Printf("... (truncated %d more lines)\n", len(ops)-printed); break }
        la := op.a
        lb := op.b
        // pad/truncate
        sa := la
        sb := lb
        if len(sa) > width { sa = sa[:width] }
        if len(sb) > width { sb = sb[:width] }
        left := fmt.Sprintf("%-*s", width, sa)
        right := fmt.Sprintf("%-*s", width, sb)
        switch op.kind {
        case opEq:
            fmt.Printf(" %s  |  %s\n", left, right)
        case opDel:
            if color { left = ansiRed + left + ansiReset }
            fmt.Printf("-%s  |  %s\n", left, right)
        case opIns:
            if color { right = ansiRed + right + ansiReset }
            fmt.Printf(" %s  | +%s\n", left, right)
        }
        printed++
    }
}

func collectSections(pctx *cliutil.ParserContext) map[sectKey]sectInfo {
    m := make(map[sectKey]sectInfo)
    for t, list := range pctx.Parser.GetSections() {
        for _, s := range list {
            key := sectKey{Offset: s.Offset(), Type: uint16(t)}
            m[key] = sectInfo{
                Key:     key,
                Name:    s.TypeName(),
                Size:    s.Size(),
                CRCType: s.CRCType(),
                Enc:     s.IsEncrypted(),
                Reader:  s,
            }
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

// parseFilterTypes converts comma-separated section names to a set
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
    // Accept formats: "start:end" where each may be hex (0x...) or decimal; empty start or end are open ranges
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

func main() {
    aPath := flag.String("a", "", "first firmware file")
    bPath := flag.String("b", "", "second firmware file")
    doRaw := flag.Bool("raw", true, "perform raw byte diff")
    doSections := flag.Bool("sections", true, "perform section-aware diff")
    maxSpans := flag.Int("max-spans", 40, "maximum raw diff spans to show")
    jsonOut := flag.Bool("json", false, "output machine-readable JSON")
    filterTypes := flag.String("filter-types", "", "comma-separated section type names to include (optional)")
    filterRange := flag.String("filter-offset", "", "limit sections by offset range start:end (hex or dec)")
    hexDump := flag.Bool("hexdump", false, "print side-by-side hex dumps for diffs")
    hexDumpContext := flag.Int("hexdump-context", 16, "bytes of context around diffs in hex dumps")
    hexDumpMax := flag.Int("hexdump-max-bytes", 256, "maximum bytes to show per diff region in hex dumps")
    hexDumpWidth := flag.Int("hexdump-width", 16, "bytes per row in hex dump")
    noColor := flag.Bool("no-color", false, "disable ANSI colors in hex dumps")
    flag.Parse()

    if *aPath == "" || *bPath == "" {
        fmt.Fprintf(os.Stderr, "usage: %s -a <fwA.bin> -b <fwB.bin> [--raw] [--sections]\n", os.Args[0])
        os.Exit(2)
    }

    var logger *zap.Logger
    if *jsonOut {
        logger = zap.NewNop() // keep JSON clean
    } else {
        logger, _ = zap.NewDevelopment()
    }
    defer logger.Sync()

    // JSON report accumulator
    rep := jsonReport{}

    // Raw diff
    if *doRaw {
        a, err := readAll(*aPath)
        if err != nil { fmt.Fprintf(os.Stderr, "read %s: %v\n", *aPath, err); os.Exit(1) }
        b, err := readAll(*bPath)
        if err != nil { fmt.Fprintf(os.Stderr, "read %s: %v\n", *bPath, err); os.Exit(1) }
        rep.Raw.Enabled = true
        rep.Raw.SizeA = len(a)
        rep.Raw.SizeB = len(b)
        rep.Raw.SHA256A = sha256Hex(a)
        rep.Raw.SHA256B = sha256Hex(b)
        rep.Raw.Identical = (len(a) == len(b) && rep.Raw.SHA256A == rep.Raw.SHA256B)
        if !rep.Raw.Identical {
            rep.Raw.Spans = diffRaw(a, b, *maxSpans)
        }
        if !*jsonOut {
            fmt.Printf("== RAW BYTES ==\n")
            fmt.Printf("A: size=%d sha256=%s\n", rep.Raw.SizeA, rep.Raw.SHA256A)
            fmt.Printf("B: size=%d sha256=%s\n", rep.Raw.SizeB, rep.Raw.SHA256B)
            if rep.Raw.Identical {
                fmt.Printf("IDENTICAL\n\n")
            } else if len(rep.Raw.Spans) == 0 {
                fmt.Printf("No differing spans within common length; sizes differ by %d bytes\n\n", rep.Raw.SizeA-rep.Raw.SizeB)
            } else {
                for i, sp := range rep.Raw.Spans {
                    if sp.Len >= 0 {
                        fmt.Printf("%2d: off=0x%08x len=0x%x A=%02x B=%02x\n", i, sp.Off, sp.Len, sp.Afirst, sp.Bfirst)
                        if *hexDump {
                            start := sp.Off - int64(*hexDumpContext)
                            if start < 0 { start = 0 }
                            end := sp.Off + sp.Len + int64(*hexDumpContext)
                            maxCommon := int64(len(a))
                            if int64(len(b)) < maxCommon { maxCommon = int64(len(b)) }
                            if end > maxCommon { end = maxCommon }
                            if end-start > int64(*hexDumpMax) { end = start + int64(*hexDumpMax) }
                            fmt.Printf("-- hexdump raw span %d (0x%08x..0x%08x) --\n", i, start, end)
                            hexDumpSideBySide(a, b, 0, 0, int(start), int(end-start), *hexDumpWidth, !*noColor)
                            fmt.Println()
                        }
                    } else {
                        fmt.Printf("%2d: tail size delta at 0x%08x: %d bytes\n", i, sp.Off, sp.Len)
                    }
                }
                fmt.Println()
            }
        }
    }

    // Section-aware diff
    if *doSections {
        if !*jsonOut { fmt.Printf("== SECTIONS ==\n") }
        ctxA, err := cliutil.InitializeFirmwareParser(*aPath, logger)
        if err != nil { fmt.Fprintf(os.Stderr, "init A: %v\n", err); os.Exit(1) }
        defer ctxA.Close()
        ctxB, err := cliutil.InitializeFirmwareParser(*bPath, logger)
        if err != nil { fmt.Fprintf(os.Stderr, "init B: %v\n", err); os.Exit(1) }
        defer ctxB.Close()

        ma := collectSections(ctxA)
        mb := collectSections(ctxB)

        // Build groups by Type (align by type index, not absolute address)
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

        // Union of all types
        allTypes := make([]uint16, 0, len(ga)+len(gb))
        seen := map[uint16]struct{}{}
        for t := range ga { seen[t] = struct{}{} }
        for t := range gb { seen[t] = struct{}{} }
        for t := range seen { allTypes = append(allTypes, t) }
        sort.Slice(allTypes, func(i,j int) bool { return allTypes[i] < allTypes[j] })

        typeFilters := parseFilterTypes(*filterTypes)
        filterOK := func(name string, off uint64) bool {
            if typeFilters != nil {
                if _, ok := typeFilters[strings.ToUpper(name)]; !ok { return false }
            }
            if !withinRange(off, *filterRange) { return false }
            return true
        }

        missing := 0
        diffcnt := 0
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
                offForFilter := uint64(0)
                if oka { name = sa.Name; offForFilter = sa.Key.Offset } else if okb { name = sb.Name; offForFilter = sb.Key.Offset }
                if !filterOK(name, offForFilter) { continue }

                if !oka || !okb {
                    // Missing on one side
                    if !*jsonOut {
                        if !oka {
                            fmt.Printf("MISSING in A: %-22s @0x%08x size=0x%x\n", sb.Name, sb.Key.Offset, sb.Size)
                        } else {
                            fmt.Printf("MISSING in B: %-22s @0x%08x size=0x%x\n", sa.Name, sa.Key.Offset, sa.Size)
                        }
                    }
                    items = append(items, sectionDiffJSON{
                        Name: name,
                        Type: t,
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

                // Both present, compare bytes
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
                if !*jsonOut {
                    // Note: Offsets may differ; show both when different
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
                        if *hexDump {
                            // Special-case DBG_FW_INI: gunzip and text diff if possible
                            if strings.EqualFold(sa.Name, "DBG_FW_INI") {
                                if aText, errA := tryGunzip(ba); errA == nil {
                                    if bText, errB := tryGunzip(bb); errB == nil {
                                        fmt.Printf("-- text diff DBG_FW_INI --\n")
                                        printSideBySideTextDiff(aText, bText, !*noColor, 400, 64)
                                        fmt.Println()
                                    }
                                }
                            } else {
                                // Show hexdump around the first diff
                                start := int64(0)
                                if fd != 0xFFFFFFFF {
                                    start = int64(fd) - int64(*hexDumpContext)
                                    if start < 0 { start = 0 }
                                }
                                // Limit by available data and flag
                                maxLen := int64(*hexDumpMax)
                                if int64(len(ba)) < start+maxLen { maxLen = int64(len(ba)) - start }
                                if int64(len(bb)) < start+maxLen { tmp := int64(len(bb)) - start; if tmp < maxLen { maxLen = tmp } }
                                if maxLen < 0 { maxLen = 0 }
                                fmt.Printf("-- hexdump %s (A:0x%08x B:0x%08x +0x%x) --\n", sa.Name, sa.Key.Offset+uint64(start), sb.Key.Offset+uint64(start), fd)
                                hexDumpSideBySide(ba, bb, int64(sa.Key.Offset), int64(sb.Key.Offset), int(start), int(maxLen), *hexDumpWidth, !*noColor)
                                fmt.Println()
                            }
                        }
                    }
                }
                items = append(items, sectionDiffJSON{
                    Name: sa.Name,
                    Type: t,
                    OffsetA: sa.Key.Offset,
                    OffsetB: sb.Key.Offset,
                    SizeA: sa.Size,
                    SizeB: sb.Size,
                    CRCA: sa.CRCType.String(),
                    CRCB: sb.CRCType.String(),
                    AlgoA: types.GetSectionCRCAlgorithm(sa.Key.Type).String(),
                    AlgoB: types.GetSectionCRCAlgorithm(sb.Key.Type).String(),
                    EncryptedA: sa.Enc,
                    EncryptedB: sb.Enc,
                    Identical: same,
                    FirstDiff: fd,
                })
                if !same { diffcnt++ }
            }
        }
        rep.Sections.Enabled = true
        rep.Sections.Diffs = diffcnt
        rep.Sections.Missing = missing
        rep.Sections.Items = items
        if !*jsonOut {
            fmt.Printf("\nSummary: diffs=%d missing=%d\n", diffcnt, missing)
        }
    }

    if *jsonOut {
        enc := json.NewEncoder(os.Stdout)
        enc.SetIndent("", "  ")
        if err := enc.Encode(rep); err != nil {
            fmt.Fprintf(os.Stderr, "encode json: %v\n", err)
            os.Exit(1)
        }
    }
}
