package extract

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.uber.org/zap"

	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/parser/fs4"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/Civil/mlx5fw-go/pkg/types/extracted"
	types_sections "github.com/Civil/mlx5fw-go/pkg/types/sections"
)

// Options contains options for firmware extraction
type Options struct {
	OutputDir       string
	ExportJSON      bool     // Deprecated: JSON is always exported for parsed sections
	IncludeMetadata bool
	RemoveCRC       bool
	KeepBinary      bool     // Keep binary representation alongside JSON
}

// Extractor handles firmware extraction
type Extractor struct {
	parser  *fs4.Parser
	logger  *zap.Logger
	options Options
}

// New creates a new Extractor
func New(parser *fs4.Parser, logger *zap.Logger, opts Options) *Extractor {
	return &Extractor{
		parser:  parser,
		logger:  logger,
		options: opts,
	}
}

// Extract performs the firmware extraction
func (e *Extractor) Extract() error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(e.options.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get sections
	sections := e.parser.GetSections()
	
	// Extract sections and get filename mapping
	sectionFileMap, err := e.extractSections(sections)
	if err != nil {
		return err
	}
	
	// Extract metadata with filename mapping
	if err := e.extractMetadataWithFilenames(sectionFileMap); err != nil {
		e.logger.Warn("Failed to extract metadata", zap.Error(err))
	}
	
	// Extract gaps and unallocated data
	return e.extractGapsAndUnallocatedData()
}

func (e *Extractor) extractSections(sections map[uint16][]interfaces.CompleteSectionInterface) (map[uint64]string, error) {
	extractedCount := 0
	// Map to store CRC values for sections
	sectionCRCs := make(map[uint64]uint32) // key: offset, value: CRC
	// Map to store filenames for sections
	sectionFileMap := make(map[uint64]string) // key: offset, value: filename
	
	for _, sectionList := range sections {
		for i, section := range sectionList {
			// Get section type name
			typeName := section.TypeName()
			
			// Clean up the type name for filename
			fileName := strings.ReplaceAll(typeName, " ", "_")
			fileName = strings.ReplaceAll(fileName, "/", "_")
			
			// Add index if multiple sections of same type
			if len(sectionList) > 1 {
				fileName = fmt.Sprintf("%s_%d_0x%08x.bin", fileName, i, section.Offset())
			} else {
				fileName = fmt.Sprintf("%s_0x%08x.bin", fileName, section.Offset())
			}
			
			filePath := filepath.Join(e.options.OutputDir, fileName)
			
			e.logger.Info("Extracting section",
				zap.String("type", typeName),
				zap.Uint64("offset", section.Offset()),
				zap.Uint32("size", section.Size()),
				zap.String("file", fileName))
			
			// Load section data if not already loaded
			if section.GetRawData() == nil || len(section.GetRawData()) == 0 {
				// Read section data from firmware
				// For sections with IN_SECTION CRC, read the extra 4 bytes to get the CRC
				readSize := section.Size()
				if section.CRCType() == types.CRCInSection && !e.parser.IsEncrypted() {
					readSize += 4
				}
				data, err := e.parser.ReadSectionData(section.Type(), section.Offset(), readSize)
				if err != nil {
					e.logger.Warn("Failed to read section data", 
						zap.String("type", typeName),
						zap.Error(err))
					continue
				}
				
				// Parse the section data with full data including CRC
				// The section will store the complete data
				if err := section.Parse(data); err != nil {
					e.logger.Warn("Failed to parse section data",
						zap.String("type", typeName),
						zap.Error(err))
				}
			}
			
			// Get section data
			data := section.GetRawData()
			if data == nil || len(data) == 0 {
				e.logger.Warn("Section has no data", zap.String("type", typeName))
				continue
			}
			
			// Handle CRC removal based on encryption status
			dataToWrite := data
			originalCRCValue := uint32(0)
			if section.CRCType() == types.CRCInSection && len(data) >= 4 {
				// Extract CRC value before removal
				// For IN_SECTION CRCs, the format is:
				// - 16-bit CRC in upper 16 bits (bytes 0-1) 
				// - Lower 16 bits are 0 (bytes 2-3)
				originalCRCValue = binary.BigEndian.Uint32(data[len(data)-4:])
				// Store CRC value for later use in metadata
				sectionCRCs[section.Offset()] = originalCRCValue
				
				if !e.parser.IsEncrypted() {
					// Only remove CRC for non-encrypted firmwares
					// For encrypted firmwares, keep the data intact
					dataToWrite = data[:len(data)-4]
					e.logger.Debug("Removed CRC from end of section",
						zap.String("type", typeName),
						zap.Int("originalSize", len(data)),
						zap.Int("newSize", len(dataToWrite)),
						zap.Uint32("crcValue", originalCRCValue),
						zap.String("crcHex", fmt.Sprintf("0x%08x", originalCRCValue)))
				} else {
					e.logger.Debug("Keeping CRC intact for encrypted firmware section",
						zap.String("type", typeName),
						zap.Int("size", len(data)))
				}
			}
			
			// Always export JSON for parsed sections (JSON is the source of truth)
			jsonPath := strings.TrimSuffix(filePath, ".bin") + ".json"
			jsonData, err := json.MarshalIndent(section, "", "  ")
			if err != nil {
				e.logger.Warn("Failed to marshal section to JSON",
					zap.String("type", typeName),
					zap.Error(err))
			}
			
			jsonExported := false
			if err == nil {
				if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
					e.logger.Warn("Failed to write JSON",
						zap.String("file", jsonPath),
						zap.Error(err))
				} else {
					e.logger.Info("Exported JSON data",
						zap.String("file", jsonPath))
					jsonExported = true
				}
			}
			
			// Check if section has extracted data (like decompressed INI)
			if jsonExported {
				// Handle special sections with extracted data
				switch typedSection := section.(type) {
				case *types_sections.DBGFwIniSection:
					if extractedData := typedSection.GetExtractedData(); extractedData != nil {
						// Save the decompressed INI file
						iniPath := strings.TrimSuffix(filePath, ".bin") + ".ini"
						if err := os.WriteFile(iniPath, extractedData, 0644); err != nil {
							e.logger.Warn("Failed to write extracted INI file",
								zap.String("file", iniPath),
								zap.Error(err))
						} else {
							e.logger.Info("Exported extracted INI data",
								zap.String("file", iniPath))
						}
					}
				}
			}
			
			// Determine if we need to write binary file
			// Binary is needed when:
			// 1. KeepBinary flag is set (user wants both JSON and binary)
			// 2. Firmware is encrypted (binary data must be preserved exactly)
			// 3. JSON export failed
			// 4. Section has raw data flag (unparsed sections)
			shouldWriteBinary := e.options.KeepBinary || e.parser.IsEncrypted() || !jsonExported
			
			e.logger.Debug("Binary write decision initial",
				zap.String("type", typeName),
				zap.Bool("shouldWriteBinary", shouldWriteBinary),
				zap.Bool("jsonExported", jsonExported),
				zap.Bool("keepBinary", e.options.KeepBinary),
				zap.Bool("isEncrypted", e.parser.IsEncrypted()))
			
			if jsonExported && !shouldWriteBinary {
				// Check if section has raw data flag
				var sectionMeta struct {
					HasRawData bool `json:"has_raw_data"`
				}
				if err := json.Unmarshal(jsonData, &sectionMeta); err != nil {
					e.logger.Debug("Failed to unmarshal section metadata",
						zap.String("type", typeName),
						zap.Error(err))
				} else {
					e.logger.Debug("Unmarshaled section metadata",
						zap.String("type", typeName),
						zap.Bool("has_raw_data", sectionMeta.HasRawData))
					shouldWriteBinary = sectionMeta.HasRawData
					if shouldWriteBinary {
						e.logger.Debug("Section has raw data flag, keeping binary",
							zap.String("type", typeName))
					}
				}
			}
			
			if shouldWriteBinary {
				if err := os.WriteFile(filePath, dataToWrite, 0644); err != nil {
					return nil, fmt.Errorf("failed to write section %s: %w", fileName, err)
				}
				e.logger.Debug("Wrote binary file",
					zap.String("type", typeName),
					zap.String("file", fileName))
			}
			
			// Store filename in map only after successful extraction
			sectionFileMap[section.Offset()] = fileName
			
			extractedCount++
		}
	}
	
	e.logger.Info("Extracted sections", zap.Int("count", extractedCount))
	return sectionFileMap, nil
}

func (e *Extractor) extractGapsAndUnallocatedData() error {
	fileInfo, err := e.parser.GetReader().GetFileInfo()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	
	// Get all sections sorted by offset
	sections := e.parser.GetSections()
	allSections := make([]interfaces.CompleteSectionInterface, 0)
	for _, sectionList := range sections {
		allSections = append(allSections, sectionList...)
	}
	
	// Sort sections by offset
	sort.Slice(allSections, func(i, j int) bool {
		return allSections[i].Offset() < allSections[j].Offset()
	})
	
	// Create gaps directory
	gapsDir := filepath.Join(e.options.OutputDir, "gaps")
	if err := os.MkdirAll(gapsDir, 0755); err != nil {
		return fmt.Errorf("failed to create gaps directory: %w", err)
	}
	
	// Track gaps and extract data
	lastEnd := uint64(0)
	gapCount := 0
	
	for _, section := range allSections {
		// Check for gap before this section
		if section.Offset() > lastEnd {
			gapSize := section.Offset() - lastEnd
			
			// Read gap data
			gapData, err := e.parser.GetReader().ReadSection(int64(lastEnd), uint32(gapSize))
			if err != nil {
				e.logger.Warn("Failed to read gap data",
					zap.Uint64("offset", lastEnd),
					zap.Uint64("size", gapSize),
					zap.Error(err))
			} else {
				if err := e.saveGap(gapData, lastEnd, section.Offset(), gapCount, gapsDir); err == nil {
					gapCount++
				}
			}
		}
		
		// Update last end position
		originalSize := section.Size()
		if section.CRCType() == types.CRCInSection && !e.parser.IsEncrypted() {
			// For non-encrypted firmwares, account for removed CRC
			originalSize += 4
		}
		lastEnd = section.Offset() + uint64(originalSize)
	}
	
	// Check for trailing gap
	if fileInfo.Size > 0 && lastEnd < uint64(fileInfo.Size) {
		gapSize := uint64(fileInfo.Size) - lastEnd
		
		// Read trailing data
		gapData, err := e.parser.GetReader().ReadSection(int64(lastEnd), uint32(gapSize))
		if err != nil {
			e.logger.Warn("Failed to read trailing gap data",
				zap.Uint64("offset", lastEnd),
				zap.Uint64("size", gapSize),
				zap.Error(err))
		} else {
			if err := e.saveTrailingGap(gapData, lastEnd, gapSize, gapCount, gapsDir); err == nil {
				gapCount++
			}
		}
	}
	
	fmt.Printf("Successfully extracted data to %s (gaps: %d)\n", e.options.OutputDir, gapCount)
	return nil
}

func (e *Extractor) saveGap(gapData []byte, startOffset, endOffset uint64, gapIndex int, gapsDir string) error {
	gapSize := endOffset - startOffset
	
	// Check if gap is uniform (all bytes the same)
	isUniform := true
	var fillByte byte
	if len(gapData) > 0 {
		fillByte = gapData[0]
		for _, b := range gapData {
			if b != fillByte {
				isUniform = false
				break
			}
		}
	}
	
	gapFileName := fmt.Sprintf("gap_%03d_0x%08x_0x%08x", gapIndex, startOffset, endOffset)
	
	if isUniform && (fillByte == 0xFF || fillByte == 0x00) {
		// Save as metadata file for uniform gaps
		metaFileName := gapFileName + ".meta"
		metaPath := filepath.Join(gapsDir, metaFileName)
		metaContent := fmt.Sprintf("size=%d\nfill=0x%02x\n", gapSize, fillByte)
		
		if err := os.WriteFile(metaPath, []byte(metaContent), 0644); err != nil {
			e.logger.Warn("Failed to write gap metadata",
				zap.String("file", metaFileName),
				zap.Error(err))
			return err
		}
		e.logger.Info("Extracted uniform gap metadata",
			zap.String("file", metaFileName),
			zap.Uint64("offset", startOffset),
			zap.Uint64("size", gapSize),
			zap.Uint8("fillByte", fillByte))
	} else {
		// Save regular gap data
		gapPath := filepath.Join(gapsDir, gapFileName + ".bin")
		if err := os.WriteFile(gapPath, gapData, 0644); err != nil {
			e.logger.Warn("Failed to write gap data",
				zap.String("file", gapFileName + ".bin"),
				zap.Error(err))
			return err
		}
		e.logger.Info("Extracted gap data",
			zap.String("file", gapFileName + ".bin"),
			zap.Uint64("offset", startOffset),
			zap.Uint64("size", gapSize))
	}
	
	return nil
}

func (e *Extractor) saveTrailingGap(gapData []byte, startOffset, gapSize uint64, gapIndex int, gapsDir string) error {
	// Check if trailing gap is uniform
	isUniform := true
	var fillByte byte
	if len(gapData) > 0 {
		fillByte = gapData[0]
		for _, b := range gapData {
			if b != fillByte {
				isUniform = false
				break
			}
		}
	}
	
	gapFileName := fmt.Sprintf("gap_%03d_0x%08x_EOF", gapIndex, startOffset)
	
	if isUniform && (fillByte == 0xFF || fillByte == 0x00) {
		// Save as metadata file for uniform gaps
		metaFileName := gapFileName + ".meta"
		metaPath := filepath.Join(gapsDir, metaFileName)
		metaContent := fmt.Sprintf("size=%d\nfill=0x%02x\n", gapSize, fillByte)
		
		if err := os.WriteFile(metaPath, []byte(metaContent), 0644); err != nil {
			e.logger.Warn("Failed to write trailing gap metadata",
				zap.String("file", metaFileName),
				zap.Error(err))
			return err
		}
		e.logger.Info("Extracted uniform trailing gap metadata",
			zap.String("file", metaFileName),
			zap.Uint64("offset", startOffset),
			zap.Uint64("size", gapSize),
			zap.Uint8("fillByte", fillByte))
	} else {
		// Save regular gap data
		gapPath := filepath.Join(gapsDir, gapFileName + ".bin")
		if err := os.WriteFile(gapPath, gapData, 0644); err != nil {
			e.logger.Warn("Failed to write trailing gap data",
				zap.String("file", gapFileName + ".bin"),
				zap.Error(err))
			return err
		}
		e.logger.Info("Extracted trailing gap data",
			zap.String("file", gapFileName + ".bin"),
			zap.Uint64("offset", startOffset),
			zap.Uint64("size", gapSize))
	}
	
	return nil
}

func (e *Extractor) extractMetadataWithFilenames(sectionFileMap map[uint64]string) error {
	// Get firmware file info
	fileInfo, err := e.parser.GetReader().GetFileInfo()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	
	// Get magic pattern info
	magicOffset := e.parser.GetMagicOffset()
	
	// Get hardware pointers
	_, fs4HwPointers, err := e.parser.GetHWPointersRaw()
	if err != nil {
		e.logger.Warn("Failed to get HW pointers", zap.Error(err))
	}
	
	// Get ITOC/DTOC raw data
	itocRawData, err := e.parser.GetITOCRawData()
	if err != nil {
		e.logger.Warn("Failed to get ITOC raw data", zap.Error(err))
	}
	
	dtocRawData, err := e.parser.GetDTOCRawData()
	if err != nil {
		e.logger.Warn("Failed to get DTOC raw data", zap.Error(err))
	}
	
	// Build structured metadata
	metadata := &extracted.FirmwareMetadata{
		Format: "FS4",
		Firmware: extracted.FirmwareFileInfo{
			OriginalSize: uint64(fileInfo.Size),
			SHA256Hash:   fileInfo.SHA256,
		},
		Magic: extracted.MagicInfo{
			Offset: magicOffset,
			Value:  fmt.Sprintf("0x%016x", types.MagicPattern),
		},
		HWPointers: extracted.HWPointersInfo{
			Offset: magicOffset + types.HWPointersOffsetFromMagic,
		},
		ITOC: extracted.TOCInfo{
			Address:     e.parser.GetITOCAddress(),
			HeaderValid: e.parser.IsITOCHeaderValid(),
			RawHeader:   base64.StdEncoding.EncodeToString(itocRawData),
		},
		DTOC: extracted.TOCInfo{
			Address:     e.parser.GetDTOCAddress(),
			HeaderValid: e.parser.IsDTOCHeaderValid(),
			RawHeader:   base64.StdEncoding.EncodeToString(dtocRawData),
		},
		IsEncrypted: e.parser.IsEncrypted(),
	}
	
	// Set HW pointers based on type
	if fs4HwPointers != nil {
		metadata.HWPointers.FS4 = fs4HwPointers
	}
	
	// Add detailed section information
	sections := e.parser.GetSections()
	sectionMetadata := make([]extracted.SectionMetadata, 0)
	
	// Track memory layout
	memoryLayout := make([]extracted.MemorySegment, 0)
	lastEnd := uint64(0)
	
	// Get all sections sorted by offset
	allSections := make([]interfaces.CompleteSectionInterface, 0)
	for _, sectionList := range sections {
		allSections = append(allSections, sectionList...)
	}
	
	// Sort sections by offset for memory layout analysis
	sort.Slice(allSections, func(i, j int) bool {
		return allSections[i].Offset() < allSections[j].Offset()
	})
	
	// Build section summary and memory layout
	for i, section := range allSections {
		// Check for gap before this section
		if section.Offset() > lastEnd {
			gapSize := section.Offset() - lastEnd
			memoryLayout = append(memoryLayout, extracted.MemorySegment{
				StartOffset: lastEnd,
				EndOffset:   section.Offset(),
				Size:        gapSize,
				Type:        "gap",
				Description: fmt.Sprintf("Gap between sections"),
			})
		}
		
		// Get original CRC info if section had CRC
		var originalCRC uint32
		var originalSize uint32 = section.Size()
		if section.CRCType() == types.CRCInSection {
			// For non-encrypted firmwares, CRC was removed
			// For encrypted firmwares, CRC is kept in the data
			if !e.parser.IsEncrypted() {
				originalSize += 4 // CRC was removed
			}
			// Try to get CRC from raw data if available
			if rawData := section.GetRawData(); rawData != nil && len(rawData) >= 4 {
				// For IN_SECTION CRCs, the format is:
				// - 16-bit CRC in upper 16 bits (bytes 0-1) of last 4 bytes
				// - Lower 16 bits are 0 (bytes 2-3)
				// Extract the 16-bit CRC from the first 2 bytes of the last dword
				originalCRC = uint32(binary.BigEndian.Uint16(rawData[len(rawData)-4:]))
				e.logger.Debug("Extracted CRC from section",
					zap.String("type", section.TypeName()),
					zap.Uint32("crc", originalCRC),
					zap.String("crcHex", fmt.Sprintf("0x%08x", originalCRC)))
			}
		}
		
		// Create section metadata
		secMeta := extracted.SectionMetadata{
			BaseSection:  &interfaces.BaseSection{
				SectionType:     types.SectionType(section.Type()),
				SectionOffset:   section.Offset(),
				SectionSize:     section.Size(),
				SectionCRCType:  section.CRCType(),
				EncryptedFlag:   section.IsEncrypted(),
				DeviceDataFlag:  section.IsDeviceData(),
				FromHWPointerFlag: section.IsFromHWPointer(),
				HasRawData:      false, // Will be set based on parsing status
			},
			OriginalSize: originalSize,
			CRCValue:     originalCRC,
			Index:        i,
		}
		
		// Add filename if available
		if fileName, ok := sectionFileMap[section.Offset()]; ok {
			secMeta.FileName = fileName
		}
		
		sectionMetadata = append(sectionMetadata, secMeta)
		
		// Add section to memory layout
		memoryLayout = append(memoryLayout, extracted.MemorySegment{
			StartOffset: section.Offset(),
			EndOffset:   section.Offset() + uint64(originalSize),
			Size:        uint64(originalSize),
			Type:        "section",
			Description: section.TypeName(),
		})
		
		lastEnd = section.Offset() + uint64(originalSize)
	}
	
	// Check for trailing gap
	if fileInfo.Size > 0 && lastEnd < uint64(fileInfo.Size) {
		memoryLayout = append(memoryLayout, extracted.MemorySegment{
			StartOffset: lastEnd,
			EndOffset:   uint64(fileInfo.Size),
			Size:        uint64(fileInfo.Size) - lastEnd,
			Type:        "gap",
			Description: "Trailing gap",
		})
	}
	
	metadata.Sections = sectionMetadata
	metadata.MemoryLayout = memoryLayout
	
	// Add CRC algorithm info
	metadata.CRCInfo = extracted.CRCInfo{
		Algorithms: struct {
			Hardware string `json:"hardware"`
			Software string `json:"software"`
		}{
			Hardware: "CRC16-ARC (calc_hw_crc)",
			Software: "CRC16-XMODEM (polynomial 0x100b)",
		},
		Note: "CRC algorithm is determined by section type - HW_PTR uses hardware CRC, all others use software CRC",
	}
	
	// Add firmware boundaries
	metadata.Boundaries = extracted.FirmwareBoundaries{
		ImageStart: 0,
		ImageEnd:   uint64(fileInfo.Size),
		SectorSize: types.SectionAlignmentSector,
	}
	
	// Add gaps information
	metadata.Gaps = e.buildGapInfo(memoryLayout)
	
	// Write metadata
	metadataPath := filepath.Join(e.options.OutputDir, "firmware_metadata.json")
	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	
	if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}
	
	return nil
}

func (e *Extractor) buildGapInfo(memoryLayout []extracted.MemorySegment) []extracted.GapInfo {
	gaps := make([]extracted.GapInfo, 0)
	gapIndex := 0
	gapsDir := filepath.Join(e.options.OutputDir, "gaps")
	
	for _, segment := range memoryLayout {
		if segment.Type == "gap" {
			// Check if this gap was saved as uniform by looking for .meta file
			isUniform := false
			fillByte := uint8(0xFF) // Default
			
			// Check if a .meta file exists for this gap
			gapFileName := fmt.Sprintf("gap_%03d_0x%08x_0x%08x", gapIndex, segment.StartOffset, segment.EndOffset)
			if segment.EndOffset == 0 || segment.Description == "Trailing gap" {
				gapFileName = fmt.Sprintf("gap_%03d_0x%08x_EOF", gapIndex, segment.StartOffset)
			}
			metaPath := filepath.Join(gapsDir, gapFileName + ".meta")
			
			if metaData, err := os.ReadFile(metaPath); err == nil {
				// Parse metadata to get actual values
				lines := strings.Split(string(metaData), "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "fill=") {
						var fill uint32
						fmt.Sscanf(line, "fill=0x%02x", &fill)
						fillByte = byte(fill)
						isUniform = true
					}
				}
			}
			
			gaps = append(gaps, extracted.GapInfo{
				Index:       gapIndex,
				StartOffset: segment.StartOffset,
				EndOffset:   segment.EndOffset,
				Size:        segment.Size,
				FillByte:    fillByte,
				IsUniform:   isUniform,
			})
			gapIndex++
		}
	}
	
	return gaps
}