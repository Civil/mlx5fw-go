package sections

import (
	"encoding/json"

	"github.com/Civil/mlx5fw-go/pkg/crc"
	"github.com/Civil/mlx5fw-go/pkg/errors"
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// DeviceInfoSection represents a Device Info section
type DeviceInfoSection struct {
	*interfaces.BaseSection
	Info *types.DevInfo
}

// NewDeviceInfoSection creates a new Device Info section
func NewDeviceInfoSection(base *interfaces.BaseSection) *DeviceInfoSection {
	return &DeviceInfoSection{
		BaseSection: base,
	}
}

// Parse parses the Device Info section data
func (s *DeviceInfoSection) Parse(data []byte) error {
	s.SetRawData(data)

	s.Info = &types.DevInfo{}

	if err := s.Info.Unmarshal(data); err != nil {
		return merry.Wrap(err)
	}

	return nil
}

// MarshalJSON returns JSON representation of the Device Info section
func (s *DeviceInfoSection) MarshalJSON() ([]byte, error) {
	// Create a wrapper struct with section metadata and embedded DevInfo
	type SectionWithDevInfo struct {
		Type         uint16         `json:"type"`
		TypeName     string         `json:"type_name"`
		Offset       uint64         `json:"offset"`
		Size         uint32         `json:"size"`
		CRCType      string         `json:"crc_type"`
		IsEncrypted  bool           `json:"is_encrypted"`
		IsDeviceData bool           `json:"is_device_data"`
		HasRawData   bool           `json:"has_raw_data"`
		DeviceInfo   *types.DevInfo `json:"device_info,omitempty"`
	}

	section := &SectionWithDevInfo{
		Type:         s.Type(),
		TypeName:     s.TypeName(),
		Offset:       s.Offset(),
		Size:         s.Size(),
		CRCType:      s.CRCType().String(),
		IsEncrypted:  s.IsEncrypted(),
		IsDeviceData: s.IsDeviceData(),
		HasRawData:   s.Info == nil,
		DeviceInfo:   s.Info,
	}

	// Set the TrailerCRC field if we have DevInfo
	if s.Info != nil {
		// The TrailerCRC is the original CRC from ITOC, stored separately
		// This preserves the original value for reference only
		s.Info.TrailerCRC = uint16(s.Info.CRC & 0xFFFF)
	}

	return json.Marshal(section)
}

// VerifyCRC verifies the CRC for the DeviceInfo section
func (s *DeviceInfoSection) VerifyCRC() error {
	// DEV_INFO has a special CRC structure:
	// - The section is 512 bytes total
	// - CRC is stored at offset 508-511 as a 32-bit big-endian value
	// - The actual CRC16 is in the lower 16 bits of this 32-bit value
	// - CRC is calculated on first 508 bytes (127 dwords)

	if s.CRCType() != types.CRCInSection {
		// Use base implementation for other CRC types
		return s.BaseSection.VerifyCRC()
	}

	data := s.GetRawData()
	if len(data) < 512 {
		return merry.New("DEV_INFO section too small")
	}

	// Calculate CRC on first 508 bytes (up to but not including the CRC field)
	crcCalc := s.GetCRCHandler()
	if crcCalc == nil {
		return nil // No CRC handler
	}

	// The expected CRC is at offset 508-511 as a 32-bit big-endian value
	// Extract the lower 16 bits for the actual CRC16 value
	crc32Value := uint32(data[508])<<24 | uint32(data[509])<<16 | uint32(data[510])<<8 | uint32(data[511])
	expectedCRC := crc32Value & 0xFFFF // Lower 16 bits contain the CRC16

	// For DEV_INFO sections, we need to calculate CRC directly on the first 508 bytes
	// We can't use the default handler's behavior for CRCInSection because it would
	// subtract 4 bytes from whatever we pass, giving us the wrong range.
	// Instead, we'll create a software calculator directly.
	calc := crc.NewSoftwareCRCCalculator()
	actualCRC := calc.Calculate(data[:508])

	if uint16(actualCRC) != uint16(expectedCRC) {
		return errors.CRCMismatchError(expectedCRC, actualCRC, "device_info")
	}

	return nil
}
