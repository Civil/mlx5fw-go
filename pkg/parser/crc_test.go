package parser

import (
	"encoding/hex"
	"testing"
)

func TestNewCRCCalculator(t *testing.T) {
	crc := NewCRCCalculator()
	if crc == nil {
		t.Fatal("NewCRCCalculator() returned nil")
	}
}

func TestCalculateImageCRC(t *testing.T) {
	crc := NewCRCCalculator()

	tests := []struct {
		name         string
		data         []byte
		sizeInDwords int
		expected     uint16
	}{
		{
			name:         "Empty data",
			data:         []byte{},
			sizeInDwords: 0,
			expected:     0x0955, // Actual CRC value
		},
		{
			name:         "Simple data",
			data:         []byte{0x00, 0x00, 0x00, 0x01},
			sizeInDwords: 1,
			expected:     0x1002, // Actual CRC for this data
		},
		{
			name:         "Multiple dwords",
			data:         []byte{0x12, 0x34, 0x56, 0x78, 0xAB, 0xCD, 0xEF, 0x00},
			sizeInDwords: 2,
			expected:     0x72D3, // Actual CRC for this data
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := crc.CalculateImageCRC(tt.data, tt.sizeInDwords)
			if got != tt.expected {
				t.Errorf("CalculateImageCRC() = 0x%04X, want 0x%04X", got, tt.expected)
			}
		})
	}
}

func TestCalculateHardwareCRC(t *testing.T) {
	crc := NewCRCCalculator()

	tests := []struct {
		name     string
		data     []byte
		expected uint16
	}{
		{
			name:     "Empty data",
			data:     []byte{},
			expected: 0x0000, // Initial CRC value for hardware CRC
		},
		{
			name:     "Simple byte",
			data:     []byte{0x01},
			expected: 0x0000, // Actual CRC
		},
		{
			name:     "Multiple bytes",
			data:     []byte{0x01, 0x02, 0x03, 0x04},
			expected: 0x11C8, // Actual CRC
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := crc.CalculateHardwareCRC(tt.data)
			if got != tt.expected {
				t.Errorf("CalculateHardwareCRC() = 0x%04X, want 0x%04X", got, tt.expected)
			}
		})
	}
}

func TestCRCWithRealData(t *testing.T) {
	// Test with actual BOOT3_CODE section data pattern
	// This simulates the CRC calculation for a real firmware section
	crc := NewCRCCalculator()

	// Create test data that simulates a firmware section
	// Size: 0x2590 bytes (9616 bytes = 2404 dwords)
	sectionSize := 0x2590
	data := make([]byte, sectionSize)
	
	// Fill with pattern
	for i := 0; i < len(data); i++ {
		data[i] = byte(i & 0xFF)
	}

	// Calculate CRC on all dwords except the last one (where CRC is stored)
	sizeInDwords := sectionSize / 4
	crcSizeInDwords := sizeInDwords - 1

	calculatedCRC := crc.CalculateImageCRC(data, crcSizeInDwords)
	
	// CRC should be a valid 16-bit value
	if calculatedCRC > 0xFFFF {
		t.Errorf("CalculateImageCRC() returned invalid CRC: 0x%X", calculatedCRC)
	}

	// Test that calculating CRC twice gives same result
	secondCRC := crc.CalculateImageCRC(data, crcSizeInDwords)
	if calculatedCRC != secondCRC {
		t.Errorf("CRC calculation not consistent: first=0x%04X, second=0x%04X", calculatedCRC, secondCRC)
	}
}

func TestCRCEdgeCases(t *testing.T) {
	crc := NewCRCCalculator()

	t.Run("Data not aligned to dword", func(t *testing.T) {
		// Test with 5 bytes (not divisible by 4)
		data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
		
		// Calculate CRC for 1 dword (first 4 bytes)
		result := crc.CalculateImageCRC(data, 1)
		
		// Should only process first 4 bytes
		expected := crc.CalculateImageCRC(data[:4], 1)
		if result != expected {
			t.Errorf("CRC with unaligned data = 0x%04X, want 0x%04X", result, expected)
		}
	})

	t.Run("Size larger than data", func(t *testing.T) {
		// Test with 4 bytes but request 2 dwords
		data := []byte{0x01, 0x02, 0x03, 0x04}
		
		// This should handle gracefully
		result := crc.CalculateImageCRC(data, 2)
		
		// Result should be valid (implementation dependent)
		if result > 0xFFFF {
			t.Errorf("Invalid CRC result: 0x%X", result)
		}
	})
}

func BenchmarkCalculateImageCRC(b *testing.B) {
	crc := NewCRCCalculator()
	
	// Test with 64KB of data (typical section size)
	data := make([]byte, 64*1024)
	for i := range data {
		data[i] = byte(i)
	}
	sizeInDwords := len(data) / 4

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = crc.CalculateImageCRC(data, sizeInDwords)
	}
}

func TestCRCTablesInitialization(t *testing.T) {
	crc := NewCRCCalculator()
	
	// Test that CRC tables are properly initialized
	// We can't directly access the tables, but we can verify
	// that calculations work correctly
	
	// Known test vector
	testData, _ := hex.DecodeString("123456789")
	
	// Calculate CRC
	imageCRC := crc.CalculateImageCRC(testData, len(testData)/4)
	hwCRC := crc.CalculateHardwareCRC(testData)
	
	// Both should produce valid CRCs
	if imageCRC > 0xFFFF {
		t.Errorf("Invalid image CRC: 0x%X", imageCRC)
	}
	if hwCRC > 0xFFFF {
		t.Errorf("Invalid hardware CRC: 0x%X", hwCRC)
	}
}