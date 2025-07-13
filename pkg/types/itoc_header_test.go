package types

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/ghostiam/binstruct"
)

func TestITOCHeader_Unmarshal(t *testing.T) {
	tests := []struct {
		name          string
		data          []byte
		wantSignature uint32
		wantCRC       uint32
		wantErr       bool
	}{
		{
			name: "Valid ITOC header",
			data: func() []byte {
				buf := &bytes.Buffer{}
				binary.Write(buf, binary.BigEndian, uint32(ITOCSignature)) // Signature0
				binary.Write(buf, binary.BigEndian, uint32(0))             // Signature1
				binary.Write(buf, binary.BigEndian, uint32(0))             // Signature2
				binary.Write(buf, binary.BigEndian, uint32(0))             // Signature3
				binary.Write(buf, binary.BigEndian, uint32(2))             // Version
				binary.Write(buf, binary.BigEndian, uint32(0))             // Reserved
				binary.Write(buf, binary.BigEndian, uint32(0))             // ITOCEntryCRC
				binary.Write(buf, binary.BigEndian, uint32(0x1234))        // CRC
				return buf.Bytes()
			}(),
			wantSignature: ITOCSignature,
			wantCRC:       0x1234,
			wantErr:       false,
		},
		{
			name: "Valid DTOC header",
			data: func() []byte {
				buf := &bytes.Buffer{}
				binary.Write(buf, binary.BigEndian, uint32(DTOCSignature)) // Signature0
				binary.Write(buf, binary.BigEndian, uint32(0))             // Signature1
				binary.Write(buf, binary.BigEndian, uint32(0))             // Signature2
				binary.Write(buf, binary.BigEndian, uint32(0))             // Signature3
				binary.Write(buf, binary.BigEndian, uint32(1))             // Version
				binary.Write(buf, binary.BigEndian, uint32(0))             // Reserved
				binary.Write(buf, binary.BigEndian, uint32(0))             // ITOCEntryCRC
				binary.Write(buf, binary.BigEndian, uint32(0x5678))        // CRC
				return buf.Bytes()
			}(),
			wantSignature: DTOCSignature,
			wantCRC:       0x5678,
			wantErr:       false,
		},
		{
			name: "Invalid signature",
			data: func() []byte {
				buf := &bytes.Buffer{}
				binary.Write(buf, binary.BigEndian, uint32(0xDEADBEEF)) // Invalid signature
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(0))
				return buf.Bytes()
			}(),
			wantSignature: 0xDEADBEEF,
			wantCRC:       0,
			wantErr:       false,
		},
		{
			name:          "Too short data",
			data:          []byte{0x49, 0x54, 0x4F, 0x43}, // Just "ITOC"
			wantSignature: 0,
			wantCRC:       0,
			wantErr:       true,
		},
		{
			name:          "Empty data",
			data:          []byte{},
			wantSignature: 0,
			wantCRC:       0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := &ITOCHeader{}
			err := binstruct.UnmarshalBE(tt.data, header)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalBE() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if err == nil {
				if header.Signature0 != tt.wantSignature {
					t.Errorf("Signature0 = 0x%08X, want 0x%08X", header.Signature0, tt.wantSignature)
				}
				if header.CRC != tt.wantCRC {
					t.Errorf("CRC = 0x%04X, want 0x%04X", header.CRC, tt.wantCRC)
				}
			}
		})
	}
}

func TestITOCHeader_Size(t *testing.T) {
	// Verify the structure is exactly 32 bytes
	header := ITOCHeader{}
	size := binary.Size(header)
	
	expectedSize := 32
	if size != expectedSize {
		t.Errorf("ITOCHeader size = %d bytes, want %d bytes", size, expectedSize)
	}
}

func TestITOCHeader_SignatureConstants(t *testing.T) {
	// Test that signature constants are correct
	tests := []struct {
		name      string
		signature uint32
		expected  string
	}{
		{
			name:      "ITOC signature",
			signature: ITOCSignature,
			expected:  "ITOC",
		},
		{
			name:      "DTOC signature",
			signature: DTOCSignature,
			expected:  "DTOC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert signature to bytes and check ASCII
			bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(bytes, tt.signature)
			
			got := string(bytes)
			if got != tt.expected {
				t.Errorf("Signature as string = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestITOCHeader_Marshal(t *testing.T) {
	// Test marshaling back to bytes
	header := ITOCHeader{
		Signature0:   ITOCSignature,
		Signature1:   0,
		Signature2:   0,
		Signature3:   0,
		Version:      2,
		Reserved:     0,
		ITOCEntryCRC: 0x1234,
		CRC:          0xABCD,
	}

	// Marshal to bytes
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.BigEndian, header)
	if err != nil {
		t.Fatalf("Failed to marshal header: %v", err)
	}

	// Verify size
	if buf.Len() != 32 {
		t.Errorf("Marshaled size = %d, want 32", buf.Len())
	}

	// Unmarshal back and verify
	var header2 ITOCHeader
	err = binstruct.UnmarshalBE(buf.Bytes(), &header2)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if header2.Signature0 != header.Signature0 {
		t.Errorf("Signature0 = 0x%08X, want 0x%08X", header2.Signature0, header.Signature0)
	}
	if header2.CRC != header.CRC {
		t.Errorf("CRC = 0x%04X, want 0x%04X", header2.CRC, header.CRC)
	}
}

func TestITOCHeader_RealWorldData(t *testing.T) {
	// Test with patterns that might appear in real firmware
	tests := []struct {
		name        string
		description string
		data        []byte
	}{
		{
			name:        "ITOC with CRC",
			description: "ITOC header with valid CRC",
			data: func() []byte {
				buf := &bytes.Buffer{}
				binary.Write(buf, binary.BigEndian, uint32(0x49544F43)) // "ITOC"
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(2))
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(0x1234))
				return buf.Bytes()
			}(),
		},
		{
			name:        "DTOC with data",
			description: "DTOC header with some reserved fields set",
			data: func() []byte {
				buf := &bytes.Buffer{}
				binary.Write(buf, binary.BigEndian, uint32(0x44544F43)) // "DTOC"
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(1))
				binary.Write(buf, binary.BigEndian, uint32(0x12345678))
				binary.Write(buf, binary.BigEndian, uint32(0xABCDEF01))
				binary.Write(buf, binary.BigEndian, uint32(0x5678))
				return buf.Bytes()
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := &ITOCHeader{}
			err := binstruct.UnmarshalBE(tt.data, header)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			// Log the parsed header
			t.Logf("%s: Sig0=0x%08X, CRC=0x%08X, Version=%d",
				tt.description, header.Signature0, header.CRC, header.Version)
			
			// Basic sanity check
			if header.Signature0 != ITOCSignature && header.Signature0 != DTOCSignature {
				t.Logf("Warning: Unexpected signature 0x%08X", header.Signature0)
			}
		})
	}
}

func BenchmarkITOCHeader_Unmarshal(b *testing.B) {
	// Benchmark unmarshaling performance
	data := make([]byte, 32)
	binary.BigEndian.PutUint32(data[0:4], ITOCSignature)
	binary.BigEndian.PutUint32(data[16:20], 2)      // Version
	binary.BigEndian.PutUint32(data[28:32], 0x1234) // CRC
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		header := &ITOCHeader{}
		_ = binstruct.UnmarshalBE(data, header)
	}
}

func TestITOCHeader_Validation(t *testing.T) {
	// Test header validation scenarios
	tests := []struct {
		name     string
		header   ITOCHeader
		isValid  bool
		tocType  string
	}{
		{
			name: "Valid ITOC",
			header: ITOCHeader{
				Signature0: ITOCSignature,
				CRC:        0x1234,
			},
			isValid: true,
			tocType: "ITOC",
		},
		{
			name: "Valid DTOC",
			header: ITOCHeader{
				Signature0: DTOCSignature,
				CRC:        0x5678,
			},
			isValid: true,
			tocType: "DTOC",
		},
		{
			name: "Invalid signature",
			header: ITOCHeader{
				Signature0: 0xBADC0DE,
			},
			isValid: false,
			tocType: "UNKNOWN",
		},
		{
			name: "Zero signature",
			header: ITOCHeader{
				Signature0: 0,
			},
			isValid: false,
			tocType: "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if signature is valid
			isValid := tt.header.Signature0 == ITOCSignature || tt.header.Signature0 == DTOCSignature
			if isValid != tt.isValid {
				t.Errorf("Validation = %v, want %v", isValid, tt.isValid)
			}

			// Determine TOC type
			var tocType string
			switch tt.header.Signature0 {
			case ITOCSignature:
				tocType = "ITOC"
			case DTOCSignature:
				tocType = "DTOC"
			default:
				tocType = "UNKNOWN"
			}
			
			if tocType != tt.tocType {
				t.Errorf("TOC type = %v, want %v", tocType, tt.tocType)
			}
		})
	}
}