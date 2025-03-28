// File: pkg/container/btree_test.go
package container

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-apfs/apfs/pkg/checksum"
	"github.com/deploymenttheory/go-apfs/apfs/pkg/types"
)

// MockBlockDevice simulates a block device for testing purposes.
type MockBlockDevice struct {
	BlockSize uint32
	Blocks    map[types.PAddr][]byte
}

func (m *MockBlockDevice) ReadBlock(addr types.PAddr) ([]byte, error) {
	data, ok := m.Blocks[addr]
	if !ok {
		return nil, errors.New("block not found")
	}
	return data, nil
}

func (m *MockBlockDevice) WriteBlock(addr types.PAddr, data []byte) error {
	m.Blocks[addr] = data
	return nil
}

func (m *MockBlockDevice) GetBlockSize() uint32 {
	return m.BlockSize
}

func (m *MockBlockDevice) GetBlockCount() uint64 {
	return uint64(len(m.Blocks))
}

func (m *MockBlockDevice) Close() error {
	return nil
}

func TestReadBTreeNodePhys(t *testing.T) {
	device := &MockBlockDevice{
		BlockSize: 4096,
		Blocks:    make(map[types.PAddr][]byte),
	}

	// Set up test data
	addr := types.PAddr(0x10)
	data := make([]byte, 4096)
	binary.LittleEndian.PutUint64(data[:8], 0)             // checksum placeholder
	binary.LittleEndian.PutUint16(data[32:34], 0x0002)     // Flags
	binary.LittleEndian.PutUint16(data[34:36], 0x0001)     // Level
	binary.LittleEndian.PutUint32(data[36:40], 0x00000001) // NKeys
	binary.LittleEndian.PutUint16(data[40:42], 64)         // TableSpace.Off
	binary.LittleEndian.PutUint16(data[42:44], 128)        // TableSpace.Len

	// Compute checksum
	checksum := checksum.Fletcher64WithZeroedChecksum(data, 0)
	binary.LittleEndian.PutUint64(data[:8], checksum)

	device.Blocks[addr] = data

	node, err := ReadBTreeNodePhys(device, addr)
	if err != nil {
		t.Fatalf("ReadBTreeNodePhys failed: %v", err)
	}

	if node.Flags != 0x0002 {
		t.Errorf("expected Flags 0x0002, got 0x%x", node.Flags)
	}
	if node.Level != 0x0001 {
		t.Errorf("expected Level 1, got %d", node.Level)
	}
	if node.NKeys != 1 {
		t.Errorf("expected NKeys 1, got %d", node.NKeys)
	}
}

func TestValidateBTreeNodePhys(t *testing.T) {
	node := &types.BTNodePhys{
		Flags: 0x0002,
		Level: 1,
		NKeys: 1,
		Data:  []byte{1, 2, 3},
	}

	if err := ValidateBTreeNodePhys(node); err != nil {
		t.Errorf("validation failed unexpectedly: %v", err)
	}

	// Test failure scenario
	node.NKeys = 0
	if err := ValidateBTreeNodePhys(node); err == nil {
		t.Error("expected validation error for NKeys=0, got nil")
	}
}

func TestSearchBTreeNode(t *testing.T) {
	node := &types.BTNodePhys{
		Flags:      0x02, // Leaf node flag
		Level:      0,
		Data:       make([]byte, 0, 512),
		TableSpace: types.NLoc{Off: 256, Len: 0},
	}

	// Insert keys in a specific order to test binary search
	keys := []string{"apple", "hello", "zebra"}
	values := []string{"red", "world", "stripes"}

	for i, key := range keys {
		err := InsertKeyValueLeaf(node, []byte(key), []byte(values[i]))
		if err != nil {
			t.Fatalf("Failed to insert key %s: %v", key, err)
		}
	}

	// Debug: print out all keys to verify insertion
	t.Log("Inserted keys:")
	for i := 0; i < int(node.NKeys); i++ {
		k, err := GetKeyAtIndex(node, i)
		if err != nil {
			t.Logf("Error getting key at index %d: %v", i, err)
			continue
		}
		t.Logf("Key at index %d: %q", i, k)
	}

	tests := []struct {
		key       string
		wantIndex int
		wantFound bool
	}{
		{"apple", 0, true},
		{"hello", 1, true},
		{"zebra", 2, true},
		{"banana", 1, false},
	}

	for _, tc := range tests {
		index, found, err := SearchBTreeNodePhys(node, []byte(tc.key), bytes.Compare)
		if err != nil {
			t.Fatalf("SearchBTreeNode failed for key '%s': %v", tc.key, err)
		}
		if found != tc.wantFound || index != tc.wantIndex {
			t.Errorf("key '%s': want index %d found %v, got index %d found %v",
				tc.key, tc.wantIndex, tc.wantFound, index, found)
		}
	}
}

func TestGetKeyAtIndex(t *testing.T) {
	node := &types.BTNodePhys{
		Flags:      0x02,
		Level:      0,
		Data:       make([]byte, 0, 512),
		TableSpace: types.NLoc{Off: 256, Len: 0},
	}
	InsertKeyValueLeaf(node, []byte("testk"), []byte("val"))

	key, err := GetKeyAtIndex(node, 0)
	if err != nil {
		t.Fatalf("GetKeyAtIndex failed: %v", err)
	}
	if string(key) != "testk" {
		t.Errorf("expected 'testk', got '%s'", key)
	}
}

func TestGetValueAtIndex(t *testing.T) {
	node := &types.BTNodePhys{
		Flags:      0x02,
		Level:      0,
		Data:       make([]byte, 0, 512),
		TableSpace: types.NLoc{Off: 256, Len: 0},
	}
	InsertKeyValueLeaf(node, []byte("k"), []byte("value"))

	value, err := GetValueAtIndex(node, 0)
	if err != nil {
		t.Fatalf("GetValueAtIndex failed: %v", err)
	}
	if string(value) != "value" {
		t.Errorf("expected 'value', got '%s'", value)
	}
}

func TestInsertKeyValueLeafAndGetKeyValue(t *testing.T) {
	node := &types.BTNodePhys{
		Flags:      0x02, // leaf
		Level:      0,
		Data:       make([]byte, 0, 512),
		TableSpace: types.NLoc{Off: 256, Len: 0},
	}

	key := []byte("foo")
	val := []byte("bar")
	err := InsertKeyValueLeaf(node, key, val)
	if err != nil {
		t.Fatalf("InsertKeyValueLeaf failed: %v", err)
	}

	if node.NKeys != 1 {
		t.Errorf("expected NKeys = 1, got %d", node.NKeys)
	}

	readKey, err := GetKeyAtIndex(node, 0)
	if err != nil {
		t.Fatalf("GetKeyAtIndex failed: %v", err)
	}
	if !bytes.Equal(readKey, key) {
		t.Errorf("expected key %q, got %q", key, readKey)
	}

	readVal, err := GetValueAtIndex(node, 0)
	if err != nil {
		t.Fatalf("GetValueAtIndex failed: %v", err)
	}
	if !bytes.Equal(readVal, val) {
		t.Errorf("expected value %q, got %q", val, readVal)
	}
}

func TestDeleteKeyValue(t *testing.T) {
	node := &types.BTNodePhys{
		Flags:      0x02,
		Level:      0,
		Data:       make([]byte, 0, 512),
		TableSpace: types.NLoc{Off: 256, Len: 0},
	}

	InsertKeyValueLeaf(node, []byte("foo"), []byte("bar"))
	InsertKeyValueLeaf(node, []byte("baz"), []byte("qux"))

	if node.NKeys != 2 {
		t.Fatalf("expected 2 keys, got %d", node.NKeys)
	}

	err := DeleteKeyValue(node, []byte("foo"), bytes.Compare)
	if err != nil {
		t.Fatalf("DeleteKeyValue failed: %v", err)
	}

	if node.NKeys != 1 {
		t.Errorf("expected NKeys = 1 after deletion, got %d", node.NKeys)
	}

	// Should fail if key doesn't exist
	err = DeleteKeyValue(node, []byte("nonexistent"), bytes.Compare)
	if err == nil {
		t.Error("expected error for deleting nonexistent key, got nil")
	}
}
