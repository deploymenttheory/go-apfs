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
	node := &types.BTreeNodePhys{
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
	nodeData := make([]byte, 50)

	// First key: "hello" at offset 20, length 5
	binary.LittleEndian.PutUint16(nodeData[0:2], 20)
	binary.LittleEndian.PutUint16(nodeData[2:4], 5)

	// Second key: "world" at offset 25, length 5
	binary.LittleEndian.PutUint16(nodeData[4:6], 25)
	binary.LittleEndian.PutUint16(nodeData[6:8], 5)

	copy(nodeData[20:], "hello")
	copy(nodeData[25:], "world")

	node := &types.BTreeNodePhys{
		NKeys: 2,
		Data:  nodeData,
		TableSpace: types.NLoc{
			Off: 0,
			Len: 20,
		},
	}

	keyCompare := bytes.Compare

	tests := []struct {
		key       string
		wantIndex int
		wantFound bool
	}{
		{"apple", 0, false},
		{"hello", 0, true},
		{"world", 1, true},
		{"zebra", 2, false},
	}

	for _, tc := range tests {
		index, found, err := SearchBTreeNodePhys(node, []byte(tc.key), keyCompare)
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
	node := &types.BTreeNodePhys{
		NKeys: 1,
		Data:  []byte("\x04\x00\x05\x00testkey"),
		TableSpace: types.NLoc{
			Off: 0,
			Len: 20,
		},
	}

	key, err := GetKeyAtIndex(node, 0)
	if err != nil {
		t.Fatalf("GetKeyAtIndex failed: %v", err)
	}
	if string(key) != "testk" {
		t.Errorf("expected 'testk', got '%s'", key)
	}
}

func TestGetValueAtIndex(t *testing.T) {
	node := &types.BTreeNodePhys{
		NKeys: 1,
		Data:  []byte("\x00\x00\x00\x00\x08\x00\x05\x00value"),
		TableSpace: types.NLoc{
			Off: 0,
			Len: 20,
		},
	}

	value, err := GetValueAtIndex(node, 0)
	if err != nil {
		t.Fatalf("GetValueAtIndex failed: %v", err)
	}
	if string(value) != "value" {
		t.Errorf("expected 'value', got '%s'", value)
	}
}
