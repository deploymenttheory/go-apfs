// File: internal/interfaces/file_system_objects.go
package interfaces

import (
	"time"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// FileSystemObjectReader provides basic information about a file system object
type FileSystemObjectReader interface {
	// ObjectType returns the type of the file system object
	ObjectType() types.JObjTypes

	// ObjectKind returns the kind of the file system object
	ObjectKind() types.JObjKinds

	// ObjectIdentifier returns the object's unique identifier
	ObjectIdentifier() uint64
}

// InodeReader provides methods for reading inode information
type InodeReader interface {
	FileSystemObjectReader

	// ParentID returns the identifier of the parent directory
	ParentID() uint64

	// PrivateID returns the unique identifier for the file's data stream
	PrivateID() uint64

	// CreationTime returns the time the inode was created
	CreationTime() time.Time

	// ModificationTime returns the time the inode was last modified
	ModificationTime() time.Time

	// ChangeTime returns the time the inode's attributes were last modified
	ChangeTime() time.Time

	// AccessTime returns the time the inode was last accessed
	AccessTime() time.Time

	// Flags returns the inode's internal flags
	Flags() types.JInodeFlags

	// Owner returns the user identifier of the inode's owner
	Owner() types.UidT

	// Group returns the group identifier of the inode's group
	Group() types.GidT

	// Mode returns the file's mode
	Mode() types.ModeT

	// IsDirectory checks if the inode represents a directory
	IsDirectory() bool

	// NumberOfChildren returns the number of directory entries (for directories)
	NumberOfChildren() int32

	// NumberOfHardLinks returns the number of hard links (for files)
	NumberOfHardLinks() int32

	// HasResourceFork checks if the inode has a resource fork
	HasResourceFork() bool

	// Size returns the file size
	Size() uint64
}

// DirectoryEntryReader provides methods for reading directory entry information
type DirectoryEntryReader interface {
	FileSystemObjectReader

	// FileName returns the name of the directory entry
	FileName() string

	// FileID returns the identifier of the inode this entry represents
	FileID() uint64

	// DateAdded returns the time this directory entry was added
	DateAdded() time.Time

	// FileType returns the type of file (directory, regular file, symlink, etc.)
	FileType() uint16
}

// DirectoryStatsReader provides methods for reading directory statistics information
type DirectoryStatsReader interface {
	FileSystemObjectReader

	// NumChildren returns the number of files and folders in the directory
	NumChildren() uint64

	// TotalSize returns the total size of all files in the directory and descendants
	TotalSize() uint64

	// ChainedKey returns the parent directory's file system object identifier
	ChainedKey() uint64

	// GenCount returns the generation counter
	GenCount() uint64
}

// ExtendedAttributeReader provides methods for reading extended attribute information
type ExtendedAttributeReader interface {
	FileSystemObjectReader

	// AttributeName returns the name of the extended attribute
	AttributeName() string

	// IsDataEmbedded checks if the attribute data is embedded in the record
	IsDataEmbedded() bool

	// IsDataStream checks if the attribute data is stored in a data stream
	IsDataStream() bool

	// IsFileSystemOwned checks if the attribute is owned by the file system
	IsFileSystemOwned() bool

	// Data returns the attribute's data
	Data() []byte
}

// FileSystemObjectTypeResolver provides methods for resolving file system object types
type FileSystemObjectTypeResolver interface {
	// ResolveObjectType converts a raw object type to a human-readable description
	ResolveObjectType(objType types.JObjTypes) string

	// ResolveObjectKind converts a raw object kind to a human-readable description
	ResolveObjectKind(objKind types.JObjKinds) string

	// ListSupportedObjectTypes returns all supported object types
	ListSupportedObjectTypes() []types.JObjTypes
}
