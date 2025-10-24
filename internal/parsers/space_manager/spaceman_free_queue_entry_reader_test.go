package spacemanager

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestNewSpacemanFreeQueueEntryReader(t *testing.T) {
	// Create free queue entry data (24 bytes)
	data := make([]byte, 24)
	endian := binary.LittleEndian

	// Fill free queue entry fields
	endian.PutUint64(data[0:8], 54321)   // SfqkXid - transaction ID
	endian.PutUint64(data[8:16], 0x1000) // SfqkPaddr - physical address
	endian.PutUint64(data[16:24], 128)   // SfqeCount - free block count

	reader, err := NewSpacemanFreeQueueEntryReader(data, endian)
	if err != nil {
		t.Fatalf("NewSpacemanFreeQueueEntryReader() failed: %v", err)
	}

	if reader.TransactionID() != 54321 {
		t.Errorf("TransactionID() = %d, want 54321", reader.TransactionID())
	}

	if reader.PhysicalAddress() != 0x1000 {
		t.Errorf("PhysicalAddress() = 0x%x, want 0x1000", reader.PhysicalAddress())
	}

	if reader.FreeBlockCount() != 128 {
		t.Errorf("FreeBlockCount() = %d, want 128", reader.FreeBlockCount())
	}

	if reader.FreeBlockCountAsUint64() != 128 {
		t.Errorf("FreeBlockCountAsUint64() = %d, want 128", reader.FreeBlockCountAsUint64())
	}
}

func TestSpacemanFreeQueueEntryReader_TooSmall(t *testing.T) {
	data := make([]byte, 20) // Too small (need 24)

	_, err := NewSpacemanFreeQueueEntryReader(data, binary.LittleEndian)
	if err == nil {
		t.Error("NewSpacemanFreeQueueEntryReader() should have failed with too small data")
	}
}

func TestSpacemanFreeQueueEntryReader_IsValidEntry(t *testing.T) {
	tests := []struct {
		name        string
		xid         uint64
		addr        uint64
		count       uint64
		expectValid bool
	}{
		{"Valid entry", 123, 0x1000, 64, true},
		{"Zero address", 123, 0, 64, false},
		{"Zero count", 123, 0x1000, 0, false},
		{"Both zero", 123, 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, 24)
			endian := binary.LittleEndian

			endian.PutUint64(data[0:8], tt.xid)
			endian.PutUint64(data[8:16], tt.addr)
			endian.PutUint64(data[16:24], tt.count)

			reader, err := NewSpacemanFreeQueueEntryReader(data, endian)
			if err != nil {
				t.Fatalf("NewSpacemanFreeQueueEntryReader() failed: %v", err)
			}

			if reader.IsValidEntry() != tt.expectValid {
				t.Errorf("IsValidEntry() = %v, want %v", reader.IsValidEntry(), tt.expectValid)
			}
		})
	}
}

func TestSpacemanFreeQueueEntryReader_IsRecentTransaction(t *testing.T) {
	data := make([]byte, 24)
	endian := binary.LittleEndian

	// Set transaction ID to 1000
	endian.PutUint64(data[0:8], 1000)

	reader, err := NewSpacemanFreeQueueEntryReader(data, endian)
	if err != nil {
		t.Fatalf("NewSpacemanFreeQueueEntryReader() failed: %v", err)
	}

	// Test with lower threshold (should be recent)
	if !reader.IsRecentTransaction(types.XidT(500)) {
		t.Error("IsRecentTransaction(500) should be true for XID 1000")
	}

	// Test with higher threshold (should not be recent)
	if reader.IsRecentTransaction(types.XidT(1500)) {
		t.Error("IsRecentTransaction(1500) should be false for XID 1000")
	}

	// Test with equal threshold (should not be recent)
	if reader.IsRecentTransaction(types.XidT(1000)) {
		t.Error("IsRecentTransaction(1000) should be false for XID 1000")
	}
}

func TestSpacemanFreeQueueEntryReader_CalculateTotalFreeSpace(t *testing.T) {
	data := make([]byte, 24)
	endian := binary.LittleEndian

	// Set free block count to 100
	endian.PutUint64(data[16:24], 100)

	reader, err := NewSpacemanFreeQueueEntryReader(data, endian)
	if err != nil {
		t.Fatalf("NewSpacemanFreeQueueEntryReader() failed: %v", err)
	}

	// Test with 4096 byte blocks
	blockSize := uint32(4096)
	expectedSpace := uint64(100 * 4096) // 409,600 bytes
	totalSpace := reader.CalculateTotalFreeSpace(blockSize)

	if totalSpace != expectedSpace {
		t.Errorf("CalculateTotalFreeSpace(%d) = %d, want %d", blockSize, totalSpace, expectedSpace)
	}
}

func TestSpacemanFreeQueueEntryReader_CompareByAddress(t *testing.T) {
	data := make([]byte, 24)
	endian := binary.LittleEndian

	// Set address to 0x2000
	endian.PutUint64(data[8:16], 0x2000)

	reader, err := NewSpacemanFreeQueueEntryReader(data, endian)
	if err != nil {
		t.Fatalf("NewSpacemanFreeQueueEntryReader() failed: %v", err)
	}

	tests := []struct {
		name      string
		otherAddr types.Paddr
		expected  int
	}{
		{"Less than", 0x3000, -1},
		{"Equal", 0x2000, 0},
		{"Greater than", 0x1000, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reader.CompareByAddress(tt.otherAddr)
			if result != tt.expected {
				t.Errorf("CompareByAddress(0x%x) = %d, want %d", tt.otherAddr, result, tt.expected)
			}
		})
	}
}

func TestSpacemanFreeQueueEntryReader_CompareByTransaction(t *testing.T) {
	data := make([]byte, 24)
	endian := binary.LittleEndian

	// Set transaction ID to 500
	endian.PutUint64(data[0:8], 500)

	reader, err := NewSpacemanFreeQueueEntryReader(data, endian)
	if err != nil {
		t.Fatalf("NewSpacemanFreeQueueEntryReader() failed: %v", err)
	}

	tests := []struct {
		name     string
		otherXid types.XidT
		expected int
	}{
		{"Less than", 600, -1},
		{"Equal", 500, 0},
		{"Greater than", 400, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reader.CompareByTransaction(tt.otherXid)
			if result != tt.expected {
				t.Errorf("CompareByTransaction(%d) = %d, want %d", tt.otherXid, result, tt.expected)
			}
		})
	}
}

func TestSpacemanFreeQueueEntryReader_IsContiguousWith(t *testing.T) {
	data := make([]byte, 24)
	endian := binary.LittleEndian

	// Set address to 0x1000 and count to 10 blocks
	endian.PutUint64(data[8:16], 0x1000) // Start address
	endian.PutUint64(data[16:24], 10)    // 10 blocks

	reader, err := NewSpacemanFreeQueueEntryReader(data, endian)
	if err != nil {
		t.Fatalf("NewSpacemanFreeQueueEntryReader() failed: %v", err)
	}

	blockSize := uint32(4096)
	// This entry covers 0x1000 to 0x1000 + (10 * 4096) = 0x1000 to 0xB000
	// So end address is 0xB000

	tests := []struct {
		name      string
		otherAddr types.Paddr
		expected  bool
	}{
		{"Contiguous", 0xB000, true},      // Touches end of this entry
		{"Not contiguous", 0xC000, false}, // Gap between
		{"Overlapping", 0xA000, false},    // Overlaps with this entry
		{"Before", 0x500, false},          // Before this entry
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reader.IsContiguousWith(tt.otherAddr, blockSize)
			if result != tt.expected {
				t.Errorf("IsContiguousWith(0x%x, %d) = %v, want %v", tt.otherAddr, blockSize, result, tt.expected)
			}
		})
	}
}

func TestSpacemanFreeQueueEntryReader_GetAddressRange(t *testing.T) {
	data := make([]byte, 24)
	endian := binary.LittleEndian

	// Set address to 0x2000 and count to 5 blocks
	endian.PutUint64(data[8:16], 0x2000) // Start address
	endian.PutUint64(data[16:24], 5)     // 5 blocks

	reader, err := NewSpacemanFreeQueueEntryReader(data, endian)
	if err != nil {
		t.Fatalf("NewSpacemanFreeQueueEntryReader() failed: %v", err)
	}

	blockSize := uint32(4096)
	startAddr, endAddr := reader.GetAddressRange(blockSize)

	expectedStart := types.Paddr(0x2000)
	expectedEnd := types.Paddr(0x2000 + 5*4096) // 0x2000 + 0x5000 = 0x7000

	if startAddr != expectedStart {
		t.Errorf("GetAddressRange() start = 0x%x, want 0x%x", startAddr, expectedStart)
	}

	if endAddr != expectedEnd {
		t.Errorf("GetAddressRange() end = 0x%x, want 0x%x", endAddr, expectedEnd)
	}
}

func TestSpacemanFreeQueueEntryReader_GetKey(t *testing.T) {
	data := make([]byte, 24)
	endian := binary.LittleEndian

	// Set key fields
	endian.PutUint64(data[0:8], 12345)   // XID
	endian.PutUint64(data[8:16], 0x5000) // Address

	reader, err := NewSpacemanFreeQueueEntryReader(data, endian)
	if err != nil {
		t.Fatalf("NewSpacemanFreeQueueEntryReader() failed: %v", err)
	}

	key := reader.GetKey()
	if key == nil {
		t.Error("GetKey() returned nil")
		return
	}

	if key.SfqkXid != 12345 {
		t.Errorf("GetKey().SfqkXid = %d, want 12345", key.SfqkXid)
	}

	if key.SfqkPaddr != 0x5000 {
		t.Errorf("GetKey().SfqkPaddr = 0x%x, want 0x5000", key.SfqkPaddr)
	}
}

func TestSpacemanFreeQueueEntryReader_String(t *testing.T) {
	data := make([]byte, 24)
	endian := binary.LittleEndian

	// Set specific values for predictable string output
	endian.PutUint64(data[0:8], 999)     // XID
	endian.PutUint64(data[8:16], 0x1234) // Address
	endian.PutUint64(data[16:24], 42)    // Count

	reader, err := NewSpacemanFreeQueueEntryReader(data, endian)
	if err != nil {
		t.Fatalf("NewSpacemanFreeQueueEntryReader() failed: %v", err)
	}

	str := reader.String()
	expected := "FreeQueueEntry{XID: 999, Addr: 0x1234, Count: 42}"

	if str != expected {
		t.Errorf("String() = %q, want %q", str, expected)
	}
}

func TestSpacemanFreeQueueEntryReader_GetFreeQueueEntry(t *testing.T) {
	data := make([]byte, 24)
	endian := binary.LittleEndian

	// Fill with test data
	endian.PutUint64(data[0:8], 77777)

	reader, err := NewSpacemanFreeQueueEntryReader(data, endian)
	if err != nil {
		t.Fatalf("NewSpacemanFreeQueueEntryReader() failed: %v", err)
	}

	entry := reader.GetFreeQueueEntry()
	if entry == nil {
		t.Error("GetFreeQueueEntry() returned nil")
		return
	}

	if entry.SfqeKey.SfqkXid != 77777 {
		t.Errorf("GetFreeQueueEntry().SfqeKey.SfqkXid = %d, want 77777", entry.SfqeKey.SfqkXid)
	}
}
