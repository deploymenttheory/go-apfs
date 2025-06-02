package volumes

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// volumeEncryptionMetadata implements the VolumeEncryptionMetadata interface
type volumeEncryptionMetadata struct {
	superblock *types.ApfsSuperblockT
}

// MetadataCryptoState returns the encryption state information for the volume's metadata
func (vem *volumeEncryptionMetadata) MetadataCryptoState() types.WrappedMetaCryptoStateT {
	return vem.superblock.ApfsMetaCrypto
}

// IsEncrypted checks if the volume is encrypted
func (vem *volumeEncryptionMetadata) IsEncrypted() bool {
	// Check if the volume is not explicitly marked as unencrypted
	return (vem.superblock.ApfsFsFlags & types.ApfsFsUnencrypted) == 0
}

// HasEncryptionKeyRotated checks if the volume's encryption has changed keys
func (vem *volumeEncryptionMetadata) HasEncryptionKeyRotated() bool {
	// Check the incompatible features flag for encryption key rotation
	return (vem.superblock.ApfsIncompatibleFeatures & types.ApfsIncompatEncRolled) != 0
}

// NewVolumeEncryptionMetadata creates a new VolumeEncryptionMetadata implementation
func NewVolumeEncryptionMetadata(superblock *types.ApfsSuperblockT) interfaces.VolumeEncryptionMetadata {
	return &volumeEncryptionMetadata{
		superblock: superblock,
	}
}
