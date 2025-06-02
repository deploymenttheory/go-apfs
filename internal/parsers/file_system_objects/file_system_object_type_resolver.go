package file_system_objects

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// fileSystemObjectTypeResolver implements the FileSystemObjectTypeResolver interface
type fileSystemObjectTypeResolver struct{}

// NewFileSystemObjectTypeResolver creates a new file system object type resolver
func NewFileSystemObjectTypeResolver() interfaces.FileSystemObjectTypeResolver {
	return &fileSystemObjectTypeResolver{}
}

// ResolveObjectType converts a raw object type to a human-readable description
func (fstr *fileSystemObjectTypeResolver) ResolveObjectType(objType types.JObjTypes) string {
	switch objType {
	case types.ApfsTypeAny:
		return "Any"
	case types.ApfsTypeSnapMetadata:
		return "Snapshot Metadata"
	case types.ApfsTypeExtent:
		return "Physical Extent"
	case types.ApfsTypeInode:
		return "Inode"
	case types.ApfsTypeXattr:
		return "Extended Attribute"
	case types.ApfsTypeSiblingLink:
		return "Sibling Link"
	case types.ApfsTypeDstreamId:
		return "Data Stream"
	case types.ApfsTypeCryptoState:
		return "Crypto State"
	case types.ApfsTypeFileExtent:
		return "File Extent"
	case types.ApfsTypeDirRec:
		return "Directory Entry"
	case types.ApfsTypeDirStats:
		return "Directory Statistics"
	case types.ApfsTypeSnapName:
		return "Snapshot Name"
	case types.ApfsTypeSiblingMap:
		return "Sibling Map"
	case types.ApfsTypeFileInfo:
		return "File Info"
	case types.ApfsTypeInvalid:
		return "Invalid"
	default:
		return "Unknown"
	}
}

// ResolveObjectKind converts a raw object kind to a human-readable description
func (fstr *fileSystemObjectTypeResolver) ResolveObjectKind(objKind types.JObjKinds) string {
	switch objKind {
	case types.ApfsKindAny:
		return "Any"
	case types.ApfsKindNew:
		return "New"
	case types.ApfsKindUpdate:
		return "Update"
	case types.ApfsKindDead:
		return "Dead"
	default:
		return "Unknown"
	}
}

// ListSupportedObjectTypes returns all supported object types
func (fstr *fileSystemObjectTypeResolver) ListSupportedObjectTypes() []types.JObjTypes {
	return []types.JObjTypes{
		types.ApfsTypeSnapMetadata,
		types.ApfsTypeExtent,
		types.ApfsTypeInode,
		types.ApfsTypeXattr,
		types.ApfsTypeSiblingLink,
		types.ApfsTypeDstreamId,
		types.ApfsTypeCryptoState,
		types.ApfsTypeFileExtent,
		types.ApfsTypeDirRec,
		types.ApfsTypeDirStats,
		types.ApfsTypeSnapName,
		types.ApfsTypeSiblingMap,
		types.ApfsTypeFileInfo,
	}
}
