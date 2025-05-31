package container

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// containerFlagManager implements the ContainerFlagManager interface
type containerFlagManager struct {
	superblock *types.NxSuperblockT
}

// NewContainerFlagManager creates a new ContainerFlagManager implementation
func NewContainerFlagManager(superblock *types.NxSuperblockT) interfaces.ContainerFlagManager {
	return &containerFlagManager{
		superblock: superblock,
	}
}

// Flags returns the container's flags
func (cfgm *containerFlagManager) Flags() uint64 {
	return cfgm.superblock.NxFlags
}

// HasFlag checks if a specific flag is set
func (cfgm *containerFlagManager) HasFlag(flag uint64) bool {
	return cfgm.superblock.NxFlags&flag == flag
}

// UsesSoftwareCryptography checks if the container uses software cryptography
func (cfgm *containerFlagManager) UsesSoftwareCryptography() bool {
	return cfgm.HasFlag(types.NxCryptoSw)
}
