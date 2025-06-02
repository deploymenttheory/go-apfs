package btrees

import (
	"encoding/binary"
	"fmt"
	"testing"

	parser "github.com/deploymenttheory/go-apfs/internal/parsers/btrees"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// MockBlockDeviceReader implements the BlockDeviceReader interface for testing
type MockBlockDeviceReader struct {
	blocks map[types.Paddr][]byte
}

func NewMockBlockDeviceReader() *MockBlockDeviceReader {
	return &MockBlockDeviceReader{
		blocks: make(map[types.Paddr][]byte),
	}
}

func (m *MockBlockDeviceReader) ReadBlock(address types.Paddr) ([]byte, error) {
	if data, exists := m.blocks[address]; exists {
		return data, nil
	}
	return nil, fmt.Errorf("block not found at address %d", address)
}

func (m *MockBlockDeviceReader) ReadBlockRange(start types.Paddr, count uint32) ([]byte, error) {
	var result []byte
	for i := uint32(0); i < count; i++ {
		block, err := m.ReadBlock(start + types.Paddr(i))
		if err != nil {
			return nil, err
		}
		result = append(result, block...)
	}
	return result, nil
}

func (m *MockBlockDeviceReader) ReadBytes(address types.Paddr, offset uint32, length uint32) ([]byte, error) {
	block, err := m.ReadBlock(address)
	if err != nil {
		return nil, err
	}
	if int(offset+length) > len(block) {
		return nil, fmt.Errorf("read beyond block boundary")
	}
	return block[offset : offset+length], nil
}

func (m *MockBlockDeviceReader) BlockSize() uint32 {
	return 4096
}

func (m *MockBlockDeviceReader) TotalBlocks() uint64 {
	return uint64(len(m.blocks))
}

func (m *MockBlockDeviceReader) TotalSize() uint64 {
	return uint64(len(m.blocks)) * uint64(m.BlockSize())
}

func (m *MockBlockDeviceReader) IsValidAddress(address types.Paddr) bool {
	_, exists := m.blocks[address]
	return exists
}

func (m *MockBlockDeviceReader) CanReadRange(start types.Paddr, count uint32) bool {
	for i := uint32(0); i < count; i++ {
		if !m.IsValidAddress(start + types.Paddr(i)) {
			return false
		}
	}
	return true
}

func (m *MockBlockDeviceReader) SetBlock(address types.Paddr, data []byte) {
	m.blocks[address] = data
}

// MockBTreeInfoReader implements the BTreeInfoReader interface for testing
type MockBTreeInfoReader struct {
	nodeSize   uint32
	keySize    uint32
	valueSize  uint32
	flags      uint32
	keyCount   uint64
	nodeCount  uint64
	longestKey uint32
	longestVal uint32
}

func NewMockBTreeInfoReader() *MockBTreeInfoReader {
	return &MockBTreeInfoReader{
		nodeSize:   4096,
		keySize:    8,
		valueSize:  8,
		flags:      0,
		keyCount:   100,
		nodeCount:  10,
		longestKey: 8,
		longestVal: 8,
	}
}

func (m *MockBTreeInfoReader) Flags() uint32                  { return m.flags }
func (m *MockBTreeInfoReader) NodeSize() uint32               { return m.nodeSize }
func (m *MockBTreeInfoReader) KeySize() uint32                { return m.keySize }
func (m *MockBTreeInfoReader) ValueSize() uint32              { return m.valueSize }
func (m *MockBTreeInfoReader) LongestKey() uint32             { return m.longestKey }
func (m *MockBTreeInfoReader) LongestValue() uint32           { return m.longestVal }
func (m *MockBTreeInfoReader) KeyCount() uint64               { return m.keyCount }
func (m *MockBTreeInfoReader) NodeCount() uint64              { return m.nodeCount }
func (m *MockBTreeInfoReader) HasUint64Keys() bool            { return false }
func (m *MockBTreeInfoReader) SupportsSequentialInsert() bool { return false }
func (m *MockBTreeInfoReader) AllowsGhosts() bool             { return false }
func (m *MockBTreeInfoReader) IsEphemeral() bool              { return false }
func (m *MockBTreeInfoReader) IsPhysical() bool               { return true }
func (m *MockBTreeInfoReader) IsPersistent() bool             { return true }
func (m *MockBTreeInfoReader) HasAlignedKV() bool             { return true }
func (m *MockBTreeInfoReader) IsHashed() bool                 { return false }
func (m *MockBTreeInfoReader) HasHeaderlessNodes() bool       { return false }

// Helper function to create test B-tree node data
func createTestNavigatorNodeData(oid types.OidT, flags uint16, level uint16, nkeys uint32, isFixedKV bool) []byte {
	// Calculate buffer size: header (56 bytes) + data
	// For internal nodes, we need space for nkeys + 1 children (each taking 16 bytes)
	// For leaf nodes, we need space for nkeys entries
	var bufferSize int
	if level > 0 && isFixedKV {
		bufferSize = 56 + int(nkeys+1)*16 // Internal node: nkeys + 1 children
	} else {
		bufferSize = 56 + int(nkeys)*16 // Leaf node: nkeys entries
	}
	data := make([]byte, bufferSize)
	endian := binary.LittleEndian

	// Object header (32 bytes)
	copy(data[0:8], []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) // Checksum
	endian.PutUint64(data[8:16], uint64(oid))
	endian.PutUint64(data[16:24], uint64(12345)) // XID
	endian.PutUint32(data[24:28], types.ObjectTypeBtreeNode)
	endian.PutUint32(data[28:32], 0) // Subtype

	// B-tree node specific fields
	nodeFlags := flags
	if isFixedKV {
		nodeFlags |= types.BtnodeFixedKvSize
	}
	endian.PutUint16(data[32:34], nodeFlags)
	endian.PutUint16(data[34:36], level)
	endian.PutUint32(data[36:40], nkeys)

	// Table space, free space, and free lists
	endian.PutUint16(data[40:42], 100) // table space offset
	endian.PutUint16(data[42:44], 200) // table space length
	endian.PutUint16(data[44:46], 300) // free space offset
	endian.PutUint16(data[46:48], 150) // free space length
	endian.PutUint16(data[48:50], 400) // key free list offset
	endian.PutUint16(data[50:52], 50)  // key free list length
	endian.PutUint16(data[52:54], 500) // value free list offset
	endian.PutUint16(data[54:56], 75)  // value free list length

	// Add test data containing child OIDs for internal nodes
	if level > 0 && isFixedKV {
		// For fixed-size key/value, store child OIDs in the data section
		// Internal nodes have nkeys + 1 children
		numChildren := nkeys + 1
		for i := uint32(0); i < numChildren; i++ {
			offset := 56 + int(i)*16 // 8 bytes key + 8 bytes value
			// Key (8 bytes) - for the last child, use a large key value
			var keyValue uint64
			if i < nkeys {
				keyValue = uint64(i + 1000)
			} else {
				keyValue = uint64(9999) // Large key for the last child pointer
			}
			endian.PutUint64(data[offset:offset+8], keyValue)
			// Value containing child OID (8 bytes)
			childOID := uint64(oid) + uint64(i) + 100
			endian.PutUint64(data[offset+8:offset+16], childOID)
		}
	}

	return data
}

func TestNewBTreeNavigator(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	btreeInfo := NewMockBTreeInfoReader()
	rootOID := types.OidT(1000)

	navigator := NewBTreeNavigator(blockReader, rootOID, btreeInfo)
	if navigator == nil {
		t.Fatal("NewBTreeNavigator returned nil")
	}

	// Test that we can cast to the concrete type to access additional methods
	if concreteNav, ok := navigator.(*btreeNavigator); ok {
		if concreteNav.rootOID != rootOID {
			t.Errorf("Root OID not set correctly: got %d, want %d", concreteNav.rootOID, rootOID)
		}
		if concreteNav.GetCacheSize() != 0 {
			t.Errorf("Initial cache size should be 0, got %d", concreteNav.GetCacheSize())
		}
	}
}

func TestBTreeNavigator_GetRootNode(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	btreeInfo := NewMockBTreeInfoReader()
	rootOID := types.OidT(1000)

	// Create root node data
	rootData := createTestNavigatorNodeData(rootOID, types.BtnodeRoot|types.BtnodeLeaf, 0, 5, true)
	blockReader.SetBlock(types.Paddr(rootOID), rootData)

	navigator := NewBTreeNavigator(blockReader, rootOID, btreeInfo)

	rootNode, err := navigator.GetRootNode()
	if err != nil {
		t.Fatalf("GetRootNode failed: %v", err)
	}

	if !rootNode.IsRoot() {
		t.Error("Root node should have root flag set")
	}

	if !rootNode.IsLeaf() {
		t.Error("Root node should have leaf flag set")
	}

	if rootNode.KeyCount() != 5 {
		t.Errorf("Root node key count: got %d, want 5", rootNode.KeyCount())
	}
}

func TestBTreeNavigator_GetHeight(t *testing.T) {
	testCases := []struct {
		name           string
		rootLevel      uint16
		expectedHeight uint16
	}{
		{"Single level tree", 0, 1},
		{"Two level tree", 1, 2},
		{"Three level tree", 2, 3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			blockReader := NewMockBlockDeviceReader()
			btreeInfo := NewMockBTreeInfoReader()
			rootOID := types.OidT(1000)

			// Create root node data with specified level
			rootData := createTestNavigatorNodeData(rootOID, types.BtnodeRoot, tc.rootLevel, 5, true)
			blockReader.SetBlock(types.Paddr(rootOID), rootData)

			navigator := NewBTreeNavigator(blockReader, rootOID, btreeInfo)

			height, err := navigator.GetHeight()
			if err != nil {
				t.Fatalf("GetHeight failed: %v", err)
			}

			if height != tc.expectedHeight {
				t.Errorf("Height: got %d, want %d", height, tc.expectedHeight)
			}
		})
	}
}

func TestBTreeNavigator_GetChildNode(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	btreeInfo := NewMockBTreeInfoReader()
	rootOID := types.OidT(1000)

	// Create internal root node with 3 keys (4 children)
	rootData := createTestNavigatorNodeData(rootOID, types.BtnodeRoot, 1, 3, true)
	blockReader.SetBlock(types.Paddr(rootOID), rootData)

	// Create child nodes
	for i := 0; i < 4; i++ {
		childOID := types.OidT(uint64(rootOID) + uint64(i) + 100)
		childData := createTestNavigatorNodeData(childOID, types.BtnodeLeaf, 0, 10, true)
		blockReader.SetBlock(types.Paddr(childOID), childData)
	}

	navigator := NewBTreeNavigator(blockReader, rootOID, btreeInfo)

	rootNode, err := navigator.GetRootNode()
	if err != nil {
		t.Fatalf("GetRootNode failed: %v", err)
	}

	// Test getting a valid child
	childNode, err := navigator.GetChildNode(rootNode, 0)
	if err != nil {
		t.Fatalf("GetChildNode failed: %v", err)
	}

	if !childNode.IsLeaf() {
		t.Error("Child node should be a leaf")
	}

	if childNode.KeyCount() != 10 {
		t.Errorf("Child node key count: got %d, want 10", childNode.KeyCount())
	}

	// Test error cases
	t.Run("leaf node error", func(t *testing.T) {
		leafData := createTestNavigatorNodeData(2000, types.BtnodeLeaf, 0, 5, true)
		blockReader.SetBlock(2000, leafData)
		leafNode, _ := parser.NewBTreeNodeReader(leafData, binary.LittleEndian)

		_, err := navigator.GetChildNode(leafNode, 0)
		if err == nil {
			t.Error("Expected error when getting child of leaf node")
		}
	})

	t.Run("invalid index error", func(t *testing.T) {
		_, err := navigator.GetChildNode(rootNode, 10)
		if err == nil {
			t.Error("Expected error for invalid child index")
		}

		_, err = navigator.GetChildNode(rootNode, -1)
		if err == nil {
			t.Error("Expected error for negative child index")
		}
	})
}

func TestBTreeNavigator_GetNodeByObjectID(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	btreeInfo := NewMockBTreeInfoReader()
	rootOID := types.OidT(1000)

	// Create test node
	nodeOID := types.OidT(2000)
	nodeData := createTestNavigatorNodeData(nodeOID, types.BtnodeLeaf, 0, 7, true)
	blockReader.SetBlock(types.Paddr(nodeOID), nodeData)

	navigator := NewBTreeNavigator(blockReader, rootOID, btreeInfo)

	// Test successful retrieval
	node, err := navigator.GetNodeByObjectID(nodeOID)
	if err != nil {
		t.Fatalf("GetNodeByObjectID failed: %v", err)
	}

	if node.KeyCount() != 7 {
		t.Errorf("Node key count: got %d, want 7", node.KeyCount())
	}

	// Test caching - second retrieval should use cache
	node2, err := navigator.GetNodeByObjectID(nodeOID)
	if err != nil {
		t.Fatalf("Second GetNodeByObjectID failed: %v", err)
	}

	if node != node2 {
		t.Error("Second retrieval should return cached node")
	}

	// Test cache size
	if concreteNav, ok := navigator.(*btreeNavigator); ok {
		if concreteNav.GetCacheSize() != 1 {
			t.Errorf("Cache size should be 1, got %d", concreteNav.GetCacheSize())
		}
	}

	// Test error case - non-existent node
	_, err = navigator.GetNodeByObjectID(types.OidT(9999))
	if err == nil {
		t.Error("Expected error for non-existent node")
	}
}

func TestBTreeNavigator_CacheOperations(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	btreeInfo := NewMockBTreeInfoReader()
	rootOID := types.OidT(1000)

	// Create test nodes
	for i := 0; i < 5; i++ {
		nodeOID := types.OidT(1000 + i)
		nodeData := createTestNavigatorNodeData(nodeOID, types.BtnodeLeaf, 0, uint32(i+1), true)
		blockReader.SetBlock(types.Paddr(nodeOID), nodeData)
	}

	navigator := NewBTreeNavigator(blockReader, rootOID, btreeInfo)
	concreteNav := navigator.(*btreeNavigator)

	// Load nodes into cache
	for i := 0; i < 5; i++ {
		_, err := navigator.GetNodeByObjectID(types.OidT(1000 + i))
		if err != nil {
			t.Fatalf("Failed to load node %d: %v", i, err)
		}
	}

	if concreteNav.GetCacheSize() != 5 {
		t.Errorf("Cache size should be 5, got %d", concreteNav.GetCacheSize())
	}

	// Clear cache
	concreteNav.ClearCache()

	if concreteNav.GetCacheSize() != 0 {
		t.Errorf("Cache size should be 0 after clear, got %d", concreteNav.GetCacheSize())
	}

	// Verify nodes need to be reloaded from block device
	_, err := navigator.GetNodeByObjectID(types.OidT(1000))
	if err != nil {
		t.Fatalf("Failed to reload node after cache clear: %v", err)
	}

	if concreteNav.GetCacheSize() != 1 {
		t.Errorf("Cache size should be 1 after reload, got %d", concreteNav.GetCacheSize())
	}
}

func TestBTreeNavigator_ExtractChildOID_Errors(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	btreeInfo := NewMockBTreeInfoReader()

	// Test with variable-size keys/values (not yet implemented)
	btreeInfo.keySize = 0
	btreeInfo.valueSize = 0

	rootOID := types.OidT(1000)
	rootData := createTestNavigatorNodeData(rootOID, types.BtnodeRoot, 1, 3, false) // variable size
	blockReader.SetBlock(types.Paddr(rootOID), rootData)

	navigator := NewBTreeNavigator(blockReader, rootOID, btreeInfo)
	rootNode, err := navigator.GetRootNode()
	if err != nil {
		t.Fatalf("GetRootNode failed: %v", err)
	}

	// This should fail because variable-size extraction is not implemented
	_, err = navigator.GetChildNode(rootNode, 0)
	if err == nil {
		t.Error("Expected error for variable-size key/value extraction")
	}
	if err.Error() != "failed to extract child OID: variable-size key/value extraction not yet implemented" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func BenchmarkBTreeNavigator_GetNodeByObjectID(b *testing.B) {
	blockReader := NewMockBlockDeviceReader()
	btreeInfo := NewMockBTreeInfoReader()
	rootOID := types.OidT(1000)

	// Create test nodes
	for i := 0; i < 100; i++ {
		nodeOID := types.OidT(1000 + i)
		nodeData := createTestNavigatorNodeData(nodeOID, types.BtnodeLeaf, 0, uint32(i+1), true)
		blockReader.SetBlock(types.Paddr(nodeOID), nodeData)
	}

	navigator := NewBTreeNavigator(blockReader, rootOID, btreeInfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nodeOID := types.OidT(1000 + (i % 100))
		_, err := navigator.GetNodeByObjectID(nodeOID)
		if err != nil {
			b.Fatalf("GetNodeByObjectID failed: %v", err)
		}
	}
}

func BenchmarkBTreeNavigator_CachedAccess(b *testing.B) {
	blockReader := NewMockBlockDeviceReader()
	btreeInfo := NewMockBTreeInfoReader()
	rootOID := types.OidT(1000)

	// Create and cache nodes
	nodeOID := types.OidT(2000)
	nodeData := createTestNavigatorNodeData(nodeOID, types.BtnodeLeaf, 0, 10, true)
	blockReader.SetBlock(types.Paddr(nodeOID), nodeData)

	navigator := NewBTreeNavigator(blockReader, rootOID, btreeInfo)

	// Prime the cache
	_, err := navigator.GetNodeByObjectID(nodeOID)
	if err != nil {
		b.Fatalf("Failed to prime cache: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := navigator.GetNodeByObjectID(nodeOID)
		if err != nil {
			b.Fatalf("Cached GetNodeByObjectID failed: %v", err)
		}
	}
}
