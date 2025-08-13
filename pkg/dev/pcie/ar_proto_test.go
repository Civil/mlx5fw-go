package pcie

import "testing"

func TestPushGetBitsBE(t *testing.T) {
	b := make([]byte, 8)
	// Write 0b10101 over 5 bits at offset 3
	pushBitsBE(b, 3, 5, 0b10101)
	// Read it back
	v := getBitsBE(b, 3, 5)
	if v != 0b10101 {
		t.Fatalf("expected 0b10101, got %b", v)
	}
}

func TestPackMCQIHeader(t *testing.T) {
	pl := make([]byte, 0x94)
	packMCQIHeader(pl, 0x1, 0x7c)
	// Verify info_type at bit 91
	if it := getBitsBE(pl, 91, 5); it != 1 {
		t.Fatalf("info_type mismatch: got %d", it)
	}
	// info_size at bit 96 (32 bits)
	if is := getBitsBE(pl, 96, 32); is != 0x7c {
		t.Fatalf("info_size mismatch: got 0x%x", is)
	}
	// data_size at bit 176 (16 bits)
	if ds := getBitsBE(pl, 176, 16); ds != 0x7c {
		t.Fatalf("data_size mismatch: got 0x%x", ds)
	}
}

func TestPackMCQSHeader(t *testing.T) {
	pl := make([]byte, 0x10)
	packMCQSHeader(pl, 0x12, 0x345, 0x8)
	if comp := getBitsBE(pl, 16, 16); comp != 0x12 {
		t.Fatalf("component_index mismatch: got 0x%x", comp)
	}
	if devIdx := getBitsBE(pl, 4, 12); devIdx != 0x345 {
		t.Fatalf("device_index mismatch: got 0x%x", devIdx)
	}
	if devType := getBitsBE(pl, 120, 8); devType != 0x8 {
		t.Fatalf("device_type mismatch: got 0x%x", devType)
	}
}
