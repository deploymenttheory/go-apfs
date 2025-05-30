package volumes

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// volumeExtendedMetadata implements the VolumeExtendedMetadata interface
type volumeExtendedMetadata struct {
	superblock *types.ApfsSuperblockT
}

// NewVolumeExtendedMetadata creates a new VolumeExtendedMetadata implementation
func NewVolumeExtendedMetadata(superblock *types.ApfsSuperblockT) interfaces.VolumeExtendedMetadata {
	return &volumeExtendedMetadata{
		superblock: superblock,
	}
}

// SnapshotMetadataExtOID returns the extended snapshot metadata object identifier
func (vem *volumeExtendedMetadata) SnapshotMetadataExtOID() types.OidT {
	return vem.superblock.ApfsSnapMetaExtOid
}

// FileExtentTreeOID returns the file extent tree object identifier
func (vem *volumeExtendedMetadata) FileExtentTreeOID() types.OidT {
	return vem.superblock.ApfsFextTreeOid
}

// FileExtentTreeType returns the file extent tree type
func (vem *volumeExtendedMetadata) FileExtentTreeType() uint32 {
	return vem.superblock.ApfsFextTreeType
}
