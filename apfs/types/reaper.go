package types

// Reaper (pages 164-168)
// The reaper is a mechanism that allows large objects to be deleted over a period
// spanning multiple transactions. There's exactly one instance of this structure in a container.

// NxReaperPhysT is the main reaper structure.
// Reference: page 164
type NxReaperPhysT struct {
	// The object's header.
	NrO ObjPhysT

	// The next reap identifier to be assigned.
	NrNextReapId uint64

	// The identifier of the last completed reap.
	NrCompletedId uint64

	// The object identifier of the head of the reaper list.
	NrHead OidT

	// The object identifier of the tail of the reaper list.
	NrTail OidT

	// The reaper flags.
	NrFlags uint32

	// The count of reaper lists.
	NrRlcount uint32

	// The type of the object being reaped.
	NrType uint32

	// The size of the object being reaped.
	NrSize uint32

	// The filesystem object identifier of the object being reaped.
	NrFsOid OidT

	// The object identifier of the object being reaped.
	NrOid OidT

	// The transaction identifier for the object being reaped.
	NrXid XidT

	// The flags for the reaper list entry.
	NrNrleFlags uint32

	// The size of the state buffer.
	NrStateBufferSize uint32

	// The state buffer for the reaper.
	NrStateBuffer []byte
}

// NxReapListPhysT represents a list of objects to be reaped.
// Reference: page 164
type NxReapListPhysT struct {
	// The object's header.
	NrlO ObjPhysT

	// The object identifier of the next reaper list.
	NrlNext OidT

	// The flags for this reaper list.
	NrlFlags uint32

	// The maximum number of entries in this list.
	NrlMax uint32

	// The number of entries currently in this list.
	NrlCount uint32

	// The index of the first entry.
	NrlFirst uint32

	// The index of the last entry.
	NrlLast uint32

	// The index of the first free entry.
	NrlFree uint32

	// The entries in this reaper list.
	NrlEntries []NxReapListEntryT
}

// NxReapListEntryT represents an entry in a reaper list.
// Reference: page 164
type NxReapListEntryT struct {
	// The index of the next entry.
	NrleNext uint32

	// The flags for this entry.
	NrleFlags uint32

	// The type of object to reap.
	NrleType uint32

	// The size of the object to reap.
	NrleSize uint32

	// The filesystem object identifier of the object to reap.
	NrleFsOid OidT

	// The object identifier of the object to reap.
	NrleOid OidT

	// The transaction identifier for the object to reap.
	NrleXid XidT
}

// Volume Reaper States (page 165)

const (
	// ApfsReapPhaseStart is the initial phase of reaper.
	// Reference: page 165
	ApfsReapPhaseStart = 0

	// ApfsReapPhaseSnapshots is the phase where snapshots are being reaped.
	// Reference: page 165
	ApfsReapPhaseSnapshots = 1

	// ApfsReapPhaseActiveFs is the phase where active filesystem objects are being reaped.
	// Reference: page 165
	ApfsReapPhaseActiveFs = 2

	// ApfsReapPhaseDestroyOmap is the phase where the object map is being destroyed.
	// Reference: page 165
	ApfsReapPhaseDestroyOmap = 3

	// ApfsReapPhaseDone is the phase when the reaper is finished.
	// Reference: page 165
	ApfsReapPhaseDone = 4
)

// Reaper Flags (page 165)

// NrBhmFlag is a reserved flag that must always be set.
// Reference: page 165
const NrBhmFlag uint32 = 0x00000001

// NrContinue indicates the current object is being reaped.
// Reference: page 165
const NrContinue uint32 = 0x00000002

// Reaper List Entry Flags (page 165)

// NrleValid indicates the entry is valid.
// Reference: page 166
const NrleValid uint32 = 0x00000001

// NrleReapIdRecord indicates this is a reap ID record.
// Reference: page 166
const NrleReapIdRecord uint32 = 0x00000002

// NrleCall indicates the entry is ready to be called.
// Reference: page 166
const NrleCall uint32 = 0x00000004

// NrleCompletion indicates the entry is a completion entry.
// Reference: page 166
const NrleCompletion uint32 = 0x00000008

// NrleCleanup indicates the entry is a cleanup entry.
// Reference: page 166
const NrleCleanup uint32 = 0x00000010

// Reaper List Flags (page 166)

// NrlIndexInvalid is an invalid index for a reaper list.
// Reference: page 166
const NrlIndexInvalid uint32 = 0xffffffff

// OmapReapStateT represents the state used when reaping an object map.
// Reference: page 166
type OmapReapStateT struct {
	// The current reaping phase.
	// For the values used in this field, see Object Map Reaper Phases.
	OmrPhase uint32

	// The key of the most recently freed entry in the object map.
	// This field allows the reaper to resume after the last entry it processed.
	OmrOk OmapKeyT
}

// OmapCleanupStateT represents the state used when reaping to clean up deleted snapshots.
// Reference: page 166
type OmapCleanupStateT struct {
	// A flag that indicates whether the structure has valid data in it.
	// If the value of this field is zero, the structure has been allocated and zeroed,
	// but doesn't yet contain valid data. Otherwise, the structure is valid.
	OmcCleaning uint32

	// The flags for the snapshot being deleted.
	// The value for this field is the same as the value of the snapshot's omap_snapshot_t.oms_flags field.
	OmcOmsflags uint32

	// The transaction identifier of the snapshot prior to the snapshots being deleted.
	OmcSxidprev XidT

	// The transaction identifier of the first snapshot being deleted.
	OmcSxidstart XidT

	// The transaction identifier of the last snapshot being deleted.
	OmcSxidend XidT

	// The transaction identifier of the snapshot after the snapshots being deleted.
	OmcSxidnext XidT

	// The key of the next object mapping to consider for deletion.
	OmcCurkey OmapKeyT
}

// ApfsReapStateT contains the state for APFS reaping.
// Reference: page 167
type ApfsReapStateT struct {
	// The last physical block number being processed.
	LastPbn uint64

	// The current snapshot's transaction identifier.
	CurSnapXid XidT

	// The current phase of reaping.
	Phase uint32
}
