package services

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/deploymenttheory/go-apfs/internal/parsers/container"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// ContainerReader provides low-level access to container data
type ContainerReader struct {
	file             *os.File
	superblock       *types.NxSuperblockT
	blockSize        uint32
	containerSize    uint64
	endianness       binary.ByteOrder
	mu               sync.RWMutex
	blockCache       map[uint64][]byte
	maxCacheSize     int
	currentCacheSize int
}

// NewContainerReader opens a container file and reads its superblock
func NewContainerReader(filePath string) (*ContainerReader, error) {
	if filePath == "" {
		return nil, fmt.Errorf("container file path cannot be empty")
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open container file: %w", err)
	}

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	containerSize := uint64(fileInfo.Size())

	// Read the container superblock (at block 0)
	// APFS superblock is typically 4096 bytes
	superblockData := make([]byte, 4096)
	n, err := file.ReadAt(superblockData, 0)
	if err != nil && err != io.EOF {
		file.Close()
		return nil, fmt.Errorf("failed to read superblock: %w", err)
	}

	if n < 1024 {
		file.Close()
		return nil, fmt.Errorf("insufficient data read for superblock: got %d bytes, need at least 1024", n)
	}

	// Parse the superblock to get block size and validate
	sbReader, err := container.NewContainerSuperblockReader(superblockData, binary.LittleEndian)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to parse container superblock: %w", err)
	}

	blockSize := sbReader.BlockSize()
	if blockSize == 0 {
		file.Close()
		return nil, fmt.Errorf("invalid block size: 0")
	}

	cr := &ContainerReader{
		file:          file,
		superblock:    sbReader.(*container.ContainerSuperblockReader).Superblock,
		blockSize:     blockSize,
		containerSize: containerSize,
		endianness:    binary.LittleEndian,
		blockCache:    make(map[uint64][]byte),
		maxCacheSize:  50 * 1024 * 1024, // 50MB cache
	}

	return cr, nil
}

// ReadBlock reads a single block from the container
func (cr *ContainerReader) ReadBlock(blockNumber uint64) ([]byte, error) {
	cr.mu.RLock()

	// Check cache first
	if cachedBlock, exists := cr.blockCache[blockNumber]; exists {
		cr.mu.RUnlock()
		return append([]byte{}, cachedBlock...), nil // Return copy
	}
	cr.mu.RUnlock()

	// Calculate offset
	offset := int64(blockNumber) * int64(cr.blockSize)
	if uint64(offset) >= cr.containerSize {
		return nil, fmt.Errorf("block %d is beyond container size", blockNumber)
	}

	// Read the block
	blockData := make([]byte, cr.blockSize)
	n, err := cr.file.ReadAt(blockData, offset)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read block %d: %w", blockNumber, err)
	}

	if n < int(cr.blockSize) {
		return nil, fmt.Errorf("incomplete block read: got %d bytes, expected %d", n, cr.blockSize)
	}

	// Cache the block
	cr.mu.Lock()
	cr.cacheBlock(blockNumber, blockData)
	cr.mu.Unlock()

	return append([]byte{}, blockData...), nil // Return copy
}

// ReadBlocks reads multiple consecutive blocks
func (cr *ContainerReader) ReadBlocks(startBlock uint64, count uint64) ([]byte, error) {
	if count == 0 {
		return []byte{}, nil
	}

	totalSize := count * uint64(cr.blockSize)
	result := make([]byte, 0, totalSize)

	for i := uint64(0); i < count; i++ {
		blockNum := startBlock + i
		blockData, err := cr.ReadBlock(blockNum)
		if err != nil {
			return nil, fmt.Errorf("failed to read block %d: %w", blockNum, err)
		}
		result = append(result, blockData...)
	}

	return result, nil
}

// cacheBlock adds a block to the cache, respecting size limits
// Must be called with mu locked
func (cr *ContainerReader) cacheBlock(blockNumber uint64, data []byte) {
	blockSize := len(data)

	// If adding this block exceeds cache size, clear cache
	if cr.currentCacheSize+blockSize > cr.maxCacheSize {
		cr.blockCache = make(map[uint64][]byte)
		cr.currentCacheSize = 0
	}

	cr.blockCache[blockNumber] = append([]byte{}, data...)
	cr.currentCacheSize += blockSize
}

// GetBlockSize returns the block size of the container
func (cr *ContainerReader) GetBlockSize() uint32 {
	return cr.blockSize
}

// GetContainerSize returns the total container size
func (cr *ContainerReader) GetContainerSize() uint64 {
	return cr.containerSize
}

// GetSuperblock returns the container superblock
func (cr *ContainerReader) GetSuperblock() *types.NxSuperblockT {
	return cr.superblock
}

// SeekToBlock positions for reading at a specific block
func (cr *ContainerReader) SeekToBlock(blockNumber uint64) error {
	offset := int64(blockNumber) * int64(cr.blockSize)
	if uint64(offset) >= cr.containerSize {
		return fmt.Errorf("block %d is beyond container size", blockNumber)
	}

	_, err := cr.file.Seek(offset, io.SeekStart)
	if err != nil {
		return fmt.Errorf("seek failed: %w", err)
	}

	return nil
}

// ClearCache removes all cached blocks
func (cr *ContainerReader) ClearCache() {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.blockCache = make(map[uint64][]byte)
	cr.currentCacheSize = 0
}

// GetCacheStats returns cache statistics
func (cr *ContainerReader) GetCacheStats() map[string]interface{} {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	return map[string]interface{}{
		"cached_blocks":    len(cr.blockCache),
		"cache_size_bytes": cr.currentCacheSize,
		"max_cache_bytes":  cr.maxCacheSize,
	}
}

// IsCached checks if a block is in cache
func (cr *ContainerReader) IsCached(blockNumber uint64) bool {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	_, exists := cr.blockCache[blockNumber]
	return exists
}

// Close closes the container file
func (cr *ContainerReader) Close() error {
	if cr.file != nil {
		return cr.file.Close()
	}
	return nil
}

// GetEndianness returns the endianness used in the container
func (cr *ContainerReader) GetEndianness() binary.ByteOrder {
	return cr.endianness
}
