// File: pkg/container/spaceman.go

package container

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/apfs/pkg/types"
)

// SpaceManager implements the SpaceManager interface for APFS
type SpaceManager struct {
	device       types.BlockDevice
	spaceman     *types.SpacemanPhys
	spacemenAddr types.PAddr
	bitmapCache  map[types.PAddr][]byte // Cache of bitmap blocks
}

// NewSpaceManager creates a new SpaceManager instance
func NewSpaceManager(device types.BlockDevice, addr types.PAddr) (*SpaceManager, error) {
	// Read the spaceman structure
	data, err := device.ReadBlock(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to read space manager at address %d: %w", addr, err)
	}

	// Parse the spaceman structure
	spaceman, err := parseSpacemanPhys(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse space manager: %w", err)
	}

	return &SpaceManager{
		device:       device,
		spaceman:     spaceman,
		spacemenAddr: addr,
		bitmapCache:  make(map[types.PAddr][]byte),
	}, nil
}

// parseSpacemanPhys parses a raw APFS block containing a spaceman_phys_t structure
// and returns a fully decoded SpacemanPhys instance. This function performs a manual,
// field-by-field binary read to ensure strict alignment with the APFS on-disk format.
// It expects the data to be little-endian and match the expected layout defined in the
// Apple File System Reference.
//
// The returned SpacemanPhys includes decoded metadata such as internal pool configuration,
// free block queues, device allocation zones, and per-zone allocation statistics.
//
// Returns an error if the data is too short or if any field fails to decode.
func parseSpacemanPhys(data []byte) (*types.SpacemanPhys, error) {
	reader := bytes.NewReader(data)
	var sm types.SpacemanPhys

	read := func(v interface{}) error {
		return binary.Read(reader, binary.LittleEndian, v)
	}

	// Object Header (obj_phys_t)
	if err := read(&sm.Header); err != nil {
		return nil, fmt.Errorf("read Header: %w", err)
	}

	if err := read(&sm.BlockSize); err != nil {
		return nil, fmt.Errorf("read BlockSize: %w", err)
	}
	if err := read(&sm.BlocksPerChunk); err != nil {
		return nil, fmt.Errorf("read BlocksPerChunk: %w", err)
	}
	if err := read(&sm.ChunksPerCIB); err != nil {
		return nil, fmt.Errorf("read ChunksPerCIB: %w", err)
	}
	if err := read(&sm.CIBsPerCAB); err != nil {
		return nil, fmt.Errorf("read CIBsPerCAB: %w", err)
	}

	// Devices [2]
	for i := 0; i < 2; i++ {
		if err := read(&sm.Devices[i]); err != nil {
			return nil, fmt.Errorf("read Devices[%d]: %w", i, err)
		}
	}

	if err := read(&sm.Flags); err != nil {
		return nil, fmt.Errorf("read Flags: %w", err)
	}
	if err := read(&sm.IPBmTxMultiplier); err != nil {
		return nil, fmt.Errorf("read IPBmTxMultiplier: %w", err)
	}
	if err := read(&sm.IPBlockCount); err != nil {
		return nil, fmt.Errorf("read IPBlockCount: %w", err)
	}
	if err := read(&sm.IPBmSizeInBlocks); err != nil {
		return nil, fmt.Errorf("read IPBmSizeInBlocks: %w", err)
	}
	if err := read(&sm.IPBmBlockCount); err != nil {
		return nil, fmt.Errorf("read IPBmBlockCount: %w", err)
	}
	if err := read(&sm.IPBmBase); err != nil {
		return nil, fmt.Errorf("read IPBmBase: %w", err)
	}
	if err := read(&sm.IPBase); err != nil {
		return nil, fmt.Errorf("read IPBase: %w", err)
	}
	if err := read(&sm.FSReserveBlockCount); err != nil {
		return nil, fmt.Errorf("read FSReserveBlockCount: %w", err)
	}
	if err := read(&sm.FSReserveAllocCount); err != nil {
		return nil, fmt.Errorf("read FSReserveAllocCount: %w", err)
	}

	// Free Queues [3]
	for i := 0; i < 3; i++ {
		if err := read(&sm.FreeQueues[i]); err != nil {
			return nil, fmt.Errorf("read FreeQueues[%d]: %w", i, err)
		}
	}

	if err := read(&sm.IPBmFreeHead); err != nil {
		return nil, fmt.Errorf("read IPBmFreeHead: %w", err)
	}
	if err := read(&sm.IPBmFreeTail); err != nil {
		return nil, fmt.Errorf("read IPBmFreeTail: %w", err)
	}
	if err := read(&sm.IPBmXidOffset); err != nil {
		return nil, fmt.Errorf("read IPBmXidOffset: %w", err)
	}
	if err := read(&sm.IPBitmapOffset); err != nil {
		return nil, fmt.Errorf("read IPBitmapOffset: %w", err)
	}
	if err := read(&sm.IPBmFreeNextOffset); err != nil {
		return nil, fmt.Errorf("read IPBmFreeNextOffset: %w", err)
	}

	if err := read(&sm.Version); err != nil {
		return nil, fmt.Errorf("read Version: %w", err)
	}
	if err := read(&sm.StructSize); err != nil {
		return nil, fmt.Errorf("read StructSize: %w", err)
	}

	// Datazone Allocation Zones [2][8]
	for dev := 0; dev < 2; dev++ {
		for zone := 0; zone < 8; zone++ {
			if err := read(&sm.Datazone.AllocationZones[dev][zone]); err != nil {
				return nil, fmt.Errorf("read Datazone.AllocationZones[%d][%d]: %w", dev, zone, err)
			}
		}
	}

	return &sm, nil
}

// AllocateBlock allocates a free block from the space manager
func (sm *SpaceManager) AllocateBlock() (types.PAddr, error) {
	// Check internal bitmap first (most efficient)
	if addr, err := sm.allocateFromInternalPool(); err == nil {
		return addr, nil
	}

	// Try main device allocation
	if addr, err := sm.allocateFromMainDevice(); err == nil {
		return addr, nil
	}

	// If this is a Fusion setup, try tier2 device
	if addr, err := sm.allocateFromTier2Device(); err == nil {
		return addr, nil
	}

	return 0, fmt.Errorf("no free blocks available")
}

// allocateFromInternalPool tries to allocate a block from the internal pool bitmap (IPBM).
// The internal pool is used for critical metadata and small structures.
//
// This function scans the bitmap starting from the first byte, looking for the first free bit (0).
// Once found, it marks the bit as allocated (1), writes the updated bitmap back to disk,
// and returns the calculated block address.
//
// Returns an error if the internal pool is unavailable or fully allocated.
func (sm *SpaceManager) allocateFromInternalPool() (types.PAddr, error) {
	// Verify internal pool is enabled
	if sm.spaceman.IPBmSizeInBlocks == 0 {
		return 0, fmt.Errorf("internal pool bitmap is not available")
	}

	bitmapAddr := sm.spaceman.IPBmBase

	// Read the bitmap from the IPBM base address
	bitmap, err := sm.getBitmapBlock(bitmapAddr)
	if err != nil {
		return 0, fmt.Errorf("failed to read internal pool bitmap: %w", err)
	}

	// Traverse each bit to find the first 0 (free block)
	for bitIndex := 0; bitIndex < len(bitmap)*8; bitIndex++ {
		byteIndex := bitIndex / 8
		bitInByte := bitIndex % 8

		// Check if this bit is free
		if bitmap[byteIndex]&(1<<bitInByte) == 0 {
			// Mark bit as allocated
			bitmap[byteIndex] |= 1 << bitInByte

			// Write back the modified bitmap
			if err := sm.device.WriteBlock(bitmapAddr, bitmap); err != nil {
				return 0, fmt.Errorf("failed to update internal pool bitmap: %w", err)
			}

			// Return the block address
			return sm.spaceman.IPBase + types.PAddr(bitIndex), nil
		}
	}

	return 0, fmt.Errorf("no free blocks available in internal pool")
}

// allocateFromMainDevice tries to allocate a block from the main device.
// It iterates through the CIBs (Chunk Info Blocks) referenced by CABs (Chunk Address Blocks),
// looking for a free bit in the chunk bitmap. Once a free block is found, it marks it as
// allocated, updates the bitmap, and returns the calculated physical block address.
// Apple File System Reference Page 159–160
// Returns an error if no free blocks are available or if reading any metadata fails.
func (sm *SpaceManager) allocateFromMainDevice() (types.PAddr, error) {
	mainDevice := sm.spaceman.Devices[0]

	// Loop over all CIBs (chunks), skipping internal pool
	for cibIndex := 0; cibIndex < int(mainDevice.CIBCount); cibIndex++ {
		cibAddr, err := sm.getCIBAddress(cibIndex)
		if err != nil {
			return 0, fmt.Errorf("failed to get CIB address for index %d: %w", cibIndex, err)
		}

		cibData, err := sm.device.ReadBlock(cibAddr)
		if err != nil {
			return 0, fmt.Errorf("failed to read CIB at %d: %w", cibAddr, err)
		}

		blockAddr, ok, err := sm.scanAndAllocateFromCIB(cibData, cibIndex)
		if err != nil {
			return 0, fmt.Errorf("failed to allocate from CIB: %w", err)
		}
		if ok {
			return blockAddr, nil
		}
	}

	return 0, fmt.Errorf("no free blocks available in main device")
}

// getCIBAddress retrieves the physical address of a Chunk Info Block (CIB)
// by reading the appropriate Chunk Address Block (CAB) and indexing into it.
// The CIBs are referenced through CABs, which are found starting at the
// main device's AddrOffset.
//
// Each CAB block contains an array of 8-byte addresses (types.PAddr), and
// typically holds up to 512 entries per 4KB block.
//
// Returns an error if the CIB index is out of bounds or the block read fails.
func (sm *SpaceManager) getCIBAddress(cibIndex int) (types.PAddr, error) {
	main := sm.spaceman.Devices[0]

	if cibIndex < 0 || cibIndex >= int(main.CIBCount) {
		return 0, fmt.Errorf("invalid CIB index: %d", cibIndex)
	}

	entriesPerCAB := int(sm.spaceman.BlockSize) / 8 // 8 bytes per address
	cabIndex := cibIndex / entriesPerCAB
	offsetInCAB := cibIndex % entriesPerCAB

	if cabIndex >= int(main.CABCount) {
		return 0, fmt.Errorf("CAB index %d out of range", cabIndex)
	}

	// CABs start at AddrOffset on disk
	cabAddr := types.PAddr(main.AddrOffset) + types.PAddr(cabIndex)

	data, err := sm.device.ReadBlock(cabAddr)
	if err != nil {
		return 0, fmt.Errorf("failed to read CAB block at %d: %w", cabAddr, err)
	}

	// Each address is 8 bytes (PAddr)
	if len(data) < (offsetInCAB+1)*8 {
		return 0, fmt.Errorf("CAB block too short")
	}

	// Extract the CIB address
	addr := binary.LittleEndian.Uint64(data[offsetInCAB*8 : offsetInCAB*8+8])
	return types.PAddr(addr), nil
}

// scanAndAllocateFromCIB parses a CIB (Chunk Info Block) and attempts to allocate
// a free block from one of its chunk_info entries. Each chunk_info entry contains
// metadata including a bitmap that tracks allocated blocks within a chunk.
//
// This function scans the bitmap, finds a free bit, marks it as allocated,
// writes back the bitmap, and returns the physical address of the allocated block.
//
// Returns (addr, true, nil) on success; (_, false, nil) if no blocks found;
// or (_, false, err) on read/parse failure.
func (sm *SpaceManager) scanAndAllocateFromCIB(cibData []byte, cibIndex int) (types.PAddr, bool, error) {
	const chunkInfoSize = 0x20 // 32 bytes per chunk_info_t

	// Determine how many chunk_info entries exist in this CIB
	numChunks := len(cibData) / chunkInfoSize
	for i := 0; i < numChunks; i++ {
		entry := cibData[i*chunkInfoSize : (i+1)*chunkInfoSize]

		chunkBase := types.PAddr(binary.LittleEndian.Uint64(entry[0:8]))   // base block of chunk
		bitmapAddr := types.PAddr(binary.LittleEndian.Uint64(entry[8:16])) // bitmap physical address
		blockCount := binary.LittleEndian.Uint32(entry[16:20])             // number of blocks in chunk
		freeCount := binary.LittleEndian.Uint32(entry[20:24])              // number of free blocks

		if freeCount == 0 || blockCount == 0 {
			continue
		}

		// Read the bitmap
		bitmap, err := sm.getBitmapBlock(bitmapAddr)
		if err != nil {
			return 0, false, fmt.Errorf("failed to read bitmap for chunk at %d: %w", chunkBase, err)
		}

		for j := 0; j < int(blockCount); j++ {
			byteIndex := j / 8
			bitInByte := j % 8

			if bitmap[byteIndex]&(1<<bitInByte) == 0 {
				// Mark bit as allocated
				bitmap[byteIndex] |= 1 << bitInByte

				// Write back bitmap
				if err := sm.device.WriteBlock(bitmapAddr, bitmap); err != nil {
					return 0, false, fmt.Errorf("failed to write updated bitmap: %w", err)
				}

				// Return actual block address
				return chunkBase + types.PAddr(j), true, nil
			}
		}
	}

	// No free blocks in this CIB
	return 0, false, nil
}

// allocateFromTier2Device tries to allocate a block from the tier2 device's chunk allocation pool.
// This is used in Fusion Drive setups where the second device represents the HDD.
// It behaves identically to main device allocation but references spaceman.Devices[1].
func (sm *SpaceManager) allocateFromTier2Device() (types.PAddr, error) {
	tier2 := sm.spaceman.Devices[1]

	// Loop over all CIBs (chunks)
	for cibIndex := 0; cibIndex < int(tier2.CIBCount); cibIndex++ {
		cibAddr, err := sm.getTier2CIBAddress(cibIndex)
		if err != nil {
			return 0, fmt.Errorf("failed to get tier2 CIB address for index %d: %w", cibIndex, err)
		}

		cibData, err := sm.device.ReadBlock(cibAddr)
		if err != nil {
			return 0, fmt.Errorf("failed to read tier2 CIB at %d: %w", cibAddr, err)
		}

		blockAddr, ok, err := sm.scanAndAllocateFromCIB(cibData, cibIndex)
		if err != nil {
			return 0, fmt.Errorf("tier2 allocation error in CIB: %w", err)
		}
		if ok {
			return blockAddr, nil
		}
	}

	return 0, fmt.Errorf("no free blocks available in tier2 device")
}

// getTier2CIBAddress retrieves the physical address of a CIB (Chunk Info Block) for the tier2 device.
// CABs on tier2 are indexed independently using Devices[1].CABCount and AddrOffset.
func (sm *SpaceManager) getTier2CIBAddress(cibIndex int) (types.PAddr, error) {
	tier2 := sm.spaceman.Devices[1]

	if cibIndex < 0 || cibIndex >= int(tier2.CIBCount) {
		return 0, fmt.Errorf("invalid tier2 CIB index: %d", cibIndex)
	}

	entriesPerCAB := int(sm.spaceman.BlockSize) / 8
	cabIndex := cibIndex / entriesPerCAB
	offsetInCAB := cibIndex % entriesPerCAB

	if cabIndex >= int(tier2.CABCount) {
		return 0, fmt.Errorf("CAB index %d out of range for tier2", cabIndex)
	}

	cabAddr := types.PAddr(tier2.AddrOffset) + types.PAddr(cabIndex)

	data, err := sm.device.ReadBlock(cabAddr)
	if err != nil {
		return 0, fmt.Errorf("failed to read tier2 CAB at %d: %w", cabAddr, err)
	}

	if len(data) < (offsetInCAB+1)*8 {
		return 0, fmt.Errorf("tier2 CAB block too short")
	}

	addr := binary.LittleEndian.Uint64(data[offsetInCAB*8 : offsetInCAB*8+8])
	return types.PAddr(addr), nil
}

// getBitmapBlock retrieves a bitmap block, using cache if available
func (sm *SpaceManager) getBitmapBlock(addr types.PAddr) ([]byte, error) {
	// Check if we have this bitmap in cache
	if bitmap, ok := sm.bitmapCache[addr]; ok {
		return bitmap, nil
	}

	// Read from disk
	bitmap, err := sm.device.ReadBlock(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to read bitmap block: %w", err)
	}

	// Cache the bitmap
	sm.bitmapCache[addr] = bitmap

	return bitmap, nil
}

// FreeBlock marks a block as free in the space manager
func (sm *SpaceManager) FreeBlock(addr types.PAddr) error {
	// Check if the block is in the internal pool
	if addr >= sm.spaceman.IPBase && addr < sm.spaceman.IPBase+types.PAddr(sm.spaceman.IPBlockCount) {
		return sm.freeInternalPoolBlock(addr)
	}

	// For main device blocks
	return sm.freeMainDeviceBlock(addr)
}

// freeInternalPoolBlock clears the allocation bit for a block in the internal pool bitmap (IPBM),
// effectively marking it as free. This only applies to addresses within the internal pool range.
//
// Returns an error if the block is outside the internal pool, already free,
// or if bitmap read/write operations fail.
func (sm *SpaceManager) freeInternalPoolBlock(addr types.PAddr) error {
	start := sm.spaceman.IPBase
	end := sm.spaceman.IPBase + types.PAddr(sm.spaceman.IPBlockCount)

	if addr < start || addr >= end {
		return fmt.Errorf("block %d is outside internal pool range", addr)
	}

	offset := addr - sm.spaceman.IPBase
	byteIndex := int(offset) / 8
	bitIndex := int(offset) % 8

	bitmapAddr := sm.spaceman.IPBmBase
	bitmap, err := sm.getBitmapBlock(bitmapAddr)
	if err != nil {
		return fmt.Errorf("failed to read IPBM bitmap: %w", err)
	}

	// Safety check
	if byteIndex >= len(bitmap) {
		return fmt.Errorf("bitmap too short for index %d", byteIndex)
	}

	// Check if already free
	if bitmap[byteIndex]&(1<<bitIndex) == 0 {
		return fmt.Errorf("block %d is already free", addr)
	}

	// Clear the bit
	bitmap[byteIndex] &^= 1 << bitIndex

	if err := sm.device.WriteBlock(bitmapAddr, bitmap); err != nil {
		return fmt.Errorf("failed to write IPBM bitmap: %w", err)
	}

	return nil
}

// freeMainDeviceBlock clears the allocation bit for a block on the main device.
// It walks the list of CIBs (via CABs), finds the chunk containing the block,
// and clears the corresponding bit in the chunk's allocation bitmap.
//
// Returns an error if the block is not part of any known chunk, or if bitmap operations fail.
func (sm *SpaceManager) freeMainDeviceBlock(addr types.PAddr) error {
	main := sm.spaceman.Devices[0]

	for cibIndex := 0; cibIndex < int(main.CIBCount); cibIndex++ {
		cibAddr, err := sm.getCIBAddress(cibIndex)
		if err != nil {
			return fmt.Errorf("failed to resolve CIB address for index %d: %w", cibIndex, err)
		}

		cibData, err := sm.device.ReadBlock(cibAddr)
		if err != nil {
			return fmt.Errorf("failed to read CIB at %d: %w", cibAddr, err)
		}

		const chunkInfoSize = 0x20
		numChunks := len(cibData) / chunkInfoSize

		for i := 0; i < numChunks; i++ {
			entry := cibData[i*chunkInfoSize : (i+1)*chunkInfoSize]

			chunkBase := types.PAddr(binary.LittleEndian.Uint64(entry[0:8]))
			bitmapAddr := types.PAddr(binary.LittleEndian.Uint64(entry[8:16]))
			blockCount := binary.LittleEndian.Uint32(entry[16:20])

			if addr < chunkBase || addr >= chunkBase+types.PAddr(blockCount) {
				continue
			}

			offset := addr - chunkBase
			byteIndex := int(offset) / 8
			bitIndex := int(offset) % 8

			bitmap, err := sm.getBitmapBlock(bitmapAddr)
			if err != nil {
				return fmt.Errorf("failed to read bitmap at %d: %w", bitmapAddr, err)
			}

			if byteIndex >= len(bitmap) {
				return fmt.Errorf("bitmap for chunk too short")
			}

			if bitmap[byteIndex]&(1<<bitIndex) == 0 {
				return fmt.Errorf("block %d is already free", addr)
			}

			// Clear bit
			bitmap[byteIndex] &^= 1 << bitIndex

			if err := sm.device.WriteBlock(bitmapAddr, bitmap); err != nil {
				return fmt.Errorf("failed to write updated bitmap: %w", err)
			}

			return nil // Success
		}
	}

	return fmt.Errorf("block %d not found in any main device chunk", addr)
}

// IsBlockAllocated checks whether the given physical address is currently allocated.
// It first checks the internal pool range (IPBM), then walks all main device CIBs/chunks
// to see if the address is tracked and marked allocated in the bitmap.
//
// Returns true if the block is allocated, false if free, or an error if the block isn't tracked.
func (sm *SpaceManager) IsBlockAllocated(addr types.PAddr) (bool, error) {
	// --- Internal Pool Check ---
	if addr >= sm.spaceman.IPBase && addr < sm.spaceman.IPBase+types.PAddr(sm.spaceman.IPBlockCount) {
		offset := addr - sm.spaceman.IPBase
		byteIndex := int(offset) / 8
		bitIndex := int(offset) % 8

		bitmapAddr := sm.spaceman.IPBmBase
		bitmap, err := sm.getBitmapBlock(bitmapAddr)
		if err != nil {
			return false, fmt.Errorf("failed to read IPBM: %w", err)
		}
		if byteIndex >= len(bitmap) {
			return false, fmt.Errorf("IPBM too short for index %d", byteIndex)
		}
		return bitmap[byteIndex]&(1<<bitIndex) != 0, nil
	}

	// --- Main Device Check ---
	main := sm.spaceman.Devices[0]
	for cibIndex := 0; cibIndex < int(main.CIBCount); cibIndex++ {
		cibAddr, err := sm.getCIBAddress(cibIndex)
		if err != nil {
			return false, fmt.Errorf("failed to resolve CIB address %d: %w", cibIndex, err)
		}

		cibData, err := sm.device.ReadBlock(cibAddr)
		if err != nil {
			return false, fmt.Errorf("failed to read CIB at %d: %w", cibAddr, err)
		}

		const chunkInfoSize = 0x20
		numChunks := len(cibData) / chunkInfoSize

		for i := 0; i < numChunks; i++ {
			entry := cibData[i*chunkInfoSize : (i+1)*chunkInfoSize]

			chunkBase := types.PAddr(binary.LittleEndian.Uint64(entry[0:8]))
			bitmapAddr := types.PAddr(binary.LittleEndian.Uint64(entry[8:16]))
			blockCount := binary.LittleEndian.Uint32(entry[16:20])

			if addr < chunkBase || addr >= chunkBase+types.PAddr(blockCount) {
				continue
			}

			offset := addr - chunkBase
			byteIndex := int(offset) / 8
			bitIndex := int(offset) % 8

			bitmap, err := sm.getBitmapBlock(bitmapAddr)
			if err != nil {
				return false, fmt.Errorf("failed to read bitmap at %d: %w", bitmapAddr, err)
			}
			if byteIndex >= len(bitmap) {
				return false, fmt.Errorf("bitmap too short")
			}

			return bitmap[byteIndex]&(1<<bitIndex) != 0, nil
		}
	}

	return false, fmt.Errorf("block %d is not tracked by internal pool or main device", addr)
}

// GetFreeBlockCount returns the total number of free blocks
// available in the internal pool and main device. It walks the internal
// bitmap and all CIBs/chunks, summing the number of unset bits (0 = free).
func (sm *SpaceManager) GetFreeBlockCount() (uint64, error) {
	var total uint64

	// --- Count free in Internal Pool ---
	internalCount, err := sm.countFreeBlocksInIPBM()
	if err != nil {
		return 0, fmt.Errorf("failed to count internal pool free blocks: %w", err)
	}
	total += internalCount

	// --- Count free in Main Device ---
	mainCount, err := sm.countFreeBlocksInDevice(0)
	if err != nil {
		return 0, fmt.Errorf("failed to count main device free blocks: %w", err)
	}
	total += mainCount

	// Optional: Add Tier2 device if available
	if sm.spaceman.Devices[1].CIBCount > 0 {
		tier2Count, err := sm.countFreeBlocksInDevice(1)
		if err != nil {
			return 0, fmt.Errorf("failed to count tier2 device free blocks: %w", err)
		}
		total += tier2Count
	}

	return total, nil
}

func (sm *SpaceManager) countFreeBlocksInDevice(deviceIndex int) (uint64, error) {
	dev := sm.spaceman.Devices[deviceIndex]
	var total uint64

	for cibIndex := 0; cibIndex < int(dev.CIBCount); cibIndex++ {
		var cibAddr types.PAddr
		var err error

		if deviceIndex == 0 {
			cibAddr, err = sm.getCIBAddress(cibIndex)
		} else {
			cibAddr, err = sm.getTier2CIBAddress(cibIndex)
		}

		if err != nil {
			return 0, err
		}

		cibData, err := sm.device.ReadBlock(cibAddr)
		if err != nil {
			return 0, err
		}

		const chunkInfoSize = 0x20
		numChunks := len(cibData) / chunkInfoSize

		for i := 0; i < numChunks; i++ {
			entry := cibData[i*chunkInfoSize : (i+1)*chunkInfoSize]
			freeCount := binary.LittleEndian.Uint32(entry[20:24]) // chunk_info.free_count
			total += uint64(freeCount)
		}
	}

	return total, nil
}

func (sm *SpaceManager) countFreeBlocksInIPBM() (uint64, error) {
	if sm.spaceman.IPBmSizeInBlocks == 0 {
		return 0, nil
	}

	bitmapAddr := sm.spaceman.IPBmBase
	bitmap, err := sm.getBitmapBlock(bitmapAddr)
	if err != nil {
		return 0, err
	}

	var free uint64
	for _, b := range bitmap {
		for i := 0; i < 8; i++ {
			if b&(1<<i) == 0 {
				free++
			}
		}
	}

	return free, nil
}

// GetContiguousBlocks attempts to allocate a sequence of `count` contiguous blocks.
// It first tries the internal pool bitmap (IPBM), then scans the main device chunk
// bitmaps (CIBs) to find and reserve a contiguous run of free blocks.
//
// Returns the starting physical address on success, or an error if no such range exists.
func (sm *SpaceManager) GetContiguousBlocks(count uint32) (types.PAddr, error) {
	// Try internal pool first
	addr, err := sm.getContiguousFromIPBM(count)
	if err == nil {
		return addr, nil
	}

	// Fall back to main device
	addr, err = sm.getContiguousFromDevice(0, count)
	if err == nil {
		return addr, nil
	}

	// Optionally: try tier2 device
	if sm.spaceman.Devices[1].CIBCount > 0 {
		return sm.getContiguousFromDevice(1, count)
	}

	return 0, fmt.Errorf("no %d contiguous blocks available", count)
}

// getContiguousFromIPBM searches the internal pool bitmap for a sequence of `count` free blocks.
// On success, it marks them allocated and returns the starting address.
func (sm *SpaceManager) getContiguousFromIPBM(count uint32) (types.PAddr, error) {
	if sm.spaceman.IPBmSizeInBlocks == 0 {
		return 0, fmt.Errorf("IPBM not available")
	}

	bitmapAddr := sm.spaceman.IPBmBase
	bitmap, err := sm.getBitmapBlock(bitmapAddr)
	if err != nil {
		return 0, err
	}

	bitLen := len(bitmap) * 8
	run := 0
	start := -1

	for i := 0; i < bitLen; i++ {
		byteIndex := i / 8
		bitIndex := i % 8

		if bitmap[byteIndex]&(1<<bitIndex) == 0 {
			if run == 0 {
				start = i
			}
			run++
			if run == int(count) {
				// Found run: mark bits as allocated
				for j := 0; j < int(count); j++ {
					b := (start + j) / 8
					k := (start + j) % 8
					bitmap[b] |= 1 << k
				}

				if err := sm.device.WriteBlock(bitmapAddr, bitmap); err != nil {
					return 0, fmt.Errorf("failed to update IPBM: %w", err)
				}

				return sm.spaceman.IPBase + types.PAddr(start), nil
			}
		} else {
			run = 0
			start = -1
		}
	}

	return 0, fmt.Errorf("no contiguous range of %d blocks in IPBM", count)
}

// getContiguousFromDevice scans all CIBs on the specified device for a run of `count` free blocks.
// On success, the blocks are marked allocated in the chunk bitmap and the starting address returned.
func (sm *SpaceManager) getContiguousFromDevice(deviceIndex int, count uint32) (types.PAddr, error) {
	dev := sm.spaceman.Devices[deviceIndex]

	for cibIndex := 0; cibIndex < int(dev.CIBCount); cibIndex++ {
		var cibAddr types.PAddr
		var err error

		if deviceIndex == 0 {
			cibAddr, err = sm.getCIBAddress(cibIndex)
		} else {
			cibAddr, err = sm.getTier2CIBAddress(cibIndex)
		}
		if err != nil {
			return 0, err
		}

		cibData, err := sm.device.ReadBlock(cibAddr)
		if err != nil {
			return 0, err
		}

		const chunkInfoSize = 0x20
		numChunks := len(cibData) / chunkInfoSize

		for i := 0; i < numChunks; i++ {
			entry := cibData[i*chunkInfoSize : (i+1)*chunkInfoSize]

			chunkBase := types.PAddr(binary.LittleEndian.Uint64(entry[0:8]))
			bitmapAddr := types.PAddr(binary.LittleEndian.Uint64(entry[8:16]))
			blockCount := binary.LittleEndian.Uint32(entry[16:20])
			if blockCount < count {
				continue
			}

			bitmap, err := sm.getBitmapBlock(bitmapAddr)
			if err != nil {
				return 0, err
			}

			run := 0
			start := -1
			for j := 0; j < int(blockCount); j++ {
				byteIndex := j / 8
				bitIndex := j % 8

				if bitmap[byteIndex]&(1<<bitIndex) == 0 {
					if run == 0 {
						start = j
					}
					run++
					if run == int(count) {
						// Found run — mark bits
						for k := 0; k < int(count); k++ {
							b := (start + k) / 8
							kb := (start + k) % 8
							bitmap[b] |= 1 << kb
						}
						if err := sm.device.WriteBlock(bitmapAddr, bitmap); err != nil {
							return 0, fmt.Errorf("failed to write chunk bitmap: %w", err)
						}
						return chunkBase + types.PAddr(start), nil
					}
				} else {
					run = 0
					start = -1
				}
			}
		}
	}

	return 0, fmt.Errorf("no contiguous block range of size %d on device[%d]", count, deviceIndex)
}
