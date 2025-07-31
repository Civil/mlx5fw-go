package types

import ()

// Marshal/Unmarshal methods for SecureBootSignatures
func (s *SecureBootSignatures) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &SecureBootSignaturesAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*s = *annotated.FromAnnotated()
	return nil
}

func (s *SecureBootSignatures) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := s.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for ResetCapabilities
func (r *ResetCapabilities) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &ResetCapabilitiesAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*r = *annotated.FromAnnotated()
	return nil
}

func (r *ResetCapabilities) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := r.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for ResetVersion
func (r *ResetVersion) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &ResetVersionAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*r = *annotated.FromAnnotated()
	return nil
}

func (r *ResetVersion) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := r.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for VersionVector
func (v *VersionVector) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &VersionVectorAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*v = *annotated.FromAnnotated()
	return nil
}

func (v *VersionVector) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := v.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for ResetInfo
func (r *ResetInfo) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &ResetInfoAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*r = *annotated.FromAnnotated()
	return nil
}

func (r *ResetInfo) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := r.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}


// Marshal/Unmarshal methods for HMACDigest
func (h *HMACDigest) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &HMACDigestAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*h = *annotated.FromAnnotated()
	return nil
}

func (h *HMACDigest) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := h.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for RSAPublicKey
func (r *RSAPublicKey) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &RSAPublicKeyAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*r = *annotated.FromAnnotated()
	return nil
}

func (r *RSAPublicKey) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := r.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for DBGFwIni
func (d *DBGFwIni) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &DBGFwIniAnnotated{}
	headerSize := 16
	if len(data) >= headerSize {
		if err := annotated.Unmarshal(data[:headerSize]); err != nil {
			return err
		}
	} else {
		if err := annotated.Unmarshal(data); err != nil {
			return err
		}
	}
	// Convert to legacy format
	*d = *annotated.FromAnnotated()
	return nil
}

func (d *DBGFwIni) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := d.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for DBGFwParams
func (d *DBGFwParams) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &DBGFwParamsAnnotated{}
	headerSize := 16
	if len(data) >= headerSize {
		if err := annotated.Unmarshal(data[:headerSize]); err != nil {
			return err
		}
	} else {
		if err := annotated.Unmarshal(data); err != nil {
			return err
		}
	}
	// Convert to legacy format
	*d = *annotated.FromAnnotated()
	return nil
}

func (d *DBGFwParams) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := d.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for FWAdb
func (f *FWAdb) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &FWAdbAnnotated{}
	headerSize := 16
	if len(data) >= headerSize {
		if err := annotated.Unmarshal(data[:headerSize]); err != nil {
			return err
		}
	} else {
		if err := annotated.Unmarshal(data); err != nil {
			return err
		}
	}
	// Convert to legacy format
	*f = *annotated.FromAnnotated()
	return nil
}

func (f *FWAdb) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := f.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for CRDumpMaskData
func (c *CRDumpMaskData) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &CRDumpMaskDataAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*c = *annotated.FromAnnotated()
	return nil
}

func (c *CRDumpMaskData) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := c.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for FWNVLog
func (f *FWNVLog) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &FWNVLogAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*f = *annotated.FromAnnotated()
	return nil
}

func (f *FWNVLog) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := f.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for VPD_R0
func (v *VPD_R0) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &VPD_R0Annotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*v = *annotated.FromAnnotated()
	return nil
}

func (v *VPD_R0) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := v.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for NVData
func (n *NVData) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &NVDataAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*n = *annotated.FromAnnotated()
	return nil
}

func (n *NVData) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := n.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for FWInternalUsage
func (f *FWInternalUsage) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &FWInternalUsageAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*f = *annotated.FromAnnotated()
	return nil
}

func (f *FWInternalUsage) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := f.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for ProgrammableHWFW
func (p *ProgrammableHWFW) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &ProgrammableHWFWAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*p = *annotated.FromAnnotated()
	return nil
}

func (p *ProgrammableHWFW) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := p.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for DigitalCertPtr
func (d *DigitalCertPtr) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &DigitalCertPtrAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*d = *annotated.FromAnnotated()
	return nil
}

func (d *DigitalCertPtr) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := d.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for DigitalCertRW
func (d *DigitalCertRW) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &DigitalCertRWAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*d = *annotated.FromAnnotated()
	return nil
}

func (d *DigitalCertRW) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := d.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for CertChain
func (c *CertChain) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &CertChainAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*c = *annotated.FromAnnotated()
	return nil
}

func (c *CertChain) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := c.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for RootCertificates
func (r *RootCertificates) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &RootCertificatesAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*r = *annotated.FromAnnotated()
	return nil
}

func (r *RootCertificates) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := r.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}

// Marshal/Unmarshal methods for CertificateChains
func (c *CertificateChains) Unmarshal(data []byte) error {
	// Use annotated version for unmarshaling
	annotated := &CertificateChainsAnnotated{}
	if err := annotated.Unmarshal(data); err != nil {
		return err
	}
	// Convert to legacy format
	*c = *annotated.FromAnnotated()
	return nil
}

func (c *CertificateChains) Marshal() ([]byte, error) {
	// Convert to annotated format
	annotated := c.ToAnnotated()
	// Marshal using annotation-based marshaling
	return annotated.Marshal()
}