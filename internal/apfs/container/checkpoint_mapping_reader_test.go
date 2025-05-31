package container

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestCheckpointMappingData creates test data for checkpoint mapping
func createTestCheckpointMappingData(cpmType, cpmSubtype, cpmSize, cpmPad uint32, cpmFsOid, cpmOid types.OidT, cpmPaddr types.Paddr, endian binary.ByteOrder) []byte {
	data := make([]byte, 40) // CheckpointMappingT is 40 bytes
	endian.PutUint32(data[0:4], cpmType)
	endian.PutUint32(data[4:8], cpmSubtype)
	endian.PutUint32(data[8:12], cpmSize)
	endian.PutUint32(data[12:16], cpmPad)
	endian.PutUint64(data[16:24], uint64(cpmFsOid))
	endian.PutUint64(data[24:32], uint64(cpmOid))
	endian.PutUint64(data[32:40], uint64(cpmPaddr))
	return data
}

func TestCheckpointMappingReader(t *testing.T) {
	tests := []struct {
		name       string
		cpmType    uint32
		cpmSubtype uint32
		cpmSize    uint32
		cpmPad     uint32
		cpmFsOid   types.OidT
		cpmOid     types.OidT
		cpmPaddr   types.Paddr
	}{
		{
			name:       "Basic mapping",
			cpmType:    0x10,
			cpmSubtype: 0x20,
			cpmSize:    4096,
			cpmPad:     0,
			cpmFsOid:   types.OidT(0x1000),
			cpmOid:     types.OidT(0x2000),
			cpmPaddr:   types.Paddr(0x3000),
		},
		{
			name:       "Zero values",
			cpmType:    0,
			cpmSubtype: 0,
			cpmSize:    0,
			cpmPad:     0,
			cpmFsOid:   types.OidT(0),
			cpmOid:     types.OidT(0),
			cpmPaddr:   types.Paddr(0),
		},
		{
			name:       "Large values",
			cpmType:    0xFFFFFFFF,
			cpmSubtype: 0xFFFFFFFE,
			cpmSize:    0x100000, // 1MB
			cpmPad:     0x12345678,
			cpmFsOid:   types.OidT(0x123456789ABCDEF0),
			cpmOid:     types.OidT(0xFEDCBA9876543210),
			cpmPaddr:   types.Paddr(0x7FFFFFFFFFFFFFFF),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endian := binary.LittleEndian
			data := createTestCheckpointMappingData(
				tt.cpmType, tt.cpmSubtype, tt.cpmSize, tt.cpmPad,
				tt.cpmFsOid, tt.cpmOid, tt.cpmPaddr, endian)

			reader, err := NewCheckpointMappingReader(data, endian)
			if err != nil {
				t.Fatalf("NewCheckpointMappingReader() error = %v", err)
			}

			// Test all methods
			if objType := reader.Type(); objType != tt.cpmType {
				t.Errorf("Type() = 0x%X, want 0x%X", objType, tt.cpmType)
			}

			if subtype := reader.Subtype(); subtype != tt.cpmSubtype {
				t.Errorf("Subtype() = 0x%X, want 0x%X", subtype, tt.cpmSubtype)
			}

			if size := reader.Size(); size != tt.cpmSize {
				t.Errorf("Size() = %d, want %d", size, tt.cpmSize)
			}

			if fsOid := reader.FilesystemOID(); fsOid != tt.cpmFsOid {
				t.Errorf("FilesystemOID() = 0x%X, want 0x%X", fsOid, tt.cpmFsOid)
			}

			if objId := reader.ObjectID(); objId != tt.cpmOid {
				t.Errorf("ObjectID() = 0x%X, want 0x%X", objId, tt.cpmOid)
			}

			if paddr := reader.PhysicalAddress(); paddr != tt.cpmPaddr {
				t.Errorf("PhysicalAddress() = 0x%X, want 0x%X", paddr, tt.cpmPaddr)
			}
		})
	}
}

func TestCheckpointMappingReader_ErrorCases(t *testing.T) {
	endian := binary.LittleEndian

	tests := []struct {
		name     string
		dataSize int
	}{
		{
			name:     "Empty data",
			dataSize: 0,
		},
		{
			name:     "Too small data - 39 bytes",
			dataSize: 39,
		},
		{
			name:     "Too small data - 32 bytes",
			dataSize: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.dataSize)

			_, err := NewCheckpointMappingReader(data, endian)
			if err == nil {
				t.Error("NewCheckpointMappingReader() expected error, got nil")
			}
		})
	}
}
