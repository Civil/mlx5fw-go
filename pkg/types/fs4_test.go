package types

import (
	"encoding/hex"
	"testing"
)

func TestFS4HWPointers(t *testing.T) {
	// Test data representing HW pointers structure
	// This simulates the structure at offset 0x18 in an FS4 firmware
	// Each pointer is 8 bytes (4 bytes Ptr + 4 bytes CRC)
	hexData := "0000000000000000" + // BootRecordPtr
		"0000100000000000" + // Boot2Ptr (0x1000)
		"0000100000000000" + // TOCPtr (0x1000)
		"0000500000000000" + // ToolsPtr (0x5000)
		"ffffffff00000000" + // AuthenticationStartPtr (0xffffffff)
		"0000000000000000" + // AuthenticationEndPtr
		"0000000000000000" + // DigestPtr
		"0000000000000000" + // DigestRecoveryKeyPtr
		"0000000000000000" + // FWWindowStartPtr
		"ffffffff00000000" + // FWWindowEndPtr (0xffffffff)
		"0000000000000000" + // ImageInfoSectionPtr
		"ffffffff00000000" + // ImageSignaturePtr (0xffffffff)
		"ffffffff00000000" + // PublicKeyPtr (0xffffffff)
		"ffffffff00000000" + // FWSecurityVersionPtr (0xffffffff)
		"ffffffff00000000" + // GCMIVDeltaPtr (0xffffffff)
		"ffffffff00000000" // HashesTablePtr (0xffffffff)
	
	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatalf("Failed to decode hex string: %v", err)
	}

	// Use annotated version for testing
	hwPointersAnnotated := &FS4HWPointersAnnotated{}
	err = hwPointersAnnotated.Unmarshal(data[:128])
	if err != nil {
		t.Fatalf("Failed to unmarshal HW pointers: %v", err)
	}
	
	// Convert to legacy format for comparison
	hwPointers := hwPointersAnnotated.FromAnnotated()

	// Test boot record pointer
	if hwPointers.BootRecordPtr.Ptr != 0 {
		t.Errorf("BootRecordPtr.Ptr = 0x%x, want 0x0", hwPointers.BootRecordPtr.Ptr)
	}

	// Test boot2 pointer
	if hwPointers.Boot2Ptr.Ptr != 0x1000 {
		t.Errorf("Boot2Ptr.Ptr = 0x%x, want 0x1000", hwPointers.Boot2Ptr.Ptr)
	}

	// Test TOC pointer
	if hwPointers.TOCPtr.Ptr != 0x1000 {
		t.Errorf("TOCPtr.Ptr = 0x%x, want 0x1000", hwPointers.TOCPtr.Ptr)
	}

	// Test tools pointer
	if hwPointers.ToolsPtr.Ptr != 0x5000 {
		t.Errorf("ToolsPtr.Ptr = 0x%x, want 0x5000", hwPointers.ToolsPtr.Ptr)
	}

	// Test authentication start pointer (should be 0xffffffff for invalid)
	if hwPointers.AuthenticationStartPtr.Ptr != 0xffffffff {
		t.Errorf("AuthenticationStartPtr.Ptr = 0x%x, want 0xffffffff", hwPointers.AuthenticationStartPtr.Ptr)
	}
}

func TestITOCHeader(t *testing.T) {
	// Test ITOC header structure
	// Signature0: 0x49544F43 (ITOC signature)
	// Signature1-3: 0x00000000
	// Version: 2
	// Reserved: 0x00
	// ITOCEntryCRC: 0x1234
	// CRC: 0x5678
	hexData := "49544F43000000000000000000000000" +
		"00000002000000000000123400005678"

	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatalf("Failed to decode hex string: %v", err)
	}

	// Use annotated version for testing
	headerAnnotated := &ITOCHeaderAnnotated{}
	err = headerAnnotated.Unmarshal(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal ITOC header: %v", err)
	}
	
	// Convert to legacy format for comparison
	header := headerAnnotated.FromAnnotated()

	// Test signature
	if header.Signature0 != ITOCSignature {
		t.Errorf("ITOCHeader.Signature0 = 0x%x, want 0x%x", header.Signature0, ITOCSignature)
	}

	// Test version
	if header.Version != 2 {
		t.Errorf("ITOCHeader.Version = %d, want 2", header.Version)
	}
	
	// Test CRC fields
	if header.ITOCEntryCRC != 0x1234 {
		t.Errorf("ITOCHeader.ITOCEntryCRC = 0x%x, want 0x1234", header.ITOCEntryCRC)
	}
	
	if header.CRC != 0x5678 {
		t.Errorf("ITOCHeader.CRC = 0x%x, want 0x5678", header.CRC)
	}
}

func TestITOCEntry_ParseFields(t *testing.T) {
	tests := []struct {
		name           string
		hexData        string
		expectedType   uint8
		expectedSize   uint32
		expectedAddr   uint32
		expectedCRC    uint16
		expectedNoCRC  bool
	}{
		{
			name:           "BOOT3_CODE entry",
			hexData:        "0f002590205000402050004400000000000000000000a00000005a5200002dda",
			expectedType:   0x0f,
			expectedSize:   0x2590,
			expectedAddr:   0xa000,
			expectedCRC:    0x5a52,
			expectedNoCRC:  false,
		},
		{
			name:           "Empty entry",
			hexData:        "ff000000000000000000000000000000000000000000000000000000000000ff",
			expectedType:   0xff,
			expectedSize:   0,
			expectedAddr:   0,
			expectedCRC:    0,
			expectedNoCRC:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := hex.DecodeString(tt.hexData)
			if err != nil {
				t.Fatalf("Failed to decode hex string: %v", err)
			}

			// Use annotated version for testing
			entryAnnotated := &ITOCEntryAnnotated{}
			err = entryAnnotated.Unmarshal(data)
			if err != nil {
				t.Fatalf("Failed to unmarshal ITOCEntry: %v", err)
			}
			
			// Convert to legacy format for comparison
			entry := entryAnnotated.FromAnnotated()

			if entry.Type != tt.expectedType {
				t.Errorf("ITOCEntry.Type = 0x%x, want 0x%x", entry.Type, tt.expectedType)
			}

			if entry.Size != tt.expectedSize {
				t.Errorf("ITOCEntry.Size = 0x%x, want 0x%x", entry.Size, tt.expectedSize)
			}

			if entry.FlashAddr != tt.expectedAddr {
				t.Errorf("ITOCEntry.FlashAddr = 0x%x, want 0x%x", entry.FlashAddr, tt.expectedAddr)
			}

			if entry.SectionCRC != tt.expectedCRC {
				t.Errorf("ITOCEntry.SectionCRC = 0x%x, want 0x%x", entry.SectionCRC, tt.expectedCRC)
			}

			if entry.GetNoCRC() != tt.expectedNoCRC {
				t.Errorf("ITOCEntry.GetNoCRC() = %v, want %v", entry.GetNoCRC(), tt.expectedNoCRC)
			}
		})
	}
}

// TestITOCEntry_GetBits removed - testing private function