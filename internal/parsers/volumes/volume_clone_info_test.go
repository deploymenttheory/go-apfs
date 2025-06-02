package volumes

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestSuperblockWithCloneInfo creates a test superblock with specific clone info details
func createTestSuperblockWithCloneInfo(cloneInfoIdEpoch, cloneInfoXid uint64) *types.ApfsSuperblockT {
	return &types.ApfsSuperblockT{
		ApfsCloneinfoIdEpoch: cloneInfoIdEpoch,
		ApfsCloneinfoXid:     cloneInfoXid,
	}
}

// TestVolumeCloneInfo tests all clone info method implementations
func TestVolumeCloneInfo(t *testing.T) {
	testCases := []struct {
		name             string
		cloneInfoIdEpoch uint64
		cloneInfoXid     uint64
		expectedIdEpoch  uint64
		expectedXid      uint64
	}{
		{
			name:             "Valid Clone Info",
			cloneInfoIdEpoch: 12345,
			cloneInfoXid:     67890,
			expectedIdEpoch:  12345,
			expectedXid:      67890,
		},
		{
			name:             "Zero Clone Info",
			cloneInfoIdEpoch: 0,
			cloneInfoXid:     0,
			expectedIdEpoch:  0,
			expectedXid:      0,
		},
		{
			name:             "Maximum Values",
			cloneInfoIdEpoch: ^uint64(0), // Maximum uint64
			cloneInfoXid:     ^uint64(0), // Maximum uint64
			expectedIdEpoch:  ^uint64(0),
			expectedXid:      ^uint64(0),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sb := createTestSuperblockWithCloneInfo(tc.cloneInfoIdEpoch, tc.cloneInfoXid)
			vci := NewVolumeCloneInfo(sb)

			// Test CloneInfoIdEpoch
			if idEpoch := vci.CloneInfoIdEpoch(); idEpoch != tc.expectedIdEpoch {
				t.Errorf("CloneInfoIdEpoch() = %d, want %d", idEpoch, tc.expectedIdEpoch)
			}

			// Test CloneInfoXID
			if xid := vci.CloneInfoXID(); xid != tc.expectedXid {
				t.Errorf("CloneInfoXID() = %d, want %d", xid, tc.expectedXid)
			}
		})
	}
}

// TestVolumeCloneInfo_NewConstructor tests the constructor with nil superblock
func TestVolumeCloneInfo_NewConstructor(t *testing.T) {
	sb := createTestSuperblockWithCloneInfo(42, 84)
	vci := NewVolumeCloneInfo(sb)

	if vci == nil {
		t.Error("NewVolumeCloneInfo() returned nil")
	}
}

// Benchmark clone info methods
func BenchmarkVolumeCloneInfo(b *testing.B) {
	sb := createTestSuperblockWithCloneInfo(12345, 67890)
	vci := NewVolumeCloneInfo(sb)

	b.Run("CloneInfoIdEpoch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = vci.CloneInfoIdEpoch()
		}
	})

	b.Run("CloneInfoXID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = vci.CloneInfoXID()
		}
	})
}
