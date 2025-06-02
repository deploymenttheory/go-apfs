package btrees

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// btreeLocationReader implements the BTreeLocationReader interface
type btreeLocationReader struct {
	location types.NlocT
}

// NewBTreeLocationReader creates a new BTreeLocationReader implementation
func NewBTreeLocationReader(location types.NlocT) interfaces.BTreeLocationReader {
	return &btreeLocationReader{
		location: location,
	}
}

// Offset returns the offset in bytes
func (lr *btreeLocationReader) Offset() uint16 {
	return lr.location.Off
}

// Length returns the length in bytes
func (lr *btreeLocationReader) Length() uint16 {
	return lr.location.Len
}

// IsValid checks if the offset is valid
func (lr *btreeLocationReader) IsValid() bool {
	return lr.location.Off != types.BtoffInvalid
}
