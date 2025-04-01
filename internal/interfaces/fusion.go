// File: internal/interfaces/fusion.go
package interfaces

import (
	"github.com/deploymenttheory/go-apfs/internal/types"
)

type FusionContainerManager interface {
	// Check if Fusion Drive is supported at the container level
	SupportsFusion() bool

	// Get the Fusion Drive UUID
	FusionUUID() types.UUID

	// Get the Fusion Middle Tree Object ID
	FusionMiddleTreeOID() types.OidT

	// Get the Fusion Write-back Cache Object ID
	FusionWriteBackCacheOID() types.OidT

	// Get Fusion-specific container features or flags
	FusionContainerFlags() uint64
}

// FusionWriteBackCacheReader provides methods for reading Fusion write-back cache information
type FusionWriteBackCacheReader interface {
	// Version returns the version of the write-back cache
	Version() uint64

	// ListHeadOID returns the object identifier of the head of the write-back cache list
	ListHeadOID() types.OidT

	// ListTailOID returns the object identifier of the tail of the write-back cache list
	ListTailOID() types.OidT

	// StableHeadOffset returns the offset of the stable head of the write-back cache
	StableHeadOffset() uint64

	// StableTailOffset returns the offset of the stable tail of the write-back cache
	StableTailOffset() uint64

	// ListBlocksCount returns the number of blocks in the write-back cache list
	ListBlocksCount() uint32

	// ReadCacheUsage returns the amount of space used by the read cache
	ReadCacheUsage() uint64

	// ReadCacheStashLocation returns the location of the read cache stash
	ReadCacheStashLocation() types.Prange
}

// FusionWriteBackCacheListReader provides methods for reading the Fusion write-back cache list
type FusionWriteBackCacheListReader interface {
	// Version returns the version of the write-back cache list
	Version() uint64

	// TailOffset returns the offset of the tail of the write-back cache list
	TailOffset() uint64

	// IndexBegin returns the beginning index of the write-back cache list
	IndexBegin() uint32

	// IndexEnd returns the ending index of the write-back cache list
	IndexEnd() uint32

	// MaxIndex returns the maximum index of the write-back cache list
	MaxIndex() uint32

	// ListEntries returns all entries in the write-back cache list
	ListEntries() []FusionWriteBackCacheEntry
}

// FusionWriteBackCacheEntry represents a single entry in the Fusion write-back cache list
type FusionWriteBackCacheEntry interface {
	// WriteBackCacheLBA returns the logical block address in the write-back cache
	WriteBackCacheLBA() types.Paddr

	// TargetLBA returns the target logical block address
	TargetLBA() types.Paddr

	// Length returns the length of the entry
	Length() uint64
}

// FusionMiddleTreeReader provides methods for reading Fusion middle tree information
type FusionMiddleTreeReader interface {
	// ListEntries returns all entries in the Fusion middle tree
	ListEntries() ([]FusionMiddleTreeEntry, error)

	// FindEntryByLBA finds a middle tree entry by logical block address
	FindEntryByLBA(lba types.Paddr) (FusionMiddleTreeEntry, bool)
}

// FusionMiddleTreeEntry represents an entry in the Fusion middle tree
type FusionMiddleTreeEntry interface {
	// LogicalBlockAddress returns the logical block address
	LogicalBlockAddress() types.Paddr

	// Length returns the length of the extent
	Length() uint32

	// IsDirty checks if the extent is marked as dirty
	IsDirty() bool

	// IsTenant checks if the extent is a tenant
	IsTenant() bool
}

// FusionDeviceInspector provides methods for inspecting Fusion storage configuration
type FusionDeviceInspector interface {
	// IsFusionDrive checks if the current storage configuration is a Fusion drive
	IsFusionDrive() bool

	// MainDeviceType returns the type of the main (typically SSD) device
	MainDeviceType() string

	// SecondaryDeviceType returns the type of the secondary (typically HDD) device
	SecondaryDeviceType() string

	// CacheUtilization returns the percentage of cache being used
	CacheUtilization() float64

	// ListCachedExtents returns information about cached extents
	ListCachedExtents() ([]FusionMiddleTreeEntry, error)
}
