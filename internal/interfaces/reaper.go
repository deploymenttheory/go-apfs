// File: internal/interfaces/reaper.go
package interfaces

import (
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// ReaperReader provides methods for reading reaper information
type ReaperReader interface {
	// NextReapID returns the next reap identifier to be assigned
	NextReapID() uint64

	// CompletedReapID returns the identifier of the last completed reap
	CompletedReapID() uint64

	// HeadOID returns the object identifier of the head of the reaper list
	HeadOID() types.OidT

	// TailOID returns the object identifier of the tail of the reaper list
	TailOID() types.OidT

	// Flags returns the reaper flags
	Flags() uint32

	// ReaperListCount returns the count of reaper lists
	ReaperListCount() uint32
}

// ReaperObjectInfo provides details about an object being reaped
type ReaperObjectInfo interface {
	// Type returns the type of the object being reaped
	Type() uint32

	// Size returns the size of the object being reaped
	Size() uint32

	// FileSystemOID returns the filesystem object identifier
	FileSystemOID() types.OidT

	// ObjectID returns the object identifier
	ObjectID() types.OidT

	// TransactionID returns the transaction identifier
	TransactionID() types.XidT
}

// ReaperListReader provides methods for reading a reaper list
type ReaperListReader interface {
	// NextListOID returns the object identifier of the next reaper list
	NextListOID() types.OidT

	// Flags returns the flags for this reaper list
	Flags() uint32

	// MaxEntries returns the maximum number of entries in this list
	MaxEntries() uint32

	// CurrentEntryCount returns the number of entries currently in the list
	CurrentEntryCount() uint32

	// Entries returns the list of reaper list entries
	Entries() []ReaperListEntryReader
}

// ReaperListEntryReader provides methods for reading a reaper list entry
type ReaperListEntryReader interface {
	// NextEntryIndex returns the index of the next entry
	NextEntryIndex() uint32

	// Flags returns the flags for this entry
	Flags() uint32

	// IsValid() checks if the entry is valid
	IsValid() bool

	// IsReapIDRecord checks if this is a reap ID record
	IsReapIDRecord() bool

	// IsReadyToCall checks if the entry is ready to be called
	IsReadyToCall() bool

	// IsCompletionEntry checks if this is a completion entry
	IsCompletionEntry() bool

	// IsCleanupEntry checks if this is a cleanup entry
	IsCleanupEntry() bool
}

// ReaperPhaseManager provides methods for managing reaper phases
type ReaperPhaseManager interface {
	// CurrentPhase returns the current reaping phase
	CurrentPhase() uint32

	// PhaseDescription returns a human-readable description of the current phase
	PhaseDescription() string

	// IsPhaseComplete checks if the current phase is complete
	IsPhaseComplete() bool
}

// ObjectMapReaperState provides state information for object map reaping
type ObjectMapReaperState interface {
	// ReapingPhase returns the current reaping phase
	ReapingPhase() uint32

	// LastProcessedKey returns the key of the most recently freed entry
	LastProcessedKey() types.OmapKeyT
}

// SnapshotCleanupState provides state information for snapshot cleanup
type SnapshotCleanupState interface {
	// IsCleaning checks if the cleanup process is active
	IsCleaning() bool

	// SnapshotFlags returns the flags for the snapshots being deleted
	SnapshotFlags() uint32

	// PreviousSnapshotXID returns the transaction ID of the snapshot before deletion
	PreviousSnapshotXID() types.XidT

	// StartSnapshotXID returns the transaction ID of the first snapshot being deleted
	StartSnapshotXID() types.XidT

	// EndSnapshotXID returns the transaction ID of the last snapshot being deleted
	EndSnapshotXID() types.XidT

	// NextSnapshotXID returns the transaction ID of the snapshot after deletion
	NextSnapshotXID() types.XidT
}

// ReaperStateReader provides a comprehensive view of the reaper's current state
type ReaperStateReader interface {
	// LastProcessedBlockNumber returns the last physical block number processed
	LastProcessedBlockNumber() uint64

	// CurrentSnapshotXID returns the current snapshot's transaction identifier
	CurrentSnapshotXID() types.XidT

	// PhaseDescription returns a human-readable description of the current reaping phase
	PhaseDescription() string
}
