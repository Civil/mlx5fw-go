package types

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestBigEndian_Uint32(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected uint32
	}{
		{
			name:     "Big endian value",
			input:    []byte{0x12, 0x34, 0x56, 0x78},
			expected: 0x12345678,
		},
		{
			name:     "All zeros",
			input:    []byte{0x00, 0x00, 0x00, 0x00},
			expected: 0,
		},
		{
			name:     "All ones",
			input:    []byte{0xFF, 0xFF, 0xFF, 0xFF},
			expected: 0xFFFFFFFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test using binary.BigEndian directly (since that's what we use)
			got := binary.BigEndian.Uint32(tt.input)
			if got != tt.expected {
				t.Errorf("binary.BigEndian.Uint32() = 0x%x, want 0x%x", got, tt.expected)
			}
		})
	}
}

func TestBigEndian_Uint16(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected uint16
	}{
		{
			name:     "Big endian value",
			input:    []byte{0x12, 0x34},
			expected: 0x1234,
		},
		{
			name:     "All zeros",
			input:    []byte{0x00, 0x00},
			expected: 0,
		},
		{
			name:     "All ones",
			input:    []byte{0xFF, 0xFF},
			expected: 0xFFFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test using binary.BigEndian directly
			got := binary.BigEndian.Uint16(tt.input)
			if got != tt.expected {
				t.Errorf("binary.BigEndian.Uint16() = 0x%x, want 0x%x", got, tt.expected)
			}
		})
	}
}

func TestBigEndian_PutUint32(t *testing.T) {
	tests := []struct {
		name     string
		value    uint32
		expected []byte
	}{
		{
			name:     "Big endian value",
			value:    0x12345678,
			expected: []byte{0x12, 0x34, 0x56, 0x78},
		},
		{
			name:     "All zeros",
			value:    0,
			expected: []byte{0x00, 0x00, 0x00, 0x00},
		},
		{
			name:     "All ones",
			value:    0xFFFFFFFF,
			expected: []byte{0xFF, 0xFF, 0xFF, 0xFF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 4)
			binary.BigEndian.PutUint32(buf, tt.value)

			if !bytes.Equal(buf, tt.expected) {
				t.Errorf("binary.BigEndian.PutUint32() = %v, want %v", buf, tt.expected)
			}
		})
	}
}

func TestBigEndian_PutUint16(t *testing.T) {
	tests := []struct {
		name     string
		value    uint16
		expected []byte
	}{
		{
			name:     "Big endian value",
			value:    0x1234,
			expected: []byte{0x12, 0x34},
		},
		{
			name:     "All zeros",
			value:    0,
			expected: []byte{0x00, 0x00},
		},
		{
			name:     "All ones",
			value:    0xFFFF,
			expected: []byte{0xFF, 0xFF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 2)
			binary.BigEndian.PutUint16(buf, tt.value)

			if !bytes.Equal(buf, tt.expected) {
				t.Errorf("binary.BigEndian.PutUint16() = %v, want %v", buf, tt.expected)
			}
		})
	}
}

func TestBinaryWrite(t *testing.T) {
	type TestStruct struct {
		A uint32
		B uint16
		C uint8
		D uint8
	}

	expected := []byte{
		0x12, 0x34, 0x56, 0x78, // A
		0xAB, 0xCD, // B
		0xEF, // C
		0x01, // D
	}

	s := TestStruct{
		A: 0x12345678,
		B: 0xABCD,
		C: 0xEF,
		D: 0x01,
	}

	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.BigEndian, s)
	if err != nil {
		t.Fatalf("binary.Write() error = %v", err)
	}

	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("binary.Write() = %v, want %v", buf.Bytes(), expected)
	}
}

func TestBinaryRead(t *testing.T) {
	type TestStruct struct {
		A uint32
		B uint16
		C uint8
		D uint8
	}

	input := []byte{
		0x12, 0x34, 0x56, 0x78, // A
		0xAB, 0xCD, // B
		0xEF, // C
		0x01, // D
	}

	var s TestStruct
	buf := bytes.NewReader(input)
	err := binary.Read(buf, binary.BigEndian, &s)
	if err != nil {
		t.Fatalf("binary.Read() error = %v", err)
	}

	if s.A != 0x12345678 {
		t.Errorf("s.A = 0x%x, want 0x12345678", s.A)
	}
	if s.B != 0xABCD {
		t.Errorf("s.B = 0x%x, want 0xABCD", s.B)
	}
	if s.C != 0xEF {
		t.Errorf("s.C = 0x%x, want 0xEF", s.C)
	}
	if s.D != 0x01 {
		t.Errorf("s.D = 0x%x, want 0x01", s.D)
	}
}
