// File: pkg/container/btree.go
package container

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/apfs/pkg/checksum"
	"github.com/deploymenttheory/go-apfs/apfs/pkg/types"
)

// BTNodePhysSize defines the fixed-size portion of a B-tree node structure
const BTNodePhysSize = 64
const kvLocSize = 8

// ReadBTreeNodePhys reads and parses a BTreeNodePhys structure from a block device.
func ReadBTreeNodePhys(device types.BlockDevice, addr types.PAddr) (*types.BTNodePhys, error) {
	blockSize := device.GetBlockSize()
	if BTNodePhysSize > int(blockSize) {
		return nil, fmt.Errorf("BTreeNodePhys size (%d) exceeds device block size (%d)", BTNodePhysSize, blockSize)
	}

	data, err := device.ReadBlock(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to read block at addr %d: %w", addr, err)
	}

	if len(data) < BTNodePhysSize {
		return nil, fmt.Errorf("data too short: expected %d bytes, got %d", BTNodePhysSize, len(data))
	}

	node := &types.BTNodePhys{
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
		Data: data[BTNodePhysSize:],
	}

	computedChecksum := checksum.Fletcher64WithZeroedChecksum(data, 0)
	expectedChecksum := binary.LittleEndian.Uint64(data[:8])
	if computedChecksum != expectedChecksum {
		return nil, fmt.Errorf("checksum mismatch: computed 0x%x, expected 0x%x", computedChecksum, expectedChecksum)
	}

	return node, nil
}

// ValidateBTreeNodePhys performs basic validation on the BTNodePhys structure.
func ValidateBTreeNodePhys(node *types.BTNodePhys) error {
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

// SearchBTreeNodePhys performs binary search on a B-tree node to find a key.
func SearchBTreeNodePhys(node *types.BTNodePhys, searchKey []byte, keyCompare func(a, b []byte) int) (int, bool, error) {
	if node == nil || node.Data == nil {
		return 0, false, fmt.Errorf("node or node data is nil")
	}

	left, right := 0, int(node.NKeys)-1

	for left <= right {
		mid := left + (right-left)/2
		key, err := GetKeyAtIndex(node, mid)
		if err != nil {
			return 0, false, fmt.Errorf("GetKeyAtIndex(mid=%d) failed: %w", mid, err)
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

	// Clamp left to NKeys range
	if left >= int(node.NKeys) {
		left = int(node.NKeys)
	}

	return left, false, nil
}

// GetKeyAtIndex returns the key bytes for the kvloc_t at index.
func GetKeyAtIndex(node *types.BTNodePhys, index int) ([]byte, error) {
	if index < 0 || uint32(index) >= node.NKeys {
		return nil, fmt.Errorf("key index out of bounds: %d", index)
	}

	locStart := int(node.TableSpace.Off) + index*kvLocSize
	if locStart+kvLocSize > len(node.Data) {
		return nil, fmt.Errorf("kvloc entry out of bounds")
	}

	keyOff := binary.LittleEndian.Uint16(node.Data[locStart : locStart+2])
	keyLen := binary.LittleEndian.Uint16(node.Data[locStart+2 : locStart+4])

	if int(keyOff)+int(keyLen) > len(node.Data) {
		return nil, fmt.Errorf("invalid key range: off=%d len=%d", keyOff, keyLen)
	}

	return node.Data[keyOff : keyOff+keyLen], nil
}

// GetValueAtIndex returns the value bytes for the kvloc_t at index.
func GetValueAtIndex(node *types.BTNodePhys, index int) ([]byte, error) {
	if index < 0 || uint32(index) >= node.NKeys {
		return nil, fmt.Errorf("value index out of bounds: %d", index)
	}

	locStart := int(node.TableSpace.Off) + index*kvLocSize
	if locStart+kvLocSize > len(node.Data) {
		return nil, fmt.Errorf("kvloc entry out of bounds")
	}

	valOff := binary.LittleEndian.Uint16(node.Data[locStart+4 : locStart+6])
	valLen := binary.LittleEndian.Uint16(node.Data[locStart+6 : locStart+8])

	if int(valOff)+int(valLen) > len(node.Data) {
		return nil, fmt.Errorf("invalid value range: off=%d len=%d", valOff, valLen)
	}

	return node.Data[valOff : valOff+valLen], nil
}

// TraverseBTree recursively traverses a B-tree and executes a callback for each leaf node encountered.
func TraverseBTree(device types.BlockDevice, addr types.PAddr, callback func(node *types.BTNodePhys) error) error {
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

// InsertKeyValueLeaf inserts a key/value pair into a leaf node using APFS-compliant kvloc_t layout.
// kvloc_t: { uint16 key_off, uint16 key_len, uint16 val_off, uint16 val_len }
func InsertKeyValueLeaf(node *types.BTNodePhys, key, value []byte) error {
	if !node.IsLeaf() {
		return fmt.Errorf("insert only supported on leaf nodes")
	}

	tableStart := int(node.TableSpace.Off)

	// Append key and value to the end of the data section
	keyOffset := len(node.Data)
	node.Data = append(node.Data, key...)
	keyLen := len(key)

	valOffset := len(node.Data)
	node.Data = append(node.Data, value...)
	valLen := len(value)

	// Build the kvloc entry
	kvloc := make([]byte, kvLocSize)
	binary.LittleEndian.PutUint16(kvloc[0:2], uint16(keyOffset))
	binary.LittleEndian.PutUint16(kvloc[2:4], uint16(keyLen))
	binary.LittleEndian.PutUint16(kvloc[4:6], uint16(valOffset))
	binary.LittleEndian.PutUint16(kvloc[6:8], uint16(valLen))

	// Determine sorted insert position
	insertIdx := 0
	for i := 0; i < int(node.NKeys); i++ {
		off := tableStart + i*kvLocSize
		if off+kvLocSize > len(node.Data) {
			break
		}
		koff := binary.LittleEndian.Uint16(node.Data[off : off+2])
		klen := binary.LittleEndian.Uint16(node.Data[off+2 : off+4])
		if int(koff)+int(klen) > len(node.Data) {
			break
		}
		existingKey := node.Data[koff : koff+klen]
		if bytes.Compare(key, existingKey) < 0 {
			break
		}
		insertIdx++
	}

	// Calculate copy boundaries
	insertPos := tableStart + insertIdx*kvLocSize
	endPos := tableStart + int(node.NKeys)*kvLocSize
	neededLen := endPos + kvLocSize
	if len(node.Data) < neededLen {
		node.Data = append(node.Data, make([]byte, neededLen-len(node.Data))...)
	}

	// Shift kvloc entries and insert the new one
	copy(node.Data[insertPos+kvLocSize:], node.Data[insertPos:endPos])
	copy(node.Data[insertPos:], kvloc)

	node.NKeys++
	node.TableSpace.Len += uint16(kvLocSize)
	return nil
}

// DeleteKeyValue removes a key from a leaf node and updates location entries.
// Note: Actual key/value bytes are not removed from node.Data, mimicking typical disk behavior.
func DeleteKeyValue(node *types.BTNodePhys, key []byte, compare func(a, b []byte) int) error {
	if !node.IsLeaf() {
		return fmt.Errorf("deletion only supported on leaf nodes")
	}

	index := -1
	for i := 0; i < int(node.NKeys); i++ {
		k, err := GetKeyAtIndex(node, i)
		if err != nil {
			return err
		}
		if compare(key, k) == 0 {
			index = i
			break
		}
	}
	if index == -1 {
		return fmt.Errorf("key not found for deletion")
	}

	locStart := int(node.TableSpace.Off) + index*kvLocSize
	locEnd := locStart + kvLocSize

	copy(node.Data[locStart:], node.Data[locEnd:int(node.TableSpace.Off)+int(node.TableSpace.Len)])
	node.Data = node.Data[:len(node.Data)-kvLocSize]
	node.TableSpace.Len -= uint16(kvLocSize)
	node.NKeys--
	return nil
}
