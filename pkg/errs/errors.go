package errs

import "github.com/ansel1/merry/v2"

var (
	// File format errors
	ErrInvalidMagic      = merry.New("invalid magic pattern")
	ErrInvalidCRC        = merry.New("CRC verification failed")
	ErrUnsupportedFormat = merry.New("unsupported firmware format")
	ErrEncryptedImage    = merry.New("encrypted images not supported")
	
	// Parsing errors
	ErrInvalidPointer   = merry.New("invalid pointer value")
	ErrInvalidITOC      = merry.New("invalid ITOC structure")
	ErrInvalidDTOC      = merry.New("invalid DTOC structure")
	ErrSectionNotFound  = merry.New("section not found")
	ErrInvalidSection   = merry.New("invalid section data")
	
	// IO errors
	ErrReadFailed       = merry.New("failed to read data")
	ErrInvalidOffset    = merry.New("invalid offset")
	ErrInvalidSize      = merry.New("invalid size")
	ErrInvalidDataSize  = merry.New("invalid data size")
)