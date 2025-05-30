package volumes

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// volumeIntegrityCheck implements the VolumeIntegrityCheck interface
type volumeIntegrityCheck struct {
	superblock *types.ApfsSuperblockT
}

// NewVolumeIntegrityCheck creates a new VolumeIntegrityCheck implementation
func NewVolumeIntegrityCheck(superblock *types.ApfsSuperblockT) interfaces.VolumeIntegrityCheck {
	return &volumeIntegrityCheck{
		superblock: superblock,
	}
}

// ValidateMagicNumber checks if the volume has the correct magic number
func (vic *volumeIntegrityCheck) ValidateMagicNumber() bool {
	return vic.superblock.ApfsMagic == types.ApfsMagic
}

// MagicNumber returns the volume's magic number
func (vic *volumeIntegrityCheck) MagicNumber() uint32 {
	return vic.superblock.ApfsMagic
}
