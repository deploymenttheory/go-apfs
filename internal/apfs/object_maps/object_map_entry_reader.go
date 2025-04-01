package objectmaps

import (
	"github.com/deploymenttheory/go-apfs/internal/types"
)

type ObjectMapEntry struct {
	Key types.OmapKeyT
	Val types.OmapValT
}

func NewObjectMapEntry(key types.OmapKeyT, val types.OmapValT) *ObjectMapEntry {
	return &ObjectMapEntry{Key: key, Val: val}
}

func (e *ObjectMapEntry) ObjectID() types.OidT {
	return e.Key.OkOid
}

func (e *ObjectMapEntry) TransactionID() types.XidT {
	return e.Key.OkXid
}

func (e *ObjectMapEntry) Flags() uint32 {
	return e.Val.OvFlags
}

func (e *ObjectMapEntry) Size() uint32 {
	return e.Val.OvSize
}

func (e *ObjectMapEntry) PhysicalAddress() types.Paddr {
	return e.Val.OvPaddr
}

func (e *ObjectMapEntry) IsDeleted() bool {
	return (e.Val.OvFlags & types.OmapValDeleted) != 0
}

func (e *ObjectMapEntry) IsEncrypted() bool {
	return (e.Val.OvFlags & types.OmapValEncrypted) != 0
}

func (e *ObjectMapEntry) HasHeader() bool {
	return (e.Val.OvFlags & types.OmapValNoheader) == 0
}
