package types

import (
	"testing"
)

func TestFirmwareFormat_String(t *testing.T) {
	tests := []struct {
		name   string
		format FirmwareFormat
		want   string
	}{
		{
			name:   "FS4 format",
			format: FormatFS4,
			want:   "FS4",
		},
		{
			name:   "FS5 format",
			format: FormatFS5,
			want:   "FS5",
		},
		{
			name:   "Unknown format",
			format: FormatUnknown,
			want:   "Unknown",
		},
		{
			name:   "Invalid format",
			format: FirmwareFormat(99),
			want:   "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.format.String(); got != tt.want {
				t.Errorf("FirmwareFormat.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCRCType(t *testing.T) {
	// Test that CRC type constants have expected values
	tests := []struct {
		name     string
		crcType  CRCType
		expected uint8
	}{
		{
			name:     "CRCInITOCEntry",
			crcType:  CRCInITOCEntry,
			expected: 0,
		},
		{
			name:     "CRCNone",
			crcType:  CRCNone,
			expected: 1,
		},
		{
			name:     "CRCInSection",
			crcType:  CRCInSection,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if uint8(tt.crcType) != tt.expected {
				t.Errorf("CRCType %s = %v, want %v", tt.name, uint8(tt.crcType), tt.expected)
			}
		})
	}
}

func TestFirmwareMetadata(t *testing.T) {
	// Test creating firmware metadata
	metadata := &FirmwareMetadata{
		Format:      FormatFS4,
		ImageStart:  0x1000,
		ImageSize:   0x100000,
		ChunkSize:   0x1000,
		IsEncrypted: false,
		HWPointers:  nil,
		ITOCHeader:  nil,
		DTOCHeader:  nil,
		ImageInfo:   nil,
		DeviceInfo:  nil,
		MFGInfo:     nil,
	}

	if metadata.Format != FormatFS4 {
		t.Errorf("FirmwareMetadata.Format = %v, want %v", metadata.Format, FormatFS4)
	}

	if metadata.ImageStart != 0x1000 {
		t.Errorf("FirmwareMetadata.ImageStart = %v, want %v", metadata.ImageStart, 0x1000)
	}

	if metadata.ImageSize != 0x100000 {
		t.Errorf("FirmwareMetadata.ImageSize = %v, want %v", metadata.ImageSize, 0x100000)
	}

	if metadata.IsEncrypted {
		t.Error("FirmwareMetadata.IsEncrypted should be false")
	}
}
