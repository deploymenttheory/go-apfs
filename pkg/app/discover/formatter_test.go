package discover

import (
	"testing"
)

func TestFormatOutput(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		wantErr  bool
		validate func(*testing.T, string)
	}{
		{
			name:    "table format",
			format:  "table",
			wantErr: false,
			validate: func(t *testing.T, output string) {
				// Implementation of validate function
			},
		},
		{
			name:    "json format",
			format:  "json",
			wantErr: false,
			validate: func(t *testing.T, output string) {
				// Implementation of validate function
			},
		},
		{
			name:    "yaml format",
			format:  "yaml",
			wantErr: false,
			validate: func(t *testing.T, output string) {
				// Implementation of validate function
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Implementation of test logic
		})
	}
}
