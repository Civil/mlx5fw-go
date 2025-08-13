package annotations

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// Endianness represents the byte order
type Endianness int

const (
	// BigEndian represents big-endian byte order
	BigEndian Endianness = iota
	// LittleEndian represents little-endian byte order
	LittleEndian
	// HostEndian represents the host's native byte order
	HostEndian
)

// FieldAnnotation represents the annotations for a struct field
type FieldAnnotation struct {
	// ByteOffset is the offset in bytes from the start of the struct
	ByteOffset int
	// BitOffset is the absolute bit offset from the start of the struct (for bitfields)
	BitOffset int
	// BitLength is the length in bits for bitfield fields
	BitLength int
	// Endianness specifies the byte order for multi-byte fields
	Endianness Endianness
	// Skip indicates this field should be skipped during marshaling/unmarshaling
	Skip bool
	// FieldName is the original struct field name
	FieldName string
	// FieldType is the reflect.Type of the field
	FieldType reflect.Type
	// ArrayLength is the length for array fields
	ArrayLength int
	// IsArray indicates if this is an array field
	IsArray bool
	// IsBitfield indicates if this is a bitfield
	IsBitfield bool
	// Reserved indicates this is a reserved/padding field
	Reserved bool
	// HexAsDec indicates values should be treated as hex digits representing decimal
	// e.g., 0x2024 should be treated as year 2024, not 8228
	HexAsDec bool
	// ListSize specifies the field name that contains the count for a list
	// When set, this field is a list whose length is determined at runtime
	ListSize string
	// ListTerminator specifies a byte sequence that terminates the list
	// When the unmarshaler encounters this sequence, it stops reading list elements
	ListTerminator []byte
}

// StructAnnotations represents all annotations for a struct
type StructAnnotations struct {
	// Fields contains annotations for each field
	Fields []FieldAnnotation
	// TotalSize is the total size of the struct in bytes
	TotalSize int
	// Name is the struct name
	Name string
}

// structCache caches parsed annotations per struct type to avoid repeated reflection work
var structCache sync.Map // map[reflect.Type]*StructAnnotations

// MarshalOptions controls marshaling behavior
type MarshalOptions struct {
	// IncludeReserved indicates whether to marshal reserved fields
	IncludeReserved bool
	// IncludeSkipped indicates whether to marshal skipped fields
	IncludeSkipped bool
	// OutputSize forces a specific output buffer size (0 means use calculated size)
	OutputSize int
}

// UnmarshalOptions controls unmarshaling behavior
type UnmarshalOptions struct {
	// IncludeReserved indicates whether to unmarshal reserved fields
	IncludeReserved bool
	// IncludeSkipped indicates whether to unmarshal skipped fields
	IncludeSkipped bool
}

// ParseTag parses the custom struct tag for annotations
// Tag format: `offset:"byte:4,bit:2,len:3,endian:be"`
func ParseTag(tag string, fieldType reflect.Type) (*FieldAnnotation, error) {
	annotation := &FieldAnnotation{
		Endianness: BigEndian, // Default to big-endian
		FieldType:  fieldType,
	}

	if tag == "" {
		return annotation, nil
	}

	parts := strings.Split(tag, ",")
	for _, part := range parts {
		kv := strings.Split(strings.TrimSpace(part), ":")
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "byte":
			offset, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid byte offset: %s", value)
			}
			annotation.ByteOffset = offset

		case "bit":
			offset, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid bit offset: %s", value)
			}
			if offset < 0 {
				return nil, fmt.Errorf("bit offset must be non-negative, got %d", offset)
			}
			// Store the absolute bit offset - this is needed for big-endian bitfield extraction
			annotation.BitOffset = offset
			// Also set ByteOffset for compatibility with code that expects it
			annotation.ByteOffset = offset / 8

		case "len":
			length, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid bit length: %s", value)
			}
			annotation.BitLength = length
			annotation.IsBitfield = true

		case "endian":
			switch strings.ToLower(value) {
			case "be", "big":
				annotation.Endianness = BigEndian
			case "le", "little":
				annotation.Endianness = LittleEndian
			case "host":
				annotation.Endianness = HostEndian
			default:
				return nil, fmt.Errorf("invalid endianness: %s", value)
			}

		case "skip":
			annotation.Skip = value == "true"

		case "reserved":
			annotation.Reserved = value == "true"

		case "hex_as_dec":
			annotation.HexAsDec = value == "true"

		case "list_size":
			annotation.ListSize = value

		case "list_terminator":
			// Parse hex string to bytes (e.g., "0x00000000" or "00000000")
			hexStr := strings.TrimPrefix(value, "0x")
			hexStr = strings.TrimPrefix(hexStr, "0X")

			// Ensure even number of characters
			if len(hexStr)%2 != 0 {
				hexStr = "0" + hexStr
			}

			terminator := make([]byte, len(hexStr)/2)
			for i := 0; i < len(terminator); i++ {
				b, err := strconv.ParseUint(hexStr[i*2:i*2+2], 16, 8)
				if err != nil {
					return nil, fmt.Errorf("invalid dynamic array terminator hex: %s", value)
				}
				terminator[i] = byte(b)
			}
			annotation.ListTerminator = terminator
		}
	}

	// Check if field is an array
	if fieldType.Kind() == reflect.Array {
		annotation.IsArray = true
		annotation.ArrayLength = fieldType.Len()
	}

	// Auto-detect bit length for bitfields if not specified
	// If bit offset is specified without explicit length, it's a bitfield
	if (annotation.IsBitfield || annotation.BitOffset > 0) && annotation.BitLength == 0 {
		annotation.IsBitfield = true
		// Determine default bit length based on field type
		switch fieldType.Kind() {
		case reflect.Bool:
			annotation.BitLength = 1
		case reflect.Uint8, reflect.Int8:
			annotation.BitLength = 8
		case reflect.Uint16, reflect.Int16:
			annotation.BitLength = 16
		case reflect.Uint32, reflect.Int32:
			annotation.BitLength = 32
		case reflect.Uint64, reflect.Int64:
			annotation.BitLength = 64
		default:
			return nil, fmt.Errorf("cannot auto-detect bit length for type %s", fieldType.Kind())
		}
	}

	return annotation, nil
}

// ParseStruct parses all annotations for a struct
func ParseStruct(structType reflect.Type) (*StructAnnotations, error) {
	if structType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct type, got %s", structType.Kind())
	}

	if cached, ok := structCache.Load(structType); ok {
		return cached.(*StructAnnotations), nil
	}

	annotations := &StructAnnotations{
		Name:   structType.Name(),
		Fields: make([]FieldAnnotation, 0, structType.NumField()),
	}

	maxOffset := 0

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}
		// Parse the offset tag
		tag := field.Tag.Get("offset")
		// Handle "-" tag which means skip this field (standard Go convention)
		if tag == "-" {
			continue
		}
		fieldAnnotation, err := ParseTag(tag, field.Type)
		if err != nil {
			return nil, fmt.Errorf("error parsing tag for field %s: %w", field.Name, err)
		}
		fieldAnnotation.FieldName = field.Name
		// If no explicit byte offset is set, calculate it based on previous fields
		// Note: bit: tags also set ByteOffset when converted from bit offsets
		if tag == "" || (!strings.Contains(tag, "byte:") && !strings.Contains(tag, "bit:")) {
			fieldAnnotation.ByteOffset = maxOffset
		}
		// Update max offset based on field size
		fieldSize := getFieldSize(field.Type)
		endOffset := fieldAnnotation.ByteOffset + fieldSize
		if endOffset > maxOffset {
			maxOffset = endOffset
		}
		annotations.Fields = append(annotations.Fields, *fieldAnnotation)
	}

	annotations.TotalSize = maxOffset
	// Cache the result for future calls
	structCache.Store(structType, annotations)
	return annotations, nil
}

// getFieldSize returns the size in bytes of a field type
func getFieldSize(fieldType reflect.Type) int {
	switch fieldType.Kind() {
	case reflect.Bool, reflect.Uint8, reflect.Int8:
		return 1
	case reflect.Uint16, reflect.Int16:
		return 2
	case reflect.Uint32, reflect.Int32, reflect.Float32:
		return 4
	case reflect.Uint64, reflect.Int64, reflect.Float64:
		return 8
	case reflect.Array:
		return fieldType.Len() * getFieldSize(fieldType.Elem())
	case reflect.Struct:
		// For structs, we need to calculate the total size recursively
		size := 0
		for i := 0; i < fieldType.NumField(); i++ {
			field := fieldType.Field(i)
			if field.PkgPath == "" { // Only exported fields
				size += getFieldSize(field.Type)
			}
		}
		return size
	default:
		// Default to pointer size for unknown types
		return 8
	}
}

// GetFieldByOffset finds a field annotation by its byte offset
func (sa *StructAnnotations) GetFieldByOffset(byteOffset int) *FieldAnnotation {
	for i := range sa.Fields {
		if sa.Fields[i].ByteOffset == byteOffset {
			return &sa.Fields[i]
		}
	}
	return nil
}

// GetFieldByName finds a field annotation by its field name
func (sa *StructAnnotations) GetFieldByName(name string) *FieldAnnotation {
	for i := range sa.Fields {
		if sa.Fields[i].FieldName == name {
			return &sa.Fields[i]
		}
	}
	return nil
}
