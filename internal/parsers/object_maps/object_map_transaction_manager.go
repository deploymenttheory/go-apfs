package objectmaps

import (
	"github.com/deploymenttheory/go-apfs/internal/types"
)

type ObjectMapTransactionManager struct {
	Header types.OmapPhysT
}

func NewObjectMapTransactionManager(header types.OmapPhysT) *ObjectMapTransactionManager {
	return &ObjectMapTransactionManager{Header: header}
}

func (o *ObjectMapTransactionManager) PendingRevertMinXID() types.XidT {
	return o.Header.OmPendingRevertMin
}

func (o *ObjectMapTransactionManager) PendingRevertMaxXID() types.XidT {
	return o.Header.OmPendingRevertMax
}

func (o *ObjectMapTransactionManager) IsRevertInProgress() bool {
	return o.Header.OmPendingRevertMin != 0 || o.Header.OmPendingRevertMax != 0
}
