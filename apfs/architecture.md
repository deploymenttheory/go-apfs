# APFS Project Architecture

## Overview

This architecture follows the layered design of Apple File System as described in the reference documentation, with a clear separation between the container layer and file-system layer.

## Directory Structure

```
apfs/
├── cmd/                     # Command-line interfaces
│   ├── apfs-info/           # Tool to display APFS container/volume info
│   ├── apfs-mount/          # Tool to mount APFS volumes
│   └── apfs-recover/        # Data recovery tool
├── internal/                # Non-exported internal packages
│   └── binary/              # Binary parsing utilities
└── pkg/                     # Exported package code
    ├── checksum/            # Checksum
    │   └── fletcher64.go    # Checksum algorithm
    ├── types/               # Core types and constants
    │   ├── constants.go     # All APFS constants from the spec
    │   ├── errors.go        # Error definitions
    │   ├── types.go         # Common data structures and interfaces
    │   ├── interfaces.go    # Extended interfaces
    │   ├── structs.go       # All APFS on-disk data structures as Go structs
    │   ├── binary.go        # Serialization/deserialization helpers
    │   └── version.go       # Version compatibility checking
    ├── container/           # Container layer
    │   ├── object.go        # Object structures (obj_phys_t)
    │   ├── checkpoint.go    # Checkpoint mechanism
    │   ├── container.go     # Container manager (nx_superblock_t)
    │   ├── omap.go          # Object maps
    │   ├── spaceman.go      # Space manager
    │   ├── btree.go         # B-tree structures
    │   └── reaper.go        # Reaper for delayed deletion
    ├── fs/                  # File system layer
    │   ├── volume.go        # Volume structures (apfs_superblock_t)
    │   ├── inode.go         # Inode structures and operations
    │   ├── dentry.go        # Directory entry structures
    │   ├── xattr.go         # Extended attributes
    │   ├── datastream.go    # File data stream handling
    │   ├── extents.go       # File extent management
    │   ├── extfields.go     # Extended field handling
    │   └── siblings.go      # Hard link management
    ├── crypto/              # Encryption support
    │   ├── keybag.go        # Keybag structures and handling
    │   ├── keys.go          # KEK/VEK key management
    │   └── crypto.go        # Encryption/decryption utilities
    ├── snapshot/            # Snapshot management
    │   ├── snapshot.go      # Snapshot structures
    │   └── operations.go    # Snapshot operations
    ├── fusion/              # Fusion drive support
    │   ├── fusion.go        # Fusion drive structures
    │   └── tier.go          # Tier management
    ├── transaction/         # Transaction handling
    │   ├── transaction.go   # Transaction structures and operations
    │   └── operations.go    # Operation interfaces
    └── util/                # Utilities
        ├── io.go            # I/O utilities
        ├── checksum.go      # Fletcher64 implementation
        ├── bits.go          # Bit manipulation utilities
        └── uuid.go          # UUID handling
```

Looking at the APFS reference document and your Go project architecture, I'd recommend implementing the system in the following logical order:

1. **Core Types and Constants** (types/constants.go, types/types.go)
   - Define fundamental types like OID, XID, PAddr, etc.
   - Set up essential constants and error types
   - Implement basic interfaces (BlockDevice, Object, etc.)

2. **IO and Utilities** (util/io.go, util/checksum.go)
   - Implement the block device interface for reading/writing blocks
   - Build Fletcher64 checksum implementation
   - Add UUID handling utilities

3. **Container Layer Core**
   - Object handling (container/object.go)
   - Container superblock parsing (container/container.go)
   - Checkpoint mechanism (container/checkpoint.go)

4. **Basic Reading Infrastructure**
   - Object map implementation (container/omap.go)
   - B-tree structures (container/btree.go)
   - Space manager basics (container/spaceman.go)

5. **File System Layer Basics**
   - Volume superblock (fs/volume.go)
   - Inode structures (fs/inode.go)
   - Directory entries (fs/dentry.go)

6. **Data Access**
   - Data streams (fs/datastream.go)
   - File extents (fs/extents.go)
   - Extended fields (fs/extfields.go)

7. **Extended Functionality**
   - Extended attributes (fs/xattr.go)
   - Hard link handling (fs/siblings.go)
   - Snapshot support (snapshot/)

8. **Encryption Support** (crypto/)
   - Keybag handling
   - Encryption/decryption utilities

9. **Advanced Features**
   - Transaction handling (transaction/)
   - Reaper implementation (container/reaper.go)
   - Fusion drive support (fusion/)

10. **Command-Line Tools**
    - Info tool (cmd/apfs-info)
    - Mount tool with FUSE (cmd/apfs-mount)
    - Recovery tool (cmd/apfs-recover)

This order follows a logical progression from the lowest-level building blocks to more advanced features, allowing you to build and test a basic read-only implementation before expanding to more complex features. I'd recommend implementing a minimal read-only file system first, then adding support for snapshots, encryption, and finally write capabilities.

## Structs and Functions by File

### pkg/types/

#### constants.go
```go
// APFS magic constants
const (
    NXMagic    uint32 = 0x4253584E // 'NXSB'
    APFSMagic  uint32 = 0x42535041 // 'APSB'
    JSDRMagic  uint32 = 0x5244534A // 'JSDR'
)

// Block size constants
const (
    MinBlockSize   uint32 = 4096
    DefBlockSize   uint32 = 4096
    MaxBlockSize   uint32 = 65536
    MinContainerSize uint64 = 1048576
)

// Object types
const (
    ObjIDMask         uint64 = 0x0fffffffffffffffULL
    ObjTypeMask       uint64 = 0xf000000000000000ULL
    ObjTypeShift      uint8  = 60
    
    // Object storage types
    ObjVirtual   uint32 = 0x00000000
    ObjEphemeral uint32 = 0x80000000
    ObjPhysical  uint32 = 0x40000000
    
    // Object types
    ObjectTypeNXSuperblock uint32 = 0x00000001
    ObjectTypeBtree        uint32 = 0x00000002
    ObjectTypeBtreeNode    uint32 = 0x00000003
    ObjectTypeSpaceman     uint32 = 0x00000005
    // ... additional object types
)

// File system record types
const (
    APFSTypeAny          uint64 = 0
    APFSTypeSnapMetadata uint64 = 1
    APFSTypeExtent       uint64 = 2
    APFSTypeInode        uint64 = 3
    APFSTypeXattr        uint64 = 4
    // ... additional record types
)

// Volume roles
const (
    APFSVolRoleNone     uint16 = 0x0000
    APFSVolRoleSystem   uint16 = 0x0001
    APFSVolRoleUser     uint16 = 0x0002
    APFSVolRoleRecovery uint16 = 0x0004
    // ... additional volume roles
)

// Inode flags
const (
    InodeIsAPFSPrivate     uint64 = 0x00000001
    InodeMaintainDirStats  uint64 = 0x00000002
    InodeHasFinderInfo     uint64 = 0x00000100
    // ... additional inode flags
)

// Other constants for encryption, features, extended attributes, etc.
```

#### errors.go
```go
package types

import "errors"

// Common errors
var (
    ErrInvalidChecksum      = errors.New("invalid checksum")
    ErrInvalidMagic         = errors.New("invalid magic number")
    ErrInvalidBlockSize     = errors.New("invalid block size")
    ErrUnsupportedVersion   = errors.New("unsupported APFS version")
    ErrInvalidBlockAddress  = errors.New("invalid block address")
    ErrNoValidCheckpoint    = errors.New("no valid checkpoint found")
    ErrNotFound             = errors.New("object not found")
    ErrNotImplemented       = errors.New("feature not implemented")
    ErrNoFreeSpace          = errors.New("no free space available")
    ErrNoKeyAvailable       = errors.New("no encryption key available")
    ErrStructTooShort       = errors.New("data too short for structure")
    ErrInvalidObjectType    = errors.New("invalid object type")
    ErrVolumeNotMounted     = errors.New("volume not mounted")
    ErrSnapshotExists       = errors.New("snapshot already exists")
)
```

#### types.go
```go
package types

// PAddr represents a physical block address
type PAddr int64

// XID represents a transaction identifier
type XID uint64

// OID represents an object identifier
type OID uint64

// PRange represents a range of physical blocks
type PRange struct {
    StartAddr PAddr  // Starting block address
    BlockCount uint64 // Number of blocks
}

// UUID is a universally unique identifier
type UUID [16]byte

// Common interfaces

// Object represents an APFS object with a header
type Object interface {
    GetOID() OID
    GetXID() XID
    GetType() uint32
    GetSubtype() uint32
    IsVirtual() bool
    IsEphemeral() bool
    IsPhysical() bool
    IsEncrypted() bool
    Checksum() [8]byte
    VerifyChecksum([]byte) bool
}

// BlockDevice provides access to the underlying storage
type BlockDevice interface {
    ReadBlock(addr PAddr) ([]byte, error)
    WriteBlock(addr PAddr, data []byte) error
    GetBlockSize() uint32
    GetBlockCount() uint64
    Close() error
}

// KeyProvider provides access to encryption keys
type KeyProvider interface {
    GetVolumeKey(volumeUUID UUID) ([]byte, error)
    GetFileKey(volumeUUID UUID, fileID uint64) ([]byte, error)
}
```

### pkg/container/

#### object.go
```go
package container

import (
    "bytes"
    "encoding/binary"
    
    "apfs/pkg/types"
    "apfs/pkg/util"
)

// ObjectPhys represents the common header of all APFS objects (obj_phys_t)
type ObjectPhys struct {
    Checksum [8]byte   // Fletcher 64 checksum
    OID      types.OID // Object identifier
    XID      types.XID // Transaction identifier
    Type     uint32    // Object type and flags
    Subtype  uint32    // Object subtype
}

// ObjectFromBytes parses an object from raw bytes
func ObjectFromBytes(data []byte) (*ObjectPhys, error) {
    // Validate data length
    // Parse fields
    // Return parsed object
}

// GetOID returns the object identifier
func (o *ObjectPhys) GetOID() types.OID {
    return o.OID
}

// GetXID returns the transaction identifier
func (o *ObjectPhys) GetXID() types.XID {
    return o.XID
}

// GetType returns the object type
func (o *ObjectPhys) GetType() uint32 {
    return o.Type & types.ObjectTypeMask
}

// GetSubtype returns the object subtype
func (o *ObjectPhys) GetSubtype() uint32 {
    return o.Subtype
}

// GetStorageType returns the storage type (virtual/physical/ephemeral)
func (o *ObjectPhys) GetStorageType() uint32 {
    return o.Type & types.ObjStorageTypeMask
}

// IsVirtual returns true if the object is virtual
func (o *ObjectPhys) IsVirtual() bool {
    return o.GetStorageType() == types.ObjVirtual
}

// IsEphemeral returns true if the object is ephemeral
func (o *ObjectPhys) IsEphemeral() bool {
    return o.GetStorageType() == types.ObjEphemeral
}

// IsPhysical returns true if the object is physical
func (o *ObjectPhys) IsPhysical() bool {
    return o.GetStorageType() == types.ObjPhysical
}

// IsEncrypted returns true if the object is encrypted
func (o *ObjectPhys) IsEncrypted() bool {
    return (o.Type & types.ObjEncrypted) != 0
}

// VerifyChecksum verifies the object's checksum
func (o *ObjectPhys) VerifyChecksum(data []byte) bool {
    // Create copy of data with checksum zeroed
    // Calculate Fletcher64 checksum
    // Compare with stored checksum
}

// SetChecksum calculates and sets the checksum for the object
func (o *ObjectPhys) SetChecksum(data []byte) {
    // Create copy of data with checksum zeroed
    // Calculate Fletcher64 checksum
    // Store the checksum
}

// Serialize converts the object to bytes
func (o *ObjectPhys) Serialize() ([]byte, error) {
    // Create buffer
    // Write fields
    // Return serialized data
}
```

#### checkpoint.go
```go
package container

import (
    "apfs/pkg/types"
)

// CheckpointMapping represents a mapping between an ephemeral object and its physical location
type CheckpointMapping struct {
    Type     uint32    // Object type
    Subtype  uint32    // Object subtype
    Size     uint32    // Size in bytes
    Pad      uint32    // Reserved
    FSOID    types.OID // Volume object ID
    OID      types.OID // Object ID
    PAddr    types.PAddr // Physical address
}

// CheckpointMapPhys represents a checkpoint mapping block
type CheckpointMapPhys struct {
    Object   ObjectPhys          // Object header
    Flags    uint32              // Flags
    Count    uint32              // Number of mappings
    Mappings []CheckpointMapping // Array of mappings
}

// IsLastCheckpointMap returns true if this is the last mapping block
func (c *CheckpointMapPhys) IsLastCheckpointMap() bool {
    return c.Flags&types.CheckpointMapLast != 0
}

// CheckpointInfo represents a container checkpoint
type CheckpointInfo struct {
    Superblock     *NXSuperblock
    CheckpointMaps []*CheckpointMapPhys
    XID            types.XID
}

// LoadCheckpoint loads a checkpoint from the specified location
func LoadCheckpoint(device types.BlockDevice, descBase types.PAddr, index uint32) (*CheckpointInfo, error) {
    // Read checkpoint descriptor
    // Verify checksums
    // Parse superblock and mapping blocks
    // Return checkpoint info
}

// FindLatestCheckpoint finds the latest valid checkpoint
func FindLatestCheckpoint(device types.BlockDevice, superblock *NXSuperblock) (*CheckpointInfo, error) {
    // Find all checkpoints
    // Verify their validity
    // Return the one with the highest XID
}

// LoadEphemeralObject loads an ephemeral object from the checkpoint
func LoadEphemeralObject(device types.BlockDevice, checkpoint *CheckpointInfo, mapping CheckpointMapping) ([]byte, error) {
    // Calculate location in checkpoint data area
    // Read the object
    // Verify checksum
    // Return object data
}
```

#### container.go
```go
package container

import (
    "apfs/pkg/types"
    "apfs/pkg/util"
)

// NXSuperblock represents a container superblock (nx_superblock_t)
type NXSuperblock struct {
    Object                 ObjectPhys        // Object header
    Magic                  uint32            // Magic ('NXSB')
    BlockSize              uint32            // Block size
    BlockCount             uint64            // Number of blocks
    Features               uint64            // Optional features
    ReadOnlyCompat         uint64            // Read-only compatible features
    IncompatFeatures       uint64            // Incompatible features
    UUID                   types.UUID        // Container UUID
    NextOID                types.OID         // Next object ID
    NextXID                types.XID         // Next transaction ID
    XPDescBlocks           uint32            // Checkpoint descriptor blocks
    XPDataBlocks           uint32            // Checkpoint data blocks
    XPDescBase             types.PAddr       // Checkpoint descriptor base
    XPDataBase             types.PAddr       // Checkpoint data base
    XPDescNext             uint32            // Next checkpoint descriptor
    XPDataNext             uint32            // Next checkpoint data
    XPDescIndex            uint32            // Checkpoint descriptor index
    XPDescLen              uint32            // Checkpoint descriptor length
    XPDataIndex            uint32            // Checkpoint data index
    XPDataLen              uint32            // Checkpoint data length
    SpacemanOID            types.OID         // Space manager OID
    OMapOID                types.OID         // Object map OID
    ReaperOID              types.OID         // Reaper OID
    TestType               uint32            // For testing
    MaxFileSystems         uint32            // Max number of volumes
    FSOID                  [100]types.OID    // Volume OIDs
    Counters               [32]uint64        // Array of counters
    BlockedOutRange        types.PRange      // Blocked out range
    EvictMappingTreeOID    types.OID         // Evict mapping tree OID
    Flags                  uint64            // Flags
    EFIJumpstart          types.PAddr       // EFI jumpstart
    FusionUUID             types.UUID        // Fusion UUID
    KeyLocker              types.PRange      // Keybag location
    EphemeralInfo          [4]uint64         // Ephemeral info
    // Additional fields...
}

// IsValid checks if the superblock is valid
func (sb *NXSuperblock) IsValid() bool {
    return sb.Magic == types.NXMagic
}

// HasValidVersion checks if the container has a supported version
func (sb *NXSuperblock) HasValidVersion() bool {
    return sb.IncompatFeatures&types.NXIncompatVersion2 != 0
}

// ContainerManager manages an APFS container
type ContainerManager struct {
    Device           types.BlockDevice
    Superblock       *NXSuperblock
    Checkpoint       *CheckpointInfo
    SpaceManager     *SpaceManager
    ObjectMap        *OMapPhys
    EphemeralObjects map[types.OID][]byte
    KeyProvider      types.KeyProvider
}

// NewContainerManager creates a new container manager
func NewContainerManager(device types.BlockDevice) (*ContainerManager, error) {
    // Read superblock from block 0
    // Find latest checkpoint
    // Load ephemeral objects
    // Initialize space manager and object map
    // Return container manager
}

// GetVolume returns a volume by index
func (cm *ContainerManager) GetVolume(index uint32) (*VolumeManager, error) {
    // Validate index
    // Get volume OID
    // Resolve volume superblock
    // Create and return volume manager
}

// ListVolumes returns information about all volumes
func (cm *ContainerManager) ListVolumes() ([]VolumeInfo, error) {
    // Iterate through volume OIDs
    // Collect information about each volume
    // Return volumes info
}

// ResolveObject resolves an object by its ID and transaction ID
func (cm *ContainerManager) ResolveObject(oid types.OID, xid types.XID) ([]byte, error) {
    // Check if it's an ephemeral object
    // Check if it's a physical object
    // Look up virtual object in object map
    // Return object data
}

// ReadPhysicalObject reads an object at a specific physical address
func (cm *ContainerManager) ReadPhysicalObject(addr types.PAddr) ([]byte, error) {
    // Read the object from the device
    // Verify the checksum
    // Return the object data
}
```

#### omap.go
```go
package container

import (
    "apfs/pkg/types"
)

// OMapKey represents a key in the object map
type OMapKey struct {
    OID types.OID // Object identifier
    XID types.XID // Transaction identifier
}

// OMapVal represents a value in the object map
type OMapVal struct {
    Flags uint32      // Flags
    Size  uint32      // Size in bytes
    PAddr types.PAddr // Physical address
}

// IsDeleted returns true if the object has been deleted
func (v *OMapVal) IsDeleted() bool {
    return v.Flags&types.OMapValDeleted != 0
}

// OMapSnapshot represents a snapshot in the object map
type OMapSnapshot struct {
    Flags uint32    // Flags
    Pad   uint32    // Reserved
    OID   types.OID // Reserved
}

// OMapPhys represents an object map
type OMapPhys struct {
    Object            ObjectPhys  // Object header
    Flags             uint32      // Flags
    SnapCount         uint32      // Number of snapshots
    TreeType          uint32      // Tree type
    SnapshotTreeType  uint32      // Snapshot tree type
    TreeOID           types.OID   // Tree OID
    SnapshotTreeOID   types.OID   // Snapshot tree OID
    MostRecentSnap    types.XID   // Most recent snapshot
    PendingRevertMin  types.XID   // Pending revert minimum
    PendingRevertMax  types.XID   // Pending revert maximum
}

// IsManuallyManaged returns true if the object map doesn't support snapshots
func (o *OMapPhys) IsManuallyManaged() bool {
    return o.Flags&types.OMapManuallyManaged != 0
}

// CreateObjectMap creates a new object map
func CreateObjectMap(container *ContainerManager, tx *Transaction) (*OMapPhys, error) {
    // Allocate blocks for the B-tree
    // Initialize the object map structure
    // Return the new object map
}

// LookupObject looks up an object in the object map
func LookupObject(container *ContainerManager, omap *OMapPhys, oid types.OID, xid types.XID) (*OMapVal, error) {
    // Load the B-tree
    // Create the key
    // Search for the key
    // Parse and return the value
}
```

#### spaceman.go
```go
package container

import (
    "apfs/pkg/types"
)

// ChunkInfo represents information about a chunk of blocks
type ChunkInfo struct {
    XID         types.XID   // Transaction ID
    Addr        types.PAddr // Base address
    BlockCount  uint32      // Total blocks
    FreeCount   uint32      // Free blocks
    BitmapAddr  types.PAddr // Bitmap address
}

// ChunkInfoBlock represents a block containing chunk information
type ChunkInfoBlock struct {
    Object      ObjectPhys  // Object header
    Index       uint32      // Index
    Count       uint32      // Number of entries
    Chunks      []ChunkInfo // Chunk entries
}

// CIBAddrBlock represents a block of chunk info block addresses
type CIBAddrBlock struct {
    Object      ObjectPhys    // Object header
    Index       uint32        // Index
    Count       uint32        // Number of entries
    Addresses   []types.PAddr // Addresses
}

// SpacemanFreeQueue represents a free queue
type SpacemanFreeQueue struct {
    Count         uint64    // Number of entries
    TreeOID       types.OID // B-tree object ID
    OldestXID     types.XID // Oldest transaction ID
    TreeNodeLimit uint16    // Tree node limit
    // Pad fields...
}

// SpacemanDevice represents device-specific space manager information
type SpacemanDevice struct {
    BlockCount  uint64 // Total blocks
    ChunkCount  uint64 // Total chunks
    CIBCount    uint32 // Chunk info blocks
    CABCount    uint32 // CIB address blocks
    FreeCount   uint64 // Free blocks
    AddrOffset  uint32 // Address offset
    // Reserved fields...
}

// SpacemanPhys represents the space manager
type SpacemanPhys struct {
    Object            ObjectPhys             // Object header
    BlockSize         uint32                 // Block size
    BlocksPerChunk    uint32                 // Blocks per chunk
    ChunksPerCIB      uint32                 // Chunks per CIB
    CIBsPerCAB        uint32                 // CIBs per CAB
    Devices           [2]SpacemanDevice      // Main and tier2 devices
    Flags             uint32                 // Flags
    IPBmTxMultiplier  uint32                 // Internal pool bitmap TX multiplier
    IPBlockCount      uint64                 // Internal pool block count
    IPBmSizeInBlocks  uint32                 // Internal pool bitmap size
    IPBmBlockCount    uint32                 // Internal pool bitmap block count
    IPBmBase          types.PAddr            // Internal pool bitmap base
    IPBase            types.PAddr            // Internal pool base
    FreeQueueCount    uint64                 // Free queue count
    FreeQueues        [3]SpacemanFreeQueue   // Free queues (IP, main, tier2)
    // Additional fields...
}

// SpaceManager manages block allocation
type SpaceManager struct {
    Container      *ContainerManager
    Spaceman       *SpacemanPhys
    InternalPool   *InternalPoolManager
    CIBAddrBlocks  map[uint32]*CIBAddrBlock
    CIBBlocks      map[types.PAddr]*ChunkInfoBlock
    BitmapBlocks   map[types.PAddr][]byte
}

// InternalPoolManager manages the internal pool allocation
type InternalPoolManager struct {
    Spaceman     *SpaceManager
    BitmapBase   types.PAddr
    BitmapBlocks uint32
    FreeHead     uint16
    FreeTail     uint16
    Bitmap       []byte
}

// NewSpaceManager creates a new space manager
func NewSpaceManager(container *ContainerManager) (*SpaceManager, error) {
    // Get space manager object from container
    // Parse space manager structure
    // Initialize internal pool if configured
    // Return space manager
}

// AllocateBlock allocates a new block
func (sm *SpaceManager) AllocateBlock() (types.PAddr, error) {
    // Try internal pool first
    // Then try main device
    // Then try tier2 device if Fusion
    // Return allocated block address
}

// FreeBlock frees a previously allocated block
func (sm *SpaceManager) FreeBlock(addr types.PAddr, immediate bool) error {
    // Check which device the block belongs to
    // If immediate, update bitmap directly
    // Otherwise, add to free queue
    // Return success/error
}

// AllocateContiguousBlocks allocates multiple contiguous blocks
func (sm *SpaceManager) AllocateContiguousBlocks(count uint32) (types.PAddr, error) {
    // Search for contiguous free space
    // Mark blocks as allocated
    // Return starting address
}
```

#### btree.go
```go
package container

import (
    "apfs/pkg/types"
)

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

// BTreeNodePhys represents a B-tree node
type BTreeNodePhys struct {
    Object        ObjectPhys // Object header
    Flags         uint16     // Flags
    Level         uint16     // Level in tree (0 = leaf)
    KeyCount      uint32     // Number of keys
    TableSpace    NLoc       // Table of contents location
    FreeSpace     NLoc       // Free space location
    KeyFreeList   NLoc       // Free list for keys
    ValFreeList   NLoc       // Free list for values
    Data          []byte     // Node data
}

// IsLeaf returns true if this is a leaf node
func (n *BTreeNodePhys) IsLeaf() bool {
    return n.Flags&types.BtnodeLeaf != 0
}

// BTreeInfoFixed represents static B-tree information
type BTreeInfoFixed struct {
    Flags     uint32 // Flags
    NodeSize  uint32 // Node size
    KeySize   uint32 // Key size (0 if variable)
    ValSize   uint32 // Value size (0 if variable)
}

// BTreeInfo represents B-tree information
type BTreeInfo struct {
    Fixed       BTreeInfoFixed // Fixed information
    LongestKey  uint32         // Longest key
    LongestVal  uint32         // Longest value
    KeyCount    uint64         // Total keys
    NodeCount   uint64         // Total nodes
}

// BTree represents a B-tree
type BTree struct {
    Container    *ContainerManager
    RootNode     *BTreeNodePhys
    Info         *BTreeInfo
    KeySize      uint32
    ValSize      uint32
    TreeType     uint32
    TreeSubtype  uint32
    AllowGhosts  bool
    IsFixed      bool
    IsHashed     bool
    IsPhysical   bool
}

// NewBTree creates a new B-tree from a root node
func NewBTree(container *ContainerManager, data []byte) (*BTree, error) {
    // Parse the root node
    // Extract tree information
    // Return the B-tree
}

// Search searches for a key in the B-tree
func (bt *BTree) Search(key []byte) ([]byte, error) {
    // Start search at root node
    // Traverse the tree
    // Return the value if found
}

// Insert inserts a key-value pair into the B-tree
func (bt *BTree) Insert(tx *Transaction, key, value []byte) error {
    // Find the appropriate leaf node
    // Insert the key-value pair
    // Split nodes if necessary
    // Update tree metadata
}

// Delete deletes a key from the B-tree
func (bt *BTree) Delete(tx *Transaction, key []byte) error {
    // Find the key
    // Remove the key-value pair
    // Merge nodes if necessary
    // Update tree metadata
}

// IterateRange iterates over a range of keys
func (bt *BTree) IterateRange(start, end []byte, callback func(key, value []byte) bool) error {
    // Find the start key
    // Iterate through keys in range
    // Call callback for each key-value pair
}
```

#### reaper.go
```go
package container

import (
    "apfs/pkg/types"
)

// NXReaperPhys represents the reaper structure (nx_reaper_phys_t)
type NXReaperPhys struct {
    Object          ObjectPhys  // Object header
    NextReapID      uint64      // Next reap ID
    CompletedID     uint64      // Completed ID
    Head            types.OID   // Head of reap list
    Tail            types.OID   // Tail of reap list
    Flags           uint32      // Flags
    RLCount         uint32      // Reap list count
    Type            uint32      // Type
    Size            uint32      // Size
    FSOID           types.OID   // File system OID
    OID             types.OID   // Object ID
    XID             types.XID   // Transaction ID
    NRLEFlags       uint32      // Reap list entry flags
    StateBufferSize uint32      // State buffer size
    StateBuffer     []byte      // State buffer
}

// NXReapListPhys represents a reap list (nx_reap_list_phys_t)
type NXReapListPhys struct {
    Object    ObjectPhys         // Object header
    Next      types.OID          // Next list
    Flags     uint32             // Flags
    Max       uint32             // Maximum entries
    Count     uint32             // Current count
    First     uint32             // First entry
    Last      uint32             // Last entry
    Free      uint32             // Free entry
    Entries   []NXReapListEntry  // Entries
}

// NXReapListEntry represents an entry in a reap list (nx_reap_list_entry_t)
type NXReapListEntry struct {
    Next      uint32     // Next entry
    Flags     uint32     // Flags
    Type      uint32     // Type
    Size      uint32     // Size
    FSOID     types.OID  // File system OID
    OID       types.OID  // Object ID
    XID       types.XID  // Transaction ID
}

// OMapReapState represents the object map reaper state (omap_reap_state_t)
type OMapReapState struct {
    Phase     uint32   // Reaper phase (OMAP_REAP_PHASE_*)
    Key       OMapKey  // Current key
}

// OMapCleanupState represents the state used when cleaning up deleted snapshots (omap_cleanup_state_t)
type OMapCleanupState struct {
    Cleaning   uint32      // Flag indicating valid data
    OMSFlags   uint32      // Snapshot flags
    SXIDPrev   types.XID   // Transaction ID of previous snapshot
    SXIDStart  types.XID   // Transaction ID of first snapshot being deleted
    SXIDEnd    types.XID   // Transaction ID of last snapshot being deleted
    SXIDNext   types.XID   // Transaction ID of next snapshot
    CurKey     OMapKey     // Current object mapping key
}

// ApfsReapState represents file system reaper state (apfs_reap_state_t)
type ApfsReapState struct {
    LastPBN    uint64     // Last physical block number
    CurSnapXID types.XID  // Current snapshot XID
    Phase      uint32     // Current phase
}

// Reaper manages deletion of large objects
type Reaper struct {
    Container  *ContainerManager
    Phys       *NXReaperPhys
    Lists      map[types.OID]*NXReapListPhys
}

// Constants for reaper phases, flags, etc.
const (
    // Volume reaper states
    ApfsReapPhaseStart       = 0
    ApfsReapPhaseSnapshots   = 1
    ApfsReapPhaseActiveFS    = 2
    ApfsReapPhaseDestroyOmap = 3
    ApfsReapPhaseDone        = 4
    
    // Reaper flags
    NRBHMFlag   uint32 = 0x00000001
    NRContinue  uint32 = 0x00000002
    
    // Reaper list entry flags
    NRLEValid         uint32 = 0x00000001
    NRLEReapIDRecord  uint32 = 0x00000002
    NRLECall          uint32 = 0x00000004
    NRLECompletion    uint32 = 0x00000008
    NRLECleanup       uint32 = 0x00000010
    
    // Other constants
    NRLIndexInvalid   uint32 = 0xffffffff
)

// NewReaper creates a new reaper from the container
func NewReaper(container *ContainerManager) (*Reaper, error) {
    // Load reaper object
    // Initialize state
    // Return reaper
}

// LoadReapList loads a reap list by OID
func (r *Reaper) LoadReapList(oid types.OID) (*NXReapListPhys, error) {
    // Check if already loaded
    // Load reap list from disk
    // Parse reap list
    // Store in cache
    // Return list
}

// AddObject adds an object to be reaped
func (r *Reaper) AddObject(tx *Transaction, fsoid, oid types.OID, xid types.XID, objType uint32, size uint32) error {
    // Get tail list
    // Check if list is full, create new if needed
    // Create reap list entry
    // Add entry to list
    // Update reaper state
    // Save changes
    // Return success/error
}

// ProcessReapQueue processes pending reap operations
func (r *Reaper) ProcessReapQueue(tx *Transaction, maxEntries uint32) error {
    // Check if reaping is in progress
    // Process up to maxEntries entries
    // Return success/error
}

// ReapObject performs reaping of a single object
func (r *Reaper) ReapObject(tx *Transaction, entry *NXReapListEntry) error {
    // Set reaper state to indicate reaping in progress
    // Determine object type and perform appropriate reaping
    // Mark reaping as complete
    // Return success/error
}

// ParseReaperFromBytes parses a reaper from bytes
func ParseReaperFromBytes(data []byte) (*NXReaperPhys, error) {
    // Validate data length
    // Parse object header
    // Parse remaining fields
    // Parse state buffer if present
    // Return parsed reaper
}

// ParseReapListFromBytes parses a reap list from bytes
func ParseReapListFromBytes(data []byte) (*NXReapListPhys, error) {
    // Validate data length
    // Parse object header
    // Parse remaining fields
    // Parse entries if present
    // Return parsed reap list
}
```

### pkg/fs/

#### volume.go
```go
package fs

import (
    "apfs/pkg/types"
    "apfs/pkg/container"
)

// APFSSuperblock represents a volume superblock (apfs_superblock_t)
type APFSSuperblock struct {
    Object                ObjectPhys      // Object header
    Magic                 uint32          // Magic ('APSB')
    FSIndex               uint32          // Index in container's array
    Features              uint64          // Optional features
    ReadOnlyCompat        uint64          // Read-only compatible features
    IncompatFeatures      uint64          // Incompatible features
    UnmountTime           uint64          // Last unmount time (nanoseconds)
    ReserveBlockCount     uint64          // Reserved block count
    QuotaBlockCount       uint64          // Quota block count
    AllocCount            uint64          // Allocated block count
    MetaCrypto            MetaCryptoState // Metadata crypto state
    RootTreeType          uint32          // Root tree type
    ExtentrefTreeType     uint32          // Extent reference tree type
    SnapMetaTreeType      uint32          // Snapshot metadata tree type
    OMapOID               types.OID       // Object map OID
    RootTreeOID           types.OID       // Root tree OID
    ExtentrefTreeOID      types.OID       // Extent reference tree OID
    SnapMetaTreeOID       types.OID       // Snapshot metadata tree OID
    RevertToXID           types.XID       // Revert to transaction ID
    RevertToSblockOID     types.OID       // Revert to superblock OID
    NextObjID             uint64          // Next object ID
    NumFiles              uint64          // Number of files
    NumDirectories        uint64          // Number of directories
    NumSymlinks           uint64          // Number of symlinks
    NumOtherFSObjects     uint64          // Number of other objects
    NumSnapshots          uint64          // Number of snapshots
    TotalBlocksAlloced    uint64          // Total blocks allocated
    TotalBlocksFreed      uint64          // Total blocks freed
    UUID                  types.UUID      // Volume UUID
    LastModTime           uint64          // Last modification time
    FSFlags               uint64          // Flags
    FormattedBy           ModifiedBy      // Formatted by
    ModifiedBy            [8]ModifiedBy   // Modified by history
    VolName               [256]byte       // Volume name
    NextDocID             uint32          // Next document ID
    Role                  uint16          // Role (system, data, etc.)
    Reserved              uint16          // Reserved
    RootToXID             types.XID       // Root from transaction ID
    ERStateOID            types.OID       // Encryption rolling state OID
    CloneinfoIDEpoch      uint64          // Clone info ID epoch
    CloneinfoXID          uint64          // Clone info transaction ID
    SnapMetaExtOID        types.OID       // Extended snapshot metadata OID
    VolumeGroupID         types.UUID      // Volume group UUID
    IntegrityMetaOID      types.OID       // Integrity metadata OID
    FextTreeOID           types.OID       // File extent tree OID
    FextTreeType          uint32          // File extent tree type
    ReservedType          uint32          // Reserved
    ReservedOID           types.OID       // Reserved
}

// MetaCryptoState represents wrapped metadata crypto state
type MetaCryptoState struct {
    MajorVersion     uint16
    MinorVersion     uint16
    Flags            uint32
    PersistentClass  uint32
    KeyOSVersion     uint32
    KeyRevision      uint16
    Unused           uint16
}

// ModifiedBy represents information about software that modified the volume
type ModifiedBy struct {
    ID         [32]byte // Identifier 
    Timestamp  uint64   // Timestamp
    LastXID    types.XID // Last transaction ID
}

// VolumeManager manages a volume
type VolumeManager struct {
    Container      *container.ContainerManager
    Superblock     *APFSSuperblock
    ObjectMap      *container.OMapPhys
    RootTree       *container.BTree
    ExtentrefTree  *container.BTree
    SnapMetaTree   *container.BTree
    IsReadOnly     bool
    IsEncrypted    bool
}

// VolumeInfo contains summary information about a volume
type VolumeInfo struct {
    Index          uint32
    Name           string
    UUID           types.UUID
    Role           uint16
    NumFiles       uint64
    NumDirectories uint64
    Capacity       uint64
    Used           uint64
    Created        uint64
    Modified       uint64
    Encrypted      bool
    CaseSensitive  bool
}

// NewVolumeManager creates a new volume manager for a volume
func NewVolumeManager(container *container.ContainerManager, superblock *APFSSuperblock) (*VolumeManager, error) {
    // Load the volume's object map
    // Load and initialize root tree
    // Check encryption status
    // Return volume manager
}

// GetFile returns a file object by ID
func (vm *VolumeManager) GetFile(oid types.OID) (*Inode, error) {
    // Look up the inode record in the root tree
    // Parse the inode record
    // Return the file object
}

// GetFileByPath returns a file object for a path
func (vm *VolumeManager) GetFileByPath(path string) (*Inode, error) {
    // Start at root directory
    // Parse path components
    // Traverse directory structure
    // Return file object for the specified path
}

// ListDirectory lists the contents of a directory
func (vm *VolumeManager) ListDirectory(dirOID types.OID) ([]*DirectoryEntry, error) {
    // Verify the OID corresponds to a directory
    // Find directory entries in the root tree
    // Parse and return entries
}

// GetVolumeInfo returns summary information about the volume
func (vm *VolumeManager) GetVolumeInfo() (*VolumeInfo, error) {
    // Extract information from superblock
    // Calculate space usage
    // Return volume info
}

// CreateSnapshot creates a new snapshot of the volume
func (vm *VolumeManager) CreateSnapshot(name string) error {
    // Validate writable status
    // Create snapshot metadata
    // Update volume structures
    // Return success/error
}

// GetSnapshots lists all snapshots
func (vm *VolumeManager) GetSnapshots() ([]*SnapshotInfo, error) {
    // Query the snapshot metadata tree
    // Parse snapshot records
    // Return snapshot info
}
```

#### inode.go

package fs

import (
    "time"
    "apfs/pkg/types"
)

// Inode represents a file system object (j_inode_val_t)
type Inode struct {
    ParentID             uint64    // Parent directory ID
    PrivateID            uint64    // Private ID for extents
    CreateTime           uint64    // Creation time
    ModTime              uint64    // Modification time
    ChangeTime           uint64    // Change time (attributes)
    AccessTime           uint64    // Last access time
    InternalFlags        uint64    // Internal flags
    NChildren            int32     // Number of children (for directories)
    NLink                int32     // Number of hard links (for files)
    DefaultProtClass     uint32    // Default protection class
    WriteGenCounter      uint32    // Write generation counter
    BSDFlags             uint32    // BSD flags
    UID                  uint32    // Owner user ID
    GID                  uint32    // Owner group ID
    Mode                 uint16    // File mode
    Pad1                 uint16    // Padding
    UncompressedSize     uint64    // Uncompressed size
    ExtendedFields       []ExtendedField // Extended fields
    
    // Runtime fields
    ObjectID             types.OID // Object ID
    Name                 string    // Name (from directory entry)
    DataStream           *DataStream // Data stream (if present)
    XAttrs               map[string]*XAttr // Extended attributes
}

// ExtendedField represents an extended field
type ExtendedField struct {
    Type     uint8
    Flags    uint8
    Size     uint16
    Data     []byte
}

// FileInfo returns file information
func (i *Inode) FileInfo() *FileInfo {
    // Extract file information
    // Return file info
}

// IsDir returns true if this is a directory
func (i *Inode) IsDir() bool {
    return (i.Mode & types.S_IFMT) == types.S_IFDIR
}

// IsRegular returns true if this is a regular file
func (i *Inode) IsRegular() bool {
    return (i.Mode & types.S_IFMT) == types.S_IFREG
}

// IsSymlink returns true if this is a symbolic link
func (i *Inode) IsSymlink() bool {
    return (i.Mode & types.S_IFMT) == types.S_IFLNK
}

// IsEncrypted returns true if this file is encrypted
func (i *Inode) IsEncrypted() bool {
    // Check encryption state
    // Return true/false
}

// HasExtendedField returns true if the inode has an extended field of the specified type
func (i *Inode) HasExtendedField(fieldType uint8) bool {
    for _, field := range i.ExtendedFields {
        if field.Type == fieldType {
            return true
        }
    }
    return false
}

// GetExtendedField returns the extended field of the specified type
func (i *Inode) GetExtendedField(fieldType uint8) *ExtendedField {
    for _, field := range i.ExtendedFields {
        if field.Type == fieldType {
            return &field
        }
    }
    return nil
}

// ReadData reads data from the file
func (i *Inode) ReadData(offset int64, size int) ([]byte, error) {
    // Verify this is a regular file
    // Find data stream
    // Locate and read extents
    // Handle encryption if needed
    // Return data
}

// FileInfo contains information about a file
type FileInfo struct {
    Name        string
    Size        int64
    Mode        uint16
    ModTime     time.Time
    IsDir       bool
    Sys         *Inode
}

#### dentry.go

package fs

import (
    "apfs/pkg/types"
)

// DirectoryEntry represents a directory entry (j_drec_val_t)
type DirectoryEntry struct {
    FileID        uint64         // File ID (inode number)
    DateAdded     uint64         // Date added
    Flags         uint16         // Flags
    ExtendedFields []ExtendedField // Extended fields
    
    // Runtime fields
    Name          string         // Entry name
    ParentID      types.OID      // Parent directory ID
    Type          uint8          // File type
}

// GetFileType returns the file type
func (d *DirectoryEntry) GetFileType() uint8 {
    return d.Flags & types.DrecTypeMask
}

// IsDirectory returns true if entry is a directory
func (d *DirectoryEntry) IsDirectory() bool {
    return d.GetFileType() == types.DT_DIR
}

// IsRegularFile returns true if entry is a regular file
func (d *DirectoryEntry) IsRegularFile() bool {
    return d.GetFileType() == types.DT_REG
}

// IsSymlink returns true if entry is a symbolic link
func (d *DirectoryEntry) IsSymlink() bool {
    return d.GetFileType() == types.DT_LNK
}

// DirectoryStats represents directory statistics (j_dir_stats_val_t)
type DirectoryStats struct {
    NumChildren    uint64    // Number of children
    TotalSize      uint64    // Total size of all files
    ChainedKey     uint64    // Parent directory object ID
    GenCount       uint64    // Generation count
}

// HashedName represents a directory entry name with precomputed hash
type HashedName struct {
    NameLenAndHash uint32   // Name length and hash
    Name           string   // Directory entry name
}

// ComputeNameHash computes the hash for a directory entry name
func ComputeNameHash(name string) uint32 {
    // Convert to UTF-8 if needed
    // Normalize using canonical decomposition (NFD)
    // Convert to UTF-32
    // Compute CRC-32C hash
    // Complement bits
    // Keep low 22 bits
    // Return hash
}

// LookupDirectoryEntry looks up a directory entry by name
func LookupDirectoryEntry(volume *VolumeManager, dirOID types.OID, name string) (*DirectoryEntry, error) {
    // Compute name hash
    // Search for directory entry in the root tree
    // Parse and return the entry
}

// ListDirectoryEntries lists all entries in a directory
func ListDirectoryEntries(volume *VolumeManager, dirOID types.OID) ([]*DirectoryEntry, error) {
    // Find directory entries in the root tree
    // Parse and return entries
}

#### xattr.go

package fs

import (
    "apfs/pkg/types"
)

// XAttr represents an extended attribute (j_xattr_val_t)
type XAttr struct {
    Flags    uint16    // Flags
    DataLen  uint16    // Data length
    Data     []byte    // Attribute data or data stream ID
    
    // Runtime fields
    Name     string    // Attribute name
    ObjectID types.OID // Object ID
}

// IsEmbedded returns true if the attribute data is embedded
func (x *XAttr) IsEmbedded() bool {
    return x.Flags&types.XattrDataEmbedded != 0
}

// IsDataStream returns true if the attribute data is stored in a data stream
func (x *XAttr) IsDataStream() bool {
    return x.Flags&types.XattrDataStream != 0
}

// IsFileSystemOwned returns true if the attribute is owned by the file system
func (x *XAttr) IsFileSystemOwned() bool {
    return x.Flags&types.XattrFileSystemOwned != 0
}

// GetXattr gets an extended attribute by name
func GetXattr(volume *VolumeManager, inodeID types.OID, name string) (*XAttr, error) {
    // Search for the attribute in the file system tree
    // Parse attribute record
    // Read attribute data (embedded or data stream)
    // Return attribute
}

// ListXattrs lists all extended attributes for a file
func ListXattrs(volume *VolumeManager, inodeID types.OID) ([]*XAttr, error) {
    // Find all extended attribute records for the inode
    // Parse records
    // Return attributes list
}

// XAttrData reads the data for an extended attribute
func XAttrData(volume *VolumeManager, xattr *XAttr) ([]byte, error) {
    // If embedded, return embedded data
    // If data stream, read from data stream
    // Return attribute data
}

#### datastream.go

package fs

import (
    "apfs/pkg/types"
)

// DataStream represents information about a data stream (j_dstream_t)
type DataStream struct {
    Size                uint64    // Logical size in bytes
    AllocedSize         uint64    // Allocated size
    DefaultCryptoID     uint64    // Default crypto ID
    TotalBytesWritten   uint64    // Total bytes written
    TotalBytesRead      uint64    // Total bytes read
}

// XAttrDataStream represents a data stream for extended attributes
type XAttrDataStream struct {
    XAttrObjectID    uint64       // Extended attribute object ID
    Stream           DataStream   // Data stream
}

// DataStreamID represents a data stream identifier record
type DataStreamID struct {
    RefCount    uint32    // Reference count
}

// NewDataStream creates a new data stream
func NewDataStream() *DataStream {
    return &DataStream{}
}

// ReadDataStream reads data from a data stream
func ReadDataStream(volume *VolumeManager, inodeID types.OID, offset int64, size int) ([]byte, error) {
    // Find data stream info for the inode
    // Find file extents that cover the requested range
    // Read data from each extent
    // Decrypt if necessary
    // Return data
}

// GetDataStreamInfo gets data stream information for an inode
func GetDataStreamInfo(volume *VolumeManager, inodeID types.OID) (*DataStream, error) {
    // Search for data stream extended field in the inode
    // If not found, search for data stream ID record
    // Return data stream info
}

#### extents.go

package fs

import (
    "apfs/pkg/types"
    "apfs/pkg/container"
)

// FileExtent represents a file extent (j_file_extent_val_t)
type FileExtent struct {
    LenAndFlags      uint64       // Length and flags
    PhysBlockNum     uint64       // Physical block number
    CryptoID         uint64       // Crypto ID
}

// FileExtentKey represents a file extent key (j_file_extent_key_t)
type FileExtentKey struct {
    Header        types.JKey   // Key header
    LogicalAddr   uint64       // Logical address
}

// PhysicalExtent represents a physical extent (j_phys_ext_val_t)
type PhysicalExtent struct {
    LenAndKind     uint64    // Length and kind
    OwningObjID    uint64    // Owning object ID
    RefCount       int32     // Reference count
}

// GetExtentLength returns the length in bytes of the file extent
func (e *FileExtent) GetExtentLength() uint64 {
    return e.LenAndFlags & types.JFileExtentLenMask
}

// GetExtentFlags returns the flags for the file extent
func (e *FileExtent) GetExtentFlags() uint8 {
    return uint8((e.LenAndFlags & types.JFileExtentFlagMask) >> types.JFileExtentFlagShift)
}

// PhysicalLenAndKind represents length and kind for physical extents
type PhysicalLenAndKind struct {
    Length    uint64       // Length in blocks
    Kind      types.JObjKind // Kind
}

// NewPhysicalLenAndKind creates a new length and kind value
func NewPhysicalLenAndKind(length uint64, kind types.JObjKind) uint64 {
    return length | (uint64(kind) << types.PextKindShift)
}

// GetPhysicalExtentLength returns the length in blocks
func GetPhysicalExtentLength(lenAndKind uint64) uint64 {
    return lenAndKind & types.PextLenMask
}

// GetPhysicalExtentKind returns the kind
func GetPhysicalExtentKind(lenAndKind uint64) types.JObjKind {
    return types.JObjKind((lenAndKind & types.PextKindMask) >> types.PextKindShift)
}

// GetFileExtents gets all extents for a file
func GetFileExtents(volume *VolumeManager, inodeID types.OID) ([]*FileExtent, error) {
    // Search for file extent records in the file system tree
    // Parse records
    // Return extents list
}

// ReadExtent reads data from a file extent
func ReadExtent(volume *VolumeManager, extent *FileExtent, offset int64, size int) ([]byte, error) {
    // Calculate physical location
    // Read data from disk
    // Decrypt if necessary
    // Return data
}

#### extfields.go

package fs

import (
    "encoding/binary"
    "apfs/pkg/types"
)

// ExtendedFieldBlob represents a collection of extended fields (xf_blob_t)
type ExtendedFieldBlob struct {
    NumExts     uint16
    UsedData    uint16
    Data        []byte
}

// DecodeExtendedFields decodes extended fields from a blob
func DecodeExtendedFields(data []byte) ([]ExtendedField, error) {
    // Verify data length
    // Parse number of fields and used data size
    // Read field descriptors
    // Read field data
    // Return extended fields
}

// EncodeExtendedFields encodes extended fields to bytes
func EncodeExtendedFields(fields []ExtendedField) ([]byte, error) {
    // Calculate total size needed
    // Create buffer
    // Write number of fields and used data size
    // Write field descriptors
    // Write field data
    // Return encoded data
}

// GetDocumentID gets the document ID extended field
func GetDocumentID(inode *Inode) (uint32, bool) {
    field := inode.GetExtendedField(types.InoExtTypeDocumentID)
    if field == nil {
        return 0, false
    }
    
    // Parse document ID from field data
    if len(field.Data) < 4 {
        return 0, false
    }
    
    return binary.LittleEndian.Uint32(field.Data), true
}

// GetDataStream gets the data stream extended field
func GetDataStream(inode *Inode) (*DataStream, bool) {
    field := inode.GetExtendedField(types.InoExtTypeDstream)
    if field == nil {
        return nil, false
    }
    
    // Parse data stream from field data
    if len(field.Data) < binary.Size(DataStream{}) {
        return nil, false
    }
    
    // Unmarshal data stream
    // Return data stream
    return nil, false // Placeholder
}

// GetFinderInfo gets the Finder info extended field
func GetFinderInfo(inode *Inode) ([]byte, bool) {
    field := inode.GetExtendedField(types.InoExtTypeFinderInfo)
    if field == nil {
        return nil, false
    }
    
    // Copy and return finder info data
    return field.Data, true
}

// AddExtendedField adds an extended field to an inode
func AddExtendedField(inode *Inode, fieldType uint8, flags uint8, data []byte) {
    // Create new extended field
    field := ExtendedField{
        Type:  fieldType,
        Flags: flags,
        Size:  uint16(len(data)),
        Data:  data,
    }
    
    // Replace existing field if present
    for i, f := range inode.ExtendedFields {
        if f.Type == fieldType {
            inode.ExtendedFields[i] = field
            return
        }
    }
    
    // Add new field
    inode.ExtendedFields = append(inode.ExtendedFields, field)
}

#### siblings.go

package fs

import (
    "apfs/pkg/types"
)

// SiblingLink represents a sibling link record (j_sibling_val_t)
type SiblingLink struct {
    ParentID    uint64    // Parent directory ID
    NameLen     uint16    // Name length
    Name        string    // Name
}

// SiblingLinkKey represents a sibling link key (j_sibling_key_t)
type SiblingLinkKey struct {
    Header      types.JKey  // Key header
    SiblingID   uint64      // Sibling ID
}

// SiblingMap represents a sibling map record (j_sibling_map_val_t)
type SiblingMap struct {
    FileID      uint64    // File ID (inode number)
}

// SiblingMapKey represents a sibling map key (j_sibling_map_key_t)
type SiblingMapKey struct {
    Header      types.JKey  // Key header
}

// GetSiblingID gets the sibling ID for a hard link
func GetSiblingID(volume *VolumeManager, dirEntryID types.OID) (uint64, error) {
    // Check if the directory entry has a sibling ID extended field
    // If found, return the sibling ID
    // Otherwise, return error
}

// GetTargetInode gets the target inode for a hard link
func GetTargetInode(volume *VolumeManager, siblingID uint64) (types.OID, error) {
    // Look up the sibling ID in the sibling map
    // Return the target inode ID
}

// GetSiblingLinks gets all siblings for an inode
func GetSiblingLinks(volume *VolumeManager, inodeID types.OID) ([]*SiblingLink, error) {
    // Search for sibling link records for the inode
    // Parse records
    // Return links
}

// GetPrimaryLink gets the primary link (original name) for an inode
func GetPrimaryLink(volume *VolumeManager, inodeID types.OID) (*SiblingLink, error) {
    // Get all sibling links
    // Find the primary link (with lowest sibling ID)
    // Return the primary link
}

### pkg/crypto/
#### keybag.go

package crypto

import (
    "apfs/pkg/types"
)

// KBLocker represents a keybag (kb_locker_t)
type KBLocker struct {
    Version     uint16          // Version
    NumKeys     uint16          // Number of keys
    NumBytes    uint32          // Number of bytes
    Padding     [8]byte         // Padding
    Entries     []KeybagEntry   // Entries
}

// KeybagEntry represents an entry in a keybag (keybag_entry_t)
type KeybagEntry struct {
    UUID        types.UUID   // UUID
    Tag         uint16       // Tag
    KeyLen      uint16       // Key data length
    Padding     [4]byte      // Padding
    KeyData     []byte       // Key data
}

// UnwrapKeybag unwraps a keybag using its UUID
func UnwrapKeybag(data []byte, uuid types.UUID) (*KBLocker, error) {
    // Verify keybag version
    // Parse keybag structure
    // Unwrap the keybag using the UUID
    // Return unwrapped keybag
}

// FindKeyByUUID finds a keybag entry with the specified UUID and tag
func FindKeyByUUID(keybag *KBLocker, uuid types.UUID, tag uint16) *KeybagEntry {
    for _, entry := range keybag.Entries {
        if entry.Tag == tag && types.UUIDEqual(entry.UUID, uuid) {
            return &entry
        }
    }
    return nil
}

// FindVolumeKey finds the volume encryption key for a volume
func FindVolumeKey(containerKeybag *KBLocker, volumeUUID types.UUID) ([]byte, error) {
    // Find volume key entry in container keybag
    // Return the key data (wrapped VEK)
}

// FindVolumeKeybag finds the volume keybag location for a volume
func FindVolumeKeybag(containerKeybag *KBLocker, volumeUUID types.UUID) (types.PRange, error) {
    // Find volume unlock records entry in container keybag
    // Parse the data as a PRange
    // Return the volume keybag location
}

// FindPassphraseKey finds the KEK for a user's password
func FindPassphraseKey(volumeKeybag *KBLocker, userUUID types.UUID) ([]byte, error) {
    // Find the volume unlock records entry for the user
    // Return the key data (wrapped KEK)
}

// ParseMediaKeybag parses a media keybag object
func ParseMediaKeybag(data []byte) (*KBLocker, error) {
    // Parse media keybag header
    // Extract keybag data
    // Return parsed keybag
}

#### keys.go

package crypto

import (
    "apfs/pkg/types"
)

// VolumeEncryptionKey represents a volume encryption key
type VolumeEncryptionKey struct {
    Key       []byte    // The key bytes
    UUID      types.UUID // The volume UUID
}

// FileEncryptionKey represents a per-file encryption key
type FileEncryptionKey struct {
    Key        []byte    // The key bytes
    FileID     uint64    // The file ID
    VolumeUUID types.UUID // The volume UUID
}

// UnwrapVEK unwraps a volume encryption key using a key encryption key
func UnwrapVEK(wrappedVEK, kek []byte) ([]byte, error) {
    // Implement RFC 3394 AES key unwrapping
    // Return unwrapped VEK
}

// UnwrapKEK unwraps a key encryption key using a user password
func UnwrapKEK(wrappedKEK []byte, password string) ([]byte, error) {
    // Derive key from password
    // Unwrap KEK using derived key
    // Return unwrapped KEK
}

// UnwrapWithPassword unwraps a volume encryption key using a user password
func UnwrapWithPassword(wrappedVEK, wrappedKEK []byte, password string) ([]byte, error) {
    // Unwrap KEK with password
    // Unwrap VEK with KEK
    // Return VEK
}

// UnwrapWithRecoveryKey unwraps a volume encryption key using a recovery key
func UnwrapWithRecoveryKey(wrappedVEK, wrappedRecoveryKEK []byte, recoveryKey string) ([]byte, error) {
    // Parse recovery key
    // Unwrap recovery KEK
    // Unwrap VEK with recovery KEK
    // Return VEK
}

// GetVolumeKey gets the volume encryption key for a volume
func GetVolumeKey(containerKeybag, volumeKeybag *KBLocker, volumeUUID types.UUID, password string) ([]byte, error) {
    // Find wrapped VEK in container keybag
    // Find wrapped KEK in volume keybag
    // Unwrap KEK using password
    // Unwrap VEK using KEK
    // Return VEK
}

#### crypto.go
```go
package crypto

import (
    "crypto/aes"
    "crypto/cipher"
    "apfs/pkg/types"
)

// DecryptData decrypts data using the specified key and tweak
func DecryptData(data, key []byte, tweak uint64) ([]byte, error) {
    // Create AES-XTS cipher
    // Set up tweak
    // Decrypt data
    // Return decrypted data
}

// EncryptData encrypts data using the specified key and tweak
func EncryptData(data, key []byte, tweak uint64) ([]byte, error) {
    // Create AES-XTS cipher
    // Set up tweak
    // Encrypt data
    // Return encrypted data
}

// DecryptBlock decrypts a block using the specified key and tweak
func DecryptBlock(blockData []byte, blockAddr uint64, key []byte) ([]byte, error) {
    // Use block address as tweak
    // Decrypt block data
    // Return decrypted block
}

// DecryptFileBlock decrypts a file block using the file's crypto ID
func DecryptFileBlock(blockData []byte, fileExtent *FileExtent, key []byte) ([]byte, error) {
    // Use crypto ID as tweak
    // Decrypt block data
    // Return decrypted block
}

// CryptoState represents crypto state for a file
type CryptoState struct {
    RefCount     uint32              // Reference count
    State        WrappedCryptoState  // Crypto state
}

// WrappedCryptoState represents wrapped crypto state for a file
type WrappedCryptoState struct {
    MajorVersion    uint16    // Major version
    MinorVersion    uint16    // Minor version
    Flags           uint32    // Flags
    PersistentClass uint32    // Protection class
    KeyOSVersion    uint32    // Key OS version
    KeyRevision     uint16    // Key revision
    KeyLen          uint16    // Key length
    PersistentKey   []byte    // Persistent key data
}

// MetaCryptoState represents wrapped metadata crypto state
type MetaCryptoState struct {
    MajorVersion    uint16    // Major version
    MinorVersion    uint16    // Minor version
    Flags           uint32    // Flags
    PersistentClass uint32    // Protection class
    KeyOSVersion    uint32    // Key OS version
    KeyRevision     uint16    // Key revision
    Unused          uint16    // Unused
}

// GetCryptoState gets the crypto state for a file
func GetCryptoState(volume *VolumeManager, cryptoID uint64) (*CryptoState, error) {
    // Look up crypto state record
    // Parse and return crypto state
}

// IsSoftwareEncryption checks if the container uses software encryption
func IsSoftwareEncryption(container *ContainerManager) bool {
    return container.Superblock.Flags&types.NXCryptoSW != 0
}

// IsFileEncrypted checks if a file is encrypted
func IsFileEncrypted(inode *Inode, extent *FileExtent) bool {
    return extent.CryptoID != 0 && extent.CryptoID != types.CryptoSwID
}
```

### pkg/snapshot/
#### snapshot.go

```go
package snapshot

import (
    "apfs/pkg/types"
)

// SnapshotMetadata represents snapshot metadata (j_snap_metadata_val_t)
type SnapshotMetadata struct {
    ExtentrefTreeOID    types.OID    // Extent reference tree OID
    SblockOID           types.OID    // Superblock OID
    CreateTime          uint64       // Creation time
    ChangeTime          uint64       // Change time
    Inum                uint64       // Inode number
    ExtentrefTreeType   uint32       // Extent reference tree type
    Flags               uint32       // Flags
    NameLen             uint16       // Name length
    Name                string       // Snapshot name
}

// SnapshotName represents a snapshot name record (j_snap_name_val_t)
type SnapshotName struct {
    SnapXID    types.XID    // Snapshot transaction ID
}

// SnapshotInfo contains information about a snapshot
type SnapshotInfo struct {
    Name       string       // Snapshot name
    XID        types.XID    // Transaction ID
    CreateTime uint64       // Creation time
    ChangeTime uint64       // Change time
    UUID       types.UUID   // Snapshot UUID (if available)
}

// IsPendingDataless returns true if the snapshot is pending conversion to dataless
func (s *SnapshotMetadata) IsPendingDataless() bool {
    return s.Flags&types.SnapMetaPendingDataless != 0
}

// IsMergeInProgress returns true if a snapshot merge is in progress
func (s *SnapshotMetadata) IsMergeInProgress() bool {
    return s.Flags&types.SnapMetaMergeInProgress != 0
}
```

#### operations.go

```go
package snapshot

import (
    "apfs/pkg/types"
    "apfs/pkg/fs"
    "apfs/pkg/container"
)

// CreateSnapshot creates a new snapshot
func CreateSnapshot(volume *fs.VolumeManager, name string, tx *container.Transaction) error {
    // Validate name
    // Create a new extent reference tree
    // Create snapshot metadata record
    // Create snapshot name record
    // Update volume structures
    // Return success/error
}

// DeleteSnapshot deletes a snapshot
func DeleteSnapshot(volume *fs.VolumeManager, xid types.XID, tx *container.Transaction) error {
    // Find snapshot metadata
    // Delete snapshot records
    // Mark object map entries for deletion
    // Return success/error
}

// GetSnapshots gets all snapshots for a volume
func GetSnapshots(volume *fs.VolumeManager) ([]*SnapshotInfo, error) {
    // Query snapshot metadata tree
    // Parse and collect snapshot information
    // Sort by creation time
    // Return snapshots
}

// RevertToSnapshot reverts a volume to a snapshot
func RevertToSnapshot(volume *fs.VolumeManager, xid types.XID) error {
    // Validate snapshot exists
    // Set revert transaction ID in volume superblock
    // Return success/error
}

// MountSnapshot mounts a volume at a specific snapshot
func MountSnapshot(container *container.ContainerManager, volumeIndex uint32, xid types.XID) (*fs.VolumeManager, error) {
    // Get volume superblock
    // Set root-to transaction ID
    // Create volume manager
    // Return mounted snapshot
}
```

### pkg/fusion/
#### fusion.go
```go
package fusion

import (
    "apfs/pkg/types"
    "apfs/pkg/container"
)

// FusionWBC represents a Fusion write-back cache state
type FusionWBC struct {
    Object           container.ObjectPhys  // Object header
    Version          uint64                // Version
    ListHeadOID      types.OID             // List head OID
    ListTailOID      types.OID             // List tail OID
    StableHeadOffset uint64                // Stable head offset
    StableTailOffset uint64                // Stable tail offset
    ListBlocksCount  uint32                // List blocks count
    Reserved         uint32                // Reserved
    UsedByRC         uint64                // Used by RC
    RCStash          types.PRange          // RC stash
}

// FusionWBCListEntry represents a Fusion write-back cache list entry
type FusionWBCListEntry struct {
    WbcLba      types.PAddr    // Write-back cache logical block address
    TargetLba   types.PAddr    // Target logical block address
    Length      uint64         // Length
}

// FusionWBCList represents a Fusion write-back cache list
type FusionWBCList struct {
    Object      container.ObjectPhys    // Object header
    Version     uint64                  // Version
    TailOffset  uint64                  // Tail offset
    IndexBegin  uint32                  // Index begin
    IndexEnd    uint32                  // Index end
    IndexMax    uint32                  // Index max
    Reserved    uint32                  // Reserved
    Entries     []FusionWBCListEntry    // List entries
}

// IsFusionContainer checks if a container is a Fusion container
func IsFusionContainer(container *container.ContainerManager) bool {
    return container.Superblock.IncompatFeatures&types.NXIncompatFusion != 0
}

// IsMainDevice checks if this is the Fusion main device
func IsMainDevice(uuid types.UUID) bool {
    // Check high bit of UUID
    return (uuid[0] & 0x80) != 0
}

// GetTier2Address converts a logical address to a tier2 address
func GetTier2Address(addr types.PAddr, blockSize uint32) types.PAddr {
    tier2Marker := types.FusionTier2DeviceByteAddr >> uint64(blockSize)
    return types.PAddr(tier2Marker | uint64(addr))
}
```
#### tier.go
```go
package fusion

import (
    "apfs/pkg/types"
    "apfs/pkg/container"
)

// FusionMTKey represents a Fusion middle tree key
type FusionMTKey types.PAddr

// FusionMTVal represents a Fusion middle tree value
type FusionMTVal struct {
    Lba       types.PAddr    // Logical block address
    Length    uint32         // Length
    Flags     uint32         // Flags
}

// IsDirty checks if the entry is dirty
func (v *FusionMTVal) IsDirty() bool {
    return v.Flags&types.FusionMTDirty != 0
}

// IsTenant checks if the entry is a tenant
func (v *FusionMTVal) IsTenant() bool {
    return v.Flags&types.FusionMTTenant != 0
}

// FusionDevice represents a Fusion device manager
type FusionDevice struct {
    Container     *container.ContainerManager
    IsMain        bool
    MiddleTree    *container.BTree
    WBCache       *FusionWBC
    WBCLists      map[types.OID]*FusionWBCList
}

// NewFusionDevice creates a new Fusion device manager
func NewFusionDevice(container *container.ContainerManager) (*FusionDevice, error) {
    // Check if this is a Fusion container
    // Determine if this is the main device
    // Load middle tree
    // Load write-back cache if main device
    // Return Fusion device manager
}

// ReadBlock reads a block from the Fusion device
func (fd *FusionDevice) ReadBlock(addr types.PAddr) ([]byte, error) {
    // Check if block is in cache
    // If not, read from disk
    // Return block data
}

// WriteBlock writes a block to the Fusion device
func (fd *FusionDevice) WriteBlock(addr types.PAddr, data []byte) error {
    // If main device, write directly
    // If tier2 device, check write policy
    // Either write to cache or to disk
    // Update middle tree if needed
    // Return success/error
}

// FlushCache flushes the write-back cache
func (fd *FusionDevice) FlushCache() error {
    // For each entry in cache
    // Write to tier2 device
    // Update cache state
    // Return success/error
}

```
### pkg/transaction/
#### transaction.go
```go
package transaction

import (
    "apfs/pkg/types"
    "apfs/pkg/container"
)

// TransactionType represents the type of transaction
type TransactionType int

const (
    TxTypeNormal TransactionType = iota    // Normal transaction
    TxTypeCheckpoint                       // Checkpoint transaction
    TxTypeFlush                            // Flush transaction
    TxTypeRevert                           // Revert transaction
)

// Transaction represents an APFS transaction
type Transaction struct {
    Container        *container.ContainerManager
    XID              types.XID
    Type             TransactionType
    VirtualObjects   map[types.OID][]byte
    PhysicalObjects  map[types.PAddr][]byte
    ObjectLocations  map[types.OID]types.PAddr
    EphemeralObjects map[types.OID][]byte
    ObjectsToDelete  []ObjectToDelete
    Checksum         func([]byte) [8]byte
}

// ObjectToDelete represents an object being deleted
type ObjectToDelete struct {
    OID   types.OID
    XID   types.XID
    Type  uint32
}

// NewTransaction creates a new transaction
func NewTransaction(container *container.ContainerManager, txType TransactionType) (*Transaction, error) {
    // Initialize transaction state
    // Assign next transaction ID
    // Set up checksum function
    // Return transaction
}

// Commit commits the transaction
func (tx *Transaction) Commit() error {
    // Allocate space for new objects
    // Calculate checksums
    // Write objects to disk
    // Update object maps
    // Create checkpoint if needed
    // Update container superblock
    // Return success/error
}

// Abort aborts the transaction
func (tx *Transaction) Abort() error {
    // Clear transaction state
    // Release any allocated blocks
    // Return success/error
}

// CreateObject creates a new object in the transaction
func (tx *Transaction) CreateObject(objType, objSubtype uint32, size uint32) (types.OID, []byte, error) {
    // Allocate object ID
    // Initialize object header
    // Allocate memory for object
    // Add to transaction
    // Return object ID and data buffer
}

// UpdateObject updates an existing object in the transaction
func (tx *Transaction) UpdateObject(oid types.OID, data []byte) error {
    // Locate existing object
    // Copy new data
    // Update transaction state
    // Return success/error
}

// DeleteObject marks an object for deletion
func (tx *Transaction) DeleteObject(oid types.OID) error {
    // Mark object for deletion
    // Add to reaper if needed
    // Update transaction state
    // Return success/error
}
```

#### operations.go
```go
package transaction

import (
    "apfs/pkg/types"
    "apfs/pkg/container"
)

// OperationType represents the type of operation
type OperationType int

const (
    OpTypeRead OperationType = iota    // Read operation
    OpTypeWrite                        // Write operation
    OpTypeDelete                       // Delete operation
    OpTypeAllocate                     // Allocate operation
    OpTypeFree                         // Free operation
)

// Operation represents a transaction operation
type Operation interface {
    Type() OperationType
    Execute(tx *Transaction) error
    Rollback(tx *Transaction) error
}

// ReadOperation represents a read operation
type ReadOperation struct {
    OID      types.OID
    XID      types.XID
    Buffer   *[]byte
}

// WriteOperation represents a write operation
type WriteOperation struct {
    OID      types.OID
    Data     []byte
    OldData  []byte
}

// DeleteOperation represents a delete operation
type DeleteOperation struct {
    OID      types.OID
    OldData  []byte
}

// AllocateOperation represents an allocate operation
type AllocateOperation struct {
    Size     uint32
    Address  *types.PAddr
}

// FreeOperation represents a free operation
type FreeOperation struct {
    Address  types.PAddr
    Size     uint32
}

// ExecuteTransaction executes operations in a transaction
func ExecuteTransaction(container *container.ContainerManager, operations []Operation) error {
    // Create transaction
    // For each operation
    //   Execute operation
    // If any operation fails
    //   Rollback all executed operations
    //   Abort transaction
    //   Return error
    // Commit transaction
    // Return success/error
}

// CreateCheckpoint creates a new checkpoint
func CreateCheckpoint(container *container.ContainerManager) error {
    // Create checkpoint transaction
    // Save ephemeral objects to checkpoint data area
    // Update checkpoint descriptor area
    // Commit transaction
    // Return success/error
}
```
### pkg/util/
#### io.go
```go
package util

import (
    "os"
    "io"
    "apfs/pkg/types"
)

// BlockDevice implements types.BlockDevice interface for a file
type BlockDevice struct {
    File         *os.File
    BlockSize    uint32
    BlockCount   uint64
    IsReadOnly   bool
}

// NewBlockDevice creates a new block device
func NewBlockDevice(path string, readOnly bool) (*BlockDevice, error) {
    // Open file
    // Get file size
    // Detect block size
    // Calculate block count
    // Return block device
}

// ReadBlock reads a block from the device
func (d *BlockDevice) ReadBlock(addr types.PAddr) ([]byte, error) {
    // Validate address
    // Seek to block position
    // Read block data
    // Return data
}

// WriteBlock writes a block to the device
func (d *BlockDevice) WriteBlock(addr types.PAddr, data []byte) error {
    // Validate address
    // Check read-only status
    // Seek to block position
    // Write block data
    // Return success/error
}

// GetBlockSize returns the block size
func (d *BlockDevice) GetBlockSize() uint32 {
    return d.BlockSize
}

// GetBlockCount returns the block count
func (d *BlockDevice) GetBlockCount() uint64 {
    return d.BlockCount
}

// Close closes the block device
func (d *BlockDevice) Close() error {
    return d.File.Close()
}

// RawBlockReader implements low-level block reading
type RawBlockReader struct {
    File          *os.File
    BlockSize     uint32
    CurrentOffset int64
}

// NewRawBlockReader creates a new raw block reader
func NewRawBlockReader(path string, blockSize uint32) (*RawBlockReader, error) {
    // Open file
    // Return raw block reader
}

// ReadBlocks reads multiple blocks
func (r *RawBlockReader) ReadBlocks(addr types.PAddr, count uint32) ([]byte, error) {
    // Calculate offset
    // Seek to position
    // Read blocks
    // Return data
}

// Seek seeks to a position
func (r *RawBlockReader) Seek(addr types.PAddr) error {
    // Calculate offset
    // Seek to position
    // Return success/error
}

// Close closes the raw block reader
func (r *RawBlockReader) Close() error {
    return r.File.Close()
}
```
#### checksum.go

```go
package util

// Fletcher64 computes a Fletcher-64 checksum
func Fletcher64(data []byte) [8]byte {
    // Zero out checksum if present in data
    // Implement Fletcher-64 algorithm
    // Return 8-byte checksum
}

// VerifyChecksum verifies a Fletcher-64 checksum
func VerifyChecksum(data []byte, checksum [8]byte) bool {
    // Make a copy of data
    // Zero out checksum fields in copy
    // Compute checksum of copy
    // Compare with provided checksum
    // Return true if match, false otherwise
}

// CRC32C computes a CRC-32C checksum
func CRC32C(data []byte) uint32 {
    // Implement CRC-32C algorithm
    // Return 4-byte checksum
}

// ComputeNameHash computes a directory entry name hash
func ComputeNameHash(name string) uint32 {
    // Normalize name using NFD
    // Convert to UTF-32
    // Compute CRC-32C hash
    // Complement bits
    // Mask to 22 bits
    // Return hash
}
```

#### uuid.go
```go
package util

import (
    "encoding/binary"
    "apfs/pkg/types"
    "strings"
)

// ParseUUID parses a UUID string into a UUID
func ParseUUID(s string) (types.UUID, error) {
    // Parse UUID string
    // Return UUID
}

// UUIDToString converts a UUID to a string
func UUIDToString(uuid types.UUID) string {
    // Format UUID as string
    // Return string
}

// UUIDEqual compares two UUIDs for equality
func UUIDEqual(a, b types.UUID) bool {
    for i := 0; i < 16; i++ {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}

// GenerateUUID generates a new UUID
func GenerateUUID() types.UUID {
    // Generate random UUID
    // Return UUID
}

// IsZeroUUID checks if a UUID is zero
func IsZeroUUID(uuid types.UUID) bool {
    for i := 0; i < 16; i++ {
        if uuid[i] != 0 {
            return false
        }
    }
    return true
}
```

### cmd/apfs-info/
#### main.go
```go
package main

import (
    "flag"
    "fmt"
    "os"
    "text/tabwriter"
    "time"
    
    "apfs/pkg/container"
    "apfs/pkg/fs"
    "apfs/pkg/types"
    "apfs/pkg/util"
)

func main() {
    // Parse command line flags
    devicePath := flag.String("device", "", "Path to APFS device")
    flag.Parse()
    
    if *devicePath == "" {
        fmt.Println("Error: Device path is required")
        flag.Usage()
        os.Exit(1)
    }
    
    // Open block device
    device, err := util.NewBlockDevice(*devicePath, true)
    if err != nil {
        fmt.Printf("Error opening device: %v\n", err)
        os.Exit(1)
    }
    defer device.Close()
    
    // Create container manager
    containerMgr, err := container.NewContainerManager(device)
    if err != nil {
        fmt.Printf("Error parsing APFS container: %v\n", err)
        os.Exit(1)
    }
    
    // Display container information
    displayContainerInfo(containerMgr)
    
    // List volumes
    volumes, err := containerMgr.ListVolumes()
    if err != nil {
        fmt.Printf("Error listing volumes: %v\n", err)
        os.Exit(1)
    }
    
    // Display volume information
    displayVolumesInfo(volumes)
}

func displayContainerInfo(containerMgr *container.ContainerManager) {
    sb := containerMgr.Superblock
    
    fmt.Println("APFS Container Information:")
    fmt.Println("===========================")
    fmt.Printf("UUID:            %s\n", util.UUIDToString(sb.UUID))
    fmt.Printf("Block Size:      %d bytes\n", sb.BlockSize)
    fmt.Printf("Block Count:     %d blocks\n", sb.BlockCount)
    fmt.Printf("Capacity:        %s\n", formatSize(uint64(sb.BlockSize)*sb.BlockCount))
    fmt.Printf("Free Space:      %s\n", formatSize(uint64(containerMgr.SpaceManager.GetFreeSpace())))
    fmt.Printf("APFS Version:    %s\n", getAPFSVersion(sb))
    fmt.Printf("Is Fusion:       %v\n", containerMgr.Superblock.IncompatFeatures&types.NXIncompatFusion != 0)
    fmt.Println()
}

func displayVolumesInfo(volumes []fs.VolumeInfo) {
    fmt.Println("APFS Volumes:")
    fmt.Println("=============")
    
    w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
    fmt.Fprintf(w, "Index\tName\tRole\tCapacity\tUsed\tEncrypted\tCase-sensitive\tCreated\n")
    fmt.Fprintf(w, "-----\t----\t----\t--------\t----\t---------\t--------------\t-------\n")
    
    for _, vol := range volumes {
        role := formatVolumeRole(vol.Role)
        created := time.Unix(0, int64(vol.Created))
        
        fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%v\t%v\t%s\n",
            vol.Index,
            vol.Name,
            role,
            formatSize(vol.Capacity),
            formatSize(vol.Used),
            vol.Encrypted,
            vol.CaseSensitive,
            created.Format("2006-01-02 15:04:05"),
        )
    }
    w.Flush()
}

func formatSize(size uint64) string {
    // Convert size to human-readable format
}

func formatVolumeRole(role uint16) string {
    // Convert role to string
}

func getAPFSVersion(sb *container.NXSuperblock) string {
    // Determine APFS version based on features
}
```

### cmd/apfs-mount/
#### main.go

```go
package main

import (
    "flag"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    
    "apfs/pkg/container"
    "apfs/pkg/fs"
    "apfs/pkg/types"
    "apfs/pkg/util"
    "apfs/cmd/apfs-mount/fuse"
)

func main() {
    // Parse command line flags
    devicePath := flag.String("device", "", "Path to APFS device")
    mountPoint := flag.String("mountpoint", "", "Mount point directory")
    volumeName := flag.String("volume", "", "Volume name to mount")
    volumeIndex := flag.Int("index", -1, "Volume index to mount")
    readOnly := flag.Bool("readonly", false, "Mount read-only")
    password := flag.String("password", "", "Password for encrypted volumes")
    debug := flag.Bool("debug", false, "Enable debug logging")
    
    flag.Parse()
    
    // Validate arguments
    if *devicePath == "" || *mountPoint == "" {
        fmt.Println("Error: Device path and mount point are required")
        flag.Usage()
        os.Exit(1)
    }
    
    if *volumeName == "" && *volumeIndex < 0 {
        fmt.Println("Error: Either volume name or volume index must be specified")
        flag.Usage()
        os.Exit(1)
    }
    
    // Open block device
    device, err := util.NewBlockDevice(*devicePath, *readOnly)
    if err != nil {
        fmt.Printf("Error opening device: %v\n", err)
        os.Exit(1)
    }
    defer device.Close()
    
    // Create container manager
    containerMgr, err := container.NewContainerManager(device)
    if err != nil {
        fmt.Printf("Error parsing APFS container: %v\n", err)
        os.Exit(1)
    }
    
    // Find volume
    var volumeManager *fs.VolumeManager
    
    if *volumeIndex >= 0 {
        volumeManager, err = containerMgr.GetVolume(uint32(*volumeIndex))
    } else {
        volumes, err := containerMgr.ListVolumes()
        if err != nil {
            fmt.Printf("Error listing volumes: %v\n", err)
            os.Exit(1)
        }
        
        for _, vol := range volumes {
            if vol.Name == *volumeName {
                volumeManager, err = containerMgr.GetVolume(vol.Index)
                break
            }
        }
        
        if volumeManager == nil {
            fmt.Printf("Error: Volume '%s' not found\n", *volumeName)
            os.Exit(1)
        }
    }
    
    if err != nil {
        fmt.Printf("Error mounting volume: %v\n", err)
        os.Exit(1)
    }
    
    // Handle encrypted volumes
    if volumeManager.IsEncrypted && *password == "" {
        fmt.Printf("Error: Volume is encrypted, password required\n")
        os.Exit(1)
    }
    
    if volumeManager.IsEncrypted {
        // Unlock volume with password
        err = volumeManager.Unlock(*password)
        if err != nil {
            fmt.Printf("Error unlocking volume: %v\n", err)
            os.Exit(1)
        }
    }
    
    // Initialize FUSE filesystem
    fuseFS := fuse.NewAPFSFileSystem(volumeManager, *readOnly, *debug)
    
    // Mount filesystem
    fmt.Printf("Mounting volume '%s' at %s\n", volumeManager.GetVolumeName(), *mountPoint)
    
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-sigCh
        fmt.Println("\nUnmounting...")
        fuseFS.Unmount()
        os.Exit(0)
    }()
    
    if err := fuseFS.Mount(*mountPoint); err != nil {
        fmt.Printf("Error mounting filesystem: %v\n", err)
        os.Exit(1)
    }
}
```

### cmd/apfs-recover/
#### main.go

```go
package main

import (
    "flag"
    "fmt"
    "os"
    "path/filepath"
    
    "apfs/pkg/container"
    "apfs/pkg/fs"
    "apfs/pkg/types"
    "apfs/pkg/util"
)

func main() {
    // Parse command line flags
    devicePath := flag.String("device", "", "Path to APFS device")
    outputDir := flag.String("output", "", "Output directory for recovered files")
    volumeName := flag.String("volume", "", "Volume name to recover from")
    volumeIndex := flag.Int("index", -1, "Volume index to recover from")
    recursive := flag.Bool("recursive", false, "Recover files recursively")
    password := flag.String("password", "", "Password for encrypted volumes")
    path := flag.String("path", "/", "Path to recover")
    all := flag.Bool("all", false, "Recover all files")
    
    flag.Parse()
    
    // Validate arguments
    if *devicePath == "" || *outputDir == "" {
        fmt.Println("Error: Device path and output directory are required")
        flag.Usage()
        os.Exit(1)
    }
    
    // Create output directory if it doesn't exist
    if err := os.MkdirAll(*outputDir, 0755); err != nil {
        fmt.Printf("Error creating output directory: %v\n", err)
        os.Exit(1)
    }
    
    // Open block device
    device, err := util.NewBlockDevice(*devicePath, true)
    if err != nil {
        fmt.Printf("Error opening device: %v\n", err)
        os.Exit(1)
    }
    defer device.Close()
    
    // Create container manager
    containerMgr, err := container.NewContainerManager(device)
    if err != nil {
        fmt.Printf("Error parsing APFS container: %v\n", err)
        os.Exit(1)
    }
    
    // Find volume
    var volumeManager *fs.VolumeManager
    
    if *volumeIndex >= 0 {
        volumeManager, err = containerMgr.GetVolume(uint32(*volumeIndex))
    } else if *volumeName != "" {
        volumes, err := containerMgr.ListVolumes()
        if err != nil {
            fmt.Printf("Error listing volumes: %v\n", err)
            os.Exit(1)
        }
        
        for _, vol := range volumes {
            if vol.Name == *volumeName {
                volumeManager, err = containerMgr.GetVolume(vol.Index)
                break
            }
        }
        
        if volumeManager == nil {
            fmt.Printf("Error: Volume '%s' not found\n", *volumeName)
            os.Exit(1)
        }
    } else {
        // If no volume specified, use the first one
        volumeManager, err = containerMgr.GetVolume(0)
    }
    
    if err != nil {
        fmt.Printf("Error accessing volume: %v\n", err)
        os.Exit(1)
    }
    
    // Handle encrypted volumes
    if volumeManager.IsEncrypted && *password == "" {
        fmt.Printf("Error: Volume is encrypted, password required\n")
        os.Exit(1)
    }
    
    if volumeManager.IsEncrypted {
        // Unlock volume with password
        err = volumeManager.Unlock(*password)
        if err != nil {
            fmt.Printf("Error unlocking volume: %v\n", err)
            os.Exit(1)
        }
    }
    
    // Start recovery
    fmt.Printf("Starting recovery from volume '%s'\n", volumeManager.GetVolumeName())
    
    if *all {
        // Recover all files
        RecoverAllFiles(volumeManager, *outputDir)
    } else {
        // Recover specific path
        fileInfo, err := volumeManager.GetFileByPath(*path)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            os.Exit(1)
        }
        
        if fileInfo.IsDir() {
            RecoverDirectory(volumeManager, fileInfo.Sys.(*fs.Inode), *outputDir, *recursive)
        } else {
            RecoverFile(volumeManager, fileInfo.Sys.(*fs.Inode), *outputDir)
        }
    }
    fmt.Println("Recovery completed successfully")
}

// RecoverFile recovers a single file
func RecoverFile(volume *fs.VolumeManager, inode *fs.Inode, outputDir string) error {
    // Get file name
    fileName := inode.Name
    if fileName == "" {
        fileName = fmt.Sprintf("file_%d", inode.ObjectID)
    }
    
    // Create output file
    outputPath := filepath.Join(outputDir, fileName)
    fmt.Printf("Recovering file: %s\n", outputPath)
    
    outFile, err := os.Create(outputPath)
    if err != nil {
        fmt.Printf("Error creating output file: %v\n", err)
        return err
    }
    defer outFile.Close()
    
    // Determine file size
    fileSize := int64(0)
    if inode.DataStream != nil {
        fileSize = int64(inode.DataStream.Size)
    }
    
    // Read and write file data in chunks
    const chunkSize = 1024 * 1024 // 1MB chunks
    var offset int64 = 0
    
    for offset < fileSize {
        size := chunkSize
        if offset+int64(size) > fileSize {
            size = int(fileSize - offset)
        }
        
        data, err := inode.ReadData(offset, size)
        if err != nil {
            fmt.Printf("Error reading file data at offset %d: %v\n", offset, err)
            return err
        }
        
        _, err = outFile.Write(data)
        if err != nil {
            fmt.Printf("Error writing to output file: %v\n", err)
            return err
        }
        
        offset += int64(size)
        fmt.Printf("\rRecovered %d/%d bytes (%.1f%%)", offset, fileSize, float64(offset*100)/float64(fileSize))
    }
    
    fmt.Println()
    
    // Set file modification time
    modTime := fs.AppleTimeToTime(inode.ModTime)
    err = os.Chtimes(outputPath, modTime, modTime)
    if err != nil {
        fmt.Printf("Warning: Could not set file times: %v\n", err)
    }
    
    return nil
}

// RecoverDirectory recovers a directory and its contents
func RecoverDirectory(volume *fs.VolumeManager, inode *fs.Inode, outputDir string, recursive bool) error {
    // Get directory name
    dirName := inode.Name
    if dirName == "" {
        dirName = fmt.Sprintf("dir_%d", inode.ObjectID)
    }
    
    // Create output directory
    outputPath := filepath.Join(outputDir, dirName)
    fmt.Printf("Recovering directory: %s\n", outputPath)
    
    err := os.MkdirAll(outputPath, 0755)
    if err != nil {
        fmt.Printf("Error creating output directory: %v\n", err)
        return err
    }
    
    // List directory contents
    entries, err := volume.ListDirectory(types.OID(inode.ObjectID))
    if err != nil {
        fmt.Printf("Error listing directory: %v\n", err)
        return err
    }
    
    // Recover each entry
    for _, entry := range entries {
        // Skip "." and ".." entries
        if entry.Name == "." || entry.Name == ".." {
            continue
        }
        
        // Get inode for the entry
        childInode, err := volume.GetFile(types.OID(entry.FileID))
        if err != nil {
            fmt.Printf("Error getting file %s: %v\n", entry.Name, err)
            continue
        }
        
        // Set the name from the directory entry
        childInode.Name = entry.Name
        
        if childInode.IsDir() {
            if recursive {
                // Recursively recover subdirectory
                err = RecoverDirectory(volume, childInode, outputPath, recursive)
                if err != nil {
                    fmt.Printf("Error recovering subdirectory %s: %v\n", entry.Name, err)
                }
            }
        } else if childInode.IsRegular() {
            // Recover regular file
            err = RecoverFile(volume, childInode, outputPath)
            if err != nil {
                fmt.Printf("Error recovering file %s: %v\n", entry.Name, err)
            }
        } else if childInode.IsSymlink() {
            // Recover symbolic link
            target, err := volume.ReadSymlink(childInode)
            if err != nil {
                fmt.Printf("Error reading symlink %s: %v\n", entry.Name, err)
                continue
            }
            
            linkPath := filepath.Join(outputPath, entry.Name)
            err = os.Symlink(target, linkPath)
            if err != nil {
                fmt.Printf("Error creating symlink %s: %v\n", entry.Name, err)
            }
        }
    }
    
    return nil
}

// RecoverAllFiles recovers all files in the volume
func RecoverAllFiles(volume *fs.VolumeManager, outputDir string) error {
    // Get root directory inode
    rootInode, err := volume.GetFile(types.ROOT_DIR_INO_NUM)
    if err != nil {
        fmt.Printf("Error getting root directory: %v\n", err)
        return err
    }
    
    rootInode.Name = "root"
    
    // Recursively recover everything from root
    return RecoverDirectory(volume, rootInode, outputDir, true)
}
```

### cmd/apfs-mount/fuse/
#### filesystem.go

```go
package fuse

import (
    "context"
    "os"
    "syscall"
    "time"
    
    "bazil.org/fuse"
    "bazil.org/fuse/fs"
    
    "apfs/pkg/fs"
    "apfs/pkg/types"
)

// APFSFileSystem implements the FUSE filesystem interface for APFS
type APFSFileSystem struct {
    volume    *fs.VolumeManager
    conn      *fuse.Conn
    readOnly  bool
    debug     bool
}

// APFSNode represents a file or directory in the FUSE filesystem
type APFSNode struct {
    fs      *APFSFileSystem
    inode   *fs.Inode
    oid     types.OID
}

// APFSDir represents a directory in the FUSE filesystem
type APFSDir struct {
    APFSNode
}

// APFSFile represents a file in the FUSE filesystem
type APFSFile struct {
    APFSNode
}

// APFSSymlink represents a symbolic link in the FUSE filesystem
type APFSSymlink struct {
    APFSNode
}

// NewAPFSFileSystem creates a new APFS FUSE filesystem
func NewAPFSFileSystem(volume *fs.VolumeManager, readOnly bool, debug bool) *APFSFileSystem {
    return &APFSFileSystem{
        volume:   volume,
        readOnly: readOnly,
        debug:    debug,
    }
}

// Mount mounts the filesystem at the specified mount point
func (afs *APFSFileSystem) Mount(mountpoint string) error {
    var err error
    
    options := []fuse.MountOption{
        fuse.FSName("apfs"),
        fuse.Subtype("apfs"),
        fuse.LocalVolume(),
        fuse.VolumeName(afs.volume.GetVolumeName()),
    }
    
    if afs.readOnly {
        options = append(options, fuse.ReadOnly())
    }
    
    if afs.debug {
        options = append(options, fuse.Debug())
    }
    
    afs.conn, err = fuse.Mount(mountpoint, options...)
    if err != nil {
        return err
    }
    
    // Serve filesystem
    err = fs.Serve(afs.conn, afs)
    if err != nil {
        return err
    }
    
    // Wait for unmount
    <-afs.conn.Ready
    return afs.conn.MountError
}

// Unmount unmounts the filesystem
func (afs *APFSFileSystem) Unmount() error {
    if afs.conn != nil {
        return fuse.Unmount(afs.conn.MountPoint)
    }
    return nil
}

// Root implements the fs.FS interface
func (afs *APFSFileSystem) Root() (fs.Node, error) {
    // Get root directory inode
    rootInode, err := afs.volume.GetFile(types.ROOT_DIR_INO_NUM)
    if err != nil {
        return nil, err
    }
    
    return &APFSDir{
        APFSNode: APFSNode{
            fs:    afs,
            inode: rootInode,
            oid:   types.ROOT_DIR_INO_NUM,
        },
    }, nil
}

// Attr implements the fs.Node interface
func (n *APFSNode) Attr(ctx context.Context, attr *fuse.Attr) error {
    attr.Inode = uint64(n.oid)
    attr.Mode = os.FileMode(n.inode.Mode)
    attr.Uid = n.inode.UID
    attr.Gid = n.inode.GID
    attr.Mtime = fs.AppleTimeToTime(n.inode.ModTime)
    attr.Ctime = fs.AppleTimeToTime(n.inode.ChangeTime)
    attr.Atime = fs.AppleTimeToTime(n.inode.AccessTime)
    attr.Crtime = fs.AppleTimeToTime(n.inode.CreateTime)
    
    if n.inode.DataStream != nil {
        attr.Size = n.inode.DataStream.Size
        attr.Blocks = (n.inode.DataStream.Size + 511) / 512
    }
    
    return nil
}

// Lookup implements the fs.NodeStringLookuper interface
func (d *APFSDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
    // Find the directory entry
    entry, err := fs.LookupDirectoryEntry(d.fs.volume, d.oid, name)
    if err != nil {
        return nil, syscall.ENOENT
    }
    
    // Get the inode
    inode, err := d.fs.volume.GetFile(types.OID(entry.FileID))
    if err != nil {
        return nil, err
    }
    
    // Set the name from directory entry
    inode.Name = entry.Name
    
    // Create the appropriate node type
    if inode.IsDir() {
        return &APFSDir{
            APFSNode: APFSNode{
                fs:    d.fs,
                inode: inode,
                oid:   types.OID(entry.FileID),
            },
        }, nil
    } else if inode.IsSymlink() {
        return &APFSSymlink{
            APFSNode: APFSNode{
                fs:    d.fs,
                inode: inode,
                oid:   types.OID(entry.FileID),
            },
        }, nil
    } else {
        return &APFSFile{
            APFSNode: APFSNode{
                fs:    d.fs,
                inode: inode,
                oid:   types.OID(entry.FileID),
            },
        }, nil
    }
}

// ReadDirAll implements the fs.HandleReadDirAller interface
func (d *APFSDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
    // List directory entries
    entries, err := d.fs.volume.ListDirectory(d.oid)
    if err != nil {
        return nil, err
    }
    
    // Convert to FUSE directory entries
    dirents := make([]fuse.Dirent, 0, len(entries))
    for _, entry := range entries {
        var direntType fuse.DirentType
        
        switch entry.GetFileType() {
        case types.DT_DIR:
            direntType = fuse.DT_Dir
        case types.DT_REG:
            direntType = fuse.DT_File
        case types.DT_LNK:
            direntType = fuse.DT_Link
        case types.DT_FIFO:
            direntType = fuse.DT_FIFO
        case types.DT_SOCK:
            direntType = fuse.DT_Socket
        case types.DT_CHR:
            direntType = fuse.DT_Char
        case types.DT_BLK:
            direntType = fuse.DT_Block
        default:
            direntType = fuse.DT_Unknown
        }
        
        dirents = append(dirents, fuse.Dirent{
            Inode: uint64(entry.FileID),
            Type:  direntType,
            Name:  entry.Name,
        })
    }
    
    return dirents, nil
}

// Open implements the fs.NodeOpener interface
func (f *APFSFile) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
    if f.fs.readOnly && (req.Flags&fuse.OpenFlags(os.O_WRONLY|os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) != 0) {
        return nil, syscall.EROFS
    }
    
    // Set direct IO flag to disable kernel caching (optional)
    if f.fs.debug {
        resp.Flags |= fuse.OpenDirectIO
    }
    
    return f, nil
}

// Read implements the fs.HandleReader interface
func (f *APFSFile) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
    // Check if the inode has a data stream
    if f.inode.DataStream == nil {
        return nil
    }
    
    // Read the data from the file
    data, err := f.inode.ReadData(req.Offset, req.Size)
    if err != nil {
        return err
    }
    
    resp.Data = data
    return nil
}

// Readlink implements the fs.NodeReadlinker interface
func (s *APFSSymlink) Readlink(ctx context.Context, req *fuse.ReadlinkRequest) (string, error) {
    // Read the symlink target
    target, err := s.fs.volume.ReadSymlink(s.inode)
    if err != nil {
        return "", err
    }
    
    return target, nil
}
```

#### helpers.go

```go
package fuse

import (
    "time"
    
    "apfs/pkg/fs"
)

// AppleTimeToTime converts Apple time (nanoseconds since 1970-01-01) to Go time
func AppleTimeToTime(appleTime uint64) time.Time {
    return time.Unix(0, int64(appleTime))
}

// TimeToAppleTime converts Go time to Apple time
func TimeToAppleTime(t time.Time) uint64 {
    return uint64(t.UnixNano())
}

// FormatFuseErrorf formats and logs a FUSE error message
func (afs *APFSFileSystem) FormatFuseErrorf(format string, args ...interface{}) {
    if afs.debug {
        log.Printf(format, args...)
    }
}
```

```go
```

```go
```

```go
```