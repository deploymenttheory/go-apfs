package sealed_volumes

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestFileInfoReader_ValidData(t *testing.T) {
	keyData := make([]byte, 16)
	valueData := make([]byte, 3+32) // HashedLen(2) + HashSize(1) + SHA256 hash(32)

	// Set up key
	binary.LittleEndian.PutUint64(keyData[0:8], 0x5000000000000123) // ObjIdAndType
	binary.LittleEndian.PutUint64(keyData[8:16], (1<<56)|0x1000)    // Type=1 (DATA_HASH), LBA=0x1000

	// Set up value
	binary.LittleEndian.PutUint16(valueData[0:2], 100) // HashedLen
	valueData[2] = 32                                  // HashSize
	copy(valueData[3:], "test_hash_data_32bytes......")

	reader, err := NewFileInfoReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewFileInfoReader failed: %v", err)
	}

	if reader.HashedLength() != 100 {
		t.Errorf("HashedLength() = %d, want 100", reader.HashedLength())
	}

	hash := reader.DataHash()
	if len(hash) != 32 {
		t.Errorf("DataHash() length = %d, want 32", len(hash))
	}
}

func TestFileInfoReader_BigEndian(t *testing.T) {
	keyData := make([]byte, 16)
	valueData := make([]byte, 3+48) // HashedLen(2) + HashSize(1) + SHA384 hash(48)

	binary.BigEndian.PutUint64(keyData[0:8], 0x5000000000000456)
	binary.BigEndian.PutUint64(keyData[8:16], (1<<56)|0x2000)

	binary.BigEndian.PutUint16(valueData[0:2], 200)
	valueData[2] = 48
	copy(valueData[3:], "test_hash_384_bits_of_data_testing")

	reader, err := NewFileInfoReader(keyData, valueData, binary.BigEndian)
	if err != nil {
		t.Fatalf("NewFileInfoReader failed: %v", err)
	}

	if reader.HashedLength() != 200 {
		t.Errorf("HashedLength() = %d, want 200", reader.HashedLength())
	}

	hash := reader.DataHash()
	if len(hash) != 48 {
		t.Errorf("DataHash() length = %d, want 48", len(hash))
	}
}

func TestFileInfoReader_VariousHashSizes(t *testing.T) {
	tests := []struct {
		name     string
		hashSize uint8
	}{
		{"SHA256", 32},
		{"SHA512/256", 32},
		{"SHA384", 48},
		{"SHA512", 64},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keyData := make([]byte, 16)
			valueData := make([]byte, 3+int(tc.hashSize))

			binary.LittleEndian.PutUint64(keyData[0:8], 0x5000000000000001)
			binary.LittleEndian.PutUint64(keyData[8:16], (1<<56)|0x0100)

			binary.LittleEndian.PutUint16(valueData[0:2], 50)
			valueData[2] = tc.hashSize

			reader, err := NewFileInfoReader(keyData, valueData, binary.LittleEndian)
			if err != nil {
				t.Fatalf("NewFileInfoReader failed: %v", err)
			}

			hash := reader.DataHash()
			if len(hash) != int(tc.hashSize) {
				t.Errorf("DataHash() length = %d, want %d", len(hash), tc.hashSize)
			}
		})
	}
}

func TestFileInfoReader_BitFieldExtraction(t *testing.T) {
	keyData := make([]byte, 16)
	valueData := make([]byte, 3+32)

	binary.LittleEndian.PutUint64(keyData[0:8], 0x5000000000000789)
	// Set InfoAndLba: Type in bits 56-63 (shift by 56), LBA in bits 0-55
	// Type = 1, LBA = 0xFF (maximum that fits in 56 bits within our test range)
	infoAndLba := (uint64(1) << 56) | uint64(0xFF)
	binary.LittleEndian.PutUint64(keyData[8:16], infoAndLba)

	binary.LittleEndian.PutUint16(valueData[0:2], 75)
	valueData[2] = 32

	reader, err := NewFileInfoReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewFileInfoReader failed: %v", err)
	}

	fir := reader.(*fileInfoReader)
	// Extract type from bits 56-63
	extractedType := types.JObjFileInfoType((infoAndLba & types.JFileInfoTypeMask) >> types.JFileInfoTypeShift)
	if fir.GetInfoType() != extractedType {
		t.Errorf("GetInfoType() = %d, want %d", fir.GetInfoType(), extractedType)
	}

	// Extract LBA from bits 0-55
	extractedLBA := infoAndLba & types.JFileInfoLbaMask
	if fir.GetLogicalBlockAddress() != extractedLBA {
		t.Errorf("GetLogicalBlockAddress() = %#x, want %#x", fir.GetLogicalBlockAddress(), extractedLBA)
	}
}

func TestFileInfoReader_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		keySize   int
		valueSize int
		hashSize  uint8
		shouldErr bool
	}{
		{"Valid", 16, 35, 32, false},
		{"Key too small", 8, 35, 32, true},
		{"Value too small", 16, 2, 32, true},
		{"Hash size exceeds max", 16, 68, 65, true},
		{"Hash data insufficient", 16, 10, 32, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keyData := make([]byte, tc.keySize)
			valueData := make([]byte, tc.valueSize)

			if len(keyData) >= 16 {
				binary.LittleEndian.PutUint64(keyData[0:8], 0x5000000000000001)
				binary.LittleEndian.PutUint64(keyData[8:16], (1<<56)|0x0100)
			}

			if len(valueData) >= 3 {
				binary.LittleEndian.PutUint16(valueData[0:2], 50)
				valueData[2] = tc.hashSize
			}

			_, err := NewFileInfoReader(keyData, valueData, binary.LittleEndian)
			if tc.shouldErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestFileInfoReader_ZeroHashSize(t *testing.T) {
	keyData := make([]byte, 16)
	valueData := make([]byte, 4) // Minimum: HashedLen(2) + HashSize(1) + 1 byte padding

	binary.LittleEndian.PutUint64(keyData[0:8], 0x5000000000000001)
	binary.LittleEndian.PutUint64(keyData[8:16], (1<<56)|0x0100)
	binary.LittleEndian.PutUint16(valueData[0:2], 0)
	valueData[2] = 0

	reader, err := NewFileInfoReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewFileInfoReader failed: %v", err)
	}

	hash := reader.DataHash()
	if len(hash) != 0 {
		t.Errorf("DataHash() length = %d, want 0", len(hash))
	}
}
