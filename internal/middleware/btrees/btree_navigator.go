package btrees

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// btreeNavigator implements the BTreeNavigator interface
type btreeNavigator struct {
	blockReader interfaces.BlockDeviceReader
	rootOID     types.OidT
	btreeInfo   interfaces.BTreeInfoReader
	nodeCache   map[types.OidT]interfaces.BTreeNodeReader
}

// NewBTreeNavigator creates a new BTreeNavigator implementation
func NewBTreeNavigator(blockReader interfaces.BlockDeviceReader, rootOID types.OidT, btreeInfo interfaces.BTreeInfoReader) interfaces.BTreeNavigator {
	return &btreeNavigator{
		blockReader: blockReader,
		rootOID:     rootOID,
		btreeInfo:   btreeInfo,
		nodeCache:   make(map[types.OidT]interfaces.BTreeNodeReader),
	}
}

// GetRootNode returns the root node of the B-tree
func (nav *btreeNavigator) GetRootNode() (interfaces.BTreeNodeReader, error) {
	return nav.GetNodeByObjectID(nav.rootOID)
}

// GetChildNode returns a child node of the given parent node at the specified index
func (nav *btreeNavigator) GetChildNode(parent interfaces.BTreeNodeReader, index int) (interfaces.BTreeNodeReader, error) {
	if parent.IsLeaf() {
		return nil, fmt.Errorf("cannot get child of leaf node")
	}

	// Internal nodes have KeyCount() + 1 children
	maxChildIndex := int(parent.KeyCount()) + 1
	if index < 0 || index >= maxChildIndex {
		return nil, fmt.Errorf("child index %d out of range [0, %d)", index, maxChildIndex)
	}

	// Extract child OID from parent node data
	childOID, err := nav.extractChildOID(parent, index)
	if err != nil {
		return nil, fmt.Errorf("failed to extract child OID: %w", err)
	}

	return nav.GetNodeByObjectID(childOID)
}

// GetNodeByObjectID returns a node with the specified object identifier
func (nav *btreeNavigator) GetNodeByObjectID(objectID types.OidT) (interfaces.BTreeNodeReader, error) {
	// Check cache first
	if node, exists := nav.nodeCache[objectID]; exists {
		return node, nil
	}

	// Read node from block device
	blockAddr := types.Paddr(objectID) // Convert OID to physical address
	blockData, err := nav.blockReader.ReadBlock(blockAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to read block at address %d: %w", blockAddr, err)
	}

	// Create node reader
	node, err := NewBTreeNodeReader(blockData, nav.getEndianness())
	if err != nil {
		return nil, fmt.Errorf("failed to create node reader: %w", err)
	}

	// Cache the node
	nav.nodeCache[objectID] = node

	return node, nil
}

// GetHeight returns the height of the B-tree
func (nav *btreeNavigator) GetHeight() (uint16, error) {
	rootNode, err := nav.GetRootNode()
	if err != nil {
		return 0, fmt.Errorf("failed to get root node: %w", err)
	}

	// Height is the level of the root node plus 1
	return rootNode.Level() + 1, nil
}

// extractChildOID extracts the child OID from a parent node at the given index
func (nav *btreeNavigator) extractChildOID(parent interfaces.BTreeNodeReader, index int) (types.OidT, error) {
	nodeData := parent.Data()

	if parent.HasFixedKVSize() {
		return nav.extractChildOIDFixed(nodeData, index)
	}

	return nav.extractChildOIDVariable(nodeData, index)
}

// extractChildOIDFixed extracts child OID from fixed-size key/value node
func (nav *btreeNavigator) extractChildOIDFixed(nodeData []byte, index int) (types.OidT, error) {
	keySize := nav.btreeInfo.KeySize()
	valueSize := nav.btreeInfo.ValueSize()
	entrySize := keySize + valueSize

	if entrySize == 0 {
		return 0, fmt.Errorf("invalid entry size: key=%d, value=%d", keySize, valueSize)
	}

	// Data is already without the header, so we start at offset 0
	offset := int(entrySize) * index

	// Debug information
	if offset+int(valueSize) > len(nodeData) {
		return 0, fmt.Errorf("entry %d extends beyond node data: offset=%d, valueSize=%d, nodeDataLen=%d, keySize=%d, entrySize=%d",
			index, offset, valueSize, len(nodeData), keySize, entrySize)
	}

	// Skip key, read value (which should contain the child OID)
	valueOffset := offset + int(keySize)
	if valueSize >= 8 {
		// Assume child OID is at the beginning of the value
		childOID := nav.getEndianness().Uint64(nodeData[valueOffset : valueOffset+8])
		return types.OidT(childOID), nil
	}

	return 0, fmt.Errorf("value size too small to contain OID: %d", valueSize)
}

// extractChildOIDVariable extracts child OID from variable-size key/value node
func (nav *btreeNavigator) extractChildOIDVariable(nodeData []byte, index int) (types.OidT, error) {
	// For variable-size entries, we need to parse the table of contents
	// This is a simplified implementation - real APFS would need more complex parsing

	if len(nodeData) < 8 {
		return 0, fmt.Errorf("insufficient node data for variable entry")
	}

	// This is a placeholder implementation
	// In real APFS, we'd need to:
	// 1. Parse the table of contents to find entry locations
	// 2. Extract the key/value at the specified index
	// 3. Parse the value to get the child OID

	return 0, fmt.Errorf("variable-size key/value extraction not yet implemented")
}

// getEndianness returns the byte order for this platform
func (nav *btreeNavigator) getEndianness() binary.ByteOrder {
	// APFS uses little-endian on all supported platforms
	return binary.LittleEndian
}

// ClearCache clears the node cache
func (nav *btreeNavigator) ClearCache() {
	nav.nodeCache = make(map[types.OidT]interfaces.BTreeNodeReader)
}

// GetCacheSize returns the number of cached nodes
func (nav *btreeNavigator) GetCacheSize() int {
	return len(nav.nodeCache)
}
