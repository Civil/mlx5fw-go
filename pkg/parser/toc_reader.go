package parser

import (
	"fmt"
	
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
	"go.uber.org/zap"
)

// TOCReader provides generic TOC reading functionality
type TOCReader struct {
	logger *zap.Logger
	sectionFactory interfaces.SectionFactory
}

// NewTOCReader creates a new TOC reader
func NewTOCReader(logger *zap.Logger) *TOCReader {
	return &TOCReader{logger: logger}
}

// NewTOCReaderWithFactory creates a new TOC reader with a section factory
func NewTOCReaderWithFactory(logger *zap.Logger, factory interfaces.SectionFactory) *TOCReader {
	return &TOCReader{
		logger: logger,
		sectionFactory: factory,
	}
}

// ReadTOCHeader reads and validates a TOC header
func (r *TOCReader) ReadTOCHeader(data []byte, tocAddr uint32, isDTOC bool) (*types.ITOCHeaderAnnotated, error) {
	if tocAddr+32 > uint32(len(data)) {
		return nil, merry.New("TOC header out of bounds")
	}

	// Use annotated version for unmarshaling
	header := &types.ITOCHeaderAnnotated{}
	headerData := data[tocAddr : tocAddr+32]
	if err := header.Unmarshal(headerData); err != nil {
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

// ReadTOCEntries reads all TOC entries until end marker
func (r *TOCReader) ReadTOCEntries(data []byte, tocAddr uint32) ([]*types.ITOCEntryAnnotated, error) {
	var entries []*types.ITOCEntryAnnotated
	entriesOffset := tocAddr + 32 // After header

	for i := 0; ; i++ {
		entryOffset := entriesOffset + uint32(i)*32
		if entryOffset+32 > uint32(len(data)) {
			break
		}

		// Use annotated version for unmarshaling
		entry := &types.ITOCEntryAnnotated{}
		entryData := data[entryOffset:entryOffset+32]
		if err := entry.Unmarshal(entryData); err != nil {
			// Log error but continue reading other entries
			r.logger.Warn("Failed to unmarshal ITOC entry",
				zap.Int("index", i),
				zap.Error(err))
			continue
		}
		
		
		// Debug: Log CRC info for all section types
		r.logger.Debug("ITOC entry CRC info",
			zap.String("type", types.GetSectionTypeName(uint16(entry.Type))),
			zap.Uint8("crc_field", entry.CRCField),
			zap.String("crc_type", entry.GetCRCType().String()),
			zap.Uint16("section_crc", entry.SectionCRC),
			zap.Bool("no_crc", entry.GetNoCRC()))

		// Stop at end marker (type 0xFF)
		if entry.GetType() == 0xFF {
			break
		}

		// Skip empty entries (all zeros except potentially the type field)
		if entry.GetSize() == 0 && entry.GetFlashAddr() == 0 {
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
	entries, err := r.ReadTOCEntries(data, tocAddr)
	if err != nil {
		return nil, err
	}

	// Convert entries to sections
	var sections []*interfaces.Section
	for _, entry := range entries {
		sectionType := entry.GetType()
		if isDTOC {
			// DTOC sections use different type mapping
			// DTOC entry types < 0x20 need to be OR'd with 0xE000
			// DTOC entry types >= 0xE0 are already in the correct range
			if sectionType < 0x20 {
				sectionType = sectionType | 0xE000
			}
		}

		section := &interfaces.Section{
			Type:         sectionType,
			Offset:       uint64(entry.GetFlashAddr()),
			Size:         entry.GetSize(),
			CRCType:      r.GetCRCType(entry),
			IsDeviceData: isDTOC,
			Entry:        entry, // Use annotated entry directly
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

// ReadTOCSectionsNew reads TOC sections and returns section interfaces
func (r *TOCReader) ReadTOCSectionsNew(data []byte, tocAddr uint32, isDTOC bool) ([]interfaces.CompleteSectionInterface, error) {
	if r.sectionFactory == nil {
		return nil, merry.New("section factory not set")
	}

	tocType := "ITOC"
	if isDTOC {
		tocType = "DTOC"
	}

	r.logger.Debug("Reading TOC sections",
		zap.String("type", tocType),
		zap.Uint32("address", tocAddr))

	// Read and validate header
	header, err := r.ReadTOCHeader(data, tocAddr, isDTOC)
	if err != nil {
		return nil, err
	}

	// Log header info
	r.logger.Debug("TOC header",
		zap.String("type", tocType),
		zap.Uint32("signature", header.Signature0),
		zap.Uint32("version", header.Version))

	// Read entries
	entries, err := r.ReadTOCEntries(data, tocAddr)
	if err != nil {
		return nil, err
	}

	// Convert entries to sections using factory
	var sections []interfaces.CompleteSectionInterface
	for _, entry := range entries {
		sectionType := entry.GetType()
		if isDTOC {
			// DTOC sections use different type mapping
			// DTOC entry types < 0x20 need to be OR'd with 0xE000
			// DTOC entry types >= 0xE0 are already in the correct range
			if sectionType < 0x20 {
				sectionType = sectionType | 0xE000
			}
		}

		flashAddr := entry.GetFlashAddr()
		if sectionType == 16 { // IMAGE_INFO
			r.logger.Debug("IMAGE_INFO ITOC entry details",
				zap.Uint32("flash_addr_dwords_raw", entry.FlashAddrDwords),
				zap.Uint32("flash_addr_bytes", flashAddr),
				zap.String("hex_addr", fmt.Sprintf("0x%x", flashAddr)))
		}
		if sectionType == 4 { // IRON_PREP_CODE
			r.logger.Debug("IRON_PREP_CODE ITOC entry details",
				zap.Uint32("size_dwords_raw", entry.SizeDwords),
				zap.Uint32("size_bytes", entry.GetSize()),
				zap.String("hex_size", fmt.Sprintf("0x%x", entry.GetSize())),
				zap.Uint32("flash_addr", flashAddr),
				zap.String("hex_addr", fmt.Sprintf("0x%x", flashAddr)))
		}
		
		// Special handling for HASHES_TABLE sections (type 0xfa)
		// HASHES_TABLE sections have dynamic size that needs to be calculated from the header
		var sectionSize uint32 = entry.GetSize()
		if sectionType == 0xfa && !isDTOC { // HASHES_TABLE sections
			calculatedSize, err := r.calculateHashesTableSize(data, flashAddr)
			if err != nil {
				r.logger.Warn("Failed to calculate HASHES_TABLE size, using ITOC entry size",
					zap.Uint32("flash_addr", flashAddr),
					zap.Uint32("itoc_size", entry.GetSize()),
					zap.Error(err))
			} else {
				r.logger.Debug("HASHES_TABLE dynamic size calculation",
					zap.Uint32("flash_addr", flashAddr),
					zap.Uint32("itoc_size", entry.GetSize()),
					zap.Uint32("calculated_size", calculatedSize))
				sectionSize = calculatedSize
			}
		}
		
		section, err := r.sectionFactory.CreateSection(
			uint16(sectionType),
			uint64(flashAddr),
			sectionSize,
			r.GetCRCType(entry),
			uint32(entry.SectionCRC),
			false, // isEncrypted - will be determined later
			isDTOC,
			entry, // Use annotated entry directly
			false, // isFromHWPointer
		)
		if err != nil {
			r.logger.Warn("Failed to create section",
				zap.String("toc", tocType),
				zap.Uint16("type", uint16(sectionType)),
				zap.Error(err))
			continue
		}

		r.logger.Debug("Found section",
			zap.String("toc", tocType),
			zap.String("name", r.getSectionName(entry.GetType(), isDTOC)),
			zap.Uint16("type", uint16(sectionType)),
			zap.Uint64("offset", section.Offset()),
			zap.Uint32("size", section.Size()))

		sections = append(sections, section)
	}

	return sections, nil
}

// GetCRCType determines the CRC type from ITOC entry
func (r *TOCReader) GetCRCType(entry *types.ITOCEntryAnnotated) types.CRCType {
	// Use the entry's GetCRCType method which extracts the CRC type
	// from the 3-bit CRCField that was parsed by annotations
	return entry.GetCRCType()
}

// getSectionName returns the section name based on type
func (r *TOCReader) getSectionName(sectionType uint16, isDTOC bool) string {
	if isDTOC {
		return types.GetDTOCSectionTypeName(uint8(sectionType))
	}
	return types.GetSectionTypeName(sectionType)
}

// GetCRCTypeLegacy determines the CRC type from legacy ITOC entry flags
func (r *TOCReader) GetCRCTypeLegacy(entry *types.ITOCEntry) types.CRCType {
	// If CRC field has bit 207 set (no_crc flag), it means NO CRC
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
		if err := entry.Unmarshal(data[entryOffset:entryOffset+32]); err != nil {
			continue
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// calculateHashesTableSize calculates the dynamic size of a HASHES_TABLE section
// by reading its header and using the formula: (4 + DwSize) * 4
func (r *TOCReader) calculateHashesTableSize(data []byte, flashAddr uint32) (uint32, error) {
	// Check bounds for header read (FS4HashesTableHeader is 12 bytes)
	if flashAddr+12 > uint32(len(data)) {
		return 0, merry.Errorf("HASHES_TABLE header out of bounds at 0x%x", flashAddr)
	}
	
	// Read header data
	headerData := data[flashAddr : flashAddr+12]
	
	// Parse header using FS4HashesTableHeader structure
	header := &types.FS4HashesTableHeader{}
	if err := header.Unmarshal(headerData); err != nil {
		return 0, merry.Wrap(err)
	}
	
	// Calculate size using mstflint formula: (4 + DwSize) * 4
	calculatedSize := (4 + header.DwSize) * 4
	
	r.logger.Debug("HASHES_TABLE header parsed",
		zap.Uint32("flash_addr", flashAddr),
		zap.Uint32("dw_size", header.DwSize),
		zap.Uint32("calculated_size", calculatedSize))
	
	return calculatedSize, nil
}