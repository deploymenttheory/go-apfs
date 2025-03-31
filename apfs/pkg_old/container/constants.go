// File: pkg/container/constants.go
package container

// APFS magic constants
const (
	NXMagic   uint32 = 0x4253584E // 'NXSB'
	APFSMagic uint32 = 0x42535041 // 'APSB'
	JSMagic   uint32 = 0x5244534A // 'JSDR' (JumpStart Driver)
	ERMagic   uint32 = 0x46414C42 // 'FLAB' (Encryption Rolling)

	MaxCksumSize uint32 = 8 // Maximum checksum size
)

// Block size constants and structural limits
const (
	MinBlockSize     uint32 = 4096    // Minimum supported block size
	DefaultBlockSize uint32 = 4096    // Default block size used in most containers
	MaxBlockSize     uint32 = 65536   // Maximum supported block size
	MinContainerSize uint64 = 1048576 // Minimum container size in bytes (1 MiB)

	NXMaxFileSystems       = 100  // Maximum number of file systems (volumes) in a container
	NXMaxXPDescBlocks      = 32   // Maximum number of checkpoint descriptor blocks
	NXMaxXPDataBlocks      = 8192 // Maximum number of checkpoint data blocks
	NXNumCounters          = 32   // Number of counters in nx_superblock_t
	NXEphemeralInfoCount   = 4    // Number of ephemeral_info[] entries in nx_superblock_t
	SpacemanFreeQueueCount = 3    // Number of spaceman free queues (internal, main, tier2)
	NXMinimumBlockSize     = 4096
	NXMaximumBlockSize     = 65536
)

// Object identifier constants
const (
	OIDInvalid       uint64 = 0
	OIDNXSuperblock  uint64 = 1
	OIDReservedCount uint64 = 1024
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

// Container feature flags
const (
	// NX flags (nx_flags field of nx_superblock_t)
	NXReserved1 uint64 = 0x00000001
	NXReserved2 uint64 = 0x00000002
	NXCryptoSW  uint64 = 0x00000004

	// Optional features (nx_features field of nx_superblock_t)
	NXFeatureDefrag uint64 = 0x0000000000000001
	NXFeatureLCFD   uint64 = 0x0000000000000002

	NXSupportedFeaturesMask uint64 = NXFeatureDefrag | NXFeatureLCFD
	UnsupportedFeaturesMask uint64 = ^NXSupportedFeaturesMask

	// Read-only compatible features (nx_readonly_compatible_features field of nx_superblock_t)
	NXSupportedROCompatMask uint64 = 0x0

	// Incompatible features (nx_incompatible_features field of nx_superblock_t)
	NXIncompatVersion1 uint64 = 0x0000000000000001
	NXIncompatVersion2 uint64 = 0x0000000000000002
	NXIncompatFusion   uint64 = 0x0000000000000100

	NXSupportedIncompatMask         uint64 = NXIncompatVersion2 | NXIncompatFusion
	UnsupportedIncompatFeaturesMask uint64 = ^NXSupportedIncompatMask
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

// BTNodeFlags represent flag values for B-tree nodes
const (
	BTNodeRoot        uint16 = 0x0001
	BTNodeLeaf        uint16 = 0x0002
	BTNodeFixedKVSize uint16 = 0x0004
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

// Reaper flags
const (
	NRBHMFlag  uint32 = 0x00000001
	NRContinue uint32 = 0x00000002
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

// Storage type constants from obj_phys_t.o_type
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

// Keybag Entry Tags
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

// Crypto Constants
const (
	// Crypto State Identifiers
	CryptoSWID             uint64 = 4
	CryptoReserved5        uint64 = 5
	ApfsUnassignedCryptoID uint64 = ^uint64(0) // ~0ULL

	// Key Wrapping and Encryption
	CPMaxWrappedKeySize uint16 = 128
	MaxKeyLength        uint16 = 128

	// Encryption Rolling Magic Number
	ERVersion uint32 = 1
)

// Encryption Rolling Phases
const (
	ERPhaseOmapRoll uint32 = 1
	ERPhaseDataRoll uint32 = 2
	ERPhaseSnapRoll uint32 = 3
)

// Encryption Rolling Flags
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

// APFS GPT partition UUID
const (
	APFSGPTPartitionUUID          string = "7C3457EF-0000-11AA-AA11-00306543ECAC"
	APFSFVPersonalRecoveryKeyUUID string = "EBC6C064-0000-11AA-AA11-00306543ECAC"
)
