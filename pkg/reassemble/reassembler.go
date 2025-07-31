package reassemble

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"github.com/Civil/mlx5fw-go/pkg/annotations"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/types"
)

// Options contains options for firmware reassembly
type Options struct {
	InputDir      string
	OutputFile    string
	VerifyCRC     bool
	BinaryOnly    bool  // Force binary-only mode, ignore JSON files
}

// Reassembler handles firmware reassembly
type Reassembler struct {
	logger  *zap.Logger
	options Options
}

// New creates a new Reassembler
func New(logger *zap.Logger, opts Options) *Reassembler {
	return &Reassembler{
		logger:  logger,
		options: opts,
	}
}

// Reassemble performs the firmware reassembly
func (r *Reassembler) Reassemble() error {
	r.logger.Info("Starting reassemble command",
		zap.String("inputDir", r.options.InputDir),
		zap.String("outputFile", r.options.OutputFile),
		zap.Bool("verifyCRC", r.options.VerifyCRC))

	// Load metadata
	metadataPath := filepath.Join(r.options.InputDir, "firmware_metadata.json")
	metadata, err := r.loadMetadata(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	// Verify we have all required files
	if err := r.verifyRequiredFiles(metadata); err != nil {
		return fmt.Errorf("missing required files: %w", err)
	}

	// Create output file
	outputFile, err := os.Create(r.options.OutputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Reassemble firmware
	if err := r.reassembleFirmware(outputFile, metadata); err != nil {
		outputFile.Close()
		os.Remove(r.options.OutputFile)
		return fmt.Errorf("failed to reassemble firmware: %w", err)
	}

	r.logger.Info("Firmware reassembled successfully",
		zap.String("outputFile", r.options.OutputFile))
	return nil
}

func (r *Reassembler) loadMetadata(path string) (*FirmwareMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata FirmwareMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return &metadata, nil
}

func (r *Reassembler) verifyRequiredFiles(metadata *FirmwareMetadata) error {
	// Build section filename map
	fileMap := r.buildSectionFileMap(metadata.Sections)
	
	// Check for section files
	for _, section := range metadata.Sections {
		// Skip zero-size sections - they weren't extracted
		if section.Size == 0 {
			continue
		}
		
		fileName := r.getSectionFileName(section, fileMap)
		filePath := filepath.Join(r.options.InputDir, fileName)
		
		// Check for either binary or JSON file
		hasBinary := false
		hasJSON := false
		
		if _, err := os.Stat(filePath); err == nil {
			hasBinary = true
		}
		
		jsonFileName := strings.TrimSuffix(fileName, ".bin") + ".json"
		jsonPath := filepath.Join(r.options.InputDir, jsonFileName)
		if _, err := os.Stat(jsonPath); err == nil {
			hasJSON = true
		}
		
		if !hasBinary && !hasJSON {
			return fmt.Errorf("missing section file: %s (neither .bin nor .json found)", fileName)
		}
	}

	// Check for gap files
	gapsDir := filepath.Join(r.options.InputDir, "gaps")
	if _, err := os.Stat(gapsDir); err != nil {
		return fmt.Errorf("missing gaps directory")
	}

	// Count expected gaps from memory layout
	expectedGaps := 0
	for _, segment := range metadata.MemoryLayout {
		if segment.Type == "gap" {
			expectedGaps++
		}
	}

	// Verify gap files exist (either .bin or .meta)
	// Note: there can be multiple files per gap index if a gap contains multiple sections
	gapBinFiles, err := filepath.Glob(filepath.Join(gapsDir, "gap_*.bin"))
	if err != nil {
		return fmt.Errorf("failed to list gap bin files: %w", err)
	}
	
	gapMetaFiles, err := filepath.Glob(filepath.Join(gapsDir, "gap_*.meta"))
	if err != nil {
		return fmt.Errorf("failed to list gap meta files: %w", err)
	}
	
	// Count unique gap indices
	uniqueGapIndices := make(map[int]bool)
	for _, file := range append(gapBinFiles, gapMetaFiles...) {
		base := filepath.Base(file)
		var gapIndex int
		if _, err := fmt.Sscanf(base, "gap_%03d_", &gapIndex); err == nil {
			uniqueGapIndices[gapIndex] = true
		}
	}
	
	if len(uniqueGapIndices) != expectedGaps {
		return fmt.Errorf("gap count mismatch: expected %d gaps, found %d unique gap indices (from %d files: %d bin + %d meta)", 
			expectedGaps, len(uniqueGapIndices), len(gapBinFiles)+len(gapMetaFiles), len(gapBinFiles), len(gapMetaFiles))
	}

	return nil
}

func (r *Reassembler) buildSectionFileMap(sections []SectionMetadata) map[string]string {
	// Group sections by type name, excluding zero-size sections
	sectionsByType := make(map[string][]SectionMetadata)
	for _, section := range sections {
		// Skip zero-size sections
		if section.Size == 0 {
			continue
		}
		sectionsByType[section.TypeName] = append(sectionsByType[section.TypeName], section)
	}
	
	// Build filename map
	fileMap := make(map[string]string)
	for typeName, sectionList := range sectionsByType {
		// Clean up the type name for filename
		fileName := strings.ReplaceAll(typeName, " ", "_")
		fileName = strings.ReplaceAll(fileName, "/", "_")
		
		// Generate filenames based on whether there are multiple sections of this type
		if len(sectionList) > 1 {
			// Multiple sections - add index
			for i, section := range sectionList {
				key := fmt.Sprintf("%s_%d", typeName, section.Offset)
				fileMap[key] = fmt.Sprintf("%s_%d_0x%08x.bin", fileName, i, section.Offset)
			}
		} else {
			// Single section - no index
			section := sectionList[0]
			key := fmt.Sprintf("%s_%d", typeName, section.Offset)
			fileMap[key] = fmt.Sprintf("%s_0x%08x.bin", fileName, section.Offset)
		}
	}
	
	return fileMap
}

func (r *Reassembler) getSectionFileName(section SectionMetadata, fileMap map[string]string) string {
	// If the section has a filename in metadata, use it
	if section.FileName != "" {
		return section.FileName
	}
	
	// Otherwise, fall back to the generated filename
	key := fmt.Sprintf("%s_%d", section.TypeName, section.Offset)
	if fileName, ok := fileMap[key]; ok {
		return fileName
	}
	// Fallback if not found in map
	fileName := strings.ReplaceAll(section.TypeName, " ", "_")
	fileName = strings.ReplaceAll(fileName, "/", "_")
	return fmt.Sprintf("%s_0x%08x.bin", fileName, section.Offset)
}

func (r *Reassembler) reassembleFirmware(output io.WriteSeeker, metadata *FirmwareMetadata) error {
	// Initialize CRC calculator
	crcCalc := parser.NewCRCCalculator()

	// Build section filename map
	fileMap := r.buildSectionFileMap(metadata.Sections)

	// Pre-allocate firmware buffer
	firmwareSize := metadata.Firmware.OriginalSize
	firmwareData := make([]byte, firmwareSize)

	// Fill with 0xFF (standard padding)
	for i := range firmwareData {
		firmwareData[i] = 0xFF
	}

	// Write magic pattern at the correct offset
	if err := r.writeMagicPattern(firmwareData, metadata.Magic.Offset); err != nil {
		return fmt.Errorf("failed to write magic pattern: %w", err)
	}

	// Write hardware pointers
	if err := r.writeHWPointers(firmwareData, metadata.HWPointers); err != nil {
		return fmt.Errorf("failed to write HW pointers: %w", err)
	}

	// Process gaps FIRST (before sections)
	// This ensures that gap data doesn't overwrite section CRCs
	if err := r.reassembleGaps(firmwareData, metadata); err != nil {
		return fmt.Errorf("failed to reassemble gaps: %w", err)
	}

	// Process sections (will overwrite gap data with section data + CRC)
	for _, section := range metadata.Sections {
		// Skip zero-size sections
		if section.Size == 0 {
			continue
		}

		// Read section data - prefer JSON over binary unless BinaryOnly mode
		sectionData, err := r.readSectionData(section, fileMap)
		if err != nil {
			return fmt.Errorf("failed to read section data: %w", err)
		}

		// Add CRC if needed (only for non-encrypted firmwares)
		if section.CRCType == "IN_SECTION" && section.OriginalSize > section.Size && !metadata.IsEncrypted {
			r.logger.Info("Adding IN_SECTION CRC",
				zap.String("section", section.TypeName),
				zap.Uint32("originalSize", section.OriginalSize),
				zap.Uint32("size", section.Size))
			
			// Determine whether to use blank CRC based on section type
			// Some sections use 0xFFFFFFFF as a placeholder CRC value
			var crcBytes []byte
			
			// Check if this is a section type that typically has blank CRCs
			// Based on observation: TOOLS_AREA, BOOT2, HASHES_TABLE, DEV_INFO, and device data sections
			if section.TypeName == "TOOLS_AREA" || section.TypeName == "BOOT2" || 
			   section.TypeName == "HASHES_TABLE" || section.TypeName == "DEV_INFO" || 
			   section.TypeName == "MFG_INFO" || section.TypeName == "IMAGE_INFO" ||
			   section.TypeName == "FORBIDDEN_VERSIONS" || section.TypeName == "PUBLIC_KEYS_2048" ||
			   section.TypeName == "PUBLIC_KEYS_4096" || section.TypeName == "IMAGE_SIGNATURE_512" ||
			   strings.HasPrefix(section.TypeName, "UNKNOWN_0xE0") {
				// Use blank CRC for these sections
				crcBytes = []byte{0xFF, 0xFF, 0xFF, 0xFF}
				r.logger.Info("Using blank CRC for section",
					zap.String("section", section.TypeName))
			} else {
				// Calculate CRC
				var crc uint16
				if r.isHardwareCRCSection(section) {
					crc = crcCalc.CalculateHardwareCRC(sectionData)
					r.logger.Info("Calculated hardware CRC",
						zap.String("section", section.TypeName),
						zap.Uint16("crc", crc),
						zap.Int("dataLen", len(sectionData)))
				} else {
					crc = crcCalc.CalculateSoftwareCRC16(sectionData)
					r.logger.Info("Calculated software CRC",
						zap.String("section", section.TypeName),
						zap.Uint16("crc", crc),
						zap.Int("dataLen", len(sectionData)))
				}

				// Append CRC (16-bit CRC in upper 16 bits of 32-bit word, big-endian)
				crcBytes = make([]byte, 4)
				// Create a simple struct to handle the CRC format
				crcStruct := struct {
					CRC      uint16 `offset:"byte:0,endian:be"`
					Reserved uint16 `offset:"byte:2,endian:be"`
				}{CRC: crc, Reserved: 0}
				crcData, _ := annotations.MarshalStruct(&crcStruct)
				crcBytes = crcData
			}
			sectionData = append(sectionData, crcBytes...)

			r.logger.Debug("Added CRC to section",
				zap.String("section", section.TypeName),
				zap.String("crcBytes", fmt.Sprintf("%02x%02x%02x%02x", 
					crcBytes[0], crcBytes[1], crcBytes[2], crcBytes[3])))
		} else if section.CRCType == "IN_SECTION" && metadata.IsEncrypted {
			r.logger.Debug("Keeping section data intact for encrypted firmware",
				zap.String("section", section.TypeName),
				zap.Int("dataSize", len(sectionData)))
		}

		// Write section data
		copy(firmwareData[section.Offset:], sectionData)
	}

	// Write ITOC and DTOC headers if they exist
	if err := r.writeTOCHeaders(firmwareData, metadata); err != nil {
		return fmt.Errorf("failed to write TOC headers: %w", err)
	}

	// Update CRCs in HW pointers
	if err := r.updateHWPointerCRCs(firmwareData, metadata, crcCalc); err != nil {
		return fmt.Errorf("failed to update HW pointer CRCs: %w", err)
	}

	// Write the complete firmware
	if _, err := output.Write(firmwareData); err != nil {
		return fmt.Errorf("failed to write firmware data: %w", err)
	}

	return nil
}

func (r *Reassembler) reassembleGaps(firmwareData []byte, metadata *FirmwareMetadata) error {
	gapsDir := filepath.Join(r.options.InputDir, "gaps")
	
	// Check if gaps directory exists
	if _, err := os.Stat(gapsDir); os.IsNotExist(err) {
		// No gaps to process
		return nil
	}
	
	// Process memory layout to identify gaps
	gapIndex := 0
	for _, segment := range metadata.MemoryLayout {
		if segment.Type != "gap" {
			continue
		}
		
		start := segment.Start
		end := segment.End
		
		// Check for metadata file first
		gapFileName := fmt.Sprintf("gap_%03d_", gapIndex)
		metaFiles, err := filepath.Glob(filepath.Join(gapsDir, gapFileName+"*.meta"))
		if err == nil && len(metaFiles) > 0 {
			// Read metadata file
			metaData, err := os.ReadFile(metaFiles[0])
			if err != nil {
				return fmt.Errorf("failed to read gap metadata %s: %w", metaFiles[0], err)
			}
			
			// Parse metadata
			var size uint64
			var fillByte byte
			lines := strings.Split(string(metaData), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "size=") {
					fmt.Sscanf(line, "size=%d", &size)
				} else if strings.HasPrefix(line, "fill=") {
					var fill uint32
					fmt.Sscanf(line, "fill=0x%02x", &fill)
					fillByte = byte(fill)
				}
			}
			
			// Fill the gap with the specified byte
			for i := start; i < end && i < uint64(len(firmwareData)); i++ {
				firmwareData[i] = fillByte
			}
			r.logger.Debug("Filled uniform gap from metadata",
				zap.Int("index", gapIndex),
				zap.Uint64("start", start),
				zap.Uint64("size", size),
				zap.Uint8("fillByte", fillByte))
		} else {
			// Try regular bin file
			gapFiles, err := filepath.Glob(filepath.Join(gapsDir, gapFileName+"*.bin"))
			if err != nil || len(gapFiles) == 0 {
				r.logger.Warn("Gap file not found",
					zap.Int("index", gapIndex),
					zap.Uint64("start", start),
					zap.Uint64("end", end))
				gapIndex++
				continue
			}
			
			gapData, err := os.ReadFile(gapFiles[0])
			if err != nil {
				return fmt.Errorf("failed to read gap file %s: %w", gapFiles[0], err)
			}
			
			// Copy gap data to firmware
			copy(firmwareData[start:end], gapData)
		}
		gapIndex++
	}
	
	return nil
}

func (r *Reassembler) writeMagicPattern(data []byte, offset uint32) error {
	if int(offset)+8 > len(data) {
		return fmt.Errorf("magic pattern offset out of bounds")
	}
	// Use MagicPatternStruct to write the magic pattern
	magicStruct := &types.MagicPatternStruct{
		Magic: types.MagicPattern,
	}
	magicData, err := magicStruct.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal magic pattern: %w", err)
	}
	copy(data[offset:], magicData)
	return nil
}

func (r *Reassembler) writeHWPointers(data []byte, hwPointers HWPointersInfo) error {
	// Decode base64 data
	decodedData, err := base64.StdEncoding.DecodeString(hwPointers.RawData)
	if err != nil {
		return fmt.Errorf("failed to decode HW pointers data: %w", err)
	}
	
	if int(hwPointers.Offset)+len(decodedData) > len(data) {
		return fmt.Errorf("HW pointers offset out of bounds")
	}
	copy(data[hwPointers.Offset:], decodedData)
	return nil
}

func (r *Reassembler) writeTOCHeaders(data []byte, metadata *FirmwareMetadata) error {
	// Write ITOC header if valid
	if metadata.ITOC.HeaderValid && len(metadata.ITOC.RawHeader) > 0 {
		decodedITOC, err := base64.StdEncoding.DecodeString(metadata.ITOC.RawHeader)
		if err != nil {
			return fmt.Errorf("failed to decode ITOC header: %w", err)
		}
		copy(data[metadata.ITOC.Address:], decodedITOC)
	}
	
	// Write DTOC header if valid
	if metadata.DTOC.HeaderValid && len(metadata.DTOC.RawHeader) > 0 {
		decodedDTOC, err := base64.StdEncoding.DecodeString(metadata.DTOC.RawHeader)
		if err != nil {
			return fmt.Errorf("failed to decode DTOC header: %w", err)
		}
		copy(data[metadata.DTOC.Address:], decodedDTOC)
	}
	
	return nil
}

func (r *Reassembler) updateHWPointerCRCs(data []byte, metadata *FirmwareMetadata, crcCalc *parser.CRCCalculator) error {
	hwPointersOffset := metadata.HWPointers.Offset
	
	// Process each HW pointer entry (16 entries, 8 bytes each)
	// The CRC for each entry is calculated on the first 6 bytes of the 8-byte entry
	for i := 0; i < 16; i++ {
		entryOffset := hwPointersOffset + uint32(i*8)
		
		// Get pointer value (first 4 bytes)
		if entryOffset+8 > uint32(len(data)) {
			break
		}
		
		// Unmarshal the HW pointer entry WITH reserved field
		// We need the reserved field for proper CRC calculation
		entry := &types.HWPointerEntry{}
		if err := entry.UnmarshalWithReserved(data[entryOffset:entryOffset+8]); err != nil {
			continue // Skip invalid entries
		}
		
		// Skip if pointer is 0 or 0xFFFFFFFF
		if entry.Ptr == 0 || entry.Ptr == 0xFFFFFFFF {
			continue
		}
		
		// Calculate CRC on 6 bytes of the pointer entry (ptr + 2 bytes of next field)
		crc := r.calcHWPointerCRC(data[entryOffset:entryOffset+8], 6)
		
		// Update CRC in the entry
		entry.CRC = crc
		
		// Marshal back to update the data WITH reserved field
		entryData, err := entry.MarshalWithReserved()
		if err != nil {
			continue
		}
		copy(data[entryOffset:entryOffset+8], entryData)
	}
	
	return nil
}

func (r *Reassembler) calcHWPointerCRC(data []byte, size int) uint16 {
	crc := uint16(0xffff)
	for i := 0; i < size; i++ {
		var d byte
		if i > 1 {
			d = data[i]
		} else {
			d = ^data[i]
		}
		tableIndex := (crc ^ uint16(d)) & 0xff
		crc = (crc >> 8) ^ types.CRC16Table2[tableIndex]
	}
	// Swap bytes
	crc = ((crc << 8) & 0xff00) | ((crc >> 8) & 0xff)
	return crc
}

func (r *Reassembler) isHardwareCRCSection(section SectionMetadata) bool {
	// Use the single source of truth for CRC algorithm determination
	return types.GetSectionCRCAlgorithm(section.Type) == types.CRCAlgorithmHardware
}

// readSectionData reads section data from JSON and binary files as needed
func (r *Reassembler) readSectionData(section SectionMetadata, fileMap map[string]string) ([]byte, error) {
	sectionFileName := r.getSectionFileName(section, fileMap)
	jsonFileName := strings.TrimSuffix(sectionFileName, ".bin") + ".json"
	jsonPath := filepath.Join(r.options.InputDir, jsonFileName)
	
	// Always read JSON file first (it should always exist)
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file %s: %w", jsonFileName, err)
	}
	
	// Parse JSON to check for raw data flag
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		return nil, fmt.Errorf("failed to parse JSON %s: %w", jsonFileName, err)
	}
	
	// Check if this section has raw data flag
	hasRawData := false
	if rawDataFlag, ok := jsonMap["has_raw_data"].(bool); ok {
		hasRawData = rawDataFlag
	}
	
	// If binary-only mode is set, always try to read binary
	if r.options.BinaryOnly || hasRawData {
		// Read binary file
		binaryPath := filepath.Join(r.options.InputDir, sectionFileName)
		sectionData, err := os.ReadFile(binaryPath)
		if err != nil {
			if hasRawData {
				return nil, fmt.Errorf("section has raw data flag but binary file not found: %s", sectionFileName)
			}
			return nil, fmt.Errorf("failed to read binary file %s: %w", sectionFileName, err)
		}
		
		r.logger.Debug("Read section from binary",
			zap.String("section", section.TypeName),
			zap.String("file", sectionFileName),
			zap.Bool("has_raw_data", hasRawData))
		
		return sectionData, nil
	}
	
	// Try to reconstruct from JSON
	r.logger.Debug("Attempting JSON reconstruction",
		zap.String("section", section.TypeName),
		zap.String("json_file", jsonFileName))
	
	reconstructed, err := r.reconstructFromJSON(section, jsonData)
	if err != nil {
		// If reconstruction fails and binary file exists, use it
		binaryPath := filepath.Join(r.options.InputDir, sectionFileName)
		if _, statErr := os.Stat(binaryPath); statErr == nil {
			sectionData, readErr := os.ReadFile(binaryPath)
			if readErr == nil {
				r.logger.Debug("Using binary file after failed JSON reconstruction",
					zap.String("section", section.TypeName),
					zap.String("file", sectionFileName))
				return sectionData, nil
			}
		}
		return nil, fmt.Errorf("failed to reconstruct from JSON: %w", err)
	}
	
	r.logger.Info("Reconstructed section from JSON",
		zap.String("section", section.TypeName),
		zap.String("file", jsonFileName))
	
	return reconstructed, nil
}

// reconstructFromJSON attempts to reconstruct section data from JSON
func (r *Reassembler) reconstructFromJSON(section SectionMetadata, jsonData []byte) ([]byte, error) {
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	// Check if this section has raw data flag (unparsed sections)
	if hasRawData, ok := jsonMap["has_raw_data"].(bool); ok && hasRawData {
		// This section requires binary file for reconstruction
		return nil, fmt.Errorf("section has raw data flag, binary file required")
	}
	
	// Check if this is a section with parsed data that can be reconstructed
	// For now, we only handle specific section types that have parsers
	switch section.Type {
	case types.SectionTypeImageInfo:
		// IMAGE_INFO can be reconstructed from its fields
		return r.reconstructImageInfo(jsonMap)
	case types.SectionTypeDevInfo:
		// DEV_INFO can be reconstructed from its fields
		return r.reconstructDevInfo(jsonMap, section.Size)
	case types.SectionTypeBoot2:
		// BOOT2 still needs binary file but we check for proper structure
		return nil, fmt.Errorf("BOOT2 section requires binary file")
	case types.SectionTypeMfgInfo:
		// MFG_INFO can be reconstructed from its fields
		return r.reconstructMfgInfo(jsonMap)
	case types.SectionTypeHashesTable:
		// HASHES_TABLE can be reconstructed from its fields
		return r.reconstructHashesTable(jsonMap)
	case types.SectionTypeToolsArea:
		// TOOLS_AREA needs binary data for TypeData and Reserved fields
		return nil, fmt.Errorf("TOOLS_AREA section requires binary file")
	case types.SectionTypeResetInfo:
		// RESET_INFO structure doesn't match firmware layout, needs binary data
		return nil, fmt.Errorf("RESET_INFO section requires binary file")
	case types.SectionTypeImageSignature256, types.SectionTypeImageSignature512:
		// Image signature sections can be reconstructed
		return r.reconstructImageSignature(jsonMap, section.Type, section.Size)
	case types.SectionTypePublicKeys2048, types.SectionTypePublicKeys4096:
		// Public keys sections can be reconstructed
		return r.reconstructPublicKeys(jsonMap, section.Type, section.Size)
	case types.SectionTypeForbiddenVersions:
		// Forbidden versions can be reconstructed
		return r.reconstructForbiddenVersions(jsonMap, section.Size)
	case types.SectionTypeDbgFWINI:
		// DBG_FW_INI needs binary data for bit-perfect reconstruction
		return nil, fmt.Errorf("DBG_FW_INI section requires binary file")
	case types.SectionTypeDbgFWParams:
		// DBG_FW_PARAMS still needs binary data
		return nil, fmt.Errorf("DBG_FW_PARAMS section requires binary file")
	case types.SectionTypeCRDumpMaskData:
		// CRDUMP sections need binary data
		return nil, fmt.Errorf("CRDUMP_MASK_DATA section requires binary file")
	case types.SectionTypeProgrammableHwFw1, types.SectionTypeProgrammableHwFw2:
		// Programmable HW FW sections need binary data
		return nil, fmt.Errorf("PROGRAMMABLE_HW_FW section requires binary file")
	case types.SectionTypeFwNvLog:
		// FW_NV_LOG sections need binary data
		return nil, fmt.Errorf("FW_NV_LOG section requires binary file")
	case types.SectionTypeNvData0, types.SectionTypeNvData1, types.SectionTypeNvData2:
		// NV_DATA sections need binary data
		return nil, fmt.Errorf("NV_DATA section requires binary file")
	case types.SectionTypeDigitalCertPtr:
		// DIGITAL_CERT_PTR needs binary data due to variable structure
		return nil, fmt.Errorf("DIGITAL_CERT_PTR section requires binary file")
	// Add more section types as we implement their reconstruction
	default:
		// For sections without specific reconstruction, we need the binary
		return nil, fmt.Errorf("section type %s cannot be reconstructed from JSON alone", section.TypeName)
	}
}

// Section-specific reconstruction methods are in separate files

func (r *Reassembler) reconstructImageSignature(jsonMap map[string]interface{}, sectionType uint16, sectionSize uint32) ([]byte, error) {
	// Extract the signature object
	sigData, ok := jsonMap["signature"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing signature field in JSON")
	}
	
	// Get signature type
	var sigType uint32
	if st, ok := sigData["signature_type"].(float64); ok {
		sigType = uint32(st)
	}
	
	// Get signature hex string
	sigHex, ok := sigData["signature"].(string)
	if !ok {
		return nil, fmt.Errorf("missing signature hex string")
	}
	
	signature, err := hex.DecodeString(sigHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature hex: %w", err)
	}
	
	// Create appropriate structure based section type
	var structData []byte
	
	if sectionType == types.SectionTypeImageSignature256 {
		sig := &types.ImageSignature{
			SignatureType: sigType,
		}
		copy(sig.Signature[:], signature)
		structData, err = sig.Marshal()
	} else if sectionType == types.SectionTypeImageSignature512 {
		sig := &types.ImageSignature2{
			SignatureType: sigType,
		}
		copy(sig.Signature[:], signature)
		structData, err = sig.Marshal()
	} else {
		return nil, fmt.Errorf("unsupported signature section type: %d", sectionType)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to marshal signature structure: %w", err)
	}
	
	// Handle padding
	if uint32(len(structData)) < sectionSize {
		// Check if padding data was provided in JSON (at the section level, not inside signature)
		if paddingHex, ok := jsonMap["padding"].(string); ok {
			paddingData, err := hex.DecodeString(paddingHex)
			if err != nil {
				return nil, fmt.Errorf("failed to decode padding hex: %w", err)
			}
			r.logger.Info("Using padding from JSON for signature section",
				zap.Int("structSize", len(structData)),
				zap.Uint32("sectionSize", sectionSize),
				zap.Int("paddingSize", len(paddingData)))
			structData = append(structData, paddingData...)
		} else {
			// Pad with zero bytes if no padding data provided
			padding := make([]byte, sectionSize - uint32(len(structData)))
			structData = append(structData, padding...)
		}
	}
	
	return structData, nil
}


func (r *Reassembler) reconstructPublicKeys(jsonMap map[string]interface{}, sectionType uint16, sectionSize uint32) ([]byte, error) {
	// Extract the keys array
	keysData, ok := jsonMap["keys"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("missing keys field in JSON")
	}
	
	// Determine entry size based on section type
	// PublicKey: 4 bytes reserved + 16 bytes UUID + 256 bytes key = 276 bytes
	// PublicKey2: 4 bytes reserved + 16 bytes UUID + 512 bytes key = 532 bytes
	entrySize := 276 // For PUBLIC_KEYS_2048
	if sectionType == types.SectionTypePublicKeys4096 {
		entrySize = 532 // For PUBLIC_KEYS_4096
	}
	
	// Create result buffer with section size to handle padding
	result := make([]byte, sectionSize)
	
	// Process each key
	for i, keyInterface := range keysData {
		keyData, ok := keyInterface.(map[string]interface{})
		if !ok {
			continue
		}
		
		// Get reserved field
		var reserved uint32
		if res, ok := keyData["reserved"].(float64); ok {
			reserved = uint32(res)
		}
		
		// Get UUID
		uuidHex, _ := keyData["uuid"].(string)
		uuid, _ := hex.DecodeString(uuidHex)
		
		// Get key
		keyHex, _ := keyData["key"].(string)
		key, _ := hex.DecodeString(keyHex)
		
		// Create key structure based on type
		if sectionType == types.SectionTypePublicKeys2048 {
			pk := &types.PublicKey{
				Reserved: reserved,
			}
			copy(pk.UUID[:], uuid)
			copy(pk.Key[:], key)
			pkData, _ := pk.Marshal()
			copy(result[i*entrySize:], pkData)
		} else {
			pk2 := &types.PublicKey2{
				Reserved: reserved,
			}
			copy(pk2.UUID[:], uuid)
			copy(pk2.Key[:], key)
			pk2Data, _ := pk2.Marshal()
			copy(result[i*entrySize:], pk2Data)
		}
	}
	
	return result, nil
}

func (r *Reassembler) reconstructForbiddenVersions(jsonMap map[string]interface{}, sectionSize uint32) ([]byte, error) {
	// Extract the forbidden_versions object
	forbiddenData, ok := jsonMap["forbidden_versions"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing forbidden_versions field in JSON")
	}
	
	// Calculate actual number of version slots based on section size
	// Section size = 8 (header) + numVersions * 4
	numVersionSlots := (sectionSize - 8) / 4
	
	// Create ForbiddenVersions structure
	fv := &types.ForbiddenVersions{
		Versions: make([]uint32, numVersionSlots),
	}
	
	// Get count
	if count, ok := forbiddenData["count"].(float64); ok {
		fv.Count = uint32(count)
	}
	
	// Get reserved
	if reserved, ok := forbiddenData["reserved"].(float64); ok {
		fv.Reserved = uint32(reserved)
	}
	
	// Get versions array
	if versionsData, ok := forbiddenData["versions"].([]interface{}); ok {
		for i, ver := range versionsData {
			if i >= int(numVersionSlots) {
				break
			}
			if version, ok := ver.(float64); ok {
				fv.Versions[i] = uint32(version)
			}
		}
	}
	
	// Marshal the structure to bytes with the expected section size
	return annotations.MarshalStructWithSize(fv, int(sectionSize))
}



func (r *Reassembler) reconstructDevInfo(jsonMap map[string]interface{}, sectionSize uint32) ([]byte, error) {
	// Extract the device_info object
	deviceInfoData, ok := jsonMap["device_info"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing device_info field in JSON")
	}
	
	// Create and populate DevInfo structure
	info := &types.DevInfoAnnotated{}
	
	// Populate signature fields
	if sig0, ok := deviceInfoData["signature0"].(float64); ok {
		info.Signature0 = uint32(sig0)
	}
	if sig1, ok := deviceInfoData["signature1"].(float64); ok {
		info.Signature1 = uint32(sig1)
	}
	if sig2, ok := deviceInfoData["signature2"].(float64); ok {
		info.Signature2 = uint32(sig2)
	}
	if sig3, ok := deviceInfoData["signature3"].(float64); ok {
		info.Signature3 = uint32(sig3)
	}
	
	// Version fields
	if minorVer, ok := deviceInfoData["minor_version"].(float64); ok {
		info.MinorVersion = uint8(minorVer)
	}
	if majorVer, ok := deviceInfoData["major_version"].(float64); ok {
		info.MajorVersion = uint16(majorVer)
	}
	
	// Reserved fields
	if reserved1, ok := deviceInfoData["reserved1"].(float64); ok {
		info.Reserved1 = uint8(reserved1)
	}
	if reserved2, ok := deviceInfoData["reserved2"].([]interface{}); ok {
		for i := 0; i < len(reserved2) && i < 12; i++ {
			if val, ok := reserved2[i].(float64); ok {
				info.Reserved2[i] = byte(val)
			}
		}
	}
	
	// GUID info
	if guidsData, ok := deviceInfoData["guids"].(map[string]interface{}); ok {
		if reserved1, ok := guidsData["reserved1"].(float64); ok {
			info.Guids.Reserved1 = uint16(reserved1)
		}
		if step, ok := guidsData["step"].(float64); ok {
			info.Guids.Step = uint8(step)
		}
		if numAlloc, ok := guidsData["num_allocated"].(float64); ok {
			info.Guids.NumAllocated = uint8(numAlloc)
		}
		if reserved2, ok := guidsData["reserved2"].(float64); ok {
			info.Guids.Reserved2 = uint32(reserved2)
		}
		if uid, ok := guidsData["uid"].(float64); ok {
			info.Guids.UID = uint64(uid)
		}
	}
	
	// MAC info
	if macsData, ok := deviceInfoData["macs"].(map[string]interface{}); ok {
		if reserved1, ok := macsData["reserved1"].(float64); ok {
			info.Macs.Reserved1 = uint16(reserved1)
		}
		if step, ok := macsData["step"].(float64); ok {
			info.Macs.Step = uint8(step)
		}
		if numAlloc, ok := macsData["num_allocated"].(float64); ok {
			info.Macs.NumAllocated = uint8(numAlloc)
		}
		if reserved2, ok := macsData["reserved2"].(float64); ok {
			info.Macs.Reserved2 = uint32(reserved2)
		}
		if uid, ok := macsData["uid"].(float64); ok {
			info.Macs.UID = uint64(uid)
		}
	}
	
	// Reserved3
	if reserved3, ok := deviceInfoData["reserved3"].([]interface{}); ok {
		for i := 0; i < len(reserved3) && i < 444; i++ {
			if val, ok := reserved3[i].(float64); ok {
				info.Reserved3[i] = byte(val)
			}
		}
	}
	
	// Reserved4 is usually 0 (padding before CRC)
	if reserved4, ok := deviceInfoData["reserved4"].(float64); ok {
		info.Reserved4 = uint16(reserved4)
	}
	
	// CRC will be recalculated when writing the section, not restored from JSON
	// The original_crc field is stored for reference only
	// Note: CRC at offset 0x1FE is part of the struct, not the trailing CRC
	
	// Clear the CRC field - it MUST be recalculated
	info.CRC = 0
	
	// Marshal the structure to binary with CRC=0
	data, err := info.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DEV_INFO: %w", err)
	}
	
	// DevInfo marshals to 512 bytes with Reserved4 at 508-509 and CRC at 510-511
	// CRC should be calculated on everything up to but not including Reserved4
	// With Reserved3[444], we need to calculate on first 508 bytes
	crcCalc := parser.NewCRCCalculator()
	crc := crcCalc.CalculateSoftwareCRC16(data[:508])
	
	r.logger.Info("DEV_INFO CRC calculation",
		zap.Uint16("calculatedCRC", crc),
		zap.Uint16("reserved4", info.Reserved4))
	
	// Set the CRC in the struct and remarshal
	info.CRC = crc
	data, err = info.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DEV_INFO with CRC: %w", err)
	}
	
	// The DevInfo struct marshals to 484 bytes, but the firmware section expects 512 bytes
	// The section has padding after the struct data
	// Return data without CRC (last 4 bytes of the 512-byte section)
	
	// Pad data to 512 bytes if needed
	if len(data) < 512 {
		paddedData := make([]byte, 512)
		copy(paddedData, data)
		// Fill padding with zeros (firmware expectation)
		for i := len(data); i < 512; i++ {
			paddedData[i] = 0
		}
		data = paddedData
	}
	
	// Return 512 bytes (the full struct including the CRC field and padding)
	// The IN_SECTION CRC trailer will be added after this by the reassembler
	return data[:512], nil
}

func (r *Reassembler) reconstructMfgInfo(jsonMap map[string]interface{}) ([]byte, error) {
	// Extract the mfg_info object
	mfgInfoData, ok := jsonMap["mfg_info"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing mfg_info field in JSON")
	}
	
	// Get the section size from JSON to handle variable sizes
	sectionSize := uint32(256) // default size
	if size, ok := jsonMap["size"].(float64); ok {
		sectionSize = uint32(size)
	}
	
	// Create and populate MFGInfo structure
	info := &types.MFGInfo{}
	
	// Populate fields from JSON (now stored as byte arrays)
	if psidData, ok := mfgInfoData["psid"].([]interface{}); ok {
		for i := 0; i < len(psidData) && i < 16; i++ {
			if val, ok := psidData[i].(float64); ok {
				info.PSID[i] = byte(val)
			}
		}
	}
	
	if partNumberData, ok := mfgInfoData["part_number"].([]interface{}); ok {
		for i := 0; i < len(partNumberData) && i < 32; i++ {
			if val, ok := partNumberData[i].(float64); ok {
				info.PartNumber[i] = byte(val)
			}
		}
	}
	
	if revisionData, ok := mfgInfoData["revision"].([]interface{}); ok {
		for i := 0; i < len(revisionData) && i < 16; i++ {
			if val, ok := revisionData[i].(float64); ok {
				info.Revision[i] = byte(val)
			}
		}
	}
	
	if productNameData, ok := mfgInfoData["product_name"].([]interface{}); ok {
		for i := 0; i < len(productNameData) && i < 64; i++ {
			if val, ok := productNameData[i].(float64); ok {
				info.ProductName[i] = byte(val)
			}
		}
	}
	
	if reservedData, ok := mfgInfoData["reserved"].([]interface{}); ok {
		for i := 0; i < len(reservedData) && i < 128; i++ {
			if val, ok := reservedData[i].(float64); ok {
				info.Reserved[i] = byte(val)
			}
		}
	}
	
	// Marshal the structure to binary
	data, err := info.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MFG_INFO: %w", err)
	}
	
	// Pad to the actual section size if needed
	if uint32(len(data)) < sectionSize {
		paddedData := make([]byte, sectionSize)
		copy(paddedData, data)
		data = paddedData
	}
	
	return data, nil
}

func (r *Reassembler) reconstructHashesTable(jsonMap map[string]interface{}) ([]byte, error) {
	// Get size from metadata
	size, ok := jsonMap["size"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing size field in JSON")
	}
	
	// Create buffer
	data := make([]byte, int(size))
	
	// Extract header data
	headerData, ok := jsonMap["header"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing header field in JSON")
	}
	
	// Create and populate HashesTableHeader structure
	header := &types.HashesTableHeader{}
	
	// Populate header fields from JSON
	if magic, ok := headerData["magic"].(float64); ok {
		header.Magic = uint32(magic)
	}
	if version, ok := headerData["version"].(float64); ok {
		header.Version = uint32(version)
	}
	if reserved1, ok := headerData["reserved1"].(float64); ok {
		header.Reserved1 = uint32(reserved1)
	}
	if reserved2, ok := headerData["reserved2"].(float64); ok {
		header.Reserved2 = uint32(reserved2)
	}
	if tableSize, ok := headerData["table_size"].(float64); ok {
		header.TableSize = uint32(tableSize)
	}
	if numEntries, ok := headerData["num_entries"].(float64); ok {
		header.NumEntries = uint32(numEntries)
	}
	if reserved3, ok := headerData["reserved3"].(float64); ok {
		header.Reserved3 = uint32(reserved3)
	}
	if crc, ok := headerData["crc"].(float64); ok {
		header.CRC = uint16(crc)
	}
	if reserved4, ok := headerData["reserved4"].(float64); ok {
		header.Reserved4 = uint16(reserved4)
	}
	
	// Marshal header to binary
	headerBytes, err := header.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal hashes table header: %w", err)
	}
	copy(data[0:32], headerBytes)
	
	// Process entries if present
	offset := 32 // After header
	if entries, ok := jsonMap["entries"].([]interface{}); ok {
		for _, entryInterface := range entries {
			if entryMap, ok := entryInterface.(map[string]interface{}); ok && offset+64 <= len(data) {
				// Create and populate HashTableEntry
				entry := &types.HashTableEntry{}
				
				// Hash
				if hashStr, ok := entryMap["hash"].(string); ok {
					if hashBytes, err := hex.DecodeString(hashStr); err == nil && len(hashBytes) == 32 {
						copy(entry.Hash[:], hashBytes)
					}
				}
				
				if entryType, ok := entryMap["type"].(float64); ok {
					entry.Type = uint32(entryType)
				}
				if entryOffset, ok := entryMap["offset"].(float64); ok {
					entry.Offset = uint32(entryOffset)
				}
				if entrySize, ok := entryMap["size"].(float64); ok {
					entry.Size = uint32(entrySize)
				}
				
				// Marshal entry to binary
				entryData, err := entry.Marshal()
				if err == nil {
					copy(data[offset:offset+64], entryData[:64]) // HashTableEntry is exactly 64 bytes
				}
				
				offset += 64
			}
		}
	}
	
	// Write reserved tail data if present
	if reservedTailStr, ok := jsonMap["reserved_tail"].(string); ok {
		if reservedTailBytes, err := hex.DecodeString(reservedTailStr); err == nil {
			copy(data[offset:], reservedTailBytes)
		}
	}
	
	r.logger.Info("Reconstructed HASHES_TABLE section from JSON",
		zap.Int("size", len(data)))
	
	return data, nil
}