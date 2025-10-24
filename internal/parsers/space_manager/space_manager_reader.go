package spacemanager

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// SpaceManagerReader implements comprehensive space manager parsing
// The space manager allocates and frees blocks where objects and file data can be stored
// There's exactly one instance of this structure in a container
type SpaceManagerReader struct {
	spaceman *types.SpacemanPhysT
	data     []byte
	endian   binary.ByteOrder
}

// NewSpaceManagerReader creates a new space manager reader
func NewSpaceManagerReader(data []byte, endian binary.ByteOrder) (*SpaceManagerReader, error) {
	// Minimum required size for spaceman_phys_t structure
	// obj_phys_t (32) + basic fields (16) + 2 devices (112) + flags/ip (48) + free queues (120) + remaining (136+) = 2664+ bytes
	minSize := 2700
	if len(data) < minSize {
		return nil, fmt.Errorf("data too small for space manager: %d bytes, need at least %d", len(data), minSize)
	}

	spaceman, err := parseSpaceManager(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse space manager: %w", err)
	}

	// Validate the object type
	objectType := spaceman.SmO.OType & types.ObjectTypeMask
	if objectType != types.ObjectTypeSpaceman {
		return nil, fmt.Errorf("invalid space manager object type: 0x%x", objectType)
	}

	return &SpaceManagerReader{
		spaceman: spaceman,
		data:     data,
		endian:   endian,
	}, nil
}

// parseSpaceManager parses raw bytes into a SpacemanPhysT structure
func parseSpaceManager(data []byte, endian binary.ByteOrder) (*types.SpacemanPhysT, error) {
	if len(data) < 512 {
		return nil, fmt.Errorf("insufficient data for space manager")
	}

	sm := &types.SpacemanPhysT{}
	offset := 0

	// Parse object header (obj_phys_t): 32 bytes
	copy(sm.SmO.OChecksum[:], data[offset:offset+8])
	offset += 8
	sm.SmO.OOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	sm.SmO.OXid = types.XidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	sm.SmO.OType = endian.Uint32(data[offset : offset+4])
	offset += 4
	sm.SmO.OSubtype = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse space manager specific fields (16 bytes)
	sm.SmBlockSize = endian.Uint32(data[offset : offset+4])
	offset += 4
	sm.SmBlocksPerChunk = endian.Uint32(data[offset : offset+4])
	offset += 4
	sm.SmChunksPerCib = endian.Uint32(data[offset : offset+4])
	offset += 4
	sm.SmCibsPerCab = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse devices array: spaceman_device_t sm_dev[SD_COUNT] where SD_COUNT = 2
	// Each spaceman_device_t is 56 bytes (including SmCabOid)
	for i := 0; i < int(types.SdCount); i++ {
		sm.SmDev[i].SmBlockCount = endian.Uint64(data[offset : offset+8])
		offset += 8
		sm.SmDev[i].SmChunkCount = endian.Uint64(data[offset : offset+8])
		offset += 8
		sm.SmDev[i].SmCibCount = endian.Uint32(data[offset : offset+4])
		offset += 4
		sm.SmDev[i].SmCabCount = endian.Uint32(data[offset : offset+4])
		offset += 4
		sm.SmDev[i].SmFreeCount = endian.Uint64(data[offset : offset+8])
		offset += 8
		sm.SmDev[i].SmAddrOffset = endian.Uint32(data[offset : offset+4])
		offset += 4
		sm.SmDev[i].SmReserved = endian.Uint32(data[offset : offset+4])
		offset += 4
		sm.SmDev[i].SmReserved2 = endian.Uint64(data[offset : offset+8])
		offset += 8
		sm.SmDev[i].SmCabOid = types.OidT(endian.Uint64(data[offset : offset+8]))
		offset += 8
	}

	// Parse space manager flags and internal pool configuration
	sm.SmFlags = endian.Uint32(data[offset : offset+4])
	offset += 4
	sm.SmIpBmTxMultiplier = endian.Uint32(data[offset : offset+4])
	offset += 4
	sm.SmIpBlockCount = endian.Uint64(data[offset : offset+8])
	offset += 8
	sm.SmIpBmSizeInBlocks = endian.Uint32(data[offset : offset+4])
	offset += 4
	sm.SmIpBmBlockCount = endian.Uint32(data[offset : offset+4])
	offset += 4
	sm.SmIpBmBase = types.Paddr(endian.Uint64(data[offset : offset+8]))
	offset += 8
	sm.SmIpBase = types.Paddr(endian.Uint64(data[offset : offset+8]))
	offset += 8
	sm.SmFsReserveBlockCount = endian.Uint64(data[offset : offset+8])
	offset += 8
	sm.SmFsReserveAllocCount = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse free queues array: spaceman_free_queue_t sm_fq[SFQ_COUNT] where SFQ_COUNT = 3
	// Each spaceman_free_queue_t is 40 bytes
	for i := 0; i < int(types.SfqCount); i++ {
		sm.SmFq[i].SfqCount = endian.Uint64(data[offset : offset+8])
		offset += 8
		sm.SmFq[i].SfqTreeOid = types.OidT(endian.Uint64(data[offset : offset+8]))
		offset += 8
		sm.SmFq[i].SfqOldestXid = types.XidT(endian.Uint64(data[offset : offset+8]))
		offset += 8
		sm.SmFq[i].SfqTreeNodeLimit = endian.Uint16(data[offset : offset+2])
		offset += 2
		sm.SmFq[i].SfqPad16 = endian.Uint16(data[offset : offset+2])
		offset += 2
		sm.SmFq[i].SfqPad32 = endian.Uint32(data[offset : offset+4])
		offset += 4
		sm.SmFq[i].SfqReserved = endian.Uint64(data[offset : offset+8])
		offset += 8
	}

	// Parse remaining internal pool bitmap management fields
	sm.SmIpBmFreeHead = endian.Uint16(data[offset : offset+2])
	offset += 2
	sm.SmIpBmFreeTail = endian.Uint16(data[offset : offset+2])
	offset += 2
	sm.SmIpBmXidOffset = endian.Uint32(data[offset : offset+4])
	offset += 4
	sm.SmIpBitmapOffset = endian.Uint32(data[offset : offset+4])
	offset += 4
	sm.SmIpBmFreeNextOffset = endian.Uint32(data[offset : offset+4])
	offset += 4
	sm.SmVersion = endian.Uint32(data[offset : offset+4])
	offset += 4
	sm.SmStructSize = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse data zone information
	err := parseDataZoneInfo(&sm.SmDatazone, data[offset:], endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse data zone info: %w", err)
	}

	return sm, nil
}

// parseDataZoneInfo parses data zone information from the data
func parseDataZoneInfo(datazone *types.SpacemanDataZoneInfoPhysT, data []byte, endian binary.ByteOrder) error {
	offset := 0

	// Parse allocation zones: sdz_allocation_zones[SD_COUNT][SM_DATAZONE_ALLOCZONE_COUNT]
	for device := 0; device < int(types.SdCount); device++ {
		for zone := 0; zone < types.SmDataZoneAllocZoneCount; zone++ {
			if offset+136 > len(data) {
				return fmt.Errorf("insufficient data for allocation zone %d:%d", device, zone)
			}

			allocZone := &datazone.SdzAllocationZones[device][zone]

			// Parse current boundaries
			allocZone.SazCurrentBoundaries.SazZoneStart = endian.Uint64(data[offset : offset+8])
			offset += 8
			allocZone.SazCurrentBoundaries.SazZoneEnd = endian.Uint64(data[offset : offset+8])
			offset += 8

			// Parse previous boundaries array [7]
			for i := 0; i < types.SmAllocZoneNumPreviousBoundaries; i++ {
				allocZone.SazPreviousBoundaries[i].SazZoneStart = endian.Uint64(data[offset : offset+8])
				offset += 8
				allocZone.SazPreviousBoundaries[i].SazZoneEnd = endian.Uint64(data[offset : offset+8])
				offset += 8
			}

			// Parse remaining allocation zone fields
			allocZone.SazZoneId = endian.Uint16(data[offset : offset+2])
			offset += 2
			allocZone.SazPreviousBoundaryIndex = endian.Uint16(data[offset : offset+2])
			offset += 2
			allocZone.SazReserved = endian.Uint32(data[offset : offset+4])
			offset += 4
		}
	}

	return nil
}

// Superblock returns the space manager's physical structure
func (smr *SpaceManagerReader) Superblock() *types.SpacemanPhysT {
	return smr.spaceman
}

// BlockSize returns the block size used by the space manager
func (smr *SpaceManagerReader) BlockSize() uint32 {
	return smr.spaceman.SmBlockSize
}

// Version returns the space manager version
func (smr *SpaceManagerReader) Version() uint32 {
	return smr.spaceman.SmVersion
}

// BlocksPerChunk returns the number of blocks per chunk
func (smr *SpaceManagerReader) BlocksPerChunk() uint32 {
	return smr.spaceman.SmBlocksPerChunk
}

// ChunksPerCIB returns the number of chunks per chunk-info block
func (smr *SpaceManagerReader) ChunksPerCIB() uint32 {
	return smr.spaceman.SmChunksPerCib
}

// CIBsPerCAB returns the number of chunk-info blocks per chunk-info address block
func (smr *SpaceManagerReader) CIBsPerCAB() uint32 {
	return smr.spaceman.SmCibsPerCab
}

// Flags returns the space manager flags
func (smr *SpaceManagerReader) Flags() uint32 {
	return smr.spaceman.SmFlags
}

// IsVersioned returns true if the space manager uses versioning
func (smr *SpaceManagerReader) IsVersioned() bool {
	return smr.spaceman.SmFlags&types.SmFlagVersioned != 0
}

// GetMainDevice returns a SpacemanDeviceReader for the main device
func (smr *SpaceManagerReader) GetMainDevice() (*SpacemanDeviceReader, error) {
	data := make([]byte, 56)
	return smr.getDeviceReader(types.SdMain, data)
}

// GetTier2Device returns a SpacemanDeviceReader for the tier2 device
func (smr *SpaceManagerReader) GetTier2Device() (*SpacemanDeviceReader, error) {
	data := make([]byte, 56)
	return smr.getDeviceReader(types.SdTier2, data)
}

// getDeviceReader helper function
func (smr *SpaceManagerReader) getDeviceReader(device types.SmdevT, data []byte) (*SpacemanDeviceReader, error) {
	dev := &smr.spaceman.SmDev[device]
	offset := 0

	smr.endian.PutUint64(data[offset:offset+8], dev.SmBlockCount)
	offset += 8
	smr.endian.PutUint64(data[offset:offset+8], dev.SmChunkCount)
	offset += 8
	smr.endian.PutUint32(data[offset:offset+4], dev.SmCibCount)
	offset += 4
	smr.endian.PutUint32(data[offset:offset+4], dev.SmCabCount)
	offset += 4
	smr.endian.PutUint64(data[offset:offset+8], dev.SmFreeCount)
	offset += 8
	smr.endian.PutUint32(data[offset:offset+4], dev.SmAddrOffset)
	offset += 4
	smr.endian.PutUint32(data[offset:offset+4], dev.SmReserved)
	offset += 4
	smr.endian.PutUint64(data[offset:offset+8], dev.SmReserved2)
	offset += 8
	smr.endian.PutUint64(data[offset:offset+8], uint64(dev.SmCabOid))
	offset += 8

	return NewSpacemanDeviceReader(data, smr.endian)
}

// HasFusionDevice returns true if both main and tier2 devices are present
func (smr *SpaceManagerReader) HasFusionDevice() bool {
	return smr.spaceman.SmDev[types.SdTier2].SmBlockCount > 0
}

// GetFreeQueue returns a SpacemanFreeQueueReader for a specific queue
func (smr *SpaceManagerReader) GetFreeQueue(queueType types.SfqT) (*SpacemanFreeQueueReader, error) {
	if int(queueType) >= len(smr.spaceman.SmFq) {
		return nil, fmt.Errorf("invalid queue type: %d", queueType)
	}

	data := make([]byte, 40)
	queue := &smr.spaceman.SmFq[queueType]
	offset := 0

	smr.endian.PutUint64(data[offset:offset+8], queue.SfqCount)
	offset += 8
	smr.endian.PutUint64(data[offset:offset+8], uint64(queue.SfqTreeOid))
	offset += 8
	smr.endian.PutUint64(data[offset:offset+8], uint64(queue.SfqOldestXid))
	offset += 8
	smr.endian.PutUint16(data[offset:offset+2], queue.SfqTreeNodeLimit)
	offset += 2
	smr.endian.PutUint16(data[offset:offset+2], queue.SfqPad16)
	offset += 2
	smr.endian.PutUint32(data[offset:offset+4], queue.SfqPad32)
	offset += 4
	smr.endian.PutUint64(data[offset:offset+8], queue.SfqReserved)
	offset += 8

	return NewSpacemanFreeQueueReader(data, smr.endian)
}

// GetDataZoneInfo returns a DataZoneInfoReader for accessing allocation zones
func (smr *SpaceManagerReader) GetDataZoneInfo() (*DataZoneInfoReader, error) {
	// 2176 bytes for all allocation zones
	data := make([]byte, 2176)
	offset := 0

	for device := 0; device < int(types.SdCount); device++ {
		for zone := 0; zone < types.SmDataZoneAllocZoneCount; zone++ {
			allocZone := &smr.spaceman.SmDatazone.SdzAllocationZones[device][zone]

			smr.endian.PutUint64(data[offset:offset+8], allocZone.SazCurrentBoundaries.SazZoneStart)
			offset += 8
			smr.endian.PutUint64(data[offset:offset+8], allocZone.SazCurrentBoundaries.SazZoneEnd)
			offset += 8

			for i := 0; i < types.SmAllocZoneNumPreviousBoundaries; i++ {
				smr.endian.PutUint64(data[offset:offset+8], allocZone.SazPreviousBoundaries[i].SazZoneStart)
				offset += 8
				smr.endian.PutUint64(data[offset:offset+8], allocZone.SazPreviousBoundaries[i].SazZoneEnd)
				offset += 8
			}

			smr.endian.PutUint16(data[offset:offset+2], allocZone.SazZoneId)
			offset += 2
			smr.endian.PutUint16(data[offset:offset+2], allocZone.SazPreviousBoundaryIndex)
			offset += 2
			smr.endian.PutUint32(data[offset:offset+4], allocZone.SazReserved)
			offset += 4
		}
	}

	return NewDataZoneInfoReader(data, smr.endian)
}

// GetStructSize returns the size of the space manager structure
func (smr *SpaceManagerReader) GetStructSize() uint32 {
	return smr.spaceman.SmStructSize
}
