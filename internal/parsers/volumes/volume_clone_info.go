package volumes

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// volumeCloneInfo implements the VolumeCloneInfo interface
type volumeCloneInfo struct {
	superblock *types.ApfsSuperblockT
}

// NewVolumeCloneInfo creates a new VolumeCloneInfo implementation
func NewVolumeCloneInfo(superblock *types.ApfsSuperblockT) interfaces.VolumeCloneInfo {
	return &volumeCloneInfo{
		superblock: superblock,
	}
}

// CloneInfoIdEpoch returns the largest object identifier used by this volume
// at the time INODE_WAS_EVER_CLONED started storing valid information
func (vci *volumeCloneInfo) CloneInfoIdEpoch() uint64 {
	return vci.superblock.ApfsCloneinfoIdEpoch
}

// CloneInfoXID returns the transaction identifier used with ApfsCloneinfoIdEpoch
func (vci *volumeCloneInfo) CloneInfoXID() uint64 {
	return vci.superblock.ApfsCloneinfoXid
}
