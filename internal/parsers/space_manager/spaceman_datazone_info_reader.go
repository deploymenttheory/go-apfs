package spacemanager

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// DataZoneInfoReader provides parsing for data zone information
// Manages allocation zones across multiple devices for efficient space allocation
type DataZoneInfoReader struct {
	datazoneInfo *types.SpacemanDataZoneInfoPhysT
	data         []byte
	endian       binary.ByteOrder
}

// NewDataZoneInfoReader creates a new data zone info reader
// Data zone info is 2176 bytes: 2 devices × 8 zones × 136 bytes per zone
func NewDataZoneInfoReader(data []byte, endian binary.ByteOrder) (*DataZoneInfoReader, error) {
	// 2 devices × 8 zones × 136 bytes = 2176 bytes
	requiredSize := int(types.SdCount) * types.SmDataZoneAllocZoneCount * 136
	if len(data) < requiredSize {
		return nil, fmt.Errorf("data too small for data zone info: %d bytes, need at least %d", len(data), requiredSize)
	}

	datazoneInfo, err := parseDataZoneInfoStruct(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse data zone info: %w", err)
	}

	return &DataZoneInfoReader{
		datazoneInfo: datazoneInfo,
		data:         data,
		endian:       endian,
	}, nil
}

// parseDataZoneInfoStruct parses raw bytes into SpacemanDataZoneInfoPhysT
func parseDataZoneInfoStruct(data []byte, endian binary.ByteOrder) (*types.SpacemanDataZoneInfoPhysT, error) {
	requiredSize := int(types.SdCount) * types.SmDataZoneAllocZoneCount * 136
	if len(data) < requiredSize {
		return nil, fmt.Errorf("insufficient data for data zone info")
	}

	sdzi := &types.SpacemanDataZoneInfoPhysT{}
	offset := 0

	// Parse allocation zones: sdz_allocation_zones[SD_COUNT][SM_DATAZONE_ALLOCZONE_COUNT]
	// SD_COUNT = 2 devices, SM_DATAZONE_ALLOCZONE_COUNT = 8 zones per device
	for device := 0; device < int(types.SdCount); device++ {
		for zone := 0; zone < types.SmDataZoneAllocZoneCount; zone++ {
			allocZone := &sdzi.SdzAllocationZones[device][zone]

			// Parse current boundaries (16 bytes)
			allocZone.SazCurrentBoundaries.SazZoneStart = endian.Uint64(data[offset : offset+8])
			offset += 8
			allocZone.SazCurrentBoundaries.SazZoneEnd = endian.Uint64(data[offset : offset+8])
			offset += 8

			// Parse previous boundaries array [7] (112 bytes: 7 × 16 bytes)
			for i := 0; i < types.SmAllocZoneNumPreviousBoundaries; i++ {
				allocZone.SazPreviousBoundaries[i].SazZoneStart = endian.Uint64(data[offset : offset+8])
				offset += 8
				allocZone.SazPreviousBoundaries[i].SazZoneEnd = endian.Uint64(data[offset : offset+8])
				offset += 8
			}

			// Parse zone ID (2 bytes)
			allocZone.SazZoneId = endian.Uint16(data[offset : offset+2])
			offset += 2

			// Parse previous boundary index (2 bytes)
			allocZone.SazPreviousBoundaryIndex = endian.Uint16(data[offset : offset+2])
			offset += 2

			// Parse reserved (4 bytes)
			allocZone.SazReserved = endian.Uint32(data[offset : offset+4])
			offset += 4
		}
	}

	return sdzi, nil
}

// GetDataZoneInfo returns the data zone info structure
func (dzi *DataZoneInfoReader) GetDataZoneInfo() *types.SpacemanDataZoneInfoPhysT {
	return dzi.datazoneInfo
}

// GetAllocationZoneReader returns an AllocationZoneInfoReader for a specific device and zone
func (dzi *DataZoneInfoReader) GetAllocationZoneReader(device types.SmdevT, zone int) (*AllocationZoneInfoReader, error) {
	if int(device) >= int(types.SdCount) || zone < 0 || zone >= types.SmDataZoneAllocZoneCount {
		return nil, fmt.Errorf("invalid device (%d) or zone (%d)", device, zone)
	}

	// Create 136-byte buffer for the allocation zone
	data := make([]byte, 136)
	allocZone := &dzi.datazoneInfo.SdzAllocationZones[device][zone]

	offset := 0

	// Current boundaries
	dzi.endian.PutUint64(data[offset:offset+8], allocZone.SazCurrentBoundaries.SazZoneStart)
	offset += 8
	dzi.endian.PutUint64(data[offset:offset+8], allocZone.SazCurrentBoundaries.SazZoneEnd)
	offset += 8

	// Previous boundaries
	for i := 0; i < types.SmAllocZoneNumPreviousBoundaries; i++ {
		dzi.endian.PutUint64(data[offset:offset+8], allocZone.SazPreviousBoundaries[i].SazZoneStart)
		offset += 8
		dzi.endian.PutUint64(data[offset:offset+8], allocZone.SazPreviousBoundaries[i].SazZoneEnd)
		offset += 8
	}

	// Zone ID
	dzi.endian.PutUint16(data[offset:offset+2], allocZone.SazZoneId)
	offset += 2

	// Previous boundary index
	dzi.endian.PutUint16(data[offset:offset+2], allocZone.SazPreviousBoundaryIndex)
	offset += 2

	// Reserved
	dzi.endian.PutUint32(data[offset:offset+4], allocZone.SazReserved)
	offset += 4

	return NewAllocationZoneInfoReader(data, dzi.endian)
}

// GetAllocationZone returns the raw allocation zone structure
func (dzi *DataZoneInfoReader) GetAllocationZone(device types.SmdevT, zone int) (*types.SpacemanAllocationZoneInfoPhysT, error) {
	if int(device) >= int(types.SdCount) || zone < 0 || zone >= types.SmDataZoneAllocZoneCount {
		return nil, fmt.Errorf("invalid device (%d) or zone (%d)", device, zone)
	}
	return &dzi.datazoneInfo.SdzAllocationZones[device][zone], nil
}

// GetDeviceZones returns all allocation zones for a specific device
func (dzi *DataZoneInfoReader) GetDeviceZones(device types.SmdevT) ([types.SmDataZoneAllocZoneCount]*types.SpacemanAllocationZoneInfoPhysT, error) {
	if int(device) >= int(types.SdCount) {
		return [types.SmDataZoneAllocZoneCount]*types.SpacemanAllocationZoneInfoPhysT{}, fmt.Errorf("invalid device %d", device)
	}

	var zones [types.SmDataZoneAllocZoneCount]*types.SpacemanAllocationZoneInfoPhysT
	for i := 0; i < types.SmDataZoneAllocZoneCount; i++ {
		zones[i] = &dzi.datazoneInfo.SdzAllocationZones[device][i]
	}
	return zones, nil
}

// GetMainDeviceZones returns all allocation zones for the main device
func (dzi *DataZoneInfoReader) GetMainDeviceZones() [types.SmDataZoneAllocZoneCount]*types.SpacemanAllocationZoneInfoPhysT {
	zones, _ := dzi.GetDeviceZones(types.SdMain)
	return zones
}

// GetTier2DeviceZones returns all allocation zones for the tier2 device
func (dzi *DataZoneInfoReader) GetTier2DeviceZones() [types.SmDataZoneAllocZoneCount]*types.SpacemanAllocationZoneInfoPhysT {
	zones, _ := dzi.GetDeviceZones(types.SdTier2)
	return zones
}

// CalculateTotalAllocatedBlocks calculates total allocated blocks across a device
func (dzi *DataZoneInfoReader) CalculateTotalAllocatedBlocks(device types.SmdevT) (uint64, error) {
	zones, err := dzi.GetDeviceZones(device)
	if err != nil {
		return 0, err
	}

	var total uint64
	for _, zone := range zones {
		if zone != nil && zone.SazCurrentBoundaries.SazZoneEnd > zone.SazCurrentBoundaries.SazZoneStart {
			total += zone.SazCurrentBoundaries.SazZoneEnd - zone.SazCurrentBoundaries.SazZoneStart
		}
	}
	return total, nil
}

// GetZoneCount returns the number of allocation zones
func (dzi *DataZoneInfoReader) GetZoneCount() int {
	return types.SmDataZoneAllocZoneCount
}

// GetDeviceCount returns the number of devices
func (dzi *DataZoneInfoReader) GetDeviceCount() int {
	return int(types.SdCount)
}
