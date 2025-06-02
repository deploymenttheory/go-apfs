package file_system_objects

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// extendedAttributeReader implements the ExtendedAttributeReader interface
type extendedAttributeReader struct {
	key   *types.JXattrKeyT
	value *types.JXattrValT
}

// NewExtendedAttributeReader creates a new extended attribute reader
func NewExtendedAttributeReader(keyData, valueData []byte, endian binary.ByteOrder) (interfaces.ExtendedAttributeReader, error) {
	key, err := parseXattrKey(keyData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse xattr key: %w", err)
	}

	value, err := parseXattrValue(valueData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse xattr value: %w", err)
	}

	return &extendedAttributeReader{
		key:   key,
		value: value,
	}, nil
}

// parseXattrKey parses raw bytes into a JXattrKeyT structure
func parseXattrKey(data []byte, endian binary.ByteOrder) (*types.JXattrKeyT, error) {
	if len(data) < 10 { // 8 bytes header + 2 bytes name_len minimum
		return nil, fmt.Errorf("insufficient data for xattr key")
	}

	key := &types.JXattrKeyT{}
	offset := 0

	key.Hdr.ObjIdAndType = endian.Uint64(data[offset : offset+8])
	offset += 8

	key.NameLen = endian.Uint16(data[offset : offset+2])
	offset += 2

	if offset+int(key.NameLen) > len(data) {
		return nil, fmt.Errorf("name length exceeds available data")
	}

	key.Name = make([]byte, key.NameLen)
	copy(key.Name, data[offset:offset+int(key.NameLen)])

	return key, nil
}

// parseXattrValue parses raw bytes into a JXattrValT structure
func parseXattrValue(data []byte, endian binary.ByteOrder) (*types.JXattrValT, error) {
	minSize := 2 + 2 // flags + xdata_len
	if len(data) < minSize {
		return nil, fmt.Errorf("insufficient data for xattr value")
	}

	value := &types.JXattrValT{}
	offset := 0

	value.Flags = endian.Uint16(data[offset : offset+2])
	offset += 2

	value.XdataLen = endian.Uint16(data[offset : offset+2])
	offset += 2

	// Parse xdata (variable length)
	if offset < len(data) {
		value.Xdata = make([]byte, len(data)-offset)
		copy(value.Xdata, data[offset:])
	}

	return value, nil
}

// FileSystemObjectReader interface methods

func (ear *extendedAttributeReader) ObjectType() types.JObjTypes {
	return types.JObjTypes((ear.key.Hdr.ObjIdAndType & types.ObjTypeMask) >> types.ObjTypeShift)
}

func (ear *extendedAttributeReader) ObjectKind() types.JObjKinds {
	return types.ApfsKindNew // Default for now
}

func (ear *extendedAttributeReader) ObjectIdentifier() uint64 {
	return ear.key.Hdr.ObjIdAndType & types.ObjIdMask
}

// ExtendedAttributeReader interface methods

func (ear *extendedAttributeReader) AttributeName() string {
	// Remove null terminator if present
	name := string(ear.key.Name)
	return strings.TrimRight(name, "\x00")
}

func (ear *extendedAttributeReader) IsDataEmbedded() bool {
	return types.JXattrFlags(ear.value.Flags)&types.XattrDataEmbedded != 0
}

func (ear *extendedAttributeReader) IsDataStream() bool {
	return types.JXattrFlags(ear.value.Flags)&types.XattrDataStream != 0
}

func (ear *extendedAttributeReader) IsFileSystemOwned() bool {
	return types.JXattrFlags(ear.value.Flags)&types.XattrFileSystemOwned != 0
}

func (ear *extendedAttributeReader) Data() []byte {
	if ear.IsDataEmbedded() {
		// Return embedded data limited by xdata_len
		if int(ear.value.XdataLen) <= len(ear.value.Xdata) {
			return ear.value.Xdata[:ear.value.XdataLen]
		}
		return ear.value.Xdata
	}
	// For stream data, this would contain the stream ID
	return ear.value.Xdata
}
