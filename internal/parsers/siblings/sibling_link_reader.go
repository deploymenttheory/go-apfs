package siblings

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// siblingLinkReader implements the SiblingLinkReader interface
type siblingLinkReader struct {
	key    *types.JSiblingKeyT
	value  *types.JSiblingValT
	endian binary.ByteOrder
}

// NewSiblingLinkReader creates a new SiblingLinkReader implementation
func NewSiblingLinkReader(keyData, valueData []byte, endian binary.ByteOrder) (interfaces.SiblingLinkReader, error) {
	if len(keyData) < 16 { // JKeyT (8 bytes) + SiblingId (8 bytes)
		return nil, fmt.Errorf("key data too small for sibling link key: %d bytes", len(keyData))
	}

	if len(valueData) < 10 { // ParentId (8 bytes) + NameLen (2 bytes) minimum
		return nil, fmt.Errorf("value data too small for sibling link value: %d bytes", len(valueData))
	}

	key, err := parseSiblingLinkKey(keyData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sibling link key: %w", err)
	}

	value, err := parseSiblingLinkValue(valueData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sibling link value: %w", err)
	}

	return &siblingLinkReader{
		key:    key,
		value:  value,
		endian: endian,
	}, nil
}

// parseSiblingLinkKey parses raw bytes into a JSiblingKeyT structure
func parseSiblingLinkKey(data []byte, endian binary.ByteOrder) (*types.JSiblingKeyT, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("insufficient data for sibling link key")
	}

	key := &types.JSiblingKeyT{}
	key.Hdr.ObjIdAndType = endian.Uint64(data[0:8])
	key.SiblingId = endian.Uint64(data[8:16])

	return key, nil
}

// parseSiblingLinkValue parses raw bytes into a JSiblingValT structure
func parseSiblingLinkValue(data []byte, endian binary.ByteOrder) (*types.JSiblingValT, error) {
	if len(data) < 10 {
		return nil, fmt.Errorf("insufficient data for sibling link value")
	}

	value := &types.JSiblingValT{}
	value.ParentId = endian.Uint64(data[0:8])
	value.NameLen = endian.Uint16(data[8:10])

	// Extract the name (variable length, includes null terminator)
	if len(data) < 10+int(value.NameLen) {
		return nil, fmt.Errorf("insufficient data for sibling link name: have %d bytes, need %d", len(data)-10, value.NameLen)
	}

	value.Name = make([]byte, value.NameLen)
	copy(value.Name, data[10:10+value.NameLen])

	return value, nil
}

// SiblingID returns the unique identifier for this sibling
func (slr *siblingLinkReader) SiblingID() uint64 {
	return slr.key.SiblingId
}

// InodeNumber returns the original inode number
func (slr *siblingLinkReader) InodeNumber() uint64 {
	return slr.key.Hdr.ObjIdAndType & types.ObjIdMask
}

// ParentDirectoryID returns the object identifier of the parent directory
func (slr *siblingLinkReader) ParentDirectoryID() uint64 {
	return slr.value.ParentId
}

// Name returns the name of the sibling link
func (slr *siblingLinkReader) Name() string {
	s := &Sibling{slr.value}
	return s.NameString()
}
