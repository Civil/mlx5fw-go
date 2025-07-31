package types

// Type aliases to use annotated versions directly
type DevInfo = DevInfoAnnotated
type UidEntry = UidEntryAnnotated

// GetNumAllocated returns the total number of allocated UIDs
func (u *UidEntry) GetNumAllocated() int {
	return int(u.NumAllocated)
}

// GetUID returns the UID with correct byte order
func (u *UidEntry) GetUID() uint64 {
	return u.UID
}