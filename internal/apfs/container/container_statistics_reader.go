package container

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// containerStatisticsReader implements the ContainerStatisticsReader interface
type containerStatisticsReader struct {
	superblock *types.NxSuperblockT
}

// NewContainerStatisticsReader creates a new ContainerStatisticsReader implementation
func NewContainerStatisticsReader(superblock *types.NxSuperblockT) interfaces.ContainerStatisticsReader {
	return &containerStatisticsReader{
		superblock: superblock,
	}
}

// Counters returns the array of counters that store information about the container
func (csr *containerStatisticsReader) Counters() []uint64 {
	// Convert the fixed-size array to a slice
	counters := make([]uint64, types.NxNumCounters)
	copy(counters, csr.superblock.NxCounters[:])
	return counters
}

// ObjectChecksumSetCount returns the number of times a checksum has been computed while writing objects to disk
func (csr *containerStatisticsReader) ObjectChecksumSetCount() uint64 {
	// Based on the counter IDs defined in types, this would be a specific counter
	// For now, we'll return the first counter as an example
	if len(csr.superblock.NxCounters) > 0 {
		return csr.superblock.NxCounters[0]
	}
	return 0
}

// ObjectChecksumFailCount returns the number of times an object's checksum was invalid when reading from disk
func (csr *containerStatisticsReader) ObjectChecksumFailCount() uint64 {
	// Based on the counter IDs defined in types, this would be a specific counter
	// For now, we'll return the second counter as an example
	if len(csr.superblock.NxCounters) > 1 {
		return csr.superblock.NxCounters[1]
	}
	return 0
}
