package types

import (
	"encoding/json"

	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// MfgInfo represents the MFG_INFO section structure using annotations
// Based on mstflint's image_layout_mfg_info structure and hex dump analysis
type MfgInfo struct {
	// PSID - offset 0x0
	PSID [16]byte `offset:"byte:0" json:"-"` // "MT_0000000911" (will be handled separately in MarshalJSON)

	// Unknown fields - offset 0x10
	Reserved1 [12]byte `offset:"byte:16,reserved:true" json:"reserved1"` // offset 0x10-0x1b
	Flags     uint32   `offset:"byte:28,endian:be" json:"flags"`         // offset 0x1c - observed 0x01000001

	// GUID allocation info - offset 0x20
	Guids UidEntry `offset:"byte:32" json:"guids"` // offset 0x20-0x2f - GUID allocation info

	// MAC allocation info - offset 0x30
	Macs UidEntry `offset:"byte:48" json:"macs"` // offset 0x30-0x3f - MAC allocation info

	// Trailing padding (variable): different families use different total sizes.
	// Observed: 0x100 (256) on some CX6Dx firmwares and 0x140 (320) on CX5.
	// To match mstflint (image_layout_mfg_info) behavior and avoid magic sizes,
	// model the tail as a slice starting at 0x40 so we accept either layout.
	// CRC/size for this section come from the DTOC/ITOC entry, not from an
	// embedded trailer.
	Reserved2 []byte `offset:"byte:64,reserved:true" json:"reserved2"`
}

// Unmarshal methods using annotations

// Unmarshal unmarshals binary data into MfgInfo
func (m *MfgInfo) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, m)
}

// Marshal marshals MfgInfo into binary data
func (m *MfgInfo) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(m)
}

// MarshalJSON implements json.Marshaler interface
func (m *MfgInfo) MarshalJSON() ([]byte, error) {
	// Create an anonymous struct that includes all fields plus computed ones
	type Alias MfgInfo
	return json.Marshal(&struct {
		*Alias
		PSID string `json:"psid"`
	}{
		Alias: (*Alias)(m),
		PSID:  nullTerminatedString(m.PSID[:]),
	})
}

// UnmarshalJSON implements json.Unmarshaler interface
func (m *MfgInfo) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary struct with all fields
	var temp struct {
		PSID      string   `json:"psid"`
		Reserved1 []byte   `json:"reserved1"`
		Flags     uint32   `json:"flags"`
		Guids     UidEntry `json:"guids"`
		Macs      UidEntry `json:"macs"`
		Reserved2 []byte   `json:"reserved2"`
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

// GetPSIDString returns PSID as a cleaned string
func (m *MfgInfo) GetPSIDString() string {
	return nullTerminatedString(m.PSID[:])
}
