package reassemble

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/Civil/mlx5fw-go/pkg/types"
)

// reconstructImageInfo reconstructs IMAGE_INFO from JSON
func (r *Reassembler) reconstructImageInfo(jsonMap map[string]interface{}) ([]byte, error) {
	// Extract the image_info object
	imageInfoData, ok := jsonMap["image_info"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing image_info field in JSON")
	}
	
	// Create and populate ImageInfo structure
	info := &types.ImageInfo{}
	
	// Populate fields from JSON
	// Reconstruct individual bitfields from security_and_version
	if securityAndVersion, ok := imageInfoData["security_and_version"].(float64); ok {
		sv := uint32(securityAndVersion)
		info.MinorVersion = uint8(sv & 0xFF)
		info.SecurityMode = uint8((sv >> 8) & 0xFF)
		info.Reserved0 = uint8((sv >> 16) & 0xFF)
		
		// Extract byte 3 which contains MajorVersion and flags
		byte3 := uint8((sv >> 24) & 0xFF)
		info.SignedFW = (byte3 & (1 << 0)) != 0  // Bit 24 of dword = bit 0 of byte 3
		info.SecureFW = (byte3 & (1 << 1)) != 0  // Bit 25 of dword = bit 1 of byte 3
		info.MCCEnabled = (byte3 & (1 << 6)) != 0  // Bit 30 of dword = bit 6 of byte 3
		info.DebugFW = (byte3 & (1 << 7)) != 0  // Bit 31 of dword = bit 7 of byte 3
		
		// MajorVersion is bits 26-29 (bits 2-5 of byte 3, 4 bits only)
		info.MajorVersion = (byte3 >> 2) & 0x0F
	}
	
	// FW version fields
	if fwVerMajor, ok := imageInfoData["fw_ver_major"].(float64); ok {
		info.FWVerMajor = uint16(fwVerMajor)
	}
	if reserved2, ok := imageInfoData["reserved2"].(float64); ok {
		info.Reserved2 = uint16(reserved2)
	}
	if fwVerSubminor, ok := imageInfoData["fw_ver_subminor"].(float64); ok {
		info.FWVerSubminor = uint16(fwVerSubminor)
	}
	if fwVerMinor, ok := imageInfoData["fw_ver_minor"].(float64); ok {
		info.FWVerMinor = uint16(fwVerMinor)
	}
	
	// Date/time fields
	if reserved3a, ok := imageInfoData["reserved3a"].(float64); ok {
		info.Reserved3a = uint8(reserved3a)
	}
	if hour, ok := imageInfoData["hour"].(float64); ok {
		info.Hour = uint8(hour)
	}
	if minutes, ok := imageInfoData["minutes"].(float64); ok {
		info.Minutes = uint8(minutes)
	}
	if seconds, ok := imageInfoData["seconds"].(float64); ok {
		info.Seconds = uint8(seconds)
	}
	if day, ok := imageInfoData["day"].(float64); ok {
		info.Day = uint8(day)
	}
	if month, ok := imageInfoData["month"].(float64); ok {
		info.Month = uint8(month)
	}
	if year, ok := imageInfoData["year"].(float64); ok {
		info.Year = uint16(year)
	}
	
	// MIC version fields
	if micVerMajor, ok := imageInfoData["mic_ver_major"].(float64); ok {
		info.MICVerMajor = uint16(micVerMajor)
	}
	if reserved4, ok := imageInfoData["reserved4"].(float64); ok {
		info.Reserved4 = uint16(reserved4)
	}
	if micVerSubminor, ok := imageInfoData["mic_ver_subminor"].(float64); ok {
		info.MICVerSubminor = uint16(micVerSubminor)
	}
	if micVerMinor, ok := imageInfoData["mic_ver_minor"].(float64); ok {
		info.MICVerMinor = uint16(micVerMinor)
	}
	
	// PCI IDs
	if pciDeviceID, ok := imageInfoData["pci_device_id"].(float64); ok {
		info.PCIDeviceID = uint16(pciDeviceID)
	}
	if pciVendorID, ok := imageInfoData["pci_vendor_id"].(float64); ok {
		info.PCIVendorID = uint16(pciVendorID)
	}
	if pciSubsystemID, ok := imageInfoData["pci_subsystem_id"].(float64); ok {
		info.PCISubsystemID = uint16(pciSubsystemID)
	}
	if pciSubVendorID, ok := imageInfoData["pci_subvendor_id"].(float64); ok {
		info.PCISubVendorID = uint16(pciSubVendorID)
	}
	
	// PSID
	if psid, ok := imageInfoData["psid"].(string); ok {
		psidBytes := []byte(psid)
		if len(psidBytes) > 16 {
			psidBytes = psidBytes[:16]
		}
		copy(info.PSID[:], psidBytes)
	}
	
	// Reserved5a and VSD vendor ID
	if reserved5a, ok := imageInfoData["reserved5a"].(float64); ok {
		info.Reserved5a = uint16(reserved5a)
	}
	if vsdVendorID, ok := imageInfoData["vsd_vendor_id"].(float64); ok {
		info.VSDVendorID = uint16(vsdVendorID)
	}
	
	// VSD
	if vsd, ok := imageInfoData["vsd"].(string); ok {
		vsdBytes := []byte(vsd)
		if len(vsdBytes) > 208 {
			vsdBytes = vsdBytes[:208]
		}
		copy(info.VSD[:], vsdBytes)
	}
	
	// Image size data
	if imageSizeData, ok := imageInfoData["image_size_data"].([]interface{}); ok && len(imageSizeData) == 8 {
		for i, val := range imageSizeData {
			if b, ok := val.(float64); ok {
				info.ImageSizeData[i] = uint8(b)
			}
		}
	}
	
	// Reserved6
	if reserved6, ok := imageInfoData["reserved6"].([]interface{}); ok && len(reserved6) == 8 {
		for i, val := range reserved6 {
			if b, ok := val.(float64); ok {
				info.Reserved6[i] = uint8(b)
			}
		}
	}
	
	// Supported HW IDs
	if supportedHWIDs, ok := imageInfoData["supported_hw_ids"].([]interface{}); ok && len(supportedHWIDs) == 4 {
		for i, val := range supportedHWIDs {
			if id, ok := val.(float64); ok {
				info.SupportedHWID[i] = uint32(id)
			}
		}
	}
	
	// INI file num
	if iniFileNum, ok := imageInfoData["ini_file_num"].(float64); ok {
		info.INIFileNum = uint32(iniFileNum)
	}
	
	// Reserved7
	if reserved7, ok := imageInfoData["reserved7"].([]interface{}); ok && len(reserved7) == 148 {
		for i, val := range reserved7 {
			if b, ok := val.(float64); ok {
				info.Reserved7[i] = uint8(b)
			}
		}
	}
	
	// Product version - use raw bytes if available
	if productVerRaw, ok := imageInfoData["product_ver_raw"].([]interface{}); ok && len(productVerRaw) == 16 {
		for i, val := range productVerRaw {
			if b, ok := val.(float64); ok {
				info.ProductVer[i] = uint8(b)
			}
		}
	} else if productVer, ok := imageInfoData["product_ver"].(string); ok {
		// Fallback to string version
		prodVerBytes := []byte(productVer)
		if len(prodVerBytes) > 16 {
			prodVerBytes = prodVerBytes[:16]
		}
		copy(info.ProductVer[:], prodVerBytes)
	}
	
	// Description
	if description, ok := imageInfoData["description"].(string); ok {
		descBytes := []byte(description)
		if len(descBytes) > 256 {
			descBytes = descBytes[:256]
		}
		copy(info.Description[:], descBytes)
	}
	
	// Reserved8
	if reserved8, ok := imageInfoData["reserved8"].([]interface{}); ok && len(reserved8) == 48 {
		for i, val := range reserved8 {
			if b, ok := val.(float64); ok {
				info.Reserved8[i] = uint8(b)
			}
		}
	}
	
	// Module versions
	if moduleVersions, ok := imageInfoData["module_versions"].([]interface{}); ok && len(moduleVersions) == 64 {
		for i, val := range moduleVersions {
			if b, ok := val.(float64); ok {
				info.ModuleVersions[i] = uint8(b)
			}
		}
	}
	
	// Name
	if name, ok := imageInfoData["name"].(string); ok {
		nameBytes := []byte(name)
		if len(nameBytes) > 64 {
			nameBytes = nameBytes[:64]
		}
		copy(info.Name[:], nameBytes)
	}
	
	// PRS name
	if prsName, ok := imageInfoData["prs_name"].(string); ok {
		prsBytes := []byte(prsName)
		if len(prsBytes) > 128 {
			prsBytes = prsBytes[:128]
		}
		copy(info.PRSName[:], prsBytes)
	}
	
	// Marshal the structure to binary
	data, err := info.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal IMAGE_INFO: %w", err)
	}
	
	// Ensure size is 1024 bytes (standard IMAGE_INFO size)
	if len(data) < 1024 {
		paddedData := make([]byte, 1024)
		copy(paddedData, data)
		data = paddedData
	}
	
	r.logger.Info("Reconstructed IMAGE_INFO section from JSON",
		zap.Int("size", len(data)))
	
	return data, nil
}