package objects

import (
	"github.com/deploymenttheory/go-apfs/internal/types"
)

type typeInfo struct {
	name     string
	category string
}

type StaticObjectTypeResolver struct {
	registry map[uint32]typeInfo
}

func NewStaticObjectTypeResolver() *StaticObjectTypeResolver {
	return &StaticObjectTypeResolver{
		registry: map[uint32]typeInfo{
			types.ObjectTypeNxSuperblock:    {"NX Superblock", "Container"},
			types.ObjectTypeBtree:           {"B-tree Root", "Metadata"},
			types.ObjectTypeBtreeNode:       {"B-tree Node", "Metadata"},
			types.ObjectTypeFs:              {"APFS Volume", "File System"},
			types.ObjectTypeOmap:            {"Object Map", "Metadata"},
			types.ObjectTypeSnapmetatree:    {"Snapshot Metadata Tree", "File System"},
			types.ObjectTypeContainerKeybag: {"Container Keybag", "Security"}, // 'keys'
			types.ObjectTypeVolumeKeybag:    {"Volume Keybag", "Security"},    // 'recs'
			types.ObjectTypeMediaKeybag:     {"Media Keybag", "Security"},     // 'mkey'
		},
	}
}

func (r *StaticObjectTypeResolver) ResolveType(objectType uint32) string {
	// Check unmasked value first (for string-based constants like 'mkey')
	if info, ok := r.registry[objectType]; ok {
		return info.name
	}

	// Fall back to masked type (for numerical types with flags)
	base := objectType & types.ObjectTypeMask
	if info, ok := r.registry[base]; ok {
		return info.name
	}

	return "Unknown"
}

func (r *StaticObjectTypeResolver) SupportedObjectTypes() []uint32 {
	types := make([]uint32, 0, len(r.registry))
	for t := range r.registry {
		types = append(types, t)
	}
	return types
}

func (r *StaticObjectTypeResolver) GetObjectTypeCategory(objectType uint32) string {
	if info, ok := r.registry[objectType]; ok {
		return info.category
	}

	base := objectType & types.ObjectTypeMask
	if info, ok := r.registry[base]; ok {
		return info.category
	}

	return "Unknown"
}
