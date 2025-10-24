// File: internal/interfaces/snapshot.go
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

// SnapMetadataReader provides methods for reading snapshot metadata records.
type SnapMetadataReader interface {
	// ExtentRefTreeOID returns the physical object identifier of the B-tree that stores extents information.
	ExtentRefTreeOID() uint64

	// SuperblockOID returns the physical object identifier of the volume superblock.
	SuperblockOID() uint64

	// CreateTime returns the time the snapshot was created.
	CreateTime() time.Time

	// ChangeTime returns the last time the snapshot was modified.
	ChangeTime() time.Time

	// InodeNumber returns the inode number associated with the snapshot.
	InodeNumber() uint64

	// ExtentRefTreeType returns the type of the B-tree that stores extents information.
	ExtentRefTreeType() uint32

	// Flags returns the snapshot metadata flags.
	Flags() uint32

	// Name returns the snapshot's name.
	Name() string

	// HasFlag checks if a specific flag is set.
	HasFlag(flag uint32) bool
}

// SnapNameReader provides methods for reading snapshot name records.
type SnapNameReader interface {
	// Name returns the snapshot's name.
	Name() string

	// SnapXID returns the last transaction identifier included in the snapshot.
	SnapXID() uint64
}

// SnapMetaExtReader provides methods for reading extended snapshot metadata.
type SnapMetaExtReader interface {
	// Version returns the version of the extended metadata structure.
	Version() uint32

	// Flags returns the extended metadata flags.
	Flags() uint32

	// SnapXID returns the snapshot's transaction identifier.
	SnapXID() uint64

	// UUID returns the snapshot's UUID.
	UUID() [16]byte

	// Token returns the opaque metadata token.
	Token() uint64
}

// SnapshotManager provides methods for managing and querying snapshot metadata.
type SnapshotManager interface {
	// GetSnapshotMetadata retrieves metadata for a snapshot by transaction ID.
	GetSnapshotMetadata(xid uint64) (SnapMetadataReader, error)

	// GetSnapshotByName finds a snapshot by name.
	GetSnapshotByName(name string) (SnapMetadataReader, error)

	// ListSnapshots returns all available snapshots.
	ListSnapshots() ([]SnapMetadataReader, error)

	// CountSnapshots returns the total number of snapshots.
	CountSnapshots() (int, error)

	// GetSnapshotExtendedMetadata retrieves extended metadata for a snapshot.
	GetSnapshotExtendedMetadata(xid uint64) (SnapMetaExtReader, error)
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

// SnapshotComparator provides methods for comparing snapshots
type SnapshotComparator interface {
	// CompareSnapshots compares two snapshots and returns the differences
	CompareSnapshots(snap1, snap2 SnapshotReader) (SnapshotDiff, error)

	// GetChangedFiles returns files that changed between two snapshots
	GetChangedFiles(snap1, snap2 SnapshotReader) ([]FileChange, error)

	// GetChangedDirectories returns directories that changed between two snapshots
	GetChangedDirectories(snap1, snap2 SnapshotReader) ([]DirectoryChange, error)

	// ComputeDelta computes the delta between two snapshots
	ComputeDelta(older, newer SnapshotReader) (SnapshotDelta, error)
}

// SnapshotDiff represents the differences between two snapshots
type SnapshotDiff struct {
	// The older snapshot being compared
	OlderSnapshot SnapshotReader

	// The newer snapshot being compared
	NewerSnapshot SnapshotReader

	// Files that were added in the newer snapshot
	AddedFiles []FileChange

	// Files that were modified between snapshots
	ModifiedFiles []FileChange

	// Files that were deleted in the newer snapshot
	DeletedFiles []FileChange

	// Directories that changed
	ChangedDirectories []DirectoryChange

	// Summary statistics
	TotalChanges int64
	BytesChanged int64
}

// FileChange represents a change to a file between snapshots
type FileChange struct {
	// The file's inode ID
	InodeID uint64

	// The file's path
	Path string

	// Type of change
	ChangeType ChangeType

	// Old file size (for modifications and deletions)
	OldSize uint64

	// New file size (for additions and modifications)
	NewSize uint64

	// Old modification time
	OldModTime time.Time

	// New modification time
	NewModTime time.Time
}

// DirectoryChange represents a change to a directory between snapshots
type DirectoryChange struct {
	// The directory's inode ID
	InodeID uint64

	// The directory's path
	Path string

	// Type of change
	ChangeType ChangeType

	// Number of files added to this directory
	FilesAdded int

	// Number of files removed from this directory
	FilesRemoved int

	// Number of files modified in this directory
	FilesModified int
}

// SnapshotDelta represents the delta between two snapshots
type SnapshotDelta struct {
	// The transaction ID range covered by this delta
	StartTransactionID types.XidT
	EndTransactionID   types.XidT

	// Changed inodes in this delta
	ChangedInodes []uint64

	// New extents allocated in this delta
	NewExtents []types.Prange

	// Freed extents in this delta
	FreedExtents []types.Prange

	// Size of the delta in bytes
	DeltaSize uint64
}

// SnapshotAnalyzer provides methods for analyzing snapshot behavior and efficiency
type SnapshotAnalyzer interface {
	// AnalyzeSnapshotEfficiency analyzes how efficiently snapshots are using space
	AnalyzeSnapshotEfficiency(snapshots []SnapshotReader) (SnapshotEfficiencyAnalysis, error)

	// GetSnapshotTimeline creates a timeline of snapshot creation and changes
	GetSnapshotTimeline() (SnapshotTimeline, error)

	// CalculateSnapshotOverhead calculates storage overhead of maintaining snapshots
	CalculateSnapshotOverhead() (SnapshotOverhead, error)
}

// SnapshotEfficiencyAnalysis contains analysis of snapshot storage efficiency
type SnapshotEfficiencyAnalysis struct {
	// Total number of snapshots analyzed
	SnapshotCount int

	// Total storage overhead from snapshots
	TotalOverhead uint64

	// Average storage overhead per snapshot
	AverageOverhead uint64

	// Most storage-efficient snapshot
	MostEfficient SnapshotReader

	// Least storage-efficient snapshot
	LeastEfficient SnapshotReader

	// Recommended snapshots to delete for space savings
	RecommendedDeletions []SnapshotReader
}

// SnapshotTimeline represents a chronological view of snapshots
type SnapshotTimeline struct {
	// Snapshots in chronological order
	Snapshots []SnapshotTimelineEntry

	// Total time span covered
	TimeSpan time.Duration

	// Average time between snapshots
	AverageInterval time.Duration
}

// SnapshotTimelineEntry represents one entry in a snapshot timeline
type SnapshotTimelineEntry struct {
	// The snapshot
	Snapshot SnapshotReader

	// Number of changes since the previous snapshot
	ChangesSincePrevious int64

	// Storage used since the previous snapshot
	StorageUsedSincePrevious uint64

	// Time since the previous snapshot
	TimeSincePrevious time.Duration
}

// SnapshotOverhead contains information about snapshot storage overhead
type SnapshotOverhead struct {
	// Total space used by all snapshots
	TotalSnapshotSpace uint64

	// Space used by the current volume state
	CurrentVolumeSpace uint64

	// Overhead percentage
	OverheadPercentage float64

	// Space that could be reclaimed by deleting all snapshots
	ReclaimableSpace uint64
}
