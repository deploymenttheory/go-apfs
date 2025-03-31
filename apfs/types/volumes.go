package types

// Volumes (pages 51-70)
// A volume contains a file system, the files and metadata that make up that file system,
// and various supporting data structures like an object map.

// ApfsSuperblockT is a volume superblock.
// Reference: page 51
type ApfsSuperblockT struct {
	// The object's header. (page 52)
	ApfsO ObjPhysT

	// A number that can be used to verify that you're reading an instance of apfs_superblock_t. (page 52)
	// The value of this field is always ApfsMagic.
	ApfsMagic uint32

	// The index of the object identifier for this volume's file system in the container's array of file systems. (page 53)
	// The container's array is stored in the nx_fs_oid field of nx_superblock_t.
	ApfsFsIndex uint32

	// A bit field of the optional features being used by this volume. (page 53)
	// For the values used in this bit field, see Optional Volume Feature Flags.
	ApfsFeatures uint64

	// A bit field of the read-only compatible features being used by this volume. (page 53)
	// For the values used in this bit field, see Read-Only Compatible Volume Feature Flags.
	ApfsReadonlyCompatibleFeatures uint64

	// A bit field of the backward-incompatible features being used by this volume. (page 53)
	// For the values used in this bit field, see Incompatible Volume Feature Flags.
	ApfsIncompatibleFeatures uint64

	// The time that this volume was last unmounted. (page 53)
	// This timestamp is represented as the number of nanoseconds since January 1, 1970 at 0:00 UTC,
	// disregarding leap seconds.
	ApfsUnmountTime uint64

	// The number of blocks that have been reserved for this volume to allocate. (page 54)
	ApfsFsReserveBlockCount uint64

	// The maximum number of blocks that this volume can allocate. (page 54)
	ApfsFsQuotaBlockCount uint64

	// The number of blocks currently allocated for this volume's file system. (page 54)
	ApfsFsAllocCount uint64

	// Information about the key used to encrypt metadata for this volume. (page 54)
	// On devices running macOS, the volume encryption key (VEK) is used to encrypt the metadata.
	ApfsMetaCrypto WrappedMetaCryptoStateT

	// The type of the root file-system tree. (page 54)
	// The value is typically OBJ_VIRTUAL | OBJECT_TYPE_BTREE, with a subtype of OBJECT_TYPE_FSTREE.
	ApfsRootTreeType uint32

	// The type of the extent-reference tree. (page 54)
	// The value is typically OBJ_PHYSICAL | OBJECT_TYPE_BTREE, with a subtype of OBJECT_TYPE_BLOCKREF.
	ApfsExtentreftreeType uint32

	// The type of the snapshot metadata tree. (page 54)
	// The value is typically OBJ_PHYSICAL | OBJECT_TYPE_BTREE, with a subtype of OBJECT_TYPE_BLOCKREF.
	ApfsSnapMetatreeType uint32

	// The physical object identifier of the volume's object map. (page 55)
	ApfsOmapOid OidT

	// The virtual object identifier of the root file-system tree. (page 55)
	ApfsRootTreeOid OidT

	// The physical object identifier of the extent-reference tree. (page 55)
	// When a snapshot is created, the current extent-reference tree is moved to the snapshot.
	// A new, empty, extent-reference tree is created and its object identifier becomes the new value of this field.
	ApfsExtentrefTreeOid OidT

	// The virtual object identifier of the snapshot metadata tree. (page 55)
	ApfsSnapMetaTreeOid OidT

	// The transaction identifier of a snapshot that the volume will revert to. (page 55)
	// When mounting a volume, if the value of this field nonzero, revert to the specified snapshot
	// by deleting all snapshots after the specified transaction identifier and deleting the current state,
	// and then setting this field to zero.
	ApfsRevertToXid XidT

	// The physical object identifier of a volume superblock that the volume will revert to. (page 55)
	// When mounting a volume, if the apfs_revert_to_xid field is nonzero, ignore the value of this field.
	// Otherwise, revert to the specified volume superblock.
	ApfsRevertToSblockOid OidT

	// The next identifier that will be assigned to a file-system object in this volume. (page 55)
	ApfsNextObjId uint64

	// The number of regular files in this volume. (page 56)
	ApfsNumFiles uint64

	// The number of directories in this volume. (page 56)
	ApfsNumDirectories uint64

	// The number of symbolic links in this volume. (page 56)
	ApfsNumSymlinks uint64

	// The number of other files in this volume. (page 56)
	// The value of this field includes all files that aren't included in the apfs_num_symlinks,
	// apfs_num_directories, or apfs_num_files fields.
	ApfsNumOtherFsobjects uint64

	// The number of snapshots in this volume. (page 56)
	ApfsNumSnapshots uint64

	// The total number of blocks that have been allocated by this volume. (page 56)
	// The value of this field increases when blocks are allocated, but isn't modified when they're freed.
	// If the volume doesn't contain any files, the value of this field matches apfs_total_blocks_freed.
	ApfsTotalBlocksAlloced uint64

	// The total number of blocks that have been freed by this volume. (page 56)
	// The value of this field isn't modified when blocks are allocated, but increases when they're freed.
	// If the volume doesn't contain any files, the value of this field matches apfs_total_blocks_alloced.
	ApfsTotalBlocksFreed uint64

	// The universally unique identifier for this volume. (page 57)
	ApfsVolUuid UUID

	// The time that this volume was last modified. (page 57)
	// This timestamp is represented as the number of nanoseconds since January 1, 1970 at 0:00 UTC,
	// disregarding leap seconds.
	ApfsLastModTime uint64

	// The volume's flags. (page 57)
	// For the values used in this bit field, see Volume Flags.
	ApfsFsFlags uint64

	// Information about the software that created this volume. (page 57)
	// This field is set only once, when the volume is created.
	ApfsFormattedBy ApfsModifiedByT

	// Information about the software that has modified this volume. (page 57)
	// The newest element in this array is stored at index zero.
	ApfsModifiedBy [ApfsMaxHist]ApfsModifiedByT

	// The name of the volume, represented as a null-terminated UTF-8 string. (page 57)
	// The APFS_INCOMPAT_NON_UTF8_FNAMES flag has no effect on this field's value.
	ApfsVolname [ApfsVolnameLen]byte

	// The next document identifier that will be assigned. (page 58)
	// A document's identifier is stored in the INO_EXT_TYPE_DOCUMENT_ID extended field of the inode.
	ApfsNextDocId uint32

	// The role of this volume within the container. (page 58)
	// For possible values, see Volume Roles.
	ApfsRole uint16

	// Reserved. (page 58)
	// Populate this field with zero when you create a new volume,
	// and preserve its value when you modify an existing volume.
	Reserved uint16

	// The transaction identifier of the snapshot to root from, or zero to root normally. (page 58)
	ApfsRootToXid XidT

	// The current state of encryption or decryption for a drive that's being encrypted or decrypted,
	// or zero if no encryption change is in progress. (page 58)
	ApfsErStateOid OidT

	// The largest object identifier used by this volume at the time
	// INODE_WAS_EVER_CLONED started storing valid information. (page 58)
	ApfsCloneinfoIdEpoch uint64

	// A transaction identifier used with apfs_cloneinfo_id_epoch. (page 59)
	// When unmounting a volume, the value of this field is set to the latest transaction identifier,
	// the same as the apfs_modified_by field.
	ApfsCloneinfoXid uint64

	// The virtual object identifier of the extended snapshot metadata object. (page 59)
	ApfsSnapMetaExtOid OidT

	// The volume group the volume belongs to. (page 59)
	// If the volume doesn't belong to a volume group, the value of this field is zero and the
	// APFS_FEATURE_VOLGRP_SYSTEM_INO_SPACE flag must not be set.
	// Otherwise, the APFS_FEATURE_VOLGRP_SYSTEM_INO_SPACE flag must be set and this field
	// must have a nonzero value.
	ApfsVolumeGroupId UUID

	// The virtual object identifier of the integrity metadata object. (page 59)
	// If the value of this field is nonzero, the APFS_INCOMPAT_SEALED_VOLUME flag must also be set.
	ApfsIntegrityMetaOid OidT

	// The virtual object identifier of the file extent tree. (page 59)
	// If the value of this field is nonzero, the APFS_INCOMPAT_SEALED_VOLUME flag must also be set.
	ApfsFextTreeOid OidT

	// The type of the file extent tree. (page 60)
	// The value is typically OBJ_PHYSICAL | OBJECT_TYPE_BTREE, with a subtype of OBJECT_TYPE_FEXT_TREE.
	ApfsFextTreeType uint32

	// Reserved. (page 60)
	ReservedType uint32

	// Reserved. (page 60)
	ReservedOid OidT
}

// ApfsModifiedByT contains information about a program that modified the volume.
// Reference: page 60
type ApfsModifiedByT struct {
	// A string that identifies the program and its version. (page 61)
	Id [ApfsModifiedNamelen]byte

	// The time that the program last modified this volume. (page 61)
	// This timestamp is represented as the number of nanoseconds since January 1, 1970 at 0:00 UTC,
	// disregarding leap seconds.
	Timestamp uint64

	// The last transaction identifier that's part of this program's modifications. (page 61)
	LastXid XidT
}

// ApfsMagic is the value of the apfs_magic field.
// This magic number was chosen because in hex dumps it appears as "APSB",
// which is an abbreviated form of APFS superblock.
// Reference: page 60
const ApfsMagic uint32 = 'B' | 'S'<<8 | 'P'<<16 | 'A'<<24 // 'BSPA'

// ApfsMaxHist is the number of entries stored in the apfs_modified_by field.
// Reference: page 60
const ApfsMaxHist = 8

// ApfsVolnameLen is the maximum length of the volume name stored in the apfs_volname field.
// Reference: page 60
const ApfsVolnameLen = 256

// ApfsModifiedNamelen is the length of the id field in ApfsModifiedByT.
// Reference: page 61
const ApfsModifiedNamelen = 32

// Volume Flags (pages 61-63)

// ApfsFsUnencrypted indicates the volume isn't encrypted.
// Reference: page 62
const ApfsFsUnencrypted uint64 = 0x00000001

// ApfsFsReserved2 is reserved.
// Reference: page 62
const ApfsFsReserved2 uint64 = 0x00000002

// ApfsFsReserved4 is reserved.
// Reference: page 62
const ApfsFsReserved4 uint64 = 0x00000004

// ApfsFsOnekey indicates files on the volume are all encrypted using the volume encryption key (VEK).
// Reference: page 62
const ApfsFsOnekey uint64 = 0x00000008

// ApfsFsSpilledover indicates the volume has run out of allocated space on the solid-state drive.
// Reference: page 62
const ApfsFsSpilledover uint64 = 0x00000010

// ApfsFsRunSpilloverCleaner indicates the volume has spilled over and the spillover cleaner must be run.
// Reference: page 62
const ApfsFsRunSpilloverCleaner uint64 = 0x00000020

// ApfsFsAlwaysCheckExtentref indicates the volume's extent reference tree is always consulted
// when deciding whether to overwrite an extent.
// Reference: page 62
const ApfsFsAlwaysCheckExtentref uint64 = 0x00000040

// ApfsFsReserved80 is reserved.
// Reference: page 63
const ApfsFsReserved80 uint64 = 0x00000080

// ApfsFsReserved100 is reserved.
// Reference: page 63
const ApfsFsReserved100 uint64 = 0x00000100

// ApfsFsFlagsValidMask is a bit mask of all volume flags.
// Reference: page 63
const ApfsFsFlagsValidMask uint64 = ApfsFsUnencrypted |
	ApfsFsReserved2 |
	ApfsFsReserved4 |
	ApfsFsOnekey |
	ApfsFsSpilledover |
	ApfsFsRunSpilloverCleaner |
	ApfsFsAlwaysCheckExtentref |
	ApfsFsReserved80 |
	ApfsFsReserved100

// ApfsFsCryptoflags is a bit mask of all encryption-related volume flags.
// Reference: page 63
const ApfsFsCryptoflags uint64 = ApfsFsUnencrypted |
	ApfsFsReserved2 |
	ApfsFsOnekey

// Volume Roles (pages 63-66)

// ApfsVolRoleNone indicates the volume has no defined role.
// Reference: page 64
const ApfsVolRoleNone uint16 = 0x0000

// ApfsVolRoleSystem indicates the volume contains a root directory for the system.
// Reference: page 64
const ApfsVolRoleSystem uint16 = 0x0001

// ApfsVolRoleUser indicates the volume contains users' home directories.
// Reference: page 64
const ApfsVolRoleUser uint16 = 0x0002

// ApfsVolRoleRecovery indicates the volume contains a recovery system.
// Reference: page 64
const ApfsVolRoleRecovery uint16 = 0x0004

// ApfsVolRoleVm indicates the volume is used as swap space for virtual memory.
// Reference: page 64
const ApfsVolRoleVm uint16 = 0x0008

// ApfsVolRolePreboot indicates the volume contains files needed to boot from an encrypted volume.
// Reference: page 65
const ApfsVolRolePreboot uint16 = 0x0010

// ApfsVolRoleInstaller indicates the volume is used by the OS installer.
// Reference: page 65
const ApfsVolRoleInstaller uint16 = 0x0020

// ApfsVolumeEnumShift is the bit shift used to separate the old and new enumeration cases.
// Reference: page 66
const ApfsVolumeEnumShift uint16 = 6

// ApfsVolRoleData indicates the volume contains mutable data.
// Reference: page 65
const ApfsVolRoleData uint16 = 1 << ApfsVolumeEnumShift

// ApfsVolRoleBaseband indicates the volume is used by the radio firmware.
// Reference: page 65
const ApfsVolRoleBaseband uint16 = 2 << ApfsVolumeEnumShift

// ApfsVolRoleUpdate indicates the volume is used by the software update mechanism.
// Reference: page 65
const ApfsVolRoleUpdate uint16 = 3 << ApfsVolumeEnumShift

// ApfsVolRoleXart indicates the volume is used to manage OS access to secure user data.
// Reference: page 65
const ApfsVolRoleXart uint16 = 4 << ApfsVolumeEnumShift

// ApfsVolRoleHardware indicates the volume is used for firmware data.
// Reference: page 65
const ApfsVolRoleHardware uint16 = 5 << ApfsVolumeEnumShift

// ApfsVolRoleBackup indicates the volume is used by Time Machine to store backups.
// Reference: page 66
const ApfsVolRoleBackup uint16 = 6 << ApfsVolumeEnumShift

// ApfsVolRoleReserved7 is reserved.
// Reference: page 66
const ApfsVolRoleReserved7 uint16 = 7 << ApfsVolumeEnumShift

// ApfsVolRoleReserved8 is reserved.
// Reference: page 66
const ApfsVolRoleReserved8 uint16 = 8 << ApfsVolumeEnumShift

// ApfsVolRoleEnterprise indicates this volume is used to store enterprise-managed data.
// Reference: page 66
const ApfsVolRoleEnterprise uint16 = 9 << ApfsVolumeEnumShift

// ApfsVolRoleReserved10 is reserved.
// Reference: page 66
const ApfsVolRoleReserved10 uint16 = 10 << ApfsVolumeEnumShift

// ApfsVolRolePrelogin indicates this volume is used to store system data used before login.
// Reference: page 66
const ApfsVolRolePrelogin uint16 = 11 << ApfsVolumeEnumShift

// Optional Volume Feature Flags (pages 67-68)

// ApfsFeatureDefragPrerelease is reserved.
// Reference: page 67
const ApfsFeatureDefragPrerelease uint64 = 0x00000001

// ApfsFeatureHardlinkMapRecords indicates the volume has hardlink map records.
// Reference: page 67
const ApfsFeatureHardlinkMapRecords uint64 = 0x00000002

// ApfsFeatureDefrag indicates the volume supports defragmentation.
// Reference: page 67
const ApfsFeatureDefrag uint64 = 0x00000004

// ApfsFeatureStrictatime indicates this volume updates file access times every time the file is read.
// Reference: page 67
const ApfsFeatureStrictatime uint64 = 0x00000008

// ApfsFeatureVolgrpSystemInoSpace indicates this volume supports mounting a system
// and data volume as a single user-visible volume.
// Reference: page 68
const ApfsFeatureVolgrpSystemInoSpace uint64 = 0x00000010

// ApfsSupportedFeaturesMask is a bit mask of all the optional volume features.
// Reference: page 68
const ApfsSupportedFeaturesMask uint64 = ApfsFeatureDefrag |
	ApfsFeatureDefragPrerelease |
	ApfsFeatureHardlinkMapRecords |
	ApfsFeatureStrictatime |
	ApfsFeatureVolgrpSystemInoSpace

// Read-Only Compatible Volume Feature Flags (page 68)

// ApfsSupportedRocompatMask is a bit mask of all read-only compatible volume features.
// Reference: page 68
const ApfsSupportedRocompatMask uint64 = 0x0

// Incompatible Volume Feature Flags (pages 68-70)

// ApfsIncompatCaseInsensitive indicates filenames on this volume are case insensitive.
// Reference: page 69
const ApfsIncompatCaseInsensitive uint64 = 0x00000001

// ApfsIncompatDatalessSnaps indicates at least one snapshot with no data exists for this volume.
// Reference: page 69
const ApfsIncompatDatalessSnaps uint64 = 0x00000002

// ApfsIncompatEncRolled indicates this volume's encryption has changed keys at least once.
// Reference: page 69
const ApfsIncompatEncRolled uint64 = 0x00000004

// ApfsIncompatNormalizationInsensitive indicates filenames on this volume are normalization insensitive.
// Reference: page 69
const ApfsIncompatNormalizationInsensitive uint64 = 0x00000008

// ApfsIncompatIncompleteRestore indicates this volume is being restored,
// or a restore operation to this volume was uncleanly aborted.
// Reference: page 69
const ApfsIncompatIncompleteRestore uint64 = 0x00000010

// ApfsIncompatSealedVolume indicates this volume can't be modified.
// Reference: page 69
const ApfsIncompatSealedVolume uint64 = 0x00000020

// ApfsIncompatReserved40 is reserved.
// Reference: page 70
const ApfsIncompatReserved40 uint64 = 0x00000040

// ApfsSupportedIncompatMask is a bit mask of all the backward-incompatible volume features.
// Reference: page 70
const ApfsSupportedIncompatMask uint64 = ApfsIncompatCaseInsensitive |
	ApfsIncompatDatalessSnaps |
	ApfsIncompatEncRolled |
	ApfsIncompatNormalizationInsensitive |
	ApfsIncompatIncompleteRestore |
	ApfsIncompatSealedVolume |
	ApfsIncompatReserved40

// Inode Numbers (page 96)

// InvalidInoNum is an invalid inode number.
// Reference: page 96
const InvalidInoNum uint64 = 0

// RootDirParent is the inode number for the root directory's parent.
// Reference: page 96
// This is a sentinel value; there's no inode on disk with this inode number.
const RootDirParent uint64 = 1

// RootDirInoNum is the inode number for the root directory of the volume.
// Reference: page 96
const RootDirInoNum uint64 = 2

// PrivDirInoNum is the inode number for the private directory.
// Reference: page 96
// The private directory's filename is "private-dir". When creating a new volume,
// you must create a directory with this name and inode number.
const PrivDirInoNum uint64 = 3

// SnapDirInoNum is the inode number for the directory where snapshot metadata is stored.
// Reference: page 97
// Snapshot inodes are stored in the snapshot metedata tree.
const SnapDirInoNum uint64 = 6

// PurgeableDirInoNum is the inode number used for storing references to purgeable files.
// Reference: page 97
// This inode number and the directory records that use it are reserved.
// Other implementations of the Apple File System must not modify them.
// There isn't an actual directory with this inode number.
const PurgeableDirInoNum uint64 = 7

// MinUserInoNum is the smallest inode number available for user content.
// Reference: page 97
// All inode numbers less than this value are reserved.
const MinUserInoNum uint64 = 16

// UnifiedIdSpaceMark is the smallest inode number used by the system volume in a volume group.
// Reference: page 97
const UnifiedIdSpaceMark uint64 = 0x0800000000000000

// Extended Attributes Constants (page 97)

// XattrMaxEmbeddedSize is the largest size, in bytes, of an extended attribute
// whose value is stored directly in the record.
// Reference: page 97
const XattrMaxEmbeddedSize uint32 = 3804

// SymlinkEaName is the name of an extended attribute for a symbolic link
// whose value is the target file on the data volume.
// Reference: page 98
const SymlinkEaName string = "com.apple.fs.symlink"

// FirmlinkEaName is the name of an extended attribute for a firm link
// whose value is the target file.
// Reference: page 98
const FirmlinkEaName string = "com.apple.fs.firmlink"

// ApfsCowExemptCountName is the number of files on the volume that don't use copy on write.
// Reference: page 98
const ApfsCowExemptCountName string = "com.apple.fs.cow-exempt-file-count"

// File-System Object Constants (page 98)

// OwningObjIdInvalid indicates an invalid object identifier.
// Reference: page 98
const OwningObjIdInvalid uint64 = ^uint64(0) // ~0ULL

// OwningObjIdUnknown indicates an unknown object identifier.
// Reference: page 98
const OwningObjIdUnknown uint64 = ^uint64(1) // ~1ULL

// JobjMaxKeySize is the maximum size of a key in a file system object.
// Reference: page 98
const JobjMaxKeySize uint32 = 832

// JobjMaxValueSize is the maximum size of a value in a file system object.
// Reference: page 98
const JobjMaxValueSize uint32 = 3808

// MinDocId is the smallest document identifier available for user content.
// Reference: page 98
// All document identifiers less than this value are reserved.
const MinDocId uint32 = 3

// File Modes (pages 98-100)

// ModeT represents a file mode.
// Reference: page 99
type ModeT uint16

// SIfmt is the bit mask used to access the file type.
// Reference: page 99
const SIfmt ModeT = 0170000

// SIfifo indicates a named pipe.
// Reference: page 99
const SIfifo ModeT = 0010000

// SIfchr indicates a character-special file.
// Reference: page 99
const SIfchr ModeT = 0020000

// SIfdir indicates a directory.
// Reference: page 99
const SIfdir ModeT = 0040000

// SIfblk indicates a block-special file.
// Reference: page 99
const SIfblk ModeT = 0060000

// SIfreg indicates a regular file.
// Reference: page 100
const SIfreg ModeT = 0100000

// SIflnk indicates a symbolic link.
// Reference: page 100
const SIflnk ModeT = 0120000

// SIfsock indicates a socket.
// Reference: page 100
const SIfsock ModeT = 0140000

// SIfwht indicates a whiteout.
// Reference: page 100
const SIfwht ModeT = 0160000

// Directory Entry File Types (pages 100-101)

// DtUnknown indicates an unknown directory entry.
// Reference: page 100
const DtUnknown uint16 = 0

// DtFifo indicates a named pipe.
// Reference: page 100
const DtFifo uint16 = 1

// DtChr indicates a character-special file.
// Reference: page 101
const DtChr uint16 = 2

// DtDir indicates a directory.
// Reference: page 101
const DtDir uint16 = 4

// DtBlk indicates a block-special file.
// Reference: page 101
const DtBlk uint16 = 6

// DtReg indicates a regular file.
// Reference: page 101
const DtReg uint16 = 8

// DtLnk indicates a symbolic link.
// Reference: page 101
const DtLnk uint16 = 10

// DtSock indicates a socket.
// Reference: page 101
const DtSock uint16 = 12

// DtWht indicates a whiteout.
// Reference: page 101
const DtWht uint16 = 14

// Unix permission bits
// These are Go-specific constants for file mode permissions and are not part of the APFS specification
const (
	// User permissions
	SUread  ModeT = 0000400 // Owner has read permission
	SUwrite ModeT = 0000200 // Owner has write permission
	SUexec  ModeT = 0000100 // Owner has execute permission

	// Group permissions
	SGread  ModeT = 0000040 // Group has read permission
	SGwrite ModeT = 0000020 // Group has write permission
	SGexec  ModeT = 0000010 // Group has execute permission

	// Others permissions
	SOread  ModeT = 0000004 // Others have read permission
	SOwrite ModeT = 0000002 // Others have write permission
	SOexec  ModeT = 0000001 // Others have execute permission
)
