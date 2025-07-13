package types

import (
	"testing"
)

func TestGetSectionTypeName(t *testing.T) {
	tests := []struct {
		name        string
		sectionType uint16
		want        string
	}{
		// Common ITOC sections
		{name: "BOOT_CODE", sectionType: SectionTypeBootCode, want: "BOOT_CODE"},
		{name: "PCI_CODE", sectionType: SectionTypePCICode, want: "PCI_CODE"},
		{name: "MAIN_CODE", sectionType: SectionTypeMainCode, want: "MAIN_CODE"},
		{name: "PCIE_LINK_CODE", sectionType: SectionTypePCIELinkCode, want: "PCIE_LINK_CODE"},
		{name: "IRON_PREP_CODE", sectionType: SectionTypeIronPrepCode, want: "IRON_PREP_CODE"},
		{name: "POST_IRON_BOOT_CODE", sectionType: SectionTypePostIronBootCode, want: "POST_IRON_BOOT_CODE"},
		{name: "UPGRADE_CODE", sectionType: SectionTypeUpgradeCode, want: "UPGRADE_CODE"},
		{name: "HW_BOOT_CFG", sectionType: SectionTypeHWBootCfg, want: "HW_BOOT_CFG"},
		{name: "HW_MAIN_CFG", sectionType: SectionTypeHWMainCfg, want: "HW_MAIN_CFG"},
		{name: "PHY_UC_CODE", sectionType: SectionTypePhyUCCode, want: "PHY_UC_CODE"},
		{name: "PHY_UC_CONSTS", sectionType: SectionTypePhyUCConsts, want: "PHY_UC_CONSTS"},
		{name: "PCIE_PHY_UC_CODE", sectionType: SectionTypePCIEPhyUCCode, want: "PCIE_PHY_UC_CODE"},
		{name: "CCIR_INFRA_CODE", sectionType: SectionTypeCCIRInfraCode, want: "CCIR_INFRA_CODE"},
		{name: "CCIR_ALGO_CODE", sectionType: SectionTypeCCIRAlgoCode, want: "CCIR_ALGO_CODE"},
		{name: "BOOT3_CODE", sectionType: SectionTypeBoot3Code, want: "BOOT3_CODE"},
		{name: "IMAGE_INFO", sectionType: SectionTypeImageInfo, want: "IMAGE_INFO"},
		{name: "FW_BOOT_CFG", sectionType: SectionTypeFWBootCfg, want: "FW_BOOT_CFG"},
		{name: "FW_MAIN_CFG", sectionType: SectionTypeFWMainCfg, want: "FW_MAIN_CFG"},
		{name: "APU_KERNEL", sectionType: SectionTypeAPUKernel, want: "APU_KERNEL"},
		{name: "ACE_CODE", sectionType: SectionTypeACECode, want: "ACE_CODE"},
		{name: "ROM_CODE", sectionType: SectionTypeROMCode, want: "ROM_CODE"},
		{name: "RESET_INFO", sectionType: SectionTypeResetInfo, want: "RESET_INFO"},
		{name: "DBG_FW_INI", sectionType: SectionTypeDbgFWINI, want: "DBG_FW_INI"},
		{name: "DBG_FW_PARAMS", sectionType: SectionTypeDbgFWParams, want: "DBG_FW_PARAMS"},
		{name: "FW_ADB", sectionType: SectionTypeFWAdb, want: "FW_ADB"},

		// Security sections
		{name: "IMAGE_SIGNATURE_256", sectionType: SectionTypeImageSignature256, want: "IMAGE_SIGNATURE_256"},
		{name: "PUBLIC_KEYS_2048", sectionType: SectionTypePublicKeys2048, want: "PUBLIC_KEYS_2048"},
		{name: "FORBIDDEN_VERSIONS", sectionType: SectionTypeForbiddenVersions, want: "FORBIDDEN_VERSIONS"},
		{name: "IMAGE_SIGNATURE_512", sectionType: SectionTypeImageSignature512, want: "IMAGE_SIGNATURE_512"},
		{name: "PUBLIC_KEYS_4096", sectionType: SectionTypePublicKeys4096, want: "PUBLIC_KEYS_4096"},
		{name: "HMAC_DIGEST", sectionType: SectionTypeHMACDigest, want: "HMAC_DIGEST"},
		{name: "RSA_PUBLIC_KEY", sectionType: SectionTypeRsaPublicKey, want: "RSA_PUBLIC_KEY"},
		{name: "RSA_4096_SIGNATURES", sectionType: SectionTypeRsa4096Signatures, want: "RSA_4096_SIGNATURES"},
		{name: "ENCRYPTION_KEY_TRANSITION", sectionType: SectionTypeEncryptionKeyTransition, want: "ENCRYPTION_KEY_TRANSITION"},

		// Other sections
		{name: "PXIR_INI", sectionType: SectionTypePxirIni, want: "PXIR_INI"},
		{name: "PXIR_INI1", sectionType: SectionTypePxirIni1, want: "PXIR_INI1"},
		{name: "NVDA_ROT_CERTIFICATES", sectionType: SectionTypeNvdaRotCertificates, want: "NVDA_ROT_CERTIFICATES"},
		{name: "EXCLKSYNC_INFO", sectionType: SectionTypeExclkSyncInfo, want: "EXCLKSYNC_INFO"},
		{name: "MAIN_PAGES_HASHES", sectionType: SectionTypeMainPagesHashes, want: "MAIN_PAGES_HASHES"},
		{name: "MAIN_PAGES_LOCKED_HASHES", sectionType: SectionTypeMainPagesLockedHashes, want: "MAIN_PAGES_LOCKED_HASHES"},
		{name: "STRN_MAIN", sectionType: SectionTypeStrnMain, want: "STRN_MAIN"},
		{name: "STRN_IRON", sectionType: SectionTypeStrnIron, want: "STRN_IRON"},
		{name: "STRN_TILE", sectionType: SectionTypeStrnTile, want: "STRN_TILE"},
		{name: "MAIN_DATA", sectionType: SectionTypeMainData, want: "MAIN_DATA"},
		{name: "FW_DEBUG_DUMP_2", sectionType: SectionTypeFwDebugDump2, want: "FW_DEBUG_DUMP_2"},
		{name: "SECURITY_LOG", sectionType: SectionTypeSecurityLog, want: "SECURITY_LOG"},

		// Device data sections
		{name: "MFG_INFO", sectionType: SectionTypeMfgInfo, want: "MFG_INFO"},
		{name: "DEV_INFO", sectionType: SectionTypeDevInfo, want: "DEV_INFO"},
		{name: "NV_DATA1", sectionType: SectionTypeNvData1, want: "NV_DATA"},
		{name: "VPD_R0", sectionType: SectionTypeVpdR0, want: "VPD_R0"},
		{name: "NV_DATA2", sectionType: SectionTypeNvData2, want: "NV_DATA"},
		{name: "FW_NV_LOG", sectionType: SectionTypeFwNvLog, want: "FW_NV_LOG"},
		{name: "NV_DATA0", sectionType: SectionTypeNvData0, want: "NV_DATA"},
		{name: "DEV_INFO1", sectionType: SectionTypeDevInfo1, want: "DEV_INFO"},
		{name: "DEV_INFO2", sectionType: SectionTypeDevInfo2, want: "DEV_INFO"},
		{name: "CRDUMP_MASK_DATA", sectionType: SectionTypeCRDumpMaskData, want: "CRDUMP_MASK_DATA"},
		{name: "FW_INTERNAL_USAGE", sectionType: SectionTypeFwInternalUsage, want: "FW_INTERNAL_USAGE"},
		{name: "PROGRAMMABLE_HW_FW1", sectionType: SectionTypeProgrammableHwFw1, want: "PROGRAMMABLE_HW_FW"},
		{name: "PROGRAMMABLE_HW_FW2", sectionType: SectionTypeProgrammableHwFw2, want: "PROGRAMMABLE_HW_FW"},

		// Certificate sections
		{name: "DIGITAL_CERT_PTR", sectionType: SectionTypeDigitalCertPtr, want: "DIGITAL_CERT_PTR"},
		{name: "DIGITAL_CERT_RW", sectionType: SectionTypeDigitalCertRw, want: "DIGITAL_CERT_RW"},
		{name: "LC_INI1_TABLE", sectionType: SectionTypeLcIni1Table, want: "LC_INI1_TABLE"},
		{name: "LC_INI2_TABLE", sectionType: SectionTypeLcIni2Table, want: "LC_INI2_TABLE"},
		{name: "LC_INI_NV_DATA", sectionType: SectionTypeLcIniNvData, want: "LC_INI_NV_DATA"},
		{name: "CERT_CHAIN_0", sectionType: SectionTypeCertChain0, want: "CERT_CHAIN_0"},
		{name: "DIGITAL_CACERT_RW", sectionType: SectionTypeDigitalCaCertRw, want: "DIGITAL_CACERT_RW"},
		{name: "CERTIFICATE_CHAINS_1", sectionType: SectionTypeCertificateChains1, want: "CERTIFICATE_CHAINS_1"},
		{name: "CERTIFICATE_CHAINS_2", sectionType: SectionTypeCertificateChains2, want: "CERTIFICATE_CHAINS_2"},
		{name: "ROOT_CERTIFICATES_1", sectionType: SectionTypeRootCertificates1, want: "ROOT_CERTIFICATES_1"},
		{name: "ROOT_CERTIFICATES_2", sectionType: SectionTypeRootCertificates2, want: "ROOT_CERTIFICATES_2"},

		// Special sections
		{name: "TOOLS_AREA", sectionType: SectionTypeToolsArea, want: "TOOLS_AREA"},
		{name: "HASHES_TABLE", sectionType: SectionTypeHashesTable, want: "HASHES_TABLE"},
		{name: "HW_PTR", sectionType: SectionTypeHwPtr, want: "HW_PTR"},
		{name: "FW_DEBUG_DUMP", sectionType: SectionTypeFwDebugDump, want: "FW_DEBUG_DUMP"},
		{name: "ITOC", sectionType: SectionTypeItoc, want: "ITOC"},
		{name: "DTOC", sectionType: SectionTypeDtoc, want: "DTOC"},
		{name: "END", sectionType: SectionTypeEnd, want: "END"},
		{name: "BOOT2", sectionType: SectionTypeBoot2, want: "BOOT2"},

		// Unknown section
		{name: "Unknown section 0x99", sectionType: 0x99, want: "UNKNOWN_0x99"},
		{name: "Unknown section 0xBB", sectionType: 0xBB, want: "UNKNOWN_0xBB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSectionTypeName(tt.sectionType); got != tt.want {
				t.Errorf("GetSectionTypeName(0x%02X) = %v, want %v", tt.sectionType, got, tt.want)
			}
		})
	}
}

func TestGetDTOCSectionTypeName(t *testing.T) {
	tests := []struct {
		name        string
		sectionType uint8
		want        string
	}{
		// DTOC sections with 0xE prefix
		{name: "MFG_INFO", sectionType: 0xE0, want: "MFG_INFO"},
		{name: "DEV_INFO", sectionType: 0xE1, want: "DEV_INFO"},
		{name: "NV_DATA E2", sectionType: 0xE2, want: "NV_DATA"},
		{name: "VPD_R0", sectionType: 0xE3, want: "VPD_R0"},
		{name: "NV_DATA E4", sectionType: 0xE4, want: "NV_DATA"},
		{name: "FW_NV_LOG", sectionType: 0xE5, want: "FW_NV_LOG"},
		{name: "NV_DATA E6", sectionType: 0xE6, want: "NV_DATA"},
		{name: "FW_INTERNAL_USAGE EA", sectionType: 0xEA, want: "FW_INTERNAL_USAGE"},
		{name: "PROGRAMMABLE_HW_FW EB", sectionType: 0xEB, want: "PROGRAMMABLE_HW_FW"},
		{name: "PROGRAMMABLE_HW_FW EC", sectionType: 0xEC, want: "PROGRAMMABLE_HW_FW"},
		{name: "DIGITAL_CERT_PTR ED", sectionType: 0xED, want: "DIGITAL_CERT_PTR"},
		{name: "DIGITAL_CERT_RW EE", sectionType: 0xEE, want: "DIGITAL_CERT_RW"},
		{name: "CERT_CHAIN_0 EF", sectionType: 0xEF, want: "CERT_CHAIN_0"},

		// DTOC sections with other prefixes
		{name: "SECURITY_LOG", sectionType: 0xD5, want: "SECURITY_LOG"},
		{name: "CERTIFICATE_CHAINS_1 30", sectionType: 0x30, want: "CERTIFICATE_CHAINS_1"},
		{name: "CERTIFICATE_CHAINS_2 31", sectionType: 0x31, want: "CERTIFICATE_CHAINS_2"},
		{name: "ROOT_CERTIFICATES_1 80", sectionType: 0x80, want: "ROOT_CERTIFICATES_1"},
		{name: "ROOT_CERTIFICATES_2 81", sectionType: 0x81, want: "ROOT_CERTIFICATES_2"},
		{name: "FW_INTERNAL_USAGE 90", sectionType: 0x90, want: "FW_INTERNAL_USAGE"},
		{name: "DIGITAL_CERT_RW 91", sectionType: 0x91, want: "DIGITAL_CERT_RW"},
		{name: "DIGITAL_CACERT_RW 92", sectionType: 0x92, want: "DIGITAL_CACERT_RW"},
		{name: "DIGITAL_CERT_PTR 99", sectionType: 0x99, want: "DIGITAL_CERT_PTR"},

		// 0xF prefix sections
		{name: "ROOT_CERTIFICATES_1 F0", sectionType: 0xF0, want: "ROOT_CERTIFICATES_1"},
		{name: "ROOT_CERTIFICATES_2 F1", sectionType: 0xF1, want: "ROOT_CERTIFICATES_2"},
		{name: "CERT_CHAIN_0 F2", sectionType: 0xF2, want: "CERT_CHAIN_0"},
		{name: "DIGITAL_CACERT_RW F3", sectionType: 0xF3, want: "DIGITAL_CACERT_RW"},
		{name: "CERTIFICATE_CHAINS_1 F4", sectionType: 0xF4, want: "CERTIFICATE_CHAINS_1"},
		{name: "CERTIFICATE_CHAINS_2 F5", sectionType: 0xF5, want: "CERTIFICATE_CHAINS_2"},
		{name: "ROOT_CERTIFICATES_1 F6", sectionType: 0xF6, want: "ROOT_CERTIFICATES_1"},
		{name: "ROOT_CERTIFICATES_2 F7", sectionType: 0xF7, want: "ROOT_CERTIFICATES_2"},
		{name: "DIGITAL_CERT_PTR F9", sectionType: 0xF9, want: "DIGITAL_CERT_PTR"},
		{name: "FW_INTERNAL_USAGE FA", sectionType: 0xFA, want: "FW_INTERNAL_USAGE"},

		// Unknown DTOC section
		{name: "Unknown DTOC 0x55", sectionType: 0x55, want: "UNKNOWN_DTOC_0x55"},
		{name: "Unknown DTOC 0xAA", sectionType: 0xAA, want: "UNKNOWN_DTOC_0xAA"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDTOCSectionTypeName(tt.sectionType); got != tt.want {
				t.Errorf("GetDTOCSectionTypeName(0x%02X) = %v, want %v", tt.sectionType, got, tt.want)
			}
		})
	}
}

func TestSectionNameConsistency(t *testing.T) {
	// Test that section types that appear in both ITOC and DTOC have consistent names
	consistencyTests := []struct {
		name         string
		itocType     uint16
		dtocType     uint8
		expectedName string
	}{
		{name: "MFG_INFO", itocType: SectionTypeMfgInfo, dtocType: 0xE0, expectedName: "MFG_INFO"},
		{name: "DEV_INFO", itocType: SectionTypeDevInfo, dtocType: 0xE1, expectedName: "DEV_INFO"},
		{name: "SECURITY_LOG", itocType: SectionTypeSecurityLog, dtocType: 0xD5, expectedName: "SECURITY_LOG"},
		{name: "FW_INTERNAL_USAGE", itocType: SectionTypeFwInternalUsage, dtocType: 0xEA, expectedName: "FW_INTERNAL_USAGE"},
	}

	for _, tt := range consistencyTests {
		t.Run(tt.name, func(t *testing.T) {
			itocName := GetSectionTypeName(tt.itocType)
			dtocName := GetDTOCSectionTypeName(tt.dtocType)
			
			if itocName != tt.expectedName {
				t.Errorf("ITOC name for %s: got %v, want %v", tt.name, itocName, tt.expectedName)
			}
			if dtocName != tt.expectedName {
				t.Errorf("DTOC name for %s: got %v, want %v", tt.name, dtocName, tt.expectedName)
			}
		})
	}
}

func TestSpecialSectionTypes(t *testing.T) {
	// Test that special section types are handled correctly
	tests := []struct {
		name        string
		sectionType uint16
		want        string
	}{
		// BOOT2 has special type 0x100
		{name: "BOOT2 special type", sectionType: 0x100, want: "BOOT2"},
		// Test high values don't cause issues
		{name: "High unknown value", sectionType: 0xFFFF, want: "UNKNOWN_0xFFFF"},
		{name: "Mid-range unknown", sectionType: 0x200, want: "UNKNOWN_0x200"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSectionTypeName(tt.sectionType); got != tt.want {
				t.Errorf("GetSectionTypeName(0x%04X) = %v, want %v", tt.sectionType, got, tt.want)
			}
		})
	}
}

func BenchmarkGetSectionTypeName(b *testing.B) {
	// Benchmark the section name lookup
	sectionTypes := []uint16{
		SectionTypeBootCode,
		SectionTypeMainCode,
		SectionTypeImageInfo,
		SectionTypeBoot2,
		0x99, // Unknown
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, st := range sectionTypes {
			_ = GetSectionTypeName(st)
		}
	}
}

func BenchmarkGetDTOCSectionTypeName(b *testing.B) {
	// Benchmark the DTOC section name lookup
	sectionTypes := []uint8{
		0xE0, // MFG_INFO
		0xE1, // DEV_INFO
		0xF0, // ROOT_CERTIFICATES_1
		0x55, // Unknown
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, st := range sectionTypes {
			_ = GetDTOCSectionTypeName(st)
		}
	}
}