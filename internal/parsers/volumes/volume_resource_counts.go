package volumes

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// volumeResourceCounts implements the VolumeResourceCounts interface
type volumeResourceCounts struct {
	superblock *types.ApfsSuperblockT
}

// NewVolumeResourceCounts creates a new VolumeResourceCounts implementation
func NewVolumeResourceCounts(superblock *types.ApfsSuperblockT) interfaces.VolumeResourceCounts {
	return &volumeResourceCounts{
		superblock: superblock,
	}
}

// TotalFiles returns the number of regular files in this volume
func (vrc *volumeResourceCounts) TotalFiles() uint64 {
	return vrc.superblock.ApfsNumFiles
}

// TotalDirectories returns the number of directories in this volume
func (vrc *volumeResourceCounts) TotalDirectories() uint64 {
	return vrc.superblock.ApfsNumDirectories
}

// TotalSymlinks returns the number of symbolic links in this volume
func (vrc *volumeResourceCounts) TotalSymlinks() uint64 {
	return vrc.superblock.ApfsNumSymlinks
}

// TotalOtherFileSystemObjects returns the number of other files in this volume
func (vrc *volumeResourceCounts) TotalOtherFileSystemObjects() uint64 {
	return vrc.superblock.ApfsNumOtherFsobjects
}
