package volumes

import (
	"math"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// volumeSpaceManagement implements the VolumeSpaceManagement interface
type volumeSpaceManagement struct {
	superblock *types.ApfsSuperblockT
}

// ReservedBlockCount returns the number of blocks reserved for this volume
func (vsm *volumeSpaceManagement) ReservedBlockCount() uint64 {
	return vsm.superblock.ApfsFsReserveBlockCount
}

// QuotaBlockCount returns the maximum number of blocks this volume can allocate
func (vsm *volumeSpaceManagement) QuotaBlockCount() uint64 {
	return vsm.superblock.ApfsFsQuotaBlockCount
}

// AllocatedBlockCount returns the number of blocks currently allocated for this volume's file system
func (vsm *volumeSpaceManagement) AllocatedBlockCount() uint64 {
	return vsm.superblock.ApfsFsAllocCount
}

// TotalBlocksAllocated returns the total number of blocks that have been allocated
func (vsm *volumeSpaceManagement) TotalBlocksAllocated() uint64 {
	return vsm.superblock.ApfsTotalBlocksAlloced
}

// TotalBlocksFreed returns the total number of blocks that have been freed
func (vsm *volumeSpaceManagement) TotalBlocksFreed() uint64 {
	return vsm.superblock.ApfsTotalBlocksFreed
}

// SpaceUtilization calculates the space utilization percentage
func (vsm *volumeSpaceManagement) SpaceUtilization() float64 {
	// Handle cases to prevent division by zero
	if vsm.superblock.ApfsFsQuotaBlockCount == 0 {
		return 0.0
	}

	// Calculate utilization as a percentage
	utilization := float64(vsm.superblock.ApfsFsAllocCount) / float64(vsm.superblock.ApfsFsQuotaBlockCount) * 100.0

	// Ensure the utilization is between 0 and 100
	return math.Min(math.Max(utilization, 0.0), 100.0)
}

// NewVolumeSpaceManagement creates a new VolumeSpaceManagement implementation
func NewVolumeSpaceManagement(superblock *types.ApfsSuperblockT) interfaces.VolumeSpaceManagement {
	return &volumeSpaceManagement{
		superblock: superblock,
	}
}
