package snapshot

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// snapMetadataReader implements the SnapMetadataReader interface
type snapMetadataReader struct {
	key    *types.JSnapMetadataKeyT
	value  *types.JSnapMetadataValT
	endian binary.ByteOrder
}

// NewSnapMetadataReader creates a new SnapMetadataReader implementation
func NewSnapMetadataReader(keyData, valueData []byte, endian binary.ByteOrder) (interfaces.SnapMetadataReader, error) {
	if len(keyData) < 8 { // JKeyT (8 bytes)
		return nil, fmt.Errorf("key data too small for snapshot metadata key: %d bytes", len(keyData))
	}

	if len(valueData) < 38 { // ExtentrefTreeOid (8) + SblockOid (8) + CreateTime (8) + ChangeTime (8) + Inum (8) + ExtentrefTreeType (4) + Flags (4) + NameLen (2) minimum
		return nil, fmt.Errorf("value data too small for snapshot metadata value: %d bytes", len(valueData))
	}

	key, err := parseSnapMetadataKey(keyData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse snapshot metadata key: %w", err)
	}

	value, err := parseSnapMetadataValue(valueData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse snapshot metadata value: %w", err)
	}

	return &snapMetadataReader{
		key:    key,
		value:  value,
		endian: endian,
	}, nil
}

// parseSnapMetadataKey parses raw bytes into a JSnapMetadataKeyT structure
func parseSnapMetadataKey(data []byte, endian binary.ByteOrder) (*types.JSnapMetadataKeyT, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("insufficient data for snapshot metadata key")
	}

	key := &types.JSnapMetadataKeyT{}
	key.Hdr.ObjIdAndType = endian.Uint64(data[0:8])

	return key, nil
}

// parseSnapMetadataValue parses raw bytes into a JSnapMetadataValT structure
func parseSnapMetadataValue(data []byte, endian binary.ByteOrder) (*types.JSnapMetadataValT, error) {
	if len(data) < 38 {
		return nil, fmt.Errorf("insufficient data for snapshot metadata value")
	}

	value := &types.JSnapMetadataValT{}
	value.ExtentrefTreeOid = types.OidT(endian.Uint64(data[0:8]))
	value.SblockOid = types.OidT(endian.Uint64(data[8:16]))
	value.CreateTime = endian.Uint64(data[16:24])
	value.ChangeTime = endian.Uint64(data[24:32])
	value.Inum = endian.Uint64(data[32:40])
	value.ExtentrefTreeType = endian.Uint32(data[40:44])
	value.Flags = endian.Uint32(data[44:48])
	value.NameLen = endian.Uint16(data[48:50])

	// Extract the name (variable length, includes null terminator)
	if len(data) < 50+int(value.NameLen) {
		return nil, fmt.Errorf("insufficient data for snapshot name: have %d bytes, need %d", len(data)-50, value.NameLen)
	}

	value.Name = make([]byte, value.NameLen)
	copy(value.Name, data[50:50+value.NameLen])

	return value, nil
}

// ExtentRefTreeOID returns the physical object identifier of the B-tree that stores extents information.
func (smr *snapMetadataReader) ExtentRefTreeOID() uint64 {
	return uint64(smr.value.ExtentrefTreeOid)
}

// SuperblockOID returns the physical object identifier of the volume superblock.
func (smr *snapMetadataReader) SuperblockOID() uint64 {
	return uint64(smr.value.SblockOid)
}

// CreateTime returns the time the snapshot was created.
func (smr *snapMetadataReader) CreateTime() time.Time {
	return time.Unix(0, int64(smr.value.CreateTime))
}

// ChangeTime returns the last time the snapshot was modified.
func (smr *snapMetadataReader) ChangeTime() time.Time {
	return time.Unix(0, int64(smr.value.ChangeTime))
}

// InodeNumber returns the inode number associated with the snapshot.
func (smr *snapMetadataReader) InodeNumber() uint64 {
	return smr.value.Inum
}

// ExtentRefTreeType returns the type of the B-tree that stores extents information.
func (smr *snapMetadataReader) ExtentRefTreeType() uint32 {
	return smr.value.ExtentrefTreeType
}

// Flags returns the snapshot metadata flags.
func (smr *snapMetadataReader) Flags() uint32 {
	return smr.value.Flags
}

// Name returns the snapshot's name.
func (smr *snapMetadataReader) Name() string {
	return strings.TrimRight(string(smr.value.Name), "\x00")
}

// HasFlag checks if a specific flag is set.
func (smr *snapMetadataReader) HasFlag(flag uint32) bool {
	return (smr.value.Flags & flag) != 0
}
