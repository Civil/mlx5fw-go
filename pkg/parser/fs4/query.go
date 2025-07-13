package fs4

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/ghostiam/binstruct"
	"go.uber.org/zap"

	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
)

// Query returns firmware information similar to mstflint query output
func (p *Parser) Query() (*interfaces.FirmwareInfo, error) {
	info := &interfaces.FirmwareInfo{
		Format:        p.GetFormat().String(),
		FormatVersion: 4,
		ImageSize:     uint64(p.reader.Size()),
		BaseGUIDNum:   8,  // Standard for most cards
		BaseMACNum:    8,  // Standard for most cards
	}

	// Get IMAGE_INFO section if available
	imageInfoSections := p.sections[types.SectionTypeImageInfo]
	if len(imageInfoSections) > 0 {
		section := imageInfoSections[0]
		p.logger.Debug("Found IMAGE_INFO section", 
			zap.Uint64("offset", section.Offset),
			zap.Uint32("size", section.Size))
		
		if section.Data == nil {
			// Read the section data if not already loaded
			data, err := p.reader.ReadSection(int64(section.Offset), section.Size)
			if err != nil {
				p.logger.Warn("Failed to read IMAGE_INFO section", zap.Error(err))
			} else {
				section.Data = data
			}
		}

		if section.Data != nil {
			// Log the actual section size
			p.logger.Debug("IMAGE_INFO section size", 
				zap.Uint32("size", section.Size),
				zap.Int("data_len", len(section.Data)))
			
			// Try parsing with the working ImageInfoBinary format first
			var imageInfoBinary types.ImageInfoBinary
			err := binstruct.UnmarshalBE(section.Data, &imageInfoBinary)
			if err != nil {
				p.logger.Warn("Failed to parse IMAGE_INFO", zap.Error(err))
			} else {
				info.FWVersion = imageInfoBinary.GetFWVersionString()
				info.FWReleaseDate = imageInfoBinary.GetFWReleaseDateString()
				// MIC version is fixed at 2.0.0 for FS4 format
				info.MICVersion = "2.0.0"
				info.PRSName = imageInfoBinary.GetPRSNameString()
				info.PartNumber = imageInfoBinary.GetPartNumberString()
				info.Description = imageInfoBinary.GetDescriptionString()
				info.PSID = imageInfoBinary.GetPSIDString()
				
				// Try to extract security attributes from the first 4 bytes
				if len(section.Data) >= 4 {
					securityAndVersion := binary.BigEndian.Uint32(section.Data[0:4])
					attrs := p.parseSecurityAttributes(securityAndVersion)
					if attrs != "" {
						info.SecurityAttrs = attrs
					}
				}
				
				// Build product version from fw version
				if info.FWVersion != "" {
					// Convert version format: "16.35.4030" -> "rel-16_35_4030"
					parts := strings.Split(info.FWVersion, ".")
					if len(parts) == 3 {
						info.ProductVersion = fmt.Sprintf("rel-%s_%s_%s", parts[0], parts[1], parts[2])
					}
				}
			}
		}
	}

	// Get GUIDs/MACs info from DEV_INFO section (0xe1 in DTOC becomes 0xe0e1)
	// Based on mstflint's fs3_ops.cpp:301 Fs3Operations::GetImageInfo
	devInfoType := uint16(types.SectionTypeDevInfo | 0xE000)
	devInfoSections := p.sections[devInfoType]
	p.logger.Debug("Looking for DEV_INFO sections", 
		zap.Int("count", len(devInfoSections)),
		zap.Uint32("type", uint32(devInfoType)))
	if len(devInfoSections) > 0 {
		// Use the first DEV_INFO section
		section := devInfoSections[0]
		p.logger.Debug("Found DEV_INFO section", 
			zap.Uint64("offset", section.Offset),
			zap.Uint32("size", section.Size))
		
		if section.Data == nil {
			// Read the section data if not already loaded
			data, err := p.reader.ReadSection(int64(section.Offset), section.Size)
			if err != nil {
				p.logger.Warn("Failed to read DEV_INFO section", zap.Error(err))
			} else {
				section.Data = data
			}
		}

		if section.Data != nil && len(section.Data) >= 0x40 {
			// Parse DEV_INFO structure
			// Note: DEV_INFO uses little-endian encoding
			var devInfo types.DevInfo
			err := binstruct.UnmarshalLE(section.Data, &devInfo)
			if err != nil {
				p.logger.Warn("Failed to parse DEV_INFO", zap.Error(err))
			} else {
				// Get GUIDs and MACs number from DEV_INFO
				info.BaseGUIDNum = devInfo.Guids.GetNumAllocated()
				info.BaseMACNum = devInfo.Macs.GetNumAllocated()
				info.BaseGUID = devInfo.Guids.GetUID()
				info.BaseMAC = devInfo.Macs.GetUID()
				p.logger.Debug("Got GUIDs/MACs from DEV_INFO",
					zap.Uint8("guid_num_allocated", devInfo.Guids.NumAllocated),
					zap.Uint8("guid_step", devInfo.Guids.Step),
					zap.Uint64("guid_uid", devInfo.Guids.UID),
					zap.Uint8("mac_num_allocated", devInfo.Macs.NumAllocated),
					zap.Uint8("mac_step", devInfo.Macs.Step),
					zap.Uint64("mac_uid", devInfo.Macs.UID),
					zap.Int("guids", info.BaseGUIDNum),
					zap.Int("macs", info.BaseMACNum))
			}
		}
	}

	// Parse ROM_CODE section for ROM info
	romCodeSections := p.sections[types.SectionTypeROMCode]
	if len(romCodeSections) > 0 {
		section := romCodeSections[0]
		if section.Data == nil {
			// Read the section data if not already loaded
			data, err := p.reader.ReadSection(int64(section.Offset), section.Size)
			if err != nil {
				p.logger.Warn("Failed to read ROM_CODE section", zap.Error(err))
			} else {
				section.Data = data
			}
		}

		if section.Data != nil {
			// Parse ROM info from ROM_CODE section
			romInfo := p.parseRomInfo(section.Data)
			if len(romInfo) > 0 {
				info.RomInfo = romInfo
			}
		}
	}

	// If DEV_INFO UIDs are empty, try to get them from MFG_INFO
	// This is the case for ConnectX7 firmware where DEV_INFO is all FFs
	if info.BaseGUID == 0 && info.BaseMAC == 0 {
		// MFG_INFO is type 0xe0 in DTOC (becomes 0xe0e0)
		mfgInfoType := uint16(0xe0 | 0xE000)
		mfgInfoSections := p.sections[mfgInfoType]
		p.logger.Debug("Looking for MFG_INFO sections", 
			zap.Int("count", len(mfgInfoSections)),
			zap.Uint32("type", uint32(mfgInfoType)))
		
		if len(mfgInfoSections) > 0 {
			section := mfgInfoSections[0]
			p.logger.Debug("Found MFG_INFO section", 
				zap.Uint64("offset", section.Offset),
				zap.Uint32("size", section.Size))
			
			if section.Data == nil {
				// Read the section data if not already loaded
				data, err := p.reader.ReadSection(int64(section.Offset), section.Size)
				if err != nil {
					p.logger.Warn("Failed to read MFG_INFO section", zap.Error(err))
				} else {
					section.Data = data
				}
			}

			if section.Data != nil && len(section.Data) >= 0x40 {
				// Parse MFG_INFO structure
				// Note: MFG_INFO uses little-endian encoding like DEV_INFO
				var mfgInfo types.MfgInfo
				err := binstruct.UnmarshalLE(section.Data, &mfgInfo)
				if err != nil {
					p.logger.Warn("Failed to parse MFG_INFO", zap.Error(err))
				} else {
					// Get GUIDs and MACs from MFG_INFO
					info.BaseGUIDNum = mfgInfo.Guids.GetNumAllocated()
					info.BaseMACNum = mfgInfo.Macs.GetNumAllocated()
					info.BaseGUID = mfgInfo.Guids.GetUID()
					info.BaseMAC = mfgInfo.Macs.GetUID()
					p.logger.Debug("Got GUIDs/MACs from MFG_INFO",
						zap.Uint8("guid_num_allocated", mfgInfo.Guids.NumAllocated),
						zap.Uint64("guid_uid", mfgInfo.Guids.GetUID()),
						zap.Uint8("mac_num_allocated", mfgInfo.Macs.NumAllocated),
						zap.Uint64("mac_uid", mfgInfo.Macs.GetUID()),
						zap.String("psid", mfgInfo.GetPSIDString()))
				}
			}
		}
	}

	// Check if firmware is encrypted
	info.IsEncrypted = p.itocHeader == nil || p.itocHeader.Signature0 != types.ITOCSignature

	// Get chunk size from HW pointers if available
	if p.hwPointers != nil {
		// Chunk size is typically log2 based, need to calculate from firmware structure
		info.ChunkSize = 0x800000 // Default 8MB for most cards
	}

	// Add section information
	for _, sections := range p.sections {
		for _, section := range sections {
			sectionInfo := interfaces.SectionInfo{
				Type:         section.Type,
				TypeName:     types.GetSectionTypeName(section.Type),
				Offset:       section.Offset,
				Size:         section.Size,
				CRCType:      section.CRCType,
				IsEncrypted:  section.IsEncrypted,
				IsDeviceData: section.IsDeviceData,
			}
			info.Sections = append(info.Sections, sectionInfo)
		}
	}

	return info, nil
}

// parseSecurityAttributes parses security attributes from the first dword of IMAGE_INFO
func (p *Parser) parseSecurityAttributes(securityAndVersion uint32) string {
	// Extract security flags from bits
	mccEn := (securityAndVersion & (1 << 8)) != 0
	debugFW := (securityAndVersion & (1 << 13)) != 0
	signedFW := (securityAndVersion & (1 << 14)) != 0
	secureFW := (securityAndVersion & (1 << 15)) != 0
	
	// Build security mode
	var mode uint32
	if mccEn {
		mode |= types.SMMFlags.MCC_EN
	}
	if debugFW {
		mode |= types.SMMFlags.DEBUG_FW
	}
	if signedFW {
		mode |= types.SMMFlags.SIGNED_FW
	}
	if secureFW {
		mode |= types.SMMFlags.SECURE_FW
	}
	
	// Build attribute string
	attrs := []string{}
	
	if mode&types.SMMFlags.SECURE_FW != 0 {
		attrs = append(attrs, "secure-fw")
	} else if mode&types.SMMFlags.SIGNED_FW != 0 {
		attrs = append(attrs, "signed-fw")
	} else {
		return "N/A"
	}
	
	if mode&types.SMMFlags.DEBUG_FW != 0 {
		attrs = append(attrs, "debug")
	}
	
	return strings.Join(attrs, ", ")
}

// parseRomInfo parses ROM information from ROM_CODE section data
// Based on mstflint's FwOperations::RomInfo::GetExpRomVersion() in fw_ops.cpp:1894
func (p *Parser) parseRomInfo(data []byte) []interfaces.RomInfo {
	var romInfoList []interfaces.RomInfo
	
	// mstflint searches for the magic string "mlxsign:" in the ROM data
	// Reference: fw_ops.cpp:1896-1904
	magicString := "mlxsign:"
	magicLen := len(magicString)
	
	// Search for magic string in ROM data
	// Reference: fw_ops.cpp:1928-1942
	for i := 0; i <= len(data)-magicLen; i++ {
		if i+magicLen > len(data) {
			break
		}
		
		// Check if we found the magic string
		found := true
		for j := 0; j < magicLen; j++ {
			if data[i+j] != magicString[j] {
				found = false
				break
			}
		}
		
		if found {
			// Parse ROM info after mlxsign:
			// Reference: fw_ops.cpp:1960-1961 - calls GetExpRomVerForOneRom
			verOffset := i + magicLen
			if romInfo := p.parseOneRomInfo(data, verOffset); romInfo != nil {
				romInfoList = append(romInfoList, *romInfo)
			}
			
			// Skip past this ROM info (ROM_INFO_SIZE = 12)
			// Reference: fw_ops.cpp:1989 and mlxfwops_com.h:151
			// Note: we add 11 because the for loop will increment i by 1
			i += 11
		}
	}
	
	return romInfoList
}

// parseOneRomInfo parses a single ROM entry after mlxsign:
// Based on mstflint's FwOperations::RomInfo::GetExpRomVerForOneRom() in fw_ops.cpp:2044
func (p *Parser) parseOneRomInfo(data []byte, verOffset int) *interfaces.RomInfo {
	if verOffset+12 > len(data) {
		return nil
	}
	
	// Get expansion ROM product ID and version info
	// Reference: fw_ops.cpp:2065-2066
	tmp := binary.LittleEndian.Uint32(data[verOffset:])
	offs4 := binary.LittleEndian.Uint32(data[verOffset+4:])
	offs8 := binary.LittleEndian.Uint32(data[verOffset+8:])
	
	productID := uint16(tmp >> 16)
	
	// Parse ROM type from product ID
	// Reference: fw_ops.cpp:2357 - expRomType2Str
	romType := ""
	switch productID {
	case 0x10:
		romType = "PXE"
	case 0x11:
		romType = "UEFI"
	case 0x12:
		romType = "CLP"
	case 0x13:
		romType = "NVMe"
	case 0xf:
		romType = "CLP"
	default:
		// Unknown type, skip
		return nil
	}
	
	// Build version string
	// Reference: fw_ops.cpp:2072-2077
	var version string
	ver0 := tmp & 0xff
	if productID != 0xf {
		ver1 := (offs4 >> 16) & 0xff
		ver2 := offs4 & 0xffff
		version = fmt.Sprintf("%d.%d.%d", ver0, ver1, ver2)
	} else {
		// For type 0xf, version is handled differently
		// Reference: fw_ops.cpp:2110-2117
		if verOffset+0x10+4 <= len(data) {
			strLen := int((data[verOffset+0xc+1]) & 0xff)
			if verOffset+0x10+strLen <= len(data) && strLen > 0 {
				version = string(data[verOffset+0x10 : verOffset+0x10+strLen])
			}
		}
	}
	
	// Get CPU architecture if product ID >= 0x10
	// Reference: fw_ops.cpp:2084-2089
	cpu := ""
	if productID >= 0x10 && verOffset+12 <= len(data) {
		suppCpuArch := (offs8 >> 8) & 0xf
		
		// Parse CPU architecture
		// Reference: mlxfwops_com.h enum ExpRomCpuArch
		switch suppCpuArch {
		case 0x0:
			// ERC_UNSPECIFIED
			cpu = ""
		case 0x1:
			// ERC_AMD64
			cpu = "AMD64"
		case 0x2:
			// ERC_AARCH64
			cpu = "AARCH64"
		case 0x3:
			// ERC_AMD64_AARCH64
			cpu = "AMD64,AARCH64"
		case 0x4:
			// ERC_IA32
			cpu = "IA32"
		}
	}
	
	return &interfaces.RomInfo{
		Type:    romType,
		Version: version,
		CPU:     cpu,
	}
}