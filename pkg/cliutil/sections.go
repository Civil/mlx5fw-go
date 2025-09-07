package cliutil

import (
    "sort"

    "github.com/Civil/mlx5fw-go/pkg/interfaces"
    "github.com/Civil/mlx5fw-go/pkg/types"
)

// SectionInfo captures common metadata used for comparisons and dumping.
type SectionInfo struct {
    Type      uint16
    Name      string
    Offset    uint64
    Size      uint32
    CRCType   types.CRCType
    Encrypted bool
    Reader    interfaces.CompleteSectionInterface
}

// CollectSectionsByType groups sections by type and sorts each slice by offset.
func CollectSectionsByType(pctx *ParserContext) map[uint16][]SectionInfo {
    g := make(map[uint16][]SectionInfo)
    for t, list := range pctx.Parser.GetSections() {
        for _, s := range list {
            info := SectionInfo{
                Type:      uint16(t),
                Name:      s.TypeName(),
                Offset:    s.Offset(),
                Size:      s.Size(),
                CRCType:   s.CRCType(),
                Encrypted: s.IsEncrypted(),
                Reader:    s,
            }
            g[uint16(t)] = append(g[uint16(t)], info)
        }
    }
    for t := range g {
        sort.Slice(g[t], func(i, j int) bool { return g[t][i].Offset < g[t][j].Offset })
    }
    return g
}

// ReadSectionBytes reads section payload plus trailer CRC if present and not encrypted.
func ReadSectionBytes(ctx *ParserContext, s SectionInfo) ([]byte, error) {
    readSize := s.Size
    if s.CRCType == types.CRCInSection && !s.Encrypted {
        readSize += 4
    }
    return ctx.Parser.ReadSectionData(s.Type, s.Offset, readSize)
}

