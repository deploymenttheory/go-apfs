// File: internal/interfaces/filesystem.go
package interfaces

import (
	"io"
	"time"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// FileSystemNavigator provides methods for navigating the filesystem tree
type FileSystemNavigator interface {
	// GetRootDirectory returns the root directory inode
	GetRootDirectory() (InodeReader, error)

	// ReadDirectory returns all entries in a directory
	ReadDirectory(inodeID uint64) ([]DirectoryEntryReader, error)

	// FindFile finds a file by its full path
	FindFile(path string) (InodeReader, error)

	// FindDirectory finds a directory by its full path
	FindDirectory(path string) (InodeReader, error)

	// ListFiles returns all files in a directory (non-recursive)
	ListFiles(directoryID uint64) ([]InodeReader, error)

	// ListDirectories returns all subdirectories in a directory
	ListDirectories(directoryID uint64) ([]InodeReader, error)

	// ResolvePath resolves a path to an inode ID
	ResolvePath(path string) (uint64, error)

	// GetFullPath constructs the full path for a given inode
	GetFullPath(inodeID uint64) (string, error)

	// GetParentDirectory returns the parent directory of a given inode
	GetParentDirectory(inodeID uint64) (InodeReader, error)

	// WalkFileSystem performs a recursive walk of the filesystem
	WalkFileSystem(rootInodeID uint64, walkFunc FileSystemWalkFunc) error
}

// FileSystemWalkFunc is called for each file and directory during filesystem traversal
type FileSystemWalkFunc func(path string, inode InodeReader, isDirectory bool) error

// PathResolver provides methods for path resolution and manipulation
type PathResolver interface {
	// NormalizePath normalizes a filesystem path
	NormalizePath(path string) string

	// JoinPath joins path components together
	JoinPath(components ...string) string

	// SplitPath splits a path into its components
	SplitPath(path string) []string

	// GetFileName returns the filename portion of a path
	GetFileName(path string) string

	// GetDirectoryPath returns the directory portion of a path
	GetDirectoryPath(path string) string

	// IsAbsolutePath checks if a path is absolute
	IsAbsolutePath(path string) bool

	// ValidatePath checks if a path is valid for the filesystem
	ValidatePath(path string) error
}

// FileDataReader provides methods for reading file content
type FileDataReader interface {
	// ReadFile reads the entire content of a file
	ReadFile(inodeID uint64) ([]byte, error)

	// ReadFileRange reads a specific range of bytes from a file
	ReadFileRange(inodeID uint64, offset, length uint64) ([]byte, error)

	// GetFileSize returns the size of a file in bytes
	GetFileSize(inodeID uint64) (uint64, error)

	// CreateFileReader creates an io.Reader for streaming file content
	CreateFileReader(inodeID uint64) (io.Reader, error)

	// CreateFileSeeker creates an io.ReadSeeker for random access to file content
	CreateFileSeeker(inodeID uint64) (io.ReadSeeker, error)

	// GetFileExtents returns all extents for a file
	GetFileExtents(inodeID uint64) ([]FileExtentReader, error)

	// VerifyFileChecksum verifies the integrity of a file's data
	VerifyFileChecksum(inodeID uint64) (bool, error)
}

// FileMetadataReader provides methods for reading file metadata
type FileMetadataReader interface {
	// GetFilePermissions returns file permissions
	GetFilePermissions(inodeID uint64) (types.ModeT, error)

	// GetFileOwnership returns owner and group information
	GetFileOwnership(inodeID uint64) (types.UidT, types.GidT, error)

	// GetFileTimes returns creation, modification, and access times
	GetFileTimes(inodeID uint64) (creation time.Time, modification time.Time, access time.Time, err error)

	// GetFileFlags returns BSD-style file flags
	GetFileFlags(inodeID uint64) (uint32, error)

	// IsSymlink checks if a file is a symbolic link
	IsSymlink(inodeID uint64) (bool, error)

	// ReadSymlink reads the target of a symbolic link
	ReadSymlink(inodeID uint64) (string, error)

	// GetHardLinkCount returns the number of hard links to a file
	GetHardLinkCount(inodeID uint64) (uint32, error)

	// FindHardLinks finds all hard links to a file
	FindHardLinks(inodeID uint64) ([]string, error)
}

// ExtendedAttributeManager provides methods for managing extended attributes
type ExtendedAttributeManager interface {
	// ListExtendedAttributes returns all extended attribute names for a file
	ListExtendedAttributes(inodeID uint64) ([]string, error)

	// ReadExtendedAttribute reads the value of an extended attribute
	ReadExtendedAttribute(inodeID uint64, name string) ([]byte, error)

	// HasExtendedAttribute checks if a file has a specific extended attribute
	HasExtendedAttribute(inodeID uint64, name string) (bool, error)

	// GetExtendedAttributeSize returns the size of an extended attribute
	GetExtendedAttributeSize(inodeID uint64, name string) (uint64, error)

	// GetResourceFork reads the resource fork of a file (macOS specific)
	GetResourceFork(inodeID uint64) ([]byte, error)

	// HasResourceFork checks if a file has a resource fork
	HasResourceFork(inodeID uint64) (bool, error)
}

// DirectoryStatisticsReader provides methods for reading directory statistics
type DirectoryStatisticsReader interface {
	// GetDirectoryStats returns comprehensive statistics for a directory
	GetDirectoryStats(inodeID uint64) (DirectoryStats, error)

	// GetChildCount returns the number of immediate children in a directory
	GetChildCount(inodeID uint64) (uint64, error)

	// GetTotalSize returns the total size of all files in a directory tree
	GetTotalSize(inodeID uint64) (uint64, error)

	// GetFileCount returns the total number of files in a directory tree
	GetFileCount(inodeID uint64) (uint64, error)
}

// DirectoryStats contains comprehensive statistics about a directory
type DirectoryStats struct {
	// Number of immediate children
	ChildCount uint64

	// Total size of all files in the directory tree
	TotalSize uint64

	// Number of files in the directory tree
	FileCount uint64

	// Number of subdirectories in the directory tree
	DirectoryCount uint64

	// Deepest nesting level in the directory tree
	MaxDepth int

	// Last modification time of any file in the tree
	LastModified time.Time
}

// FileSystemSearch provides methods for searching the filesystem
type FileSystemSearch interface {
	// FindFilesByName finds files matching a name pattern
	FindFilesByName(pattern string) ([]InodeReader, error)

	// FindFilesByExtension finds files with a specific extension
	FindFilesByExtension(extension string) ([]InodeReader, error)

	// FindFilesBySize finds files within a size range
	FindFilesBySize(minSize, maxSize uint64) ([]InodeReader, error)

	// FindFilesByDate finds files modified within a date range
	FindFilesByDate(since, until time.Time) ([]InodeReader, error)

	// FindFilesByType finds files of a specific type
	FindFilesByType(fileType types.ModeT) ([]InodeReader, error)

	// FindLargestFiles finds the largest files in the filesystem
	FindLargestFiles(count int) ([]FileSearchResult, error)

	// FindNewestFiles finds the most recently modified files
	FindNewestFiles(count int) ([]FileSearchResult, error)

	// FindOldestFiles finds the oldest files in the filesystem
	FindOldestFiles(count int) ([]FileSearchResult, error)
}

// FileSearchResult represents a file found during a search operation
type FileSearchResult struct {
	// The inode of the found file
	Inode InodeReader

	// The full path to the file
	Path string

	// Size of the file
	Size uint64

	// Last modification time
	ModificationTime time.Time

	// File type
	FileType types.ModeT
}

// FileSystemIntegrityChecker provides methods for checking filesystem integrity
type FileSystemIntegrityChecker interface {
	// CheckFileIntegrity verifies the integrity of a specific file
	CheckFileIntegrity(inodeID uint64) (FileIntegrityResult, error)

	// CheckDirectoryIntegrity verifies the integrity of a directory and its contents
	CheckDirectoryIntegrity(inodeID uint64, recursive bool) (DirectoryIntegrityResult, error)

	// CheckFilesystemIntegrity performs a comprehensive integrity check
	CheckFilesystemIntegrity() (FilesystemIntegrityResult, error)

	// VerifyInodeConsistency checks if an inode's metadata is consistent
	VerifyInodeConsistency(inodeID uint64) (bool, []string, error)
}

// FileIntegrityResult contains the result of checking a file's integrity
type FileIntegrityResult struct {
	// The inode ID that was checked
	InodeID uint64

	// Whether the file passed all integrity checks
	IsValid bool

	// List of any integrity issues found
	Issues []IntegrityIssue

	// Checksum verification result
	ChecksumValid bool

	// Whether all extents are accessible
	ExtentsAccessible bool
}

// DirectoryIntegrityResult contains the result of checking a directory's integrity
type DirectoryIntegrityResult struct {
	// The directory inode ID that was checked
	InodeID uint64

	// Whether the directory structure is valid
	IsValid bool

	// Issues found with the directory itself
	DirectoryIssues []IntegrityIssue

	// Results for individual files (if recursive check was performed)
	FileResults []FileIntegrityResult

	// Number of files checked
	FilesChecked int

	// Number of files with issues
	FilesWithIssues int
}

// FilesystemIntegrityResult contains the result of checking the entire filesystem
type FilesystemIntegrityResult struct {
	// Whether the filesystem passed all integrity checks
	IsValid bool

	// Total number of files checked
	TotalFiles int

	// Number of files with issues
	FilesWithIssues int

	// Issues found at the filesystem level
	FilesystemIssues []IntegrityIssue

	// Detailed results for individual directories
	DirectoryResults []DirectoryIntegrityResult

	// Summary of all issues found
	IssueSummary map[IntegrityIssueType]int
}

// IntegrityIssue represents a specific integrity problem
type IntegrityIssue struct {
	// Type of integrity issue
	Type IntegrityIssueType

	// Severity level of the issue
	Severity IntegrityIssueSeverity

	// Description of the issue
	Description string

	// The inode ID affected (if applicable)
	AffectedInode uint64

	// The path affected (if known)
	AffectedPath string

	// Additional details about the issue
	Details map[string]any
}

// IntegrityIssueType represents the type of integrity issue
type IntegrityIssueType int

const (
	IntegrityIssueCorruptedInode IntegrityIssueType = iota
	IntegrityIssueInvalidChecksum
	IntegrityIssueMissingExtent
	IntegrityIssueInvalidReference
	IntegrityIssueOrphanedFile
	IntegrityIssueInvalidPermissions
	IntegrityIssueCorruptedExtendedAttribute
	IntegrityIssueInconsistentTimestamp
)

// IntegrityIssueSeverity represents the severity of an integrity issue
type IntegrityIssueSeverity int

const (
	IntegrityIssueSeverityInfo IntegrityIssueSeverity = iota
	IntegrityIssueSeverityWarning
	IntegrityIssueSeverityError
	IntegrityIssueSeverityCritical
)
