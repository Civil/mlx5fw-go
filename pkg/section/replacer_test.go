package section

import (
	"encoding/binary"
	"testing"

	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestDetermineFirmwareSizeLimit(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	tests := []struct {
		name         string
		firmwareSize int
		expectedSize int
	}{
		{
			name:         "32MB firmware",
			firmwareSize: 0x2000000 + 0x10000, // ~32MB
			expectedSize: FirmwareSize32MB,
		},
		{
			name:         "64MB firmware",
			firmwareSize: 0x4000000 + 0x10000, // ~64MB
			expectedSize: FirmwareSize64MB,
		},
		{
			name:         "small firmware defaults to 32MB",
			firmwareSize: 0x1000000, // 16MB
			expectedSize: FirmwareSize32MB,
		},
		{
			name:         "large firmware defaults to 64MB",
			firmwareSize: 0x3000000, // 48MB
			expectedSize: FirmwareSize64MB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Replacer{
				firmwareData: make([]byte, tt.firmwareSize),
				logger:       logger,
			}
			
			limit := r.determineFirmwareSizeLimit()
			assert.Equal(t, tt.expectedSize, limit)
		})
	}
}

func TestGetCRCType(t *testing.T) {
	logger := zaptest.NewLogger(t)
	r := &Replacer{logger: logger}

	tests := []struct {
		name        string
		entry       *types.ITOCEntry
		expectedCRC types.CRCType
		setupEntry  func(*types.ITOCEntry)
	}{
		{
			name: "no CRC flag set",
			entry: &types.ITOCEntry{
				Data: [32]byte{},
			},
			expectedCRC: types.CRCNone,
			setupEntry: func(e *types.ITOCEntry) {
				// Set CRC field to 1 (NOCRC) at bits 205-207
				// Bit 207 is the LSB of the 3-bit field
				e.Data[25] |= 0x01 // Set bit 207 (LSB of 3-bit field for value 1)
			},
		},
		{
			name: "CRC in ITOC entry",
			entry: &types.ITOCEntry{
				Data: [32]byte{},
			},
			expectedCRC: types.CRCInITOCEntry,
			setupEntry: func(e *types.ITOCEntry) {
				// Set SectionCRC bits (208-223) to non-zero value
				// Bits 208-223 = byte 26 bit 0 to byte 27 bit 7
				e.Data[26] = 0x12
				e.Data[27] = 0x34
			},
		},
		{
			name: "CRC in section",
			entry: &types.ITOCEntry{
				Data: [32]byte{},
			},
			expectedCRC: types.CRCInSection,
			setupEntry: func(e *types.ITOCEntry) {
				// Set CRC field to 7 at bits 205-207
				// For value 7, we need bit pattern 111
				e.Data[25] |= 0x07 // Set bits 205-207 to 111 (value 7)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupEntry != nil {
				tt.setupEntry(tt.entry)
			}
			tt.entry.ParseFields()
			
			crcType := r.tocReader.GetCRCType(tt.entry)
			assert.Equal(t, tt.expectedCRC, crcType)
		})
	}
}

func TestSetBits(t *testing.T) {
	logger := zaptest.NewLogger(t)
	r := &Replacer{logger: logger}

	tests := []struct {
		name      string
		dataSize  int
		bitOffset int
		bitCount  int
		value     uint32
		expected  []byte
	}{
		{
			name:      "set single bit",
			dataSize:  1,
			bitOffset: 3,
			bitCount:  1,
			value:     1,
			expected:  []byte{0x10}, // bit 3 set
		},
		{
			name:      "set multiple bits",
			dataSize:  2,
			bitOffset: 4,
			bitCount:  8,
			value:     0xFF,
			expected:  []byte{0x0F, 0xF0},
		},
		{
			name:      "clear bits",
			dataSize:  1,
			bitOffset: 0,
			bitCount:  4,
			value:     0,
			expected:  []byte{0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.dataSize)
			if tt.name == "clear bits" {
				data[0] = 0xFF // Start with all bits set
				tt.expected[0] = 0x0F // Expect lower 4 bits cleared
			}
			
			r.setBits(data, tt.bitOffset, tt.bitCount, tt.value)
			assert.Equal(t, tt.expected, data)
		})
	}
}

func TestFindMagicPattern(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	tests := []struct {
		name           string
		setupData      func() []byte
		expectedOffset uint32
		expectError    bool
	}{
		{
			name: "magic at offset 0",
			setupData: func() []byte {
				data := make([]byte, 0x1000)
				binary.BigEndian.PutUint64(data[0:8], types.MagicPattern)
				return data
			},
			expectedOffset: 0,
		},
		{
			name: "magic at offset 0x10000",
			setupData: func() []byte {
				data := make([]byte, 0x20000)
				binary.BigEndian.PutUint64(data[0x10000:0x10008], types.MagicPattern)
				return data
			},
			expectedOffset: 0x10000,
		},
		{
			name: "magic not found",
			setupData: func() []byte {
				return make([]byte, 0x1000)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Replacer{
				firmwareData: tt.setupData(),
				logger:       logger,
			}
			
			offset, err := r.findMagicPattern(r.firmwareData)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedOffset, offset)
			}
		})
	}
}

func TestWriteHWPointers(t *testing.T) {
	logger := zaptest.NewLogger(t)
	r := &Replacer{logger: logger}

	hw := &types.FS4HWPointers{
		BootRecordPtr: types.HWPointerEntry{
			Ptr: 0x12345678,
			CRC: 0xABCD,
		},
		Boot2Ptr: types.HWPointerEntry{
			Ptr: 0x87654321,
			CRC: 0xDCBA,
		},
		TOCPtr: types.HWPointerEntry{
			Ptr: 0x11223344,
			CRC: 0x5566,
		},
	}

	data := make([]byte, 128)
	r.writeHWPointers(data, hw)

	// Verify written values
	assert.Equal(t, uint32(0x12345678), binary.BigEndian.Uint32(data[0:4]))
	assert.Equal(t, uint32(0xABCD), binary.BigEndian.Uint32(data[4:8]))
	assert.Equal(t, uint32(0x87654321), binary.BigEndian.Uint32(data[8:12]))
	assert.Equal(t, uint32(0xDCBA), binary.BigEndian.Uint32(data[12:16]))
	assert.Equal(t, uint32(0x11223344), binary.BigEndian.Uint32(data[16:20]))
	assert.Equal(t, uint32(0x5566), binary.BigEndian.Uint32(data[20:24]))
}

func TestUpdateSectionCRC(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	tests := []struct {
		name        string
		section     *interfaces.Section
		workingData []byte
		newSize     uint32
		expectError bool
	}{
		{
			name: "CRC none - no update",
			section: &interfaces.Section{
				Type:    types.SectionTypeDbgFWINI,
				Offset:  0,
				Size:    100,
				CRCType: types.CRCNone,
			},
			workingData: make([]byte, 200),
			newSize:     100,
		},
		{
			name: "CRC in ITOC entry - deferred",
			section: &interfaces.Section{
				Type:    types.SectionTypeItoc,
				Offset:  0,
				Size:    100,
				CRCType: types.CRCInITOCEntry,
			},
			workingData: make([]byte, 200),
			newSize:     100,
		},
		{
			name: "CRC in section",
			section: &interfaces.Section{
				Type:    types.SectionTypeDbgFWINI,
				Offset:  0,
				Size:    100,
				CRCType: types.CRCInSection,
			},
			workingData: func() []byte {
				data := make([]byte, 200)
				// Fill with test data
				for i := 0; i < 100; i++ {
					data[i] = byte(i)
				}
				return data
			}(),
			newSize: 100,
		},
		{
			name: "section too small for CRC",
			section: &interfaces.Section{
				Type:    types.SectionTypeDbgFWINI,
				Offset:  0,
				Size:    2,
				CRCType: types.CRCInSection,
			},
			workingData: make([]byte, 10),
			newSize:     2,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Replacer{
				logger: logger,
			}
			
			err := r.updateSectionCRC(tt.workingData, tt.section, tt.newSize)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				
				// For CRC in section, verify CRC was written
				if tt.section.CRCType == types.CRCInSection {
					lastDwordOffset := tt.section.Offset + uint64(tt.newSize) - 4
					lastDword := binary.BigEndian.Uint32(tt.workingData[lastDwordOffset:])
					crc := uint16(lastDword & 0xFFFF)
					assert.NotZero(t, crc, "CRC should be calculated and written")
				}
			}
		})
	}
}

func TestSerializeITOCEntry(t *testing.T) {
	logger := zaptest.NewLogger(t)
	r := &Replacer{logger: logger}

	entry := &types.ITOCEntry{
		Type:       types.SectionTypeDbgFWINI,
		Size:       0x1000, // 4KB
		Param0:     0x12345,
		Param1:     0xABCDEF00,
		FlashAddr:  0x100000, // 1MB offset
		Encrypted:  true,
		CRC:        7,
		SectionCRC: 0x5678,
	}

	// Initialize entry data
	for i := range entry.Data {
		entry.Data[i] = 0
	}

	data := make([]byte, 32)
	err := r.serializeITOCEntry(entry, data)
	require.NoError(t, err)

	// Verify key fields were serialized correctly
	// Type should be in bits 0-7
	assert.Equal(t, uint8(types.SectionTypeDbgFWINI), data[0])
	
	// Size in dwords should be in bits 8-29
	sizeDwords := entry.Size >> 2
	// Read bits 8-29 
	sizeFromData := (uint32(data[1])<<14 | uint32(data[2])<<6 | uint32(data[3])>>2) & 0x3FFFFF
	assert.Equal(t, sizeDwords, sizeFromData)
	
	// Param1 should be at bytes 8-11
	assert.Equal(t, entry.Param1, binary.BigEndian.Uint32(data[8:12]))
}

func TestRelocateSections(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	// Create test data with sections
	workingData := make([]byte, 1024)
	
	// Fill sections with identifiable data
	section1Data := []byte("SECTION1")
	section2Data := []byte("SECTION2")
	section3Data := []byte("SECTION3")
	
	copy(workingData[100:], section1Data)
	copy(workingData[200:], section2Data)
	copy(workingData[300:], section3Data)
	
	relocMap := map[uint32]*relocationInfo{
		200: {
			newOffset: 250, // Move forward by 50
			size:      uint32(len(section2Data)),
		},
		300: {
			newOffset: 350, // Move forward by 50
			size:      uint32(len(section3Data)),
		},
	}
	
	r := &Replacer{
		logger: logger,
	}
	
	err := r.relocateSections(workingData, relocMap, 150, 50)
	require.NoError(t, err)
	
	// Verify sections were moved
	assert.Equal(t, section1Data, workingData[100:100+len(section1Data)])
	assert.Equal(t, section2Data, workingData[250:250+len(section2Data)])
	assert.Equal(t, section3Data, workingData[350:350+len(section3Data)])
}

func TestUpdateHWPointers(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	// Create test firmware data with magic pattern and HW pointers
	workingData := make([]byte, 0x1000)
	
	// Write magic pattern at offset 0
	binary.BigEndian.PutUint64(workingData[0:8], types.MagicPattern)
	
	// Create HW pointers at offset 0x18
	hwPointersOffset := uint32(0x18)
	hw := &types.FS4HWPointers{
		Boot2Ptr: types.HWPointerEntry{
			Ptr: 0x100,
			CRC: 0x1234,
		},
		TOCPtr: types.HWPointerEntry{
			Ptr: 0x200,
			CRC: 0x5678,
		},
	}
	
	r := &Replacer{
		firmwareData: workingData,
		logger:       logger,
	}
	
	// Write initial HW pointers
	r.writeHWPointers(workingData[hwPointersOffset:], hw)
	
	// Create relocation map
	relocMap := map[uint32]*relocationInfo{
		0x100: {
			newOffset: 0x150,
			size:      100,
		},
		0x200: {
			newOffset: 0x250,
			size:      200,
		},
	}
	
	err := r.updateHWPointers(workingData, relocMap)
	require.NoError(t, err)
	
	// Read back HW pointers
	hwData := workingData[hwPointersOffset : hwPointersOffset+128]
	
	// Manually parse the updated pointers
	boot2Ptr := binary.BigEndian.Uint32(hwData[8:12])
	tocPtr := binary.BigEndian.Uint32(hwData[16:20])
	
	assert.Equal(t, uint32(0x150), boot2Ptr)
	assert.Equal(t, uint32(0x250), tocPtr)
}

func TestPadFirmware(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	tests := []struct {
		name           string
		inputSize      int
		expectedSize   int
		expectedFill   byte
	}{
		{
			name:         "small firmware padded to 32MB",
			inputSize:    0x1000000, // 16MB
			expectedSize: FirmwareSize32MB,
			expectedFill: 0xFF,
		},
		{
			name:         "32MB firmware stays same",
			inputSize:    FirmwareSize32MB,
			expectedSize: FirmwareSize32MB,
			expectedFill: 0xFF,
		},
		{
			name:         "large firmware padded to 64MB",
			inputSize:    0x3000000, // 48MB
			expectedSize: FirmwareSize64MB,
			expectedFill: 0xFF,
		},
		{
			name:         "oversized firmware truncated to 64MB",
			inputSize:    FirmwareSize64MB + 0x100000,
			expectedSize: FirmwareSize64MB,
			expectedFill: 0xFF,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create input data with pattern
			inputData := make([]byte, tt.inputSize)
			for i := range inputData {
				inputData[i] = byte(i & 0xFF)
			}
			
			r := &Replacer{
				firmwareData: inputData,
				logger:       logger,
			}
			
			result := r.padFirmware(inputData)
			
			// Check size
			assert.Equal(t, tt.expectedSize, len(result))
			
			// Check original data preserved
			preservedSize := tt.inputSize
			if preservedSize > tt.expectedSize {
				preservedSize = tt.expectedSize
			}
			for i := 0; i < preservedSize; i++ {
				assert.Equal(t, byte(i&0xFF), result[i], "Original data should be preserved at index %d", i)
			}
			
			// Check padding
			if tt.inputSize < tt.expectedSize {
				for i := tt.inputSize; i < tt.expectedSize; i++ {
					assert.Equal(t, tt.expectedFill, result[i], "Padding should be 0xFF at index %d", i)
				}
			}
		})
	}
}