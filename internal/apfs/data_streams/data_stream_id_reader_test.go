package datastreams

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestDataStreamIDData creates test data for data stream ID key and value
func createTestDataStreamIDData(objectID uint64, refCount uint32, endian binary.ByteOrder) ([]byte, []byte) {
	// Create key data (8 bytes)
	keyData := make([]byte, 8)
	objIdAndType := objectID & types.ObjIdMask // No type bits set for simplicity
	endian.PutUint64(keyData[0:8], objIdAndType)

	// Create value data (4 bytes)
	valueData := make([]byte, 4)
	endian.PutUint32(valueData[0:4], refCount)

	return keyData, valueData
}

func TestDataStreamIDReader(t *testing.T) {
	tests := []struct {
		name     string
		objectID uint64
		refCount uint32
	}{
		{
			name:     "Zero values",
			objectID: 0,
			refCount: 0,
		},
		{
			name:     "Small object ID",
			objectID: 0x1000,
			refCount: 1,
		},
		{
			name:     "Large object ID",
			objectID: 0x0123456789ABCDEF,
			refCount: 100,
		},
		{
			name:     "High reference count",
			objectID: 0x2000,
			refCount: 0xFFFFFFFF, // Max uint32
		},
		{
			name:     "Typical data stream",
			objectID: 0x42424242,
			refCount: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endian := binary.LittleEndian

			keyData, valueData := createTestDataStreamIDData(tt.objectID, tt.refCount, endian)

			reader, err := NewDataStreamIDReader(keyData, valueData, endian)
			if err != nil {
				t.Fatalf("NewDataStreamIDReader() error = %v", err)
			}

			// Test ObjectID
			if objectID := reader.ObjectID(); objectID != tt.objectID {
				t.Errorf("ObjectID() = 0x%X, want 0x%X", objectID, tt.objectID)
			}

			// Test ReferenceCount
			if refCount := reader.ReferenceCount(); refCount != tt.refCount {
				t.Errorf("ReferenceCount() = %d, want %d", refCount, tt.refCount)
			}
		})
	}
}

func TestDataStreamIDReader_ObjectIDMasking(t *testing.T) {
	endian := binary.LittleEndian

	// Test that type bits are properly masked out when extracting object ID
	objectIDWithType := uint64(0x0123456789ABCDEF) | (uint64(0x80) << 56) // Add some type bits
	expectedObjectID := uint64(0x0123456789ABCDEF) & types.ObjIdMask

	keyData, valueData := createTestDataStreamIDData(objectIDWithType, 1, endian)

	reader, err := NewDataStreamIDReader(keyData, valueData, endian)
	if err != nil {
		t.Fatalf("NewDataStreamIDReader() error = %v", err)
	}

	if objectID := reader.ObjectID(); objectID != expectedObjectID {
		t.Errorf("ObjectID() = 0x%X, want 0x%X (type bits should be masked)", objectID, expectedObjectID)
	}
}

func TestDataStreamIDReader_ErrorCases(t *testing.T) {
	endian := binary.LittleEndian

	tests := []struct {
		name      string
		keySize   int
		valueSize int
	}{
		{
			name:      "Empty key data",
			keySize:   0,
			valueSize: 4,
		},
		{
			name:      "Too small key data",
			keySize:   7,
			valueSize: 4,
		},
		{
			name:      "Empty value data",
			keySize:   8,
			valueSize: 0,
		},
		{
			name:      "Too small value data",
			keySize:   8,
			valueSize: 3,
		},
		{
			name:      "Both too small",
			keySize:   4,
			valueSize: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyData := make([]byte, tt.keySize)
			valueData := make([]byte, tt.valueSize)

			_, err := NewDataStreamIDReader(keyData, valueData, endian)
			if err == nil {
				t.Error("NewDataStreamIDReader() expected error, got nil")
			}
		})
	}
}
