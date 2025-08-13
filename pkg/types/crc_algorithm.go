package types

// CRCAlgorithm represents the algorithm used for CRC calculation
type CRCAlgorithm uint8

const (
	// CRCAlgorithmSoftware uses software CRC16 (polynomial 0x100b)
	CRCAlgorithmSoftware CRCAlgorithm = iota
	// CRCAlgorithmHardware uses hardware CRC16 (calc_hw_crc with crc16table2)
	CRCAlgorithmHardware
)

// GetSectionCRCAlgorithm returns the CRC algorithm used for a specific section type
// This is the single source of truth for CRC algorithm selection
func GetSectionCRCAlgorithm(sectionType uint16) CRCAlgorithm {
	// Based on mstflint source code analysis:
	// - HW pointers use calc_hw_crc (hardware CRC)
	// - All other sections use software CRC16 (Crc16 class)
	switch sectionType {
	case SectionTypeHwPtr:
		return CRCAlgorithmHardware
	default:
		// All other sections including BOOT2, TOOLS_AREA, IMAGE_INFO, etc. use software CRC
		return CRCAlgorithmSoftware
	}
}

// String returns the string representation of the CRC algorithm
func (c CRCAlgorithm) String() string {
	switch c {
	case CRCAlgorithmSoftware:
		return "SOFTWARE"
	case CRCAlgorithmHardware:
		return "HARDWARE"
	default:
		return "UNKNOWN"
	}
}
