// File: internal/interfaces/siblings.go
package interfaces

// SiblingLinkReader provides information about a sibling link
type SiblingLinkReader interface {
	// SiblingID returns the unique identifier for this sibling
	SiblingID() uint64

	// InodeNumber returns the original inode number
	InodeNumber() uint64

	// ParentDirectoryID returns the object identifier of the parent directory
	ParentDirectoryID() uint64

	// Name returns the name of the sibling link
	Name() string
}

// SiblingMapReader provides mapping information for siblings
type SiblingMapReader interface {
	// SiblingID returns the unique identifier for this sibling map
	SiblingID() uint64

	// FileID returns the inode number of the underlying file
	FileID() uint64
}

// SiblingManager provides methods for managing and querying sibling links
type SiblingManager interface {
	// ListSiblingLinks returns all sibling links for a given inode
	ListSiblingLinks(inodeNumber uint64) ([]SiblingLinkReader, error)

	// FindSiblingLinkByName finds a sibling link by its name
	FindSiblingLinkByName(name string) (SiblingLinkReader, error)

	// FindSiblingMapByID finds a sibling map by its unique identifier
	FindSiblingMapByID(siblingID uint64) (SiblingMapReader, error)

	// CountSiblingLinks counts the number of sibling links for a given inode
	CountSiblingLinks(inodeNumber uint64) (int, error)
}

// HardLinkInfo provides comprehensive information about hard links
type HardLinkInfo interface {
	// OriginalInode returns the original inode number
	OriginalInode() uint64

	// LinkedFiles returns all file paths linked to this inode
	LinkedFiles() ([]string, error)

	// IsHardLink checks if the file is a hard link
	IsHardLink() bool
}
