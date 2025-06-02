package encryption

import (
	"encoding/binary"
	"testing"

	helpers "github.com/deploymenttheory/go-apfs/internal/helpers"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestCryptoData creates test crypto key and value data
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

func TestNewCryptoStateReader(t *testing.T) {
	endian := binary.LittleEndian

	tests := []struct {
		name            string
		objID           types.OidT
		objType         uint32
		refCount        uint32
		protectionClass types.CpKeyClassT
		keyVersion      types.CpKeyRevisionT
		keyLen          uint16
		expectError     bool
	}{
		{
			name:            "Valid crypto state",
			objID:           12345,
			objType:         types.ObjectTypeTest,
			refCount:        1,
			protectionClass: types.ProtectionClassC,
			keyVersion:      1,
			keyLen:          32,
			expectError:     false,
		},
		{
			name:            "Protection class A",
			objID:           67890,
			objType:         types.ObjectTypeTest,
			refCount:        2,
			protectionClass: types.ProtectionClassA,
			keyVersion:      2,
			keyLen:          64,
			expectError:     false,
		},
		{
			name:            "Maximum key length",
			objID:           11111,
			objType:         types.ObjectTypeTest,
			refCount:        1,
			protectionClass: types.ProtectionClassD,
			keyVersion:      1,
			keyLen:          types.CpMaxWrappedkeysize,
			expectError:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keyData, valueData := createTestCryptoData(tc.objID, tc.objType, tc.refCount, tc.protectionClass, tc.keyVersion, tc.keyLen, endian)

			csr, err := NewCryptoStateReader(keyData, valueData, endian)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Test basic properties
			if csr.ReferenceCount() != tc.refCount {
				t.Errorf("ReferenceCount() = %d, want %d", csr.ReferenceCount(), tc.refCount)
			}

			if csr.ProtectionClass() != tc.protectionClass {
				t.Errorf("ProtectionClass() = %d, want %d", csr.ProtectionClass(), tc.protectionClass)
			}

			if csr.KeyVersion() != tc.keyVersion {
				t.Errorf("KeyVersion() = %d, want %d", csr.KeyVersion(), tc.keyVersion)
			}

			if csr.KeyLength() != tc.keyLen {
				t.Errorf("KeyLength() = %d, want %d", csr.KeyLength(), tc.keyLen)
			}

			if csr.ObjectID() != tc.objID {
				t.Errorf("ObjectID() = %d, want %d", csr.ObjectID(), tc.objID)
			}

			if csr.ObjectType() != (tc.objType & 0xF) {
				t.Errorf("ObjectType() = %d, want %d (4-bit masked)", csr.ObjectType(), tc.objType&0xF)
			}

			// Test version information
			if csr.MajorVersion() != 5 {
				t.Errorf("MajorVersion() = %d, want 5", csr.MajorVersion())
			}

			if csr.MinorVersion() != 0 {
				t.Errorf("MinorVersion() = %d, want 0", csr.MinorVersion())
			}

			// Test wrapped key data
			wrappedKey := csr.WrappedKeyData()
			if len(wrappedKey) != int(tc.keyLen) {
				t.Errorf("WrappedKeyData() length = %d, want %d", len(wrappedKey), tc.keyLen)
			}

			// Verify key data pattern
			for i := 0; i < len(wrappedKey); i++ {
				expected := byte(i % 256)
				if wrappedKey[i] != expected {
					t.Errorf("WrappedKeyData()[%d] = %d, want %d", i, wrappedKey[i], expected)
				}
			}
		})
	}
}

func TestCryptoStateReaderErrors(t *testing.T) {
	endian := binary.LittleEndian

	tests := []struct {
		name          string
		keyDataSize   int
		valueDataSize int
	}{
		{"Empty key data", 0, 20},
		{"Too small key data", 7, 20},
		{"Empty value data", 8, 0},
		{"Too small value data", 8, 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keyData := make([]byte, tc.keyDataSize)
			valueData := make([]byte, tc.valueDataSize)

			_, err := NewCryptoStateReader(keyData, valueData, endian)
			if err == nil {
				t.Errorf("Expected error but got none")
			}
		})
	}
}

func TestCryptoStateReaderIsValid(t *testing.T) {
	endian := binary.LittleEndian

	tests := []struct {
		name            string
		refCount        uint32
		protectionClass types.CpKeyClassT
		expectValid     bool
	}{
		{
			name:            "Valid state with Class C",
			refCount:        1,
			protectionClass: types.ProtectionClassC,
			expectValid:     true,
		},
		{
			name:            "Valid state with Class A",
			refCount:        2,
			protectionClass: types.ProtectionClassA,
			expectValid:     true,
		},
		{
			name:            "Invalid - zero reference count",
			refCount:        0,
			protectionClass: types.ProtectionClassC,
			expectValid:     false,
		},
		{
			name:            "Invalid - unknown protection class",
			refCount:        1,
			protectionClass: types.CpKeyClassT(999),
			expectValid:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keyData, valueData := createTestCryptoData(12345, types.ObjectTypeTest, tc.refCount, tc.protectionClass, 1, 32, endian)

			csr, err := NewCryptoStateReader(keyData, valueData, endian)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if csr.IsValid() != tc.expectValid {
				t.Errorf("IsValid() = %t, want %t", csr.IsValid(), tc.expectValid)
			}
		})
	}
}

func TestCryptoStateReaderOSVersion(t *testing.T) {
	endian := binary.LittleEndian

	// Test OS version packing/unpacking
	majorVersion := uint16(14)
	minorLetter := byte('A')
	buildNumber := uint32(12345)

	packedVersion := helpers.PackOsVersion(majorVersion, minorLetter, buildNumber)

	keyData, valueData := createTestCryptoData(12345, types.ObjectTypeTest, 1, types.ProtectionClassC, 1, 32, endian)

	// Manually set the OS version in the value data
	endian.PutUint32(valueData[16:20], uint32(packedVersion))

	csr, err := NewCryptoStateReader(keyData, valueData, endian)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	osVersion := csr.OSVersion()
	unpackedMajor, unpackedMinor, unpackedBuild := helpers.UnpackOsVersion(osVersion)

	if unpackedMajor != majorVersion {
		t.Errorf("Unpacked major version = %d, want %d", unpackedMajor, majorVersion)
	}

	if unpackedMinor != minorLetter {
		t.Errorf("Unpacked minor letter = %c, want %c", unpackedMinor, minorLetter)
	}

	if unpackedBuild != buildNumber {
		t.Errorf("Unpacked build number = %d, want %d", unpackedBuild, buildNumber)
	}
}

func TestCryptoStateReaderEndianness(t *testing.T) {
	tests := []struct {
		name   string
		endian binary.ByteOrder
	}{
		{"Little Endian", binary.LittleEndian},
		{"Big Endian", binary.BigEndian},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			objID := types.OidT(0x023456789ABCDEF0) // Use proper 60-bit value (high nibble = 0)
			objType := uint32(types.ObjectTypeTest)
			refCount := uint32(0x12345678)

			keyData, valueData := createTestCryptoData(objID, objType, refCount, types.ProtectionClassC, 1, 32, tc.endian)

			csr, err := NewCryptoStateReader(keyData, valueData, tc.endian)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if csr.ObjectID() != objID {
				t.Errorf("ObjectID() = 0x%X, want 0x%X", csr.ObjectID(), objID)
			}

			if csr.ReferenceCount() != refCount {
				t.Errorf("ReferenceCount() = 0x%X, want 0x%X", csr.ReferenceCount(), refCount)
			}
		})
	}
}

func BenchmarkNewCryptoStateReader(b *testing.B) {
	endian := binary.LittleEndian
	keyData, valueData := createTestCryptoData(12345, types.ObjectTypeTest, 1, types.ProtectionClassC, 1, 64, endian)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewCryptoStateReader(keyData, valueData, endian)
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkCryptoStateReaderOperations(b *testing.B) {
	endian := binary.LittleEndian
	keyData, valueData := createTestCryptoData(12345, types.ObjectTypeTest, 1, types.ProtectionClassC, 1, 64, endian)

	csr, err := NewCryptoStateReader(keyData, valueData, endian)
	if err != nil {
		b.Errorf("Unexpected error: %v", err)
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = csr.ReferenceCount()
		_ = csr.ProtectionClass()
		_ = csr.KeyVersion()
		_ = csr.IsValid()
		_ = csr.WrappedKeyData()
	}
}
