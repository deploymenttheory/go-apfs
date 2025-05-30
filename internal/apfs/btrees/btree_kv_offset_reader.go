package btrees

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// btreeKVOffsetReader implements the BTreeKVOffsetReader interface
type btreeKVOffsetReader struct {
	kvOffset types.KvoffT
}

// NewBTreeKVOffsetReader creates a new BTreeKVOffsetReader implementation
func NewBTreeKVOffsetReader(kvOffset types.KvoffT) interfaces.BTreeKVOffsetReader {
	return &btreeKVOffsetReader{
		kvOffset: kvOffset,
	}
}

// KeyOffset returns the offset of the key
func (kvor *btreeKVOffsetReader) KeyOffset() uint16 {
	return kvor.kvOffset.K
}

// ValueOffset returns the offset of the value
func (kvor *btreeKVOffsetReader) ValueOffset() uint16 {
	return kvor.kvOffset.V
}
