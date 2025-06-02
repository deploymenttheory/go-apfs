package btrees

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// Helper function to create a simple test tree with fixed-size keys/values
func createTestSearchTree() (interfaces.BTreeNavigator, interfaces.BTreeInfoReader, *MockBlockDeviceReader) {
	blockReader := NewMockBlockDeviceReader()
	btreeInfo := NewMockBTreeInfoReader()
	btreeInfo.keySize = 8
	btreeInfo.valueSize = 8

	// Create a simple 2-level tree:
	// Root (level 1) with 3 keys pointing to 3 leaf nodes
	// Each leaf node has 3 keys
	rootOID := types.OidT(1000)

	// Create root node with keys [10, 20, 30] pointing to children [1100, 1200, 1300, 1400]
	// Note: Internal nodes with n keys have n+1 children
	rootData := createTestSearchNodeData(rootOID, types.BtnodeRoot, 1, 3, []uint64{10, 20, 30}, []uint64{1100, 1200, 1300, 1400})
	blockReader.SetBlock(types.Paddr(rootOID), rootData)

	// Create leaf node 1100 with keys [1, 5, 9]
	leaf1Data := createTestSearchNodeData(1100, types.BtnodeLeaf, 0, 3, []uint64{1, 5, 9}, []uint64{101, 105, 109})
	blockReader.SetBlock(1100, leaf1Data)

	// Create leaf node 1200 with keys [11, 15, 19]
	leaf2Data := createTestSearchNodeData(1200, types.BtnodeLeaf, 0, 3, []uint64{11, 15, 19}, []uint64{111, 115, 119})
	blockReader.SetBlock(1200, leaf2Data)

	// Create leaf node 1300 with keys [21, 25, 29]
	leaf3Data := createTestSearchNodeData(1300, types.BtnodeLeaf, 0, 3, []uint64{21, 25, 29}, []uint64{121, 125, 129})
	blockReader.SetBlock(1300, leaf3Data)

	// Create leaf node 1400 with keys [31, 35, 39] (for keys > 30)
	leaf4Data := createTestSearchNodeData(1400, types.BtnodeLeaf, 0, 3, []uint64{31, 35, 39}, []uint64{131, 135, 139})
	blockReader.SetBlock(1400, leaf4Data)

	navigator := NewBTreeNavigator(blockReader, rootOID, btreeInfo)

	return navigator, btreeInfo, blockReader
}

// Helper function to create test node data with specific keys and values
func createTestSearchNodeData(oid types.OidT, flags uint16, level uint16, nkeys uint32, keys []uint64, values []uint64) []byte {
	// For internal nodes, we may have more values (child pointers) than keys
	numEntries := len(values)
	if level == 0 {
		// For leaf nodes, use the number of keys
		numEntries = int(nkeys)
	}

	data := make([]byte, 56+numEntries*16) // Header + key-value pairs
	endian := binary.LittleEndian

	// Object header (32 bytes)
	copy(data[0:8], []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08})
	endian.PutUint64(data[8:16], uint64(oid))
	endian.PutUint64(data[16:24], 12345)
	endian.PutUint32(data[24:28], types.ObjectTypeBtreeNode)
	endian.PutUint32(data[28:32], 0)

	// B-tree node specific fields
	endian.PutUint16(data[32:34], flags|types.BtnodeFixedKvSize)
	endian.PutUint16(data[34:36], level)
	endian.PutUint32(data[36:40], nkeys)

	// Table space, free space, and free lists
	endian.PutUint16(data[40:42], 100)
	endian.PutUint16(data[42:44], 200)
	endian.PutUint16(data[44:46], 300)
	endian.PutUint16(data[46:48], 150)
	endian.PutUint16(data[48:50], 400)
	endian.PutUint16(data[50:52], 50)
	endian.PutUint16(data[52:54], 500)
	endian.PutUint16(data[54:56], 75)

	// Add key-value data
	for i := 0; i < numEntries && i < len(values); i++ {
		offset := 56 + i*16

		// For internal nodes, we may have more child pointers than keys
		var keyValue uint64
		if i < len(keys) {
			keyValue = keys[i]
		} else {
			// For the extra child pointer, use a large key value
			keyValue = 9999
		}

		endian.PutUint64(data[offset:offset+8], keyValue)
		endian.PutUint64(data[offset+8:offset+16], values[i])
	}

	return data
}

// Helper function to create key from uint64
func uint64ToKey(val uint64) []byte {
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, val)
	return key
}

// Helper function to extract uint64 from value
func valueToUint64(val []byte) uint64 {
	if len(val) >= 8 {
		return binary.LittleEndian.Uint64(val)
	}
	return 0
}

func TestNewBTreeSearcher(t *testing.T) {
	navigator, btreeInfo, _ := createTestSearchTree()

	// Test with default key comparer
	searcher := NewBTreeSearcher(navigator, btreeInfo, nil)
	if searcher == nil {
		t.Fatal("NewBTreeSearcher returned nil")
	}

	// Test with custom key comparer
	customComparer := func(a, b []byte) int {
		return bytes.Compare(a, b)
	}
	searcher2 := NewBTreeSearcher(navigator, btreeInfo, customComparer)
	if searcher2 == nil {
		t.Fatal("NewBTreeSearcher with custom comparer returned nil")
	}
}

func TestBTreeSearcher_Find(t *testing.T) {
	navigator, btreeInfo, _ := createTestSearchTree()
	searcher := NewBTreeSearcher(navigator, btreeInfo, nil)

	testCases := []struct {
		name          string
		searchKey     uint64
		expectedValue uint64
		shouldFind    bool
	}{
		{"Find key 1", 1, 101, true},
		{"Find key 5", 5, 105, true},
		{"Find key 9", 9, 109, true},
		{"Find key 11", 11, 111, true},
		{"Find key 15", 15, 115, true},
		{"Find key 19", 19, 119, true},
		{"Find key 21", 21, 121, true},
		{"Find key 25", 25, 125, true},
		{"Find key 29", 29, 129, true},
		{"Find non-existent key", 100, 0, false},
		{"Find key 0", 0, 0, false},
		{"Find key 50", 50, 0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key := uint64ToKey(tc.searchKey)
			value, err := searcher.Find(key)

			if tc.shouldFind {
				if err != nil {
					t.Fatalf("Find(%d) failed: %v", tc.searchKey, err)
				}
				actualValue := valueToUint64(value)
				if actualValue != tc.expectedValue {
					t.Errorf("Find(%d) = %d, want %d", tc.searchKey, actualValue, tc.expectedValue)
				}
			} else {
				if err == nil {
					t.Errorf("Find(%d) should have failed but returned value %v", tc.searchKey, value)
				}
			}
		})
	}
}

func TestBTreeSearcher_ContainsKey(t *testing.T) {
	navigator, btreeInfo, _ := createTestSearchTree()
	searcher := NewBTreeSearcher(navigator, btreeInfo, nil)

	testCases := []struct {
		name       string
		searchKey  uint64
		shouldFind bool
	}{
		{"Contains key 1", 1, true},
		{"Contains key 15", 15, true},
		{"Contains key 29", 29, true},
		{"Does not contain key 100", 100, false},
		{"Does not contain key 0", 0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key := uint64ToKey(tc.searchKey)
			found, err := searcher.ContainsKey(key)

			if err != nil {
				t.Fatalf("ContainsKey(%d) failed: %v", tc.searchKey, err)
			}

			if found != tc.shouldFind {
				t.Errorf("ContainsKey(%d) = %v, want %v", tc.searchKey, found, tc.shouldFind)
			}
		})
	}
}

func BenchmarkBTreeSearcher_Find(b *testing.B) {
	navigator, btreeInfo, _ := createTestSearchTree()
	searcher := NewBTreeSearcher(navigator, btreeInfo, nil)

	// Test keys that exist in the tree
	keys := []uint64{1, 5, 9, 11, 15, 19, 21, 25, 29}
	keyBytes := make([][]byte, len(keys))
	for i, k := range keys {
		keyBytes[i] = uint64ToKey(k)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keyBytes[i%len(keyBytes)]
		_, err := searcher.Find(key)
		if err != nil {
			b.Fatalf("Find failed: %v", err)
		}
	}
}
