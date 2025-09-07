package fs4

import (
    "encoding/binary"
    "fmt"

    "github.com/ansel1/merry/v2"
    "go.uber.org/zap"

    "github.com/Civil/mlx5fw-go/pkg/interfaces"
    "github.com/Civil/mlx5fw-go/pkg/types"
)

// parseFS3 implements a minimal FS3 (CIB) parser that discovers the ITOC and builds section list
func (p *Parser) parseFS3() error {
    p.logger.Info("Parsing FS3 firmware")

    // Heuristic: FS3 images have ITOC aligned to 4KB sectors, typically at 0x2000
    const sectorSize = 0x1000
    fileSize := int(p.reader.Size())

    found := false
    var itocOffset int

    // Scan sectors starting from 0x1000 up to end-0x1000
    for off := sectorSize; off+types.ITOCHeaderSize <= fileSize; off += sectorSize {
        headerData, err := p.reader.ReadSection(int64(off), types.ITOCHeaderSize)
        if err != nil {
            return merry.Wrap(err)
        }
        // Check for ITOC signature pattern: 0x49544F43, 0x04081516, 0x2342CAFA, 0xBACAFE00
        sig0 := binary.BigEndian.Uint32(headerData[0:4])
        sig1 := binary.BigEndian.Uint32(headerData[4:8])
        sig2 := binary.BigEndian.Uint32(headerData[8:12])
        sig3 := binary.BigEndian.Uint32(headerData[12:16])
        if sig0 == types.ITOCSignature && sig1 == 0x04081516 && sig2 == 0x2342cafa && sig3 == 0xbacafe00 {
            itocOffset = off
            found = true
            break
        }
    }

    if !found {
        return merry.New("no valid FS3 ITOC found")
    }

    p.itocAddr = uint32(itocOffset)
    p.itocHeaderValid = true // Treat as valid once signature matched
    p.isEncrypted = false
    p.format = types.FormatFS3

    p.logger.Debug("Found FS3 ITOC", zap.Uint32("address", p.itocAddr), zap.String("address_hex", fmt.Sprintf("0x%x", p.itocAddr)))

    // Parse ITOC entries
    // Entries follow the header; iterate until type == 0xFF (END)
    for idx := 0; ; idx++ {
        entryOff := itocOffset + types.ITOCHeaderSize + idx*types.ITOCEntrySize
        if entryOff+types.ITOCEntrySize > fileSize {
            break
        }
        entryData, err := p.reader.ReadSection(int64(entryOff), types.ITOCEntrySize)
        if err != nil {
            return merry.Wrap(err)
        }

        var e types.FS3ITOCEntry
        if err := e.Unmarshal(entryData); err != nil {
            // Stop on parse error
            return merry.Wrap(err)
        }

        if e.Type == 0xFF { // END
            break
        }

        // Compute address/size in bytes
        addr := uint64(e.FlashAddrDwords << 2)
        size := uint32(e.SizeDwords << 2)

        // Determine CRC handling for FS3: either IN_ITOC_ENTRY or NONE
        crcType := types.CRCInITOCEntry
        if e.NoCRC {
            crcType = types.CRCNone
        }

        // Expected CRC from entry (16-bit)
        expCRC := uint32(e.SectionCRC)

        // Create a synthetic FS4-style ITOC entry with the basics filled for downstream consumers
        fs4Entry := &types.ITOCEntry{}
        // Fill fields used by factory/consumers
        // Type
        fs4Entry.SetType(uint8(e.Type))
        // Size in bytes
        fs4Entry.SetSize(size)
        // Flash address (bytes)
        fs4Entry.SetFlashAddr(uint32(addr))
        // Section CRC (16-bit)
        fs4Entry.SetSectionCRC(uint16(expCRC))
        // CRC flag encoding: CRCInITOCEntry=0, NONE=1
        if crcType == types.CRCNone {
            fs4Entry.SetParam0(1) // not used, but avoid zeroing other fields
        }

        var section interfaces.CompleteSectionInterface
        var serr error
        if size == 0 {
            // Zero-length sections (e.g., VPD_R0) should still be listed
            section, serr = p.sectionFactory.CreateSection(uint16(e.Type), addr, size, crcType, expCRC, false /*isEncrypted*/, e.DeviceData, fs4Entry, false /*fromHWPointer*/)
        } else {
            // Read section data (without embedded CRC; FS3 uses ITOC CRC)
            data, derr := p.reader.ReadSection(int64(addr), size)
            if derr != nil {
                p.logger.Warn("Failed to read FS3 section", zap.Error(derr), zap.Uint8("type", e.Type), zap.Uint64("addr", addr), zap.Uint32("size", size))
                continue
            }
            // Build section via factory
            section, serr = p.sectionFactory.CreateSectionFromData(uint16(e.Type), addr, size, crcType, expCRC, false /*isEncrypted*/, e.DeviceData, fs4Entry, false /*fromHWPointer*/, data)
        }
        if serr != nil {
            p.logger.Warn("Failed to create FS3 section", zap.Error(serr), zap.Uint8("type", e.Type))
            continue
        }

        p.addSection(section)
    }

    // Try to add BOOT2 section heuristically (optional): FS3 boot header at 0x38 contains size at [4..8]
    // This mirrors FS4 parser's display expectations but is best-effort.
    if p.reader.Size() >= 0x40 {
        hdr, err := p.reader.ReadSection(0x38, 16)
        if err == nil {
            sz := binary.BigEndian.Uint32(hdr[4:8])
            // FS3 boot2 size appears to be measured in 8-byte units; convert to bytes
            szBytes := sz * 8
            if szBytes > 0 && szBytes < 0x100000 {
                // Treat as bytes and add BOOT2 section starting right after FS3 header (0x38)
                bootOff := uint64(0x38)
                // Ensure we do not exceed file
                if int(bootOff)+int(szBytes) <= fileSize {
                    data, _ := p.reader.ReadSection(int64(bootOff), szBytes)
                    // CRC type: in-section for BOOT2 to align with existing logic
                    p.logger.Debug("Adding FS3 BOOT2 section", zap.Uint64("offset", bootOff), zap.Uint32("size", szBytes))
                    section, err := p.sectionFactory.CreateSectionFromData(uint16(types.SectionTypeBoot2), bootOff, szBytes, types.CRCInSection, 0, false, false, nil, false, data)
                    if err == nil {
                        p.addSection(section)
                    }
                }
            }
        }
    }

    return nil
}
