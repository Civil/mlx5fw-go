package types

import (
	"reflect"
	
	"github.com/Civil/mlx5fw-go/pkg/annotations"
)

// DeviceInfoAnnotated represents the DEV_INFO section structure using annotations
// This structure contains device-specific information
type DeviceInfoAnnotated struct {
	// Device identification
	DeviceID       uint16   `offset:"byte:0,endian:be"`   // Device ID
	VendorID       uint16   `offset:"byte:2,endian:be"`   // Vendor ID (usually 0x15b3 for Mellanox)
	SubsystemID    uint16   `offset:"byte:4,endian:be"`   // Subsystem ID
	SubsystemVendorID uint16 `offset:"byte:6,endian:be"`   // Subsystem vendor ID
	
	// Version information
	HWVersion      uint32   `offset:"byte:8,endian:be"`   // Hardware version
	HWRevision     uint32   `offset:"byte:12,endian:be"`  // Hardware revision
	
	// Device capabilities
	Capabilities   uint64   `offset:"byte:16,endian:be"`  // Device capabilities bitmap
	
	// MAC addresses and GUIDs
	MACGUID        [8]byte  `offset:"byte:24"`            // Base MAC/GUID
	NumMACs        uint32   `offset:"byte:32,endian:be"`  // Number of MAC addresses
	
	// Additional info
	Reserved       [64]byte `offset:"byte:36,reserved:true"` // Reserved for future use
}

// MFGInfoAnnotated represents the MFG_INFO section structure using annotations
// This structure contains manufacturing information
type MFGInfoAnnotated struct {
	PSID           [16]byte  `offset:"byte:0"`   // Product Serial ID
	PartNumber     [32]byte  `offset:"byte:16"`  // Manufacturer part number
	Revision       [16]byte  `offset:"byte:48"`  // Revision
	ProductName    [64]byte  `offset:"byte:64"`  // Product name
	Reserved       [128]byte `offset:"byte:128,reserved:true"` // Reserved for future use
}

// VPDDataAnnotated represents the VPD_R0 section structure using annotations
// VPD (Vital Product Data) contains additional product information
type VPDDataAnnotated struct {
	// VPD header
	Signature      [3]byte  `offset:"byte:0"`  // Should be "VPD"
	Length         uint8    `offset:"byte:3"`  // Length of VPD data
	
	// VPD fields are variable length with tag-length-value format
	// For sections with variable length data, we typically handle the fixed header
	// and then process the remaining data separately
}

// Unmarshal methods using annotations

// Unmarshal unmarshals binary data into DeviceInfoAnnotated
func (d *DeviceInfoAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, d)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (d *DeviceInfoAnnotated) UnmarshalWithReserved(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*d))
	if err != nil {
		return err
	}
	opts := &annotations.UnmarshalOptions{
		IncludeReserved: true,
	}
	return annotations.UnmarshalWithOptions(data, d, annot, opts)
}

// Marshal marshals DeviceInfoAnnotated into binary data
func (d *DeviceInfoAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(d)
}

// MarshalWithReserved marshals DeviceInfoAnnotated including reserved fields
func (d *DeviceInfoAnnotated) MarshalWithReserved() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*d))
	if err != nil {
		return nil, err
	}
	opts := &annotations.MarshalOptions{
		IncludeReserved: true,
	}
	return annotations.MarshalWithOptions(d, annot, opts)
}

// Unmarshal unmarshals binary data into MFGInfoAnnotated
func (m *MFGInfoAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, m)
}

// UnmarshalWithReserved unmarshals binary data including reserved fields
func (m *MFGInfoAnnotated) UnmarshalWithReserved(data []byte) error {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*m))
	if err != nil {
		return err
	}
	opts := &annotations.UnmarshalOptions{
		IncludeReserved: true,
	}
	return annotations.UnmarshalWithOptions(data, m, annot, opts)
}

// Marshal marshals MFGInfoAnnotated into binary data
func (m *MFGInfoAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(m)
}

// MarshalWithReserved marshals MFGInfoAnnotated including reserved fields
func (m *MFGInfoAnnotated) MarshalWithReserved() ([]byte, error) {
	annot, err := annotations.ParseStruct(reflect.TypeOf(*m))
	if err != nil {
		return nil, err
	}
	opts := &annotations.MarshalOptions{
		IncludeReserved: true,
	}
	return annotations.MarshalWithOptions(m, annot, opts)
}

// Conversion methods to maintain compatibility

// ToAnnotated converts DeviceInfo to DeviceInfoAnnotated
func (d *DeviceInfo) ToAnnotated() *DeviceInfoAnnotated {
	return &DeviceInfoAnnotated{
		DeviceID:          d.DeviceID,
		VendorID:          d.VendorID,
		SubsystemID:       d.SubsystemID,
		SubsystemVendorID: d.SubsystemVendorID,
		HWVersion:         d.HWVersion,
		HWRevision:        d.HWRevision,
		Capabilities:      d.Capabilities,
		MACGUID:           d.MACGUID,
		NumMACs:           d.NumMACs,
		Reserved:          d.Reserved,
	}
}

// FromAnnotated converts DeviceInfoAnnotated to DeviceInfo
func (d *DeviceInfoAnnotated) FromAnnotated() *DeviceInfo {
	return &DeviceInfo{
		DeviceID:          d.DeviceID,
		VendorID:          d.VendorID,
		SubsystemID:       d.SubsystemID,
		SubsystemVendorID: d.SubsystemVendorID,
		HWVersion:         d.HWVersion,
		HWRevision:        d.HWRevision,
		Capabilities:      d.Capabilities,
		MACGUID:           d.MACGUID,
		NumMACs:           d.NumMACs,
		Reserved:          d.Reserved,
	}
}

// ToAnnotated converts MFGInfo to MFGInfoAnnotated
func (m *MFGInfo) ToAnnotated() *MFGInfoAnnotated {
	return &MFGInfoAnnotated{
		PSID:        m.PSID,
		PartNumber:  m.PartNumber,
		Revision:    m.Revision,
		ProductName: m.ProductName,
		Reserved:    m.Reserved,
	}
}

// FromAnnotated converts MFGInfoAnnotated to MFGInfo
func (m *MFGInfoAnnotated) FromAnnotated() *MFGInfo {
	return &MFGInfo{
		PSID:        m.PSID,
		PartNumber:  m.PartNumber,
		Revision:    m.Revision,
		ProductName: m.ProductName,
		Reserved:    m.Reserved,
	}
}

// Unmarshal unmarshals binary data into VPDDataAnnotated
func (v *VPDDataAnnotated) Unmarshal(data []byte) error {
	return annotations.UnmarshalStruct(data, v)
}

// Marshal marshals VPDDataAnnotated into binary data
func (v *VPDDataAnnotated) Marshal() ([]byte, error) {
	return annotations.MarshalStruct(v)
}

// ToAnnotated converts VPDData to VPDDataAnnotated
func (v *VPDData) ToAnnotated() *VPDDataAnnotated {
	return &VPDDataAnnotated{
		Signature: v.Signature,
		Length:    v.Length,
	}
}

// FromAnnotated converts VPDDataAnnotated to VPDData
func (v *VPDDataAnnotated) FromAnnotated() *VPDData {
	return &VPDData{
		Signature: v.Signature,
		Length:    v.Length,
		// Data field will be handled separately after unmarshaling the header
	}
}

// Helper methods

// GetPSIDString returns the PSID as a string
func (m *MFGInfoAnnotated) GetPSIDString() string {
	return nullTerminatedString(m.PSID[:])
}

// GetPartNumberString returns the part number as a string  
func (m *MFGInfoAnnotated) GetPartNumberString() string {
	return nullTerminatedString(m.PartNumber[:])
}

// GetRevisionString returns the revision as a string
func (m *MFGInfoAnnotated) GetRevisionString() string {
	return nullTerminatedString(m.Revision[:])
}

// GetProductNameString returns the product name as a string
func (m *MFGInfoAnnotated) GetProductNameString() string {
	return nullTerminatedString(m.ProductName[:])
}