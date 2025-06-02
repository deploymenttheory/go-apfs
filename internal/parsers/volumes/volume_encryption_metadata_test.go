package volumes

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestSuperblockWithEncryption creates a test superblock with specific encryption details
func createTestSuperblockWithEncryption(
	fsFlags uint64,
	incompatFeatures uint64,
	metaCrypto types.WrappedMetaCryptoStateT,
) *types.ApfsSuperblockT {
	return &types.ApfsSuperblockT{
		ApfsFsFlags:              fsFlags,
		ApfsIncompatibleFeatures: incompatFeatures,
		ApfsMetaCrypto:           metaCrypto,
	}
}

// TestVolumeEncryptionMetadata tests all encryption metadata method implementations
func TestVolumeEncryptionMetadata(t *testing.T) {
	testCases := []struct {
		name                string
		fsFlags             uint64
		incompatFeatures    uint64
		metaCrypto          types.WrappedMetaCryptoStateT
		expectedIsEncrypted bool
		expectedKeyRotated  bool
	}{
		{
			name:             "Encrypted Volume",
			fsFlags:          0, // Not unencrypted
			incompatFeatures: types.ApfsIncompatEncRolled,
			metaCrypto: types.WrappedMetaCryptoStateT{
				MajorVersion:    5,
				MinorVersion:    0,
				PersistentClass: 1, // ProtectionClassA
			},
			expectedIsEncrypted: true,
			expectedKeyRotated:  true,
		},
		{
			name:                "Unencrypted Volume",
			fsFlags:             types.ApfsFsUnencrypted,
			incompatFeatures:    0,
			metaCrypto:          types.WrappedMetaCryptoStateT{},
			expectedIsEncrypted: false,
			expectedKeyRotated:  false,
		},
		{
			name:             "Encrypted Volume Without Key Rotation",
			fsFlags:          0,
			incompatFeatures: 0,
			metaCrypto: types.WrappedMetaCryptoStateT{
				MajorVersion:    5,
				MinorVersion:    0,
				PersistentClass: 2, // ProtectionClassB
			},
			expectedIsEncrypted: true,
			expectedKeyRotated:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sb := createTestSuperblockWithEncryption(
				tc.fsFlags,
				tc.incompatFeatures,
				tc.metaCrypto,
			)
			vem := NewVolumeEncryptionMetadata(sb)

			// Test MetadataCryptoState
			metaCrypto := vem.MetadataCryptoState()
			if metaCrypto.MajorVersion != tc.metaCrypto.MajorVersion {
				t.Errorf("MetadataCryptoState().MajorVersion = %d, want %d",
					metaCrypto.MajorVersion, tc.metaCrypto.MajorVersion)
			}

			// Test IsEncrypted
			if isEncrypted := vem.IsEncrypted(); isEncrypted != tc.expectedIsEncrypted {
				t.Errorf("IsEncrypted() = %v, want %v", isEncrypted, tc.expectedIsEncrypted)
			}

			// Test HasEncryptionKeyRotated
			if keyRotated := vem.HasEncryptionKeyRotated(); keyRotated != tc.expectedKeyRotated {
				t.Errorf("HasEncryptionKeyRotated() = %v, want %v", keyRotated, tc.expectedKeyRotated)
			}
		})
	}
}

// Benchmark encryption metadata methods
func BenchmarkVolumeEncryptionMetadata(b *testing.B) {
	sb := createTestSuperblockWithEncryption(
		0,
		types.ApfsIncompatEncRolled,
		types.WrappedMetaCryptoStateT{
			MajorVersion:    5,
			MinorVersion:    0,
			PersistentClass: 1,
		},
	)
	vem := NewVolumeEncryptionMetadata(sb)

	// Benchmark individual method calls
	b.Run("MetadataCryptoState", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = vem.MetadataCryptoState()
		}
	})

	b.Run("IsEncrypted", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = vem.IsEncrypted()
		}
	})

	b.Run("HasEncryptionKeyRotated", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = vem.HasEncryptionKeyRotated()
		}
	})
}

// TestMetadataCryptoStateCopy ensures the returned crypto state is a copy
func TestMetadataCryptoStateCopy(t *testing.T) {
	sb := createTestSuperblockWithEncryption(
		0,
		0,
		types.WrappedMetaCryptoStateT{
			MajorVersion:    5,
			MinorVersion:    0,
			PersistentClass: 1,
		},
	)
	vem := NewVolumeEncryptionMetadata(sb)

	// Get crypto state
	cryptoState1 := vem.MetadataCryptoState()
	cryptoState2 := vem.MetadataCryptoState()

	// Modify one instance
	cryptoState1.MajorVersion = 6

	// Verify the original instance is unchanged
	if cryptoState1.MajorVersion == cryptoState2.MajorVersion {
		t.Errorf("MetadataCryptoState did not return a copy")
	}
}
