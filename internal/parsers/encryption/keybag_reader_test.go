package encryption

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestKeybagData creates test keybag data with specified entries
func createTestKeybagData(version uint16, entries []TestKeybagEntry, endian binary.ByteOrder) []byte {
	// Calculate total size needed
	headerSize := 16 // version(2) + nkeys(2) + nbytes(4) + padding(8)
	totalDataSize := 0

	for _, entry := range entries {
		entrySize := 16 + 2 + 2 + 4 + len(entry.KeyData) // UUID + tag + keylen + padding + data
		totalDataSize += entrySize
	}

	data := make([]byte, headerSize+totalDataSize)
	offset := 0

	// Write header
	endian.PutUint16(data[offset:offset+2], version)
	offset += 2
	endian.PutUint16(data[offset:offset+2], uint16(len(entries)))
	offset += 2
	endian.PutUint32(data[offset:offset+4], uint32(totalDataSize))
	offset += 4

	// Skip padding (8 bytes)
	offset += 8

	// Write entries
	for _, entry := range entries {
		// UUID
		copy(data[offset:offset+16], entry.UUID[:])
		offset += 16

		// Tag
		endian.PutUint16(data[offset:offset+2], uint16(entry.Tag))
		offset += 2

		// Key length
		endian.PutUint16(data[offset:offset+2], uint16(len(entry.KeyData)))
		offset += 2

		// Padding (4 bytes)
		offset += 4

		// Key data
		copy(data[offset:offset+len(entry.KeyData)], entry.KeyData)
		offset += len(entry.KeyData)
	}

	return data
}

type TestKeybagEntry struct {
	UUID    types.UUID
	Tag     types.KbTag
	KeyData []byte
}

func TestNewKeybagReader(t *testing.T) {
	endian := binary.LittleEndian

	tests := []struct {
		name          string
		version       uint16
		entries       []TestKeybagEntry
		expectError   bool
		expectedError string
	}{
		{
			name:    "Valid keybag with volume key",
			version: types.ApfsKeybagVersion,
			entries: []TestKeybagEntry{
				{
					UUID:    types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
					Tag:     types.KbTagVolumeKey,
					KeyData: []byte{0xDE, 0xAD, 0xBE, 0xEF, 0xCA, 0xFE, 0xBA, 0xBE},
				},
			},
			expectError: false,
		},
		{
			name:    "Valid keybag with multiple entries",
			version: types.ApfsKeybagVersion,
			entries: []TestKeybagEntry{
				{
					UUID:    types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
					Tag:     types.KbTagVolumeKey,
					KeyData: []byte{0xDE, 0xAD, 0xBE, 0xEF},
				},
				{
					UUID:    types.ApfsFvPersonalRecoveryKeyUuid,
					Tag:     types.KbTagVolumeUnlockRecords,
					KeyData: []byte{0xFE, 0xED, 0xFA, 0xCE},
				},
			},
			expectError: false,
		},
		{
			name:    "Valid keybag with personal recovery key",
			version: types.ApfsKeybagVersion,
			entries: []TestKeybagEntry{
				{
					UUID:    types.ApfsFvPersonalRecoveryKeyUuid,
					Tag:     types.KbTagVolumeUnlockRecords,
					KeyData: []byte("recovery-key-data-here"),
				},
			},
			expectError: false,
		},
		{
			name:        "Empty keybag",
			version:     types.ApfsKeybagVersion,
			entries:     []TestKeybagEntry{},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := createTestKeybagData(tc.version, tc.entries, endian)

			reader, err := NewKeybagReader(data, endian)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if tc.expectedError != "" && err.Error() != tc.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tc.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify basic properties
			if reader.Version() != tc.version {
				t.Errorf("Version() = %d, want %d", reader.Version(), tc.version)
			}

			if reader.EntryCount() != uint16(len(tc.entries)) {
				t.Errorf("EntryCount() = %d, want %d", reader.EntryCount(), len(tc.entries))
			}

			if !reader.IsValid() {
				t.Errorf("IsValid() = false, want true")
			}

			// Verify entries
			entries := reader.ListEntries()
			if len(entries) != len(tc.entries) {
				t.Errorf("ListEntries() length = %d, want %d", len(entries), len(tc.entries))
			}

			for i, entry := range entries {
				expectedEntry := tc.entries[i]

				if entry.UUID() != expectedEntry.UUID {
					t.Errorf("Entry[%d].UUID() = %v, want %v", i, entry.UUID(), expectedEntry.UUID)
				}

				if entry.Tag() != expectedEntry.Tag {
					t.Errorf("Entry[%d].Tag() = %v, want %v", i, entry.Tag(), expectedEntry.Tag)
				}

				if entry.KeyLength() != uint16(len(expectedEntry.KeyData)) {
					t.Errorf("Entry[%d].KeyLength() = %d, want %d", i, entry.KeyLength(), len(expectedEntry.KeyData))
				}

				keyData := entry.KeyData()
				if len(keyData) != len(expectedEntry.KeyData) {
					t.Errorf("Entry[%d].KeyData() length = %d, want %d", i, len(keyData), len(expectedEntry.KeyData))
				}

				for j, b := range keyData {
					if b != expectedEntry.KeyData[j] {
						t.Errorf("Entry[%d].KeyData()[%d] = 0x%02X, want 0x%02X", i, j, b, expectedEntry.KeyData[j])
					}
				}
			}
		})
	}
}

func TestKeybagReaderErrors(t *testing.T) {
	endian := binary.LittleEndian

	tests := []struct {
		name        string
		dataSize    int
		description string
	}{
		{"Empty data", 0, "insufficient data for keybag"},
		{"Too small header", 15, "insufficient data for keybag"},
		{"Header only", 16, "should work with no entries"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := make([]byte, tc.dataSize)
			if tc.dataSize >= 2 {
				endian.PutUint16(data[0:2], types.ApfsKeybagVersion)
			}

			_, err := NewKeybagReader(data, endian)

			if tc.dataSize < 16 {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for valid header: %v", err)
				}
			}
		})
	}
}

func TestKeybagEntryReaderMethods(t *testing.T) {
	endian := binary.LittleEndian

	tests := []struct {
		name                            string
		entry                           TestKeybagEntry
		expectedIsPersonalKey           bool
		expectedIsInstitutionalKey      bool
		expectedIsInstitutionalUser     bool
		expectedIsVolumeKey             bool
		expectedIsUnlockRecord          bool
		expectedTagDescription          string
	}{
		{
			name: "Volume key entry",
			entry: TestKeybagEntry{
				UUID:    types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
				Tag:     types.KbTagVolumeKey,
				KeyData: []byte{0xDE, 0xAD, 0xBE, 0xEF},
			},
			expectedIsPersonalKey:       false,
			expectedIsInstitutionalKey:  false,
			expectedIsInstitutionalUser: false,
			expectedIsVolumeKey:         true,
			expectedIsUnlockRecord:      false,
			expectedTagDescription:      "Volume Key",
		},
		{
			name: "Personal recovery key entry",
			entry: TestKeybagEntry{
				UUID:    types.ApfsFvPersonalRecoveryKeyUuid,
				Tag:     types.KbTagVolumeUnlockRecords,
				KeyData: []byte("recovery-key"),
			},
			expectedIsPersonalKey:       true,
			expectedIsInstitutionalKey:  false,
			expectedIsInstitutionalUser: false,
			expectedIsVolumeKey:         false,
			expectedIsUnlockRecord:      true,
			expectedTagDescription:      "Volume Unlock Records",
		},
		{
			name: "Unlock records entry",
			entry: TestKeybagEntry{
				UUID:    types.UUID{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x00},
				Tag:     types.KbTagVolumeUnlockRecords,
				KeyData: []byte{0x12, 0x34},
			},
			expectedIsPersonalKey:       false,
			expectedIsInstitutionalKey:  false,
			expectedIsInstitutionalUser: false,
			expectedIsVolumeKey:         false,
			expectedIsUnlockRecord:      true,
			expectedTagDescription:      "Volume Unlock Records",
		},
		{
			name: "Institutional recovery key entry",
			entry: TestKeybagEntry{
				UUID:    types.ApfsFvInstitutionalRecoveryKeyUuid,
				Tag:     types.KbTagVolumeUnlockRecords,
				KeyData: []byte("institutional-key"),
			},
			expectedIsPersonalKey:       false,
			expectedIsInstitutionalKey:  true,
			expectedIsInstitutionalUser: false,
			expectedIsVolumeKey:         false,
			expectedIsUnlockRecord:      true,
			expectedTagDescription:      "Volume Unlock Records",
		},
		{
			name: "Institutional user entry",
			entry: TestKeybagEntry{
				UUID:    types.ApfsFvInstitutionalUserUuid,
				Tag:     types.KbTagVolumeUnlockRecords,
				KeyData: []byte("institutional-user-key"),
			},
			expectedIsPersonalKey:       false,
			expectedIsInstitutionalKey:  false,
			expectedIsInstitutionalUser: true,
			expectedIsVolumeKey:         false,
			expectedIsUnlockRecord:      true,
			expectedTagDescription:      "Volume Unlock Records",
		},
		{
			name: "Passphrase hint entry",
			entry: TestKeybagEntry{
				UUID:    types.UUID{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99},
				Tag:     types.KbTagVolumePassphraseHint,
				KeyData: []byte("password hint"),
			},
			expectedIsPersonalKey:       false,
			expectedIsInstitutionalKey:  false,
			expectedIsInstitutionalUser: false,
			expectedIsVolumeKey:         false,
			expectedIsUnlockRecord:      false,
			expectedTagDescription:      "Volume Passphrase Hint",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := createTestKeybagData(types.ApfsKeybagVersion, []TestKeybagEntry{tc.entry}, endian)

			reader, err := NewKeybagReader(data, endian)
			if err != nil {
				t.Fatalf("Failed to create keybag reader: %v", err)
			}

			entries := reader.ListEntries()
			if len(entries) != 1 {
				t.Fatalf("Expected 1 entry, got %d", len(entries))
			}

			entry := entries[0]

			if entry.IsPersonalRecoveryKey() != tc.expectedIsPersonalKey {
				t.Errorf("IsPersonalRecoveryKey() = %t, want %t", entry.IsPersonalRecoveryKey(), tc.expectedIsPersonalKey)
			}

			if entry.IsInstitutionalRecoveryKey() != tc.expectedIsInstitutionalKey {
				t.Errorf("IsInstitutionalRecoveryKey() = %t, want %t", entry.IsInstitutionalRecoveryKey(), tc.expectedIsInstitutionalKey)
			}

			if entry.IsInstitutionalUser() != tc.expectedIsInstitutionalUser {
				t.Errorf("IsInstitutionalUser() = %t, want %t", entry.IsInstitutionalUser(), tc.expectedIsInstitutionalUser)
			}

			if entry.IsVolumeKey() != tc.expectedIsVolumeKey {
				t.Errorf("IsVolumeKey() = %t, want %t", entry.IsVolumeKey(), tc.expectedIsVolumeKey)
			}

			if entry.IsUnlockRecord() != tc.expectedIsUnlockRecord {
				t.Errorf("IsUnlockRecord() = %t, want %t", entry.IsUnlockRecord(), tc.expectedIsUnlockRecord)
			}

			if entry.TagDescription() != tc.expectedTagDescription {
				t.Errorf("TagDescription() = '%s', want '%s'", entry.TagDescription(), tc.expectedTagDescription)
			}
		})
	}
}

func TestKeybagReaderInvalidVersion(t *testing.T) {
	endian := binary.LittleEndian

	// Test with version 1 (should fail)
	data := createTestKeybagData(1, []TestKeybagEntry{}, endian)

	_, err := NewKeybagReader(data, endian)
	if err == nil {
		t.Errorf("Expected error for version 1, but got none")
	}
}

func TestKeybagReaderTotalDataSize(t *testing.T) {
	endian := binary.LittleEndian

	entries := []TestKeybagEntry{
		{
			UUID:    types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
			Tag:     types.KbTagVolumeKey,
			KeyData: []byte{0xDE, 0xAD, 0xBE, 0xEF}, // 4 bytes
		},
		{
			UUID:    types.UUID{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x00},
			Tag:     types.KbTagVolumeUnlockRecords,
			KeyData: []byte{0xFE, 0xED, 0xFA, 0xCE, 0xBA, 0xBE}, // 6 bytes
		},
	}

	data := createTestKeybagData(types.ApfsKeybagVersion, entries, endian)
	reader, err := NewKeybagReader(data, endian)
	if err != nil {
		t.Fatalf("Failed to create keybag reader: %v", err)
	}

	// Each entry has: 16 (UUID) + 2 (tag) + 2 (keylen) + 4 (padding) + keylen = 24 + keylen
	// Entry 1: 24 + 4 = 28 bytes
	// Entry 2: 24 + 6 = 30 bytes
	// Total: 58 bytes
	expectedTotalSize := uint32(58)

	if reader.TotalDataSize() != expectedTotalSize {
		t.Errorf("TotalDataSize() = %d, want %d", reader.TotalDataSize(), expectedTotalSize)
	}
}
