package types

const (
	// Magic patterns
	MagicPattern = 0x4D544657ABCDEF00 // "MTFW\xAB\xCD\xEF\x00" in big endian
	ITOCSignature = 0x49544F43         // "ITOC"
	DTOCSignature = 0x44544F43         // "DTOC"
	
	
	// CRC polynomial and initial values
	CRCPolynomial     = 0x100b
	CRCInitial        = 0xffff
	CRCXorOut         = 0xffff
	
	// Section types (from mstflint fs3_ops.h)
	SectionTypeBootCode          = 0x1
	SectionTypePCICode           = 0x2
	SectionTypeMainCode          = 0x3
	SectionTypePCIELinkCode      = 0x4
	SectionTypeIronPrepCode      = 0x5
	SectionTypePostIronBootCode  = 0x6
	SectionTypeUpgradeCode       = 0x7
	SectionTypeHWBootCfg         = 0x8
	SectionTypeHWMainCfg         = 0x9
	SectionTypePhyUCCode         = 0xa
	SectionTypePhyUCConsts       = 0xb
	SectionTypePCIEPhyUCCode     = 0xc
	SectionTypeCCIRInfraCode     = 0xd
	SectionTypeCCIRAlgoCode      = 0xe
	SectionTypeBoot3Code         = 0xf
	SectionTypeImageInfo         = 0x10
	SectionTypeFWBootCfg         = 0x11
	SectionTypeFWMainCfg         = 0x12
	SectionTypeAPUKernel         = 0x14
	SectionTypeACECode           = 0x15
	SectionTypeROMCode           = 0x18
	SectionTypeResetInfo         = 0x20
	SectionTypeDbgFWINI          = 0x30
	SectionTypeDbgFWParams       = 0x32
	SectionTypeFWAdb             = 0x33
	SectionTypeImageSignature256 = 0xa0
	SectionTypePublicKeys2048    = 0xa1
	SectionTypeForbiddenVersions = 0xa2
	SectionTypeImageSignature512 = 0xa3
	SectionTypePublicKeys4096    = 0xa4
	SectionTypeHMACDigest        = 0xa5
	SectionTypeRsaPublicKey      = 0xa6
	SectionTypeRsa4096Signatures     = 0xa7
	SectionTypeEncryptionKeyTransition = 0xa9
	SectionTypePxirIni               = 0xaa
	SectionTypePxirIni1              = 0xab
	SectionTypeNvdaRotCertificates   = 0xad
	SectionTypeExclkSyncInfo         = 0xb0
	SectionTypeMainPagesHashes       = 0xb1
	SectionTypeMainPagesLockedHashes = 0xb2
	SectionTypeStrnMain              = 0xb4
	SectionTypeStrnIron              = 0xb5
	SectionTypeStrnTile              = 0xb6
	SectionTypeMainData              = 0xd3
	SectionTypeFwDebugDump2          = 0xd4
	SectionTypeSecurityLog           = 0xd5
	SectionTypeMfgInfo               = 0xe0
	SectionTypeDevInfo               = 0xe1
	SectionTypeNvData1               = 0xe2
	SectionTypeVpdR0                 = 0xe3
	SectionTypeNvData2               = 0xe4
	SectionTypeFwNvLog               = 0xe5
	SectionTypeNvData0               = 0xe6
	SectionTypeDevInfo1              = 0xe7
	SectionTypeDevInfo2              = 0xe8
	SectionTypeCRDumpMaskData        = 0xe9
	SectionTypeFwInternalUsage       = 0xea
	SectionTypeProgrammableHwFw1     = 0xeb
	SectionTypeProgrammableHwFw2     = 0xec
	SectionTypeDigitalCertPtr        = 0xed
	SectionTypeDigitalCertRw         = 0xee
	SectionTypeLcIni1Table           = 0xef
	SectionTypeLcIni2Table           = 0xf0
	SectionTypeLcIniNvData           = 0xf1
	SectionTypeCertChain0            = 0xf2
	SectionTypeDigitalCaCertRw       = 0xf3
	SectionTypeCertificateChains1    = 0xf4
	SectionTypeCertificateChains2    = 0xf5
	SectionTypeRootCertificates1     = 0xf6
	SectionTypeRootCertificates2     = 0xf7
	SectionTypeToolsArea             = 0xf9
	SectionTypeHashesTable           = 0xfa
	SectionTypeHwPtr                 = 0xfb
	SectionTypeFwDebugDump           = 0xfc
	SectionTypeItoc                  = 0xfd
	SectionTypeDtoc                  = 0xfe
	SectionTypeEnd                   = 0xff
	
	// Special section types for sections not in ITOC/DTOC
	SectionTypeBoot2                 = 0x100 // Special marker for BOOT2 section
	
	// NOCRC flag value
	NOCRC = 1
)

// Search offsets for magic pattern
var MagicSearchOffsets = []uint32{
	0x0, 0x10000, 0x20000, 0x40000, 0x80000,
	0x100000, 0x200000, 0x400000, 0x800000,
	0x1000000, 0x2000000,
}