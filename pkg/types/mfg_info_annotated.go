package types

import (
	"encoding/json"
	
	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// MfgInfoAnnotated represents the MFG_INFO section structure using annotations
// Based on mstflint's image_layout_mfg_info structure and hex dump analysis
type MfgInfoAnnotated struct {
	// PSID - offset 0x0
	PSID           [16]byte  `offset:"byte:0" json:"-"`  // "MT_0000000911" (will be handled separately in MarshalJSON)
	
	// Unknown fields - offset 0x10
	Reserved1      [12]byte  `offset:"byte:16,reserved:true" json:"reserved1"`  // offset 0x10-0x1b
	Flags          uint32    `offset:"byte:28,endian:be" json:"flags"`      // offset 0x1c - observed 0x01000001
	
	// GUID allocation info - offset 0x20
	Guids          UidEntryAnnotated  `offset:"byte:32" json:"guids"`  // offset 0x20-0x2f - GUID allocation info
	
	// MAC allocation info - offset 0x30  
	Macs           UidEntryAnnotated  `offset:"byte:48" json:"macs"`  // offset 0x30-0x3f - MAC allocation info
	
	// Remaining data
	Reserved2      [448]byte `offset:"byte:64,reserved:true" json:"reserved2"`  // offset 0x40-0x1ff - padding to 0x200
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

// MarshalJSON implements json.Marshaler interface
func (m *MfgInfoAnnotated) MarshalJSON() ([]byte, error) {
	// Create an anonymous struct that includes all fields plus computed ones
	type Alias MfgInfoAnnotated
	return json.Marshal(&struct {
		*Alias
		PSID string `json:"psid"`
	}{
		Alias: (*Alias)(m),
		PSID:  nullTerminatedString(m.PSID[:]),
	})
}

// UnmarshalJSON implements json.Unmarshaler interface
func (m *MfgInfoAnnotated) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary struct with all fields
	var temp struct {
		PSID      string            `json:"psid"`
		Reserved1 []byte            `json:"reserved1"`
		Flags     uint32            `json:"flags"`
		Guids     UidEntryAnnotated `json:"guids"`
		Macs      UidEntryAnnotated `json:"macs"`
		Reserved2 []byte            `json:"reserved2"`
	}
	
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	
	// Copy fields
	psidBytes := []byte(temp.PSID)
	if len(psidBytes) > 16 {
		psidBytes = psidBytes[:16]
	}
	copy(m.PSID[:], psidBytes)
	
	copy(m.Reserved1[:], temp.Reserved1)
	m.Flags = temp.Flags
	m.Guids = temp.Guids
	m.Macs = temp.Macs
	copy(m.Reserved2[:], temp.Reserved2)
	
	return nil
}

