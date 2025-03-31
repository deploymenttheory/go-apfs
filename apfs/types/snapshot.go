package types

// Snapshot Metadata (pages 117-121)
// Snapshots let you get a stable, read-only copy of the filesystem at a given point in time.
// Snapshots are designed to be fast and inexpensive to create;
// however, deleting a snapshot involves more work.

// JSnapMetadataKeyT is the key half of a record containing metadata about a snapshot.
// Reference: page 117
type JSnapMetadataKeyT struct {
	// The record's header. (page 117)
	// The object identifier in the header is the snapshot's transaction identifier.
	// The type in the header is always APFS_TYPE_SNAP_METADATA.
	Hdr JKeyT
}

// JSnapMetadataValT is the value half of a record containing metadata about a snapshot.
// Reference: page 117
type JSnapMetadataValT struct {
	// The physical object identifier of the B-tree that stores extents information. (page 117)
	ExtentrefTreeOid OidT

	// The physical object identifier of the volume superblock. (page 117)
	SblockOid OidT

	// The time that this snapshot was created. (page 118)
	// This timestamp is represented as the number of nanoseconds since January 1, 1970 at 0:00 UTC,
	// disregarding leap seconds.
	CreateTime uint64

	// The time that this snapshot was last modified. (page 118)
	// This timestamp is represented as the number of nanoseconds since January 1, 1970 at 0:00 UTC,
	// disregarding leap seconds.
	ChangeTime uint64

	// No overview available. (page 118)
	Inum uint64

	// The type of the B-tree that stores extents information. (page 118)
	ExtentrefTreeType uint32

	// A bit field that contains additional information about a snapshot metadata record. (page 118)
	// For the values used in this bit field, see snap_meta_flags.
	Flags uint32

	// The length of the snapshot's name, including the final null character (U+0000). (page 118)
	NameLen uint16

	// The snapshot's name, represented as a null-terminated UTF-8 string. (page 118)
	Name []byte
}

// JSnapNameKeyT is the key half of a snapshot name record.
// Reference: page 119
type JSnapNameKeyT struct {
	// The record's header. (page 119)
	// The object identifier in the header is always ~0ULL.
	// The type in the header is always APFS_TYPE_SNAP_NAME.
	Hdr JKeyT

	// The length of the extended attribute's name, including the final null character (U+0000). (page 119)
	NameLen uint16

	// The extended attribute's name, represented as a null-terminated UTF-8 string. (page 119)
	Name []byte
}

// JSnapNameValT is the value half of a snapshot name record.
// Reference: page 119
type JSnapNameValT struct {
	// The last transaction identifier included in the snapshot. (page 119)
	SnapXid XidT
}

// SnapMetaFlags contains snapshot metadata flags.
// Reference: page 119
type SnapMetaFlags uint32

const (
	// SnapMetaPendingDataless indicates the snapshot is pending being made dataless.
	// Reference: page 119
	SnapMetaPendingDataless SnapMetaFlags = 0x00000001

	// SnapMetaMergeInProgress indicates a merge is in progress for this snapshot.
	// Reference: page 119
	SnapMetaMergeInProgress SnapMetaFlags = 0x00000002
)

// SnapMetaExtObjPhysT contains additional metadata about snapshots.
// Reference: page 120
type SnapMetaExtObjPhysT struct {
	// The object's header. (page 120)
	SmeOpO ObjPhysT

	// Additional snapshot metadata. (page 120)
	SmeOpSme SnapMetaExtT
}

// SnapMetaExtT contains extended snapshot metadata.
// Reference: page 120
type SnapMetaExtT struct {
	// The version of this structure. (page 120)
	SmeVersion uint32

	// The flags for this extended snapshot metadata. (page 120)
	SmeFlags uint32

	// The snapshot's transaction identifier. (page 121)
	SmeSnapXid XidT

	// The snapshot's UUID. (page 121)
	SmeUuid UUID

	// Opaque metadata. (page 121)
	SmeToken uint64
}
