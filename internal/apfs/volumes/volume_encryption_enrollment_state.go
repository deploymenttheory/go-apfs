package volumes

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// volumeEncryptionRollingState implements the VolumeEncryptionRollingState interface
type volumeEncryptionRollingState struct {
	superblock *types.ApfsSuperblockT
}

// NewVolumeEncryptionRollingState creates a new VolumeEncryptionRollingState implementation
func NewVolumeEncryptionRollingState(superblock *types.ApfsSuperblockT) interfaces.VolumeEncryptionRollingState {
	return &volumeEncryptionRollingState{
		superblock: superblock,
	}
}

// EncryptionRollingStateOID returns the encryption rolling state object identifier
func (vers *volumeEncryptionRollingState) EncryptionRollingStateOID() types.OidT {
	return vers.superblock.ApfsErStateOid
}

// IsEncryptionRollingInProgress checks if encryption rolling is currently in progress
func (vers *volumeEncryptionRollingState) IsEncryptionRollingInProgress() bool {
	return vers.superblock.ApfsErStateOid != types.OidInvalid && vers.superblock.ApfsErStateOid != 0
}
