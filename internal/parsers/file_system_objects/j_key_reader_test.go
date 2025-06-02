package file_system_objects

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestNewJKeyReader(t *testing.T) {
	tests := []struct {
		name        string
		objectID    uint64
		objectType  types.JObjTypes
		expectError bool
	}{
		{
			name:       "inode key",
			objectID:   123,
			objectType: types.ApfsTypeInode,
		},
		{
			name:       "directory entry key",
			objectID:   456,
			objectType: types.ApfsTypeDirRec,
		},
		{
			name:       "extended attribute key",
			objectID:   789,
			objectType: types.ApfsTypeXattr,
		},
		{
			name:       "directory stats key",
			objectID:   101112,
			objectType: types.ApfsTypeDirStats,
		},
		{
			name:       "maximum object ID",
			objectID:   types.ObjIdMask,
			objectType: types.ApfsTypeInode,
		},
		{
			name:       "minimum object ID",
			objectID:   1,
			objectType: types.ApfsTypeFileExtent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createJKeyTestData(tt.objectID, tt.objectType)

			reader, err := NewJKeyReader(data, binary.LittleEndian)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewJKeyReader failed: %v", err)
			}

			// Test object identifier
			if reader.ObjectIdentifier() != tt.objectID {
				t.Errorf("Expected object identifier %d, got %d", tt.objectID, reader.ObjectIdentifier())
			}

			// Test object type
			if reader.ObjectType() != tt.objectType {
				t.Errorf("Expected object type %d, got %d", tt.objectType, reader.ObjectType())
			}

			// Test raw field
			expectedRaw := (uint64(tt.objectType) << types.ObjTypeShift) | tt.objectID
			if reader.RawObjIdAndType() != expectedRaw {
				t.Errorf("Expected raw obj_id_and_type %d, got %d", expectedRaw, reader.RawObjIdAndType())
			}
		})
	}
}

func TestJKeyReaderErrorCases(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "insufficient data",
			data: make([]byte, 4),
		},
		{
			name: "empty data",
			data: make([]byte, 0),
		},
		{
			name: "nil data",
			data: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewJKeyReader(tt.data, binary.LittleEndian)
			if err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

func TestJKeyReaderBitMasking(t *testing.T) {
	// Test that bit masking works correctly for object ID and type extraction
	objectID := uint64(0x123456789ABCDEF)
	objectType := types.ApfsTypeInode

	data := createJKeyTestData(objectID, objectType)
	reader, err := NewJKeyReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewJKeyReader failed: %v", err)
	}

	// Object ID should be masked to only use the lower 60 bits
	expectedObjectID := objectID & types.ObjIdMask
	if reader.ObjectIdentifier() != expectedObjectID {
		t.Errorf("Expected masked object identifier %d, got %d", expectedObjectID, reader.ObjectIdentifier())
	}

	// Object type should be extracted from upper 4 bits
	if reader.ObjectType() != objectType {
		t.Errorf("Expected object type %d, got %d", objectType, reader.ObjectType())
	}
}

func TestJKeyReaderEndianness(t *testing.T) {
	objectID := uint64(0x123456789ABCDEF)
	objectType := types.ApfsTypeInode

	tests := []struct {
		name   string
		endian binary.ByteOrder
	}{
		{
			name:   "little endian",
			endian: binary.LittleEndian,
		},
		{
			name:   "big endian",
			endian: binary.BigEndian,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, 8)
			objIdAndType := (uint64(objectType) << types.ObjTypeShift) | objectID
			tt.endian.PutUint64(data, objIdAndType)

			reader, err := NewJKeyReader(data, tt.endian)
			if err != nil {
				t.Fatalf("NewJKeyReader failed: %v", err)
			}

			expectedObjectID := objectID & types.ObjIdMask
			if reader.ObjectIdentifier() != expectedObjectID {
				t.Errorf("Expected object identifier %d, got %d", expectedObjectID, reader.ObjectIdentifier())
			}

			if reader.ObjectType() != objectType {
				t.Errorf("Expected object type %d, got %d", objectType, reader.ObjectType())
			}
		})
	}
}

func TestJKeyReaderSystemObjectIDs(t *testing.T) {
	// Test with system object IDs that use the upper range
	systemObjectID := types.SystemObjIdMark | 123
	objectType := types.ApfsTypeInode

	data := createJKeyTestData(systemObjectID, objectType)
	reader, err := NewJKeyReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewJKeyReader failed: %v", err)
	}

	// Should still extract the object ID correctly
	expectedObjectID := systemObjectID & types.ObjIdMask
	if reader.ObjectIdentifier() != expectedObjectID {
		t.Errorf("Expected object identifier %d, got %d", expectedObjectID, reader.ObjectIdentifier())
	}

	if reader.ObjectType() != objectType {
		t.Errorf("Expected object type %d, got %d", objectType, reader.ObjectType())
	}
}

func createJKeyTestData(objectID uint64, objectType types.JObjTypes) []byte {
	data := make([]byte, 8)
	objIdAndType := (uint64(objectType) << types.ObjTypeShift) | objectID
	binary.LittleEndian.PutUint64(data, objIdAndType)
	return data
}
