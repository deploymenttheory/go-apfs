package extendedfields

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// ExtendedField implements interfaces.ExtendedField using a parsed XFieldT header and raw data.
type ExtendedField struct {
	header types.XFieldT
	data   []byte
}

// NewExtendedField constructs an ExtendedField from a types.XFieldT header and associated data.
func NewExtendedField(header types.XFieldT, data []byte) interfaces.ExtendedField {
	return &ExtendedField{header: header, data: data}
}

// Type returns the extended field's type identifier (XType).
func (e *ExtendedField) Type() uint8 {
	return e.header.XType
}

// Flags returns the raw flags byte (XFlags) associated with this extended field.
func (e *ExtendedField) Flags() uint8 {
	return e.header.XFlags
}

// Size returns the size of the data segment, in bytes (XSize).
func (e *ExtendedField) Size() uint16 {
	return e.header.XSize
}

// Data returns the raw byte slice for the extended field's data payload.
func (e *ExtendedField) Data() []byte {
	return e.data
}

// IsDataDependent returns true if the extended field's flags indicate
// the value is data-dependent (XfDataDependent).
func (e *ExtendedField) IsDataDependent() bool {
	return e.header.XFlags&types.XfDataDependent != 0
}

// ShouldCopy returns true if the field should be copied on duplication.
// It returns false if the XfDoNotCopy flag is set.
func (e *ExtendedField) ShouldCopy() bool {
	return e.header.XFlags&types.XfDoNotCopy == 0
}

// IsUserField returns true if the XfUserField flag is set, indicating a user-space defined field.
func (e *ExtendedField) IsUserField() bool {
	return e.header.XFlags&types.XfUserField != 0
}

// IsSystemField returns true if the XfSystemField flag is set, indicating a kernel/APFS-defined field.
func (e *ExtendedField) IsSystemField() bool {
	return e.header.XFlags&types.XfSystemField != 0
}
