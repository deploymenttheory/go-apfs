// File: internal/interfaces/analysis.go
package interfaces

import (
	"time"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// APFSAnalyzer provides comprehensive analysis of APFS containers and volumes
type APFSAnalyzer interface {
	// AnalyzeContainer performs a comprehensive analysis of a container
	AnalyzeContainer(container ContainerManager) (ContainerAnalysis, error)

	// AnalyzeVolume performs a comprehensive analysis of a volume
	AnalyzeVolume(volume Volume) (VolumeAnalysis, error)

	// CompareContainers compares two containers and highlights differences
	CompareContainers(container1, container2 ContainerManager) (ContainerComparison, error)

	// CompareVolumes compares two volumes and highlights differences
	CompareVolumes(volume1, volume2 Volume) (VolumeComparison, error)

	// GenerateReport generates a detailed report of the APFS structure
	GenerateReport(container ContainerManager, options ReportOptions) (AnalysisReport, error)
}

// ContainerAnalysis contains comprehensive analysis results for a container
type ContainerAnalysis struct {
	// Basic container information
	ContainerInfo ContainerInfo

	// Space utilization analysis
	SpaceAnalysis SpaceAnalysis

	// Volume analysis for each volume in the container
	VolumeAnalyses []VolumeAnalysis

	// Snapshot analysis across all volumes
	SnapshotAnalysis SnapshotEfficiencyAnalysis

	// Health and integrity assessment
	HealthAssessment HealthAssessment

	// Performance characteristics
	PerformanceAnalysis PerformanceAnalysis

	// Security analysis
	SecurityAnalysis SecurityAnalysis

	// Recommendations for optimization
	Recommendations []Recommendation
}

// ContainerInfo contains basic information about a container
type ContainerInfo struct {
	// Container UUID
	UUID types.UUID

	// APFS version
	APFSVersion string

	// Block size in bytes
	BlockSize uint32

	// Total number of blocks
	BlockCount uint64

	// Total size in bytes
	TotalSize uint64

	// Number of volumes
	VolumeCount int

	// Container creation time (if available)
	CreationTime *time.Time

	// Features in use
	Features []string

	// Encryption status
	IsEncrypted bool
}

// SpaceAnalysis contains detailed space utilization information
type SpaceAnalysis struct {
	// Total space in bytes
	TotalSpace uint64

	// Used space in bytes
	UsedSpace uint64

	// Free space in bytes
	FreeSpace uint64

	// Space utilization percentage
	UtilizationPercentage float64

	// Per-volume space breakdown
	VolumeSpaceBreakdown []VolumeSpaceInfo

	// Snapshot overhead
	SnapshotOverhead uint64

	// Metadata overhead
	MetadataOverhead uint64

	// Fragmentation analysis
	FragmentationLevel float64

	// Wasted space due to fragmentation
	WastedSpace uint64

	// Space efficiency score (0-100)
	EfficiencyScore float64
}

// VolumeSpaceInfo contains space information for a single volume
type VolumeSpaceInfo struct {
	// Volume name
	Name string

	// Volume UUID
	UUID types.UUID

	// Space used by this volume
	UsedSpace uint64

	// Number of files in the volume
	FileCount uint64

	// Average file size
	AverageFileSize uint64

	// Largest file size
	LargestFileSize uint64
}

// VolumeAnalysis contains comprehensive analysis results for a volume
type VolumeAnalysis struct {
	// Basic volume information
	VolumeInfo VolumeInfo

	// File system structure analysis
	FileSystemAnalysis FileSystemAnalysis

	// Directory structure analysis
	DirectoryAnalysis DirectoryAnalysis

	// File type distribution
	FileTypeDistribution FileTypeDistribution

	// Access pattern analysis
	AccessPatternAnalysis AccessPatternAnalysis

	// Security and permissions analysis
	SecurityAnalysis SecurityAnalysis

	// Performance characteristics
	PerformanceAnalysis PerformanceAnalysis

	// Recommendations specific to this volume
	Recommendations []Recommendation
}

// VolumeInfo contains basic information about a volume
type VolumeInfo struct {
	// Volume name
	Name string

	// Volume UUID
	UUID types.UUID

	// Volume role
	Role string

	// Case sensitivity
	IsCaseSensitive bool

	// Encryption status
	IsEncrypted bool

	// Snapshot count
	SnapshotCount int

	// Volume creation time
	CreationTime time.Time

	// Last modification time
	LastModified time.Time

	// Reserved space
	ReservedSpace uint64

	// Quota (if set)
	Quota *uint64
}

// FileSystemAnalysis contains analysis of the file system structure
type FileSystemAnalysis struct {
	// Total number of inodes
	TotalInodes uint64

	// Number of files
	FileCount uint64

	// Number of directories
	DirectoryCount uint64

	// Number of symbolic links
	SymlinkCount uint64

	// B-tree statistics
	BTreeStatistics BTreeStatistics

	// Extent tree analysis
	ExtentTreeAnalysis ExtentTreeAnalysis

	// Orphaned inode count
	OrphanedInodes uint64

	// Integrity issues found
	IntegrityIssues []IntegrityIssue
}

// BTreeStatistics contains statistics about B-tree structures
type BTreeStatistics struct {
	// Number of B-tree nodes
	NodeCount uint64

	// Tree depth
	MaxDepth int

	// Average node utilization
	AverageUtilization float64

	// Fragmentation level
	FragmentationLevel float64

	// Number of split operations estimated
	SplitOperations uint64
}

// ExtentTreeAnalysis contains analysis of extent allocation
type ExtentTreeAnalysis struct {
	// Total number of extents
	ExtentCount uint64

	// Average extent size
	AverageExtentSize uint64

	// Largest extent size
	LargestExtentSize uint64

	// Smallest extent size
	SmallestExtentSize uint64

	// Fragmentation ratio
	FragmentationRatio float64

	// Number of single-block extents (highly fragmented)
	SingleBlockExtents uint64
}

// DirectoryAnalysis contains analysis of directory structure
type DirectoryAnalysis struct {
	// Maximum directory depth
	MaxDepth int

	// Average directory depth
	AverageDepth float64

	// Largest directory (by entry count)
	LargestDirectorySize uint64

	// Average directory size
	AverageDirectorySize float64

	// Number of empty directories
	EmptyDirectories uint64

	// Directories with very few entries (potentially inefficient)
	SparseDictionaries uint64

	// Long path analysis
	LongPaths []LongPathInfo
}

// LongPathInfo contains information about unusually long paths
type LongPathInfo struct {
	// The long path
	Path string

	// Length in characters
	Length int

	// Depth (number of path components)
	Depth int
}

// FileTypeDistribution contains statistics about file types
type FileTypeDistribution struct {
	// Distribution by file extension
	ExtensionDistribution map[string]FileTypeStats

	// Distribution by MIME type (if detectable)
	MIMETypeDistribution map[string]FileTypeStats

	// Distribution by file size ranges
	SizeDistribution map[string]FileTypeStats

	// Binary vs text file ratio
	BinaryFileRatio float64

	// Compressed file ratio
	CompressedFileRatio float64
}

// FileTypeStats contains statistics for a specific file type
type FileTypeStats struct {
	// Number of files of this type
	Count uint64

	// Total size of files of this type
	TotalSize uint64

	// Average size of files of this type
	AverageSize uint64

	// Percentage of total files
	CountPercentage float64

	// Percentage of total space
	SizePercentage float64
}

// AccessPatternAnalysis contains analysis of file access patterns
type AccessPatternAnalysis struct {
	// Recently accessed files
	RecentlyAccessed []FileAccessInfo

	// Files that haven't been accessed in a long time
	StaleFiles []FileAccessInfo

	// Most frequently accessed directories
	HotDirectories []DirectoryAccessInfo

	// Access time distribution
	AccessTimeDistribution map[string]uint64

	// Modification time distribution
	ModificationTimeDistribution map[string]uint64
}

// FileAccessInfo contains access information for a file
type FileAccessInfo struct {
	// File path
	Path string

	// File size
	Size uint64

	// Last access time
	LastAccess time.Time

	// Last modification time
	LastModification time.Time

	// Creation time
	CreationTime time.Time

	// Access frequency estimate
	AccessFrequency AccessFrequency
}

// DirectoryAccessInfo contains access information for a directory
type DirectoryAccessInfo struct {
	// Directory path
	Path string

	// Number of entries
	EntryCount uint64

	// Last access time
	LastAccess time.Time

	// Access frequency estimate
	AccessFrequency AccessFrequency
}

// AccessFrequency represents how frequently a file or directory is accessed
type AccessFrequency int

const (
	AccessFrequencyVeryLow AccessFrequency = iota
	AccessFrequencyLow
	AccessFrequencyMedium
	AccessFrequencyHigh
	AccessFrequencyVeryHigh
)

// SecurityAnalysis contains security-related analysis
type SecurityAnalysis struct {
	// Files with unusual permissions
	UnusualPermissions []FilePermissionInfo

	// World-writable files
	WorldWritableFiles []string

	// SetUID/SetGID files
	SetUIDFiles []string

	// Files with no owner
	OrphanedFiles []string

	// Extended attribute usage
	ExtendedAttributeUsage ExtendedAttributeStats

	// Encryption status analysis
	EncryptionAnalysis EncryptionAnalysis
}

// FilePermissionInfo contains permission information for a file
type FilePermissionInfo struct {
	// File path
	Path string

	// File permissions
	Permissions types.ModeT

	// Owner UID
	OwnerUID types.UidT

	// Group GID
	GroupGID types.GidT

	// Reason why permissions are considered unusual
	Reason string
}

// ExtendedAttributeStats contains statistics about extended attribute usage
type ExtendedAttributeStats struct {
	// Files with extended attributes
	FilesWithExtendedAttributes uint64

	// Most common extended attribute names
	CommonAttributes map[string]uint64

	// Total size of extended attributes
	TotalExtendedAttributeSize uint64

	// Files with resource forks
	FilesWithResourceForks uint64
}

// EncryptionAnalysis contains analysis of encryption usage
type EncryptionAnalysis struct {
	// Whether the container is encrypted
	ContainerEncrypted bool

	// Whether individual volumes are encrypted
	VolumeEncryptionStatus map[string]bool

	// Encryption algorithms in use
	EncryptionAlgorithms []string

	// Files with per-file encryption
	PerFileEncryptionCount uint64
}

// PerformanceAnalysis contains performance-related analysis
type PerformanceAnalysis struct {
	// Storage efficiency metrics
	StorageEfficiency StorageEfficiencyMetrics

	// Predicted I/O performance characteristics
	IOPerformance IOPerformanceMetrics

	// B-tree performance analysis
	BTreePerformance BTreePerformanceMetrics

	// Fragmentation impact on performance
	FragmentationImpact FragmentationImpactMetrics
}

// StorageEfficiencyMetrics contains storage efficiency measurements
type StorageEfficiencyMetrics struct {
	// Compression ratio achieved
	CompressionRatio float64

	// Deduplication opportunities
	DeduplicationPotential uint64

	// Sparse file effectiveness
	SparseFileEffectiveness float64

	// Snapshot efficiency
	SnapshotEfficiency float64
}

// IOPerformanceMetrics contains I/O performance predictions
type IOPerformanceMetrics struct {
	// Estimated random read performance impact
	RandomReadImpact float64

	// Estimated sequential read performance impact
	SequentialReadImpact float64

	// Estimated write performance impact
	WriteImpact float64

	// Fragmentation impact on I/O
	FragmentationIOImpact float64
}

// BTreePerformanceMetrics contains B-tree performance analysis
type BTreePerformanceMetrics struct {
	// Search efficiency (based on tree balance)
	SearchEfficiency float64

	// Insert efficiency (based on node utilization)
	InsertEfficiency float64

	// Tree maintenance overhead
	MaintenanceOverhead float64
}

// FragmentationImpactMetrics contains fragmentation impact analysis
type FragmentationImpactMetrics struct {
	// Overall fragmentation score (0-100, higher is more fragmented)
	FragmentationScore float64

	// Performance impact percentage
	PerformanceImpact float64

	// Recommended defragmentation priority
	DefragmentationPriority DefragmentationPriority
}

// DefragmentationPriority represents the priority level for defragmentation
type DefragmentationPriority int

const (
	DefragmentationPriorityLow DefragmentationPriority = iota
	DefragmentationPriorityMedium
	DefragmentationPriorityHigh
	DefragmentationPriorityCritical
)

// HealthAssessment contains overall health assessment
type HealthAssessment struct {
	// Overall health score (0-100)
	OverallHealthScore float64

	// Critical issues found
	CriticalIssues []HealthIssue

	// Warning-level issues
	Warnings []HealthIssue

	// Information-level observations
	Observations []HealthIssue

	// Predicted time until maintenance needed
	MaintenanceRecommendation *time.Duration
}

// HealthIssue represents a health-related issue
type HealthIssue struct {
	// Type of issue
	Type HealthIssueType

	// Severity level
	Severity HealthIssueSeverity

	// Description of the issue
	Description string

	// Affected component
	Component string

	// Recommended action
	RecommendedAction string

	// Impact on performance/reliability
	Impact string
}

// HealthIssueType represents the type of health issue
type HealthIssueType int

const (
	HealthIssueFragmentation HealthIssueType = iota
	HealthIssueCorruption
	HealthIssueSpaceExhaustion
	HealthIssuePerformanceDegradation
	HealthIssueSecurityConcern
	HealthIssueCompatibilityIssue
)

// HealthIssueSeverity represents the severity of a health issue
type HealthIssueSeverity int

const (
	HealthIssueSeverityInfo HealthIssueSeverity = iota
	HealthIssueSeverityWarning
	HealthIssueSeverityError
	HealthIssueSeverityCritical
)

// Recommendation represents a recommendation for optimization or maintenance
type Recommendation struct {
	// Type of recommendation
	Type RecommendationType

	// Priority level
	Priority RecommendationPriority

	// Title/summary of the recommendation
	Title string

	// Detailed description
	Description string

	// Expected benefit
	ExpectedBenefit string

	// Implementation difficulty
	Difficulty RecommendationDifficulty

	// Estimated time to implement
	EstimatedTime *time.Duration

	// Risk level of implementing the recommendation
	RiskLevel RecommendationRisk
}

// RecommendationType represents the type of recommendation
type RecommendationType int

const (
	RecommendationTypeDefragmentation RecommendationType = iota
	RecommendationTypeSpaceReclamation
	RecommendationTypeSecurityImprovement
	RecommendationTypePerformanceOptimization
	RecommendationTypeMaintenanceScheduling
	RecommendationTypeCapacityPlanning
)

// RecommendationPriority represents the priority of a recommendation
type RecommendationPriority int

const (
	RecommendationPriorityLow RecommendationPriority = iota
	RecommendationPriorityMedium
	RecommendationPriorityHigh
	RecommendationPriorityCritical
)

// RecommendationDifficulty represents how difficult a recommendation is to implement
type RecommendationDifficulty int

const (
	RecommendationDifficultyEasy RecommendationDifficulty = iota
	RecommendationDifficultyMedium
	RecommendationDifficultyHard
	RecommendationDifficultyVeryHard
)

// RecommendationRisk represents the risk level of implementing a recommendation
type RecommendationRisk int

const (
	RecommendationRiskLow RecommendationRisk = iota
	RecommendationRiskMedium
	RecommendationRiskHigh
)

// ContainerComparison contains the results of comparing two containers
type ContainerComparison struct {
	// The two containers being compared
	Container1 ContainerManager
	Container2 ContainerManager

	// Differences found between the containers
	Differences []ContainerDifference

	// Summary of differences
	DifferenceSummary ComparisonSummary

	// Similarity score (0-100)
	SimilarityScore float64
}

// VolumeComparison contains the results of comparing two volumes
type VolumeComparison struct {
	// The two volumes being compared
	Volume1 Volume
	Volume2 Volume

	// Differences found between the volumes
	Differences []VolumeDifference

	// Summary of differences
	DifferenceSummary ComparisonSummary

	// Similarity score (0-100)
	SimilarityScore float64
}

// ContainerDifference represents a difference between two containers
type ContainerDifference struct {
	// Type of difference
	Type DifferenceType

	// Component that differs
	Component string

	// Value in first container
	Value1 interface{}

	// Value in second container
	Value2 interface{}

	// Description of the difference
	Description string
}

// VolumeDifference represents a difference between two volumes
type VolumeDifference struct {
	// Type of difference
	Type DifferenceType

	// Component that differs
	Component string

	// Value in first volume
	Value1 interface{}

	// Value in second volume
	Value2 interface{}

	// Description of the difference
	Description string
}

// DifferenceType represents the type of difference found
type DifferenceType int

const (
	DifferenceTypeStructural DifferenceType = iota
	DifferenceTypeConfiguration
	DifferenceTypeContent
	DifferenceTypeMetadata
	DifferenceTypePerformance
	DifferenceTypeSecurity
)

// ComparisonSummary contains a summary of comparison results
type ComparisonSummary struct {
	// Total number of differences found
	TotalDifferences int

	// Number of differences by type
	DifferencesByType map[DifferenceType]int

	// Most significant differences
	MajorDifferences []string

	// Overall comparison verdict
	Verdict string
}

// ReportOptions contains options for generating analysis reports
type ReportOptions struct {
	// Include detailed file listings
	IncludeFileListings bool

	// Include directory tree structure
	IncludeDirectoryTree bool

	// Include performance analysis
	IncludePerformanceAnalysis bool

	// Include security analysis
	IncludeSecurityAnalysis bool

	// Include recommendations
	IncludeRecommendations bool

	// Output format
	OutputFormat ReportFormat

	// Maximum number of items to include in lists
	MaxListItems int

	// Whether to include raw data
	IncludeRawData bool
}

// ReportFormat represents the format for analysis reports
type ReportFormat int

const (
	ReportFormatText ReportFormat = iota
	ReportFormatJSON
	ReportFormatXML
	ReportFormatHTML
	ReportFormatPDF
)

// AnalysisReport contains the complete analysis report
type AnalysisReport struct {
	// Report metadata
	Metadata ReportMetadata

	// Container analysis (if applicable)
	ContainerAnalysis *ContainerAnalysis

	// Volume analyses
	VolumeAnalyses []VolumeAnalysis

	// Executive summary
	ExecutiveSummary string

	// Detailed findings
	DetailedFindings string

	// Recommendations summary
	RecommendationsSummary string

	// Raw data (if requested)
	RawData map[string]interface{}
}

// ReportMetadata contains metadata about the analysis report
type ReportMetadata struct {
	// When the analysis was performed
	AnalysisTime time.Time

	// Version of the analysis tool
	AnalyzerVersion string

	// Analysis duration
	AnalysisDuration time.Duration

	// Source of the data analyzed
	DataSource string

	// Report generation options used
	Options ReportOptions
}
