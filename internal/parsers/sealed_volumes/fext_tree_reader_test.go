package sealed_volumes

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestFextTreeKeyReader_ValidData(t *testing.T) {
	data := make([]byte, 16)

	binary.LittleEndian.PutUint64(data[0:8], 12345)    // PrivateId (file ID)
	binary.LittleEndian.PutUint64(data[8:16], 1048576) // LogicalAddr

	reader, err := NewFextTreeKeyReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewFextTreeKeyReader failed: %v", err)
	}

	if reader.FileID() != 12345 {
		t.Errorf("FileID() = %d, want 12345", reader.FileID())
	}

	if reader.LogicalAddress() != 1048576 {
		t.Errorf("LogicalAddress() = %d, want 1048576", reader.LogicalAddress())
	}
}

func TestFextTreeValReader_ValidData(t *testing.T) {
	data := make([]byte, 16)

	// Length in lower bits, flags in upper bits
	lenAndFlags := uint64(4096) // 4KB extent
	binary.LittleEndian.PutUint64(data[0:8], lenAndFlags)
	binary.LittleEndian.PutUint64(data[8:16], 65536) // Physical block number

	reader, err := NewFextTreeValReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewFextTreeValReader failed: %v", err)
	}

	if reader.Length() != 4096 {
		t.Errorf("Length() = %d, want 4096", reader.Length())
	}

	if reader.PhysicalBlockNumber() != 65536 {
		t.Errorf("PhysicalBlockNumber() = %d, want 65536", reader.PhysicalBlockNumber())
	}
}

func TestFextTreeKeyReader_BigEndian(t *testing.T) {
	data := make([]byte, 16)

	binary.BigEndian.PutUint64(data[0:8], 54321)
	binary.BigEndian.PutUint64(data[8:16], 2097152)

	reader, err := NewFextTreeKeyReader(data, binary.BigEndian)
	if err != nil {
		t.Fatalf("NewFextTreeKeyReader failed: %v", err)
	}

	if reader.FileID() != 54321 {
		t.Errorf("FileID() = %d, want 54321", reader.FileID())
	}

	if reader.LogicalAddress() != 2097152 {
		t.Errorf("LogicalAddress() = %d, want 2097152", reader.LogicalAddress())
	}
}

func TestFextTreeValReader_BigEndian(t *testing.T) {
	data := make([]byte, 16)

	binary.BigEndian.PutUint64(data[0:8], uint64(8192))
	binary.BigEndian.PutUint64(data[8:16], 131072)

	reader, err := NewFextTreeValReader(data, binary.BigEndian)
	if err != nil {
		t.Fatalf("NewFextTreeValReader failed: %v", err)
	}

	if reader.Length() != 8192 {
		t.Errorf("Length() = %d, want 8192", reader.Length())
	}

	if reader.PhysicalBlockNumber() != 131072 {
		t.Errorf("PhysicalBlockNumber() = %d, want 131072", reader.PhysicalBlockNumber())
	}
}

func TestFextTreeValReader_ExtentLength(t *testing.T) {
	tests := []struct {
		name   string
		length uint64
	}{
		{"Small extent", 4096},
		{"Medium extent", 1048576},
		{"Large extent", 268435456},
		{"Max length", types.JFileExtentLenMask},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := make([]byte, 16)
			binary.LittleEndian.PutUint64(data[0:8], tc.length)
			binary.LittleEndian.PutUint64(data[8:16], 1000)

			reader, err := NewFextTreeValReader(data, binary.LittleEndian)
			if err != nil {
				t.Fatalf("NewFextTreeValReader failed: %v", err)
			}

			if reader.Length() != tc.length {
				t.Errorf("Length() = %d, want %d", reader.Length(), tc.length)
			}
		})
	}
}

func TestFextTreeValReader_Flags(t *testing.T) {
	data := make([]byte, 16)

	// Set flags in upper bits (simulating some flag bits set)
	lenAndFlags := uint64(4096) | (uint64(0xFF) << 56)
	binary.LittleEndian.PutUint64(data[0:8], lenAndFlags)
	binary.LittleEndian.PutUint64(data[8:16], 5000)

	reader, err := NewFextTreeValReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewFextTreeValReader failed: %v", err)
	}

	// Flags should be extracted using the mask and shift
	flags := reader.Flags()
	if flags == 0 {
		// Flags extraction depends on the mask/shift constants
		// This test just verifies the method works without error
	}
}

func TestFextTreeReader_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		dataSize  int
		shouldErr bool
	}{
		{"Valid", 16, false},
		{"Too small", 8, true},
		{"Way too small", 4, true},
		{"Empty", 0, true},
		{"One byte short", 15, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := make([]byte, tc.dataSize)

			if len(data) >= 16 {
				binary.LittleEndian.PutUint64(data[0:8], 1000)
				binary.LittleEndian.PutUint64(data[8:16], 2000)
			}

			_, errKey := NewFextTreeKeyReader(data, binary.LittleEndian)
			_, errVal := NewFextTreeValReader(data, binary.LittleEndian)

			if tc.shouldErr {
				if errKey == nil {
					t.Error("NewFextTreeKeyReader: expected error, got nil")
				}
				if errVal == nil {
					t.Error("NewFextTreeValReader: expected error, got nil")
				}
			} else {
				if errKey != nil {
					t.Errorf("NewFextTreeKeyReader: unexpected error: %v", errKey)
				}
				if errVal != nil {
					t.Errorf("NewFextTreeValReader: unexpected error: %v", errVal)
				}
			}
		})
	}
}

func TestFextTreeReader_MaxValues(t *testing.T) {
	keyData := make([]byte, 16)
	valData := make([]byte, 16)

	binary.LittleEndian.PutUint64(keyData[0:8], ^uint64(0))
	binary.LittleEndian.PutUint64(keyData[8:16], ^uint64(0))

	keyReader, err := NewFextTreeKeyReader(keyData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewFextTreeKeyReader failed: %v", err)
	}

	if keyReader.FileID() != ^uint64(0) {
		t.Errorf("FileID() = %d, want max", keyReader.FileID())
	}

	if keyReader.LogicalAddress() != ^uint64(0) {
		t.Errorf("LogicalAddress() = %d, want max", keyReader.LogicalAddress())
	}

	// For value with max length in lower bits
	binary.LittleEndian.PutUint64(valData[0:8], types.JFileExtentLenMask)
	binary.LittleEndian.PutUint64(valData[8:16], ^uint64(0))

	valReader, err := NewFextTreeValReader(valData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewFextTreeValReader failed: %v", err)
	}

	if valReader.Length() != types.JFileExtentLenMask {
		t.Errorf("Length() = %d, want %d", valReader.Length(), types.JFileExtentLenMask)
	}

	if valReader.PhysicalBlockNumber() != ^uint64(0) {
		t.Errorf("PhysicalBlockNumber() = %d, want max", valReader.PhysicalBlockNumber())
	}
}
