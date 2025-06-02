package file_system_objects

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestNewExtendedAttributeReader(t *testing.T) {
	tests := []struct {
		name          string
		attributeName string
		objectID      uint64
		flags         types.JXattrFlags
		data          []byte
		expectError   bool
	}{
		{
			name:          "embedded attribute",
			attributeName: "com.apple.metadata",
			objectID:      123,
			flags:         types.XattrDataEmbedded,
			data:          []byte("test data"),
		},
		{
			name:          "stream attribute",
			attributeName: "com.apple.resourcefork",
			objectID:      456,
			flags:         types.XattrDataStream,
			data:          []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, // 8-byte stream ID
		},
		{
			name:          "filesystem owned attribute",
			attributeName: "com.apple.system",
			objectID:      789,
			flags:         types.XattrDataEmbedded | types.XattrFileSystemOwned,
			data:          []byte("system data"),
		},
		{
			name:          "empty attribute name",
			attributeName: "",
			objectID:      101112,
			flags:         types.XattrDataEmbedded,
			data:          []byte("data"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyData, valueData := createExtendedAttributeTestData(tt.attributeName, tt.objectID, tt.flags, tt.data)

			reader, err := NewExtendedAttributeReader(keyData, valueData, binary.LittleEndian)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewExtendedAttributeReader failed: %v", err)
			}

			// Test FileSystemObjectReader methods
			if reader.ObjectIdentifier() != tt.objectID {
				t.Errorf("Expected object identifier %d, got %d", tt.objectID, reader.ObjectIdentifier())
			}

			if reader.ObjectType() != types.ApfsTypeXattr {
				t.Errorf("Expected object type %d, got %d", types.ApfsTypeXattr, reader.ObjectType())
			}

			// Test ExtendedAttributeReader methods
			if reader.AttributeName() != tt.attributeName {
				t.Errorf("Expected attribute name %q, got %q", tt.attributeName, reader.AttributeName())
			}

			// Test flag methods
			expectedEmbedded := tt.flags&types.XattrDataEmbedded != 0
			if reader.IsDataEmbedded() != expectedEmbedded {
				t.Errorf("Expected IsDataEmbedded %v, got %v", expectedEmbedded, reader.IsDataEmbedded())
			}

			expectedStream := tt.flags&types.XattrDataStream != 0
			if reader.IsDataStream() != expectedStream {
				t.Errorf("Expected IsDataStream %v, got %v", expectedStream, reader.IsDataStream())
			}

			expectedFSOwned := tt.flags&types.XattrFileSystemOwned != 0
			if reader.IsFileSystemOwned() != expectedFSOwned {
				t.Errorf("Expected IsFileSystemOwned %v, got %v", expectedFSOwned, reader.IsFileSystemOwned())
			}

			// Test data retrieval
			data := reader.Data()
			if reader.IsDataEmbedded() {
				// For embedded data, should get the actual data
				expectedLen := len(tt.data)
				if len(data) != expectedLen {
					t.Errorf("Expected data length %d, got %d", expectedLen, len(data))
				}
			} else {
				// For stream data, should get the stream ID
				if len(data) != len(tt.data) {
					t.Errorf("Expected stream ID length %d, got %d", len(tt.data), len(data))
				}
			}
		})
	}
}

func TestExtendedAttributeReaderErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		keyData   []byte
		valueData []byte
	}{
		{
			name:      "insufficient key data",
			keyData:   make([]byte, 6),
			valueData: make([]byte, 8),
		},
		{
			name:      "insufficient value data",
			keyData:   make([]byte, 12),
			valueData: make([]byte, 2),
		},
		{
			name:      "name length exceeds data",
			keyData:   createMalformedXattrKeyData(),
			valueData: make([]byte, 8),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewExtendedAttributeReader(tt.keyData, tt.valueData, binary.LittleEndian)
			if err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

func TestExtendedAttributeReaderNullTermination(t *testing.T) {
	// Test that null-terminated strings are handled correctly
	attributeName := "test.attr\x00\x00"
	keyData, valueData := createExtendedAttributeTestData(attributeName, 123, types.XattrDataEmbedded, []byte("data"))

	reader, err := NewExtendedAttributeReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewExtendedAttributeReader failed: %v", err)
	}

	// Should trim null terminators
	if reader.AttributeName() != "test.attr" {
		t.Errorf("Expected attribute name %q, got %q", "test.attr", reader.AttributeName())
	}
}

func TestExtendedAttributeDataLengthLimiting(t *testing.T) {
	// Test that embedded data is limited by xdata_len
	data := []byte("this is a long test data string")
	limitedLen := uint16(10)

	keyData, valueData := createExtendedAttributeTestDataWithLength("test.attr", 123, types.XattrDataEmbedded, data, limitedLen)

	reader, err := NewExtendedAttributeReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewExtendedAttributeReader failed: %v", err)
	}

	retrievedData := reader.Data()
	if len(retrievedData) != int(limitedLen) {
		t.Errorf("Expected data length %d, got %d", limitedLen, len(retrievedData))
	}

	expectedData := data[:limitedLen]
	for i, b := range retrievedData {
		if b != expectedData[i] {
			t.Errorf("Data mismatch at index %d: expected %v, got %v", i, expectedData[i], b)
		}
	}
}

func createExtendedAttributeTestData(attributeName string, objectID uint64, flags types.JXattrFlags, data []byte) ([]byte, []byte) {
	return createExtendedAttributeTestDataWithLength(attributeName, objectID, flags, data, uint16(len(data)))
}

func createExtendedAttributeTestDataWithLength(attributeName string, objectID uint64, flags types.JXattrFlags, data []byte, dataLen uint16) ([]byte, []byte) {
	// Create key data
	nameBytes := []byte(attributeName)
	keyData := make([]byte, 8+2+len(nameBytes))

	// Header
	objIdAndType := (uint64(types.ApfsTypeXattr) << types.ObjTypeShift) | objectID
	binary.LittleEndian.PutUint64(keyData[0:8], objIdAndType)

	// Name length
	binary.LittleEndian.PutUint16(keyData[8:10], uint16(len(nameBytes)))

	// Name
	copy(keyData[10:], nameBytes)

	// Create value data
	valueData := make([]byte, 4+len(data)) // 2 + 2 + data
	offset := 0

	// Flags
	binary.LittleEndian.PutUint16(valueData[offset:offset+2], uint16(flags))
	offset += 2

	// Data length
	binary.LittleEndian.PutUint16(valueData[offset:offset+2], dataLen)
	offset += 2

	// Data
	copy(valueData[offset:], data)

	return keyData, valueData
}

func createMalformedXattrKeyData() []byte {
	keyData := make([]byte, 12)

	// Header
	objIdAndType := (uint64(types.ApfsTypeXattr) << types.ObjTypeShift) | 123
	binary.LittleEndian.PutUint64(keyData[0:8], objIdAndType)

	// Name length that exceeds available data
	binary.LittleEndian.PutUint16(keyData[8:10], 100)

	return keyData
}
