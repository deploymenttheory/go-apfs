package container

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestSuperblockForEphemeral creates a test superblock with specific ephemeral info
func createTestSuperblockForEphemeral(ephInfo [types.NxEphInfoCount]uint64) *types.NxSuperblockT {
	return &types.NxSuperblockT{
		NxEphemeralInfo: ephInfo,
	}
}

func TestContainerEphemeralManager(t *testing.T) {
	// Create test ephemeral info array
	var testEphInfo [types.NxEphInfoCount]uint64
	for i := 0; i < types.NxEphInfoCount; i++ {
		testEphInfo[i] = uint64(1000 + i*100) // Test values: 1000, 1100, 1200, 1300
	}

	superblock := createTestSuperblockForEphemeral(testEphInfo)
	manager := NewContainerEphemeralManager(superblock)

	// Test EphemeralInfo() method
	t.Run("EphemeralInfo", func(t *testing.T) {
		ephInfo := manager.EphemeralInfo()

		if len(ephInfo) != types.NxEphInfoCount {
			t.Errorf("EphemeralInfo() length = %d, want %d", len(ephInfo), types.NxEphInfoCount)
		}

		for i, info := range ephInfo {
			expected := uint64(1000 + i*100)
			if info != expected {
				t.Errorf("EphemeralInfo()[%d] = %d, want %d", i, info, expected)
			}
		}
	})

	// Test MinimumBlockCount() method
	t.Run("MinimumBlockCount", func(t *testing.T) {
		minBlocks := manager.MinimumBlockCount()
		expected := uint32(types.NxEphMinBlockCount)

		if minBlocks != expected {
			t.Errorf("MinimumBlockCount() = %d, want %d", minBlocks, expected)
		}
	})

	// Test MaxEphemeralStructures() method
	t.Run("MaxEphemeralStructures", func(t *testing.T) {
		maxStructs := manager.MaxEphemeralStructures()
		expected := uint32(types.NxMaxFileSystemEphStructs)

		if maxStructs != expected {
			t.Errorf("MaxEphemeralStructures() = %d, want %d", maxStructs, expected)
		}
	})

	// Test EphemeralInfoVersion() method
	t.Run("EphemeralInfoVersion", func(t *testing.T) {
		version := manager.EphemeralInfoVersion()
		expected := uint32(types.NxEphInfoVersion1)

		if version != expected {
			t.Errorf("EphemeralInfoVersion() = %d, want %d", version, expected)
		}
	})
}

func TestContainerEphemeralManager_ZeroValues(t *testing.T) {
	var zeroEphInfo [types.NxEphInfoCount]uint64
	superblock := createTestSuperblockForEphemeral(zeroEphInfo)
	manager := NewContainerEphemeralManager(superblock)

	// Test with all zero ephemeral info
	ephInfo := manager.EphemeralInfo()
	for i, info := range ephInfo {
		if info != 0 {
			t.Errorf("Expected zero ephemeral info at index %d, got %d", i, info)
		}
	}

	// Constants should still return correct values
	if manager.MinimumBlockCount() != types.NxEphMinBlockCount {
		t.Errorf("MinimumBlockCount() = %d, want %d", manager.MinimumBlockCount(), types.NxEphMinBlockCount)
	}

	if manager.MaxEphemeralStructures() != types.NxMaxFileSystemEphStructs {
		t.Errorf("MaxEphemeralStructures() = %d, want %d", manager.MaxEphemeralStructures(), types.NxMaxFileSystemEphStructs)
	}

	if manager.EphemeralInfoVersion() != types.NxEphInfoVersion1 {
		t.Errorf("EphemeralInfoVersion() = %d, want %d", manager.EphemeralInfoVersion(), types.NxEphInfoVersion1)
	}
}

func TestContainerEphemeralManager_SliceIndependence(t *testing.T) {
	var testEphInfo [types.NxEphInfoCount]uint64
	for i := 0; i < types.NxEphInfoCount; i++ {
		testEphInfo[i] = uint64(i + 1)
	}

	superblock := createTestSuperblockForEphemeral(testEphInfo)
	manager := NewContainerEphemeralManager(superblock)

	// Get ephemeral info slice and modify it
	ephInfo := manager.EphemeralInfo()
	ephInfo[0] = 9999

	// Get ephemeral info again and verify original data is unchanged
	ephInfo2 := manager.EphemeralInfo()
	if ephInfo2[0] != 1 {
		t.Errorf("Modifying returned slice affected internal data: got %d, want 1", ephInfo2[0])
	}
}
