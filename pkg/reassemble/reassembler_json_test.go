package reassemble

import (
    "encoding/base64"
    "encoding/hex"
    "testing"

    "go.uber.org/zap"

    "github.com/Civil/mlx5fw-go/pkg/interfaces"
    "github.com/Civil/mlx5fw-go/pkg/types"
    "github.com/Civil/mlx5fw-go/pkg/types/extracted"
)

func newReassemblerForTest() *Reassembler {
    return New(zap.NewNop(), Options{})
}

func mkMeta(t uint16, size uint32) extracted.SectionMetadata {
    return extracted.SectionMetadata{
        BaseSection: &interfaces.BaseSection{
            SectionType:    types.SectionType(t),
            SectionOffset:  0,
            SectionSize:    size,
            SectionCRCType: types.CRCInSection,
        },
        OriginalSize: size, // not used by reconstruction path but set for completeness
    }
}

func TestReconstruct_ImageSignature256_PadsToSectionSize(t *testing.T) {
    r := newReassemblerForTest()
    // Create 256-byte signature of zeros
    sig := make([]byte, 256)
    json := []byte(`{"image_signature": {"signature_type": 1, "signature": "` + base64.StdEncoding.EncodeToString(sig) + `"}}`)
    meta := mkMeta(types.SectionTypeImageSignature256, 320)
    out, err := r.reconstructFromJSONByType(json, meta)
    if err != nil {
        t.Fatalf("reconstruct failed: %v", err)
    }
    if len(out) != int(meta.Size()) {
        t.Fatalf("size mismatch: got %d want %d", len(out), meta.Size())
    }
}

func TestReconstruct_ImageSignature512_PadsToSectionSize(t *testing.T) {
    r := newReassemblerForTest()
    sig := make([]byte, 512)
    json := []byte(`{"image_signature": {"signature_type": 1, "signature": "` + base64.StdEncoding.EncodeToString(sig) + `"}}`)
    meta := mkMeta(types.SectionTypeImageSignature512, 576)
    out, err := r.reconstructFromJSONByType(json, meta)
    if err != nil {
        t.Fatalf("reconstruct failed: %v", err)
    }
    if len(out) != int(meta.Size()) {
        t.Fatalf("size mismatch: got %d want %d", len(out), meta.Size())
    }
}

func TestReconstruct_PublicKeys2048_PadsToSectionSize(t *testing.T) {
    r := newReassemblerForTest()
    key := make([]byte, 256)
    uuid := make([]byte, 16)
    uuidHex := hex.EncodeToString(uuid)
    keyB64 := base64.StdEncoding.EncodeToString(key)
    // Build 8 identical keys JSON
    keys := `[{"reserved":0,"uuid":"` + uuidHex + `","key":"` + keyB64 + `"},` +
        `{"reserved":0,"uuid":"` + uuidHex + `","key":"` + keyB64 + `"},` +
        `{"reserved":0,"uuid":"` + uuidHex + `","key":"` + keyB64 + `"},` +
        `{"reserved":0,"uuid":"` + uuidHex + `","key":"` + keyB64 + `"},` +
        `{"reserved":0,"uuid":"` + uuidHex + `","key":"` + keyB64 + `"},` +
        `{"reserved":0,"uuid":"` + uuidHex + `","key":"` + keyB64 + `"},` +
        `{"reserved":0,"uuid":"` + uuidHex + `","key":"` + keyB64 + `"},` +
        `{"reserved":0,"uuid":"` + uuidHex + `","key":"` + keyB64 + `"}]`
    json := []byte(`{"public_keys": {"keys": ` + keys + `}}`)
    meta := mkMeta(types.SectionTypePublicKeys2048, 2304)
    out, err := r.reconstructFromJSONByType(json, meta)
    if err != nil {
        t.Fatalf("reconstruct failed: %v", err)
    }
    if len(out) != int(meta.Size()) {
        t.Fatalf("size mismatch: got %d want %d", len(out), meta.Size())
    }
}

func TestReconstruct_PublicKeys4096_PadsToSectionSize(t *testing.T) {
    r := newReassemblerForTest()
    key := make([]byte, 512)
    uuid := make([]byte, 16)
    uuidHex := hex.EncodeToString(uuid)
    keyB64 := base64.StdEncoding.EncodeToString(key)
    keys := `[{"reserved":0,"uuid":"` + uuidHex + `","key":"` + keyB64 + `"},` +
        `{"reserved":0,"uuid":"` + uuidHex + `","key":"` + keyB64 + `"},` +
        `{"reserved":0,"uuid":"` + uuidHex + `","key":"` + keyB64 + `"},` +
        `{"reserved":0,"uuid":"` + uuidHex + `","key":"` + keyB64 + `"},` +
        `{"reserved":0,"uuid":"` + uuidHex + `","key":"` + keyB64 + `"},` +
        `{"reserved":0,"uuid":"` + uuidHex + `","key":"` + keyB64 + `"},` +
        `{"reserved":0,"uuid":"` + uuidHex + `","key":"` + keyB64 + `"},` +
        `{"reserved":0,"uuid":"` + uuidHex + `","key":"` + keyB64 + `"}]`
    json := []byte(`{"public_keys": {"keys": ` + keys + `}}`)
    meta := mkMeta(types.SectionTypePublicKeys4096, 4352)
    out, err := r.reconstructFromJSONByType(json, meta)
    if err != nil {
        t.Fatalf("reconstruct failed: %v", err)
    }
    if len(out) != int(meta.Size()) {
        t.Fatalf("size mismatch: got %d want %d", len(out), meta.Size())
    }
}

func TestReconstruct_MfgInfo_PadsTo256And320(t *testing.T) {
    r := newReassemblerForTest()
    // Minimal MFG_INFO JSON with PSID and zeroed fields
    mfgJSON := []byte(`{"mfg_info": {"psid":"MT_0000000000","reserved1":"","flags":1,"guids":{"start":0,"count":0},"macs":{"start":0,"count":0},"reserved2":""}}`)

    // Case 1: 256 bytes
    meta256 := mkMeta(types.SectionTypeMfgInfo, 256)
    out256, err := r.reconstructFromJSONByType(mfgJSON, meta256)
    if err != nil { t.Fatalf("mfg 256 reconstruct failed: %v", err) }
    if len(out256) != int(meta256.Size()) {
        t.Fatalf("mfg 256 size mismatch: got %d want %d", len(out256), meta256.Size())
    }

    // Case 2: 320 bytes
    meta320 := mkMeta(types.SectionTypeMfgInfo, 320)
    out320, err := r.reconstructFromJSONByType(mfgJSON, meta320)
    if err != nil { t.Fatalf("mfg 320 reconstruct failed: %v", err) }
    if len(out320) != int(meta320.Size()) {
        t.Fatalf("mfg 320 size mismatch: got %d want %d", len(out320), meta320.Size())
    }
}

func TestReconstruct_HashesTable_ReservedTail(t *testing.T) {
    r := newReassemblerForTest()
    // Build one entry with 32-byte zero hash
    entry := `{"type":1,"offset":4096,"size":64,"reserved":0,"hash":"` + hex.EncodeToString(make([]byte,32)) + `"}`
    // Header values (magic/version arbitrary for test)
    header := `{"magic":305419896,"version":1,"reserved1":0,"reserved2":0,"table_size":124,"num_entries":1,"reserved3":0,"crc":0,"reserved4":0}`
    // Reserved tail of 28 bytes (56 hex chars)
    tail := hex.EncodeToString(make([]byte, 28))
    jsonData := []byte(`{"header": ` + header + `, "entries": [` + entry + `], "reserved_tail": "` + tail + `"}`)
    // Section size = 32 (hdr) + 64 (entry) + 28 (tail) = 124
    meta := mkMeta(types.SectionTypeHashesTable, 124)
    out, err := r.reconstructFromJSONByType(jsonData, meta)
    if err != nil { t.Fatalf("hashes_table reconstruct failed: %v", err) }
    if len(out) != int(meta.Size()) {
        t.Fatalf("hashes_table size mismatch: got %d want %d", len(out), meta.Size())
    }
    // Verify last 28 bytes are equal to tail (zeros)
    for i := len(out)-28; i < len(out); i++ {
        if out[i] != 0x00 { t.Fatalf("reserved tail byte at %d not zero", i) }
    }
}
