package discover

import (
	"fmt"
	"time"

	"github.com/deploymenttheory/go-apfs/pkg/app"
)

// Request represents a file discovery request
type Request struct {
	ContainerPath string
	Target        app.VolumeTarget

	// Search criteria
	NamePattern    string
	NameRegex      string
	Extensions     []string
	CaseSensitive  bool
	MinSize        string
	MaxSize        string
	ModifiedAfter  string
	ModifiedBefore string
	ContentSearch  string
	IncludeDeleted bool
	MaxResults     int
}

// Response represents discovery results
type Response struct {
	Files       []FileResult  `json:"files"`
	TotalFound  int           `json:"total_found"`
	SearchTime  time.Duration `json:"search_time"`
	VolumeInfo  VolumeInfo    `json:"volume_info"`
	Truncated   bool          `json:"truncated"`
	SearchQuery SearchQuery   `json:"search_query"`
}

// FileResult represents a discovered file
type FileResult struct {
	Path        string    `json:"path"`
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	Modified    time.Time `json:"modified"`
	Created     time.Time `json:"created"`
	Type        string    `json:"type"`
	Deleted     bool      `json:"deleted"`
	VolumeID    uint64    `json:"volume_id"`
	InodeID     uint64    `json:"inode_id"`
	Permissions string    `json:"permissions"`
	Owner       string    `json:"owner"`
	Group       string    `json:"group"`
	Extension   string    `json:"extension"`
	Compressed  bool      `json:"compressed"`
	Encrypted   bool      `json:"encrypted"`
}

// VolumeInfo represents information about the searched volume
type VolumeInfo struct {
	ID            uint64 `json:"id"`
	Name          string `json:"name"`
	UUID          string `json:"uuid"`
	Role          string `json:"role"`
	Encrypted     bool   `json:"encrypted"`
	CaseSensitive bool   `json:"case_sensitive"`
}

// SearchQuery represents the executed search parameters
type SearchQuery struct {
	NamePattern    string   `json:"name_pattern,omitempty"`
	NameRegex      string   `json:"name_regex,omitempty"`
	Extensions     []string `json:"extensions,omitempty"`
	CaseSensitive  bool     `json:"case_sensitive"`
	MinSize        string   `json:"min_size,omitempty"`
	MaxSize        string   `json:"max_size,omitempty"`
	ModifiedAfter  string   `json:"modified_after,omitempty"`
	ModifiedBefore string   `json:"modified_before,omitempty"`
	ContentSearch  string   `json:"content_search,omitempty"`
	IncludeDeleted bool     `json:"include_deleted"`
	MaxResults     int      `json:"max_results"`
}

// SizeClass represents file size categories for display
type SizeClass string

const (
	SizeClassTiny   SizeClass = "tiny"   // < 1KB
	SizeClassSmall  SizeClass = "small"  // < 1MB
	SizeClassMedium SizeClass = "medium" // < 100MB
	SizeClassLarge  SizeClass = "large"  // < 1GB
	SizeClassHuge   SizeClass = "huge"   // >= 1GB
)

// GetSizeClass returns the size class for display purposes
func (f *FileResult) GetSizeClass() SizeClass {
	switch {
	case f.Size < 1024:
		return SizeClassTiny
	case f.Size < 1024*1024:
		return SizeClassSmall
	case f.Size < 100*1024*1024:
		return SizeClassMedium
	case f.Size < 1024*1024*1024:
		return SizeClassLarge
	default:
		return SizeClassHuge
	}
}

// FormatSize returns a human-readable size string
func (f *FileResult) FormatSize() string {
	const unit = 1024
	if f.Size < unit {
		return fmt.Sprintf("%d B", f.Size)
	}
	div, exp := int64(unit), 0
	for n := f.Size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(f.Size)/float64(div), "KMGTPE"[exp])
}
