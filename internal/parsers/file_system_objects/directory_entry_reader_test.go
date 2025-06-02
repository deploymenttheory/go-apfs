package file_system_objects

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestNewDirectoryEntryReader(t *testing.T) {
	tests := []struct {
		name        string
		isHashed    bool
		fileName    string
		fileID      uint64
		objectID    uint64
		expectError bool
	}{
		{
			name:     "basic directory entry",
			isHashed: false,
			fileName: "test.txt",
			fileID:   123,
			objectID: 456,
		},
		{
			name:     "hashed directory entry",
			isHashed: true,
			fileName: "test_file.dat",
			fileID:   789,
			objectID: 101112,
		},
		{
			name:     "empty filename",
			isHashed: false,
			fileName: "",
			fileID:   1,
			objectID: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyData, valueData := createDirectoryEntryTestData(tt.isHashed, tt.fileName, tt.fileID, tt.objectID)

			reader, err := NewDirectoryEntryReader(keyData, valueData, binary.LittleEndian, tt.isHashed)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewDirectoryEntryReader failed: %v", err)
			}

			// Test FileSystemObjectReader methods
			if reader.ObjectIdentifier() != tt.objectID {
				t.Errorf("Expected object identifier %d, got %d", tt.objectID, reader.ObjectIdentifier())
			}

			if reader.ObjectType() != types.ApfsTypeDirRec {
				t.Errorf("Expected object type %d, got %d", types.ApfsTypeDirRec, reader.ObjectType())
			}

			// Test DirectoryEntryReader methods
			if reader.FileName() != tt.fileName {
				t.Errorf("Expected filename %q, got %q", tt.fileName, reader.FileName())
			}

			if reader.FileID() != tt.fileID {
				t.Errorf("Expected file ID %d, got %d", tt.fileID, reader.FileID())
			}

			// Test that DateAdded returns a valid time
			dateAdded := reader.DateAdded()
			if dateAdded.IsZero() {
				t.Error("DateAdded should not be zero")
			}
		})
	}
}

func TestDirectoryEntryReaderErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		keyData   []byte
		valueData []byte
		isHashed  bool
	}{
		{
			name:      "insufficient key data",
			keyData:   make([]byte, 4),
			valueData: make([]byte, 18),
			isHashed:  false,
		},
		{
			name:      "insufficient value data",
			keyData:   make([]byte, 12),
			valueData: make([]byte, 10),
			isHashed:  false,
		},
		{
			name:      "name length exceeds data",
			keyData:   createMalformedKeyData(),
			valueData: make([]byte, 18),
			isHashed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDirectoryEntryReader(tt.keyData, tt.valueData, binary.LittleEndian, tt.isHashed)
			if err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

func TestDirectoryEntryReaderNullTermination(t *testing.T) {
	// Test that null-terminated strings are handled correctly
	fileName := "test\x00\x00"
	keyData, valueData := createDirectoryEntryTestData(false, fileName, 123, 456)

	reader, err := NewDirectoryEntryReader(keyData, valueData, binary.LittleEndian, false)
	if err != nil {
		t.Fatalf("NewDirectoryEntryReader failed: %v", err)
	}

	// Should trim null terminators
	if reader.FileName() != "test" {
		t.Errorf("Expected filename %q, got %q", "test", reader.FileName())
	}
}

func createDirectoryEntryTestData(isHashed bool, fileName string, fileID, objectID uint64) ([]byte, []byte) {
	// Create key data
	var keyData []byte
	if isHashed {
		keyData = createHashedDirectoryKeyData(fileName, objectID)
	} else {
		keyData = createDirectoryKeyData(fileName, objectID)
	}

	// Create value data
	valueData := createDirectoryValueData(fileID)

	return keyData, valueData
}

func createDirectoryKeyData(fileName string, objectID uint64) []byte {
	nameBytes := []byte(fileName)
	keyData := make([]byte, 8+2+len(nameBytes))

	// Header
	objIdAndType := (uint64(types.ApfsTypeDirRec) << types.ObjTypeShift) | objectID
	binary.LittleEndian.PutUint64(keyData[0:8], objIdAndType)

	// Name length
	binary.LittleEndian.PutUint16(keyData[8:10], uint16(len(nameBytes)))

	// Name
	copy(keyData[10:], nameBytes)

	return keyData
}

func createHashedDirectoryKeyData(fileName string, objectID uint64) []byte {
	nameBytes := []byte(fileName)
	keyData := make([]byte, 8+4+len(nameBytes))

	// Header
	objIdAndType := (uint64(types.ApfsTypeDirRec) << types.ObjTypeShift) | objectID
	binary.LittleEndian.PutUint64(keyData[0:8], objIdAndType)

	// Name length and hash (simplified hash)
	nameLen := uint32(len(nameBytes))
	hash := uint32(0x12345) << types.JDrecHashShift // Simple test hash
	nameLenAndHash := (hash & types.JDrecHashMask) | (nameLen & types.JDrecLenMask)
	binary.LittleEndian.PutUint32(keyData[8:12], nameLenAndHash)

	// Name
	copy(keyData[12:], nameBytes)

	return keyData
}

func createDirectoryValueData(fileID uint64) []byte {
	valueData := make([]byte, 18) // 8 + 8 + 2

	// File ID
	binary.LittleEndian.PutUint64(valueData[0:8], fileID)

	// Date added (current time in nanoseconds)
	dateAdded := uint64(time.Now().UnixNano())
	binary.LittleEndian.PutUint64(valueData[8:16], dateAdded)

	// Flags
	binary.LittleEndian.PutUint16(valueData[16:18], 0x4000) // Directory flag

	return valueData
}

func createMalformedKeyData() []byte {
	keyData := make([]byte, 12)

	// Header
	objIdAndType := (uint64(types.ApfsTypeDirRec) << types.ObjTypeShift) | 123
	binary.LittleEndian.PutUint64(keyData[0:8], objIdAndType)

	// Name length that exceeds available data
	binary.LittleEndian.PutUint16(keyData[8:10], 100)

	return keyData
}
