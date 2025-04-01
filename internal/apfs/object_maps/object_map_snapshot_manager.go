package objectmaps

import (
	"errors"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// ObjectMapSnapshotInfo provides basic access to snapshot metadata
type ObjectMapSnapshotInfo interface {
	Oid() types.OidT
	Flags() uint32
	IsDeleted() bool
	IsReverted() bool
}

// ObjectMapSnapshot implements ObjectMapSnapshotInfo
type ObjectMapSnapshot struct {
	snapshot types.OmapSnapshotT
}

func (s *ObjectMapSnapshot) Oid() types.OidT {
	return s.snapshot.OmsOid
}

func (s *ObjectMapSnapshot) Flags() uint32 {
	return s.snapshot.OmsFlags
}

func (s *ObjectMapSnapshot) IsDeleted() bool {
	return (s.snapshot.OmsFlags & types.OmapSnapshotDeleted) != 0
}

func (s *ObjectMapSnapshot) IsReverted() bool {
	return (s.snapshot.OmsFlags & types.OmapSnapshotReverted) != 0
}

// ObjectMapSnapshotManagerImpl provides access to a list of OmapSnapshotT
type ObjectMapSnapshotManagerImpl struct {
	snapshots []types.OmapSnapshotT
}

func NewObjectMapSnapshotManager(snapshots []types.OmapSnapshotT) *ObjectMapSnapshotManagerImpl {
	return &ObjectMapSnapshotManagerImpl{snapshots: snapshots}
}

func (m *ObjectMapSnapshotManagerImpl) ListSnapshots() ([]ObjectMapSnapshotInfo, error) {
	result := make([]ObjectMapSnapshotInfo, 0, len(m.snapshots))
	for _, snap := range m.snapshots {
		result = append(result, &ObjectMapSnapshot{snapshot: snap})
	}
	return result, nil
}

func (m *ObjectMapSnapshotManagerImpl) FindSnapshotByOID(oid types.OidT) (ObjectMapSnapshotInfo, error) {
	for _, snap := range m.snapshots {
		if snap.OmsOid == oid {
			return &ObjectMapSnapshot{snapshot: snap}, nil
		}
	}
	return nil, errors.New("snapshot not found")
}
