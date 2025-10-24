package spacemanager

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// AllocationZoneBoundariesReader provides parsing for allocation zone boundaries
// Allocation zone boundaries define the start and end addresses of an allocation zone
type AllocationZoneBoundariesReader struct {
	boundaries *types.SpacemanAllocationZoneBoundariesT
	data       []byte
	endian     binary.ByteOrder
}

// NewAllocationZoneBoundariesReader creates a new allocation zone boundaries reader
// Allocation zone boundaries are 16 bytes: zone_start (uint64) + zone_end (uint64)
func NewAllocationZoneBoundariesReader(data []byte, endian binary.ByteOrder) (*AllocationZoneBoundariesReader, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("data too small for allocation zone boundaries: %d bytes, need at least 16", len(data))
	}

	boundaries, err := parseAllocationZoneBoundaries(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse allocation zone boundaries: %w", err)
	}

	return &AllocationZoneBoundariesReader{
		boundaries: boundaries,
		data:       data,
		endian:     endian,
	}, nil
}

// parseAllocationZoneBoundaries parses raw bytes into SpacemanAllocationZoneBoundariesT
func parseAllocationZoneBoundaries(data []byte, endian binary.ByteOrder) (*types.SpacemanAllocationZoneBoundariesT, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("insufficient data for allocation zone boundaries")
	}

	saz := &types.SpacemanAllocationZoneBoundariesT{}
	offset := 0

	// Parse zone start (uint64)
	saz.SazZoneStart = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse zone end (uint64)
	saz.SazZoneEnd = endian.Uint64(data[offset : offset+8])
	offset += 8

	return saz, nil
}

// GetBoundaries returns the boundaries structure
func (azb *AllocationZoneBoundariesReader) GetBoundaries() *types.SpacemanAllocationZoneBoundariesT {
	return azb.boundaries
}

// ZoneStart returns the starting block address of the zone
func (azb *AllocationZoneBoundariesReader) ZoneStart() uint64 {
	return azb.boundaries.SazZoneStart
}

// ZoneEnd returns the ending block address of the zone
func (azb *AllocationZoneBoundariesReader) ZoneEnd() uint64 {
	return azb.boundaries.SazZoneEnd
}

// ZoneSize returns the size of the zone in blocks
func (azb *AllocationZoneBoundariesReader) ZoneSize() uint64 {
	if azb.boundaries.SazZoneEnd >= azb.boundaries.SazZoneStart {
		return azb.boundaries.SazZoneEnd - azb.boundaries.SazZoneStart
	}
	return 0
}

// IsValid returns true if the boundaries are valid (start < end)
func (azb *AllocationZoneBoundariesReader) IsValid() bool {
	return azb.boundaries.SazZoneStart < azb.boundaries.SazZoneEnd
}

// IsEmpty returns true if the zone has no size (invalid end boundary)
func (azb *AllocationZoneBoundariesReader) IsEmpty() bool {
	return azb.boundaries.SazZoneEnd == types.SmAllocZoneInvalidEndBoundary
}

// ContainsAddress returns true if the given address is within the zone boundaries
func (azb *AllocationZoneBoundariesReader) ContainsAddress(addr uint64) bool {
	return addr >= azb.boundaries.SazZoneStart && addr < azb.boundaries.SazZoneEnd
}
