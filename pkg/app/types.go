package app

import (
	"errors"
	"fmt"
	"time"
)

// VolumeTarget represents volume selection across commands
type VolumeTarget struct {
	VolumeID   uint64
	VolumeName string
	Snapshot   string
}

// Validate ensures volume target is valid
func (vt *VolumeTarget) Validate() error {
	if vt.VolumeID != 0 && vt.VolumeName != "" {
		return errors.New("cannot specify both volume-id and volume-name")
	}
	return nil
}

// IsEmpty returns true if no volume target is specified
func (vt *VolumeTarget) IsEmpty() bool {
	return vt.VolumeID == 0 && vt.VolumeName == "" && vt.Snapshot == ""
}

// String returns a string representation of the volume target
func (vt *VolumeTarget) String() string {
	if vt.VolumeName != "" {
		result := "Volume: " + vt.VolumeName
		if vt.Snapshot != "" {
			result += " (Snapshot: " + vt.Snapshot + ")"
		}
		return result
	}
	if vt.VolumeID != 0 {
		result := fmt.Sprintf("Volume ID: %d", vt.VolumeID)
		if vt.Snapshot != "" {
			result += " (Snapshot: " + vt.Snapshot + ")"
		}
		return result
	}
	if vt.Snapshot != "" {
		return "Snapshot: " + vt.Snapshot
	}
	return "All volumes"
}

// ProgressUpdate represents progress information
type ProgressUpdate struct {
	Message     string
	Completed   int64
	Total       int64
	StartedAt   time.Time
	ElapsedTime time.Duration
}

// Percent calculates completion percentage
func (p *ProgressUpdate) Percent() int {
	if p.Total == 0 {
		return 0
	}
	return int((p.Completed * 100) / p.Total)
}

// Rate calculates items per second
func (p *ProgressUpdate) Rate() float64 {
	if p.ElapsedTime == 0 {
		return 0
	}
	return float64(p.Completed) / p.ElapsedTime.Seconds()
}

// ETA estimates time to completion
func (p *ProgressUpdate) ETA() time.Duration {
	if p.Completed == 0 || p.Total == 0 {
		return 0
	}
	rate := p.Rate()
	if rate == 0 {
		return 0
	}
	remaining := p.Total - p.Completed
	return time.Duration(float64(remaining)/rate) * time.Second
}

// CommonError represents application-level errors
type CommonError struct {
	Code    string
	Message string
	Cause   error
}

func (e *CommonError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *CommonError) Unwrap() error {
	return e.Cause
}

// Common error codes
const (
	ErrCodeInvalidInput    = "INVALID_INPUT"
	ErrCodeContainerAccess = "CONTAINER_ACCESS"
	ErrCodeVolumeNotFound  = "VOLUME_NOT_FOUND"
	ErrCodePermission      = "PERMISSION_DENIED"
	ErrCodeTimeout         = "TIMEOUT"
	ErrCodeNotImplemented  = "NOT_IMPLEMENTED"
)

// NewError creates a new CommonError
func NewError(code, message string, cause error) *CommonError {
	return &CommonError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}
