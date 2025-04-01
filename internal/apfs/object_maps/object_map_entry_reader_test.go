package objectmaps

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestObjectMapEntry(t *testing.T) {
	key := types.OmapKeyT{
		OkOid: 1234,
		OkXid: 5678,
	}
	val := types.OmapValT{
		OvFlags: types.OmapValEncrypted | types.OmapValDeleted,
		OvSize:  4096,
		OvPaddr: 0xABCDEF,
	}

	entry := NewObjectMapEntry(key, val)

	if entry.ObjectID() != key.OkOid {
		t.Errorf("ObjectID() = %d, want %d", entry.ObjectID(), key.OkOid)
	}
	if entry.TransactionID() != key.OkXid {
		t.Errorf("TransactionID() = %d, want %d", entry.TransactionID(), key.OkXid)
	}
	if entry.Flags() != val.OvFlags {
		t.Errorf("Flags() = %d, want %d", entry.Flags(), val.OvFlags)
	}
	if entry.Size() != val.OvSize {
		t.Errorf("Size() = %d, want %d", entry.Size(), val.OvSize)
	}
	if entry.PhysicalAddress() != val.OvPaddr {
		t.Errorf("PhysicalAddress() = %d, want %d", entry.PhysicalAddress(), val.OvPaddr)
	}
	if !entry.IsDeleted() {
		t.Error("Expected IsDeleted() = true")
	}
	if !entry.IsEncrypted() {
		t.Error("Expected IsEncrypted() = true")
	}
	if !entry.HasHeader() {
		t.Error("Expected HasHeader() = true")
	}

	// Now set flag for no-header
	val.OvFlags = types.OmapValNoheader
	entry = NewObjectMapEntry(key, val)
	if entry.HasHeader() {
		t.Error("Expected HasHeader() = false due to OmapValNoheader")
	}
}
