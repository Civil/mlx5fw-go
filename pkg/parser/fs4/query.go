package fs4

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

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
		BaseGUIDNum:   0,  // Will be set from DEV_INFO or MFG_INFO
		BaseMACNum:    0,  // Will be set from DEV_INFO or MFG_INFO
	}

	// Get IMAGE_INFO section if available
	imageInfoSections := p.sections[types.SectionTypeImageInfo]
	p.logger.Debug("Looking for IMAGE_INFO sections", 
		zap.Int("count", len(imageInfoSections)),
		zap.Uint32("type", uint32(types.SectionTypeImageInfo)))
	if len(imageInfoSections) > 0 {
		section := imageInfoSections[0]
		p.logger.Debug("Found IMAGE_INFO section", 
			zap.Uint64("offset", section.Offset()),
			zap.Uint32("size", section.Size()),
			zap.String("crc_type", section.CRCType().String()))
		
		if section.GetRawData() == nil {
			// Read the section data if not already loaded
			// For sections with IN_SECTION CRC, read the extra 4 bytes to get the CRC
			readSize := section.Size()
			if section.CRCType() == types.CRCInSection && !p.IsEncrypted() {
				readSize += 4
			}
			data, err := p.reader.ReadSection(int64(section.Offset()), readSize)
			if err != nil {
				p.logger.Warn("Failed to read IMAGE_INFO section", zap.Error(err))
			} else {
				// Parse the section data (this will store it internally)
				if err := section.Parse(data); err != nil {
					p.logger.Warn("Failed to parse IMAGE_INFO section", zap.Error(err))
				}
			}
		}

		if section.GetRawData() != nil {
			// Log the actual section size
			p.logger.Debug("IMAGE_INFO section size", 
				zap.Uint32("size", section.Size()),
				zap.Int("data_len", len(section.GetRawData())))
			
			// Debug: Log first 64 bytes of data
			if len(section.GetRawData()) >= 64 {
				p.logger.Debug("IMAGE_INFO first 64 bytes",
					zap.String("hex", fmt.Sprintf("%x", section.GetRawData()[:64])))
			}
			
			// Debug: Log specific FW version fields at offsets 4, 8, 10
			if len(section.GetRawData()) >= 12 {
				data := section.GetRawData()
				p.logger.Debug("IMAGE_INFO FW version raw bytes",
					zap.String("major_at_4", fmt.Sprintf("%02x %02x", data[4], data[5])),
					zap.String("subminor_at_8", fmt.Sprintf("%02x %02x", data[8], data[9])),
					zap.String("minor_at_10", fmt.Sprintf("%02x %02x", data[10], data[11])))
			}
			
			// Check if data is all 0xFF
			allFF := true
			data := section.GetRawData()
			for i := 0; i < len(data) && i < 1024; i++ {
				if data[i] != 0xFF {
					allFF = false
					break
				}
			}
			if allFF {
				p.logger.Warn("IMAGE_INFO section contains all 0xFF bytes!")
			}
			
			// Try parsing with the ImageInfo format
			var imageInfo types.ImageInfo
			err := imageInfo.Unmarshal(section.GetRawData())
			if err != nil {
				p.logger.Warn("Failed to parse IMAGE_INFO", zap.Error(err))
			} else {
				info.FWVersion = imageInfo.GetFWVersionString()
				info.FWReleaseDate = imageInfo.GetFWReleaseDateString()
				// MIC version is fixed at 2.0.0 for FS4 format
				info.MICVersion = "2.0.0"
				info.PRSName = imageInfo.GetPRSNameString()
				info.PartNumber = imageInfo.GetPartNumberString()
				info.Description = imageInfo.GetDescriptionString()
				info.PSID = imageInfo.GetPSIDString()
				
				// Get security attributes using the parsed fields
				attrs := p.parseSecurityAttributesFromImageInfo(&imageInfo)
				if attrs != "" {
					info.SecurityAttrs = attrs
				}
				
				// Get product version from IMAGE_INFO
				productVer := imageInfo.GetProductVerString()
				if productVer != "" {
					info.ProductVersion = productVer
				}
			}
		}
	}
	
	// Get security attributes from signature sections (but don't apply yet)
	sigAttrs := p.getSignatureSecurityAttributes()

	// Get GUIDs/MACs info from DEV_INFO section (0xe1 in DTOC)
	// Based on mstflint's fs3_ops.cpp:301 Fs3Operations::GetImageInfo
	devInfoType := uint16(types.SectionTypeDevInfo)
	devInfoSections := p.sections[devInfoType]
	p.logger.Debug("Looking for DEV_INFO sections", 
		zap.Int("count", len(devInfoSections)),
		zap.Uint32("type", uint32(devInfoType)))
	if len(devInfoSections) > 0 {
		// Use the first DEV_INFO section
		section := devInfoSections[0]
		p.logger.Debug("Found DEV_INFO section", 
			zap.Uint64("offset", section.Offset()),
			zap.Uint32("size", section.Size()))
		
		if section.GetRawData() == nil {
			// Read the section data if not already loaded
			// For sections with IN_SECTION CRC, read the extra 4 bytes to get the CRC
			readSize := section.Size()
			if section.CRCType() == types.CRCInSection && !p.IsEncrypted() {
				readSize += 4
			}
			data, err := p.reader.ReadSection(int64(section.Offset()), readSize)
			if err != nil {
				p.logger.Warn("Failed to read DEV_INFO section", zap.Error(err))
			} else {
				// Parse the section data (this will store it internally)
				if err := section.Parse(data); err != nil {
					p.logger.Warn("Failed to parse DEV_INFO section", zap.Error(err))
				}
			}
		}

		if section.GetRawData() != nil && len(section.GetRawData()) >= 0x40 {
			// Parse DEV_INFO structure
			// Note: DEV_INFO uses little-endian encoding
			var devInfo types.DevInfo
			err := devInfo.Unmarshal(section.GetRawData())
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
	p.logger.Debug("Looking for ROM_CODE sections", 
		zap.Int("count", len(romCodeSections)))
	if len(romCodeSections) > 0 {
		section := romCodeSections[0]
		p.logger.Debug("Found ROM_CODE section", 
			zap.Uint64("offset", section.Offset()),
			zap.Uint32("size", section.Size()))
		if section.GetRawData() == nil {
			// Read the section data if not already loaded
			// For sections with IN_SECTION CRC, read the extra 4 bytes to get the CRC
			readSize := section.Size()
			if section.CRCType() == types.CRCInSection && !p.IsEncrypted() {
				readSize += 4
			}
			data, err := p.reader.ReadSection(int64(section.Offset()), readSize)
			if err != nil {
				p.logger.Warn("Failed to read ROM_CODE section", zap.Error(err))
			} else {
				// Parse the section data (this will store it internally)
				if err := section.Parse(data); err != nil {
					p.logger.Warn("Failed to parse ROM_CODE section", zap.Error(err))
				}
			}
		}

		if section.GetRawData() != nil {
			// Parse ROM info from ROM_CODE section
			p.logger.Debug("Parsing ROM info", 
				zap.Int("data_len", len(section.GetRawData())))
			romInfo := p.parseRomInfo(section.GetRawData())
			p.logger.Debug("Parsed ROM info entries", 
				zap.Int("count", len(romInfo)))
			if len(romInfo) > 0 {
				info.RomInfo = romInfo
			}
		}
	}

	// If DEV_INFO UIDs are empty, try to get them from MFG_INFO
	// This is the case for ConnectX7 firmware where DEV_INFO is all FFs
	if info.BaseGUID == 0 && info.BaseMAC == 0 {
		// MFG_INFO is type 0xe0 in DTOC
		mfgInfoType := uint16(types.SectionTypeMfgInfo)
		mfgInfoSections := p.sections[mfgInfoType]
		p.logger.Debug("Looking for MFG_INFO sections because UIDs are 0", 
			zap.Int("count", len(mfgInfoSections)),
			zap.Uint32("type", uint32(mfgInfoType)))
		
		if len(mfgInfoSections) > 0 {
			section := mfgInfoSections[0]
			p.logger.Debug("Found MFG_INFO section", 
				zap.Uint64("offset", section.Offset()),
				zap.Uint32("size", section.Size()))
			
			if section.GetRawData() == nil {
				// Read the section data if not already loaded
				// For sections with IN_SECTION CRC, read the extra 4 bytes to get the CRC
				readSize := section.Size()
				if section.CRCType() == types.CRCInSection && !p.IsEncrypted() {
					readSize += 4
				}
				data, err := p.reader.ReadSection(int64(section.Offset()), readSize)
				if err != nil {
					p.logger.Warn("Failed to read MFG_INFO section", zap.Error(err))
				} else {
					// Parse the section data (this will store it internally)
					if err := section.Parse(data); err != nil {
						p.logger.Warn("Failed to parse MFG_INFO section", zap.Error(err))
					}
				}
			}

			if section.GetRawData() != nil && len(section.GetRawData()) >= 0x40 {
				// Parse MFG_INFO structure
				// Note: MFG_INFO uses little-endian encoding like DEV_INFO
				var mfgInfo types.MfgInfo
				err := mfgInfo.Unmarshal(section.GetRawData())
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
						zap.String("psid", mfgInfo.GetPSIDString()),
						zap.Int("final_guids", info.BaseGUIDNum),
						zap.Int("final_macs", info.BaseMACNum))
				}
			}
		}
	}

	// Check if firmware is encrypted
	info.IsEncrypted = p.itocHeader == nil || p.itocHeader.Signature0 != types.ITOCSignature
	
	// Debug logging for encryption detection
	p.logger.Debug("Encryption detection", 
		zap.Bool("is_encrypted", info.IsEncrypted),
		zap.Bool("itoc_header_nil", p.itocHeader == nil),
		zap.Uint32("itoc_signature", func() uint32 {
			if p.itocHeader != nil {
				return p.itocHeader.Signature0
			}
			return 0
		}()),
		zap.Uint32("expected_signature", types.ITOCSignature))
	
	// Apply security attributes based on encryption status and signature sections
	// For encrypted firmware, mstflint returns "secure-fw" regardless of signature sections
	if info.IsEncrypted {
		if info.SecurityAttrs == "N/A" || info.SecurityAttrs == "" {
			info.SecurityAttrs = "secure-fw"
		}
	} else if sigAttrs != "" {
		if info.SecurityAttrs == "N/A" || info.SecurityAttrs == "" {
			// If IMAGE_INFO didn't provide attributes, use signature attributes
			info.SecurityAttrs = sigAttrs
		} else {
			// Merge with existing attributes
			existingAttrs := strings.Split(info.SecurityAttrs, ", ")
			sigAttrsList := strings.Split(sigAttrs, ", ")
			
			// Create a map to avoid duplicates
			attrMap := make(map[string]bool)
			for _, attr := range existingAttrs {
				attrMap[attr] = true
			}
			for _, attr := range sigAttrsList {
				attrMap[attr] = true
			}
			
			// Build combined list
			var combinedAttrs []string
			// Order matters: secure-fw, signed-fw, debug, dev
			if attrMap["secure-fw"] {
				combinedAttrs = append(combinedAttrs, "secure-fw")
			} else if attrMap["signed-fw"] {
				combinedAttrs = append(combinedAttrs, "signed-fw")
			}
			if attrMap["debug"] {
				combinedAttrs = append(combinedAttrs, "debug")
			}
			if attrMap["dev"] {
				combinedAttrs = append(combinedAttrs, "dev")
			}
			
			if len(combinedAttrs) > 0 {
				info.SecurityAttrs = strings.Join(combinedAttrs, ", ")
			}
		}
	}
	
	// Determine if we should use dual GUID/MAC format
	// This is used for certain encrypted firmwares (like CX7) that don't have DEV_INFO sections
	// The check is: encrypted AND no DEV_INFO sections found
	devInfoCount := 0
	for _, sections := range p.sections {
		for _, section := range sections {
			if section.Type() == types.SectionTypeDevInfo {
				devInfoCount++
			}
		}
	}
	
	if info.IsEncrypted && devInfoCount == 0 {
		info.UseDualFormat = true
		// For encrypted firmware without GUID/MAC data, all values are N/A
		info.GUIDStep = 0
		info.MACStep = 0
	}

	// Get chunk size from HW pointers if available
	if p.hwPointers != nil {
		// Chunk size is typically log2 based, need to calculate from firmware structure
		info.ChunkSize = 0x800000 // Default 8MB for most cards
	}

	// Add section information
	for _, sections := range p.sections {
		for _, section := range sections {
			sectionInfo := interfaces.SectionInfo{
				Type:         section.Type(),
				TypeName:     types.GetSectionTypeName(section.Type()),
				Offset:       section.Offset(),
				Size:         section.Size(),
				CRCType:      section.CRCType(),
				IsEncrypted:  section.IsEncrypted(),
				IsDeviceData: section.IsDeviceData(),
			}
			info.Sections = append(info.Sections, sectionInfo)
		}
	}

	return info, nil
}

// parseSecurityAttributesFromImageInfo parses security attributes from parsed IMAGE_INFO
// This replicates the original parseSecurityAttributes logic exactly
func (p *Parser) parseSecurityAttributesFromImageInfo(imageInfo *types.ImageInfo) string {
	// Build security mode using the same logic as original parseSecurityAttributes
	var mode uint32
	if imageInfo.IsMCCEnabled() {
		mode |= types.SMMFlags.MCC_EN
	}
	if imageInfo.IsDebugFW() {
		mode |= types.SMMFlags.DEBUG_FW
	}
	if imageInfo.IsSignedFW() {
		mode |= types.SMMFlags.SIGNED_FW
	}
	if imageInfo.IsSecureFW() {
		mode |= types.SMMFlags.SECURE_FW
	}
	
	// Check IMAGE_SIGNATURE sections for dev_fw flag
	// Based on mstflint's Fs3Operations::GetImgSigInfo
	devFw := p.checkDevFwFromSignature()
	if devFw {
		mode |= types.SMMFlags.DEV_FW
	}
	
	// Build attribute string using exact same logic as original
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
	
	if mode&types.SMMFlags.DEV_FW != 0 {
		attrs = append(attrs, "dev")
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
	
	p.logger.Debug("Searching for ROM signatures", 
		zap.String("magic", magicString),
		zap.Int("data_len", len(data)))
	
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
			p.logger.Debug("Found ROM signature", 
				zap.Int("offset", i))
			// Parse ROM info after mlxsign:
			// Reference: fw_ops.cpp:1960-1961 - calls GetExpRomVerForOneRom
			verOffset := i + magicLen
			if romInfo := p.parseOneRomInfo(data, verOffset); romInfo != nil {
				p.logger.Debug("Parsed ROM entry", 
					zap.String("type", romInfo.Type),
					zap.String("version", romInfo.Version),
					zap.String("cpu", romInfo.CPU))
				romInfoList = append(romInfoList, *romInfo)
			}
			
			// Skip past this ROM info (ROM_INFO_SIZE = 12)
			// Reference: fw_ops.cpp:1989 and mlxfwops_com.h:151
			// Note: we add 11 because the for loop will increment i by 1
			i += 11
		}
	}
	
	p.logger.Debug("Finished parsing ROM info", 
		zap.Int("entries_found", len(romInfoList)))
	
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
	p.logger.Debug("ROM product ID", 
		zap.Uint16("product_id", productID),
		zap.String("hex", fmt.Sprintf("0x%x", productID)))
	
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
	case 0x14:
		romType = "UEFI Virtio net"
	case 0x15:
		romType = "UEFI Virtio blk"
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

// checkDevFwFromSignature checks IMAGE_SIGNATURE sections for dev_fw flag
// Based on mstflint's Fs3Operations::GetImgSigInfo
func (p *Parser) checkDevFwFromSignature() bool {
	// Check IMAGE_SIGNATURE_256 sections
	sig256Sections := p.sections[types.SectionTypeImageSignature256]
	for _, section := range sig256Sections {
		if section.GetRawData() == nil {
			// Read the section data if not already loaded
			// For sections with IN_SECTION CRC, read the extra 4 bytes to get the CRC
			readSize := section.Size()
			if section.CRCType() == types.CRCInSection && !p.IsEncrypted() {
				readSize += 4
			}
			data, err := p.reader.ReadSection(int64(section.Offset()), readSize)
			if err != nil {
				p.logger.Warn("Failed to read IMAGE_SIGNATURE_256 section", zap.Error(err))
				continue
			}
			// Parse the section data (this will store it internally)
			if err := section.Parse(data); err != nil {
				p.logger.Warn("Failed to parse IMAGE_SIGNATURE_256 section", zap.Error(err))
				continue
			}
		}
		
		if section.GetRawData() != nil && len(section.GetRawData()) >= 0x20 { // Need at least 32 bytes for keypair_uuid
			// Parse the signature structure
			var sig types.FS4ImageSignatureStruct
			if err := binary.Read(bytes.NewReader(section.GetRawData()), binary.BigEndian, &sig); err != nil {
				p.logger.Warn("Failed to parse IMAGE_SIGNATURE_256", zap.Error(err))
				continue
			}
			
			// Check keypair_uuid condition from mstflint
			// if (((keypair_uuid[0] != 0) || (keypair_uuid[1] != 0) || (keypair_uuid[2] != 0)) && 
			//     (keypair_uuid[3] == 0) && ((EXTRACT(keypair_uuid[2], 0, 16)) == 0))
			if (sig.KeypairUUID[0] != 0 || sig.KeypairUUID[1] != 0 || sig.KeypairUUID[2] != 0) &&
			   sig.KeypairUUID[3] == 0 && (sig.KeypairUUID[2] & 0xFFFF) == 0 {
				return true
			}
		}
	}
	
	// Check IMAGE_SIGNATURE_512 sections
	sig512Sections := p.sections[types.SectionTypeImageSignature512]
	for _, section := range sig512Sections {
		if section.GetRawData() == nil {
			// Read the section data if not already loaded
			// For sections with IN_SECTION CRC, read the extra 4 bytes to get the CRC
			readSize := section.Size()
			if section.CRCType() == types.CRCInSection && !p.IsEncrypted() {
				readSize += 4
			}
			data, err := p.reader.ReadSection(int64(section.Offset()), readSize)
			if err != nil {
				p.logger.Warn("Failed to read IMAGE_SIGNATURE_512 section", zap.Error(err))
				continue
			}
			// Parse the section data (this will store it internally)
			if err := section.Parse(data); err != nil {
				p.logger.Warn("Failed to parse IMAGE_SIGNATURE_512 section", zap.Error(err))
				continue
			}
		}
		
		if section.GetRawData() != nil && len(section.GetRawData()) >= 0x20 { // Need at least 32 bytes for keypair_uuid
			// Parse the signature structure
			var sig types.FS4ImageSignature2Struct
			if err := binary.Read(bytes.NewReader(section.GetRawData()), binary.BigEndian, &sig); err != nil {
				p.logger.Warn("Failed to parse IMAGE_SIGNATURE_512", zap.Error(err))
				continue
			}
			
			// Check keypair_uuid condition from mstflint
			if (sig.KeypairUUID[0] != 0 || sig.KeypairUUID[1] != 0 || sig.KeypairUUID[2] != 0) &&
			   sig.KeypairUUID[3] == 0 && (sig.KeypairUUID[2] & 0xFFFF) == 0 {
				return true
			}
		}
	}
	
	return false
}

// getSignatureSecurityAttributes gets security attributes from IMAGE_SIGNATURE sections
// This only returns dev attribute from signature analysis, matching original behavior
func (p *Parser) getSignatureSecurityAttributes() string {
	devFw := p.checkDevFwFromSignature()
	
	// Debug logging
	p.logger.Debug("getSignatureSecurityAttributes", 
		zap.Bool("devFw", devFw),
		zap.Int("sig256", len(p.sections[types.SectionTypeImageSignature256])),
		zap.Int("sig512", len(p.sections[types.SectionTypeImageSignature512])),
		zap.Int("sig256", len(p.sections[types.SectionTypeImageSignature256])),
		zap.Int("sig512", len(p.sections[types.SectionTypeImageSignature512])))
	
	// Only return dev attribute if found - this matches the original behavior
	if devFw {
		return "dev"
	}
	
	return ""
}
