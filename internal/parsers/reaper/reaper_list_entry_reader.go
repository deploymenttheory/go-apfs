package reaper

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// reaperListEntryReader implements ReaperListEntryReader interface
type reaperListEntryReader struct {
	entry  *types.NxReapListEntryT
	endian binary.ByteOrder
}

// NewReaperListEntryReader creates a new reader for reaper list entries
func NewReaperListEntryReader(data []byte, endian binary.ByteOrder) (interfaces.ReaperListEntryReader, error) {
	if len(data) < 40 { // NrleNext(4) + NrleFlags(4) + NrleType(4) + NrleSize(4) + NrleFsOid(8) + NrleOid(8) + NrleXid(8)
		return nil, fmt.Errorf("data too small for reaper list entry: %d bytes", len(data))
	}

	entry, err := parseReaperListEntry(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reaper list entry: %w", err)
	}

	return &reaperListEntryReader{
		entry:  entry,
		endian: endian,
	}, nil
}

// parseReaperListEntry parses raw bytes into a NxReapListEntryT structure
func parseReaperListEntry(data []byte, endian binary.ByteOrder) (*types.NxReapListEntryT, error) {
	if len(data) < 40 {
		return nil, fmt.Errorf("insufficient data for reaper list entry: need 40 bytes, got %d", len(data))
	}

	entry := &types.NxReapListEntryT{}
	entry.NrleNext = endian.Uint32(data[0:4])
	entry.NrleFlags = endian.Uint32(data[4:8])
	entry.NrleType = endian.Uint32(data[8:12])
	entry.NrleSize = endian.Uint32(data[12:16])
	entry.NrleFsOid = types.OidT(endian.Uint64(data[16:24]))
	entry.NrleOid = types.OidT(endian.Uint64(data[24:32]))
	entry.NrleXid = types.XidT(endian.Uint64(data[32:40]))

	return entry, nil
}

// NextEntryIndex returns the index of the next entry
func (rler *reaperListEntryReader) NextEntryIndex() uint32 {
	return rler.entry.NrleNext
}

// Flags returns the flags for this entry
func (rler *reaperListEntryReader) Flags() uint32 {
	return rler.entry.NrleFlags
}

// IsValid checks if the entry is valid
func (rler *reaperListEntryReader) IsValid() bool {
	return (rler.entry.NrleFlags & types.NrleValid) != 0
}

// IsReapIDRecord checks if this is a reap ID record
func (rler *reaperListEntryReader) IsReapIDRecord() bool {
	return (rler.entry.NrleFlags & types.NrleReapIdRecord) != 0
}

// IsReadyToCall checks if the entry is ready to be called
func (rler *reaperListEntryReader) IsReadyToCall() bool {
	return (rler.entry.NrleFlags & types.NrleCall) != 0
}

// IsCompletionEntry checks if this is a completion entry
func (rler *reaperListEntryReader) IsCompletionEntry() bool {
	return (rler.entry.NrleFlags & types.NrleCompletion) != 0
}

// IsCleanupEntry checks if this is a cleanup entry
func (rler *reaperListEntryReader) IsCleanupEntry() bool {
	return (rler.entry.NrleFlags & types.NrleCleanup) != 0
}

// Type returns the type of object to reap
func (rler *reaperListEntryReader) Type() uint32 {
	return rler.entry.NrleType
}

// Size returns the size of the object to reap
func (rler *reaperListEntryReader) Size() uint32 {
	return rler.entry.NrleSize
}

// FileSystemOID returns the filesystem object identifier
func (rler *reaperListEntryReader) FileSystemOID() types.OidT {
	return rler.entry.NrleFsOid
}

// ObjectID returns the object identifier
func (rler *reaperListEntryReader) ObjectID() types.OidT {
	return rler.entry.NrleOid
}

// TransactionID returns the transaction identifier
func (rler *reaperListEntryReader) TransactionID() types.XidT {
	return rler.entry.NrleXid
}
