package extract

import (
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

	// Get new sections
	sections := e.parser.GetSectionsNew()
	
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

func (e *Extractor) extractSections(sections map[uint16][]interfaces.SectionInterface) (map[uint64]string, error) {
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
			
			// Store filename in map
			sectionFileMap[section.Offset()] = fileName
			
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
			jsonExported := false
			var jsonData []byte
			var err error
			if jsonData, err = section.MarshalJSON(); err == nil {
				// Pretty print JSON
				var prettyJSON map[string]interface{}
				if err := json.Unmarshal(jsonData, &prettyJSON); err == nil {
					if prettyData, err := json.MarshalIndent(prettyJSON, "", "  "); err == nil {
						if err := os.WriteFile(jsonPath, prettyData, 0644); err != nil {
							e.logger.Warn("Failed to write JSON",
								zap.String("file", jsonPath),
								zap.Error(err))
						} else {
							e.logger.Info("Exported JSON data",
								zap.String("file", jsonPath))
							jsonExported = true
						}
					}
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
			// 4. Section doesn't have meaningful parsed data (only has basic metadata)
			shouldWriteBinary := e.options.KeepBinary || 
				e.parser.IsEncrypted() || 
				!jsonExported
			
			// Check if JSON has meaningful parsed data beyond basic metadata
			if jsonExported && !shouldWriteBinary {
				// Check if JSON contains only basic metadata fields
				var jsonCheck map[string]interface{}
				if err := json.Unmarshal(jsonData, &jsonCheck); err == nil {
					// Count non-metadata fields
					parsedFields := 0
					hasRawDataFlag := false
					for key, value := range jsonCheck {
						switch key {
						case "type", "type_name", "offset", "size", "crc_type", 
						     "is_encrypted", "is_device_data", "data_size":
							// These are basic metadata fields
						case "has_raw_data":
							// Check if has_raw_data flag is set
							if v, ok := value.(bool); ok && v {
								hasRawDataFlag = true
							}
						default:
							parsedFields++
						}
					}
					// If only metadata fields exist or has_raw_data flag is set, we need the binary
					if parsedFields == 0 || hasRawDataFlag {
						shouldWriteBinary = true
						e.logger.Debug("Section has no parsed fields or has raw data flag, keeping binary",
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
	sections := e.parser.GetSectionsNew()
	allSections := make([]interfaces.SectionInterface, 0)
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
	hwPointersData, hwPointers, err := e.parser.GetHWPointersRaw()
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
	
	metadata := map[string]interface{}{
		"format": "FS4",
		"firmware_file": map[string]interface{}{
			"original_size": fileInfo.Size,
			"sha256_hash": fileInfo.SHA256,
		},
		"magic_pattern": map[string]interface{}{
			"offset": magicOffset,
			"pattern": fmt.Sprintf("0x%016x", types.MagicPattern),
			"search_offsets": types.MagicSearchOffsets,
		},
		"hw_pointers": map[string]interface{}{
			"offset": magicOffset + types.HWPointersOffsetFromMagic,
			"raw_data": hwPointersData,
			"parsed": hwPointers,
		},
		"itoc": map[string]interface{}{
			"address": e.parser.GetITOCAddress(),
			"header_valid": e.parser.IsITOCHeaderValid(),
			"raw_header": itocRawData,
		},
		"dtoc": map[string]interface{}{
			"address": e.parser.GetDTOCAddress(),
			"header_valid": e.parser.IsDTOCHeaderValid(),
			"raw_header": dtocRawData,
		},
		"is_encrypted": e.parser.IsEncrypted(),
	}
	
	// Add detailed section information
	sections := e.parser.GetSectionsNew()
	sectionSummary := make([]map[string]interface{}, 0)
	
	// Track memory layout
	memoryLayout := make([]map[string]interface{}, 0)
	lastEnd := uint64(0)
	
	// Get all sections sorted by offset
	allSections := make([]interfaces.SectionInterface, 0)
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
			memoryLayout = append(memoryLayout, map[string]interface{}{
				"type": "gap",
				"start": lastEnd,
				"end": section.Offset(),
				"size": gapSize,
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
				// - 16-bit CRC in upper 16 bits (bytes 0-1) 
				// - Lower 16 bits are 0 (bytes 2-3)
				originalCRC = uint32(rawData[len(rawData)-4])<<8 | uint32(rawData[len(rawData)-3])
				e.logger.Debug("Extracted CRC from section",
					zap.String("type", section.TypeName()),
					zap.Uint32("crc", originalCRC),
					zap.String("crcHex", fmt.Sprintf("0x%08x", originalCRC)))
			}
		}
		
		summary := map[string]interface{}{
			"type": section.Type(),
			"type_name": section.TypeName(),
			"offset": section.Offset(),
			"size": section.Size(),
			"original_size_with_crc": originalSize,
			"crc_type": section.CRCType().String(),
			"crc_value": originalCRC,
			"is_encrypted": section.IsEncrypted(),
			"is_device_data": section.IsDeviceData(),
			"is_from_hw_pointer": section.IsFromHWPointer(),
			"index": i,
		}
		
		// Add filename if available
		if fileName, ok := sectionFileMap[section.Offset()]; ok {
			summary["file_name"] = fileName
		}
		
		// Add ITOC entry info if available
		if entry := section.GetITOCEntry(); entry != nil {
			summary["itoc_entry"] = map[string]interface{}{
				"type": entry.Type,
				"flash_addr": entry.GetFlashAddr(),
				"section_crc": entry.SectionCRC,
				"encrypted": entry.Encrypted,
			}
		}
		
		sectionSummary = append(sectionSummary, summary)
		
		// Add section to memory layout
		memoryLayout = append(memoryLayout, map[string]interface{}{
			"type": "section",
			"section_type": section.TypeName(),
			"start": section.Offset(),
			"end": section.Offset() + uint64(originalSize),
			"size": originalSize,
		})
		
		lastEnd = section.Offset() + uint64(originalSize)
	}
	
	// Check for trailing gap
	if fileInfo.Size > 0 && lastEnd < uint64(fileInfo.Size) {
		memoryLayout = append(memoryLayout, map[string]interface{}{
			"type": "gap",
			"start": lastEnd,
			"end": fileInfo.Size,
			"size": uint64(fileInfo.Size) - lastEnd,
		})
	}
	
	metadata["sections"] = sectionSummary
	metadata["memory_layout"] = memoryLayout
	
	// Add CRC algorithm info
	metadata["crc_info"] = map[string]interface{}{
		"algorithms": map[string]string{
			"hardware": "CRC16-ARC (calc_hw_crc)",
			"software": "CRC16-XMODEM (polynomial 0x100b)",
		},
		"note": "CRC algorithm is determined by section type - HW_PTR uses hardware CRC, all others use software CRC",
	}
	
	// Add firmware boundaries
	metadata["boundaries"] = map[string]interface{}{
		"image_start": 0,
		"image_end": fileInfo.Size,
		"sector_size": types.SectionAlignmentSector,
	}
	
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