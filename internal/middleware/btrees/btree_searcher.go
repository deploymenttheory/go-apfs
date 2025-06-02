package btrees

import (
	"bytes"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
)

// btreeSearcher implements the BTreeSearcher interface
type btreeSearcher struct {
	navigator   interfaces.BTreeNavigator
	btreeInfo   interfaces.BTreeInfoReader
	keyComparer KeyComparer
}

// KeyComparer defines a function type for comparing keys
type KeyComparer func(a, b []byte) int

// NewBTreeSearcher creates a new BTreeSearcher implementation
func NewBTreeSearcher(navigator interfaces.BTreeNavigator, btreeInfo interfaces.BTreeInfoReader, keyComparer KeyComparer) interfaces.BTreeSearcher {
	if keyComparer == nil {
		keyComparer = DefaultKeyComparer
	}

	return &btreeSearcher{
		navigator:   navigator,
		btreeInfo:   btreeInfo,
		keyComparer: keyComparer,
	}
}

// DefaultKeyComparer provides default byte-wise key comparison
func DefaultKeyComparer(a, b []byte) int {
	return bytes.Compare(a, b)
}

// Find looks for a key in the B-tree and returns its associated value
func (searcher *btreeSearcher) Find(key []byte) ([]byte, error) {
	rootNode, err := searcher.navigator.GetRootNode()
	if err != nil {
		return nil, fmt.Errorf("failed to get root node: %w", err)
	}

	return searcher.findInNode(rootNode, key)
}

// FindRange returns all key-value pairs within a given key range
func (searcher *btreeSearcher) FindRange(startKey []byte, endKey []byte) ([]interfaces.KeyValuePair, error) {
	if searcher.keyComparer(startKey, endKey) > 0 {
		return nil, fmt.Errorf("start key must be less than or equal to end key")
	}

	var results []interfaces.KeyValuePair

	err := searcher.traverseRange(startKey, endKey, func(key, value []byte) error {
		results = append(results, interfaces.KeyValuePair{
			Key:   append([]byte(nil), key...),   // Make a copy
			Value: append([]byte(nil), value...), // Make a copy
		})
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to traverse range: %w", err)
	}

	return results, nil
}

// ContainsKey checks if a key exists in the B-tree
func (searcher *btreeSearcher) ContainsKey(key []byte) (bool, error) {
	_, err := searcher.Find(key)
	if err != nil {
		return false, nil // Key not found
	}
	return true, nil
}

// findInNode searches for a key within a specific node
func (searcher *btreeSearcher) findInNode(node interfaces.BTreeNodeReader, key []byte) ([]byte, error) {
	if node.IsLeaf() {
		return searcher.findInLeaf(node, key)
	}

	return searcher.findInInternal(node, key)
}

// findInLeaf searches for a key in a leaf node
func (searcher *btreeSearcher) findInLeaf(node interfaces.BTreeNodeReader, key []byte) ([]byte, error) {
	entries, err := searcher.extractNodeEntries(node)
	if err != nil {
		return nil, fmt.Errorf("failed to extract node entries: %w", err)
	}

	for _, entry := range entries {
		if searcher.keyComparer(entry.Key, key) == 0 {
			return entry.Value, nil
		}
	}

	return nil, fmt.Errorf("key not found")
}

// findInInternal searches for a key in an internal node
func (searcher *btreeSearcher) findInInternal(node interfaces.BTreeNodeReader, key []byte) ([]byte, error) {
	childIndex, err := searcher.findChildIndex(node, key)
	if err != nil {
		return nil, fmt.Errorf("failed to find child index: %w", err)
	}

	childNode, err := searcher.navigator.GetChildNode(node, childIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get child node: %w", err)
	}

	return searcher.findInNode(childNode, key)
}

// findChildIndex determines which child to follow for a given key
func (searcher *btreeSearcher) findChildIndex(node interfaces.BTreeNodeReader, key []byte) (int, error) {
	entries, err := searcher.extractNodeEntries(node)
	if err != nil {
		return 0, fmt.Errorf("failed to extract node entries: %w", err)
	}

	// For internal nodes, find the appropriate child
	for i, entry := range entries {
		if searcher.keyComparer(key, entry.Key) < 0 {
			return i, nil
		}
	}

	// If key is greater than all keys, go to the rightmost child
	return len(entries), nil
}

// extractNodeEntries extracts all key-value pairs from a node
func (searcher *btreeSearcher) extractNodeEntries(node interfaces.BTreeNodeReader) ([]interfaces.KeyValuePair, error) {
	if node.HasFixedKVSize() {
		return searcher.extractFixedSizeEntries(node)
	}

	return searcher.extractVariableSizeEntries(node)
}

// extractFixedSizeEntries extracts entries from a fixed-size key/value node
func (searcher *btreeSearcher) extractFixedSizeEntries(node interfaces.BTreeNodeReader) ([]interfaces.KeyValuePair, error) {
	keySize := searcher.btreeInfo.KeySize()
	valueSize := searcher.btreeInfo.ValueSize()
	entrySize := keySize + valueSize
	keyCount := node.KeyCount()
	nodeData := node.Data()

	if entrySize == 0 {
		return nil, fmt.Errorf("invalid entry size: key=%d, value=%d", keySize, valueSize)
	}

	entries := make([]interfaces.KeyValuePair, 0, keyCount)

	for i := uint32(0); i < keyCount; i++ {
		offset := int(entrySize) * int(i)
		if offset+int(entrySize) > len(nodeData) {
			return nil, fmt.Errorf("entry %d extends beyond node data", i)
		}

		key := make([]byte, keySize)
		copy(key, nodeData[offset:offset+int(keySize)])

		value := make([]byte, valueSize)
		copy(value, nodeData[offset+int(keySize):offset+int(entrySize)])

		entries = append(entries, interfaces.KeyValuePair{
			Key:   key,
			Value: value,
		})
	}

	return entries, nil
}

// extractVariableSizeEntries extracts entries from a variable-size key/value node
func (searcher *btreeSearcher) extractVariableSizeEntries(node interfaces.BTreeNodeReader) ([]interfaces.KeyValuePair, error) {
	// This is a placeholder implementation for variable-size entries
	// Real APFS implementation would need to:
	// 1. Parse the table of contents to get key/value locations
	// 2. Extract each key/value pair based on the locations
	// 3. Handle the complex layout of variable-size entries

	return nil, fmt.Errorf("variable-size entry extraction not yet implemented")
}

// traverseRange traverses all key-value pairs within a range
func (searcher *btreeSearcher) traverseRange(startKey, endKey []byte, visitor func(key, value []byte) error) error {
	rootNode, err := searcher.navigator.GetRootNode()
	if err != nil {
		return fmt.Errorf("failed to get root node: %w", err)
	}

	return searcher.traverseRangeInNode(rootNode, startKey, endKey, visitor)
}

// traverseRangeInNode recursively traverses nodes within a key range
func (searcher *btreeSearcher) traverseRangeInNode(node interfaces.BTreeNodeReader, startKey, endKey []byte, visitor func(key, value []byte) error) error {
	if node.IsLeaf() {
		return searcher.visitLeafRange(node, startKey, endKey, visitor)
	}

	return searcher.visitInternalRange(node, startKey, endKey, visitor)
}

// visitLeafRange visits all entries in a leaf node within the key range
func (searcher *btreeSearcher) visitLeafRange(node interfaces.BTreeNodeReader, startKey, endKey []byte, visitor func(key, value []byte) error) error {
	entries, err := searcher.extractNodeEntries(node)
	if err != nil {
		return fmt.Errorf("failed to extract node entries: %w", err)
	}

	for _, entry := range entries {
		if searcher.keyComparer(entry.Key, startKey) >= 0 && searcher.keyComparer(entry.Key, endKey) <= 0 {
			if err := visitor(entry.Key, entry.Value); err != nil {
				return err
			}
		}
	}

	return nil
}

// visitInternalRange visits children of an internal node within the key range
func (searcher *btreeSearcher) visitInternalRange(node interfaces.BTreeNodeReader, startKey, endKey []byte, visitor func(key, value []byte) error) error {
	entries, err := searcher.extractNodeEntries(node)
	if err != nil {
		return fmt.Errorf("failed to extract node entries: %w", err)
	}

	// Determine which children to visit
	for i := 0; i <= len(entries); i++ {
		shouldVisit := false

		if i == 0 {
			// First child: visit if startKey is less than first key
			if len(entries) == 0 || searcher.keyComparer(startKey, entries[0].Key) < 0 {
				shouldVisit = true
			}
		} else if i == len(entries) {
			// Last child: visit if endKey is greater than last key
			if searcher.keyComparer(endKey, entries[i-1].Key) > 0 {
				shouldVisit = true
			}
		} else {
			// Middle child: visit if range overlaps with this section
			if searcher.keyComparer(startKey, entries[i].Key) < 0 && searcher.keyComparer(endKey, entries[i-1].Key) > 0 {
				shouldVisit = true
			}
		}

		if shouldVisit {
			childNode, err := searcher.navigator.GetChildNode(node, i)
			if err != nil {
				return fmt.Errorf("failed to get child node %d: %w", i, err)
			}

			if err := searcher.traverseRangeInNode(childNode, startKey, endKey, visitor); err != nil {
				return err
			}
		}
	}

	return nil
}
