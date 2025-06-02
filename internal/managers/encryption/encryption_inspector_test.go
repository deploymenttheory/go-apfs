package encryption

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	cryptoParser "github.com/deploymenttheory/go-apfs/internal/parsers/encryption"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestCryptoData creates test crypto key and value data - copied from parser package
func createTestCryptoData(objID types.OidT, objType uint32, refCount uint32, protectionClass types.CpKeyClassT, keyVersion types.CpKeyRevisionT, keyLen uint16, endian binary.ByteOrder) ([]byte, []byte) {
	// Create key data (8 bytes for JKeyT)
	keyData := make([]byte, 8)
	objIdAndType := uint64(objID) | (uint64(objType) << types.ObjTypeShift)
	endian.PutUint64(keyData[0:8], objIdAndType)

	// Create value data
	valueSize := 4 + 20 + int(keyLen) // refcnt + wrapped state header + key data
	valueData := make([]byte, valueSize)
	offset := 0

	// Reference count
	endian.PutUint32(valueData[offset:offset+4], refCount)
	offset += 4

	// Wrapped crypto state
	endian.PutUint16(valueData[offset:offset+2], 5) // major version
	offset += 2
	endian.PutUint16(valueData[offset:offset+2], 0) // minor version
	offset += 2
	endian.PutUint32(valueData[offset:offset+4], 0) // flags
	offset += 4
	endian.PutUint32(valueData[offset:offset+4], uint32(protectionClass)) // protection class
	offset += 4
	endian.PutUint32(valueData[offset:offset+4], 0x12345678) // OS version
	offset += 4
	endian.PutUint16(valueData[offset:offset+2], uint16(keyVersion)) // key revision
	offset += 2
	endian.PutUint16(valueData[offset:offset+2], keyLen) // key length
	offset += 2

	// Key data (fill with test pattern)
	for i := uint16(0); i < keyLen; i++ {
		valueData[offset+int(i)] = byte(i % 256)
	}

	return keyData, valueData
}

func createMockCryptoStateReader(objID types.OidT, refCount uint32, protectionClass types.CpKeyClassT, keyVersion types.CpKeyRevisionT, keyLen uint16, valid bool) interfaces.CryptoStateReader {
	endian := binary.LittleEndian
	keyData, valueData := createTestCryptoData(objID, types.ObjectTypeTest, refCount, protectionClass, keyVersion, keyLen, endian)

	if !valid {
		// Corrupt the data to make it invalid
		if len(valueData) > 4 {
			valueData[0] = 0 // Set reference count to 0
		}
	}

	reader, _ := cryptoParser.NewCryptoStateReader(keyData, valueData, endian)
	return reader
}

func TestNewEncryptionInspector(t *testing.T) {
	mockReader := createMockCryptoStateReader(12345, 1, types.ProtectionClassC, 1, 32, true)

	inspector := NewEncryptionInspector(mockReader)
	if inspector == nil {
		t.Error("NewEncryptionInspector() returned nil")
	}
}

func TestEncryptionInspectorIsEncryptionEnabled(t *testing.T) {
	tests := []struct {
		name            string
		reader          interfaces.CryptoStateReader
		expectedEnabled bool
	}{
		{
			name:            "Valid encryption state",
			reader:          createMockCryptoStateReader(12345, 1, types.ProtectionClassC, 1, 32, true),
			expectedEnabled: true,
		},
		{
			name:            "Invalid encryption state",
			reader:          createMockCryptoStateReader(12345, 0, types.ProtectionClassC, 1, 32, false),
			expectedEnabled: false,
		},
		{
			name:            "Nil reader",
			reader:          nil,
			expectedEnabled: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			inspector := NewEncryptionInspector(tc.reader)
			enabled := inspector.IsEncryptionEnabled()

			if enabled != tc.expectedEnabled {
				t.Errorf("IsEncryptionEnabled() = %t, want %t", enabled, tc.expectedEnabled)
			}
		})
	}
}

func TestEncryptionInspectorGetCryptoIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		objID      types.OidT
		expectedID uint64
	}{
		{
			name:       "Valid object ID",
			objID:      12345,
			expectedID: 12345,
		},
		{
			name:       "Large object ID",
			objID:      0x123456789ABCDEF,
			expectedID: 0x123456789ABCDEF,
		},
		{
			name:       "Zero object ID",
			objID:      0,
			expectedID: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reader := createMockCryptoStateReader(tc.objID, 1, types.ProtectionClassC, 1, 32, true)
			inspector := NewEncryptionInspector(reader)

			cryptoID := inspector.GetCryptoIdentifier()
			if cryptoID != tc.expectedID {
				t.Errorf("GetCryptoIdentifier() = 0x%016X, want 0x%016X", cryptoID, tc.expectedID)
			}
		})
	}
}

func TestEncryptionInspectorGetCryptoIdentifierNilReader(t *testing.T) {
	inspector := NewEncryptionInspector(nil)
	cryptoID := inspector.GetCryptoIdentifier()

	if cryptoID != 0 {
		t.Errorf("GetCryptoIdentifier() with nil reader = %d, want 0", cryptoID)
	}
}

func TestEncryptionInspectorAnalyzeEncryptionState(t *testing.T) {
	tests := []struct {
		name            string
		objID           types.OidT
		refCount        uint32
		protectionClass types.CpKeyClassT
		keyVersion      types.CpKeyRevisionT
		keyLen          uint16
		valid           bool
		expectError     bool
	}{
		{
			name:            "Valid encryption state",
			objID:           12345,
			refCount:        1,
			protectionClass: types.ProtectionClassC,
			keyVersion:      1,
			keyLen:          32,
			valid:           true,
			expectError:     false,
		},
		{
			name:            "Complete protection class",
			objID:           67890,
			refCount:        2,
			protectionClass: types.ProtectionClassA,
			keyVersion:      3,
			keyLen:          64,
			valid:           true,
			expectError:     false,
		},
		{
			name:            "No protection class",
			objID:           11111,
			refCount:        1,
			protectionClass: types.ProtectionClassD,
			keyVersion:      1,
			keyLen:          16,
			valid:           true,
			expectError:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reader := createMockCryptoStateReader(tc.objID, tc.refCount, tc.protectionClass, tc.keyVersion, tc.keyLen, tc.valid)
			inspector := NewEncryptionInspector(reader)

			analysis, err := inspector.AnalyzeEncryptionState()

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify analysis fields
			if analysis.KeyVersion != uint16(tc.keyVersion) {
				t.Errorf("Analysis.KeyVersion = %d, want %d", analysis.KeyVersion, tc.keyVersion)
			}

			if analysis.IsValid != tc.valid {
				t.Errorf("Analysis.IsValid = %t, want %t", analysis.IsValid, tc.valid)
			}

			// Verify protection class name
			resolver := NewProtectionClassResolver()
			expectedProtectionClassName := resolver.ResolveName(tc.protectionClass)
			if analysis.ProtectionClass != expectedProtectionClassName {
				t.Errorf("Analysis.ProtectionClass = '%s', want '%s'", analysis.ProtectionClass, expectedProtectionClassName)
			}

			// Verify metadata fields
			if analysis.Metadata == nil {
				t.Error("Analysis.Metadata is nil")
				return
			}

			// Check required metadata fields
			requiredFields := []string{
				"protection_class_description",
				"effective_class",
				"security_level",
				"ios_only",
				"macos_only",
				"reference_count",
				"major_version",
				"minor_version",
				"key_length",
				"object_id",
				"object_type",
				"os_major_version",
				"os_minor_letter",
				"os_build_number",
				"os_version_raw",
				"crypto_flags",
			}

			for _, field := range requiredFields {
				if _, exists := analysis.Metadata[field]; !exists {
					t.Errorf("Missing required metadata field: %s", field)
				}
			}
		})
	}
}

func TestEncryptionInspectorAnalyzeEncryptionStateNilReader(t *testing.T) {
	inspector := NewEncryptionInspector(nil)

	_, err := inspector.AnalyzeEncryptionState()
	if err == nil {
		t.Error("Expected error for nil reader but got none")
	}
}

func TestNewKeyRollingManager(t *testing.T) {
	currentReader := createMockCryptoStateReader(12345, 1, types.ProtectionClassC, 2, 32, true)
	previousReader := createMockCryptoStateReader(12345, 1, types.ProtectionClassC, 1, 32, true)

	manager := NewKeyRollingManager(currentReader, previousReader)
	if manager == nil {
		t.Error("NewKeyRollingManager() returned nil")
	}
}

func TestKeyRollingManagerIsKeyRollingInProgress(t *testing.T) {
	tests := []struct {
		name               string
		currentKeyVersion  types.CpKeyRevisionT
		previousKeyVersion types.CpKeyRevisionT
		currentReader      interfaces.CryptoStateReader
		previousReader     interfaces.CryptoStateReader
		expectedInProgress bool
	}{
		{
			name:               "Key rolling in progress",
			currentKeyVersion:  2,
			previousKeyVersion: 1,
			expectedInProgress: true,
		},
		{
			name:               "No key rolling",
			currentKeyVersion:  1,
			previousKeyVersion: 1,
			expectedInProgress: false,
		},
		{
			name:               "Nil current reader",
			currentReader:      nil,
			previousReader:     createMockCryptoStateReader(12345, 1, types.ProtectionClassC, 1, 32, true),
			expectedInProgress: false,
		},
		{
			name:               "Nil previous reader",
			currentReader:      createMockCryptoStateReader(12345, 1, types.ProtectionClassC, 2, 32, true),
			previousReader:     nil,
			expectedInProgress: false,
		},
		{
			name:               "Both readers nil",
			currentReader:      nil,
			previousReader:     nil,
			expectedInProgress: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var currentReader, previousReader interfaces.CryptoStateReader

			if tc.currentReader != nil {
				currentReader = tc.currentReader
			} else if tc.name != "Nil current reader" && tc.name != "Both readers nil" {
				currentReader = createMockCryptoStateReader(12345, 1, types.ProtectionClassC, tc.currentKeyVersion, 32, true)
			}

			if tc.previousReader != nil {
				previousReader = tc.previousReader
			} else if tc.name != "Nil previous reader" && tc.name != "Both readers nil" {
				previousReader = createMockCryptoStateReader(12345, 1, types.ProtectionClassC, tc.previousKeyVersion, 32, true)
			}

			manager := NewKeyRollingManager(currentReader, previousReader)
			inProgress := manager.IsKeyRollingInProgress()

			if inProgress != tc.expectedInProgress {
				t.Errorf("IsKeyRollingInProgress() = %t, want %t", inProgress, tc.expectedInProgress)
			}
		})
	}
}

func TestKeyRollingManagerGetKeyVersions(t *testing.T) {
	currentVersion := types.CpKeyRevisionT(3)
	previousVersion := types.CpKeyRevisionT(2)

	currentReader := createMockCryptoStateReader(12345, 1, types.ProtectionClassC, currentVersion, 32, true)
	previousReader := createMockCryptoStateReader(12345, 1, types.ProtectionClassC, previousVersion, 32, true)

	manager := NewKeyRollingManager(currentReader, previousReader)

	if manager.GetCurrentKeyVersion() != currentVersion {
		t.Errorf("GetCurrentKeyVersion() = %d, want %d", manager.GetCurrentKeyVersion(), currentVersion)
	}

	if manager.GetPreviousKeyVersion() != previousVersion {
		t.Errorf("GetPreviousKeyVersion() = %d, want %d", manager.GetPreviousKeyVersion(), previousVersion)
	}
}

func TestKeyRollingManagerGetKeyVersionsNilReaders(t *testing.T) {
	tests := []struct {
		name                    string
		currentReader           interfaces.CryptoStateReader
		previousReader          interfaces.CryptoStateReader
		expectedCurrentVersion  types.CpKeyRevisionT
		expectedPreviousVersion types.CpKeyRevisionT
	}{
		{
			name:                    "Nil current reader",
			currentReader:           nil,
			previousReader:          createMockCryptoStateReader(12345, 1, types.ProtectionClassC, 2, 32, true),
			expectedCurrentVersion:  0,
			expectedPreviousVersion: 2,
		},
		{
			name:                    "Nil previous reader",
			currentReader:           createMockCryptoStateReader(12345, 1, types.ProtectionClassC, 3, 32, true),
			previousReader:          nil,
			expectedCurrentVersion:  3,
			expectedPreviousVersion: 0,
		},
		{
			name:                    "Both readers nil",
			currentReader:           nil,
			previousReader:          nil,
			expectedCurrentVersion:  0,
			expectedPreviousVersion: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager := NewKeyRollingManager(tc.currentReader, tc.previousReader)

			if manager.GetCurrentKeyVersion() != tc.expectedCurrentVersion {
				t.Errorf("GetCurrentKeyVersion() = %d, want %d", manager.GetCurrentKeyVersion(), tc.expectedCurrentVersion)
			}

			if manager.GetPreviousKeyVersion() != tc.expectedPreviousVersion {
				t.Errorf("GetPreviousKeyVersion() = %d, want %d", manager.GetPreviousKeyVersion(), tc.expectedPreviousVersion)
			}
		})
	}
}
