package objectmaps

import (
	"github.com/deploymenttheory/go-apfs/internal/types"
)

type ObjectMapReader struct {
	Header types.OmapPhysT
}

func NewObjectMapReader(header types.OmapPhysT) *ObjectMapReader {
	return &ObjectMapReader{Header: header}
}

func (r *ObjectMapReader) Flags() uint32 {
	return r.Header.OmFlags
}

func (r *ObjectMapReader) SnapshotCount() uint32 {
	return r.Header.OmSnapCount
}

func (r *ObjectMapReader) TreeType() uint32 {
	return r.Header.OmTreeType
}

func (r *ObjectMapReader) SnapshotTreeType() uint32 {
	return r.Header.OmSnapshotTreeType
}

func (r *ObjectMapReader) TreeOID() types.OidT {
	return r.Header.OmTreeOid
}

func (r *ObjectMapReader) SnapshotTreeOID() types.OidT {
	return r.Header.OmSnapshotTreeOid
}

func (r *ObjectMapReader) MostRecentSnapshotXID() types.XidT {
	return r.Header.OmMostRecentSnap
}
