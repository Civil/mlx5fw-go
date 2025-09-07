package bytesutil

import (
    "crypto/sha256"
    "encoding/hex"
)

// RawDiffSpan describes a differing span between two byte slices.
// If Len < 0, it indicates a tail size delta (a.Size - b.Size) at Off.
type RawDiffSpan struct {
    Off    int64  `json:"off"`
    Len    int64  `json:"len"`
    Afirst byte   `json:"a_first"`
    Bfirst byte   `json:"b_first"`
}

// DiffRaw returns up to maxSpans differing spans between a and b.
func DiffRaw(a, b []byte, maxSpans int) []RawDiffSpan {
    var spans []RawDiffSpan
    la, lb := int64(len(a)), int64(len(b))
    i := int64(0)
    max := la
    if lb < max { max = lb }
    for i < max && len(spans) < maxSpans {
        if a[i] == b[i] { i++; continue }
        start := i
        afirst, bfirst := a[i], b[i]
        for i < max && a[i] != b[i] { i++ }
        spans = append(spans, RawDiffSpan{Off: start, Len: i - start, Afirst: afirst, Bfirst: bfirst})
    }
    if len(spans) < maxSpans && la != lb {
        spans = append(spans, RawDiffSpan{Off: max, Len: (la - lb)})
    }
    return spans
}

// SHA256Hex computes SHA-256 and returns lowercase hex string.
func SHA256Hex(b []byte) string {
    h := sha256.Sum256(b)
    return hex.EncodeToString(h[:])
}

