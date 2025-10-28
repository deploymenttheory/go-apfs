// File: internal/interfaces/block_device.go
package interfaces

import (
	"io"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// BlockDeviceReader provides methods for reading from block devices
type BlockDeviceReader interface {
	// ReadBlock reads a single block at the specified address
	ReadBlock(address types.Paddr) ([]byte, error)

	// ReadBlockRange reads multiple consecutive blocks
	ReadBlockRange(start types.Paddr, count uint32) ([]byte, error)

	// ReadBytes reads a specific number of bytes starting at a block address and offset
	ReadBytes(address types.Paddr, offset uint32, length uint32) ([]byte, error)

	// BlockSize returns the size of a single block in bytes
	BlockSize() uint32

	// TotalBlocks returns the total number of blocks on the device
	TotalBlocks() uint64

	// TotalSize returns the total size of the device in bytes
	TotalSize() uint64

	// IsValidAddress checks if a block address is valid
	IsValidAddress(address types.Paddr) bool

	// CanReadRange checks if a range of blocks can be read
	CanReadRange(start types.Paddr, count uint32) bool
}

// BlockDeviceWriter provides methods for writing to block devices
type BlockDeviceWriter interface {
	// WriteBlock writes a single block at the specified address
	WriteBlock(address types.Paddr, data []byte) error

	// WriteBlockRange writes multiple consecutive blocks
	WriteBlockRange(start types.Paddr, data []byte) error

	// WriteBytes writes a specific number of bytes starting at a block address and offset
	WriteBytes(address types.Paddr, offset uint32, data []byte) error

	// FlushWrites ensures all pending writes are committed to storage
	FlushWrites() error

	// IsReadOnly checks if the device is read-only
	IsReadOnly() bool

	// CanWriteRange checks if a range of blocks can be written
	CanWriteRange(start types.Paddr, count uint32) bool
}

// BlockDeviceInfo provides information about a block device
type BlockDeviceInfo interface {
	// DevicePath returns the system path to the device
	DevicePath() string

	// DeviceType returns the type of device (e.g., "disk", "image", "partition")
	DeviceType() string

	// IsRemovable checks if the device is removable
	IsRemovable() bool

	// IsWritable checks if the device supports writing
	IsWritable() bool

	// VendorInfo returns vendor information about the device
	VendorInfo() (vendor, model string)

	// SerialNumber returns the device serial number
	SerialNumber() string

	// FirmwareVersion returns the device firmware version
	FirmwareVersion() string
}

// BlockDeviceManager provides methods for managing block device access
type BlockDeviceManager interface {
	// OpenDevice opens a block device for reading/writing
	OpenDevice(devicePath string) (BlockDevice, error)

	// CloseDevice closes a block device
	CloseDevice(device BlockDevice) error

	// ListDevices returns all available block devices
	ListDevices() ([]BlockDeviceInfo, error)

	// FindDeviceByPath finds a device by its system path
	FindDeviceByPath(path string) (BlockDeviceInfo, error)

	// DetectAPFSDevices returns devices that contain APFS containers
	DetectAPFSDevices() ([]BlockDeviceInfo, error)
}

// BlockDevice represents a complete block device interface
type BlockDevice interface {
	BlockDeviceReader
	BlockDeviceWriter
	BlockDeviceInfo
	io.Closer
}

// BlockCache provides caching functionality for block device operations
type BlockCache interface {
	// GetBlock retrieves a block from cache or reads it from device
	GetBlock(address types.Paddr) ([]byte, error)

	// PutBlock stores a block in the cache
	PutBlock(address types.Paddr, data []byte) error

	// InvalidateBlock removes a block from the cache
	InvalidateBlock(address types.Paddr) error

	// FlushCache writes all dirty blocks to the device
	FlushCache() error

	// ClearCache removes all blocks from the cache
	ClearCache() error

	// CacheStatistics returns cache performance statistics
	CacheStatistics() BlockCacheStats
}

// BlockCacheStats contains cache performance statistics
type BlockCacheStats struct {
	// Total number of cache hits
	Hits uint64

	// Total number of cache misses
	Misses uint64

	// Current number of blocks in cache
	BlocksInCache uint32

	// Maximum number of blocks the cache can hold
	MaxBlocks uint32

	// Cache hit ratio as a percentage
	HitRatio float64

	// Total bytes currently cached
	BytesCached uint64
}

// SparseBlockReader provides methods for reading sparse block ranges efficiently
type SparseBlockReader interface {
	// ReadSparseRange reads blocks, skipping over unallocated ranges
	ReadSparseRange(ranges []types.Prange) ([]SparseBlock, error)

	// IsBlockAllocated checks if a block is allocated
	IsBlockAllocated(address types.Paddr) (bool, error)

	// FindAllocatedBlocks finds all allocated blocks in a range
	FindAllocatedBlocks(start types.Paddr, count uint32) ([]types.Paddr, error)
}

// SparseBlock represents a block with its allocation status
type SparseBlock struct {
	// Physical address of the block
	Address types.Paddr

	// Block data (nil if not allocated)
	Data []byte

	// Whether the block is allocated
	Allocated bool
}

// BlockDeviceValidator provides methods for validating block device operations
type BlockDeviceValidator interface {
	// ValidateDevice performs comprehensive device validation
	ValidateDevice() (BlockDeviceValidationResult, error)

	// CheckBlockIntegrity verifies the integrity of specific blocks
	CheckBlockIntegrity(addresses []types.Paddr) ([]BlockIntegrityResult, error)

	// VerifyReadWrite tests read/write functionality (on non-critical blocks)
	VerifyReadWrite() (bool, error)
}

// BlockDeviceValidationResult contains the result of device validation
type BlockDeviceValidationResult struct {
	// Whether the device passed validation
	IsValid bool

	// Issues found during validation
	Issues []BlockDeviceIssue

	// Device capabilities detected
	Capabilities BlockDeviceCapabilities

	// Performance metrics
	Performance BlockDevicePerformance
}

// BlockDeviceIssue represents a problem found during device validation
type BlockDeviceIssue struct {
	// Type of issue
	Type BlockDeviceIssueType

	// Severity of the issue
	Severity BlockDeviceIssueSeverity

	// Description of the issue
	Description string

	// Affected block address (if applicable)
	AffectedBlock types.Paddr

	// Additional details
	Details map[string]any
}

// BlockDeviceIssueType represents the type of device issue
type BlockDeviceIssueType int

const (
	BlockDeviceIssueUnreadableBlock BlockDeviceIssueType = iota
	BlockDeviceIssueUnwritableBlock
	BlockDeviceIssueSlowResponse
	BlockDeviceIssueInconsistentSize
	BlockDeviceIssuePermissionDenied
	BlockDeviceIssueHardwareError
)

// BlockDeviceIssueSeverity represents the severity of a device issue
type BlockDeviceIssueSeverity int

const (
	BlockDeviceIssueSeverityInfo BlockDeviceIssueSeverity = iota
	BlockDeviceIssueSeverityWarning
	BlockDeviceIssueSeverityError
	BlockDeviceIssueSeverityCritical
)

// BlockDeviceCapabilities describes what a device can do
type BlockDeviceCapabilities struct {
	// Whether the device supports reading
	CanRead bool

	// Whether the device supports writing
	CanWrite bool

	// Whether the device supports sparse reads
	SupportsSparseReads bool

	// Whether the device supports efficient random access
	SupportsRandomAccess bool

	// Maximum transfer size in bytes
	MaxTransferSize uint32

	// Optimal transfer size in bytes
	OptimalTransferSize uint32
}

// BlockDevicePerformance contains performance metrics for a device
type BlockDevicePerformance struct {
	// Average read latency in microseconds
	AverageReadLatency uint32

	// Average write latency in microseconds
	AverageWriteLatency uint32

	// Read throughput in bytes per second
	ReadThroughput uint64

	// Write throughput in bytes per second
	WriteThroughput uint64

	// Number of read operations tested
	ReadOperationsTested uint32

	// Number of write operations tested
	WriteOperationsTested uint32
}

// BlockIntegrityResult contains the result of checking a block's integrity
type BlockIntegrityResult struct {
	// Address of the block that was checked
	Address types.Paddr

	// Whether the block is readable
	IsReadable bool

	// Whether the block is writable (if tested)
	IsWritable bool

	// Whether the block data is consistent
	IsConsistent bool

	// Any errors encountered
	Error error
}
