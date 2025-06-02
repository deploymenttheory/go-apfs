package discover

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/deploymenttheory/go-apfs/pkg/app"
)

// Handle processes a discovery request
func Handle(ctx *app.Context, req *Request) (*Response, error) {
	startTime := time.Now()

	// 1. Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	ctx.Log(fmt.Sprintf("Starting file discovery in: %s", req.ContainerPath))
	ctx.Progress("Validating container...", 5)

	// 2. Log search criteria
	logSearchCriteria(ctx, req)

	// 3. TODO: Call orchestration layer when implemented
	// For now, return mock data
	ctx.Progress("Scanning filesystem...", 25)

	response := createMockResponse(req)
	response.SearchTime = time.Since(startTime)
	response.SearchQuery = createSearchQuery(req)

	ctx.Progress("Processing results...", 90)

	// Truncate results if over limit
	if len(response.Files) > req.MaxResults {
		response.Files = response.Files[:req.MaxResults]
		response.Truncated = true
	}

	ctx.Progress("Complete", 100)
	ctx.Log(fmt.Sprintf("Discovery completed: found %d files in %v", response.TotalFound, response.SearchTime))

	return response, nil
}

// logSearchCriteria logs the search criteria for verbose output
func logSearchCriteria(ctx *app.Context, req *Request) {
	if !ctx.Verbose {
		return
	}

	ctx.Log("Search criteria:")
	if !req.Target.IsEmpty() {
		ctx.Log("  " + req.Target.String())
	}
	if req.NamePattern != "" {
		ctx.Log(fmt.Sprintf("  Name pattern: %s", req.NamePattern))
	}
	if req.NameRegex != "" {
		ctx.Log(fmt.Sprintf("  Name regex: %s", req.NameRegex))
	}
	if len(req.Extensions) > 0 {
		ctx.Log(fmt.Sprintf("  Extensions: %s", strings.Join(req.Extensions, ", ")))
	}
	if req.ContentSearch != "" {
		ctx.Log(fmt.Sprintf("  Content search: \"%s\"", req.ContentSearch))
	}
	if req.MinSize != "" || req.MaxSize != "" {
		ctx.Log(fmt.Sprintf("  Size range: %s - %s", req.MinSize, req.MaxSize))
	}
	if req.IncludeDeleted {
		ctx.Log("  Including deleted files")
	}
}

// createMockResponse creates mock data for testing
func createMockResponse(req *Request) *Response {
	// Mock file results based on search criteria
	files := []FileResult{}

	// Generate some mock files that match the criteria
	mockFiles := []struct {
		path string
		size int64
		ext  string
	}{
		{"/Users/alice/Documents/report.pdf", 2048576, "pdf"},
		{"/Users/alice/Pictures/vacation.jpg", 5242880, "jpg"},
		{"/Users/bob/Downloads/password.txt", 1024, "txt"},
		{"/System/Library/secret_config.plist", 4096, "plist"},
		{"/Users/alice/Desktop/presentation.pptx", 15728640, "pptx"},
	}

	for i, mock := range mockFiles {
		// Simple filtering based on extensions
		if len(req.Extensions) > 0 {
			matches := false
			for _, ext := range req.Extensions {
				if strings.EqualFold(ext, mock.ext) {
					matches = true
					break
				}
			}
			if !matches {
				continue
			}
		}

		// Simple name pattern matching
		if req.NamePattern != "" {
			matched, _ := filepath.Match(strings.ToLower(req.NamePattern), strings.ToLower(filepath.Base(mock.path)))
			if !matched {
				continue
			}
		}

		file := FileResult{
			Path:        mock.path,
			Name:        filepath.Base(mock.path),
			Size:        mock.size,
			Modified:    time.Now().Add(-time.Duration(i*24) * time.Hour),
			Created:     time.Now().Add(-time.Duration(i*30*24) * time.Hour),
			Type:        "file",
			VolumeID:    1,
			InodeID:     uint64(100000 + i),
			Permissions: "-rw-r--r--",
			Owner:       "alice",
			Group:       "staff",
			Extension:   mock.ext,
		}

		files = append(files, file)
	}

	return &Response{
		Files:      files,
		TotalFound: len(files),
		VolumeInfo: VolumeInfo{
			ID:            1,
			Name:          "Macintosh HD",
			UUID:          "12345678-1234-1234-1234-123456789ABC",
			Role:          "System",
			Encrypted:     true,
			CaseSensitive: true,
		},
	}
}

// createSearchQuery creates a SearchQuery from the request
func createSearchQuery(req *Request) SearchQuery {
	return SearchQuery{
		NamePattern:    req.NamePattern,
		NameRegex:      req.NameRegex,
		Extensions:     req.Extensions,
		CaseSensitive:  req.CaseSensitive,
		MinSize:        req.MinSize,
		MaxSize:        req.MaxSize,
		ModifiedAfter:  req.ModifiedAfter,
		ModifiedBefore: req.ModifiedBefore,
		ContentSearch:  req.ContentSearch,
		IncludeDeleted: req.IncludeDeleted,
		MaxResults:     req.MaxResults,
	}
}
