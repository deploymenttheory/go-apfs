package reaper

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// apfsReapStateReader implements ReaperStateReader interface
type apfsReapStateReader struct {
	state  *types.ApfsReapStateT
	endian binary.ByteOrder
}

// NewApfsReapStateReader creates a new reader for APFS reap state
func NewApfsReapStateReader(data []byte, endian binary.ByteOrder) (interfaces.ReaperStateReader, error) {
	if len(data) < 20 { // LastPbn(8) + CurSnapXid(8) + Phase(4)
		return nil, fmt.Errorf("data too small for APFS reap state: %d bytes", len(data))
	}

	state, err := parseApfsReapState(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse APFS reap state: %w", err)
	}

	return &apfsReapStateReader{
		state:  state,
		endian: endian,
	}, nil
}

// parseApfsReapState parses raw bytes into an ApfsReapStateT structure
func parseApfsReapState(data []byte, endian binary.ByteOrder) (*types.ApfsReapStateT, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("insufficient data for APFS reap state: need 20 bytes, got %d", len(data))
	}

	state := &types.ApfsReapStateT{}
	state.LastPbn = endian.Uint64(data[0:8])
	state.CurSnapXid = types.XidT(endian.Uint64(data[8:16]))
	state.Phase = endian.Uint32(data[16:20])

	return state, nil
}

// LastProcessedBlockNumber returns the last physical block number processed
func (arsr *apfsReapStateReader) LastProcessedBlockNumber() uint64 {
	return arsr.state.LastPbn
}

// CurrentSnapshotXID returns the current snapshot's transaction identifier
func (arsr *apfsReapStateReader) CurrentSnapshotXID() types.XidT {
	return arsr.state.CurSnapXid
}

// PhaseDescription returns a human-readable description of the current reaping phase
func (arsr *apfsReapStateReader) PhaseDescription() string {
	switch arsr.state.Phase {
	case types.ApfsReapPhaseStart:
		return "Start"
	case types.ApfsReapPhaseSnapshots:
		return "Snapshots"
	case types.ApfsReapPhaseActiveFs:
		return "Active Filesystem"
	case types.ApfsReapPhaseDestroyOmap:
		return "Destroy Object Map"
	case types.ApfsReapPhaseDone:
		return "Done"
	default:
		return "Unknown"
	}
}

// Phase returns the current reaping phase
func (arsr *apfsReapStateReader) Phase() uint32 {
	return arsr.state.Phase
}
