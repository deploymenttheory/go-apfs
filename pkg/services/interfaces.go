package services

import (
	"context"
	"io"
	"time"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// ContainerInfo represents basic container metadata
type ContainerInfo struct {
	DevicePath      string
	BlockSize       uint32
	BlockCount      uint64
	VolumeCount     uint32
	CheckpointID    uint64
	Volumes         []VolumeInfo
	SpaceManager    SpaceManagerInfo
	Features        []string
	Encrypted       bool
	CaseInsensitive bool
}

// VolumeInfo represents basic volume metadata
type VolumeInfo struct {
	ObjectID      uint64
	Name          string
	Role          string
	Reserved      uint64
	Quota         uint64
	Allocated     uint64
	FileCount     uint64
	DirCount      uint64
	SnapshotCount uint32
	Encrypted     bool
	Features      []string
	CaseSensitive bool
	LastModified  time.Time
}

// SpaceManagerInfo represents space management information
type SpaceManagerInfo struct {
	BlockSize          uint32
	ChunkCount         uint32
	FreeBlocks         uint64
	UsedBlocks         uint64
	ReservedBlocks     uint64
	FragmentationRatio float64
}

// FileInfo represents detailed file information
type FileInfo struct {
	InodeID       uint64
	Name          string
	Path          string
	Type          string
	Size          uint64
	Blocks        uint64
	Owner         uint32
	Group         uint32
	Mode          uint32
	Created       time.Time
	Modified      time.Time
	Accessed      time.Time
	Changed       time.Time
	Flags         uint64
	HardLinks     int32
	ExtendedAttrs map[string][]byte
	ResourceFork  []byte
	Compressed    bool
	Encrypted     bool
}

// DirectoryInfo represents directory information with statistics
type DirectoryInfo struct {
	FileInfo
	ChildCount uint64
	TotalSize  uint64
	Children   []FileInfo
	Recursive  bool
}

// ExtractionOptions configures extraction behavior
type ExtractionOptions struct {
	PreserveMetadata    bool
	PreservePermissions bool
	PreserveTimestamps  bool
	PreserveExtAttrs    bool
	OverwriteExisting   bool
	CreateDirectories   bool
	VerifyIntegrity     bool
	MaxDepth            int
	IncludeHidden       bool
	FollowSymlinks      bool
}

// AnalysisResult represents filesystem analysis results
type AnalysisResult struct {
	ContainerInfo    ContainerInfo
	VolumeAnalysis   []VolumeAnalysis
	StructuralHealth StructuralHealth
	EncryptionInfo   EncryptionInfo
	Performance      PerformanceMetrics
	Recommendations  []string
	Warnings         []string
	Errors           []string
	GeneratedAt      time.Time
}

// VolumeAnalysis represents analysis of a single volume
type VolumeAnalysis struct {
	VolumeInfo
	BTreeStats      BTreeStats
	ObjectStats     ObjectStats
	FileSystemStats FileSystemStats
	IntegrityStatus IntegrityStatus
}

// BTreeStats represents B-tree analysis
type BTreeStats struct {
	NodeCount     uint64
	LeafNodes     uint64
	InternalNodes uint64
	Height        int
	KeyCount      uint64
	Fragmentation float64
	HealthScore   float64
}

// ObjectStats represents object-level statistics
type ObjectStats struct {
	TotalObjects     uint64
	InodeObjects     uint64
	ExtentObjects    uint64
	XattrObjects     uint64
	DirRecObjects    uint64
	SiblingObjects   uint64
	SnapshotObjects  uint64
	CorruptedObjects uint64
}

// FileSystemStats represents filesystem-level statistics
type FileSystemStats struct {
	TotalFiles       uint64
	TotalDirectories uint64
	TotalSymlinks    uint64
	TotalSize        uint64
	UsedSpace        uint64
	FreeSpace        uint64
	FragmentedFiles  uint64
	CompressedFiles  uint64
	EncryptedFiles   uint64
}

// IntegrityStatus represents integrity check results
type IntegrityStatus struct {
	ChecksumErrors    uint64
	StructuralErrors  uint64
	OrphanedObjects   uint64
	HealthScore       float64
	LastChecked       time.Time
	RecommendedAction string
}

// StructuralHealth represents overall structural health
type StructuralHealth struct {
	OverallScore    float64
	ContainerHealth float64
	VolumeHealth    []float64
	CriticalIssues  []string
	WarningIssues   []string
	LastValidated   time.Time
}

// EncryptionInfo represents encryption analysis
type EncryptionInfo struct {
	ContainerEncrypted bool
	EncryptionMethod   string
	ProtectionClasses  []string
	KeybagInfo         KeybagInfo
	VolumeEncryption   []VolumeEncryption
}

// KeybagInfo represents keybag analysis
type KeybagInfo struct {
	Type          string
	Version       string
	KeyCount      int
	Locked        bool
	HardwareBound bool
}

// VolumeEncryption represents per-volume encryption info
type VolumeEncryption struct {
	VolumeID          uint64
	VolumeName        string
	EncryptionEnabled bool
	ProtectionClass   string
	FileVaultEnabled  bool
	EncryptedFiles    uint64
	UnencryptedFiles  uint64
}

// PerformanceMetrics represents performance analysis
type PerformanceMetrics struct {
	AvgSeekTime         time.Duration
	AvgReadTime         time.Duration
	ThroughputMBps      float64
	IOPSCapability      uint64
	FragmentationImpact float64
	OptimizationScore   float64
}

// ContainerService provides container-level operations
type ContainerService interface {
	// DiscoverContainers finds APFS containers on accessible devices
	DiscoverContainers(ctx context.Context) ([]ContainerInfo, error)

	// OpenContainer opens a container at the specified path
	OpenContainer(ctx context.Context, devicePath string) (ContainerInfo, error)

	// ReadSuperblock reads and parses the container superblock
	ReadSuperblock(ctx context.Context, devicePath string) (*types.NxSuperblockT, error)

	// ListVolumes enumerates all volumes in the container
	ListVolumes(ctx context.Context, devicePath string) ([]VolumeInfo, error)

	// GetSpaceManagerInfo retrieves space management information
	GetSpaceManagerInfo(ctx context.Context, devicePath string) (SpaceManagerInfo, error)

	// VerifyCheckpoints validates container checkpoints
	VerifyCheckpoints(ctx context.Context, devicePath string) error

	// Close closes the container and releases resources
	Close() error
}

// VolumeService provides volume-level operations
type VolumeService interface {
	// OpenVolume opens a specific volume by ID or name
	OpenVolume(ctx context.Context, containerPath string, volumeID uint64) (VolumeInfo, error)

	// OpenVolumeByName opens a volume by name
	OpenVolumeByName(ctx context.Context, containerPath string, volumeName string) (VolumeInfo, error)

	// ReadVolumeSuperblock reads the volume superblock
	ReadVolumeSuperblock(ctx context.Context, containerPath string, volumeID uint64) (*types.ApfsSuperblockT, error)

	// GetVolumeStatistics calculates volume statistics
	GetVolumeStatistics(ctx context.Context, containerPath string, volumeID uint64) (FileSystemStats, error)

	// ListSnapshots enumerates volume snapshots
	ListSnapshots(ctx context.Context, containerPath string, volumeID uint64) ([]SnapshotInfo, error)

	// CheckVolumeIntegrity performs volume integrity checking
	CheckVolumeIntegrity(ctx context.Context, containerPath string, volumeID uint64) (IntegrityStatus, error)

	// Close closes the volume and releases resources
	Close() error
}

// SnapshotInfo represents snapshot metadata
type SnapshotInfo struct {
	ObjectID     uint64
	Name         string
	CreatedAt    time.Time
	VolumeID     uint64
	FileCount    uint64
	DirCount     uint64
	DataSize     uint64
	MetadataSize uint64
}

// FilesystemService provides filesystem navigation and operations
type FilesystemService interface {
	// ListDirectory lists files and directories at the specified path
	ListDirectory(ctx context.Context, containerPath string, volumeID uint64, dirPath string, recursive bool) ([]FileInfo, error)

	// GetFileInfo retrieves detailed information about a specific file
	GetFileInfo(ctx context.Context, containerPath string, volumeID uint64, filePath string) (FileInfo, error)

	// GetDirectoryInfo retrieves directory information with statistics
	GetDirectoryInfo(ctx context.Context, containerPath string, volumeID uint64, dirPath string, includeChildren bool) (DirectoryInfo, error)

	// FindFiles searches for files matching specified criteria
	FindFiles(ctx context.Context, containerPath string, volumeID uint64, searchPath string, pattern string, maxResults int) ([]FileInfo, error)

	// GetInode retrieves file information by inode ID
	GetInode(ctx context.Context, containerPath string, volumeID uint64, inodeID uint64) (FileInfo, error)

	// WalkFilesystem performs a depth-first traversal of the filesystem
	WalkFilesystem(ctx context.Context, containerPath string, volumeID uint64, rootPath string, walkFunc func(FileInfo) error) error

	// CheckAccess determines if a file/directory is accessible (not encrypted)
	CheckAccess(ctx context.Context, containerPath string, volumeID uint64, filePath string) (bool, error)
}

// ExtractionService provides file and directory extraction
type ExtractionService interface {
	// ExtractFile extracts a single file to the specified destination
	ExtractFile(ctx context.Context, containerPath string, volumeID uint64, filePath string, destPath string, options ExtractionOptions) error

	// ExtractDirectory extracts a directory tree to the specified destination
	ExtractDirectory(ctx context.Context, containerPath string, volumeID uint64, sourcePath string, destPath string, options ExtractionOptions) error

	// ExtractVolume extracts an entire volume to the specified destination
	ExtractVolume(ctx context.Context, containerPath string, volumeID uint64, destPath string, options ExtractionOptions) error

	// StreamFile provides streaming access to file content
	StreamFile(ctx context.Context, containerPath string, volumeID uint64, filePath string) (io.ReadCloser, error)

	// ExtractMetadata extracts only metadata without file content
	ExtractMetadata(ctx context.Context, containerPath string, volumeID uint64, filePath string) (FileInfo, error)

	// EstimateExtractionSize calculates the total size of an extraction operation
	EstimateExtractionSize(ctx context.Context, containerPath string, volumeID uint64, sourcePath string, recursive bool) (uint64, error)

	// ValidateExtraction verifies an extraction completed successfully
	ValidateExtraction(ctx context.Context, sourcePath string, destPath string) error
}

// AnalysisService provides deep structural analysis
type AnalysisService interface {
	// AnalyzeContainer performs comprehensive container analysis
	AnalyzeContainer(ctx context.Context, devicePath string) (AnalysisResult, error)

	// AnalyzeVolume performs deep analysis of a specific volume
	AnalyzeVolume(ctx context.Context, containerPath string, volumeID uint64) (VolumeAnalysis, error)

	// AnalyzeBTrees analyzes B-tree structures and health
	AnalyzeBTrees(ctx context.Context, containerPath string, volumeID uint64) (BTreeStats, error)

	// CheckStructuralIntegrity performs comprehensive integrity checking
	CheckStructuralIntegrity(ctx context.Context, devicePath string) (StructuralHealth, error)

	// AnalyzeEncryption analyzes encryption configuration and status
	AnalyzeEncryption(ctx context.Context, devicePath string) (EncryptionInfo, error)

	// MeasurePerformance performs performance analysis
	MeasurePerformance(ctx context.Context, devicePath string) (PerformanceMetrics, error)

	// GenerateReport creates a comprehensive analysis report
	GenerateReport(ctx context.Context, devicePath string, format string) ([]byte, error)
}
