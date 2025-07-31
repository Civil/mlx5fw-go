package interfaces

import (
	"io"
	
	"github.com/Civil/mlx5fw-go/pkg/types"
)

// FirmwareParser is the main interface for parsing firmware files
type FirmwareParser interface {
	// Parse reads and parses the firmware from the provided reader
	Parse(reader io.ReaderAt) error
	
	// Query returns firmware information similar to mstflint query output
	Query() (*FirmwareInfo, error)
	
	// GetFormat returns the firmware format (FS4 or FS5)
	GetFormat() int
	
	// GetSections returns all parsed sections
	GetSections() map[uint16]*Section
	
	// GetSection returns a specific section by type
	GetSection(sectionType uint16) (*Section, error)
}

// FirmwareInfo represents the query output information
type FirmwareInfo struct {
	// Basic information
	Format          string
	FormatVersion   int
	
	// Version information
	FWVersion       string
	FWReleaseDate   string
	MICVersion      string
	ProductVersion  string
	
	// Product information
	PartNumber      string
	Description     string
	PSID            string
	OrigPSID        string  // Original PSID (shown when different from PSID)
	PRSName         string
	
	// ROM information
	RomInfo         []RomInfo
	
	// GUID/MAC information
	BaseGUID        uint64
	BaseGUIDNum     int
	BaseMAC         uint64
	BaseMACNum      int
	
	// Additional GUID/MAC for dual format (encrypted firmware)
	BaseGUID2       uint64
	BaseGUID2Num    int
	BaseMAC2        uint64
	BaseMAC2Num     int
	GUIDStep        uint8
	MACStep         uint8
	UseDualFormat   bool  // Whether to display GUID1/GUID2 format
	
	// VSD information
	ImageVSD        string
	DeviceVSD       string
	
	// Security information
	SecurityAttrs   string
	SecurityVer     int
	IsEncrypted     bool
	
	// Device information
	DeviceID        uint16
	VendorID        uint16
	
	// Size information
	ImageSize       uint64
	ChunkSize       uint64
	
	// Additional metadata
	Sections        []SectionInfo
}

// RomInfo represents ROM/expansion ROM information
type RomInfo struct {
	Type    string
	Version string
	CPU     string
}

// SectionInfo represents information about a single section
type SectionInfo struct {
	Type            uint16
	TypeName        string
	Offset          uint64
	Size            uint32
	CRCType         types.CRCType
	IsEncrypted     bool
	IsDeviceData    bool
}

// Section represents a parsed firmware section
type Section struct {
	Type            uint16
	Offset          uint64
	Size            uint32
	Data            []byte
	CRCType         types.CRCType
	CRC             uint32
	IsEncrypted     bool
	IsDeviceData    bool
	Entry           *types.ITOCEntry
	IsFromHWPointer bool  // True if section was discovered from HW pointer in encrypted firmware
}

// CRCVerifier provides CRC calculation and verification
type CRCVerifier interface {
	// CalculateSoftwareCRC calculates software CRC16 for data
	CalculateSoftwareCRC(data []byte) uint16
	
	// CalculateHardwareCRC calculates hardware CRC for data
	CalculateHardwareCRC(data []byte) uint16
	
	// VerifyCRC verifies CRC for a data buffer
	VerifyCRC(data []byte, expectedCRC uint32, useHardwareCRC bool) error
}