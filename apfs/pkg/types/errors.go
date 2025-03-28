package types

import (
	"errors"
	"fmt"
)

// Common errors that can occur when working with APFS
var (
	// General errors
	ErrInvalidArgument = errors.New("invalid argument")
	ErrNotImplemented  = errors.New("feature not implemented")
	ErrUnknownFormat   = errors.New("unknown format")
	ErrNotSupported    = errors.New("operation not supported")
	ErrBadRequest      = errors.New("bad request")

	// File system errors
	ErrNotFound          = errors.New("object not found")
	ErrNotDirectory      = errors.New("not a directory")
	ErrNotFile           = errors.New("not a file")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrReadOnly          = errors.New("file system is read-only")
	ErrInvalidPath       = errors.New("invalid path")
	ErrDirectoryNotEmpty = errors.New("directory not empty")
	ErrFileExists        = errors.New("file already exists")
	ErrNoSpace           = errors.New("no space available")
	ErrIsDot             = errors.New("file is . or ..")

	// Block device errors
	ErrInvalidBlockSize = errors.New("invalid block size")
	ErrInvalidBlockAddr = errors.New("invalid block address")
	ErrIOError          = errors.New("I/O error")
	ErrDeviceFull       = errors.New("device is full")
	ErrDeviceOffline    = errors.New("device is offline")

	// Container errors
	ErrInvalidMagic         = errors.New("invalid magic number")
	ErrInvalidChecksum      = errors.New("invalid checksum")
	ErrInvalidSuperblock    = errors.New("invalid superblock")
	ErrUnsupportedVersion   = errors.New("unsupported APFS version")
	ErrNoValidCheckpoint    = errors.New("no valid checkpoint found")
	ErrInvalidObjectType    = errors.New("invalid object type")
	ErrObjectNotFound       = errors.New("object not found")
	ErrInvalidObjectID      = errors.New("invalid object ID")
	ErrInvalidTransactionID = errors.New("invalid transaction ID")
	ErrObjectMapCorrupted   = errors.New("object map corrupted")

	// Reaper errors
	ErrReaperBusy       = errors.New("reaper is busy")
	ErrReaperNotRunning = errors.New("reaper is not running")

	// Volume errors
	ErrInvalidVolume        = errors.New("invalid volume")
	ErrNoFreeSpace          = errors.New("no free space available")
	ErrVolumeNotMounted     = errors.New("volume not mounted")
	ErrVolumeAlreadyMounted = errors.New("volume already mounted")
	ErrIncompatibleFeature  = errors.New("incompatible feature")
	ErrReadOnlyVolume       = errors.New("volume is read-only")

	// B-tree errors
	ErrInvalidBtree        = errors.New("invalid B-tree")
	ErrInvalidBtreeNode    = errors.New("invalid B-tree node")
	ErrBtreeNodeFull       = errors.New("b-tree node is full")
	ErrInvalidBTreeKey     = errors.New("invalid b tree key")
	ErrInvalidBTreeKeySize = errors.New("invalid b tree key size")
	ErrKeyNotFound         = errors.New("key not found")

	// Encryption errors
	ErrNoKeyAvailable       = errors.New("no encryption key available")
	ErrInvalidEncryptionKey = errors.New("invalid encryption key")
	ErrWrongPassword        = errors.New("wrong password")
	ErrEncryptionFailed     = errors.New("encryption failed")
	ErrDecryptionFailed     = errors.New("decryption failed")
	ErrKeyUnwrapFailed      = errors.New("key unwrap failed")
	ErrInvalidKeybag        = errors.New("invalid keybag")
	ErrVolumeEncrypted      = errors.New("volume is encrypted")

	// Snapshot errors
	ErrSnapshotExists   = errors.New("snapshot already exists")
	ErrSnapshotNotFound = errors.New("snapshot not found")
	ErrInvalidSnapshot  = errors.New("invalid snapshot")
	ErrTooManySnapshots = errors.New("too many snapshots")
	ErrSnapshotInUse    = errors.New("snapshot is in use")

	// Fusion drive errors
	ErrNotFusionDrive = errors.New("not a fusion drive")
	ErrTierNotFound   = errors.New("tier not found")
	ErrInvalidTier    = errors.New("invalid tier")

	// Parsing errors
	ErrStructTooShort      = errors.New("data too short for structure")
	ErrInvalidStructSize   = errors.New("invalid structure size")
	ErrInvalidFieldValue   = errors.New("invalid field value")
	ErrInvalidObjectHeader = errors.New("invalid object header")
	ErrInvalidObjectFlags  = errors.New("invalid object flags")

	// Transaction errors
	ErrTransactionAborted    = errors.New("transaction aborted")
	ErrTransactionInProgress = errors.New("transaction in progress")
	ErrNoActiveTransaction   = errors.New("no active transaction")
)

// APFSError represents an error with additional APFS-specific context
type APFSError struct {
	Err       error  // The underlying error
	Operation string // The operation that caused the error
	Object    string // The object on which the operation was performed (path, OID, etc.)
	Detail    string // Additional details about the error
}

// Error implements the error interface
func (e *APFSError) Error() string {
	if e.Object != "" && e.Detail != "" {
		return fmt.Sprintf("%s: %s [%s]: %v", e.Operation, e.Object, e.Detail, e.Err)
	} else if e.Object != "" {
		return fmt.Sprintf("%s: %s: %v", e.Operation, e.Object, e.Err)
	} else if e.Detail != "" {
		return fmt.Sprintf("%s: %v [%s]", e.Operation, e.Err, e.Detail)
	}
	return fmt.Sprintf("%s: %v", e.Operation, e.Err)
}

// Unwrap returns the underlying error
func (e *APFSError) Unwrap() error {
	return e.Err
}

// NewAPFSError creates a new APFSError with the given details
func NewAPFSError(err error, operation string, object string, detail string) error {
	return &APFSError{
		Err:       err,
		Operation: operation,
		Object:    object,
		Detail:    detail,
	}
}

// IsNotFound returns true if the error indicates a "not found" condition
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound) || errors.Is(err, ErrObjectNotFound) || errors.Is(err, ErrKeyNotFound) || errors.Is(err, ErrSnapshotNotFound)
}

// IsReadOnly returns true if the error indicates a read-only condition
func IsReadOnly(err error) bool {
	return errors.Is(err, ErrReadOnly) || errors.Is(err, ErrReadOnlyVolume)
}

// IsChecksumError returns true if the error is related to invalid checksums
func IsChecksumError(err error) bool {
	return errors.Is(err, ErrInvalidChecksum)
}

// IsInvalidData returns true if the error indicates invalid data
func IsInvalidData(err error) bool {
	return errors.Is(err, ErrInvalidObjectType) || errors.Is(err, ErrInvalidMagic) ||
		errors.Is(err, ErrInvalidSuperblock) || errors.Is(err, ErrInvalidChecksum) ||
		errors.Is(err, ErrInvalidBtree) || errors.Is(err, ErrInvalidObjectHeader)
}

// IsEncryptionError returns true if the error is related to encryption
func IsEncryptionError(err error) bool {
	return errors.Is(err, ErrNoKeyAvailable) || errors.Is(err, ErrInvalidEncryptionKey) ||
		errors.Is(err, ErrWrongPassword) || errors.Is(err, ErrEncryptionFailed) ||
		errors.Is(err, ErrDecryptionFailed) || errors.Is(err, ErrKeyUnwrapFailed) ||
		errors.Is(err, ErrInvalidKeybag) || errors.Is(err, ErrVolumeEncrypted)
}

// IsIOError returns true if the error is related to I/O operations
func IsIOError(err error) bool {
	return errors.Is(err, ErrIOError) || errors.Is(err, ErrDeviceOffline)
}

// IsSpaceError returns true if the error indicates space-related issues
func IsSpaceError(err error) bool {
	return errors.Is(err, ErrNoFreeSpace) || errors.Is(err, ErrDeviceFull)
}
