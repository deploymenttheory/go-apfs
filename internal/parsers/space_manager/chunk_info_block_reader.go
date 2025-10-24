package spacemanager

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// ChunkInfoBlockReader provides parsing capabilities for chunk_info_block_t structures
// A chunk info block contains an array of chunk-info structures for efficient management
type ChunkInfoBlockReader struct {
	chunkInfoBlock *types.ChunkInfoBlockT
	data           []byte
	endian         binary.ByteOrder
}

// NewChunkInfoBlockReader creates a new chunk info block reader
// Chunk info blocks organize multiple chunk info structures into manageable blocks
func NewChunkInfoBlockReader(data []byte, endian binary.ByteOrder) (*ChunkInfoBlockReader, error) {
	// Minimum size: obj_phys_t (32) + index (4) + count (4) = 40 bytes + variable chunk info array
	if len(data) < 40 {
		return nil, fmt.Errorf("data too small for chunk info block: %d bytes, need at least 40", len(data))
	}

	chunkInfoBlock, err := parseChunkInfoBlock(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chunk info block: %w", err)
	}

	// Validate the object type
	objectType := chunkInfoBlock.CibO.OType & types.ObjectTypeMask
	if objectType != types.ObjectTypeSpacemanCib {
		return nil, fmt.Errorf("invalid chunk info block object type: 0x%x", objectType)
	}

	return &ChunkInfoBlockReader{
		chunkInfoBlock: chunkInfoBlock,
		data:           data,
		endian:         endian,
	}, nil
}

// parseChunkInfoBlock parses raw bytes into a ChunkInfoBlockT structure
// This follows the exact layout of chunk_info_block_t from Apple File System Reference
func parseChunkInfoBlock(data []byte, endian binary.ByteOrder) (*types.ChunkInfoBlockT, error) {
	if len(data) < 40 {
		return nil, fmt.Errorf("insufficient data for chunk info block")
	}

	cib := &types.ChunkInfoBlockT{}
	offset := 0

	// Parse object header (obj_phys_t): 32 bytes
	// Contains checksum, object ID, transaction ID, type, and subtype
	copy(cib.CibO.OChecksum[:], data[offset:offset+8])
	offset += 8
	cib.CibO.OOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	cib.CibO.OXid = types.XidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	cib.CibO.OType = endian.Uint32(data[offset : offset+4])
	offset += 4
	cib.CibO.OSubtype = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse chunk info block specific fields
	// uint32_t cib_index - Index of this chunk info block
	cib.CibIndex = endian.Uint32(data[offset : offset+4])
	offset += 4

	// uint32_t cib_chunk_info_count - Number of chunk info entries in this block
	cib.CibChunkInfoCount = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse chunk_info_t array: cib_chunk_info[]
	// Each chunk_info_t is 32 bytes
	if cib.CibChunkInfoCount > 0 {
		chunkInfoSize := 32 // Size of chunk_info_t
		requiredSize := offset + int(cib.CibChunkInfoCount)*chunkInfoSize
		if len(data) < requiredSize {
			return nil, fmt.Errorf("insufficient data for chunk info entries: need %d bytes, have %d",
				requiredSize, len(data))
		}

		cib.CibChunkInfo = make([]types.ChunkInfoT, cib.CibChunkInfoCount)
		for i := uint32(0); i < cib.CibChunkInfoCount; i++ {
			chunkInfo := &cib.CibChunkInfo[i]

			// Parse each chunk_info_t (32 bytes)
			chunkInfo.CiXid = endian.Uint64(data[offset : offset+8])
			offset += 8
			chunkInfo.CiAddr = endian.Uint64(data[offset : offset+8])
			offset += 8
			chunkInfo.CiBlockCount = endian.Uint32(data[offset : offset+4])
			offset += 4
			chunkInfo.CiFreeCount = endian.Uint32(data[offset : offset+4])
			offset += 4
			chunkInfo.CiBitmapAddr = types.Paddr(endian.Uint64(data[offset : offset+8]))
			offset += 8
		}
	}

	return cib, nil
}

// GetChunkInfoBlock returns the chunk info block structure
// Provides direct access to the underlying chunk_info_block_t for advanced operations
func (cibr *ChunkInfoBlockReader) GetChunkInfoBlock() *types.ChunkInfoBlockT {
	return cibr.chunkInfoBlock
}

// ObjectHeader returns the object header
// Provides access to object ID, transaction ID, and other metadata
func (cibr *ChunkInfoBlockReader) ObjectHeader() *types.ObjPhysT {
	return &cibr.chunkInfoBlock.CibO
}

// Index returns the index of this chunk info block
// Used to identify the position of this block in the overall chunk info hierarchy
func (cibr *ChunkInfoBlockReader) Index() uint32 {
	return cibr.chunkInfoBlock.CibIndex
}

// ChunkInfoCount returns the number of chunk info entries in this block
// Indicates how many chunks are managed by this block
func (cibr *ChunkInfoBlockReader) ChunkInfoCount() uint32 {
	return cibr.chunkInfoBlock.CibChunkInfoCount
}

// GetChunkInfo returns a specific chunk info by index
// Provides access to individual chunk information within this block
func (cibr *ChunkInfoBlockReader) GetChunkInfo(index uint32) (*types.ChunkInfoT, error) {
	if index >= cibr.chunkInfoBlock.CibChunkInfoCount {
		return nil, fmt.Errorf("chunk info index %d out of range (have %d entries)",
			index, cibr.chunkInfoBlock.CibChunkInfoCount)
	}
	return &cibr.chunkInfoBlock.CibChunkInfo[index], nil
}

// GetAllChunkInfos returns all chunk info entries in this block
// Provides bulk access to all chunk information
func (cibr *ChunkInfoBlockReader) GetAllChunkInfos() []types.ChunkInfoT {
	return cibr.chunkInfoBlock.CibChunkInfo
}

// GetChunkInfoReader returns a ChunkInfoReader for a specific chunk
// Provides enhanced access to chunk information with utility methods
func (cibr *ChunkInfoBlockReader) GetChunkInfoReader(index uint32) (*ChunkInfoReader, error) {
	if index >= cibr.chunkInfoBlock.CibChunkInfoCount {
		return nil, fmt.Errorf("chunk info index %d out of range (have %d entries)",
			index, cibr.chunkInfoBlock.CibChunkInfoCount)
	}

	// Create a 32-byte buffer for the chunk info
	chunkData := make([]byte, 32)
	chunkInfo := &cibr.chunkInfoBlock.CibChunkInfo[index]

	// Serialize the chunk info back to bytes for the reader
	offset := 0
	cibr.endian.PutUint64(chunkData[offset:offset+8], chunkInfo.CiXid)
	offset += 8
	cibr.endian.PutUint64(chunkData[offset:offset+8], chunkInfo.CiAddr)
	offset += 8
	cibr.endian.PutUint32(chunkData[offset:offset+4], chunkInfo.CiBlockCount)
	offset += 4
	cibr.endian.PutUint32(chunkData[offset:offset+4], chunkInfo.CiFreeCount)
	offset += 4
	cibr.endian.PutUint64(chunkData[offset:offset+8], uint64(chunkInfo.CiBitmapAddr))
	offset += 8

	return NewChunkInfoReader(chunkData, cibr.endian)
}

// CalculateTotalBlocks returns the sum of all block counts in this chunk info block
// Provides aggregate statistics for space management
func (cibr *ChunkInfoBlockReader) CalculateTotalBlocks() uint64 {
	var total uint64
	for i := uint32(0); i < cibr.chunkInfoBlock.CibChunkInfoCount; i++ {
		total += uint64(cibr.chunkInfoBlock.CibChunkInfo[i].CiBlockCount)
	}
	return total
}

// CalculateTotalFreeBlocks returns the sum of all free block counts in this chunk info block
// Provides aggregate free space statistics
func (cibr *ChunkInfoBlockReader) CalculateTotalFreeBlocks() uint64 {
	var total uint64
	for i := uint32(0); i < cibr.chunkInfoBlock.CibChunkInfoCount; i++ {
		total += uint64(cibr.chunkInfoBlock.CibChunkInfo[i].CiFreeCount)
	}
	return total
}

// CalculateOverallUtilization returns the utilization percentage across all chunks in this block
// Returns a value between 0.0 (all free) and 100.0 (all used)
func (cibr *ChunkInfoBlockReader) CalculateOverallUtilization() float64 {
	totalBlocks := cibr.CalculateTotalBlocks()
	if totalBlocks == 0 {
		return 0.0
	}
	totalFree := cibr.CalculateTotalFreeBlocks()
	totalUsed := totalBlocks - totalFree
	return (float64(totalUsed) / float64(totalBlocks)) * 100.0
}