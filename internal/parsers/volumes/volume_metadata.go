package volumes

import (
	"time"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// volumeMetadata implements the VolumeMetadata interface
type volumeMetadata struct {
	superblock *types.ApfsSuperblockT
}

// LastUnmountTime returns the time the volume was last unmounted
func (vm *volumeMetadata) LastUnmountTime() time.Time {
	// Convert nanoseconds since epoch to time.Time
	return time.Unix(0, int64(vm.superblock.ApfsUnmountTime))
}

// LastModifiedTime returns the time the volume was last modified
func (vm *volumeMetadata) LastModifiedTime() time.Time {
	// Convert nanoseconds since epoch to time.Time
	return time.Unix(0, int64(vm.superblock.ApfsLastModTime))
}

// FormattedBy returns information about the software that created the volume
func (vm *volumeMetadata) FormattedBy() types.ApfsModifiedByT {
	return vm.superblock.ApfsFormattedBy
}

// ModificationHistory returns the history of modifications to the volume
func (vm *volumeMetadata) ModificationHistory() []types.ApfsModifiedByT {
	// Return a copy of the modification history to prevent direct modification
	history := make([]types.ApfsModifiedByT, len(vm.superblock.ApfsModifiedBy))
	copy(history, vm.superblock.ApfsModifiedBy[:])
	return history
}

// NextObjectID returns the next identifier to be assigned to a file-system object
func (vm *volumeMetadata) NextObjectID() uint64 {
	return vm.superblock.ApfsNextObjId
}

// NextDocumentID returns the next document identifier to be assigned
func (vm *volumeMetadata) NextDocumentID() uint32 {
	return vm.superblock.ApfsNextDocId
}

// NewVolumeMetadata creates a new VolumeMetadata implementation
func NewVolumeMetadata(superblock *types.ApfsSuperblockT) interfaces.VolumeMetadata {
	return &volumeMetadata{
		superblock: superblock,
	}
}
