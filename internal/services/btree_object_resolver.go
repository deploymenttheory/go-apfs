package services

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/parsers/btrees"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// BTreeObjectResolver resolves virtual object IDs using B-tree traversal
type BTreeObjectResolver struct {
	container *ContainerReader
}

// NewBTreeObjectResolver creates a new B-tree based object resolver
func NewBTreeObjectResolver(container *ContainerReader) *BTreeObjectResolver {
	return &BTreeObjectResolver{
		container: container,
	}
}

// ResolveVirtualObject resolves a virtual object ID to its physical address using B-tree traversal
func (btor *BTreeObjectResolver) ResolveVirtualObject(virtualOID types.OidT, transactionID types.XidT) (types.Paddr, error) {
	if btor.container == nil {
		return 0, fmt.Errorf("container reader is nil")
	}

	containerSB := btor.container.GetSuperblock()
	if containerSB == nil {
		return 0, fmt.Errorf("container superblock is nil")
	}

	// Get the container's object map OID (this is a physical OID)
	omapOID := containerSB.NxOmapOid
	if omapOID == 0 {
		return 0, fmt.Errorf("container object map OID is zero")
	}

	// Read and parse the object map
	omapData, err := btor.container.ReadBlock(uint64(omapOID))
	if err != nil {
		return 0, fmt.Errorf("failed to read object map at block %d: %w", omapOID, err)
	}

	omap, err := btor.parseObjectMapHeader(omapData, binary.LittleEndian)
	if err != nil {
		return 0, fmt.Errorf("failed to parse object map header: %w", err)
	}

	// Check if this is a manually managed object map (no B-tree)
	if omap.OmTreeOid == 0 {
		// Try manually managed object map parsing
		return btor.searchManuallyManagedObjectMap(omapData, virtualOID, transactionID)
	}

	// This object map uses a B-tree - traverse it to find the mapping
	fmt.Printf("DEBUG: Object map uses B-tree at OID %d, searching for virtual OID %d\n", omap.OmTreeOid, virtualOID)
	return btor.searchBTreeObjectMap(omap.OmTreeOid, virtualOID, transactionID)
}

// parseObjectMapHeader parses the object map header from raw data
func (btor *BTreeObjectResolver) parseObjectMapHeader(data []byte, endian binary.ByteOrder) (*types.OmapPhysT, error) {
	if len(data) < 72 {
		return nil, fmt.Errorf("insufficient data for object map header")
	}

	omap := &types.OmapPhysT{}

	// Parse object header (first 32 bytes)
	copy(omap.OmO.OChecksum[:], data[0:8])
	omap.OmO.OOid = types.OidT(endian.Uint64(data[8:16]))
	omap.OmO.OXid = types.XidT(endian.Uint64(data[16:24]))
	omap.OmO.OType = endian.Uint32(data[24:28])
	omap.OmO.OSubtype = endian.Uint32(data[28:32])

	// Parse object map specific fields
	offset := 32
	omap.OmFlags = endian.Uint32(data[offset : offset+4])
	offset += 4
	omap.OmSnapCount = endian.Uint32(data[offset : offset+4])
	offset += 4
	omap.OmTreeType = endian.Uint32(data[offset : offset+4])
	offset += 4
	omap.OmSnapshotTreeType = endian.Uint32(data[offset : offset+4])
	offset += 4
	omap.OmTreeOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	omap.OmSnapshotTreeOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	omap.OmMostRecentSnap = types.XidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	omap.OmPendingRevertMin = types.XidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	omap.OmPendingRevertMax = types.XidT(endian.Uint64(data[offset : offset+8]))

	return omap, nil
}

// searchManuallyManagedObjectMap searches for object mappings in a manually managed object map
func (btor *BTreeObjectResolver) searchManuallyManagedObjectMap(omapData []byte, virtualOID types.OidT, transactionID types.XidT) (types.Paddr, error) {
	// Object map header is 72 bytes, entries start after that
	entryOffset := 72
	entrySize := 32 // OmapKeyT (16 bytes) + OmapValT (16 bytes) = 32 bytes total

	for entryOffset+entrySize <= len(omapData) {
		// Parse the key (OID + XID)
		entryOID := types.OidT(binary.LittleEndian.Uint64(omapData[entryOffset : entryOffset+8]))
		entryXID := types.XidT(binary.LittleEndian.Uint64(omapData[entryOffset+8 : entryOffset+16]))

		// Safety check - if we hit all zeros, we've reached the end
		if entryOID == 0 && entryXID == 0 {
			break
		}

		// Check if this is the object we're looking for
		// For object maps, we need the OID to match and the XID to be <= transactionID
		if entryOID == virtualOID && entryXID <= transactionID {
			// Parse the value to get the physical address
			entryPaddr := types.Paddr(binary.LittleEndian.Uint64(omapData[entryOffset+24 : entryOffset+32]))
			return entryPaddr, nil
		}

		entryOffset += entrySize
	}

	return 0, fmt.Errorf("virtual object %d not found in manually managed object map", virtualOID)
}

// searchBTreeObjectMap searches for object mappings in a B-tree based object map
func (btor *BTreeObjectResolver) searchBTreeObjectMap(treeOID types.OidT, virtualOID types.OidT, transactionID types.XidT) (types.Paddr, error) {
	// The B-tree itself is stored as a virtual object, which creates a circular dependency
	// To resolve this, we need to either:
	// 1. Have a bootstrap mechanism where the root B-tree is stored physically
	// 2. Use a different resolution method for the B-tree itself
	
	// For now, we'll assume the B-tree root is stored at a physical address equal to its OID
	// This is often the case in APFS implementations
	fmt.Printf("DEBUG: Reading B-tree root from block %d\n", treeOID)
	treeData, err := btor.container.ReadBlock(uint64(treeOID))
	if err != nil {
		fmt.Printf("DEBUG: Failed to read B-tree root: %v\n", err)
		return 0, fmt.Errorf("failed to read B-tree root at block %d: %w", treeOID, err)
	}

	// Parse the B-tree node
	fmt.Printf("DEBUG: Parsing B-tree node (%d bytes)\n", len(treeData))
	
	// Debug: Check the first few bytes to understand node structure
	if len(treeData) >= 64 {
		fmt.Printf("DEBUG: Node header bytes: %02x\n", treeData[:64])
		fmt.Printf("DEBUG: Node object type: 0x%08x\n", binary.LittleEndian.Uint32(treeData[24:28]))
		fmt.Printf("DEBUG: Node flags: 0x%04x\n", binary.LittleEndian.Uint16(treeData[32:34]))
		fmt.Printf("DEBUG: Node level: %d\n", binary.LittleEndian.Uint16(treeData[34:36]))
		fmt.Printf("DEBUG: Node key count: %d\n", binary.LittleEndian.Uint32(treeData[36:40]))
	}
	
	nodeReader, err := btrees.NewBTreeNodeReader(treeData, binary.LittleEndian)
	if err != nil {
		fmt.Printf("DEBUG: Failed to parse B-tree node: %v\n", err)
		return 0, fmt.Errorf("failed to parse B-tree node: %w", err)
	}

	// Create the search key
	searchKey := types.OmapKeyT{
		OkOid: virtualOID,
		OkXid: transactionID,
	}

	fmt.Printf("DEBUG: Searching B-tree for key OID=%d, XID=%d\n", searchKey.OkOid, searchKey.OkXid)

	// Search the B-tree for the mapping
	result, err := btor.searchBTreeNode(nodeReader, searchKey)
	if err != nil {
		fmt.Printf("DEBUG: B-tree search failed: %v\n", err)
	} else {
		fmt.Printf("DEBUG: B-tree search succeeded: physical address %d\n", result)
	}
	return result, err
}

// searchBTreeNode recursively searches a B-tree node for an object mapping
func (btor *BTreeObjectResolver) searchBTreeNode(nodeReader interface{}, searchKey types.OmapKeyT) (types.Paddr, error) {
	// Cast the interface to the actual B-tree node reader
	node, ok := nodeReader.(interfaces.BTreeNodeReader)
	if !ok {
		return 0, fmt.Errorf("invalid node reader type")
	}

	// Check if this is a leaf node
	if node.IsLeaf() {
		// Search for the key in this leaf node
		return btor.searchLeafNode(node, searchKey)
	} else {
		// This is an internal node - find the appropriate child to follow
		childOID, err := btor.findChildNode(node, searchKey)
		if err != nil {
			return 0, fmt.Errorf("failed to find child node: %w", err)
		}

		// Recursively search the child node
		return btor.searchChildNode(childOID, searchKey)
	}
}

// searchLeafNode searches for a key in a leaf node and returns the associated physical address
func (btor *BTreeObjectResolver) searchLeafNode(node interfaces.BTreeNodeReader, searchKey types.OmapKeyT) (types.Paddr, error) {
	// Get node data
	nodeData := node.Data()
	keyCount := node.KeyCount()

	// Parse the table of contents to find key/value pairs
	entries, err := btor.parseTableOfContents(node, nodeData)
	if err != nil {
		return 0, fmt.Errorf("failed to parse table of contents: %w", err)
	}

	// Search for the key (with debug output)
	fmt.Printf("DEBUG: Leaf node has %d keys, %d entries\n", keyCount, len(entries))
	for i := uint32(0); i < keyCount; i++ {
		if i >= uint32(len(entries)) {
			break
		}

		// Extract key and value from the entry
		key, value, err := btor.extractKeyValue(nodeData, entries[i], node.HasFixedKVSize())
		if err != nil {
			fmt.Printf("DEBUG: Entry %d: Failed to extract key/value: %v\n", i, err)
			continue // Skip malformed entries
		}

		// Parse the key as an object map key
		if len(key) >= 16 {
			fmt.Printf("DEBUG: Entry %d: Raw key bytes: %02x\n", i, key[:16])
			entryOID := types.OidT(binary.LittleEndian.Uint64(key[0:8]))
			entryXID := types.XidT(binary.LittleEndian.Uint64(key[8:16]))

			fmt.Printf("DEBUG: Entry %d: OID=%d (0x%016x), XID=%d (searching for OID=%d, XID<=%d)\n", 
				i, entryOID, entryOID, entryXID, searchKey.OkOid, searchKey.OkXid)

			// Check if this matches our search criteria
			// For object maps, we want exact OID match and XID <= search XID
			if entryOID == searchKey.OkOid && entryXID <= searchKey.OkXid {
				// Parse the value as an object map value to get physical address
				if len(value) >= 16 {
					physAddr := types.Paddr(binary.LittleEndian.Uint64(value[8:16])) // paddr is at offset 8
					fmt.Printf("DEBUG: Found matching entry! Physical address: %d\n", physAddr)
					return physAddr, nil
				}
			}
		} else {
			fmt.Printf("DEBUG: Entry %d: Key too short (%d bytes)\n", i, len(key))
		}
	}

	return 0, fmt.Errorf("key not found in leaf node")
}

// findChildNode finds the appropriate child node to follow in an internal node
func (btor *BTreeObjectResolver) findChildNode(node interfaces.BTreeNodeReader, searchKey types.OmapKeyT) (types.OidT, error) {
	// Get node data
	nodeData := node.Data()
	keyCount := node.KeyCount()

	// Parse the table of contents
	entries, err := btor.parseTableOfContents(node, nodeData)
	if err != nil {
		return 0, fmt.Errorf("failed to parse table of contents: %w", err)
	}

	// Find the correct child by comparing keys
	// In B-trees, keys in internal nodes are separators
	for i := uint32(0); i < keyCount; i++ {
		if i >= uint32(len(entries)) {
			break
		}

		// Extract key and value
		key, value, err := btor.extractKeyValue(nodeData, entries[i], node.HasFixedKVSize())
		if err != nil {
			continue
		}

		// Parse the key
		if len(key) >= 16 {
			entryOID := types.OidT(binary.LittleEndian.Uint64(key[0:8]))
			entryXID := types.XidT(binary.LittleEndian.Uint64(key[8:16]))

			// If our search key is less than or equal to this entry's key, 
			// we should follow this child
			if btor.compareKeys(searchKey, types.OmapKeyT{OkOid: entryOID, OkXid: entryXID}) <= 0 {
				// The value in internal nodes contains the child OID
				if len(value) >= 8 {
					childOID := types.OidT(binary.LittleEndian.Uint64(value[0:8]))
					return childOID, nil
				}
			}
		}
	}

	// If we didn't find a suitable child, take the rightmost child
	// This happens when our search key is larger than all separator keys
	if keyCount > 0 && uint32(len(entries)) > 0 {
		lastEntry := entries[keyCount-1]
		_, value, err := btor.extractKeyValue(nodeData, lastEntry, node.HasFixedKVSize())
		if err == nil && len(value) >= 8 {
			childOID := types.OidT(binary.LittleEndian.Uint64(value[0:8]))
			return childOID, nil
		}
	}

	return 0, fmt.Errorf("no suitable child node found")
}

// searchChildNode loads and searches a child node
func (btor *BTreeObjectResolver) searchChildNode(childOID types.OidT, searchKey types.OmapKeyT) (types.Paddr, error) {
	// Read the child node (this may require recursive virtual object resolution)
	// For now, assume child nodes are stored physically at their OID
	childData, err := btor.container.ReadBlock(uint64(childOID))
	if err != nil {
		return 0, fmt.Errorf("failed to read child node at block %d: %w", childOID, err)
	}

	// Parse the child node
	childNodeReader, err := btrees.NewBTreeNodeReader(childData, binary.LittleEndian)
	if err != nil {
		return 0, fmt.Errorf("failed to parse child node: %w", err)
	}

	// Recursively search the child node
	return btor.searchBTreeNode(childNodeReader, searchKey)
}

// parseTableOfContents parses the table of contents from a B-tree node
func (btor *BTreeObjectResolver) parseTableOfContents(node interfaces.BTreeNodeReader, nodeData []byte) ([]tableEntry, error) {
	keyCount := node.KeyCount()
	tableSpace := node.TableSpace()

	// The table offset is relative to the btn_data field, which starts after the btree_node_phys_t header
	// btree_node_phys_t header is 56 bytes, so btn_data starts at offset 56
	btnDataStart := 56
	tableOffset := btnDataStart + int(tableSpace.Off)
	
	fmt.Printf("DEBUG: Table parsing - keyCount=%d, tableSpace.Off=%d, tableOffset=%d, nodeDataLen=%d\n", 
		keyCount, tableSpace.Off, tableOffset, len(nodeData))
	
	// Debug: Show what's at the start of btn_data
	if len(nodeData) > btnDataStart+32 {
		fmt.Printf("DEBUG: btn_data start (offset %d): %02x\n", btnDataStart, nodeData[btnDataStart:btnDataStart+32])
	}
	
	if tableOffset >= len(nodeData) {
		return nil, fmt.Errorf("table offset %d exceeds node data length %d", tableOffset, len(nodeData))
	}

	var entries []tableEntry

	if node.HasFixedKVSize() {
		fmt.Printf("DEBUG: Fixed-size key/value table\n")
		// Fixed-size keys and values - table contains kvoff_t entries
		entrySize := 4 // kvoff_t is 4 bytes (2 uint16s)
		for i := uint32(0); i < keyCount; i++ {
			offset := tableOffset + int(i)*entrySize
			if offset+entrySize > len(nodeData) {
				break
			}
			
			keyOffset := binary.LittleEndian.Uint16(nodeData[offset : offset+2])
			valueOffset := binary.LittleEndian.Uint16(nodeData[offset+2 : offset+4])
			
			fmt.Printf("DEBUG: Entry %d: keyOffset=%d, valueOffset=%d\n", i, keyOffset, valueOffset)
			
			entries = append(entries, tableEntry{
				keyOffset:   keyOffset,
				valueOffset: valueOffset,
				isFixedSize: true,
			})
		}
	} else {
		fmt.Printf("DEBUG: Variable-size key/value table\n")
		// Variable-size keys and values - table contains kvloc_t entries
		entrySize := 8 // kvloc_t is 8 bytes (2 nloc_t, each 4 bytes)
		for i := uint32(0); i < keyCount; i++ {
			offset := tableOffset + int(i)*entrySize
			if offset+entrySize > len(nodeData) {
				break
			}
			
			keyOffset := binary.LittleEndian.Uint16(nodeData[offset : offset+2])
			keyLen := binary.LittleEndian.Uint16(nodeData[offset+2 : offset+4])
			valueOffset := binary.LittleEndian.Uint16(nodeData[offset+4 : offset+6])
			valueLen := binary.LittleEndian.Uint16(nodeData[offset+6 : offset+8])
			
			fmt.Printf("DEBUG: Entry %d: keyOffset=%d, keyLen=%d, valueOffset=%d, valueLen=%d\n", 
				i, keyOffset, keyLen, valueOffset, valueLen)
			
			entries = append(entries, tableEntry{
				keyOffset:   keyOffset,
				keyLen:      keyLen,
				valueOffset: valueOffset,
				valueLen:    valueLen,
				isFixedSize: false,
			})
		}
	}

	return entries, nil
}

// tableEntry represents an entry in the table of contents
type tableEntry struct {
	keyOffset   uint16
	keyLen      uint16
	valueOffset uint16
	valueLen    uint16
	isFixedSize bool
}

// extractKeyValue extracts the key and value data from a table entry
func (btor *BTreeObjectResolver) extractKeyValue(nodeData []byte, entry tableEntry, isFixedSize bool) ([]byte, []byte, error) {
	var key, value []byte

	// Key and value offsets are relative to the btn_data field (starts at offset 56)
	btnDataStart := 56

	// Extract key
	keyStart := btnDataStart + int(entry.keyOffset)
	if isFixedSize {
		// For fixed-size, we need to determine the size from the B-tree info
		// For object maps, keys are 16 bytes (OID + XID)
		keyEnd := keyStart + 16
		if keyEnd > len(nodeData) {
			return nil, nil, fmt.Errorf("key extends beyond node data: keyStart=%d, keyEnd=%d, nodeDataLen=%d", 
				keyStart, keyEnd, len(nodeData))
		}
		key = nodeData[keyStart:keyEnd]
	} else {
		keyEnd := keyStart + int(entry.keyLen)
		if keyEnd > len(nodeData) {
			return nil, nil, fmt.Errorf("key extends beyond node data: keyStart=%d, keyEnd=%d, nodeDataLen=%d", 
				keyStart, keyEnd, len(nodeData))
		}
		key = nodeData[keyStart:keyEnd]
	}

	// Extract value
	valueStart := btnDataStart + int(entry.valueOffset)
	if isFixedSize {
		// For object maps, values are 16 bytes (flags + size + paddr)
		valueEnd := valueStart + 16
		if valueEnd > len(nodeData) {
			return nil, nil, fmt.Errorf("value extends beyond node data: valueStart=%d, valueEnd=%d, nodeDataLen=%d", 
				valueStart, valueEnd, len(nodeData))
		}
		value = nodeData[valueStart:valueEnd]
	} else {
		valueEnd := valueStart + int(entry.valueLen)
		if valueEnd > len(nodeData) {
			return nil, nil, fmt.Errorf("value extends beyond node data: valueStart=%d, valueEnd=%d, nodeDataLen=%d", 
				valueStart, valueEnd, len(nodeData))
		}
		value = nodeData[valueStart:valueEnd]
	}

	fmt.Printf("DEBUG: Extracted key (%d bytes) from offset %d: %02x\n", len(key), keyStart, key)
	fmt.Printf("DEBUG: Extracted value (%d bytes) from offset %d: %02x\n", len(value), valueStart, value)

	return key, value, nil
}

// compareKeys compares two object map keys for ordering
func (btor *BTreeObjectResolver) compareKeys(key1, key2 types.OmapKeyT) int {
	// First compare by OID
	if key1.OkOid < key2.OkOid {
		return -1
	} else if key1.OkOid > key2.OkOid {
		return 1
	}
	
	// If OIDs are equal, compare by XID
	if key1.OkXid < key2.OkXid {
		return -1
	} else if key1.OkXid > key2.OkXid {
		return 1
	}
	
	return 0 // Keys are equal
}