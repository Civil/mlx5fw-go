package types

import ()

// CalculateNumSections calculates the actual number of valid sections in an ITOC
// by reading entries until an invalid entry is found
func CalculateNumSections(firmwareData []byte, itocOffset uint32) (int, error) {
	// Skip the ITOC header (32 bytes)
	entriesOffset := itocOffset + 32
	maxEntries := 512 // Maximum entries to check
	numSections := 0
	
	for i := 0; i < maxEntries; i++ {
		entryOffset := entriesOffset + uint32(i)*32
		
		// Check bounds
		if entryOffset+32 > uint32(len(firmwareData)) {
			break
		}
		
		// Read entry data
		entryData := firmwareData[entryOffset : entryOffset+32]
		
		// Parse entry
		entry := &ITOCEntry{}
		if err := entry.Unmarshal(entryData); err != nil {
			return 0, err
		}
		
		// Check if this is a valid entry
		// Invalid entries have type 0xFF or are empty (size and address are 0)
		if entry.GetType() == 0xFF {
			break
		}
		
		if entry.GetSize() == 0 && entry.GetFlashAddr() == 0 {
			// Check if entire entry is zero (end marker)
			allZero := true
			for _, b := range entryData {
				if b != 0 {
					allZero = false
					break
				}
			}
			if allZero {
				break
			}
			// Skip empty entries that aren't end markers
			continue
		}
		
		// Valid entry found
		numSections++
	}
	
	return numSections, nil
}

// GetNumSections is a helper method for ITOCHeader to get the number of sections
// Note: This requires access to the firmware data and ITOC offset
func (h *ITOCHeader) GetNumSections(firmwareData []byte, itocOffset uint32) (int, error) {
	return CalculateNumSections(firmwareData, itocOffset)
}