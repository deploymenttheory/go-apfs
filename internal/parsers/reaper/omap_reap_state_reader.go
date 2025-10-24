package reaper

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// omapReapStateReader implements ObjectMapReaperState interface
type omapReapStateReader struct {
	state  *types.OmapReapStateT
	endian binary.ByteOrder
}

// NewOmapReapStateReader creates a new reader for object map reap state
func NewOmapReapStateReader(data []byte, endian binary.ByteOrder) (interfaces.ObjectMapReaperState, error) {
	if len(data) < 28 { // OmrPhase(4) + OmrOk(OmapKeyT - 24 bytes)
		return nil, fmt.Errorf("data too small for omap reap state: %d bytes", len(data))
	}

	state, err := parseOmapReapState(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse omap reap state: %w", err)
	}

	return &omapReapStateReader{
		state:  state,
		endian: endian,
	}, nil
}

// parseOmapReapState parses raw bytes into an OmapReapStateT structure
func parseOmapReapState(data []byte, endian binary.ByteOrder) (*types.OmapReapStateT, error) {
	if len(data) < 28 {
		return nil, fmt.Errorf("insufficient data for omap reap state: need 28 bytes, got %d", len(data))
	}

	state := &types.OmapReapStateT{}
	state.OmrPhase = endian.Uint32(data[0:4])

	// Parse OmapKeyT (16 bytes): OkOid(8) + OkXid(8)
	state.OmrOk.OkOid = types.OidT(endian.Uint64(data[4:12]))
	state.OmrOk.OkXid = types.XidT(endian.Uint64(data[12:20]))
	// Padding/reserved at data[20:28]

	return state, nil
}

// ReapingPhase returns the current reaping phase
func (orsr *omapReapStateReader) ReapingPhase() uint32 {
	return orsr.state.OmrPhase
}

// LastProcessedKey returns the key of the most recently freed entry
func (orsr *omapReapStateReader) LastProcessedKey() types.OmapKeyT {
	return orsr.state.OmrOk
}

// PhaseDescription returns a human-readable description of the reaping phase
func (orsr *omapReapStateReader) PhaseDescription() string {
	switch orsr.state.OmrPhase {
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
