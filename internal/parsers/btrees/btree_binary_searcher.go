package btrees

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// BinarySearcher provides binary search capabilities for B-tree nodes
type BinarySearcher struct {
	endian binary.ByteOrder
}

// NewBinarySearcher creates a new binary searcher with the specified endianness
func NewBinarySearcher(endian binary.ByteOrder) *BinarySearcher {
	if endian == nil {
		endian = binary.LittleEndian
	}
	return &BinarySearcher{endian: endian}
}

// SearchEntry represents a search result from a B-tree node
type SearchEntry struct {
	Index       int
	KeyData     []byte
	ValueData   []byte
	ChildOID    types.OidT
	IsChildNode bool
}

// FindEntryByOID performs a binary search to find an entry matching the target OID
// For internal nodes, this returns the child pointer; for leaf nodes, this returns the matching entry
func (bs *BinarySearcher) FindEntryByOID(node interfaces.BTreeNodeReader, targetOID types.OidT) (*SearchEntry, error) {
	if node.KeyCount() == 0 {
		return nil, fmt.Errorf("node has no entries")
	}

	nodeData := node.Data()
	tableSpace := node.TableSpace()
	btnDataStart := 56
	tableOffset := btnDataStart + int(tableSpace.Off)

	if tableOffset >= len(nodeData) {
		return nil, fmt.Errorf("table offset out of bounds")
	}

	// Binary search for the entry
	if node.HasFixedKVSize() {
		return bs.binarySearchFixedSize(node, nodeData, targetOID, btnDataStart, tableOffset)
	} else {
		return bs.binarySearchVariableSize(node, nodeData, targetOID, btnDataStart, tableOffset)
	}
}

// FindEntryByOIDAndXID performs a binary search for composite OID+XID key
// Used for object map B-tree searches where both OID and XID matter
func (bs *BinarySearcher) FindEntryByOIDAndXID(node interfaces.BTreeNodeReader, targetOID types.OidT, targetXID types.XidT) (*SearchEntry, error) {
	if node.KeyCount() == 0 {
		return nil, fmt.Errorf("node has no entries")
	}

	nodeData := node.Data()
	tableSpace := node.TableSpace()
	btnDataStart := 56
	tableOffset := btnDataStart + int(tableSpace.Off)

	if tableOffset >= len(nodeData) {
		return nil, fmt.Errorf("table offset out of bounds")
	}

	// For composite keys, use OID as primary sort, XID as secondary
	if node.HasFixedKVSize() {
		return bs.binarySearchFixedSizeComposite(node, nodeData, targetOID, targetXID, btnDataStart, tableOffset)
	} else {
		return bs.binarySearchVariableSizeComposite(node, nodeData, targetOID, targetXID, btnDataStart, tableOffset)
	}
}

// binarySearchFixedSize performs binary search on fixed-size kvoff_t entries
func (bs *BinarySearcher) binarySearchFixedSize(node interfaces.BTreeNodeReader, nodeData []byte, targetOID types.OidT, btnDataStart, tableOffset int) (*SearchEntry, error) {
	low := 0
	high := int(node.KeyCount()) - 1
	entrySize := 4 // kvoff_t: 2 bytes key offset + 2 bytes value offset

	for low <= high {
		mid := (low + high) / 2

		entry, err := bs.extractFixedSizeEntry(node, nodeData, mid, btnDataStart, tableOffset, entrySize)
		if err != nil {
			return nil, err
		}

		// Extract OID from key for comparison
		keyOID, err := bs.extractOIDFromKey(entry.KeyData)
		if err != nil {
			return nil, err
		}

		if keyOID < uint64(targetOID) {
			low = mid + 1
		} else if keyOID > uint64(targetOID) {
			high = mid - 1
		} else {
			// Found exact match
			entry.Index = mid
			return entry, nil
		}
	}

	// Not found, but 'high' points to the last entry with OID < targetOID
	// For internal nodes, this is the child we should follow
	if high >= 0 && high < int(node.KeyCount()) {
		entry, err := bs.extractFixedSizeEntry(node, nodeData, high, btnDataStart, tableOffset, entrySize)
		if err == nil {
			entry.Index = high
			return entry, nil
		}
	}

	return nil, fmt.Errorf("OID %d not found in node", targetOID)
}

// binarySearchFixedSizeComposite performs binary search with OID+XID comparison
func (bs *BinarySearcher) binarySearchFixedSizeComposite(node interfaces.BTreeNodeReader, nodeData []byte, targetOID types.OidT, targetXID types.XidT, btnDataStart, tableOffset int) (*SearchEntry, error) {
	low := 0
	high := int(node.KeyCount()) - 1
	entrySize := 4

	bestMatch := -1
	bestMatchEntry := (*SearchEntry)(nil)

	for low <= high {
		mid := (low + high) / 2

		entry, err := bs.extractFixedSizeEntry(node, nodeData, mid, btnDataStart, tableOffset, entrySize)
		if err != nil {
			return nil, err
		}

		// Extract OID and XID from key
		keyOID, keyXID, err := bs.extractOIDXIDFromKey(entry.KeyData)
		if err != nil {
			return nil, err
		}

		// Compare composite key: [OID, XID]
		if keyOID < uint64(targetOID) {
			low = mid + 1
		} else if keyOID > uint64(targetOID) {
			high = mid - 1
		} else {
			// OIDs match, check XID
			if keyXID <= uint64(targetXID) {
				// This is a valid candidate (XID <= target XID)
				bestMatch = mid
				bestMatchEntry = entry
				// But continue searching for potentially newer XID
				low = mid + 1
			} else {
				// keyXID > targetXID, search earlier
				high = mid - 1
			}
		}
	}

	if bestMatch >= 0 && bestMatchEntry != nil {
		bestMatchEntry.Index = bestMatch
		return bestMatchEntry, nil
	}

	return nil, fmt.Errorf("OID %d with XID <= %d not found in node", targetOID, targetXID)
}

// binarySearchVariableSize performs binary search on variable-size kvloc_t entries
func (bs *BinarySearcher) binarySearchVariableSize(node interfaces.BTreeNodeReader, nodeData []byte, targetOID types.OidT, btnDataStart, tableOffset int) (*SearchEntry, error) {
	low := 0
	high := int(node.KeyCount()) - 1
	entrySize := 8 // kvloc_t: 4x 2-byte values

	for low <= high {
		mid := (low + high) / 2

		entry, err := bs.extractVariableSizeEntry(node, nodeData, mid, btnDataStart, tableOffset, entrySize)
		if err != nil {
			return nil, err
		}

		// Extract OID from key
		keyOID, err := bs.extractOIDFromKey(entry.KeyData)
		if err != nil {
			return nil, err
		}

		if keyOID < uint64(targetOID) {
			low = mid + 1
		} else if keyOID > uint64(targetOID) {
			high = mid - 1
		} else {
			// Found exact match
			entry.Index = mid
			return entry, nil
		}
	}

	if high >= 0 && high < int(node.KeyCount()) {
		entry, err := bs.extractVariableSizeEntry(node, nodeData, high, btnDataStart, tableOffset, entrySize)
		if err == nil {
			entry.Index = high
			return entry, nil
		}
	}

	return nil, fmt.Errorf("OID %d not found in node", targetOID)
}

// binarySearchVariableSizeComposite performs binary search for variable-size composite keys
func (bs *BinarySearcher) binarySearchVariableSizeComposite(node interfaces.BTreeNodeReader, nodeData []byte, targetOID types.OidT, targetXID types.XidT, btnDataStart, tableOffset int) (*SearchEntry, error) {
	low := 0
	high := int(node.KeyCount()) - 1
	entrySize := 8

	bestMatch := -1
	bestMatchEntry := (*SearchEntry)(nil)

	for low <= high {
		mid := (low + high) / 2

		entry, err := bs.extractVariableSizeEntry(node, nodeData, mid, btnDataStart, tableOffset, entrySize)
		if err != nil {
			return nil, err
		}

		keyOID, keyXID, err := bs.extractOIDXIDFromKey(entry.KeyData)
		if err != nil {
			return nil, err
		}

		if keyOID < uint64(targetOID) {
			low = mid + 1
		} else if keyOID > uint64(targetOID) {
			high = mid - 1
		} else {
			if keyXID <= uint64(targetXID) {
				bestMatch = mid
				bestMatchEntry = entry
				low = mid + 1
			} else {
				high = mid - 1
			}
		}
	}

	if bestMatch >= 0 && bestMatchEntry != nil {
		bestMatchEntry.Index = bestMatch
		return bestMatchEntry, nil
	}

	return nil, fmt.Errorf("OID %d with XID <= %d not found in node", targetOID, targetXID)
}

// extractFixedSizeEntry extracts a kvoff_t entry
func (bs *BinarySearcher) extractFixedSizeEntry(node interfaces.BTreeNodeReader, nodeData []byte, index int, btnDataStart, tableOffset, entrySize int) (*SearchEntry, error) {
	offset := tableOffset + index*entrySize
	if offset+entrySize > len(nodeData) {
		return nil, fmt.Errorf("entry offset out of bounds")
	}

	keyOffset := bs.endian.Uint16(nodeData[offset : offset+2])
	valueOffset := bs.endian.Uint16(nodeData[offset+2 : offset+4])

	entry := &SearchEntry{
		IsChildNode: !node.IsLeaf(),
	}

	// Extract key data
	keyStart := btnDataStart + int(keyOffset)
	if keyStart+8 <= len(nodeData) {
		entry.KeyData = nodeData[keyStart : keyStart+8]
	}

	// Extract value data
	valueStart := btnDataStart + int(valueOffset)
	if valueStart+8 <= len(nodeData) {
		valueData := nodeData[valueStart : valueStart+8]
		entry.ValueData = valueData

		// If this is an internal node, interpret value as child OID
		if !node.IsLeaf() {
			entry.ChildOID = types.OidT(bs.endian.Uint64(valueData))
		}
	}

	return entry, nil
}

// extractVariableSizeEntry extracts a kvloc_t entry
func (bs *BinarySearcher) extractVariableSizeEntry(node interfaces.BTreeNodeReader, nodeData []byte, index int, btnDataStart, tableOffset, entrySize int) (*SearchEntry, error) {
	offset := tableOffset + index*entrySize
	if offset+entrySize > len(nodeData) {
		return nil, fmt.Errorf("entry offset out of bounds")
	}

	keyOffset := bs.endian.Uint16(nodeData[offset : offset+2])
	keySize := bs.endian.Uint16(nodeData[offset+2 : offset+4])
	valueOffset := bs.endian.Uint16(nodeData[offset+4 : offset+6])
	valueSize := bs.endian.Uint16(nodeData[offset+6 : offset+8])

	entry := &SearchEntry{
		IsChildNode: !node.IsLeaf(),
	}

	// Extract key data
	keyStart := btnDataStart + int(keyOffset)
	if keyStart+int(keySize) <= len(nodeData) {
		entry.KeyData = nodeData[keyStart : keyStart+int(keySize)]
	}

	// Extract value data
	valueStart := btnDataStart + int(valueOffset)
	if valueStart+int(valueSize) <= len(nodeData) {
		valueData := nodeData[valueStart : valueStart+int(valueSize)]
		entry.ValueData = valueData

		// If this is an internal node and value is 8 bytes, interpret as child OID
		if !node.IsLeaf() && len(valueData) == 8 {
			entry.ChildOID = types.OidT(bs.endian.Uint64(valueData))
		}
	}

	return entry, nil
}

// extractOIDFromKey extracts the OID from a key (first 8 bytes)
func (bs *BinarySearcher) extractOIDFromKey(keyData []byte) (uint64, error) {
	if len(keyData) < 8 {
		return 0, fmt.Errorf("key too short for OID extraction")
	}
	return bs.endian.Uint64(keyData[0:8]), nil
}

// extractOIDXIDFromKey extracts OID (first 8 bytes) and XID (second 8 bytes)
func (bs *BinarySearcher) extractOIDXIDFromKey(keyData []byte) (uint64, uint64, error) {
	if len(keyData) < 16 {
		return 0, 0, fmt.Errorf("key too short for OID+XID extraction")
	}
	oid := bs.endian.Uint64(keyData[0:8])
	xid := bs.endian.Uint64(keyData[8:16])
	return oid, xid, nil
}
