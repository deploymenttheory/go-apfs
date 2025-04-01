package interfaces

import (
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// VolumeReader defines operations for reading and interacting with an APFS volume
type VolumeReader interface {
	// Name returns the volume name as a string
	Name() string

	// IsEncrypted checks if the volume is encrypted
	IsEncrypted() bool

	// RoleName returns a human-readable name for the volume's role
	RoleName() string

	// HasFeature checks if a specific optional feature is enabled
	HasFeature(flag uint64) bool

	// HasIncompatibleFeature checks if a specific incompatibility flag is enabled
	HasIncompatibleFeature(flag uint64) bool

	// Superblock returns the underlying superblock
	Superblock() *types.ApfsSuperblockT

	// UUID returns the volume's unique identifier
	UUID() types.UUID

	// Role returns the raw role value
	Role() uint16

	// Flags returns the volume's flags
	Flags() uint64
}

// VolumeInspector provides methods for volume discovery and inspection
type VolumeInspector interface {
	// ListVolumes returns all volumes
	ListVolumes() ([]VolumeReader, error)

	// FindVolumeByName finds a volume with a specific name
	FindVolumeByName(name string) (VolumeReader, error)

	// FindVolumeByRole finds volumes with a specific role
	FindVolumesByRole(role uint16) ([]VolumeReader, error)

	// FindEncryptedVolumes returns all encrypted volumes
	FindEncryptedVolumes() ([]VolumeReader, error)

	// FindVolumeByUUID finds a volume with a specific UUID
	FindVolumeByUUID(uuid types.UUID) (VolumeReader, error)
}

// VolumeEncryptionInfo provides encryption-related volume information
type VolumeEncryptionInfo interface {
	// IsEncrypted checks if the volume is encrypted
	IsEncrypted() bool

	// EncryptionType returns the type of encryption
	EncryptionType() uint64

	// KeybagLocation returns the location of the volume's keybag
	KeybagLocation() types.Prange
}

// VolumeSnapshotManager provides snapshot-related operations
type VolumeSnapshotManager interface {
	// ListSnapshots returns all snapshots for the volume
	ListSnapshots() ([]SnapshotInfo, error)

	// GetSnapshot retrieves a specific snapshot by transaction ID
	GetSnapshot(xid types.XidT) (SnapshotInfo, error)
}

// SnapshotInfo represents metadata for a volume snapshot
type SnapshotInfo struct {
	// Transaction ID of the snapshot
	XID types.XidT

	// Creation time of the snapshot
	CreateTime uint64

	// Last modification time of the snapshot
	ChangeTime uint64

	// Name of the snapshot
	Name string
}

// VolumeStatistics provides volume-level metrics
type VolumeStatistics interface {
	// TotalFiles returns the number of files in the volume
	TotalFiles() uint64

	// TotalDirectories returns the number of directories
	TotalDirectories() uint64

	// TotalSymlinks returns the number of symbolic links
	TotalSymlinks() uint64

	// TotalSnapshots returns the number of snapshots
	TotalSnapshots() uint64

	// AllocatedBlocks returns the total number of allocated blocks
	TotalAllocatedBlocks() uint64

	// TotalBlocksFree returns the total number of free unallocated blocks
	TotalBlocksFree() uint64
}
