package types

import (
	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// MfgInfoAnnotated represents the MFG_INFO section structure using annotations
// Based on mstflint's image_layout_mfg_info structure and hex dump analysis
type MfgInfoAnnotated struct {
	// PSID - offset 0x0
	PSID           [16]byte  `offset:"byte:0"`  // "MT_0000000911"
	
	// Unknown fields - offset 0x10
	Reserved1      [12]byte  `offset:"byte:16,reserved:true"`  // offset 0x10-0x1b
	Flags          uint32    `offset:"byte:28,endian:be"`      // offset 0x1c - observed 0x01000001
	
	// GUID allocation info - offset 0x20
	Guids          UidEntryAnnotated  `offset:"byte:32"`  // offset 0x20-0x2f - GUID allocation info
	
	// MAC allocation info - offset 0x30  
	Macs           UidEntryAnnotated  `offset:"byte:48"`  // offset 0x30-0x3f - MAC allocation info
	
	// Remaining data
	Reserved2      [448]byte `offset:"byte:64,reserved:true"`  // offset 0x40-0x1ff - padding to 0x200
}

// Unmarshal methods using annotations

// Unmarshal unmarshals binary data into MfgInfoAnnotated
func (m *MfgInfoAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, m)
}

// Marshal marshals MfgInfoAnnotated into binary data
func (m *MfgInfoAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(m)
}

