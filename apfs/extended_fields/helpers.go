package extendedfields

import (
	"github.com/deploymenttheory/go-apfs/apfs/types"
)

// Field wraps types.XFieldT to provide helper methods.
type Field struct {
	*types.XFieldT
}

// IsSystemField returns true if the XField has the XfSystemField flag set.
//
// This indicates that the field was created by the kernel or APFS implementation,
// and cannot be removed or modified by user-space code.
//
// Reference: APFS Reference, page 113.
func (xf *Field) IsSystemField() bool {
	return xf.XFlags&uint8(types.XfSystemField) != 0
}

// IsUserField returns true if the XField has the XfUserField flag set.
//
// This indicates that the field was created by a user-space program.
//
// Reference: APFS Reference, page 113.
func (xf *Field) IsUserField() bool {
	return xf.XFlags&uint8(types.XfUserField) != 0
}
