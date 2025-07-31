package sections

import (
	"encoding/hex"
	"encoding/json"
	
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// ImageSignatureSection represents an IMAGE_SIGNATURE_256 section
type ImageSignatureSection struct {
	*interfaces.BaseSection
	Signature *types.ImageSignature
}

// NewImageSignatureSection creates a new ImageSignature section
func NewImageSignatureSection(base *interfaces.BaseSection) *ImageSignatureSection {
	return &ImageSignatureSection{
		BaseSection: base,
	}
}

// Parse parses the IMAGE_SIGNATURE section data
func (s *ImageSignatureSection) Parse(data []byte) error {
	s.SetRawData(data)
	
	if len(data) < 260 { // 4 bytes type + 256 bytes signature
		return merry.Errorf("IMAGE_SIGNATURE section too small: expected at least 260 bytes, got %d", len(data))
	}
	
	s.Signature = &types.ImageSignature{}
	if err := s.Signature.Unmarshal(data); err != nil {
		return merry.Wrap(err)
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the IMAGE_SIGNATURE section
func (s *ImageSignatureSection) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type":      s.Type(),
		"type_name": s.TypeName(),
		"offset":    s.Offset(),
		"size":      s.Size(),
	}
	
	// Check if signature is blank (all 0xFF)
	isBlank := true
	if s.Signature != nil {
		if s.Signature.SignatureType != 0xFFFFFFFF {
			isBlank = false
		} else {
			for _, b := range s.Signature.Signature {
				if b != 0xFF {
					isBlank = false
					break
				}
			}
		}
		
		// If blank signature with non-standard padding, mark as needing raw data
		if isBlank && s.Size() != 260 { // 4 + 256 
			result["has_raw_data"] = true
		}
		
		result["signature"] = map[string]interface{}{
			"signature_type": s.Signature.SignatureType,
			"signature":      hex.EncodeToString(s.Signature.Signature[:]),
		}
		
		// Check for padding data after the signature
		rawData := s.GetRawData()
		if len(rawData) > 260 {
			paddingData := rawData[260:]
			// Check if padding is non-zero
			hasNonZeroPadding := false
			for _, b := range paddingData {
				if b != 0 {
					hasNonZeroPadding = true
					break
				}
			}
			if hasNonZeroPadding {
				result["padding"] = hex.EncodeToString(paddingData)
			}
		}
	}
	
	return json.Marshal(result)
}

// ImageSignature2Section represents an IMAGE_SIGNATURE_512 section
type ImageSignature2Section struct {
	*interfaces.BaseSection
	Signature *types.ImageSignature2
}

// NewImageSignature2Section creates a new ImageSignature2 section
func NewImageSignature2Section(base *interfaces.BaseSection) *ImageSignature2Section {
	return &ImageSignature2Section{
		BaseSection: base,
	}
}

// Parse parses the IMAGE_SIGNATURE_512 section data
func (s *ImageSignature2Section) Parse(data []byte) error {
	s.SetRawData(data)
	
	if len(data) < 516 { // 4 bytes type + 512 bytes signature
		return merry.Errorf("IMAGE_SIGNATURE_512 section too small: expected at least 516 bytes, got %d", len(data))
	}
	
	s.Signature = &types.ImageSignature2{}
	if err := s.Signature.Unmarshal(data); err != nil {
		return merry.Wrap(err)
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the IMAGE_SIGNATURE_512 section
func (s *ImageSignature2Section) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type":      s.Type(),
		"type_name": s.TypeName(),
		"offset":    s.Offset(),
		"size":      s.Size(),
	}
	
	// Check if signature is blank (all 0xFF)
	isBlank := true
	if s.Signature != nil {
		if s.Signature.SignatureType != 0xFFFFFFFF {
			isBlank = false
		} else {
			for _, b := range s.Signature.Signature {
				if b != 0xFF {
					isBlank = false
					break
				}
			}
		}
		
		// If blank signature with non-standard padding, mark as needing raw data
		if isBlank && s.Size() != 516 { // 4 + 512
			result["has_raw_data"] = true
		}
		
		result["signature"] = map[string]interface{}{
			"signature_type": s.Signature.SignatureType,
			"signature":      hex.EncodeToString(s.Signature.Signature[:]),
		}
		
		// Check for padding data after the signature
		rawData := s.GetRawData()
		if len(rawData) > 516 {
			paddingData := rawData[516:]
			// Check if padding is non-zero
			hasNonZeroPadding := false
			for _, b := range paddingData {
				if b != 0 {
					hasNonZeroPadding = true
					break
				}
			}
			if hasNonZeroPadding {
				result["padding"] = hex.EncodeToString(paddingData)
			}
		}
	}
	
	return json.Marshal(result)
}

// PublicKeysSection represents a PUBLIC_KEYS_2048 section
type PublicKeysSection struct {
	*interfaces.BaseSection
	Keys *types.PublicKeys
}

// NewPublicKeysSection creates a new PublicKeys section
func NewPublicKeysSection(base *interfaces.BaseSection) *PublicKeysSection {
	return &PublicKeysSection{
		BaseSection: base,
	}
}

// Parse parses the PUBLIC_KEYS section data
func (s *PublicKeysSection) Parse(data []byte) error {
	s.SetRawData(data)
	
	s.Keys = &types.PublicKeys{}
	if err := s.Keys.Unmarshal(data); err != nil {
		return merry.Wrap(err)
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the PUBLIC_KEYS section
func (s *PublicKeysSection) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type":      s.Type(),
		"type_name": s.TypeName(),
		"offset":    s.Offset(),
		"size":      s.Size(),
	}
	
	if s.Keys != nil {
		keys := make([]map[string]interface{}, 0)
		for i, key := range s.Keys.Keys {
			keys = append(keys, map[string]interface{}{
				"index":    i,
				"reserved": key.Reserved,
				"uuid":     hex.EncodeToString(key.UUID[:]),
				"key":      hex.EncodeToString(key.Key[:]),
			})
		}
		result["keys"] = keys
	}
	
	return json.Marshal(result)
}

// PublicKeys2Section represents a PUBLIC_KEYS_4096 section
type PublicKeys2Section struct {
	*interfaces.BaseSection
	Keys *types.PublicKeys2
}

// NewPublicKeys2Section creates a new PublicKeys2 section
func NewPublicKeys2Section(base *interfaces.BaseSection) *PublicKeys2Section {
	return &PublicKeys2Section{
		BaseSection: base,
	}
}

// Parse parses the PUBLIC_KEYS_4096 section data
func (s *PublicKeys2Section) Parse(data []byte) error {
	s.SetRawData(data)
	
	s.Keys = &types.PublicKeys2{}
	if err := s.Keys.Unmarshal(data); err != nil {
		return merry.Wrap(err)
	}
	
	return nil
}

// MarshalJSON returns JSON representation of the PUBLIC_KEYS_4096 section
func (s *PublicKeys2Section) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type":      s.Type(),
		"type_name": s.TypeName(),
		"offset":    s.Offset(),
		"size":      s.Size(),
	}
	
	if s.Keys != nil {
		keys := make([]map[string]interface{}, 0)
		for i, key := range s.Keys.Keys {
			keys = append(keys, map[string]interface{}{
				"index":    i,
				"reserved": key.Reserved,
				"uuid":     hex.EncodeToString(key.UUID[:]),
				"key":      hex.EncodeToString(key.Key[:]),
			})
		}
		result["keys"] = keys
	}
	
	return json.Marshal(result)
}