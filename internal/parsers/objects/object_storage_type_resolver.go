package objects

import (
	"strings"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

type StaticObjectStorageTypeResolver struct{}

// NewStaticObjectStorageTypeResolver returns a resolver that knows about APFS storage type flags.
func NewStaticObjectStorageTypeResolver() *StaticObjectStorageTypeResolver {
	return &StaticObjectStorageTypeResolver{}
}

func (r *StaticObjectStorageTypeResolver) DetermineStorageType(objectType uint32) string {
	storageBits := objectType & types.ObjStorageTypeMask

	switch storageBits {
	case types.ObjVirtual:
		return "virtual"
	case types.ObjEphemeral:
		return "ephemeral"
	case types.ObjPhysical:
		return "physical"
	default:
		// If multiple storage bits are set (e.g. physical + ephemeral), that's invalid
		return "unknown"
	}
}

func (r *StaticObjectStorageTypeResolver) IsStorageTypeSupported(storageType string) bool {
	switch strings.ToLower(storageType) {
	case "virtual", "ephemeral", "physical":
		return true
	default:
		return false
	}
}
