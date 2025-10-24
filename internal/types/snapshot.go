package types

// Snapshot Metadata (pages 1545-1799)
// Snapshots let you get a stable, read-only copy of the filesystem at a given point in time.

// JSnapMetadataKeyT is the key half of a record containing metadata about a snapshot.
// Reference: page 1551
type JSnapMetadataKeyT struct {
	// The record's header. (page 1565)
	// The object identifier in the header is the snapshot's transaction identifier.
	// The type in the header is always APFS_TYPE_SNAP_METADATA.
	Hdr JKeyT
}

// JSnapMetadataValT is the value half of a record containing metadata about a snapshot.
// Reference: page 1571
type JSnapMetadataValT struct {
	// The physical object identifier of the B-tree that stores extents information. (page 1592)
	ExtentrefTreeOid OidT

	// The physical object identifier of the volume superblock. (page 1596)
	SblockOid OidT

	// The time that this snapshot was created. (page 1599)
	// Represented as nanoseconds since January 1, 1970 at 000 UTC, disregarding leap seconds.
	CreateTime uint64

	// The time that this snapshot was last modified. (page 1603)
	// Represented as nanoseconds since January 1, 1970 at 000 UTC, disregarding leap seconds.
	ChangeTime uint64

	// The inode number associated with this snapshot. (page 1608)
	Inum uint64

	// The type of the B-tree that stores extents information. (page 1611)
	ExtentrefTreeType uint32

	// A bit field that contains additional information about a snapshot metadata record. (page 1615)
	// For the values used in this bit field, see snap_meta_flags.
	Flags uint32

	// The length of the snapshot's name, including the final null character (U+0000). (page 1620)
	NameLen uint16

	// The snapshot's name, represented as a null-terminated UTF-8 string. (page 1623)
	Name []byte
}

// JSnapNameKeyT is the key half of a snapshot name record.
// Reference: page 1652
type JSnapNameKeyT struct {
	// The record's header. (page 1665)
	// The object identifier in the header is always ~0ULL.
	// The type in the header is always APFS_TYPE_SNAP_NAME.
	Hdr JKeyT

	// The length of the snapshot's name, including the final null character (U+0000). (page 1671)
	NameLen uint16

	// The snapshot's name, represented as a null-terminated UTF-8 string. (page 1674)
	Name []byte
}

// JSnapNameValT is the value half of a snapshot name record.
// Reference: page 1682
type JSnapNameValT struct {
	// The last transaction identifier included in the snapshot. (page 1689)
	SnapXid XidT
}

// SnapMetaFlags represents bit flags for snapshot metadata records.
// Reference: page 1699
type SnapMetaFlags uint32

const (
	// SnapMetaPendingDataless indicates that the snapshot is pending dataless operation.
	// Reference: page 1715
	SnapMetaPendingDataless SnapMetaFlags = 0x00000001

	// SnapMetaMergeInProgress indicates that a merge operation is in progress for this snapshot.
	// Reference: page 1721
	SnapMetaMergeInProgress SnapMetaFlags = 0x00000002
)

// SnapMetaExtObjPhysT is additional metadata about snapshots.
// Reference: page 1727
type SnapMetaExtObjPhysT struct {
	// The physical object header. (page 1733)
	SmeObjPhys ObjPhysT

	// The snapshot metadata extension. (page 1738)
	SmeopSme SnapMetaExtT
}

// SnapMetaExtT is extended metadata for snapshots.
// Reference: page 1744
type SnapMetaExtT struct {
	// The version of this structure. (page 1772)
	SmeVersion uint32

	// Flags for this snapshot metadata extension. (page 1775)
	SmeFlags uint32

	// The snapshot's transaction identifier. (page 1781)
	SmeSnapXid XidT

	// The snapshot's UUID. (page 1787)
	SmeUuid UUID

	// Opaque metadata token. (page 1793)
	SmeToken uint64
}
