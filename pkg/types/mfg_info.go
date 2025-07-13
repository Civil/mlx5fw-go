package types

// MfgInfo represents the MFG_INFO section structure
// Based on mstflint's image_layout_mfg_info structure and hex dump analysis
type MfgInfo struct {
	// PSID - offset 0x0
	PSID           [16]byte  `bin:"len:16"` // offset 0x0-0xf - "MT_0000000911"
	
	// Unknown fields - offset 0x10
	Reserved1      [12]byte  `bin:"len:12"` // offset 0x10-0x1b
	Flags          uint32    `bin:"len:4"`  // offset 0x1c - observed 0x01000001
	
	// GUID allocation info - offset 0x20
	Guids          UidEntry  `bin:"len:16"` // offset 0x20-0x2f - GUID allocation info
	
	// MAC allocation info - offset 0x30  
	Macs           UidEntry  `bin:"len:16"` // offset 0x30-0x3f - MAC allocation info
	
	// Remaining data
	Reserved2      [448]byte `bin:"len:448"` // offset 0x40-0x1ff - padding to 0x200
}

// GetPSIDString returns the PSID as a string
func (m *MfgInfo) GetPSIDString() string {
	return nullTerminatedString(m.PSID[:])
}