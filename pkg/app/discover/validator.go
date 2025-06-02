package discover

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/deploymenttheory/go-apfs/pkg/app"
)

// Validate validates a discovery request
func (r *Request) Validate() error {
	// Container path is required
	if r.ContainerPath == "" {
		return app.NewError(app.ErrCodeInvalidInput, "container path is required", nil)
	}

	// Validate volume target
	if err := r.Target.Validate(); err != nil {
		return app.NewError(app.ErrCodeInvalidInput, "invalid volume target", err)
	}

	// Validate regex pattern if provided
	if r.NameRegex != "" {
		if _, err := regexp.Compile(r.NameRegex); err != nil {
			return app.NewError(app.ErrCodeInvalidInput, "invalid regex pattern", err)
		}
	}

	// Validate size formats
	if r.MinSize != "" {
		if err := validateSizeFormat(r.MinSize); err != nil {
			return app.NewError(app.ErrCodeInvalidInput, "invalid min-size format", err)
		}
	}
	if r.MaxSize != "" {
		if err := validateSizeFormat(r.MaxSize); err != nil {
			return app.NewError(app.ErrCodeInvalidInput, "invalid max-size format", err)
		}
	}

	// Validate date formats
	if r.ModifiedAfter != "" {
		if _, err := time.Parse("2006-01-02", r.ModifiedAfter); err != nil {
			return app.NewError(app.ErrCodeInvalidInput, "invalid date format for modified-after, use YYYY-MM-DD", err)
		}
	}
	if r.ModifiedBefore != "" {
		if _, err := time.Parse("2006-01-02", r.ModifiedBefore); err != nil {
			return app.NewError(app.ErrCodeInvalidInput, "invalid date format for modified-before, use YYYY-MM-DD", err)
		}
	}

	// Validate max results
	if r.MaxResults < 1 || r.MaxResults > 10000 {
		return app.NewError(app.ErrCodeInvalidInput, "max results must be between 1 and 10000", nil)
	}

	// Check for conflicting search criteria
	if r.NamePattern != "" && r.NameRegex != "" {
		return app.NewError(app.ErrCodeInvalidInput, "cannot specify both name pattern and regex", nil)
	}

	return nil
}

// validateSizeFormat validates size format strings like "10MB", "1GB"
func validateSizeFormat(size string) error {
	size = strings.ToUpper(strings.TrimSpace(size))
	if size == "" {
		return fmt.Errorf("empty size")
	}

	// Remove all spaces from the string to handle cases like " 10 MB "
	size = strings.ReplaceAll(size, " ", "")

	// Extract numeric part and unit
	var numPart string
	var unit string

	for i, char := range size {
		if char >= '0' && char <= '9' || char == '.' {
			numPart += string(char)
		} else {
			unit = size[i:]
			break
		}
	}

	// Validate numeric part
	if numPart == "" {
		return fmt.Errorf("no numeric value found")
	}
	if _, err := strconv.ParseFloat(numPart, 64); err != nil {
		return fmt.Errorf("invalid numeric value: %s", numPart)
	}

	// Validate unit
	validUnits := []string{"B", "KB", "MB", "GB", "TB"}
	validUnit := false
	for _, validU := range validUnits {
		if unit == validU {
			validUnit = true
			break
		}
	}
	if !validUnit {
		return fmt.Errorf("invalid size unit: %s (valid: %s)", unit, strings.Join(validUnits, ", "))
	}

	return nil
}

// ParseSize converts size string to bytes
func ParseSize(size string) (int64, error) {
	if err := validateSizeFormat(size); err != nil {
		return 0, err
	}

	size = strings.ToUpper(strings.TrimSpace(size))
	// Remove all spaces to handle cases like " 10 MB "
	size = strings.ReplaceAll(size, " ", "")

	// Extract numeric part and unit
	var numPart string
	var unit string

	for i, char := range size {
		if char >= '0' && char <= '9' || char == '.' {
			numPart += string(char)
		} else {
			unit = size[i:]
			break
		}
	}

	value, _ := strconv.ParseFloat(numPart, 64)

	multiplier := map[string]int64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}

	return int64(value * float64(multiplier[unit])), nil
}
