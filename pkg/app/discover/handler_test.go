package discover

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deploymenttheory/go-apfs/pkg/app"
)

func TestHandle(t *testing.T) {
	tests := []struct {
		name     string
		request  *Request
		wantErr  bool
		validate func(*testing.T, *Response)
	}{
		{
			name: "basic request",
			request: &Request{
				ContainerPath: "/test/container.dmg",
				MaxResults:    1000,
			},
			wantErr: false,
			validate: func(t *testing.T, resp *Response) {
				assert.NotNil(t, resp)
				assert.GreaterOrEqual(t, resp.TotalFound, 0)
				assert.Greater(t, resp.SearchTime, time.Duration(0))
				assert.Equal(t, "Macintosh HD", resp.VolumeInfo.Name)
			},
		},
		{
			name: "request with PDF filter",
			request: &Request{
				ContainerPath: "/test/container.dmg",
				Extensions:    []string{"pdf"},
				MaxResults:    1000,
			},
			wantErr: false,
			validate: func(t *testing.T, resp *Response) {
				assert.NotNil(t, resp)
				// All returned files should be PDFs
				for _, file := range resp.Files {
					assert.Equal(t, "pdf", file.Extension)
					assert.True(t, strings.HasSuffix(file.Name, ".pdf"))
				}
			},
		},
		{
			name: "request with name pattern",
			request: &Request{
				ContainerPath: "/test/container.dmg",
				NamePattern:   "*password*",
				MaxResults:    1000,
			},
			wantErr: false,
			validate: func(t *testing.T, resp *Response) {
				assert.NotNil(t, resp)
				// All returned files should match pattern
				for _, file := range resp.Files {
					assert.True(t, strings.Contains(strings.ToLower(file.Name), "password"))
				}
			},
		},
		{
			name: "request with volume target",
			request: &Request{
				ContainerPath: "/test/container.dmg",
				Target: app.VolumeTarget{
					VolumeName: "Data",
					Snapshot:   "backup-1",
				},
				MaxResults: 1000,
			},
			wantErr: false,
			validate: func(t *testing.T, resp *Response) {
				assert.NotNil(t, resp)
				assert.Equal(t, "backup-1", resp.SearchQuery.IncludeDeleted) // This would be properly set in real implementation
			},
		},
		{
			name: "invalid request - missing container path",
			request: &Request{
				MaxResults: 1000,
			},
			wantErr: true,
		},
		{
			name: "invalid request - bad regex",
			request: &Request{
				ContainerPath: "/test/container.dmg",
				NameRegex:     "[invalid",
				MaxResults:    1000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := app.NewContext()
			ctx.Verbose = false
			ctx.Quiet = true

			resp, err := Handle(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				if tt.validate != nil {
					tt.validate(t, resp)
				}
			}
		})
	}
}

func TestCreateMockResponse(t *testing.T) {
	tests := []struct {
		name     string
		request  *Request
		validate func(*testing.T, *Response)
	}{
		{
			name: "no filters",
			request: &Request{
				ContainerPath: "/test/container.dmg",
				MaxResults:    1000,
			},
			validate: func(t *testing.T, resp *Response) {
				assert.NotEmpty(t, resp.Files)
				assert.Equal(t, len(resp.Files), resp.TotalFound)
			},
		},
		{
			name: "PDF extension filter",
			request: &Request{
				ContainerPath: "/test/container.dmg",
				Extensions:    []string{"pdf"},
				MaxResults:    1000,
			},
			validate: func(t *testing.T, resp *Response) {
				for _, file := range resp.Files {
					assert.Equal(t, "pdf", file.Extension)
				}
			},
		},
		{
			name: "multiple extension filter",
			request: &Request{
				ContainerPath: "/test/container.dmg",
				Extensions:    []string{"pdf", "jpg"},
				MaxResults:    1000,
			},
			validate: func(t *testing.T, resp *Response) {
				for _, file := range resp.Files {
					assert.Contains(t, []string{"pdf", "jpg"}, file.Extension)
				}
			},
		},
		{
			name: "name pattern filter",
			request: &Request{
				ContainerPath: "/test/container.dmg",
				NamePattern:   "*password*",
				MaxResults:    1000,
			},
			validate: func(t *testing.T, resp *Response) {
				for _, file := range resp.Files {
					assert.True(t, strings.Contains(strings.ToLower(file.Name), "password"))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := createMockResponse(tt.request)
			require.NotNil(t, resp)

			// Common validations
			assert.NotEmpty(t, resp.VolumeInfo.Name)
			assert.Greater(t, resp.VolumeInfo.ID, uint64(0))
			assert.NotEmpty(t, resp.VolumeInfo.UUID)

			if tt.validate != nil {
				tt.validate(t, resp)
			}
		})
	}
}

func TestCreateSearchQuery(t *testing.T) {
	request := &Request{
		ContainerPath:  "/test/container.dmg",
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

	query := createSearchQuery(request)

	assert.Equal(t, request.NamePattern, query.NamePattern)
	assert.Equal(t, request.Extensions, query.Extensions)
	assert.Equal(t, request.CaseSensitive, query.CaseSensitive)
	assert.Equal(t, request.MinSize, query.MinSize)
	assert.Equal(t, request.MaxSize, query.MaxSize)
	assert.Equal(t, request.ModifiedAfter, query.ModifiedAfter)
	assert.Equal(t, request.ModifiedBefore, query.ModifiedBefore)
	assert.Equal(t, request.ContentSearch, query.ContentSearch)
	assert.Equal(t, request.IncludeDeleted, query.IncludeDeleted)
	assert.Equal(t, request.MaxResults, query.MaxResults)
}

func TestLogSearchCriteria(t *testing.T) {
	// Test that verbose logging works without panicking
	ctx := app.NewContext()
	ctx.Verbose = true

	request := &Request{
		ContainerPath: "/test/container.dmg",
		Target: app.VolumeTarget{
			VolumeName: "Test Volume",
		},
		NamePattern:    "*.pdf",
		Extensions:     []string{"pdf"},
		ContentSearch:  "secret",
		MinSize:        "1MB",
		MaxSize:        "100MB",
		IncludeDeleted: true,
	}

	// This should not panic
	logSearchCriteria(ctx, request)

	// Test with non-verbose mode
	ctx.Verbose = false
	logSearchCriteria(ctx, request)
}
