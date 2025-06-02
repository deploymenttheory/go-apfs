package file_system_objects

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// directoryStatsReader implements directory statistics parsing
type directoryStatsReader struct {
	key   *types.JDirStatsKeyT
	value *types.JDirStatsValT
}

// NewDirectoryStatsReader creates a new directory stats reader
func NewDirectoryStatsReader(keyData, valueData []byte, endian binary.ByteOrder) (interfaces.DirectoryStatsReader, error) {
	key, err := parseDirStatsKey(keyData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dir stats key: %w", err)
	}

	value, err := parseDirStatsValue(valueData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dir stats value: %w", err)
	}

	return &directoryStatsReader{
		key:   key,
		value: value,
	}, nil
}

// parseDirStatsKey parses raw bytes into a JDirStatsKeyT structure
func parseDirStatsKey(data []byte, endian binary.ByteOrder) (*types.JDirStatsKeyT, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("insufficient data for dir stats key")
	}

	key := &types.JDirStatsKeyT{}
	key.Hdr.ObjIdAndType = endian.Uint64(data[0:8])

	return key, nil
}

// parseDirStatsValue parses raw bytes into a JDirStatsValT structure
func parseDirStatsValue(data []byte, endian binary.ByteOrder) (*types.JDirStatsValT, error) {
	minSize := 8 + 8 + 8 + 8 // num_children + total_size + chained_key + gen_count
	if len(data) < minSize {
		return nil, fmt.Errorf("insufficient data for dir stats value")
	}

	value := &types.JDirStatsValT{}
	offset := 0

	value.NumChildren = endian.Uint64(data[offset : offset+8])
	offset += 8

	value.TotalSize = endian.Uint64(data[offset : offset+8])
	offset += 8

	value.ChainedKey = endian.Uint64(data[offset : offset+8])
	offset += 8

	value.GenCount = endian.Uint64(data[offset : offset+8])

	return value, nil
}

// ObjectIdentifier returns the object's unique identifier
func (dsr *directoryStatsReader) ObjectIdentifier() uint64 {
	return dsr.key.Hdr.ObjIdAndType & types.ObjIdMask
}

// ObjectType returns the type of the file system object
func (dsr *directoryStatsReader) ObjectType() types.JObjTypes {
	return types.JObjTypes((dsr.key.Hdr.ObjIdAndType & types.ObjTypeMask) >> types.ObjTypeShift)
}

// NumChildren returns the number of files and folders in the directory
func (dsr *directoryStatsReader) NumChildren() uint64 {
	return dsr.value.NumChildren
}

// TotalSize returns the total size of all files in the directory and descendants
func (dsr *directoryStatsReader) TotalSize() uint64 {
	return dsr.value.TotalSize
}

// ChainedKey returns the parent directory's file system object identifier
func (dsr *directoryStatsReader) ChainedKey() uint64 {
	return dsr.value.ChainedKey
}

// GenCount returns the generation counter
func (dsr *directoryStatsReader) GenCount() uint64 {
	return dsr.value.GenCount
}

// ObjectKind returns the kind of the file system object
func (dsr *directoryStatsReader) ObjectKind() types.JObjKinds {
	return types.ApfsKindNew // Default for now
}
