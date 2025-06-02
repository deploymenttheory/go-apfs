package services

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"
)

// filesystemService implements the FilesystemService interface
type filesystemService struct {
	containerService ContainerService
}

// NewFilesystemService creates a new filesystem service instance
func NewFilesystemService(containerService ContainerService) FilesystemService {
	return &filesystemService{
		containerService: containerService,
	}
}

// ListDirectory lists files and directories at the specified path
func (fs *filesystemService) ListDirectory(ctx context.Context, containerPath string, volumeID uint64, dirPath string, recursive bool) ([]FileInfo, error) {
	// This is a simplified implementation that would need full B-tree traversal
	// For now, return a mock result to show the interface structure

	// Ensure container is open
	_, err := fs.containerService.OpenContainer(ctx, containerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open container: %w", err)
	}

	// Mock directory listing - real implementation would:
	// 1. Read volume superblock to get filesystem tree root
	// 2. Traverse B-tree to find directory inode
	// 3. Read directory entries
	// 4. Recursively traverse if recursive=true

	var files []FileInfo

	if dirPath == "/" || dirPath == "" {
		// Mock root directory entries
		files = append(files, FileInfo{
			InodeID:  2, // Root directory typically has inode 2
			Name:     "Applications",
			Path:     "/Applications",
			Type:     "directory",
			Size:     0,
			Mode:     0755,
			Created:  time.Now().Add(-24 * time.Hour),
			Modified: time.Now().Add(-1 * time.Hour),
			Accessed: time.Now(),
			Changed:  time.Now().Add(-1 * time.Hour),
		})

		files = append(files, FileInfo{
			InodeID:  3,
			Name:     "Users",
			Path:     "/Users",
			Type:     "directory",
			Size:     0,
			Mode:     0755,
			Created:  time.Now().Add(-30 * 24 * time.Hour),
			Modified: time.Now().Add(-2 * time.Hour),
			Accessed: time.Now(),
			Changed:  time.Now().Add(-2 * time.Hour),
		})

		files = append(files, FileInfo{
			InodeID:  4,
			Name:     "System",
			Path:     "/System",
			Type:     "directory",
			Size:     0,
			Mode:     0755,
			Created:  time.Now().Add(-30 * 24 * time.Hour),
			Modified: time.Now().Add(-6 * time.Hour),
			Accessed: time.Now(),
			Changed:  time.Now().Add(-6 * time.Hour),
		})
	}

	return files, nil
}

// GetFileInfo retrieves detailed information about a specific file
func (fs *filesystemService) GetFileInfo(ctx context.Context, containerPath string, volumeID uint64, filePath string) (FileInfo, error) {
	// Ensure container is open
	_, err := fs.containerService.OpenContainer(ctx, containerPath)
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to open container: %w", err)
	}

	// Mock file info - real implementation would:
	// 1. Parse file path to traverse directory structure
	// 2. Look up file in B-tree by name
	// 3. Read inode to get detailed metadata
	// 4. Read extended attributes if present

	fileName := path.Base(filePath)
	dirPath := path.Dir(filePath)

	fileInfo := FileInfo{
		InodeID:  100, // Mock inode ID
		Name:     fileName,
		Path:     filePath,
		Type:     fs.determineFileType(fileName),
		Size:     1024, // Mock size
		Mode:     0644,
		Created:  time.Now().Add(-7 * 24 * time.Hour),
		Modified: time.Now().Add(-1 * 24 * time.Hour),
		Accessed: time.Now(),
		Changed:  time.Now().Add(-1 * 24 * time.Hour),
	}

	// Check if this appears to be a directory
	if strings.HasSuffix(filePath, "/") || fs.isKnownDirectory(dirPath, fileName) {
		fileInfo.Type = "directory"
		fileInfo.Mode = 0755
		fileInfo.Size = 0
	}

	return fileInfo, nil
}

// GetDirectoryInfo retrieves directory information with statistics
func (fs *filesystemService) GetDirectoryInfo(ctx context.Context, containerPath string, volumeID uint64, dirPath string, includeChildren bool) (DirectoryInfo, error) {
	// Get basic directory file info
	fileInfo, err := fs.GetFileInfo(ctx, containerPath, volumeID, dirPath)
	if err != nil {
		return DirectoryInfo{}, err
	}

	dirInfo := DirectoryInfo{
		FileInfo:   fileInfo,
		ChildCount: 0,
		TotalSize:  0,
		Recursive:  includeChildren,
	}

	if includeChildren {
		children, err := fs.ListDirectory(ctx, containerPath, volumeID, dirPath, false)
		if err != nil {
			return dirInfo, fmt.Errorf("failed to list children: %w", err)
		}

		dirInfo.Children = children
		dirInfo.ChildCount = uint64(len(children))

		// Calculate total size
		for _, child := range children {
			dirInfo.TotalSize += child.Size
		}
	}

	return dirInfo, nil
}

// FindFiles searches for files matching specified criteria
func (fs *filesystemService) FindFiles(ctx context.Context, containerPath string, volumeID uint64, searchPath string, pattern string, maxResults int) ([]FileInfo, error) {
	// Ensure container is open
	_, err := fs.containerService.OpenContainer(ctx, containerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open container: %w", err)
	}

	// Mock search results - real implementation would:
	// 1. Traverse filesystem tree starting from searchPath
	// 2. Match filenames against pattern (glob or regex)
	// 3. Collect matching results up to maxResults

	var results []FileInfo

	// Simple mock: if pattern matches known files, return them
	if strings.Contains(pattern, "app") || strings.Contains(pattern, "App") {
		results = append(results, FileInfo{
			InodeID:  200,
			Name:     "Calculator.app",
			Path:     "/Applications/Calculator.app",
			Type:     "directory",
			Size:     0,
			Mode:     0755,
			Created:  time.Now().Add(-30 * 24 * time.Hour),
			Modified: time.Now().Add(-5 * 24 * time.Hour),
			Accessed: time.Now(),
			Changed:  time.Now().Add(-5 * 24 * time.Hour),
		})
	}

	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results, nil
}

// GetInode retrieves file information by inode ID
func (fs *filesystemService) GetInode(ctx context.Context, containerPath string, volumeID uint64, inodeID uint64) (FileInfo, error) {
	// Ensure container is open
	_, err := fs.containerService.OpenContainer(ctx, containerPath)
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to open container: %w", err)
	}

	// Mock inode lookup - real implementation would:
	// 1. Use inode ID to directly read inode record from B-tree
	// 2. Parse inode data to extract metadata
	// 3. Handle extended attributes and special file types

	fileInfo := FileInfo{
		InodeID:  inodeID,
		Name:     fmt.Sprintf("inode_%d", inodeID),
		Path:     fmt.Sprintf("/unknown/inode_%d", inodeID),
		Type:     "file",
		Size:     512,
		Mode:     0644,
		Created:  time.Now().Add(-10 * 24 * time.Hour),
		Modified: time.Now().Add(-2 * 24 * time.Hour),
		Accessed: time.Now(),
		Changed:  time.Now().Add(-2 * 24 * time.Hour),
	}

	return fileInfo, nil
}

// WalkFilesystem performs a depth-first traversal of the filesystem
func (fs *filesystemService) WalkFilesystem(ctx context.Context, containerPath string, volumeID uint64, rootPath string, walkFunc func(FileInfo) error) error {
	// Ensure container is open
	_, err := fs.containerService.OpenContainer(ctx, containerPath)
	if err != nil {
		return fmt.Errorf("failed to open container: %w", err)
	}

	// Start with root directory
	files, err := fs.ListDirectory(ctx, containerPath, volumeID, rootPath, false)
	if err != nil {
		return fmt.Errorf("failed to list root directory: %w", err)
	}

	// Walk each file/directory
	for _, file := range files {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Call walk function
		if err := walkFunc(file); err != nil {
			return err
		}

		// If this is a directory, recursively walk it
		if file.Type == "directory" {
			err := fs.WalkFilesystem(ctx, containerPath, volumeID, file.Path, walkFunc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// CheckAccess determines if a file/directory is accessible (not encrypted)
func (fs *filesystemService) CheckAccess(ctx context.Context, containerPath string, volumeID uint64, filePath string) (bool, error) {
	// Ensure container is open
	_, err := fs.containerService.OpenContainer(ctx, containerPath)
	if err != nil {
		return false, fmt.Errorf("failed to open container: %w", err)
	}

	// Mock accessibility check - real implementation would:
	// 1. Check if volume is encrypted
	// 2. Check file protection class
	// 3. Verify if file data is accessible with current keys

	// For now, assume system directories are accessible, user data may not be
	if strings.HasPrefix(filePath, "/System") ||
		strings.HasPrefix(filePath, "/Applications") ||
		strings.HasPrefix(filePath, "/Library") {
		return true, nil
	}

	// User directories might be encrypted
	if strings.HasPrefix(filePath, "/Users") {
		return false, nil // Assume encrypted
	}

	return true, nil // Default to accessible
}

// Helper functions

// determineFileType determines file type based on name/extension
func (fs *filesystemService) determineFileType(fileName string) string {
	if strings.HasSuffix(fileName, ".app") {
		return "directory" // .app bundles are directories
	}
	if strings.HasSuffix(fileName, ".txt") ||
		strings.HasSuffix(fileName, ".log") ||
		strings.HasSuffix(fileName, ".conf") {
		return "file"
	}
	if strings.HasSuffix(fileName, ".dylib") ||
		strings.HasSuffix(fileName, ".so") {
		return "file"
	}
	if strings.Contains(fileName, ".") {
		return "file"
	}
	return "file" // Default to file
}

// isKnownDirectory checks if a filename represents a known directory
func (fs *filesystemService) isKnownDirectory(dirPath, fileName string) bool {
	knownDirs := []string{
		"Applications", "Library", "System", "Users", "bin", "usr",
		"sbin", "var", "tmp", "etc", "private", "opt",
	}

	for _, known := range knownDirs {
		if fileName == known {
			return true
		}
	}

	return false
}
