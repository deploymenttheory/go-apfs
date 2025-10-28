// File: internal/interfaces/mounting.go
package interfaces

import (
	"time"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// APFSMounter provides methods for mounting APFS containers and volumes
type APFSMounter interface {
	// MountContainer mounts an APFS container from a device or image
	MountContainer(devicePath string) (ContainerManager, error)

	// MountContainerWithOptions mounts a container with specific options
	MountContainerWithOptions(devicePath string, options MountOptions) (ContainerManager, error)

	// MountVolume mounts a specific volume within a container
	MountVolume(container ContainerManager, volumeIndex uint32) (Volume, error)

	// MountVolumeByName mounts a volume by its name
	MountVolumeByName(container ContainerManager, volumeName string) (Volume, error)

	// MountVolumeByUUID mounts a volume by its UUID
	MountVolumeByUUID(container ContainerManager, volumeUUID types.UUID) (Volume, error)

	// UnmountVolume unmounts a specific volume
	UnmountVolume(volume Volume) error

	// UnmountContainer unmounts a container and all its volumes
	UnmountContainer(container ContainerManager) error

	// ListMountedContainers returns all currently mounted containers
	ListMountedContainers() ([]ContainerManager, error)

	// ListMountedVolumes returns all currently mounted volumes
	ListMountedVolumes() ([]Volume, error)
}

// MountOptions contains options for mounting APFS containers and volumes
type MountOptions struct {
	// Whether to mount in read-only mode
	ReadOnly bool

	// Whether to verify checksums during mounting
	VerifyChecksums bool

	// Whether to mount volumes automatically
	AutoMountVolumes bool

	// Maximum number of volumes to auto-mount (0 = no limit)
	MaxAutoMountVolumes uint32

	// Whether to use block caching
	EnableCaching bool

	// Cache size in megabytes (0 = default)
	CacheSizeMB uint32

	// Whether to perform integrity checks during mount
	PerformIntegrityChecks bool

	// Timeout for mount operations
	MountTimeout time.Duration

	// Custom block device to use instead of opening from path
	BlockDevice BlockDevice

	// Whether to mount specific snapshots
	MountSnapshot bool

	// Transaction ID of snapshot to mount (if MountSnapshot is true)
	SnapshotTransactionID types.XidT
}

// APFSImageReader provides methods for reading APFS disk images
type APFSImageReader interface {
	// OpenImage opens an APFS disk image file
	OpenImage(imagePath string) (APFSImage, error)

	// OpenImageWithPassword opens an encrypted APFS disk image
	OpenImageWithPassword(imagePath string, password string) (APFSImage, error)

	// CreateImageFromDevice creates a disk image from a physical device
	CreateImageFromDevice(devicePath string, imagePath string, options ImageCreationOptions) error

	// ValidateImage validates the integrity of a disk image
	ValidateImage(imagePath string) (ImageValidationResult, error)

	// GetImageInfo returns information about a disk image without fully opening it
	GetImageInfo(imagePath string) (ImageInfo, error)
}

// APFSImage represents an APFS disk image
type APFSImage interface {
	BlockDevice

	// ImagePath returns the path to the image file
	ImagePath() string

	// IsEncrypted checks if the image is encrypted
	IsEncrypted() bool

	// ImageFormat returns the format of the image (e.g., "raw", "dmg", "sparse")
	ImageFormat() string

	// CreationTime returns when the image was created
	CreationTime() time.Time

	// ImageSize returns the logical size of the image
	ImageSize() uint64

	// ActualSize returns the actual size of the image file on disk
	ActualSize() uint64

	// CompressionRatio returns the compression ratio (actual/logical)
	CompressionRatio() float64
}

// ImageCreationOptions contains options for creating disk images
type ImageCreationOptions struct {
	// Whether to create a sparse image
	Sparse bool

	// Whether to compress the image
	Compress bool

	// Whether to encrypt the image
	Encrypt bool

	// Password for encryption (if Encrypt is true)
	Password string

	// Format of the image to create
	Format string

	// Whether to verify the copy
	Verify bool

	// Progress callback for long operations
	ProgressCallback func(bytesProcessed, totalBytes uint64)
}

// ImageValidationResult contains the result of validating a disk image
type ImageValidationResult struct {
	// Whether the image is valid
	IsValid bool

	// Issues found during validation
	Issues []ImageValidationIssue

	// Checksum validation results
	ChecksumValid bool

	// Whether the image format is supported
	FormatSupported bool

	// Image metadata
	Metadata ImageInfo
}

// ImageValidationIssue represents an issue found during image validation
type ImageValidationIssue struct {
	// Type of issue
	Type ImageValidationIssueType

	// Severity of the issue
	Severity ImageValidationIssueSeverity

	// Description of the issue
	Description string

	// Byte offset where the issue was found (if applicable)
	ByteOffset uint64

	// Additional details
	Details map[string]any
}

// ImageValidationIssueType represents the type of image validation issue
type ImageValidationIssueType int

const (
	ImageValidationIssueCorruptedHeader ImageValidationIssueType = iota
	ImageValidationIssueInvalidChecksum
	ImageValidationIssueUnsupportedFormat
	ImageValidationIssueInconsistentSize
	ImageValidationIssueMissingData
	ImageValidationIssueEncryptionError
)

// ImageValidationIssueSeverity represents the severity of an image validation issue
type ImageValidationIssueSeverity int

const (
	ImageValidationIssueSeverityInfo ImageValidationIssueSeverity = iota
	ImageValidationIssueSeverityWarning
	ImageValidationIssueSeverityError
	ImageValidationIssueSeverityCritical
)

// ImageInfo contains information about a disk image
type ImageInfo struct {
	// Path to the image file
	Path string

	// Format of the image
	Format string

	// Logical size of the image
	LogicalSize uint64

	// Actual size of the image file
	ActualSize uint64

	// Whether the image is encrypted
	IsEncrypted bool

	// Whether the image is sparse
	IsSparse bool

	// Whether the image is compressed
	IsCompressed bool

	// Creation time of the image
	CreationTime time.Time

	// Modification time of the image
	ModificationTime time.Time

	// Checksum of the image (if available)
	Checksum string

	// Checksum algorithm used
	ChecksumAlgorithm string
}

// ContainerMountState provides information about the mount state of a container
type ContainerMountState interface {
	// MountTime returns when the container was mounted
	MountTime() time.Time

	// MountOptions returns the options used when mounting
	MountOptions() MountOptions

	// SourcePath returns the path to the source device or image
	SourcePath() string

	// IsMounted checks if the container is currently mounted
	IsMounted() bool

	// MountedVolumeCount returns the number of volumes currently mounted
	MountedVolumeCount() int

	// GetMountedVolumes returns all mounted volumes in this container
	GetMountedVolumes() ([]Volume, error)
}

// VolumeMountState provides information about the mount state of a volume
type VolumeMountState interface {
	// MountTime returns when the volume was mounted
	MountTime() time.Time

	// Container returns the container this volume belongs to
	Container() ContainerManager

	// IsMounted checks if the volume is currently mounted
	IsMounted() bool

	// IsReadOnly checks if the volume was mounted read-only
	IsReadOnly() bool

	// SnapshotMounted returns the snapshot transaction ID if mounted from a snapshot
	SnapshotMounted() (types.XidT, bool)
}

// MountValidator provides methods for validating mount operations
type MountValidator interface {
	// CanMountContainer checks if a container can be mounted
	CanMountContainer(devicePath string) (bool, []string, error)

	// CanMountVolume checks if a volume can be mounted
	CanMountVolume(container ContainerManager, volumeIndex uint32) (bool, []string, error)

	// ValidateMountPrerequisites checks system prerequisites for mounting
	ValidateMountPrerequisites() (bool, []string, error)

	// CheckPermissions verifies the necessary permissions for mounting
	CheckPermissions(devicePath string) (bool, error)
}

// MountManager provides comprehensive mount management functionality
type MountManager interface {
	APFSMounter
	APFSImageReader
	MountValidator

	// GetContainerMountState returns mount state information for a container
	GetContainerMountState(container ContainerManager) (ContainerMountState, error)

	// GetVolumeMountState returns mount state information for a volume
	GetVolumeMountState(volume Volume) (VolumeMountState, error)

	// RefreshMountState refreshes the cached mount state information
	RefreshMountState() error

	// SetMountEventCallback sets a callback for mount/unmount events
	SetMountEventCallback(callback MountEventCallback)
}

// MountEventCallback is called when mount/unmount events occur
type MountEventCallback func(event MountEvent)

// MountEvent represents a mount or unmount event
type MountEvent struct {
	// Type of event
	Type MountEventType

	// Time when the event occurred
	Timestamp time.Time

	// Container involved in the event (if applicable)
	Container ContainerManager

	// Volume involved in the event (if applicable)
	Volume Volume

	// Source path of the mount
	SourcePath string

	// Error that occurred (if any)
	Error error
}

// MountEventType represents the type of mount event
type MountEventType int

const (
	MountEventContainerMounted MountEventType = iota
	MountEventContainerUnmounted
	MountEventVolumeMounted
	MountEventVolumeUnmounted
	MountEventMountError
)

// PasswordProvider provides methods for obtaining passwords for encrypted volumes
type PasswordProvider interface {
	// GetPassword requests a password for an encrypted volume
	GetPassword(volumeUUID types.UUID, volumeName string) (string, error)

	// GetPasswordWithHint requests a password with a hint
	GetPasswordWithHint(volumeUUID types.UUID, volumeName string, hint string) (string, error)

	// CachePassword caches a password for future use
	CachePassword(volumeUUID types.UUID, password string) error

	// ClearPasswordCache clears all cached passwords
	ClearPasswordCache() error
}
