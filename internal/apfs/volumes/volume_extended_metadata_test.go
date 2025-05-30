package volumes

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestSuperblockWithExtendedMetadata creates a test superblock with specific extended metadata details
func createTestSuperblockWithExtendedMetadata(snapMetaExtOid, fextTreeOid types.OidT, fextTreeType uint32) *types.ApfsSuperblockT {
	return &types.ApfsSuperblockT{
		ApfsSnapMetaExtOid: snapMetaExtOid,
		ApfsFextTreeOid:    fextTreeOid,
		ApfsFextTreeType:   fextTreeType,
	}
}

// TestVolumeExtendedMetadata tests all extended metadata method implementations
func TestVolumeExtendedMetadata(t *testing.T) {
	testCases := []struct {
		name                 string
		snapMetaExtOid       types.OidT
		fextTreeOid          types.OidT
		fextTreeType         uint32
		expectedSnapMetaOid  types.OidT
		expectedFextTreeOid  types.OidT
		expectedFextTreeType uint32
	}{
		{
			name:                 "Valid Extended Metadata",
			snapMetaExtOid:       12345,
			fextTreeOid:          67890,
			fextTreeType:         1,
			expectedSnapMetaOid:  12345,
			expectedFextTreeOid:  67890,
			expectedFextTreeType: 1,
		},
		{
			name:                 "Zero Values",
			snapMetaExtOid:       0,
			fextTreeOid:          0,
			fextTreeType:         0,
			expectedSnapMetaOid:  0,
			expectedFextTreeOid:  0,
			expectedFextTreeType: 0,
		},
		{
			name:                 "Maximum Values",
			snapMetaExtOid:       ^types.OidT(0), // Maximum OidT
			fextTreeOid:          ^types.OidT(0), // Maximum OidT
			fextTreeType:         ^uint32(0),     // Maximum uint32
			expectedSnapMetaOid:  ^types.OidT(0),
			expectedFextTreeOid:  ^types.OidT(0),
			expectedFextTreeType: ^uint32(0),
		},
		{
			name:                 "Mixed Values",
			snapMetaExtOid:       types.OidInvalid,
			fextTreeOid:          54321,
			fextTreeType:         2,
			expectedSnapMetaOid:  types.OidInvalid,
			expectedFextTreeOid:  54321,
			expectedFextTreeType: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sb := createTestSuperblockWithExtendedMetadata(tc.snapMetaExtOid, tc.fextTreeOid, tc.fextTreeType)
			vem := NewVolumeExtendedMetadata(sb)

			// Test SnapshotMetadataExtOID
			if snapMetaOid := vem.SnapshotMetadataExtOID(); snapMetaOid != tc.expectedSnapMetaOid {
				t.Errorf("SnapshotMetadataExtOID() = %d, want %d", snapMetaOid, tc.expectedSnapMetaOid)
			}

			// Test FileExtentTreeOID
			if fextTreeOid := vem.FileExtentTreeOID(); fextTreeOid != tc.expectedFextTreeOid {
				t.Errorf("FileExtentTreeOID() = %d, want %d", fextTreeOid, tc.expectedFextTreeOid)
			}

			// Test FileExtentTreeType
			if fextTreeType := vem.FileExtentTreeType(); fextTreeType != tc.expectedFextTreeType {
				t.Errorf("FileExtentTreeType() = %d, want %d", fextTreeType, tc.expectedFextTreeType)
			}
		})
	}
}

// TestVolumeExtendedMetadata_NewConstructor tests the constructor
func TestVolumeExtendedMetadata_NewConstructor(t *testing.T) {
	sb := createTestSuperblockWithExtendedMetadata(42, 84, 2)
	vem := NewVolumeExtendedMetadata(sb)

	if vem == nil {
		t.Error("NewVolumeExtendedMetadata() returned nil")
	}
}

// TestVolumeExtendedMetadata_IndividualMethods tests each method independently
func TestVolumeExtendedMetadata_IndividualMethods(t *testing.T) {
	sb := createTestSuperblockWithExtendedMetadata(111, 222, 333)
	vem := NewVolumeExtendedMetadata(sb)

	t.Run("SnapshotMetadataExtOID", func(t *testing.T) {
		if got := vem.SnapshotMetadataExtOID(); got != 111 {
			t.Errorf("SnapshotMetadataExtOID() = %d, want %d", got, 111)
		}
	})

	t.Run("FileExtentTreeOID", func(t *testing.T) {
		if got := vem.FileExtentTreeOID(); got != 222 {
			t.Errorf("FileExtentTreeOID() = %d, want %d", got, 222)
		}
	})

	t.Run("FileExtentTreeType", func(t *testing.T) {
		if got := vem.FileExtentTreeType(); got != 333 {
			t.Errorf("FileExtentTreeType() = %d, want %d", got, 333)
		}
	})
}

// Benchmark extended metadata methods
func BenchmarkVolumeExtendedMetadata(b *testing.B) {
	sb := createTestSuperblockWithExtendedMetadata(12345, 67890, 1)
	vem := NewVolumeExtendedMetadata(sb)

	b.Run("SnapshotMetadataExtOID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = vem.SnapshotMetadataExtOID()
		}
	})

	b.Run("FileExtentTreeOID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = vem.FileExtentTreeOID()
		}
	})

	b.Run("FileExtentTreeType", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = vem.FileExtentTreeType()
		}
	})
}
