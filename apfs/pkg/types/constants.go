package types

// APFS magic constants
const (
	NXMagic   uint32 = 0x4253584E // 'NXSB'
	APFSMagic uint32 = 0x42535041 // 'APSB'
	JSMagic   uint32 = 0x5244534A // 'JSDR' (JumpStart Driver)
	ERMagic   uint32 = 0x46414C42 // 'FLAB' (Encryption Rolling)

	MaxCksumSize uint32 = 8 // Maximum checksum size
)

// -------------------------------------------------------------------
// Limits and structural bounds
// Reference: Apple File System Reference — nx_superblock_t (page ~32)
// -------------------------------------------------------------------

// Block size constants and structural limits
const (
	MinBlockSize     uint32 = 4096    // Minimum supported block size
	DefaultBlockSize uint32 = 4096    // Default block size used in most containers
	MaxBlockSize     uint32 = 65536   // Maximum supported block size
	MinContainerSize uint64 = 1048576 // Minimum container size in bytes (1 MiB)

	NXMaxFileSystems           = 100  // Maximum number of file systems (volumes) in a container
	NXMaxXPDescBlocks          = 32   // Maximum number of checkpoint descriptor blocks
	NXMaxXPDataBlocks          = 8192 // Maximum number of checkpoint data blocks
	NXNumCounters              = 32   // Number of counters in nx_superblock_t
	NXEphemeralInfoCount       = 4    // Number of ephemeral_info[] entries in nx_superblock_t
	SpacemanFreeQueueCount     = 3    // Number of spaceman free queues (internal, main, tier2)
	MaxSnapshotMetadataRecords = 1024 // Max entries in the snapshot metadata tree
	APFSMaxVolumeRoles         = 32   // Number of roles representable in a volume role bitmask
)

// Object identifier constants
const (
	OIDNXSuperblock  uint64 = 1
	OIDInvalid       uint64 = 0
	OIDReservedCount uint32 = 1024
)

// Object type masks
const (
	ObjectTypeMask         uint32 = 0x0000FFFF
	ObjectTypeFlagsMask    uint32 = 0xFFFF0000
	ObjStorageTypeMask     uint32 = 0xC0000000
	ObjectTypeFlagsDefMask uint32 = 0xF8000000
)

// Object storage type flags
const (
	ObjVirtual       uint32 = 0x00000000
	ObjEphemeral     uint32 = 0x80000000
	ObjPhysical      uint32 = 0x40000000
	ObjNoheader      uint32 = 0x20000000
	ObjEncrypted     uint32 = 0x10000000
	ObjNonpersistent uint32 = 0x08000000
)

// Object types
const (
	ObjectTypeNXSuperblock      uint32 = 0x00000001
	ObjectTypeBtree             uint32 = 0x00000002
	ObjectTypeBtreeNode         uint32 = 0x00000003
	ObjectTypeSpaceman          uint32 = 0x00000005
	ObjectTypeSpacemanCAB       uint32 = 0x00000006
	ObjectTypeSpacemanCIB       uint32 = 0x00000007
	ObjectTypeSpacemanBitmap    uint32 = 0x00000008
	ObjectTypeSpacemanFreeQueue uint32 = 0x00000009
	ObjectTypeExtentListTree    uint32 = 0x0000000A
	ObjectTypeOMap              uint32 = 0x0000000B
	ObjectTypeCheckpointMap     uint32 = 0x0000000C
	ObjectTypeFS                uint32 = 0x0000000D
	ObjectTypeFSTree            uint32 = 0x0000000E
	ObjectTypeBlockrefTree      uint32 = 0x0000000F
	ObjectTypeSnapMetaTree      uint32 = 0x00000010
	ObjectTypeNXReaper          uint32 = 0x00000011
	ObjectTypeNXReapList        uint32 = 0x00000012
	ObjectTypeOMapSnapshot      uint32 = 0x00000013
	ObjectTypeEFIJumpstart      uint32 = 0x00000014
	ObjectTypeFusionMiddleTree  uint32 = 0x00000015
	ObjectTypeNXFusionWBC       uint32 = 0x00000016
	ObjectTypeNXFusionWBCList   uint32 = 0x00000017
	ObjectTypeERState           uint32 = 0x00000018
	ObjectTypeGBitmap           uint32 = 0x00000019
	ObjectTypeGBitmapTree       uint32 = 0x0000001A
	ObjectTypeGBitmapBlock      uint32 = 0x0000001B
	ObjectTypeERRecoveryBlock   uint32 = 0x0000001C
	ObjectTypeSnapMetaExt       uint32 = 0x0000001D
	ObjectTypeIntegrityMeta     uint32 = 0x0000001E
	ObjectTypeFextTree          uint32 = 0x0000001F
	ObjectTypeReserved20        uint32 = 0x00000020
	ObjectTypeInvalid           uint32 = 0x00000000
	ObjectTypeTest              uint32 = 0x000000FF
)

// Special object types - 4 character codes instead of integers
const (
	ObjectTypeContainerKeybag uint32 = 0x6B657973 // 'keys'
	ObjectTypeVolumeKeybag    uint32 = 0x72656373 // 'recs'
	ObjectTypeMediaKeybag     uint32 = 0x6D6B6579 // 'mkey'
)

// File-system record types (j_obj_types)
const (
	APFSTypeAny          uint8 = 0
	APFSTypeSnapMetadata uint8 = 1
	APFSTypeExtent       uint8 = 2
	APFSTypeInode        uint8 = 3
	APFSTypeXattr        uint8 = 4
	APFSTypeSiblingLink  uint8 = 5
	APFSTypeDStreamID    uint8 = 6
	APFSTypeCryptoState  uint8 = 7
	APFSTypeFileExtent   uint8 = 8
	APFSTypeDirRecord    uint8 = 9
	APFSTypeDirStats     uint8 = 10
	APFSTypeSnapName     uint8 = 11
	APFSTypeSiblingMap   uint8 = 12
	APFSTypeFileInfo     uint8 = 13
	APFSTypeMaxValid     uint8 = 13
	APFSTypeMax          uint8 = 15
	APFSTypeInvalid      uint8 = 15
)

// File-system record kinds (j_obj_kinds)
const (
	APFSKindAny            uint8 = 0
	APFSKindNew            uint8 = 1
	APFSKindUpdate         uint8 = 2
	APFSKindDead           uint8 = 3
	APFSKindUpdateRefCount uint8 = 4
	APFSKindInvalid        uint8 = 255
)

// Container feature flags
const (
	// NX flags (nx_flags field of nx_superblock_t)
	NXReserved1 uint64 = 0x00000001
	NXReserved2 uint64 = 0x00000002
	NXCryptoSW  uint64 = 0x00000004

	// Optional features (nx_features field of nx_superblock_t)
	NXFeatureDefrag         uint64 = 0x0000000000000001
	NXFeatureLCFD           uint64 = 0x0000000000000002
	NXSupportedFeaturesMask uint64 = (NXFeatureDefrag | NXFeatureLCFD)

	// Read-only compatible features (nx_readonly_compatible_features field of nx_superblock_t)
	NXSupportedROCompatMask uint64 = 0x0

	// Incompatible features (nx_incompatible_features field of nx_superblock_t)
	NXIncompatVersion1      uint64 = 0x0000000000000001
	NXIncompatVersion2      uint64 = 0x0000000000000002
	NXIncompatFusion        uint64 = 0x0000000000000100
	NXSupportedIncompatMask uint64 = (NXIncompatVersion2 | NXIncompatFusion)
)

// Volume feature flags
const (
	// Volume flags (apfs_fs_flags field of apfs_superblock_t)
	APFSFSUnencrypted          uint64 = 0x00000001
	APFSFSReserved2            uint64 = 0x00000002
	APFSFSReserved4            uint64 = 0x00000004
	APFSFSOnekey               uint64 = 0x00000008
	APFSFSSpilledover          uint64 = 0x00000010
	APFSFSRunSpilloverCleaner  uint64 = 0x00000020
	APFSFSAlwaysCheckExtentref uint64 = 0x00000040
	APFSFSReserved80           uint64 = 0x00000080
	APFSFSReserved100          uint64 = 0x00000100
	APFSFSFlagsValidMask       uint64 = 0x000001FF
	APFSFSCryptoFlags          uint64 = 0x0000000B // Unencrypted | Reserved2 | Onekey

	// Optional features (apfs_features field of apfs_superblock_t)
	APFSFeatureDefragPrerelease     uint64 = 0x00000001
	APFSFeatureHardlinkMapRecords   uint64 = 0x00000002
	APFSFeatureDefrag               uint64 = 0x00000004
	APFSFeatureStrictAtime          uint64 = 0x00000008
	APFSFeatureVolgrpSystemInoSpace uint64 = 0x00000010
	APFSSupportedFeaturesMask       uint64 = 0x0000001F

	// Read-only compatible features (apfs_readonly_compatible_features field of apfs_superblock_t)
	APFSSupportedROCompatMask uint64 = 0x0

	// Incompatible features (apfs_incompatible_features field of apfs_superblock_t)
	APFSIncompatCaseInsensitive          uint64 = 0x00000001
	APFSIncompatDatalessSnaps            uint64 = 0x00000002
	APFSIncompatEncRolled                uint64 = 0x00000004
	APFSIncompatNormalizationInsensitive uint64 = 0x00000008
	APFSIncompatIncompleteRestore        uint64 = 0x00000010
	APFSIncompatSealedVolume             uint64 = 0x00000020
	APFSIncompatReserved40               uint64 = 0x00000040
	APFSSupportedIncompatMask            uint64 = 0x0000007F
)

// Volume roles
const (
	APFSVolRoleNone      uint16 = 0x0000
	APFSVolRoleSystem    uint16 = 0x0001
	APFSVolRoleUser      uint16 = 0x0002
	APFSVolRoleRecovery  uint16 = 0x0004
	APFSVolRoleVM        uint16 = 0x0008
	APFSVolRolePreboot   uint16 = 0x0010
	APFSVolRoleInstaller uint16 = 0x0020

	APFSVolumeEnumShift uint16 = 6

	APFSVolRoleData       uint16 = 1 << APFSVolumeEnumShift
	APFSVolRoleBaseband   uint16 = 2 << APFSVolumeEnumShift
	APFSVolRoleUpdate     uint16 = 3 << APFSVolumeEnumShift
	APFSVolRoleXart       uint16 = 4 << APFSVolumeEnumShift
	APFSVolRoleHardware   uint16 = 5 << APFSVolumeEnumShift
	APFSVolRoleBackup     uint16 = 6 << APFSVolumeEnumShift
	APFSVolRoleReserved7  uint16 = 7 << APFSVolumeEnumShift
	APFSVolRoleReserved8  uint16 = 8 << APFSVolumeEnumShift
	APFSVolRoleEnterprise uint16 = 9 << APFSVolumeEnumShift
	APFSVolRoleReserved10 uint16 = 10 << APFSVolumeEnumShift
	APFSVolRolePrelogin   uint16 = 11 << APFSVolumeEnumShift
)

// Checkpoint map flags
const (
	CheckpointMapLast uint32 = 0x00000001
)

// OMap flags and constants
const (
	// OMap flags (om_flags field of omap_phys_t)
	OMapManuallyManaged  uint32 = 0x00000001
	OMapEncrypting       uint32 = 0x00000002
	OMapDecrypting       uint32 = 0x00000004
	OMapKeyrolling       uint32 = 0x00000008
	OMapCryptoGeneration uint32 = 0x00000010
	OMapValidFlags       uint32 = 0x0000001F

	// OMap value flags (ov_flags field of omap_val_t)
	OMapValDeleted          uint32 = 0x00000001
	OMapValSaved            uint32 = 0x00000002
	OMapValEncrypted        uint32 = 0x00000004
	OMapValNoheader         uint32 = 0x00000008
	OMapValCryptoGeneration uint32 = 0x00000010

	// Snapshot flags (oms_flags field of omap_snapshot_t)
	OMapSnapshotDeleted  uint32 = 0x00000001
	OMapSnapshotReverted uint32 = 0x00000002

	// OMap reaper phases
	OMapReapPhaseMapTree      uint32 = 1
	OMapReapPhaseSnapshotTree uint32 = 2

	// OMap constants
	OMapMaxSnapCount uint32 = 0xFFFFFFFF
)

// B-tree flags and constants
const (
	// B-tree flags (bt_flags field of btree_info_fixed_t)
	BtreeUint64Keys       uint32 = 0x00000001
	BtreeSequentialInsert uint32 = 0x00000002
	BtreeAllowGhosts      uint32 = 0x00000004
	BtreeEphemeral        uint32 = 0x00000008
	BtreePhysical         uint32 = 0x00000010
	BtreeNonpersistent    uint32 = 0x00000020
	BtreeKVNonaligned     uint32 = 0x00000040
	BtreeHashed           uint32 = 0x00000080
	BtreeNoheader         uint32 = 0x00000100

	// B-tree node flags (btn_flags field of btree_node_phys_t)
	BtnodeRoot           uint16 = 0x0001
	BtnodeLeaf           uint16 = 0x0002
	BtnodeFixedKVSize    uint16 = 0x0004
	BtnodeHashed         uint16 = 0x0008
	BtnodeNoheader       uint16 = 0x0010
	BtnodeCheckKoffInval uint16 = 0x8000

	// B-tree constants
	BtreeNodeSizeDefault   uint32 = 4096
	BtreeNodeMinEntryCount uint32 = 4
	BtreeTocEntryIncrement uint32 = 8
	BtreeTocEntryMaxUnused uint32 = 16
	BtreeNodeHashSizeMax   uint32 = 64
	BtoffInvalid           uint16 = 0xFFFF
)

// Inode flags (internal_flags field of j_inode_val_t)
const (
	InodeIsAPFSPrivate         uint64 = 0x00000001
	InodeMaintainDirStats      uint64 = 0x00000002
	InodeDirStatsOrigin        uint64 = 0x00000004
	InodeProtClassExplicit     uint64 = 0x00000008
	InodeWasCloned             uint64 = 0x00000010
	InodeFlagUnused            uint64 = 0x00000020
	InodeHasSecurityEA         uint64 = 0x00000040
	InodeBeingTruncated        uint64 = 0x00000080
	InodeHasFinderInfo         uint64 = 0x00000100
	InodeIsSparse              uint64 = 0x00000200
	InodeWasEverCloned         uint64 = 0x00000400
	InodeActiveFileTrimmed     uint64 = 0x00000800
	InodePinnedToMain          uint64 = 0x00001000
	InodePinnedToTier2         uint64 = 0x00002000
	InodeHasRsrcFork           uint64 = 0x00004000
	InodeNoRsrcFork            uint64 = 0x00008000
	InodeAllocationSpilledOver uint64 = 0x00010000
	InodeFastPromote           uint64 = 0x00020000
	InodeHasUncompressedSize   uint64 = 0x00040000
	InodeIsPurgeable           uint64 = 0x00080000
	InodeWantsToBePurgeable    uint64 = 0x00100000
	InodeIsSyncRoot            uint64 = 0x00200000
	InodeSnapshotCowExemption  uint64 = 0x00400000

	InodeInheritedInternalFlags uint64 = (InodeMaintainDirStats | InodeSnapshotCowExemption)
	InodeClonedInternalFlags    uint64 = (InodeHasRsrcFork | InodeNoRsrcFork | InodeHasFinderInfo | InodeSnapshotCowExemption)
	APFSValidInternalInodeFlags uint64 = 0x007FFFFF
	APFSInodePinnedMask         uint64 = (InodePinnedToMain | InodePinnedToTier2)
)

// Directory record flags (flags field of j_drec_val_t)
const (
	DrecTypeMask   uint16 = 0x000F
	DrecReserved10 uint16 = 0x0010
)

// Extended attribute flags (flags field of j_xattr_val_t)
const (
	XattrDataStream      uint16 = 0x00000001
	XattrDataEmbedded    uint16 = 0x00000002
	XattrFileSystemOwned uint16 = 0x00000004
	XattrReserved8       uint16 = 0x00000008
)

// Extended field types
const (
	// Directory extended field types
	DrecExtTypeSiblingID uint8 = 1

	// Inode extended field types
	InoExtTypeSnapXID        uint8 = 1
	InoExtTypeDeltaTreeOID   uint8 = 2
	InoExtTypeDocumentID     uint8 = 3
	InoExtTypeName           uint8 = 4
	InoExtTypePrevFsize      uint8 = 5
	InoExtTypeReserved6      uint8 = 6
	InoExtTypeFinderInfo     uint8 = 7
	InoExtTypeDstream        uint8 = 8
	InoExtTypeReserved9      uint8 = 9
	InoExtTypeDirStatsKey    uint8 = 10
	InoExtTypeFSUUID         uint8 = 11
	InoExtTypeReserved12     uint8 = 12
	InoExtTypeSparseBytes    uint8 = 13
	InoExtTypeRdev           uint8 = 14
	InoExtTypePurgeableFlags uint8 = 15
	InoExtTypeOrigSyncRootID uint8 = 16
)

// Extended field flags
const (
	XFDataDependent   uint8 = 0x0001
	XFDoNotCopy       uint8 = 0x0002
	XFReserved4       uint8 = 0x0004
	XFChildrenInherit uint8 = 0x0008
	XFUserField       uint8 = 0x0010
	XFSystemField     uint8 = 0x0020
	XFReserved40      uint8 = 0x0040
	XFReserved80      uint8 = 0x0080
)

// Physical extent constants
const (
	PextLenMask   uint64 = 0x0FFFFFFFFFFFFFFF
	PextKindMask  uint64 = 0xF000000000000000
	PextKindShift uint8  = 60
)

// File extent constants
const (
	JFileExtentLenMask   uint64 = 0x00FFFFFFFFFFFFFF
	JFileExtentFlagMask  uint64 = 0xFF00000000000000
	JFileExtentFlagShift uint8  = 56
	FextCryptoIDIsTweak  uint8  = 0x01
)

// File info constants
const (
	JFileInfoLBAMask   uint64 = 0x00FFFFFFFFFFFFFF
	JFileInfoTypeMask  uint64 = 0xFF00000000000000
	JFileInfoTypeShift uint8  = 56
)

// Reserved inode numbers
const (
	InvalidInoNum      uint64 = 0
	RootDirParent      uint64 = 1
	RootDirInoNum      uint64 = 2
	PrivDirInoNum      uint64 = 3
	SnapDirInoNum      uint64 = 6
	PurgeableDirInoNum uint64 = 7
	MinUserInoNum      uint64 = 16
	UnifiedIDSpaceMark uint64 = 0x0800000000000000
)

// Extended attribute constants
const (
	XattrMaxEmbeddedSize   uint32 = 3804
	SymlinkEAName          string = "com.apple.fs.symlink"
	FirmlinkEAName         string = "com.apple.fs.firmlink"
	ApfsCowExemptCountName string = "com.apple.fs.cow-exempt-file-count"
)

// File modes (equivalent to stat.h values)
const (
	SIFmt   uint16 = 0170000
	SIFifo  uint16 = 0010000
	SIFchr  uint16 = 0020000
	SIFdir  uint16 = 0040000
	SIFblk  uint16 = 0060000
	SIFreg  uint16 = 0100000
	SIFlnk  uint16 = 0120000
	SIFsock uint16 = 0140000
	SIFwht  uint16 = 0160000
)

// Directory entry file types
const (
	DTUnknown uint8 = 0
	DTFifo    uint8 = 1
	DTChr     uint8 = 2
	DTDir     uint8 = 4
	DTBlk     uint8 = 6
	DTReg     uint8 = 8
	DTLnk     uint8 = 10
	DTSock    uint8 = 12
	DTWht     uint8 = 14
)

// File system object constants
const (
	OwningObjIDInvalid uint64 = ^uint64(0) // ~0ULL
	OwningObjIDUnknown uint64 = ^uint64(1) // ~1ULL
	JObjMaxKeySize     uint32 = 832
	JObjMaxValueSize   uint32 = 3808
	MinDocID           uint32 = 3
)

// Crypto constants
const (
	CryptoSWID             uint64 = 4
	CryptoReserved5        uint64 = 5
	ApfsUnassignedCryptoID uint64 = ^uint64(0) // ~0ULL
	CPMaxWrappedKeySize    uint16 = 128
)

// Protection classes
const (
	ProtectionClassDirNone uint32 = 0
	ProtectionClassA       uint32 = 1
	ProtectionClassB       uint32 = 2
	ProtectionClassC       uint32 = 3
	ProtectionClassD       uint32 = 4
	ProtectionClassF       uint32 = 6
	ProtectionClassM       uint32 = 14
	CPEffectiveClassMask   uint32 = 0x0000001F
)

// Keybag tags
const (
	KBTagUnknown              uint16 = 0
	KBTagReserved1            uint16 = 1
	KBTagVolumeKey            uint16 = 2
	KBTagVolumeUnlockRecords  uint16 = 3
	KBTagVolumePassphraseHint uint16 = 4
	KBTagWrappingMKey         uint16 = 5
	KBTagVolumeMKey           uint16 = 6
	KBTagReservedF8           uint16 = 0xF8
)

// Snapshot metadata flags
const (
	SnapMetaPendingDataless uint32 = 0x00000001
	SnapMetaMergeInProgress uint32 = 0x00000002
)

// Volume reaper states
const (
	ApfsReapPhaseStart       uint32 = 0
	ApfsReapPhaseSnapshots   uint32 = 1
	ApfsReapPhaseActiveFS    uint32 = 2
	ApfsReapPhaseDestroyOmap uint32 = 3
	ApfsReapPhaseDone        uint32 = 4
)

// Reaper flags
const (
	NRBHMFlag  uint32 = 0x00000001
	NRContinue uint32 = 0x00000002
)

// Reaper list entry flags
const (
	NRLEValid        uint32 = 0x00000001
	NRLEReapIDRecord uint32 = 0x00000002
	NRLECall         uint32 = 0x00000004
	NRLECompletion   uint32 = 0x00000008
	NRLECleanup      uint32 = 0x00000010
)

// Reaper list flags
const (
	NRLIndexInvalid uint32 = 0xFFFFFFFF
)

// Integrity metadata flags
const (
	APFSSealBroken uint32 = 0x00000001
)

// Integrity metadata version constants
const (
	IntegrityMetaVersionInvalid uint32 = 0
	IntegrityMetaVersion1       uint32 = 1
	IntegrityMetaVersion2       uint32 = 2
	IntegrityMetaVersionHighest uint32 = IntegrityMetaVersion2
)

// Hash types
const (
	APFSHashInvalid    uint32 = 0
	APFSHashSHA256     uint32 = 1
	APFSHashSHA512_256 uint32 = 2
	APFSHashSHA384     uint32 = 3
	APFSHashSHA512     uint32 = 4
	APFSHashMin        uint32 = APFSHashSHA256
	APFSHashMax        uint32 = APFSHashSHA512
	APFSHashDefault    uint32 = APFSHashSHA256
)

// Hash sizes
const (
	APFSHashCCSHA256Size     uint32 = 32
	APFSHashCCSHA512_256Size uint32 = 32
	APFSHashCCSHA384Size     uint32 = 48
	APFSHashCCSHA512Size     uint32 = 64
	APFSHashMaxSize          uint32 = 64
)

// Fusion middle-tree flags
const (
	FusionMTDirty    uint32 = 0x00000001
	FusionMTTenant   uint32 = 0x00000002
	FusionMTAllFlags uint32 = 0x00000003
)

// Fusion address markers
const (
	FusionTier2DeviceByteAddr uint64 = 0x4000000000000000
)

// Encryption rolling flags
const (
	ERSBFlagEncrypting       uint64 = 0x0000000000000001
	ERSBFlagDecrypting       uint64 = 0x0000000000000002
	ERSBFlagKeyrolling       uint64 = 0x0000000000000004
	ERSBFlagPaused           uint64 = 0x0000000000000008
	ERSBFlagFailed           uint64 = 0x0000000000000010
	ERSBFlagCidIsTweak       uint64 = 0x0000000000000020
	ERSBFlagFree1            uint64 = 0x0000000000000040
	ERSBFlagFree2            uint64 = 0x0000000000000080
	ERSBFlagCMBlockSizeMask  uint64 = 0x0000000000000F00
	ERSBFlagCMBlockSizeShift        = 8
	ERSBFlagERPhaseMask      uint64 = 0x0000000000003000
	ERSBFlagERPhaseShift            = 12
	ERSBFlagFromOnekey       uint64 = 0x0000000000004000
)

// Encryption rolling phases
const (
	ERPhaseOmapRoll uint32 = 1
	ERPhaseDataRoll uint32 = 2
	ERPhaseSnapRoll uint32 = 3
)

// APFS GPT partition UUID
const (
	APFSGPTPartitionUUID          string = "7C3457EF-0000-11AA-AA11-00306543ECAC"
	APFSFVPersonalRecoveryKeyUUID string = "EBC6C064-0000-11AA-AA11-00306543ECAC"
)

// -------------------------------------------------------------------
// Constants for obj_phys_t.o_type (32-bit type + flags)
// Reference: Apple File System Reference — Object Type Masks (page 13–14)
// -------------------------------------------------------------------

const (
	// Mask to extract the object type from the low 16 bits of o_type
	OBJECT_TYPE_MASK = 0x0000ffff

	// Mask to extract all flags from the high 16 bits of o_type
	OBJECT_TYPE_FLAGS_MASK = 0xffff0000

	// Mask to extract only defined flag bits (subset of high bits)
	OBJECT_TYPE_FLAGS_DEFINED_MASK = 0xf8000000

	// Mask to extract storage type bits (used to distinguish physical/virtual/ephemeral)
	OBJ_STORAGETYPE_MASK = 0xc0000000
)

// Storage type flags for o_type (mutually exclusive)
const (
	OBJ_VIRTUAL       = 0x00000000 // No bits set = virtual object
	OBJ_EPHEMERAL     = 0x80000000 // Bit 31 set
	OBJ_PHYSICAL      = 0x40000000 // Bit 30 set
	OBJ_NOHEADER      = 0x20000000 // Object does not use obj_phys_t header
	OBJ_ENCRYPTED     = 0x10000000 // Object is encrypted
	OBJ_NONPERSISTENT = 0x08000000 // Ephemeral object that does not persist across unmount
)

// -------------------------------------------------------------------
// Constants for j_key_t.j_oid (64-bit packed object ID and type)
// Reference: Apple File System Reference — j_key_t / j_inode_key_t (page ~73)
// -------------------------------------------------------------------

const (
	// Mask to extract the object ID from j_oid (low 60 bits)
	OBJ_ID_MASK = 0x0fffffffffffffff

	// Bit position where the object type begins in j_oid (bits 60–63)
	OBJ_TYPE_SHIFT = 60

	// Mask to extract the object type from j_oid (high 4 bits)
	OBJECT_TYPE_MASK_JKEY = 0xf000000000000000
)

// -------------------------------------------------------------------
// Directory Record (j_drec_hashed_key_t) masks
// Reference: Apple File System Reference, page ~78
// -------------------------------------------------------------------

const (
	JDrecLenMask   uint32 = 0x000003FF // Lower 10 bits: name length
	JDrecHashMask  uint32 = 0xFFFFFC00 // Upper 22 bits: name hash
	JDrecHashShift uint32 = 10         // Hash is shifted by 10 bits
)
