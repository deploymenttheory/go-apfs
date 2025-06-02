package btrees

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// btreeInfoReader implements the BTreeInfoReader interface
type btreeInfoReader struct {
	info *types.BtreeInfoT
}

// NewBTreeInfoReader creates a new BTreeInfoReader implementation
func NewBTreeInfoReader(data []byte, endian binary.ByteOrder) (interfaces.BTreeInfoReader, error) {
	if len(data) < 48 { // Minimum size for btree_info_t
		return nil, fmt.Errorf("data too small for B-tree info: %d bytes", len(data))
	}

	info, err := parseBTreeInfo(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse B-tree info: %w", err)
	}

	return &btreeInfoReader{
		info: info,
	}, nil
}

// parseBTreeInfo parses raw bytes into a BtreeInfoT structure
func parseBTreeInfo(data []byte, endian binary.ByteOrder) (*types.BtreeInfoT, error) {
	if len(data) < 48 {
		return nil, fmt.Errorf("insufficient data for B-tree info")
	}

	info := &types.BtreeInfoT{}

	// Parse fixed B-tree info (first 16 bytes)
	info.BtFixed.BtFlags = endian.Uint32(data[0:4])
	info.BtFixed.BtNodeSize = endian.Uint32(data[4:8])
	info.BtFixed.BtKeySize = endian.Uint32(data[8:12])
	info.BtFixed.BtValSize = endian.Uint32(data[12:16])

	// Parse variable B-tree info (next 32 bytes)
	info.BtLongestKey = endian.Uint32(data[16:20])
	info.BtLongestVal = endian.Uint32(data[20:24])
	info.BtKeyCount = endian.Uint64(data[24:32])
	info.BtNodeCount = endian.Uint64(data[32:40])

	return info, nil
}

// Flags returns the B-tree's flags
func (bir *btreeInfoReader) Flags() uint32 {
	return bir.info.BtFixed.BtFlags
}

// NodeSize returns the on-disk size in bytes of a node in this B-tree
func (bir *btreeInfoReader) NodeSize() uint32 {
	return bir.info.BtFixed.BtNodeSize
}

// KeySize returns the size of a key, or zero if keys have variable size
func (bir *btreeInfoReader) KeySize() uint32 {
	return bir.info.BtFixed.BtKeySize
}

// ValueSize returns the size of a value, or zero if values have variable size
func (bir *btreeInfoReader) ValueSize() uint32 {
	return bir.info.BtFixed.BtValSize
}

// LongestKey returns the length in bytes of the longest key ever stored in the B-tree
func (bir *btreeInfoReader) LongestKey() uint32 {
	return bir.info.BtLongestKey
}

// LongestValue returns the length in bytes of the longest value ever stored in the B-tree
func (bir *btreeInfoReader) LongestValue() uint32 {
	return bir.info.BtLongestVal
}

// KeyCount returns the number of keys stored in the B-tree
func (bir *btreeInfoReader) KeyCount() uint64 {
	return bir.info.BtKeyCount
}

// NodeCount returns the number of nodes stored in the B-tree
func (bir *btreeInfoReader) NodeCount() uint64 {
	return bir.info.BtNodeCount
}

// HasUint64Keys checks if the B-tree uses 64-bit unsigned integer keys
func (bir *btreeInfoReader) HasUint64Keys() bool {
	return bir.info.BtFixed.BtFlags&types.BtreeUint64Keys != 0
}

// SupportsSequentialInsert checks if the B-tree is optimized for sequential insertions
func (bir *btreeInfoReader) SupportsSequentialInsert() bool {
	return bir.info.BtFixed.BtFlags&types.BtreeSequentialInsert != 0
}

// AllowsGhosts checks if the table of contents can contain keys with no corresponding value
func (bir *btreeInfoReader) AllowsGhosts() bool {
	return bir.info.BtFixed.BtFlags&types.BtreeAllowGhosts != 0
}

// IsEphemeral checks if the nodes use ephemeral object identifiers
func (bir *btreeInfoReader) IsEphemeral() bool {
	return bir.info.BtFixed.BtFlags&types.BtreeEphemeral != 0
}

// IsPhysical checks if the nodes use physical object identifiers
func (bir *btreeInfoReader) IsPhysical() bool {
	return bir.info.BtFixed.BtFlags&types.BtreePhysical != 0
}

// IsPersistent checks if the B-tree is persisted across unmounting
func (bir *btreeInfoReader) IsPersistent() bool {
	return bir.info.BtFixed.BtFlags&types.BtreeNonpersistent == 0
}

// HasAlignedKV checks if keys and values are aligned to eight-byte boundaries
func (bir *btreeInfoReader) HasAlignedKV() bool {
	return bir.info.BtFixed.BtFlags&types.BtreeKvNonaligned == 0
}

// IsHashed checks if nonleaf nodes store a hash of their child nodes
func (bir *btreeInfoReader) IsHashed() bool {
	return bir.info.BtFixed.BtFlags&types.BtreeHashed != 0
}

// HasHeaderlessNodes checks if nodes are stored without object headers
func (bir *btreeInfoReader) HasHeaderlessNodes() bool {
	return bir.info.BtFixed.BtFlags&types.BtreeNoheader != 0
}
