package file_system_objects

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// inodeReader implements the InodeReader interface
type inodeReader struct {
	key   *types.JInodeKeyT
	value *types.JInodeValT
}

// NewInodeReader creates a new inode reader
func NewInodeReader(keyData, valueData []byte, endian binary.ByteOrder) (interfaces.InodeReader, error) {
	key, err := parseInodeKey(keyData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inode key: %w", err)
	}

	value, err := parseInodeValue(valueData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inode value: %w", err)
	}

	return &inodeReader{
		key:   key,
		value: value,
	}, nil
}

// parseInodeKey parses raw bytes into a JInodeKeyT structure
func parseInodeKey(data []byte, endian binary.ByteOrder) (*types.JInodeKeyT, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("insufficient data for inode key")
	}

	key := &types.JInodeKeyT{}
	key.Hdr.ObjIdAndType = endian.Uint64(data[0:8])

	return key, nil
}

// parseInodeValue parses raw bytes into a JInodeValT structure
func parseInodeValue(data []byte, endian binary.ByteOrder) (*types.JInodeValT, error) {
	minSize := 8 + 8 + 8 + 8 + 8 + 8 + 8 + 4 + 4 + 4 + 4 + 4 + 4 + 2 + 2 + 8 // 98 bytes minimum
	if len(data) < minSize {
		return nil, fmt.Errorf("insufficient data for inode value: %d bytes", len(data))
	}

	value := &types.JInodeValT{}
	offset := 0

	// Parse fixed fields according to JInodeValT structure
	value.ParentId = endian.Uint64(data[offset : offset+8])
	offset += 8

	value.PrivateId = endian.Uint64(data[offset : offset+8])
	offset += 8

	value.CreateTime = endian.Uint64(data[offset : offset+8])
	offset += 8

	value.ModTime = endian.Uint64(data[offset : offset+8])
	offset += 8

	value.ChangeTime = endian.Uint64(data[offset : offset+8])
	offset += 8

	value.AccessTime = endian.Uint64(data[offset : offset+8])
	offset += 8

	value.InternalFlags = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse union field (nchildren/nlink)
	value.NchildrenOrNlink = int32(endian.Uint32(data[offset : offset+4]))
	offset += 4

	value.DefaultProtectionClass = types.CpKeyClassT(endian.Uint32(data[offset : offset+4]))
	offset += 4

	value.WriteGenerationCounter = endian.Uint32(data[offset : offset+4])
	offset += 4

	value.BsdFlags = endian.Uint32(data[offset : offset+4])
	offset += 4

	value.Owner = types.UidT(endian.Uint32(data[offset : offset+4]))
	offset += 4

	value.Group = types.GidT(endian.Uint32(data[offset : offset+4]))
	offset += 4

	value.Mode = types.ModeT(endian.Uint16(data[offset : offset+2]))
	offset += 2

	value.Pad1 = endian.Uint16(data[offset : offset+2])
	offset += 2

	value.UncompressedSize = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse extended fields (variable length)
	if offset < len(data) {
		value.XFields = make([]byte, len(data)-offset)
		copy(value.XFields, data[offset:])
	}

	return value, nil
}

// FileSystemObjectReader interface methods

func (ir *inodeReader) ObjectType() types.JObjTypes {
	return types.JObjTypes((ir.key.Hdr.ObjIdAndType & types.ObjTypeMask) >> types.ObjTypeShift)
}

func (ir *inodeReader) ObjectKind() types.JObjKinds {
	return types.ApfsKindNew // Default for now
}

func (ir *inodeReader) ObjectIdentifier() uint64 {
	return ir.key.Hdr.ObjIdAndType & types.ObjIdMask
}

// InodeReader interface methods

func (ir *inodeReader) ParentID() uint64 {
	return ir.value.ParentId
}

func (ir *inodeReader) PrivateID() uint64 {
	return ir.value.PrivateId
}

func (ir *inodeReader) CreationTime() time.Time {
	return time.Unix(0, int64(ir.value.CreateTime))
}

func (ir *inodeReader) ModificationTime() time.Time {
	return time.Unix(0, int64(ir.value.ModTime))
}

func (ir *inodeReader) ChangeTime() time.Time {
	return time.Unix(0, int64(ir.value.ChangeTime))
}

func (ir *inodeReader) AccessTime() time.Time {
	return time.Unix(0, int64(ir.value.AccessTime))
}

func (ir *inodeReader) Flags() types.JInodeFlags {
	return types.JInodeFlags(ir.value.InternalFlags)
}

func (ir *inodeReader) Owner() types.UidT {
	return ir.value.Owner
}

func (ir *inodeReader) Group() types.GidT {
	return ir.value.Group
}

func (ir *inodeReader) Mode() types.ModeT {
	return ir.value.Mode
}

func (ir *inodeReader) IsDirectory() bool {
	// Check mode bits for directory flag
	return ir.value.Mode&types.ModeT(types.ModeIFDIR) != 0
}

func (ir *inodeReader) NumberOfChildren() int32 {
	if ir.IsDirectory() {
		return ir.value.NchildrenOrNlink
	}
	return 0
}

func (ir *inodeReader) NumberOfHardLinks() int32 {
	if !ir.IsDirectory() {
		return ir.value.NchildrenOrNlink
	}
	return 0
}

func (ir *inodeReader) HasResourceFork() bool {
	// This would need the actual flag constant from the types
	// For now, returning false as placeholder
	return false
}

func (ir *inodeReader) Size() uint64 {
	return ir.value.UncompressedSize
}
