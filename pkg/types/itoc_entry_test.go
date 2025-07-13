package types

import (
	"encoding/hex"
	"testing"
)

func TestITOCEntry_ParseFields_Detailed(t *testing.T) {
	tests := []struct {
		name            string
		data            string // hex encoded data
		expectedType    uint8
		expectedSize    uint32
		expectedFlash   uint32
		expectedCRC     uint16
		expectedNoCRC   bool
		expectedDevData bool
	}{
		{
			name: "Basic ITOC entry",
			// Type=0x10, Size=0x600, FlashAddr=0x2000, CRC=0x1234
			// Based on the working example, let me create valid bit-packed data
			data:            "1000060000000000000000000000000000000000000020000000123400000000",
			expectedType:    0x10,
			expectedSize:    0x600,
			expectedFlash:   0x2000,
			expectedCRC:     0x1234,
			expectedNoCRC:   false,
			expectedDevData: false,
		},
		{
			name: "Entry with NoCRC flag",
			// Type=0x03, Size=0x1000, FlashAddr=0x5000, NoCRC=1
			data:            "0300100000000000000000000000000000000000000050000001000000000000",
			expectedType:    0x03,
			expectedSize:    0x1000,
			expectedFlash:   0x5000,
			expectedCRC:     0x0000,
			expectedNoCRC:   true,
			expectedDevData: false,
		},
		{
			name: "Device data entry",
			// Type=0xE0, Size=0x400, FlashAddr=0x100000, DeviceData=1
			data:            "e000040000000000000000000000000000000000001000000000000000000000",
			expectedType:    0xE0,
			expectedSize:    0x400,
			expectedFlash:   0x100000,
			expectedCRC:     0x0000,
			expectedNoCRC:   false,
			expectedDevData: false, // ParseFields doesn't set this, it's external
		},
		{
			name: "Empty entry (all zeros)",
			data:            "0000000000000000000000000000000000000000000000000000000000000000",
			expectedType:    0x00,
			expectedSize:    0x0,
			expectedFlash:   0x0,
			expectedCRC:     0x0000,
			expectedNoCRC:   false,
			expectedDevData: false,
		},
		{
			name: "End marker entry",
			data:            "ff00000000000000000000000000000000000000000000000000000000000000",
			expectedType:    0xFF,
			expectedSize:    0x0,
			expectedFlash:   0x0,
			expectedCRC:     0x0000,
			expectedNoCRC:   false,
			expectedDevData: false,
		},
		{
			name: "Large section entry",
			// Type=0x01, Size=0x100000 (1MB), FlashAddr=0x40000
			data:            "0110000000000000000000000000000000000000000400000000567800000000",
			expectedType:    0x01,
			expectedSize:    0x100000,
			expectedFlash:   0x40000,
			expectedCRC:     0x5678,
			expectedNoCRC:   false,
			expectedDevData: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Decode hex data
			data, err := hex.DecodeString(tt.data)
			if err != nil {
				t.Fatalf("Failed to decode hex: %v", err)
			}

			// Create ITOC entry and copy data
			entry := &ITOCEntry{}
			copy(entry.Data[:], data)

			// Parse fields
			entry.ParseFields()

			// Check parsed values
			if entry.Type != tt.expectedType {
				t.Errorf("Type = 0x%02X, want 0x%02X", entry.Type, tt.expectedType)
			}
			if entry.Size != tt.expectedSize {
				t.Errorf("Size = 0x%X, want 0x%X", entry.Size, tt.expectedSize)
			}
			if entry.FlashAddr != tt.expectedFlash {
				t.Errorf("FlashAddr = 0x%X, want 0x%X", entry.FlashAddr, tt.expectedFlash)
			}
			if entry.SectionCRC != tt.expectedCRC {
				t.Errorf("SectionCRC = 0x%04X, want 0x%04X", entry.SectionCRC, tt.expectedCRC)
			}
			if entry.GetNoCRC() != tt.expectedNoCRC {
				t.Errorf("GetNoCRC() = %v, want %v", entry.GetNoCRC(), tt.expectedNoCRC)
			}
			if entry.GetType() != uint16(tt.expectedType) {
				t.Errorf("GetType() = 0x%02X, want 0x%02X", entry.GetType(), tt.expectedType)
			}
		})
	}
}

func TestITOCEntry_GetMethods(t *testing.T) {
	// Test the getter methods
	entry := &ITOCEntry{
		Type:       0x10,
		CRC:        1,    // NoCRC
		FlashAddr:  0x2000,
		Size:       0x600,
		SectionCRC: 0x1234,
	}

	t.Run("GetType", func(t *testing.T) {
		if got := entry.GetType(); got != 0x10 {
			t.Errorf("GetType() = 0x%02X, want 0x10", got)
		}
	})

	t.Run("GetNoCRC", func(t *testing.T) {
		if got := entry.GetNoCRC(); !got {
			t.Error("GetNoCRC() = false, want true")
		}
	})

	t.Run("GetNoCRC when zero", func(t *testing.T) {
		entry.CRC = 0
		if got := entry.GetNoCRC(); got {
			t.Error("GetNoCRC() = true, want false")
		}
	})
}

func TestITOCEntry_ParseFields_EdgeCases(t *testing.T) {
	t.Run("Partial data", func(t *testing.T) {
		// Test with less than 32 bytes
		entry := &ITOCEntry{}
		// Only fill first few bytes
		entry.Data[0] = 0x10 // Type

		entry.ParseFields()

		// Should still parse what it can
		if entry.Type != 0x10 {
			t.Errorf("Type = 0x%02X, want 0x10", entry.Type)
		}
		// Other fields should be zero
		if entry.Size != 0 {
			t.Errorf("Size = %d, want 0", entry.Size)
		}
	})

	t.Run("CRC flag variations", func(t *testing.T) {
		testCases := []struct {
			name     string
			data     string
			wantNoCRC bool
		}{
			{
				name:     "CRC flag clear",
				data:     "1000060000000000000000000000000000000000000020000000123400000000",
				wantNoCRC: false,
			},
			{
				name:     "CRC flag set",
				data:     "1000060000000000000000000000000000000000000020000001123400000000",
				wantNoCRC: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				data, _ := hex.DecodeString(tc.data)
				entry := &ITOCEntry{}
				copy(entry.Data[:], data)
				entry.ParseFields()

				if got := entry.GetNoCRC(); got != tc.wantNoCRC {
					t.Errorf("GetNoCRC() = %v, want %v", got, tc.wantNoCRC)
				}
			})
		}
	})
}

func TestITOCEntry_SectionTypes(t *testing.T) {
	// Test various section types
	sectionTypes := []uint8{
		0x01, // BOOT_CODE
		0x03, // MAIN_CODE
		0x10, // IMAGE_INFO
		0xE0, // MFG_INFO
		0xFF, // END
	}

	for _, sType := range sectionTypes {
		t.Run(GetSectionTypeName(uint16(sType)), func(t *testing.T) {
			entry := &ITOCEntry{
				Type: sType,
			}
			
			if got := entry.GetType(); got != uint16(sType) {
				t.Errorf("GetType() = 0x%02X, want 0x%02X", got, sType)
			}
		})
	}
}

func TestITOCEntry_RealWorldData(t *testing.T) {
	// Test with real-world ITOC entry patterns
	realEntries := []struct {
		name        string
		description string
		data        string
	}{
		{
			name:        "IMAGE_INFO section",
			description: "Typical IMAGE_INFO section entry",
			data:        "10000006000000000002A00000000000000000000000000000000000009B48",
		},
		{
			name:        "BOOT3_CODE section",
			description: "BOOT3_CODE with CRC",
			data:        "0F000025900000000017E00000000000000000000000000000000000008E9C",
		},
		{
			name:        "MFG_INFO device data",
			description: "Device data section",
			data:        "E00000100000000001FED00000000000000000000000000000000000000000",
		},
	}

	for _, re := range realEntries {
		t.Run(re.name, func(t *testing.T) {
			data, err := hex.DecodeString(re.data)
			if err != nil {
				t.Fatalf("Failed to decode hex: %v", err)
			}

			entry := &ITOCEntry{}
			copy(entry.Data[:], data)
			entry.ParseFields()

			// Basic sanity checks
			if entry.Type == 0 && entry.Size == 0 && entry.FlashAddr == 0 {
				t.Error("All fields are zero, parsing likely failed")
			}

			// Log parsed values for debugging
			t.Logf("%s: Type=0x%02X, Size=0x%X, FlashAddr=0x%X, CRC=0x%04X, NoCRC=%v",
				re.description, entry.Type, entry.Size, entry.FlashAddr, 
				entry.SectionCRC, entry.GetNoCRC())
		})
	}
}

func BenchmarkITOCEntry_ParseFields(b *testing.B) {
	// Benchmark parsing performance
	data, _ := hex.DecodeString("10000006000000000002A00000000000000000000000000000000000009B48")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry := &ITOCEntry{}
		copy(entry.Data[:], data)
		entry.ParseFields()
	}
}

func TestITOCEntry_DataIntegrity(t *testing.T) {
	// Test that raw data is preserved correctly
	testData := "10000006000000000002A00000000000000000000000000000000000009B48"
	data, _ := hex.DecodeString(testData)
	
	entry := &ITOCEntry{}
	copy(entry.Data[:], data)
	
	// Verify data was copied correctly
	for i, b := range data {
		if entry.Data[i] != b {
			t.Errorf("Data[%d] = 0x%02X, want 0x%02X", i, entry.Data[i], b)
		}
	}
	
	// Parse and verify data is still intact
	entry.ParseFields()
	
	for i, b := range data {
		if entry.Data[i] != b {
			t.Errorf("After parse: Data[%d] = 0x%02X, want 0x%02X", i, entry.Data[i], b)
		}
	}
}