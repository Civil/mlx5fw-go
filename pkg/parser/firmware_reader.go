package parser

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/ansel1/merry/v2"
	"go.uber.org/zap"

	"github.com/Civil/mlx5fw-go/pkg/errs"
	"github.com/Civil/mlx5fw-go/pkg/types"
)

// FirmwareReader provides low-level firmware file reading operations
type FirmwareReader struct {
	file     *os.File
	filePath string
	size     int64
	logger   *zap.Logger
}

// NewFirmwareReader creates a new firmware reader
func NewFirmwareReader(filePath string, logger *zap.Logger) (*FirmwareReader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, merry.Wrap(err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, merry.Wrap(err)
	}

	return &FirmwareReader{
		file:     file,
		filePath: filePath,
		size:     stat.Size(),
		logger:   logger,
	}, nil
}

// Close closes the firmware file
func (r *FirmwareReader) Close() error {
	if r.file != nil {
		err := r.file.Close()
		r.file = nil
		return err
	}
	return nil
}

// Size returns the size of the firmware file
func (r *FirmwareReader) Size() int64 {
	return r.size
}

// ReadAt implements io.ReaderAt interface
func (r *FirmwareReader) ReadAt(p []byte, off int64) (n int, err error) {
	return r.file.ReadAt(p, off)
}

// FindMagicPattern searches for the firmware magic pattern at standard offsets
func (r *FirmwareReader) FindMagicPattern() (uint32, error) {
	buf := make([]byte, 8)
	for _, offset := range types.MagicSearchOffsets {
		if int64(offset) >= r.size {
			continue
		}

		_, err := r.ReadAt(buf, int64(offset))
		if err != nil {
			if err == io.EOF {
				continue
			}
			return 0, err
		}

		magic := binary.BigEndian.Uint64(buf)
		if magic == types.MagicPattern {
			r.logger.Debug("Found magic pattern", zap.Uint32("offset", offset))
			return offset, nil
		}
	}
	return 0, errs.ErrInvalidMagic
}

// ReadSection reads a section of data from the firmware
func (r *FirmwareReader) ReadSection(offset int64, size uint32) ([]byte, error) {
	if offset < 0 || offset >= r.size {
		return nil, merry.Errorf("invalid offset: %d", offset)
	}
	
	if int64(offset+int64(size)) > r.size {
		return nil, merry.Errorf("section extends beyond file size")
	}

	buf := make([]byte, size)
	_, err := r.ReadAt(buf, offset)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	return buf, nil
}

// FileInfo represents firmware file information
type FileInfo struct {
	Size   int64
	SHA256 string
}

// GetFileInfo returns information about the firmware file
func (r *FirmwareReader) GetFileInfo() (*FileInfo, error) {
	// Calculate SHA256 hash
	hasher := sha256.New()
	_, err := r.file.Seek(0, 0)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	
	if _, err := io.Copy(hasher, r.file); err != nil {
		return nil, merry.Wrap(err)
	}
	
	return &FileInfo{
		Size:   r.size,
		SHA256: fmt.Sprintf("%x", hasher.Sum(nil)),
	}, nil
}