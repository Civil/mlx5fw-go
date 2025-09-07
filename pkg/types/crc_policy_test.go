package types

import "testing"

func TestGetInSectionCRCPolicy_Mappings(t *testing.T) {
    // Software CRC cases
    if p := GetInSectionCRCPolicy(SectionTypeToolsArea); p != InSectionCRCPolicySoftware {
        t.Fatalf("TOOLS_AREA policy: got %v want Software", p)
    }

    // Blank sentinel cases
    blanks := []uint16{
        SectionTypeBoot2,
        SectionTypeDevInfo, SectionTypeDevInfo1, SectionTypeDevInfo2,
        SectionTypeMfgInfo,
        SectionTypeImageInfo,
        SectionTypeForbiddenVersions,
        SectionTypePublicKeys2048, SectionTypePublicKeys4096,
        SectionTypeImageSignature512,
        SectionTypeHashesTable,
    }
    for _, st := range blanks {
        if p := GetInSectionCRCPolicy(st); p != InSectionCRCPolicyBlank {
            t.Fatalf("type 0x%x policy: got %v want Blank", st, p)
        }
    }

    // Hardware policy only for HW_PTR if ever IN_SECTION (it is not in practice)
    if p := GetInSectionCRCPolicy(SectionTypeHwPtr); p != InSectionCRCPolicyHardware {
        t.Fatalf("HW_PTR policy: got %v want Hardware", p)
    }
}

