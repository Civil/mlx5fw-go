package parser

import (
	"encoding/binary"
	"io"
	"os"

	"github.com/ansel1/merry/v2"
	"go.uber.org/zap"

	"github.com/Civil/mlx5fw-go/pkg/errs"
	"github.com/Civil/mlx5fw-go/pkg/types"
)

// FirmwareReader provides low-level firmware file reading operations
type FirmwareReader struct {
	file   *os.File
	size   int64
	logger *zap.Logger
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
		file:   file,
		size:   stat.Size(),
		logger: logger,
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

// ReadUint32BE reads a big-endian uint32 at the specified offset
func (r *FirmwareReader) ReadUint32BE(offset int64) (uint32, error) {
	buf := make([]byte, 4)
	_, err := r.ReadAt(buf, offset)
	if err != nil {
		return 0, merry.Wrap(err)
	}
	return binary.BigEndian.Uint32(buf), nil
}

// ReadUint64BE reads a big-endian uint64 at the specified offset
func (r *FirmwareReader) ReadUint64BE(offset int64) (uint64, error) {
	buf := make([]byte, 8)
	_, err := r.ReadAt(buf, offset)
	if err != nil {
		return 0, merry.Wrap(err)
	}
	return binary.BigEndian.Uint64(buf), nil
}

// FindMagicPattern searches for the firmware magic pattern at standard offsets
func (r *FirmwareReader) FindMagicPattern() (uint32, error) {
	for _, offset := range types.MagicSearchOffsets {
		if int64(offset) >= r.size {
			continue
		}

		magic, err := r.ReadUint64BE(int64(offset))
		if err != nil {
			if err == io.EOF {
				continue
			}
			return 0, err
		}

		if magic == types.MagicPattern {
			r.logger.Debug("Found magic pattern", zap.Uint32("offset", offset))
			return offset, nil
		}
	}
	return 0, errs.ErrInvalidMagic
}

// ReadHWPointers reads hardware pointers at the given offset
func (r *FirmwareReader) ReadHWPointers(offset int64, size int) ([]byte, error) {
	buf := make([]byte, size)
	_, err := r.ReadAt(buf, offset)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	return buf, nil
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