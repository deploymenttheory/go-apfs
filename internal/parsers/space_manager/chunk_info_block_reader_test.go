package spacemanager

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestNewChunkInfoBlockReader(t *testing.T) {
	// Create chunk info block data with 2 chunk info entries
	entryCount := uint32(2)
	entrySize := 32 // Size of chunk_info_t
	dataSize := 40 + int(entryCount)*entrySize // Header + entries
	data := make([]byte, dataSize)
	endian := binary.LittleEndian

	// Fill object header
	endian.PutUint64(data[8:16], 123)                                    // OOid
	endian.PutUint64(data[16:24], 456)                                   // OXid
	endian.PutUint32(data[24:28], types.ObjectTypeSpacemanCib)           // OType
	endian.PutUint32(data[28:32], 0)                                     // OSubtype

	// Fill chunk info block specific fields
	endian.PutUint32(data[32:36], 5)          // CibIndex
	endian.PutUint32(data[36:40], entryCount) // CibChunkInfoCount

	// Fill first chunk info entry at offset 40
	offset := 40
	endian.PutUint64(data[offset:offset+8], 1001)   // CiXid
	endian.PutUint64(data[offset+8:offset+16], 2001) // CiAddr
	endian.PutUint32(data[offset+16:offset+20], 1024) // CiBlockCount
	endian.PutUint32(data[offset+20:offset+24], 256)  // CiFreeCount
	endian.PutUint64(data[offset+24:offset+32], 3001) // CiBitmapAddr

	// Fill second chunk info entry
	offset += entrySize
	endian.PutUint64(data[offset:offset+8], 1002)   // CiXid
	endian.PutUint64(data[offset+8:offset+16], 2002) // CiAddr
	endian.PutUint32(data[offset+16:offset+20], 2048) // CiBlockCount
	endian.PutUint32(data[offset+20:offset+24], 512)  // CiFreeCount
	endian.PutUint64(data[offset+24:offset+32], 3002) // CiBitmapAddr

	reader, err := NewChunkInfoBlockReader(data, endian)
	if err != nil {
		t.Fatalf("NewChunkInfoBlockReader() failed: %v", err)
	}

	if reader.Index() != 5 {
		t.Errorf("Index() = %d, want 5", reader.Index())
	}

	if reader.ChunkInfoCount() != entryCount {
		t.Errorf("ChunkInfoCount() = %d, want %d", reader.ChunkInfoCount(), entryCount)
	}

	// Test first chunk info
	chunkInfo0, err := reader.GetChunkInfo(0)
	if err != nil {
		t.Fatalf("GetChunkInfo(0) failed: %v", err)
	}

	if chunkInfo0.CiXid != 1001 {
		t.Errorf("ChunkInfo[0].CiXid = %d, want 1001", chunkInfo0.CiXid)
	}

	if chunkInfo0.CiBlockCount != 1024 {
		t.Errorf("ChunkInfo[0].CiBlockCount = %d, want 1024", chunkInfo0.CiBlockCount)
	}

	// Test second chunk info
	chunkInfo1, err := reader.GetChunkInfo(1)
	if err != nil {
		t.Fatalf("GetChunkInfo(1) failed: %v", err)
	}

	if chunkInfo1.CiAddr != 2002 {
		t.Errorf("ChunkInfo[1].CiAddr = %d, want 2002", chunkInfo1.CiAddr)
	}
}

func TestChunkInfoBlockReader_InvalidType(t *testing.T) {
	data := make([]byte, 64)
	endian := binary.LittleEndian

	// Set invalid object type
	endian.PutUint32(data[24:28], types.ObjectTypeNxSuperblock) // Wrong type

	_, err := NewChunkInfoBlockReader(data, endian)
	if err == nil {
		t.Error("NewChunkInfoBlockReader() should have failed with invalid object type")
	}
}

func TestChunkInfoBlockReader_TooSmall(t *testing.T) {
	data := make([]byte, 30) // Too small (need at least 40)

	_, err := NewChunkInfoBlockReader(data, binary.LittleEndian)
	if err == nil {
		t.Error("NewChunkInfoBlockReader() should have failed with too small data")
	}
}

func TestChunkInfoBlockReader_InsufficientDataForEntries(t *testing.T) {
	data := make([]byte, 50) // Header + partial entry
	endian := binary.LittleEndian

	// Set object header
	endian.PutUint32(data[24:28], types.ObjectTypeSpacemanCib)

	// Set count that requires more data than available
	endian.PutUint32(data[36:40], 5) // CibChunkInfoCount = 5, but not enough data

	_, err := NewChunkInfoBlockReader(data, endian)
	if err == nil {
		t.Error("NewChunkInfoBlockReader() should have failed with insufficient data for entries")
	}
}

func TestChunkInfoBlockReader_GetChunkInfoOutOfRange(t *testing.T) {
	data := make([]byte, 72) // Header + 1 entry
	endian := binary.LittleEndian

	// Set object header
	endian.PutUint32(data[24:28], types.ObjectTypeSpacemanCib)
	endian.PutUint32(data[36:40], 1) // CibChunkInfoCount = 1

	reader, err := NewChunkInfoBlockReader(data, endian)
	if err != nil {
		t.Fatalf("NewChunkInfoBlockReader() failed: %v", err)
	}

	// Try to access index 5 when only 1 entry exists
	_, err = reader.GetChunkInfo(5)
	if err == nil {
		t.Error("GetChunkInfo(5) should have failed with out of range index")
	}
}

func TestChunkInfoBlockReader_GetAllChunkInfos(t *testing.T) {
	entryCount := uint32(3)
	dataSize := 40 + int(entryCount)*32
	data := make([]byte, dataSize)
	endian := binary.LittleEndian

	// Set object header
	endian.PutUint32(data[24:28], types.ObjectTypeSpacemanCib)
	endian.PutUint32(data[36:40], entryCount) // CibChunkInfoCount

	reader, err := NewChunkInfoBlockReader(data, endian)
	if err != nil {
		t.Fatalf("NewChunkInfoBlockReader() failed: %v", err)
	}

	allChunkInfos := reader.GetAllChunkInfos()
	if len(allChunkInfos) != int(entryCount) {
		t.Errorf("GetAllChunkInfos() returned %d entries, want %d", len(allChunkInfos), entryCount)
	}
}

func TestChunkInfoBlockReader_CalculateAggregates(t *testing.T) {
	entryCount := uint32(2)
	dataSize := 40 + int(entryCount)*32
	data := make([]byte, dataSize)
	endian := binary.LittleEndian

	// Set object header
	endian.PutUint32(data[24:28], types.ObjectTypeSpacemanCib)
	endian.PutUint32(data[36:40], entryCount) // CibChunkInfoCount

	// Fill first chunk info: 1000 total, 200 free
	offset := 40
	endian.PutUint32(data[offset+16:offset+20], 1000) // CiBlockCount
	endian.PutUint32(data[offset+20:offset+24], 200)  // CiFreeCount

	// Fill second chunk info: 2000 total, 400 free
	offset += 32
	endian.PutUint32(data[offset+16:offset+20], 2000) // CiBlockCount
	endian.PutUint32(data[offset+20:offset+24], 400)  // CiFreeCount

	reader, err := NewChunkInfoBlockReader(data, endian)
	if err != nil {
		t.Fatalf("NewChunkInfoBlockReader() failed: %v", err)
	}

	// Test aggregate calculations
	totalBlocks := reader.CalculateTotalBlocks()
	expectedTotal := uint64(3000) // 1000 + 2000
	if totalBlocks != expectedTotal {
		t.Errorf("CalculateTotalBlocks() = %d, want %d", totalBlocks, expectedTotal)
	}

	totalFree := reader.CalculateTotalFreeBlocks()
	expectedFree := uint64(600) // 200 + 400
	if totalFree != expectedFree {
		t.Errorf("CalculateTotalFreeBlocks() = %d, want %d", totalFree, expectedFree)
	}

	// Test utilization: (3000 - 600) / 3000 * 100 = 80%
	utilization := reader.CalculateOverallUtilization()
	expectedUtilization := 80.0
	if utilization != expectedUtilization {
		t.Errorf("CalculateOverallUtilization() = %.1f, want %.1f", utilization, expectedUtilization)
	}
}

func TestChunkInfoBlockReader_CalculateAggregatesZeroBlocks(t *testing.T) {
	data := make([]byte, 40) // Header only, no entries
	endian := binary.LittleEndian

	// Set object header
	endian.PutUint32(data[24:28], types.ObjectTypeSpacemanCib)
	endian.PutUint32(data[36:40], 0) // CibChunkInfoCount = 0

	reader, err := NewChunkInfoBlockReader(data, endian)
	if err != nil {
		t.Fatalf("NewChunkInfoBlockReader() failed: %v", err)
	}

	// Should return 0 for all aggregates
	if reader.CalculateTotalBlocks() != 0 {
		t.Errorf("CalculateTotalBlocks() = %d, want 0", reader.CalculateTotalBlocks())
	}

	if reader.CalculateTotalFreeBlocks() != 0 {
		t.Errorf("CalculateTotalFreeBlocks() = %d, want 0", reader.CalculateTotalFreeBlocks())
	}

	if reader.CalculateOverallUtilization() != 0.0 {
		t.Errorf("CalculateOverallUtilization() = %.1f, want 0.0", reader.CalculateOverallUtilization())
	}
}

func TestChunkInfoBlockReader_GetChunkInfoReader(t *testing.T) {
	entryCount := uint32(1)
	dataSize := 40 + int(entryCount)*32
	data := make([]byte, dataSize)
	endian := binary.LittleEndian

	// Set object header
	endian.PutUint32(data[24:28], types.ObjectTypeSpacemanCib)
	endian.PutUint32(data[36:40], entryCount) // CibChunkInfoCount

	// Fill chunk info with specific values
	offset := 40
	endian.PutUint64(data[offset:offset+8], 9999)   // CiXid
	endian.PutUint64(data[offset+8:offset+16], 8888) // CiAddr
	endian.PutUint32(data[offset+16:offset+20], 500) // CiBlockCount
	endian.PutUint32(data[offset+20:offset+24], 100) // CiFreeCount
	endian.PutUint64(data[offset+24:offset+32], 7777) // CiBitmapAddr

	reader, err := NewChunkInfoBlockReader(data, endian)
	if err != nil {
		t.Fatalf("NewChunkInfoBlockReader() failed: %v", err)
	}

	// Get chunk info reader for index 0
	chunkReader, err := reader.GetChunkInfoReader(0)
	if err != nil {
		t.Fatalf("GetChunkInfoReader(0) failed: %v", err)
	}

	// Verify the chunk reader has the correct values
	if chunkReader.TransactionID() != 9999 {
		t.Errorf("ChunkReader.TransactionID() = %d, want 9999", chunkReader.TransactionID())
	}

	if chunkReader.ChunkAddress() != 8888 {
		t.Errorf("ChunkReader.ChunkAddress() = %d, want 8888", chunkReader.ChunkAddress())
	}

	if chunkReader.BlockCount() != 500 {
		t.Errorf("ChunkReader.BlockCount() = %d, want 500", chunkReader.BlockCount())
	}

	if chunkReader.UsedCount() != 400 { // 500 - 100
		t.Errorf("ChunkReader.UsedCount() = %d, want 400", chunkReader.UsedCount())
	}
}