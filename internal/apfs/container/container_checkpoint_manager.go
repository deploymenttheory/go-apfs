package container

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// containerCheckpointManager implements the ContainerCheckpointManager interface
type containerCheckpointManager struct {
	superblock *types.NxSuperblockT
}

// NewContainerCheckpointManager creates a new ContainerCheckpointManager implementation
func NewContainerCheckpointManager(superblock *types.NxSuperblockT) interfaces.ContainerCheckpointManager {
	return &containerCheckpointManager{
		superblock: superblock,
	}
}

// CheckpointDescriptorBlockCount returns the number of blocks used by the checkpoint descriptor area
func (ccm *containerCheckpointManager) CheckpointDescriptorBlockCount() uint32 {
	// Mask out the highest bit which is used as a flag
	return ccm.superblock.NxXpDescBlocks & 0x7FFFFFFF
}

// CheckpointDataBlockCount returns the number of blocks used by the checkpoint data area
func (ccm *containerCheckpointManager) CheckpointDataBlockCount() uint32 {
	// Mask out the highest bit which is used as a flag
	return ccm.superblock.NxXpDataBlocks & 0x7FFFFFFF
}

// CheckpointDescriptorBase returns the base address of the checkpoint descriptor area
func (ccm *containerCheckpointManager) CheckpointDescriptorBase() types.Paddr {
	return ccm.superblock.NxXpDescBase
}

// CheckpointDataBase returns the base address of the checkpoint data area
func (ccm *containerCheckpointManager) CheckpointDataBase() types.Paddr {
	return ccm.superblock.NxXpDataBase
}

// CheckpointDescriptorNext returns the next index to use in the checkpoint descriptor area
func (ccm *containerCheckpointManager) CheckpointDescriptorNext() uint32 {
	return ccm.superblock.NxXpDescNext
}

// CheckpointDataNext returns the next index to use in the checkpoint data area
func (ccm *containerCheckpointManager) CheckpointDataNext() uint32 {
	return ccm.superblock.NxXpDataNext
}

// CheckpointDescriptorIndex returns the index of the first valid item in the checkpoint descriptor area
func (ccm *containerCheckpointManager) CheckpointDescriptorIndex() uint32 {
	return ccm.superblock.NxXpDescIndex
}

// CheckpointDescriptorLength returns the number of blocks in the checkpoint descriptor area used by the current checkpoint
func (ccm *containerCheckpointManager) CheckpointDescriptorLength() uint32 {
	return ccm.superblock.NxXpDescLen
}

// CheckpointDataIndex returns the index of the first valid item in the checkpoint data area
func (ccm *containerCheckpointManager) CheckpointDataIndex() uint32 {
	return ccm.superblock.NxXpDataIndex
}

// CheckpointDataLength returns the number of blocks in the checkpoint data area used by the current checkpoint
func (ccm *containerCheckpointManager) CheckpointDataLength() uint32 {
	return ccm.superblock.NxXpDataLen
}
