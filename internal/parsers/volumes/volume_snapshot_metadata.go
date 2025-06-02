package volumes

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// volumeSnapshotMetadata implements the VolumeSnapshotMetadata interface
type volumeSnapshotMetadata struct {
	superblock *types.ApfsSuperblockT
}

// NewVolumeSnapshotMetadata creates a new VolumeSnapshotMetadata implementation
func NewVolumeSnapshotMetadata(superblock *types.ApfsSuperblockT) interfaces.VolumeSnapshotMetadata {
	return &volumeSnapshotMetadata{
		superblock: superblock,
	}
}

// TotalSnapshots returns the total number of snapshots
func (vsm *volumeSnapshotMetadata) TotalSnapshots() uint64 {
	return vsm.superblock.ApfsNumSnapshots
}

// RevertToSnapshotXID returns the transaction identifier of a snapshot to revert to
func (vsm *volumeSnapshotMetadata) RevertToSnapshotXID() types.XidT {
	return vsm.superblock.ApfsRevertToXid
}

// RevertToSuperblockOID returns the physical object identifier of a volume superblock to revert to
func (vsm *volumeSnapshotMetadata) RevertToSuperblockOID() types.OidT {
	return vsm.superblock.ApfsRevertToSblockOid
}

// RootToSnapshotXID returns the transaction identifier of the snapshot to root from
func (vsm *volumeSnapshotMetadata) RootToSnapshotXID() types.XidT {
	return vsm.superblock.ApfsRootToXid
}
