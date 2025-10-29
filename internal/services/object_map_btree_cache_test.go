package services

import (
	"bytes"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// MockBTreeNodeForCache is a mock B-tree node for testing cache
type MockBTreeNodeForCache struct {
	oid types.OidT
}

func (m *MockBTreeNodeForCache) Flags() uint16              { return 0 }
func (m *MockBTreeNodeForCache) Level() uint16              { return 0 }
func (m *MockBTreeNodeForCache) KeyCount() uint32           { return 0 }
func (m *MockBTreeNodeForCache) TableSpace() types.NlocT    { return types.NlocT{} }
func (m *MockBTreeNodeForCache) FreeSpace() types.NlocT     { return types.NlocT{} }
func (m *MockBTreeNodeForCache) KeyFreeList() types.NlocT   { return types.NlocT{} }
func (m *MockBTreeNodeForCache) ValueFreeList() types.NlocT { return types.NlocT{} }
func (m *MockBTreeNodeForCache) Data() []byte               { return []byte{} }
func (m *MockBTreeNodeForCache) IsRoot() bool               { return false }
func (m *MockBTreeNodeForCache) IsLeaf() bool               { return false }
func (m *MockBTreeNodeForCache) HasFixedKVSize() bool       { return false }
func (m *MockBTreeNodeForCache) IsHashed() bool             { return false }
func (m *MockBTreeNodeForCache) HasHeader() bool            { return false }

func TestNewObjectMapBTreeCache(t *testing.T) {
	config := CacheConfig{
		MaxNodes:     100,
		MaxBlockSize: 10 * 1024 * 1024,
	}

	cache := NewObjectMapBTreeCache(config)
	if cache == nil {
		t.Error("expected cache to be created")
	}

	if cache.maxNodes != 100 {
		t.Errorf("expected maxNodes=100, got %d", cache.maxNodes)
	}

	if cache.maxBlockSize != 10*1024*1024 {
		t.Errorf("expected maxBlockSize=10MB, got %d", cache.maxBlockSize)
	}
}

func TestDefaultCacheConfig(t *testing.T) {
	config := DefaultCacheConfig()

	if config.MaxNodes <= 0 {
		t.Error("expected positive MaxNodes")
	}

	if config.MaxBlockSize <= 0 {
		t.Error("expected positive MaxBlockSize")
	}
}

func TestNodeCacheGetPut(t *testing.T) {
	cache := NewObjectMapBTreeCache(CacheConfig{MaxNodes: 10, MaxBlockSize: 1024})

	node1 := &MockBTreeNodeForCache{oid: 1}
	err := cache.PutNode(types.OidT(1), node1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	retrieved, found := cache.GetNode(types.OidT(1))
	if !found {
		t.Error("expected node to be found in cache")
	}

	if retrieved != node1 {
		t.Error("expected same node object")
	}

	// Check statistics
	stats := cache.GetStats()
	if stats.NodeCachedCount != 1 {
		t.Errorf("expected 1 cached node, got %d", stats.NodeCachedCount)
	}
	if stats.NodeHits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.NodeHits)
	}
}

func TestNodeCacheMiss(t *testing.T) {
	cache := NewObjectMapBTreeCache(CacheConfig{MaxNodes: 10, MaxBlockSize: 1024})

	_, found := cache.GetNode(types.OidT(999))
	if found {
		t.Error("expected node not to be found")
	}

	stats := cache.GetStats()
	if stats.NodeMisses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.NodeMisses)
	}
}

func TestBlockCacheGetPut(t *testing.T) {
	cache := NewObjectMapBTreeCache(CacheConfig{MaxNodes: 10, MaxBlockSize: 10 * 1024})

	blockData := make([]byte, 1024)
	for i := range blockData {
		blockData[i] = byte(i % 256)
	}

	err := cache.PutBlock(uint64(1), blockData)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	retrieved, found := cache.GetBlock(uint64(1))
	if !found {
		t.Error("expected block to be found")
	}

	if !bytes.Equal(retrieved, blockData) {
		t.Error("expected same block data")
	}

	stats := cache.GetStats()
	if stats.BlockCachedCount != 1 {
		t.Errorf("expected 1 cached block, got %d", stats.BlockCachedCount)
	}
	if stats.BlockCachedSize != 1024 {
		t.Errorf("expected 1024 bytes cached, got %d", stats.BlockCachedSize)
	}
}

func TestNodeCacheEviction(t *testing.T) {
	cache := NewObjectMapBTreeCache(CacheConfig{MaxNodes: 3, MaxBlockSize: 1024})

	// Add 3 nodes
	for i := 1; i <= 3; i++ {
		node := &MockBTreeNodeForCache{oid: types.OidT(i)}
		err := cache.PutNode(types.OidT(i), node)
		if err != nil {
			t.Errorf("failed to put node %d: %v", i, err)
		}
	}

	stats := cache.GetStats()
	if stats.NodeCachedCount != 3 {
		t.Errorf("expected 3 cached nodes, got %d", stats.NodeCachedCount)
	}

	// Add 4th node - should evict the oldest (node 1)
	node4 := &MockBTreeNodeForCache{oid: 4}
	err := cache.PutNode(types.OidT(4), node4)
	if err != nil {
		t.Errorf("failed to put node 4: %v", err)
	}

	stats = cache.GetStats()
	if stats.NodeCachedCount != 3 {
		t.Errorf("expected 3 cached nodes after eviction, got %d", stats.NodeCachedCount)
	}
	if stats.NodeEvictions != 1 {
		t.Errorf("expected 1 eviction, got %d", stats.NodeEvictions)
	}

	// Node 1 should be evicted
	_, found := cache.GetNode(types.OidT(1))
	if found {
		t.Error("expected node 1 to be evicted")
	}

	// Node 4 should still be there
	_, found = cache.GetNode(types.OidT(4))
	if !found {
		t.Error("expected node 4 to be cached")
	}
}

func TestBlockCacheEviction(t *testing.T) {
	cache := NewObjectMapBTreeCache(CacheConfig{MaxNodes: 10, MaxBlockSize: 3 * 1024})

	blockData := make([]byte, 1024)

	// Add 3 blocks
	for i := 1; i <= 3; i++ {
		err := cache.PutBlock(uint64(i), blockData)
		if err != nil {
			t.Errorf("failed to put block %d: %v", i, err)
		}
	}

	stats := cache.GetStats()
	if stats.BlockCachedCount != 3 {
		t.Errorf("expected 3 cached blocks, got %d", stats.BlockCachedCount)
	}

	// Add 4th block - should evict the oldest
	err := cache.PutBlock(uint64(4), blockData)
	if err != nil {
		t.Errorf("failed to put block 4: %v", err)
	}

	stats = cache.GetStats()
	if stats.BlockCachedCount != 3 {
		t.Errorf("expected 3 cached blocks after eviction, got %d", stats.BlockCachedCount)
	}
	if stats.BlockEvictions != 1 {
		t.Errorf("expected 1 eviction, got %d", stats.BlockEvictions)
	}

	// Block 1 should be evicted
	_, found := cache.GetBlock(uint64(1))
	if found {
		t.Error("expected block 1 to be evicted")
	}
}

func TestCacheInvalidateNode(t *testing.T) {
	cache := NewObjectMapBTreeCache(CacheConfig{MaxNodes: 10, MaxBlockSize: 1024})

	node := &MockBTreeNodeForCache{oid: 1}
	cache.PutNode(types.OidT(1), node)

	// Verify it's cached
	_, found := cache.GetNode(types.OidT(1))
	if !found {
		t.Error("expected node to be cached")
	}

	// Invalidate it
	cache.InvalidateNode(types.OidT(1))

	// Verify it's gone
	_, found = cache.GetNode(types.OidT(1))
	if found {
		t.Error("expected node to be invalidated")
	}
}

func TestCacheInvalidateBlock(t *testing.T) {
	cache := NewObjectMapBTreeCache(CacheConfig{MaxNodes: 10, MaxBlockSize: 10 * 1024})

	blockData := make([]byte, 1024)
	cache.PutBlock(uint64(1), blockData)

	// Verify it's cached
	_, found := cache.GetBlock(uint64(1))
	if !found {
		t.Error("expected block to be cached")
	}

	// Invalidate it
	cache.InvalidateBlock(uint64(1))

	// Verify it's gone
	_, found = cache.GetBlock(uint64(1))
	if found {
		t.Error("expected block to be invalidated")
	}
}

func TestCacheClear(t *testing.T) {
	cache := NewObjectMapBTreeCache(CacheConfig{MaxNodes: 10, MaxBlockSize: 10 * 1024})

	// Add nodes and blocks
	node := &MockBTreeNodeForCache{oid: 1}
	cache.PutNode(types.OidT(1), node)

	blockData := make([]byte, 1024)
	cache.PutBlock(uint64(1), blockData)

	stats := cache.GetStats()
	if stats.NodeCachedCount == 0 || stats.BlockCachedCount == 0 {
		t.Error("expected data to be cached")
	}

	// Clear all
	cache.Clear()

	stats = cache.GetStats()
	if stats.NodeCachedCount != 0 {
		t.Errorf("expected 0 nodes after clear, got %d", stats.NodeCachedCount)
	}
	if stats.BlockCachedCount != 0 {
		t.Errorf("expected 0 blocks after clear, got %d", stats.BlockCachedCount)
	}
	if stats.BlockCachedSize != 0 {
		t.Errorf("expected 0 bytes cached after clear, got %d", stats.BlockCachedSize)
	}
}

func TestCacheStatistics(t *testing.T) {
	cache := NewObjectMapBTreeCache(CacheConfig{MaxNodes: 10, MaxBlockSize: 10 * 1024})

	// Add node
	node := &MockBTreeNodeForCache{oid: 1}
	cache.PutNode(types.OidT(1), node)

	// Hit
	cache.GetNode(types.OidT(1))

	// Miss
	cache.GetNode(types.OidT(999))

	stats := cache.GetStats()
	if stats.NodeHits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.NodeHits)
	}
	if stats.NodeMisses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.NodeMisses)
	}

	expectedHitRate := 0.5 // 1 hit out of 2 accesses
	if stats.NodeHitRate != expectedHitRate {
		t.Errorf("expected hit rate %.2f, got %.2f", expectedHitRate, stats.NodeHitRate)
	}
}

func TestCacheResetStats(t *testing.T) {
	cache := NewObjectMapBTreeCache(CacheConfig{MaxNodes: 10, MaxBlockSize: 10 * 1024})

	node := &MockBTreeNodeForCache{oid: 1}
	cache.PutNode(types.OidT(1), node)
	cache.GetNode(types.OidT(1))

	stats := cache.GetStats()
	if stats.NodeHits != 1 {
		t.Error("expected statistics to be recorded")
	}

	cache.ResetStats()

	stats = cache.GetStats()
	if stats.NodeHits != 0 {
		t.Errorf("expected 0 hits after reset, got %d", stats.NodeHits)
	}
	if stats.NodeMisses != 0 {
		t.Errorf("expected 0 misses after reset, got %d", stats.NodeMisses)
	}
}

func TestCacheThreadSafety(t *testing.T) {
	cache := NewObjectMapBTreeCache(CacheConfig{MaxNodes: 100, MaxBlockSize: 100 * 1024})

	// This test verifies that the cache doesn't panic under concurrent access
	// In a real test, you might use race detector: go test -race

	done := make(chan bool, 2)

	// Goroutine 1: Add nodes
	go func() {
		for i := 0; i < 10; i++ {
			node := &MockBTreeNodeForCache{oid: types.OidT(i)}
			cache.PutNode(types.OidT(i), node)
		}
		done <- true
	}()

	// Goroutine 2: Read nodes
	go func() {
		for i := 0; i < 10; i++ {
			cache.GetNode(types.OidT(i))
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	stats := cache.GetStats()
	if stats.NodeCachedCount == 0 {
		t.Error("expected some nodes to be cached")
	}
}

func TestCacheNilErrors(t *testing.T) {
	cache := NewObjectMapBTreeCache(CacheConfig{MaxNodes: 10, MaxBlockSize: 10 * 1024})

	// Try to put nil node
	err := cache.PutNode(types.OidT(1), nil)
	if err == nil {
		t.Error("expected error when putting nil node")
	}

	// Try to put nil block
	err = cache.PutBlock(uint64(1), nil)
	if err == nil {
		t.Error("expected error when putting nil block")
	}
}
