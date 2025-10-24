package extendedfields

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

type InodeExtendedFieldReader struct {
	fields []interfaces.ExtendedField
}

func NewInodeExtendedFieldReader(fields []interfaces.ExtendedField) *InodeExtendedFieldReader {
	return &InodeExtendedFieldReader{fields: fields}
}

func (r *InodeExtendedFieldReader) SnapshotTransactionID() (uint64, bool) {
	return readUint64Field(r.fields, 5)
}

func (r *InodeExtendedFieldReader) DeltaTreeOID() (types.OidT, bool) {
	v, ok := readUint64Field(r.fields, 6)
	return types.OidT(v), ok
}

func (r *InodeExtendedFieldReader) DocumentID() (uint32, bool) {
	return readUint32Field(r.fields, 1)
}

func (r *InodeExtendedFieldReader) PreviousFileSize() (uint64, bool) {
	return readUint64Field(r.fields, 4)
}

func (r *InodeExtendedFieldReader) FinderInfo() ([]byte, bool) {
	for _, f := range r.fields {
		if f.Type() == uint8(types.JObjTypeExtent) {
			return f.Data(), true
		}
	}
	return nil, false
}

func (r *InodeExtendedFieldReader) SparseByteCount() (uint64, bool) {
	return readUint64Field(r.fields, 3)
}

func (r *InodeExtendedFieldReader) DeviceIdentifier() (uint32, bool) {
	return readUint32Field(r.fields, 7)
}

func (r *InodeExtendedFieldReader) OriginalSyncRootID() (uint64, bool) {
	return readUint64Field(r.fields, 8)
}
