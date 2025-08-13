package types

// ParseFields parses the bit-packed fields from the raw data
// Based on image_layout_itoc_entry_unpack from mstflint
func (e *ITOCEntry) ParseFields() {
	// Fields are parsed by Unmarshal; kept for compatibility
}

// Compatibility helpers
func (e *ITOCEntry) GetCRC() uint8         { return e.CRCField }
func (e *ITOCEntry) SetType(t uint8)       { e.Type = t }
func (e *ITOCEntry) SetSize(s uint32)      { e.SizeDwords = s / 4 }
func (e *ITOCEntry) SetFlashAddr(a uint32) { e.FlashAddrDwords = a / 8 }
func (e *ITOCEntry) SetParam0(p uint32) {
	e.Param0Low = uint32(p & 0xF)
	e.Param0High = p >> 4
}
func (e *ITOCEntry) SetParam1(p uint32)       { e.Param1 = p }
func (e *ITOCEntry) SetCRC(c uint8)           { e.CRCField = c }
func (e *ITOCEntry) SetSectionCRC(c uint16)   { e.SectionCRC = c }
func (e *ITOCEntry) SetEncrypted(enc bool)    { e.Encrypted = enc }
func (e *ITOCEntry) SetITOCEntryCRC(c uint16) { e.ITOCEntryCRC = c }
