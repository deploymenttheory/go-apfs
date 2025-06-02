package discover

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deploymenttheory/go-apfs/pkg/app"
)

func TestRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request Request
		wantErr bool
		errCode string
	}{
		{
			name: "valid basic request",
			request: Request{
				ContainerPath: "/dev/disk2",
				MaxResults:    1000,
			},
			wantErr: false,
		},
		{
			name: "missing container path",
			request: Request{
				MaxResults: 1000,
			},
			wantErr: true,
			errCode: app.ErrCodeInvalidInput,
		},
		{
			name: "invalid volume target - both ID and name",
			request: Request{
				ContainerPath: "/dev/disk2",
				Target: app.VolumeTarget{
					VolumeID:   1,
					VolumeName: "Test",
				},
				MaxResults: 1000,
			},
			wantErr: true,
			errCode: app.ErrCodeInvalidInput,
		},
		{
			name: "invalid regex pattern",
			request: Request{
				ContainerPath: "/dev/disk2",
				NameRegex:     "[invalid",
				MaxResults:    1000,
			},
			wantErr: true,
			errCode: app.ErrCodeInvalidInput,
		},
		{
			name: "invalid min size format",
			request: Request{
				ContainerPath: "/dev/disk2",
				MinSize:       "invalid",
				MaxResults:    1000,
			},
			wantErr: true,
			errCode: app.ErrCodeInvalidInput,
		},
		{
			name: "invalid max size format",
			request: Request{
				ContainerPath: "/dev/disk2",
				MaxSize:       "10XB",
				MaxResults:    1000,
			},
			wantErr: true,
			errCode: app.ErrCodeInvalidInput,
		},
		{
			name: "invalid date format - after",
			request: Request{
				ContainerPath: "/dev/disk2",
				ModifiedAfter: "2024-13-01",
				MaxResults:    1000,
			},
			wantErr: true,
			errCode: app.ErrCodeInvalidInput,
		},
		{
			name: "invalid date format - before",
			request: Request{
				ContainerPath:  "/dev/disk2",
				ModifiedBefore: "not-a-date",
				MaxResults:     1000,
			},
			wantErr: true,
			errCode: app.ErrCodeInvalidInput,
		},
		{
			name: "max results too small",
			request: Request{
				ContainerPath: "/dev/disk2",
				MaxResults:    0,
			},
			wantErr: true,
			errCode: app.ErrCodeInvalidInput,
		},
		{
			name: "max results too large",
			request: Request{
				ContainerPath: "/dev/disk2",
				MaxResults:    20000,
			},
			wantErr: true,
			errCode: app.ErrCodeInvalidInput,
		},
		{
			name: "conflicting name criteria",
			request: Request{
				ContainerPath: "/dev/disk2",
				NamePattern:   "*.pdf",
				NameRegex:     ".*\\.pdf$",
				MaxResults:    1000,
			},
			wantErr: true,
			errCode: app.ErrCodeInvalidInput,
		},
		{
			name: "valid complete request",
			request: Request{
				ContainerPath: "/dev/disk2",
				Target: app.VolumeTarget{
					VolumeName: "Macintosh HD",
					Snapshot:   "backup-1",
				},
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
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					var appErr *app.CommonError
					require.ErrorAs(t, err, &appErr)
					assert.Equal(t, tt.errCode, appErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSizeFormat(t *testing.T) {
	tests := []struct {
		name    string
		size    string
		wantErr bool
	}{
		{"valid bytes", "123B", false},
		{"valid kilobytes", "10KB", false},
		{"valid megabytes", "5MB", false},
		{"valid gigabytes", "2GB", false},
		{"valid terabytes", "1TB", false},
		{"valid decimal", "1.5MB", false},
		{"lowercase unit", "10mb", false},
		{"with spaces", " 10 MB ", false},

		{"empty string", "", true},
		{"no number", "MB", true},
		{"no unit", "123", true},
		{"invalid unit", "10XB", true},
		{"invalid number", "abc MB", true},
		{"multiple units", "10MBGB", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSizeFormat(tt.size)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		name     string
		size     string
		expected int64
		wantErr  bool
	}{
		{"bytes", "123B", 123, false},
		{"kilobytes", "10KB", 10 * 1024, false},
		{"megabytes", "5MB", 5 * 1024 * 1024, false},
		{"gigabytes", "2GB", 2 * 1024 * 1024 * 1024, false},
		{"decimal megabytes", "1.5MB", int64(1.5 * 1024 * 1024), false},
		{"lowercase", "10mb", 10 * 1024 * 1024, false},
		{"with spaces", " 10 MB ", 10 * 1024 * 1024, false},

		{"invalid format", "invalid", 0, true},
		{"empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSize(tt.size)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
