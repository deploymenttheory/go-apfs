package reaper

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// omapCleanupStateReader implements SnapshotCleanupState interface
type omapCleanupStateReader struct {
	state  *types.OmapCleanupStateT
	endian binary.ByteOrder
}

// NewOmapCleanupStateReader creates a new reader for object map cleanup state
func NewOmapCleanupStateReader(data []byte, endian binary.ByteOrder) (interfaces.SnapshotCleanupState, error) {
	// OmcCleaning(4) + OmcOmsflags(4) + OmcSxidprev(8) + OmcSxidstart(8) + OmcSxidend(8) + OmcSxidnext(8) + OmcCurkey(16)
	if len(data) < 56 {
		return nil, fmt.Errorf("data too small for omap cleanup state: %d bytes", len(data))
	}

	state, err := parseOmapCleanupState(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse omap cleanup state: %w", err)
	}

	return &omapCleanupStateReader{
		state:  state,
		endian: endian,
	}, nil
}

// parseOmapCleanupState parses raw bytes into an OmapCleanupStateT structure
func parseOmapCleanupState(data []byte, endian binary.ByteOrder) (*types.OmapCleanupStateT, error) {
	if len(data) < 56 {
		return nil, fmt.Errorf("insufficient data for omap cleanup state: need 56 bytes, got %d", len(data))
	}

	state := &types.OmapCleanupStateT{}
	state.OmcCleaning = endian.Uint32(data[0:4])
	state.OmcOmsflags = endian.Uint32(data[4:8])
	state.OmcSxidprev = types.XidT(endian.Uint64(data[8:16]))
	state.OmcSxidstart = types.XidT(endian.Uint64(data[16:24]))
	state.OmcSxidend = types.XidT(endian.Uint64(data[24:32]))
	state.OmcSxidnext = types.XidT(endian.Uint64(data[32:40]))

	// Parse OmapKeyT (16 bytes): OkOid(8) + OkXid(8)
	state.OmcCurkey.OkOid = types.OidT(endian.Uint64(data[40:48]))
	state.OmcCurkey.OkXid = types.XidT(endian.Uint64(data[48:56]))

	return state, nil
}

// IsCleaning checks if the cleanup process is active
func (ocsr *omapCleanupStateReader) IsCleaning() bool {
	return ocsr.state.OmcCleaning != 0
}

// SnapshotFlags returns the flags for the snapshots being deleted
func (ocsr *omapCleanupStateReader) SnapshotFlags() uint32 {
	return ocsr.state.OmcOmsflags
}

// PreviousSnapshotXID returns the transaction ID of the snapshot before deletion
func (ocsr *omapCleanupStateReader) PreviousSnapshotXID() types.XidT {
	return ocsr.state.OmcSxidprev
}

// StartSnapshotXID returns the transaction ID of the first snapshot being deleted
func (ocsr *omapCleanupStateReader) StartSnapshotXID() types.XidT {
	return ocsr.state.OmcSxidstart
}

// EndSnapshotXID returns the transaction ID of the last snapshot being deleted
func (ocsr *omapCleanupStateReader) EndSnapshotXID() types.XidT {
	return ocsr.state.OmcSxidend
}

// NextSnapshotXID returns the transaction ID of the snapshot after deletion
func (ocsr *omapCleanupStateReader) NextSnapshotXID() types.XidT {
	return ocsr.state.OmcSxidnext
}
