package parser

import (
	"bytes"
	"encoding/binary"

	"github.com/Civil/mlx5fw-go/pkg/types"
)

// GapHandler handles gaps between sections in firmware
type GapHandler struct {
	// Minimum gap size to consider (smaller gaps are ignored)
	MinGapSize uint64
	// Whether to skip gaps that are entirely 0xFF
	SkipEmptyGaps bool
}

// NewGapHandler creates a new gap handler with default settings
func NewGapHandler() *GapHandler {
	return &GapHandler{
		MinGapSize:    1, // By default, consider all gaps
		SkipEmptyGaps: true,
	}
}

// GapInfo contains information about a gap
type GapInfo struct {
	Start    uint64
	End      uint64
	Size     uint64
	IsEmpty  bool   // True if gap contains only 0xFF bytes
	IsHeader bool   // True if this is the header gap (before first section)
}

// AnalyzeGap analyzes a gap in the firmware data
func (h *GapHandler) AnalyzeGap(data []byte, start, end uint64) *GapInfo {
	size := end - start
	if size < h.MinGapSize {
		return nil
	}

	gap := &GapInfo{
		Start: start,
		End:   end,
		Size:  size,
	}

	// Check if this is the header gap (before magic pattern)
	if start == 0 && size <= types.MagicPatternOffset {
		gap.IsHeader = true
	}

	// Check if gap is empty (all 0xFF)
	if h.SkipEmptyGaps && size > 0 && int(start+size) <= len(data) {
		gap.IsEmpty = h.isEmptyGap(data[start:end])
	}

	return gap
}

// isEmptyGap checks if a gap contains only 0xFF bytes
func (h *GapHandler) isEmptyGap(data []byte) bool {
	// For performance, check in chunks
	const chunkSize = 1024
	emptyChunk := make([]byte, chunkSize)
	for i := range emptyChunk {
		emptyChunk[i] = 0xFF
	}

	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := data[i:end]
		
		if len(chunk) == chunkSize {
			if !bytes.Equal(chunk, emptyChunk) {
				return false
			}
		} else {
			// Check remaining bytes
			for _, b := range chunk {
				if b != 0xFF {
					return false
				}
			}
		}
	}
	return true
}

// ShouldExtractGap determines if a gap should be extracted
func (h *GapHandler) ShouldExtractGap(gap *GapInfo) bool {
	if gap == nil {
		return false
	}

	// Skip header gap as it can be reconstructed
	if gap.IsHeader {
		return false
	}

	// Skip empty gaps if configured
	if h.SkipEmptyGaps && gap.IsEmpty {
		return false
	}

	return true
}

// GenerateHeaderGap generates the standard header gap for FS4 firmware
func GenerateFS4HeaderGap() []byte {
	// FS4 header is standardized:
	// - 0x00-0x07: All zeros (boot signature area)
	// - 0x08-0x0F: Magic pattern
	// - 0x10-0x17: All 0xFF
	// - 0x18+: HW pointers (handled separately)
	
	header := make([]byte, types.MagicPatternOffset)
	
	// Fill with 0xFF by default
	for i := range header {
		header[i] = 0xFF
	}
	
	// Clear boot signature area
	for i := 0; i < 8; i++ {
		header[i] = 0
	}
	
	// Magic pattern will be written separately
	return header
}

// ReconstructHeader reconstructs the standard firmware header
func ReconstructFS4Header(magicOffset uint32) []byte {
	if magicOffset != types.MagicPatternOffset {
		// Non-standard magic offset, can't use standard header
		return nil
	}
	
	header := GenerateFS4HeaderGap()
	
	// Write magic pattern
	binary.BigEndian.PutUint64(header[magicOffset:], types.MagicPattern)
	
	return header
}