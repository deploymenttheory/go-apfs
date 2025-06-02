package extendedfields

import (
	"testing"
)

func TestExtendedFieldTypeResolver_ResolveName(t *testing.T) {
	resolver := NewExtendedFieldTypeResolver()

	tests := []struct {
		fieldType uint8
		expected  string
	}{
		{1, "Document ID"},
		{2, "Finder Info"},
		{3, "Sparse Bytes"},
		{4, "Previous File Size"},
		{5, "Snapshot XID"},
		{6, "Delta Tree OID"},
		{7, "Device ID"},
		{8, "Original Sync Root ID"},
		{9, "Sibling ID"},
		{10, "Filesystem UUID"},
		{99, "Unknown"}, // unknown
	}

	for _, tt := range tests {
		name := resolver.ResolveName(tt.fieldType)
		if name != tt.expected {
			t.Errorf("ResolveName(%d) = %q; want %q", tt.fieldType, name, tt.expected)
		}
	}
}

func TestExtendedFieldTypeResolver_ResolveDescription(t *testing.T) {
	resolver := NewExtendedFieldTypeResolver()

	tests := map[uint8]string{
		1:  "The APFS document ID for the file.",
		2:  "Extended Finder information blob.",
		3:  "Count of sparse bytes in the file's data stream.",
		4:  "Used for crash recovery; file size prior to last update.",
		5:  "Transaction ID of the snapshot that created this inode.",
		6:  "Tree OID for snapshot delta information.",
		7:  "Device identifier for special files.",
		8:  "Original hierarchy inode ID for sync root.",
		9:  "Hardlink sibling ID.",
		10: "UUID of mounted filesystem target.",
		42: "No description available",
	}

	for fieldType, expected := range tests {
		desc := resolver.ResolveDescription(fieldType)
		if desc != expected {
			t.Errorf("ResolveDescription(%d) = %q; want %q", fieldType, desc, expected)
		}
	}
}

func TestExtendedFieldTypeResolver_ListSupportedFieldTypes(t *testing.T) {
	resolver := NewExtendedFieldTypeResolver()
	supported := resolver.ListSupportedFieldTypes()

	expectedCount := 10
	if len(supported) != expectedCount {
		t.Errorf("Expected %d supported types, got %d", expectedCount, len(supported))
	}

	found := make(map[uint8]bool)
	for _, typ := range supported {
		found[typ] = true
	}

	for _, typ := range []uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} {
		if !found[typ] {
			t.Errorf("Expected field type %d to be listed", typ)
		}
	}
}
