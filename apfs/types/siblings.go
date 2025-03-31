package types

// Siblings (pages 115-116)
// Hard links that all refer to the same inode are called siblings.
// Each sibling has its own identifier that's used instead of the shared inode number
// when siblings need to be distinguished.

// JSiblingKeyT is the key half of a sibling-link record.
// Reference: page 115
type JSiblingKeyT struct {
	// The record's header. (page 115)
	// The object identifier in the header is the file-system object's identifier, that is, its inode number.
	// The type in the header is always APFS_TYPE_SIBLING_LINK.
	Hdr JKeyT

	// The sibling's unique identifier. (page 115)
	// This value matches the object identifier for the sibling map record (j_sibling_key_t).
	SiblingId uint64
}

// JSiblingValT is the value half of a sibling-link record.
// Reference: page 115
type JSiblingValT struct {
	// The object identifier for the inode that's the parent directory. (page 116)
	ParentId uint64

	// The length of the name, including the final null character (U+0000). (page 116)
	NameLen uint16

	// The name, represented as a null-terminated UTF-8 string. (page 116)
	Name []byte
}

// JSiblingMapKeyT is the key half of a sibling-map record.
// Reference: page 116
type JSiblingMapKeyT struct {
	// The record's header. (page 116)
	// The object identifier in the header is the sibling's unique identifier,
	// which matches the sibling_id field of j_sibling_key_t.
	// The type in the header is always APFS_TYPE_SIBLING_MAP.
	Hdr JKeyT
}

// JSiblingMapValT is the value half of a sibling-map record.
// Reference: page 116
type JSiblingMapValT struct {
	// The inode number of the underlying file. (page 116)
	FileId uint64
}
