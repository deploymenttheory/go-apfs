package siblings

import (
	"encoding/binary"
	"testing"
)

func TestSiblingLinkReader_ValidData(t *testing.T) {
	// Create test data
	keyData := make([]byte, 16)
	valueData := make([]byte, 10+10) // ParentId + NameLen + "test.txt\0"

	// Set up key
	binary.LittleEndian.PutUint64(keyData[0:8], 0x1000000000000005) // ObjIdAndType with type=5 (sibling link)
	binary.LittleEndian.PutUint64(keyData[8:16], 12345)             // SiblingId

	// Set up value
	binary.LittleEndian.PutUint64(valueData[0:8], 999) // ParentId
	binary.LittleEndian.PutUint16(valueData[8:10], 10) // NameLen (includes null terminator)
	copy(valueData[10:], "test.txt\x00")

	reader, err := NewSiblingLinkReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewSiblingLinkReader failed: %v", err)
	}

	if reader.SiblingID() != 12345 {
		t.Errorf("SiblingID() = %d, want 12345", reader.SiblingID())
	}

	if reader.ParentDirectoryID() != 999 {
		t.Errorf("ParentDirectoryID() = %d, want 999", reader.ParentDirectoryID())
	}

	if reader.Name() != "test.txt" {
		t.Errorf("Name() = %q, want %q", reader.Name(), "test.txt")
	}
}

func TestSiblingLinkReader_BigEndian(t *testing.T) {
	keyData := make([]byte, 16)
	valueData := make([]byte, 10+9) // ParentId + NameLen + "file.c\0"

	// Set up key (big endian)
	binary.BigEndian.PutUint64(keyData[0:8], 0x0500000000001000)
	binary.BigEndian.PutUint64(keyData[8:16], 54321)

	// Set up value (big endian)
	binary.BigEndian.PutUint64(valueData[0:8], 888)
	binary.BigEndian.PutUint16(valueData[8:10], 9)
	copy(valueData[10:], "file.c\x00")

	reader, err := NewSiblingLinkReader(keyData, valueData, binary.BigEndian)
	if err != nil {
		t.Fatalf("NewSiblingLinkReader failed: %v", err)
	}

	if reader.SiblingID() != 54321 {
		t.Errorf("SiblingID() = %d, want 54321", reader.SiblingID())
	}

	if reader.Name() != "file.c" {
		t.Errorf("Name() = %q, want %q", reader.Name(), "file.c")
	}
}

func TestSiblingLinkReader_InodeNumber(t *testing.T) {
	keyData := make([]byte, 16)
	valueData := make([]byte, 10+1)

	// ObjId is extracted using ObjIdMask (0x0000FFFFFFFFFFFF), create a simple value
	objIdAndType := uint64(0x5000000000000005) // Type=5 in upper bits, ObjId=5 in lower 48 bits
	binary.LittleEndian.PutUint64(keyData[0:8], objIdAndType)
	binary.LittleEndian.PutUint64(keyData[8:16], 111)

	binary.LittleEndian.PutUint64(valueData[0:8], 777)
	binary.LittleEndian.PutUint16(valueData[8:10], 1)
	copy(valueData[10:], "\x00")

	reader, err := NewSiblingLinkReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewSiblingLinkReader failed: %v", err)
	}

	inodeNum := reader.InodeNumber()
	if inodeNum != 5 {
		t.Errorf("InodeNumber() = %d, want 5", inodeNum)
	}
}

func TestSiblingLinkReader_VariableLengthNames(t *testing.T) {
	tests := []struct {
		name      string
		nameBytes string
		sibling   uint64
		parentId  uint64
	}{
		{"simple", "a\x00", 100, 10},
		{"long", "this-is-a-very-long-filename-test.txt\x00", 200, 20},
		{"special", "file@#$.txt\x00", 300, 30},
		{"unicode", "файл.txt\x00", 400, 40},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keyData := make([]byte, 16)
			valueData := make([]byte, 10+len(tc.nameBytes))

			binary.LittleEndian.PutUint64(keyData[0:8], 0x5000000000001) // Type=5
			binary.LittleEndian.PutUint64(keyData[8:16], tc.sibling)

			binary.LittleEndian.PutUint64(valueData[0:8], tc.parentId)
			binary.LittleEndian.PutUint16(valueData[8:10], uint16(len(tc.nameBytes)))
			copy(valueData[10:], tc.nameBytes)

			reader, err := NewSiblingLinkReader(keyData, valueData, binary.LittleEndian)
			if err != nil {
				t.Fatalf("NewSiblingLinkReader failed: %v", err)
			}

			if reader.SiblingID() != tc.sibling {
				t.Errorf("SiblingID() = %d, want %d", reader.SiblingID(), tc.sibling)
			}

			if reader.ParentDirectoryID() != tc.parentId {
				t.Errorf("ParentDirectoryID() = %d, want %d", reader.ParentDirectoryID(), tc.parentId)
			}
		})
	}
}

func TestSiblingLinkReader_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		keySize   int
		valueSize int
		shouldErr bool
	}{
		{"Valid key and value", 16, 11, false}, // Changed from 10 to 11 to include name byte
		{"Key too small", 8, 10, true},
		{"Value too small", 16, 5, true},
		{"Both too small", 8, 5, true},
		{"Key empty", 0, 10, true},
		{"Value empty", 16, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keyData := make([]byte, tc.keySize)
			valueData := make([]byte, tc.valueSize)

			// Fill with valid data if size permits
			if len(keyData) >= 16 {
				binary.LittleEndian.PutUint64(keyData[0:8], 0x5000000000001)
				binary.LittleEndian.PutUint64(keyData[8:16], 12345)
			}
			if len(valueData) >= 10 {
				binary.LittleEndian.PutUint64(valueData[0:8], 999)
				binary.LittleEndian.PutUint16(valueData[8:10], uint16(len(valueData)-10)) // NameLen = remaining bytes
				if len(valueData) > 10 {
					copy(valueData[10:], "\x00")
				}
			}

			_, err := NewSiblingLinkReader(keyData, valueData, binary.LittleEndian)
			if tc.shouldErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSiblingLinkReader_InsufficientNameData(t *testing.T) {
	keyData := make([]byte, 16)
	valueData := make([]byte, 10)

	binary.LittleEndian.PutUint64(keyData[0:8], 0x5000000000001)
	binary.LittleEndian.PutUint64(keyData[8:16], 12345)

	// Claim 20 bytes of name data but only provide 10 total bytes
	binary.LittleEndian.PutUint64(valueData[0:8], 999)
	binary.LittleEndian.PutUint16(valueData[8:10], 20)

	_, err := NewSiblingLinkReader(keyData, valueData, binary.LittleEndian)
	if err == nil {
		t.Error("expected error for insufficient name data")
	}
}

func TestSiblingLinkReader_ZeroLengthName(t *testing.T) {
	keyData := make([]byte, 16)
	valueData := make([]byte, 10)

	binary.LittleEndian.PutUint64(keyData[0:8], 0x5000000000001)
	binary.LittleEndian.PutUint64(keyData[8:16], 12345)

	binary.LittleEndian.PutUint64(valueData[0:8], 999)
	binary.LittleEndian.PutUint16(valueData[8:10], 0)

	reader, err := NewSiblingLinkReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewSiblingLinkReader failed: %v", err)
	}

	if reader.Name() != "" {
		t.Errorf("Name() = %q, want empty string", reader.Name())
	}
}

func TestSiblingLinkReader_MaxValues(t *testing.T) {
	keyData := make([]byte, 16)
	valueData := make([]byte, 10+10)

	// Use maximum values
	binary.LittleEndian.PutUint64(keyData[0:8], ^uint64(0)) // All bits set
	binary.LittleEndian.PutUint64(keyData[8:16], ^uint64(0))

	binary.LittleEndian.PutUint64(valueData[0:8], ^uint64(0))
	binary.LittleEndian.PutUint16(valueData[8:10], 10)
	copy(valueData[10:], "maxvalue\x00")

	reader, err := NewSiblingLinkReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewSiblingLinkReader failed: %v", err)
	}

	if reader.SiblingID() != ^uint64(0) {
		t.Errorf("SiblingID() = %d, want %d", reader.SiblingID(), ^uint64(0))
	}

	if reader.ParentDirectoryID() != ^uint64(0) {
		t.Errorf("ParentDirectoryID() = %d, want %d", reader.ParentDirectoryID(), ^uint64(0))
	}
}
