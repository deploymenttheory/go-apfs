package interfaces

import (
	"time"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// SnapshotReader provides basic information about a snapshot
type SnapshotReader interface {
	// Name returns the snapshot's name
	Name() string

	// CreationTime returns the time the snapshot was created
	CreationTime() time.Time

	// LastModifiedTime returns the time the snapshot was last modified
	LastModifiedTime() time.Time

	// TransactionID returns the snapshot's transaction identifier
	TransactionID() types.XidT

	// UUID returns the snapshot's unique identifier
	UUID() types.UUID
}

// SnapshotMetadataReader provides detailed metadata about a snapshot
type SnapshotMetadataReader interface {
	// ExtentReferenceTreeOID returns the physical object identifier of the extent reference tree
	ExtentReferenceTreeOID() types.OidT

	// SuperblockOID returns the physical object identifier of the volume superblock
	SuperblockOID() types.OidT

	// ExtentReferenceTreeType returns the type of the extent reference tree
	ExtentReferenceTreeType() uint32

	// Flags returns the snapshot metadata flags
	Flags() types.SnapMetaFlags

	// IsPendingDataless checks if the snapshot is pending being made dataless
	IsPendingDataless() bool

	// IsMergeInProgress checks if a merge is in progress for this snapshot
	IsMergeInProgress() bool
}

// SnapshotExtendedMetadataReader provides additional extended metadata about a snapshot
type SnapshotExtendedMetadataReader interface {
	// Version returns the version of the extended metadata structure
	Version() uint32

	// ExtendedFlags returns the extended metadata flags
	ExtendedFlags() uint32

	// Token returns the opaque metadata token
	Token() uint64
}

// SnapshotManager provides methods for managing and querying snapshots
type SnapshotManager interface {
	// ListSnapshots returns all snapshots for a volume
	ListSnapshots() ([]SnapshotReader, error)

	// FindSnapshotByName finds a snapshot by its name
	FindSnapshotByName(name string) (SnapshotReader, error)

	// FindSnapshotByUUID finds a snapshot by its UUID
	FindSnapshotByUUID(uuid types.UUID) (SnapshotReader, error)

	// FindSnapshotByTransactionID finds a snapshot by its transaction ID
	FindSnapshotByTransactionID(xid types.XidT) (SnapshotReader, error)
}

// SnapshotRestoreManager provides methods for snapshot restoration
type SnapshotRestoreManager interface {
	// Restore attempts to restore the volume to the state of this snapshot
	Restore() error

	// PreviewChanges shows what would change if the snapshot were restored
	PreviewChanges() ([]SnapshotChange, error)
}

// SnapshotChange represents a single change that would occur during snapshot restoration
type SnapshotChange struct {
	Type        ChangeType
	Path        string
	Size        int64
	Permissions string
}

// ChangeType represents the type of change in a snapshot
type ChangeType int

const (
	ChangeTypeAdded ChangeType = iota
	ChangeTypeModified
	ChangeTypeDeleted
)
