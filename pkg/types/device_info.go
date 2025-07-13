package types

// DeviceInfo represents the DEV_INFO section structure
// This structure contains device-specific information
type DeviceInfo struct {
	// Device identification
	DeviceID       uint16   `bin:"BE"`                // Device ID
	VendorID       uint16   `bin:"BE"`                // Vendor ID (usually 0x15b3 for Mellanox)
	SubsystemID    uint16   `bin:"BE"`                // Subsystem ID
	SubsystemVendorID uint16 `bin:"BE"`                // Subsystem vendor ID
	
	// Version information
	HWVersion      uint32   `bin:"BE"`                // Hardware version
	HWRevision     uint32   `bin:"BE"`                // Hardware revision
	
	// Device capabilities
	Capabilities   uint64   `bin:"BE"`                // Device capabilities bitmap
	
	// MAC addresses and GUIDs
	MACGUID        [8]byte  `bin:""`                  // Base MAC/GUID
	NumMACs        uint32   `bin:"BE"`                // Number of MAC addresses
	
	// Additional info
	Reserved       [64]byte `bin:""`                  // Reserved for future use
}

// MFGInfo represents the MFG_INFO section structure
// This structure contains manufacturing information
type MFGInfo struct {
	PSID           [16]byte `bin:""`                  // Product Serial ID
	PartNumber     [32]byte `bin:""`                  // Manufacturer part number
	Revision       [16]byte `bin:""`                  // Revision
	ProductName    [64]byte `bin:""`                  // Product name
	Reserved       [128]byte `bin:""`                 // Reserved for future use
}

// VPDData represents the VPD_R0 section structure
// VPD (Vital Product Data) contains additional product information
type VPDData struct {
	// VPD header
	Signature      [3]byte  `bin:""`                  // Should be "VPD"
	Length         uint8    `bin:""`                  // Length of VPD data
	
	// VPD fields are variable length with tag-length-value format
	// We'll store raw data and parse it separately
	Data           []byte   `bin:""`                  // Raw VPD data
}