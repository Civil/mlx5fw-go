package reassemble

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"github.com/Civil/mlx5fw-go/pkg/annotations"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/Civil/mlx5fw-go/pkg/types/extracted"
)

// Helper function for hex decoding
func hexToBytes(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

// readSectionDataJSON reads section data using the new JSON format
func (r *Reassembler) readSectionDataJSON(section extracted.SectionMetadata, fileMap map[string]string) ([]byte, error) {
	sectionFileName := r.getSectionFileName(section, fileMap)
	jsonFileName := strings.TrimSuffix(sectionFileName, ".bin") + ".json"
	jsonPath := filepath.Join(r.options.InputDir, jsonFileName)

	// Always read JSON file first (it should always exist)
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file %s: %w", jsonFileName, err)
	}

	// First parse to get section type
	var baseSection struct {
		Type       types.SectionType `json:"type"`
		HasRawData bool              `json:"has_raw_data"`
	}
	if err := json.Unmarshal(jsonData, &baseSection); err != nil {
		return nil, fmt.Errorf("failed to parse base section JSON %s: %w", jsonFileName, err)
	}

	// Check if this section has raw data flag or binary-only mode
	if r.options.BinaryOnly || baseSection.HasRawData {
		// Read binary file
		binaryPath := filepath.Join(r.options.InputDir, sectionFileName)
		sectionData, err := os.ReadFile(binaryPath)
		if err != nil {
			if baseSection.HasRawData {
				return nil, fmt.Errorf("section has raw data flag but binary file not found: %s", sectionFileName)
			}
			return nil, fmt.Errorf("failed to read binary file %s: %w", sectionFileName, err)
		}

		r.logger.Debug("Read section from binary",
			zap.String("section", section.TypeName()),
			zap.String("file", sectionFileName),
			zap.Bool("has_raw_data", baseSection.HasRawData))

		return sectionData, nil
	}

	// Try to reconstruct from JSON
	r.logger.Debug("Attempting JSON reconstruction",
		zap.String("section", section.TypeName()),
		zap.String("json_file", jsonFileName))

	reconstructed, err := r.reconstructFromJSONByType(jsonData, section)
	if err != nil {
		// If reconstruction fails and binary file exists, use it
		binaryPath := filepath.Join(r.options.InputDir, sectionFileName)
		if _, statErr := os.Stat(binaryPath); statErr == nil {
			sectionData, readErr := os.ReadFile(binaryPath)
			if readErr == nil {
				r.logger.Debug("Using binary file after failed JSON reconstruction",
					zap.String("section", section.TypeName()),
					zap.String("file", sectionFileName))
				return sectionData, nil
			}
		}
		return nil, fmt.Errorf("failed to reconstruct from JSON: %w", err)
	}

	r.logger.Info("Reconstructed section from JSON",
		zap.String("section", section.TypeName()),
		zap.String("file", jsonFileName))

	return reconstructed, nil
}

// reconstructFromJSONByType reconstructs section data based on section type
func (r *Reassembler) reconstructFromJSONByType(jsonData []byte, metadata extracted.SectionMetadata) ([]byte, error) {
	switch uint16(metadata.Type()) {
    case types.SectionTypeImageInfo:
		// Parse JSON into wrapper struct
		var sectionData struct {
			ImageInfo *types.ImageInfo `json:"image_info"`
		}
		if err := json.Unmarshal(jsonData, &sectionData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal IMAGE_INFO JSON: %w", err)
		}
		if sectionData.ImageInfo == nil {
			return nil, fmt.Errorf("missing image_info data in JSON")
		}

        // Marshal to fixed size (1024 bytes)
        data, err := annotations.MarshalStructWithSize(sectionData.ImageInfo, int(types.ImageInfoSize))
        if err != nil {
            return nil, fmt.Errorf("failed to marshal IMAGE_INFO: %w", err)
        }

        return data, nil

    case types.SectionTypeDevInfo:
		// Parse JSON into wrapper struct
		var sectionData struct {
			DeviceInfo *types.DevInfo `json:"device_info"`
		}
		if err := json.Unmarshal(jsonData, &sectionData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal DEV_INFO JSON: %w", err)
		}
		if sectionData.DeviceInfo == nil {
			return nil, fmt.Errorf("missing device_info data in JSON")
		}

        // Marshal to the fixed DEV_INFO structure size (512 bytes)
        data, err := annotations.MarshalStructWithSize(sectionData.DeviceInfo, int(types.DevInfoSize))
        if err != nil {
            return nil, fmt.Errorf("failed to marshal DEV_INFO: %w", err)
        }

		r.logger.Debug("DEV_INFO marshaled data",
			zap.Int("dataLen", len(data)),
			zap.Uint32("expectedSize", metadata.Size()))

		// Don't add CRC here - the main reassembler will handle it if needed
		// The marshaled data already includes the structure without CRC

		// Ensure data is exactly the expected size
		if len(data) > int(metadata.Size()) {
			// Trim to expected size if larger
			data = data[:metadata.Size()]
		} else if len(data) < int(metadata.Size()) {
			// Pad with zeros if smaller
			paddedData := make([]byte, metadata.Size())
			copy(paddedData, data)
			data = paddedData
		}

		return data, nil

	case types.SectionTypeMfgInfo:
		// Parse JSON into wrapper struct
		var sectionData struct {
			MfgInfo *types.MfgInfo `json:"mfg_info"`
		}
		if err := json.Unmarshal(jsonData, &sectionData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal MFG_INFO JSON: %w", err)
		}
		if sectionData.MfgInfo == nil {
			return nil, fmt.Errorf("missing mfg_info data in JSON")
		}

        // Marshal to the exact section size so reserved tail matches image
        data, err := annotations.MarshalStructWithSize(sectionData.MfgInfo, int(metadata.Size()))
        if err != nil {
            return nil, fmt.Errorf("failed to marshal MFG_INFO: %w", err)
        }
        return data, nil

	case types.SectionTypeHashesTable:
		// Parse JSON into wrapper struct
		var sectionData struct {
			Header       *types.HashesTableHeader `json:"header"`
			Entries      []*types.HashTableEntry  `json:"entries"`
			ReservedTail string                   `json:"reserved_tail,omitempty"`
		}
		if err := json.Unmarshal(jsonData, &sectionData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal HASHES_TABLE JSON: %w", err)
		}
		if sectionData.Header == nil {
			return nil, fmt.Errorf("missing header data in JSON")
		}

		// Create buffer
		data := make([]byte, metadata.Size())

		// Marshal header to binary
		headerBytes, err := sectionData.Header.Marshal()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal hashes table header: %w", err)
		}
		copy(data[0:32], headerBytes)

		// Process entries
		offset := 32 // After header
		for _, entry := range sectionData.Entries {
			if offset+64 > len(data) {
				break
			}

			// Marshal entry to binary
			entryData, err := entry.Marshal()
			if err == nil {
				copy(data[offset:offset+64], entryData[:64])
			}

			offset += 64
		}

		// Write reserved tail data if present
		if sectionData.ReservedTail != "" {
			if reservedTailBytes, err := hexToBytes(sectionData.ReservedTail); err == nil {
				copy(data[offset:], reservedTailBytes)
			}
		}

		r.logger.Info("Reconstructed HASHES_TABLE section from JSON",
			zap.Int("size", len(data)))

		return data, nil

	case types.SectionTypeForbiddenVersions:
		// Parse JSON into wrapper struct
		var sectionData struct {
			ForbiddenVersions *types.ForbiddenVersions `json:"forbidden_versions"`
		}
		if err := json.Unmarshal(jsonData, &sectionData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal FORBIDDEN_VERSIONS JSON: %w", err)
		}
		if sectionData.ForbiddenVersions == nil {
			return nil, fmt.Errorf("missing forbidden_versions data in JSON")
		}

		// Marshal to get raw data
		data, err := annotations.MarshalStructWithSize(sectionData.ForbiddenVersions, int(metadata.Size()))
		if err != nil {
			return nil, fmt.Errorf("failed to marshal FORBIDDEN_VERSIONS: %w", err)
		}

		return data, nil

	case types.SectionTypeImageSignature256:
		// Parse JSON into wrapper struct
		var sectionData struct {
			ImageSignature *types.ImageSignature `json:"image_signature"`
			Padding        types.FWByteSlice     `json:"padding,omitempty"`
		}
		if err := json.Unmarshal(jsonData, &sectionData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal IMAGE_SIGNATURE_256 JSON: %w", err)
		}
		if sectionData.ImageSignature == nil {
			return nil, fmt.Errorf("missing image_signature data in JSON")
		}

		// Marshal signature structure
		sigData, err := sectionData.ImageSignature.Marshal()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal IMAGE_SIGNATURE_256: %w", err)
		}

		// Add padding if present
		data := sigData
		if len(sectionData.Padding) > 0 {
			data = append(data, sectionData.Padding...)
		}

		// Pad to section size if needed
		if uint32(len(data)) < metadata.Size() {
			paddedData := make([]byte, metadata.Size())
			copy(paddedData, data)
			data = paddedData
		}

		return data, nil

	case types.SectionTypeImageSignature512:
		// Parse JSON into wrapper struct
		var sectionData struct {
			ImageSignature *types.ImageSignature2 `json:"image_signature"`
			Padding        types.FWByteSlice      `json:"padding,omitempty"`
		}
		if err := json.Unmarshal(jsonData, &sectionData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal IMAGE_SIGNATURE_512 JSON: %w", err)
		}
		if sectionData.ImageSignature == nil {
			return nil, fmt.Errorf("missing image_signature data in JSON")
		}

		// Marshal signature structure
		sigData, err := sectionData.ImageSignature.Marshal()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal IMAGE_SIGNATURE_512: %w", err)
		}

		// Add padding if present
		data := sigData
		if len(sectionData.Padding) > 0 {
			data = append(data, sectionData.Padding...)
		}

		// Pad to section size if needed
		if uint32(len(data)) < metadata.Size() {
			paddedData := make([]byte, metadata.Size())
			copy(paddedData, data)
			data = paddedData
		}

		return data, nil

	case types.SectionTypePublicKeys2048:
		// Parse JSON into wrapper struct
		var sectionData struct {
			PublicKeys *types.PublicKeys `json:"public_keys"`
		}
		if err := json.Unmarshal(jsonData, &sectionData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal PUBLIC_KEYS_2048 JSON: %w", err)
		}
		if sectionData.PublicKeys == nil {
			return nil, fmt.Errorf("missing public_keys data in JSON")
		}

		// Marshal to get raw data
		data, err := sectionData.PublicKeys.Marshal()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal PUBLIC_KEYS_2048: %w", err)
		}

		// Pad to section size if needed
		if uint32(len(data)) < metadata.Size() {
			paddedData := make([]byte, metadata.Size())
			copy(paddedData, data)
			data = paddedData
		}

		return data, nil

	case types.SectionTypePublicKeys4096:
		// Parse JSON into wrapper struct
		var sectionData struct {
			PublicKeys *types.PublicKeys2 `json:"public_keys"`
		}
		if err := json.Unmarshal(jsonData, &sectionData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal PUBLIC_KEYS_4096 JSON: %w", err)
		}
		if sectionData.PublicKeys == nil {
			return nil, fmt.Errorf("missing public_keys data in JSON")
		}

		// Marshal to get raw data
		data, err := sectionData.PublicKeys.Marshal()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal PUBLIC_KEYS_4096: %w", err)
		}

		// Pad to section size if needed
		if uint32(len(data)) < metadata.Size() {
			paddedData := make([]byte, metadata.Size())
			copy(paddedData, data)
			data = paddedData
		}

		return data, nil

	default:
		// For other sections, we need binary data
		return nil, fmt.Errorf("section type %s requires binary file", metadata.TypeName())
	}
}

// Section-specific reconstruction methods for new JSON format

func (r *Reassembler) reconstructDevInfoFromJSON(devInfo *types.DevInfoJSON, sectionSize uint32) ([]byte, error) {
	// Create and populate DevInfo structure
	info := &types.DevInfo{}

	// Populate fields from JSON
	info.Signature0 = devInfo.Signature0
	info.Signature1 = devInfo.Signature1
	info.Signature2 = devInfo.Signature2
	info.Signature3 = devInfo.Signature3
	info.MinorVersion = devInfo.MinorVersion
	info.MajorVersion = devInfo.MajorVersion
	info.Reserved1 = devInfo.Reserved1
	copy(info.Reserved2[:], devInfo.Reserved2)

	// GUID info
	info.Guids.Reserved1 = devInfo.Guids.Reserved1
	info.Guids.Step = devInfo.Guids.Step
	info.Guids.NumAllocated = devInfo.Guids.NumAllocated
	info.Guids.Reserved2 = devInfo.Guids.Reserved2
	info.Guids.UID = devInfo.Guids.UID

	// MAC info
	info.Macs.Reserved1 = devInfo.Macs.Reserved1
	info.Macs.Step = devInfo.Macs.Step
	info.Macs.NumAllocated = devInfo.Macs.NumAllocated
	info.Macs.Reserved2 = devInfo.Macs.Reserved2
	info.Macs.UID = devInfo.Macs.UID

	// Reserved fields
	copy(info.Reserved3[:], devInfo.Reserved3)
	// Reserved4 field was removed from DevInfo

	// Clear the CRC field - it MUST be recalculated
	info.CRC = 0

	// Marshal the structure to binary with CRC=0
	data, err := info.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DEV_INFO: %w", err)
	}

	// Calculate CRC on first 508 bytes
	crcCalc := parser.NewCRCCalculator()
	crc := crcCalc.CalculateSoftwareCRC16(data[:508])

	r.logger.Info("DEV_INFO CRC calculation",
		zap.Uint16("calculatedCRC", crc))

	// Set the CRC in the struct and remarshal
	info.CRC = uint32(crc)
	data, err = info.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DEV_INFO with CRC: %w", err)
	}

	// Pad data to 512 bytes if needed
	if len(data) < 512 {
		paddedData := make([]byte, 512)
		copy(paddedData, data)
		data = paddedData
	}

	return data[:512], nil
}

func (r *Reassembler) reconstructMfgInfoFromJSON(mfgInfo *types.MfgInfoJSON, sectionSize uint32) ([]byte, error) {
	// Create and populate MFGInfo structure (annotated layout)
	info := &types.MfgInfo{}
	// Only PSID is defined explicitly in annotated layout; copy it
	copy(info.PSID[:], mfgInfo.PSID)

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

func (r *Reassembler) reconstructHashesTableFromJSON(hashesTable *types.HashesTableJSON, sectionSize uint32) ([]byte, error) {
	// Create buffer
	data := make([]byte, sectionSize)

	// Create and populate HashesTableHeader structure
	header := &types.HashesTableHeader{
		Magic:      hashesTable.Header.Magic,
		Version:    hashesTable.Header.Version,
		Reserved1:  hashesTable.Header.Reserved1,
		Reserved2:  hashesTable.Header.Reserved2,
		TableSize:  hashesTable.Header.TableSize,
		NumEntries: hashesTable.Header.NumEntries,
		Reserved3:  hashesTable.Header.Reserved3,
		CRC:        hashesTable.Header.CRC,
		Reserved4:  hashesTable.Header.Reserved4,
	}

	// Marshal header to binary
	headerBytes, err := header.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal hashes table header: %w", err)
	}
	copy(data[0:32], headerBytes)

	// Process entries
	offset := 32 // After header
	for _, entryJSON := range hashesTable.Entries {
		if offset+64 > len(data) {
			break
		}

		// Create and populate HashTableEntry
		entry := &types.HashTableEntry{
			Type:   entryJSON.Type,
			Offset: entryJSON.Offset,
			Size:   entryJSON.Size,
		}

		// Decode hash
		if hashBytes, err := hexToBytes(entryJSON.Hash); err == nil && len(hashBytes) == 32 {
			copy(entry.Hash[:], hashBytes)
		}

		// Marshal entry to binary
		entryData, err := entry.Marshal()
		if err == nil {
			copy(data[offset:offset+64], entryData[:64])
		}

		offset += 64
	}

	// Write reserved tail data if present
	if reservedTailBytes, err := hexToBytes(hashesTable.ReservedTail); err == nil {
		copy(data[offset:], reservedTailBytes)
	}

	r.logger.Info("Reconstructed HASHES_TABLE section from JSON",
		zap.Int("size", len(data)))

	return data, nil
}

func (r *Reassembler) reconstructImageSignatureFromJSON(sig *types.ImageSignatureJSON,
	sectionType uint16, sectionSize uint32, paddingHex string) ([]byte, error) {

	// Decode signature
	signature, err := hexToBytes(sig.Signature)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature hex: %w", err)
	}

	// Create appropriate structure based on section type
	var structData []byte

	if sectionType == types.SectionTypeImageSignature256 {
		imgSig := &types.ImageSignature{
			SignatureType: sig.SignatureType,
		}
		copy(imgSig.Signature[:], signature)
		structData, err = imgSig.Marshal()
	} else if sectionType == types.SectionTypeImageSignature512 {
		imgSig := &types.ImageSignature2{
			SignatureType: sig.SignatureType,
		}
		copy(imgSig.Signature[:], signature)
		structData, err = imgSig.Marshal()
	} else {
		return nil, fmt.Errorf("unsupported signature section type: %d", sectionType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to marshal signature structure: %w", err)
	}

	// Handle padding
	if uint32(len(structData)) < sectionSize {
		if paddingHex != "" {
			paddingData, err := hexToBytes(paddingHex)
			if err != nil {
				return nil, fmt.Errorf("failed to decode padding hex: %w", err)
			}
			structData = append(structData, paddingData...)
		} else {
			// Pad with zero bytes if no padding data provided
			padding := make([]byte, sectionSize-uint32(len(structData)))
			structData = append(structData, padding...)
		}
	}

	return structData, nil
}

func (r *Reassembler) reconstructPublicKeysFromJSON(keys []types.PublicKeyJSON,
	sectionType uint16, sectionSize uint32) ([]byte, error) {

	// Determine entry size based on section type
	entrySize := 276 // For PUBLIC_KEYS_2048
	if sectionType == types.SectionTypePublicKeys4096 {
		entrySize = 532 // For PUBLIC_KEYS_4096
	}

	// Create result buffer with section size to handle padding
	result := make([]byte, sectionSize)

	// Process each key
	for i, keyJSON := range keys {
		// Decode UUID and key
		uuid, _ := hexToBytes(keyJSON.UUID)
		key, _ := hexToBytes(keyJSON.Key)

		// Create key structure based on type
		if sectionType == types.SectionTypePublicKeys2048 {
			pk := &types.PublicKey{
				Reserved: keyJSON.Reserved,
			}
			copy(pk.UUID[:], uuid)
			copy(pk.Key[:], key)
			pkData, _ := pk.Marshal()
			copy(result[i*entrySize:], pkData)
		} else {
			pk2 := &types.PublicKey2{
				Reserved: keyJSON.Reserved,
			}
			copy(pk2.UUID[:], uuid)
			copy(pk2.Key[:], key)
			pk2Data, _ := pk2.Marshal()
			copy(result[i*entrySize:], pk2Data)
		}
	}

	return result, nil
}

func (r *Reassembler) reconstructForbiddenVersionsFromJSON(fv *types.ForbiddenVersionsJSON,
	sectionSize uint32) ([]byte, error) {

	// Calculate actual number of version slots based on section size
	numVersionSlots := (sectionSize - 8) / 4

	// Create ForbiddenVersions structure
	forbiddenVer := &types.ForbiddenVersions{
		Count:    fv.Count,
		Reserved: fv.Reserved,
		Versions: make([]uint32, numVersionSlots),
	}

	// Copy versions
	for i, ver := range fv.Versions {
		if i >= int(numVersionSlots) {
			break
		}
		forbiddenVer.Versions[i] = ver
	}

	// Marshal the structure to bytes with the expected section size
	return annotations.MarshalStructWithSize(forbiddenVer, int(sectionSize))
}
