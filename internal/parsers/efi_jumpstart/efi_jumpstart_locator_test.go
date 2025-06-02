package efijumpstart

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Creates a mock container superblock with specific properties
func createMockSuperblock(jumpstartPaddr types.Paddr, xpDescBase types.Paddr, xpDescBlocks, xpDescIndex, xpDescNext uint32) []byte {
	// Create a buffer large enough for the full superblock structure (minimum 256 bytes for tests)
	buf := make([]byte, 1536) // Size of NxSuperblockT structure

	// Write ObjPhysT header (32 bytes) - we'll just zero it for testing
	// In a real implementation, this would have checksum, OID, XID, type, subtype

	// Write the magic number at offset 32 (after ObjPhysT header)
	// The magic is "NXSB" which is 0x4E585342 in little-endian
	binary.LittleEndian.PutUint32(buf[32:36], 0x4E585342) // "NXSB" magic

	// Write the block size at offset 36
	binary.LittleEndian.PutUint32(buf[36:40], testBlockSize)

	// Write block count at offset 40
	binary.LittleEndian.PutUint64(buf[40:48], 1000000) // Some reasonable block count

	// Skip features, UUID, etc. for now and go to checkpoint fields

	// NxXpDescBlocks at offset 104
	binary.LittleEndian.PutUint32(buf[104:108], xpDescBlocks)

	// NxXpDescBase at offset 112
	binary.LittleEndian.PutUint64(buf[112:120], uint64(xpDescBase))

	// NxXpDescNext at offset 128
	binary.LittleEndian.PutUint32(buf[128:132], xpDescNext)

	// NxXpDescIndex at offset 136
	binary.LittleEndian.PutUint32(buf[136:140], xpDescIndex)

	// Skip to EFI jumpstart field - need to calculate the correct offset
	// Based on the container reader, this should be around offset where NxEfiJumpstart is parsed
	// Looking at the container reader, it's parsed after many fields, let me find the right offset

	// Calculate the exact offset as done in the container reader:
	// Start at 184 (after fixed fields)
	// + NxMaxFileSystems * 8 (volume OIDs array)
	// + NxNumCounters * 8 (counters array)
	// + 16 (blocked out range)
	// + 8 (evict mapping tree OID)
	// + 8 (flags)
	efiJumpstartOffset := 184 + (types.NxMaxFileSystems * 8) + (types.NxNumCounters * 8) + 16 + 8 + 8
	binary.LittleEndian.PutUint64(buf[efiJumpstartOffset:efiJumpstartOffset+8], uint64(jumpstartPaddr))

	return buf
}

// Creates a mock checkpoint map with specified mappings
func createMockCheckpointMap(objType uint32, flags uint32, mappings []types.CheckpointMappingT) []byte {
	// Create a buffer of full block size
	buf := make([]byte, testBlockSize)

	// Write ObjPhysT header (32 bytes)
	// For testing, we'll just write the object type at the correct offset
	binary.LittleEndian.PutUint32(buf[24:28], objType) // OType at offset 24 in ObjPhysT
	binary.LittleEndian.PutUint32(buf[28:32], 0)       // OSubtype at offset 28

	// Write the checkpoint map flags at offset 32 (after ObjPhysT)
	binary.LittleEndian.PutUint32(buf[32:36], flags)

	// Write the count of mappings at offset 36
	binary.LittleEndian.PutUint32(buf[36:40], uint32(len(mappings)))

	// Write each mapping starting at offset 40
	mappingOffset := 40
	for _, mapping := range mappings {
		// Write the mapping type
		binary.LittleEndian.PutUint32(buf[mappingOffset:mappingOffset+4], mapping.CpmType)

		// Write the mapping subtype
		binary.LittleEndian.PutUint32(buf[mappingOffset+4:mappingOffset+8], mapping.CpmSubtype)

		// Write the mapping size
		binary.LittleEndian.PutUint32(buf[mappingOffset+8:mappingOffset+12], mapping.CpmSize)

		// Write padding (CpmPad)
		binary.LittleEndian.PutUint32(buf[mappingOffset+12:mappingOffset+16], 0)

		// Write the FS OID
		binary.LittleEndian.PutUint64(buf[mappingOffset+16:mappingOffset+24], uint64(mapping.CpmFsOid))

		// Write the mapping OID
		binary.LittleEndian.PutUint64(buf[mappingOffset+24:mappingOffset+32], uint64(mapping.CpmOid))

		// Write the mapping physical address
		binary.LittleEndian.PutUint64(buf[mappingOffset+32:mappingOffset+40], uint64(mapping.CpmPaddr))

		// Move to next mapping (each mapping is 48 bytes)
		mappingOffset += 48
	}

	return buf
}

// Creates a mock EFI jumpstart structure
func createMockJumpstart() []byte {
	// Create a buffer large enough for the whole jumpstart structure
	buf := make([]byte, 512)

	// Create the object header (typically first 24 bytes but we only need the first few)
	// Skip writing full ObjPhysT header, just write essential data

	// Write the magic number where it should be (after obj header)
	binary.LittleEndian.PutUint32(buf[8:12], types.NxEfiJumpstartMagic)

	// Write the version
	binary.LittleEndian.PutUint32(buf[12:16], types.NxEfiJumpstartVersion)

	// Write file length
	binary.LittleEndian.PutUint32(buf[16:20], 12345)

	// Write number of extents
	binary.LittleEndian.PutUint32(buf[20:24], 2)

	// Skip reserved fields (16 * 8 = 128 bytes)

	// Write extents at offset 152 (after header + reserved)
	extentOffset := 152

	// First extent
	binary.LittleEndian.PutUint64(buf[extentOffset:extentOffset+8], uint64(100)) // Start paddr
	binary.LittleEndian.PutUint64(buf[extentOffset+8:extentOffset+16], 1)        // Block count

	// Second extent
	binary.LittleEndian.PutUint64(buf[extentOffset+16:extentOffset+24], uint64(200)) // Start paddr
	binary.LittleEndian.PutUint64(buf[extentOffset+24:extentOffset+32], 2)           // Block count

	return buf
}

// Creates mock disk data with superblock, checkpoint maps, and jumpstart structure
func createMockDisk(
	jumpstartPaddr types.Paddr,
	xpDescBase types.Paddr,
	xpDescBlocks, xpDescIndex, xpDescNext uint32,
	checkpointMaps map[types.Paddr][]byte,
	includeJumpstart bool,
	signatureScanPaddr types.Paddr,
) []byte {
	// Determine the size of the mock disk
	maxPaddr := types.Paddr(10) // Default size for basic container

	// Find the maximum address that needs to be included
	if jumpstartPaddr > maxPaddr {
		maxPaddr = jumpstartPaddr
	}
	if xpDescBase+types.Paddr(xpDescBlocks) > maxPaddr {
		maxPaddr = xpDescBase + types.Paddr(xpDescBlocks)
	}
	for paddr := range checkpointMaps {
		if paddr > maxPaddr {
			maxPaddr = paddr
		}
	}
	if signatureScanPaddr > maxPaddr {
		maxPaddr = signatureScanPaddr
	}

	// Add some buffer
	maxPaddr += 5

	// Create the mock disk with enough space for all blocks
	diskSize := int64(maxPaddr+1) * int64(testBlockSize)
	disk := make([]byte, diskSize)

	// Write the superblock at block 0
	superblock := createMockSuperblock(jumpstartPaddr, xpDescBase, xpDescBlocks, xpDescIndex, xpDescNext)
	copy(disk[0:], superblock)

	// Write checkpoint maps at their respective addresses
	for paddr, cpm := range checkpointMaps {
		offset := int64(paddr) * int64(testBlockSize)
		copy(disk[offset:offset+int64(len(cpm))], cpm)
	}

	// Write jumpstart structure if needed
	if includeJumpstart && jumpstartPaddr > 0 {
		jumpstartData := createMockJumpstart()
		offset := int64(jumpstartPaddr) * int64(testBlockSize)
		copy(disk[offset:offset+int64(len(jumpstartData))], jumpstartData)
	}

	// Write jumpstart signature for scan test if needed
	if signatureScanPaddr > 0 {
		// Just write the magic and version at the specified address
		offset := int64(signatureScanPaddr) * int64(testBlockSize)
		binary.LittleEndian.PutUint32(disk[offset:], types.NxEfiJumpstartMagic)
		binary.LittleEndian.PutUint32(disk[offset+4:], types.NxEfiJumpstartVersion)
	}

	return disk
}

// --- Test FindEFIJumpstart with Jumpstart in Superblock ---

func TestFindEFIJumpstart_JumpstartInSuperblock(t *testing.T) {
	// Create mock disk with jumpstart paddr in superblock
	mockDisk := createMockDisk(
		testJumpstartPaddr, // Jumpstart address in superblock
		0,                  // No checkpoint descriptor area
		0, 0, 0,            // No checkpoint blocks/indices
		nil,  // No checkpoint maps
		true, // Include jumpstart structure
		0,    // No signature scan
	)

	reader := bytes.NewReader(mockDisk)
	locator, err := NewAPFSJumpstartLocator(reader, testBlockSize, 0)
	require.NoError(t, err)

	// Test finding the jumpstart
	paddr, err := locator.FindEFIJumpstart()
	require.NoError(t, err)
	assert.Equal(t, testJumpstartPaddr, paddr)
}

// --- Test FindEFIJumpstart with Invalid Superblock ---

func TestFindEFIJumpstart_InvalidSuperblock(t *testing.T) {
	// Create an invalid superblock (wrong magic)
	mockSuperblock := createMockSuperblock(0, 0, 0, 0, 0)

	// Corrupt the magic number at the correct offset (32)
	binary.LittleEndian.PutUint32(mockSuperblock[32:36], 0xBADBADBA) // Wrong magic

	reader := bytes.NewReader(mockSuperblock)
	locator, err := NewAPFSJumpstartLocator(reader, testBlockSize, 0)
	require.NoError(t, err)

	// Test finding the jumpstart - should fail
	_, err = locator.FindEFIJumpstart()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid container superblock magic")
}

// --- Test FindEFIJumpstart with Jumpstart in Checkpoint Map ---

func TestFindEFIJumpstart_JumpstartInCheckpointMap(t *testing.T) {
	// Create checkpoint map with jumpstart mapping
	jumpstartMapping := types.CheckpointMappingT{
		CpmType:    types.ObjectTypeEfiJumpstart,
		CpmSubtype: 0,
		CpmSize:    testBlockSize,
		CpmOid:     123,
		CpmPaddr:   testJumpstartPaddr,
	}

	// Another mapping for testing
	otherMapping := types.CheckpointMappingT{
		CpmType:    types.ObjectTypeFs,
		CpmSubtype: 0,
		CpmSize:    testBlockSize,
		CpmOid:     456,
		CpmPaddr:   50,
	}

	// Create a checkpoint map with these mappings
	checkpointMap := createMockCheckpointMap(
		types.ObjectTypeCheckpointMap,
		types.CheckpointMapLast, // Last flag set
		[]types.CheckpointMappingT{otherMapping, jumpstartMapping},
	)

	checkpointMaps := map[types.Paddr][]byte{
		1: checkpointMap, // Place at block 1
	}

	// Create mock disk
	mockDisk := createMockDisk(
		0,       // No jumpstart in superblock
		1,       // Checkpoint descriptor at block 1
		2, 0, 1, // 2 checkpoint blocks, index 0, next 1
		checkpointMaps, // Include our checkpoint map
		true,           // Include jumpstart structure
		0,              // No signature scan
	)

	reader := bytes.NewReader(mockDisk)
	locator, err := NewAPFSJumpstartLocator(reader, testBlockSize, 0)
	require.NoError(t, err)

	// Test finding the jumpstart
	paddr, err := locator.FindEFIJumpstart()
	require.NoError(t, err)
	assert.Equal(t, testJumpstartPaddr, paddr)
}

// --- Test FindEFIJumpstart with Multi-Block Checkpoint Map Chain ---

func TestFindEFIJumpstart_ChainedCheckpointMaps(t *testing.T) {
	// Create first checkpoint map with link to second map
	firstMapPaddr := types.Paddr(1)
	secondMapPaddr := types.Paddr(2)

	// Mapping to second checkpoint map
	secondMapMapping := types.CheckpointMappingT{
		CpmType:    types.ObjectTypeCheckpointMap,
		CpmSubtype: 0,
		CpmSize:    testBlockSize,
		CpmOid:     456,
		CpmPaddr:   secondMapPaddr,
	}

	// Create first checkpoint map (no last flag, has link to second map)
	firstMap := createMockCheckpointMap(
		types.ObjectTypeCheckpointMap,
		0, // No last flag
		[]types.CheckpointMappingT{secondMapMapping},
	)

	// Jumpstart mapping for second map
	jumpstartMapping := types.CheckpointMappingT{
		CpmType:    types.ObjectTypeEfiJumpstart,
		CpmSubtype: 0,
		CpmSize:    testBlockSize,
		CpmOid:     789,
		CpmPaddr:   testJumpstartPaddr,
	}

	// Create second checkpoint map with jumpstart mapping
	secondMap := createMockCheckpointMap(
		types.ObjectTypeCheckpointMap,
		types.CheckpointMapLast, // Last flag set
		[]types.CheckpointMappingT{jumpstartMapping},
	)

	checkpointMaps := map[types.Paddr][]byte{
		firstMapPaddr:  firstMap,
		secondMapPaddr: secondMap,
	}

	// Create mock disk
	mockDisk := createMockDisk(
		0,             // No jumpstart in superblock
		firstMapPaddr, // Checkpoint descriptor at first map
		2, 0, 1,       // 2 checkpoint blocks, index 0, next 1
		checkpointMaps, // Include our checkpoint maps
		true,           // Include jumpstart structure
		0,              // No signature scan
	)

	reader := bytes.NewReader(mockDisk)
	locator, err := NewAPFSJumpstartLocator(reader, testBlockSize, 0)
	require.NoError(t, err)

	// Test finding the jumpstart through the chain
	paddr, err := locator.FindEFIJumpstart()
	require.NoError(t, err)
	assert.Equal(t, testJumpstartPaddr, paddr)
}

// --- Test FindEFIJumpstart with Signature Scan Fallback ---

func TestFindEFIJumpstart_SignatureScanFallback(t *testing.T) {
	scanPaddr := types.Paddr(5) // Signature at block 5

	// Create a valid checkpoint map but with no jumpstart mapping
	otherMapping := types.CheckpointMappingT{
		CpmType:    types.ObjectTypeFs,
		CpmSubtype: 0,
		CpmSize:    testBlockSize,
		CpmOid:     456,
		CpmPaddr:   50,
	}

	checkpointMap := createMockCheckpointMap(
		types.ObjectTypeCheckpointMap,
		types.CheckpointMapLast, // Last flag set
		[]types.CheckpointMappingT{otherMapping},
	)

	checkpointMaps := map[types.Paddr][]byte{
		1: checkpointMap, // Place at block 1
	}

	// Create mock disk with signature scan location
	mockDisk := createMockDisk(
		0,        // No jumpstart in superblock
		1,        // Checkpoint descriptor at block 1
		10, 0, 1, // 10 checkpoint blocks (for scan), index 0, next 1
		checkpointMaps, // Include our checkpoint map without jumpstart mapping
		false,          // Don't include full jumpstart structure
		scanPaddr,      // Include jumpstart signature at this address
	)

	reader := bytes.NewReader(mockDisk)
	locator, err := NewAPFSJumpstartLocator(reader, testBlockSize, 0)
	require.NoError(t, err)

	// Test finding the jumpstart - should find via signature scan
	paddr, err := locator.FindEFIJumpstart()
	require.NoError(t, err)
	assert.Equal(t, scanPaddr, paddr)
}

// --- Test FindEFIJumpstart with No Jumpstart Found ---

func TestFindEFIJumpstart_NoJumpstartFound(t *testing.T) {
	// Create a valid checkpoint map but with no jumpstart mapping
	otherMapping := types.CheckpointMappingT{
		CpmType:    types.ObjectTypeFs,
		CpmSubtype: 0,
		CpmSize:    testBlockSize,
		CpmOid:     456,
		CpmPaddr:   50,
	}

	checkpointMap := createMockCheckpointMap(
		types.ObjectTypeCheckpointMap,
		types.CheckpointMapLast, // Last flag set
		[]types.CheckpointMappingT{otherMapping},
	)

	checkpointMaps := map[types.Paddr][]byte{
		1: checkpointMap, // Place at block 1
	}

	// Create mock disk without any jumpstart or signature scan location
	mockDisk := createMockDisk(
		0,        // No jumpstart in superblock
		1,        // Checkpoint descriptor at block 1
		10, 0, 1, // 10 checkpoint blocks (for scan), index 0, next 1
		checkpointMaps, // Include our checkpoint map without jumpstart mapping
		false,          // Don't include full jumpstart structure
		0,              // No signature scan location
	)

	reader := bytes.NewReader(mockDisk)
	locator, err := NewAPFSJumpstartLocator(reader, testBlockSize, 0)
	require.NoError(t, err)

	// Test finding the jumpstart - should fail with "not found"
	_, err = locator.FindEFIJumpstart()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "EFI jumpstart structure not found")
}

// --- Test FindEFIJumpstart with Invalid Checkpoint Map ---

func TestFindEFIJumpstart_InvalidCheckpointMap(t *testing.T) {
	// Create an invalid checkpoint map (wrong type)
	invalidMap := createMockCheckpointMap(
		types.ObjectTypeFs, // Wrong type for checkpoint map
		0,
		[]types.CheckpointMappingT{},
	)

	checkpointMaps := map[types.Paddr][]byte{
		1: invalidMap,
	}

	// Create mock disk with invalid checkpoint map
	mockDisk := createMockDisk(
		0,       // No jumpstart in superblock
		1,       // Checkpoint descriptor at block 1
		2, 0, 1, // 2 checkpoint blocks, index 0, next 1
		checkpointMaps, // Include our invalid checkpoint map
		false,          // Don't include jumpstart structure
		5,              // Add signature scan location at block 5
	)

	reader := bytes.NewReader(mockDisk)
	locator, err := NewAPFSJumpstartLocator(reader, testBlockSize, 0)
	require.NoError(t, err)

	// Test finding the jumpstart - should find via signature scan fallback
	paddr, err := locator.FindEFIJumpstart()
	require.NoError(t, err)
	assert.Equal(t, types.Paddr(5), paddr)
}

// --- Test ReadJumpstartObjectFromPaddr ---

func TestReadJumpstartObjectFromPaddr(t *testing.T) {
	// Create mock disk with jumpstart structure
	mockDisk := createMockDisk(
		testJumpstartPaddr, // Jumpstart at this address
		0, 0, 0, 0,         // No checkpoint data
		nil,  // No checkpoint maps
		true, // Include jumpstart structure
		0,    // No signature scan
	)

	reader := bytes.NewReader(mockDisk)
	locator, err := NewAPFSJumpstartLocator(reader, testBlockSize, 0)
	require.NoError(t, err)

	// Test reading the jumpstart object
	jumpstart, err := locator.ReadJumpstartObjectFromPaddr(testJumpstartPaddr)
	require.NoError(t, err)
	assert.NotNil(t, jumpstart)
	assert.Equal(t, types.NxEfiJumpstartMagic, jumpstart.NejMagic)
	assert.Equal(t, types.NxEfiJumpstartVersion, jumpstart.NejVersion)
	assert.Equal(t, uint32(12345), jumpstart.NejEfiFileLen)
	assert.Equal(t, uint32(2), jumpstart.NejNumExtents)
	assert.Len(t, jumpstart.NejRecExtents, 2)
	assert.Equal(t, types.Paddr(100), jumpstart.NejRecExtents[0].PrStartPaddr)
	assert.Equal(t, uint64(1), jumpstart.NejRecExtents[0].PrBlockCount)
	assert.Equal(t, types.Paddr(200), jumpstart.NejRecExtents[1].PrStartPaddr)
	assert.Equal(t, uint64(2), jumpstart.NejRecExtents[1].PrBlockCount)
}

// --- Test with Partition Offset ---

func TestFindEFIJumpstart_WithPartitionOffset(t *testing.T) {
	partitionOffset := int64(1024 * 1024) // 1MB offset

	// Create mock disk with jumpstart paddr in superblock
	mockDisk := createMockDisk(
		testJumpstartPaddr, // Jumpstart address in superblock
		0,                  // No checkpoint descriptor area
		0, 0, 0,            // No checkpoint blocks/indices
		nil,  // No checkpoint maps
		true, // Include jumpstart structure
		0,    // No signature scan
	)

	// Create a disk with partition offset
	fullDisk := make([]byte, partitionOffset+int64(len(mockDisk)))
	copy(fullDisk[partitionOffset:], mockDisk)

	reader := bytes.NewReader(fullDisk)
	locator, err := NewAPFSJumpstartLocator(reader, testBlockSize, partitionOffset)
	require.NoError(t, err)

	// Test finding the jumpstart
	paddr, err := locator.FindEFIJumpstart()
	require.NoError(t, err)
	assert.Equal(t, testJumpstartPaddr, paddr)
}

// --- Test readContainerSuperblock Error Handling ---

func TestReadContainerSuperblock_ReadError(t *testing.T) {
	// Create a mock reader that returns an error
	errReader := &mockReaderAt{
		data:        []byte{},
		failOnRead:  true,
		failAfterN:  0,
		errToReturn: errors.New("read error"),
	}

	locator, err := NewAPFSJumpstartLocator(errReader, testBlockSize, 0)
	require.NoError(t, err)

	// Test reading the superblock - should fail
	_, err = locator.readContainerSuperblock()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read container superblock")
}

// --- Test ReadCheckpointMap Error Handling ---

func TestReadCheckpointMap_ReadError(t *testing.T) {
	// Create a mock reader that fails on specific reads
	mockDisk := createMockDisk(
		0, 1, 2, 0, 1, // Setup with checkpoint data
		nil, false, 0, // No maps or jumpstart
	)

	errReader := &mockReaderAt{
		data:        mockDisk,
		failOnRead:  true,
		failAfterN:  1, // Fail after reading superblock
		errToReturn: errors.New("checkpoint read error"),
	}

	locator, err := NewAPFSJumpstartLocator(errReader, testBlockSize, 0)
	require.NoError(t, err)

	// We need to test an internal method, so we'll test through FindEFIJumpstart
	_, err = locator.FindEFIJumpstart()
	require.Error(t, err)
	// The error could be wrapped, so test for containment
	assert.Contains(t, err.Error(), "failed to find jumpstart from checkpoints")
}

// --- Helper function tests ---

func TestIsValidCheckpointMap(t *testing.T) {
	locator, err := NewAPFSJumpstartLocator(bytes.NewReader([]byte{}), testBlockSize, 0)
	require.NoError(t, err)

	// Test with nil map
	assert.False(t, locator.isValidCheckpointMap(nil))

	// Test with wrong object type
	wrongTypeMap := &types.CheckpointMapPhysT{
		CpmO: types.ObjPhysT{
			OType: types.ObjectTypeFs, // Wrong type
		},
		CpmCount: 1,
		CpmMap:   make([]types.CheckpointMappingT, 1),
	}
	assert.False(t, locator.isValidCheckpointMap(wrongTypeMap))

	// Test with zero count
	zeroCountMap := &types.CheckpointMapPhysT{
		CpmO: types.ObjPhysT{
			OType: types.ObjectTypeCheckpointMap,
		},
		CpmCount: 0,
		CpmMap:   []types.CheckpointMappingT{},
	}
	assert.False(t, locator.isValidCheckpointMap(zeroCountMap))

	// Test with too large count
	largeCountMap := &types.CheckpointMapPhysT{
		CpmO: types.ObjPhysT{
			OType: types.ObjectTypeCheckpointMap,
		},
		CpmCount: 2000, // Larger than limit
		CpmMap:   make([]types.CheckpointMappingT, 2000),
	}
	assert.False(t, locator.isValidCheckpointMap(largeCountMap))

	// Test with mismatched map length
	mismatchMap := &types.CheckpointMapPhysT{
		CpmO: types.ObjPhysT{
			OType: types.ObjectTypeCheckpointMap,
		},
		CpmCount: 2,
		CpmMap:   make([]types.CheckpointMappingT, 1), // Only 1 element
	}
	assert.False(t, locator.isValidCheckpointMap(mismatchMap))

	// Test with valid map
	validMap := &types.CheckpointMapPhysT{
		CpmO: types.ObjPhysT{
			OType: types.ObjectTypeCheckpointMap,
		},
		CpmCount: 1,
		CpmMap:   make([]types.CheckpointMappingT, 1),
	}
	assert.True(t, locator.isValidCheckpointMap(validMap))
}

// --- Edge Case Tests ---

func TestFindEFIJumpstart_InvalidCheckpointIndices(t *testing.T) {
	// Create superblock with invalid indices
	mockDisk := createMockDisk(
		0,       // No jumpstart in superblock
		1,       // Checkpoint descriptor at block 1
		1, 2, 2, // Invalid indices (both 2, but max is 1)
		nil,   // No checkpoint maps
		false, // Don't include jumpstart structure
		0,     // No signature scan
	)

	reader := bytes.NewReader(mockDisk)
	locator, err := NewAPFSJumpstartLocator(reader, testBlockSize, 0)
	require.NoError(t, err)

	// Test should handle invalid indices gracefully
	_, err = locator.FindEFIJumpstart()
	require.Error(t, err)
	// Changed error message check to match the way errors are wrapped
	assert.Contains(t, err.Error(), "failed to find jumpstart from checkpoints")
}

func TestFindEFIJumpstart_ZeroBlockSize(t *testing.T) {
	// Test constructor with zero block size
	_, err := NewAPFSJumpstartLocator(bytes.NewReader([]byte{}), 0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "block size cannot be zero")
}

func TestFindEFIJumpstart_NegativePartitionOffset(t *testing.T) {
	// Test constructor with negative partition offset
	_, err := NewAPFSJumpstartLocator(bytes.NewReader([]byte{}), testBlockSize, -1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "partition offset cannot be negative")
}

func TestFindEFIJumpstart_NilReader(t *testing.T) {
	// Test constructor with nil reader
	_, err := NewAPFSJumpstartLocator(nil, testBlockSize, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reader cannot be nil")
}
