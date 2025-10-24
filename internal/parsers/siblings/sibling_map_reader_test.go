package siblings

import (
	"encoding/binary"
	"testing"
)

func TestSiblingMapReader_ValidData(t *testing.T) {
	keyData := make([]byte, 8)
	valueData := make([]byte, 8)

	// Set up key with sibling ID in ObjIdAndType (lower 48 bits)
	binary.LittleEndian.PutUint64(keyData[0:8], 0xC00000000000ABCD) // Type=12 (sibling map), ObjId=0xABCD

	// Set up value with FileId
	binary.LittleEndian.PutUint64(valueData[0:8], 0x1000000000000FFF) // Target inode

	reader, err := NewSiblingMapReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewSiblingMapReader failed: %v", err)
	}

	if reader.FileID() != 0x1000000000000FFF {
		t.Errorf("FileID() = %#x, want %#x", reader.FileID(), 0x1000000000000FFF)
	}
}

func TestSiblingMapReader_BigEndian(t *testing.T) {
	keyData := make([]byte, 8)
	valueData := make([]byte, 8)

	binary.BigEndian.PutUint64(keyData[0:8], 0x0C0000000000DCBA)
	binary.BigEndian.PutUint64(valueData[0:8], 0xFFF0000000000100)

	reader, err := NewSiblingMapReader(keyData, valueData, binary.BigEndian)
	if err != nil {
		t.Fatalf("NewSiblingMapReader failed: %v", err)
	}

	if reader.FileID() != 0xFFF0000000000100 {
		t.Errorf("FileID() = %#x, want %#x", reader.FileID(), uint64(0xFFF0000000000100))
	}
}

func TestSiblingMapReader_SiblingIDExtraction(t *testing.T) {
	tests := []struct {
		name       string
		keyValue   uint64
		expectedID uint64
	}{
		{"Simple ID", 0xC000000000000001, 1},
		{"Large ID", 0xC000FFFFFFFFFFFF, 0xFFFFFFFFFFFF}, // 48-bit mask
		{"Zero ID", 0xC000000000000000, 0},
		{"Max 48-bit", 0xC0000000FFFFFFFF, 0xFFFFFFFF},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keyData := make([]byte, 8)
			valueData := make([]byte, 8)

			binary.LittleEndian.PutUint64(keyData[0:8], tc.keyValue)
			binary.LittleEndian.PutUint64(valueData[0:8], 12345)

			reader, err := NewSiblingMapReader(keyData, valueData, binary.LittleEndian)
			if err != nil {
				t.Fatalf("NewSiblingMapReader failed: %v", err)
			}

			if reader.SiblingID() != tc.expectedID {
				t.Errorf("SiblingID() = %#x, want %#x", reader.SiblingID(), tc.expectedID)
			}
		})
	}
}

func TestSiblingMapReader_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		keySize   int
		valueSize int
		shouldErr bool
	}{
		{"Valid", 8, 8, false},
		{"Key too small", 4, 8, true},
		{"Value too small", 8, 4, true},
		{"Both too small", 4, 4, true},
		{"Key empty", 0, 8, true},
		{"Value empty", 8, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keyData := make([]byte, tc.keySize)
			valueData := make([]byte, tc.valueSize)

			// Fill with valid data if size permits
			if len(keyData) >= 8 {
				binary.LittleEndian.PutUint64(keyData[0:8], 0xC000000000000001)
			}
			if len(valueData) >= 8 {
				binary.LittleEndian.PutUint64(valueData[0:8], 999)
			}

			_, err := NewSiblingMapReader(keyData, valueData, binary.LittleEndian)
			if tc.shouldErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSiblingMapReader_MaxValues(t *testing.T) {
	keyData := make([]byte, 8)
	valueData := make([]byte, 8)

	// All bits set
	binary.LittleEndian.PutUint64(keyData[0:8], ^uint64(0))
	binary.LittleEndian.PutUint64(valueData[0:8], ^uint64(0))

	reader, err := NewSiblingMapReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewSiblingMapReader failed: %v", err)
	}

	// FileID should be all bits set
	if reader.FileID() != ^uint64(0) {
		t.Errorf("FileID() = %#x, want %#x", reader.FileID(), ^uint64(0))
	}
}

func TestSiblingMapReader_ZeroValues(t *testing.T) {
	keyData := make([]byte, 8)
	valueData := make([]byte, 8)

	// All zeros
	binary.LittleEndian.PutUint64(keyData[0:8], 0)
	binary.LittleEndian.PutUint64(valueData[0:8], 0)

	reader, err := NewSiblingMapReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewSiblingMapReader failed: %v", err)
	}

	if reader.SiblingID() != 0 {
		t.Errorf("SiblingID() = %d, want 0", reader.SiblingID())
	}

	if reader.FileID() != 0 {
		t.Errorf("FileID() = %d, want 0", reader.FileID())
	}
}

func TestSiblingMapReader_ConsistentEndianness(t *testing.T) {
	// Test that different endianness produces different results
	testValue := uint64(0x0102030405060708)

	// Little endian
	leKeyData := make([]byte, 8)
	leValueData := make([]byte, 8)
	binary.LittleEndian.PutUint64(leKeyData[0:8], 0xC000000000000001)
	binary.LittleEndian.PutUint64(leValueData[0:8], testValue)

	leReader, err := NewSiblingMapReader(leKeyData, leValueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("Little-endian reader failed: %v", err)
	}

	// Big endian
	beKeyData := make([]byte, 8)
	beValueData := make([]byte, 8)
	binary.BigEndian.PutUint64(beKeyData[0:8], 0x0C0000000000001)
	binary.BigEndian.PutUint64(beValueData[0:8], testValue)

	beReader, err := NewSiblingMapReader(beKeyData, beValueData, binary.BigEndian)
	if err != nil {
		t.Fatalf("Big-endian reader failed: %v", err)
	}

	// Both should read the same value despite different byte order
	if leReader.FileID() != beReader.FileID() {
		t.Errorf("Endian mismatch: LE=%#x, BE=%#x", leReader.FileID(), beReader.FileID())
	}
}
