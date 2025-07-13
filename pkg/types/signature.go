package types

// SignatureBlock represents a firmware signature block
// Used for IMAGE_SIGNATURE_256 and IMAGE_SIGNATURE_512 sections
type SignatureBlock struct {
	SignatureType   uint32   `bin:"BE"`    // Type of signature (RSA-2048, RSA-4096, etc.)
	SignatureLength uint32   `bin:"BE"`    // Length of signature data
	Reserved1       uint32   `bin:"BE"`    // Reserved field
	Reserved2       uint32   `bin:"BE"`    // Reserved field
	Signature       []byte   `bin:""`      // Variable length signature data
}

// PublicKeyBlock represents a public key block
// Used for PUBLIC_KEYS_2048 and PUBLIC_KEYS_4096 sections
type PublicKeyBlock struct {
	KeyType         uint32   `bin:"BE"`    // Type of key (RSA-2048, RSA-4096, etc.)
	KeyLength       uint32   `bin:"BE"`    // Length of key data
	KeyID           uint32   `bin:"BE"`    // Key identifier
	Reserved        uint32   `bin:"BE"`    // Reserved field
	PublicKey       []byte   `bin:""`      // Variable length public key data
}

// ForbiddenVersionsHeader represents the header for forbidden versions section
type ForbiddenVersionsHeader struct {
	Magic           uint32   `bin:"BE"`    // Magic value
	Version         uint32   `bin:"BE"`    // Format version
	NumEntries      uint32   `bin:"BE"`    // Number of forbidden version entries
	Reserved        uint32   `bin:"BE"`    // Reserved field
}

// ForbiddenVersionEntry represents a single forbidden version entry
type ForbiddenVersionEntry struct {
	MinVersion      uint32   `bin:"BE"`    // Minimum forbidden version
	MaxVersion      uint32   `bin:"BE"`    // Maximum forbidden version
	Flags           uint32   `bin:"BE"`    // Flags for this entry
	Reserved        uint32   `bin:"BE"`    // Reserved field
}