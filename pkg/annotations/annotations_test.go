package annotations

import (
	"reflect"
	"testing"
)

// Test struct with various field types and annotations
type TestStruct struct {
	// Basic fields
	Field1 uint32 `offset:"byte:0,endian:be"`
	Field2 uint16 `offset:"byte:4,endian:be"`
	Field3 uint8  `offset:"byte:6"`

	// Bitfield
	BitField1 uint8 `offset:"byte:7,bit:0,len:3"`
	BitField2 uint8 `offset:"byte:7,bit:3,len:5"`

	// Array
	Array1 [4]uint32 `offset:"byte:8,endian:be"`

	// Skip and reserved
	_        uint32 `offset:"byte:24,skip:true"`
	Reserved uint32 `offset:"byte:28,reserved:true"`
}

func TestParseAnnotations(t *testing.T) {
	annotations, err := ParseStruct(reflect.TypeOf(TestStruct{}))
	if err != nil {
		t.Fatalf("Failed to parse annotations: %v", err)
	}

	if annotations.Name != "TestStruct" {
		t.Errorf("Expected struct name TestStruct, got %s", annotations.Name)
	}

	// Check Field1 annotations
	field1 := annotations.GetFieldByName("Field1")
	if field1 == nil {
		t.Fatal("Field1 not found")
	}
	if field1.ByteOffset != 0 {
		t.Errorf("Expected Field1 byte offset 0, got %d", field1.ByteOffset)
	}
	if field1.Endianness != BigEndian {
		t.Error("Expected Field1 to be big-endian")
	}

	// Check bitfield annotations
	bitfield1 := annotations.GetFieldByName("BitField1")
	if bitfield1 == nil {
		t.Fatal("BitField1 not found")
	}
	if !bitfield1.IsBitfield {
		t.Error("Expected BitField1 to be marked as bitfield")
	}
	if bitfield1.BitOffset != 0 || bitfield1.BitLength != 3 {
		t.Errorf("Expected BitField1 bit offset 0, length 3, got offset %d, length %d",
			bitfield1.BitOffset, bitfield1.BitLength)
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	// Create test data
	original := TestStruct{
		Field1:    0x12345678,
		Field2:    0xABCD,
		Field3:    0xEF,
		BitField1: 0x5,  // 101 in binary (3 bits)
		BitField2: 0x15, // 10101 in binary (5 bits)
		Array1:    [4]uint32{0x11111111, 0x22222222, 0x33333333, 0x44444444},
	}

	// Parse annotations
	annotations, err := ParseStruct(reflect.TypeOf(TestStruct{}))
	if err != nil {
		t.Fatalf("Failed to parse annotations: %v", err)
	}

	// Marshal
	data, err := Marshal(original, annotations)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Check some specific bytes
	// Field1 (big-endian uint32 at offset 0)
	if data[0] != 0x12 || data[1] != 0x34 || data[2] != 0x56 || data[3] != 0x78 {
		t.Errorf("Field1 not marshaled correctly: %x", data[0:4])
	}

	// Field2 (big-endian uint16 at offset 4)
	if data[4] != 0xAB || data[5] != 0xCD {
		t.Errorf("Field2 not marshaled correctly: %x", data[4:6])
	}

	// Field3 (uint8 at offset 6)
	if data[6] != 0xEF {
		t.Errorf("Field3 not marshaled correctly: %x", data[6])
	}

	// Bitfields (both in byte at offset 7)
	// BitField1 (3 bits) = 5 = 101
	// BitField2 (5 bits) = 21 = 10101
	// Combined byte should be: 10101101 = 0xAD
	expectedBitfield := uint8((original.BitField2 << 3) | original.BitField1)
	if data[7] != expectedBitfield {
		t.Errorf("Bitfields not marshaled correctly: expected %x, got %x", expectedBitfield, data[7])
	}

	// Unmarshal
	var unmarshaled TestStruct
	err = Unmarshal(data, &unmarshaled, annotations)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Compare
	if unmarshaled.Field1 != original.Field1 {
		t.Errorf("Field1 mismatch: expected %x, got %x", original.Field1, unmarshaled.Field1)
	}
	if unmarshaled.Field2 != original.Field2 {
		t.Errorf("Field2 mismatch: expected %x, got %x", original.Field2, unmarshaled.Field2)
	}
	if unmarshaled.Field3 != original.Field3 {
		t.Errorf("Field3 mismatch: expected %x, got %x", original.Field3, unmarshaled.Field3)
	}
	if unmarshaled.BitField1 != original.BitField1 {
		t.Errorf("BitField1 mismatch: expected %x, got %x", original.BitField1, unmarshaled.BitField1)
	}
	if unmarshaled.BitField2 != original.BitField2 {
		t.Errorf("BitField2 mismatch: expected %x, got %x", original.BitField2, unmarshaled.BitField2)
	}
	for i := 0; i < len(original.Array1); i++ {
		if unmarshaled.Array1[i] != original.Array1[i] {
			t.Errorf("Array1[%d] mismatch: expected %x, got %x", i, original.Array1[i], unmarshaled.Array1[i])
		}
	}
}

func TestBitfieldEdgeCases(t *testing.T) {
	type BitfieldTest struct {
		// Bitfield spanning multiple bytes
		Field1 uint16 `offset:"byte:0,bit:6,len:10,endian:be"`
		// Bitfield at byte boundary
		Field2 uint8 `offset:"byte:2,bit:0,len:8"`
	}

	original := BitfieldTest{
		Field1: 0x3FF, // All 10 bits set
		Field2: 0xAA,
	}

	annotations, err := ParseStruct(reflect.TypeOf(BitfieldTest{}))
	if err != nil {
		t.Fatalf("Failed to parse annotations: %v", err)
	}

	data, err := Marshal(original, annotations)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var unmarshaled BitfieldTest
	err = Unmarshal(data, &unmarshaled, annotations)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if unmarshaled.Field1 != original.Field1 {
		t.Errorf("Field1 mismatch: expected %x, got %x", original.Field1, unmarshaled.Field1)
	}
	if unmarshaled.Field2 != original.Field2 {
		t.Errorf("Field2 mismatch: expected %x, got %x", original.Field2, unmarshaled.Field2)
	}
}
