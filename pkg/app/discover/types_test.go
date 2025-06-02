package discover

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFileResult_GetSizeClass(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected SizeClass
	}{
		{"tiny file", 512, SizeClassTiny},
		{"small file", 512 * 1024, SizeClassSmall},
		{"medium file", 50 * 1024 * 1024, SizeClassMedium},
		{"large file", 500 * 1024 * 1024, SizeClassLarge},
		{"huge file", 2 * 1024 * 1024 * 1024, SizeClassHuge},
		{"edge case - 1KB", 1024, SizeClassSmall},
		{"edge case - 1MB", 1024 * 1024, SizeClassSmall},
		{"edge case - 100MB", 100 * 1024 * 1024, SizeClassLarge},
		{"edge case - 1GB", 1024 * 1024 * 1024, SizeClassHuge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := FileResult{Size: tt.size}
			assert.Equal(t, tt.expected, file.GetSizeClass())
		})
	}
}

func TestFileResult_FormatSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{"bytes", 512, "512 B"},
		{"kilobytes", 1536, "1.5 KB"},
		{"megabytes", int64(2.5 * 1024 * 1024), "2.5 MB"},
		{"gigabytes", (32 * 1024 * 1024 * 1024) / 10, "3.2 GB"},
		{"zero bytes", 0, "0 B"},
		{"one byte", 1, "1 B"},
		{"exact kilobyte", 1024, "1.0 KB"},
		{"exact megabyte", 1024 * 1024, "1.0 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := FileResult{Size: tt.size}
			result := file.FormatSize()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSearchQuery_JSONSerialization(t *testing.T) {
	query := SearchQuery{
		NamePattern:    "*.pdf",
		Extensions:     []string{"pdf", "doc"},
		CaseSensitive:  true,
		MinSize:        "1MB",
		MaxSize:        "100MB",
		ModifiedAfter:  "2024-01-01",
		ModifiedBefore: "2024-12-31",
		ContentSearch:  "secret",
		IncludeDeleted: true,
		MaxResults:     500,
	}

	// Test that all fields are properly tagged for JSON
	assert.Equal(t, "*.pdf", query.NamePattern)
	assert.Equal(t, []string{"pdf", "doc"}, query.Extensions)
	assert.True(t, query.CaseSensitive)
	assert.Equal(t, "1MB", query.MinSize)
	assert.Equal(t, "100MB", query.MaxSize)
	assert.Equal(t, "2024-01-01", query.ModifiedAfter)
	assert.Equal(t, "2024-12-31", query.ModifiedBefore)
	assert.Equal(t, "secret", query.ContentSearch)
	assert.True(t, query.IncludeDeleted)
	assert.Equal(t, 500, query.MaxResults)
}

func TestFileResult_CompleteStructure(t *testing.T) {
	now := time.Now()
	created := now.Add(-24 * time.Hour)

	file := FileResult{
		Path:        "/Users/test/document.pdf",
		Name:        "document.pdf",
		Size:        1024 * 1024, // 1MB
		Modified:    now,
		Created:     created,
		Type:        "file",
		Deleted:     false,
		VolumeID:    1,
		InodeID:     12345,
		Permissions: "-rw-r--r--",
		Owner:       "test",
		Group:       "staff",
		Extension:   "pdf",
		Compressed:  false,
		Encrypted:   false,
	}

	assert.Equal(t, "/Users/test/document.pdf", file.Path)
	assert.Equal(t, "document.pdf", file.Name)
	assert.Equal(t, int64(1024*1024), file.Size)
	assert.Equal(t, now, file.Modified)
	assert.Equal(t, created, file.Created)
	assert.Equal(t, "file", file.Type)
	assert.False(t, file.Deleted)
	assert.Equal(t, uint64(1), file.VolumeID)
	assert.Equal(t, uint64(12345), file.InodeID)
	assert.Equal(t, "-rw-r--r--", file.Permissions)
	assert.Equal(t, "test", file.Owner)
	assert.Equal(t, "staff", file.Group)
	assert.Equal(t, "pdf", file.Extension)
	assert.False(t, file.Compressed)
	assert.False(t, file.Encrypted)

	// Test computed properties
	assert.Equal(t, SizeClassSmall, file.GetSizeClass())
	assert.Equal(t, "1.0 MB", file.FormatSize())
}

func TestVolumeInfo_Structure(t *testing.T) {
	volume := VolumeInfo{
		ID:            1,
		Name:          "Macintosh HD",
		UUID:          "12345678-1234-1234-1234-123456789ABC",
		Role:          "System",
		Encrypted:     true,
		CaseSensitive: true,
	}

	assert.Equal(t, uint64(1), volume.ID)
	assert.Equal(t, "Macintosh HD", volume.Name)
	assert.Equal(t, "12345678-1234-1234-1234-123456789ABC", volume.UUID)
	assert.Equal(t, "System", volume.Role)
	assert.True(t, volume.Encrypted)
	assert.True(t, volume.CaseSensitive)
}

func TestResponse_Structure(t *testing.T) {
	response := Response{
		Files: []FileResult{
			{Name: "file1.pdf", Size: 1024},
			{Name: "file2.doc", Size: 2048},
		},
		TotalFound: 2,
		SearchTime: 250 * time.Millisecond,
		VolumeInfo: VolumeInfo{
			ID:   1,
			Name: "Test Volume",
		},
		Truncated: false,
		SearchQuery: SearchQuery{
			Extensions: []string{"pdf", "doc"},
			MaxResults: 1000,
		},
	}

	assert.Len(t, response.Files, 2)
	assert.Equal(t, 2, response.TotalFound)
	assert.Equal(t, 250*time.Millisecond, response.SearchTime)
	assert.Equal(t, "Test Volume", response.VolumeInfo.Name)
	assert.False(t, response.Truncated)
	assert.Equal(t, []string{"pdf", "doc"}, response.SearchQuery.Extensions)
}
