//go:build linux
// +build linux

package pcie

// Access Register protocol scaffolding: Operation TLV + Reg TLV packing helpers
// This mirrors mstflint's layout for Access Register encapsulation.

const (
	// Register IDs (HCA)
	RegID_MGIR = 0x9020
	RegID_MCQI = 0x9061
	RegID_MCQS = 0x9060

	// Methods
	RegMethodGet = 1
	RegMethodSet = 2
)

// Operation TLV (packed in network byte order in mstflint, but values are small)
type opTLV struct {
	Class  uint8 // Access class (SMP/GMP/etc.), not used for PCI path
	Method uint8 // 1=GET, 2=SET
	_pad   uint16
	RegID  uint16
	Status uint16
}

// regTLV header
type regTLV struct {
	Type uint16 // TLV type for Reg
	Len  uint16 // dwords length including header
}

// packOpTLV packs Operation TLV into the buffer; returns bytes written.
func packOpTLV(dst []byte, method uint8, regID uint16) int {
	// Minimal packing: Type/class fields are not surfaced to device in PCI path
	// We only care about method + regID; status is written by device.
	if len(dst) < 8 {
		return 0
	}
	dst[0] = 0 // class placeholder
	dst[1] = method
	dst[2] = 0
	dst[3] = 0
	dst[4] = byte(regID >> 8)
	dst[5] = byte(regID)
	dst[6] = 0 // status hi (out)
	dst[7] = 0 // status lo (out)
	return 8
}

// packRegTLV packs the Reg TLV header and returns bytes written (not including body).
func packRegTLV(dst []byte, dwords int) int {
	if len(dst) < 4 {
		return 0
	}
	// Simple header: Type is fixed for register container in our usage (opaque)
	// We only need to set length in dwords including this header
	dst[0] = 0
	dst[1] = 0 // type placeholder
	l := uint16(dwords)
	dst[2] = byte(l >> 8)
	dst[3] = byte(l)
	return 4
}

// ARClient.transact is expected to send the packed TLV over ICMD/PCI;
// the transport is device-specific and implemented elsewhere.

// pushBitsBE sets bitLen bits of val into dst starting at bitOff (MSB-first within each byte).
func pushBitsBE(dst []byte, bitOff uint32, bitLen uint32, val uint32) {
	for i := uint32(0); i < bitLen; i++ {
		bit := (val >> (bitLen - 1 - i)) & 1
		pos := bitOff + i
		byteIdx := pos / 8
		bitInByte := pos % 8 // 0 is MSB in a byte
		if int(byteIdx) >= len(dst) {
			break
		}
		mask := byte(1 << (7 - bitInByte))
		if bit != 0 {
			dst[byteIdx] |= mask
		}
	}
}

// packMCQIHeader packs the MCQI reg header fields into payload per reg_access_hca_mcqi_reg_ext_pack
// Only sets the header; data union left zero for GET.
func packMCQIHeader(payload []byte, infoType byte, dataSize uint16) {
	for i := range payload {
		payload[i] = 0
	}
	// component_index (16 bits) at bit offset 16
	pushBitsBE(payload, 16, 16, 0)
	// device_index (12 bits) at bit offset 4
	pushBitsBE(payload, 4, 12, 0)
	// read_pending_component (1 bit) at bit offset 0
	pushBitsBE(payload, 0, 1, 0)
	// device_type (8 bits) at bit offset 56 (0 = NIC/Switch)
	pushBitsBE(payload, 56, 8, 0)
	// info_type (5 bits) at bit offset 91
	pushBitsBE(payload, 91, 5, uint32(infoType))
	// info_size (32 bits) at bit offset 96: version ext size is 0x7c
	pushBitsBE(payload, 96, 32, 0x7c)
	// offset (32 bits) at bit offset 128: 0
	pushBitsBE(payload, 128, 32, 0)
	// data_size (16 bits) at bit offset 176
	pushBitsBE(payload, 176, 16, uint32(dataSize))
}

// getBitsBE reads bitLen bits starting at bitOff (MSB-first within byte) and returns value
func getBitsBE(src []byte, bitOff uint32, bitLen uint32) uint32 {
	var v uint32
	for i := uint32(0); i < bitLen; i++ {
		pos := bitOff + i
		byteIdx := pos / 8
		bitInByte := pos % 8
		if int(byteIdx) >= len(src) {
			break
		}
		b := (src[byteIdx] >> (7 - bitInByte)) & 1
		v = (v << 1) | uint32(b)
	}
	return v
}

// packMCQSHeader packs the MCQS request header: component_index, device_index, device_type.
func packMCQSHeader(payload []byte, componentIndex uint16, deviceIndex uint16, deviceType uint8) {
	for i := range payload {
		payload[i] = 0
	}
	pushBitsBE(payload, 16, 16, uint32(componentIndex))
	pushBitsBE(payload, 4, 12, uint32(deviceIndex))
	pushBitsBE(payload, 120, 8, uint32(deviceType))
}
