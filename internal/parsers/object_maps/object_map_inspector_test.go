package objectmaps

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// makeTestEntry creates a test ObjectMapEntryReader with given OID, XID, and flags
func makeTestEntry(oid types.OidT, xid types.XidT, flags uint32) interfaces.ObjectMapEntryReader {
	return NewObjectMapEntry(
		types.OmapKeyT{OkOid: oid, OkXid: xid},
		types.OmapValT{OvFlags: flags},
	)
}

func TestObjectMapInspectorImpl(t *testing.T) {
	entries := []interfaces.ObjectMapEntryReader{
		makeTestEntry(1001, 1, types.OmapValEncrypted),
		makeTestEntry(2002, 2, types.OmapValDeleted),
		makeTestEntry(3003, 3, 0),
	}

	inspector := NewObjectMapInspector(entries)

	t.Run("ListObjects returns all entries", func(t *testing.T) {
		list, err := inspector.ListObjects()
		if err != nil {
			t.Fatalf("ListObjects() error: %v", err)
		}
		if len(list) != 3 {
			t.Errorf("Expected 3 entries, got %d", len(list))
		}
	})

	t.Run("FindObjectByID finds by OID only", func(t *testing.T) {
		entry, err := inspector.FindObjectByID(2002)
		if err != nil {
			t.Fatalf("FindObjectByID(2002) error: %v", err)
		}
		if entry.ObjectID() != 2002 {
			t.Errorf("Expected ObjectID 2002, got %d", entry.ObjectID())
		}
	})

	t.Run("FindObjectByID finds by OID and XID", func(t *testing.T) {
		entry, err := inspector.FindObjectByID(1001, 1)
		if err != nil {
			t.Fatalf("FindObjectByID(1001, 1) error: %v", err)
		}
		if entry.TransactionID() != 1 {
			t.Errorf("Expected TransactionID 1, got %d", entry.TransactionID())
		}
	})

	t.Run("FindObjectByID returns error when not found", func(t *testing.T) {
		_, err := inspector.FindObjectByID(9999)
		if err == nil {
			t.Error("Expected error for missing object, got nil")
		}
	})

	t.Run("CountObjects returns total count", func(t *testing.T) {
		count, err := inspector.CountObjects()
		if err != nil {
			t.Fatalf("CountObjects() error: %v", err)
		}
		if count != 3 {
			t.Errorf("Expected count 3, got %d", count)
		}
	})

	t.Run("FindDeletedObjects returns only deleted", func(t *testing.T) {
		deleted, err := inspector.FindDeletedObjects()
		if err != nil {
			t.Fatalf("FindDeletedObjects() error: %v", err)
		}
		if len(deleted) != 1 || !deleted[0].IsDeleted() {
			t.Errorf("Expected 1 deleted entry, got %d", len(deleted))
		}
	})

	t.Run("FindEncryptedObjects returns only encrypted", func(t *testing.T) {
		encrypted, err := inspector.FindEncryptedObjects()
		if err != nil {
			t.Fatalf("FindEncryptedObjects() error: %v", err)
		}
		if len(encrypted) != 1 || !encrypted[0].IsEncrypted() {
			t.Errorf("Expected 1 encrypted entry, got %d", len(encrypted))
		}
	})
}
