package objectmaps

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestObjectMapTransactionManagerImpl(t *testing.T) {
	header := types.OmapPhysT{
		OmPendingRevertMin: 100,
		OmPendingRevertMax: 200,
	}

	manager := NewObjectMapTransactionManager(header)

	if got := manager.PendingRevertMinXID(); got != header.OmPendingRevertMin {
		t.Errorf("PendingRevertMinXID() = %d, want %d", got, header.OmPendingRevertMin)
	}
	if got := manager.PendingRevertMaxXID(); got != header.OmPendingRevertMax {
		t.Errorf("PendingRevertMaxXID() = %d, want %d", got, header.OmPendingRevertMax)
	}
	if !manager.IsRevertInProgress() {
		t.Error("Expected IsRevertInProgress() = true")
	}

	// Test with no revert in progress
	headerEmpty := types.OmapPhysT{}
	managerEmpty := NewObjectMapTransactionManager(headerEmpty)

	if managerEmpty.IsRevertInProgress() {
		t.Error("Expected IsRevertInProgress() = false")
	}
}
