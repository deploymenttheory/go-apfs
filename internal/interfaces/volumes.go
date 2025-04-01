// File: internal/interfaces/volumes.go
package interfaces

import (
	"time"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// VolumeIdentity provides core volume identification details
type VolumeIdentity interface {
	// Unique volume identifier
	UUID() types.UUID

	// Volume name
	Name() string

	// Volume role identifier
	Role() uint16

	// Human-readable role description
	RoleName() string

	// Index in container's filesystem array
	Index() uint32
}

// VolumeFeatures provides information about volume capabilities and flags
type VolumeFeatures interface {
	// Optional feature flags
	Features() uint64

	// Read-only compatible feature flags
	ReadonlyCompatibleFeatures() uint64

	// Incompatible feature flags
	IncompatibleFeatures() uint64

	// Specific feature checks
	SupportsDefragmentation() bool
	SupportsHardlinkMapRecords() bool
	IsStrictAccessTimeEnabled() bool
	IsCaseInsensitive() bool
	IsNormalizationInsensitive() bool
	IsSealed() bool

	// Specific flag checks
	IsUnencrypted() bool
	IsOneKeyEncryption() bool
	IsSpilledOver() bool
	RequiresSpilloverCleaner() bool
	AlwaysChecksExtentReference() bool
}

// VolumeSpaceManagement provides space allocation and quota information
type VolumeSpaceManagement interface {
	// Block allocation details
	ReservedBlockCount() uint64
	QuotaBlockCount() uint64
	AllocatedBlockCount() uint64

	// Allocation tracking
	TotalBlocksAllocated() uint64
	TotalBlocksFreed() uint64

	// Space utilization
	SpaceUtilization() float64
}

// VolumeTreeStructure provides object identifiers for key filesystem trees
type VolumeTreeStructure interface {
	// Root filesystem tree details
	RootTreeOID() types.OidT
	RootTreeType() uint32

	// Extent reference tree details
	ExtentReferenceTreeOID() types.OidT
	ExtentReferenceTreeType() uint32

	// Snapshot metadata tree details
	SnapshotMetadataTreeOID() types.OidT
	SnapshotMetadataTreeType() uint32

	// Object map identifier
	ObjectMapOID() types.OidT
}

// VolumeMetadata provides additional volume-level metadata
type VolumeMetadata interface {
	// Timestamp information
	LastUnmountTime() time.Time
	LastModifiedTime() time.Time

	// Modification tracking
	FormattedBy() types.ApfsModifiedByT
	ModificationHistory() []types.ApfsModifiedByT

	// Identifier tracking
	NextObjectID() uint64
	NextDocumentID() uint32
}

// VolumeEncryptionMetadata provides encryption-related volume information
type VolumeEncryptionMetadata interface {
	// Encryption state details
	MetadataCryptoState() types.WrappedMetaCryptoStateT

	// Encryption status checks
	IsEncrypted() bool
	HasEncryptionKeyRotated() bool
}

// VolumeSnapshotMetadata provides snapshot-related information
type VolumeSnapshotMetadata interface {
	// Snapshot details
	TotalSnapshots() uint64

	// Snapshot reversion information
	RevertToSnapshotXID() types.XidT
	RevertToSuperblockOID() types.OidT

	// Snapshot tree details
	RootToSnapshotXID() types.XidT
}

// VolumeResourceCounts provides counts of filesystem objects
type VolumeResourceCounts interface {
	// Filesystem object counts
	TotalFiles() uint64
	TotalDirectories() uint64
	TotalSymlinks() uint64
	TotalOtherFileSystemObjects() uint64
}

// VolumeGroupInfo provides volume group information
type VolumeGroupInfo interface {
	// Volume group details
	VolumeGroupID() types.UUID

	// Metadata object identifiers
	IntegrityMetadataOID() types.OidT
}

// VolumeIntegrityCheck provides methods for verifying volume integrity
type VolumeIntegrityCheck interface {
	// Magic number validation
	ValidateMagicNumber() bool
	MagicNumber() uint32
}

// VolumeEncryptionRollingState provides information about ongoing encryption changes
type VolumeEncryptionRollingState interface {
	// Encryption rolling state object identifier
	EncryptionRollingStateOID() types.OidT

	// Check if encryption rolling is in progress
	IsEncryptionRollingInProgress() bool
}

// VolumeCloneInfo provides information about cloning operations
type VolumeCloneInfo interface {
	// Cloning information
	CloneInfoIdEpoch() uint64
	CloneInfoXID() uint64
}

// VolumeExtendedMetadata provides access to extended metadata details
type VolumeExtendedMetadata interface {
	// Extended metadata object identifiers
	SnapshotMetadataExtOID() types.OidT
	FileExtentTreeOID() types.OidT
	FileExtentTreeType() uint32
}

// Comprehensive Volume Interface
type Volume interface {
	VolumeIdentity
	VolumeFeatures
	VolumeSpaceManagement
	VolumeTreeStructure
	VolumeMetadata
	VolumeEncryptionMetadata
	VolumeSnapshotMetadata
	VolumeResourceCounts
	VolumeGroupInfo
	VolumeIntegrityCheck
	VolumeEncryptionRollingState
	VolumeCloneInfo
	VolumeExtendedMetadata
}
