package file_system_objects

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// directoryEntryReader implements the DirectoryEntryReader interface
type directoryEntryReader struct {
	key   interface{} // Can be *types.JDrecKeyT or *types.JDrecHashedKeyT
	value *types.JDrecValT
}

// NewDirectoryEntryReader creates a new directory entry reader
func NewDirectoryEntryReader(keyData, valueData []byte, endian binary.ByteOrder, isHashed bool) (interfaces.DirectoryEntryReader, error) {
	var key interface{}
	var err error

	if isHashed {
		key, err = parseHashedDirectoryKey(keyData, endian)
	} else {
		key, err = parseDirectoryKey(keyData, endian)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse directory key: %w", err)
	}

	value, err := parseDirectoryValue(valueData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse directory value: %w", err)
	}

	return &directoryEntryReader{
		key:   key,
		value: value,
	}, nil
}

// parseDirectoryKey parses raw bytes into a JDrecKeyT structure
func parseDirectoryKey(data []byte, endian binary.ByteOrder) (*types.JDrecKeyT, error) {
	if len(data) < 10 { // 8 bytes header + 2 bytes name_len minimum
		return nil, fmt.Errorf("insufficient data for directory key")
	}

	key := &types.JDrecKeyT{}
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

// parseHashedDirectoryKey parses raw bytes into a JDrecHashedKeyT structure
func parseHashedDirectoryKey(data []byte, endian binary.ByteOrder) (*types.JDrecHashedKeyT, error) {
	if len(data) < 12 { // 8 bytes header + 4 bytes name_len_and_hash minimum
		return nil, fmt.Errorf("insufficient data for hashed directory key")
	}

	key := &types.JDrecHashedKeyT{}
	offset := 0

	key.Hdr.ObjIdAndType = endian.Uint64(data[offset : offset+8])
	offset += 8

	key.NameLenAndHash = endian.Uint32(data[offset : offset+4])
	offset += 4

	nameLen := key.NameLenAndHash & types.JDrecLenMask

	if offset+int(nameLen) > len(data) {
		return nil, fmt.Errorf("name length exceeds available data")
	}

	key.Name = make([]byte, nameLen)
	copy(key.Name, data[offset:offset+int(nameLen)])

	return key, nil
}

// parseDirectoryValue parses raw bytes into a JDrecValT structure
func parseDirectoryValue(data []byte, endian binary.ByteOrder) (*types.JDrecValT, error) {
	minSize := 8 + 8 + 2 // file_id + date_added + flags
	if len(data) < minSize {
		return nil, fmt.Errorf("insufficient data for directory value")
	}

	value := &types.JDrecValT{}
	offset := 0

	value.FileId = endian.Uint64(data[offset : offset+8])
	offset += 8

	value.DateAdded = endian.Uint64(data[offset : offset+8])
	offset += 8

	value.Flags = endian.Uint16(data[offset : offset+2])
	offset += 2

	// Parse extended fields (variable length)
	if offset < len(data) {
		value.XFields = make([]byte, len(data)-offset)
		copy(value.XFields, data[offset:])
	}

	return value, nil
}

// FileSystemObjectReader interface methods

func (der *directoryEntryReader) ObjectType() types.JObjTypes {
	switch k := der.key.(type) {
	case *types.JDrecKeyT:
		return types.JObjTypes((k.Hdr.ObjIdAndType & types.ObjTypeMask) >> types.ObjTypeShift)
	case *types.JDrecHashedKeyT:
		return types.JObjTypes((k.Hdr.ObjIdAndType & types.ObjTypeMask) >> types.ObjTypeShift)
	default:
		return types.ApfsTypeInvalid
	}
}

func (der *directoryEntryReader) ObjectKind() types.JObjKinds {
	return types.ApfsKindNew // Default for now
}

func (der *directoryEntryReader) ObjectIdentifier() uint64 {
	switch k := der.key.(type) {
	case *types.JDrecKeyT:
		return k.Hdr.ObjIdAndType & types.ObjIdMask
	case *types.JDrecHashedKeyT:
		return k.Hdr.ObjIdAndType & types.ObjIdMask
	default:
		return 0
	}
}

// DirectoryEntryReader interface methods

func (der *directoryEntryReader) FileName() string {
	switch k := der.key.(type) {
	case *types.JDrecKeyT:
		// Remove null terminator if present
		name := string(k.Name)
		return strings.TrimRight(name, "\x00")
	case *types.JDrecHashedKeyT:
		// Remove null terminator if present
		name := string(k.Name)
		return strings.TrimRight(name, "\x00")
	default:
		return ""
	}
}

func (der *directoryEntryReader) FileID() uint64 {
	return der.value.FileId
}

func (der *directoryEntryReader) DateAdded() time.Time {
	return time.Unix(0, int64(der.value.DateAdded))
}

func (der *directoryEntryReader) FileType() uint16 {
	return der.value.Flags
}
