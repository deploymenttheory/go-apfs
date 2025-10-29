package services

import (
	"container/list"
	"fmt"
	"sync"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// ObjectMapBTreeCache provides two-level caching for B-tree nodes and block data
// Level 1: Parsed B-tree nodes (key: OID)
// Level 2: Raw block data (key: block number)
// Both use LRU eviction policy with thread-safe access
type ObjectMapBTreeCache struct {
	// Node cache: OID -> parsed B-tree node
	nodeCache map[types.OidT]*lruNode
	nodeOrder *list.List // LRU order

	// Block cache: block number -> raw block data
	blockCache map[uint64]*lruBlock
	blockOrder *list.List // LRU order

	// Configuration
	maxNodes         int
	maxBlockSize     int64 // in bytes
	currentNodeCount int
	currentBlockSize int64

	// Statistics
	nodeHits       int64
	nodeMisses     int64
	blockHits      int64
	blockMisses    int64
	nodeEvictions  int64
	blockEvictions int64

	// Thread safety
	mu sync.RWMutex
}

// lruNode represents a cached B-tree node with LRU metadata
type lruNode struct {
	oid     types.OidT
	node    interfaces.BTreeNodeReader
	element *list.Element
	touched int64 // for LRU tracking
}

// lruBlock represents a cached block with LRU metadata
type lruBlock struct {
	blockNum uint64
	data     []byte
	element  *list.Element
	touched  int64
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	MaxNodes     int   // Maximum number of nodes to cache
	MaxBlockSize int64 // Maximum block cache size in bytes
}

// DefaultCacheConfig returns recommended cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MaxNodes:     1000,              // ~32MB for nodes (assuming 32KB avg)
		MaxBlockSize: 100 * 1024 * 1024, // 100MB for blocks
	}
}

// NewObjectMapBTreeCache creates a new two-level cache with the given configuration
func NewObjectMapBTreeCache(config CacheConfig) *ObjectMapBTreeCache {
	if config.MaxNodes <= 0 {
		config.MaxNodes = DefaultCacheConfig().MaxNodes
	}
	if config.MaxBlockSize <= 0 {
		config.MaxBlockSize = DefaultCacheConfig().MaxBlockSize
	}

	return &ObjectMapBTreeCache{
		nodeCache:        make(map[types.OidT]*lruNode),
		nodeOrder:        list.New(),
		blockCache:       make(map[uint64]*lruBlock),
		blockOrder:       list.New(),
		maxNodes:         config.MaxNodes,
		maxBlockSize:     config.MaxBlockSize,
		currentNodeCount: 0,
		currentBlockSize: 0,
	}
}

// GetNode retrieves a cached B-tree node by OID
func (c *ObjectMapBTreeCache) GetNode(oid types.OidT) (interfaces.BTreeNodeReader, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if node, exists := c.nodeCache[oid]; exists {
		// Move to front (most recently used)
		c.nodeOrder.MoveToFront(node.element)
		c.nodeHits++
		return node.node, true
	}

	c.nodeMisses++
	return nil, false
}

// PutNode stores a B-tree node in the cache
func (c *ObjectMapBTreeCache) PutNode(oid types.OidT, node interfaces.BTreeNodeReader) error {
	if node == nil {
		return fmt.Errorf("cannot cache nil node")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already cached
	if existing, exists := c.nodeCache[oid]; exists {
		c.nodeOrder.MoveToFront(existing.element)
		return nil
	}

	// Add to cache
	element := c.nodeOrder.PushFront(&lruNode{
		oid:     oid,
		node:    node,
		touched: 0,
	})
	c.nodeCache[oid] = &lruNode{
		oid:     oid,
		node:    node,
		element: element,
	}
	c.currentNodeCount++

	// Evict if necessary
	for c.currentNodeCount > c.maxNodes && c.nodeOrder.Len() > 0 {
		c.evictOldestNode()
	}

	return nil
}

// GetBlock retrieves cached block data by block number
func (c *ObjectMapBTreeCache) GetBlock(blockNum uint64) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if block, exists := c.blockCache[blockNum]; exists {
		c.blockOrder.MoveToFront(block.element)
		c.blockHits++
		return block.data, true
	}

	c.blockMisses++
	return nil, false
}

// PutBlock stores block data in the cache
func (c *ObjectMapBTreeCache) PutBlock(blockNum uint64, data []byte) error {
	if data == nil {
		return fmt.Errorf("cannot cache nil block data")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	blockSize := int64(len(data))

	// Check if already cached
	if existing, exists := c.blockCache[blockNum]; exists {
		c.blockOrder.MoveToFront(existing.element)
		return nil
	}

	// Add to cache
	element := c.blockOrder.PushFront(&lruBlock{
		blockNum: blockNum,
		data:     data,
		touched:  0,
	})
	c.blockCache[blockNum] = &lruBlock{
		blockNum: blockNum,
		data:     data,
		element:  element,
	}
	c.currentBlockSize += blockSize

	// Evict if necessary
	for c.currentBlockSize > c.maxBlockSize && c.blockOrder.Len() > 0 {
		c.evictOldestBlock()
	}

	return nil
}

// evictOldestNode removes the least recently used node from cache
func (c *ObjectMapBTreeCache) evictOldestNode() {
	if c.nodeOrder.Len() == 0 {
		return
	}

	element := c.nodeOrder.Back()
	if element == nil {
		return
	}

	node := element.Value.(*lruNode)
	delete(c.nodeCache, node.oid)
	c.nodeOrder.Remove(element)
	c.currentNodeCount--
	c.nodeEvictions++
}

// evictOldestBlock removes the least recently used block from cache
func (c *ObjectMapBTreeCache) evictOldestBlock() {
	if c.blockOrder.Len() == 0 {
		return
	}

	element := c.blockOrder.Back()
	if element == nil {
		return
	}

	block := element.Value.(*lruBlock)
	c.currentBlockSize -= int64(len(block.data))
	delete(c.blockCache, block.blockNum)
	c.blockOrder.Remove(element)
	c.blockEvictions++
}

// InvalidateNode removes a specific node from cache
func (c *ObjectMapBTreeCache) InvalidateNode(oid types.OidT) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if node, exists := c.nodeCache[oid]; exists {
		delete(c.nodeCache, oid)
		c.nodeOrder.Remove(node.element)
		c.currentNodeCount--
	}
}

// InvalidateBlock removes a specific block from cache
func (c *ObjectMapBTreeCache) InvalidateBlock(blockNum uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if block, exists := c.blockCache[blockNum]; exists {
		c.currentBlockSize -= int64(len(block.data))
		delete(c.blockCache, blockNum)
		c.blockOrder.Remove(block.element)
	}
}

// ClearNodeCache removes all cached nodes
func (c *ObjectMapBTreeCache) ClearNodeCache() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.nodeCache = make(map[types.OidT]*lruNode)
	c.nodeOrder = list.New()
	c.currentNodeCount = 0
}

// ClearBlockCache removes all cached blocks
func (c *ObjectMapBTreeCache) ClearBlockCache() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.blockCache = make(map[uint64]*lruBlock)
	c.blockOrder = list.New()
	c.currentBlockSize = 0
}

// Clear removes all cached data
func (c *ObjectMapBTreeCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.nodeCache = make(map[types.OidT]*lruNode)
	c.nodeOrder = list.New()
	c.currentNodeCount = 0

	c.blockCache = make(map[uint64]*lruBlock)
	c.blockOrder = list.New()
	c.currentBlockSize = 0
}

// CacheStats returns current cache statistics
type CacheStats struct {
	NodeCachedCount int
	NodeHits        int64
	NodeMisses      int64
	NodeHitRate     float64
	NodeEvictions   int64

	BlockCachedCount int
	BlockCachedSize  int64
	BlockHits        int64
	BlockMisses      int64
	BlockHitRate     float64
	BlockEvictions   int64
}

// GetStats returns statistics about cache performance
func (c *ObjectMapBTreeCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := CacheStats{
		NodeCachedCount: len(c.nodeCache),
		NodeHits:        c.nodeHits,
		NodeMisses:      c.nodeMisses,
		NodeEvictions:   c.nodeEvictions,

		BlockCachedCount: len(c.blockCache),
		BlockCachedSize:  c.currentBlockSize,
		BlockHits:        c.blockHits,
		BlockMisses:      c.blockMisses,
		BlockEvictions:   c.blockEvictions,
	}

	// Calculate hit rates
	if totalNodeAccess := c.nodeHits + c.nodeMisses; totalNodeAccess > 0 {
		stats.NodeHitRate = float64(c.nodeHits) / float64(totalNodeAccess)
	}

	if totalBlockAccess := c.blockHits + c.blockMisses; totalBlockAccess > 0 {
		stats.BlockHitRate = float64(c.blockHits) / float64(totalBlockAccess)
	}

	return stats
}

// ResetStats clears all statistics counters
func (c *ObjectMapBTreeCache) ResetStats() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.nodeHits = 0
	c.nodeMisses = 0
	c.blockHits = 0
	c.blockMisses = 0
	c.nodeEvictions = 0
	c.blockEvictions = 0
}
