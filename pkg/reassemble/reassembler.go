package reassemble

import (
	"encoding/base64"
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
	"github.com/Civil/mlx5fw-go/pkg/types/extracted"
)

// Options contains options for firmware reassembly
type Options struct {
	InputDir   string
	OutputFile string
	VerifyCRC  bool
	BinaryOnly bool // Force binary-only mode, ignore JSON files
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

func (r *Reassembler) loadMetadata(path string) (*extracted.FirmwareMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata extracted.FirmwareMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return &metadata, nil
}

func (r *Reassembler) verifyRequiredFiles(metadata *extracted.FirmwareMetadata) error {
	// Build section filename map
	fileMap := r.buildSectionFileMap(metadata.Sections)

	// Check for section files
	for _, section := range metadata.Sections {
		// Skip zero-size sections - they weren't extracted
		if section.Size() == 0 {
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

func (r *Reassembler) buildSectionFileMap(sections []extracted.SectionMetadata) map[string]string {
	// Group sections by type name, excluding zero-size sections
	sectionsByType := make(map[string][]extracted.SectionMetadata)
	for _, section := range sections {
		// Skip zero-size sections
		if section.Size() == 0 {
			continue
		}
		sectionsByType[section.TypeName()] = append(sectionsByType[section.TypeName()], section)
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
				key := fmt.Sprintf("%s_%d", typeName, section.Offset())
				fileMap[key] = fmt.Sprintf("%s_%d_0x%08x.bin", fileName, i, section.Offset())
			}
		} else {
			// Single section - no index
			section := sectionList[0]
			key := fmt.Sprintf("%s_%d", typeName, section.Offset())
			fileMap[key] = fmt.Sprintf("%s_0x%08x.bin", fileName, section.Offset())
		}
	}

	return fileMap
}

func (r *Reassembler) getSectionFileName(section extracted.SectionMetadata, fileMap map[string]string) string {
	// If the section has a filename in metadata, use it
	if section.FileName != "" {
		return section.FileName
	}

	// Otherwise, fall back to the generated filename
	key := fmt.Sprintf("%s_%d", section.TypeName(), section.Offset())
	if fileName, ok := fileMap[key]; ok {
		return fileName
	}
	// Fallback if not found in map
	fileName := strings.ReplaceAll(section.TypeName(), " ", "_")
	fileName = strings.ReplaceAll(fileName, "/", "_")
	return fmt.Sprintf("%s_0x%08x.bin", fileName, section.Offset())
}

func (r *Reassembler) reassembleFirmware(output io.WriteSeeker, metadata *extracted.FirmwareMetadata) error {
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
    // FS4/FS5 have a standardized 16-byte magic structure. FS3 does not.
    // For FS3 images, the initial header (including MTFW) is restored via gap files,
    // so we must not overwrite it here.
    if strings.EqualFold(metadata.Format, "FS3") {
        r.logger.Debug("Skipping magic pattern write for FS3 format")
    } else {
        if err := r.writeMagicPattern(firmwareData, metadata.Magic.Offset); err != nil {
            return fmt.Errorf("failed to write magic pattern: %w", err)
        }
    }

    // Write hardware pointers (FS4/FS5 only). FS3 has no HW pointers block.
    if metadata.HWPointers.FS4 != nil || metadata.HWPointers.FS5 != nil {
        if err := r.writeHWPointers(firmwareData, metadata.HWPointers); err != nil {
            return fmt.Errorf("failed to write HW pointers: %w", err)
        }
    } else {
        r.logger.Debug("No HW pointers present (likely FS3); skipping write")
    }

	// Process gaps FIRST (before sections)
	// This ensures that gap data doesn't overwrite section CRCs
	if err := r.reassembleGaps(firmwareData, metadata); err != nil {
		return fmt.Errorf("failed to reassemble gaps: %w", err)
	}

	// Process sections (will overwrite gap data with section data + CRC)
	for _, section := range metadata.Sections {
		// Skip zero-size sections
		if section.Size() == 0 {
			continue
		}

		// Read section data - prefer JSON over binary unless BinaryOnly mode
		sectionData, err := r.readSectionDataJSON(section, fileMap)
		if err != nil {
			return fmt.Errorf("failed to read section data: %w", err)
		}

		// Add CRC if needed (only for non-encrypted firmwares)
		// Check if we need to add CRC: only if the data doesn't already include it
		// The extractor removes CRC from binary files for non-encrypted firmwares,
		// so we need to add it back during reassembly
		needsCRC := section.CRCType() == types.CRCInSection &&
			section.OriginalSize > section.Size() &&
			!metadata.IsEncrypted &&
			len(sectionData) < int(section.OriginalSize) // Data doesn't already include CRC

		if needsCRC {
			r.logger.Info("Adding IN_SECTION CRC",
				zap.String("section", section.TypeName()),
				zap.Uint32("originalSize", section.OriginalSize),
				zap.Uint32("size", section.Size()),
				zap.Int("dataLen", len(sectionData)))

			// Determine whether to use blank CRC based on section type
			// Some sections use 0xFFFFFFFF as a placeholder CRC value
			var crcBytes []byte

			// Check if this is a section type that typically has blank CRCs
			// Based on observation: BOOT2, HASHES_TABLE, DEV_INFO, and device data sections
			// NOTE: TOOLS_AREA should have its CRC calculated, not blank
			if section.TypeName() == "BOOT2" ||
				section.TypeName() == "HASHES_TABLE" || section.TypeName() == "DEV_INFO" ||
				section.TypeName() == "MFG_INFO" || section.TypeName() == "IMAGE_INFO" ||
				section.TypeName() == "FORBIDDEN_VERSIONS" || section.TypeName() == "PUBLIC_KEYS_2048" ||
				section.TypeName() == "PUBLIC_KEYS_4096" || section.TypeName() == "IMAGE_SIGNATURE_512" ||
				strings.HasPrefix(section.TypeName(), "UNKNOWN_0xE0") {
				// Use blank CRC for these sections
				crcBytes = []byte{0xFF, 0xFF, 0xFF, 0xFF}
				r.logger.Info("Using blank CRC for section",
					zap.String("section", section.TypeName()))
			} else {
				// Calculate CRC
				var crc uint16
				if r.isHardwareCRCSection(section) {
					crc = crcCalc.CalculateHardwareCRC(sectionData)
					r.logger.Info("Calculated hardware CRC",
						zap.String("section", section.TypeName()),
						zap.Uint16("crc", crc),
						zap.Int("dataLen", len(sectionData)))
				} else {
					crc = crcCalc.CalculateSoftwareCRC16(sectionData)
					r.logger.Info("Calculated software CRC",
						zap.String("section", section.TypeName()),
						zap.Uint16("crc", crc),
						zap.Int("dataLen", len(sectionData)))
				}

				// Append CRC (16-bit CRC in lower 16 bits of 32-bit word, big-endian)
				crcBytes = make([]byte, 4)
				// Create a simple struct to handle the CRC format
				// Based on mstflint source: CRC is stored in lower 16 bits
				crcStruct := struct {
					Reserved uint16 `offset:"byte:0,endian:be"`
					CRC      uint16 `offset:"byte:2,endian:be"`
				}{Reserved: 0, CRC: crc}
				crcData, _ := annotations.MarshalStruct(&crcStruct)
				crcBytes = crcData
			}
			sectionData = append(sectionData, crcBytes...)

			r.logger.Debug("Added CRC to section",
				zap.String("section", section.TypeName()),
				zap.String("crcBytes", fmt.Sprintf("%02x%02x%02x%02x",
					crcBytes[0], crcBytes[1], crcBytes[2], crcBytes[3])))
		} else if section.CRCType() == types.CRCInSection && metadata.IsEncrypted {
			r.logger.Debug("Keeping section data intact for encrypted firmware",
				zap.String("section", section.TypeName()),
				zap.Int("dataSize", len(sectionData)))
		}

		// Write section data
		copy(firmwareData[section.Offset():], sectionData)
	}

	// Write ITOC and DTOC headers if they exist
	if err := r.writeTOCHeaders(firmwareData, metadata); err != nil {
		return fmt.Errorf("failed to write TOC headers: %w", err)
	}

    // Update CRCs in HW pointers if present
    if metadata.HWPointers.FS4 != nil || metadata.HWPointers.FS5 != nil {
        if err := r.updateHWPointerCRCs(firmwareData, metadata, crcCalc); err != nil {
            return fmt.Errorf("failed to update HW pointer CRCs: %w", err)
        }
    } else {
        r.logger.Debug("No HW pointers present; skipping CRC update")
    }

	// Write the complete firmware
	if _, err := output.Write(firmwareData); err != nil {
		return fmt.Errorf("failed to write firmware data: %w", err)
	}

	return nil
}

func (r *Reassembler) reassembleGaps(firmwareData []byte, metadata *extracted.FirmwareMetadata) error {
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

		start := segment.StartOffset
		end := segment.EndOffset

		// Check for binary file first (non-uniform gaps)
		gapFileName := fmt.Sprintf("gap_%03d_", gapIndex)
		binFiles, err := filepath.Glob(filepath.Join(gapsDir, gapFileName+"*.bin"))
		if err == nil && len(binFiles) > 0 {
			// Binary file exists - use it
			gapData, err := os.ReadFile(binFiles[0])
			if err != nil {
				return fmt.Errorf("failed to read gap file %s: %w", binFiles[0], err)
			}

			// Copy gap data to firmware
			copy(firmwareData[start:end], gapData)
			r.logger.Debug("Restored gap from binary file",
				zap.Int("index", gapIndex),
				zap.Uint64("start", start),
				zap.Uint64("size", segment.Size),
				zap.String("file", filepath.Base(binFiles[0])))
		} else {
			// No binary file, check for metadata file
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
				// No gap files found
				r.logger.Warn("Gap file not found",
					zap.Int("index", gapIndex),
					zap.Uint64("start", start),
					zap.Uint64("end", end))
			}
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

func (r *Reassembler) writeHWPointers(data []byte, hwPointers extracted.HWPointersInfo) error {
    // If neither FS4 nor FS5 pointers exist (e.g., FS3 image), do nothing
    if hwPointers.FS4 == nil && hwPointers.FS5 == nil {
        return nil
    }
    // Marshal HW pointers structure based on type
    var hwPointersData []byte
    var err error

    if hwPointers.FS4 != nil {
		hwPointersData, err = hwPointers.FS4.Marshal()
		if err != nil {
			return fmt.Errorf("failed to marshal FS4 HW pointers: %w", err)
		}
	} else if hwPointers.FS5 != nil {
		hwPointersData, err = hwPointers.FS5.Marshal()
		if err != nil {
			return fmt.Errorf("failed to marshal FS5 HW pointers: %w", err)
		}
	} else {
		return fmt.Errorf("no HW pointers data found")
	}

	if int(hwPointers.Offset)+len(hwPointersData) > len(data) {
		return fmt.Errorf("HW pointers offset out of bounds")
	}
	copy(data[hwPointers.Offset:], hwPointersData)
	return nil
}

func (r *Reassembler) writeTOCHeaders(data []byte, metadata *extracted.FirmwareMetadata) error {
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

func (r *Reassembler) updateHWPointerCRCs(data []byte, metadata *extracted.FirmwareMetadata, crcCalc *parser.CRCCalculator) error {
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
		if err := entry.UnmarshalWithReserved(data[entryOffset : entryOffset+8]); err != nil {
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

func (r *Reassembler) isHardwareCRCSection(section extracted.SectionMetadata) bool {
	// Use the single source of truth for CRC algorithm determination
	return types.GetSectionCRCAlgorithm(section.Type()) == types.CRCAlgorithmHardware
}
