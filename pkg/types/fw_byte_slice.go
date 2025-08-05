package types

import (
	"encoding/hex"
	"encoding/json"
)

// FWByteSlice is a helper type for byte slices that marshals/unmarshals as hex strings in JSON
type FWByteSlice []byte

// MarshalJSON implements json.Marshaler interface
func (b FWByteSlice) MarshalJSON() ([]byte, error) {
	if len(b) == 0 {
		return json.Marshal("")
	}
	return json.Marshal(hex.EncodeToString(b))
}

// UnmarshalJSON implements json.Unmarshaler interface
func (b *FWByteSlice) UnmarshalJSON(data []byte) error {
	var hexStr string
	if err := json.Unmarshal(data, &hexStr); err != nil {
		return err
	}
	
	if hexStr == "" {
		*b = nil
		return nil
	}
	
	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		return err
	}
	
	*b = decoded
	return nil
}

// Bytes returns the underlying byte slice
func (b FWByteSlice) Bytes() []byte {
	return []byte(b)
}

// SetBytes sets the byte slice from a byte array
func (b *FWByteSlice) SetBytes(data []byte) {
	*b = FWByteSlice(data)
}

// String returns hex representation
func (b FWByteSlice) String() string {
	return hex.EncodeToString(b)
}