package container

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// containerEphemeralManager implements the ContainerEphemeralManager interface
type containerEphemeralManager struct {
	superblock *types.NxSuperblockT
}

// NewContainerEphemeralManager creates a new ContainerEphemeralManager implementation
func NewContainerEphemeralManager(superblock *types.NxSuperblockT) interfaces.ContainerEphemeralManager {
	return &containerEphemeralManager{
		superblock: superblock,
	}
}

// EphemeralInfo returns the array of fields used in management of ephemeral data
func (cem *containerEphemeralManager) EphemeralInfo() []uint64 {
	// Convert the fixed-size array to a slice
	ephInfo := make([]uint64, types.NxEphInfoCount)
	copy(ephInfo, cem.superblock.NxEphemeralInfo[:])
	return ephInfo
}

// MinimumBlockCount returns the default minimum size in blocks for structures containing ephemeral data
func (cem *containerEphemeralManager) MinimumBlockCount() uint32 {
	return types.NxEphMinBlockCount
}

// MaxEphemeralStructures returns the number of structures containing ephemeral data that a volume can have
func (cem *containerEphemeralManager) MaxEphemeralStructures() uint32 {
	return types.NxMaxFileSystemEphStructs
}

// EphemeralInfoVersion returns the version number for structures containing ephemeral data
func (cem *containerEphemeralManager) EphemeralInfoVersion() uint32 {
	return types.NxEphInfoVersion1
}
