package sections

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// DBGFwIniSection represents a DBG_FW_INI section
type DBGFwIniSection struct {
	*interfaces.BaseSection
	Header  *types.DBGFwIni
	IniData []byte // Stores the decompressed INI file content
}

// NewDBGFwIniSection creates a new DBGFwIni section
func NewDBGFwIniSection(base *interfaces.BaseSection) *DBGFwIniSection {
	return &DBGFwIniSection{
		BaseSection: base,
	}
}

// Parse parses the DBG_FW_INI section data
func (s *DBGFwIniSection) Parse(data []byte) error {
	s.SetRawData(data)

	// According to mstflint source code analysis:
	// DBG_FW_INI section is always compressed with zlib in practice,
	// but this is NOT indicated in the ITOC entry flags.
	// mstflint just attempts to decompress unconditionally.

	// Try to decompress the entire data (no header)
	decompressed, err := s.decompressZlib(data)
	if err == nil {
		// Successfully decompressed - this is the INI data
		s.IniData = decompressed

		// Create a synthetic header for consistency
		s.Header = &types.DBGFwIni{
			CompressionMethod: 1, // Zlib
			UncompressedSize:  uint32(len(decompressed)),
			CompressedSize:    uint32(len(data)),
			Reserved:          0,
		}

	} else {
		// Decompression failed, keep the raw data but don't try to parse as header
		// The data is likely compressed but we couldn't decompress it
		s.IniData = nil
		// Don't set a header to avoid garbage values in JSON
		s.Header = nil
	}

	return nil
}

// MarshalJSON returns JSON representation of the DBG_FW_INI section
func (s *DBGFwIniSection) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type":      s.Type(),
		"type_name": s.TypeName(),
		"offset":    s.Offset(),
		"size":      s.Size(),
	}

	// Always need raw data for bit-perfect reconstruction
	// (zlib compression is not deterministic)
	result["has_raw_data"] = true

	// Also mark if we have extracted data for convenience
	if s.IniData != nil && len(s.IniData) > 0 {
		result["has_extracted_data"] = true
		result["extracted_data_type"] = "ini"
	}

	if s.Header != nil {
		compressionMethod := "Unknown"
		switch s.Header.CompressionMethod {
		case 0:
			compressionMethod = "Uncompressed"
		case 1:
			compressionMethod = "Zlib"
		case 2:
			compressionMethod = "LZMA"
		}

		result["dbg_fw_ini"] = map[string]interface{}{
			"compression_method": compressionMethod,
			"uncompressed_size":  s.Header.UncompressedSize,
			"compressed_size":    s.Header.CompressedSize,
			"ini_data_size":      len(s.IniData),
		}
	} else {
		// No header, likely compressed data that we couldn't decompress
		result["dbg_fw_ini"] = map[string]interface{}{
			"note": "Compressed data without header, decompression failed",
		}
	}

	return json.Marshal(result)
}

// GetExtractedData returns the decompressed INI data if available
func (s *DBGFwIniSection) GetExtractedData() []byte {
	return s.IniData
}

// Marshal marshals the DBGFwIniSection back to binary format
func (s *DBGFwIniSection) Marshal() ([]byte, error) {
	// Compress the INI data
	compressedData, err := s.compressZlib(s.IniData)
	if err != nil {
		return nil, fmt.Errorf("failed to compress INI data: %w", err)
	}

	// Based on how mstflint works, just return the compressed data without header
	return compressedData, nil
}

// decompressZlib decompresses zlib compressed data
func (s *DBGFwIniSection) decompressZlib(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer reader.Close()

	// Read all data
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress: %w", err)
	}

	return decompressed, nil
}

// compressZlib compresses data using zlib
func (s *DBGFwIniSection) compressZlib(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	// Use BestCompression (level 9) to match original firmware
	w, err := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(data); err != nil {
		w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DBGFwParamsSection represents a DBG_FW_PARAMS section
type DBGFwParamsSection struct {
	*interfaces.BaseSection
	Header *types.DBGFwParams
	Data   []byte
}

// NewDBGFwParamsSection creates a new DBGFwParams section
func NewDBGFwParamsSection(base *interfaces.BaseSection) *DBGFwParamsSection {
	return &DBGFwParamsSection{
		BaseSection: base,
	}
}

// Parse parses the DBG_FW_PARAMS section data
func (s *DBGFwParamsSection) Parse(data []byte) error {
	s.SetRawData(data)

	// Based on mstflint analysis, DBG_FW_PARAMS can be very small (8 bytes)
	// and may contain compressed data without a header structure
	if len(data) < 16 {
		// Small section, likely compressed data without header
		// Check if it's zlib compressed (starts with 0x78)
		if len(data) >= 2 && data[0] == 0x78 {
			// This is compressed data, treat it as raw data
			s.Data = data
			s.Header = nil
		} else {
			// Unknown format, keep raw data
			s.Data = data
			s.Header = nil
		}
		return nil
	}

	// Normal case: has header
	s.Header = &types.DBGFwParams{}
	if err := s.Header.Unmarshal(data[:16]); err != nil {
		return merry.Wrap(err)
	}

	if len(data) > 16 {
		s.Data = data[16:]
	}

	return nil
}

// MarshalJSON returns JSON representation of the DBG_FW_PARAMS section
func (s *DBGFwParamsSection) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type":         s.Type(),
		"type_name":    s.TypeName(),
		"offset":       s.Offset(),
		"size":         s.Size(),
		"has_raw_data": true, // DBG_FW_PARAMS needs binary data
	}

	if s.Header != nil {
		compressionMethod := "Unknown"
		switch s.Header.CompressionMethod {
		case 0:
			compressionMethod = "Uncompressed"
		case 1:
			compressionMethod = "Zlib"
		case 2:
			compressionMethod = "LZMA"
		}

		result["dbg_fw_params"] = map[string]interface{}{
			"compression_method": compressionMethod,
			"uncompressed_size":  s.Header.UncompressedSize,
			"compressed_size":    s.Header.CompressedSize,
			"data_size":          len(s.Data),
		}
	} else if s.Data != nil && len(s.Data) > 0 {
		// No header, just raw data
		result["dbg_fw_params"] = map[string]interface{}{
			"note":      "Small section with raw data (no header)",
			"data_size": len(s.Data),
		}

		// Check if it looks like compressed data
		if len(s.Data) >= 2 && s.Data[0] == 0x78 {
			result["dbg_fw_params"].(map[string]interface{})["possible_compression"] = "zlib"
		}
	}

	return json.Marshal(result)
}

// FWAdbSection represents a FW_ADB section
type FWAdbSection struct {
	*interfaces.BaseSection
	Header *types.FWAdb
	Data   []byte
}

// NewFWAdbSection creates a new FWAdb section
func NewFWAdbSection(base *interfaces.BaseSection) *FWAdbSection {
	return &FWAdbSection{
		BaseSection: base,
	}
}

// Parse parses the FW_ADB section data
func (s *FWAdbSection) Parse(data []byte) error {
	s.SetRawData(data)

	if len(data) < 16 {
		return merry.New("FW_ADB section too small")
	}

	s.Header = &types.FWAdb{}
	if err := s.Header.Unmarshal(data[:16]); err != nil {
		return merry.Wrap(err)
	}

	if len(data) > 16 {
		s.Data = data[16:]
	}

	return nil
}

// MarshalJSON returns JSON representation of the FW_ADB section
func (s *FWAdbSection) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type":         s.Type(),
		"type_name":    s.TypeName(),
		"offset":       s.Offset(),
		"size":         s.Size(),
		"has_raw_data": true, // FW_ADB needs binary data
	}

	if s.Header != nil {
		result["fw_adb"] = map[string]interface{}{
			"version":   s.Header.Version,
			"size":      s.Header.Size,
			"data_size": len(s.Data),
		}
	}

	return json.Marshal(result)
}
