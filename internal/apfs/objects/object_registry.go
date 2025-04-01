package objects

import (
	"github.com/deploymenttheory/go-apfs/internal/types"
)

type StaticObjectRegistry struct {
	registry map[uint32]ObjectTypeInfo
}

// ObjectTypeInfo contains detailed information about an APFS object type
type ObjectTypeInfo struct {
	// Numeric type identifier
	Type uint32

	// Human-readable name
	Name string

	// Detailed description
	Description string

	// Categorization (e.g., metadata, file system, container)
	Category string
}

// NewStaticObjectRegistry initializes a hardcoded APFS object registry
func NewStaticObjectRegistry() *StaticObjectRegistry {
	return &StaticObjectRegistry{
		registry: map[uint32]ObjectTypeInfo{
			types.ObjectTypeNxSuperblock: {
				Type:        types.ObjectTypeNxSuperblock,
				Name:        "NX Superblock",
				Description: "Container superblock (nx_superblock_t)",
				Category:    "Container",
			},
			types.ObjectTypeBtree: {
				Type:        types.ObjectTypeBtree,
				Name:        "B-tree Root",
				Description: "B-tree root node (btree_node_phys_t)",
				Category:    "Metadata",
			},
			types.ObjectTypeBtreeNode: {
				Type:        types.ObjectTypeBtreeNode,
				Name:        "B-tree Node",
				Description: "B-tree child node (btree_node_phys_t)",
				Category:    "Metadata",
			},
			types.ObjectTypeFs: {
				Type:        types.ObjectTypeFs,
				Name:        "APFS Volume",
				Description: "APFS volume superblock (apfs_superblock_t)",
				Category:    "File System",
			},
			types.ObjectTypeOmap: {
				Type:        types.ObjectTypeOmap,
				Name:        "Object Map",
				Description: "Mapping of virtual to physical object identifiers",
				Category:    "Metadata",
			},
			types.ObjectTypeContainerKeybag: {
				Type:        types.ObjectTypeContainerKeybag,
				Name:        "Container Keybag",
				Description: "Keybag for container-level encryption",
				Category:    "Security",
			},
			types.ObjectTypeVolumeKeybag: {
				Type:        types.ObjectTypeVolumeKeybag,
				Name:        "Volume Keybag",
				Description: "Keybag for volume-level encryption",
				Category:    "Security",
			},
			types.ObjectTypeMediaKeybag: {
				Type:        types.ObjectTypeMediaKeybag,
				Name:        "Media Keybag",
				Description: "Keybag for media encryption",
				Category:    "Security",
			},
		},
	}
}

func (r *StaticObjectRegistry) LookupType(objectType uint32) (ObjectTypeInfo, bool) {
	// Attempt full-width match first (e.g. for 'mkey' style types)
	if info, ok := r.registry[objectType]; ok {
		return info, true
	}

	// Fallback to base type for masked numeric types
	base := objectType & types.ObjectTypeMask
	if info, ok := r.registry[base]; ok {
		return info, true
	}

	return ObjectTypeInfo{}, false
}

func (r *StaticObjectRegistry) ListObjectTypes() []ObjectTypeInfo {
	list := make([]ObjectTypeInfo, 0, len(r.registry))
	for _, info := range r.registry {
		list = append(list, info)
	}
	return list
}
