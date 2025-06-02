package volumes

import (
	"strings"
	"unicode/utf8"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// volumeIdentity implements the VolumeIdentity interface
type volumeIdentity struct {
	superblock *types.ApfsSuperblockT
}

// NewVolumeIdentity creates a new VolumeIdentity implementation
func NewVolumeIdentity(superblock *types.ApfsSuperblockT) interfaces.VolumeIdentity {
	return &volumeIdentity{
		superblock: superblock,
	}
}

// UUID returns the unique volume identifier
func (vi *volumeIdentity) UUID() types.UUID {
	return vi.superblock.ApfsVolUuid
}

// Name returns the volume name as a string
func (vi *volumeIdentity) Name() string {
	// Convert byte array to string, trimming null terminators
	name := string(vi.superblock.ApfsVolname[:])
	// Trim null bytes and any trailing spaces
	name = strings.TrimRight(name, "\x00 ")

	// Validate UTF-8
	if utf8.ValidString(name) {
		return name
	}

	// Fallback if name is not valid UTF-8
	return "[Invalid Volume Name]"
}

// Role returns the raw role value
func (vi *volumeIdentity) Role() uint16 {
	return vi.superblock.ApfsRole
}

// RoleName provides a human-readable description of the volume's role
func (vi *volumeIdentity) RoleName() string {
	switch vi.superblock.ApfsRole {
	case types.ApfsVolRoleNone:
		return "None"
	case types.ApfsVolRoleSystem:
		return "System"
	case types.ApfsVolRoleUser:
		return "User"
	case types.ApfsVolRoleRecovery:
		return "Recovery"
	case types.ApfsVolRoleVm:
		return "Virtual Memory"
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
		return "XART (Secure User Data)"
	case types.ApfsVolRoleHardware:
		return "Hardware"
	case types.ApfsVolRoleBackup:
		return "Backup"
	case types.ApfsVolRoleEnterprise:
		return "Enterprise"
	case types.ApfsVolRolePrelogin:
		return "Prelogin"
	default:
		return "Unknown"
	}
}

// Index returns the index in the container's filesystem array
func (vi *volumeIdentity) Index() uint32 {
	return vi.superblock.ApfsFsIndex
}
