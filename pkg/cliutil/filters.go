package cliutil

import (
    "fmt"
    "strings"
)

// ParseFilterTypes converts comma-separated section names to a set (uppercased)
func ParseFilterTypes(s string) map[string]struct{} {
    if s == "" { return nil }
    parts := strings.Split(s, ",")
    m := make(map[string]struct{}, len(parts))
    for _, p := range parts {
        n := strings.TrimSpace(strings.ToUpper(p))
        if n != "" { m[n] = struct{}{} }
    }
    return m
}

// WithinRange checks whether an offset is inside an optional start:end range where each bound can be hex (0x...) or decimal.
// Empty start or end act as open bounds; malformed inputs fall back to allowing everything.
func WithinRange(off uint64, rng string) bool {
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

