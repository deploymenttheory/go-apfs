package spacemanager

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// SpacemanFreeQueueReader provides high-level access to free queue information
// Wraps spaceman_free_queue_t to provide convenient methods for queue metadata
// Note: The actual free queue entries are stored in a B-tree referenced by SfqTreeOid
type SpacemanFreeQueueReader struct {
	queue  *types.SpacemanFreeQueueT
	data   []byte
	endian binary.ByteOrder
}

// NewSpacemanFreeQueueReader creates a new spaceman free queue reader
// Free queue structure is 40 bytes
func NewSpacemanFreeQueueReader(data []byte, endian binary.ByteOrder) (*SpacemanFreeQueueReader, error) {
	if len(data) < 40 {
		return nil, fmt.Errorf("data too small for spaceman free queue: %d bytes, need at least 40", len(data))
	}

	queue, err := parseSpacemanFreeQueue(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spaceman free queue: %w", err)
	}

	return &SpacemanFreeQueueReader{
		queue:  queue,
		data:   data,
		endian: endian,
	}, nil
}

// parseSpacemanFreeQueue parses raw bytes into SpacemanFreeQueueT
func parseSpacemanFreeQueue(data []byte, endian binary.ByteOrder) (*types.SpacemanFreeQueueT, error) {
	if len(data) < 40 {
		return nil, fmt.Errorf("insufficient data for spaceman free queue")
	}

	sfq := &types.SpacemanFreeQueueT{}
	offset := 0

	// Parse count (uint64)
	sfq.SfqCount = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse tree OID (uint64)
	sfq.SfqTreeOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8

	// Parse oldest XID (uint64)
	sfq.SfqOldestXid = types.XidT(endian.Uint64(data[offset : offset+8]))
	offset += 8

	// Parse tree node limit (uint16)
	sfq.SfqTreeNodeLimit = endian.Uint16(data[offset : offset+2])
	offset += 2

	// Parse pad16 (uint16)
	sfq.SfqPad16 = endian.Uint16(data[offset : offset+2])
	offset += 2

	// Parse pad32 (uint32)
	sfq.SfqPad32 = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse reserved (uint64)
	sfq.SfqReserved = endian.Uint64(data[offset : offset+8])
	offset += 8

	return sfq, nil
}

// GetQueue returns the queue structure
func (sqr *SpacemanFreeQueueReader) GetQueue() *types.SpacemanFreeQueueT {
	return sqr.queue
}

// Count returns the number of entries in this free queue
func (sqr *SpacemanFreeQueueReader) Count() uint64 {
	return sqr.queue.SfqCount
}

// TreeOID returns the Object ID of the B-tree containing queue entries
// This B-tree stores SpacemanFreeQueueEntryT structures
func (sqr *SpacemanFreeQueueReader) TreeOID() types.OidT {
	return sqr.queue.SfqTreeOid
}

// OldestXID returns the oldest transaction identifier in this queue
func (sqr *SpacemanFreeQueueReader) OldestXID() types.XidT {
	return sqr.queue.SfqOldestXid
}

// TreeNodeLimit returns the limit on the number of nodes in the B-tree
func (sqr *SpacemanFreeQueueReader) TreeNodeLimit() uint16 {
	return sqr.queue.SfqTreeNodeLimit
}

// IsEmpty returns true if the queue has no entries
func (sqr *SpacemanFreeQueueReader) IsEmpty() bool {
	return sqr.queue.SfqCount == 0
}

// HasTreeOID returns true if this queue references a valid B-tree
func (sqr *SpacemanFreeQueueReader) HasTreeOID() bool {
	return sqr.queue.SfqTreeOid != 0
}

// IsValid returns true if the queue has valid metadata
func (sqr *SpacemanFreeQueueReader) IsValid() bool {
	return sqr.HasTreeOID() || sqr.queue.SfqCount == 0
}

// Summary returns a human-readable summary of the free queue
func (sqr *SpacemanFreeQueueReader) Summary() string {
	status := "empty"
	if !sqr.IsEmpty() {
		status = fmt.Sprintf("%d entries", sqr.queue.SfqCount)
	}
	return fmt.Sprintf("FreeQueue{Status: %s, TreeOID: %d, OldestXID: %d, NodeLimit: %d}",
		status,
		sqr.queue.SfqTreeOid,
		sqr.queue.SfqOldestXid,
		sqr.queue.SfqTreeNodeLimit)
}
