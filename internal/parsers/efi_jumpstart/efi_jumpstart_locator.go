// File: internal/efijumpstart/efi_jumpstart_locator.go
package efijumpstart

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/parsers/container"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// Constants for APFS container structure
// APFS Container Constants
const (
	// Block 0 is the container superblock
	containerSuperblockOffset = 0
	// Standard APFS block size (default, can be overridden)
	defaultBlockSize = 4096
	// Typical location of checkpoint descriptor area in blocks from the start
	checkpointDescAreaStartOffset = 1
)

// APFSJumpstartLocator provides a complete implementation to find
// EFI jumpstart structures in APFS containers
type APFSJumpstartLocator struct {
	reader          io.ReaderAt
	blockSize       uint32
	partitionOffset int64 // Offset in bytes to the start of the APFS container
}

// Ensure APFSJumpstartLocator implements the EFIJumpstartLocator interface
var _ interfaces.EFIJumpstartLocator = (*APFSJumpstartLocator)(nil)

// NewAPFSJumpstartLocator creates a locator that can find EFI jumpstart structures
// in an APFS container.
//
// Parameters:
//   - reader: Interface to read the raw disk/image
//   - blockSize: The block size used by the APFS container (usually 4096)
//   - partitionOffset: The byte offset to the start of the APFS container partition
func NewAPFSJumpstartLocator(reader io.ReaderAt, blockSize uint32, partitionOffset int64) (*APFSJumpstartLocator, error) {
	if reader == nil {
		return nil, fmt.Errorf("reader cannot be nil")
	}
	if blockSize == 0 {
		return nil, fmt.Errorf("block size cannot be zero")
	}
	if partitionOffset < 0 {
		return nil, fmt.Errorf("partition offset cannot be negative, got %d", partitionOffset)
	}

	return &APFSJumpstartLocator{
		reader:          reader,
		blockSize:       blockSize,
		partitionOffset: partitionOffset,
	}, nil
}

// Using types.NxSuperblockT from the types package for the container superblock

// FindEFIJumpstart locates and returns the physical address of the EFI jumpstart
// structure within the APFS container.
func (l *APFSJumpstartLocator) FindEFIJumpstart() (types.Paddr, error) {
	// First, attempt to get jumpstart address from container superblock (block 0)
	superblock, err := l.readContainerSuperblock()
	if err != nil {
		return 0, fmt.Errorf("failed to read container superblock: %w", err)
	}

	// Check if magic is valid
	if superblock.NxMagic != types.NxMagic {
		return 0, fmt.Errorf("invalid container superblock magic: expected %#x, got %#x", types.NxMagic, superblock.NxMagic)
	}

	// If the EFI jumpstart pointer is valid in the superblock, use it
	if superblock.NxEfiJumpstart != 0 {
		return superblock.NxEfiJumpstart, nil
	}

	// If not found in superblock, check if we need to look in the checkpoint descriptor area
	// This requires parsing checkpoint descriptors which are more complex
	// Let's implement checkpoint-based search as a fallback
	jumpstartAddr, err := l.findJumpstartFromCheckpoints(superblock)
	if err != nil {
		return 0, fmt.Errorf("failed to find jumpstart from checkpoints: %w", err)
	}

	if jumpstartAddr == 0 {
		return 0, fmt.Errorf("EFI jumpstart structure not found in container")
	}

	return jumpstartAddr, nil
}

// FindEFIJumpstartInPartition locates the EFI jumpstart structure in a specific APFS partition.
// This implementation searches for the partition in a GPT-formatted disk, then locates the
// jumpstart within that partition.
func (l *APFSJumpstartLocator) FindEFIJumpstartInPartition(partitionUUID string) (types.Paddr, error) {
	// Create a GPT partition manager to find the APFS partition
	partManager, err := NewGPTPartitionManager(l.reader, uint64(l.blockSize), "")
	if err != nil {
		return 0, fmt.Errorf("failed to create partition manager: %w", err)
	}

	// Check if the provided UUID is a valid APFS partition
	if !partManager.IsAPFSPartition(partitionUUID) {
		return 0, fmt.Errorf("partition UUID %s is not a valid APFS partition", partitionUUID)
	}

	// Get all partitions to find the one with matching UUID
	partitions, err := l.getAllPartitions()
	if err != nil {
		return 0, fmt.Errorf("failed to get partition list: %w", err)
	}

	var targetPartition *PartitionInfo
	for i := range partitions {
		if partitions[i].UUID == partitionUUID {
			targetPartition = &partitions[i]
			break
		}
	}

	if targetPartition == nil {
		return 0, fmt.Errorf("partition with UUID %s not found on disk", partitionUUID)
	}

	// Create a new locator with the offset of the target partition
	partitionLocator, err := NewAPFSJumpstartLocator(
		l.reader,
		l.blockSize,
		targetPartition.Offset,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create partition-specific locator: %w", err)
	}

	// Find jumpstart in this specific partition
	jumpstartAddr, err := partitionLocator.FindEFIJumpstart()
	if err != nil {
		return 0, fmt.Errorf("failed to find jumpstart in partition %s: %w", partitionUUID, err)
	}

	return jumpstartAddr, nil
}

// PartitionInfo represents basic partition information
type PartitionInfo struct {
	UUID   string
	Name   string
	Offset int64
	Size   uint64
}

// getAllPartitions is a helper to get all partitions from the disk
func (l *APFSJumpstartLocator) getAllPartitions() ([]PartitionInfo, error) {
	// Use the GPT manager to get all partitions
	partManager, err := NewGPTPartitionManager(l.reader, uint64(l.blockSize), "")
	if err != nil {
		return nil, fmt.Errorf("failed to create partition manager: %w", err)
	}

	// For simplicity, we'll just get EFI partitions here, but in a real implementation
	// you'd want to get all partitions including APFS ones
	efiParts, err := partManager.ListEFIPartitions()
	if err != nil {
		return nil, fmt.Errorf("failed to list partitions: %w", err)
	}

	// Convert to our internal type
	partitions := make([]PartitionInfo, len(efiParts))
	for i, part := range efiParts {
		partitions[i] = PartitionInfo{
			UUID:   part.UUID,
			Name:   part.Name,
			Offset: int64(part.Offset),
			Size:   part.Size,
		}
	}

	return partitions, nil
}

// readContainerSuperblock reads and parses the APFS container superblock.
func (l *APFSJumpstartLocator) readContainerSuperblock() (*types.NxSuperblockT, error) {
	// Container superblock is at block 0 of the APFS container
	offset := l.partitionOffset + containerSuperblockOffset

	// Read enough data for the superblock
	bufSize := 1536 // Size of the complete nx_superblock_t structure
	buf := make([]byte, bufSize)

	// Read the superblock data
	n, err := l.reader.ReadAt(buf, offset)
	if err != nil && (err != io.EOF || n == 0) {
		return nil, fmt.Errorf("failed to read container superblock: %w", err)
	}

	if n < 256 { // At minimum need the essential fields
		return nil, fmt.Errorf("short read for container superblock: got %d bytes, need at least 256", n)
	}

	// Use the existing container superblock reader for proper parsing
	csr, err := container.NewContainerSuperblockReader(buf, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to create container superblock reader: %w", err)
	}

	// Extract the parsed superblock data
	// We need to construct a NxSuperblockT from the reader's methods
	superblock := &types.NxSuperblockT{
		NxMagic:          csr.Magic(),
		NxBlockSize:      csr.BlockSize(),
		NxBlockCount:     csr.BlockCount(),
		NxUuid:           csr.UUID(),
		NxNextOid:        csr.NextObjectID(),
		NxNextXid:        csr.NextTransactionID(),
		NxSpacemanOid:    csr.SpaceManagerOID(),
		NxOmapOid:        csr.ObjectMapOID(),
		NxReaperOid:      csr.ReaperOID(),
		NxMaxFileSystems: csr.MaxFileSystems(),
		NxEfiJumpstart:   csr.EFIJumpstart(),
		NxFusionUuid:     csr.FusionUUID(),
		NxKeylocker:      csr.KeylockerLocation(),
		NxMkbLocker:      csr.MediaKeyLocation(),
	}

	// Copy volume OIDs
	volumeOIDs := csr.VolumeOIDs()
	for i, oid := range volumeOIDs {
		if i < types.NxMaxFileSystems {
			superblock.NxFsOid[i] = oid
		}
	}

	// For checkpoint-related fields, we need to parse them manually since the container reader doesn't expose them
	// Parse the essential checkpoint fields we need
	if n >= 152 { // Ensure we have enough data for checkpoint fields
		superblock.NxXpDescBlocks = binary.LittleEndian.Uint32(buf[104:108])
		superblock.NxXpDescBase = types.Paddr(binary.LittleEndian.Uint64(buf[112:120]))
		superblock.NxXpDescNext = binary.LittleEndian.Uint32(buf[128:132])
		superblock.NxXpDescIndex = binary.LittleEndian.Uint32(buf[136:140])
	}

	return superblock, nil
}

// findJumpstartFromCheckpoints attempts to find the EFI jumpstart address by parsing
// checkpoint data in the APFS container.
func (l *APFSJumpstartLocator) findJumpstartFromCheckpoints(superblock *types.NxSuperblockT) (types.Paddr, error) {
	// Determine the base address of the checkpoint descriptor area
	var checkpointDescBase types.Paddr
	if superblock.NxXpDescBase != 0 {
		checkpointDescBase = superblock.NxXpDescBase
	} else {
		// Default to typical location (fallback)
		checkpointDescBase = checkpointDescAreaStartOffset
	}

	// Determine the number of blocks in the checkpoint descriptor area
	var checkpointDescBlocks uint32
	if superblock.NxXpDescBlocks > 0 {
		// Remove the highest bit which is used as a flag
		// Refer to APFS spec: "The highest bit of this number is used as a flag"
		checkpointDescBlocks = superblock.NxXpDescBlocks & 0x7FFFFFFF
	} else {
		// Default to a reasonable search range
		checkpointDescBlocks = 64 // Conservative default
	}

	// Find the most recent valid checkpoint
	checkpointMapPhys, checkpointIndex, err := l.findLatestValidCheckpoint(checkpointDescBase, checkpointDescBlocks)
	if err != nil {
		// Only fall back to signature scanning for "no valid checkpoint map found" errors
		// Preserve other errors like read failures or invalid indices
		if strings.Contains(err.Error(), "no valid checkpoint map found") {
			jumpstartAddr, scanErr := l.scanForJumpstartSignature(checkpointDescBase, checkpointDescBlocks)
			if scanErr != nil {
				return 0, fmt.Errorf("failed to find valid checkpoint: %w, and signature scan failed: %v", err, scanErr)
			}
			return jumpstartAddr, nil
		}
		// For other errors (read failures, invalid indices), don't fall back to signature scanning
		return 0, fmt.Errorf("failed to find valid checkpoint: %w", err)
	}
	if checkpointMapPhys == nil {
		// If no checkpoint map found, fall back to signature scanning
		jumpstartAddr, err := l.scanForJumpstartSignature(checkpointDescBase, checkpointDescBlocks)
		if err != nil {
			return 0, fmt.Errorf("no valid checkpoint found and signature scan failed: %w", err)
		}
		return jumpstartAddr, nil
	}

	// Search through checkpoint mappings for EFI jumpstart object
	jumpstartAddr, err := l.findJumpstartInCheckpointMap(checkpointMapPhys, checkpointIndex)
	if err != nil {
		return 0, fmt.Errorf("failed to find jumpstart in checkpoint map: %w", err)
	}

	if jumpstartAddr == 0 {
		// If not found in checkpoint mappings, fall back to signature scanning
		// This is less reliable but provides a fallback mechanism
		jumpstartAddr, err = l.scanForJumpstartSignature(checkpointDescBase, checkpointDescBlocks)
		if err != nil {
			return 0, fmt.Errorf("failed to scan for jumpstart signature: %w", err)
		}
	}

	return jumpstartAddr, nil
}

// findLatestValidCheckpoint locates and reads the most recent valid checkpoint in the descriptor area
func (l *APFSJumpstartLocator) findLatestValidCheckpoint(descBase types.Paddr, descBlocks uint32) (*types.CheckpointMapPhysT, uint32, error) {
	// We need to find the latest valid checkpoint in the checkpoint descriptor area
	// The checkpoint descriptor area contains a circular buffer of checkpoint mappings

	// First, get the checkpoint descriptor indexes from the superblock
	checkpointDescIndex := uint32(0)
	checkpointDescNext := uint32(0)

	// Read the container superblock again to get latest checkpoint indexes
	// (In a more efficient implementation, we'd pass these from the caller)
	superblock, err := l.readContainerSuperblock()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read container superblock for checkpoint info: %w", err)
	}

	// Get checkpoint index values
	checkpointDescIndex = superblock.NxXpDescIndex
	checkpointDescNext = superblock.NxXpDescNext

	// Validate the index values
	if checkpointDescIndex >= descBlocks || checkpointDescNext >= descBlocks {
		return nil, 0, fmt.Errorf("invalid checkpoint index values: index=%d, next=%d, blocks=%d",
			checkpointDescIndex, checkpointDescNext, descBlocks)
	}

	// Find the most recent checkpoint index
	var latestIndex uint32
	if checkpointDescNext == 0 {
		// If next is 0, the latest is right before it (circular buffer)
		latestIndex = descBlocks - 1
	} else {
		// Otherwise, the latest is right before next
		latestIndex = checkpointDescNext - 1
	}

	// Read the checkpoint map at the latest index
	checkpointMapPhys, err := l.readCheckpointMap(descBase, latestIndex)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read checkpoint map at index %d: %w", latestIndex, err)
	}

	// Validate the checkpoint map
	if !l.isValidCheckpointMap(checkpointMapPhys) {
		// If the latest isn't valid, try the one the superblock points to
		checkpointMapPhys, err = l.readCheckpointMap(descBase, checkpointDescIndex)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to read fallback checkpoint map at index %d: %w", checkpointDescIndex, err)
		}

		if !l.isValidCheckpointMap(checkpointMapPhys) {
			// Last resort: scan for a valid checkpoint
			for i := uint32(0); i < descBlocks; i++ {
				checkpointMapPhys, err = l.readCheckpointMap(descBase, i)
				if err == nil && l.isValidCheckpointMap(checkpointMapPhys) {
					return checkpointMapPhys, i, nil
				}
			}
			return nil, 0, fmt.Errorf("no valid checkpoint map found after full scan")
		}

		return checkpointMapPhys, checkpointDescIndex, nil
	}

	return checkpointMapPhys, latestIndex, nil
}

// readCheckpointMap reads a checkpoint map from the specified base address and index
func (l *APFSJumpstartLocator) readCheckpointMap(descBase types.Paddr, index uint32) (*types.CheckpointMapPhysT, error) {
	// Calculate the offset for this checkpoint map
	offset := l.partitionOffset + int64(descBase)*int64(l.blockSize) + int64(index)*int64(l.blockSize)

	// Read one block for the checkpoint map header and initial mappings
	buf := make([]byte, l.blockSize)
	n, err := l.reader.ReadAt(buf, offset)
	if err != nil && (err != io.EOF || n == 0) {
		return nil, fmt.Errorf("failed to read checkpoint map block: %w", err)
	}

	if n < 40 { // Minimum size needed: ObjPhysT (32) + flags (4) + count (4)
		return nil, fmt.Errorf("short read for checkpoint map: got %d bytes, need at least 40", n)
	}

	var checkpointMap types.CheckpointMapPhysT

	// Parse ObjPhysT header manually (32 bytes)
	copy(checkpointMap.CpmO.OChecksum[:], buf[0:8])
	checkpointMap.CpmO.OOid = types.OidT(binary.LittleEndian.Uint64(buf[8:16]))
	checkpointMap.CpmO.OXid = types.XidT(binary.LittleEndian.Uint64(buf[16:24]))
	checkpointMap.CpmO.OType = binary.LittleEndian.Uint32(buf[24:28])
	checkpointMap.CpmO.OSubtype = binary.LittleEndian.Uint32(buf[28:32])

	// Parse flags and count at correct offsets
	checkpointMap.CpmFlags = binary.LittleEndian.Uint32(buf[32:36])
	checkpointMap.CpmCount = binary.LittleEndian.Uint32(buf[36:40])

	// Validate the count to avoid allocation issues
	if checkpointMap.CpmCount > 1000 { // Arbitrary reasonable limit
		return nil, fmt.Errorf("checkpoint map count too large: %d", checkpointMap.CpmCount)
	}

	// Allocate and read the mapping entries
	checkpointMap.CpmMap = make([]types.CheckpointMappingT, checkpointMap.CpmCount)

	// Calculate size of all mappings
	mappingSize := int(checkpointMap.CpmCount) * 48 // Each mapping is 48 bytes

	// If mappings don't fit in the initial buffer, read more data
	if 40+mappingSize > n {
		// Need to read more data to get all mappings
		extendedBuf := make([]byte, 40+mappingSize)
		copy(extendedBuf[:n], buf) // Copy what we've already read

		// Read the rest of the mappings
		_, err := l.reader.ReadAt(extendedBuf[n:], offset+int64(n))
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read extended checkpoint map data: %w", err)
		}

		buf = extendedBuf // Use the extended buffer
	}

	// Parse all mapping entries manually starting at offset 40
	mappingOffset := 40
	for i := uint32(0); i < checkpointMap.CpmCount; i++ {
		if mappingOffset+48 > len(buf) {
			return nil, fmt.Errorf("insufficient data for checkpoint mapping %d", i)
		}

		mapping := &checkpointMap.CpmMap[i]
		mapping.CpmType = binary.LittleEndian.Uint32(buf[mappingOffset : mappingOffset+4])
		mapping.CpmSubtype = binary.LittleEndian.Uint32(buf[mappingOffset+4 : mappingOffset+8])
		mapping.CpmSize = binary.LittleEndian.Uint32(buf[mappingOffset+8 : mappingOffset+12])
		mapping.CpmPad = binary.LittleEndian.Uint32(buf[mappingOffset+12 : mappingOffset+16])
		mapping.CpmFsOid = types.OidT(binary.LittleEndian.Uint64(buf[mappingOffset+16 : mappingOffset+24]))
		mapping.CpmOid = types.OidT(binary.LittleEndian.Uint64(buf[mappingOffset+24 : mappingOffset+32]))
		mapping.CpmPaddr = types.Paddr(binary.LittleEndian.Uint64(buf[mappingOffset+32 : mappingOffset+40]))

		mappingOffset += 48
	}

	return &checkpointMap, nil
}

// isValidCheckpointMap checks if the checkpoint map is valid
func (l *APFSJumpstartLocator) isValidCheckpointMap(checkpointMap *types.CheckpointMapPhysT) bool {
	if checkpointMap == nil {
		return false
	}

	// Check if the object type is checkpoint map
	if checkpointMap.CpmO.OType&types.ObjectTypeMask != types.ObjectTypeCheckpointMap {
		return false
	}

	// Check if count is reasonable
	if checkpointMap.CpmCount == 0 || checkpointMap.CpmCount > 1000 {
		return false
	}

	// Check if we have mapping data
	if len(checkpointMap.CpmMap) != int(checkpointMap.CpmCount) {
		return false
	}

	return true
}

// findJumpstartInCheckpointMap searches for the EFI jumpstart object in a checkpoint map
func (l *APFSJumpstartLocator) findJumpstartInCheckpointMap(checkpointMap *types.CheckpointMapPhysT, checkpointIndex uint32) (types.Paddr, error) {
	if checkpointMap == nil {
		return 0, fmt.Errorf("checkpoint map is nil")
	}

	// Look for EFI jumpstart object type in the mappings
	for _, mapping := range checkpointMap.CpmMap {
		if mapping.CpmType&types.ObjectTypeMask == types.ObjectTypeEfiJumpstart {
			// Found a jumpstart object mapping
			return mapping.CpmPaddr, nil
		}
	}

	// If the checkpoint map flag indicates it's the last one, we're done
	if checkpointMap.CpmFlags&types.CheckpointMapLast != 0 {
		// This is the last checkpoint map, so no more to check
		return 0, nil
	}

	// Check if there might be another checkpoint map to check
	// Look for a mapping to another checkpoint map
	for _, mapping := range checkpointMap.CpmMap {
		if mapping.CpmType&types.ObjectTypeMask == types.ObjectTypeCheckpointMap {
			// Found another checkpoint map, recursively search it
			nextCheckpointMap, err := l.readCheckpointMapFromPaddr(mapping.CpmPaddr)
			if err != nil {
				return 0, fmt.Errorf("failed to read next checkpoint map: %w", err)
			}

			return l.findJumpstartInCheckpointMap(nextCheckpointMap, checkpointIndex)
		}
	}

	// No jumpstart found in this checkpoint map chain
	return 0, nil
}

// readCheckpointMapFromPaddr reads a checkpoint map from a physical address
func (l *APFSJumpstartLocator) readCheckpointMapFromPaddr(paddr types.Paddr) (*types.CheckpointMapPhysT, error) {
	offset := l.partitionOffset + int64(paddr)*int64(l.blockSize)

	// Read one block for the checkpoint map
	buf := make([]byte, l.blockSize)
	n, err := l.reader.ReadAt(buf, offset)
	if err != nil && (err != io.EOF || n == 0) {
		return nil, fmt.Errorf("failed to read checkpoint map from paddr %d: %w", paddr, err)
	}

	if n < 40 { // Minimum size needed: ObjPhysT (32) + flags (4) + count (4)
		return nil, fmt.Errorf("short read for checkpoint map: got %d bytes, need at least 40", n)
	}

	var checkpointMap types.CheckpointMapPhysT

	// Parse ObjPhysT header manually (32 bytes)
	copy(checkpointMap.CpmO.OChecksum[:], buf[0:8])
	checkpointMap.CpmO.OOid = types.OidT(binary.LittleEndian.Uint64(buf[8:16]))
	checkpointMap.CpmO.OXid = types.XidT(binary.LittleEndian.Uint64(buf[16:24]))
	checkpointMap.CpmO.OType = binary.LittleEndian.Uint32(buf[24:28])
	checkpointMap.CpmO.OSubtype = binary.LittleEndian.Uint32(buf[28:32])

	// Parse flags and count at correct offsets
	checkpointMap.CpmFlags = binary.LittleEndian.Uint32(buf[32:36])
	checkpointMap.CpmCount = binary.LittleEndian.Uint32(buf[36:40])

	// Validate the count
	if checkpointMap.CpmCount > 1000 { // Arbitrary reasonable limit
		return nil, fmt.Errorf("checkpoint map count too large: %d", checkpointMap.CpmCount)
	}

	// Allocate and read the mapping entries
	checkpointMap.CpmMap = make([]types.CheckpointMappingT, checkpointMap.CpmCount)

	// Parse all mapping entries manually starting at offset 40
	mappingOffset := 40
	for i := uint32(0); i < checkpointMap.CpmCount; i++ {
		if mappingOffset+48 > n {
			return nil, fmt.Errorf("insufficient data for checkpoint mapping %d", i)
		}

		mapping := &checkpointMap.CpmMap[i]
		mapping.CpmType = binary.LittleEndian.Uint32(buf[mappingOffset : mappingOffset+4])
		mapping.CpmSubtype = binary.LittleEndian.Uint32(buf[mappingOffset+4 : mappingOffset+8])
		mapping.CpmSize = binary.LittleEndian.Uint32(buf[mappingOffset+8 : mappingOffset+12])
		mapping.CpmPad = binary.LittleEndian.Uint32(buf[mappingOffset+12 : mappingOffset+16])
		mapping.CpmFsOid = types.OidT(binary.LittleEndian.Uint64(buf[mappingOffset+16 : mappingOffset+24]))
		mapping.CpmOid = types.OidT(binary.LittleEndian.Uint64(buf[mappingOffset+24 : mappingOffset+32]))
		mapping.CpmPaddr = types.Paddr(binary.LittleEndian.Uint64(buf[mappingOffset+32 : mappingOffset+40]))

		mappingOffset += 48
	}

	return &checkpointMap, nil
}

// scanForJumpstartSignature falls back to scanning for the jumpstart signature
// This is a last resort if the checkpoint map parsing fails
func (l *APFSJumpstartLocator) scanForJumpstartSignature(startBlock types.Paddr, blockCount uint32) (types.Paddr, error) {
	// Scan the entire container, not just the checkpoint area
	// Start from block 0 and scan up to a reasonable limit
	maxScanBlocks := uint32(1024) // Limit scan range for performance

	// Search for the jumpstart structure in blocks starting from block 0
	buf := make([]byte, l.blockSize)
	for blockAddr := types.Paddr(0); blockAddr < types.Paddr(maxScanBlocks); blockAddr++ {
		offset := l.partitionOffset + int64(blockAddr)*int64(l.blockSize)

		// Read the block
		n, err := l.reader.ReadAt(buf, offset)
		if err != nil && (err != io.EOF || n == 0) {
			// Skip blocks we can't read
			continue
		}

		// Look for jumpstart magic "JSDR" (little endian: 0x5244534A)
		for i := 0; i < n-4; i += 8 { // Assume 8-byte alignment
			magic := binary.LittleEndian.Uint32(buf[i:])
			if magic == types.NxEfiJumpstartMagic {
				// Found a potential jumpstart structure
				// Verify it by checking the version too
				if i+8 < n { // Make sure we can read the version
					version := binary.LittleEndian.Uint32(buf[i+4:])
					if version == types.NxEfiJumpstartVersion {
						// Found a valid jumpstart structure at this block
						return blockAddr, nil
					}
				}
			}
		}
	}

	// Jumpstart not found
	return 0, nil
}

// ReadJumpstartObjectFromPaddr reads and parses the NxEfiJumpstartT structure from the given
// physical address.
// This is a helper method to verify that we've found the correct jumpstart structure.
func (l *APFSJumpstartLocator) ReadJumpstartObjectFromPaddr(paddr types.Paddr) (*types.NxEfiJumpstartT, error) {
	offset := l.partitionOffset + int64(paddr)*int64(l.blockSize)

	// First read the basic jumpstart struct (without the variable extents)
	// Minimum size is ObjPhysT (8) + 4 header fields (16) + reserved (128) = 152 bytes
	minSize := 152
	buf := make([]byte, minSize)

	n, err := l.reader.ReadAt(buf, offset)
	if err != nil && (err != io.EOF || n < minSize) {
		return nil, fmt.Errorf("failed to read jumpstart structure: %w", err)
	}

	reader := bytes.NewReader(buf)

	// Skip ObjPhysT (obj_id + obj_type) which is 8 bytes
	if _, err := reader.Seek(8, io.SeekCurrent); err != nil {
		return nil, fmt.Errorf("failed to seek past ObjPhysT: %w", err)
	}

	var jumpstart types.NxEfiJumpstartT

	// Read magic
	if err := binary.Read(reader, binary.LittleEndian, &jumpstart.NejMagic); err != nil {
		return nil, fmt.Errorf("failed to read magic: %w", err)
	}

	// Read version
	if err := binary.Read(reader, binary.LittleEndian, &jumpstart.NejVersion); err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}

	// Read file length
	if err := binary.Read(reader, binary.LittleEndian, &jumpstart.NejEfiFileLen); err != nil {
		return nil, fmt.Errorf("failed to read file length: %w", err)
	}

	// Read extent count
	if err := binary.Read(reader, binary.LittleEndian, &jumpstart.NejNumExtents); err != nil {
		return nil, fmt.Errorf("failed to read extent count: %w", err)
	}

	// Skip reserved fields (16 * uint64 = 128 bytes)
	if _, err := reader.Seek(128, io.SeekCurrent); err != nil {
		return nil, fmt.Errorf("failed to seek past reserved fields: %w", err)
	}

	// Now read the extents - each is 16 bytes (paddr 8 + block_count 8)
	if jumpstart.NejNumExtents > 0 {
		extentSize := 16 // Size of each Prange struct
		extentsBufferSize := int(jumpstart.NejNumExtents) * extentSize

		// Read additional buffer for extents
		extentsBuf := make([]byte, extentsBufferSize)
		n, err := l.reader.ReadAt(extentsBuf, offset+int64(minSize))
		if err != nil && (err != io.EOF || n < extentsBufferSize) {
			return nil, fmt.Errorf("failed to read jumpstart extents: %w", err)
		}

		// Parse extents
		jumpstart.NejRecExtents = make([]types.Prange, jumpstart.NejNumExtents)
		extReader := bytes.NewReader(extentsBuf)

		for i := uint32(0); i < jumpstart.NejNumExtents; i++ {
			if err := binary.Read(extReader, binary.LittleEndian, &jumpstart.NejRecExtents[i].PrStartPaddr); err != nil {
				return nil, fmt.Errorf("failed to read extent %d start address: %w", i, err)
			}
			if err := binary.Read(extReader, binary.LittleEndian, &jumpstart.NejRecExtents[i].PrBlockCount); err != nil {
				return nil, fmt.Errorf("failed to read extent %d block count: %w", i, err)
			}
		}
	}

	// Verify the structure
	if jumpstart.NejMagic != types.NxEfiJumpstartMagic {
		return nil, fmt.Errorf("invalid jumpstart magic: %#x", jumpstart.NejMagic)
	}
	if jumpstart.NejVersion != types.NxEfiJumpstartVersion {
		return nil, fmt.Errorf("invalid jumpstart version: %d", jumpstart.NejVersion)
	}

	return &jumpstart, nil
}
