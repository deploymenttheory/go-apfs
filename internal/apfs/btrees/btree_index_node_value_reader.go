package btrees

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// btreeIndexNodeValueReader implements the BTreeIndexNodeValueReader interface
type btreeIndexNodeValueReader struct {
	indexValue *types.BtnIndexNodeValT
}

// NewBTreeIndexNodeValueReader creates a new BTreeIndexNodeValueReader implementation
func NewBTreeIndexNodeValueReader(indexValue *types.BtnIndexNodeValT) interfaces.BTreeIndexNodeValueReader {
	return &btreeIndexNodeValueReader{
		indexValue: indexValue,
	}
}

// ChildObjectID returns the object identifier of the child node
func (inr *btreeIndexNodeValueReader) ChildObjectID() types.OidT {
	return inr.indexValue.BinvChildOid
}

// ChildHash returns the hash of the child node
func (inr *btreeIndexNodeValueReader) ChildHash() [types.BtreeNodeHashSizeMax]byte {
	return inr.indexValue.BinvChildHash
}
