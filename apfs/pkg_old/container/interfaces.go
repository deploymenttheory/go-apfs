package container

import (
	"github.com/yourusername/apfs/common"
)

// Container defines the interface for APFS container operations
type Container interface {
	// Open opens an APFS container from a block device
	Open() error

	// Close closes the container
	Close() error

	// GetSuperblock returns the container superblock
	GetSuperblock() (*NXSuperblock, error)

	// GetBlockSize returns the container's block size in bytes
	GetBlockSize() uint32

	// GetUUID returns the container's UUID
	GetUUID() common.UUID

	// GetMaxFileSystems returns the maximum number of file systems supported
	GetMaxFileSystems() uint32

	// GetVersion returns the container version
	GetVersion() APFSVersion

	// GetFeatures returns the container features flags
	GetFeatures() (uint64, uint64, uint64)

	// IsFusion returns true if the container is a Fusion Drive container
	IsFusion() bool

	// IsEncrypted returns true if the container uses encryption
	IsEncrypted() bool

	// IsReadOnly returns true if the container is mounted read-only
	IsReadOnly() bool

	// FindVolume finds a volume by name or UUID
	FindVolume(name string) (VolumeInfo, error)
	FindVolumeByUUID(uuid common.UUID) (VolumeInfo, error)

	// GetVolumeCount returns the number of volumes in the container
	GetVolumeCount() (uint32, error)

	// ListVolumes returns information about all volumes
	ListVolumes() ([]VolumeInfo, error)

	// GetCheckpointInfo returns information about the latest checkpoint
	GetCheckpointInfo() (CheckpointInfo, error)

	// GetSpaceInfo returns space usage information
	GetSpaceInfo() (SpaceInfo, error)

	// GetObjectMap returns the container's object map
	GetObjectMap() (ObjectMap, error)

	// GetSpaceManager returns the container's space manager
	GetSpaceManager() (SpaceManager, error)

	// GetDeviceInfo returns information about the underlying device
	GetDeviceInfo() (*common.BlockDeviceInfo, error)

	// GetTransaction creates a new transaction for container-level operations
	GetTransaction(options *common.TransactionOptions) (Transaction, error)
}

// VolumeInfo contains information about a volume within a container
type VolumeInfo struct {
	Index           uint32      // Volume index in container
	Role            uint16      // Volume role
	Name            string      // Volume name
	UUID            common.UUID // Volume UUID
	SuperblockOID   common.OID  // Volume superblock OID
	Reserved        uint64      // Reserved flags
	EncryptionState uint32      // Encryption state
	CaseSensitive   bool        // Whether filesystem is case-sensitive
}

// CheckpointInfo contains information about a container checkpoint
type CheckpointInfo struct {
	XID            common.XID   // Transaction ID
	CreateTime     uint64       // Creation timestamp
	DescBlocks     uint32       // Number of descriptor blocks
	DataBlocks     uint32       // Number of data blocks
	DescBase       common.PAddr // Descriptor base address
	DataBase       common.PAddr // Data base address
	DescNext       uint32       // Next descriptor index
	DataNext       uint32       // Next data index
	DescIndex      uint32       // Current descriptor index
	DescLen        uint32       // Descriptor length
	DataIndex      uint32       // Current data index
	DataLen        uint32       // Data length
	NumVirtualOIDs uint64       // Number of virtual OIDs
}

// SpaceInfo contains information about container space usage
type SpaceInfo struct {
	TotalBlocks      uint64 // Total blocks in container
	FreeBlocks       uint64 // Free blocks in container
	UsedByVolumes    uint64 // Blocks used by volumes
	UsedByContainer  uint64 // Blocks used by container metadata
	Reserved         uint64 // Reserved blocks
	Fragmentation    uint32 // Fragmentation percentage
	InternalPoolSize uint64 // Internal pool size
}

// Transaction defines the interface for container-level transactions
type Transaction interface {
	// Begin begins a transaction
	Begin() error

	// Commit commits the transaction
	Commit() error

	// Abort aborts the transaction
	Abort() error

	// GetXID returns the transaction ID
	GetXID() common.XID

	// IsActive returns true if the transaction is active
	IsActive() bool

	// CreateObject creates a new object
	CreateObject(objType, objSubtype uint32, size uint32) (common.OID, []byte, error)

	// UpdateObject updates an existing object
	UpdateObject(oid common.OID, data []byte) error

	// DeleteObject marks an object for deletion
	DeleteObject(oid common.OID) error

	// AllocateBlocks allocates blocks from the space manager
	AllocateBlocks(count uint64) (common.PAddr, error)

	// FreeBlocks returns blocks to the space manager
	FreeBlocks(addr common.PAddr, count uint64) error
}

// ObjectMap defines the interface for the container's object map
type ObjectMap interface {
	// Lookup looks up an object by OID and transaction ID
	Lookup(oid common.OID, xid common.XID) (common.PAddr, uint32, error)

	// LookupLatest looks up the latest version of an object
	LookupLatest(oid common.OID) (common.PAddr, uint32, common.XID, error)

	// Insert inserts or updates a mapping
	Insert(oid common.OID, xid common.XID, paddr common.PAddr, flags uint32) error

	// Delete deletes a mapping
	Delete(oid common.OID, xid common.XID) error

	// Iterate iterates over all mappings
	Iterate(callback func(oid common.OID, xid common.XID, paddr common.PAddr, flags uint32) error) error

	// GetSnapshotCount returns the number of snapshots
	GetSnapshotCount() uint32

	// GetInfo returns information about the object map
	GetInfo() (*OMapPhys, error)
}

// SpaceManager defines the interface for the container's space manager
type SpaceManager interface {
	// AllocateBlocks allocates free blocks
	AllocateBlocks(count uint64) (common.PAddr, error)

	// FreeBlocks marks blocks as free
	FreeBlocks(addr common.PAddr, count uint64) error

	// IsBlockAllocated checks if a block is allocated
	IsBlockAllocated(addr common.PAddr) (bool, error)

	// GetFreeBlockCount returns the number of free blocks
	GetFreeBlockCount() (uint64, error)

	// GetDeviceInfo returns information about device-specific space manager
	GetDeviceInfo(deviceIndex uint32) (*SpacemanDeviceInfo, error)

	// GetInfo returns information about the space manager
	GetInfo() (*SpacemanPhys, error)
}

// KeyManager defines the interface for container encryption key management
type KeyManager interface {
	// GetContainerKey gets the container encryption key
	GetContainerKey() ([]byte, error)

	// GetVolumeKey gets a volume encryption key
	GetVolumeKey(volumeUUID common.UUID) ([]byte, error)

	// UnlockContainer attempts to unlock the container with a password
	UnlockContainer(password string) error

	// UnlockVolume attempts to unlock a volume with a password
	UnlockVolume(volumeUUID common.UUID, password string) error

	// IsContainerUnlocked returns true if the container is unlocked
	IsContainerUnlocked() bool

	// IsVolumeUnlocked returns true if a volume is unlocked
	IsVolumeUnlocked(volumeUUID common.UUID) bool

	// GetKeybag returns the container keybag
	GetKeybag() (*KBLocker, error)
}

// Reaper defines the interface for the container reaper
type Reaper interface {
	// Start starts the reaper
	Start() error

	// Stop stops the reaper
	Stop() error

	// IsRunning returns true if the reaper is running
	IsRunning() bool

	// AddObject adds an object to be reaped
	AddObject(oid common.OID, xid common.XID) error

	// GetInfo returns information about the reaper
	GetInfo() (*NXReaperPhys, error)
}

// Factory creates container components
type Factory interface {
	// CreateContainer creates a new container instance
	CreateContainer(device common.BlockDevice) (Container, error)

	// OpenContainer opens an existing container
	OpenContainer(device common.BlockDevice) (Container, error)

	// CreateObjectMap creates an object map instance
	CreateObjectMap(container Container, omapOID common.OID) (ObjectMap, error)

	// CreateSpaceManager creates a space manager instance
	CreateSpaceManager(container Container, smOID common.OID) (SpaceManager, error)

	// CreateKeyManager creates a key manager instance
	CreateKeyManager(container Container) (KeyManager, error)

	// CreateReaper creates a reaper instance
	CreateReaper(container Container, reaperOID common.OID) (Reaper, error)
}
