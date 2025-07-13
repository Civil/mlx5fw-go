package main

import (
	"encoding/json"
	"fmt"
	"os"
	
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
)

// QueryJSONOutput represents the JSON structure for query command output
type QueryJSONOutput struct {
	ImageType       string          `json:"image_type"`
	FWVersion       string          `json:"fw_version"`
	FWReleaseDate   string          `json:"fw_release_date"`
	MICVersion      string          `json:"mic_version"`
	PRSName         string          `json:"prs_name"`
	PartNumber      string          `json:"part_number"`
	Description     string          `json:"description"`
	ProductVersion  string          `json:"product_version,omitempty"`
	RomInfo         []RomInfoJSON   `json:"rom_info,omitempty"`
	BaseGUID        *UIDInfo        `json:"base_guid"`
	BaseMAC         *UIDInfo        `json:"base_mac"`
	ImageVSD        string          `json:"image_vsd"`
	DeviceVSD       string          `json:"device_vsd"`
	PSID            string          `json:"psid"`
	SecurityAttrs   string          `json:"security_attributes"`
	SecurityVer     int             `json:"security_version"`
	DefaultUpdateMethod string      `json:"default_update_method"`
}

// UIDInfo represents UID information in JSON
type UIDInfo struct {
	UID   string `json:"uid"`
	Count int    `json:"count"`
}

// RomInfoJSON represents ROM information in JSON
type RomInfoJSON struct {
	Type    string `json:"type"`
	Version string `json:"version"`
	CPU     string `json:"cpu,omitempty"`
}

// SectionJSONOutput represents the JSON structure for sections command output
type SectionJSONOutput struct {
	Sections []SectionInfoJSON `json:"sections"`
}

// SectionInfoJSON represents a single section in JSON
type SectionInfoJSON struct {
	Type         string `json:"type"`
	TypeName     string `json:"type_name"`
	Offset       string `json:"offset"`
	Size         string `json:"size"`
	CRCType      string `json:"crc_type"`
	IsEncrypted  bool   `json:"is_encrypted"`
	IsDeviceData bool   `json:"is_device_data"`
	Content      string `json:"content,omitempty"`
}

// convertToQueryJSON converts FirmwareInfo to JSON output structure
func convertToQueryJSON(info *interfaces.FirmwareInfo) *QueryJSONOutput {
	output := &QueryJSONOutput{
		ImageType:       info.Format,
		FWVersion:       info.FWVersion,
		FWReleaseDate:   info.FWReleaseDate,
		MICVersion:      info.MICVersion,
		PRSName:         info.PRSName,
		PartNumber:      info.PartNumber,
		Description:     info.Description,
		ProductVersion:  info.ProductVersion,
		ImageVSD:        formatJSONNA(info.ImageVSD),
		DeviceVSD:       formatJSONNA(info.DeviceVSD),
		PSID:            info.PSID,
		SecurityAttrs:   formatJSONNA(info.SecurityAttrs),
		SecurityVer:     info.SecurityVer,
		DefaultUpdateMethod: "fw_ctrl",
	}
	
	// Convert ROM info
	for _, rom := range info.RomInfo {
		output.RomInfo = append(output.RomInfo, RomInfoJSON{
			Type:    rom.Type,
			Version: rom.Version,
			CPU:     rom.CPU,
		})
	}
	
	// Convert GUID info
	if info.BaseGUID != 0 {
		output.BaseGUID = &UIDInfo{
			UID:   fmt.Sprintf("%016x", info.BaseGUID),
			Count: info.BaseGUIDNum,
		}
	} else {
		output.BaseGUID = &UIDInfo{
			UID:   "N/A",
			Count: info.BaseGUIDNum,
		}
	}
	
	// Convert MAC info
	if info.BaseMAC != 0 {
		output.BaseMAC = &UIDInfo{
			UID:   fmt.Sprintf("%012x", info.BaseMAC),
			Count: info.BaseMACNum,
		}
	} else {
		output.BaseMAC = &UIDInfo{
			UID:   "N/A",
			Count: info.BaseMACNum,
		}
	}
	
	return output
}

// getCRCTypeName returns a human-readable name for CRC type
func getCRCTypeName(crcType types.CRCType) string {
	switch crcType {
	case types.CRCInITOCEntry:
		return "IN_ITOC_ENTRY"
	case types.CRCNone:
		return "NONE"
	case types.CRCInSection:
		return "IN_SECTION"
	default:
		return fmt.Sprintf("UNKNOWN_%d", crcType)
	}
}

// formatJSONNA formats empty strings as "N/A" for JSON output
func formatJSONNA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

// outputJSON outputs data as JSON
func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}