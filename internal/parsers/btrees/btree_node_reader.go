package btrees

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/parsers/objects"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// btreeNodeReader implements the BTreeNodeReader interface
type btreeNodeReader struct {
	node   *types.BtreeNodePhysT
	data   []byte
	endian binary.ByteOrder
}

// NewBTreeNodeReader creates a new BTreeNodeReader implementation
func NewBTreeNodeReader(data []byte, endian binary.ByteOrder) (interfaces.BTreeNodeReader, error) {
	if len(data) < 56 { // Minimum size for btree_node_phys_t header
		return nil, fmt.Errorf("data too small for B-tree node: %d bytes", len(data))
	}

	node, err := parseBTreeNode(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse B-tree node: %w", err)
	}

	// Verify Fletcher-64 checksum for metadata integrity
	// APFS uses checksums for all metadata structures including B-tree nodes
	checksumVerifier := objects.NewChecksumInspector(&node.BtnO, data)
	if !checksumVerifier.VerifyChecksum() {
		return nil, fmt.Errorf("B-tree node checksum verification failed (OID: %d, XID: %d)", node.BtnO.OOid, node.BtnO.OXid)
	}

	return &btreeNodeReader{
		node:   node,
		data:   data,
		endian: endian,
	}, nil
}

// parseBTreeNode parses raw bytes into a BtreeNodePhysT structure
func parseBTreeNode(data []byte, endian binary.ByteOrder) (*types.BtreeNodePhysT, error) {
	if len(data) < 56 {
		return nil, fmt.Errorf("insufficient data for B-tree node header")
	}

	node := &types.BtreeNodePhysT{}

	// Parse object header (first 32 bytes based on obj_phys_t)
	copy(node.BtnO.OChecksum[:], data[0:8])
	node.BtnO.OOid = types.OidT(endian.Uint64(data[8:16]))
	node.BtnO.OXid = types.XidT(endian.Uint64(data[16:24]))
	node.BtnO.OType = endian.Uint32(data[24:28])
	node.BtnO.OSubtype = endian.Uint32(data[28:32])

	// Parse B-tree node specific fields
	node.BtnFlags = endian.Uint16(data[32:34])
	node.BtnLevel = endian.Uint16(data[34:36])
	node.BtnNkeys = endian.Uint32(data[36:40])

	// Parse table space location
	node.BtnTableSpace.Off = endian.Uint16(data[40:42])
	node.BtnTableSpace.Len = endian.Uint16(data[42:44])

	// Parse free space location
	node.BtnFreeSpace.Off = endian.Uint16(data[44:46])
	node.BtnFreeSpace.Len = endian.Uint16(data[46:48])

	// Parse key free list location
	node.BtnKeyFreeList.Off = endian.Uint16(data[48:50])
	node.BtnKeyFreeList.Len = endian.Uint16(data[50:52])

	// Parse value free list location
	node.BtnValFreeList.Off = endian.Uint16(data[52:54])
	node.BtnValFreeList.Len = endian.Uint16(data[54:56])

	// Store the node data (everything after the fixed header)
	if len(data) > 56 {
		node.BtnData = make([]byte, len(data)-56)
		copy(node.BtnData, data[56:])
	}

	return node, nil
}

// Flags returns the B-tree node's flags
func (br *btreeNodeReader) Flags() uint16 {
	return br.node.BtnFlags
}

// Level returns the number of child levels below this node
func (br *btreeNodeReader) Level() uint16 {
	return br.node.BtnLevel
}

// KeyCount returns the number of keys stored in this node
func (br *btreeNodeReader) KeyCount() uint32 {
	return br.node.BtnNkeys
}

// TableSpace returns the location of the table of contents
func (br *btreeNodeReader) TableSpace() types.NlocT {
	return br.node.BtnTableSpace
}

// FreeSpace returns the location of the shared free space for keys and values
func (br *btreeNodeReader) FreeSpace() types.NlocT {
	return br.node.BtnFreeSpace
}

// KeyFreeList returns the linked list that tracks free key space
func (br *btreeNodeReader) KeyFreeList() types.NlocT {
	return br.node.BtnKeyFreeList
}

// ValueFreeList returns the linked list that tracks free value space
func (br *btreeNodeReader) ValueFreeList() types.NlocT {
	return br.node.BtnValFreeList
}

// Data returns the node's storage area
func (br *btreeNodeReader) Data() []byte {
	return br.node.BtnData
}

// IsRoot checks if the node is a root node
func (br *btreeNodeReader) IsRoot() bool {
	return br.node.BtnFlags&types.BtnodeRoot != 0
}

// IsLeaf checks if the node is a leaf node
func (br *btreeNodeReader) IsLeaf() bool {
	return br.node.BtnFlags&types.BtnodeLeaf != 0
}

// HasFixedKVSize checks if the node has keys and values of fixed size
func (br *btreeNodeReader) HasFixedKVSize() bool {
	return br.node.BtnFlags&types.BtnodeFixedKvSize != 0
}

// IsHashed checks if the node contains child hashes
func (br *btreeNodeReader) IsHashed() bool {
	return br.node.BtnFlags&types.BtnodeHashed != 0
}

// HasHeader checks if the node is stored with an object header
func (br *btreeNodeReader) HasHeader() bool {
	return br.node.BtnFlags&types.BtnodeNoheader == 0
}
