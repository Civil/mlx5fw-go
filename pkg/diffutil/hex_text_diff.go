package diffutil

import (
    "bytes"
    "compress/gzip"
    "fmt"
    "io"
    "strings"
)

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

// HexDumpSideBySide prints a side-by-side hex dump for a and b.
// baseA/baseB are absolute base offsets; start is relative index into both slices.
func HexDumpSideBySide(a, b []byte, baseA, baseB int64, start, length, width int, color bool) {
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
            ch := " "
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
            ch := " "
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

// TryGunzip attempts to gunzip the provided data and returns string content.
func TryGunzip(data []byte) (string, error) {
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

// PrintSideBySideTextDiff renders a side-by-side text diff using LCS alignment.
func PrintSideBySideTextDiff(aText, bText string, color bool, maxLines int, width int) {
    al := strings.Split(aText, "\n")
    bl := strings.Split(bText, "\n")
    ops := lcsLines(al, bl)
    if width <= 0 { width = 60 }
    printed := 0
    for _, op := range ops {
        if printed >= maxLines { fmt.Printf("... (truncated %d more lines)\n", len(ops)-printed); break }
        sa := op.a
        sb := op.b
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

