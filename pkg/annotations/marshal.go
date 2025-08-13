package annotations

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"unsafe"
)

// decToHexByte converts a byte from decimal representation to hex
// e.g., 27 -> 0x27
func decToHexByte(val uint8) uint8 {
	return ((val / 10) << 4) | (val % 10)
}

// decToHexUint16 converts a uint16 from decimal representation to hex
// e.g., 2024 -> 0x2024
func decToHexUint16(val uint16) uint16 {
	return uint16(((val/1000)%10)<<12) |
		uint16(((val/100)%10)<<8) |
		uint16(((val/10)%10)<<4) |
		uint16(val%10)
}

// decToHexUint32 converts a uint32 from decimal representation to hex
// e.g., 20240627 -> 0x20240627
func decToHexUint32(val uint32) uint32 {
	result := uint32(0)
	for i := 0; i < 8; i++ {
		digit := (val / uint32(pow10(i))) % 10
		result |= digit << (i * 4)
	}
	return result
}

func pow10(n int) int {
	result := 1
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
}

// Marshal marshals a struct to bytes using the field annotations
func Marshal(v interface{}, annotations *StructAnnotations) ([]byte, error) {
	return MarshalWithOptions(v, annotations, nil)
}

// MarshalWithOptions marshals a struct to bytes using the field annotations and options
func MarshalWithOptions(v interface{}, annotations *StructAnnotations, opts *MarshalOptions) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", rv.Kind())
	}

	// Default options
	if opts == nil {
		opts = &MarshalOptions{
			IncludeSkipped: false,
		}
	}

	// Allocate buffer for the result
	bufferSize := annotations.TotalSize
	if opts.OutputSize > 0 {
		bufferSize = opts.OutputSize
	}
	buffer := make([]byte, bufferSize)

	// Process each field
	for _, fieldAnnot := range annotations.Fields {
		// Check if we should skip this field based on options
		if fieldAnnot.Skip && !opts.IncludeSkipped {
			continue
		}
		// Reserved fields are treated the same as regular fields

		// Get field value
		field := rv.FieldByName(fieldAnnot.FieldName)
		if !field.IsValid() {
			continue
		}

		// Marshal the field
		if err := marshalField(buffer, field, &fieldAnnot); err != nil {
			return nil, fmt.Errorf("error marshaling field %s: %w", fieldAnnot.FieldName, err)
		}
	}

	return buffer, nil
}

// marshalField marshals a single field into the buffer
func marshalField(buffer []byte, fieldValue reflect.Value, annotation *FieldAnnotation) error {
	if annotation.IsBitfield {
		return marshalBitfield(buffer, fieldValue, annotation)
	}

	// Handle arrays
	if annotation.IsArray {
		return marshalArray(buffer, fieldValue, annotation)
	}

	// Handle lists (slices)
	if fieldValue.Kind() == reflect.Slice {
		return marshalList(buffer, fieldValue, annotation)
	}

	// Get byte order
	byteOrder := getByteOrder(annotation.Endianness)

	// Create a temporary buffer for the field
	fieldBuffer := &bytes.Buffer{}

	// Write the field value
	switch fieldValue.Kind() {
	case reflect.Bool:
		if fieldValue.Bool() {
			buffer[annotation.ByteOffset] = 1
		} else {
			buffer[annotation.ByteOffset] = 0
		}
		return nil
	case reflect.Uint8:
		val := uint8(fieldValue.Uint())
		if annotation.HexAsDec {
			val = decToHexByte(val)
		}
		buffer[annotation.ByteOffset] = val
		return nil
	case reflect.Uint16:
		val := uint16(fieldValue.Uint())
		if annotation.HexAsDec {
			val = decToHexUint16(val)
		}
		binary.Write(fieldBuffer, byteOrder, val)
	case reflect.Uint32:
		val := uint32(fieldValue.Uint())
		if annotation.HexAsDec {
			val = decToHexUint32(val)
		}
		binary.Write(fieldBuffer, byteOrder, val)
	case reflect.Uint64:
		binary.Write(fieldBuffer, byteOrder, fieldValue.Uint())
	case reflect.Int8:
		buffer[annotation.ByteOffset] = uint8(fieldValue.Int())
		return nil
	case reflect.Int16:
		binary.Write(fieldBuffer, byteOrder, int16(fieldValue.Int()))
	case reflect.Int32:
		binary.Write(fieldBuffer, byteOrder, int32(fieldValue.Int()))
	case reflect.Int64:
		binary.Write(fieldBuffer, byteOrder, fieldValue.Int())
	case reflect.Struct:
		// For nested structs, we need to recursively marshal
		nestedAnnotations, err := ParseStruct(fieldValue.Type())
		if err != nil {
			return err
		}
		nestedBytes, err := Marshal(fieldValue.Interface(), nestedAnnotations)
		if err != nil {
			return err
		}
		copy(buffer[annotation.ByteOffset:], nestedBytes)
		return nil
	default:
		return fmt.Errorf("unsupported field type: %s", fieldValue.Kind())
	}

	// Copy field bytes to buffer
	copy(buffer[annotation.ByteOffset:], fieldBuffer.Bytes())
	return nil
}

// marshalBitfield handles bitfield marshaling
func marshalBitfield(buffer []byte, fieldValue reflect.Value, annotation *FieldAnnotation) error {
	// Get the value as uint64
	var value uint64
	switch fieldValue.Kind() {
	case reflect.Bool:
		if fieldValue.Bool() {
			value = 1
		} else {
			value = 0
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value = fieldValue.Uint()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = uint64(fieldValue.Int())
	default:
		return fmt.Errorf("bitfield must be integer or bool type, got %s", fieldValue.Kind())
	}

	// Mask the value to fit the bit length
	mask := uint64((1 << annotation.BitLength) - 1)
	value &= mask

	// Special handling for big-endian bitfields
	if annotation.Endianness == BigEndian {
		// Calculate which bytes contain this bitfield
		// BitOffset is absolute from start of structure
		startByte := annotation.BitOffset / 8
		endBit := annotation.BitOffset + annotation.BitLength - 1
		endByte := endBit / 8
		numBytes := endByte - startByte + 1

		// Check bounds
		if endByte >= len(buffer) {
			return fmt.Errorf("buffer too small for bitfield at offset %d", annotation.BitOffset)
		}

		// Read existing bytes
		var existingValue uint64
		for i := 0; i < numBytes; i++ {
			existingValue = (existingValue << 8) | uint64(buffer[startByte+i])
		}

		// Calculate where our field starts within the bytes we read
		bitOffsetInValue := annotation.BitOffset - (startByte * 8)

		// Calculate unused bits at the end
		totalBitsInValue := numBytes * 8
		unusedBitsAtEnd := totalBitsInValue - bitOffsetInValue - annotation.BitLength

		// Create mask to clear the bits we're about to set
		mask := ((uint64(1) << annotation.BitLength) - 1)
		if unusedBitsAtEnd > 0 {
			mask = mask << uint(unusedBitsAtEnd)
		}
		existingValue &= ^mask

		// Position the new value
		if unusedBitsAtEnd > 0 {
			value = value << uint(unusedBitsAtEnd)
		}
		existingValue |= value

		// Write back the bytes
		for i := numBytes - 1; i >= 0; i-- {
			buffer[startByte+i] = uint8(existingValue & 0xFF)
			existingValue >>= 8
		}

		return nil
	}

	// Original implementation for little-endian bitfields
	bytePos := annotation.ByteOffset + annotation.BitOffset/8
	bitPos := annotation.BitOffset % 8
	endBit := bitPos + annotation.BitLength

	// Keep the special handling for aligned multi-byte fields if needed
	if false && annotation.Endianness == BigEndian && getFieldSize(annotation.FieldType) > 1 &&
		annotation.ByteOffset%4 == 0 && bytePos == annotation.ByteOffset &&
		endBit <= annotation.ByteOffset*8+32 {
		// Get the full field size based on the struct field type
		fieldSize := getFieldSize(annotation.FieldType)

		// Read the existing field value
		if annotation.ByteOffset+fieldSize > len(buffer) {
			return fmt.Errorf("buffer too small for field at offset %d (need %d bytes, have %d)", annotation.ByteOffset, annotation.ByteOffset+fieldSize, len(buffer))
		}

		// Convert the existing field bytes to a big-endian integer
		var fullValue uint64
		fieldData := buffer[annotation.ByteOffset : annotation.ByteOffset+fieldSize]
		reader := bytes.NewReader(fieldData)

		switch fieldSize {
		case 2:
			var val uint16
			binary.Read(reader, binary.BigEndian, &val)
			fullValue = uint64(val)
		case 4:
			var val uint32
			binary.Read(reader, binary.BigEndian, &val)
			fullValue = uint64(val)
		case 8:
			binary.Read(reader, binary.BigEndian, &fullValue)
		default:
			return fmt.Errorf("unsupported field size %d for big-endian bitfield", fieldSize)
		}

		// Clear the bits we're about to set
		clearMask := uint64((1<<annotation.BitLength)-1) << annotation.BitOffset
		fullValue &= ^clearMask

		// Set the new bits
		fullValue |= (value << annotation.BitOffset)

		// Write back the full value
		switch fieldSize {
		case 2:
			binary.BigEndian.PutUint16(buffer[annotation.ByteOffset:], uint16(fullValue))
		case 4:
			binary.BigEndian.PutUint32(buffer[annotation.ByteOffset:], uint32(fullValue))
		case 8:
			binary.BigEndian.PutUint64(buffer[annotation.ByteOffset:], fullValue)
		}

		return nil
	}

	// Original implementation for byte-level bitfields
	// Variables already declared above

	// Calculate how many bytes this bitfield spans
	numBytes := (endBit + 7) / 8

	// Read existing bytes, modify, and write back
	for i := 0; i < numBytes && bytePos+i < len(buffer); i++ {
		byteIdx := bytePos + i

		// Calculate bit range for this byte
		startBit := 0
		if i == 0 {
			startBit = bitPos
		}
		endBitInByte := 8
		if i == numBytes-1 && endBit%8 != 0 {
			endBitInByte = endBit % 8
		}
		bitsInByte := endBitInByte - startBit

		// Extract bits for this byte
		shift := i*8 - bitPos
		if shift < 0 {
			shift = 0
		}
		byteMask := uint8((1 << bitsInByte) - 1)
		byteValue := uint8((value >> shift) & uint64(byteMask))

		// Merge with existing byte
		existingMask := ^(byteMask << startBit)
		buffer[byteIdx] = (buffer[byteIdx] & existingMask) | (byteValue << startBit)
	}

	return nil
}

// marshalArray handles array marshaling
func marshalArray(buffer []byte, fieldValue reflect.Value, annotation *FieldAnnotation) error {
	elemSize := getFieldSize(annotation.FieldType.Elem())
	byteOrder := getByteOrder(annotation.Endianness)

	for i := 0; i < annotation.ArrayLength; i++ {
		elem := fieldValue.Index(i)
		offset := annotation.ByteOffset + i*elemSize

		// Create a temporary buffer for the element
		elemBuffer := &bytes.Buffer{}

		switch elem.Kind() {
		case reflect.Uint8:
			buffer[offset] = uint8(elem.Uint())
		case reflect.Uint16:
			binary.Write(elemBuffer, byteOrder, uint16(elem.Uint()))
			copy(buffer[offset:], elemBuffer.Bytes())
		case reflect.Uint32:
			binary.Write(elemBuffer, byteOrder, uint32(elem.Uint()))
			copy(buffer[offset:], elemBuffer.Bytes())
		case reflect.Uint64:
			binary.Write(elemBuffer, byteOrder, elem.Uint())
			copy(buffer[offset:], elemBuffer.Bytes())
		case reflect.Struct:
			// For struct array elements, recursively marshal
			nestedAnnotations, err := ParseStruct(elem.Type())
			if err != nil {
				return err
			}
			nestedBytes, err := Marshal(elem.Interface(), nestedAnnotations)
			if err != nil {
				return fmt.Errorf("error marshaling array element %d: %w", i, err)
			}
			copy(buffer[offset:], nestedBytes)
		default:
			return fmt.Errorf("unsupported array element type: %s", elem.Kind())
		}
	}

	return nil
}

// getByteOrder returns the appropriate byte order
func getByteOrder(endianness Endianness) binary.ByteOrder {
	switch endianness {
	case LittleEndian:
		return binary.LittleEndian
	case HostEndian:
		if isLittleEndian() {
			return binary.LittleEndian
		}
		return binary.BigEndian
	default: // BigEndian
		return binary.BigEndian
	}
}

// isLittleEndian checks if the host is little-endian
func isLittleEndian() bool {
	var i uint32 = 0x01020304
	return *(*byte)(unsafe.Pointer(&i)) == 0x04
}

// marshalList handles list (slice) marshaling
func marshalList(buffer []byte, fieldValue reflect.Value, annotation *FieldAnnotation) error {
	elemType := fieldValue.Type().Elem()
	elemSize := getFieldSize(elemType)
	byteOrder := getByteOrder(annotation.Endianness)
	numElements := fieldValue.Len()

	// Marshal each element
	for i := 0; i < numElements; i++ {
		elem := fieldValue.Index(i)
		offset := annotation.ByteOffset + (i * elemSize)

		// Ensure buffer is large enough
		if offset+elemSize > len(buffer) {
			// Lists might need to extend the buffer
			// This should be handled by the caller, for now return error
			return fmt.Errorf("buffer too small for list element %d at offset %d", i, offset)
		}

		switch elem.Kind() {
		case reflect.Uint8:
			buffer[offset] = uint8(elem.Uint())
		case reflect.Uint16:
			if byteOrder == binary.BigEndian {
				binary.BigEndian.PutUint16(buffer[offset:], uint16(elem.Uint()))
			} else {
				binary.LittleEndian.PutUint16(buffer[offset:], uint16(elem.Uint()))
			}
		case reflect.Uint32:
			if byteOrder == binary.BigEndian {
				binary.BigEndian.PutUint32(buffer[offset:], uint32(elem.Uint()))
			} else {
				binary.LittleEndian.PutUint32(buffer[offset:], uint32(elem.Uint()))
			}
		case reflect.Uint64:
			if byteOrder == binary.BigEndian {
				binary.BigEndian.PutUint64(buffer[offset:], elem.Uint())
			} else {
				binary.LittleEndian.PutUint64(buffer[offset:], elem.Uint())
			}
		case reflect.Struct:
			// For struct elements, recursively marshal
			nestedAnnotations, err := ParseStruct(elem.Type())
			if err != nil {
				return err
			}
			nestedData, err := MarshalWithOptions(elem.Interface(), nestedAnnotations, nil)
			if err != nil {
				return fmt.Errorf("error marshaling list element %d: %w", i, err)
			}
			copy(buffer[offset:], nestedData)
		default:
			return fmt.Errorf("unsupported list element type: %s", elem.Kind())
		}
	}

	// Add terminator if specified
	if len(annotation.ListTerminator) > 0 {
		terminatorOffset := annotation.ByteOffset + (numElements * elemSize)
		if terminatorOffset+len(annotation.ListTerminator) > len(buffer) {
			return fmt.Errorf("buffer too small for list terminator at offset %d", terminatorOffset)
		}
		copy(buffer[terminatorOffset:], annotation.ListTerminator)
	}

	return nil
}

// MarshalStruct marshals a struct to bytes by automatically parsing its annotations
func MarshalStruct(v interface{}) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("v must be a struct or pointer to struct")
	}

	annotations, err := ParseStruct(rv.Type())
	if err != nil {
		return nil, fmt.Errorf("failed to parse struct annotations: %w", err)
	}

	return Marshal(v, annotations)
}

// MarshalStructWithSize marshals a struct to bytes with a specific output size
func MarshalStructWithSize(v interface{}, outputSize int) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("v must be a struct or pointer to struct")
	}

	annotations, err := ParseStruct(rv.Type())
	if err != nil {
		return nil, fmt.Errorf("failed to parse struct annotations: %w", err)
	}

	opts := &MarshalOptions{
		OutputSize: outputSize,
	}

	return MarshalWithOptions(v, annotations, opts)
}

// MarshalWithOptionsStruct marshals a struct to bytes by parsing annotations from v automatically,
// applying the provided options.
func MarshalWithOptionsStruct(v interface{}, opts *MarshalOptions) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("v must be a struct or pointer to struct")
	}
	annotations, err := ParseStruct(rv.Type())
	if err != nil {
		return nil, fmt.Errorf("failed to parse struct annotations: %w", err)
	}
	return MarshalWithOptions(v, annotations, opts)
}
