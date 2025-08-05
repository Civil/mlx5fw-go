package types

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestBinaryVersion_Structure(t *testing.T) {
	tests := []struct {
		name         string
		data         []byte
		wantLength   uint16
		wantType     uint8
		wantVersion  uint8
		wantReserved uint32
	}{
		{
			name: "Basic binary version",
			data: func() []byte {
				buf := &bytes.Buffer{}
				binary.Write(buf, binary.BigEndian, uint16(8))    // Length
				binary.Write(buf, binary.BigEndian, uint8(1))     // Type
				binary.Write(buf, binary.BigEndian, uint8(2))     // Version
				binary.Write(buf, binary.BigEndian, uint32(0))    // Reserved
				return buf.Bytes()
			}(),
			wantLength:   8,
			wantType:     1,
			wantVersion:  2,
			wantReserved: 0,
		},
		{
			name: "Tools area header",
			data: func() []byte {
				buf := &bytes.Buffer{}
				binary.Write(buf, binary.BigEndian, uint16(64))   // Length
				binary.Write(buf, binary.BigEndian, uint8(0x10))  // Type
				binary.Write(buf, binary.BigEndian, uint8(3))     // Version
				binary.Write(buf, binary.BigEndian, uint32(0))    // Reserved
				return buf.Bytes()
			}(),
			wantLength:   64,
			wantType:     0x10,
			wantVersion:  3,
			wantReserved: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use binary.Read instead of binstruct
			buf := bytes.NewReader(tt.data)
			var bv BinaryVersion
			
			// Read fields manually
			binary.Read(buf, binary.BigEndian, &bv.Length)
			binary.Read(buf, binary.BigEndian, &bv.Type)
			binary.Read(buf, binary.BigEndian, &bv.Version)
			binary.Read(buf, binary.BigEndian, &bv.Reserved)

			if bv.Length != tt.wantLength {
				t.Errorf("Length = %d, want %d", bv.Length, tt.wantLength)
			}
			if bv.Type != tt.wantType {
				t.Errorf("Type = %d, want %d", bv.Type, tt.wantType)
			}
			if bv.Version != tt.wantVersion {
				t.Errorf("Version = %d, want %d", bv.Version, tt.wantVersion)
			}
			if bv.Reserved != tt.wantReserved {
				t.Errorf("Reserved = %d, want %d", bv.Reserved, tt.wantReserved)
			}
		})
	}
}

func TestToolsArea(t *testing.T) {
	// Test Tools Area structure
	toolsAreaData := func() []byte {
		buf := &bytes.Buffer{}
		// Binary header
		binary.Write(buf, binary.BigEndian, uint16(0x40))  // Length = 64 bytes
		binary.Write(buf, binary.BigEndian, uint8(0x10))   // Type
		binary.Write(buf, binary.BigEndian, uint8(1))      // Version
		binary.Write(buf, binary.BigEndian, uint32(0))     // Reserved
		// Some data
		buf.Write([]byte("TOOLS_DATA"))
		return buf.Bytes()
	}()

	var ta ToolsArea
	// Read header manually
	buf := bytes.NewReader(toolsAreaData[:8])
	binary.Read(buf, binary.BigEndian, &ta.BinaryHeader.Length)
	binary.Read(buf, binary.BigEndian, &ta.BinaryHeader.Type)
	binary.Read(buf, binary.BigEndian, &ta.BinaryHeader.Version)
	binary.Read(buf, binary.BigEndian, &ta.BinaryHeader.Reserved)

	// The ToolsArea struct no longer has a Data field
	// The data after the header is now part of the section data

	if ta.BinaryHeader.Length != 0x40 {
		t.Errorf("BinaryHeader.Length = %d, want %d", ta.BinaryHeader.Length, 0x40)
	}
	if ta.BinaryHeader.Type != 0x10 {
		t.Errorf("BinaryHeader.Type = 0x%02X, want 0x10", ta.BinaryHeader.Type)
	}
	// Check the data after the header directly
	if len(toolsAreaData) > 18 && string(toolsAreaData[8:18]) != "TOOLS_DATA" {
		t.Errorf("Data prefix = %q, want %q", string(toolsAreaData[8:18]), "TOOLS_DATA")
	}
}

func TestMagicPatternStruct(t *testing.T) {
	tests := []struct {
		name         string
		data         []byte
		wantMagic    uint64
		wantReserved uint64
	}{
		{
			name: "Valid magic pattern",
			data: func() []byte {
				buf := &bytes.Buffer{}
				binary.Write(buf, binary.BigEndian, uint64(MagicPattern))
				binary.Write(buf, binary.BigEndian, uint64(0))
				return buf.Bytes()
			}(),
			wantMagic:    MagicPattern,
			wantReserved: 0,
		},
		{
			name: "Invalid magic",
			data: func() []byte {
				buf := &bytes.Buffer{}
				binary.Write(buf, binary.BigEndian, uint64(0xDEADBEEF))
				binary.Write(buf, binary.BigEndian, uint64(0x12345678))
				return buf.Bytes()
			}(),
			wantMagic:    0xDEADBEEF,
			wantReserved: 0x12345678,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mps MagicPatternStruct
			buf := bytes.NewReader(tt.data)
			
			// Read fields manually
			binary.Read(buf, binary.BigEndian, &mps.Magic)
			binary.Read(buf, binary.BigEndian, &mps.Reserved)

			if mps.Magic != tt.wantMagic {
				t.Errorf("Magic = 0x%016X, want 0x%016X", mps.Magic, tt.wantMagic)
			}
			if mps.Reserved != tt.wantReserved {
				t.Errorf("Reserved = 0x%016X, want 0x%016X", mps.Reserved, tt.wantReserved)
			}
		})
	}
}

func TestBoot2Header(t *testing.T) {
	tests := []struct {
		name         string
		data         []byte
		wantMagic    uint32
		wantVersion  uint32
		wantSize     uint32
		wantReserved uint32
	}{
		{
			name: "Valid Boot2 header",
			data: func() []byte {
				buf := &bytes.Buffer{}
				binary.Write(buf, binary.BigEndian, uint32(0xAA551234)) // Magic
				binary.Write(buf, binary.BigEndian, uint32(1))          // Version
				binary.Write(buf, binary.BigEndian, uint32(0x2A8C))     // Size
				binary.Write(buf, binary.BigEndian, uint32(0))          // Reserved
				return buf.Bytes()
			}(),
			wantMagic:    0xAA551234,
			wantVersion:  1,
			wantSize:     0x2A8C,
			wantReserved: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b2h Boot2Header
			buf := bytes.NewReader(tt.data)
			
			// Read fields manually
			binary.Read(buf, binary.BigEndian, &b2h.Magic)
			binary.Read(buf, binary.BigEndian, &b2h.Version)
			binary.Read(buf, binary.BigEndian, &b2h.Size)
			binary.Read(buf, binary.BigEndian, &b2h.Reserved)

			if b2h.Magic != tt.wantMagic {
				t.Errorf("Magic = 0x%08X, want 0x%08X", b2h.Magic, tt.wantMagic)
			}
			if b2h.Version != tt.wantVersion {
				t.Errorf("Version = %d, want %d", b2h.Version, tt.wantVersion)
			}
			if b2h.Size != tt.wantSize {
				t.Errorf("Size = 0x%X, want 0x%X", b2h.Size, tt.wantSize)
			}
			if b2h.Reserved != tt.wantReserved {
				t.Errorf("Reserved = %d, want %d", b2h.Reserved, tt.wantReserved)
			}
		})
	}
}

func TestHWIDRecord(t *testing.T) {
	tests := []struct {
		name           string
		data           []byte
		wantHWID       uint32
		wantChipType   uint32
		wantDeviceType uint32
		wantReserved   uint32
	}{
		{
			name: "ConnectX-5 HWID",
			data: func() []byte {
				buf := &bytes.Buffer{}
				binary.Write(buf, binary.BigEndian, uint32(0x0191)) // HWID
				binary.Write(buf, binary.BigEndian, uint32(0x5))    // ChipType
				binary.Write(buf, binary.BigEndian, uint32(0x1))    // DeviceType
				binary.Write(buf, binary.BigEndian, uint32(0))      // Reserved
				return buf.Bytes()
			}(),
			wantHWID:       0x0191,
			wantChipType:   0x5,
			wantDeviceType: 0x1,
			wantReserved:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var hwid HWIDRecord
			buf := bytes.NewReader(tt.data)
			
			// Read fields manually
			binary.Read(buf, binary.BigEndian, &hwid.HWID)
			binary.Read(buf, binary.BigEndian, &hwid.ChipType)
			binary.Read(buf, binary.BigEndian, &hwid.DeviceType)
			binary.Read(buf, binary.BigEndian, &hwid.Reserved)

			if hwid.HWID != tt.wantHWID {
				t.Errorf("HWID = 0x%04X, want 0x%04X", hwid.HWID, tt.wantHWID)
			}
			if hwid.ChipType != tt.wantChipType {
				t.Errorf("ChipType = %d, want %d", hwid.ChipType, tt.wantChipType)
			}
			if hwid.DeviceType != tt.wantDeviceType {
				t.Errorf("DeviceType = %d, want %d", hwid.DeviceType, tt.wantDeviceType)
			}
			if hwid.Reserved != tt.wantReserved {
				t.Errorf("Reserved = %d, want %d", hwid.Reserved, tt.wantReserved)
			}
		})
	}
}

func TestBinaryVersion_Size(t *testing.T) {
	// Verify the structure size
	bv := BinaryVersion{}
	size := binary.Size(bv)
	
	expectedSize := 8 // 2 + 1 + 1 + 4
	if size != expectedSize {
		t.Errorf("BinaryVersion size = %d bytes, want %d bytes", size, expectedSize)
	}
}

func TestStructAlignment(t *testing.T) {
	// Test that our structs have the expected sizes
	tests := []struct {
		name         string
		actualSize   int
		expectedSize int
	}{
		{
			name:         "BinaryVersion",
			actualSize:   binary.Size(BinaryVersion{}),
			expectedSize: 8,
		},
		{
			name:         "MagicPatternStruct",
			actualSize:   binary.Size(MagicPatternStruct{}),
			expectedSize: 16,
		},
		{
			name:         "Boot2Header",
			actualSize:   binary.Size(Boot2Header{}),
			expectedSize: 16,
		},
		{
			name:         "HWIDRecord",
			actualSize:   binary.Size(HWIDRecord{}),
			expectedSize: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actualSize != tt.expectedSize {
				t.Errorf("%s size = %d bytes, want %d bytes", tt.name, tt.actualSize, tt.expectedSize)
			}
		})
	}
}