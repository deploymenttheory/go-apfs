package objects

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// ObjectIdentifier implements the ObjectIdentifier interface
type ObjectIdentifier struct {
	id  types.OidT
	xid types.XidT
}

// Ensure ObjectIdentifier implements interfaces.ObjectIdentifier
var _ interfaces.ObjectIdentifier = (*ObjectIdentifier)(nil)

// NewObjectIdentifier creates a new ObjectIdentifier
func NewObjectIdentifier(id types.OidT, xid types.XidT) *ObjectIdentifier {
	return &ObjectIdentifier{
		id:  id,
		xid: xid,
	}
}

// ID returns the object's unique identifier
func (o *ObjectIdentifier) ID() types.OidT {
	return o.id
}

// TransactionID returns the transaction identifier of the most recent modification
func (o *ObjectIdentifier) TransactionID() types.XidT {
	return o.xid
}

// IsValid checks if the object identifier is valid
func (o *ObjectIdentifier) IsValid() bool {
	return o.id != types.OidInvalid && o.xid != types.XidInvalid
}
