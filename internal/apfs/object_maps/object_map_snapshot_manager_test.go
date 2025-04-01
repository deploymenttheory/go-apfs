package objectmaps

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestObjectMapSnapshotManagerImpl(t *testing.T) {
	snapshots := []types.OmapSnapshotT{
		{OmsOid: 100, OmsFlags: types.OmapSnapshotDeleted},
		{OmsOid: 200, OmsFlags: types.OmapSnapshotReverted},
		{OmsOid: 300, OmsFlags: 0},
	}

	manager := NewObjectMapSnapshotManager(snapshots)

	t.Run("ListSnapshots returns all snapshots", func(t *testing.T) {
		list, err := manager.ListSnapshots()
		if err != nil {
			t.Fatalf("ListSnapshots() failed: %v", err)
		}
		if len(list) != len(snapshots) {
			t.Fatalf("Expected %d snapshots, got %d", len(snapshots), len(list))
		}
	})

	t.Run("FindSnapshotByOID returns correct snapshot", func(t *testing.T) {
		snap, err := manager.FindSnapshotByOID(200)
		if err != nil {
			t.Fatalf("FindSnapshotByOID(200) failed: %v", err)
		}
		if snap.Oid() != 200 {
			t.Errorf("Expected OID 200, got %d", snap.Oid())
		}
		if !snap.IsReverted() {
			t.Errorf("Expected IsReverted() = true")
		}
	})

	t.Run("FindSnapshotByOID returns error for missing snapshot", func(t *testing.T) {
		_, err := manager.FindSnapshotByOID(9999)
		if err == nil {
			t.Fatal("Expected error for missing snapshot, got nil")
		}
	})
}
