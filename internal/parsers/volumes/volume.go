package volumes

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// volume implements the complete Volume interface by embedding all component interfaces
type volume struct {
	interfaces.VolumeIdentity
	interfaces.VolumeFeatures
	interfaces.VolumeSpaceManagement
	interfaces.VolumeTreeStructure
	interfaces.VolumeMetadata
	interfaces.VolumeEncryptionMetadata
	interfaces.VolumeSnapshotMetadata
	interfaces.VolumeResourceCounts
	interfaces.VolumeGroupInfo
	interfaces.VolumeIntegrityCheck
	interfaces.VolumeEncryptionRollingState
	interfaces.VolumeCloneInfo
	interfaces.VolumeExtendedMetadata

	superblock *types.ApfsSuperblockT
}

// NewVolume creates a new Volume implementation from an APFS superblock
func NewVolume(superblock *types.ApfsSuperblockT) interfaces.Volume {
	return &volume{
		VolumeIdentity:               NewVolumeIdentity(superblock),
		VolumeFeatures:               NewVolumeFeatures(superblock),
		VolumeSpaceManagement:        NewVolumeSpaceManagement(superblock),
		VolumeTreeStructure:          NewVolumeTreeStructure(superblock),
		VolumeMetadata:               NewVolumeMetadata(superblock),
		VolumeEncryptionMetadata:     NewVolumeEncryptionMetadata(superblock),
		VolumeSnapshotMetadata:       NewVolumeSnapshotMetadata(superblock),
		VolumeResourceCounts:         NewVolumeResourceCounts(superblock),
		VolumeGroupInfo:              NewVolumeGroupInfo(superblock),
		VolumeIntegrityCheck:         NewVolumeIntegrityCheck(superblock),
		VolumeEncryptionRollingState: NewVolumeEncryptionRollingState(superblock),
		VolumeCloneInfo:              NewVolumeCloneInfo(superblock),
		VolumeExtendedMetadata:       NewVolumeExtendedMetadata(superblock),
		superblock:                   superblock,
	}
}
