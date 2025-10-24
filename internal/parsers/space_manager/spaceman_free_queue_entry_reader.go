package spacemanager

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// SpacemanFreeQueueEntryReader provides parsing capabilities for spaceman_free_queue_entry_t structures
// Free queue entries represent available blocks in the space manager's free queue B-trees
type SpacemanFreeQueueEntryReader struct {
	freeQueueEntry *types.SpacemanFreeQueueEntryT
	data           []byte
	endian         binary.ByteOrder
}

// NewSpacemanFreeQueueEntryReader creates a new free queue entry reader
// Free queue entries track available space with transaction identifiers for consistency
func NewSpacemanFreeQueueEntryReader(data []byte, endian binary.ByteOrder) (*SpacemanFreeQueueEntryReader, error) {
	// spaceman_free_queue_entry_t structure size:
	// spaceman_free_queue_key_t (16 bytes) + spaceman_free_queue_val_t (8 bytes) = 24 bytes
	if len(data) < 24 {
		return nil, fmt.Errorf("data too small for free queue entry: %d bytes, need at least 24", len(data))
	}

	freeQueueEntry, err := parseSpacemanFreeQueueEntry(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse free queue entry: %w", err)
	}

	return &SpacemanFreeQueueEntryReader{
		freeQueueEntry: freeQueueEntry,
		data:           data,
		endian:         endian,
	}, nil
}

// parseSpacemanFreeQueueEntry parses raw bytes into a SpacemanFreeQueueEntryT structure
// This follows the exact layout of spaceman_free_queue_entry_t from Apple File System Reference
func parseSpacemanFreeQueueEntry(data []byte, endian binary.ByteOrder) (*types.SpacemanFreeQueueEntryT, error) {
	if len(data) < 24 {
		return nil, fmt.Errorf("insufficient data for free queue entry")
	}

	sfqe := &types.SpacemanFreeQueueEntryT{}
	offset := 0

	// Parse spaceman_free_queue_key_t sfqe_key (16 bytes)
	// The key contains transaction ID and physical address
	// xid_t sfqk_xid - Transaction identifier (8 bytes)
	sfqe.SfqeKey.SfqkXid = types.XidT(endian.Uint64(data[offset : offset+8]))
	offset += 8

	// paddr_t sfqk_paddr - Physical address (8 bytes)
	sfqe.SfqeKey.SfqkPaddr = types.Paddr(endian.Uint64(data[offset : offset+8]))
	offset += 8

	// Parse spaceman_free_queue_val_t sfqe_count (8 bytes)
	// uint64_t - Count of free blocks at this location
	sfqe.SfqeCount = types.SpacemanFreeQueueValT(endian.Uint64(data[offset : offset+8]))
	offset += 8

	return sfqe, nil
}

// GetFreeQueueEntry returns the free queue entry structure
// Provides direct access to the underlying spaceman_free_queue_entry_t for advanced operations
func (sfqer *SpacemanFreeQueueEntryReader) GetFreeQueueEntry() *types.SpacemanFreeQueueEntryT {
	return sfqer.freeQueueEntry
}

// GetKey returns the free queue key
// Provides access to the transaction ID and physical address that form the entry key
func (sfqer *SpacemanFreeQueueEntryReader) GetKey() *types.SpacemanFreeQueueKeyT {
	return &sfqer.freeQueueEntry.SfqeKey
}

// TransactionID returns the transaction identifier for this free queue entry
// Indicates when this free space was made available
func (sfqer *SpacemanFreeQueueEntryReader) TransactionID() types.XidT {
	return sfqer.freeQueueEntry.SfqeKey.SfqkXid
}

// PhysicalAddress returns the physical address of the free space
// Location on the storage device where the free blocks are located
func (sfqer *SpacemanFreeQueueEntryReader) PhysicalAddress() types.Paddr {
	return sfqer.freeQueueEntry.SfqeKey.SfqkPaddr
}

// FreeBlockCount returns the count of free blocks
// Number of contiguous free blocks available at the physical address
func (sfqer *SpacemanFreeQueueEntryReader) FreeBlockCount() types.SpacemanFreeQueueValT {
	return sfqer.freeQueueEntry.SfqeCount
}

// FreeBlockCountAsUint64 returns the free block count as a standard uint64
// Convenience method for numeric operations
func (sfqer *SpacemanFreeQueueEntryReader) FreeBlockCountAsUint64() uint64 {
	return uint64(sfqer.freeQueueEntry.SfqeCount)
}

// IsValidEntry returns true if the entry has valid data
// Checks for non-zero address and block count
func (sfqer *SpacemanFreeQueueEntryReader) IsValidEntry() bool {
	return sfqer.freeQueueEntry.SfqeKey.SfqkPaddr != 0 && sfqer.freeQueueEntry.SfqeCount > 0
}

// IsRecentTransaction returns true if the transaction ID is greater than the specified threshold
// Useful for identifying recently freed space
func (sfqer *SpacemanFreeQueueEntryReader) IsRecentTransaction(thresholdXid types.XidT) bool {
	return sfqer.freeQueueEntry.SfqeKey.SfqkXid > thresholdXid
}

// CalculateTotalFreeSpace returns the total free space in bytes for a given block size
// Multiplies the free block count by the provided block size
func (sfqer *SpacemanFreeQueueEntryReader) CalculateTotalFreeSpace(blockSize uint32) uint64 {
	return uint64(sfqer.freeQueueEntry.SfqeCount) * uint64(blockSize)
}

// CompareByAddress compares this entry's address with another address
// Returns: -1 if this < other, 0 if equal, 1 if this > other
func (sfqer *SpacemanFreeQueueEntryReader) CompareByAddress(otherAddr types.Paddr) int {
	thisAddr := sfqer.freeQueueEntry.SfqeKey.SfqkPaddr
	if thisAddr < otherAddr {
		return -1
	} else if thisAddr > otherAddr {
		return 1
	}
	return 0
}

// CompareByTransaction compares this entry's transaction ID with another transaction ID
// Returns: -1 if this < other, 0 if equal, 1 if this > other
func (sfqer *SpacemanFreeQueueEntryReader) CompareByTransaction(otherXid types.XidT) int {
	thisXid := sfqer.freeQueueEntry.SfqeKey.SfqkXid
	if thisXid < otherXid {
		return -1
	} else if thisXid > otherXid {
		return 1
	}
	return 0
}

// IsContiguousWith checks if this entry is physically contiguous with another address
// Returns true if the end of this entry's blocks touches the start of the other address
func (sfqer *SpacemanFreeQueueEntryReader) IsContiguousWith(otherAddr types.Paddr, blockSize uint32) bool {
	thisEndAddr := sfqer.freeQueueEntry.SfqeKey.SfqkPaddr + types.Paddr(uint64(sfqer.freeQueueEntry.SfqeCount)*uint64(blockSize))
	return thisEndAddr == otherAddr
}

// GetAddressRange returns the start and end addresses of the free space
// Calculates the physical address range covered by this free queue entry
func (sfqer *SpacemanFreeQueueEntryReader) GetAddressRange(blockSize uint32) (types.Paddr, types.Paddr) {
	startAddr := sfqer.freeQueueEntry.SfqeKey.SfqkPaddr
	endAddr := startAddr + types.Paddr(uint64(sfqer.freeQueueEntry.SfqeCount)*uint64(blockSize))
	return startAddr, endAddr
}

// String returns a human-readable representation of the free queue entry
// Useful for debugging and logging
func (sfqer *SpacemanFreeQueueEntryReader) String() string {
	return fmt.Sprintf("FreeQueueEntry{XID: %d, Addr: 0x%x, Count: %d}",
		sfqer.freeQueueEntry.SfqeKey.SfqkXid,
		sfqer.freeQueueEntry.SfqeKey.SfqkPaddr,
		sfqer.freeQueueEntry.SfqeCount)
}