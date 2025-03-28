package types

import (
	"io"
	"time"
)

// Basic type definitions for APFS

// OID represents an object identifier
type OID uint64

// XID represents a transaction identifier
type XID uint64

// PAddr represents a physical block address
type PAddr int64

// UUID represents a universally unique identifier
type UUID [16]byte

// PRange represents a range of physical blocks
type PRange struct {
	StartAddr  PAddr  // Starting block address
	BlockCount uint64 // Number of blocks
}

// Checksum represents a Fletcher 64 checksum
type Checksum [8]byte

// JKey represents the common header for file-system keys
type JKey struct {
	ObjIDAndType uint64 // Object ID and type
}

// GetObjID returns the object ID from a j_key
func (k *JKey) GetObjID() OID {
	return OID(k.ObjIDAndType & OBJ_ID_MASK)
}

// GetType returns the type from a j_key
func (k *JKey) GetType() uint8 {
	return uint8((k.ObjIDAndType & OBJECT_TYPE_MASK) >> OBJ_TYPE_SHIFT)
}

// NewJKey creates a new JKey with the given object ID and type
func NewJKey(objID OID, objType uint8) JKey {
	return JKey{
		ObjIDAndType: uint64(objID) | (uint64(objType) << OBJ_TYPE_SHIFT),
	}
}

// ObjectHeader represents the common header for all APFS objects (obj_phys_t)
type ObjectHeader struct {
	Cksum   Checksum // Fletcher 64 checksum
	OID     OID      // Object identifier
	XID     XID      // Transaction identifier
	Type    uint32   // Object type and flags
	Subtype uint32   // Object subtype
}

// IsEphemeral returns true if the object is ephemeral
func (o *ObjectHeader) IsEphemeral() bool {
	return (o.Type & ObjEphemeral) != 0
}

// IsPhysical returns true if the object is physical
func (o *ObjectHeader) IsPhysical() bool {
	return (o.Type & ObjPhysical) != 0
}

// IsVirtual returns true if the object is virtual
func (o *ObjectHeader) IsVirtual() bool {
	return (o.Type & ObjStorageTypeMask) == ObjVirtual
}

// IsEncrypted returns true if the object is encrypted
func (o *ObjectHeader) IsEncrypted() bool {
	return (o.Type & ObjEncrypted) != 0
}

// GetObjectType returns the object type without flags
func (o *ObjectHeader) GetObjectType() uint32 {
	return o.Type & ObjectTypeMask
}

// GetObjectFlags returns the object flags
func (o *ObjectHeader) GetObjectFlags() uint32 {
	return o.Type & ObjectTypeFlagsMask
}

// NLoc represents a location within a B-tree node
type NLoc struct {
	Off uint16 // Offset
	Len uint16 // Length
}

// KVLoc represents the location of a key and value
type KVLoc struct {
	Key NLoc // Key location
	Val NLoc // Value location
}

// KVOff represents offsets for fixed-size keys and values
type KVOff struct {
	Key uint16 // Key offset
	Val uint16 // Value offset
}

// BlockDevice defines the interface for accessing storage blocks
type BlockDevice interface {
	// ReadBlock reads a block at the given address
	ReadBlock(addr PAddr) ([]byte, error)

	// WriteBlock writes data to the block at the given address
	WriteBlock(addr PAddr, data []byte) error

	// GetBlockSize returns the size of blocks in bytes
	GetBlockSize() uint32

	// GetBlockCount returns the total number of blocks
	GetBlockCount() uint64

	// Close releases resources associated with the device
	Close() error
}

// KeyProvider defines the interface for accessing encryption keys
type KeyProvider interface {
	// GetVolumeKey retrieves the volume encryption key for a volume
	GetVolumeKey(volumeUUID UUID) ([]byte, error)

	// GetFileKey retrieves the encryption key for a file
	GetFileKey(volumeUUID UUID, fileID uint64) ([]byte, error)

	// Unlock attempts to unlock encrypted volumes with a password
	Unlock(volumeUUID UUID, password string) error
}

// Object defines the interface for APFS objects
type Object interface {
	// GetOID returns the object identifier
	GetOID() OID

	// GetXID returns the transaction identifier
	GetXID() XID

	// GetType returns the object type
	GetType() uint32

	// GetSubtype returns the object subtype
	GetSubtype() uint32

	// IsVirtual returns true if the object is virtual
	IsVirtual() bool

	// IsEphemeral returns true if the object is ephemeral
	IsEphemeral() bool

	// IsPhysical returns true if the object is physical
	IsPhysical() bool

	// IsEncrypted returns true if the object is encrypted
	IsEncrypted() bool

	// GetChecksum returns the object's checksum
	GetChecksum() Checksum

	// VerifyChecksum verifies the object's checksum
	VerifyChecksum([]byte) bool
}

// TimeSpec represents a timestamp with nanosecond precision
type TimeSpec uint64

// ToTime converts an APFS timestamp to a Go time.Time
func (ts TimeSpec) ToTime() time.Time {
	return time.Unix(0, int64(ts))
}

// FromTime converts a Go time.Time to an APFS timestamp
func FromTime(t time.Time) TimeSpec {
	return TimeSpec(t.UnixNano())
}

// FSTree represents a file system B-tree
type FSTree interface {
	// Lookup looks up a key in the tree
	Lookup(key []byte) ([]byte, error)

	// Insert inserts a key-value pair into the tree
	Insert(key, value []byte) error

	// Delete deletes a key from the tree
	Delete(key []byte) error

	// Iterate iterates over all entries in the tree
	Iterate(callback func(key, value []byte) error) error

	// IterateRange iterates over a range of keys
	IterateRange(startKey, endKey []byte, callback func(key, value []byte) error) error
}

// VolumeInfo contains summary information about a volume
type VolumeInfo struct {
	Index          uint32   // Volume index in container
	Name           string   // Volume name
	UUID           UUID     // Volume UUID
	Role           uint16   // Volume role
	NumFiles       uint64   // Number of files
	NumDirectories uint64   // Number of directories
	Capacity       uint64   // Total capacity in bytes
	Used           uint64   // Used space in bytes
	Created        TimeSpec // Creation time
	Modified       TimeSpec // Last modification time
	Encrypted      bool     // Whether the volume is encrypted
	CaseSensitive  bool     // Whether filenames are case-sensitive
}

// SnapshotInfo contains information about a snapshot
type SnapshotInfo struct {
	Name       string   // Snapshot name
	XID        XID      // Transaction ID
	UUID       UUID     // Snapshot UUID
	CreateTime TimeSpec // Creation time
	ChangeTime TimeSpec // Change time
}

// FileInfo represents information about a file or directory
type FileInfo struct {
	ObjectID    OID      // Object ID (inode number)
	ParentID    OID      // Parent directory ID
	Name        string   // File name
	Size        int64    // File size in bytes
	AllocSize   int64    // Allocated size in bytes
	Mode        uint16   // File mode
	UID         uint32   // Owner user ID
	GID         uint32   // Owner group ID
	CreateTime  TimeSpec // Creation time
	ModTime     TimeSpec // Modification time
	ChangeTime  TimeSpec // Change time (attributes)
	AccessTime  TimeSpec // Last access time
	Flags       uint64   // Flags
	LinkCount   int32    // Number of hard links
	IsDir       bool     // Whether this is a directory
	IsSymlink   bool     // Whether this is a symbolic link
	IsEncrypted bool     // Whether the file is encrypted
}

// DirectoryEntry represents an entry in a directory
type DirectoryEntry struct {
	ObjectID  OID      // Object ID (inode number)
	ParentID  OID      // Parent directory ID
	Name      string   // Entry name
	Type      uint8    // File type
	DateAdded TimeSpec // Date added to directory
	Flags     uint16   // Flags
}

// DataStream represents information about a data stream
type DataStream struct {
	Size              uint64 // Logical size in bytes
	AllocSize         uint64 // Allocated size in bytes
	DefaultCryptoID   uint64 // Default crypto ID
	TotalBytesWritten uint64 // Total bytes written
	TotalBytesRead    uint64 // Total bytes read
}

// FileExtent represents a file extent
type FileExtent struct {
	LogicalOffset uint64 // Logical offset in the file
	Length        uint64 // Length in bytes
	PhysicalAddr  PAddr  // Physical block address
	CryptoID      uint64 // Crypto ID
}

// TransactionOptions defines options for a transaction
type TransactionOptions struct {
	ReadOnly   bool // Whether the transaction is read-only
	Checkpoint bool // Whether to checkpoint after the transaction
}

// Transaction defines the interface for APFS transactions
type Transaction interface {
	// Commit commits the transaction
	Commit() error

	// Abort aborts the transaction
	Abort() error

	// CreateObject creates a new object
	CreateObject(objType, objSubtype uint32, size uint32) (OID, []byte, error)

	// UpdateObject updates an existing object
	UpdateObject(oid OID, data []byte) error

	// DeleteObject marks an object for deletion
	DeleteObject(oid OID) error

	// GetXID returns the transaction identifier
	GetXID() XID
}

// FileSystem defines the interface for accessing an APFS volume
type FileSystem interface {
	// GetRootDirectory returns the root directory
	GetRootDirectory() (FileInfo, error)

	// OpenFile opens a file at the specified path
	OpenFile(path string) (File, error)

	// Stat returns information about a file
	Stat(path string) (FileInfo, error)

	// Readdir reads a directory's contents
	Readdir(path string) ([]DirectoryEntry, error)

	// GetXattr gets an extended attribute
	GetXattr(path, name string) ([]byte, error)

	// ListXattr lists extended attributes
	ListXattr(path string) ([]string, error)

	// ReadSymlink reads a symbolic link
	ReadSymlink(path string) (string, error)

	// ListSnapshots lists snapshots
	ListSnapshots() ([]SnapshotInfo, error)

	// GetVolumeInfo returns information about the volume
	GetVolumeInfo() (VolumeInfo, error)

	// Close unmounts the file system
	Close() error
}

// File defines the interface for APFS files
type File interface {
	io.Reader
	io.ReaderAt
	io.Closer

	// Stat returns information about the file
	Stat() (FileInfo, error)

	// ReadExtent reads a specific extent
	ReadExtent(extent FileExtent, offset int64, size int) ([]byte, error)

	// GetExtents returns all extents for the file
	GetExtents() ([]FileExtent, error)
}

// ObjectResolver defines the interface for resolving object IDs
type ObjectResolver interface {
	// ResolveObject resolves an object by ID and transaction ID
	ResolveObject(oid OID, xid XID) ([]byte, error)

	// ReadPhysicalObject reads an object at a specific physical address
	ReadPhysicalObject(addr PAddr) ([]byte, error)
}
