package volumes

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// volumeFeatures implements the VolumeFeatures interface
type volumeFeatures struct {
	superblock *types.ApfsSuperblockT
}

// Features returns the optional feature flags
func (vf *volumeFeatures) Features() uint64 {
	return vf.superblock.ApfsFeatures
}

// ReadonlyCompatibleFeatures returns the read-only compatible feature flags
func (vf *volumeFeatures) ReadonlyCompatibleFeatures() uint64 {
	return vf.superblock.ApfsReadonlyCompatibleFeatures
}

// IncompatibleFeatures returns the incompatible feature flags
func (vf *volumeFeatures) IncompatibleFeatures() uint64 {
	return vf.superblock.ApfsIncompatibleFeatures
}

// SupportsDefragmentation checks if defragmentation is supported
func (vf *volumeFeatures) SupportsDefragmentation() bool {
	return vf.superblock.ApfsFeatures&types.ApfsFeatureDefrag != 0
}

// SupportsHardlinkMapRecords checks if hardlink map records are supported
func (vf *volumeFeatures) SupportsHardlinkMapRecords() bool {
	return vf.superblock.ApfsFeatures&types.ApfsFeatureHardlinkMapRecords != 0
}

// IsStrictAccessTimeEnabled checks if strict access time is enabled
func (vf *volumeFeatures) IsStrictAccessTimeEnabled() bool {
	return vf.superblock.ApfsFeatures&types.ApfsFeatureStrictatime != 0
}

// IsCaseInsensitive checks if the volume is case insensitive
func (vf *volumeFeatures) IsCaseInsensitive() bool {
	return vf.superblock.ApfsIncompatibleFeatures&types.ApfsIncompatCaseInsensitive != 0
}

// IsNormalizationInsensitive checks if the volume is normalization insensitive
func (vf *volumeFeatures) IsNormalizationInsensitive() bool {
	return vf.superblock.ApfsIncompatibleFeatures&types.ApfsIncompatNormalizationInsensitive != 0
}

// IsSealed checks if the volume is sealed
func (vf *volumeFeatures) IsSealed() bool {
	return vf.superblock.ApfsIncompatibleFeatures&types.ApfsIncompatSealedVolume != 0
}

// IsUnencrypted checks if the volume is unencrypted
func (vf *volumeFeatures) IsUnencrypted() bool {
	return vf.superblock.ApfsFsFlags&types.ApfsFsUnencrypted != 0
}

// IsOneKeyEncryption checks if all files are encrypted with the volume encryption key
func (vf *volumeFeatures) IsOneKeyEncryption() bool {
	return vf.superblock.ApfsFsFlags&types.ApfsFsOnekey != 0
}

// IsSpilledOver checks if the volume has run out of allocated space on the SSD
func (vf *volumeFeatures) IsSpilledOver() bool {
	return vf.superblock.ApfsFsFlags&types.ApfsFsSpilledover != 0
}

// RequiresSpilloverCleaner checks if the volume requires a spillover cleaner
func (vf *volumeFeatures) RequiresSpilloverCleaner() bool {
	return vf.superblock.ApfsFsFlags&types.ApfsFsRunSpilloverCleaner != 0
}

// AlwaysChecksExtentReference checks if the extent reference tree is always consulted
func (vf *volumeFeatures) AlwaysChecksExtentReference() bool {
	return vf.superblock.ApfsFsFlags&types.ApfsFsAlwaysCheckExtentref != 0
}

// NewVolumeFeatures creates a new VolumeFeatures implementation
func NewVolumeFeatures(superblock *types.ApfsSuperblockT) interfaces.VolumeFeatures {
	return &volumeFeatures{
		superblock: superblock,
	}
}
