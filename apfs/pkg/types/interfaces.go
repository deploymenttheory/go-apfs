// interfaces.go extends the types of types.go
package types

import (
	"io"
	"time"
)

// === Extended BlockDevice Interface ===

// BlockDeviceInfo contains information about a block device
type BlockDeviceInfo struct {
	Path       string // Path to the device
	BlockSize  uint32 // Block size in bytes
	BlockCount uint64 // Total number of blocks
	TotalSize  uint64 // Total size in bytes
	ReadOnly   bool   // Whether the device is read-only
	IsFusion   bool   // Whether this is a Fusion drive
	Identifier string // Device identifier
	Model      string // Device model
	Vendor     string // Device vendor
}

// BlockDeviceExtended extends the BlockDevice interface with additional operations
type BlockDeviceExtended interface {
	BlockDevice

	// ReadBlocks reads multiple contiguous blocks
	ReadBlocks(addr PAddr, count uint32) ([]byte, error)

	// WriteBlocks writes multiple contiguous blocks
	WriteBlocks(addr PAddr, data []byte) error

	// FlushWrites ensures all pending writes are committed to disk
	FlushWrites() error

	// GetDeviceInfo returns information about the device
	GetDeviceInfo() (*BlockDeviceInfo, error)

	// IsReadOnly returns true if the device is read-only
	IsReadOnly() bool

	// GetPartitions returns information about partitions on the device
	GetPartitions() ([]PartitionInfo, error)
}

// PartitionInfo contains information about a partition
type PartitionInfo struct {
	Index      uint32 // Partition index
	StartBlock uint64 // Starting block
	BlockCount uint64 // Number of blocks
	GUID       UUID   // Partition GUID
	TypeGUID   UUID   // Partition type GUID
	Name       string // Partition name
	IsAPFS     bool   // Whether this is an APFS partition
}

// === Extended Object Interface ===

// ObjectExtended extends the Object interface with additional operations
type ObjectExtended interface {
	Object

	// GetData returns the object's raw data
	GetData() ([]byte, error)

	// SetData updates the object's data
	SetData(data []byte) error

	// GetSize returns the object's size in bytes
	GetSize() uint32

	// GetCreationTime returns the object's creation time
	GetCreationTime() time.Time

	// GetModificationTime returns the object's modification time
	GetModificationTime() time.Time

	// IsInTransaction returns true if the object is part of an active transaction
	IsInTransaction() bool
}

// === Extended Transaction Interface ===

// TransactionExtended extends the Transaction interface with additional operations
type TransactionExtended interface {
	Transaction

	// IsActive returns true if the transaction is active
	IsActive() bool

	// GetChanges returns the changes made in this transaction
	GetChanges() ([]ObjectChange, error)

	// GetObjectCount returns the number of objects affected by this transaction
	GetObjectCount() int

	// GetDuration returns the duration of the transaction
	GetDuration() time.Duration

	// AddCallback registers a callback to be executed when the transaction commits
	AddCallback(callback func(xid XID) error) error

	// CanRollback returns true if the transaction can be rolled back
	CanRollback() bool

	// Rollback rolls back the transaction
	Rollback() error
}

// ObjectChange represents a change to an object in a transaction
type ObjectChange struct {
	OID       OID    // Object ID
	Type      uint32 // Object type
	Operation string // Operation (create, update, delete)
	BeforeXID XID    // Transaction ID before change
	AfterXID  XID    // Transaction ID after change
}

// === Extended File System Interface ===

// FileOperations represents operations that can be performed on files
type FileOperations uint32

const (
	// FileOpRead allows reading from the file
	FileOpRead FileOperations = 1 << iota
	// FileOpWrite allows writing to the file
	FileOpWrite
	// FileOpAppend allows appending to the file
	FileOpAppend
	// FileOpTruncate allows truncating the file
	FileOpTruncate
	// FileOpCreate allows creating the file if it doesn't exist
	FileOpCreate
	// FileOpExclusive fails if the file already exists with FileOpCreate
	FileOpExclusive
)

// FileSystemExtended extends the FileSystem interface with additional operations
type FileSystemExtended interface {
	FileSystem

	// CreateFile creates a new file
	CreateFile(path string, mode uint16) (File, error)

	// CreateDirectory creates a new directory
	CreateDirectory(path string, mode uint16) error

	// CreateSymlink creates a new symbolic link
	CreateSymlink(path, target string) error

	// Remove removes a file, directory, or symbolic link
	Remove(path string) error

	// RemoveAll removes a path and any children it contains
	RemoveAll(path string) error

	// Rename renames a file, directory, or symbolic link
	Rename(oldPath, newPath string) error

	// OpenFileExtended opens a file with specific access flags
	OpenFileExtended(path string, ops FileOperations) (FileExtended, error)

	// SetXattr sets an extended attribute
	SetXattr(path, name string, data []byte) error

	// RemoveXattr removes an extended attribute
	RemoveXattr(path, name string) error

	// Sync commits changes to disk
	Sync() error

	// CreateSnapshot creates a new snapshot
	CreateSnapshot(name string) (SnapshotInfo, error)

	// DeleteSnapshot deletes a snapshot
	DeleteSnapshot(name string) error

	// MountSnapshot mounts a snapshot
	MountSnapshot(name string) (FileSystemExtended, error)

	// GetFreespace returns the amount of free space in bytes
	GetFreespace() (uint64, error)
}

// === Extended File Interface ===

// FileExtended extends the File interface with additional operations
type FileExtended interface {
	File
	io.Writer
	io.WriterAt

	// Truncate changes the size of the file
	Truncate(size int64) error

	// Sync commits changes to disk
	Sync() error

	// Chown changes the numeric owner and group of the file
	Chown(uid, gid uint32) error

	// Chmod changes the mode of the file
	Chmod(mode uint16) error

	// SetFlags sets the flags on the file
	SetFlags(flags uint64) error

	// SetTimes sets the access and modification times
	SetTimes(atime, mtime time.Time) error

	// SetXattr sets an extended attribute
	SetXattr(name string, data []byte) error

	// RemoveXattr removes an extended attribute
	RemoveXattr(name string) error
}

// === Extended ObjectResolver Interface ===

// ObjectResolverExtended extends the ObjectResolver interface with additional operations
type ObjectResolverExtended interface {
	ObjectResolver

	// ResolveObjectWithXID resolves an object by ID and transaction ID
	ResolveObjectWithXID(oid OID, xid XID) ([]byte, error)

	// ResolveLatestObject resolves the latest version of an object
	ResolveLatestObject(oid OID) ([]byte, error)

	// ListSnapshots lists snapshots for an object
	ListObjectSnapshots(oid OID) ([]XID, error)

	// WritePhysicalObject writes an object at a specific physical address
	WritePhysicalObject(addr PAddr, data []byte) error

	// AllocateObject allocates space for a new object
	AllocateObject(size uint32) (PAddr, error)

	// DeallocateObject deallocates an object
	DeallocateObject(addr PAddr) error
}

// === Extended KeyProvider Interface ===

// KeyProviderExtended extends the KeyProvider interface with additional operations
type KeyProviderExtended interface {
	KeyProvider

	// GetVolumeUnlockRecords gets the volume unlock records
	GetVolumeUnlockRecords(volumeUUID UUID) ([]byte, error)

	// GetRecoveryKey gets the recovery key for a volume
	GetRecoveryKey(volumeUUID UUID) ([]byte, error)

	// UnlockWithRecoveryKey unlocks a volume with a recovery key
	UnlockWithRecoveryKey(volumeUUID UUID, recoveryKey string) error

	// ChangePassword changes the password for a volume
	ChangePassword(volumeUUID UUID, oldPassword, newPassword string) error

	// AddRecoveryKey adds a recovery key to a volume
	AddRecoveryKey(volumeUUID UUID, password string) (string, error)

	// RemoveRecoveryKey removes a recovery key from a volume
	RemoveRecoveryKey(volumeUUID UUID, password string) error

	// IsUnlocked returns true if the volume is unlocked
	IsUnlocked(volumeUUID UUID) bool
}
