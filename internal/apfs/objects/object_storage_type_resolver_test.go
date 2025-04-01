package objects

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestDetermineStorageType(t *testing.T) {
	resolver := NewStaticObjectStorageTypeResolver()

	tests := []struct {
		name       string
		input      uint32
		wantResult string
	}{
		{"Virtual", types.ObjectTypeFs | types.ObjVirtual, "virtual"},
		{"Ephemeral", types.ObjectTypeFs | types.ObjEphemeral, "ephemeral"},
		{"Physical", types.ObjectTypeFs | types.ObjPhysical, "physical"},
		{"Virtual (no storage bits set)", 0x30000000, "virtual"},
		{"Invalid combo (ephemeral | physical)", types.ObjEphemeral | types.ObjPhysical, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.DetermineStorageType(tt.input)
			if result != tt.wantResult {
				t.Errorf("DetermineStorageType(%#x) = %q; want %q", tt.input, result, tt.wantResult)
			}
		})
	}
}

func TestIsStorageTypeSupported(t *testing.T) {
	resolver := NewStaticObjectStorageTypeResolver()

	tests := []struct {
		input    string
		expected bool
	}{
		{"virtual", true},
		{"Virtual", true},
		{"EPHEMERAL", true},
		{"physical", true},
		{"garbage", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := resolver.IsStorageTypeSupported(tt.input)
			if got != tt.expected {
				t.Errorf("IsStorageTypeSupported(%q) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}
