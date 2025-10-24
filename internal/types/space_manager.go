package types

// Space Manager (pages 159-163)
// The space manager allocates and frees blocks where objects and file data can be stored.
// There's exactly one instance of this structure in a container.

// ChunkInfoT manages information about a chunk of storage.
// Reference: page 159
type ChunkInfoT struct {
	// The transaction identifier for this chunk information.
	CiXid uint64

	// The address of the chunk.
	CiAddr uint64

	// The number of blocks in the chunk.
	CiBlockCount uint32

	// The number of free blocks in the chunk.
	CiFreeCount uint32

	// The address of the bitmap for this chunk.
	CiBitmapAddr Paddr
}

// ChunkInfoBlockT is a block that contains an array of chunk-info structures.
// Reference: page 159
type ChunkInfoBlockT struct {
	// The object's header.
	CibO ObjPhysT

	// The index of this chunk info block.
	CibIndex uint32

	// The number of chunk info entries in this block.
	CibChunkInfoCount uint32

	// Array of chunk info entries.
	CibChunkInfo []ChunkInfoT
}

// CibAddrBlockT is a block that contains an array of chunk-info block addresses.
// Reference: page 159
type CibAddrBlockT struct {
	// The object's header.
	CabO ObjPhysT

	// The index of this chunk-info address block.
	CabIndex uint32

	// The number of chunk-info blocks referenced by this address block.
	CabCibCount uint32

	// Array of addresses of chunk-info blocks.
	CabCibAddr []Paddr
}

// SpacemanFreeQueueEntryT represents an entry in the space manager's free queue.
// Reference: page 159
type SpacemanFreeQueueEntryT struct {
	// The key for this entry.
	SfqeKey SpacemanFreeQueueKeyT

	// The value for this entry.
	SfqeCount SpacemanFreeQueueValT
}

// SpacemanFreeQueueValT is the count of free blocks.
// Reference: page 160
type SpacemanFreeQueueValT uint64

// SpacemanFreeQueueKeyT is the key for a space manager free queue entry.
// Reference: page 160
type SpacemanFreeQueueKeyT struct {
	// The transaction identifier.
	SfqkXid XidT

	// The physical address.
	SfqkPaddr Paddr
}

// SpacemanFreeQueueT represents a queue of free space in the space manager.
// Reference: page 160
type SpacemanFreeQueueT struct {
	// The count of entries in this queue.
	SfqCount uint64

	// The object identifier of the B-tree for this queue.
	SfqTreeOid OidT

	// The oldest transaction identifier in this queue.
	SfqOldestXid XidT

	// The limit on the number of nodes in the tree.
	SfqTreeNodeLimit uint16

	// Padding.
	SfqPad16 uint16

	// Padding.
	SfqPad32 uint32

	// Reserved.
	SfqReserved uint64
}

// SpacemanDeviceT contains information about a device managed by the space manager.
// Reference: page 160
type SpacemanDeviceT struct {
	// The total number of blocks in this device.
	SmBlockCount uint64

	// The number of chunks in this device.
	SmChunkCount uint64

	// The number of chunk-info blocks.
	SmCibCount uint32

	// The number of chunk-info address blocks.
	SmCabCount uint32

	// The number of free blocks in this device.
	SmFreeCount uint64

	// The address offset for this device.
	SmAddrOffset uint32

	// Reserved.
	SmReserved uint32

	// Reserved.
	SmReserved2 uint64

	// The object identifier of the root chunk-info address block.
	SmCabOid OidT
}

// SpacemanAllocationZoneBoundariesT defines the boundaries of an allocation zone.
// Reference: page 161
type SpacemanAllocationZoneBoundariesT struct {
	// The starting block of the zone.
	SazZoneStart uint64

	// The ending block of the zone.
	SazZoneEnd uint64
}

// SmAllocZoneInvalidEndBoundary indicates an invalid end boundary for an allocation zone.
// Reference: page 161
const SmAllocZoneInvalidEndBoundary uint64 = 0

// SmAllocZoneNumPreviousBoundaries is the number of previous boundaries to store for an allocation zone.
// Reference: page 161
const SmAllocZoneNumPreviousBoundaries = 7

// SpacemanAllocationZoneInfoPhysT contains allocation zone information.
// Reference: page 161
type SpacemanAllocationZoneInfoPhysT struct {
	// The current boundaries for this allocation zone.
	SazCurrentBoundaries SpacemanAllocationZoneBoundariesT

	// The previous boundaries for this allocation zone.
	SazPreviousBoundaries [SmAllocZoneNumPreviousBoundaries]SpacemanAllocationZoneBoundariesT

	// The zone ID.
	SazZoneId uint16

	// The index of the previous boundary.
	SazPreviousBoundaryIndex uint16

	// Reserved.
	SazReserved uint32
}

// SmDataZoneAllocZoneCount is the number of allocation zones in a data zone.
// Reference: page 161
const SmDataZoneAllocZoneCount = 8

// SpacemanDataZoneInfoPhysT contains information about a data zone.
// Reference: page 161
type SpacemanDataZoneInfoPhysT struct {
	// Array of allocation zones for each device.
	SdzAllocationZones [SdCount][SmDataZoneAllocZoneCount]SpacemanAllocationZoneInfoPhysT
}

// SpacemanPhysT is the main structure for the space manager.
// Reference: page 161
type SpacemanPhysT struct {
	// The object's header.
	SmO ObjPhysT

	// The block size.
	SmBlockSize uint32

	// The number of blocks per chunk.
	SmBlocksPerChunk uint32

	// The number of chunks per chunk-info block.
	SmChunksPerCib uint32

	// The number of chunk-info blocks per chunk-info address block.
	SmCibsPerCab uint32

	// The devices managed by this space manager.
	SmDev [SdCount]SpacemanDeviceT

	// The space manager flags.
	SmFlags uint32

	// The transaction multiplier for the internal-pool bitmap.
	SmIpBmTxMultiplier uint32

	// The number of blocks in the internal pool.
	SmIpBlockCount uint64

	// The size of the internal-pool bitmap in blocks.
	SmIpBmSizeInBlocks uint32

	// The number of blocks in the internal-pool bitmap.
	SmIpBmBlockCount uint32

	// The base address of the internal-pool bitmap.
	SmIpBmBase Paddr

	// The base address of the internal pool.
	SmIpBase Paddr

	// The number of blocks reserved for the file system.
	SmFsReserveBlockCount uint64

	// The number of reserved blocks that have been allocated.
	SmFsReserveAllocCount uint64

	// The free queues for this space manager.
	SmFq [SfqCount]SpacemanFreeQueueT

	// The head of the free list for the internal-pool bitmap.
	SmIpBmFreeHead uint16

	// The tail of the free list for the internal-pool bitmap.
	SmIpBmFreeTail uint16

	// The transaction identifier offset for the internal-pool bitmap.
	SmIpBmXidOffset uint32

	// The offset to the internal-pool bitmap.
	SmIpBitmapOffset uint32

	// The offset to the next free entry in the internal-pool bitmap.
	SmIpBmFreeNextOffset uint32

	// The version of the space manager.
	SmVersion uint32

	// The size of the space manager structure.
	SmStructSize uint32

	// Information about data zones.
	SmDatazone SpacemanDataZoneInfoPhysT
}

// SfqT indicates which free queue is being referenced.
// Reference: page 162
type SfqT int

const (
	// SfqIp represents the internal pool free queue.
	// Reference: page 162
	SfqIp SfqT = 0

	// SfqMain represents the main device free queue.
	// Reference: page 162
	SfqMain SfqT = 1

	// SfqTier2 represents the tier2 device free queue.
	// Reference: page 162
	SfqTier2 SfqT = 2

	// SfqCount is the number of free queues.
	// Reference: page 162
	SfqCount SfqT = 3
)

// SmdevT indicates which device is being referenced in the space manager.
// Reference: page 162
type SmdevT int

const (
	// SdMain represents the main device.
	// Reference: page 162
	SdMain SmdevT = 0

	// SdTier2 represents the tier2 device.
	// Reference: page 162
	SdTier2 SmdevT = 1

	// SdCount is the number of devices.
	// Reference: page 162
	SdCount SmdevT = 2
)

// CiCountMask is the bit mask used to access the count field in a chunk info block.
// Reference: page 162
const CiCountMask uint32 = 0x000fffff

// CiCountReservedMask is the bit mask for reserved bits in a chunk info block.
// Reference: page 163
const CiCountReservedMask uint32 = 0xfff00000

// Internal-Pool Bitmap constants (page 163)

// SpacemanIpBmTxMultiplier is the transaction multiplier for the internal-pool bitmap.
// Reference: page 163
const SpacemanIpBmTxMultiplier uint32 = 16

// SpacemanIpBmIndexInvalid indicates an invalid index into the internal-pool bitmap.
// Reference: page 163
const SpacemanIpBmIndexInvalid uint16 = 0xffff

// SpacemanIpBmBlockCountMax is the maximum number of blocks in the internal-pool bitmap.
// Reference: page 163
const SpacemanIpBmBlockCountMax uint32 = 0xfffe

// SmFlagVersioned is a flag indicating that the space manager is versioned.
// Reference: page 162
const SmFlagVersioned uint32 = 0x00000001
