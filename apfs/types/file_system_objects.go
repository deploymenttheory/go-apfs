package types

// File-System Objects (pages 71-101)
// A file-system object stores information about a part of the file system,
// like a directory or a file on disk. These objects are stored as one or more records.

// JKeyT is a header used at the beginning of all file-system keys.
// Reference: page 72
type JKeyT struct {
	// A bit field that contains the object's identifier and its type. (page 72)
	// The object's identifier is a uint64_t value accessed as obj_id_and_type & OBJ_ID_MASK.
	// The object's type is a uint8_t value accessed as (obj_id_and_type & OBJ_TYPE_MASK) >> OBJ_TYPE_SHIFT.
	ObjIdAndType uint64
}

// ObjIdMask is the bit mask used to access the object identifier.
// Reference: page 73
const ObjIdMask uint64 = 0x0fffffffffffffff

// ObjTypeMask is the bit mask used to access the object type.
// Reference: page 73
const ObjTypeMask uint64 = 0xf000000000000000

// ObjTypeShift is the bit shift used to access the object type.
// Reference: page 73
const ObjTypeShift uint64 = 60

// SystemObjIdMark is the smallest object identifier used by the system volume.
// Reference: page 73
const SystemObjIdMark uint64 = 0x0fffffff00000000

// JInodeKeyT is the key half of a directory-information record.
// Reference: page 73
type JInodeKeyT struct {
	// The record's header. (page 73)
	// The object identifier in the header is the file-system object's identifier, also known as its inode number.
	// The type in the header is always APFS_TYPE_INODE.
	Hdr JKeyT
}

// JInodeValT is the value half of an inode record.
// Reference: page 73-77
type JInodeValT struct {
	// The identifier of the file system record for the parent directory. (page 74)
	ParentId uint64

	// The unique identifier used by this file's data stream. (page 74)
	// This identifier appears in the owning_obj_id field of j_phys_ext_val_t records
	// that describe the extents where the data is stored.
	// For an inode that doesn't have data, the value of this field is the file-system object's identifier.
	PrivateId uint64

	// The time that this record was created. (page 75)
	// This timestamp is represented as the number of nanoseconds since January 1, 1970 at 0:00 UTC,
	// disregarding leap seconds.
	CreateTime uint64

	// The time that this record was last modified. (page 75)
	// This timestamp is represented as the number of nanoseconds since January 1, 1970 at 0:00 UTC,
	// disregarding leap seconds.
	ModTime uint64

	// The time that this record's attributes were last modified. (page 75)
	// This timestamp is represented as the number of nanoseconds since January 1, 1970 at 0:00 UTC,
	// disregarding leap seconds.
	ChangeTime uint64

	// The time that this record was last accessed. (page 75)
	// This timestamp is represented as the number of nanoseconds since January 1, 1970 at 0:00 UTC,
	// disregarding leap seconds.
	AccessTime uint64

	// The inode's flags. (page 75)
	// For the values used in this bit field, see j_inode_flags.
	InternalFlags uint64

	// A union field that can be either nchildren or nlink.
	// Only one of these fields is valid at a time, depending on whether the inode is a directory.

	// Union field: nchildren for directories, nlink for files
	// For directories: The number of directory entries
	// For files: The number of hard links whose target is this inode
	NchildrenOrNlink int32

	// The default protection class for this inode. (page 76)
	// Files in this directory that have a protection class of PROTECTION_CLASS_DIR_NONE use
	// the directory's default protection class.
	DefaultProtectionClass CpKeyClassT

	// A monotonically increasing counter that's incremented each time this inode or its data is modified. (page 76)
	// This value is allowed to overflow and restart from zero.
	WriteGenerationCounter uint32

	// The inode's BSD flags. (page 76)
	// For information about these flags, see the chflags(2) man page and the <sys/stat.h> header file.
	BsdFlags uint32

	// The user identifier of the inode's owner. (page 76)
	Owner UidT

	// The group identifier of the inode's group. (page 76)
	Group GidT

	// The file's mode. (page 77)
	// For possible values, see File Modes.
	Mode ModeT

	// Reserved. (page 77)
	// Populate this field with zero when you create a new inode,
	// and preserve its value when you modify an existing inode.
	// This field is padding.
	Pad1 uint16

	// The size of the file without compression. (page 77)
	// This field is populated only for files that have the INODE_HAS_UNCOMPRESSED_SIZE flag
	// set on the internal_flags field.
	// For files that don't have the flag set, this field is treated as padding.
	UncompressedSize uint64

	// The inode's extended fields. (page 77)
	// This location on disk contains several pieces of data that have variable sizes.
	// For information about reading extended fields, see Extended Fields.
	XFields []byte
}

// Nchildren returns the number of directory entries.
// This method is only valid if the inode represents a directory.
// Calling this on a non-directory inode will return an undefined value.
func (v *JInodeValT) Nchildren() int32 {
	// Only valid if inode is a directory
	return v.NchildrenOrNlink
}

// Nlink returns the number of hard links to this inode.
// This method is only valid if the inode does not represent a directory.
// Calling this on a directory inode will return an undefined value.
func (v *JInodeValT) Nlink() int32 {
	// Only valid if inode is not a directory
	return v.NchildrenOrNlink
}

// UidT is a user identifier.
// Reference: page 77
type UidT uint32

// GidT is a group identifier.
// Reference: page 77
type GidT uint32

// JDrecKeyT is the key half of a directory entry record.
// Reference: page 78
type JDrecKeyT struct {
	// The record's header. (page 78)
	// The object identifier in the header is the file-system object's identifier.
	// The type in the header is always APFS_TYPE_DIR_REC.
	Hdr JKeyT

	// The length of the name, including the final null character (U+0000). (page 78)
	NameLen uint16

	// The name, represented as a null-terminated UTF-8 string. (page 78)
	Name []byte
}

// JDrecHashedKeyT is the key half of a directory entry record, including a precomputed hash of its name.
// Reference: page 78
type JDrecHashedKeyT struct {
	// The record's header. (page 78)
	// The object identifier in the header is the file-system object's identifier.
	// The type in the header is always APFS_TYPE_DIR_REC.
	Hdr JKeyT

	// The hash and length of the name. (page 79)
	// The length is a 10-bit unsigned integer, accessed as name_len_and_hash & J_DREC_LEN_MASK.
	// The length includes the final null character (U+0000).
	// The hash is an unsigned 22-bit integer, accessed as
	// (name_len_and_hash & J_DREC_HASH_MASK) >> J_DREC_HASH_SHIFT.
	NameLenAndHash uint32

	// The name, represented as a null-terminated UTF-8 string. (page 79)
	Name []byte
}

// JDrecLenMask is the bit mask used to access the length of the name.
// Reference: page 79
const JDrecLenMask uint32 = 0x000003ff

// JDrecHashMask is the bit mask used to access the hash of the name.
// Reference: page 79
const JDrecHashMask uint32 = 0xfffff400

// JDrecHashShift is the bit shift used to access the hash of the name.
// Reference: page 79
const JDrecHashShift uint32 = 10

// JDrecValT is the value half of a directory entry record.
// Reference: page 79
type JDrecValT struct {
	// The identifier of the inode that this directory entry represents. (page 80)
	FileId uint64

	// The time that this directory entry was added to the directory. (page 80)
	// This timestamp is represented as the number of nanoseconds since January 1, 1970 at 0:00 UTC,
	// disregarding leap seconds. It's not updated when modifying the directory entry.
	DateAdded uint64

	// The directory entry's flags. (page 80)
	// The bits that are set in DREC_TYPE_MASK store the inode's file type,
	// and the remaining bits are reserved.
	// For possible values, see Directory Entry File Types.
	Flags uint16

	// The directory entry's extended fields. (page 80)
	// This location on disk contains several pieces of data that have variable sizes.
	// For information about reading extended fields, see Extended Fields.
	XFields []byte
}

// JDirStatsKeyT is the key half of a directory-information record.
// Reference: page 80
type JDirStatsKeyT struct {
	// The record's header. (page 81)
	// The object identifier in the header is the file-system object's identifier.
	// The type in the header is always APFS_TYPE_DIR_STATS.
	Hdr JKeyT
}

// JDirStatsValT is the value half of a directory-information record.
// Reference: page 81
type JDirStatsValT struct {
	// The number of files and folders contained by the directory. (page 81)
	NumChildren uint64

	// The total size, in bytes, of all the files stored in this directory
	// and all of this directory's descendants. (page 81)
	// Hard links contribute to the total_size of every directory they appear in.
	TotalSize uint64

	// The parent directory's file system object identifier. (page 81)
	ChainedKey uint64

	// A monotonically increasing counter that's incremented each time this inode
	// or any of its children is modified. (page 81)
	// If this counter can't be incremented without overflow, that's an unrecoverable error.
	GenCount uint64
}

// JXattrKeyT is the key half of an extended attribute record.
// Reference: page 82
type JXattrKeyT struct {
	// The record's header. (page 82)
	// The object identifier in the header is the file-system object's identifier.
	// The type in the header is always APFS_TYPE_XATTR.
	Hdr JKeyT

	// The length of the extended attribute's name, including the final null character (U+0000). (page 82)
	NameLen uint16

	// The extended attribute's name, represented as a null-terminated UTF-8 string. (page 82)
	Name []byte
}

// JXattrValT is the value half of an extended attribute record.
// Reference: page 82
type JXattrValT struct {
	// The extended attribute record's flags. (page 82)
	// For the values used in this bit field, see j_xattr_flags.
	// Either the XATTR_DATA_EMBEDDED or XATTR_DATA_STREAM flag must be set.
	Flags uint16

	// The length of the extended attribute data. (page 83)
	// If the XATTR_DATA_EMBEDDED flag is set, this field is the length of the data in the xdata field.
	// Otherwise, this field is ignored.
	XdataLen uint16

	// The extended attribute data or the identifier of a data stream that contains the data. (page 83)
	// If the XATTR_DATA_EMBEDDED flag is set, the extended attribute data is stored directly in this field.
	// Otherwise, this field contains the identifier (uint64_t) for a data stream record
	// that stores the extended attribute data.
	Xdata []byte
}

// JObjTypes represents the type of a file-system record.
// Reference: page 84
type JObjTypes uint8

const (
	// ApfsTypeAny is a record of any type.
	// Reference: page 84
	// This enumeration case is used only in search queries and in tests when iterating over objects.
	// It's not valid as the type of a file-system object.
	ApfsTypeAny JObjTypes = 0

	// ApfsTypeSnapMetadata is metadata about a snapshot.
	// Reference: page 84
	// The key is an instance of j_snap_metadata_key_t and the value is an instance of j_snap_metadata_val_t.
	ApfsTypeSnapMetadata JObjTypes = 1

	// ApfsTypeExtent is a physical extent record.
	// Reference: page 85
	// The key is an instance of j_phys_ext_key_t and the value is an instance of j_phys_ext_val_t.
	ApfsTypeExtent JObjTypes = 2

	// ApfsTypeInode is an inode.
	// Reference: page 85
	// The key is an instance of j_inode_key_t and the value is an instance of j_inode_val_t.
	ApfsTypeInode JObjTypes = 3

	// ApfsTypeXattr is an extended attribute.
	// Reference: page 85
	// The key is an instance of j_xattr_key_t and the value is an instance of j_xattr_val_t.
	ApfsTypeXattr JObjTypes = 4

	// ApfsTypeSiblingLink is a mapping from an inode to hard links that the inode is the target of.
	// Reference: page 85
	// The key is an instance of j_sibling_key_t and the value is an instance of j_sibling_val_t.
	ApfsTypeSiblingLink JObjTypes = 5

	// ApfsTypeDstreamId is a data stream.
	// Reference: page 85
	// The key is an instance of j_dstream_id_key_t and the value is an instance of j_dstream_id_val_t.
	ApfsTypeDstreamId JObjTypes = 6

	// ApfsTypeCryptoState is a per-file encryption state.
	// Reference: page 85
	// The key is an instance of j_crypto_key_t and the value is an instance of j_crypto_val_t.
	// This object type is used only by iOS devices, except for a placeholder object whose identifier
	// is always CRYPTO_SW_ID.
	ApfsTypeCryptoState JObjTypes = 7

	// ApfsTypeFileExtent is a physical extent record for a file.
	// Reference: page 85
	// The key is an instance of j_file_extent_key_t and the value is an instance of j_file_extent_val_t.
	ApfsTypeFileExtent JObjTypes = 8

	// ApfsTypeDirRec is a directory entry.
	// Reference: page 86
	// The key is an instance of j_drec_key_t and the value is an instance of j_drec_val_t.
	ApfsTypeDirRec JObjTypes = 9

	// ApfsTypeDirStats is information about a directory.
	// Reference: page 86
	// The key is an instance of j_dir_stats_key_t and the value is an instance of j_drec_val_t.
	ApfsTypeDirStats JObjTypes = 10

	// ApfsTypeSnapName is the name of a snapshot.
	// Reference: page 86
	// The key is an instance of j_snap_name_key_t and the value is an instance of j_snap_name_val_t.
	ApfsTypeSnapName JObjTypes = 11

	// ApfsTypeSiblingMap is a mapping from a hard link to its target inode.
	// Reference: page 86
	// The key is an instance of j_sibling_map_key_t and the value is an instance of j_sibling_map_val_t.
	ApfsTypeSiblingMap JObjTypes = 12

	// ApfsTypeFileInfo is additional information about file data.
	// Reference: page 86
	// The key is an instance of j_file_info_key_t and the value is an instance of j_file_info_val_t.
	ApfsTypeFileInfo JObjTypes = 13

	// ApfsTypeMaxValid is the largest valid value for a file-system object's type.
	// Reference: page 86
	ApfsTypeMaxValid JObjTypes = 13

	_ // Reserved slot to match apple's spec

	// ApfsTypeMax is the largest value for a file-system object's type.
	// Reference: page 86
	ApfsTypeMax JObjTypes = 15

	// ApfsTypeInvalid is an invalid object type.
	// Reference: page 87
	ApfsTypeInvalid JObjTypes = 15
)

// JObjKinds represents the kind of a file-system record.
// Reference: page 87
type JObjKinds uint8

const (
	// ApfsKindAny is a record of any kind.
	// Reference: page 87
	// This value isn't valid as the kind of a file-system record on disk.
	// However, implementations of Apple File System can use it internally —
	// for example, in search queries and in tests when iterating over objects.
	ApfsKindAny JObjKinds = 0

	// ApfsKindNew is a new record.
	// Reference: page 87
	// This record adds data that isn't part of any snapshots.
	ApfsKindNew JObjKinds = 1

	// ApfsKindUpdate is an updated record.
	// Reference: page 87
	// This record changes data that's part of an existing snapshot.
	ApfsKindUpdate JObjKinds = 2

	// ApfsKindDead is a record that's being deleted.
	// Reference: page 87
	// This value isn't valid as the kind of a file-system record on disk.
	// However, implementations of Apple File System can use it internally.
	ApfsKindDead JObjKinds = 3

	// ApfsKindUpdateRefcnt is an update to the reference count of a record.
	// Reference: page 88
	// This value isn't valid as the kind of a file-system record on disk.
	// However, implementations of Apple File System can use it internally.
	ApfsKindUpdateRefcnt JObjKinds = 4

	// ApfsKindInvalid is an invalid record kind.
	// Reference: page 88
	ApfsKindInvalid JObjKinds = 255
)

// JInodeFlags represents the flags used by inodes.
// Reference: page 88
type JInodeFlags uint64

const (
	// InodeIsApfsPrivate indicates the inode is used internally by an implementation of Apple File System.
	// Reference: page 89
	// Inodes with this flag set aren't considered part of the volume.
	// They can't be cloned, renamed, or deleted.
	InodeIsApfsPrivate JInodeFlags = 0x00000001

	// InodeMaintainDirStats indicates the inode tracks the size of all of its children.
	// Reference: page 89
	// This flag is only valid on a directory, and must also be set on the directory's subdirectories.
	InodeMaintainDirStats JInodeFlags = 0x00000002

	// InodeDirStatsOrigin indicates the inode has the INODE_MAINTAIN_DIR_STATS flag set explicitly,
	// not due to inheritance.
	// Reference: page 90
	// More than one directory in a hierarchy can have this flag set.
	InodeDirStatsOrigin JInodeFlags = 0x00000004

	// InodeProtClassExplicit indicates the inode's data protection class was set explicitly
	// when the inode was created.
	// Reference: page 90
	InodeProtClassExplicit JInodeFlags = 0x00000008

	// InodeWasCloned indicates the inode was created by cloning another inode.
	// Reference: page 90
	InodeWasCloned JInodeFlags = 0x00000010

	// InodeFlagUnused is reserved.
	// Reference: page 90
	// Leave this flag unset when you create a new inode,
	// and preserve its value when you modify an existing inode.
	InodeFlagUnused JInodeFlags = 0x00000020

	// InodeHasSecurityEa indicates the inode has an access control list.
	// Reference: page 90
	InodeHasSecurityEa JInodeFlags = 0x00000040

	// InodeBeingTruncated indicates the inode was truncated.
	// Reference: page 90
	InodeBeingTruncated JInodeFlags = 0x00000080

	// InodeHasFinderInfo indicates the inode has a Finder info extended field.
	// Reference: page 91
	InodeHasFinderInfo JInodeFlags = 0x00000100

	// InodeIsSparse indicates the inode has a sparse byte count extended field.
	// Reference: page 91
	InodeIsSparse JInodeFlags = 0x00000200

	// InodeWasEverCloned indicates the inode has been cloned at least once.
	// Reference: page 91
	// If this flag is set, the blocks on disk that store this inode might also be in use with another inode.
	InodeWasEverCloned JInodeFlags = 0x00000400

	// InodeActiveFileTrimmed indicates the inode is an overprovisioning file that has been trimmed.
	// Reference: page 91
	// This file type is used only on devices running iOS.
	InodeActiveFileTrimmed JInodeFlags = 0x00000800

	// InodePinnedToMain indicates the inode's file content is always on the main storage device.
	// Reference: page 91
	// This flag is only valid for Fusion systems. The main storage is a solid-state drive.
	InodePinnedToMain JInodeFlags = 0x00001000

	// InodePinnedToTier2 indicates the inode's file content is always on the secondary storage device.
	// Reference: page 92
	// This flag is only valid for Fusion systems. The secondary storage is a hard drive.
	InodePinnedToTier2 JInodeFlags = 0x00002000

	// InodeHasRsrcFork indicates the inode has a resource fork.
	// Reference: page 92
	// If this flag is set, INODE_NO_RSRC_FORK must not be set.
	// It's also valid for neither flag to be set, which implicitly indicates
	// that the inode doesn't have a resource fork.
	InodeHasRsrcFork JInodeFlags = 0x00004000

	// InodeNoRsrcFork indicates the inode doesn't have a resource fork.
	// Reference: page 92
	// If this flag is set, INODE_HAS_RSRC_FORK must not be set.
	// It's also valid for neither flag to be set, which implicitly indicates
	// that the inode doesn't have a resource fork.
	InodeNoRsrcFork JInodeFlags = 0x00008000

	// InodeAllocationSpilledover indicates the inode's file content has some space allocated outside
	// of the preferred storage tier for that file.
	// Reference: page 92
	// See also APFS_FS_SPILLEDOVER.
	InodeAllocationSpilledover JInodeFlags = 0x00010000

	// InodeFastPromote indicates this inode is scheduled for promotion from slow storage to fast storage.
	// Reference: page 92
	// The promotion between tiers will happen the first time this inode is read.
	InodeFastPromote JInodeFlags = 0x00020000

	// InodeHasUncompressedSize indicates this inode stores its uncompressed size in the inode.
	// Reference: page 92
	// The uncompressed size is stored in the uncompressed_size field of j_inode_val_t.
	// Prior to macOS 10.15 and iOS 13.1, this flag was ignored and Apple's implementation always treated
	// the uncompressed_size field as padding.
	InodeHasUncompressedSize JInodeFlags = 0x00040000

	// InodeIsPurgeable indicates this inode will be deleted at the next purge.
	// Reference: page 93
	// A purge is requested from user space by part of the operating system,
	// and the process of deleting purgeable files is the responsibility of the operating system.
	InodeIsPurgeable JInodeFlags = 0x00080000

	// InodeWantsToBePurgeable indicates this inode should become purgeable when
	// its link count drops to one.
	// Reference: page 93
	InodeWantsToBePurgeable JInodeFlags = 0x00100000

	// InodeIsSyncRoot indicates this inode is the root of a sync hierarchy for fileproviderd.
	// Reference: page 93
	// Don't add or remove this flag, but preserve the flag if it already exists.
	InodeIsSyncRoot JInodeFlags = 0x00200000

	// InodeSnapshotCowExemption indicates this inode is exempt from copy-on-write behavior
	// if the data is part of a snapshot.
	// Reference: page 93
	// Don't add or remove this flag, but preserve the flag if it already exists.
	// The number of files with this flag is tracked by the APFS_COW_EXEMPT_COUNT_NAME extended attribute.
	InodeSnapshotCowExemption JInodeFlags = 0x00400000

	// InodeInheritedInternalFlags is a bit mask of the flags that are inherited by the files
	// and subdirectories in a directory.
	// Reference: page 93
	InodeInheritedInternalFlags JInodeFlags = (InodeMaintainDirStats | InodeSnapshotCowExemption)

	// InodeClonedInternalFlags is a bit mask of the flags that are preserved when cloning.
	// Reference: page 93
	InodeClonedInternalFlags JInodeFlags = (InodeHasRsrcFork | InodeNoRsrcFork | InodeHasFinderInfo | InodeSnapshotCowExemption)
)

// ApfsValidInternalInodeFlags is a bit mask of all valid flags.
// Reference: page 94
const ApfsValidInternalInodeFlags JInodeFlags = JInodeFlags(InodeIsApfsPrivate |
	InodeMaintainDirStats |
	InodeDirStatsOrigin |
	InodeProtClassExplicit |
	InodeWasCloned |
	InodeHasSecurityEa |
	InodeBeingTruncated |
	InodeHasFinderInfo |
	InodeIsSparse |
	InodeWasEverCloned |
	InodeActiveFileTrimmed |
	InodePinnedToMain |
	InodePinnedToTier2 |
	InodeHasRsrcFork |
	InodeNoRsrcFork |
	InodeAllocationSpilledover |
	InodeFastPromote |
	InodeHasUncompressedSize |
	InodeIsPurgeable |
	InodeWantsToBePurgeable |
	InodeIsSyncRoot |
	InodeSnapshotCowExemption)

// ApfsInodePinnedMask is a bit mask of the flags that are related to pinning.
// Reference: page 94
const ApfsInodePinnedMask JInodeFlags = (InodePinnedToMain | InodePinnedToTier2)

// JXattrFlags represents the flags used in an extended attribute record to provide additional information.
// Reference: page 94
type JXattrFlags uint16

const (
	// XattrDataStream indicates the attribute data is stored in a data stream.
	// Reference: page 94
	// If this flag is set, XATTR_DATA_EMBEDDED must not be set.
	XattrDataStream JXattrFlags = 0x00000001

	// XattrDataEmbedded indicates the attribute data is stored directly in the record.
	// Reference: page 95
	// If this flag is set, the size of the value be smaller than XATTR_MAX_EMBEDDED_SIZE,
	// and XATTR_DATA_STREAM must not be set.
	XattrDataEmbedded JXattrFlags = 0x00000002

	// XattrFileSystemOwned indicates the extended attribute record is owned by the file system.
	// Reference: page 95
	// For example, this flag is used on symbolic links. The links have an extended attribute
	// whose name is SYMLINK_EA_NAME, and this flag is set on that attribute.
	XattrFileSystemOwned JXattrFlags = 0x00000004

	// XattrReserved8 is reserved.
	// Reference: page 95
	// Don't add this flag to an extended attribute record, but preserve the flag if it already exists.
	XattrReserved8 JXattrFlags = 0x00000008
)

// DirRecFlags represents the flags used by directory records.
// Reference: page 95
type DirRecFlags uint16

const (
	// DrecTypeMask is the bit mask used to access the type.
	// Reference: page 95
	// This bit mask is used with the flags field of j_drec_val_t.
	DrecTypeMask DirRecFlags = 0x000f

	// Reserved10 is reserved.
	// Reference: page 95
	// Don't set this flag. If you find a directory record with this flag set in production,
	// file a bug against the Apple File System implementation.
	Reserved10 DirRecFlags = 0x0010
)
