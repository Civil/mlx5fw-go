package cliutil

import (
    "encoding/json"
    "io"
)

// NewIndentedEncoder returns a JSON encoder that writes to w with 2‑space indentation.
func NewIndentedEncoder(w io.Writer) *json.Encoder {
    enc := json.NewEncoder(w)
    enc.SetIndent("", "  ")
    return enc
}

// EncodeJSONIndent writes v as pretty JSON to w with 2‑space indentation.
func EncodeJSONIndent(w io.Writer, v any) error {
    return NewIndentedEncoder(w).Encode(v)
}

