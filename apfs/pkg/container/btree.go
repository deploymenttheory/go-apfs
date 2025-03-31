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

// WriteBTreeNodePhys writes a BTreeNodePhys back to the device with proper checksum calculation.
func WriteBTreeNodePhys(device types.BlockDevice, addr types.PAddr, node *types.BTNodePhys) error {
	blockSize := device.GetBlockSize()

	// Create a new buffer for the complete node data
	data := make([]byte, blockSize)

	// Copy header fields to data buffer
	// Zero out the checksum field which will be calculated
	for i := 0; i < 8; i++ {
		data[i] = 0
	}

	// Copy OID, XID, Type, Subtype from the original header
	// The provided node.Data should be the raw data without the object header
	// So we need to preserve the header fields from the original block

	// Set node-specific fields
	binary.LittleEndian.PutUint16(data[32:34], node.Flags)
	binary.LittleEndian.PutUint16(data[34:36], node.Level)
	binary.LittleEndian.PutUint32(data[36:40], node.NKeys)

	// Table space
	binary.LittleEndian.PutUint16(data[40:42], node.TableSpace.Off)
	binary.LittleEndian.PutUint16(data[42:44], node.TableSpace.Len)

	// Free space
	binary.LittleEndian.PutUint16(data[44:46], node.FreeSpace.Off)
	binary.LittleEndian.PutUint16(data[46:48], node.FreeSpace.Len)

	// Key free list
	binary.LittleEndian.PutUint16(data[48:50], node.KeyFreeList.Off)
	binary.LittleEndian.PutUint16(data[50:52], node.KeyFreeList.Len)

	// Value free list
	binary.LittleEndian.PutUint16(data[52:54], node.ValFreeList.Off)
	binary.LittleEndian.PutUint16(data[54:56], node.ValFreeList.Len)

	// Copy the Data section
	if len(node.Data) > 0 {
		copy(data[BTNodePhysSize:], node.Data)
	}

	// Calculate the checksum
	computedChecksum := checksum.Fletcher64WithZeroedChecksum(data, 0)

	// Update the checksum in the data buffer
	binary.LittleEndian.PutUint64(data[0:8], computedChecksum)

	// Write the buffer to the device
	if err := device.WriteBlock(addr, data); err != nil {
		return fmt.Errorf("failed to write B-tree node at address %d: %w", addr, err)
	}

	return nil
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

// InitializeEmptyNode sets up a new empty B-tree node with properly initialized layout.
// It allocates the node's data buffer to match the device's block size for compatibility with APFS.
// If the node is a leaf, the leaf flag is applied. Root/non-root and level can be set later.
func InitializeEmptyNode(node *types.BTNodePhys, isLeaf bool, blockSize int) {
	if blockSize < BTNodePhysSize {
		blockSize = 4096 // Fallback to standard APFS minimum block size
	}

	node.Data = make([]byte, blockSize-BTNodePhysSize)

	// Set flags
	node.Flags = 0
	if isLeaf {
		node.Flags |= types.BTNodeLeaf
	}

	// Common fields
	node.Level = 0
	node.NKeys = 0

	// Table of contents (KVLoc table) goes at the start of node.Data
	node.TableSpace.Off = 0
	node.TableSpace.Len = 0

	// Free space starts immediately after the initial TOC area
	tocStart := int(node.TableSpace.Off)
	keyAreaStart := tocStart + 0 // Initially 0-length TOC

	// Value area starts at the end of Data and grows backwards
	valueAreaEnd := len(node.Data)

	// Free space is everything between key area and value area
	node.FreeSpace.Off = uint16(keyAreaStart)
	node.FreeSpace.Len = uint16(valueAreaEnd - keyAreaStart)

	// Free lists start empty
	node.KeyFreeList.Off = BTOFFInvalid
	node.KeyFreeList.Len = 0
	node.ValFreeList.Off = BTOFFInvalid
	node.ValFreeList.Len = 0
}

// InsertKeyValueLeaf inserts a key/value pair into a leaf node.
// It updates the node's Table of Contents (TOC), stores the key/value data,
// and adjusts the node's free space accordingly.
func InsertKeyValueLeaf(node *types.BTNodePhys, key, value []byte) error {
	if node == nil {
		return fmt.Errorf("node is nil")
	}
	if !node.IsLeaf() {
		return fmt.Errorf("insert only supported on leaf nodes")
	}
	if node.Data == nil {
		return fmt.Errorf("node data buffer not initialized")
	}

	keyLen := len(key)
	valLen := len(value)
	requiredKey := keyLen
	requiredVal := valLen

	// Allocate key space
	keyOff, newKeyFree, ok := allocateFromFreeList(node.Data, node.KeyFreeList, requiredKey)
	if !ok {
		keyOff = int(node.FreeSpace.Off)
	}
	node.KeyFreeList = newKeyFree

	// Allocate value space
	valOff, newValFree, ok := allocateFromFreeList(node.Data, node.ValFreeList, requiredVal)
	if !ok {
		valOff = int(node.FreeSpace.Off) + keyLen
	}
	node.ValFreeList = newValFree

	// Update FreeSpace if we’re using it
	end := valOff + valLen
	if end > len(node.Data) {
		return fmt.Errorf("not enough free space")
	}
	if valOff+valLen > int(node.FreeSpace.Off)+int(node.FreeSpace.Len) {
		return fmt.Errorf("exceeds node FreeSpace")
	}

	// Find insert index for sorted order
	insertIdx := 0
	for i := 0; i < int(node.NKeys); i++ {
		existingKey, err := GetKeyAtIndex(node, i)
		if err != nil {
			break
		}
		if bytes.Compare(key, existingKey) > 0 {
			insertIdx = i + 1
		} else {
			break
		}
	}

	// Build TOC entry
	tocEntry := make([]byte, KVLocSize)
	binary.LittleEndian.PutUint16(tocEntry[0:2], uint16(keyOff))
	binary.LittleEndian.PutUint16(tocEntry[2:4], uint16(keyLen))
	binary.LittleEndian.PutUint16(tocEntry[4:6], uint16(valOff))
	binary.LittleEndian.PutUint16(tocEntry[6:8], uint16(valLen))

	tocInsertOffset := int(node.TableSpace.Off) + insertIdx*KVLocSize
	endOfTOC := int(node.TableSpace.Off) + int(node.TableSpace.Len)

	if tocInsertOffset > endOfTOC || endOfTOC+KVLocSize > len(node.Data) {
		return fmt.Errorf("TOC overflow")
	}

	// Shift TOC entries
	if insertIdx < int(node.NKeys) {
		copy(node.Data[tocInsertOffset+KVLocSize:], node.Data[tocInsertOffset:endOfTOC])
	}
	copy(node.Data[tocInsertOffset:tocInsertOffset+KVLocSize], tocEntry)

	// Write key and value
	copy(node.Data[keyOff:keyOff+keyLen], key)
	copy(node.Data[valOff:valOff+valLen], value)

	// Update metadata
	node.NKeys++
	node.TableSpace.Len += KVLocSize
	node.FreeSpace.Off = uint16(end)
	node.FreeSpace.Len = uint16(len(node.Data) - end)

	return nil
}

func allocateFromFreeList(data []byte, head types.NLoc, length int) (offset int, updatedHead types.NLoc, ok bool) {
	cur := int(head.Off)

	for cur != int(BTOFFInvalid) && cur+8 <= len(data) {
		entry := data[cur : cur+8] // Slice of the free_node

		regionOffset := int(binary.LittleEndian.Uint16(entry[0:2]))
		regionLen := int(binary.LittleEndian.Uint16(entry[2:4]))
		next := int(binary.LittleEndian.Uint16(entry[4:6]))

		if regionLen >= length {
			remaining := regionLen - length

			if remaining >= 8 {
				// Shrink this entry in-place
				newOffset := regionOffset + length
				binary.LittleEndian.PutUint16(entry[0:2], uint16(newOffset))
				binary.LittleEndian.PutUint16(entry[2:4], uint16(remaining))
				return regionOffset, head, true
			}

			// Remove this entry from the list
			return regionOffset, types.NLoc{Off: uint16(next), Len: 0}, true
		}

		cur = next
	}

	return 0, head, false
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

	// Locate TOC entry
	tocOffset := int(node.TableSpace.Off) + index*KVLocSize
	if tocOffset+KVLocSize > len(node.Data) {
		return fmt.Errorf("TOC entry out of bounds")
	}

	// Read key/value offsets and lengths from TOC
	kvloc := &types.KVLoc{
		K: types.NLoc{
			Off: binary.LittleEndian.Uint16(node.Data[tocOffset:]),
			Len: binary.LittleEndian.Uint16(node.Data[tocOffset+2:]),
		},
		V: types.NLoc{
			Off: binary.LittleEndian.Uint16(node.Data[tocOffset+4:]),
			Len: binary.LittleEndian.Uint16(node.Data[tocOffset+6:]),
		},
	}

	// Push key and value into respective free lists
	insertFreeNode(node.Data, &node.KeyFreeList, kvloc.K.Off, kvloc.K.Len)
	insertFreeNode(node.Data, &node.ValFreeList, kvloc.V.Off, kvloc.V.Len)

	// Shift TOC entries to remove this one
	endOfTOC := int(node.TableSpace.Off) + int(node.TableSpace.Len)
	if index < int(node.NKeys)-1 {
		copy(node.Data[tocOffset:], node.Data[tocOffset+KVLocSize:endOfTOC])
	}

	// Zero out the trailing entry (optional cleanup)
	copy(node.Data[endOfTOC-KVLocSize:endOfTOC], make([]byte, KVLocSize))

	// Update metadata
	node.NKeys--
	node.TableSpace.Len -= KVLocSize

	return nil
}

func insertFreeNode(data []byte, list *types.NLoc, off, length uint16) {
	// Find place to write the new free_node — use FreeSpace.Off
	freeNodeOffset := int(list.Off)
	if freeNodeOffset == int(BTOFFInvalid) || freeNodeOffset+8 > len(data) {
		// Allocate from FreeSpace
		freeNodeOffset = int(binary.LittleEndian.Uint16(data[44:46])) // node.FreeSpace.Off
		if freeNodeOffset+8 > len(data) {
			return // no room — ignore for now
		}
	}

	// Write new free_node struct
	binary.LittleEndian.PutUint16(data[freeNodeOffset:], off)
	binary.LittleEndian.PutUint16(data[freeNodeOffset+2:], length)
	binary.LittleEndian.PutUint16(data[freeNodeOffset+4:], list.Off) // next = old head
	binary.LittleEndian.PutUint16(data[freeNodeOffset+6:], 0)        // padding

	// Update head of list
	list.Off = uint16(freeNodeOffset)
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

// LookupBTree finds a key in the B-tree, traversing from root to leaf.
func LookupBTree(device types.BlockDevice, rootAddr types.PAddr, key []byte, keyCompare func(a, b []byte) int) ([]byte, bool, error) {
	// Read the root node
	node, err := ReadBTreeNodePhys(device, rootAddr)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read root node: %w", err)
	}

	// If the node is a leaf, search it directly
	if node.IsLeaf() {
		idx, found, err := SearchBTreeNodePhys(node, key, keyCompare)
		if err != nil {
			return nil, false, err
		}
		if !found {
			return nil, false, nil
		}

		// Get the value from the leaf node
		value, err := GetValueAtIndex(node, idx)
		if err != nil {
			return nil, false, err
		}
		return value, true, nil
	}

	// For non-leaf nodes, find the appropriate child and recurse
	// Perform a search to find the insertion point
	idx, _, err := SearchBTreeNodePhys(node, key, keyCompare)
	if err != nil {
		return nil, false, err
	}

	// If the key is greater than all keys in this node, we need the last child
	if idx > 0 && idx >= int(node.NKeys) {
		idx = int(node.NKeys) - 1
	}

	// Get the child node address from the value
	value, err := GetValueAtIndex(node, idx)
	if err != nil {
		return nil, false, err
	}

	// Child address is the first 8 bytes of the value in a non-leaf node
	if len(value) < 8 {
		return nil, false, fmt.Errorf("invalid child node reference: too short")
	}
	childAddr := types.PAddr(binary.LittleEndian.Uint64(value[:8]))

	// Recursively search the child node
	return LookupBTree(device, childAddr, key, keyCompare)
}

// SplitNode splits a B-tree node that has become full.
// It returns the address of the new node, the new node itself, and the middle key.
func SplitNode(device types.BlockDevice, nodeAddr types.PAddr, node *types.BTNodePhys, spaceman types.SpaceManager) (types.PAddr, *types.BTNodePhys, []byte, error) {
	// Create a new node with the same level and flags
	newNode := &types.BTNodePhys{
		Flags: node.Flags,
		Level: node.Level,
		NKeys: 0,
	}

	// Allocate a new block using the space manager
	newNodeAddr, err := spaceman.AllocateBlock()
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to allocate block for new node: %w", err)
	}

	// Initialize the new node's storage
	InitializeEmptyNode(newNode, node.IsLeaf(), int(device.GetBlockSize()))

	// Find the middle key index
	middleIdx := int(node.NKeys) / 2

	// Get the middle key for returning to the parent
	middleKey, err := GetKeyAtIndex(node, middleIdx)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to get middle key: %w", err)
	}

	// Copy the second half of keys/values to the new node
	for i := middleIdx; i < int(node.NKeys); i++ {
		key, err := GetKeyAtIndex(node, i)
		if err != nil {
			return 0, nil, nil, fmt.Errorf("failed to get key at index %d: %w", i, err)
		}

		value, err := GetValueAtIndex(node, i)
		if err != nil {
			return 0, nil, nil, fmt.Errorf("failed to get value at index %d: %w", i, err)
		}

		// Insert into new node
		if err := InsertKeyValueLeaf(newNode, key, value); err != nil {
			return 0, nil, nil, fmt.Errorf("failed to insert into new node: %w", err)
		}
	}

	// Update the original node
	// We need to reconstruct the original node with only the first half of keys
	updatedNode := &types.BTNodePhys{
		Flags: node.Flags,
		Level: node.Level,
		NKeys: 0,
	}
	InitializeEmptyNode(updatedNode, node.IsLeaf(), int(device.GetBlockSize()))

	// Copy the first half of keys/values to the updated node
	for i := 0; i < middleIdx; i++ {
		key, err := GetKeyAtIndex(node, i)
		if err != nil {
			return 0, nil, nil, fmt.Errorf("failed to get key at index %d: %w", i, err)
		}

		value, err := GetValueAtIndex(node, i)
		if err != nil {
			return 0, nil, nil, fmt.Errorf("failed to get value at index %d: %w", i, err)
		}

		// Insert into updated node
		if err := InsertKeyValueLeaf(updatedNode, key, value); err != nil {
			return 0, nil, nil, fmt.Errorf("failed to insert into updated node: %w", err)
		}
	}

	// Write both nodes to disk
	if err := WriteBTreeNodePhys(device, nodeAddr, updatedNode); err != nil {
		return 0, nil, nil, fmt.Errorf("failed to write updated original node: %w", err)
	}

	if err := WriteBTreeNodePhys(device, newNodeAddr, newNode); err != nil {
		return 0, nil, nil, fmt.Errorf("failed to write new node: %w", err)
	}

	return newNodeAddr, newNode, middleKey, nil
}

// InsertBTree inserts a key-value pair into the B-tree, handling splits as needed.
// It returns the address of the root node (which may change if the root splits).
func InsertBTree(device types.BlockDevice, rootAddr types.PAddr, key, value []byte, keyCompare func(a, b []byte) int, spaceman types.SpaceManager) (types.PAddr, error) {
	// Read the root node
	rootNode, err := ReadBTreeNodePhys(device, rootAddr)
	if err != nil {
		return 0, fmt.Errorf("failed to read root node: %w", err)
	}

	// If root is a leaf, insert directly and handle potential split
	if rootNode.IsLeaf() {
		idx, found, err := SearchBTreeNodePhys(rootNode, key, keyCompare)
		if err != nil {
			return 0, fmt.Errorf("failed to search node: %w", err)
		}

		if found {
			kvloc, err := GetKeyValueLocation(rootNode, idx)
			if err != nil {
				return 0, fmt.Errorf("failed to get kvloc: %w", err)
			}

			existingSize := int(kvloc.K.Len) + int(kvloc.V.Len)
			newSize := len(key) + len(value)

			if newSize <= existingSize {
				keyOffset := int(kvloc.K.Off)
				valOffset := int(kvloc.V.Off)
				copy(rootNode.Data[keyOffset:keyOffset+len(key)], key)
				copy(rootNode.Data[valOffset:valOffset+len(value)], value)

				if len(key) != int(kvloc.K.Len) {
					tocOffset := int(rootNode.TableSpace.Off) + idx*KVLocSize + 2
					binary.LittleEndian.PutUint16(rootNode.Data[tocOffset:], uint16(len(key)))
				}
				if len(value) != int(kvloc.V.Len) {
					tocOffset := int(rootNode.TableSpace.Off) + idx*KVLocSize + 6
					binary.LittleEndian.PutUint16(rootNode.Data[tocOffset:], uint16(len(value)))
				}
			} else {
				if err := DeleteKeyValue(rootNode, key, keyCompare); err != nil {
					return 0, fmt.Errorf("delete before reinsert failed: %w", err)
				}
				if err := InsertKeyValueLeaf(rootNode, key, value); err != nil {
					return 0, fmt.Errorf("reinsert failed: %w", err)
				}
			}
		} else {
			if err := InsertKeyValueLeaf(rootNode, key, value); err != nil {
				return 0, fmt.Errorf("insert failed: %w", err)
			}
		}

		estimatedSize := len(key) + len(value) + KVLocSize
		pad := (KVAlignment - (estimatedSize % KVAlignment)) % KVAlignment
		required := uint16(estimatedSize + pad + KVLocSize)

		if rootNode.FreeSpace.Len < required {
			newRightAddr, newRightNode, middleKey, err := SplitNode(device, rootAddr, rootNode, spaceman)
			if err != nil {
				return 0, fmt.Errorf("split failed: %w", err)
			}

			// Write the right node (post-split)
			if err := WriteBTreeNodePhys(device, newRightAddr, newRightNode); err != nil {
				return 0, fmt.Errorf("write right node failed: %w", err)
			}

			newRootNode := &types.BTNodePhys{}
			InitializeEmptyNode(newRootNode, false, int(device.GetBlockSize()))
			newRootNode.Level = rootNode.Level + 1
			newRootNode.Flags = (rootNode.Flags & ^uint16(types.BTNodeLeaf)) | types.BTNodeRoot

			newRootAddr, err := spaceman.AllocateBlock()
			if err != nil {
				return 0, fmt.Errorf("alloc new root failed: %w", err)
			}

			leftPtr := make([]byte, 8)
			binary.LittleEndian.PutUint64(leftPtr, uint64(rootAddr))
			if err := InsertKeyValueLeaf(newRootNode, middleKey, leftPtr); err != nil {
				return 0, fmt.Errorf("insert left ptr failed: %w", err)
			}

			rightPtr := make([]byte, 8)
			binary.LittleEndian.PutUint64(rightPtr, uint64(newRightAddr))
			if err := InsertKeyValueLeaf(newRootNode, middleKey, rightPtr); err != nil {
				return 0, fmt.Errorf("insert right ptr failed: %w", err)
			}

			if err := WriteBTreeNodePhys(device, newRootAddr, newRootNode); err != nil {
				return 0, fmt.Errorf("write new root failed: %w", err)
			}

			return newRootAddr, nil
		}

		if err := WriteBTreeNodePhys(device, rootAddr, rootNode); err != nil {
			return 0, fmt.Errorf("write updated root failed: %w", err)
		}
		return rootAddr, nil
	}

	// Non-leaf path
	idx, exact, err := SearchBTreeNodePhys(rootNode, key, keyCompare)
	if err != nil {
		return 0, fmt.Errorf("search failed: %w", err)
	}

	childIdx := idx
	if exact && childIdx < int(rootNode.NKeys)-1 {
		childIdx++
	}
	if childIdx >= int(rootNode.NKeys) {
		childIdx = int(rootNode.NKeys) - 1
	}

	childVal, err := GetValueAtIndex(rootNode, childIdx)
	if err != nil {
		return 0, fmt.Errorf("get child ptr failed: %w", err)
	}
	if len(childVal) < 8 {
		return 0, fmt.Errorf("child ptr too short")
	}
	childAddr := types.PAddr(binary.LittleEndian.Uint64(childVal[:8]))

	newChildAddr, err := InsertBTree(device, childAddr, key, value, keyCompare, spaceman)
	if err != nil {
		return 0, fmt.Errorf("recursive insert failed: %w", err)
	}

	if newChildAddr != childAddr {
		newChildVal := make([]byte, 8)
		binary.LittleEndian.PutUint64(newChildVal, uint64(newChildAddr))

		keyAtIdx, err := GetKeyAtIndex(rootNode, childIdx)
		if err != nil {
			return 0, fmt.Errorf("get key for update failed: %w", err)
		}

		if err := DeleteKeyValue(rootNode, keyAtIdx, keyCompare); err != nil {
			return 0, fmt.Errorf("delete child ref failed: %w", err)
		}
		if err := InsertKeyValueLeaf(rootNode, keyAtIdx, newChildVal); err != nil {
			return 0, fmt.Errorf("insert child ref failed: %w", err)
		}

		estimatedSize := len(keyAtIdx) + 8 + KVLocSize
		pad := (KVAlignment - (estimatedSize % KVAlignment)) % KVAlignment
		required := uint16(estimatedSize + pad + KVLocSize)

		if rootNode.FreeSpace.Len < required {
			newRightAddr, newRightNode, middleKey, err := SplitNode(device, rootAddr, rootNode, spaceman)
			if err != nil {
				return 0, fmt.Errorf("split non-leaf failed: %w", err)
			}

			if err := WriteBTreeNodePhys(device, newRightAddr, newRightNode); err != nil {
				return 0, fmt.Errorf("write split non-leaf failed: %w", err)
			}

			newRootNode := &types.BTNodePhys{}
			InitializeEmptyNode(newRootNode, false, int(device.GetBlockSize()))
			newRootNode.Level = rootNode.Level + 1
			newRootNode.Flags = (rootNode.Flags & ^uint16(types.BTNodeLeaf)) | types.BTNodeRoot

			newRootAddr, err := spaceman.AllocateBlock()
			if err != nil {
				return 0, fmt.Errorf("alloc new root failed: %w", err)
			}

			leftPtr := make([]byte, 8)
			binary.LittleEndian.PutUint64(leftPtr, uint64(rootAddr))

			rightPtr := make([]byte, 8)
			binary.LittleEndian.PutUint64(rightPtr, uint64(newRightAddr))

			if err := InsertKeyValueLeaf(newRootNode, middleKey, leftPtr); err != nil {
				return 0, fmt.Errorf("insert left in new root failed: %w", err)
			}
			if err := InsertKeyValueLeaf(newRootNode, middleKey, rightPtr); err != nil {
				return 0, fmt.Errorf("insert right in new root failed: %w", err)
			}

			if err := WriteBTreeNodePhys(device, newRootAddr, newRootNode); err != nil {
				return 0, fmt.Errorf("write new root failed: %w", err)
			}

			return newRootAddr, nil
		}

		if err := WriteBTreeNodePhys(device, rootAddr, rootNode); err != nil {
			return 0, fmt.Errorf("write root after child update failed: %w", err)
		}
	}

	return rootAddr, nil
}

// DeleteBTree removes a key-value pair from the B-tree.
func DeleteBTree(device types.BlockDevice, rootAddr types.PAddr, key []byte, keyCompare func(a, b []byte) int, spaceman types.SpaceManager) (types.PAddr, error) {
	// Read the root node
	rootNode, err := ReadBTreeNodePhys(device, rootAddr)
	if err != nil {
		return 0, fmt.Errorf("failed to read root node: %w", err)
	}

	// If root is a leaf, delete directly
	if rootNode.IsLeaf() {
		// Check if the key exists
		_, found, err := SearchBTreeNodePhys(rootNode, key, keyCompare)
		if err != nil {
			return 0, err
		}

		if !found {
			return rootAddr, nil // Key not found, nothing to delete
		}

		// Delete the key
		if err := DeleteKeyValue(rootNode, key, keyCompare); err != nil {
			return 0, err
		}

		// Write the updated root
		if err := WriteBTreeNodePhys(device, rootAddr, rootNode); err != nil {
			return 0, err
		}

		return rootAddr, nil
	}

	// Non-leaf node: find the right child to delete from
	idx, _, err := SearchBTreeNodePhys(rootNode, key, keyCompare)
	if err != nil {
		return 0, err
	}

	// If the index is beyond the end of the array, use the last child
	if idx >= int(rootNode.NKeys) {
		idx = int(rootNode.NKeys) - 1
	}

	// Get the child node address
	childVal, err := GetValueAtIndex(rootNode, idx)
	if err != nil {
		return 0, err
	}

	if len(childVal) < 8 {
		return 0, fmt.Errorf("invalid child pointer value")
	}

	childAddr := types.PAddr(binary.LittleEndian.Uint64(childVal[:8]))

	// Recursively delete from the child
	newChildAddr, err := DeleteBTree(device, childAddr, key, keyCompare, spaceman)
	if err != nil {
		return 0, err
	}

	// Check if the child was modified
	if newChildAddr != childAddr {
		// Check if the child was removed completely
		if newChildAddr == 0 {
			// Remove this key from the parent if it has other children
			if rootNode.NKeys > 1 {
				keyAtIdx, err := GetKeyAtIndex(rootNode, idx)
				if err != nil {
					return 0, err
				}

				if err := DeleteKeyValue(rootNode, keyAtIdx, keyCompare); err != nil {
					return 0, err
				}

				// Write the updated node
				if err := WriteBTreeNodePhys(device, rootAddr, rootNode); err != nil {
					return 0, err
				}
			} else {
				// This node has no children left, so remove it too
				if err := spaceman.FreeBlock(rootAddr); err != nil {
					return 0, fmt.Errorf("failed to free block for empty node: %w", err)
				}
				return 0, nil
			}
		} else {
			// Update the pointer in the parent
			newChildVal := make([]byte, 8)
			binary.LittleEndian.PutUint64(newChildVal, uint64(newChildAddr))

			keyAtIdx, err := GetKeyAtIndex(rootNode, idx)
			if err != nil {
				return 0, err
			}

			// Remove the old entry
			if err := DeleteKeyValue(rootNode, keyAtIdx, keyCompare); err != nil {
				return 0, err
			}

			// Add the new entry with updated pointer
			if err := InsertKeyValueLeaf(rootNode, keyAtIdx, newChildVal); err != nil {
				return 0, err
			}

			// Write the updated node
			if err := WriteBTreeNodePhys(device, rootAddr, rootNode); err != nil {
				return 0, err
			}
		}
	}

	// Check if the root now has only one child and is not a leaf
	if !rootNode.IsLeaf() && rootNode.NKeys == 1 {
		// Get the single child
		childVal, err := GetValueAtIndex(rootNode, 0)
		if err != nil {
			return 0, err
		}

		if len(childVal) < 8 {
			return 0, fmt.Errorf("invalid child pointer value")
		}

		childAddr := types.PAddr(binary.LittleEndian.Uint64(childVal[:8]))

		// Free the old root
		if err := spaceman.FreeBlock(rootAddr); err != nil {
			return 0, fmt.Errorf("failed to free old root block: %w", err)
		}

		// The child becomes the new root
		return childAddr, nil
	}

	return rootAddr, nil
}

// RangeBTree finds all key-value pairs within a specified range.
func RangeBTree(device types.BlockDevice, rootAddr types.PAddr, startKey, endKey []byte, keyCompare func(a, b []byte) int) ([][2][]byte, error) {
	results := make([][2][]byte, 0)

	// Function to collect matching entries from leaf nodes
	collectEntries := func(node *types.BTNodePhys) error {
		if !node.IsLeaf() {
			return nil // Skip non-leaf nodes
		}

		for i := 0; i < int(node.NKeys); i++ {
			key, err := GetKeyAtIndex(node, i)
			if err != nil {
				return err
			}

			// Check if key is in range
			if (startKey == nil || keyCompare(key, startKey) >= 0) &&
				(endKey == nil || keyCompare(key, endKey) <= 0) {
				value, err := GetValueAtIndex(node, i)
				if err != nil {
					return err
				}

				// Add to results
				results = append(results, [2][]byte{key, value})
			}
		}
		return nil
	}

	// Define a recursive function to traverse the tree
	var traverse func(addr types.PAddr) error
	traverse = func(addr types.PAddr) error {
		node, err := ReadBTreeNodePhys(device, addr)
		if err != nil {
			return fmt.Errorf("failed to read node at address %d: %w", addr, err)
		}

		if node.IsLeaf() {
			return collectEntries(node)
		}

		// For non-leaf nodes, determine which children to traverse
		for i := 0; i < int(node.NKeys); i++ {
			// Get the child node address
			childVal, err := GetValueAtIndex(node, i)
			if err != nil {
				return err
			}

			if len(childVal) < 8 {
				return fmt.Errorf("invalid child pointer value at index %d", i)
			}

			childAddr := types.PAddr(binary.LittleEndian.Uint64(childVal[:8]))

			// Determine if this child might contain keys in our range
			shouldTraverse := true

			if i < int(node.NKeys)-1 && startKey != nil {
				// Check if all keys in this subtree are less than startKey
				nextKey, err := GetKeyAtIndex(node, i+1)
				if err != nil {
					return err
				}

				if keyCompare(nextKey, startKey) <= 0 {
					shouldTraverse = false // Skip this child
				}
			}

			if i > 0 && endKey != nil {
				// Check if all keys in this subtree are greater than endKey
				prevKey, err := GetKeyAtIndex(node, i)
				if err != nil {
					return err
				}

				if keyCompare(prevKey, endKey) > 0 {
					shouldTraverse = false // Skip this child
				}
			}

			if shouldTraverse {
				// Recursively traverse the child
				if err := traverse(childAddr); err != nil {
					return err
				}
			}
		}

		return nil
	}

	// Start the traversal
	if err := traverse(rootAddr); err != nil {
		return nil, err
	}

	return results, nil
}
