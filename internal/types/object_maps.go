package types

import "math"

// Object Maps (pages 44-50)
// An object map uses a B-tree to store a mapping from virtual object identifiers
// and transaction identifiers to the physical addresses where those objects are stored.

// OmapPhysT is an object map.
// Reference: page 44
type OmapPhysT struct {
	// The object's header. (page 45)
	OmO ObjPhysT

	// The object map's flags. (page 45)
	// For the values used in this bit field, see Object Map Flags.
	OmFlags uint32

	// The number of snapshots that this object map has. (page 45)
	OmSnapCount uint32

	// The type of tree being used for object mappings. (page 45)
	OmTreeType uint32

	// The type of tree being used for snapshots. (page 45)
	OmSnapshotTreeType uint32

	// The virtual object identifier of the tree being used for object mappings. (page 45)
	OmTreeOid OidT

	// The virtual object identifier of the tree being used to hold snapshot information. (page 45)
	OmSnapshotTreeOid OidT

	// The transaction identifier of the most recent snapshot that's stored in this object map. (page 45)
	OmMostRecentSnap XidT

	// The smallest transaction identifier for an in-progress revert. (page 46)
	OmPendingRevertMin XidT

	// The largest transaction identifier for an in-progress revert. (page 46)
	OmPendingRevertMax XidT
}

// OmapKeyT is a key used to access an entry in the object map.
// Reference: page 46
type OmapKeyT struct {
	// The object identifier. (page 46)
	OkOid OidT

	// The transaction identifier. (page 46)
	OkXid XidT
}

// OmapValT is a value in the object map.
// Reference: page 46
type OmapValT struct {
	// A bit field of flags. (page 46)
	// For the values used in this bit field, see Object Map Value Flags.
	OvFlags uint32

	// The size, in bytes, of the object. (page 47)
	// This value must be a multiple of the container's logical block size.
	// If the object is smaller than one logical block, the value of this field
	// is the size of one logical block.
	OvSize uint32

	// The address of the object. (page 47)
	OvPaddr Paddr
}

// OmapSnapshotT is information about a snapshot of an object map.
// Reference: page 47
type OmapSnapshotT struct {
	// The snapshot's flags. (page 47)
	// For the values used in this bit field, see Snapshot Flags.
	OmsFlags uint32

	// Reserved. (page 47)
	// Populate this field with zero when you create a new snapshot,
	// and preserve its value when you modify an existing snapshot.
	// This field is padding.
	OmsPad uint32

	// Reserved. (page 47-48)
	// Populate this field with zero when you create a new snapshot,
	// and preserve its value when you modify an existing snapshot.
	OmsOid OidT
}

// Object Map Value Flags (page 48)

// OmapValDeleted indicates the object has been deleted, and this mapping is a placeholder.
// Reference: page 48
const OmapValDeleted uint32 = 0x00000001

// OmapValSaved indicates this object mapping shouldn't be replaced when the object is updated.
// Reference: page 48
const OmapValSaved uint32 = 0x00000002

// OmapValEncrypted indicates the object is encrypted.
// Reference: page 48
const OmapValEncrypted uint32 = 0x00000004

// OmapValNoheader indicates the object is stored without an obj_phys_t header.
// Reference: page 48
const OmapValNoheader uint32 = 0x00000008

// OmapValCryptoGeneration is a one-bit flag that tracks encryption configuration.
// Reference: page 48
const OmapValCryptoGeneration uint32 = 0x00000010

// Snapshot Flags (page 49)

// OmapSnapshotDeleted indicates the snapshot has been deleted.
// Reference: page 49
const OmapSnapshotDeleted uint32 = 0x00000001

// OmapSnapshotReverted indicates the snapshot has been deleted as part of a revert.
// Reference: page 49
const OmapSnapshotReverted uint32 = 0x00000002

// Object Map Flags (pages 49-50)

// OmapManuallyManaged indicates the object map doesn't support snapshots.
// Reference: page 49
const OmapManuallyManaged uint32 = 0x00000001

// OmapEncrypting indicates a transition is in progress from unencrypted storage to encrypted storage.
// Reference: page 49
const OmapEncrypting uint32 = 0x00000002

// OmapDecrypting indicates a transition is in progress from encrypted storage to unencrypted storage.
// Reference: page 49
const OmapDecrypting uint32 = 0x00000004

// OmapKeyrolling indicates a transition is in progress from encrypted storage using an old key
// to encrypted storage using a new key.
// Reference: page 50
const OmapKeyrolling uint32 = 0x00000008

// OmapCryptoGeneration is a one-bit flag that tracks encryption configuration.
// Reference: page 50
const OmapCryptoGeneration uint32 = 0x00000010

// OmapValidFlags is a bit mask of all valid object map flags.
// Reference: page 50
const OmapValidFlags uint32 = 0x0000001f

// Object Map Constants (page 50)

// OmapMaxSnapCount is the maximum number of snapshots that can be stored in an object map.
// Reference: page 50
const OmapMaxSnapCount uint32 = math.MaxUint32

// Object Map Reaper Phases (page 50)

// OmapReapPhaseMapTree indicates the reaper is deleting entries from the object mapping tree.
// Reference: page 50
const OmapReapPhaseMapTree uint32 = 1

// OmapReapPhaseSnapshotTree indicates the reaper is deleting entries from the snapshot tree.
// Reference: page 50
const OmapReapPhaseSnapshotTree uint32 = 2
