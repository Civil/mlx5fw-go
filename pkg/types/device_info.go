package types

import (
	"github.com/Civil/mlx5fw-go/pkg/errs"
)

// DeviceInfo represents the DEV_INFO section structure
// This structure contains device-specific information
type DeviceInfo struct {
	// Device identification
	DeviceID       uint16   // Device ID
	VendorID       uint16   // Vendor ID (usually 0x15b3 for Mellanox)
	SubsystemID    uint16   // Subsystem ID
	SubsystemVendorID uint16 // Subsystem vendor ID
	
	// Version information
	HWVersion      uint32   // Hardware version
	HWRevision     uint32   // Hardware revision
	
	// Device capabilities
	Capabilities   uint64   // Device capabilities bitmap
	
	// MAC addresses and GUIDs
	MACGUID        [8]byte  // Base MAC/GUID
	NumMACs        uint32   // Number of MAC addresses
	
	// Additional info
	Reserved       [64]byte // Reserved for future use
}

// MFGInfo represents the MFG_INFO section structure
// This structure contains manufacturing information
type MFGInfo struct {
	PSID           [16]byte  // Product Serial ID
	PartNumber     [32]byte  // Manufacturer part number
	Revision       [16]byte  // Revision
	ProductName    [64]byte  // Product name
	Reserved       [128]byte // Reserved for future use
}

// VPDData represents the VPD_R0 section structure
// VPD (Vital Product Data) contains additional product information
type VPDData struct {
	// VPD header
	Signature      [3]byte  // Should be "VPD"
	Length         uint8    // Length of VPD data
	
	// VPD fields are variable length with tag-length-value format
	// We'll store raw data and parse it separately
	Data           []byte   // Raw VPD data
}

// Unmarshal unmarshals binary data into DeviceInfo structure
func (d *DeviceInfo) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &DeviceInfoAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*d = *annotated.FromAnnotated()
	return nil
}

// Marshal marshals DeviceInfo structure into binary data
func (d *DeviceInfo) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := d.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Unmarshal unmarshals binary data into MFGInfo structure
func (m *MFGInfo) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &MFGInfoAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*m = *annotated.FromAnnotated()
	return nil
}

// Marshal marshals MFGInfo structure into binary data
func (m *MFGInfo) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := m.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Unmarshal unmarshals binary data into VPDData structure
func (v *VPDData) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling the header
	annotated := &VPDDataAnnotated{}
	if len(data) < 4 {
		return errs.ErrInvalidDataSize
	}
	
	// Unmarshal the fixed header part
	if err := annotated.Unmarshal(data[:4]); err != nil {
		return err
	}
	
	// Convert header to legacy format
	v.Signature = annotated.Signature
	v.Length = annotated.Length
	
	// Handle variable length data
	if len(data) > 4 {
		v.Data = make([]byte, len(data)-4)
		copy(v.Data, data[4:])
	}
	return nil
}

// Marshal marshals VPDData structure into binary data
func (v *VPDData) Marshal() ([]byte, error) {
	// Convert to annotated format for header
	annotated := v.ToAnnotated()
	
	// Marshal the header using annotation-based marshaling
	headerData, err := annotated.Marshal()
	if err != nil {
		return nil, err
	}
	
	// Combine header with variable data
	result := make([]byte, len(headerData)+len(v.Data))
	copy(result, headerData)
	copy(result[len(headerData):], v.Data)
	
	return result, nil
}