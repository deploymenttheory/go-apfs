package snapshot

import (
	"encoding/binary"
	"testing"
)

func TestSnapNameReader_ValidData(t *testing.T) {
	keyData := make([]byte, 10+7) // JKeyT (8) + NameLen (2) + "test\x00"
	valueData := make([]byte, 8)

	// Set up key
	binary.LittleEndian.PutUint64(keyData[0:8], 0xC0000000FFFFFFFF) // Type=12, ObjId=~0ULL
	binary.LittleEndian.PutUint16(keyData[8:10], 7)
	copy(keyData[10:], "test\x00")

	// Set up value
	binary.LittleEndian.PutUint64(valueData[0:8], 100)

	reader, err := NewSnapNameReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewSnapNameReader failed: %v", err)
	}

	if reader.Name() != "test" {
		t.Errorf("Name() = %q, want %q", reader.Name(), "test")
	}

	if reader.SnapXID() != 100 {
		t.Errorf("SnapXID() = %d, want 100", reader.SnapXID())
	}
}

func TestSnapNameReader_BigEndian(t *testing.T) {
	keyData := make([]byte, 10+6) // "snap\x00" is 5 bytes, so 6 with padding
	valueData := make([]byte, 8)

	binary.BigEndian.PutUint64(keyData[0:8], 0x0CFFFFFFFF000000)
	binary.BigEndian.PutUint16(keyData[8:10], 6)
	copy(keyData[10:], "snap\x00")

	binary.BigEndian.PutUint64(valueData[0:8], 200)

	reader, err := NewSnapNameReader(keyData, valueData, binary.BigEndian)
	if err != nil {
		t.Fatalf("NewSnapNameReader failed: %v", err)
	}

	if reader.Name() != "snap" {
		t.Errorf("Name() = %q, want %q", reader.Name(), "snap")
	}

	if reader.SnapXID() != 200 {
		t.Errorf("SnapXID() = %d, want 200", reader.SnapXID())
	}
}

func TestSnapNameReader_VariableLengthNames(t *testing.T) {
	tests := []struct {
		name    string
		nameStr string
	}{
		{"single char", "a\x00"},
		{"normal", "my-snapshot\x00"},
		{"long", "very-long-snapshot-name\x00"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keyData := make([]byte, 10+len(tc.nameStr))
			valueData := make([]byte, 8)

			binary.LittleEndian.PutUint64(keyData[0:8], 0xC000FFFFFFFFFF)
			binary.LittleEndian.PutUint16(keyData[8:10], uint16(len(tc.nameStr)))
			copy(keyData[10:], tc.nameStr)

			binary.LittleEndian.PutUint64(valueData[0:8], 300+uint64(len(tc.nameStr)))

			reader, err := NewSnapNameReader(keyData, valueData, binary.LittleEndian)
			if err != nil {
				t.Fatalf("NewSnapNameReader failed: %v", err)
			}

			expected := tc.nameStr[:len(tc.nameStr)-1]
			if reader.Name() != expected {
				t.Errorf("Name() = %q, want %q", reader.Name(), expected)
			}
		})
	}
}

func TestSnapNameReader_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		keySize   int
		valueSize int
		shouldErr bool
	}{
		{"Valid", 11, 8, false}, // Changed from 10 to 11 to include 1 byte of name
		{"Key too small", 5, 8, true},
		{"Value too small", 10, 4, true},
		{"Both too small", 5, 4, true},
		{"Key empty", 0, 8, true},
		{"Value empty", 10, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keyData := make([]byte, tc.keySize)
			valueData := make([]byte, tc.valueSize)

			if len(keyData) >= 10 {
				binary.LittleEndian.PutUint64(keyData[0:8], 0xC000FFFFFFFFFF)
				binary.LittleEndian.PutUint16(keyData[8:10], uint16(len(keyData)-10)) // NameLen = remaining bytes
				if len(keyData) > 10 {
					copy(keyData[10:], "\x00")
				}
			}
			if len(valueData) >= 8 {
				binary.LittleEndian.PutUint64(valueData[0:8], 100)
			}

			_, err := NewSnapNameReader(keyData, valueData, binary.LittleEndian)
			if tc.shouldErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSnapNameReader_InsufficientNameData(t *testing.T) {
	keyData := make([]byte, 10)
	valueData := make([]byte, 8)

	binary.LittleEndian.PutUint64(keyData[0:8], 0xC000FFFFFFFFFF)
	binary.LittleEndian.PutUint16(keyData[8:10], 50) // Claim 50 bytes but only have 0

	binary.LittleEndian.PutUint64(valueData[0:8], 100)

	_, err := NewSnapNameReader(keyData, valueData, binary.LittleEndian)
	if err == nil {
		t.Error("expected error for insufficient name data")
	}
}

func TestSnapNameReader_ZeroLengthName(t *testing.T) {
	keyData := make([]byte, 10)
	valueData := make([]byte, 8)

	binary.LittleEndian.PutUint64(keyData[0:8], 0xC000FFFFFFFFFF)
	binary.LittleEndian.PutUint16(keyData[8:10], 0)

	binary.LittleEndian.PutUint64(valueData[0:8], 100)

	reader, err := NewSnapNameReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewSnapNameReader failed: %v", err)
	}

	if reader.Name() != "" {
		t.Errorf("Name() = %q, want empty string", reader.Name())
	}
}

func TestSnapNameReader_MaxValues(t *testing.T) {
	keyData := make([]byte, 10+10)
	valueData := make([]byte, 8)

	binary.LittleEndian.PutUint64(keyData[0:8], ^uint64(0))
	binary.LittleEndian.PutUint16(keyData[8:10], 10)
	copy(keyData[10:], "maxname\x00")

	binary.LittleEndian.PutUint64(valueData[0:8], ^uint64(0))

	reader, err := NewSnapNameReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewSnapNameReader failed: %v", err)
	}

	if reader.SnapXID() != ^uint64(0) {
		t.Errorf("SnapXID() = %d, want max uint64", reader.SnapXID())
	}
}
