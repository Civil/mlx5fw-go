package reassemble

// ITOCEntryMetadata represents ITOC entry information in metadata
type ITOCEntryMetadata struct {
	Type       uint16 `json:"type"`
	FlashAddr  uint32 `json:"flash_addr"`
	SectionCRC uint16 `json:"section_crc"`
	Encrypted  bool   `json:"encrypted"`
}
