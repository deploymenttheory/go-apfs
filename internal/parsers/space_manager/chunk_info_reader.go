package spacemanager

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// ChunkInfoReader provides parsing capabilities for chunk_info_t structures
// Chunk info manages information about a chunk of storage in the space manager
type ChunkInfoReader struct {
	chunkInfo *types.ChunkInfoT
	data      []byte
	endian    binary.ByteOrder
}

// NewChunkInfoReader creates a new chunk info reader
// Chunk info structures track allocation state for groups of blocks
func NewChunkInfoReader(data []byte, endian binary.ByteOrder) (*ChunkInfoReader, error) {
	// chunk_info_t structure size: 8 + 8 + 4 + 4 + 8 = 32 bytes
	if len(data) < 32 {
		return nil, fmt.Errorf("data too small for chunk info: %d bytes, need at least 32", len(data))
	}

	chunkInfo, err := parseChunkInfo(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chunk info: %w", err)
	}

	return &ChunkInfoReader{
		chunkInfo: chunkInfo,
		data:      data,
		endian:    endian,
	}, nil
}

// parseChunkInfo parses raw bytes into a ChunkInfoT structure
// This follows the exact layout of chunk_info_t from Apple File System Reference
func parseChunkInfo(data []byte, endian binary.ByteOrder) (*types.ChunkInfoT, error) {
	if len(data) < 32 {
		return nil, fmt.Errorf("insufficient data for chunk info")
	}

	ci := &types.ChunkInfoT{}
	offset := 0

	// Parse chunk info fields (32 bytes total)
	// uint64_t ci_xid - Transaction identifier for this chunk information
	ci.CiXid = endian.Uint64(data[offset : offset+8])
	offset += 8

	// uint64_t ci_addr - Address of the chunk
	ci.CiAddr = endian.Uint64(data[offset : offset+8])
	offset += 8

	// uint32_t ci_block_count - Number of blocks in the chunk
	ci.CiBlockCount = endian.Uint32(data[offset : offset+4])
	offset += 4

	// uint32_t ci_free_count - Number of free blocks in the chunk
	ci.CiFreeCount = endian.Uint32(data[offset : offset+4])
	offset += 4

	// paddr_t ci_bitmap_addr - Address of the bitmap for this chunk
	ci.CiBitmapAddr = types.Paddr(endian.Uint64(data[offset : offset+8]))
	offset += 8

	return ci, nil
}

// GetChunkInfo returns the chunk info structure
// Provides direct access to the underlying chunk_info_t for advanced operations
func (cir *ChunkInfoReader) GetChunkInfo() *types.ChunkInfoT {
	return cir.chunkInfo
}

// TransactionID returns the transaction identifier for this chunk information
// Tracks when this chunk info was last modified
func (cir *ChunkInfoReader) TransactionID() uint64 {
	return cir.chunkInfo.CiXid
}

// ChunkAddress returns the address of the chunk
// Physical location of the chunk on the storage device
func (cir *ChunkInfoReader) ChunkAddress() uint64 {
	return cir.chunkInfo.CiAddr
}

// BlockCount returns the total number of blocks in the chunk
// Defines the size of this chunk in blocks
func (cir *ChunkInfoReader) BlockCount() uint32 {
	return cir.chunkInfo.CiBlockCount
}

// FreeCount returns the number of free blocks in the chunk
// Indicates available space within this chunk
func (cir *ChunkInfoReader) FreeCount() uint32 {
	return cir.chunkInfo.CiFreeCount
}

// UsedCount returns the number of used blocks in the chunk
// Calculated as total blocks minus free blocks
func (cir *ChunkInfoReader) UsedCount() uint32 {
	if cir.chunkInfo.CiBlockCount >= cir.chunkInfo.CiFreeCount {
		return cir.chunkInfo.CiBlockCount - cir.chunkInfo.CiFreeCount
	}
	return 0
}

// BitmapAddress returns the address of the bitmap for this chunk
// Points to the allocation bitmap that tracks which blocks are free/used
func (cir *ChunkInfoReader) BitmapAddress() types.Paddr {
	return cir.chunkInfo.CiBitmapAddr
}

// IsFull returns true if the chunk has no free blocks
// Indicates the chunk is completely allocated
func (cir *ChunkInfoReader) IsFull() bool {
	return cir.chunkInfo.CiFreeCount == 0
}

// IsEmpty returns true if the chunk has no used blocks
// Indicates the chunk is completely free
func (cir *ChunkInfoReader) IsEmpty() bool {
	return cir.chunkInfo.CiFreeCount == cir.chunkInfo.CiBlockCount
}

// UtilizationPercentage returns the percentage of blocks used in the chunk
// Returns a value between 0.0 (empty) and 100.0 (full)
func (cir *ChunkInfoReader) UtilizationPercentage() float64 {
	if cir.chunkInfo.CiBlockCount == 0 {
		return 0.0
	}
	used := cir.UsedCount()
	return (float64(used) / float64(cir.chunkInfo.CiBlockCount)) * 100.0
}