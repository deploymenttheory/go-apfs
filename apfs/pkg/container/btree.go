// File: pkg/container/btree.go
package container

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/apfs/pkg/checksum"
	"github.com/deploymenttheory/go-apfs/apfs/pkg/types"
)

// BTreeNodePhysSize defines the fixed-size portion of a B-tree node structure
const BTreeNodePhysSize = 64

// ReadBTreeNodePhys reads and parses a BTreeNodePhys structure from a block device.
func ReadBTreeNodePhys(device types.BlockDevice, addr types.PAddr) (*types.BTreeNodePhys, error) {
	blockSize := device.GetBlockSize()
	if BTreeNodePhysSize > int(blockSize) {
		return nil, fmt.Errorf("BTreeNodePhys size (%d) exceeds device block size (%d)", BTreeNodePhysSize, blockSize)
	}

	data, err := device.ReadBlock(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to read block at addr %d: %w", addr, err)
	}

	if len(data) < BTreeNodePhysSize {
		return nil, fmt.Errorf("data too short: expected %d bytes, got %d", BTreeNodePhysSize, len(data))
	}

	node := &types.BTreeNodePhys{
		Flags: binary.LittleEndian.Uint16(data[32:34]),
		Level: binary.LittleEndian.Uint16(data[34:36]),
		NKeys: binary.LittleEndian.Uint32(data[36:40]),
		TableSpace: types.NLoc{
			Off: binary.LittleEndian.Uint16(data[40:42]),
			Len: binary.LittleEndian.Uint16(data[42:44]),
		},
		FreeSpace: types.NLoc{
			Off: binary.LittleEndian.Uint16(data[44:46]),
			Len: binary.LittleEndian.Uint16(data[46:48]),
		},
		KeyFreeList: types.NLoc{
			Off: binary.LittleEndian.Uint16(data[48:50]),
			Len: binary.LittleEndian.Uint16(data[50:52]),
		},
		ValFreeList: types.NLoc{
			Off: binary.LittleEndian.Uint16(data[52:54]),
			Len: binary.LittleEndian.Uint16(data[54:56]),
		},
		Data: data[BTreeNodePhysSize:],
	}

	computedChecksum := checksum.Fletcher64WithZeroedChecksum(data, 0)
	expectedChecksum := binary.LittleEndian.Uint64(data[:8])
	if computedChecksum != expectedChecksum {
		return nil, fmt.Errorf("checksum mismatch: computed 0x%x, expected 0x%x", computedChecksum, expectedChecksum)
	}

	return node, nil
}

// ValidateBTreeNodePhys performs basic validation on the BTreeNodePhys structure.
func ValidateBTreeNodePhys(node *types.BTreeNodePhys) error {
	if node.NKeys == 0 {
		return fmt.Errorf("invalid number of keys: %d", node.NKeys)
	}
	if node.Level > 16 { // arbitrary sanity check
		return fmt.Errorf("unreasonable B-tree level: %d", node.Level)
	}
	if node.Data == nil {
		return fmt.Errorf("data section cannot be nil")
	}

	// Additional checks can be implemented here
	return nil
}

// SearchBTreeNodePhys performs binary search on a B-tree node to find a key
func SearchBTreeNodePhys(node *types.BTreeNodePhys, searchKey []byte, keyCompare func(a, b []byte) int) (int, bool, error) {
	if node == nil || node.Data == nil {
		return 0, false, fmt.Errorf("node or node data is nil")
	}

	left, right := 0, int(node.NKeys)-1
	for left <= right {
		mid := left + (right-left)/2
		key, err := GetKeyAtIndex(node, mid)
		if err != nil {
			return 0, false, err
		}

		cmp := keyCompare(searchKey, key)
		if cmp == 0 {
			return mid, true, nil
		} else if cmp < 0 {
			right = mid - 1
		} else {
			left = mid + 1
		}
	}

	if left > int(node.NKeys) {
		left = int(node.NKeys)
	}

	return left, false, nil
}

// GetKeyAtIndex retrieves the key from a B-tree node at a given index.
func GetKeyAtIndex(node *types.BTreeNodePhys, index int) ([]byte, error) {
	if index < 0 || uint32(index) >= node.NKeys {
		return nil, fmt.Errorf("key index out of bounds: %d", index)
	}

	const kvLocSize = 4 // explicitly two uint16 fields (offset, length)
	locStart := int(node.TableSpace.Off) + index*kvLocSize

	if locStart+kvLocSize > len(node.Data) {
		return nil, fmt.Errorf("key location out of bounds")
	}
	keyOff := binary.LittleEndian.Uint16(node.Data[locStart : locStart+2])
	keyLen := binary.LittleEndian.Uint16(node.Data[locStart+2 : locStart+4])

	keyStart, keyEnd := int(keyOff), int(keyOff)+int(keyLen)
	if keyEnd > len(node.Data) {
		return nil, fmt.Errorf("key at index %d exceeds node data bounds", index)
	}

	return node.Data[keyStart:keyEnd], nil
}

// TraverseBTree recursively traverses a B-tree and executes a callback for each leaf node encountered.
func TraverseBTree(device types.BlockDevice, addr types.PAddr, callback func(node *types.BTreeNodePhys) error) error {
	node, err := ReadBTreeNodePhys(device, addr)
	if err != nil {
		return err
	}

	if node.IsLeaf() {
		return callback(node)
	}

	// Traverse child nodes if not a leaf
	for i := 0; i < int(node.NKeys); i++ {
		val, err := GetValueAtIndex(node, i)
		if err != nil {
			return err
		}

		// Extract child node address from val
		if len(val) < 8 {
			return fmt.Errorf("invalid child node value length")
		}
		childAddr := types.PAddr(binary.LittleEndian.Uint64(val[:8]))
		if err := TraverseBTree(device, childAddr, callback); err != nil {
			return err
		}
	}

	return nil
}

func GetValueAtIndex(node *types.BTreeNodePhys, index int) ([]byte, error) {
	if index < 0 || uint32(index) >= node.NKeys {
		return nil, fmt.Errorf("value index out of bounds: %d", index)
	}

	const kvLocSize = 4
	locStart := int(node.TableSpace.Off) + index*kvLocSize

	if locStart+kvLocSize > len(node.Data) {
		return nil, fmt.Errorf("value location out of bounds")
	}

	valOff := binary.LittleEndian.Uint16(node.Data[locStart+kvLocSize : locStart+kvLocSize+2])
	valLen := binary.LittleEndian.Uint16(node.Data[locStart+kvLocSize+2 : locStart+kvLocSize+4])

	valStart, valEnd := int(valOff), int(valOff)+int(valLen)
	if valEnd > len(node.Data) {
		return nil, fmt.Errorf("value at index %d exceeds node data bounds", index)
	}

	return node.Data[valStart:valEnd], nil
}
