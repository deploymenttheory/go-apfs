package btrees

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// btreeKVLocationReader implements the BTreeKVLocationReader interface
type btreeKVLocationReader struct {
	kvLocation types.KvlocT
}

// NewBTreeKVLocationReader creates a new BTreeKVLocationReader implementation
func NewBTreeKVLocationReader(kvLocation types.KvlocT) interfaces.BTreeKVLocationReader {
	return &btreeKVLocationReader{
		kvLocation: kvLocation,
	}
}

// KeyLocation returns the location of the key
func (kvlr *btreeKVLocationReader) KeyLocation() types.NlocT {
	return kvlr.kvLocation.K
}

// ValueLocation returns the location of the value
func (kvlr *btreeKVLocationReader) ValueLocation() types.NlocT {
	return kvlr.kvLocation.V
}
