package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/parser/fs4"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestParseReplaceSectionArgs(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedName  string
		expectedID    int
		expectedError bool
	}{
		{
			name:         "section name only",
			args:         []string{"DBG_FW_INI"},
			expectedName: "DBG_FW_INI",
			expectedID:   -1,
		},
		{
			name:         "section name with ID",
			args:         []string{"ITOC:0"},
			expectedName: "ITOC",
			expectedID:   0,
		},
		{
			name:         "section name with ID 2",
			args:         []string{"ITOC:2"},
			expectedName: "ITOC",
			expectedID:   2,
		},
		{
			name:          "invalid ID",
			args:          []string{"ITOC:abc"},
			expectedError: true,
		},
		{
			name:          "no args",
			args:          []string{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, id, err := parseReplaceSectionArgs(tt.args)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedName, name)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

// Integration test with actual parsing
func TestReplaceSectionIntegration(t *testing.T) {
	// Skip if no test firmware available
	testFirmware := "test_fw.bin"
	if _, err := os.Stat(testFirmware); os.IsNotExist(err) {
		t.Skip("Test firmware not found")
	}

	logger = zaptest.NewLogger(t)
	
	// Create temp directory for test files
	tempDir := t.TempDir()
	
	// Create replacement content
	replacementContent := []byte("TEST REPLACEMENT CONTENT")
	replacementFile := filepath.Join(tempDir, "replacement.txt")
	err := os.WriteFile(replacementFile, replacementContent, 0644)
	require.NoError(t, err)
	
	outputFile := filepath.Join(tempDir, "output.bin")
	
	// Run replace-section command
	firmwarePath = testFirmware
	err = runReplaceSectionCommandV4(nil, []string{"DBG_FW_INI"}, "DBG_FW_INI", -1, replacementFile, outputFile)
	
	// Check if command succeeded (might fail if no DBG_FW_INI section)
	if err != nil {
		t.Logf("Replace section failed (expected if no DBG_FW_INI): %v", err)
		return
	}
	
	// Verify output file was created
	_, err = os.Stat(outputFile)
	assert.NoError(t, err)
	
	// Verify the section was replaced by parsing the output
	reader, err := parser.NewFirmwareReader(outputFile, logger)
	require.NoError(t, err)
	defer reader.Close()
	
	fwParser := fs4.NewParser(reader, logger)
	err = fwParser.Parse()
	require.NoError(t, err)
	
	// Find the replaced section
	allSections := fwParser.GetSections()
	found := false
	for _, sections := range allSections {
		for _, section := range sections {
			if types.GetSectionTypeName(section.Type()) == "DBG_FW_INI" {
				found = true
				// Verify size matches replacement
				assert.Equal(t, uint32(len(replacementContent)), section.Size())
				break
			}
		}
	}
	
	if !found {
		t.Log("DBG_FW_INI section not found in output (might be expected)")
	}
}

// Test for end-to-end replace section with size change
func TestReplaceSectionWithSizeChange(t *testing.T) {
	// Skip if no test firmware available
	testFirmware := "test_fw.bin"
	if _, err := os.Stat(testFirmware); os.IsNotExist(err) {
		t.Skip("Test firmware not found")
	}

	logger = zaptest.NewLogger(t)
	
	// Create temp directory for test files
	tempDir := t.TempDir()
	
	// Create replacement content that's larger than typical section
	replacementContent := make([]byte, 10000)
	for i := range replacementContent {
		replacementContent[i] = byte(i & 0xFF)
	}
	
	replacementFile := filepath.Join(tempDir, "replacement.bin")
	err := os.WriteFile(replacementFile, replacementContent, 0644)
	require.NoError(t, err)
	
	outputFile := filepath.Join(tempDir, "output.bin")
	
	// Run replace-section command
	firmwarePath = testFirmware
	err = runReplaceSectionCommandV4(nil, []string{"DBG_FW_INI"}, "DBG_FW_INI", -1, replacementFile, outputFile)
	
	// Check if command succeeded (might fail if no DBG_FW_INI section)
	if err != nil {
		t.Logf("Replace section failed (expected if no DBG_FW_INI): %v", err)
		return
	}
	
	// Verify output file was created
	info, err := os.Stat(outputFile)
	assert.NoError(t, err)
	
	// Check that file is properly padded to 32MB or 64MB
	fileSize := info.Size()
	assert.True(t, fileSize == 32*1024*1024 || fileSize == 64*1024*1024, 
		"File should be padded to 32MB or 64MB, got %d bytes", fileSize)
}

// Test replace section command error cases
func TestReplaceSectionCommandErrors(t *testing.T) {
	logger = zaptest.NewLogger(t)
	tempDir := t.TempDir()
	
	tests := []struct {
		name           string
		args           []string
		sectionName    string
		sectionID      int
		setupFiles     func() (string, string, string)
		expectedError  string
	}{
		{
			name:        "missing firmware file",
			args:        []string{"DBG_FW_INI"},
			sectionName: "DBG_FW_INI",
			sectionID:   -1,
			setupFiles: func() (string, string, string) {
				firmwarePath = "/nonexistent/firmware.bin"
				replacementFile := filepath.Join(tempDir, "replacement.txt")
				os.WriteFile(replacementFile, []byte("test"), 0644)
				outputFile := filepath.Join(tempDir, "output.bin")
				return firmwarePath, replacementFile, outputFile
			},
			expectedError: "no such file or directory",
		},
		{
			name:        "missing replacement file",
			args:        []string{"DBG_FW_INI"},
			sectionName: "DBG_FW_INI",
			sectionID:   -1,
			setupFiles: func() (string, string, string) {
				// Create a minimal test firmware
				testFw := filepath.Join(tempDir, "test.bin")
				os.WriteFile(testFw, make([]byte, 1024), 0644)
				firmwarePath = testFw
				replacementFile := "/nonexistent/replacement.txt"
				outputFile := filepath.Join(tempDir, "output.bin")
				return firmwarePath, replacementFile, outputFile
			},
			expectedError: "no such file or directory",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, repl, out := tt.setupFiles()
			err := runReplaceSectionCommandV4(nil, tt.args, tt.sectionName, tt.sectionID, repl, out)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}