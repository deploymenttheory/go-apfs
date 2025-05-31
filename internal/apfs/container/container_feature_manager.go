package container

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// containerFeatureManager implements the ContainerFeatureManager interface
type containerFeatureManager struct {
	superblock *types.NxSuperblockT
}

// NewContainerFeatureManager creates a new ContainerFeatureManager implementation
func NewContainerFeatureManager(superblock *types.NxSuperblockT) interfaces.ContainerFeatureManager {
	return &containerFeatureManager{
		superblock: superblock,
	}
}

// Features returns the optional features being used by the container
func (cfm *containerFeatureManager) Features() uint64 {
	return cfm.superblock.NxFeatures
}

// ReadOnlyCompatibleFeatures returns the read-only compatible features being used
func (cfm *containerFeatureManager) ReadOnlyCompatibleFeatures() uint64 {
	return cfm.superblock.NxReadonlyCompatibleFeatures
}

// IncompatibleFeatures returns the backward-incompatible features being used
func (cfm *containerFeatureManager) IncompatibleFeatures() uint64 {
	return cfm.superblock.NxIncompatibleFeatures
}

// HasFeature checks if a specific optional feature is enabled
func (cfm *containerFeatureManager) HasFeature(feature uint64) bool {
	return cfm.superblock.NxFeatures&feature != 0
}

// HasReadOnlyCompatibleFeature checks if a specific read-only compatible feature is enabled
func (cfm *containerFeatureManager) HasReadOnlyCompatibleFeature(feature uint64) bool {
	return cfm.superblock.NxReadonlyCompatibleFeatures&feature != 0
}

// HasIncompatibleFeature checks if a specific incompatible feature is enabled
func (cfm *containerFeatureManager) HasIncompatibleFeature(feature uint64) bool {
	return cfm.superblock.NxIncompatibleFeatures&feature != 0
}

// SupportsDefragmentation checks if the container supports defragmentation
func (cfm *containerFeatureManager) SupportsDefragmentation() bool {
	return cfm.HasFeature(types.NxFeatureDefrag)
}

// IsLowCapacityFusionDrive checks if the container is using low-capacity Fusion Drive mode
func (cfm *containerFeatureManager) IsLowCapacityFusionDrive() bool {
	return cfm.HasFeature(types.NxFeatureLcfd)
}

// GetAPFSVersion returns the APFS version used by the container
func (cfm *containerFeatureManager) GetAPFSVersion() string {
	if cfm.HasIncompatibleFeature(types.NxIncompatVersion2) {
		return "2.0"
	}
	if cfm.HasIncompatibleFeature(types.NxIncompatVersion1) {
		return "1.0"
	}
	return "Unknown"
}

// SupportsFusion checks if the container supports Fusion Drives
func (cfm *containerFeatureManager) SupportsFusion() bool {
	return cfm.HasIncompatibleFeature(types.NxIncompatFusion)
}
