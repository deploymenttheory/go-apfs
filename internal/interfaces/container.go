// File: internal/interfaces/container.go
package interfaces

import (
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// ContainerSuperblockReader provides methods for reading the container superblock information
type ContainerSuperblockReader interface {
	// Magic returns the magic number for validating the container superblock
	Magic() uint32

	// BlockSize returns the logical block size used in the container
	BlockSize() uint32

	// BlockCount returns the total number of logical blocks available in the container
	BlockCount() uint64

	// UUID returns the universally unique identifier of the container
	UUID() types.UUID

	// NextObjectID returns the next object identifier to be used for new ephemeral or virtual objects
	NextObjectID() types.OidT

	// NextTransactionID returns the next transaction to be used
	NextTransactionID() types.XidT

	// SpaceManagerOID returns the ephemeral object identifier for the space manager
	SpaceManagerOID() types.OidT

	// ObjectMapOID returns the physical object identifier for the container's object map
	ObjectMapOID() types.OidT

	// ReaperOID returns the ephemeral object identifier for the reaper
	ReaperOID() types.OidT

	// MaxFileSystems returns the maximum number of volumes that can be stored in this container
	MaxFileSystems() uint32

	// VolumeOIDs returns the array of virtual object identifiers for volumes
	VolumeOIDs() []types.OidT

	// EFIJumpstart returns the physical object identifier of the object that contains EFI driver data
	EFIJumpstart() types.Paddr

	// FusionUUID returns the UUID of the container's Fusion set
	FusionUUID() types.UUID

	// KeylockerLocation returns the location of the container's keybag
	KeylockerLocation() types.Prange

	// MediaKeyLocation returns the wrapped media key location
	MediaKeyLocation() types.Prange

	// BlockedOutRange returns the blocked-out physical address range
	BlockedOutRange() types.Prange

	// EvictMappingTreeOID returns the object identifier of the evict-mapping tree
	EvictMappingTreeOID() types.OidT

	// TestType returns the container's test type for debugging
	TestType() uint32

	// TestOID returns the test object identifier for debugging
	TestOID() types.OidT

	// NewestMountedVersion returns the newest version of APFS that has mounted this container
	NewestMountedVersion() uint64
}

// ContainerFeatureManager provides methods for managing container features
type ContainerFeatureManager interface {
	// Features returns the optional features being used by the container
	Features() uint64

	// ReadOnlyCompatibleFeatures returns the read-only compatible features being used
	ReadOnlyCompatibleFeatures() uint64

	// IncompatibleFeatures returns the backward-incompatible features being used
	IncompatibleFeatures() uint64

	// HasFeature checks if a specific optional feature is enabled
	HasFeature(feature uint64) bool

	// HasReadOnlyCompatibleFeature checks if a specific read-only compatible feature is enabled
	HasReadOnlyCompatibleFeature(feature uint64) bool

	// HasIncompatibleFeature checks if a specific incompatible feature is enabled
	HasIncompatibleFeature(feature uint64) bool

	// SupportsDefragmentation checks if the container supports defragmentation
	SupportsDefragmentation() bool

	// IsLowCapacityFusionDrive checks if the container is using low-capacity Fusion Drive mode
	IsLowCapacityFusionDrive() bool

	// GetAPFSVersion returns the APFS version used by the container
	GetAPFSVersion() string

	// SupportsFusion checks if the container supports Fusion Drives
	SupportsFusion() bool
}

// ContainerFlagManager provides methods for working with container flags
type ContainerFlagManager interface {
	// Flags returns the container's flags
	Flags() uint64

	// HasFlag checks if a specific flag is set
	HasFlag(flag uint64) bool

	// UsesSoftwareCryptography checks if the container uses software cryptography
	UsesSoftwareCryptography() bool
}

// ContainerCheckpointManager provides methods for managing container checkpoints
type ContainerCheckpointManager interface {
	// CheckpointDescriptorBlockCount returns the number of blocks used by the checkpoint descriptor area
	CheckpointDescriptorBlockCount() uint32

	// CheckpointDataBlockCount returns the number of blocks used by the checkpoint data area
	CheckpointDataBlockCount() uint32

	// CheckpointDescriptorBase returns the base address of the checkpoint descriptor area
	CheckpointDescriptorBase() types.Paddr

	// CheckpointDataBase returns the base address of the checkpoint data area
	CheckpointDataBase() types.Paddr

	// CheckpointDescriptorNext returns the next index to use in the checkpoint descriptor area
	CheckpointDescriptorNext() uint32

	// CheckpointDataNext returns the next index to use in the checkpoint data area
	CheckpointDataNext() uint32

	// CheckpointDescriptorIndex returns the index of the first valid item in the checkpoint descriptor area
	CheckpointDescriptorIndex() uint32

	// CheckpointDescriptorLength returns the number of blocks in the checkpoint descriptor area used by the current checkpoint
	CheckpointDescriptorLength() uint32

	// CheckpointDataIndex returns the index of the first valid item in the checkpoint data area
	CheckpointDataIndex() uint32

	// CheckpointDataLength returns the number of blocks in the checkpoint data area used by the current checkpoint
	CheckpointDataLength() uint32
}

// CheckpointMappingReader provides methods for reading checkpoint mappings
type CheckpointMappingReader interface {
	// Type returns the object's type
	Type() uint32

	// Subtype returns the object's subtype
	Subtype() uint32

	// Size returns the size of the object in bytes
	Size() uint32

	// FilesystemOID returns the virtual object identifier of the volume
	FilesystemOID() types.OidT

	// ObjectID returns the ephemeral object identifier
	ObjectID() types.OidT

	// PhysicalAddress returns the address in the checkpoint data area where the object is stored
	PhysicalAddress() types.Paddr
}

// CheckpointMapReader provides methods for reading checkpoint maps
type CheckpointMapReader interface {
	// Flags returns the checkpoint map flags
	Flags() uint32

	// Count returns the number of checkpoint mappings in the array
	Count() uint32

	// Mappings returns the array of checkpoint mappings
	Mappings() []CheckpointMappingReader

	// IsLast checks if this is the last checkpoint-mapping block in a given checkpoint
	IsLast() bool
}

// EvictMappingReader provides methods for reading evict mappings
type EvictMappingReader interface {
	// DestinationAddress returns the address where the destination starts
	DestinationAddress() types.Paddr

	// Length returns the number of blocks being moved
	Length() uint64
}

// ContainerStatisticsReader provides methods for reading container-level statistics
type ContainerStatisticsReader interface {
	// Counters returns the array of counters that store information about the container
	Counters() []uint64

	// ObjectChecksumSetCount returns the number of times a checksum has been computed while writing objects to disk
	ObjectChecksumSetCount() uint64

	// ObjectChecksumFailCount returns the number of times an object's checksum was invalid when reading from disk
	ObjectChecksumFailCount() uint64
}

// ContainerEphemeralManager provides methods for managing ephemeral data
type ContainerEphemeralManager interface {
	// EphemeralInfo returns the array of fields used in management of ephemeral data
	EphemeralInfo() []uint64

	// MinimumBlockCount returns the default minimum size in blocks for structures containing ephemeral data
	MinimumBlockCount() uint32

	// MaxEphemeralStructures returns the number of structures containing ephemeral data that a volume can have
	MaxEphemeralStructures() uint32

	// EphemeralInfoVersion returns the version number for structures containing ephemeral data
	EphemeralInfoVersion() uint32
}

// ContainerManager provides high-level methods for managing APFS containers
type ContainerManager interface {
	// Volume Discovery and Management
	ListVolumes() ([]Volume, error)
	FindVolumeByName(name string) (Volume, error)
	FindVolumeByUUID(uuid types.UUID) (Volume, error)
	FindVolumesByRole(role uint16) ([]Volume, error)

	// Container Space Management
	TotalSize() uint64
	FreeSpace() uint64
	UsedSpace() uint64
	SpaceUtilization() float64

	// Blocked and Reserved Space
	BlockedOutRange() types.Prange
	EvictMappingTreeOID() types.OidT

	// Container Metadata
	UUID() types.UUID
	NextObjectID() types.OidT
	NextTransactionID() types.XidT

	// Features and Compatibility
	Features() uint64
	IncompatibleFeatures() uint64
	ReadonlyCompatibleFeatures() uint64

	// Encryption and Security
	IsEncrypted() bool
	CryptoType() uint64

	// Snapshots and Versioning
	TotalSnapshots() uint64
	LatestSnapshotXID() types.XidT

	// Health and Integrity
	CheckIntegrity() (bool, []string)
	IsHealthy() bool

	// Object Map Operations
	GetObjectMap() (ObjectMapReader, error)
	ResolveVirtualObject(oid types.OidT, xid types.XidT) (types.Paddr, error)
}
