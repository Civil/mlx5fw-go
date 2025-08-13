package fs4

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/types"
)

func TestVerifyITOCHeaderCRC(t *testing.T) {
	crcCalculator := parser.NewCRCCalculator()

	tests := []struct {
		name        string
		header      []byte
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid ITOC header with correct CRC",
			header: func() []byte {
				h := make([]byte, 32)
				binary.BigEndian.PutUint32(h[0:4], types.ITOCSignature)
				binary.BigEndian.PutUint32(h[16:20], 2) // Version
				// Calculate and set CRC
				UpdateITOCHeaderCRC(h, crcCalculator)
				return h
			}(),
			expectError: false,
		},
		{
			name: "Valid DTOC header with correct CRC",
			header: func() []byte {
				h := make([]byte, 32)
				binary.BigEndian.PutUint32(h[0:4], types.DTOCSignature)
				binary.BigEndian.PutUint32(h[16:20], 1) // Version
				// Calculate and set CRC
				UpdateITOCHeaderCRC(h, crcCalculator)
				return h
			}(),
			expectError: false,
		},
		{
			name: "Header with wrong CRC",
			header: func() []byte {
				h := make([]byte, 32)
				binary.BigEndian.PutUint32(h[0:4], types.ITOCSignature)
				binary.BigEndian.PutUint32(h[28:32], 0x1234) // Wrong CRC
				return h
			}(),
			expectError: true,
			errorMsg:    "CRC mismatch",
		},
		{
			name: "Header with blank CRC (0xFFFF)",
			header: func() []byte {
				h := make([]byte, 32)
				binary.BigEndian.PutUint32(h[0:4], types.ITOCSignature)
				binary.BigEndian.PutUint32(h[28:32], 0xFFFF) // Blank CRC
				return h
			}(),
			expectError: true,
			errorMsg:    "blank ITOC header CRC",
		},
		{
			name:        "Header too small",
			header:      make([]byte, 20),
			expectError: true,
			errorMsg:    "too small",
		},
		{
			name: "Real-world ITOC header pattern",
			header: func() []byte {
				h := make([]byte, 32)
				binary.BigEndian.PutUint32(h[0:4], 0x49544F43) // "ITOC"
				binary.BigEndian.PutUint32(h[4:8], 0)          // Signature1
				binary.BigEndian.PutUint32(h[8:12], 0)         // Signature2
				binary.BigEndian.PutUint32(h[12:16], 0)        // Signature3
				binary.BigEndian.PutUint32(h[16:20], 2)        // Version
				binary.BigEndian.PutUint32(h[20:24], 0)        // Reserved
				binary.BigEndian.PutUint32(h[24:28], 0x1234)   // ITOCEntryCRC
				// Calculate correct CRC
				crc := CalculateITOCHeaderCRC(h, crcCalculator)
				binary.BigEndian.PutUint32(h[28:32], uint32(crc))
				return h
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyITOCHeaderCRC(tt.header, crcCalculator)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else if tt.expectError && tt.errorMsg != "" && !bytes.Contains([]byte(err.Error()), []byte(tt.errorMsg)) {
				t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
			}
		})
	}
}

func TestCalculateITOCHeaderCRC(t *testing.T) {
	crcCalculator := parser.NewCRCCalculator()

	// Test with known header
	header := make([]byte, 32)
	binary.BigEndian.PutUint32(header[0:4], types.ITOCSignature)
	binary.BigEndian.PutUint32(header[16:20], 2) // Version

	crc := CalculateITOCHeaderCRC(header, crcCalculator)

	// The CRC should be consistent
	expectedCRC := uint16(0x0876) // From our demo output
	if crc != expectedCRC {
		t.Errorf("CRC mismatch: got 0x%04X, expected 0x%04X", crc, expectedCRC)
	}

	// Test that changing header data changes CRC
	header[16] = 3 // Change version
	crc2 := CalculateITOCHeaderCRC(header, crcCalculator)
	if crc == crc2 {
		t.Error("CRC should change when header data changes")
	}
}

func TestUpdateITOCHeaderCRC(t *testing.T) {
	crcCalculator := parser.NewCRCCalculator()

	// Create header with invalid CRC
	header := make([]byte, 32)
	binary.BigEndian.PutUint32(header[0:4], types.ITOCSignature)
	binary.BigEndian.PutUint32(header[16:20], 2)          // Version
	binary.BigEndian.PutUint32(header[28:32], 0xDEADBEEF) // Wrong CRC

	// Update CRC
	UpdateITOCHeaderCRC(header, crcCalculator)

	// Verify it's now correct
	err := VerifyITOCHeaderCRC(header, crcCalculator)
	if err != nil {
		t.Errorf("CRC verification failed after update: %v", err)
	}

	// Check the actual CRC value
	storedCRC := binary.BigEndian.Uint32(header[28:32])
	if storedCRC == 0xDEADBEEF {
		t.Error("CRC was not updated")
	}
	if storedCRC > 0xFFFF {
		t.Errorf("CRC should be 16-bit value, got 0x%08X", storedCRC)
	}
}

// Benchmark CRC calculation
func BenchmarkCalculateITOCHeaderCRC(b *testing.B) {
	crcCalculator := parser.NewCRCCalculator()
	header := make([]byte, 32)
	binary.BigEndian.PutUint32(header[0:4], types.ITOCSignature)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateITOCHeaderCRC(header, crcCalculator)
	}
}
