package spacemanager

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// AllocationZoneInfoReader provides parsing for allocation zone information
// Contains current and historical zone boundaries for space allocation tracking
type AllocationZoneInfoReader struct {
	zoneInfo *types.SpacemanAllocationZoneInfoPhysT
	data     []byte
	endian   binary.ByteOrder
}

// NewAllocationZoneInfoReader creates a new allocation zone info reader
// Zone info is 136 bytes: current_boundaries (16) + previous_boundaries[7] (112) + zone_id (2) + index (2) + reserved (4)
func NewAllocationZoneInfoReader(data []byte, endian binary.ByteOrder) (*AllocationZoneInfoReader, error) {
	if len(data) < 136 {
		return nil, fmt.Errorf("data too small for allocation zone info: %d bytes, need at least 136", len(data))
	}

	zoneInfo, err := parseAllocationZoneInfo(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse allocation zone info: %w", err)
	}

	return &AllocationZoneInfoReader{
		zoneInfo: zoneInfo,
		data:     data,
		endian:   endian,
	}, nil
}

// parseAllocationZoneInfo parses raw bytes into SpacemanAllocationZoneInfoPhysT
func parseAllocationZoneInfo(data []byte, endian binary.ByteOrder) (*types.SpacemanAllocationZoneInfoPhysT, error) {
	if len(data) < 136 {
		return nil, fmt.Errorf("insufficient data for allocation zone info")
	}

	sazi := &types.SpacemanAllocationZoneInfoPhysT{}
	offset := 0

	// Parse current boundaries (16 bytes)
	sazi.SazCurrentBoundaries.SazZoneStart = endian.Uint64(data[offset : offset+8])
	offset += 8
	sazi.SazCurrentBoundaries.SazZoneEnd = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse previous boundaries array [7] (112 bytes: 7 × 16 bytes)
	for i := 0; i < types.SmAllocZoneNumPreviousBoundaries; i++ {
		sazi.SazPreviousBoundaries[i].SazZoneStart = endian.Uint64(data[offset : offset+8])
		offset += 8
		sazi.SazPreviousBoundaries[i].SazZoneEnd = endian.Uint64(data[offset : offset+8])
		offset += 8
	}

	// Parse zone ID (2 bytes)
	sazi.SazZoneId = endian.Uint16(data[offset : offset+2])
	offset += 2

	// Parse previous boundary index (2 bytes)
	sazi.SazPreviousBoundaryIndex = endian.Uint16(data[offset : offset+2])
	offset += 2

	// Parse reserved (4 bytes)
	sazi.SazReserved = endian.Uint32(data[offset : offset+4])
	offset += 4

	return sazi, nil
}

// GetZoneInfo returns the zone info structure
func (azi *AllocationZoneInfoReader) GetZoneInfo() *types.SpacemanAllocationZoneInfoPhysT {
	return azi.zoneInfo
}

// ZoneID returns the unique identifier for this zone
func (azi *AllocationZoneInfoReader) ZoneID() uint16 {
	return azi.zoneInfo.SazZoneId
}

// GetCurrentBoundaries returns the current zone boundaries reader
func (azi *AllocationZoneInfoReader) GetCurrentBoundaries() *AllocationZoneBoundariesReader {
	data := make([]byte, 16)
	azi.endian.PutUint64(data[0:8], azi.zoneInfo.SazCurrentBoundaries.SazZoneStart)
	azi.endian.PutUint64(data[8:16], azi.zoneInfo.SazCurrentBoundaries.SazZoneEnd)
	reader, _ := NewAllocationZoneBoundariesReader(data, azi.endian)
	return reader
}

// GetCurrentStart returns the starting block address of the current zone
func (azi *AllocationZoneInfoReader) GetCurrentStart() uint64 {
	return azi.zoneInfo.SazCurrentBoundaries.SazZoneStart
}

// GetCurrentEnd returns the ending block address of the current zone
func (azi *AllocationZoneInfoReader) GetCurrentEnd() uint64 {
	return azi.zoneInfo.SazCurrentBoundaries.SazZoneEnd
}

// GetCurrentSize returns the size of the current zone
func (azi *AllocationZoneInfoReader) GetCurrentSize() uint64 {
	if azi.zoneInfo.SazCurrentBoundaries.SazZoneEnd >= azi.zoneInfo.SazCurrentBoundaries.SazZoneStart {
		return azi.zoneInfo.SazCurrentBoundaries.SazZoneEnd - azi.zoneInfo.SazCurrentBoundaries.SazZoneStart
	}
	return 0
}

// PreviousBoundaryIndex returns the index into the previous boundaries array
func (azi *AllocationZoneInfoReader) PreviousBoundaryIndex() uint16 {
	return azi.zoneInfo.SazPreviousBoundaryIndex
}

// GetPreviousBoundary returns a previous boundary by index (0-6)
func (azi *AllocationZoneInfoReader) GetPreviousBoundary(index int) (*AllocationZoneBoundariesReader, error) {
	if index < 0 || index >= types.SmAllocZoneNumPreviousBoundaries {
		return nil, fmt.Errorf("previous boundary index %d out of range (0-%d)", index, types.SmAllocZoneNumPreviousBoundaries-1)
	}

	data := make([]byte, 16)
	azi.endian.PutUint64(data[0:8], azi.zoneInfo.SazPreviousBoundaries[index].SazZoneStart)
	azi.endian.PutUint64(data[8:16], azi.zoneInfo.SazPreviousBoundaries[index].SazZoneEnd)
	reader, _ := NewAllocationZoneBoundariesReader(data, azi.endian)
	return reader, nil
}

// GetPreviousBoundaryRaw returns raw boundary data without reader wrapper
func (azi *AllocationZoneInfoReader) GetPreviousBoundaryRaw(index int) (uint64, uint64, error) {
	if index < 0 || index >= types.SmAllocZoneNumPreviousBoundaries {
		return 0, 0, fmt.Errorf("previous boundary index %d out of range", index)
	}

	return azi.zoneInfo.SazPreviousBoundaries[index].SazZoneStart,
		azi.zoneInfo.SazPreviousBoundaries[index].SazZoneEnd, nil
}

// GetAllPreviousBoundaries returns all previous boundaries
func (azi *AllocationZoneInfoReader) GetAllPreviousBoundaries() []types.SpacemanAllocationZoneBoundariesT {
	return azi.zoneInfo.SazPreviousBoundaries[:]
}

// IsCurrentZoneValid returns true if the current zone boundaries are valid
func (azi *AllocationZoneInfoReader) IsCurrentZoneValid() bool {
	return azi.zoneInfo.SazCurrentBoundaries.SazZoneStart < azi.zoneInfo.SazCurrentBoundaries.SazZoneEnd
}

// CurrentZoneIsEmpty returns true if the current zone has an invalid end boundary
func (azi *AllocationZoneInfoReader) CurrentZoneIsEmpty() bool {
	return azi.zoneInfo.SazCurrentBoundaries.SazZoneEnd == types.SmAllocZoneInvalidEndBoundary
}

// HasPreviousBoundaryHistory returns true if there are any valid previous boundaries
func (azi *AllocationZoneInfoReader) HasPreviousBoundaryHistory() bool {
	for _, boundary := range azi.zoneInfo.SazPreviousBoundaries {
		if boundary.SazZoneEnd != types.SmAllocZoneInvalidEndBoundary {
			return true
		}
	}
	return false
}
