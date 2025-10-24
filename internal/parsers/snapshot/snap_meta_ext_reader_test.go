package snapshot

import (
	"encoding/binary"
	"testing"
)

func TestSnapMetaExtReader_ValidData(t *testing.T) {
	data := make([]byte, 40)

	binary.LittleEndian.PutUint32(data[0:4], 1)      // SmeVersion
	binary.LittleEndian.PutUint32(data[4:8], 0)      // SmeFlags
	binary.LittleEndian.PutUint64(data[8:16], 1000)  // SmeSnapXid
	copy(data[16:32], "1234567890123456")            // SmeUuid (16 bytes)
	binary.LittleEndian.PutUint64(data[32:40], 5000) // SmeToken

	reader, err := NewSnapMetaExtReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewSnapMetaExtReader failed: %v", err)
	}

	if reader.Version() != 1 {
		t.Errorf("Version() = %d, want 1", reader.Version())
	}

	if reader.Flags() != 0 {
		t.Errorf("Flags() = %d, want 0", reader.Flags())
	}

	if reader.SnapXID() != 1000 {
		t.Errorf("SnapXID() = %d, want 1000", reader.SnapXID())
	}

	if reader.Token() != 5000 {
		t.Errorf("Token() = %d, want 5000", reader.Token())
	}

	uuid := reader.UUID()
	expected := [16]byte{}
	copy(expected[:], "1234567890123456")
	if uuid != expected {
		t.Errorf("UUID() = %v, want %v", uuid, expected)
	}
}

func TestSnapMetaExtReader_BigEndian(t *testing.T) {
	data := make([]byte, 40)

	binary.BigEndian.PutUint32(data[0:4], 2)
	binary.BigEndian.PutUint32(data[4:8], 0x12345678)
	binary.BigEndian.PutUint64(data[8:16], 2000)
	copy(data[16:32], "abcdefghijklmnop")
	binary.BigEndian.PutUint64(data[32:40], 6000)

	reader, err := NewSnapMetaExtReader(data, binary.BigEndian)
	if err != nil {
		t.Fatalf("NewSnapMetaExtReader failed: %v", err)
	}

	if reader.Version() != 2 {
		t.Errorf("Version() = %d, want 2", reader.Version())
	}

	if reader.Flags() != 0x12345678 {
		t.Errorf("Flags() = 0x%x, want 0x12345678", reader.Flags())
	}

	if reader.SnapXID() != 2000 {
		t.Errorf("SnapXID() = %d, want 2000", reader.SnapXID())
	}

	if reader.Token() != 6000 {
		t.Errorf("Token() = %d, want 6000", reader.Token())
	}
}

func TestSnapMetaExtReader_MaxValues(t *testing.T) {
	data := make([]byte, 40)

	binary.LittleEndian.PutUint32(data[0:4], ^uint32(0))
	binary.LittleEndian.PutUint32(data[4:8], ^uint32(0))
	binary.LittleEndian.PutUint64(data[8:16], ^uint64(0))
	for i := 16; i < 32; i++ {
		data[i] = 0xFF
	}
	binary.LittleEndian.PutUint64(data[32:40], ^uint64(0))

	reader, err := NewSnapMetaExtReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewSnapMetaExtReader failed: %v", err)
	}

	if reader.Version() != ^uint32(0) {
		t.Errorf("Version() = %d, want max uint32", reader.Version())
	}

	if reader.Flags() != ^uint32(0) {
		t.Errorf("Flags() = %d, want max uint32", reader.Flags())
	}

	if reader.SnapXID() != ^uint64(0) {
		t.Errorf("SnapXID() = %d, want max uint64", reader.SnapXID())
	}

	if reader.Token() != ^uint64(0) {
		t.Errorf("Token() = %d, want max uint64", reader.Token())
	}
}

func TestSnapMetaExtReader_ZeroValues(t *testing.T) {
	data := make([]byte, 40)
	// All zeros by default

	reader, err := NewSnapMetaExtReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewSnapMetaExtReader failed: %v", err)
	}

	if reader.Version() != 0 {
		t.Errorf("Version() = %d, want 0", reader.Version())
	}

	if reader.Flags() != 0 {
		t.Errorf("Flags() = %d, want 0", reader.Flags())
	}

	if reader.SnapXID() != 0 {
		t.Errorf("SnapXID() = %d, want 0", reader.SnapXID())
	}

	if reader.Token() != 0 {
		t.Errorf("Token() = %d, want 0", reader.Token())
	}
}

func TestSnapMetaExtReader_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		dataSize  int
		shouldErr bool
	}{
		{"Valid", 40, false},
		{"Too small", 20, true},
		{"Way too small", 8, true},
		{"Empty", 0, true},
		{"One less byte", 39, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := make([]byte, tc.dataSize)

			if len(data) >= 40 {
				binary.LittleEndian.PutUint32(data[0:4], 1)
				binary.LittleEndian.PutUint32(data[4:8], 0)
				binary.LittleEndian.PutUint64(data[8:16], 1000)
				copy(data[16:32], "1234567890123456")
				binary.LittleEndian.PutUint64(data[32:40], 5000)
			}

			_, err := NewSnapMetaExtReader(data, binary.LittleEndian)
			if tc.shouldErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSnapMetaExtReader_UUIDExtraction(t *testing.T) {
	data := make([]byte, 40)

	// Set a specific UUID pattern
	expectedUUID := [16]byte{
		0x01, 0x02, 0x03, 0x04,
		0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0x0C,
		0x0D, 0x0E, 0x0F, 0x10,
	}
	copy(data[16:32], expectedUUID[:])

	binary.LittleEndian.PutUint32(data[0:4], 1)
	binary.LittleEndian.PutUint32(data[4:8], 0)
	binary.LittleEndian.PutUint64(data[8:16], 100)
	binary.LittleEndian.PutUint64(data[32:40], 200)

	reader, err := NewSnapMetaExtReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewSnapMetaExtReader failed: %v", err)
	}

	uuid := reader.UUID()
	if uuid != expectedUUID {
		t.Errorf("UUID() = %v, want %v", uuid, expectedUUID)
	}
}

func TestSnapMetaExtReader_ConsistentEndianness(t *testing.T) {
	testValue := uint64(0x0102030405060708)

	// Little endian
	leData := make([]byte, 40)
	binary.LittleEndian.PutUint32(leData[0:4], 1)
	binary.LittleEndian.PutUint32(leData[4:8], 0)
	binary.LittleEndian.PutUint64(leData[8:16], testValue)
	copy(leData[16:32], "test-uuid-sixteen")
	binary.LittleEndian.PutUint64(leData[32:40], testValue)

	leReader, err := NewSnapMetaExtReader(leData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("Little-endian reader failed: %v", err)
	}

	// Big endian
	beData := make([]byte, 40)
	binary.BigEndian.PutUint32(beData[0:4], 1)
	binary.BigEndian.PutUint32(beData[4:8], 0)
	binary.BigEndian.PutUint64(beData[8:16], testValue)
	copy(beData[16:32], "test-uuid-sixteen")
	binary.BigEndian.PutUint64(beData[32:40], testValue)

	beReader, err := NewSnapMetaExtReader(beData, binary.BigEndian)
	if err != nil {
		t.Fatalf("Big-endian reader failed: %v", err)
	}

	// Both should read the same value despite different byte order
	if leReader.SnapXID() != beReader.SnapXID() {
		t.Errorf("SnapXID mismatch: LE=%d, BE=%d", leReader.SnapXID(), beReader.SnapXID())
	}

	if leReader.Token() != beReader.Token() {
		t.Errorf("Token mismatch: LE=%d, BE=%d", leReader.Token(), beReader.Token())
	}
}
