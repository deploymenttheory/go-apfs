package interfaces

import (
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// SpaceManagerReader provides methods for reading space manager information
type SpaceManagerReader interface {
	// Superblock returns the space manager's physical structure
	Superblock() *types.SpacemanPhysT

	// BlockSize returns the logical block size
	BlockSize() uint32

	// BlocksPerChunk returns the number of blocks per chunk
	BlocksPerChunk() uint32

	// Version returns the space manager version
	Version() uint32
}

// DeviceSpaceInfo provides space-related information for a specific device
type DeviceSpaceInfo interface {
	// TotalBlocks returns the total number of blocks in the device
	TotalBlocks() uint64

	// FreeBlocks returns the number of free blocks in the device
	FreeBlocks() uint64

	// ChunkCount returns the number of chunks in the device
	ChunkCount() uint64

	// ChunkInfoBlockCount returns the number of chunk-info blocks
	ChunkInfoBlockCount() uint32

	// AddressOffset returns the address offset for the device
	AddressOffset() uint32
}

// SpaceAllocationManager provides methods for managing space allocation
type SpaceAllocationManager interface {
	// ReservedBlockCount returns the number of blocks reserved for the file system
	ReservedBlockCount() uint64

	// AllocatedReservedBlocks returns the number of reserved blocks that have been allocated
	AllocatedReservedBlocks() uint64

	// GetAllocationZones returns the allocation zones for the devices
	GetAllocationZones() *types.SpacemanDataZoneInfoPhysT
}

// FreeSpaceQueue provides methods for managing free space queues
type FreeSpaceQueue interface {
	// Count returns the number of entries in the queue
	Count() uint64

	// OldestTransactionID returns the oldest transaction identifier in the queue
	OldestTransactionID() types.XidT

	// TreeNodeLimit returns the limit on the number of nodes in the tree
	TreeNodeLimit() uint16
}

// InternalPoolManager provides methods for managing the internal pool
type InternalPoolManager interface {
	// BlockCount returns the number of blocks in the internal pool
	BlockCount() uint64

	// BitmapBaseAddress returns the base address of the internal-pool bitmap
	BitmapBaseAddress() types.Paddr

	// InternalPoolBaseAddress returns the base address of the internal pool
	InternalPoolBaseAddress() types.Paddr

	// BitmapSize returns the size of the internal-pool bitmap in blocks
	BitmapSize() uint32
}

// SpaceManagerInspector provides methods for inspecting and managing space
type SpaceManagerInspector interface {
	// ListDevices returns information about all managed devices
	ListDevices() ([]DeviceSpaceInfo, error)

	// GetFreeQueue retrieves a specific free queue
	GetFreeQueue(queueType types.SfqT) (FreeSpaceQueue, error)

	// CalculateFreeSpace calculates the total free space across all devices
	CalculateFreeSpace() uint64
}
