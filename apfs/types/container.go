package types

// Container (pages 26-43)
// The container includes several top-level objects that are shared by all of the container's volumes.

// NxSuperblockT is a container superblock.
// Reference: page 27
type NxSuperblockT struct {
	// The object's header. (page 27)
	NxO ObjPhysT
	// A number that can be used to verify that you're reading an instance of nx_superblock_t. (page 27)
	// The value of this field is always NxMagic.
	NxMagic uint32
	// The logical block size used in the Apple File System container. (page 29)
	NxBlockSize uint32
	// The total number of logical blocks available in the container. (page 29)
	NxBlockCount uint64
	// A bit field of the optional features being used by this container. (page 29)
	NxFeatures uint64
	// A bit field of the read-only compatible features being used by this container. (page 29)
	NxReadonlyCompatibleFeatures uint64
	// A bit field of the backward-incompatible features being used by this container. (page 29)
	NxIncompatibleFeatures uint64
	// The universally unique identifier of this container. (page 29)
	NxUuid UUID
	// The next object identifier to be used for a new ephemeral or virtual object. (page 30)
	NxNextOid OidT
	// The next transaction to be used. (page 30)
	NxNextXid XidT
	// The number of blocks used by the checkpoint descriptor area. (page 30)
	// The highest bit of this number is used as a flag, as discussed in nx_xp_desc_base.
	// Ignore that bit when accessing this field as a count.
	NxXpDescBlocks uint32
	// The number of blocks used by the checkpoint data area. (page 30)
	// The highest bit of this number is used as a flag, as discussed in nx_xp_data_base.
	// Ignore that bit when accessing this field as a count.
	NxXpDataBlocks uint32
	// The base address of the checkpoint descriptor area or the physical object identifier
	// of a tree that contains the address information. (page 30)
	NxXpDescBase Paddr
	// The base address of the checkpoint data area or the physical object identifier of a tree
	// that contains the address information. (page 30)
	NxXpDataBase Paddr
	// The next index to use in the checkpoint descriptor area. (page 31)
	NxXpDescNext uint32
	// The next index to use in the checkpoint data area. (page 31)
	NxXpDataNext uint32
	// The index of the first valid item in the checkpoint descriptor area. (page 31)
	NxXpDescIndex uint32
	// The number of blocks in the checkpoint descriptor area used by the checkpoint
	// that this superblock belongs to. (page 31)
	NxXpDescLen uint32
	// The index of the first valid item in the checkpoint data area. (page 31)
	NxXpDataIndex uint32
	// The number of blocks in the checkpoint data area used by the checkpoint
	// that this superblock belongs to. (page 31)
	NxXpDataLen uint32
	// The ephemeral object identifier for the space manager. (page 32)
	NxSpacemanOid OidT
	// The physical object identifier for the container's object map. (page 32)
	NxOmapOid OidT
	// The ephemeral object identifier for the reaper. (page 32)
	NxReaperOid OidT
	// Reserved for testing. (page 32)
	NxTestType uint32
	// The maximum number of volumes that can be stored in this container. (page 32)
	NxMaxFileSystems uint32
	// An array of virtual object identifiers for volumes. (page 32)
	// The objects' types are all OBJECT_TYPE_BTREE and their subtypes are all OBJECT_TYPE_FSTREE.
	NxFsOid [NxMaxFileSystems]OidT
	// An array of counters that store information about the container. (page 33)
	// These counters are primarily intended to help during development and debugging.
	NxCounters [NxNumCounters]uint64
	// The physical range of blocks where space will not be allocated. (page 33)
	NxBlockedOutPrange Prange
	// The physical object identifier of a tree used to keep track of objects
	// that must be moved out of blocked-out storage. (page 33)
	NxEvictMappingTreeOid OidT
	// Other container flags. (page 33)
	NxFlags uint64
	// The physical object identifier of the object that contains EFI driver data extents. (page 33)
	NxEfiJumpstart Paddr
	// The universally unique identifier of the container's Fusion set, or zero for non-Fusion containers. (page 34)
	NxFusionUuid UUID
	// The location of the container's keybag. (page 34)
	NxKeylocker Prange
	// An array of fields used in the management of ephemeral data. (page 34)
	NxEphemeralInfo [NxEphInfoCount]uint64
	// Reserved for testing. (page 34)
	NxTestOid OidT
	// The physical object identifier of the Fusion middle tree, or zero if for non-Fusion drives. (page 34)
	NxFusionMtOid OidT
	// The ephemeral object identifier of the Fusion write-back cache state, or zero for non-Fusion drives. (page 35)
	NxFusionWbcOid OidT
	// The blocks used for the Fusion write-back cache area, or zero for non-Fusion drives. (page 35)
	NxFusionWbc Prange
	// Reserved. (page 35)
	NxNewestMountedVersion uint64
	// Wrapped media key. (page 35)
	NxMkbLocker Prange
}

// NxMagic is the value of the nx_magic field.
// This magic number was chosen because in hex dumps it appears as "NXSB",
// which is an abbreviated form of NX superblock.
// Reference: page 35
const NxMagic uint32 = 'B' | 'S'<<8 | 'X'<<16 | 'N'<<24 // 'BSXN'

// NxMaxFileSystems is the maximum number of volumes that can be in a single container.
// Reference: page 35
const NxMaxFileSystems = 100

// NxEphInfoCount is the length of the array in the nx_ephemeral_info field.
// Reference: page 35
const NxEphInfoCount = 4

// NxEphMinBlockCount is the default minimum size, in blocks, for structures that contain ephemeral data.
// Reference: page 36
const NxEphMinBlockCount = 8

// NxMaxFileSystemEphStructs is the number of structures that contain ephemeral data that a volume can have.
// Reference: page 36
const NxMaxFileSystemEphStructs = 4

// NxTxMinCheckpointCount is the minimum number of checkpoints that can fit in the checkpoint data area.
// Reference: page 36
const NxTxMinCheckpointCount = 4

// NxEphInfoVersion1 is the version number for structures that contain ephemeral data.
// Reference: page 36
const NxEphInfoVersion1 = 1

// Container Flags (pages 36-37)

// NxReserved1 is a reserved flag.
// Don't set this flag, but preserve it if it's already set.
// Reference: page 36
const NxReserved1 uint64 = 0x00000001

// NxReserved2 is a reserved flag.
// Don't add this flag to a container. If this flag is set, preserve it when reading
// the container, and remove it when modifying the container.
// Reference: page 37
const NxReserved2 uint64 = 0x00000002

// NxCryptoSw indicates the container uses software cryptography.
// Reference: page 37
const NxCryptoSw uint64 = 0x00000004

// Optional Container Feature Flags (page 37)

// NxFeatureDefrag indicates the volumes in this container support defragmentation.
// Reference: page 37
const NxFeatureDefrag uint64 = 0x0000000000000001

// NxFeatureLcfd indicates this container is using low-capacity Fusion Drive mode.
// Reference: page 37
const NxFeatureLcfd uint64 = 0x0000000000000002

// NxSupportedFeaturesMask is a bit mask of all the optional features.
// Reference: page 37
const NxSupportedFeaturesMask uint64 = NxFeatureDefrag | NxFeatureLcfd

// Read-Only Compatible Container Feature Flags (page 38)

// NxSupportedRocompatMask is a bit mask of all read-only compatible features.
// Reference: page 38
const NxSupportedRocompatMask uint64 = 0x0

// Incompatible Container Feature Flags (pages 38-39)

// NxIncompatVersion1 indicates the container uses version 1 of Apple File System,
// as implemented in macOS 10.12.
// Reference: page 38
const NxIncompatVersion1 uint64 = 0x0000000000000001

// NxIncompatVersion2 indicates the container uses version 2 of Apple File System,
// as implemented in macOS 10.13 and iOS 10.3.
// Reference: page 38
const NxIncompatVersion2 uint64 = 0x0000000000000002

// NxIncompatFusion indicates the container supports Fusion Drives.
// Reference: page 38
const NxIncompatFusion uint64 = 0x0000000000000100

// NxSupportedIncompatMask is a bit mask of all the backward-incompatible features.
// Reference: page 39
const NxSupportedIncompatMask uint64 = NxIncompatVersion2 | NxIncompatFusion

// Block and Container Sizes (page 39)

// NxMinimumBlockSize is the smallest supported size, in bytes, for a block.
// Reference: page 39
const NxMinimumBlockSize = 4096

// NxDefaultBlockSize is the default size, in bytes, for a block.
// Reference: page 39
const NxDefaultBlockSize = 4096

// NxMaximumBlockSize is the largest supported size, in bytes, for a block.
// Reference: page 39
const NxMaximumBlockSize = 65536

// NxMinimumContainerSize is the smallest supported size, in bytes, for a container.
// Reference: page 39
const NxMinimumContainerSize = 1048576

// NxCounterIdT contains indexes into a container superblock's array of counters.
// Reference: page 39-40
type NxCounterIdT int

const (
	// NxCntrObjCksumSet is the number of times a checksum has been computed while writing objects to disk.
	// Reference: page 40
	NxCntrObjCksumSet NxCounterIdT = 0

	// NxCntrObjCksumFail is the number of times an object's checksum was invalid when reading from disk.
	// Reference: page 40
	NxCntrObjCksumFail NxCounterIdT = 1

	// NxNumCounters is the maximum number of counters.
	// Reference: page 40
	NxNumCounters = 32
)

// CheckpointMappingT is a mapping from an ephemeral object identifier to its physical address in the checkpoint data area.
// Reference: page 40
type CheckpointMappingT struct {
	// The object's type. (page 40)
	// An object type is a 32-bit value: The low 16 bits indicate the type using the values listed in Object Types,
	// and the high 16 bits are flags using the values listed in Object Type Flags.
	// This field has the same meaning and behavior as the o_type field of obj_phys_t.
	CpmType uint32

	// The object's subtype. (page 41)
	// One of the values listed in Object Types.
	// Subtypes indicate the type of data stored in a data structure such as a B-tree.
	// For example, a leaf node in a B-tree that contains file-system records has a type of OBJECT_TYPE_BTREE_NODE
	// and a subtype of OBJECT_TYPE_FSTREE.
	// This field has the same meaning and behavior as the o_subtype field of obj_phys_t.
	CpmSubtype uint32

	// The size, in bytes, of the object. (page 41)
	CpmSize uint32

	// Reserved. (page 41)
	// Populate this field with zero when you create a new mapping, and preserve its value
	// when you modify an existing mapping.
	// This field is padding.
	CpmPad uint32

	// The virtual object identifier of the volume that the object is associated with. (page 41)
	CpmFsOid OidT

	// The ephemeral object identifier. (page 41)
	CpmOid OidT

	// The address in the checkpoint data area where the object is stored. (page 41)
	CpmPaddr Paddr
}

// CheckpointMapPhysT is a checkpoint-mapping block.
// Reference: page 41
type CheckpointMapPhysT struct {
	// The object's header. (page 42)
	CpmO ObjPhysT

	// A bit field that contains additional information about the list of checkpoint mappings. (page 42)
	// For the values used in this bit field, see Checkpoint Flags.
	CpmFlags uint32

	// The number of checkpoint mappings in the array. (page 42)
	CpmCount uint32

	// The array of checkpoint mappings. (page 42)
	// This is defined as a variable-sized slice in Go to accommodate the array of mappings.
	CpmMap []CheckpointMappingT
}

// Checkpoint Flags (page 42)

// CheckpointMapLast is a flag marking the last checkpoint-mapping block in a given checkpoint.
// Reference: page 42
const CheckpointMapLast uint32 = 0x00000001

// EvictMappingValT is a range of physical addresses that data is being moved into.
// Reference: page 42
type EvictMappingValT struct {
	// The address where the destination starts. (page 43)
	DstPaddr Paddr

	// The number of blocks being moved. (page 43)
	Len uint64
}
