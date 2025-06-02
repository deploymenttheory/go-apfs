package file_system_objects

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// jKeyReader implements file system object key parsing
type jKeyReader struct {
	key *types.JKeyT
}

// NewJKeyReader creates a new file system key reader
func NewJKeyReader(data []byte, endian binary.ByteOrder) (*jKeyReader, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("data too small for J-key: %d bytes", len(data))
	}

	key, err := parseJKey(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse J-key: %w", err)
	}

	return &jKeyReader{
		key: key,
	}, nil
}

// parseJKey parses raw bytes into a JKeyT structure
func parseJKey(data []byte, endian binary.ByteOrder) (*types.JKeyT, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("insufficient data for J-key")
	}

	key := &types.JKeyT{}
	key.ObjIdAndType = endian.Uint64(data[0:8])

	return key, nil
}

// ObjectIdentifier returns the object's unique identifier
func (jr *jKeyReader) ObjectIdentifier() uint64 {
	return jr.key.ObjIdAndType & types.ObjIdMask
}

// ObjectType returns the type of the file system object
func (jr *jKeyReader) ObjectType() types.JObjTypes {
	return types.JObjTypes((jr.key.ObjIdAndType & types.ObjTypeMask) >> types.ObjTypeShift)
}

// RawObjIdAndType returns the raw combined field
func (jr *jKeyReader) RawObjIdAndType() uint64 {
	return jr.key.ObjIdAndType
}
