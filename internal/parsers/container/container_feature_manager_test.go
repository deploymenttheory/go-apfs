package container

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestSuperblockForFeatures creates a test superblock with specific feature flags
func createTestSuperblockForFeatures(features, roCompatFeatures, incompatFeatures uint64) *types.NxSuperblockT {
	return &types.NxSuperblockT{
		NxFeatures:                   features,
		NxReadonlyCompatibleFeatures: roCompatFeatures,
		NxIncompatibleFeatures:       incompatFeatures,
	}
}

func TestContainerFeatureManager(t *testing.T) {
	tests := []struct {
		name               string
		features           uint64
		roCompatFeatures   uint64
		incompatFeatures   uint64
		expectedVersion    string
		expectDefrag       bool
		expectLowCapFusion bool
		expectFusion       bool
	}{
		{
			name:               "No features",
			features:           0,
			roCompatFeatures:   0,
			incompatFeatures:   0,
			expectedVersion:    "Unknown",
			expectDefrag:       false,
			expectLowCapFusion: false,
			expectFusion:       false,
		},
		{
			name:               "APFS Version 1",
			features:           0,
			roCompatFeatures:   0,
			incompatFeatures:   types.NxIncompatVersion1,
			expectedVersion:    "1.0",
			expectDefrag:       false,
			expectLowCapFusion: false,
			expectFusion:       false,
		},
		{
			name:               "APFS Version 2 with defrag",
			features:           types.NxFeatureDefrag,
			roCompatFeatures:   0,
			incompatFeatures:   types.NxIncompatVersion2,
			expectedVersion:    "2.0",
			expectDefrag:       true,
			expectLowCapFusion: false,
			expectFusion:       false,
		},
		{
			name:               "Fusion drive with low capacity",
			features:           types.NxFeatureLcfd,
			roCompatFeatures:   0,
			incompatFeatures:   types.NxIncompatFusion | types.NxIncompatVersion2,
			expectedVersion:    "2.0",
			expectDefrag:       false,
			expectLowCapFusion: true,
			expectFusion:       true,
		},
		{
			name:               "All features enabled",
			features:           types.NxFeatureDefrag | types.NxFeatureLcfd,
			roCompatFeatures:   0,
			incompatFeatures:   types.NxIncompatVersion2 | types.NxIncompatFusion,
			expectedVersion:    "2.0",
			expectDefrag:       true,
			expectLowCapFusion: true,
			expectFusion:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			superblock := createTestSuperblockForFeatures(tt.features, tt.roCompatFeatures, tt.incompatFeatures)
			manager := NewContainerFeatureManager(superblock)

			// Test basic feature getters
			if features := manager.Features(); features != tt.features {
				t.Errorf("Features() = 0x%X, want 0x%X", features, tt.features)
			}

			if roFeatures := manager.ReadOnlyCompatibleFeatures(); roFeatures != tt.roCompatFeatures {
				t.Errorf("ReadOnlyCompatibleFeatures() = 0x%X, want 0x%X", roFeatures, tt.roCompatFeatures)
			}

			if incompatFeatures := manager.IncompatibleFeatures(); incompatFeatures != tt.incompatFeatures {
				t.Errorf("IncompatibleFeatures() = 0x%X, want 0x%X", incompatFeatures, tt.incompatFeatures)
			}

			// Test feature detection methods
			if defrag := manager.SupportsDefragmentation(); defrag != tt.expectDefrag {
				t.Errorf("SupportsDefragmentation() = %v, want %v", defrag, tt.expectDefrag)
			}

			if lowCap := manager.IsLowCapacityFusionDrive(); lowCap != tt.expectLowCapFusion {
				t.Errorf("IsLowCapacityFusionDrive() = %v, want %v", lowCap, tt.expectLowCapFusion)
			}

			if fusion := manager.SupportsFusion(); fusion != tt.expectFusion {
				t.Errorf("SupportsFusion() = %v, want %v", fusion, tt.expectFusion)
			}

			// Test version detection
			if version := manager.GetAPFSVersion(); version != tt.expectedVersion {
				t.Errorf("GetAPFSVersion() = %s, want %s", version, tt.expectedVersion)
			}
		})
	}
}

func TestContainerFeatureManager_HasFeature(t *testing.T) {
	superblock := createTestSuperblockForFeatures(
		types.NxFeatureDefrag,
		0,
		types.NxIncompatVersion2|types.NxIncompatFusion,
	)
	manager := NewContainerFeatureManager(superblock)

	tests := []struct {
		name     string
		feature  uint64
		expected bool
	}{
		{
			name:     "Has defrag feature",
			feature:  types.NxFeatureDefrag,
			expected: true,
		},
		{
			name:     "Does not have LCFD feature",
			feature:  types.NxFeatureLcfd,
			expected: false,
		},
		{
			name:     "Custom feature flag",
			feature:  0x1000,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := manager.HasFeature(tt.feature); result != tt.expected {
				t.Errorf("HasFeature(0x%X) = %v, want %v", tt.feature, result, tt.expected)
			}
		})
	}
}

func TestContainerFeatureManager_HasIncompatibleFeature(t *testing.T) {
	superblock := createTestSuperblockForFeatures(
		0,
		0,
		types.NxIncompatVersion2|types.NxIncompatFusion,
	)
	manager := NewContainerFeatureManager(superblock)

	tests := []struct {
		name     string
		feature  uint64
		expected bool
	}{
		{
			name:     "Has version 2 feature",
			feature:  types.NxIncompatVersion2,
			expected: true,
		},
		{
			name:     "Has fusion feature",
			feature:  types.NxIncompatFusion,
			expected: true,
		},
		{
			name:     "Does not have version 1 feature",
			feature:  types.NxIncompatVersion1,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := manager.HasIncompatibleFeature(tt.feature); result != tt.expected {
				t.Errorf("HasIncompatibleFeature(0x%X) = %v, want %v", tt.feature, result, tt.expected)
			}
		})
	}
}

func TestContainerFeatureManager_HasReadOnlyCompatibleFeature(t *testing.T) {
	superblock := createTestSuperblockForFeatures(0, 0x123, 0)
	manager := NewContainerFeatureManager(superblock)

	tests := []struct {
		name     string
		feature  uint64
		expected bool
	}{
		{
			name:     "Has RO compatible feature",
			feature:  0x100,
			expected: true,
		},
		{
			name:     "Does not have RO compatible feature",
			feature:  0x200,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := manager.HasReadOnlyCompatibleFeature(tt.feature); result != tt.expected {
				t.Errorf("HasReadOnlyCompatibleFeature(0x%X) = %v, want %v", tt.feature, result, tt.expected)
			}
		})
	}
}
