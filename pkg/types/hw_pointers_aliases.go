package types

// Type aliases to use annotated versions directly
type HWPointerEntry = HWPointerEntryAnnotated
type FS4HWPointers = FS4HWPointersAnnotated
type FS5HWPointers = FS5HWPointersAnnotated

// Compatibility method to get CRC as uint32
func (h *HWPointerEntry) GetCRC32() uint32 {
	return uint32(h.CRC)
}