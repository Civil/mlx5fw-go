package types

// Type aliases to use annotated versions directly
type MfgInfo = MfgInfoAnnotated

// GetPSIDString returns the PSID as a string
func (m *MfgInfo) GetPSIDString() string {
	return nullTerminatedString(m.PSID[:])
}