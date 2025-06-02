package volumes

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestSuperblockWithTreeStructure creates a test superblock with specific tree structure details
func createTestSuperblockWithTreeStructure(
	rootTreeOID types.OidT,
	rootTreeType uint32,
	extentRefTreeOID types.OidT,
	extentRefTreeType uint32,
	snapMetaTreeOID types.OidT,
	snapMetaTreeType uint32,
	omapOID types.OidT,
) *types.ApfsSuperblockT {
	return &types.ApfsSuperblockT{
		ApfsRootTreeOid:       rootTreeOID,
		ApfsRootTreeType:      rootTreeType,
		ApfsExtentrefTreeOid:  extentRefTreeOID,
		ApfsExtentreftreeType: extentRefTreeType,
		ApfsSnapMetaTreeOid:   snapMetaTreeOID,
		ApfsSnapMetatreeType:  snapMetaTreeType,
		ApfsOmapOid:           omapOID,
	}
}

// TestVolumeTreeStructure tests all tree structure method implementations
func TestVolumeTreeStructure(t *testing.T) {
	testCases := []struct {
		name              string
		rootTreeOID       types.OidT
		rootTreeType      uint32
		extentRefTreeOID  types.OidT
		extentRefTreeType uint32
		snapMetaTreeOID   types.OidT
		snapMetaTreeType  uint32
		omapOID           types.OidT
	}{
		{
			name:              "Typical Volume Tree Structure",
			rootTreeOID:       1234,
			rootTreeType:      0x0000000E, // OBJECT_TYPE_FSTREE
			extentRefTreeOID:  5678,
			extentRefTreeType: 0x0000000F, // OBJECT_TYPE_BLOCKREF
			snapMetaTreeOID:   9012,
			snapMetaTreeType:  0x00000010, // OBJECT_TYPE_BLOCKREF
			omapOID:           3456,
		},
		{
			name:              "Zero OIDs",
			rootTreeOID:       0,
			rootTreeType:      0,
			extentRefTreeOID:  0,
			extentRefTreeType: 0,
			snapMetaTreeOID:   0,
			snapMetaTreeType:  0,
			omapOID:           0,
		},
		{
			name:              "Large OIDs",
			rootTreeOID:       0xFFFFFFFFFFFFFFFF,
			rootTreeType:      0xFFFFFFFF,
			extentRefTreeOID:  0xEEEEEEEEEEEEEEEE,
			extentRefTreeType: 0xEEEEEEEE,
			snapMetaTreeOID:   0xDDDDDDDDDDDDDDDD,
			snapMetaTreeType:  0xDDDDDDDD,
			omapOID:           0xCCCCCCCCCCCCCCCC,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sb := createTestSuperblockWithTreeStructure(
				tc.rootTreeOID,
				tc.rootTreeType,
				tc.extentRefTreeOID,
				tc.extentRefTreeType,
				tc.snapMetaTreeOID,
				tc.snapMetaTreeType,
				tc.omapOID,
			)
			vts := NewVolumeTreeStructure(sb)

			// Test each method
			assertOidEqual(t, "RootTreeOID", vts.RootTreeOID(), tc.rootTreeOID)
			assertUint32Equal(t, "RootTreeType", vts.RootTreeType(), tc.rootTreeType)

			assertOidEqual(t, "ExtentReferenceTreeOID", vts.ExtentReferenceTreeOID(), tc.extentRefTreeOID)
			assertUint32Equal(t, "ExtentReferenceTreeType", vts.ExtentReferenceTreeType(), tc.extentRefTreeType)

			assertOidEqual(t, "SnapshotMetadataTreeOID", vts.SnapshotMetadataTreeOID(), tc.snapMetaTreeOID)
			assertUint32Equal(t, "SnapshotMetadataTreeType", vts.SnapshotMetadataTreeType(), tc.snapMetaTreeType)

			assertOidEqual(t, "ObjectMapOID", vts.ObjectMapOID(), tc.omapOID)
		})
	}
}

// Benchmark tree structure methods
func BenchmarkVolumeTreeStructure(b *testing.B) {
	sb := createTestSuperblockWithTreeStructure(
		1234, 0x0000000E,
		5678, 0x0000000F,
		9012, 0x00000010,
		3456,
	)
	vts := NewVolumeTreeStructure(sb)

	// Benchmark individual method calls
	benchmarkTreeStructureMethod(b, "RootTreeOID", vts.RootTreeOID)
	benchmarkTreeStructureMethod(b, "ExtentReferenceTreeOID", vts.ExtentReferenceTreeOID)
	benchmarkTreeStructureMethod(b, "SnapshotMetadataTreeOID", vts.SnapshotMetadataTreeOID)
	benchmarkTreeStructureType(b, "RootTreeType", vts.RootTreeType)
}

// Helper function to assert OidT equality
func assertOidEqual(t *testing.T, name string, actual, expected types.OidT) {
	if actual != expected {
		t.Errorf("%s: expected %d, got %d", name, expected, actual)
	}
}

// Helper function to assert uint32 equality
func assertUint32Equal(t *testing.T, name string, actual, expected uint32) {
	if actual != expected {
		t.Errorf("%s: expected %d, got %d", name, expected, actual)
	}
}

// Benchmark helper for OidT method
func benchmarkTreeStructureMethod(b *testing.B, name string, method func() types.OidT) {
	b.Run(name, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = method()
		}
	})
}

// Benchmark helper for uint32 method
func benchmarkTreeStructureType(b *testing.B, name string, method func() uint32) {
	b.Run(name, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = method()
		}
	})
}
