package extendedfields

import (
	"encoding/binary"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
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

// readUint32Field attempts to extract a 4-byte little-endian uint32 from the
// data of the extended field with the given type.
//
// If a matching field is found and contains at least 4 bytes of data,
// the function returns the parsed value and true. Otherwise, it returns 0 and false.
func readUint32Field(fields []interfaces.ExtendedField, target uint8) (uint32, bool) {
	for _, f := range fields {
		if f.Type() == target && len(f.Data()) >= 4 {
			return binary.LittleEndian.Uint32(f.Data()), true
		}
	}
	return 0, false
}

// readUint64Field attempts to extract an 8-byte little-endian uint64 from the
// data of the extended field with the given type.
//
// If a matching field is found and contains at least 8 bytes of data,
// the function returns the parsed value and true. Otherwise, it returns 0 and false.
func readUint64Field(fields []interfaces.ExtendedField, target uint8) (uint64, bool) {
	for _, f := range fields {
		if f.Type() == target && len(f.Data()) >= 8 {
			return binary.LittleEndian.Uint64(f.Data()), true
		}
	}
	return 0, false
}
