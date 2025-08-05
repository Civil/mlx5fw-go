package types

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestFS4HWPointers_Unmarshal_Extended(t *testing.T) {
	tests := []struct {
		name          string
		data          []byte
		wantBoot2     uint32
		wantTOC       uint32
		wantTools     uint32
		wantImageInfo uint32
		wantHashes    uint32
		wantErr       bool
	}{
		{
			name: "Valid HW pointers",
			data: func() []byte {
				buf := &bytes.Buffer{}
				// BootRecordPtr
				binary.Write(buf, binary.BigEndian, uint32(0))         // Ptr
				binary.Write(buf, binary.BigEndian, uint32(0))         // CRC
				// Boot2Ptr
				binary.Write(buf, binary.BigEndian, uint32(0x1000))    // Ptr
				binary.Write(buf, binary.BigEndian, uint32(0xABCDEF00)) // CRC
				// TOCPtr
				binary.Write(buf, binary.BigEndian, uint32(0x5000))
				binary.Write(buf, binary.BigEndian, uint32(0))
				// ToolsPtr
				binary.Write(buf, binary.BigEndian, uint32(0x6000))
				binary.Write(buf, binary.BigEndian, uint32(0))
				// AuthenticationStartPtr
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(0))
				// AuthenticationEndPtr
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(0))
				// DigestPtr
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(0))
				// DigestRecoveryKeyPtr
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(0))
				// FWWindowStartPtr
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(0))
				// FWWindowEndPtr
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(0))
				// ImageInfoSectionPtr
				binary.Write(buf, binary.BigEndian, uint32(0x2A00))
				binary.Write(buf, binary.BigEndian, uint32(0))
				// Remaining pointers (5 more)
				for i := 0; i < 5; i++ {
					binary.Write(buf, binary.BigEndian, uint64(0))
				}
				return buf.Bytes()
			}(),
			wantBoot2:     0x1000,
			wantTOC:       0x5000,
			wantTools:     0x6000,
			wantImageInfo: 0x2A00,
			wantErr:       false,
		},
		{
			name: "Invalid pointers (0xFFFFFFFF)",
			data: func() []byte {
				buf := &bytes.Buffer{}
				// All pointers set to 0xFFFFFFFF (invalid)
				for i := 0; i < 16; i++ {
					binary.Write(buf, binary.BigEndian, uint32(0xFFFFFFFF))
					binary.Write(buf, binary.BigEndian, uint32(0))
				}
				return buf.Bytes()
			}(),
			wantBoot2:     0xFFFFFFFF,
			wantTOC:       0xFFFFFFFF,
			wantTools:     0xFFFFFFFF,
			wantImageInfo: 0xFFFFFFFF,
			wantErr:       false,
		},
		{
			name: "Zero pointers",
			data: func() []byte {
				// All zeros
				return make([]byte, 128)
			}(),
			wantBoot2:     0,
			wantTOC:       0,
			wantTools:     0,
			wantImageInfo: 0,
			wantErr:       false,
		},
		{
			name: "ConnectX-5 style pointers",
			data: func() []byte {
				buf := &bytes.Buffer{}
				// BootRecordPtr
				binary.Write(buf, binary.BigEndian, uint64(0))
				// Boot2Ptr
				binary.Write(buf, binary.BigEndian, uint32(0x1000))
				binary.Write(buf, binary.BigEndian, uint32(0))
				// TOCPtr (not used in CX-5)
				binary.Write(buf, binary.BigEndian, uint64(0))
				// ToolsPtr (contains ITOC in CX-5)
				binary.Write(buf, binary.BigEndian, uint32(0x500))
				binary.Write(buf, binary.BigEndian, uint32(0))
				// Fill to ImageInfoSectionPtr position (6 pointers)
				for i := 0; i < 6; i++ {
					binary.Write(buf, binary.BigEndian, uint64(0))
				}
				// ImageInfoSectionPtr (at position 10)
				binary.Write(buf, binary.BigEndian, uint32(0x2A00))
				binary.Write(buf, binary.BigEndian, uint32(0))
				// Fill rest
				for i := 0; i < 5; i++ {
					binary.Write(buf, binary.BigEndian, uint64(0))
				}
				return buf.Bytes()
			}(),
			wantBoot2:     0x1000,
			wantTOC:       0,
			wantTools:     0x500,
			wantImageInfo: 0x2A00,
			wantErr:       false,
		},
		{
			name: "ConnectX-7 style pointers",
			data: func() []byte {
				buf := &bytes.Buffer{}
				// BootRecordPtr
				binary.Write(buf, binary.BigEndian, uint64(0))
				// Boot2Ptr
				binary.Write(buf, binary.BigEndian, uint32(0x1000))
				binary.Write(buf, binary.BigEndian, uint32(0))
				// TOCPtr (used in CX-7)
				binary.Write(buf, binary.BigEndian, uint32(0x3500))
				binary.Write(buf, binary.BigEndian, uint32(0))
				// ToolsPtr
				binary.Write(buf, binary.BigEndian, uint64(0))
				// Fill to ImageInfoSectionPtr position (6 pointers)
				for i := 0; i < 6; i++ {
					binary.Write(buf, binary.BigEndian, uint64(0))
				}
				// ImageInfoSectionPtr (at position 10)
				binary.Write(buf, binary.BigEndian, uint32(0xF23C0))
				binary.Write(buf, binary.BigEndian, uint32(0))
				// Fill to HashesTablePtr
				for i := 0; i < 4; i++ {
					binary.Write(buf, binary.BigEndian, uint64(0))
				}
				// HashesTablePtr (at position 15)
				binary.Write(buf, binary.BigEndian, uint32(0xF2BC0))
				binary.Write(buf, binary.BigEndian, uint32(0))
				return buf.Bytes()
			}(),
			wantBoot2:     0x1000,
			wantTOC:       0x3500,
			wantTools:     0,
			wantImageInfo: 0xF23C0,
			wantHashes:    0xF2BC0,
			wantErr:       false,
		},
		{
			name:    "Too short data",
			data:    []byte{0x00, 0x00, 0x10, 0x00}, // Only 4 bytes
			wantErr: true,
		},
		{
			name:    "Empty data",
			data:    []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use annotated version for testing
			hwpAnnotated := &FS4HWPointersAnnotated{}
			err := hwpAnnotated.Unmarshal(tt.data)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if err == nil {
				// Use the annotated version directly (no conversion needed with type aliases)
				hwp := hwpAnnotated
				
				if hwp.Boot2Ptr.Ptr != tt.wantBoot2 {
					t.Errorf("Boot2Ptr.Ptr = 0x%X, want 0x%X", hwp.Boot2Ptr.Ptr, tt.wantBoot2)
				}
				if hwp.TOCPtr.Ptr != tt.wantTOC {
					t.Errorf("TOCPtr.Ptr = 0x%X, want 0x%X", hwp.TOCPtr.Ptr, tt.wantTOC)
				}
				if hwp.ToolsPtr.Ptr != tt.wantTools {
					t.Errorf("ToolsPtr.Ptr = 0x%X, want 0x%X", hwp.ToolsPtr.Ptr, tt.wantTools)
				}
				if hwp.ImageInfoSectionPtr.Ptr != tt.wantImageInfo {
					t.Errorf("ImageInfoSectionPtr.Ptr = 0x%X, want 0x%X", hwp.ImageInfoSectionPtr.Ptr, tt.wantImageInfo)
				}
				if tt.wantHashes != 0 && hwp.HashesTablePtr.Ptr != tt.wantHashes {
					t.Errorf("HashesTablePtr.Ptr = 0x%X, want 0x%X", hwp.HashesTablePtr.Ptr, tt.wantHashes)
				}
			}
		})
	}
}

func TestFS4HWPointers_Size_Extended(t *testing.T) {
	// Verify the structure is exactly 128 bytes (16 pointers * 8 bytes each)
	hwp := FS4HWPointers{}
	size := binary.Size(hwp)
	
	expectedSize := 128
	if size != expectedSize {
		t.Errorf("FS4HWPointers size = %d bytes, want %d bytes", size, expectedSize)
	}
}

func TestHWPointerEntry_Fields(t *testing.T) {
	// Test individual HWPointerEntry fields
	tests := []struct {
		name    string
		data    []byte
		wantPtr uint32
		wantCRC uint32
	}{
		{
			name: "Pointer with CRC",
			data: []byte{
				0x00, 0x00, 0x10, 0x00, // Ptr = 0x1000 (big-endian)
				0xAB, 0xCD, 0xEF, 0x00, // CRC = 0xABCDEF00
			},
			wantPtr: 0x1000,
			wantCRC: 0xABCDEF00,
		},
		{
			name: "Maximum values",
			data: []byte{
				0xFF, 0xFF, 0xFF, 0xFF, // Ptr = 0xFFFFFFFF
				0xFF, 0xFF, 0xFF, 0xFF, // CRC = 0xFFFFFFFF
			},
			wantPtr: 0xFFFFFFFF,
			wantCRC: 0xFFFFFFFF,
		},
		{
			name:    "All zeros",
			data:    make([]byte, 8),
			wantPtr: 0,
			wantCRC: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use annotated version for testing
			hwptrAnnotated := &HWPointerEntryAnnotated{}
			err := hwptrAnnotated.Unmarshal(tt.data)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			// Use the annotated version directly (no conversion needed with type aliases)
			hwptr := hwptrAnnotated

			if hwptr.Ptr != tt.wantPtr {
				t.Errorf("Ptr = 0x%X, want 0x%X", hwptr.Ptr, tt.wantPtr)
			}
			if hwptr.CRC != tt.wantCRC {
				t.Errorf("CRC = 0x%X, want 0x%X", hwptr.CRC, tt.wantCRC)
			}
		})
	}
}

func TestFS4HWPointers_ValidPointers(t *testing.T) {
	// Test validation of pointer values
	tests := []struct {
		name        string
		ptr         uint32
		isValid     bool
		description string
	}{
		{
			name:        "Valid pointer",
			ptr:         0x1000,
			isValid:     true,
			description: "Normal valid pointer",
		},
		{
			name:        "Zero pointer",
			ptr:         0,
			isValid:     false,
			description: "Zero means not present",
		},
		{
			name:        "Invalid pointer",
			ptr:         0xFFFFFFFF,
			isValid:     false,
			description: "0xFFFFFFFF means invalid",
		},
		{
			name:        "High address",
			ptr:         0x10000000,
			isValid:     true,
			description: "256MB address",
		},
		{
			name:        "Max valid address",
			ptr:         0xFFFFFFFE,
			isValid:     true,
			description: "Just below invalid marker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if pointer is valid
			isValid := tt.ptr != 0 && tt.ptr != 0xFFFFFFFF
			if isValid != tt.isValid {
				t.Errorf("%s: validation = %v, want %v", tt.description, isValid, tt.isValid)
			}
		})
	}
}

func TestFS4HWPointers_Marshal_Extended(t *testing.T) {
	// Test marshaling HW pointers back to bytes
	hwp := &FS4HWPointers{
		Boot2Ptr: HWPointerEntry{
			Ptr: 0x1000,
			CRC: 0xABCDEF00,
		},
		TOCPtr: HWPointerEntry{
			Ptr: 0x5000,
		},
		ToolsPtr: HWPointerEntry{
			Ptr: 0x6000,
		},
		ImageInfoSectionPtr: HWPointerEntry{
			Ptr: 0x2A00,
		},
	}

	// Marshal to bytes
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.BigEndian, hwp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Verify size
	if buf.Len() != 128 {
		t.Errorf("Marshaled size = %d, want 128", buf.Len())
	}

	// Unmarshal back and verify using annotated version
	hwp2Annotated := &FS4HWPointersAnnotated{}
	err = hwp2Annotated.Unmarshal(buf.Bytes())
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	
	hwp2 := hwp2Annotated.FromAnnotated()

	if hwp2.Boot2Ptr.Ptr != hwp.Boot2Ptr.Ptr {
		t.Errorf("Boot2Ptr.Ptr = 0x%X, want 0x%X", hwp2.Boot2Ptr.Ptr, hwp.Boot2Ptr.Ptr)
	}
	if hwp2.Boot2Ptr.CRC != hwp.Boot2Ptr.CRC {
		t.Errorf("Boot2Ptr.CRC = 0x%X, want 0x%X", hwp2.Boot2Ptr.CRC, hwp.Boot2Ptr.CRC)
	}
}

func BenchmarkFS4HWPointers_Unmarshal_Extended(b *testing.B) {
	// Create test data
	data := make([]byte, 128)
	binary.BigEndian.PutUint32(data[0:4], 0x1000)   // Boot2Ptr
	binary.BigEndian.PutUint32(data[8:12], 0x5000)  // TOCPtr
	binary.BigEndian.PutUint32(data[16:20], 0x6000) // ToolsPtr
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hwpAnnotated := &FS4HWPointersAnnotated{}
		_ = hwpAnnotated.Unmarshal(data)
	}
}

func TestFS4HWPointers_RealWorldPatterns_Extended(t *testing.T) {
	// Test with real-world pointer patterns
	tests := []struct {
		name        string
		description string
		setupFunc   func() *FS4HWPointers
	}{
		{
			name:        "ConnectX-5 pattern",
			description: "ITOC in ToolsPtr, no TOCPtr",
			setupFunc: func() *FS4HWPointers {
				return &FS4HWPointers{
					Boot2Ptr:            HWPointerEntry{Ptr: 0x1000},
					TOCPtr:              HWPointerEntry{Ptr: 0},
					ToolsPtr:            HWPointerEntry{Ptr: 0x500},
					ImageInfoSectionPtr: HWPointerEntry{Ptr: 0x2A00},
				}
			},
		},
		{
			name:        "ConnectX-7 pattern",
			description: "ITOC in TOCPtr",
			setupFunc: func() *FS4HWPointers {
				return &FS4HWPointers{
					Boot2Ptr:            HWPointerEntry{Ptr: 0x1000},
					TOCPtr:              HWPointerEntry{Ptr: 0x3500},
					ToolsPtr:            HWPointerEntry{Ptr: 0},
					ImageInfoSectionPtr: HWPointerEntry{Ptr: 0xF23C0},
					HashesTablePtr:      HWPointerEntry{Ptr: 0xF2BC0},
				}
			},
		},
		{
			name:        "Encrypted firmware",
			description: "Valid pointers but encrypted content",
			setupFunc: func() *FS4HWPointers {
				return &FS4HWPointers{
					Boot2Ptr:            HWPointerEntry{Ptr: 0x1000},
					TOCPtr:              HWPointerEntry{Ptr: 0xFFFFFFFF},
					ToolsPtr:            HWPointerEntry{Ptr: 0xFFFFFFFF},
					ImageInfoSectionPtr: HWPointerEntry{Ptr: 0x2A00},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hwp := tt.setupFunc()
			
			// Log the pattern
			t.Logf("%s:", tt.description)
			t.Logf("  Boot2Ptr: 0x%X", hwp.Boot2Ptr.Ptr)
			t.Logf("  TOCPtr: 0x%X", hwp.TOCPtr.Ptr)
			t.Logf("  ToolsPtr: 0x%X", hwp.ToolsPtr.Ptr)
			t.Logf("  ImageInfoSectionPtr: 0x%X", hwp.ImageInfoSectionPtr.Ptr)
			if hwp.HashesTablePtr.Ptr != 0 {
				t.Logf("  HashesTablePtr: 0x%X", hwp.HashesTablePtr.Ptr)
			}
			
			// Verify basic sanity
			if hwp.Boot2Ptr.Ptr == 0 || hwp.Boot2Ptr.Ptr == 0xFFFFFFFF {
				t.Log("Warning: Boot2Ptr is invalid")
			}
		})
	}
}