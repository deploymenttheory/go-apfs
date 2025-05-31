package encryption

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestNewMediaKeybagReader(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		endian      binary.ByteOrder
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Insufficient data",
			data:        make([]byte, 30), // Less than minimum 48 bytes
			endian:      binary.LittleEndian,
			expectError: true,
			errorMsg:    "insufficient data for media keybag",
		},
		{
			name:        "Nil endian (should default to LittleEndian)",
			data:        createValidMediaKeybagData(),
			endian:      nil,
			expectError: false,
		},
		{
			name:        "Valid media keybag data",
			data:        createValidMediaKeybagData(),
			endian:      binary.LittleEndian,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewMediaKeybagReader(tt.data, tt.endian)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if reader == nil {
					t.Errorf("Expected reader but got nil")
				}
			}
		})
	}
}

func TestMediaKeybagReaderVersion(t *testing.T) {
	data := createValidMediaKeybagData()
	reader, err := NewMediaKeybagReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	version := reader.Version()
	expectedVersion := types.ApfsKeybagVersion

	if version != expectedVersion {
		t.Errorf("Expected version %d, got %d", expectedVersion, version)
	}
}

func TestMediaKeybagReaderEntryCount(t *testing.T) {
	data := createValidMediaKeybagData()
	reader, err := NewMediaKeybagReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	count := reader.EntryCount()
	expectedCount := uint16(1) // Our test data has 1 entry

	if count != expectedCount {
		t.Errorf("Expected entry count %d, got %d", expectedCount, count)
	}
}

func TestMediaKeybagReaderTotalDataSize(t *testing.T) {
	data := createValidMediaKeybagData()
	reader, err := NewMediaKeybagReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	size := reader.TotalDataSize()
	// Test data size is calculated based on our test entry (24 bytes header + 32 bytes key data)
	expectedSize := uint32(56)

	if size != expectedSize {
		t.Errorf("Expected total data size %d, got %d", expectedSize, size)
	}
}

func TestMediaKeybagReaderListEntries(t *testing.T) {
	data := createValidMediaKeybagData()
	reader, err := NewMediaKeybagReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	entries := reader.ListEntries()
	expectedCount := 1

	if len(entries) != expectedCount {
		t.Errorf("Expected %d entries, got %d", expectedCount, len(entries))
	}

	if len(entries) > 0 {
		entry := entries[0]
		if entry.Tag() != types.KbTagVolumeMKey {
			t.Errorf("Expected tag %d, got %d", types.KbTagVolumeMKey, entry.Tag())
		}
	}
}

func TestMediaKeybagReaderIsValid(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
		expected    bool
	}{
		{
			name:        "Valid media keybag",
			data:        createValidMediaKeybagData(),
			expectError: false,
			expected:    true,
		},
		{
			name:        "Invalid object type",
			data:        createInvalidMediaKeybagData(invalidObjectType),
			expectError: false,
			expected:    false,
		},
		{
			name:        "Invalid keybag version",
			data:        createInvalidMediaKeybagData(invalidKeybagVersion),
			expectError: true, // Reader creation should fail
			expected:    false,
		},
		{
			name:        "Mismatched entry count",
			data:        createInvalidMediaKeybagData(mismatchedEntryCount),
			expectError: false, // Creation should succeed but IsValid should return false
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewMediaKeybagReader(tt.data, binary.LittleEndian)
			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}
				return // Don't test IsValid() if reader creation failed as expected
			}

			if err != nil {
				t.Fatalf("Failed to create reader: %v", err)
			}

			result := reader.IsValid()
			if result != tt.expected {
				t.Errorf("Expected IsValid() = %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMediaKeybagReaderObjectProperties(t *testing.T) {
	data := createValidMediaKeybagData()
	reader, err := NewMediaKeybagReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// Test ObjectID
	objectID := reader.(*mediaKeybagReader).ObjectID()
	expectedObjectID := types.OidT(0x123456789ABCDEF0)
	if objectID != expectedObjectID {
		t.Errorf("Expected ObjectID 0x%016X, got 0x%016X", expectedObjectID, objectID)
	}

	// Test TransactionID
	transactionID := reader.(*mediaKeybagReader).TransactionID()
	expectedTransactionID := types.XidT(0xFEDCBA9876543210)
	if transactionID != expectedTransactionID {
		t.Errorf("Expected TransactionID 0x%016X, got 0x%016X", expectedTransactionID, transactionID)
	}

	// Test ObjectType
	objectType := reader.(*mediaKeybagReader).ObjectType()
	expectedObjectType := types.ObjectTypeMediaKeybag
	if objectType != expectedObjectType {
		t.Errorf("Expected ObjectType 0x%08X, got 0x%08X", expectedObjectType, objectType)
	}

	// Test ObjectSubtype
	objectSubtype := reader.(*mediaKeybagReader).ObjectSubtype()
	expectedObjectSubtype := uint32(0)
	if objectSubtype != expectedObjectSubtype {
		t.Errorf("Expected ObjectSubtype 0x%08X, got 0x%08X", expectedObjectSubtype, objectSubtype)
	}
}

// Helper functions for test data creation

type invalidationType int

const (
	invalidObjectType invalidationType = iota
	invalidKeybagVersion
	mismatchedEntryCount
)

func createValidMediaKeybagData() []byte {
	data := make([]byte, 128) // Increased from 120 to ensure enough space
	offset := 0

	// Object header (32 bytes)
	// Checksum (8 bytes) - zeros
	offset += 8

	// Object ID
	binary.LittleEndian.PutUint64(data[offset:], uint64(0x123456789ABCDEF0))
	offset += 8

	// Transaction ID
	binary.LittleEndian.PutUint64(data[offset:], uint64(0xFEDCBA9876543210))
	offset += 8

	// Object Type
	binary.LittleEndian.PutUint32(data[offset:], types.ObjectTypeMediaKeybag)
	offset += 4

	// Object Subtype
	binary.LittleEndian.PutUint32(data[offset:], 0)
	offset += 4

	// Keybag header (16 bytes)
	// Version
	binary.LittleEndian.PutUint16(data[offset:], types.ApfsKeybagVersion)
	offset += 2

	// Number of keys
	binary.LittleEndian.PutUint16(data[offset:], 1)
	offset += 2

	// Number of bytes - should be exactly the size needed for one entry (56 bytes)
	binary.LittleEndian.PutUint32(data[offset:], 56) // UUID(16) + tag(2) + keylen(2) + padding(4) + keydata(32) = 56
	offset += 4

	// Padding (8 bytes)
	offset += 8

	// Keybag entry (56 bytes total)
	// UUID (16 bytes) - test UUID
	for i := 0; i < 16; i++ {
		data[offset+i] = byte(i)
	}
	offset += 16

	// Tag
	binary.LittleEndian.PutUint16(data[offset:], uint16(types.KbTagVolumeMKey))
	offset += 2

	// Key length
	binary.LittleEndian.PutUint16(data[offset:], 32)
	offset += 2

	// Padding (4 bytes)
	offset += 4

	// Key data (32 bytes)
	for i := 0; i < 32; i++ {
		data[offset+i] = byte(i + 100)
	}
	offset += 32

	// Ensure we have used exactly the expected amount of data
	// Total used should be: obj_header(32) + keybag_header(16) + entry(56) = 104
	if offset != 104 {
		panic(fmt.Sprintf("Expected offset 104, got %d", offset))
	}

	return data
}

func createInvalidMediaKeybagData(invalidation invalidationType) []byte {
	data := createValidMediaKeybagData()

	switch invalidation {
	case invalidObjectType:
		// Change object type to something invalid
		binary.LittleEndian.PutUint32(data[24:], 0xDEADBEEF)
	case invalidKeybagVersion:
		// Change keybag version to something too low
		binary.LittleEndian.PutUint16(data[32:], 1)
	case mismatchedEntryCount:
		// Change entry count to mismatch actual entries
		binary.LittleEndian.PutUint16(data[34:], 2) // Say we have 2 entries but only provide 1
		// Set a non-zero key length for the second entry to ensure validation fails
		// The second entry starts at offset 48 + 56 = 104 in the data
		// Key length is at offset 104 + 16 + 2 = 122 (UUID + tag + keylen)
		if len(data) > 122 {
			binary.LittleEndian.PutUint16(data[122:], 32) // Set key length to 32 for second entry
		}
	}

	return data
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		(len(s) > len(substr) && containsString(s[1:], substr))
}
