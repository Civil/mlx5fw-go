package sections

import (
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// ImageSignatureSection represents an IMAGE_SIGNATURE_256 section
type ImageSignatureSection struct {
	*interfaces.BaseSection
	ImageSignature *types.ImageSignature `json:"image_signature,omitempty"`
	Padding        types.FWByteSlice     `json:"padding,omitempty"`
}

// NewImageSignatureSection creates a new ImageSignature section
func NewImageSignatureSection(base *interfaces.BaseSection) *ImageSignatureSection {
	base.HasRawData = true // Default to true until successfully parsed
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

	s.ImageSignature = &types.ImageSignature{}
	if err := s.ImageSignature.Unmarshal(data); err != nil {
		return merry.Wrap(err)
	}

	// Check for padding data after the signature
	if len(data) > 260 {
		paddingData := data[260:]
		// Check if padding is non-zero
		hasNonZeroPadding := false
		for _, b := range paddingData {
			if b != 0 {
				hasNonZeroPadding = true
				break
			}
		}
		if hasNonZeroPadding {
			s.Padding = types.FWByteSlice(paddingData)
		}
	}

	s.BaseSection.HasRawData = false // Successfully parsed
	return nil
}

// ImageSignature2Section represents an IMAGE_SIGNATURE_512 section
type ImageSignature2Section struct {
	*interfaces.BaseSection
	ImageSignature *types.ImageSignature2 `json:"image_signature,omitempty"`
	Padding        types.FWByteSlice      `json:"padding,omitempty"`
}

// NewImageSignature2Section creates a new ImageSignature2 section
func NewImageSignature2Section(base *interfaces.BaseSection) *ImageSignature2Section {
	base.HasRawData = true // Default to true until successfully parsed
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

	s.ImageSignature = &types.ImageSignature2{}
	if err := s.ImageSignature.Unmarshal(data); err != nil {
		return merry.Wrap(err)
	}

	// Check for padding data after the signature
	if len(data) > 516 {
		paddingData := data[516:]
		// Check if padding is non-zero
		hasNonZeroPadding := false
		for _, b := range paddingData {
			if b != 0 {
				hasNonZeroPadding = true
				break
			}
		}
		if hasNonZeroPadding {
			s.Padding = types.FWByteSlice(paddingData)
		}
	}

	s.BaseSection.HasRawData = false // Successfully parsed
	return nil
}

// PublicKeysSection represents a PUBLIC_KEYS_2048 section
type PublicKeysSection struct {
	*interfaces.BaseSection
	PublicKeys *types.PublicKeys `json:"public_keys,omitempty"`
}

// NewPublicKeysSection creates a new PublicKeys section
func NewPublicKeysSection(base *interfaces.BaseSection) *PublicKeysSection {
	base.HasRawData = true // Default to true until successfully parsed
	return &PublicKeysSection{
		BaseSection: base,
	}
}

// Parse parses the PUBLIC_KEYS section data
func (s *PublicKeysSection) Parse(data []byte) error {
	s.SetRawData(data)

	s.PublicKeys = &types.PublicKeys{}
	if err := s.PublicKeys.Unmarshal(data); err != nil {
		return merry.Wrap(err)
	}

	s.BaseSection.HasRawData = false // Successfully parsed
	return nil
}

// PublicKeys2Section represents a PUBLIC_KEYS_4096 section
type PublicKeys2Section struct {
	*interfaces.BaseSection
	PublicKeys *types.PublicKeys2 `json:"public_keys,omitempty"`
}

// NewPublicKeys2Section creates a new PublicKeys2 section
func NewPublicKeys2Section(base *interfaces.BaseSection) *PublicKeys2Section {
	base.HasRawData = true // Default to true until successfully parsed
	return &PublicKeys2Section{
		BaseSection: base,
	}
}

// Parse parses the PUBLIC_KEYS_4096 section data
func (s *PublicKeys2Section) Parse(data []byte) error {
	s.SetRawData(data)

	s.PublicKeys = &types.PublicKeys2{}
	if err := s.PublicKeys.Unmarshal(data); err != nil {
		return merry.Wrap(err)
	}

	s.BaseSection.HasRawData = false // Successfully parsed
	return nil
}
