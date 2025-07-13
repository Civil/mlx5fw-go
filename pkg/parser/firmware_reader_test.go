package parser

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/Civil/mlx5fw-go/pkg/types"
)

// createTestFile creates a temporary file with the given content
func createTestFile(t *testing.T, content []byte) string {
	tmpfile, err := os.CreateTemp("", "firmware_test_*.bin")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	if _, err := tmpfile.Write(content); err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	
	if err := tmpfile.Close(); err != nil {
		os.Remove(tmpfile.Name())
		t.Fatalf("Failed to close temp file: %v", err)
	}
	
	return tmpfile.Name()
}

func TestNewFirmwareReader(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	t.Run("Valid file", func(t *testing.T) {
		content := []byte("test firmware content")
		filename := createTestFile(t, content)
		defer os.Remove(filename)
		
		reader, err := NewFirmwareReader(filename, logger)
		if err != nil {
			t.Fatalf("NewFirmwareReader() error = %v", err)
		}
		defer reader.Close()
		
		if reader.Size() != int64(len(content)) {
			t.Errorf("Size() = %v, want %v", reader.Size(), len(content))
		}
	})
	
	t.Run("Non-existent file", func(t *testing.T) {
		reader, err := NewFirmwareReader("/non/existent/file.bin", logger)
		if err == nil {
			reader.Close()
			t.Fatal("NewFirmwareReader() expected error for non-existent file")
		}
	})
}

func TestFirmwareReader_ReadAt(t *testing.T) {
	logger := zaptest.NewLogger(t)
	content := []byte("0123456789ABCDEF")
	filename := createTestFile(t, content)
	defer os.Remove(filename)
	
	reader, err := NewFirmwareReader(filename, logger)
	if err != nil {
		t.Fatalf("NewFirmwareReader() error = %v", err)
	}
	defer reader.Close()
	
	tests := []struct {
		name      string
		offset    int64
		length    int
		expected  []byte
		expectErr bool
	}{
		{
			name:     "Read from beginning",
			offset:   0,
			length:   4,
			expected: []byte("0123"),
		},
		{
			name:     "Read from middle",
			offset:   5,
			length:   5,
			expected: []byte("56789"),
		},
		{
			name:     "Read from end",
			offset:   12,
			length:   4,
			expected: []byte("CDEF"),
		},
		{
			name:      "Read beyond EOF",
			offset:    20,
			length:    4,
			expectErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, tt.length)
			n, err := reader.ReadAt(buf, tt.offset)
			
			if tt.expectErr {
				if err == nil {
					t.Error("ReadAt() expected error but got none")
				}
			} else {
				if err != nil && err != io.EOF {
					t.Errorf("ReadAt() unexpected error = %v", err)
				}
				if n != len(tt.expected) {
					t.Errorf("ReadAt() read %v bytes, want %v", n, len(tt.expected))
				}
				if !bytes.Equal(buf[:n], tt.expected) {
					t.Errorf("ReadAt() = %v, want %v", buf[:n], tt.expected)
				}
			}
		})
	}
}

func TestFirmwareReader_ReadUint32BE(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	// Create test data with known uint32 values
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint32(0x12345678))
	binary.Write(&buf, binary.BigEndian, uint32(0xABCDEF00))
	
	filename := createTestFile(t, buf.Bytes())
	defer os.Remove(filename)
	
	reader, err := NewFirmwareReader(filename, logger)
	if err != nil {
		t.Fatalf("NewFirmwareReader() error = %v", err)
	}
	defer reader.Close()
	
	tests := []struct {
		name      string
		offset    int64
		expected  uint32
		expectErr bool
	}{
		{
			name:     "Read first uint32",
			offset:   0,
			expected: 0x12345678,
		},
		{
			name:     "Read second uint32",
			offset:   4,
			expected: 0xABCDEF00,
		},
		{
			name:      "Read beyond EOF",
			offset:    8,
			expectErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := reader.ReadUint32BE(tt.offset)
			
			if tt.expectErr {
				if err == nil {
					t.Error("ReadUint32BE() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ReadUint32BE() error = %v", err)
				}
				if got != tt.expected {
					t.Errorf("ReadUint32BE() = 0x%08X, want 0x%08X", got, tt.expected)
				}
			}
		})
	}
}

func TestFirmwareReader_ReadUint64BE(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	// Create test data with known uint64 values
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint64(0x123456789ABCDEF0))
	binary.Write(&buf, binary.BigEndian, uint64(0xFEDCBA9876543210))
	
	filename := createTestFile(t, buf.Bytes())
	defer os.Remove(filename)
	
	reader, err := NewFirmwareReader(filename, logger)
	if err != nil {
		t.Fatalf("NewFirmwareReader() error = %v", err)
	}
	defer reader.Close()
	
	tests := []struct {
		name      string
		offset    int64
		expected  uint64
		expectErr bool
	}{
		{
			name:     "Read first uint64",
			offset:   0,
			expected: 0x123456789ABCDEF0,
		},
		{
			name:     "Read second uint64",
			offset:   8,
			expected: 0xFEDCBA9876543210,
		},
		{
			name:      "Read beyond EOF",
			offset:    16,
			expectErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := reader.ReadUint64BE(tt.offset)
			
			if tt.expectErr {
				if err == nil {
					t.Error("ReadUint64BE() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ReadUint64BE() error = %v", err)
				}
				if got != tt.expected {
					t.Errorf("ReadUint64BE() = 0x%016X, want 0x%016X", got, tt.expected)
				}
			}
		})
	}
}

func TestFirmwareReader_FindMagicPattern(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	tests := []struct {
		name           string
		createContent  func() []byte
		expectedOffset uint32
		expectErr      bool
	}{
		{
			name: "Magic at offset 0",
			createContent: func() []byte {
				buf := make([]byte, 108)
				binary.BigEndian.PutUint64(buf[0:], types.MagicPattern)
				return buf
			},
			expectedOffset: 0,
		},
		{
			name: "Magic at offset 0x10000",
			createContent: func() []byte {
				buf := make([]byte, 0x11000)
				binary.BigEndian.PutUint64(buf[0x10000:], types.MagicPattern)
				return buf
			},
			expectedOffset: 0x10000,
		},
		{
			name: "No magic pattern",
			createContent: func() []byte {
				return make([]byte, 0x2000)
			},
			expectErr: true,
		},
		{
			name: "Magic at multiple locations",
			createContent: func() []byte {
				buf := make([]byte, 0x30000)
				// Put magic at two locations
				binary.BigEndian.PutUint64(buf[0x10000:], types.MagicPattern)
				binary.BigEndian.PutUint64(buf[0x20000:], types.MagicPattern)
				return buf
			},
			expectedOffset: 0x10000, // Should find first occurrence
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := tt.createContent()
			filename := createTestFile(t, content)
			defer os.Remove(filename)
			
			reader, err := NewFirmwareReader(filename, logger)
			if err != nil {
				t.Fatalf("NewFirmwareReader() error = %v", err)
			}
			defer reader.Close()
			
			offset, err := reader.FindMagicPattern()
			
			if tt.expectErr {
				if err == nil {
					t.Error("FindMagicPattern() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("FindMagicPattern() error = %v", err)
				}
				if offset != tt.expectedOffset {
					t.Errorf("FindMagicPattern() = 0x%X, want 0x%X", offset, tt.expectedOffset)
				}
			}
		})
	}
}

func TestFirmwareReader_ReadSection(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	content := make([]byte, 0x1000)
	for i := range content {
		content[i] = byte(i & 0xFF)
	}
	
	filename := createTestFile(t, content)
	defer os.Remove(filename)
	
	reader, err := NewFirmwareReader(filename, logger)
	if err != nil {
		t.Fatalf("NewFirmwareReader() error = %v", err)
	}
	defer reader.Close()
	
	tests := []struct {
		name      string
		offset    int64
		size      uint32
		expectErr bool
		validate  func(data []byte) bool
	}{
		{
			name:   "Read valid section",
			offset: 0x100,
			size:   0x200,
			validate: func(data []byte) bool {
				if len(data) != 0x200 {
					return false
				}
				// Check first few bytes
				return data[0] == 0x00 && data[1] == 0x01 && data[2] == 0x02
			},
		},
		{
			name:      "Negative offset",
			offset:    -1,
			size:      0x100,
			expectErr: true,
		},
		{
			name:      "Offset beyond file",
			offset:    0x2000,
			size:      0x100,
			expectErr: true,
		},
		{
			name:      "Size extends beyond file",
			offset:    0x800,
			size:      0x900,
			expectErr: true,
		},
		{
			name:   "Zero size",
			offset: 0x100,
			size:   0,
			validate: func(data []byte) bool {
				return len(data) == 0
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := reader.ReadSection(tt.offset, tt.size)
			
			if tt.expectErr {
				if err == nil {
					t.Error("ReadSection() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ReadSection() error = %v", err)
				}
				if tt.validate != nil && !tt.validate(data) {
					t.Error("ReadSection() returned invalid data")
				}
			}
		})
	}
}

func TestFirmwareReader_Close(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	filename := createTestFile(t, []byte("test"))
	defer os.Remove(filename)
	
	reader, err := NewFirmwareReader(filename, logger)
	if err != nil {
		t.Fatalf("NewFirmwareReader() error = %v", err)
	}
	
	// Close should work
	if err := reader.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
	
	// Close again should be safe
	if err := reader.Close(); err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}

func TestFirmwareReader_ReadHWPointers(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	// Create test data with HW pointers
	hwPointers := make([]byte, 128)
	for i := 0; i < len(hwPointers); i += 8 {
		binary.BigEndian.PutUint32(hwPointers[i:], uint32(0x1000+i))
		binary.BigEndian.PutUint32(hwPointers[i+4:], 0)
	}
	
	filename := createTestFile(t, hwPointers)
	defer os.Remove(filename)
	
	reader, err := NewFirmwareReader(filename, logger)
	if err != nil {
		t.Fatalf("NewFirmwareReader() error = %v", err)
	}
	defer reader.Close()
	
	data, err := reader.ReadHWPointers(0, 128)
	if err != nil {
		t.Errorf("ReadHWPointers() error = %v", err)
	}
	
	if len(data) != 128 {
		t.Errorf("ReadHWPointers() returned %d bytes, want 128", len(data))
	}
	
	// Verify first pointer
	firstPtr := binary.BigEndian.Uint32(data[0:4])
	if firstPtr != 0x1000 {
		t.Errorf("First HW pointer = 0x%X, want 0x1000", firstPtr)
	}
}

func BenchmarkFirmwareReader_ReadSection(b *testing.B) {
	logger := zap.NewNop()
	
	// Create a 1MB test file
	content := make([]byte, 1024*1024)
	for i := range content {
		content[i] = byte(i)
	}
	
	tmpfile, err := os.CreateTemp("", "benchmark_*.bin")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	
	if _, err := tmpfile.Write(content); err != nil {
		b.Fatal(err)
	}
	tmpfile.Close()
	
	reader, err := NewFirmwareReader(tmpfile.Name(), logger)
	if err != nil {
		b.Fatal(err)
	}
	defer reader.Close()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Read 64KB sections at different offsets
		offset := int64((i * 0x10000) % (len(content) - 0x10000))
		_, err := reader.ReadSection(offset, 0x10000)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestFirmwareReader_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// This test would use actual firmware files if available
	// For now, we'll create a mock firmware structure
	logger := zaptest.NewLogger(t)
	
	// Create a mock FS4 firmware structure
	var firmware bytes.Buffer
	
	// Add padding before magic - use a standard search offset
	firmware.Write(make([]byte, 0x10000))
	
	// Add magic pattern
	magicBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(magicBytes, types.MagicPattern)
	firmware.Write(magicBytes)
	
	// Add some padding
	firmware.Write(make([]byte, 0x10))
	
	// Add HW pointers at offset 0x18 from magic
	hwPtrs := make([]byte, 128)
	binary.BigEndian.PutUint32(hwPtrs[0x00:], 0x0000)     // Boot record
	binary.BigEndian.PutUint32(hwPtrs[0x08:], 0x2000)     // Boot2
	binary.BigEndian.PutUint32(hwPtrs[0x10:], 0x5000)     // TOC
	binary.BigEndian.PutUint32(hwPtrs[0x18:], 0x5000)     // Tools
	firmware.Write(hwPtrs)
	
	// Pad to make file larger - must be larger than the magic offset
	targetSize := 0x20000
	if firmware.Len() < targetSize {
		firmware.Write(make([]byte, targetSize-firmware.Len()))
	}
	
	filename := filepath.Join(t.TempDir(), "test_firmware.bin")
	if err := os.WriteFile(filename, firmware.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	
	reader, err := NewFirmwareReader(filename, logger)
	if err != nil {
		t.Fatalf("NewFirmwareReader() error = %v", err)
	}
	defer reader.Close()
	
	// Test finding magic
	magicOffset, err := reader.FindMagicPattern()
	if err != nil {
		t.Fatalf("FindMagicPattern() error = %v", err)
	}
	if magicOffset != 0x10000 {
		t.Errorf("FindMagicPattern() = 0x%X, want 0x10000", magicOffset)
	}
	
	// Test reading HW pointers
	hwData, err := reader.ReadHWPointers(int64(magicOffset+0x18), 128)
	if err != nil {
		t.Fatalf("ReadHWPointers() error = %v", err)
	}
	
	// Verify Boot2 pointer
	boot2Ptr := binary.BigEndian.Uint32(hwData[0x08:0x0C])
	if boot2Ptr != 0x2000 {
		t.Errorf("Boot2 pointer = 0x%X, want 0x2000", boot2Ptr)
	}
}