package types

// DevInfo represents the DEV_INFO section structure
// Based on mstflint's image_layout_device_info structure
type DevInfo struct {
	Signature0     uint32    `bin:"len:4"`  // offset 0x0 - "mDevInfo"
	Signature1     uint32    `bin:"len:4"`  // offset 0x4 - "#B.."
	Signature2     uint32    `bin:"len:4"`  // offset 0x8 - "baca"
	Signature3     uint32    `bin:"len:4"`  // offset 0xc - "fe00"
	MinorVersion   uint8     `bin:"len:1"`  // offset 0x10
	MajorVersion   uint16    `bin:"len:2"`  // offset 0x11
	Reserved1      uint8     `bin:"len:1"`  // offset 0x13
	Reserved2      [12]byte  `bin:"len:12"` // offset 0x14-0x1f
	Guids          UidEntry  `bin:"len:16"` // offset 0x20 - GUID allocation info  
	Macs           UidEntry  `bin:"len:16"` // offset 0x30 - MAC allocation info
	Reserved3      [416]byte `bin:"len:416"` // offset 0x40 - padding to 0x1fc
	CRC            uint16    `bin:"len:2"`  // offset 0x1fc
	Reserved4      uint16    `bin:"len:2"`  // offset 0x1fe
}

// UidEntry represents UID allocation information
// Based on mstflint's image_layout_uid_entry structure
// Note: The first 4 bytes contain the allocation info in a specific format
type UidEntry struct {
	Reserved1       uint16 `bin:"len:2"` // offset 0x0-0x1 - reserved (0x0000)
	Step            uint8  `bin:"len:1"` // offset 0x2 - Step size (not used for CX4+)
	NumAllocated    uint8  `bin:"len:1"` // offset 0x3 - Number of allocated UIDs
	Reserved2       uint32 `bin:"len:4"` // offset 0x4-0x7 - reserved
	UID             uint64 `bin:"len:8"` // offset 0x8-0xf - Base UID value
}

// GetNumAllocated returns the total number of allocated UIDs
func (u *UidEntry) GetNumAllocated() int {
	return int(u.NumAllocated)
}

// GetUID returns the UID with correct byte order
// The UID is stored as big-endian in the firmware but when we unmarshal
// with UnmarshalLE, we need to swap the bytes
func (u *UidEntry) GetUID() uint64 {
	// Swap bytes from little-endian interpretation to big-endian
	return ((u.UID & 0xFF) << 56) |
		((u.UID & 0xFF00) << 40) |
		((u.UID & 0xFF0000) << 24) |
		((u.UID & 0xFF000000) << 8) |
		((u.UID & 0xFF00000000) >> 8) |
		((u.UID & 0xFF0000000000) >> 24) |
		((u.UID & 0xFF000000000000) >> 40) |
		((u.UID & 0xFF00000000000000) >> 56)
}