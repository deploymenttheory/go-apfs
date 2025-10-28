package services

import (
	"time"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// FileNode represents a file or directory in the filesystem
type FileNode struct {
	Inode         uint64
	Path          string
	Name          string
	Mode          uint16
	Size          uint64
	CreatedTime   time.Time
	ModifiedTime  time.Time
	ChangedTime   time.Time
	AccessedTime  time.Time
	UID           uint32
	GID           uint32
	IsDirectory   bool
	IsSymlink     bool
	IsEncrypted   bool
	ParentInode   uint64
	HardLinkCount uint32
	Flags         uint32
}

// ExtentMapping describes file data to physical blocks mapping
type ExtentMapping struct {
	LogicalOffset   uint64
	LogicalSize     uint64
	PhysicalBlock   uint64
	PhysicalSize    uint64
	IsCompressed    bool
	CompressionType string
	IsEncrypted     bool
}

// SnapshotInfo contains metadata about a snapshot
type SnapshotInfo struct {
	XID         uint64
	Name        string
	CreatedTime time.Time
	RootInode   uint64
	Size        uint64
	FileCount   uint64
	IsDataless  bool
	ParentXID   uint64
	UUID        [16]byte
}

// SpaceStats contains filesystem space usage statistics
type SpaceStats struct {
	TotalCapacity       uint64
	UsedSpace           uint64
	FreeSpace           uint64
	SnapshotSpace       uint64
	MetadataSpace       uint64
	AllocationBlockSize uint32
	UsagePercentage     float64
	FreeBlocks          uint64
	AllocatedBlocks     uint64
	FragmentationRatio  float64
}

// FileChange describes a change to a file between snapshots
type FileChange struct {
	Inode       uint64
	Path        string
	ChangeType  string
	OldMetadata *FileNode
	NewMetadata *FileNode
	OldSize     uint64
	NewSize     uint64
	OldModTime  time.Time
	NewModTime  time.Time
}

// RecoverableFile describes a file that can potentially be recovered
type RecoverableFile struct {
	Inode                   uint64
	OriginalName            string
	Size                    uint64
	RecoveryProbability     float64
	Extents                 []ExtentMapping
	Mode                    uint16
	ApproximateDeletionTime time.Time
}

// RecoveryReport contains statistics about data recovery potential
type RecoveryReport struct {
	RecoverableFileCount  int
	RecoverableDataSize   uint64
	HighConfidenceCount   int
	MediumConfidenceCount int
	LowConfidenceCount    int
	ScanTime              time.Duration
	Files                 []RecoverableFile
}

// VolumeCorruptionAnomaly describes a potential corruption or inconsistency
type VolumeCorruptionAnomaly struct {
	Type              string
	Severity          string
	AffectedInode     uint64
	Description       string
	RecommendedAction string
}

// VolumeReport is a comprehensive report of volume status
type VolumeReport struct {
	ContainerOID   uint64
	VolumeOID      uint64
	Name           string
	UUID           [16]byte
	SpaceStats     SpaceStats
	Anomalies      []VolumeCorruptionAnomaly
	FileCount      uint64
	DirectoryCount uint64
	SymlinkCount   uint64
	SnapshotCount  uint64
	IsEncrypted    bool
	IsSealed       bool
	GeneratedAt    time.Time
}

// EncryptionState describes the encryption status of the volume
type EncryptionState struct {
	IsEncrypted           bool
	ProtectionClass       string
	KeyRollingStatus      string
	KeyCount              int
	KeyRollingProgress    uint32
	LastKeyRotation       time.Time
	PerFileEncryptedCount uint64
	ValidationResult      string
	ValidationErrors      []string
}

// ObjectReference describes a reference from one object to another
type ObjectReference struct {
	FromOID        types.OidT
	ToOID          types.OidT
	ReferenceType  string
	ReferenceCount uint32
}

// ReferenceGraph shows how objects reference each other
type ReferenceGraph struct {
	Objects            map[types.OidT]string
	References         []ObjectReference
	OrphanedObjects    []types.OidT
	CircularReferences [][]types.OidT
}

// DiffReport shows differences between two snapshots
type DiffReport struct {
	Snapshot1XID  uint64
	Snapshot2XID  uint64
	Changes       []FileChange
	AddedFiles    int
	DeletedFiles  int
	ModifiedFiles int
	RenamedFiles  int
	DataAdded     uint64
	DataRemoved   uint64
}

// FileEntry is exported in filesystem_service.go

// FileReaderAdapter describes an adapter for reading file data
type FileReaderAdapter struct {
	fs      *FileSystemServiceImpl
	inodeID uint64
	size    uint64
	offset  uint64
}

// FileSeekerAdapter describes an adapter for seeking within a file
type FileSeekerAdapter struct {
	fs      *FileSystemServiceImpl
	inodeID uint64
	size    uint64
	offset  uint64
}
