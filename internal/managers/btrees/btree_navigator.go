package btrees

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/parsers/btrees"
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
	node, err := btrees.NewBTreeNodeReader(blockData, nav.getEndianness())
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
	// to find the key/value locations, then extract the child OID from the value

	if len(nodeData) < 8 {
		return 0, fmt.Errorf("insufficient node data for variable entry")
	}

	// Each TOC entry is typically 4 bytes (2 bytes key offset + 2 bytes value offset)
	// or 8 bytes for nodes with aligned KV (4 bytes each)
	tocEntrySize := 4
	if nav.btreeInfo.HasAlignedKV() {
		tocEntrySize = 8 // Aligned nodes use larger offsets
	}

	// Calculate offset for the TOC entry at the given index
	tocOffset := index * tocEntrySize
	if tocOffset+tocEntrySize > len(nodeData) {
		return 0, fmt.Errorf("TOC entry %d extends beyond node data", index)
	}

	endian := nav.getEndianness()

	var keyOffset, valueOffset uint16
	if tocEntrySize == 4 {
		// Standard 4-byte TOC entries
		keyOffset = endian.Uint16(nodeData[tocOffset : tocOffset+2])
		valueOffset = endian.Uint16(nodeData[tocOffset+2 : tocOffset+4])
	} else {
		// 8-byte aligned TOC entries
		keyOffset = uint16(endian.Uint32(nodeData[tocOffset : tocOffset+4]))
		valueOffset = uint16(endian.Uint32(nodeData[tocOffset+4 : tocOffset+8]))
	}

	// Validate offsets are within node bounds
	if int(keyOffset) >= len(nodeData) || int(valueOffset) >= len(nodeData) {
		return 0, fmt.Errorf("invalid TOC entry %d: key offset %d or value offset %d exceeds node size %d",
			index, keyOffset, valueOffset, len(nodeData))
	}

	// For internal nodes, the value should contain the child OID
	// Calculate value length - we need to determine where this value ends
	var valueLength int

	// Check if this is the last entry by seeing if there's another TOC entry
	nextTocOffset := (index + 1) * tocEntrySize
	if nextTocOffset+tocEntrySize <= len(nodeData) {
		// Not the last entry, calculate length from next entry
		var nextValueOffset uint16
		if tocEntrySize == 4 {
			nextValueOffset = endian.Uint16(nodeData[nextTocOffset+2 : nextTocOffset+4])
		} else {
			nextValueOffset = uint16(endian.Uint32(nodeData[nextTocOffset+4 : nextTocOffset+8]))
		}
		valueLength = int(nextValueOffset - valueOffset)
	} else {
		// Last entry, value extends to end of node data
		valueLength = len(nodeData) - int(valueOffset)
	}

	// Ensure value is large enough to contain an OID (8 bytes)
	if valueLength < 8 {
		return 0, fmt.Errorf("value at index %d too small to contain OID: %d bytes", index, valueLength)
	}

	// Extract the child OID from the beginning of the value
	if int(valueOffset)+8 > len(nodeData) {
		return 0, fmt.Errorf("child OID at index %d extends beyond node data", index)
	}

	childOID := endian.Uint64(nodeData[valueOffset : valueOffset+8])
	return types.OidT(childOID), nil
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
