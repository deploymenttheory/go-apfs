package file_system_objects

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestFileSystemObjectTypeResolver_ResolveObjectType(t *testing.T) {
	resolver := NewFileSystemObjectTypeResolver()

	tests := []struct {
		objType  types.JObjTypes
		expected string
	}{
		{types.ApfsTypeAny, "Any"},
		{types.ApfsTypeSnapMetadata, "Snapshot Metadata"},
		{types.ApfsTypeExtent, "Physical Extent"},
		{types.ApfsTypeInode, "Inode"},
		{types.ApfsTypeXattr, "Extended Attribute"},
		{types.ApfsTypeSiblingLink, "Sibling Link"},
		{types.ApfsTypeDstreamId, "Data Stream"},
		{types.ApfsTypeCryptoState, "Crypto State"},
		{types.ApfsTypeFileExtent, "File Extent"},
		{types.ApfsTypeDirRec, "Directory Entry"},
		{types.ApfsTypeDirStats, "Directory Statistics"},
		{types.ApfsTypeSnapName, "Snapshot Name"},
		{types.ApfsTypeSiblingMap, "Sibling Map"},
		{types.ApfsTypeFileInfo, "File Info"},
		{types.ApfsTypeInvalid, "Invalid"},
		{types.JObjTypes(255), "Unknown"}, // Unknown type
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := resolver.ResolveObjectType(tt.objType)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFileSystemObjectTypeResolver_ResolveObjectKind(t *testing.T) {
	resolver := NewFileSystemObjectTypeResolver()

	tests := []struct {
		objKind  types.JObjKinds
		expected string
	}{
		{types.ApfsKindAny, "Any"},
		{types.ApfsKindNew, "New"},
		{types.ApfsKindUpdate, "Update"},
		{types.ApfsKindDead, "Dead"},
		{types.JObjKinds(255), "Unknown"}, // Unknown kind
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := resolver.ResolveObjectKind(tt.objKind)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFileSystemObjectTypeResolver_ListSupportedObjectTypes(t *testing.T) {
	resolver := NewFileSystemObjectTypeResolver()

	supportedTypes := resolver.ListSupportedObjectTypes()

	// Check that we have the expected number of supported types
	expectedTypes := []types.JObjTypes{
		types.ApfsTypeSnapMetadata,
		types.ApfsTypeExtent,
		types.ApfsTypeInode,
		types.ApfsTypeXattr,
		types.ApfsTypeSiblingLink,
		types.ApfsTypeDstreamId,
		types.ApfsTypeCryptoState,
		types.ApfsTypeFileExtent,
		types.ApfsTypeDirRec,
		types.ApfsTypeDirStats,
		types.ApfsTypeSnapName,
		types.ApfsTypeSiblingMap,
		types.ApfsTypeFileInfo,
	}

	if len(supportedTypes) != len(expectedTypes) {
		t.Errorf("Expected %d supported types, got %d", len(expectedTypes), len(supportedTypes))
	}

	// Check that all expected types are present
	typeMap := make(map[types.JObjTypes]bool)
	for _, objType := range supportedTypes {
		typeMap[objType] = true
	}

	for _, expectedType := range expectedTypes {
		if !typeMap[expectedType] {
			t.Errorf("Expected type %d not found in supported types", expectedType)
		}
	}

	// Verify that ApfsTypeAny and ApfsTypeInvalid are not in the list
	if typeMap[types.ApfsTypeAny] {
		t.Error("ApfsTypeAny should not be in supported types list")
	}
	if typeMap[types.ApfsTypeInvalid] {
		t.Error("ApfsTypeInvalid should not be in supported types list")
	}
}

func TestFileSystemObjectTypeResolver_ComprehensiveMapping(t *testing.T) {
	resolver := NewFileSystemObjectTypeResolver()

	// Test that all supported types can be resolved to non-"Unknown" strings
	supportedTypes := resolver.ListSupportedObjectTypes()

	for _, objType := range supportedTypes {
		result := resolver.ResolveObjectType(objType)
		if result == "Unknown" {
			t.Errorf("Supported type %d resolved to 'Unknown'", objType)
		}
		if result == "" {
			t.Errorf("Supported type %d resolved to empty string", objType)
		}
	}
}

func TestFileSystemObjectTypeResolver_Interface(t *testing.T) {
	// Test that the type resolver properly implements the interface
	var resolver any = NewFileSystemObjectTypeResolver()

	// This should compile without issues if the interface is properly implemented
	if _, ok := resolver.(interfaces.FileSystemObjectTypeResolver); !ok {
		t.Error("fileSystemObjectTypeResolver does not implement FileSystemObjectTypeResolver interface")
	}
}
