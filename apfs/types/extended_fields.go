package types

// Extended Fields (pages 108-114)
// Directory entries and inodes use extended fields to store a dynamically extensible set of member fields.

// XfBlobT is a collection of extended attributes.
// Reference: page 108
type XfBlobT struct {
	// The number of extended attributes. (page 108)
	XfNumExts uint16

	// The amount of space, in bytes, used to store the extended attributes. (page 108)
	// This total includes both the space used to store metadata, as instances of x_field_t, and values.
	XfUsedData uint16

	// The extended fields. (page 109)
	// This field contains an array of instances of x_field_t, followed by the extended field data.
	XfData []byte
}

// XFieldT is an extended field's metadata.
// Reference: page 109
type XFieldT struct {
	// The extended field's data type. (page 109)
	// For possible values, see Extended-Field Types.
	XType uint8

	// The extended field's flags. (page 109)
	// For the values used in this bit field, see Extended-Field Flags.
	XFlags uint8

	// The size, in bytes, of the data stored in the extended field. (page 109)
	XSize uint16
}

// Extended-Field Types (pages 109-112)

// DrecExtTypeSiblingId is the sibling identifier for a directory record (uint64_t).
// Reference: page 110
// The corresponding sibling-link record has the same identifier in the sibling_id field of j_sibling_key_t.
// This extended field is used only for hard links.
const DrecExtTypeSiblingId uint8 = 1

// InoExtTypeSnapXid is the transaction identifier for a snapshot (xid_t).
// Reference: page 110
const InoExtTypeSnapXid uint8 = 1

// InoExtTypeDeltaTreeOid is the virtual object identifier of the file-system tree
// that corresponds to a snapshot's extent delta list (oid_t).
// Reference: page 110
// The tree object's subtype is always OBJECT_TYPE_FSTREE.
const InoExtTypeDeltaTreeOid uint8 = 2

// InoExtTypeDocumentId is the file's document identifier (uint32_t).
// Reference: page 110
// The document identifier lets applications keep track of the document during operations like atomic save,
// where one folder replaces another. The document identifier remains associated with the full path,
// not just with the inode that's currently at that path.
// Implementations of Apple File System must preserve the document identifier when the inode
// at that path is replaced.
const InoExtTypeDocumentId uint8 = 3

// InoExtTypeName is the name of the file, represented as a null-terminated UTF-8 string.
// Reference: page 111
// This extended field is used only for hard links: The name stored in the inode is the name
// of the primary link to the file, and the name of the hard link is stored in this extended field.
const InoExtTypeName uint8 = 4

// InoExtTypePrevFsize is the file's previous size (uint64_t).
// Reference: page 111
// This extended field is used for recovering after a crash.
// If it's set on an inode, truncate the file back to the size contained in this field.
const InoExtTypePrevFsize uint8 = 5

// InoExtTypeReserved6 is reserved.
// Reference: page 111
// Don't create extended fields of this type in your own code.
// Preserve the value of any extended fields of this type.
const InoExtTypeReserved6 uint8 = 6

// InoExtTypeFinderInfo is opaque data stored and used by Finder (32 bytes).
// Reference: page 111
const InoExtTypeFinderInfo uint8 = 7

// InoExtTypeDstream is a data stream (j_dstream_t).
// Reference: page 111
const InoExtTypeDstream uint8 = 8

// InoExtTypeReserved9 is reserved.
// Reference: page 111
// Don't create extended fields of this type.
// When you modify an existing volume, preserve the contents of any extended fields of this type.
const InoExtTypeReserved9 uint8 = 9

// InoExtTypeDirStatsKey is statistics about a directory (j_dir_stats_val_t).
// Reference: page 111
const InoExtTypeDirStatsKey uint8 = 10

// InoExtTypeFsUuid is the UUID of a file system that's automatically mounted in this directory (uuid_t).
// Reference: page 112
// This value matches the value of the apfs_vol_uuid field of apfs_superblock_t.
const InoExtTypeFsUuid uint8 = 11

// InoExtTypeReserved12 is reserved.
// Reference: page 112
// Don't create extended fields of this type.
// If you find an object of this type in production, file a bug against the Apple File System implementation.
const InoExtTypeReserved12 uint8 = 12

// InoExtTypeSparseBytes is the number of sparse bytes in the data stream (uint64_t).
// Reference: page 112
const InoExtTypeSparseBytes uint8 = 13

// InoExtTypeRdev is the device identifier for a block- or character-special device (uint32_t).
// Reference: page 112
// This extended field stores the same information as the st_rdev field of the stat structure
// defined in <sys/stat.h>.
const InoExtTypeRdev uint8 = 14

// InoExtTypePurgeableFlags is information about a purgeable file.
// Reference: page 112
// The value of this extended field is reserved.
// Don't create new extended fields of this type.
// When duplicating a file or directory, omit this extended field from the new copy.
const InoExtTypePurgeableFlags uint8 = 15

// InoExtTypeOrigSyncRootId is the inode number of the sync-root hierarchy
// that this file originally belonged to.
// Reference: page 112
// The specified inode always has the INODE_IS_SYNC_ROOT flag set.
const InoExtTypeOrigSyncRootId uint8 = 16

// Extended-Field Flags (pages 113-114)

// XfDataDependent indicates the data in this extended field depends on the file's data.
// Reference: page 113
// When the file data changes, this extended field must be updated to match the new data.
// If it's not possible to update the field, the field must be removed.
const XfDataDependent uint16 = 0x0001

// XfDoNotCopy indicates when copying this file, omit this extended field from the copy.
// Reference: page 113
const XfDoNotCopy uint16 = 0x0002

// XfReserved4 is reserved.
// Reference: page 113
// Don't set this flag, but preserve it if it's already set.
const XfReserved4 uint16 = 0x0004

// XfChildrenInherit indicates when creating a new entry in this directory,
// copy this extended field to the new directory entry.
// Reference: page 113
const XfChildrenInherit uint16 = 0x0008

// XfUserField indicates this extended field was added by a user-space program.
// Reference: page 113
const XfUserField uint16 = 0x0010

// XfSystemField indicates this extended field was added by the kernel, by the implementation
// of Apple File System, or by another system component.
// Reference: page 113
// Extended fields with this flag set can't be removed or modified by a program running in user space.
const XfSystemField uint16 = 0x0020

// XfReserved40 is reserved.
// Reference: page 114
// Don't set this flag, but preserve it if it's already set.
const XfReserved40 uint16 = 0x0040

// XfReserved80 is reserved.
// Reference: page 114
// Don't set this flag, but preserve it if it's already set.
const XfReserved80 uint16 = 0x0080
