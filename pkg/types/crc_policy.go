package types

// InSectionCRCPolicy defines how to handle the 4-byte trailer for IN_SECTION CRC sections
// during reassembly.
type InSectionCRCPolicy uint8

const (
    // InSectionCRCPolicySoftware: calculate CRC16 in software on section payload (without trailer)
    InSectionCRCPolicySoftware InSectionCRCPolicy = iota
    // InSectionCRCPolicyHardware: calculate hardware CRC16 (rare for trailer; HW pointers handled elsewhere)
    InSectionCRCPolicyHardware
    // InSectionCRCPolicyBlank: write 0xFFFFFFFF sentinel as trailer (observed in several sections)
    InSectionCRCPolicyBlank
)

// GetInSectionCRCPolicy returns the policy for adding the trailing 4-byte word
// for sections that use CRCInSection. This centralizes behavior used by the reassembler.
// Notes:
// - TOOLS_AREA: policy is Software (CRC over first 60 bytes; the reassembler passes the payload without trailer).
// - BOOT2: trailer is usually 0xFFFFFFFF; the internal CRC is validated separately.
// - DEV_INFO/MFG_INFO/IMAGE_INFO/SIGNATURES/PUBLIC_KEYS/FORBIDDEN_VERSIONS/HASHES_TABLE: trailer is observed as 0xFFFFFFFF.
// - All other sections default to Software.
func GetInSectionCRCPolicy(sectionType uint16) InSectionCRCPolicy {
    switch sectionType {
    case SectionTypeToolsArea:
        return InSectionCRCPolicySoftware

    case SectionTypeBoot2,
        SectionTypeDevInfo, SectionTypeDevInfo1, SectionTypeDevInfo2,
        SectionTypeMfgInfo,
        SectionTypeImageInfo,
        SectionTypeForbiddenVersions,
        SectionTypePublicKeys2048, SectionTypePublicKeys4096,
        SectionTypeImageSignature512,
        SectionTypeHashesTable:
        return InSectionCRCPolicyBlank

    default:
        // For completeness: if any section explicitly uses hardware CRC as its algorithm
        // and also places CRC in-section (uncommon), return Hardware. Otherwise Software.
        if GetSectionCRCAlgorithm(sectionType) == CRCAlgorithmHardware {
            return InSectionCRCPolicyHardware
        }
        return InSectionCRCPolicySoftware
    }
}

