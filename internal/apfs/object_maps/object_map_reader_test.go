package objectmaps

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestObjectMapReaderImpl(t *testing.T) {
	header := types.OmapPhysT{
		OmFlags:            0x01,
		OmSnapCount:        3,
		OmTreeType:         2,
		OmSnapshotTreeType: 4,
		OmTreeOid:          1001,
		OmSnapshotTreeOid:  2002,
		OmMostRecentSnap:   123456789,
	}

	reader := NewObjectMapReader(header)

	if reader.Flags() != header.OmFlags {
		t.Errorf("Flags() = %d, want %d", reader.Flags(), header.OmFlags)
	}
	if reader.SnapshotCount() != header.OmSnapCount {
		t.Errorf("SnapshotCount() = %d, want %d", reader.SnapshotCount(), header.OmSnapCount)
	}
	if reader.TreeType() != header.OmTreeType {
		t.Errorf("TreeType() = %d, want %d", reader.TreeType(), header.OmTreeType)
	}
	if reader.SnapshotTreeType() != header.OmSnapshotTreeType {
		t.Errorf("SnapshotTreeType() = %d, want %d", reader.SnapshotTreeType(), header.OmSnapshotTreeType)
	}
	if reader.TreeOID() != header.OmTreeOid {
		t.Errorf("TreeOID() = %d, want %d", reader.TreeOID(), header.OmTreeOid)
	}
	if reader.SnapshotTreeOID() != header.OmSnapshotTreeOid {
		t.Errorf("SnapshotTreeOID() = %d, want %d", reader.SnapshotTreeOID(), header.OmSnapshotTreeOid)
	}
	if reader.MostRecentSnapshotXID() != header.OmMostRecentSnap {
		t.Errorf("MostRecentSnapshotXID() = %d, want %d", reader.MostRecentSnapshotXID(), header.OmMostRecentSnap)
	}
}
