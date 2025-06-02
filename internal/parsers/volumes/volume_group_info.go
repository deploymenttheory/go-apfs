package volumes

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// volumeGroupInfo implements the VolumeGroupInfo interface
type volumeGroupInfo struct {
	superblock *types.ApfsSuperblockT
}

// NewVolumeGroupInfo creates a new VolumeGroupInfo implementation
func NewVolumeGroupInfo(superblock *types.ApfsSuperblockT) interfaces.VolumeGroupInfo {
	return &volumeGroupInfo{
		superblock: superblock,
	}
}

// VolumeGroupID returns the volume group the volume belongs to
func (vgi *volumeGroupInfo) VolumeGroupID() types.UUID {
	return vgi.superblock.ApfsVolumeGroupId
}

// IntegrityMetadataOID returns the virtual object identifier of the integrity metadata object
func (vgi *volumeGroupInfo) IntegrityMetadataOID() types.OidT {
	return vgi.superblock.ApfsIntegrityMetaOid
}
