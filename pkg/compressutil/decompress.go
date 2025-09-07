package compressutil

import (
    "bytes"
    "compress/gzip"
    "compress/zlib"
    "io"
)

// TryGunzip attempts to decompress gzip data and returns it as string.
func TryGunzip(data []byte) (string, error) {
    r, err := gzip.NewReader(bytes.NewReader(data))
    if err != nil { return "", err }
    defer r.Close()
    out, err := io.ReadAll(r)
    if err != nil { return "", err }
    return string(out), nil
}

// DecompressZlib inflates zlib-compressed data.
func DecompressZlib(data []byte) ([]byte, error) {
    r, err := zlib.NewReader(bytes.NewReader(data))
    if err != nil { return nil, err }
    defer r.Close()
    out, err := io.ReadAll(r)
    if err != nil { return nil, err }
    return out, nil
}

