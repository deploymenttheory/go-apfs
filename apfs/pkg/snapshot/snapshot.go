package snapshot

import (
	"github.com/deploymenttheory/go-apfs/pkg/types"
)

// SnapMetadata represents snapshot metadata (j_snap_metadata_val_t)
type SnapMetadata struct {
	ExtentrefTreeOID  types.OID // Extent reference tree OID
	SblockOID         types.OID // Superblock OID
	CreateTime        uint64    // Creation time
	ChangeTime        uint64    // Change time
	Inum              uint64    // Inode number
	ExtentrefTreeType uint32    // Extent reference tree type
	Flags             uint32    // Flags
	NameLen           uint16    // Name length
	Name              string    // Snapshot name
}

// SnapMetadataKey represents a snapshot metadata key (j_snap_metadata_key_t)
type SnapMetadataKey struct {
	Header types.JKey // Key header
}

// SnapName represents a snapshot name (j_snap_name_val_t)
type SnapName struct {
	SnapXID types.XID // Snapshot transaction ID
}

// SnapNameKey represents a snapshot name key (j_snap_name_key_t)
type SnapNameKey struct {
	Header  types.JKey // Key header
	NameLen uint16     // Name length
	Name    string     // Snapshot name
}

// SnapMetaFlags represents snapshot metadata flags
type SnapMetaFlags uint32

const (
	// SnapMetaPendingDataless indicates a snapshot is pending conversion to dataless
	SnapMetaPendingDataless SnapMetaFlags = 0x00000001
	// SnapMetaMergeInProgress indicates a snapshot merge is in progress
	SnapMetaMergeInProgress SnapMetaFlags = 0x00000002
)

// SnapMetaExtObjPhys represents additional snapshot metadata (snap_meta_ext_obj_phys_t)
type SnapMetaExtObjPhys struct {
	Object types.ObjectHeader // Object header
	SME    SnapMetaExt        // Snapshot metadata extension
}

// SnapMetaExt represents extended snapshot metadata (snap_meta_ext_t)
type SnapMetaExt struct {
	Version uint32     // Structure version
	Flags   uint32     // Flags
	SnapXID types.XID  // Snapshot XID
	UUID    types.UUID // Snapshot UUID
	Token   uint64     // Opaque metadata token
}

// SnapshotInfo contains information about a snapshot
type SnapshotInfo struct {
	Name       string         // Snapshot name
	XID        types.XID      // Transaction ID
	UUID       types.UUID     // Snapshot UUID
	CreateTime types.TimeSpec // Creation time
	ChangeTime types.TimeSpec // Change time
	Flags      SnapMetaFlags  // Flags
}

// IsPendingDataless returns true if the snapshot is pending conversion to dataless
func (s *SnapMetadata) IsPendingDataless() bool {
	return (s.Flags & uint32(SnapMetaPendingDataless)) != 0
}

// IsMergeInProgress returns true if a snapshot merge is in progress
func (s *SnapMetadata) IsMergeInProgress() bool {
	return (s.Flags & uint32(SnapMetaMergeInProgress)) != 0
}

// Snapshot manages an APFS snapshot
type Snapshot struct {
	XID           types.XID     // Transaction ID
	Name          string        // Snapshot name
	Metadata      *SnapMetadata // Snapshot metadata
	ExtMetadata   *SnapMetaExt  // Extended metadata
	ExtentRefTree types.OID     // Reference to extent tree
	SuperblockOID types.OID     // Reference to volume superblock
}

// GetInfo returns information about the snapshot
func (s *Snapshot) GetInfo() *SnapshotInfo {
	var uuid types.UUID
	if s.ExtMetadata != nil {
		uuid = s.ExtMetadata.UUID
	}

	return &SnapshotInfo{
		Name:       s.Name,
		XID:        s.XID,
		UUID:       uuid,
		CreateTime: types.TimeSpec(s.Metadata.CreateTime),
		ChangeTime: types.TimeSpec(s.Metadata.ChangeTime),
		Flags:      SnapMetaFlags(s.Metadata.Flags),
	}
}

// SnapshotManager manages snapshots for a volume
type SnapshotManager struct {
	VolumeOID    types.OID               // Volume object ID
	SnapMetaTree types.OID               // Snapshot metadata tree OID
	SnapMetaExt  types.OID               // Extended snapshot metadata OID
	Snapshots    map[types.XID]*Snapshot // Map of snapshots by XID
}

// NewSnapshotManager creates a new snapshot manager
func NewSnapshotManager(volumeOID, snapMetaTree, snapMetaExt types.OID) *SnapshotManager {
	return &SnapshotManager{
		VolumeOID:    volumeOID,
		SnapMetaTree: snapMetaTree,
		SnapMetaExt:  snapMetaExt,
		Snapshots:    make(map[types.XID]*Snapshot),
	}
}

// GetSnapshot gets a snapshot by XID
func (sm *SnapshotManager) GetSnapshot(xid types.XID) (*Snapshot, error) {
	snapshot, exists := sm.Snapshots[xid]
	if !exists {
		return nil, types.ErrSnapshotNotFound
	}
	return snapshot, nil
}

// GetSnapshotByName gets a snapshot by name
func (sm *SnapshotManager) GetSnapshotByName(name string) (*Snapshot, error) {
	for _, snapshot := range sm.Snapshots {
		if snapshot.Name == name {
			return snapshot, nil
		}
	}
	return nil, types.ErrSnapshotNotFound
}

// ListSnapshots lists all snapshots
func (sm *SnapshotManager) ListSnapshots() []*SnapshotInfo {
	snapshots := make([]*SnapshotInfo, 0, len(sm.Snapshots))
	for _, snapshot := range sm.Snapshots {
		snapshots = append(snapshots, snapshot.GetInfo())
	}
	return snapshots
}
