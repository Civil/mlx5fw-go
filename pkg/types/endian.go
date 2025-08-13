package types

import (
	"encoding/binary"
	"io"
)

// Reader provides endian-aware reading capabilities
type Reader struct {
	r     io.ReaderAt
	order binary.ByteOrder
}

// NewBigEndianReader creates a new big-endian reader
func NewBigEndianReader(r io.ReaderAt) *Reader {
	return &Reader{
		r:     r,
		order: binary.BigEndian,
	}
}

// ReadAt implements io.ReaderAt
func (r *Reader) ReadAt(p []byte, off int64) (int, error) {
	return r.r.ReadAt(p, off)
}

// ReadUint32At reads a uint32 at the specified offset
func (r *Reader) ReadUint32At(off int64) (uint32, error) {
	buf := make([]byte, 4)
	if _, err := r.r.ReadAt(buf, off); err != nil {
		return 0, err
	}
	return r.order.Uint32(buf), nil
}

// ReadUint64At reads a uint64 at the specified offset
func (r *Reader) ReadUint64At(off int64) (uint64, error) {
	buf := make([]byte, 8)
	if _, err := r.r.ReadAt(buf, off); err != nil {
		return 0, err
	}
	return r.order.Uint64(buf), nil
}

// ReadBytes reads a byte slice at the specified offset
func (r *Reader) ReadBytes(off int64, size int) ([]byte, error) {
	buf := make([]byte, size)
	n, err := r.r.ReadAt(buf, off)
	if err != nil {
		return nil, err
	}
	if n != size {
		return nil, io.ErrUnexpectedEOF
	}
	return buf, nil
}
