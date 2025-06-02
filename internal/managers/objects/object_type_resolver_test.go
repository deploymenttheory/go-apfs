package objects

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestResolveType(t *testing.T) {
	resolver := NewStaticObjectTypeResolver()

	tests := []struct {
		name     string
		input    uint32
		expected string
	}{
		{"FS", types.ObjectTypeFs, "APFS Volume"},
		{"Btree Node", types.ObjectTypeBtreeNode, "B-tree Node"},
		{"NX Superblock", types.ObjectTypeNxSuperblock, "NX Superblock"},
		{"Media Keybag", types.ObjectTypeMediaKeybag, "Media Keybag"},
		{"Encrypted Keybag (with flags)", types.ObjectTypeMediaKeybag | types.ObjEncrypted, "Media Keybag"},
		{"Unknown Type", 0xDEADBEEF, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.ResolveType(tt.input)
			if result != tt.expected {
				t.Errorf("ResolveType(%#x) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetObjectTypeCategory(t *testing.T) {
	resolver := NewStaticObjectTypeResolver()

	tests := []struct {
		name     string
		input    uint32
		expected string
	}{
		{"Volume", types.ObjectTypeFs, "File System"},
		{"Container Keybag", types.ObjectTypeContainerKeybag, "Security"},
		{"Volume Keybag", types.ObjectTypeVolumeKeybag, "Security"},
		{"Unknown Type", 0x12345678, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.GetObjectTypeCategory(tt.input)
			if result != tt.expected {
				t.Errorf("GetObjectTypeCategory(%#x) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSupportedObjectTypes(t *testing.T) {
	resolver := NewStaticObjectTypeResolver()
	supported := resolver.SupportedObjectTypes()

	if len(supported) == 0 {
		t.Fatal("SupportedObjectTypes() returned an empty list")
	}

	typeSeen := make(map[uint32]bool)
	for _, typ := range supported {
		if typeSeen[typ] {
			t.Errorf("Duplicate type found in SupportedObjectTypes: %#x", typ)
		}
		typeSeen[typ] = true

		// Validate resolve/type category work on every supported type
		name := resolver.ResolveType(typ)
		if name == "Unknown" {
			t.Errorf("SupportedObjectTypes() includes %#x but ResolveType() returns Unknown", typ)
		}

		category := resolver.GetObjectTypeCategory(typ)
		if category == "Unknown" {
			t.Errorf("SupportedObjectTypes() includes %#x but GetObjectTypeCategory() returns Unknown", typ)
		}
	}
}
