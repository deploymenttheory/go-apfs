package objects

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestLookupType(t *testing.T) {
	registry := NewStaticObjectRegistry()

	tests := []struct {
		name         string
		input        uint32
		wantFound    bool
		wantTypeName string
	}{
		{"Exact Match - Volume", types.ObjectTypeFs, true, "APFS Volume"},
		{"Exact Match - Container Keybag", types.ObjectTypeContainerKeybag, true, "Container Keybag"},
		{"With Flags - Volume (Encrypted)", types.ObjectTypeFs | types.ObjEncrypted, true, "APFS Volume"},
		{"Unknown Type", 0xdeadbeef, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, ok := registry.LookupType(tt.input)
			if ok != tt.wantFound {
				t.Fatalf("LookupType(%#x): found = %v, want %v", tt.input, ok, tt.wantFound)
			}
			if ok && info.Name != tt.wantTypeName {
				t.Errorf("LookupType(%#x): name = %q, want %q", tt.input, info.Name, tt.wantTypeName)
			}
		})
	}
}

func TestListObjectTypes(t *testing.T) {
	registry := NewStaticObjectRegistry()
	all := registry.ListObjectTypes()

	if len(all) == 0 {
		t.Fatal("ListObjectTypes() returned 0 results")
	}

	// Ensure a known type is present
	var found bool
	for _, info := range all {
		if info.Type == types.ObjectTypeFs {
			found = true
			if info.Name != "APFS Volume" {
				t.Errorf("Expected 'APFS Volume', got %q", info.Name)
			}
		}
	}

	if !found {
		t.Error("Expected ObjectTypeFs to be in ListObjectTypes()")
	}

	// Ensure no duplicates
	seen := map[uint32]bool{}
	for _, info := range all {
		if seen[info.Type] {
			t.Errorf("Duplicate type found in ListObjectTypes: %#x", info.Type)
		}
		seen[info.Type] = true
	}
}
