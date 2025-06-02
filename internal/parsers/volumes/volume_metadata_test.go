package volumes

import (
	"testing"
	"time"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestSuperblockWithMetadata creates a test superblock with specific metadata
func createTestSuperblockWithMetadata(
	unmountTime uint64,
	lastModTime uint64,
	formattedBy types.ApfsModifiedByT,
	modifiedBy [types.ApfsMaxHist]types.ApfsModifiedByT,
	nextObjId uint64,
	nextDocId uint32,
) *types.ApfsSuperblockT {
	return &types.ApfsSuperblockT{
		ApfsUnmountTime: unmountTime,
		ApfsLastModTime: lastModTime,
		ApfsFormattedBy: formattedBy,
		ApfsModifiedBy:  modifiedBy,
		ApfsNextObjId:   nextObjId,
		ApfsNextDocId:   nextDocId,
	}
}

// TestVolumeMetadata tests all metadata method implementations
func TestVolumeMetadata(t *testing.T) {
	// Prepare test modification history
	modifiedBy := [types.ApfsMaxHist]types.ApfsModifiedByT{}
	for i := 0; i < len(modifiedBy); i++ {
		modifiedBy[i] = types.ApfsModifiedByT{
			Id:        [32]byte{'T', 'e', 's', 't', byte('0' + i)},
			Timestamp: uint64(1000000000 * (i + 1)), // Incrementing timestamps
			LastXid:   types.XidT(i + 100),
		}
	}

	// Test case with various metadata
	sb := createTestSuperblockWithMetadata(
		1600000000000000000, // Unmount time (nanoseconds since epoch)
		1700000000000000000, // Last modified time
		types.ApfsModifiedByT{
			Id:        [32]byte{'I', 'n', 'i', 't', 'i', 'a', 'l'},
			Timestamp: 1500000000000000000,
			LastXid:   50,
		},
		modifiedBy,
		12345, // Next object ID
		67,    // Next document ID
	)

	vm := NewVolumeMetadata(sb)

	// Test LastUnmountTime
	expectedUnmountTime := time.Unix(0, int64(1600000000000000000))
	if unmountTime := vm.LastUnmountTime(); unmountTime != expectedUnmountTime {
		t.Errorf("LastUnmountTime() = %v, want %v", unmountTime, expectedUnmountTime)
	}

	// Test LastModifiedTime
	expectedModTime := time.Unix(0, int64(1700000000000000000))
	if modTime := vm.LastModifiedTime(); modTime != expectedModTime {
		t.Errorf("LastModifiedTime() = %v, want %v", modTime, expectedModTime)
	}

	// Test FormattedBy
	formattedBy := vm.FormattedBy()
	if string(formattedBy.Id[:7]) != "Initial" {
		t.Errorf("FormattedBy().Id = %s, want 'Initial'", string(formattedBy.Id[:7]))
	}

	// Test ModificationHistory
	history := vm.ModificationHistory()
	if len(history) != types.ApfsMaxHist {
		t.Errorf("ModificationHistory() length = %d, want %d", len(history), types.ApfsMaxHist)
	}

	// Verify first modification entry
	if string(history[0].Id[:5]) != "Test0" {
		t.Errorf("First modification entry Id = %s, want 'Test0'", string(history[0].Id[:5]))
	}

	// Test NextObjectID
	if nextObjId := vm.NextObjectID(); nextObjId != 12345 {
		t.Errorf("NextObjectID() = %d, want 12345", nextObjId)
	}

	// Test NextDocumentID
	if nextDocId := vm.NextDocumentID(); nextDocId != 67 {
		t.Errorf("NextDocumentID() = %d, want 67", nextDocId)
	}
}

// Benchmark metadata methods
func BenchmarkVolumeMetadata(b *testing.B) {
	sb := createTestSuperblockWithMetadata(
		1600000000000000000,
		1700000000000000000,
		types.ApfsModifiedByT{},
		[types.ApfsMaxHist]types.ApfsModifiedByT{},
		12345,
		67,
	)
	vm := NewVolumeMetadata(sb)

	// Benchmark individual method calls
	b.Run("LastUnmountTime", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = vm.LastUnmountTime()
		}
	})

	b.Run("LastModifiedTime", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = vm.LastModifiedTime()
		}
	})

	b.Run("NextObjectID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = vm.NextObjectID()
		}
	})
}

// TestModificationHistoryCopy ensures the returned slice is a copy
func TestModificationHistoryCopy(t *testing.T) {
	sb := createTestSuperblockWithMetadata(
		1600000000000000000,
		1700000000000000000,
		types.ApfsModifiedByT{},
		[types.ApfsMaxHist]types.ApfsModifiedByT{},
		12345,
		67,
	)
	vm := NewVolumeMetadata(sb)

	// Get modification history
	history1 := vm.ModificationHistory()
	history2 := vm.ModificationHistory()

	// Modify one slice
	history1[0].Id[0] = 'X'

	// Verify the original slice is unchanged
	if history1[0].Id[0] == history2[0].Id[0] {
		t.Errorf("ModificationHistory did not return a copy of the slice")
	}
}
