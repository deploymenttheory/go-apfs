package volumes

import (
	"strings"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// Volume wraps types.ApfsSuperblockT to provide helper methods.
type Volume struct {
	*types.ApfsSuperblockT
}

// Name returns the volume name as a string, stripping the null terminator.
func (v *Volume) Name() string {
	return strings.TrimRight(string(v.ApfsVolname[:]), "\x00")
}

// IsEncrypted returns true if the volume is encrypted (i.e., ApfsFsUnencrypted is not set).
func (v *Volume) IsEncrypted() bool {
	return v.ApfsFsFlags&types.ApfsFsUnencrypted == 0
}

// HasFeature checks if the volume has a specific optional feature enabled.
func (v *Volume) HasFeature(flag uint64) bool {
	return v.ApfsFeatures&flag != 0
}

// HasIncompatibleFeature checks if a given incompatibility flag is enabled.
func (v *Volume) HasIncompatibleFeature(flag uint64) bool {
	return v.ApfsIncompatibleFeatures&flag != 0
}

// RoleName returns a human-readable name for the volume's role.
func (v *Volume) RoleName() string {
	switch v.ApfsRole {
	case types.ApfsVolRoleSystem:
		return "System"
	case types.ApfsVolRoleUser:
		return "User"
	case types.ApfsVolRoleRecovery:
		return "Recovery"
	case types.ApfsVolRoleVm:
		return "VM"
	case types.ApfsVolRolePreboot:
		return "Preboot"
	case types.ApfsVolRoleInstaller:
		return "Installer"
	case types.ApfsVolRoleData:
		return "Data"
	case types.ApfsVolRoleBaseband:
		return "Baseband"
	case types.ApfsVolRoleUpdate:
		return "Update"
	case types.ApfsVolRoleXart:
		return "XART"
	case types.ApfsVolRoleHardware:
		return "Hardware"
	case types.ApfsVolRoleBackup:
		return "Backup"
	case types.ApfsVolRoleEnterprise:
		return "Enterprise"
	case types.ApfsVolRolePrelogin:
		return "Prelogin"
	default:
		return "None"
	}
}
