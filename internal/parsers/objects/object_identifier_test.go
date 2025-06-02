package objects

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestNewObjectIdentifier(t *testing.T) {
	id := types.OidT(42)
	xid := types.XidT(100)

	objID := NewObjectIdentifier(id, xid)
	if objID == nil {
		t.Fatal("Expected non-nil BaseObjectIdentifier")
	}

	if objID.ID() != id {
		t.Errorf("Expected ID %d, got %d", id, objID.ID())
	}

	if objID.TransactionID() != xid {
		t.Errorf("Expected TransactionID %d, got %d", xid, objID.TransactionID())
	}
}

func TestObjectIdentifier_IsValid_Valid(t *testing.T) {
	id := types.OidT(1234)
	xid := types.XidT(5678)

	objID := NewObjectIdentifier(id, xid)
	if !objID.IsValid() {
		t.Errorf("Expected IsValid() to be true, got false")
	}
}

func TestObjectIdentifier_IsValid_Invalid(t *testing.T) {
	objID := NewObjectIdentifier(types.OidInvalid, types.XidInvalid)
	if objID.IsValid() {
		t.Errorf("Expected IsValid() to be false, got true")
	}
}
