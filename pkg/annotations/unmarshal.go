package annotations

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

// hexToDecByte converts a byte from hex representation to decimal
// e.g., 0x27 -> 27
func hexToDecByte(b uint8) uint8 {
	return ((b >> 4) & 0xF) * 10 + (b & 0xF)
}

// hexToDecUint16 converts a uint16 from hex representation to decimal
// e.g., 0x2024 -> 2024
func hexToDecUint16(val uint16) uint16 {
	return uint16(((val >> 12) & 0xF) * 1000 +
		((val >> 8) & 0xF) * 100 +
		((val >> 4) & 0xF) * 10 +
		(val & 0xF))
}

// hexToDecUint32 converts a uint32 from hex representation to decimal
// e.g., 0x20240627 -> 20240627
func hexToDecUint32(val uint32) uint32 {
	result := uint32(0)
	multiplier := uint32(1)
	for i := 0; i < 8; i++ {
		digit := (val >> (i * 4)) & 0xF
		result += digit * multiplier
		multiplier *= 10
	}
	return result
}

// Unmarshal unmarshals bytes into a struct using the field annotations
func Unmarshal(data []byte, v interface{}, annotations *StructAnnotations) error {
	return UnmarshalWithOptions(data, v, annotations, nil)
}

// UnmarshalWithOptions unmarshals bytes into a struct using the field annotations and options
func UnmarshalWithOptions(data []byte, v interface{}, annotations *StructAnnotations, opts *UnmarshalOptions) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("v must be a pointer to struct")
	}

	rv = rv.Elem()

	// Default options
	if opts == nil {
		opts = &UnmarshalOptions{
			IncludeSkipped: false,
		}
	}

	// Process each field
	for _, fieldAnnot := range annotations.Fields {
		// Check if we should skip this field based on options
		if fieldAnnot.Skip && !opts.IncludeSkipped {
			continue
		}
		// Reserved fields are treated the same as regular fields

		// Get field value
		field := rv.FieldByName(fieldAnnot.FieldName)
		if !field.IsValid() || !field.CanSet() {
			continue
		}

		// Unmarshal the field
		if err := unmarshalField(data, field, &fieldAnnot, opts, rv); err != nil {
			return fmt.Errorf("error unmarshaling field %s: %w", fieldAnnot.FieldName, err)
		}
	}

	return nil
}

// unmarshalField unmarshals a single field from the buffer
func unmarshalField(data []byte, fieldValue reflect.Value, annotation *FieldAnnotation, opts *UnmarshalOptions, structValue reflect.Value) error {
	if annotation.IsBitfield {
		return unmarshalBitfield(data, fieldValue, annotation)
	}

	// Handle arrays and slices
	if annotation.IsArray {
		return unmarshalArray(data, fieldValue, annotation, opts)
	}
	
	// Handle lists (slices)
	if fieldValue.Kind() == reflect.Slice {
		return unmarshalList(data, fieldValue, annotation, opts, structValue)
	}

	// Check bounds
	fieldSize := getFieldSize(annotation.FieldType)
	if annotation.ByteOffset+fieldSize > len(data) {
		return fmt.Errorf("data too short for field at offset %d", annotation.ByteOffset)
	}

	// Get byte order
	byteOrder := getByteOrder(annotation.Endianness)

	// Create a reader for the field data
	fieldData := data[annotation.ByteOffset : annotation.ByteOffset+fieldSize]
	reader := bytes.NewReader(fieldData)

	// Read the field value
	switch fieldValue.Kind() {
	case reflect.Bool:
		fieldValue.SetBool(fieldData[0] != 0)
		return nil
	case reflect.Uint8:
		val := fieldData[0]
		if annotation.HexAsDec {
			val = hexToDecByte(val)
		}
		fieldValue.SetUint(uint64(val))
		return nil
	case reflect.Uint16:
		var val uint16
		binary.Read(reader, byteOrder, &val)
		if annotation.HexAsDec {
			val = hexToDecUint16(val)
		}
		fieldValue.SetUint(uint64(val))
	case reflect.Uint32:
		var val uint32
		binary.Read(reader, byteOrder, &val)
		if annotation.HexAsDec {
			val = hexToDecUint32(val)
		}
		fieldValue.SetUint(uint64(val))
	case reflect.Uint64:
		var val uint64
		binary.Read(reader, byteOrder, &val)
		fieldValue.SetUint(val)
	case reflect.Int8:
		fieldValue.SetInt(int64(int8(fieldData[0])))
		return nil
	case reflect.Int16:
		var val int16
		binary.Read(reader, byteOrder, &val)
		fieldValue.SetInt(int64(val))
	case reflect.Int32:
		var val int32
		binary.Read(reader, byteOrder, &val)
		fieldValue.SetInt(int64(val))
	case reflect.Int64:
		var val int64
		binary.Read(reader, byteOrder, &val)
		fieldValue.SetInt(val)
	case reflect.Struct:
		// For nested structs, we need to recursively unmarshal
		nestedAnnotations, err := ParseStruct(fieldValue.Type())
		if err != nil {
			return err
		}
		return UnmarshalWithOptions(fieldData, fieldValue.Addr().Interface(), nestedAnnotations, opts)
	default:
		return fmt.Errorf("unsupported field type: %s", fieldValue.Kind())
	}

	return nil
}

// unmarshalBitfield handles bitfield unmarshaling
func unmarshalBitfield(data []byte, fieldValue reflect.Value, annotation *FieldAnnotation) error {
	// Calculate byte and bit positions
	bytePos := annotation.ByteOffset + annotation.BitOffset/8
	bitPos := annotation.BitOffset % 8

	// Calculate how many bytes this bitfield spans
	endBit := bitPos + annotation.BitLength
	numBytes := (endBit + 7) / 8

	// For big-endian bitfields, we need to handle them differently
	// Big-endian bit numbering: bit 0 is the MSB of the first byte
	if annotation.Endianness == BigEndian {
		// Calculate which bytes contain this bitfield
		// BitOffset is absolute from start of structure
		startByte := annotation.BitOffset / 8
		endBit := annotation.BitOffset + annotation.BitLength - 1
		endByte := endBit / 8
		numBytes := endByte - startByte + 1
		
		// Check bounds
		if endByte >= len(data) {
			return fmt.Errorf("data too short for bitfield at offset %d", annotation.BitOffset)
		}
		
		// Read all bytes that contain our bitfield
		var value uint64
		for i := 0; i < numBytes; i++ {
			value = (value << 8) | uint64(data[startByte+i])
		}
		
		// For big-endian bitfields, we have the value as a big-endian integer
		// We need to extract the specific bits requested
		
		// Calculate bit positions within the value we read
		// startByte is the byte where our field starts
		// We read numBytes starting from startByte
		// Our field starts at bit (BitOffset % 8) within the first byte
		startBitInValue := annotation.BitOffset - (startByte * 8)
		
		// Total bits in the value we read
		totalBitsInValue := numBytes * 8
		
		// How many bits from the right edge do we need to shift?
		// If we have 32 bits total and want bits 0-28 (29 bits), we shift right by 3
		bitsFromRight := totalBitsInValue - startBitInValue - annotation.BitLength
		
		// Shift right to align our field to the LSB position
		if bitsFromRight > 0 {
			value = value >> uint(bitsFromRight)
		}
		
		// Mask to get only the bits we want
		mask := (uint64(1) << annotation.BitLength) - 1
		value = value & mask
		
		// Set the field value
		switch fieldValue.Kind() {
		case reflect.Bool:
			fieldValue.SetBool(value != 0)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fieldValue.SetUint(value)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			// Handle sign extension for signed types
			if annotation.BitLength < 64 && (value&(1<<(annotation.BitLength-1))) != 0 {
				// Sign extend
				value |= ^uint64(0) << annotation.BitLength
			}
			fieldValue.SetInt(int64(value))
		default:
			return fmt.Errorf("bitfield must be integer or bool type, got %s", fieldValue.Kind())
		}
		return nil
	}
	
	// Little-endian bit extraction (existing code)
	var value uint64
	for i := 0; i < numBytes && bytePos+i < len(data); i++ {
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

		// Extract bits from this byte
		byteMask := uint8((1 << bitsInByte) - 1)
		byteValue := (data[byteIdx] >> startBit) & byteMask

		// Shift and add to value
		shift := i * 8 - bitPos
		if shift < 0 {
			shift = 0
		}
		value |= uint64(byteValue) << shift
	}

	// Mask to the correct bit length
	mask := uint64((1 << annotation.BitLength) - 1)
	value &= mask

	// Set the field value
	switch fieldValue.Kind() {
	case reflect.Bool:
		fieldValue.SetBool(value != 0)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fieldValue.SetUint(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Handle sign extension for signed types
		if annotation.BitLength < 64 && (value&(1<<(annotation.BitLength-1))) != 0 {
			// Sign extend
			value |= ^uint64(0) << annotation.BitLength
		}
		fieldValue.SetInt(int64(value))
	default:
		return fmt.Errorf("bitfield must be integer or bool type, got %s", fieldValue.Kind())
	}

	return nil
}

// unmarshalArray handles array unmarshaling
func unmarshalArray(data []byte, fieldValue reflect.Value, annotation *FieldAnnotation, opts *UnmarshalOptions) error {
	elemSize := getFieldSize(annotation.FieldType.Elem())
	byteOrder := getByteOrder(annotation.Endianness)

	for i := 0; i < annotation.ArrayLength; i++ {
		elem := fieldValue.Index(i)
		offset := annotation.ByteOffset + i*elemSize

		if offset+elemSize > len(data) {
			return fmt.Errorf("data too short for array element %d at offset %d", i, offset)
		}

		elemData := data[offset : offset+elemSize]
		reader := bytes.NewReader(elemData)

		switch elem.Kind() {
		case reflect.Uint8:
			elem.SetUint(uint64(elemData[0]))
		case reflect.Uint16:
			var val uint16
			binary.Read(reader, byteOrder, &val)
			elem.SetUint(uint64(val))
		case reflect.Uint32:
			var val uint32
			binary.Read(reader, byteOrder, &val)
			elem.SetUint(uint64(val))
		case reflect.Uint64:
			var val uint64
			binary.Read(reader, byteOrder, &val)
			elem.SetUint(val)
		case reflect.Struct:
			// For struct array elements, recursively unmarshal
			nestedAnnotations, err := ParseStruct(elem.Type())
			if err != nil {
				return err
			}
			if err := UnmarshalWithOptions(elemData, elem.Addr().Interface(), nestedAnnotations, opts); err != nil {
				return fmt.Errorf("error unmarshaling array element %d: %w", i, err)
			}
		default:
			return fmt.Errorf("unsupported array element type: %s", elem.Kind())
		}
	}

	return nil
}

// unmarshalList handles list (slice) unmarshaling
func unmarshalList(data []byte, fieldValue reflect.Value, annotation *FieldAnnotation, opts *UnmarshalOptions, structValue reflect.Value) error {
	elemType := fieldValue.Type().Elem()
	elemSize := getFieldSize(elemType)
	byteOrder := getByteOrder(annotation.Endianness)
	
	// Determine the number of elements
	var numElements int
	
	if annotation.ListSize != "" {
		// Count-based list
		// Find the field that contains the count
		countField := structValue.FieldByName(annotation.ListSize)
		if !countField.IsValid() {
			return fmt.Errorf("list size field %s not found", annotation.ListSize)
		}
		
		switch countField.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			numElements = int(countField.Uint())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			numElements = int(countField.Int())
		default:
			return fmt.Errorf("list size field must be numeric, got %s", countField.Kind())
		}
	} else if len(annotation.ListTerminator) > 0 {
		// Terminator-based list
		// Count elements until we find the terminator
		offset := annotation.ByteOffset
		numElements = 0
		
		for offset+elemSize <= len(data) {
			// Check if current element matches terminator
			if bytes.Equal(data[offset:offset+elemSize], annotation.ListTerminator) {
				break
			}
			numElements++
			offset += elemSize
		}
	} else {
		// Use remaining data
		remainingBytes := len(data) - annotation.ByteOffset
		numElements = remainingBytes / elemSize
	}
	
	// Create slice with appropriate capacity
	slice := reflect.MakeSlice(fieldValue.Type(), numElements, numElements)
	
	// Unmarshal each element
	for i := 0; i < numElements; i++ {
		elem := slice.Index(i)
		offset := annotation.ByteOffset + (i * elemSize)
		
		if offset+elemSize > len(data) {
			return fmt.Errorf("data too short for list element %d at offset %d", i, offset)
		}
		
		elemData := data[offset : offset+elemSize]
		reader := bytes.NewReader(elemData)
		
		switch elem.Kind() {
		case reflect.Uint8:
			elem.SetUint(uint64(elemData[0]))
		case reflect.Uint16:
			var val uint16
			binary.Read(reader, byteOrder, &val)
			elem.SetUint(uint64(val))
		case reflect.Uint32:
			var val uint32
			binary.Read(reader, byteOrder, &val)
			elem.SetUint(uint64(val))
		case reflect.Uint64:
			var val uint64
			binary.Read(reader, byteOrder, &val)
			elem.SetUint(val)
		case reflect.Struct:
			// For struct elements, recursively unmarshal
			nestedAnnotations, err := ParseStruct(elem.Type())
			if err != nil {
				return err
			}
			if err := UnmarshalWithOptions(elemData, elem.Addr().Interface(), nestedAnnotations, opts); err != nil {
				return fmt.Errorf("error unmarshaling list element %d: %w", i, err)
			}
		default:
			return fmt.Errorf("unsupported list element type: %s", elem.Kind())
		}
	}
	
	fieldValue.Set(slice)
	return nil
}

// UnmarshalStruct unmarshals bytes into a struct by automatically parsing its annotations
func UnmarshalStruct(data []byte, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("v must be a pointer to struct")
	}
	
	annotations, err := ParseStruct(rv.Elem().Type())
	if err != nil {
		return fmt.Errorf("failed to parse struct annotations: %w", err)
	}
	
	return Unmarshal(data, v, annotations)
}