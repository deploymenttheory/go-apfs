package types

// Objects (pages 10-21)
// Depending on how they're stored, objects have some differences, the most important
// of which is the way you use an object identifier to find an object.

// OidT is an object identifier.
// For a physical object, its identifier is the logical block address on disk where the object is stored.
// For an ephemeral object, its identifier is a number.
// For a virtual object, its identifier is a number.
// Reference: page 12
type OidT uint64

// XidT is a transaction identifier.
// Transactions are uniquely identified by a monotonically increasing number.
// The number zero isn't a valid transaction identifier.
// Reference: page 12
type XidT uint64

// ObjPhysT is a header used at the beginning of all objects.
// Reference: page 10
type ObjPhysT struct {
	// The Fletcher 64 checksum of the object, with length matching MaxCksumSize. (page 10)
	OChecksum [MaxCksumSize]byte
	// The object's identifier. (page 11)
	OOid OidT
	// The identifier of the most recent transaction that this object was modified in. (page 11)
	OXid XidT
	// The object's type and flags. (page 11)
	// An object type is a 32-bit value: The low 16 bits indicate the type, and the high 16 bits are flags.
	OType uint32
	// The object's subtype. (page 11)
	// Subtypes indicate the type of data stored in a data structure such as a B-tree.
	OSubtype uint32
}

// Object Identifier Constants (pages 12-13)

// XidInvalid is an invalid transaction identifier.
// Reference: page 12
const XidInvalid XidT = 0

// OidNxSuperblock is the ephemeral object identifier for the container superblock.
// Reference: page 13
const OidNxSuperblock OidT = 1

// OidInvalid is an invalid object identifier.
// Reference: page 13
const OidInvalid OidT = 0

// OidReservedCount is the number of object identifiers that are reserved for objects with a fixed object identifier.
// Reference: page 13
const OidReservedCount uint64 = 1024

// Object Type Masks (pages 13-14)

// ObjectTypeMask is the bit mask used to access the type.
// Reference: page 13
const ObjectTypeMask uint32 = 0x0000ffff

// ObjectTypeFlagsMask is the bit mask used to access the flags.
// Reference: page 13
const ObjectTypeFlagsMask uint32 = 0xffff0000

// ObjStorageTypeMask is the bit mask used to access the storage portion of the object type.
// Reference: page 14
const ObjStorageTypeMask uint32 = 0xc0000000

// ObjectTypeFlagsDefinedMask is a bit mask of all bits for which flags are defined.
// Reference: page 14
const ObjectTypeFlagsDefinedMask uint32 = 0xf8000000

// MaxCksumSize is the number of bytes used for an object checksum.
// Reference: page 11
const MaxCksumSize = 8

// Object Types (pages 14-19)

// ObjectTypeNxSuperblock is a container superblock (nx_superblock_t).
// Reference: page 15
const ObjectTypeNxSuperblock uint32 = 0x00000001

// ObjectTypeBtree is a B-tree root node (btree_node_phys_t).
// Reference: page 15
const ObjectTypeBtree uint32 = 0x00000002

// ObjectTypeBtreeNode is a B-tree node (btree_node_phys_t).
// Reference: page 15
const ObjectTypeBtreeNode uint32 = 0x00000003

// ObjectTypeSpaceman is a space manager (spaceman_phys_t).
// Reference: page 15
const ObjectTypeSpaceman uint32 = 0x00000005

// ObjectTypeSpacemanCab is a chunk-info address block (cib_addr_block) used by the space manager.
// Reference: page 15
const ObjectTypeSpacemanCab uint32 = 0x00000006

// ObjectTypeSpacemanCib is a chunk-info block (chunk_info_block) used by the space manager.
// Reference: page 15
const ObjectTypeSpacemanCib uint32 = 0x00000007

// ObjectTypeSpacemanBitmap is a free-space bitmap used by the space manager.
// Reference: page 16
const ObjectTypeSpacemanBitmap uint32 = 0x00000008

// ObjectTypeSpacemanFreeQueue is a free-space queue used by the space manager.
// Reference: page 16
const ObjectTypeSpacemanFreeQueue uint32 = 0x00000009

// ObjectTypeExtentListTree is an extents-list tree.
// Reference: page 16
const ObjectTypeExtentListTree uint32 = 0x0000000a

// ObjectTypeOmap is an object map.
// Reference: page 16
const ObjectTypeOmap uint32 = 0x0000000b

// ObjectTypeCheckpointMap is a checkpoint map.
// Reference: page 16
const ObjectTypeCheckpointMap uint32 = 0x0000000c

// ObjectTypeFs is a volume (apfs_superblock_t).
// Reference: page 16
const ObjectTypeFs uint32 = 0x0000000d

// ObjectTypeFstree is a tree containing file-system records.
// Reference: page 16
const ObjectTypeFstree uint32 = 0x0000000e

// ObjectTypeBlockreftree is a tree containing extent references.
// Reference: page 17
const ObjectTypeBlockreftree uint32 = 0x0000000f

// ObjectTypeSnapmetatree is a tree containing snapshot metadata for a volume.
// Reference: page 17
const ObjectTypeSnapmetatree uint32 = 0x00000010

// ObjectTypeNxReaper is a reaper (nx_reaper_phys_t).
// Reference: page 17
const ObjectTypeNxReaper uint32 = 0x00000011

// ObjectTypeNxReapList is a reaper list (nx_reap_list_phys_t).
// Reference: page 17
const ObjectTypeNxReapList uint32 = 0x00000012

// ObjectTypeOmapSnapshot is a tree containing information about snapshots of an object map.
// Reference: page 17
const ObjectTypeOmapSnapshot uint32 = 0x00000013

// ObjectTypeEfiJumpstart is EFI information used for booting (nx_efi_jumpstart_t).
// Reference: page 17
const ObjectTypeEfiJumpstart uint32 = 0x00000014

// ObjectTypeFusionMiddleTree is a tree used for Fusion devices to track blocks.
// Reference: page 17
const ObjectTypeFusionMiddleTree uint32 = 0x00000015

// ObjectTypeNxFusionWbc is a write-back cache state used for Fusion devices.
// Reference: page 18
const ObjectTypeNxFusionWbc uint32 = 0x00000016

// ObjectTypeNxFusionWbcList is a write-back cache list used for Fusion devices.
// Reference: page 18
const ObjectTypeNxFusionWbcList uint32 = 0x00000017

// ObjectTypeErState is an encryption-rolling state.
// Reference: page 18
const ObjectTypeErState uint32 = 0x00000018

// ObjectTypeGbitmap is a general-purpose bitmap.
// Reference: page 18
const ObjectTypeGbitmap uint32 = 0x00000019

// ObjectTypeGbitmapTree is a B-tree of general-purpose bitmaps.
// Reference: page 18
const ObjectTypeGbitmapTree uint32 = 0x0000001a

// ObjectTypeGbitmapBlock is a block containing a general-purpose bitmap.
// Reference: page 18
const ObjectTypeGbitmapBlock uint32 = 0x0000001b

// ObjectTypeErRecoveryBlock is information that can be used to recover from a system crash
// during the encryption rolling process.
// Reference: page 18
const ObjectTypeErRecoveryBlock uint32 = 0x0000001c

// ObjectTypeSnapMetaExt is additional metadata about snapshots.
// Reference: page 18
const ObjectTypeSnapMetaExt uint32 = 0x0000001d

// ObjectTypeIntegrityMeta is an integrity metadata object.
// Reference: page 19
const ObjectTypeIntegrityMeta uint32 = 0x0000001e

// ObjectTypeFextTree is a B-tree of file extents.
// Reference: page 19
const ObjectTypeFextTree uint32 = 0x0000001f

// ObjectTypeReserved20 is reserved.
// Reference: page 19
const ObjectTypeReserved20 uint32 = 0x00000020

// ObjectTypeInvalid indicates an invalid object.
// Reference: page 19
const ObjectTypeInvalid uint32 = 0x00000000

// ObjectTypeTest is reserved for testing.
// Reference: page 19
const ObjectTypeTest uint32 = 0x000000ff

// ObjectTypeContainerKeybag represents a container's keybag.
// Reference: page 19
const ObjectTypeContainerKeybag uint32 = 'k' | 'e'<<8 | 'y'<<16 | 's'<<24 // 'keys'

// ObjectTypeVolumeKeybag represents a volume's keybag.
// Reference: page 19
const ObjectTypeVolumeKeybag uint32 = 'r' | 'e'<<8 | 'c'<<16 | 's'<<24 // 'recs'

// ObjectTypeMediaKeybag represents a media keybag.
// Reference: page 19
const ObjectTypeMediaKeybag uint32 = 'm' | 'k'<<8 | 'e'<<16 | 'y'<<24 // 'mkey'

// Object Type Flags (pages 20-21)

// ObjVirtual indicates a virtual object.
// Reference: page 20
const ObjVirtual uint32 = 0x00000000

// ObjEphemeral indicates an ephemeral object.
// Reference: page 20
const ObjEphemeral uint32 = 0x80000000

// ObjPhysical indicates a physical object.
// Reference: page 20
const ObjPhysical uint32 = 0x40000000

// ObjNoheader indicates an object stored without an obj_phys_t header.
// Reference: page 20
const ObjNoheader uint32 = 0x20000000

// ObjEncrypted indicates an encrypted object.
// Reference: page 21
const ObjEncrypted uint32 = 0x10000000

// ObjNonpersistent indicates an ephemeral object that isn't persisted across unmounting.
// Reference: page 21
const ObjNonpersistent uint32 = 0x08000000
