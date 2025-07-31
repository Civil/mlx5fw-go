package types

import (
	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// UidEntryAnnotated represents UID allocation information using annotations
// Based on mstflint's image_layout_uid_entry structure
type UidEntryAnnotated struct {
	Reserved1       uint16 `offset:"byte:0,endian:be,reserved:true"`  // reserved (0x0000)
	Step            uint8  `offset:"byte:2"`  // Step size (not used for CX4+)
	NumAllocated    uint8  `offset:"byte:3"`  // Number of allocated UIDs
	Reserved2       uint32 `offset:"byte:4,endian:be,reserved:true"`  // reserved
	UID             uint64 `offset:"byte:8,endian:be"`  // Base UID value
}

// DevInfoAnnotated represents the DEV_INFO section structure using annotations
// Based on mstflint's image_layout_device_info structure
type DevInfoAnnotated struct {
	Signature0     uint32    `offset:"byte:0,endian:be"`   // "mDevInfo"
	Signature1     uint32    `offset:"byte:4,endian:be"`   // "#B.."
	Signature2     uint32    `offset:"byte:8,endian:be"`   // "baca"
	Signature3     uint32    `offset:"byte:12,endian:be"`  // "fe00"
	MinorVersion   uint8     `offset:"byte:16"`            // offset 0x10
	MajorVersion   uint16    `offset:"byte:17,endian:be"`  // offset 0x11
	Reserved1      uint8     `offset:"byte:19,reserved:true"`  // offset 0x13
	Reserved2      [12]byte  `offset:"byte:20,reserved:true"`  // offset 0x14-0x1f
	Guids          UidEntryAnnotated  `offset:"byte:32"`   // offset 0x20 - GUID allocation info  
	Macs           UidEntryAnnotated  `offset:"byte:48"`   // offset 0x30 - MAC allocation info
	Reserved3      [444]byte `offset:"byte:64,reserved:true"`  // offset 0x40 - padding to 0x1fc
	Reserved4      uint16    `offset:"byte:508,endian:be,reserved:true"`  // offset 0x1fc
	CRC            uint16    `offset:"byte:510,endian:be"`  // offset 0x1fe
	// Note: The firmware has a 4-byte CRC trailer after this structure, not part of the struct itself
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

