package section

import (
	"encoding/binary"

	"github.com/Civil/mlx5fw-go/pkg/errors"
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/parser/fs4"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
	"go.uber.org/zap"
)

const (
	// Firmware size limits
	FirmwareSize32MB = 32 * 1024 * 1024 // 32MB for older cards
	FirmwareSize64MB = 64 * 1024 * 1024 // 64MB for newer cards
	
	// ITOC entry size
	ITOCEntrySize = 32
	
	// Entry size in dwords conversion
	DwordSize = 4
	
	// CRC sizes
	CRCDwordSize = 7 // 28 bytes = 7 dwords for ITOC/DTOC CRC
)

// Replacer handles section replacement with proper firmware structure updates
type Replacer struct {
	parser       *fs4.Parser
	firmwareData []byte
	logger       *zap.Logger
	tocReader    *parser.TOCReader
}

// NewReplacer creates a new section replacer
func NewReplacer(fwParser *fs4.Parser, firmwareData []byte, logger *zap.Logger) *Replacer {
	return &Replacer{
		parser:       fwParser,
		firmwareData: firmwareData,
		logger:       logger,
		tocReader:    parser.NewTOCReader(logger),
	}
}

// ReplaceSection replaces a section and updates all affected structures
func (r *Replacer) ReplaceSection(targetSection interfaces.CompleteSectionInterface, newData []byte) ([]byte, error) {
	oldSize := targetSection.Size()
	newSize := uint32(len(newData))
	sizeDiff := int32(newSize) - int32(oldSize)

	// Check firmware size limit
	fwSizeLimit := r.determineFirmwareSizeLimit()
	if sizeDiff > 0 {
		newTotalSize := len(r.firmwareData) + int(sizeDiff)
		if newTotalSize > fwSizeLimit {
			return nil, merry.Errorf("firmware size would exceed limit: %d bytes > %d MB limit", 
				newTotalSize, fwSizeLimit/(1024*1024))
		}
	}

	// Create a working copy
	workingData := make([]byte, len(r.firmwareData))
	copy(workingData, r.firmwareData)

	if sizeDiff == 0 {
		// Simple case: same size replacement
		copy(workingData[targetSection.Offset():], newData)
		
		// Update CRC
		err := r.updateSectionCRC(workingData, targetSection, newSize)
		if err != nil {
			return nil, merry.Wrap(err)
		}
		
		// Pad firmware to proper size
		workingData = r.padFirmware(workingData)
		return workingData, nil
	}

	// Complex case: size changed, need to relocate sections
	r.logger.Info("Section size changed, relocating sections",
		zap.Int32("sizeDiff", sizeDiff))

	// Get parser internals through reflection or exposed methods
	itocAddr := r.parser.GetITOCAddress()
	dtocAddr := r.parser.GetDTOCAddress()
	
	// Build relocation map
	relocMap, err := r.buildRelocationMap(targetSection, sizeDiff, itocAddr, dtocAddr)
	if err != nil {
		return nil, merry.Wrap(err)
	}

	// Extend firmware if needed
	if sizeDiff > 0 {
		maxNewEnd := uint32(0)
		for oldOffset, reloc := range relocMap {
			sectionEnd := reloc.newOffset + reloc.size
			if sectionEnd > maxNewEnd {
				maxNewEnd = sectionEnd
			}
			r.logger.Debug("Relocation entry",
				zap.Uint32("oldOffset", oldOffset),
				zap.Uint32("newOffset", reloc.newOffset),
				zap.Uint32("size", reloc.size))
		}
		
		if maxNewEnd > uint32(len(workingData)) {
			newWorkingData := make([]byte, maxNewEnd)
			copy(newWorkingData, workingData)
			workingData = newWorkingData
			r.logger.Info("Extended firmware size", zap.Uint32("newSize", maxNewEnd))
		}
	}

	// Relocate sections
	err = r.relocateSections(workingData, relocMap, targetSection.Offset(), sizeDiff)
	if err != nil {
		return nil, merry.Wrap(err)
	}

	// Copy new data for target section
	copy(workingData[targetSection.Offset():], newData)

	// Update all firmware structures
	err = r.updateFirmwareStructures(workingData, relocMap, targetSection, newSize, itocAddr, dtocAddr)
	if err != nil {
		return nil, merry.Wrap(err)
	}

	// Pad firmware to proper size
	workingData = r.padFirmware(workingData)
	return workingData, nil
}

// determineFirmwareSizeLimit determines the firmware size limit based on DTOC location
func (r *Replacer) determineFirmwareSizeLimit() int {
	fileSize := len(r.firmwareData)
	
	// Check for specific sizes that indicate card type
	if fileSize >= 0x2000000 && fileSize <= 0x2000000+0x100000 { // ~32MB
		return FirmwareSize32MB
	}
	if fileSize >= 0x4000000 && fileSize <= 0x4000000+0x100000 { // ~64MB
		return FirmwareSize64MB
	}

	// Default based on file size
	if fileSize > FirmwareSize32MB {
		return FirmwareSize64MB
	}
	return FirmwareSize32MB
}

// padFirmware pads the firmware to 32MB or 64MB with 0xFF
func (r *Replacer) padFirmware(data []byte) []byte {
	targetSize := r.determineFirmwareSizeLimit()
	currentSize := len(data)
	
	if currentSize >= targetSize {
		return data[:targetSize]
	}
	
	// Create padded firmware
	paddedData := make([]byte, targetSize)
	copy(paddedData, data)
	
	// Fill remaining with 0xFF
	for i := currentSize; i < targetSize; i++ {
		paddedData[i] = 0xFF
	}
	
	r.logger.Info("Padded firmware to target size",
		zap.Int("originalSize", currentSize),
		zap.Int("targetSize", targetSize))
	
	return paddedData
}

// relocationInfo stores relocation information for a section
type relocationInfo struct {
	newOffset   uint32
	size        uint32
	sectionType uint16
	crcType     types.CRCType
	isITOC      bool
	isDTOC      bool
	entryIndex  int
}

// buildRelocationMap builds a map of all sections that need relocation
func (r *Replacer) buildRelocationMap(targetSection interfaces.SectionInterface, sizeDiff int32, itocAddr, dtocAddr uint32) (map[uint32]*relocationInfo, error) {
	relocMap := make(map[uint32]*relocationInfo)
	
	// Process ITOC sections
	itocSections, err := r.readTOCSections(itocAddr, false)
	if err != nil {
		return nil, merry.Wrap(err)
	}

	for i, entry := range itocSections {
		if entry.GetFlashAddr() == 0 || entry.GetType() == 0xFF {
			continue
		}
		
		reloc := &relocationInfo{
			newOffset:   entry.GetFlashAddr(),
			size:        entry.GetSize(),
			sectionType: entry.GetType(),
			crcType:     r.tocReader.GetCRCTypeLegacy(entry),
			isITOC:      true,
			entryIndex:  i,
		}
		
		// Relocate if after target section
		if entry.GetFlashAddr() > uint32(targetSection.Offset()) {
			reloc.newOffset = uint32(int32(entry.GetFlashAddr()) + sizeDiff)
		}
		
		relocMap[entry.GetFlashAddr()] = reloc
	}

	// Process DTOC sections
	if dtocAddr > 0 {
		dtocSections, err := r.readTOCSections(dtocAddr, true)
		if err == nil { // DTOC might not exist
			for i, entry := range dtocSections {
				if entry.GetFlashAddr() == 0 || entry.GetType() == 0xFF {
					continue
				}
				
				// Don't duplicate entries already in relocMap
				if _, exists := relocMap[entry.GetFlashAddr()]; exists {
					continue
				}
				
				reloc := &relocationInfo{
					newOffset:   entry.GetFlashAddr(),
					size:        entry.GetSize(),
					sectionType: entry.GetType() | 0xE000, // DTOC type mapping
					crcType:     r.tocReader.GetCRCTypeLegacy(entry),
					isDTOC:      true,
					entryIndex:  i,
				}
				
				// Relocate if after target section
				if entry.GetFlashAddr() > uint32(targetSection.Offset()) {
					reloc.newOffset = uint32(int32(entry.GetFlashAddr()) + sizeDiff)
				}
				
				relocMap[entry.GetFlashAddr()] = reloc
			}
		}
	}

	return relocMap, nil
}

// readTOCSections reads all entries from a TOC
func (r *Replacer) readTOCSections(tocAddr uint32, isDTOC bool) ([]*types.ITOCEntry, error) {
	if tocAddr == 0 || tocAddr >= uint32(len(r.firmwareData)) {
		return nil, errors.InvalidParameterError("TOC address", "cannot be zero")
	}

	// Use the generic TOC reader to get raw entries
	return r.tocReader.ReadTOCRawEntries(r.firmwareData, tocAddr, isDTOC)
}


// relocateSections moves section data to new offsets
func (r *Replacer) relocateSections(workingData []byte, relocMap map[uint32]*relocationInfo, targetOffset uint64, sizeDiff int32) error {
	// Create sorted list of sections by offset
	var sortedOffsets []uint32
	for offset := range relocMap {
		sortedOffsets = append(sortedOffsets, offset)
	}
	
	// Sort offsets
	for i := 0; i < len(sortedOffsets)-1; i++ {
		for j := i + 1; j < len(sortedOffsets); j++ {
			if sortedOffsets[i] > sortedOffsets[j] {
				sortedOffsets[i], sortedOffsets[j] = sortedOffsets[j], sortedOffsets[i]
			}
		}
	}

	// Move sections (from end to beginning if growing, beginning to end if shrinking)
	if sizeDiff > 0 {
		// Growing - move from end to beginning
		for i := len(sortedOffsets) - 1; i >= 0; i-- {
			oldOffset := sortedOffsets[i]
			reloc := relocMap[oldOffset]
			
			if oldOffset <= uint32(targetOffset) || oldOffset == reloc.newOffset {
				continue
			}
			
			// Move section data
			sectionData := make([]byte, reloc.size)
			copy(sectionData, workingData[oldOffset:oldOffset+reloc.size])
			copy(workingData[reloc.newOffset:], sectionData)
			
			r.logger.Debug("Relocated section",
				zap.Uint32("oldOffset", oldOffset),
				zap.Uint32("newOffset", reloc.newOffset),
				zap.Uint32("size", reloc.size))
		}
	} else {
		// Shrinking - move from beginning to end
		for _, oldOffset := range sortedOffsets {
			reloc := relocMap[oldOffset]
			
			if oldOffset <= uint32(targetOffset) || oldOffset == reloc.newOffset {
				continue
			}
			
			// Move section data
			copy(workingData[reloc.newOffset:], workingData[oldOffset:oldOffset+reloc.size])
			
			r.logger.Debug("Relocated section",
				zap.Uint32("oldOffset", oldOffset),
				zap.Uint32("newOffset", reloc.newOffset),
				zap.Uint32("size", reloc.size))
		}
	}

	return nil
}

// updateFirmwareStructures updates ITOC, DTOC, and HW pointers
func (r *Replacer) updateFirmwareStructures(workingData []byte, relocMap map[uint32]*relocationInfo, targetSection interfaces.SectionInterface, newSize uint32, itocAddr, dtocAddr uint32) error {
	// Update ITOC entries
	err := r.updateTOCEntries(workingData, relocMap, itocAddr, false, targetSection, newSize)
	if err != nil {
		return merry.Wrap(err)
	}

	// Update DTOC entries
	if dtocAddr > 0 {
		err = r.updateTOCEntries(workingData, relocMap, dtocAddr, true, nil, 0)
		if err != nil {
			r.logger.Warn("Failed to update DTOC", zap.Error(err))
		}
	}

	// Update HW pointers
	err = r.updateHWPointers(workingData, relocMap)
	if err != nil {
		return merry.Wrap(err)
	}

	// Update CRCs for all affected sections
	for oldOffset, reloc := range relocMap {
		if oldOffset >= uint32(targetSection.Offset()) {
			// TODO: Fix CRC update for relocated sections
			// This needs to be refactored to use SectionInterface instead of the legacy Section struct
			// For now, skip CRC update for relocated sections
			r.logger.Warn("CRC update for relocated sections not yet implemented",
				zap.Uint16("type", reloc.sectionType),
				zap.Uint32("offset", reloc.newOffset))
		}
	}

	return nil
}

// updateTOCEntries updates TOC entries with new offsets and sizes
func (r *Replacer) updateTOCEntries(workingData []byte, relocMap map[uint32]*relocationInfo, tocAddr uint32, isDTOC bool, targetSection interfaces.SectionInterface, newSize uint32) error {
	if tocAddr == 0 {
		return nil
	}

	entries, err := r.readTOCSections(tocAddr, isDTOC)
	if err != nil {
		return merry.Wrap(err)
	}

	crcCalc := parser.NewCRCCalculator()
	entriesOffset := tocAddr + ITOCEntrySize

	for i, entry := range entries {
		if entry.GetFlashAddr() == 0 || entry.GetType() == 0xFF {
			continue
		}

		// Check if this entry needs updating
		reloc, exists := relocMap[entry.GetFlashAddr()]
		if !exists {
			continue
		}

		// Only update if this TOC type matches
		if (isDTOC && !reloc.isDTOC) || (!isDTOC && !reloc.isITOC) {
			continue
		}

		// Update entry fields
		if reloc.newOffset != entry.GetFlashAddr() {
			entry.SetFlashAddr(reloc.newOffset)
		}

		// Update size for target section
		if targetSection != nil && entry.GetFlashAddr() == uint32(targetSection.Offset()) {
			entry.SetSize(newSize)
			r.logger.Info("Updating target section in ITOC",
				zap.Uint32("offset", entry.GetFlashAddr()),
				zap.Uint32("oldSize", reloc.size), 
				zap.Uint32("newSize", newSize))
		}

		// Calculate section CRC if it's stored in ITOC entry
		if reloc.crcType == types.CRCInITOCEntry {
			// Read section data at new offset
			sectionOffset := reloc.newOffset
			sectionSize := reloc.size
			if targetSection != nil && entry.GetFlashAddr() == uint32(targetSection.Offset()) {
				sectionSize = newSize
			}
			
			if sectionOffset+sectionSize <= uint32(len(workingData)) {
				sectionData := workingData[sectionOffset : sectionOffset+sectionSize]
				
				// Align to dwords and calculate CRC
				alignedSize := (sectionSize + 3) & ^uint32(3)
				paddedData := make([]byte, alignedSize)
				copy(paddedData, sectionData)
				
				sizeInDwords := alignedSize / DwordSize
				sectionCRC := crcCalc.CalculateImageCRC(paddedData, int(sizeInDwords))
				entry.SectionCRC = sectionCRC
				
				r.logger.Debug("Calculated section CRC for ITOC entry",
					zap.Uint16("type", reloc.sectionType),
					zap.Uint16("crc", sectionCRC),
					zap.Uint32("offset", sectionOffset),
					zap.Uint32("size", sectionSize))
			}
		}

		// Serialize entry back to binary
		entryOffset := entriesOffset + uint32(i)*ITOCEntrySize
		err := r.serializeITOCEntry(entry, workingData[entryOffset:entryOffset+ITOCEntrySize])
		if err != nil {
			return merry.Wrap(err)
		}

		// Calculate and update entry CRC
		// Use Software CRC16 (CalculateImageCRC) for ITOC entries - same as mstflint
		// CRC is calculated over first 28 bytes (7 dwords) of the 32-byte entry
		entryCRCData := workingData[entryOffset : entryOffset+28]
		entryCRC := crcCalc.CalculateImageCRC(entryCRCData, CRCDwordSize)  
		// Store CRC in last 2 bytes of the entry (offset 30-31)
		binary.BigEndian.PutUint16(workingData[entryOffset+30:entryOffset+32], entryCRC)
	}

	// Update TOC header CRC
	// ITOC/DTOC header uses Software CRC16 (CalculateImageCRC) on first 28 bytes
	// The CRC is stored as 16-bit value in the lower half of the 32-bit field
	headerCRCData := workingData[tocAddr : tocAddr+28]
	headerCRC := crcCalc.CalculateImageCRC(headerCRCData, CRCDwordSize)
	
	// Keep upper 16 bits unchanged, update lower 16 bits with CRC
	currentCRCField := binary.BigEndian.Uint32(workingData[tocAddr+28:tocAddr+32])
	newCRCField := (currentCRCField & 0xFFFF0000) | uint32(headerCRC)
	binary.BigEndian.PutUint32(workingData[tocAddr+28:tocAddr+32], newCRCField)

	return nil
}

// serializeITOCEntry converts an ITOCEntry back to binary format
func (r *Replacer) serializeITOCEntry(entry *types.ITOCEntry, data []byte) error {
	if len(data) < ITOCEntrySize {
		return errors.DataTooShortError(40, len(data), "ITOC entry buffer")
	}

	// Marshal using annotation-based marshaling
	marshaledData, err := entry.Marshal()
	if err != nil {
		return merry.Wrap(err)
	}
	
	// Copy marshaled data
	copy(data, marshaledData)
	
	return nil
}


// updateHWPointers updates hardware pointers that reference relocated sections
func (r *Replacer) updateHWPointers(workingData []byte, relocMap map[uint32]*relocationInfo) error {
	// Find magic pattern to locate HW pointers
	magicOffset, err := r.findMagicPattern(workingData)
	if err != nil {
		return merry.Wrap(err)
	}

	hwPointersOffset := magicOffset + 0x18
	if hwPointersOffset+128 > uint32(len(workingData)) {
		return merry.Wrap(errors.ErrInvalidParameter, merry.WithMessagef("HW pointers offset %d out of bounds", hwPointersOffset))
	}

	// Read HW pointers using annotation-based unmarshaling
	hwPointers := &types.FS4HWPointers{}
	hwPointersData := workingData[hwPointersOffset : hwPointersOffset+128]
	if err := hwPointers.Unmarshal(hwPointersData); err != nil {
		return merry.Wrap(err)
	}

	crcCalc := parser.NewCRCCalculator()
	updated := false

	// Helper to update a pointer
	updatePointer := func(name string, ptr *types.HWPointerEntry) {
		if ptr.Ptr == 0 || ptr.Ptr == 0xFFFFFFFF {
			return
		}

		if reloc, exists := relocMap[ptr.Ptr]; exists && reloc.newOffset != ptr.Ptr {
			r.logger.Debug("Updating HW pointer",
				zap.String("name", name),
				zap.Uint32("old", ptr.Ptr),
				zap.Uint32("new", reloc.newOffset))

			ptr.Ptr = reloc.newOffset
			// Recalculate pointer CRC
			ptrBytes := make([]byte, 4)
			binary.BigEndian.PutUint32(ptrBytes, ptr.Ptr)
			ptr.CRC = crcCalc.CalculateHardwareCRC(ptrBytes)
			updated = true
		}
	}

	// Update all pointers
	updatePointer("Boot2Ptr", &hwPointers.Boot2Ptr)
	updatePointer("TOCPtr", &hwPointers.TOCPtr)
	updatePointer("ToolsPtr", &hwPointers.ToolsPtr)
	updatePointer("ImageInfoSectionPtr", &hwPointers.ImageInfoSectionPtr)
	updatePointer("HashesTablePtr", &hwPointers.HashesTablePtr)
	// Add other pointers as needed

	if updated {
		// Write back HW pointers
		r.writeHWPointers(workingData[hwPointersOffset:], hwPointers)
		r.logger.Info("Updated HW pointers")
	}

	return nil
}

// findMagicPattern finds the FS4 magic pattern
func (r *Replacer) findMagicPattern(data []byte) (uint32, error) {
	// Use the same search offsets as FirmwareReader for consistency
	for _, offset := range types.MagicSearchOffsets {
		if offset+8 > uint32(len(data)) {
			continue
		}
		
		if binary.BigEndian.Uint64(data[offset:offset+8]) == types.MagicPattern {
			return offset, nil
		}
	}

	return 0, merry.Wrap(errors.ErrInvalidMagic, merry.WithMessage("magic pattern not found in firmware"))
}

// writeHWPointers writes HW pointers structure back to binary
func (r *Replacer) writeHWPointers(data []byte, hw *types.FS4HWPointers) {
	// Marshal using annotation-based marshaling
	marshaledData, err := hw.Marshal()
	if err != nil {
		r.logger.Error("Failed to marshal HW pointers", zap.Error(err))
		return
	}
	
	// Copy marshaled data
	copy(data, marshaledData)
}

// updateSectionCRC updates the CRC for a section
func (r *Replacer) updateSectionCRC(workingData []byte, section interfaces.SectionInterface, newSize uint32) error {
	switch section.CRCType() {
	case types.CRCNone:
		return nil
		
	case types.CRCInITOCEntry:
		// For CRCInITOCEntry, the CRC is stored in the ITOC entry's SectionCRC field
		// We'll handle this in updateTOCEntries where we have access to the entry
		r.logger.Debug("Section uses CRC in ITOC entry, will update during TOC processing",
			zap.String("type", types.GetSectionTypeName(section.Type())),
			zap.Uint32("offset", uint32(section.Offset())),
			zap.Uint32("size", newSize))
		return nil
		
	case types.CRCInSection:
		// CRC at end of section
		if newSize < 4 {
			return errors.DataTooShortError(4, int(newSize), "section for CRC")
		}

		crcCalc := parser.NewCRCCalculator()
		sectionData := workingData[section.Offset() : section.Offset()+uint64(newSize)]

		// Calculate CRC on all data except last dword
		crcSizeInDwords := (newSize / DwordSize) - 1
		alignedSize := crcSizeInDwords * DwordSize
		crcData := sectionData[:alignedSize]
		
		// Ensure data is aligned
		if len(crcData) != int(alignedSize) {
			paddedData := make([]byte, alignedSize)
			copy(paddedData, crcData)
			crcData = paddedData
		}

		newCRC := crcCalc.CalculateImageCRC(crcData, int(crcSizeInDwords))

		// Write CRC in lower 16 bits of last dword
		lastDwordOffset := section.Offset() + uint64(newSize) - 4
		currentLastDword := binary.BigEndian.Uint32(workingData[lastDwordOffset:])
		newLastDword := (currentLastDword & 0xFFFF0000) | uint32(newCRC)
		binary.BigEndian.PutUint32(workingData[lastDwordOffset:], newLastDword)

		r.logger.Debug("Updated section CRC",
			zap.String("type", types.GetSectionTypeName(section.Type())),
			zap.Uint16("crc", newCRC))
		return nil
		
	default:
		return merry.Errorf("unknown CRC type: %d", section.CRCType())
	}
}