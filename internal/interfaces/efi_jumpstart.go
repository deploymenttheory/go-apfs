package interfaces

import (
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// EFIJumpstartReader provides methods for reading the EFI jumpstart information
type EFIJumpstartReader interface {
	// Magic returns the magic number for validating the EFI jumpstart structure
	Magic() uint32

	// Version returns the version number of the EFI jumpstart structure
	Version() uint32

	// EFIFileLength returns the size in bytes of the embedded EFI driver
	EFIFileLength() uint32

	// ExtentCount returns the number of extents where the EFI driver is stored
	ExtentCount() uint32

	// Extents returns the locations where the EFI driver is stored
	Extents() []types.Prange

	// IsValid checks if the EFI jumpstart structure is valid based on magic number and version
	IsValid() bool
}

// EFIDriverExtractor provides methods for extracting the embedded EFI driver
type EFIDriverExtractor interface {
	// ExtractEFIDriver extracts the EFI driver to the specified output path
	ExtractEFIDriver(outputPath string) error

	// GetEFIDriverData returns the raw data of the EFI driver
	GetEFIDriverData() ([]byte, error)

	// ValidateEFIDriver checks if the EFI driver data is intact and valid
	ValidateEFIDriver() error
}

// EFIPartitionManager provides methods for working with EFI-related partition information
type EFIPartitionManager interface {
	// GetPartitionUUID returns the partition UUID for an APFS partition
	GetPartitionUUID() string

	// IsAPFSPartition checks if a partition contains an APFS filesystem based on its UUID
	IsAPFSPartition(partitionUUID string) bool

	// ListEFIPartitions returns all EFI partitions on the device
	ListEFIPartitions() ([]EFIPartitionInfo, error)
}

// EFIPartitionInfo represents information about an EFI partition
type EFIPartitionInfo struct {
	// The UUID of the partition
	UUID string

	// The name of the partition if available
	Name string

	// The size of the partition in bytes
	Size uint64

	// The starting offset of the partition
	Offset uint64
}

// EFIJumpstartAnalyzer provides methods for analyzing EFI jumpstart structures
type EFIJumpstartAnalyzer interface {
	// AnalyzeEFIJumpstart performs a detailed analysis of the EFI jumpstart structure
	AnalyzeEFIJumpstart() (EFIJumpstartAnalysis, error)

	// VerifyEFIJumpstart checks if the EFI jumpstart structure is valid and consistent
	VerifyEFIJumpstart() error
}

// EFIJumpstartAnalysis contains detailed information from analyzing an EFI jumpstart structure
type EFIJumpstartAnalysis struct {
	// Indicates whether the structure is valid
	IsValid bool

	// The version of the EFI jumpstart structure
	Version uint32

	// The size of the EFI driver in bytes
	DriverSize uint32

	// The number of extents used to store the driver
	ExtentCount uint32

	// Detailed information about each extent
	ExtentDetails []EFIExtentDetail

	// Information about the driver, if it could be parsed
	DriverInfo map[string]string
}

// EFIExtentDetail provides details about an individual extent in the EFI jumpstart
type EFIExtentDetail struct {
	// The starting physical block address of the extent
	StartAddress types.Paddr

	// The number of blocks in the extent
	BlockCount uint64

	// The size of the extent in bytes
	SizeInBytes uint64
}

// EFIJumpstartLocator provides methods for locating EFI jumpstart structures
type EFIJumpstartLocator interface {
	// FindEFIJumpstart locates the EFI jumpstart structure in a container
	FindEFIJumpstart() (types.Paddr, error)

	// FindEFIJumpstartInPartition locates the EFI jumpstart structure in a specific partition
	FindEFIJumpstartInPartition(partitionUUID string) (types.Paddr, error)
}

// BootabilityChecker provides methods for checking if a volume is bootable
type BootabilityChecker interface {
	// IsBootable checks if the APFS container/volume has a valid EFI jumpstart
	IsBootable() bool

	// GetBootRequirements returns what is needed for the container/volume to be bootable
	GetBootRequirements() ([]string, error)

	// VerifyBootConfiguration checks if the boot configuration is valid
	VerifyBootConfiguration() error
}
