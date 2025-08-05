// Package errors provides domain-specific error types and helpers for mlx5fw-go
package errors

import (
	"github.com/ansel1/merry/v2"
)

// Domain-specific error types
var (
	// ErrInvalidData indicates invalid or corrupted data
	ErrInvalidData = merry.New("invalid data")
	
	// ErrDataTooShort indicates data is shorter than expected
	ErrDataTooShort = merry.New("data too short")
	
	// ErrCRCMismatch indicates CRC validation failed
	ErrCRCMismatch = merry.New("CRC mismatch")
	
	// ErrNotSupported indicates an unsupported operation or format
	ErrNotSupported = merry.New("not supported")
	
	// ErrSectionNotFound indicates a required section was not found
	ErrSectionNotFound = merry.New("section not found")
	
	// ErrInvalidMagic indicates invalid magic pattern
	ErrInvalidMagic = merry.New("invalid magic pattern")
	
	// ErrFileTooLarge indicates file exceeds size limits
	ErrFileTooLarge = merry.New("file too large")
	
	// ErrInvalidParameter indicates invalid function parameter
	ErrInvalidParameter = merry.New("invalid parameter")
)

// DataTooShortError creates an error for insufficient data with detailed context
func DataTooShortError(expected, actual int, context string) error {
	return merry.Wrap(ErrDataTooShort, 
		merry.WithMessagef("%s: expected at least %d bytes, got %d", context, expected, actual))
}

// CRCMismatchData holds CRC mismatch details
type CRCMismatchData struct {
	Expected uint32
	Actual   uint32
	Section  string
}

// CRCMismatchError creates an error for CRC validation failure with details
func CRCMismatchError(expected, actual uint32, section string) error {
	return merry.Wrap(ErrCRCMismatch,
		merry.WithMessagef("section %s: expected 0x%X, got 0x%X", section, expected, actual),
		merry.WithValue("crc_data", &CRCMismatchData{
			Expected: expected,
			Actual:   actual,
			Section:  section,
		}))
}

// GetCRCMismatchData extracts CRC mismatch details from an error
func GetCRCMismatchData(err error) (*CRCMismatchData, bool) {
	if val := merry.Value(err, "crc_data"); val != nil {
		if data, ok := val.(*CRCMismatchData); ok {
			return data, true
		}
	}
	return nil, false
}

// NotSupportedError creates an error for unsupported operations
func NotSupportedError(operation string) error {
	return merry.Wrap(ErrNotSupported, merry.WithMessage(operation))
}

// SectionNotFoundError creates an error for missing sections
func SectionNotFoundError(sectionType string, offset uint64) error {
	return merry.Wrap(ErrSectionNotFound,
		merry.WithMessagef("section type %s at offset 0x%X", sectionType, offset))
}

// InvalidMagicError creates an error for invalid magic pattern
func InvalidMagicError(expected, actual uint32, offset uint64) error {
	return merry.Wrap(ErrInvalidMagic,
		merry.WithMessagef("at offset 0x%X: expected 0x%08X, got 0x%08X", offset, expected, actual))
}

// FileTooLargeError creates an error for files exceeding size limits
func FileTooLargeError(size, limit int64) error {
	return merry.Wrap(ErrFileTooLarge,
		merry.WithMessagef("size %d bytes exceeds limit of %d bytes", size, limit))
}

// InvalidParameterError creates an error for invalid function parameters
func InvalidParameterError(parameter, reason string) error {
	return merry.Wrap(ErrInvalidParameter,
		merry.WithMessagef("parameter '%s': %s", parameter, reason))
}