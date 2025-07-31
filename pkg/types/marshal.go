package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

// MarshalLE marshals a struct to binary little-endian format
func MarshalLE(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}
	return buf.Bytes(), nil
}

// MarshalBE marshals a struct to binary big-endian format
func MarshalBE(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}
	return buf.Bytes(), nil
}

// marshalStruct is a helper that properly handles struct marshaling with correct endianness
func marshalStruct(v interface{}, order binary.ByteOrder) ([]byte, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", val.Kind())
	}
	
	buf := new(bytes.Buffer)
	
	// Walk through struct fields
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		
		// Handle different field types
		switch field.Kind() {
		case reflect.Array:
			// For byte arrays, write directly
			if field.Type().Elem().Kind() == reflect.Uint8 {
				for j := 0; j < field.Len(); j++ {
					if err := binary.Write(buf, order, uint8(field.Index(j).Uint())); err != nil {
						return nil, err
					}
				}
			} else {
				// For other arrays, write each element
				for j := 0; j < field.Len(); j++ {
					if err := binary.Write(buf, order, field.Index(j).Interface()); err != nil {
						return nil, err
					}
				}
			}
		default:
			// For other types, use binary.Write
			if err := binary.Write(buf, order, field.Interface()); err != nil {
				return nil, err
			}
		}
	}
	
	return buf.Bytes(), nil
}