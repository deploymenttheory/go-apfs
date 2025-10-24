package reaper

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// reaperListReader implements ReaperListReader interface
type reaperListReader struct {
	list    *types.NxReapListPhysT
	entries []interfaces.ReaperListEntryReader
	endian  binary.ByteOrder
}

// NewReaperListReader creates a new reader for reaper lists
func NewReaperListReader(data []byte, endian binary.ByteOrder) (interfaces.ReaperListReader, error) {
	// ObjPhysT(40) + NrlNext(8) + NrlFlags(4) + NrlMax(4) + NrlCount(4) + NrlFirst(4) + NrlLast(4) + NrlFree(4) = 72 bytes minimum
	if len(data) < 72 {
		return nil, fmt.Errorf("data too small for reaper list: %d bytes", len(data))
	}

	list, entries, err := parseReaperList(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reaper list: %w", err)
	}

	return &reaperListReader{
		list:    list,
		entries: entries,
		endian:  endian,
	}, nil
}

// parseReaperList parses raw bytes into a NxReapListPhysT structure and its entries
func parseReaperList(data []byte, endian binary.ByteOrder) (*types.NxReapListPhysT, []interfaces.ReaperListEntryReader, error) {
	if len(data) < 72 {
		return nil, nil, fmt.Errorf("insufficient data for reaper list: need 72 bytes, got %d", len(data))
	}

	list := &types.NxReapListPhysT{}

	// Parse ObjPhysT header (40 bytes)
	copy(list.NrlO.OChecksum[:], data[0:32])
	list.NrlO.OOid = types.OidT(endian.Uint64(data[32:40]))
	list.NrlO.OXid = types.XidT(endian.Uint64(data[40:48]))
	list.NrlO.OType = endian.Uint32(data[48:52])
	list.NrlO.OSubtype = endian.Uint32(data[52:56])

	// Parse list fields
	list.NrlNext = types.OidT(endian.Uint64(data[56:64]))
	list.NrlFlags = endian.Uint32(data[64:68])
	list.NrlMax = endian.Uint32(data[68:72])
	list.NrlCount = endian.Uint32(data[72:76])
	list.NrlFirst = endian.Uint32(data[76:80])
	list.NrlLast = endian.Uint32(data[80:84])
	list.NrlFree = endian.Uint32(data[84:88])

	// Parse entries (each entry is 40 bytes: NxReapListEntryT)
	entries := make([]interfaces.ReaperListEntryReader, 0)
	entrySize := 40
	offset := 88

	for i := uint32(0); i < list.NrlCount; i++ {
		if offset+entrySize > len(data) {
			break
		}

		entryReader, err := NewReaperListEntryReader(data[offset:offset+entrySize], endian)
		if err != nil {
			continue
		}
		entries = append(entries, entryReader)
		offset += entrySize
	}

	list.NrlEntries = make([]types.NxReapListEntryT, len(entries))

	return list, entries, nil
}

// NextListOID returns the object identifier of the next reaper list
func (rlr *reaperListReader) NextListOID() types.OidT {
	return rlr.list.NrlNext
}

// Flags returns the flags for this reaper list
func (rlr *reaperListReader) Flags() uint32 {
	return rlr.list.NrlFlags
}

// MaxEntries returns the maximum number of entries in this list
func (rlr *reaperListReader) MaxEntries() uint32 {
	return rlr.list.NrlMax
}

// CurrentEntryCount returns the number of entries currently in the list
func (rlr *reaperListReader) CurrentEntryCount() uint32 {
	return rlr.list.NrlCount
}

// Entries returns the list of reaper list entries
func (rlr *reaperListReader) Entries() []interfaces.ReaperListEntryReader {
	return rlr.entries
}
