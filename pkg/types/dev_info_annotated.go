package types

import (
	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// UidEntryAnnotated represents UID allocation information using annotations
// Based on mstflint's image_layout_uid_entry structure
type UidEntryAnnotated struct {
	Reserved1       uint16 `offset:"byte:0,endian:be,reserved:true" json:"reserved1"`  // reserved (0x0000)
	Step            uint8  `offset:"byte:2" json:"step"`  // Step size (not used for CX4+)
	NumAllocated    uint8  `offset:"byte:3" json:"num_allocated"`  // Number of allocated UIDs
	Reserved2       uint32 `offset:"byte:4,endian:be,reserved:true" json:"reserved2"`  // reserved
	UID             uint64 `offset:"byte:8,endian:be" json:"uid"`  // Base UID value
}

// DevInfoAnnotated represents the DEV_INFO section structure using annotations
// Based on mstflint's image_layout_device_info structure
type DevInfoAnnotated struct {
	Signature0     uint32    `offset:"byte:0,endian:be" json:"signature0"`   // "mDevInfo"
	Signature1     uint32    `offset:"byte:4,endian:be" json:"signature1"`   // "#B.."
	Signature2     uint32    `offset:"byte:8,endian:be" json:"signature2"`   // "baca"
	Signature3     uint32    `offset:"byte:12,endian:be" json:"signature3"`  // "fe00"
	MinorVersion   uint8     `offset:"byte:16" json:"minor_version"`            // offset 0x10
	MajorVersion   uint16    `offset:"byte:17,endian:be" json:"major_version"`  // offset 0x11
	Reserved1      uint8     `offset:"byte:19,reserved:true" json:"reserved1"`  // offset 0x13
	Reserved2      [12]byte  `offset:"byte:20,reserved:true" json:"reserved2"`  // offset 0x14-0x1f
	Guids          UidEntryAnnotated  `offset:"byte:32" json:"guids"`   // offset 0x20 - GUID allocation info  
	Macs           UidEntryAnnotated  `offset:"byte:48" json:"macs"`   // offset 0x30 - MAC allocation info
	Reserved3      [444]byte `offset:"byte:64,reserved:true" json:"reserved3"`  // offset 0x40 - padding to 0x1fc
	CRC            uint32    `offset:"byte:508,endian:be" json:"crc"`  // offset 0x1fc - CRC32 field (lower 16 bits contain CRC16)
	
	// The firmware has a 4-byte CRC trailer after this structure at offset 0x200
	// This field is not part of the binary structure but is used for JSON marshaling
	TrailerCRC     uint16    `offset:"-" json:"original_crc,omitempty"`  // CRC after the struct (from ITOC)
}

// Unmarshal methods using annotations

// Unmarshal unmarshals binary data into UidEntryAnnotated
func (u *UidEntryAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, u)
}

// Marshal marshals UidEntryAnnotated into binary data
func (u *UidEntryAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(u)
}

// Unmarshal unmarshals binary data into DevInfoAnnotated
func (d *DevInfoAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, d)
}

// Marshal marshals DevInfoAnnotated into binary data
func (d *DevInfoAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(d)
}

