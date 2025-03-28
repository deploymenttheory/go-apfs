// File: pkg/container/btree.go
package container

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/apfs/pkg/checksum"
	"github.com/deploymenttheory/go-apfs/apfs/pkg/types"
)

// Constants based on APFS documentation
const (
	BTNodePhysSize = 64             // Size of the header portion
	KVLocSize      = 8              // Size of a kvloc_t entry (2 bytes key offset, 2 bytes key length, 2 bytes value offset, 2 bytes value length)
	KVAlignment    = 8              // Alignment for keys and values (8-byte boundary)
	BTOFFInvalid   = uint16(0xffff) // Invalid offset marker
)

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

	// Parse the node structure
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

	// Verify checksum
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

	return nil
}

// isFixedKVSize checks if the node has fixed key-value sizes
func isFixedKVSize(node *types.BTNodePhys) bool {
	return (node.Flags & 0x0004) != 0 // BTNODE_FIXED_KV_SIZE flag
}

// GetKeyValueLocation retrieves the kvloc_t entry at the specified index
func GetKeyValueLocation(node *types.BTNodePhys, index int) (*types.KVLoc, error) {
	if index < 0 || uint32(index) >= node.NKeys {
		return nil, fmt.Errorf("index out of bounds: %d", index)
	}

	if node.TableSpace.Off == 0 || node.TableSpace.Len == 0 {
		return nil, fmt.Errorf("table space not initialized")
	}

	// Calculate where in Data the kvloc entry is stored
	locOffset := int(node.TableSpace.Off) + index*KVLocSize
	if locOffset+KVLocSize > len(node.Data) {
		return nil, fmt.Errorf("kvloc entry out of bounds: locOffset=%d, Data length=%d",
			locOffset, len(node.Data))
	}

	// Extract the kvloc fields
	kvloc := &types.KVLoc{
		K: types.NLoc{
			Off: binary.LittleEndian.Uint16(node.Data[locOffset : locOffset+2]),
			Len: binary.LittleEndian.Uint16(node.Data[locOffset+2 : locOffset+4]),
		},
		V: types.NLoc{
			Off: binary.LittleEndian.Uint16(node.Data[locOffset+4 : locOffset+6]),
			Len: binary.LittleEndian.Uint16(node.Data[locOffset+6 : locOffset+8]),
		},
	}

	return kvloc, nil
}

// GetKeyAtIndex returns the key bytes for the entry at index
func GetKeyAtIndex(node *types.BTNodePhys, index int) ([]byte, error) {
	kvloc, err := GetKeyValueLocation(node, index)
	if err != nil {
		return nil, err
	}

	// Validate key offset and length
	keyOff := int(kvloc.K.Off)
	keyLen := int(kvloc.K.Len)

	if keyOff+keyLen > len(node.Data) {
		return nil, fmt.Errorf("invalid key range: off=%d len=%d, Data length=%d",
			keyOff, keyLen, len(node.Data))
	}

	return node.Data[keyOff : keyOff+keyLen], nil
}

// GetValueAtIndex returns the value bytes for the entry at index
func GetValueAtIndex(node *types.BTNodePhys, index int) ([]byte, error) {
	kvloc, err := GetKeyValueLocation(node, index)
	if err != nil {
		return nil, err
	}

	// Validate value offset and length
	valOff := int(kvloc.V.Off)
	valLen := int(kvloc.V.Len)

	if valOff+valLen > len(node.Data) {
		return nil, fmt.Errorf("invalid value range: off=%d len=%d, Data length=%d",
			valOff, valLen, len(node.Data))
	}

	return node.Data[valOff : valOff+valLen], nil
}

// SearchBTreeNodePhys performs binary search on a B-tree node to find a key
func SearchBTreeNodePhys(node *types.BTNodePhys, searchKey []byte, keyCompare func(a, b []byte) int) (int, bool, error) {
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

	// Key not found - return insertion point
	return left, false, nil
}

// InitializeEmptyNode sets up a new empty node with properly initialized areas
func InitializeEmptyNode(node *types.BTNodePhys, isLeaf bool) {
	// Initialize the Data slice if it's nil
	if node.Data == nil {
		// Use a reasonable default size if not specified
		node.Data = make([]byte, 512)
	}

	dataSize := len(node.Data)

	// Set up node properties
	if isLeaf {
		node.Flags = types.BTNodeLeaf
	}
	node.Level = 0
	node.NKeys = 0

	// Set up table space at the beginning
	tocSize := 64 // Initial space for table of contents
	node.TableSpace.Off = 0
	node.TableSpace.Len = 0

	// Key area starts after table space
	keyAreaStart := tocSize

	// Value area starts at the end
	valueAreaStart := dataSize

	// Free space is between key area and value area
	node.FreeSpace.Off = uint16(keyAreaStart)
	node.FreeSpace.Len = uint16(valueAreaStart - keyAreaStart)

	// Initialize free lists as empty
	node.KeyFreeList.Off = BTOFFInvalid
	node.KeyFreeList.Len = 0
	node.ValFreeList.Off = BTOFFInvalid
	node.ValFreeList.Len = 0
}

// InsertKeyValueLeaf inserts a key/value pair into a leaf node
func InsertKeyValueLeaf(node *types.BTNodePhys, key, value []byte) error {
	if node == nil {
		return fmt.Errorf("node is nil")
	}

	// Initialize the node if needed
	if len(node.Data) == 0 {
		node.Data = make([]byte, 512) // Default size for testing
		node.Flags = types.BTNodeLeaf // Mark as leaf node
		node.FreeSpace.Off = 64       // Table of contents starts at offset 0, keys start at 64
		node.FreeSpace.Len = 448      // Most of the node is free space initially (512 - 64)
	}

	// Make sure node is a leaf
	if !node.IsLeaf() {
		return fmt.Errorf("insert only supported on leaf nodes")
	}

	// Find insertion position to maintain sorted order
	insertIdx := 0
	for i := 0; i < int(node.NKeys); i++ {
		existingKey, err := GetKeyAtIndex(node, i)
		if err != nil {
			break // Can't read key, assume we insert at this position
		}
		if bytes.Compare(key, existingKey) > 0 {
			insertIdx = i + 1 // Insert after this key
		} else {
			break
		}
	}

	// Simple offset calculation - in a production implementation, this would
	// handle alignment and space management more robustly
	keyOffset := int(node.FreeSpace.Off) // Keys go right after the table space
	valueOffset := keyOffset + len(key)  // Values go right after keys for simplicity

	// Create a new TOC entry
	tocEntry := make([]byte, KVLocSize)
	binary.LittleEndian.PutUint16(tocEntry[0:2], uint16(keyOffset))
	binary.LittleEndian.PutUint16(tocEntry[2:4], uint16(len(key)))
	binary.LittleEndian.PutUint16(tocEntry[4:6], uint16(valueOffset))
	binary.LittleEndian.PutUint16(tocEntry[6:8], uint16(len(value)))

	// Make space for TOC entry
	tocStart := int(node.TableSpace.Off)
	tocInsertPos := tocStart + insertIdx*KVLocSize

	// Expand Data if needed
	neededSize := max(valueOffset+len(value), tocInsertPos+KVLocSize)
	if neededSize > len(node.Data) {
		newData := make([]byte, neededSize)
		copy(newData, node.Data)
		node.Data = newData
	}

	// Shift existing TOC entries if inserting in the middle
	if insertIdx < int(node.NKeys) {
		copy(node.Data[tocInsertPos+KVLocSize:],
			node.Data[tocInsertPos:tocStart+int(node.NKeys)*KVLocSize])
	}

	// Insert the TOC entry
	copy(node.Data[tocInsertPos:tocInsertPos+KVLocSize], tocEntry)

	// Store the key and value
	copy(node.Data[keyOffset:keyOffset+len(key)], key)
	copy(node.Data[valueOffset:valueOffset+len(value)], value)

	// Update node metadata
	node.NKeys++
	node.TableSpace.Len += uint16(KVLocSize)
	node.FreeSpace.Off = uint16(valueOffset + len(value)) // Update free space pointer
	node.FreeSpace.Len -= uint16(len(key) + len(value) + KVLocSize)

	return nil
}

// DeleteKeyValue removes a key/value pair from a leaf node
func DeleteKeyValue(node *types.BTNodePhys, key []byte, compare func(a, b []byte) int) error {
	if !node.IsLeaf() {
		return fmt.Errorf("deletion only supported on leaf nodes")
	}

	// Find the key
	index, found, err := SearchBTreeNodePhys(node, key, compare)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("key not found for deletion")
	}

	// Calculate location in table of contents
	tocStart := int(node.TableSpace.Off)
	entryOffset := tocStart + index*KVLocSize

	// Shift remaining entries in the table of contents
	if index < int(node.NKeys)-1 {
		copy(node.Data[entryOffset:], node.Data[entryOffset+KVLocSize:tocStart+int(node.TableSpace.Len)])
	}

	// Update metadata
	node.NKeys--
	node.TableSpace.Len -= uint16(KVLocSize)

	// Note: We don't actually remove the key/value data - just the reference to it
	// This is consistent with the APFS documentation where free space is managed through free lists

	return nil
}

// TraverseBTree recursively traverses a B-tree and executes a callback for each leaf node encountered
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
