package types

import (
	"encoding/binary"
	"time"
)

// =============================================================================
// Container Layer Structures
// =============================================================================

// NXSuperblock represents a container superblock (nx_superblock_t)
type NXSuperblock struct {
	Header                 ObjectHeader // Object header
	Magic                  uint32       // Magic ('NXSB')
	BlockSize              uint32       // Block size
	BlockCount             uint64       // Number of blocks
	Features               uint64       // Optional features
	ReadOnlyCompatFeatures uint64       // Read-only compatible features
	IncompatFeatures       uint64       // Incompatible features
	UUID                   UUID         // Container UUID
	NextOID                OID          // Next object ID
	NextXID                XID          // Next transaction ID
	XPDescBlocks           uint32       // Checkpoint descriptor blocks
	XPDataBlocks           uint32       // Checkpoint data blocks
	XPDescBase             PAddr        // Checkpoint descriptor base
	XPDataBase             PAddr        // Checkpoint data base
	XPDescNext             uint32       // Next checkpoint descriptor
	XPDataNext             uint32       // Next checkpoint data
	XPDescIndex            uint32       // Checkpoint descriptor index
	XPDescLen              uint32       // Checkpoint descriptor length
	XPDataIndex            uint32       // Checkpoint data index
	XPDataLen              uint32       // Checkpoint data length
	SpacemanOID            OID          // Space manager OID
	OMapOID                OID          // Object map OID
	ReaperOID              OID          // Reaper OID
	TestType               uint32       // For testing
	MaxFileSystems         uint32       // Max number of volumes
	FSOID                  [100]OID     // Volume OIDs
	Counters               [32]uint64   // Array of counters
	BlockedOutRange        PRange       // Blocked out range
	EvictMappingTreeOID    OID          // Evict mapping tree OID
	Flags                  uint64       // Flags
	EFIJumpstart           PAddr        // EFI jumpstart
	FusionUUID             UUID         // Fusion UUID
	KeyLocker              PRange       // Keybag location
	EphemeralInfo          [4]uint64    // Ephemeral info
	TestOID                OID          // Test OID
	FusionMtOID            OID          // Fusion middle tree OID
	FusionWbcOID           OID          // Fusion write-back cache OID
	FusionWbc              PRange       // Fusion write-back cache
	NewestMountedVersion   uint64       // Newest mounted version
	MkbLocker              PRange       // Media key locker
}

// IsValid checks if the superblock has a valid magic number
func (sb *NXSuperblock) IsValid() bool {
	return sb.Magic == NXMagic
}

// IsCryptoSW checks if the container uses software encryption
func (sb *NXSuperblock) IsCryptoSW() bool {
	return (sb.Flags & NXCryptoSW) != 0
}

// IsFusion checks if the container is a Fusion container
func (sb *NXSuperblock) IsFusion() bool {
	return (sb.IncompatFeatures & NXIncompatFusion) != 0
}

// SupportsDefrag checks if the container supports defragmentation
func (sb *NXSuperblock) SupportsDefrag() bool {
	return (sb.Features & NXFeatureDefrag) != 0
}

// IsVersion2 checks if the container is APFS version 2
func (sb *NXSuperblock) IsVersion2() bool {
	return (sb.IncompatFeatures & NXIncompatVersion2) != 0
}

// CheckpointMapPhys represents a checkpoint mapping block (checkpoint_map_phys_t)
type CheckpointMapPhys struct {
	Header ObjectHeader        // Object header
	Flags  uint32              // Flags
	Count  uint32              // Number of mappings
	Map    []CheckpointMapping // Checkpoint mappings
}

// IsLast returns true if this is the last checkpoint mapping block
func (cmp *CheckpointMapPhys) IsLast() bool {
	return (cmp.Flags & CheckpointMapLast) != 0
}

// CheckpointMapping represents a mapping for a checkpoint (checkpoint_mapping_t)
type CheckpointMapping struct {
	Type    uint32 // Object type
	Subtype uint32 // Object subtype
	Size    uint32 // Size in bytes
	Pad     uint32 // Reserved
	FSOID   OID    // Volume object ID
	OID     OID    // Object ID
	PAddr   PAddr  // Physical address
}

// EFIJumpstart represents EFI jumpstart information (nx_efi_jumpstart_t)
type EFIJumpstart struct {
	Header     ObjectHeader // Object header
	Magic      uint32       // Magic ('JSDR')
	Version    uint32       // Version
	EFIFileLen uint32       // EFI file length
	NumExtents uint32       // Number of extents
	Reserved   [16]uint64   // Reserved
	Extents    []PRange     // EFI extents
}

// IsValid checks if the jumpstart has a valid magic number
func (js *EFIJumpstart) IsValid() bool {
	return js.Magic == JSMagic
}

// OMapPhys represents an object map (omap_phys_t)
type OMapPhys struct {
	Header           ObjectHeader // Object header
	Flags            uint32       // Flags
	SnapCount        uint32       // Number of snapshots
	TreeType         uint32       // Tree type
	SnapshotTreeType uint32       // Snapshot tree type
	TreeOID          OID          // Tree OID
	SnapshotTreeOID  OID          // Snapshot tree OID
	MostRecentSnap   XID          // Most recent snapshot
	PendingRevertMin XID          // Pending revert minimum
	PendingRevertMax XID          // Pending revert maximum
}

// IsManuallyManaged returns true if the object map doesn't support snapshots
func (o *OMapPhys) IsManuallyManaged() bool {
	return (o.Flags & OMapManuallyManaged) != 0
}

// IsEncrypting returns true if encryption is in progress
func (o *OMapPhys) IsEncrypting() bool {
	return (o.Flags & OMapEncrypting) != 0
}

// IsDecrypting returns true if decryption is in progress
func (o *OMapPhys) IsDecrypting() bool {
	return (o.Flags & OMapDecrypting) != 0
}

// IsKeyrolling returns true if key rolling is in progress
func (o *OMapPhys) IsKeyrolling() bool {
	return (o.Flags & OMapKeyrolling) != 0
}

// OMapKey represents a key in the object map (omap_key_t)
type OMapKey struct {
	OID OID // Object identifier
	XID XID // Transaction identifier
}

// OMapVal represents a value in the object map (omap_val_t)
type OMapVal struct {
	Flags uint32 // Flags
	Size  uint32 // Size in bytes
	PAddr PAddr  // Physical address
}

// IsDeleted returns true if the object has been deleted
func (v *OMapVal) IsDeleted() bool {
	return (v.Flags & OMapValDeleted) != 0
}

// IsEncrypted returns true if the object is encrypted
func (v *OMapVal) IsEncrypted() bool {
	return (v.Flags & OMapValEncrypted) != 0
}

// HasNoHeader returns true if the object has no header
func (v *OMapVal) HasNoHeader() bool {
	return (v.Flags & OMapValNoheader) != 0
}

// OMapSnapshot represents a snapshot in the object map (omap_snapshot_t)
type OMapSnapshot struct {
	Flags uint32 // Flags
	Pad   uint32 // Reserved
	OID   OID    // Reserved
}

// IsDeleted returns true if the snapshot has been deleted
func (s *OMapSnapshot) IsDeleted() bool {
	return (s.Flags & OMapSnapshotDeleted) != 0
}

// IsReverted returns true if the snapshot has been reverted
func (s *OMapSnapshot) IsReverted() bool {
	return (s.Flags & OMapSnapshotReverted) != 0
}

// SpacemanDeviceInfo represents device-specific space manager information (spaceman_device_t)
type SpacemanDeviceInfo struct {
	BlockCount uint64 // Total blocks
	ChunkCount uint64 // Total chunks
	CIBCount   uint32 // Chunk info blocks
	CABCount   uint32 // CIB address blocks
	FreeCount  uint64 // Free blocks
	AddrOffset uint32 // Address offset
	Reserved   uint32 // Reserved
	Reserved2  uint64 // Reserved
}

// SpacemanFreeQueue represents a free queue (spaceman_free_queue_t)
type SpacemanFreeQueue struct {
	Count         uint64 // Number of entries
	TreeOID       OID    // B-tree object ID
	OldestXID     XID    // Oldest transaction ID
	TreeNodeLimit uint16 // Tree node limit
	Pad16         uint16 // Padding
	Pad32         uint32 // Padding
	Reserved      uint64 // Reserved
}

// SpacemanPhys represents the space manager (spaceman_phys_t)
type SpacemanPhys struct {
	Header              ObjectHeader          // Object header
	BlockSize           uint32                // Block size
	BlocksPerChunk      uint32                // Blocks per chunk
	ChunksPerCIB        uint32                // Chunks per CIB
	CIBsPerCAB          uint32                // CIBs per CAB
	Devices             [2]SpacemanDeviceInfo // Main and tier2 devices
	Flags               uint32                // Flags
	IPBmTxMultiplier    uint32                // Internal pool bitmap TX multiplier
	IPBlockCount        uint64                // Internal pool block count
	IPBmSizeInBlocks    uint32                // Internal pool bitmap size
	IPBmBlockCount      uint32                // Internal pool bitmap block count
	IPBmBase            PAddr                 // Internal pool bitmap base
	IPBase              PAddr                 // Internal pool base
	FSReserveBlockCount uint64                // FS reserve block count
	FSReserveAllocCount uint64                // FS reserve alloc count
	FreeQueues          [3]SpacemanFreeQueue  // Free queues (IP, main, tier2)
	IPBmFreeHead        uint16                // Internal pool bitmap free head
	IPBmFreeTail        uint16                // Internal pool bitmap free tail
	IPBmXidOffset       uint32                // Internal pool bitmap XID offset
	IPBitmapOffset      uint32                // Internal pool bitmap offset
	IPBmFreeNextOffset  uint32                // Internal pool bitmap free next offset
	Version             uint32                // Version
	StructSize          uint32                // Structure size
	// Skip datazone_info_phys_t for now - complex structure
}

// IsVersioned returns true if the space manager is versioned
func (sm *SpacemanPhys) IsVersioned() bool {
	return (sm.Flags & 0x00000001) != 0 // SM_FLAG_VERSIONED
}

// ChunkInfo represents information about a chunk of blocks (chunk_info_t)
type ChunkInfo struct {
	XID        XID    // Transaction ID
	Addr       PAddr  // Base address
	BlockCount uint32 // Total blocks
	FreeCount  uint32 // Free blocks
	BitmapAddr PAddr  // Bitmap address
}

// ChunkInfoBlock represents a block containing chunk information (chunk_info_block)
type ChunkInfoBlock struct {
	Header ObjectHeader // Object header
	Index  uint32       // Index
	Count  uint32       // Number of entries
	Chunks []ChunkInfo  // Chunk entries
}

// CIBAddrBlock represents a block of chunk info block addresses (cib_addr_block)
type CIBAddrBlock struct {
	Header    ObjectHeader // Object header
	Index     uint32       // Index
	Count     uint32       // Number of entries
	Addresses []PAddr      // Addresses
}

// EvictMappingVal represents a range of physical addresses for eviction (evict_mapping_val_t)
type EvictMappingVal struct {
	DstPAddr PAddr  // Destination address
	Length   uint64 // Length
}

// BTNodePhys represents a B-tree node (btree_node_phys_t)
type BTNodePhys struct {
	Header      ObjectHeader // Object header
	Flags       uint16       // Flags
	Level       uint16       // Level in tree (0=leaf)
	KeyCount    uint32       // Number of keys
	TableSpace  NLoc         // Table of contents location
	FreeSpace   NLoc         // Free space location
	KeyFreeList NLoc         // Free list for keys
	ValFreeList NLoc         // Free list for values
	Data        []byte       // Node data (variable length)
}

// IsLeaf returns true if this is a leaf node
func (n *BTNodePhys) IsLeaf() bool {
	return (n.Flags & BtnodeLeaf) != 0
}

// IsRoot returns true if this is a root node
func (n *BTNodePhys) IsRoot() bool {
	return (n.Flags & BtnodeRoot) != 0
}

// HasFixedKVSize returns true if keys and values have fixed size
func (n *BTNodePhys) HasFixedKVSize() bool {
	return (n.Flags & BtnodeFixedKVSize) != 0
}

// IsHashed returns true if node contains hashed values
func (n *BTNodePhys) IsHashed() bool {
	return (n.Flags & BtnodeHashed) != 0
}

// HasNoHeader returns true if the node has no header
func (n *BTNodePhys) HasNoHeader() bool {
	return (n.Flags & BtnodeNoheader) != 0
}

// BTInfoFixed represents static B-tree information (btree_info_fixed_t)
type BTInfoFixed struct {
	Flags    uint32 // Flags
	NodeSize uint32 // Node size
	KeySize  uint32 // Key size (0 if variable)
	ValSize  uint32 // Value size (0 if variable)
}

// AllowsGhosts returns true if the B-tree allows ghost entries
func (bi *BTInfoFixed) AllowsGhosts() bool {
	return (bi.Flags & BtreeAllowGhosts) != 0
}

// IsEphemeral returns true if the B-tree is ephemeral
func (bi *BTInfoFixed) IsEphemeral() bool {
	return (bi.Flags & BtreeEphemeral) != 0
}

// IsPhysical returns true if the B-tree is physical
func (bi *BTInfoFixed) IsPhysical() bool {
	return (bi.Flags & BtreePhysical) != 0
}

// IsNonPersistent returns true if the B-tree is non-persistent
func (bi *BTInfoFixed) IsNonPersistent() bool {
	return (bi.Flags & BtreeNonpersistent) != 0
}

// IsHashed returns true if the B-tree is hashed
func (bi *BTInfoFixed) IsHashed() bool {
	return (bi.Flags & BtreeHashed) != 0
}

// BTreeInfo represents B-tree information (btree_info_t)
type BTreeInfo struct {
	Fixed      BTInfoFixed // Fixed information
	LongestKey uint32      // Longest key
	LongestVal uint32      // Longest value
	KeyCount   uint64      // Total keys
	NodeCount  uint64      // Total nodes
}

// BTreeNodeIndexVal represents index node values in a hashed tree (btn_index_node_val_t)
type BTreeNodeIndexVal struct {
	ChildOID  OID    // Child node OID
	ChildHash []byte // Child node hash (variable length)
}

// NXReaperPhys represents the reaper structure (nx_reaper_phys_t)
type NXReaperPhys struct {
	Header          ObjectHeader // Object header
	NextReapID      uint64       // Next reap ID
	CompletedID     uint64       // Completed ID
	Head            OID          // Head of reap list
	Tail            OID          // Tail of reap list
	Flags           uint32       // Flags
	RLCount         uint32       // Reap list count
	Type            uint32       // Type
	Size            uint32       // Size
	FSOID           OID          // File system OID
	OID             OID          // Object ID
	XID             XID          // Transaction ID
	NRLEFlags       uint32       // Reap list entry flags
	StateBufferSize uint32       // State buffer size
	StateBuffer     []byte       // State buffer
}

// IsContinuing returns true if reaping is in progress
func (r *NXReaperPhys) IsContinuing() bool {
	return (r.Flags & NRContinue) != 0
}

// NXReapListPhys represents a reap list (nx_reap_list_phys_t)
type NXReapListPhys struct {
	Header  ObjectHeader      // Object header
	Next    OID               // Next list
	Flags   uint32            // Flags
	Max     uint32            // Maximum entries
	Count   uint32            // Current count
	First   uint32            // First entry
	Last    uint32            // Last entry
	Free    uint32            // Free entry
	Entries []NXReapListEntry // Entries
}

// NXReapListEntry represents an entry in a reap list (nx_reap_list_entry_t)
type NXReapListEntry struct {
	Next  uint32 // Next entry
	Flags uint32 // Flags
	Type  uint32 // Type
	Size  uint32 // Size
	FSOID OID    // File system OID
	OID   OID    // Object ID
	XID   XID    // Transaction ID
}

// IsValid returns true if the entry is valid
func (e *NXReapListEntry) IsValid() bool {
	return (e.Flags & NRLEValid) != 0
}

// FusionMTVal represents a Fusion middle tree value (fusion_mt_val_t)
type FusionMTVal struct {
	Lba    PAddr  // Logical block address
	Length uint32 // Length
	Flags  uint32 // Flags
}

// IsDirty returns true if the entry is dirty
func (v *FusionMTVal) IsDirty() bool {
	return (v.Flags & FusionMTDirty) != 0
}

// IsTenant returns true if the entry is a tenant
func (v *FusionMTVal) IsTenant() bool {
	return (v.Flags & FusionMTTenant) != 0
}

// FusionWBC represents a Fusion write-back cache state (fusion_wbc_phys_t)
type FusionWBC struct {
	Header           ObjectHeader // Object header
	Version          uint64       // Version
	ListHeadOID      OID          // List head OID
	ListTailOID      OID          // List tail OID
	StableHeadOffset uint64       // Stable head offset
	StableTailOffset uint64       // Stable tail offset
	ListBlocksCount  uint32       // List blocks count
	Reserved         uint32       // Reserved
	UsedByRC         uint64       // Used by RC
	RCStash          PRange       // RC stash
}

// FusionWBCListEntry represents a Fusion write-back cache list entry (fusion_wbc_list_entry_t)
type FusionWBCListEntry struct {
	WbcLba    PAddr  // Write-back cache logical block address
	TargetLba PAddr  // Target logical block address
	Length    uint64 // Length
}

// FusionWBCList represents a Fusion write-back cache list (fusion_wbc_list_phys_t)
type FusionWBCList struct {
	Header     ObjectHeader         // Object header
	Version    uint64               // Version
	TailOffset uint64               // Tail offset
	IndexBegin uint32               // Index begin
	IndexEnd   uint32               // Index end
	IndexMax   uint32               // Index max
	Reserved   uint32               // Reserved
	Entries    []FusionWBCListEntry // List entries
}

// ERStatePhys represents the encryption rolling state (er_state_phys_t)
type ERStatePhys struct {
	Header               ObjectHeader // Object header
	Magic                uint32       // Magic ('FLAB')
	Version              uint32       // Version
	Flags                uint64       // Flags
	SnapXID              uint64       // Snapshot XID
	CurrentFextObjID     uint64       // Current file extent object ID
	FileOffset           uint64       // File offset
	Progress             uint64       // Progress
	TotalBlkToEncrypt    uint64       // Total blocks to encrypt
	BlockmapOID          OID          // Block map OID
	TidemarkObjID        uint64       // Tidemark object ID
	RecoveryExtentsCount uint64       // Recovery extents count
	RecoveryListOID      OID          // Recovery list OID
	RecoveryLength       uint64       // Recovery length
}

// IsEncrypting returns true if encryption is in progress
func (er *ERStatePhys) IsEncrypting() bool {
	return (er.Flags & ERSBFlagEncrypting) != 0
}

// IsDecrypting returns true if decryption is in progress
func (er *ERStatePhys) IsDecrypting() bool {
	return (er.Flags & ERSBFlagDecrypting) != 0
}

// IsKeyrolling returns true if key rolling is in progress
func (er *ERStatePhys) IsKeyrolling() bool {
	return (er.Flags & ERSBFlagKeyrolling) != 0
}

// IsPaused returns true if encryption rolling is paused
func (er *ERStatePhys) IsPaused() bool {
	return (er.Flags & ERSBFlagPaused) != 0
}

// HasFailed returns true if encryption rolling has failed
func (er *ERStatePhys) HasFailed() bool {
	return (er.Flags & ERSBFlagFailed) != 0
}

// IntegrityMetaPhys represents integrity metadata for a sealed volume (integrity_meta_phys_t)
type IntegrityMetaPhys struct {
	Header         ObjectHeader // Object header
	Version        uint32       // Version
	Flags          uint32       // Flags
	HashType       uint32       // Hash type
	RootHashOffset uint32       // Root hash offset
	BrokenXID      XID          // XID that broke the seal
	Reserved       [9]uint64    // Reserved
}

// IsSealBroken returns true if the seal has been broken
func (im *IntegrityMetaPhys) IsSealBroken() bool {
	return (im.Flags & APFSSealBroken) != 0
}

// IsVersionValid returns true if the version is valid
func (im *IntegrityMetaPhys) IsVersionValid() bool {
	return im.Version >= IntegrityMetaVersion1 && im.Version <= IntegrityMetaVersionHighest
}

// =============================================================================
// File System Layer Structures
// =============================================================================

// APFSSuperblock represents a volume superblock (apfs_superblock_t)
type APFSSuperblock struct {
	Header                 ObjectHeader    // Object header
	Magic                  uint32          // Magic ('APSB')
	FSIndex                uint32          // Index in container's array
	Features               uint64          // Optional features
	ReadOnlyCompatFeatures uint64          // Read-only compatible features
	IncompatFeatures       uint64          // Incompatible features
	UnmountTime            uint64          // Last unmount time (nanoseconds)
	ReserveBlockCount      uint64          // Reserved block count
	QuotaBlockCount        uint64          // Quota block count
	AllocCount             uint64          // Allocated block count
	MetaCrypto             MetaCryptoState // Metadata crypto state
	RootTreeType           uint32          // Root tree type
	ExtentrefTreeType      uint32          // Extent reference tree type
	SnapMetaTreeType       uint32          // Snapshot metadata tree type
	OMapOID                OID             // Object map OID
	RootTreeOID            OID             // Root tree OID
	ExtentrefTreeOID       OID             // Extent reference tree OID
	SnapMetaTreeOID        OID             // Snapshot metadata tree OID
	RevertToXID            XID             // Revert to transaction ID
	RevertToSblockOID      OID             // Revert to superblock OID
	NextObjID              uint64          // Next object ID
	NumFiles               uint64          // Number of files
	NumDirectories         uint64          // Number of directories
	NumSymlinks            uint64          // Number of symlinks
	NumOtherFSObjects      uint64          // Number of other objects
	NumSnapshots           uint64          // Number of snapshots
	TotalBlocksAlloced     uint64          // Total blocks allocated
	TotalBlocksFreed       uint64          // Total blocks freed
	UUID                   UUID            // Volume UUID
	LastModTime            uint64          // Last modification time
	FSFlags                uint64          // Flags
	FormattedBy            ModifiedBy      // Formatted by
	ModifiedBy             [8]ModifiedBy   // Modified by history
	VolName                [256]byte       // Volume name
	NextDocID              uint32          // Next document ID
	Role                   uint16          // Role (system, data, etc.)
	Reserved               uint16          // Reserved
	RootToXID              XID             // Root from transaction ID
	ERStateOID             OID             // Encryption rolling state OID
	CloneinfoIDEpoch       uint64          // Clone info ID epoch
	CloneinfoXID           uint64          // Clone info transaction ID
	SnapMetaExtOID         OID             // Extended snapshot metadata OID
	VolumeGroupID          UUID            // Volume group UUID
	IntegrityMetaOID       OID             // Integrity metadata OID
	FextTreeOID            OID             // File extent tree OID
	FextTreeType           uint32          // File extent tree type
	ReservedType           uint32          // Reserved
	ReservedOID            OID             // Reserved
}

// IsValid checks if the superblock has a valid magic number
func (sb *APFSSuperblock) IsValid() bool {
	return sb.Magic == APFSMagic
}

// IsUnencrypted returns true if the volume is unencrypted
func (sb *APFSSuperblock) IsUnencrypted() bool {
	return (sb.FSFlags & APFSFSUnencrypted) != 0
}

// UsesOneKey returns true if the volume uses a single encryption key
func (sb *APFSSuperblock) UsesOneKey() bool {
	return (sb.FSFlags & APFSFSOnekey) != 0
}

// IsSpilledOver returns true if the volume has spilled over
func (sb *APFSSuperblock) IsSpilledOver() bool {
	return (sb.FSFlags & APFSFSSpilledover) != 0
}

// ShouldRunSpilloverCleaner returns true if the volume should run the spillover cleaner
func (sb *APFSSuperblock) ShouldRunSpilloverCleaner() bool {
	return (sb.FSFlags & APFSFSRunSpilloverCleaner) != 0
}

// AlwaysCheckExtentref returns true if the volume always checks extent references
func (sb *APFSSuperblock) AlwaysCheckExtentref() bool {
	return (sb.FSFlags & APFSFSAlwaysCheckExtentref) != 0
}

// IsCaseSensitive returns true if the volume is case-sensitive
func (sb *APFSSuperblock) IsCaseSensitive() bool {
	return (sb.IncompatFeatures & APFSIncompatCaseInsensitive) == 0
}

// HasDatalessSnaps returns true if the volume has dataless snapshots
func (sb *APFSSuperblock) HasDatalessSnaps() bool {
	return (sb.IncompatFeatures & APFSIncompatDatalessSnaps) != 0
}

// IsNormalizationInsensitive returns true if the volume is normalization-insensitive
func (sb *APFSSuperblock) IsNormalizationInsensitive() bool {
	return (sb.IncompatFeatures & APFSIncompatNormalizationInsensitive) != 0
}

// IsSealed returns true if the volume is sealed
func (sb *APFSSuperblock) IsSealed() bool {
	return (sb.IncompatFeatures & APFSIncompatSealedVolume) != 0
}

// IsSystem returns true if the volume has the system role
func (sb *APFSSuperblock) IsSystem() bool {
	return (sb.Role & APFSVolRoleSystem) != 0
}

// IsData returns true if the volume has the data role
func (sb *APFSSuperblock) IsData() bool {
	return (sb.Role & APFSVolRoleData) != 0
}

// VolumeName returns the volume name as a string
func (sb *APFSSuperblock) VolumeName() string {
	// Find null terminator
	nameLen := 0
	for ; nameLen < len(sb.VolName); nameLen++ {
		if sb.VolName[nameLen] == 0 {
			break
		}
	}
	return string(sb.VolName[:nameLen])
}

// HasIntegrityMeta returns true if the volume has integrity metadata
func (sb *APFSSuperblock) HasIntegrityMeta() bool {
	return sb.IntegrityMetaOID != 0 && sb.IsSealed()
}

// ModifiedBy represents information about software that modified the volume (apfs_modified_by_t)
type ModifiedBy struct {
	ID        [32]byte // Identifier
	Timestamp uint64   // Timestamp
	LastXID   XID      // Last transaction ID
}

// GetIDString returns the ID as a string
func (m *ModifiedBy) GetIDString() string {
	// Find null terminator
	idLen := 0
	for ; idLen < len(m.ID); idLen++ {
		if m.ID[idLen] == 0 {
			break
		}
	}
	return string(m.ID[:idLen])
}

// GetTimestamp returns the timestamp as a time.Time
func (m *ModifiedBy) GetTimestamp() time.Time {
	return time.Unix(0, int64(m.Timestamp))
}

// JKey represents the common header for file-system keys (j_key_t)
type JKeyStruct struct {
	ObjIDAndType uint64 // Object ID and type
}

// GetObjID returns the object ID from a j_key
func (k *JKeyStruct) GetObjID() OID {
	return OID(k.ObjIDAndType & OBJ_ID_MASK)
}

// GetType returns the type from a j_key
func (k *JKeyStruct) GetType() uint8 {
	return uint8((k.ObjIDAndType & OBJECT_TYPE_MASK_JKEY) >> OBJ_TYPE_SHIFT)
}

// NewJKeyStruct creates a new JKey with the given object ID and type
func NewJKeyStruct(objID OID, objType uint8) JKeyStruct {
	return JKeyStruct{
		ObjIDAndType: uint64(objID) | (uint64(objType) << OBJ_TYPE_SHIFT),
	}
}

// JInodeKey represents an inode key (j_inode_key_t)
type JInodeKey struct {
	Header JKeyStruct // Key header
}

// JInodeVal represents an inode value (j_inode_val_t)
type JInodeVal struct {
	ParentID         uint64          // Parent directory ID
	PrivateID        uint64          // Private ID for extents
	CreateTime       uint64          // Creation time
	ModTime          uint64          // Modification time
	ChangeTime       uint64          // Change time (attributes)
	AccessTime       uint64          // Last access time
	InternalFlags    uint64          // Internal flags
	NChildren        int32           // Number of children (for directories)
	NLink            int32           // Number of hard links (for files)
	DefaultProtClass uint32          // Default protection class
	WriteGenCounter  uint32          // Write generation counter
	BSDFlags         uint32          // BSD flags
	UID              uint32          // Owner user ID
	GID              uint32          // Owner group ID
	Mode             uint16          // File mode
	Pad1             uint16          // Padding
	UncompressedSize uint64          // Uncompressed size
	XFields          []ExtendedField // Extended fields
}

// IsApfsPrivate returns true if the inode is used internally by APFS
func (i *JInodeVal) IsApfsPrivate() bool {
	return (i.InternalFlags & InodeIsAPFSPrivate) != 0
}

// MaintainsDirStats returns true if the inode maintains directory statistics
func (i *JInodeVal) MaintainsDirStats() bool {
	return (i.InternalFlags & InodeMaintainDirStats) != 0
}

// IsDirStatsOrigin returns true if the inode is a directory statistics origin
func (i *JInodeVal) IsDirStatsOrigin() bool {
	return (i.InternalFlags & InodeDirStatsOrigin) != 0
}

// HasExplicitProtClass returns true if the inode has an explicit protection class
func (i *JInodeVal) HasExplicitProtClass() bool {
	return (i.InternalFlags & InodeProtClassExplicit) != 0
}

// WasCloned returns true if the inode was created by cloning
func (i *JInodeVal) WasCloned() bool {
	return (i.InternalFlags & InodeWasCloned) != 0
}

// HasSecurityEA returns true if the inode has a security extended attribute
func (i *JInodeVal) HasSecurityEA() bool {
	return (i.InternalFlags & InodeHasSecurityEA) != 0
}

// IsBeingTruncated returns true if the inode is being truncated
func (i *JInodeVal) IsBeingTruncated() bool {
	return (i.InternalFlags & InodeBeingTruncated) != 0
}

// HasFinderInfo returns true if the inode has finder info
func (i *JInodeVal) HasFinderInfo() bool {
	return (i.InternalFlags & InodeHasFinderInfo) != 0
}

// IsSparse returns true if the inode is sparse
func (i *JInodeVal) IsSparse() bool {
	return (i.InternalFlags & InodeIsSparse) != 0
}

// WasEverCloned returns true if the inode has ever been cloned
func (i *JInodeVal) WasEverCloned() bool {
	return (i.InternalFlags & InodeWasEverCloned) != 0
}

// HasRsrcFork returns true if the inode has a resource fork
func (i *JInodeVal) HasRsrcFork() bool {
	return (i.InternalFlags & InodeHasRsrcFork) != 0
}

// HasNoRsrcFork returns true if the inode explicitly has no resource fork
func (i *JInodeVal) HasNoRsrcFork() bool {
	return (i.InternalFlags & InodeNoRsrcFork) != 0
}

// IsAllocationSpilledover returns true if the inode's allocation has spilled over
func (i *JInodeVal) IsAllocationSpilledover() bool {
	return (i.InternalFlags & InodeAllocationSpilledOver) != 0
}

// HasUncompressedSize returns true if the inode has an uncompressed size
func (i *JInodeVal) HasUncompressedSize() bool {
	return (i.InternalFlags & InodeHasUncompressedSize) != 0
}

// IsPurgeable returns true if the inode is purgeable
func (i *JInodeVal) IsPurgeable() bool {
	return (i.InternalFlags & InodeIsPurgeable) != 0
}

// GetExtendedField returns an extended field by type
func (i *JInodeVal) GetExtendedField(fieldType uint8) *ExtendedField {
	for idx := range i.XFields {
		if i.XFields[idx].Type == fieldType {
			return &i.XFields[idx]
		}
	}
	return nil
}

// IsDir returns true if this is a directory
func (i *JInodeVal) IsDir() bool {
	return (i.Mode & SIFmt) == SIFdir
}

// IsRegular returns true if this is a regular file
func (i *JInodeVal) IsRegular() bool {
	return (i.Mode & SIFmt) == SIFreg
}

// IsSymlink returns true if this is a symbolic link
func (i *JInodeVal) IsSymlink() bool {
	return (i.Mode & SIFmt) == SIFlnk
}

// JDrecKey represents a directory entry key (j_drec_key_t)
type JDrecKey struct {
	Header  JKeyStruct // Key header
	NameLen uint16     // Name length
	Name    []byte     // Name
}

// JDrecHashedKey represents a directory entry key with hash (j_drec_hashed_key_t)
type JDrecHashedKey struct {
	Header         JKeyStruct // Key header
	NameLenAndHash uint32     // Name length and hash
	Name           []byte     // Name
}

// GetNameLen returns the name length
func (k *JDrecHashedKey) GetNameLen() uint16 {
	return uint16(k.NameLenAndHash & JDrecLenMask)
}

// GetNameHash returns the name hash
func (k *JDrecHashedKey) GetNameHash() uint32 {
	return (k.NameLenAndHash & JDrecHashMask) >> JDrecHashShift
}

// GetNameString returns the name as a string
func (k *JDrecHashedKey) GetNameString() string {
	// Name should be null-terminated
	nameLen := k.GetNameLen()
	if nameLen > 0 && uint16(len(k.Name)) >= nameLen {
		return string(k.Name[:nameLen-1])
	}
	return ""
}

// JDrecVal represents a directory entry value (j_drec_val_t)
type JDrecVal struct {
	FileID    uint64          // File ID (inode number)
	DateAdded uint64          // Date added
	Flags     uint16          // Flags
	XFields   []ExtendedField // Extended fields
}

// GetFileType returns the file type
func (v *JDrecVal) GetFileType() uint8 {
	return uint8(v.Flags & DrecTypeMask)
}

// GetSiblingID returns the sibling ID from the extended fields
func (v *JDrecVal) GetSiblingID() (uint64, bool) {
	for _, field := range v.XFields {
		if field.Type == DrecExtTypeSiblingID && len(field.Data) >= 8 {
			return binary.LittleEndian.Uint64(field.Data), true
		}
	}
	return 0, false
}

// JDirStatsKey represents a directory stats key (j_dir_stats_key_t)
type JDirStatsKey struct {
	Header JKeyStruct // Key header
}

// JDirStatsVal represents a directory stats value (j_dir_stats_val_t)
type JDirStatsVal struct {
	NumChildren uint64 // Number of children
	TotalSize   uint64 // Total size
	ChainedKey  uint64 // Chained key
	GenCount    uint64 // Generation count
}

// JXattrKey represents an extended attribute key (j_xattr_key_t)
type JXattrKey struct {
	Header  JKeyStruct // Key header
	NameLen uint16     // Name length
	Name    []byte     // Name
}

// GetNameString returns the name as a string
func (k *JXattrKey) GetNameString() string {
	// Name should be null-terminated
	if k.NameLen > 0 && uint16(len(k.Name)) >= k.NameLen {
		return string(k.Name[:k.NameLen-1])
	}
	return ""
}

// JXattrVal represents an extended attribute value (j_xattr_val_t)
type JXattrVal struct {
	Flags   uint16 // Flags
	DataLen uint16 // Data length
	Data    []byte // Attribute data or data stream ID
}

// IsEmbedded returns true if the attribute data is embedded
func (v *JXattrVal) IsEmbedded() bool {
	return (v.Flags & XattrDataEmbedded) != 0
}

// IsDataStream returns true if the attribute data is stored in a data stream
func (v *JXattrVal) IsDataStream() bool {
	return (v.Flags & XattrDataStream) != 0
}

// IsFileSystemOwned returns true if the attribute is owned by the file system
func (v *JXattrVal) IsFileSystemOwned() bool {
	return (v.Flags & XattrFileSystemOwned) != 0
}

// JPhysExtKey represents a physical extent key (j_phys_ext_key_t)
type JPhysExtKey struct {
	Header JKeyStruct // Key header
}

// JPhysExtVal represents a physical extent value (j_phys_ext_val_t)
type JPhysExtVal struct {
	LenAndKind  uint64 // Length and kind
	OwningObjID uint64 // Owning object ID
	RefCount    int32  // Reference count
}

// GetLength returns the length in blocks
func (v *JPhysExtVal) GetLength() uint64 {
	return v.LenAndKind & PextLenMask
}

// GetKind returns the kind
func (v *JPhysExtVal) GetKind() uint8 {
	return uint8((v.LenAndKind & PextKindMask) >> PextKindShift)
}

// JFileExtentKey represents a file extent key (j_file_extent_key_t)
type JFileExtentKey struct {
	Header      JKeyStruct // Key header
	LogicalAddr uint64     // Logical address
}

// JFileExtentVal represents a file extent value (j_file_extent_val_t)
type JFileExtentVal struct {
	LenAndFlags  uint64 // Length and flags
	PhysBlockNum uint64 // Physical block number
	CryptoID     uint64 // Crypto ID
}

// GetLength returns the length in bytes
func (v *JFileExtentVal) GetLength() uint64 {
	return v.LenAndFlags & JFileExtentLenMask
}

// GetFlags returns the flags
func (v *JFileExtentVal) GetFlags() uint8 {
	return uint8((v.LenAndFlags & JFileExtentFlagMask) >> JFileExtentFlagShift)
}

// IsCryptoIDTweak returns true if the crypto ID is a tweak
func (v *JFileExtentVal) IsCryptoIDTweak() bool {
	return (v.GetFlags() & FextCryptoIDIsTweak) != 0
}

// JDstreamIDKey represents a data stream ID key (j_dstream_id_key_t)
type JDstreamIDKey struct {
	Header JKeyStruct // Key header
}

// JDstreamIDVal represents a data stream ID value (j_dstream_id_val_t)
type JDstreamIDVal struct {
	RefCount uint32 // Reference count
}

// JDstream represents a data stream (j_dstream_t)
type JDstream struct {
	Size              uint64 // Logical size in bytes
	AllocSize         uint64 // Allocated size
	DefaultCryptoID   uint64 // Default crypto ID
	TotalBytesWritten uint64 // Total bytes written
	TotalBytesRead    uint64 // Total bytes read
}

// JXattrDstream represents a data stream for extended attributes (j_xattr_dstream_t)
type JXattrDstream struct {
	XattrObjID uint64   // Extended attribute object ID
	Dstream    JDstream // Data stream
}

// MetaCryptoState represents wrapped metadata crypto state (wrapped_meta_crypto_state_t)
type MetaCryptoState struct {
	MajorVersion    uint16 // Major version
	MinorVersion    uint16 // Minor version
	Flags           uint32 // Flags
	PersistentClass uint32 // Protection class
	KeyOSVersion    uint32 // Key OS version
	KeyRevision     uint16 // Key revision
	Unused          uint16 // Unused
}

// WrappedCryptoState represents wrapped crypto state for a file (wrapped_crypto_state_t)
type WrappedCryptoState struct {
	MajorVersion    uint16 // Major version
	MinorVersion    uint16 // Minor version
	Flags           uint32 // Flags
	PersistentClass uint32 // Protection class
	KeyOSVersion    uint32 // Key OS version
	KeyRevision     uint16 // Key revision
	KeyLen          uint16 // Key length
	PersistentKey   []byte // Persistent key data
}

// JCryptoKey represents a per-file encryption state key (j_crypto_key_t)
type JCryptoKey struct {
	Header JKeyStruct // Key header
}

// JCryptoVal represents a per-file encryption state value (j_crypto_val_t)
type JCryptoVal struct {
	RefCount uint32             // Reference count
	State    WrappedCryptoState // Crypto state
}

// JSnapMetadataKey represents a snapshot metadata key (j_snap_metadata_key_t)
type JSnapMetadataKey struct {
	Header JKeyStruct // Key header
}

// JSnapMetadataVal represents a snapshot metadata value (j_snap_metadata_val_t)
type JSnapMetadataVal struct {
	ExtentrefTreeOID  OID    // Extent reference tree OID
	SblockOID         OID    // Superblock OID
	CreateTime        uint64 // Creation time
	ChangeTime        uint64 // Change time
	Inum              uint64 // Inode number
	ExtentrefTreeType uint32 // Extent reference tree type
	Flags             uint32 // Flags
	NameLen           uint16 // Name length
	Name              []byte // Snapshot name
}

// IsPendingDataless returns true if the snapshot is pending conversion to dataless
func (v *JSnapMetadataVal) IsPendingDataless() bool {
	return (v.Flags & SnapMetaPendingDataless) != 0
}

// IsMergeInProgress returns true if a snapshot merge is in progress
func (v *JSnapMetadataVal) IsMergeInProgress() bool {
	return (v.Flags & SnapMetaMergeInProgress) != 0
}

// GetNameString returns the name as a string
func (v *JSnapMetadataVal) GetNameString() string {
	// Name should be null-terminated
	if v.NameLen > 0 && uint16(len(v.Name)) >= v.NameLen {
		return string(v.Name[:v.NameLen-1])
	}
	return ""
}

// JSnapNameKey represents a snapshot name key (j_snap_name_key_t)
type JSnapNameKey struct {
	Header  JKeyStruct // Key header
	NameLen uint16     // Name length
	Name    []byte     // Name
}

// GetNameString returns the name as a string
func (k *JSnapNameKey) GetNameString() string {
	// Name should be null-terminated
	if k.NameLen > 0 && uint16(len(k.Name)) >= k.NameLen {
		return string(k.Name[:k.NameLen-1])
	}
	return ""
}

// JSnapNameVal represents a snapshot name value (j_snap_name_val_t)
type JSnapNameVal struct {
	SnapXID XID // Snapshot transaction ID
}

// JSiblingKey represents a sibling link key (j_sibling_key_t)
type JSiblingKey struct {
	Header    JKeyStruct // Key header
	SiblingID uint64     // Sibling ID
}

// JSiblingVal represents a sibling link value (j_sibling_val_t)
type JSiblingVal struct {
	ParentID uint64 // Parent directory ID
	NameLen  uint16 // Name length
	Name     []byte // Name
}

// GetNameString returns the name as a string
func (v *JSiblingVal) GetNameString() string {
	// Name should be null-terminated
	if v.NameLen > 0 && uint16(len(v.Name)) >= v.NameLen {
		return string(v.Name[:v.NameLen-1])
	}
	return ""
}

// JSiblingMapKey represents a sibling map key (j_sibling_map_key_t)
type JSiblingMapKey struct {
	Header JKeyStruct // Key header
}

// JSiblingMapVal represents a sibling map value (j_sibling_map_val_t)
type JSiblingMapVal struct {
	FileID uint64 // File ID (inode number)
}

// FextTreeKey represents a file extent tree key (fext_tree_key_t)
type FextTreeKey struct {
	PrivateID   uint64 // Private ID
	LogicalAddr uint64 // Logical address
}

// FextTreeVal represents a file extent tree value (fext_tree_val_t)
type FextTreeVal struct {
	LenAndFlags  uint64 // Length and flags
	PhysBlockNum uint64 // Physical block number
}

// GetLength returns the length in bytes
func (v *FextTreeVal) GetLength() uint64 {
	return v.LenAndFlags & JFileExtentLenMask
}

// GetFlags returns the flags
func (v *FextTreeVal) GetFlags() uint8 {
	return uint8((v.LenAndFlags & JFileExtentFlagMask) >> JFileExtentFlagShift)
}

// JFileInfoKey represents a file info key (j_file_info_key_t)
type JFileInfoKey struct {
	Header     JKeyStruct // Key header
	InfoAndLBA uint64     // Info type and LBA
}

// GetLBA returns the logical block address
func (k *JFileInfoKey) GetLBA() uint64 {
	return k.InfoAndLBA & JFileInfoLBAMask
}

// GetInfoType returns the info type
func (k *JFileInfoKey) GetInfoType() uint8 {
	return uint8((k.InfoAndLBA & JFileInfoTypeMask) >> JFileInfoTypeShift)
}

// JFileDataHashVal represents a file data hash value (j_file_data_hash_val_t)
type JFileDataHashVal struct {
	HashedLen uint16 // Length in blocks of hashed data
	HashSize  uint8  // Hash size
	Hash      []byte // Hash data
}

// KBLocker represents a keybag (kb_locker_t)
type KBLocker struct {
	Version  uint16        // Version
	NumKeys  uint16        // Number of keys
	NumBytes uint32        // Number of bytes
	Padding  [8]byte       // Padding
	Entries  []KeybagEntry // Entries
}

// KeybagEntry represents an entry in a keybag (keybag_entry_t)
type KeybagEntry struct {
	UUID    UUID    // UUID
	Tag     uint16  // Tag
	KeyLen  uint16  // Key data length
	Padding [4]byte // Padding
	KeyData []byte  // Key data
}

// MediaKeybag represents a keybag as a container-layer object (media_keybag_t)
type MediaKeybag struct {
	Header ObjectHeader // Object header
	Locker KBLocker     // Keybag
}

// ExtendedAttribute represents an extended attribute
type ExtendedAttribute struct {
	Name   string // Attribute name
	Data   []byte // Attribute data
	Flags  uint16 // Flags
	FileID OID    // File ID
}

// ExtendedField represents an extended field in an inode or directory entry
type ExtendedField struct {
	Type  uint8  // Field type
	Flags uint8  // Field flags
	Size  uint16 // Field size
	Data  []byte // Field data
}

// IsDataDependent returns true if the field depends on the file's data
func (f *ExtendedField) IsDataDependent() bool {
	return (f.Flags & XFDataDependent) != 0
}

// DoNotCopy returns true if the field should not be copied
func (f *ExtendedField) DoNotCopy() bool {
	return (f.Flags & XFDoNotCopy) != 0
}

// ChildrenInherit returns true if children should inherit this field
func (f *ExtendedField) ChildrenInherit() bool {
	return (f.Flags & XFChildrenInherit) != 0
}

// IsUserField returns true if the field was added by a user-space program
func (f *ExtendedField) IsUserField() bool {
	return (f.Flags & XFUserField) != 0
}

// IsSystemField returns true if the field was added by the system
func (f *ExtendedField) IsSystemField() bool {
	return (f.Flags & XFSystemField) != 0
}

// XFBlob represents a collection of extended fields (xf_blob_t)
type XFBlob struct {
	NumExts  uint16 // Number of extended fields
	UsedData uint16 // Used data size
	Data     []byte // Field data
}
