// File: internal/interfaces/data_recovery.go
package interfaces

import (
	"time"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// DataRecoveryManager provides comprehensive data recovery capabilities
type DataRecoveryManager interface {
	// ScanForRecoverableData scans for recoverable data structures
	ScanForRecoverableData(options ScanOptions) (RecoveryScanResult, error)

	// RecoverVolume attempts to recover a damaged or deleted volume
	RecoverVolume(volumeID types.OidT, options VolumeRecoveryOptions) (VolumeRecoveryResult, error)

	// RecoverFiles attempts to recover deleted or corrupted files
	RecoverFiles(criteria FileRecoveryCriteria, options FileRecoveryOptions) (FileRecoveryResult, error)

	// RecoverSnapshot attempts to recover a damaged snapshot
	RecoverSnapshot(snapshotXID types.XidT, options SnapshotRecoveryOptions) (SnapshotRecoveryResult, error)

	// RecoverBTree attempts to rebuild a corrupted B-tree
	RecoverBTree(btreeOID types.OidT, options BTreeRecoveryOptions) (BTreeRecoveryResult, error)

	// RecoverExtentTree attempts to recover the extent allocation tree
	RecoverExtentTree(options ExtentTreeRecoveryOptions) (ExtentTreeRecoveryResult, error)

	// ValidateRecoveredData validates the integrity of recovered data
	ValidateRecoveredData(data RecoveredData) (ValidationResult, error)
}

// ScanOptions contains options for scanning for recoverable data
type ScanOptions struct {
	// Scan entire container or specific regions
	ScanRegions []types.Prange

	// Types of data to scan for
	DataTypes []RecoverableDataType

	// Deep scan (slower but more thorough)
	DeepScan bool

	// Maximum time to spend scanning
	MaxScanTime time.Duration

	// Progress callback for long operations
	ProgressCallback func(completed, total uint64)

	// Whether to scan free space for deleted data
	ScanFreeSpace bool

	// Whether to use heuristic pattern matching
	UseHeuristics bool
}

// RecoverableDataType represents the type of data that can be recovered
type RecoverableDataType int

const (
	RecoverableDataTypeVolume RecoverableDataType = iota
	RecoverableDataTypeInode
	RecoverableDataTypeDirectory
	RecoverableDataTypeFile
	RecoverableDataTypeSnapshot
	RecoverableDataTypeBTree
	RecoverableDataTypeExtentTree
	RecoverableDataTypeCheckpoint
	RecoverableDataTypeObjectMap
	RecoverableDataTypeMetadata
)

// RecoveryScanResult contains the results of scanning for recoverable data
type RecoveryScanResult struct {
	// Total regions scanned
	RegionsScanned []types.Prange

	// Time taken for the scan
	ScanDuration time.Duration

	// Recoverable items found
	RecoverableItems []RecoverableItem

	// Number of items by type
	ItemCounts map[RecoverableDataType]int

	// Confidence level of the scan results
	ConfidenceLevel ScanConfidenceLevel

	// Issues encountered during scanning
	ScanIssues []ScanIssue
}

// RecoverableItem represents an item that can potentially be recovered
type RecoverableItem struct {
	// Type of recoverable item
	Type RecoverableDataType

	// Physical location of the item
	Location types.Prange

	// Object ID (if known)
	ObjectID *types.OidT

	// Transaction ID (if known)
	TransactionID *types.XidT

	// Estimated size of the item
	EstimatedSize uint64

	// Confidence level that this item can be recovered
	RecoveryConfidence RecoveryConfidenceLevel

	// Condition assessment
	Condition ItemCondition

	// Metadata about the item (if available)
	Metadata map[string]interface{}

	// Dependencies on other items for successful recovery
	Dependencies []types.OidT
}

// ScanConfidenceLevel represents confidence in scan results
type ScanConfidenceLevel int

const (
	ScanConfidenceLow ScanConfidenceLevel = iota
	ScanConfidenceMedium
	ScanConfidenceHigh
	ScanConfidenceVeryHigh
)

// RecoveryConfidenceLevel represents confidence in recovery success
type RecoveryConfidenceLevel int

const (
	RecoveryConfidenceLow RecoveryConfidenceLevel = iota
	RecoveryConfidenceMedium
	RecoveryConfidenceHigh
	RecoveryConfidenceVeryHigh
)

// ItemCondition represents the condition of a recoverable item
type ItemCondition int

const (
	ItemConditionGood ItemCondition = iota
	ItemConditionPartiallyCorrupted
	ItemConditionSeverelyCorrupted
	ItemConditionFragmented
	ItemConditionIncomplete
)

// ScanIssue represents an issue encountered during scanning
type ScanIssue struct {
	// Type of issue
	Type ScanIssueType

	// Severity of the issue
	Severity ScanIssueSeverity

	// Description of the issue
	Description string

	// Location where the issue was encountered
	Location *types.Prange

	// Suggested resolution
	SuggestedResolution string
}

// ScanIssueType represents the type of scan issue
type ScanIssueType int

const (
	ScanIssueTypeReadError ScanIssueType = iota
	ScanIssueTypeCorruptedStructure
	ScanIssueTypeInconsistentData
	ScanIssueTypeUnknownStructure
	ScanIssueTypePartialData
)

// ScanIssueSeverity represents the severity of a scan issue
type ScanIssueSeverity int

const (
	ScanIssueSeverityInfo ScanIssueSeverity = iota
	ScanIssueSeverityWarning
	ScanIssueSeverityError
	ScanIssueSeverityCritical
)

// VolumeRecoveryOptions contains options for volume recovery
type VolumeRecoveryOptions struct {
	// Whether to attempt to recover the volume superblock
	RecoverSuperblock bool

	// Whether to attempt to recover B-trees
	RecoverBTrees bool

	// Whether to attempt to recover file data
	RecoverFileData bool

	// Whether to recover to a new location (vs in-place)
	RecoverToNewLocation bool

	// Target location for recovery (if RecoverToNewLocation is true)
	TargetLocation string

	// Maximum number of files to attempt to recover
	MaxFilesToRecover int

	// Whether to validate recovered data
	ValidateRecoveredData bool

	// Recovery strategy to use
	Strategy VolumeRecoveryStrategy
}

// VolumeRecoveryStrategy represents the strategy for volume recovery
type VolumeRecoveryStrategy int

const (
	VolumeRecoveryStrategyConservative VolumeRecoveryStrategy = iota
	VolumeRecoveryStrategyAggressive
	VolumeRecoveryStrategyMinimal
	VolumeRecoveryStrategyComplete
)

// VolumeRecoveryResult contains the results of volume recovery
type VolumeRecoveryResult struct {
	// Whether the recovery was successful
	Success bool

	// Recovered volume (if successful)
	RecoveredVolume Volume

	// Number of files recovered
	FilesRecovered int

	// Number of directories recovered
	DirectoriesRecovered int

	// Total data recovered in bytes
	DataRecovered uint64

	// Recovery issues encountered
	Issues []RecoveryIssue

	// Recovery statistics
	Statistics VolumeRecoveryStatistics

	// Time taken for recovery
	RecoveryDuration time.Duration
}

// VolumeRecoveryStatistics contains statistics about volume recovery
type VolumeRecoveryStatistics struct {
	// Attempted recoveries by type
	AttemptedRecoveries map[RecoverableDataType]int

	// Successful recoveries by type
	SuccessfulRecoveries map[RecoverableDataType]int

	// Failed recoveries by type
	FailedRecoveries map[RecoverableDataType]int

	// Recovery success rate
	SuccessRate float64

	// Data integrity score of recovered volume
	IntegrityScore float64
}

// FileRecoveryCriteria specifies criteria for file recovery
type FileRecoveryCriteria struct {
	// File name patterns to look for
	NamePatterns []string

	// File extensions to recover
	Extensions []string

	// Size range of files to recover
	MinSize uint64
	MaxSize uint64

	// Date range for file modification
	ModifiedAfter  *time.Time
	ModifiedBefore *time.Time

	// Specific inode IDs to recover
	InodeIDs []uint64

	// Parent directory to search in
	ParentDirectory *uint64

	// Whether to search recursively
	Recursive bool
}

// FileRecoveryOptions contains options for file recovery
type FileRecoveryOptions struct {
	// Maximum number of files to recover
	MaxFiles int

	// Whether to recover file metadata
	RecoverMetadata bool

	// Whether to recover extended attributes
	RecoverExtendedAttributes bool

	// Whether to recover resource forks (macOS)
	RecoverResourceForks bool

	// Recovery destination
	DestinationPath string

	// Whether to preserve directory structure
	PreserveStructure bool

	// Whether to validate recovered files
	ValidateFiles bool

	// File naming strategy for recovered files
	NamingStrategy FileNamingStrategy
}

// FileNamingStrategy represents how recovered files should be named
type FileNamingStrategy int

const (
	FileNamingStrategyOriginal FileNamingStrategy = iota
	FileNamingStrategyWithTimestamp
	FileNamingStrategyWithInodeID
	FileNamingStrategySequential
)

// FileRecoveryResult contains the results of file recovery
type FileRecoveryResult struct {
	// List of successfully recovered files
	RecoveredFiles []RecoveredFile

	// List of files that couldn't be recovered
	FailedFiles []FailedFileRecovery

	// Total files attempted
	FilesAttempted int

	// Total files successfully recovered
	FilesRecovered int

	// Total data recovered in bytes
	DataRecovered uint64

	// Recovery duration
	RecoveryDuration time.Duration

	// Recovery issues
	Issues []RecoveryIssue
}

// RecoveredFile represents a successfully recovered file
type RecoveredFile struct {
	// Original inode ID
	OriginalInodeID uint64

	// Original file path (if known)
	OriginalPath string

	// Recovered file path
	RecoveredPath string

	// File size
	Size uint64

	// Recovery confidence level
	ConfidenceLevel RecoveryConfidenceLevel

	// Data integrity status
	IntegrityStatus FileIntegrityStatus

	// Recovery timestamp
	RecoveryTime time.Time

	// Original file metadata (if recovered)
	Metadata *FileMetadata
}

// FailedFileRecovery represents a file that couldn't be recovered
type FailedFileRecovery struct {
	// Inode ID that was attempted
	InodeID uint64

	// Reason for failure
	FailureReason string

	// Error details
	Error error

	// Partial data recovered (if any)
	PartialDataSize uint64
}

// FileIntegrityStatus represents the integrity status of a recovered file
type FileIntegrityStatus int

const (
	FileIntegrityComplete FileIntegrityStatus = iota
	FileIntegrityPartial
	FileIntegrityCorrupted
	FileIntegrityUnverifiable
)

// FileMetadata contains metadata about a recovered file
type FileMetadata struct {
	// File permissions
	Permissions types.ModeT

	// Owner and group
	OwnerUID types.UidT
	GroupGID types.GidT

	// Timestamps
	CreationTime     time.Time
	ModificationTime time.Time
	AccessTime       time.Time

	// Extended attributes
	ExtendedAttributes map[string][]byte

	// Resource fork data (macOS)
	ResourceFork []byte
}

// SnapshotRecoveryOptions contains options for snapshot recovery
type SnapshotRecoveryOptions struct {
	// Whether to recover the snapshot metadata
	RecoverMetadata bool

	// Whether to recover the snapshot data
	RecoverData bool

	// Whether to validate the recovered snapshot
	ValidateSnapshot bool

	// Target location for recovery
	TargetLocation string
}

// SnapshotRecoveryResult contains the results of snapshot recovery
type SnapshotRecoveryResult struct {
	// Whether recovery was successful
	Success bool

	// Recovered snapshot (if successful)
	RecoveredSnapshot SnapshotReader

	// Recovery confidence level
	ConfidenceLevel RecoveryConfidenceLevel

	// Issues encountered during recovery
	Issues []RecoveryIssue

	// Recovery duration
	RecoveryDuration time.Duration
}

// BTreeRecoveryOptions contains options for B-tree recovery
type BTreeRecoveryOptions struct {
	// Whether to attempt to rebuild the tree structure
	RebuildStructure bool

	// Whether to recover individual nodes
	RecoverNodes bool

	// Whether to validate the recovered tree
	ValidateTree bool

	// Recovery strategy
	Strategy BTreeRecoveryStrategy
}

// BTreeRecoveryStrategy represents the strategy for B-tree recovery
type BTreeRecoveryStrategy int

const (
	BTreeRecoveryStrategyConservative BTreeRecoveryStrategy = iota
	BTreeRecoveryStrategyAggressive
	BTreeRecoveryStrategyRebuild
)

// BTreeRecoveryResult contains the results of B-tree recovery
type BTreeRecoveryResult struct {
	// Whether recovery was successful
	Success bool

	// Number of nodes recovered
	NodesRecovered int

	// Number of records recovered
	RecordsRecovered int

	// Tree integrity after recovery
	IntegrityScore float64

	// Issues encountered during recovery
	Issues []RecoveryIssue

	// Recovery duration
	RecoveryDuration time.Duration
}

// ExtentTreeRecoveryOptions contains options for extent tree recovery
type ExtentTreeRecoveryOptions struct {
	// Whether to rebuild the free space bitmap
	RebuildBitmap bool

	// Whether to validate extent allocations
	ValidateAllocations bool

	// Recovery strategy
	Strategy ExtentTreeRecoveryStrategy
}

// ExtentTreeRecoveryStrategy represents the strategy for extent tree recovery
type ExtentTreeRecoveryStrategy int

const (
	ExtentTreeRecoveryStrategyConservative ExtentTreeRecoveryStrategy = iota
	ExtentTreeRecoveryStrategyAggressive
	ExtentTreeRecoveryStrategyRebuild
)

// ExtentTreeRecoveryResult contains the results of extent tree recovery
type ExtentTreeRecoveryResult struct {
	// Whether recovery was successful
	Success bool

	// Number of extents recovered
	ExtentsRecovered int

	// Free space recovered in bytes
	FreeSpaceRecovered uint64

	// Allocation consistency after recovery
	AllocationConsistency float64

	// Issues encountered during recovery
	Issues []RecoveryIssue

	// Recovery duration
	RecoveryDuration time.Duration
}

// RecoveryIssue represents an issue encountered during recovery
type RecoveryIssue struct {
	// Type of issue
	Type RecoveryIssueType

	// Severity of the issue
	Severity RecoveryIssueSeverity

	// Description of the issue
	Description string

	// Component affected
	Component string

	// Impact on recovery
	Impact string

	// Suggested resolution
	SuggestedResolution string
}

// RecoveryIssueType represents the type of recovery issue
type RecoveryIssueType int

const (
	RecoveryIssueTypeDataCorruption RecoveryIssueType = iota
	RecoveryIssueTypeIncompleteData
	RecoveryIssueTypeStructuralDamage
	RecoveryIssueTypeIncompatibleVersion
	RecoveryIssueTypeInsufficientSpace
	RecoveryIssueTypePermissionDenied
)

// RecoveryIssueSeverity represents the severity of a recovery issue
type RecoveryIssueSeverity int

const (
	RecoveryIssueSeverityInfo RecoveryIssueSeverity = iota
	RecoveryIssueSeverityWarning
	RecoveryIssueSeverityError
	RecoveryIssueSeverityCritical
)

// RecoveredData represents data that has been recovered
type RecoveredData struct {
	// Type of recovered data
	Type RecoverableDataType

	// Raw data
	Data []byte

	// Metadata about the recovery
	RecoveryMetadata RecoveryMetadata

	// Validation status
	ValidationStatus ValidationStatus
}

// RecoveryMetadata contains metadata about a recovery operation
type RecoveryMetadata struct {
	// When the recovery was performed
	RecoveryTime time.Time

	// Method used for recovery
	RecoveryMethod string

	// Confidence level of the recovery
	ConfidenceLevel RecoveryConfidenceLevel

	// Source location of the data
	SourceLocation types.Prange

	// Recovery tool version
	ToolVersion string
}

// ValidationResult contains the result of validating recovered data
type ValidationResult struct {
	// Whether the data is valid
	IsValid bool

	// Confidence level in the validation
	ConfidenceLevel ValidationConfidenceLevel

	// Issues found during validation
	Issues []ValidationIssue

	// Data completeness percentage
	CompletenessPercentage float64

	// Data integrity score
	IntegrityScore float64
}

// ValidationStatus represents the validation status of recovered data
type ValidationStatus int

const (
	ValidationStatusNotValidated ValidationStatus = iota
	ValidationStatusValid
	ValidationStatusPartiallyValid
	ValidationStatusInvalid
	ValidationStatusUnverifiable
)

// ValidationConfidenceLevel represents confidence in validation results
type ValidationConfidenceLevel int

const (
	ValidationConfidenceLow ValidationConfidenceLevel = iota
	ValidationConfidenceMedium
	ValidationConfidenceHigh
	ValidationConfidenceVeryHigh
)

// ValidationIssue represents an issue found during data validation
type ValidationIssue struct {
	// Type of validation issue
	Type ValidationIssueType

	// Severity of the issue
	Severity ValidationIssueSeverity

	// Description of the issue
	Description string

	// Location of the issue (if applicable)
	Location *types.Prange

	// Impact on data usability
	Impact string
}

// ValidationIssueType represents the type of validation issue
type ValidationIssueType int

const (
	ValidationIssueTypeCorruptedData ValidationIssueType = iota
	ValidationIssueTypeInconsistentChecksum
	ValidationIssueTypeMissingData
	ValidationIssueTypeInvalidStructure
	ValidationIssueTypeVersionMismatch
)

// ValidationIssueSeverity represents the severity of a validation issue
type ValidationIssueSeverity int

const (
	ValidationIssueSeverityInfo ValidationIssueSeverity = iota
	ValidationIssueSeverityWarning
	ValidationIssueSeverityError
	ValidationIssueSeverityCritical
)

// EmergencyRecoveryManager provides emergency recovery capabilities for severely damaged containers
type EmergencyRecoveryManager interface {
	// AssessContainerDamage assesses the level of damage to a container
	AssessContainerDamage() (DamageAssessment, error)

	// RecoverPartialContainer attempts to recover what can be salvaged from a damaged container
	RecoverPartialContainer(options EmergencyRecoveryOptions) (EmergencyRecoveryResult, error)

	// ExtractRawData extracts raw data from specific locations for manual analysis
	ExtractRawData(locations []types.Prange, outputPath string) error

	// GenerateRecoveryReport generates a comprehensive report of recovery possibilities
	GenerateRecoveryReport() (RecoveryReport, error)
}

// DamageAssessment contains assessment of container damage
type DamageAssessment struct {
	// Overall damage level
	DamageLevel DamageLevel

	// Specific areas of damage
	DamagedAreas []DamagedArea

	// Recoverable components
	RecoverableComponents []RecoverableComponent

	// Estimated recovery success rate
	EstimatedSuccessRate float64

	// Recommended recovery strategy
	RecommendedStrategy RecoveryStrategy
}

// DamageLevel represents the level of damage to a container
type DamageLevel int

const (
	DamageLevelMinor DamageLevel = iota
	DamageLevelModerate
	DamageLevelSevere
	DamageLevelCritical
	DamageLevelTotal
)

// DamagedArea represents an area of the container that is damaged
type DamagedArea struct {
	// Location of the damage
	Location types.Prange

	// Type of component damaged
	ComponentType string

	// Severity of damage
	Severity DamageLevel

	// Description of the damage
	Description string

	// Impact on overall container
	Impact string
}

// RecoverableComponent represents a component that can potentially be recovered
type RecoverableComponent struct {
	// Type of component
	Type RecoverableDataType

	// Location of the component
	Location types.Prange

	// Condition of the component
	Condition ItemCondition

	// Recovery confidence
	RecoveryConfidence RecoveryConfidenceLevel

	// Dependencies for recovery
	Dependencies []string
}

// RecoveryStrategy represents the recommended strategy for recovery
type RecoveryStrategy int

const (
	RecoveryStrategyMinimal RecoveryStrategy = iota
	RecoveryStrategyConservative
	RecoveryStrategyAggressive
	RecoveryStrategyEmergency
)

// EmergencyRecoveryOptions contains options for emergency recovery
type EmergencyRecoveryOptions struct {
	// Recovery strategy to use
	Strategy RecoveryStrategy

	// Whether to attempt risky recovery operations
	AllowRiskyOperations bool

	// Output location for recovered data
	OutputLocation string

	// Maximum time to spend on recovery
	MaxRecoveryTime time.Duration

	// Whether to create a recovery log
	CreateRecoveryLog bool
}

// EmergencyRecoveryResult contains the results of emergency recovery
type EmergencyRecoveryResult struct {
	// Whether any data was recovered
	DataRecovered bool

	// Amount of data recovered in bytes
	RecoveredDataSize uint64

	// Number of files recovered
	FilesRecovered int

	// Recovery confidence level
	OverallConfidence RecoveryConfidenceLevel

	// Issues encountered
	Issues []RecoveryIssue

	// Path to recovery log (if created)
	RecoveryLogPath string

	// Recovery duration
	RecoveryDuration time.Duration
}

// RecoveryReport contains a comprehensive recovery report
type RecoveryReport struct {
	// Assessment of damage
	DamageAssessment DamageAssessment

	// Recovery recommendations
	Recommendations []RecoveryRecommendation

	// Estimated costs and benefits
	CostBenefitAnalysis CostBenefitAnalysis

	// Risk assessment
	RiskAssessment RiskAssessment

	// Timeline for recovery
	RecoveryTimeline RecoveryTimeline
}

// RecoveryRecommendation represents a recommendation for recovery
type RecoveryRecommendation struct {
	// Title of the recommendation
	Title string

	// Detailed description
	Description string

	// Priority level
	Priority RecommendationPriority

	// Expected outcome
	ExpectedOutcome string

	// Risk level
	RiskLevel RecommendationRisk

	// Estimated time required
	EstimatedTime time.Duration
}

// CostBenefitAnalysis contains cost-benefit analysis for recovery
type CostBenefitAnalysis struct {
	// Estimated time cost
	TimeCost time.Duration

	// Estimated resource requirements
	ResourceRequirements string

	// Expected data recovery percentage
	ExpectedRecoveryPercentage float64

	// Value of recoverable data
	DataValue string

	// Risk of further data loss
	RiskOfDataLoss float64
}

// RiskAssessment contains risk assessment for recovery operations
type RiskAssessment struct {
	// Overall risk level
	OverallRisk RiskLevel

	// Specific risks identified
	IdentifiedRisks []Risk

	// Risk mitigation strategies
	MitigationStrategies []string
}

// RiskLevel represents the level of risk
type RiskLevel int

const (
	RiskLevelLow RiskLevel = iota
	RiskLevelMedium
	RiskLevelHigh
	RiskLevelCritical
)

// Risk represents a specific risk
type Risk struct {
	// Description of the risk
	Description string

	// Probability of occurrence
	Probability float64

	// Impact if it occurs
	Impact string

	// Severity level
	Severity RiskLevel
}

// RecoveryTimeline contains timeline information for recovery
type RecoveryTimeline struct {
	// Total estimated time
	TotalEstimatedTime time.Duration

	// Recovery phases
	Phases []RecoveryPhase

	// Critical path items
	CriticalPath []string
}

// RecoveryPhase represents a phase in the recovery process
type RecoveryPhase struct {
	// Name of the phase
	Name string

	// Description of activities
	Description string

	// Estimated duration
	EstimatedDuration time.Duration

	// Dependencies on other phases
	Dependencies []string

	// Risk level for this phase
	RiskLevel RiskLevel
}
