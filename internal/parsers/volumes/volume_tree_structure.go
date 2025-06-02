package volumes

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// volumeTreeStructure implements the VolumeTreeStructure interface
type volumeTreeStructure struct {
	superblock *types.ApfsSuperblockT
}

// RootTreeOID returns the virtual object identifier of the root file-system tree
func (vts *volumeTreeStructure) RootTreeOID() types.OidT {
	return vts.superblock.ApfsRootTreeOid
}

// RootTreeType returns the type of the root file-system tree
func (vts *volumeTreeStructure) RootTreeType() uint32 {
	return vts.superblock.ApfsRootTreeType
}

// ExtentReferenceTreeOID returns the physical object identifier of the extent-reference tree
func (vts *volumeTreeStructure) ExtentReferenceTreeOID() types.OidT {
	return vts.superblock.ApfsExtentrefTreeOid
}

// ExtentReferenceTreeType returns the type of the extent-reference tree
func (vts *volumeTreeStructure) ExtentReferenceTreeType() uint32 {
	return vts.superblock.ApfsExtentreftreeType
}

// SnapshotMetadataTreeOID returns the virtual object identifier of the snapshot metadata tree
func (vts *volumeTreeStructure) SnapshotMetadataTreeOID() types.OidT {
	return vts.superblock.ApfsSnapMetaTreeOid
}

// SnapshotMetadataTreeType returns the type of the snapshot metadata tree
func (vts *volumeTreeStructure) SnapshotMetadataTreeType() uint32 {
	return vts.superblock.ApfsSnapMetatreeType
}

// ObjectMapOID returns the physical object identifier of the volume's object map
func (vts *volumeTreeStructure) ObjectMapOID() types.OidT {
	return vts.superblock.ApfsOmapOid
}

// NewVolumeTreeStructure creates a new VolumeTreeStructure implementation
func NewVolumeTreeStructure(superblock *types.ApfsSuperblockT) interfaces.VolumeTreeStructure {
	return &volumeTreeStructure{
		superblock: superblock,
	}
}
