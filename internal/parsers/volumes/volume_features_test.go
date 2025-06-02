package volumes

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestSuperblockWithFlags creates a test superblock with specified flags
func createTestSuperblockWithFlags(
	features uint64,
	incompatFeatures uint64,
	fsFlags uint64,
) *types.ApfsSuperblockT {
	return &types.ApfsSuperblockT{
		ApfsFeatures:                   features,
		ApfsReadonlyCompatibleFeatures: 0,
		ApfsIncompatibleFeatures:       incompatFeatures,
		ApfsFsFlags:                    fsFlags,
	}
}

// TestVolumeFeatures tests all feature flag checks
func TestVolumeFeatures(t *testing.T) {
	testCases := []struct {
		name                           string
		features                       uint64
		incompatFeatures               uint64
		fsFlags                        uint64
		expectDefrag                   bool
		expectHardlinkMapRecords       bool
		expectStrictAccessTime         bool
		expectCaseInsensitive          bool
		expectNormalizationInsensitive bool
		expectSealed                   bool
		expectUnencrypted              bool
		expectOneKeyEncryption         bool
		expectSpilledOver              bool
		expectSpilloverCleaner         bool
		expectAlwaysCheckExtentRef     bool
	}{
		{
			name:                           "All Features Enabled",
			features:                       types.ApfsFeatureDefrag | types.ApfsFeatureHardlinkMapRecords | types.ApfsFeatureStrictatime,
			incompatFeatures:               types.ApfsIncompatCaseInsensitive | types.ApfsIncompatNormalizationInsensitive | types.ApfsIncompatSealedVolume,
			fsFlags:                        types.ApfsFsUnencrypted | types.ApfsFsOnekey | types.ApfsFsSpilledover | types.ApfsFsRunSpilloverCleaner | types.ApfsFsAlwaysCheckExtentref,
			expectDefrag:                   true,
			expectHardlinkMapRecords:       true,
			expectStrictAccessTime:         true,
			expectCaseInsensitive:          true,
			expectNormalizationInsensitive: true,
			expectSealed:                   true,
			expectUnencrypted:              true,
			expectOneKeyEncryption:         true,
			expectSpilledOver:              true,
			expectSpilloverCleaner:         true,
			expectAlwaysCheckExtentRef:     true,
		},
		{
			name:                           "No Features Enabled",
			features:                       0,
			incompatFeatures:               0,
			fsFlags:                        0,
			expectDefrag:                   false,
			expectHardlinkMapRecords:       false,
			expectStrictAccessTime:         false,
			expectCaseInsensitive:          false,
			expectNormalizationInsensitive: false,
			expectSealed:                   false,
			expectUnencrypted:              false,
			expectOneKeyEncryption:         false,
			expectSpilledOver:              false,
			expectSpilloverCleaner:         false,
			expectAlwaysCheckExtentRef:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sb := createTestSuperblockWithFlags(tc.features, tc.incompatFeatures, tc.fsFlags)
			vf := NewVolumeFeatures(sb)

			// Feature checks
			assertBoolEqual(t, "SupportsDefragmentation", vf.SupportsDefragmentation(), tc.expectDefrag)
			assertBoolEqual(t, "SupportsHardlinkMapRecords", vf.SupportsHardlinkMapRecords(), tc.expectHardlinkMapRecords)
			assertBoolEqual(t, "IsStrictAccessTimeEnabled", vf.IsStrictAccessTimeEnabled(), tc.expectStrictAccessTime)

			// Incompatibility checks
			assertBoolEqual(t, "IsCaseInsensitive", vf.IsCaseInsensitive(), tc.expectCaseInsensitive)
			assertBoolEqual(t, "IsNormalizationInsensitive", vf.IsNormalizationInsensitive(), tc.expectNormalizationInsensitive)
			assertBoolEqual(t, "IsSealed", vf.IsSealed(), tc.expectSealed)

			// Filesystem flag checks
			assertBoolEqual(t, "IsUnencrypted", vf.IsUnencrypted(), tc.expectUnencrypted)
			assertBoolEqual(t, "IsOneKeyEncryption", vf.IsOneKeyEncryption(), tc.expectOneKeyEncryption)
			assertBoolEqual(t, "IsSpilledOver", vf.IsSpilledOver(), tc.expectSpilledOver)
			assertBoolEqual(t, "RequiresSpilloverCleaner", vf.RequiresSpilloverCleaner(), tc.expectSpilloverCleaner)
			assertBoolEqual(t, "AlwaysChecksExtentReference", vf.AlwaysChecksExtentReference(), tc.expectAlwaysCheckExtentRef)
		})
	}
}

// Benchmark feature checks
func BenchmarkVolumeFeatures(b *testing.B) {
	sb := createTestSuperblockWithFlags(
		types.ApfsFeatureDefrag,
		types.ApfsIncompatCaseInsensitive,
		types.ApfsFsUnencrypted,
	)
	vf := NewVolumeFeatures(sb)

	// Benchmark individual feature checks
	benchmarkFeatureCheck(b, "SupportsDefragmentation", vf.SupportsDefragmentation)
	benchmarkFeatureCheck(b, "IsCaseInsensitive", vf.IsCaseInsensitive)
	benchmarkFeatureCheck(b, "IsUnencrypted", vf.IsUnencrypted)
}

// Helper function to assert boolean equality
func assertBoolEqual(t *testing.T, name string, actual, expected bool) {
	if actual != expected {
		t.Errorf("%s: expected %v, got %v", name, expected, actual)
	}
}

// Benchmark helper for feature checks
func benchmarkFeatureCheck(b *testing.B, name string, checkFunc func() bool) {
	b.Run(name, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = checkFunc()
		}
	})
}

// Test Raw Flag Accessors
func TestRawFlagAccessors(t *testing.T) {
	sb := createTestSuperblockWithFlags(
		0x1234,
		0x5678,
		0x9ABC,
	)
	vf := NewVolumeFeatures(sb)

	if got := vf.Features(); got != 0x1234 {
		t.Errorf("Features(): expected 0x1234, got 0x%x", got)
	}

	if got := vf.ReadonlyCompatibleFeatures(); got != 0 {
		t.Errorf("ReadonlyCompatibleFeatures(): expected 0, got 0x%x", got)
	}

	if got := vf.IncompatibleFeatures(); got != 0x5678 {
		t.Errorf("IncompatibleFeatures(): expected 0x5678, got 0x%x", got)
	}
}
