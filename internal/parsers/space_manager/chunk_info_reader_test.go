package spacemanager

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestNewChunkInfoReader(t *testing.T) {
	// Create chunk info data (32 bytes)
	data := make([]byte, 32)
	endian := binary.LittleEndian

	// Fill chunk info fields
	endian.PutUint64(data[0:8], 12345)   // CiXid - transaction ID
	endian.PutUint64(data[8:16], 67890)  // CiAddr - chunk address
	endian.PutUint32(data[16:20], 1024)  // CiBlockCount - total blocks
	endian.PutUint32(data[20:24], 256)   // CiFreeCount - free blocks
	endian.PutUint64(data[24:32], 54321) // CiBitmapAddr - bitmap address

	reader, err := NewChunkInfoReader(data, endian)
	if err != nil {
		t.Fatalf("NewChunkInfoReader() failed: %v", err)
	}

	if reader.TransactionID() != 12345 {
		t.Errorf("TransactionID() = %d, want 12345", reader.TransactionID())
	}

	if reader.ChunkAddress() != 67890 {
		t.Errorf("ChunkAddress() = %d, want 67890", reader.ChunkAddress())
	}

	if reader.BlockCount() != 1024 {
		t.Errorf("BlockCount() = %d, want 1024", reader.BlockCount())
	}

	if reader.FreeCount() != 256 {
		t.Errorf("FreeCount() = %d, want 256", reader.FreeCount())
	}

	if reader.BitmapAddress() != types.Paddr(54321) {
		t.Errorf("BitmapAddress() = %d, want 54321", reader.BitmapAddress())
	}
}

func TestChunkInfoReader_TooSmall(t *testing.T) {
	data := make([]byte, 20) // Too small (need 32)

	_, err := NewChunkInfoReader(data, binary.LittleEndian)
	if err == nil {
		t.Error("NewChunkInfoReader() should have failed with too small data")
	}
}

func TestChunkInfoReader_UsedCount(t *testing.T) {
	data := make([]byte, 32)
	endian := binary.LittleEndian

	// Set block counts
	endian.PutUint32(data[16:20], 1000) // CiBlockCount - total blocks
	endian.PutUint32(data[20:24], 300)  // CiFreeCount - free blocks

	reader, err := NewChunkInfoReader(data, endian)
	if err != nil {
		t.Fatalf("NewChunkInfoReader() failed: %v", err)
	}

	expectedUsed := uint32(700) // 1000 - 300
	if reader.UsedCount() != expectedUsed {
		t.Errorf("UsedCount() = %d, want %d", reader.UsedCount(), expectedUsed)
	}
}

func TestChunkInfoReader_UsedCountOverflow(t *testing.T) {
	data := make([]byte, 32)
	endian := binary.LittleEndian

	// Set invalid counts (free > total)
	endian.PutUint32(data[16:20], 100) // CiBlockCount - total blocks
	endian.PutUint32(data[20:24], 200) // CiFreeCount - free blocks (invalid)

	reader, err := NewChunkInfoReader(data, endian)
	if err != nil {
		t.Fatalf("NewChunkInfoReader() failed: %v", err)
	}

	// Should return 0 when free count exceeds total count
	if reader.UsedCount() != 0 {
		t.Errorf("UsedCount() = %d, want 0 for invalid free count", reader.UsedCount())
	}
}

func TestChunkInfoReader_IsFull(t *testing.T) {
	data := make([]byte, 32)
	endian := binary.LittleEndian

	// Set chunk as full (no free blocks)
	endian.PutUint32(data[16:20], 1000) // CiBlockCount - total blocks
	endian.PutUint32(data[20:24], 0)    // CiFreeCount - no free blocks

	reader, err := NewChunkInfoReader(data, endian)
	if err != nil {
		t.Fatalf("NewChunkInfoReader() failed: %v", err)
	}

	if !reader.IsFull() {
		t.Error("IsFull() should be true when no free blocks")
	}

	if reader.IsEmpty() {
		t.Error("IsEmpty() should be false when no free blocks")
	}
}

func TestChunkInfoReader_IsEmpty(t *testing.T) {
	data := make([]byte, 32)
	endian := binary.LittleEndian

	// Set chunk as empty (all blocks free)
	endian.PutUint32(data[16:20], 1000) // CiBlockCount - total blocks
	endian.PutUint32(data[20:24], 1000) // CiFreeCount - all blocks free

	reader, err := NewChunkInfoReader(data, endian)
	if err != nil {
		t.Fatalf("NewChunkInfoReader() failed: %v", err)
	}

	if !reader.IsEmpty() {
		t.Error("IsEmpty() should be true when all blocks are free")
	}

	if reader.IsFull() {
		t.Error("IsFull() should be false when all blocks are free")
	}
}

func TestChunkInfoReader_UtilizationPercentage(t *testing.T) {
	tests := []struct {
		name        string
		totalBlocks uint32
		freeBlocks  uint32
		expected    float64
	}{
		{"Half used", 1000, 500, 50.0},
		{"Fully used", 1000, 0, 100.0},
		{"Empty", 1000, 1000, 0.0},
		{"Quarter used", 1000, 750, 25.0},
		{"Zero blocks", 0, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, 32)
			endian := binary.LittleEndian

			endian.PutUint32(data[16:20], tt.totalBlocks) // CiBlockCount
			endian.PutUint32(data[20:24], tt.freeBlocks)  // CiFreeCount

			reader, err := NewChunkInfoReader(data, endian)
			if err != nil {
				t.Fatalf("NewChunkInfoReader() failed: %v", err)
			}

			utilization := reader.UtilizationPercentage()
			if utilization != tt.expected {
				t.Errorf("UtilizationPercentage() = %.1f, want %.1f", utilization, tt.expected)
			}
		})
	}
}

func TestChunkInfoReader_GetChunkInfo(t *testing.T) {
	data := make([]byte, 32)
	endian := binary.LittleEndian

	// Fill with test data
	endian.PutUint64(data[0:8], 99999)

	reader, err := NewChunkInfoReader(data, endian)
	if err != nil {
		t.Fatalf("NewChunkInfoReader() failed: %v", err)
	}

	chunkInfo := reader.GetChunkInfo()
	if chunkInfo == nil {
		t.Error("GetChunkInfo() returned nil")
		return
	}

	if chunkInfo.CiXid != 99999 {
		t.Errorf("GetChunkInfo().CiXid = %d, want 99999", chunkInfo.CiXid)
	}
}
