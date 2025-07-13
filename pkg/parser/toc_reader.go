package parser

import (
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
	"github.com/ghostiam/binstruct"
	"go.uber.org/zap"
)

// TOCReader provides generic TOC reading functionality
type TOCReader struct {
	logger *zap.Logger
}

// NewTOCReader creates a new TOC reader
func NewTOCReader(logger *zap.Logger) *TOCReader {
	return &TOCReader{logger: logger}
}

// ReadTOCHeader reads and validates a TOC header
func (r *TOCReader) ReadTOCHeader(data []byte, tocAddr uint32, isDTOC bool) (*types.ITOCHeader, error) {
	if tocAddr+32 > uint32(len(data)) {
		return nil, merry.New("TOC header out of bounds")
	}

	header := &types.ITOCHeader{}
	headerData := data[tocAddr : tocAddr+32]
	if err := binstruct.UnmarshalBE(headerData, header); err != nil {
		return nil, merry.Wrap(err)
	}

	// Verify signature
	expectedSig := uint32(types.ITOCSignature)
	if isDTOC {
		expectedSig = uint32(types.DTOCSignature)
	}
	if header.Signature0 != expectedSig {
		return nil, merry.Errorf("invalid TOC signature: 0x%08X", header.Signature0)
	}

	return header, nil
}

// ReadTOCEntries reads all TOC entries
func (r *TOCReader) ReadTOCEntries(data []byte, tocAddr uint32, maxEntries int) ([]*types.ITOCEntry, error) {
	var entries []*types.ITOCEntry
	entriesOffset := tocAddr + 32 // After header

	for i := 0; i < maxEntries; i++ {
		entryOffset := entriesOffset + uint32(i)*32
		if entryOffset+32 > uint32(len(data)) {
			break
		}

		entry := &types.ITOCEntry{}
		copy(entry.Data[:], data[entryOffset:entryOffset+32])
		entry.ParseFields()

		// Stop at end marker (type 0xFF)
		if entry.GetType() == 0xFF {
			break
		}

		// Skip empty entries (all zeros except potentially the type field)
		if entry.Size == 0 && entry.FlashAddr == 0 {
			continue
		}

		// Don't filter out entries based on flash address
		// mstflint processes all entries regardless of address value

		entries = append(entries, entry)
	}

	return entries, nil
}

// ReadTOCSections reads all sections from a TOC
func (r *TOCReader) ReadTOCSections(data []byte, tocAddr uint32, isDTOC bool) ([]*interfaces.Section, error) {
	// Read and validate header
	header, err := r.ReadTOCHeader(data, tocAddr, isDTOC)
	if err != nil {
		return nil, err
	}

	// Log header info
	tocType := "ITOC"
	if isDTOC {
		tocType = "DTOC"
	}
	r.logger.Debug("Parsing TOC",
		zap.String("type", tocType),
		zap.Uint32("addr", tocAddr),
		zap.Uint32("signature", header.Signature0))

	// Read entries
	entries, err := r.ReadTOCEntries(data, tocAddr, 256)
	if err != nil {
		return nil, err
	}

	// Convert entries to sections
	var sections []*interfaces.Section
	for _, entry := range entries {
		sectionType := entry.GetType()
		if isDTOC {
			// DTOC sections use different type mapping
			sectionType = entry.GetType() | 0xE000
		}

		section := &interfaces.Section{
			Type:         sectionType,
			Offset:       uint64(entry.FlashAddr),
			Size:         entry.Size,
			CRCType:      r.GetCRCType(entry),
			IsDeviceData: isDTOC,
			Entry:        entry,
		}

		r.logger.Debug("Found section",
			zap.String("toc", tocType),
			zap.String("name", r.getSectionName(entry.GetType(), isDTOC)),
			zap.Uint16("type", sectionType),
			zap.Uint64("offset", section.Offset),
			zap.Uint32("size", section.Size))

		sections = append(sections, section)
	}

	return sections, nil
}

// GetCRCType determines the CRC type from ITOC entry flags
func (r *TOCReader) GetCRCType(entry *types.ITOCEntry) types.CRCType {
	// If CRC field is 1, it means NO CRC
	if entry.GetNoCRC() {
		return types.CRCNone
	}
	// If there's a section CRC value in the ITOC entry, use it
	if entry.SectionCRC != 0 {
		return types.CRCInITOCEntry
	}
	// Otherwise, CRC is at the end of the section
	return types.CRCInSection
}

// getSectionName returns the section name based on type
func (r *TOCReader) getSectionName(sectionType uint16, isDTOC bool) string {
	if isDTOC {
		return types.GetDTOCSectionTypeName(uint8(sectionType))
	}
	return types.GetSectionTypeName(sectionType)
}

// ReadTOCRawEntries reads raw TOC entries without converting to sections
// This is useful for the replacer which needs the raw entries
func (r *TOCReader) ReadTOCRawEntries(data []byte, tocAddr uint32, isDTOC bool) ([]*types.ITOCEntry, error) {
	// Validate header first
	_, err := r.ReadTOCHeader(data, tocAddr, isDTOC)
	if err != nil {
		return nil, err
	}

	// Use CalculateNumSections to get the exact count
	numSections, err := types.CalculateNumSections(data, tocAddr)
	if err != nil {
		return nil, merry.Wrap(err)
	}

	// Read exact number of entries
	var entries []*types.ITOCEntry
	entriesOffset := tocAddr + 32

	for i := 0; i < numSections; i++ {
		entryOffset := entriesOffset + uint32(i)*32
		if entryOffset+32 > uint32(len(data)) {
			break
		}

		entry := &types.ITOCEntry{}
		copy(entry.Data[:], data[entryOffset:entryOffset+32])
		entry.ParseFields()

		entries = append(entries, entry)
	}

	return entries, nil
}