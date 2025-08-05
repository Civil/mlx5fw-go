package interfaces

import (
	"testing"

	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/stretchr/testify/assert"
)

// MockCRCHandler for testing
type MockCRCHandler struct{}

func (m *MockCRCHandler) CalculateCRC(data []byte, crcType types.CRCType) (uint32, error) {
	return 0x1234, nil
}

func (m *MockCRCHandler) VerifyCRC(data []byte, expectedCRC uint32, crcType types.CRCType) error {
	return nil
}

func (m *MockCRCHandler) GetCRCOffset() int {
	return -4
}

func (m *MockCRCHandler) HasEmbeddedCRC() bool {
	return false
}

func TestSectionOptions(t *testing.T) {
	mockHandler := &MockCRCHandler{}
	
	t.Run("NewBaseSectionWithOptions_Basic", func(t *testing.T) {
		section := NewBaseSectionWithOptions(0x100, 0x1000, 512)
		
		assert.NotNil(t, section)
		assert.Equal(t, uint16(0x100), section.Type())
		assert.Equal(t, uint64(0x1000), section.Offset())
		assert.Equal(t, uint32(512), section.Size())
		
		// Check defaults
		assert.False(t, section.IsEncrypted())
		assert.False(t, section.IsDeviceData())
		assert.False(t, section.IsFromHWPointer())
		assert.False(t, section.HasCRC())
		assert.Nil(t, section.GetITOCEntry())
	})
	
	t.Run("WithCRC", func(t *testing.T) {
		section := NewBaseSectionWithOptions(0x100, 0x1000, 512,
			WithCRC(types.CRCInSection, 0x5678),
		)
		
		assert.True(t, section.HasCRC())
		assert.Equal(t, types.CRCInSection, section.CRCType())
		assert.Equal(t, uint32(0x5678), section.GetCRC())
	})
	
	t.Run("WithEncryption", func(t *testing.T) {
		section := NewBaseSectionWithOptions(0x100, 0x1000, 512,
			WithEncryption(),
		)
		
		assert.True(t, section.IsEncrypted())
	})
	
	t.Run("WithDeviceData", func(t *testing.T) {
		section := NewBaseSectionWithOptions(0x100, 0x1000, 512,
			WithDeviceData(),
		)
		
		assert.True(t, section.IsDeviceData())
	})
	
	t.Run("WithFromHWPointer", func(t *testing.T) {
		section := NewBaseSectionWithOptions(0x100, 0x1000, 512,
			WithFromHWPointer(),
		)
		
		assert.True(t, section.IsFromHWPointer())
	})
	
	t.Run("WithRawData", func(t *testing.T) {
		testData := []byte{0x01, 0x02, 0x03, 0x04}
		section := NewBaseSectionWithOptions(0x100, 0x1000, 512,
			WithRawData(testData),
		)
		
		assert.Equal(t, testData, section.GetRawData())
	})
	
	t.Run("WithITOCEntry", func(t *testing.T) {
		entry := &types.ITOCEntry{
			Type:   0x10,  // Use a value that fits in uint8
			SizeDwords: 128, // 512 bytes / 4
			FlashAddrDwords: 0x1000,
		}
		section := NewBaseSectionWithOptions(0x100, 0x1000, 512,
			WithITOCEntry(entry),
		)
		
		assert.Equal(t, entry, section.GetITOCEntry())
	})
	
	
	t.Run("WithCRCHandler", func(t *testing.T) {
		section := NewBaseSectionWithOptions(0x100, 0x1000, 512,
			WithCRCHandler(mockHandler),
		)
		
		// Set some test data
		section.rawData = []byte{1, 2, 3, 4}
		
		// Test that the handler is used
		crc, err := section.CalculateCRC()
		assert.NoError(t, err)
		assert.Equal(t, uint32(0x1234), crc)
	})
	
	t.Run("MultipleOptions", func(t *testing.T) {
		entry := &types.ITOCEntry{
			Type:   0x10,  // Use a value that fits in uint8
			SizeDwords: 128, // 512 bytes / 4
			FlashAddrDwords: 0x1000,
		}
		
		section := NewBaseSectionWithOptions(0x100, 0x1000, 512,
			WithCRC(types.CRCInSection, 0x5678),
			WithEncryption(),
			WithDeviceData(),
			WithFromHWPointer(),
			WithRawData([]byte{1, 2, 3, 4}),
			WithITOCEntry(entry),
			WithCRCHandler(mockHandler),
		)
		
		// Verify all options were applied
		assert.True(t, section.HasCRC())
		assert.Equal(t, types.CRCInSection, section.CRCType())
		assert.Equal(t, uint32(0x5678), section.GetCRC())
		assert.True(t, section.IsEncrypted())
		assert.True(t, section.IsDeviceData())
		assert.True(t, section.IsFromHWPointer())
		assert.Equal(t, []byte{1, 2, 3, 4}, section.GetRawData())
		assert.Equal(t, entry, section.GetITOCEntry())
		
		// Test CRC calculation with handler
		crc, err := section.CalculateCRC()
		assert.NoError(t, err)
		assert.Equal(t, uint32(0x1234), crc)
	})
	
	t.Run("BackwardCompatibility", func(t *testing.T) {
		// Test that the old constructor still works
		section := NewBaseSection(
			0x100,      // sectionType
			0x1000,     // offset
			512,        // size
			types.CRCInSection, // crcType
			0x5678,     // crc
			false,      // isEncrypted
			false,      // isDeviceData
			nil,        // itocEntry
			false,      // isFromHWPointer
		)
		
		assert.NotNil(t, section)
		assert.Equal(t, uint16(0x100), section.Type())
		assert.Equal(t, uint64(0x1000), section.Offset())
		assert.Equal(t, uint32(512), section.Size())
		assert.Equal(t, types.CRCInSection, section.CRCType())
		assert.Equal(t, uint32(0x5678), section.GetCRC())
	})
}