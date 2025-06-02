package file_system_objects

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestNewDirectoryStatsReader(t *testing.T) {
	tests := []struct {
		name        string
		objectID    uint64
		numChildren uint64
		totalSize   uint64
		chainedKey  uint64
		genCount    uint64
		expectError bool
	}{
		{
			name:        "basic directory stats",
			objectID:    123,
			numChildren: 5,
			totalSize:   1024,
			chainedKey:  456,
			genCount:    1,
		},
		{
			name:        "empty directory",
			objectID:    789,
			numChildren: 0,
			totalSize:   0,
			chainedKey:  999,
			genCount:    2,
		},
		{
			name:        "large directory",
			objectID:    101112,
			numChildren: 1000000,
			totalSize:   1099511627776, // 1TB
			chainedKey:  131415,
			genCount:    999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyData, valueData := createDirectoryStatsTestData(tt.objectID, tt.numChildren, tt.totalSize, tt.chainedKey, tt.genCount)

			reader, err := NewDirectoryStatsReader(keyData, valueData, binary.LittleEndian)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewDirectoryStatsReader failed: %v", err)
			}

			// Test basic properties
			if reader.ObjectIdentifier() != tt.objectID {
				t.Errorf("Expected object identifier %d, got %d", tt.objectID, reader.ObjectIdentifier())
			}

			if reader.ObjectType() != types.ApfsTypeDirStats {
				t.Errorf("Expected object type %d, got %d", types.ApfsTypeDirStats, reader.ObjectType())
			}

			// Test directory stats specific methods
			if reader.NumChildren() != tt.numChildren {
				t.Errorf("Expected num children %d, got %d", tt.numChildren, reader.NumChildren())
			}

			if reader.TotalSize() != tt.totalSize {
				t.Errorf("Expected total size %d, got %d", tt.totalSize, reader.TotalSize())
			}

			if reader.ChainedKey() != tt.chainedKey {
				t.Errorf("Expected chained key %d, got %d", tt.chainedKey, reader.ChainedKey())
			}

			if reader.GenCount() != tt.genCount {
				t.Errorf("Expected gen count %d, got %d", tt.genCount, reader.GenCount())
			}
		})
	}
}

func TestDirectoryStatsReaderErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		keyData   []byte
		valueData []byte
	}{
		{
			name:      "insufficient key data",
			keyData:   make([]byte, 4),
			valueData: make([]byte, 32),
		},
		{
			name:      "insufficient value data",
			keyData:   make([]byte, 8),
			valueData: make([]byte, 16),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDirectoryStatsReader(tt.keyData, tt.valueData, binary.LittleEndian)
			if err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

func createDirectoryStatsTestData(objectID, numChildren, totalSize, chainedKey, genCount uint64) ([]byte, []byte) {
	// Create key data
	keyData := make([]byte, 8)
	objIdAndType := (uint64(types.ApfsTypeDirStats) << types.ObjTypeShift) | objectID
	binary.LittleEndian.PutUint64(keyData[0:8], objIdAndType)

	// Create value data
	valueData := make([]byte, 32) // 4 * 8 bytes
	offset := 0

	binary.LittleEndian.PutUint64(valueData[offset:offset+8], numChildren)
	offset += 8

	binary.LittleEndian.PutUint64(valueData[offset:offset+8], totalSize)
	offset += 8

	binary.LittleEndian.PutUint64(valueData[offset:offset+8], chainedKey)
	offset += 8

	binary.LittleEndian.PutUint64(valueData[offset:offset+8], genCount)

	return keyData, valueData
}
