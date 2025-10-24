package snapshot

import (
	"encoding/binary"
	"testing"
	"time"
)

func TestSnapMetadataReader_ValidData(t *testing.T) {
	keyData := make([]byte, 8)
	valueData := make([]byte, 50+10) // Fixed fields + "backup\x00" (8 bytes)

	// Set up key
	binary.LittleEndian.PutUint64(keyData[0:8], 0x5000000000000001) // Type=5, ObjId=1

	// Set up value
	binary.LittleEndian.PutUint64(valueData[0:8], 1000)                  // ExtentrefTreeOid
	binary.LittleEndian.PutUint64(valueData[8:16], 2000)                 // SblockOid
	binary.LittleEndian.PutUint64(valueData[16:24], 1609459200000000000) // CreateTime
	binary.LittleEndian.PutUint64(valueData[24:32], 1609459200000000000) // ChangeTime
	binary.LittleEndian.PutUint64(valueData[32:40], 123)                 // Inum
	binary.LittleEndian.PutUint32(valueData[40:44], 1)                   // ExtentrefTreeType
	binary.LittleEndian.PutUint32(valueData[44:48], 0)                   // Flags
	binary.LittleEndian.PutUint16(valueData[48:50], 8)                   // NameLen (includes null)
	copy(valueData[50:], "backup\x00")

	reader, err := NewSnapMetadataReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewSnapMetadataReader failed: %v", err)
	}

	if reader.ExtentRefTreeOID() != 1000 {
		t.Errorf("ExtentRefTreeOID() = %d, want 1000", reader.ExtentRefTreeOID())
	}

	if reader.SuperblockOID() != 2000 {
		t.Errorf("SuperblockOID() = %d, want 2000", reader.SuperblockOID())
	}

	if reader.InodeNumber() != 123 {
		t.Errorf("InodeNumber() = %d, want 123", reader.InodeNumber())
	}

	if reader.ExtentRefTreeType() != 1 {
		t.Errorf("ExtentRefTreeType() = %d, want 1", reader.ExtentRefTreeType())
	}

	if reader.Name() != "backup" {
		t.Errorf("Name() = %q, want %q", reader.Name(), "backup")
	}
}

func TestSnapMetadataReader_BigEndian(t *testing.T) {
	keyData := make([]byte, 8)
	valueData := make([]byte, 50+5)

	binary.BigEndian.PutUint64(keyData[0:8], 0x0500000000000001)
	binary.BigEndian.PutUint64(valueData[0:8], 3000)
	binary.BigEndian.PutUint64(valueData[8:16], 4000)
	binary.BigEndian.PutUint64(valueData[16:24], 1609459200000000000)
	binary.BigEndian.PutUint64(valueData[24:32], 1609459200000000000)
	binary.BigEndian.PutUint64(valueData[32:40], 456)
	binary.BigEndian.PutUint32(valueData[40:44], 2)
	binary.BigEndian.PutUint32(valueData[44:48], 0)
	binary.BigEndian.PutUint16(valueData[48:50], 5)
	copy(valueData[50:], "snap\x00")

	reader, err := NewSnapMetadataReader(keyData, valueData, binary.BigEndian)
	if err != nil {
		t.Fatalf("NewSnapMetadataReader failed: %v", err)
	}

	if reader.ExtentRefTreeOID() != 3000 {
		t.Errorf("ExtentRefTreeOID() = %d, want 3000", reader.ExtentRefTreeOID())
	}

	if reader.Name() != "snap" {
		t.Errorf("Name() = %q, want %q", reader.Name(), "snap")
	}
}

func TestSnapMetadataReader_TimestampConversion(t *testing.T) {
	keyData := make([]byte, 8)
	valueData := make([]byte, 50+1)

	binary.LittleEndian.PutUint64(keyData[0:8], 0x5000000000000001)
	binary.LittleEndian.PutUint64(valueData[0:8], 1000)
	binary.LittleEndian.PutUint64(valueData[8:16], 2000)
	// 2021-01-01 00:00:00 UTC
	expectedNano := int64(1609459200000000000)
	binary.LittleEndian.PutUint64(valueData[16:24], uint64(expectedNano))
	binary.LittleEndian.PutUint64(valueData[24:32], uint64(expectedNano))
	binary.LittleEndian.PutUint64(valueData[32:40], 100)
	binary.LittleEndian.PutUint32(valueData[40:44], 1)
	binary.LittleEndian.PutUint32(valueData[44:48], 0)
	binary.LittleEndian.PutUint16(valueData[48:50], 1)
	valueData[50] = 0

	reader, err := NewSnapMetadataReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewSnapMetadataReader failed: %v", err)
	}

	expected := time.Unix(0, expectedNano)
	if reader.CreateTime() != expected {
		t.Errorf("CreateTime() = %v, want %v", reader.CreateTime(), expected)
	}
}

func TestSnapMetadataReader_Flags(t *testing.T) {
	keyData := make([]byte, 8)
	valueData := make([]byte, 50+1)

	binary.LittleEndian.PutUint64(keyData[0:8], 0x5000000000000001)
	binary.LittleEndian.PutUint64(valueData[0:8], 1000)
	binary.LittleEndian.PutUint64(valueData[8:16], 2000)
	binary.LittleEndian.PutUint64(valueData[16:24], 100)
	binary.LittleEndian.PutUint64(valueData[24:32], 100)
	binary.LittleEndian.PutUint64(valueData[32:40], 50)
	binary.LittleEndian.PutUint32(valueData[40:44], 1)
	binary.LittleEndian.PutUint32(valueData[44:48], 0x00000001) // SNAP_META_PENDING_DATALESS
	binary.LittleEndian.PutUint16(valueData[48:50], 1)
	valueData[50] = 0

	reader, err := NewSnapMetadataReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewSnapMetadataReader failed: %v", err)
	}

	if !reader.HasFlag(0x00000001) {
		t.Error("HasFlag(0x00000001) = false, want true")
	}

	if reader.HasFlag(0x00000002) {
		t.Error("HasFlag(0x00000002) = true, want false")
	}
}

func TestSnapMetadataReader_VariableLengthNames(t *testing.T) {
	tests := []struct {
		name     string
		nameStr  string
		extraLen int
	}{
		{"single char", "a\x00", 2},
		{"normal", "my-snapshot\x00", 13},
		{"long", "very-long-snapshot-name-with-special-chars\x00", 44},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keyData := make([]byte, 8)
			valueData := make([]byte, 50+len(tc.nameStr))

			binary.LittleEndian.PutUint64(keyData[0:8], 0x5000000000000001)
			binary.LittleEndian.PutUint64(valueData[0:8], 1000)
			binary.LittleEndian.PutUint64(valueData[8:16], 2000)
			binary.LittleEndian.PutUint64(valueData[16:24], 100)
			binary.LittleEndian.PutUint64(valueData[24:32], 100)
			binary.LittleEndian.PutUint64(valueData[32:40], 50)
			binary.LittleEndian.PutUint32(valueData[40:44], 1)
			binary.LittleEndian.PutUint32(valueData[44:48], 0)
			binary.LittleEndian.PutUint16(valueData[48:50], uint16(len(tc.nameStr)))
			copy(valueData[50:], tc.nameStr)

			reader, err := NewSnapMetadataReader(keyData, valueData, binary.LittleEndian)
			if err != nil {
				t.Fatalf("NewSnapMetadataReader failed: %v", err)
			}

			expected := tc.nameStr[:len(tc.nameStr)-1] // Remove null terminator
			if reader.Name() != expected {
				t.Errorf("Name() = %q, want %q", reader.Name(), expected)
			}
		})
	}
}

func TestSnapMetadataReader_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		keySize   int
		valueSize int
		shouldErr bool
	}{
		{"Valid", 8, 51, false}, // Changed from 50 to 51 to include 1 byte of name
		{"Key too small", 4, 50, true},
		{"Value too small", 8, 30, true},
		{"Both too small", 4, 30, true},
		{"Key empty", 0, 50, true},
		{"Value empty", 8, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keyData := make([]byte, tc.keySize)
			valueData := make([]byte, tc.valueSize)

			if len(keyData) >= 8 {
				binary.LittleEndian.PutUint64(keyData[0:8], 0x5000000000000001)
			}
			if len(valueData) >= 50 {
				binary.LittleEndian.PutUint64(valueData[0:8], 1000)
				binary.LittleEndian.PutUint64(valueData[8:16], 2000)
				binary.LittleEndian.PutUint64(valueData[16:24], 100)
				binary.LittleEndian.PutUint64(valueData[24:32], 100)
				binary.LittleEndian.PutUint64(valueData[32:40], 50)
				binary.LittleEndian.PutUint32(valueData[40:44], 1)
				binary.LittleEndian.PutUint32(valueData[44:48], 0)
				binary.LittleEndian.PutUint16(valueData[48:50], uint16(len(valueData)-50)) // NameLen = remaining bytes
				if len(valueData) > 50 {
					copy(valueData[50:], "\x00")
				}
			}

			_, err := NewSnapMetadataReader(keyData, valueData, binary.LittleEndian)
			if tc.shouldErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSnapMetadataReader_InsufficientNameData(t *testing.T) {
	keyData := make([]byte, 8)
	valueData := make([]byte, 50)

	binary.LittleEndian.PutUint64(keyData[0:8], 0x5000000000000001)
	binary.LittleEndian.PutUint64(valueData[0:8], 1000)
	binary.LittleEndian.PutUint64(valueData[8:16], 2000)
	binary.LittleEndian.PutUint64(valueData[16:24], 100)
	binary.LittleEndian.PutUint64(valueData[24:32], 100)
	binary.LittleEndian.PutUint64(valueData[32:40], 50)
	binary.LittleEndian.PutUint32(valueData[40:44], 1)
	binary.LittleEndian.PutUint32(valueData[44:48], 0)
	binary.LittleEndian.PutUint16(valueData[48:50], 50) // Claim 50 bytes but only have 0

	_, err := NewSnapMetadataReader(keyData, valueData, binary.LittleEndian)
	if err == nil {
		t.Error("expected error for insufficient name data")
	}
}

func TestSnapMetadataReader_MaxValues(t *testing.T) {
	keyData := make([]byte, 8)
	valueData := make([]byte, 50+1)

	binary.LittleEndian.PutUint64(keyData[0:8], ^uint64(0))
	binary.LittleEndian.PutUint64(valueData[0:8], ^uint64(0))
	binary.LittleEndian.PutUint64(valueData[8:16], ^uint64(0))
	binary.LittleEndian.PutUint64(valueData[16:24], ^uint64(0))
	binary.LittleEndian.PutUint64(valueData[24:32], ^uint64(0))
	binary.LittleEndian.PutUint64(valueData[32:40], ^uint64(0))
	binary.LittleEndian.PutUint32(valueData[40:44], ^uint32(0))
	binary.LittleEndian.PutUint32(valueData[44:48], ^uint32(0))
	binary.LittleEndian.PutUint16(valueData[48:50], 1)
	valueData[50] = 0

	reader, err := NewSnapMetadataReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewSnapMetadataReader failed: %v", err)
	}

	if reader.ExtentRefTreeOID() != ^uint64(0) {
		t.Errorf("ExtentRefTreeOID() = %d, want max uint64", reader.ExtentRefTreeOID())
	}

	if reader.Flags() != ^uint32(0) {
		t.Errorf("Flags() = %d, want max uint32", reader.Flags())
	}
}
