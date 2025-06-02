package extendedfields

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

type DirectoryExtendedFieldReader struct {
	fields []interfaces.ExtendedField
}

func NewDirectoryExtendedFieldReader(fields []interfaces.ExtendedField) *DirectoryExtendedFieldReader {
	return &DirectoryExtendedFieldReader{fields: fields}
}

func (r *DirectoryExtendedFieldReader) SiblingID() (uint64, bool) {
	return readUint64Field(r.fields, 9)
}

func (r *DirectoryExtendedFieldReader) FileSystemUUID() (types.UUID, bool) {
	for _, f := range r.fields {
		if f.Type() == 10 && len(f.Data()) == 16 {
			var uuid types.UUID
			copy(uuid[:], f.Data())
			return uuid, true
		}
	}
	return types.UUID{}, false
}
