package siblings

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// siblingMapReader implements the SiblingMapReader interface
type siblingMapReader struct {
	key    *types.JSiblingMapKeyT
	value  *types.JSiblingMapValT
	endian binary.ByteOrder
}

// NewSiblingMapReader creates a new SiblingMapReader implementation
func NewSiblingMapReader(keyData, valueData []byte, endian binary.ByteOrder) (interfaces.SiblingMapReader, error) {
	if len(keyData) < 8 { // JKeyT (8 bytes)
		return nil, fmt.Errorf("key data too small for sibling map key: %d bytes", len(keyData))
	}

	if len(valueData) < 8 { // FileId (8 bytes)
		return nil, fmt.Errorf("value data too small for sibling map value: %d bytes", len(valueData))
	}

	key, err := parseSiblingMapKey(keyData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sibling map key: %w", err)
	}

	value, err := parseSiblingMapValue(valueData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sibling map value: %w", err)
	}

	return &siblingMapReader{
		key:    key,
		value:  value,
		endian: endian,
	}, nil
}

// parseSiblingMapKey parses raw bytes into a JSiblingMapKeyT structure
func parseSiblingMapKey(data []byte, endian binary.ByteOrder) (*types.JSiblingMapKeyT, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("insufficient data for sibling map key")
	}

	key := &types.JSiblingMapKeyT{}
	key.Hdr.ObjIdAndType = endian.Uint64(data[0:8])

	return key, nil
}

// parseSiblingMapValue parses raw bytes into a JSiblingMapValT structure
func parseSiblingMapValue(data []byte, endian binary.ByteOrder) (*types.JSiblingMapValT, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("insufficient data for sibling map value")
	}

	value := &types.JSiblingMapValT{}
	value.FileId = endian.Uint64(data[0:8])

	return value, nil
}

// SiblingID returns the unique identifier for this sibling map
func (smr *siblingMapReader) SiblingID() uint64 {
	return smr.key.Hdr.ObjIdAndType & types.ObjIdMask
}

// FileID returns the inode number of the underlying file
func (smr *siblingMapReader) FileID() uint64 {
	return smr.value.FileId
}
