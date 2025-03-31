package spacemanager

import (
	"github.com/deploymenttheory/go-apfs/apfs/types"
)

// Manager wraps types.SpacemanPhysT to provide helper methods.
type Manager struct {
	*types.SpacemanPhysT
}

// IsVersioned returns true if the SmFlagVersioned flag is set.
//
// This indicates that the space manager uses versioning to track metadata.
// Reference: APFS Reference, page 162.
func (sm *Manager) IsVersioned() bool {
	return sm.SmFlags&types.SmFlagVersioned != 0
}

// InternalPoolFreeBlockCount returns the number of free blocks
// available in the internal pool bitmap.
//
// Reference: APFS Reference, page 161–163.
func (sm *Manager) InternalPoolFreeBlockCount() uint64 {
	return sm.SmIpBlockCount - uint64(sm.SmIpBmBlockCount)
}

// HasFusionDevice returns true if both the main and tier2 devices are present.
//
// Reference: APFS Reference, page 162.
func (sm *Manager) HasFusionDevice() bool {
	return sm.SmDev[types.SdTier2].SmBlockCount > 0
}

// MainDevice returns the SpacemanDeviceT for the main SSD device.
func (sm *Manager) MainDevice() *types.SpacemanDeviceT {
	return &sm.SmDev[types.SdMain]
}

// Tier2Device returns the SpacemanDeviceT for the tier2 HDD device.
func (sm *Manager) Tier2Device() *types.SpacemanDeviceT {
	return &sm.SmDev[types.SdTier2]
}
