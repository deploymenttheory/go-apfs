package sealed_volumes

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestIntegrityMetaReader_ValidData(t *testing.T) {
	data := make([]byte, 152)

	// Set up ObjPhysT header (40 bytes)
	// OChecksum (32 bytes) - just fill with test data
	for i := 0; i < 32; i++ {
		data[i] = 0xAB
	}
	// OOid (8 bytes)
	binary.LittleEndian.PutUint64(data[32:40], 1000)
	// OXid (8 bytes)
	binary.LittleEndian.PutUint64(data[40:48], 100)
	// OType (4 bytes)
	binary.LittleEndian.PutUint32(data[48:52], 0x1E)
	// OSubtype (4 bytes)
	binary.LittleEndian.PutUint32(data[52:56], 0)

	// Set up integrity metadata fields
	binary.LittleEndian.PutUint32(data[56:60], types.IntegrityMetaVersion2)  // ImVersion
	binary.LittleEndian.PutUint32(data[60:64], types.ApfsSealBroken)         // ImFlags
	binary.LittleEndian.PutUint32(data[64:68], uint32(types.ApfsHashSha256)) // ImHashType
	binary.LittleEndian.PutUint32(data[68:72], 1024)                         // ImRootHashOffset
	binary.LittleEndian.PutUint64(data[72:80], 999)                          // ImBrokenXid
	// ImReserved already initialized to zeros

	reader, err := NewIntegrityMetaReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewIntegrityMetaReader failed: %v", err)
	}

	if reader.Version() != types.IntegrityMetaVersion2 {
		t.Errorf("Version() = %d, want %d", reader.Version(), types.IntegrityMetaVersion2)
	}

	if reader.Flags() != types.ApfsSealBroken {
		t.Errorf("Flags() = %d, want %d", reader.Flags(), types.ApfsSealBroken)
	}

	if reader.HashType() != types.ApfsHashSha256 {
		t.Errorf("HashType() = %d, want %d", reader.HashType(), types.ApfsHashSha256)
	}

	if reader.RootHashOffset() != 1024 {
		t.Errorf("RootHashOffset() = %d, want 1024", reader.RootHashOffset())
	}
}

func TestIntegrityMetaReader_BigEndian(t *testing.T) {
	data := make([]byte, 152)

	// Set up ObjPhysT header (40 bytes)
	for i := 0; i < 32; i++ {
		data[i] = 0xCD
	}
	binary.BigEndian.PutUint64(data[32:40], 1000)
	binary.BigEndian.PutUint64(data[40:48], 100)
	binary.BigEndian.PutUint32(data[48:52], 0x1E)
	binary.BigEndian.PutUint32(data[52:56], 0)

	binary.BigEndian.PutUint32(data[56:60], types.IntegrityMetaVersion1)
	binary.BigEndian.PutUint32(data[60:64], 0)
	binary.BigEndian.PutUint32(data[64:68], uint32(types.ApfsHashSha512))
	binary.BigEndian.PutUint32(data[68:72], 2048)
	binary.BigEndian.PutUint64(data[72:80], 0)

	reader, err := NewIntegrityMetaReader(data, binary.BigEndian)
	if err != nil {
		t.Fatalf("NewIntegrityMetaReader failed: %v", err)
	}

	if reader.Version() != types.IntegrityMetaVersion1 {
		t.Errorf("Version() = %d, want %d", reader.Version(), types.IntegrityMetaVersion1)
	}

	if reader.HashType() != types.ApfsHashSha512 {
		t.Errorf("HashType() = %d, want %d", reader.HashType(), types.ApfsHashSha512)
	}
}

func TestIntegrityMetaReader_SealStatus(t *testing.T) {
	data := make([]byte, 152)

	// Setup minimal valid data
	binary.LittleEndian.PutUint32(data[56:60], types.IntegrityMetaVersion2)
	binary.LittleEndian.PutUint32(data[64:68], uint32(types.ApfsHashSha256))

	// Test with seal broken
	binary.LittleEndian.PutUint32(data[60:64], types.ApfsSealBroken)
	binary.LittleEndian.PutUint64(data[72:80], 777)

	reader, err := NewIntegrityMetaReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewIntegrityMetaReader failed: %v", err)
	}

	if !reader.(*integrityMetaReader).IsSealBroken() {
		t.Error("IsSealBroken() = false, want true")
	}

	if reader.(*integrityMetaReader).BrokenTransactionID() != 777 {
		t.Errorf("BrokenTransactionID() = %d, want 777", reader.(*integrityMetaReader).BrokenTransactionID())
	}
}

func TestIntegrityMetaReader_HashTypes(t *testing.T) {
	tests := []struct {
		name     string
		hashType types.ApfsHashTypeT
	}{
		{"SHA256", types.ApfsHashSha256},
		{"SHA512/256", types.ApfsHashSha512256},
		{"SHA384", types.ApfsHashSha384},
		{"SHA512", types.ApfsHashSha512},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := make([]byte, 152)
			binary.LittleEndian.PutUint32(data[56:60], types.IntegrityMetaVersion2)
			binary.LittleEndian.PutUint32(data[64:68], uint32(tc.hashType))

			reader, err := NewIntegrityMetaReader(data, binary.LittleEndian)
			if err != nil {
				t.Fatalf("NewIntegrityMetaReader failed: %v", err)
			}

			if reader.HashType() != tc.hashType {
				t.Errorf("HashType() = %d, want %d", reader.HashType(), tc.hashType)
			}

			if !reader.(*integrityMetaReader).IsHashTypeValid() {
				t.Error("IsHashTypeValid() = false, want true")
			}
		})
	}
}

func TestIntegrityMetaReader_Validation(t *testing.T) {
	data := make([]byte, 152)

	binary.LittleEndian.PutUint32(data[56:60], types.IntegrityMetaVersion2)
	binary.LittleEndian.PutUint32(data[64:68], uint32(types.ApfsHashSha256))

	reader, err := NewIntegrityMetaReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewIntegrityMetaReader failed: %v", err)
	}

	if !reader.(*integrityMetaReader).IsVersionValid() {
		t.Error("IsVersionValid() = false, want true")
	}

	if !reader.(*integrityMetaReader).IsHashTypeValid() {
		t.Error("IsHashTypeValid() = false, want true")
	}
}

func TestIntegrityMetaReader_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		dataSize  int
		shouldErr bool
	}{
		{"Valid", 152, false},
		{"Too small", 100, true},
		{"Way too small", 16, true},
		{"Empty", 0, true},
		{"One byte short", 151, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := make([]byte, tc.dataSize)

			if len(data) >= 152 {
				binary.LittleEndian.PutUint32(data[56:60], types.IntegrityMetaVersion2)
				binary.LittleEndian.PutUint32(data[64:68], uint32(types.ApfsHashSha256))
			}

			_, err := NewIntegrityMetaReader(data, binary.LittleEndian)
			if tc.shouldErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestIntegrityMetaReader_MaxValues(t *testing.T) {
	data := make([]byte, 152)

	binary.LittleEndian.PutUint32(data[56:60], types.IntegrityMetaVersionHighest)
	binary.LittleEndian.PutUint32(data[60:64], ^uint32(0))
	binary.LittleEndian.PutUint32(data[64:68], uint32(types.ApfsHashMax))
	binary.LittleEndian.PutUint32(data[68:72], ^uint32(0))
	binary.LittleEndian.PutUint64(data[72:80], ^uint64(0))

	reader, err := NewIntegrityMetaReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewIntegrityMetaReader failed: %v", err)
	}

	if reader.Flags() != ^uint32(0) {
		t.Errorf("Flags() = %d, want %d", reader.Flags(), ^uint32(0))
	}

	if reader.RootHashOffset() != ^uint32(0) {
		t.Errorf("RootHashOffset() = %d, want %d", reader.RootHashOffset(), ^uint32(0))
	}
}
