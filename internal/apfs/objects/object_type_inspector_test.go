package objects

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func makeObj(oType, oSubtype uint32) *types.ObjPhysT {
	return &types.ObjPhysT{
		OType:    oType,
		OSubtype: oSubtype,
	}
}

func TestTypeAndSubtype(t *testing.T) {
	obj := makeObj(types.ObjectTypeFs|types.ObjPhysical, 42)
	inspector := NewObjectInspector(obj)

	if got := inspector.Type(); got != types.ObjectTypeFs {
		t.Errorf("Expected Type() %x, got %x", types.ObjectTypeFs, got)
	}

	if got := inspector.Subtype(); got != 42 {
		t.Errorf("Expected Subtype() 42, got %d", got)
	}
}

func TestTypeName(t *testing.T) {
	cases := []struct {
		oType    uint32
		expected string
	}{
		{types.ObjectTypeNxSuperblock, "NX Superblock"},
		{types.ObjectTypeBtree, "B-tree Root"},
		{types.ObjectTypeBtreeNode, "B-tree Node"},
		{types.ObjectTypeFs, "APFS Volume"},
		{types.ObjectTypeOmap, "Object Map"},
		{types.ObjectTypeInvalid, "Invalid"},
		{0x9999, "Unknown"},
	}

	for _, c := range cases {
		obj := makeObj(c.oType, 0)
		inspector := NewObjectInspector(obj)
		if got := inspector.TypeName(); got != c.expected {
			t.Errorf("TypeName() = %q, want %q for type 0x%x", got, c.expected, c.oType)
		}
	}
}

func TestStorageTypeChecks(t *testing.T) {
	tests := []struct {
		name      string
		oType     uint32
		virtual   bool
		ephemeral bool
		physical  bool
	}{
		{"Virtual", types.ObjectTypeFs | types.ObjVirtual, true, false, false},
		{"Ephemeral", types.ObjectTypeFs | types.ObjEphemeral, false, true, false},
		{"Physical", types.ObjectTypeFs | types.ObjPhysical, false, false, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			obj := makeObj(tc.oType, 0)
			inspector := NewObjectInspector(obj)

			if got := inspector.IsVirtual(); got != tc.virtual {
				t.Errorf("IsVirtual() = %v, want %v", got, tc.virtual)
			}
			if got := inspector.IsEphemeral(); got != tc.ephemeral {
				t.Errorf("IsEphemeral() = %v, want %v", got, tc.ephemeral)
			}
			if got := inspector.IsPhysical(); got != tc.physical {
				t.Errorf("IsPhysical() = %v, want %v", got, tc.physical)
			}
		})
	}
}

func TestEncryptionAndHeaderFlags(t *testing.T) {
	obj := makeObj(types.ObjectTypeFs|types.ObjEncrypted|types.ObjNonpersistent, 0)
	inspector := NewObjectInspector(obj)

	if !inspector.IsEncrypted() {
		t.Error("Expected IsEncrypted() = true")
	}
	if !inspector.IsNonpersistent() {
		t.Error("Expected IsNonpersistent() = true")
	}

	// Add no-header flag
	obj.OType |= types.ObjNoheader
	if inspector.HasHeader() {
		t.Error("Expected HasHeader() = false due to ObjNoheader flag")
	}
}
