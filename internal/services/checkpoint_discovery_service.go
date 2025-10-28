package services

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// CheckpointDiscoveryService handles finding the latest valid container superblock
// from the checkpoint descriptor area, following the Apple APFS mounting procedure
type CheckpointDiscoveryService struct {
	container *ContainerReader
}

// NewCheckpointDiscoveryService creates a new checkpoint discovery service
func NewCheckpointDiscoveryService(container *ContainerReader) *CheckpointDiscoveryService {
	return &CheckpointDiscoveryService{
		container: container,
	}
}

// CheckpointCandidate represents a potential checkpoint superblock
type CheckpointCandidate struct {
	Superblock   *types.NxSuperblockT
	TransactionID types.XidT
	BlockAddress  uint64
	IsValid      bool
	ErrorMsg     string
}

// FindLatestValidSuperblock implements the Apple APFS mounting procedure:
// 1. Read block zero to get initial superblock
// 2. Use nx_xp_desc_base to locate checkpoint descriptor area
//    - If MSB is clear: descriptor area is contiguous (simple case)
//    - If MSB is set: descriptor area is non-contiguous, requires B-tree parsing
// 3. Read entries in checkpoint descriptor area (checkpoint_map_phys_t or nx_superblock_t)
// 4. Find container superblock with largest XID that isn't malformed
func (cds *CheckpointDiscoveryService) FindLatestValidSuperblock() (*CheckpointCandidate, error) {
	// Step 1: Read block zero superblock (this might be stale)
	blockZeroSB := cds.container.GetSuperblock()
	if blockZeroSB == nil {
		return nil, fmt.Errorf("no container superblock available at block zero")
	}

	fmt.Printf("DEBUG: Block zero superblock - XID=%d, DescBase=0x%X, DescBlocks=%d\n",
		blockZeroSB.NxNextXid, blockZeroSB.NxXpDescBase, blockZeroSB.NxXpDescBlocks)

	// Step 2: Use nx_xp_desc_base to locate checkpoint descriptor area
	// Check MSB of nx_xp_desc_base to determine if descriptor area is contiguous
	descBaseRaw := uint64(blockZeroSB.NxXpDescBase)
	descBlocks := blockZeroSB.NxXpDescBlocks & 0x7FFFFFFF // Clear high bit flag

	const msbMask uint64 = 0x8000000000000000
	isNonContiguous := (descBaseRaw & msbMask) != 0

	var blockRanges []blockRange
	var err error

	if isNonContiguous {
		// MSB is set: LSBs contain B-Tree root node address
		// The B-Tree maps logical offsets to physical block ranges (prange_t)
		btreeRootAddr := descBaseRaw & ^msbMask // Clear MSB to get address
		fmt.Printf("DEBUG: Non-contiguous descriptor area - B-tree root at block %d\n", btreeRootAddr)

		blockRanges, err = cds.parseDescriptorAreaBTree(btreeRootAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse non-contiguous descriptor area B-tree: %w", err)
		}
		fmt.Printf("DEBUG: Found %d non-contiguous ranges for descriptor area\n", len(blockRanges))
	} else {
		// MSB is clear: descriptor area is contiguous
		descBase := descBaseRaw
		fmt.Printf("DEBUG: Contiguous descriptor area at block %d, %d blocks\n", descBase, descBlocks)
		blockRanges = []blockRange{{start: descBase, count: uint64(descBlocks)}}
	}

	// Step 3: Read entries in checkpoint descriptor area
	candidates := []*CheckpointCandidate{}

	// Add block zero candidate first
	blockZeroCandidate := &CheckpointCandidate{
		Superblock:    blockZeroSB,
		TransactionID: blockZeroSB.NxNextXid,
		BlockAddress:  0,
		IsValid:       cds.validateSuperblock(blockZeroSB),
		ErrorMsg:      "",
	}
	if !blockZeroCandidate.IsValid {
		blockZeroCandidate.ErrorMsg = "block zero superblock validation failed"
	}
	candidates = append(candidates, blockZeroCandidate)

	// Scan checkpoint descriptor area for additional superblocks
	// Iterate through all block ranges (contiguous or non-contiguous)
	for _, brange := range blockRanges {
		for blockOffset := uint64(0); blockOffset < brange.count; blockOffset++ {
			blockAddr := brange.start + blockOffset

			blockData, err := cds.container.ReadBlock(blockAddr)
			if err != nil {
				fmt.Printf("DEBUG: Failed to read checkpoint block %d: %v\n", blockAddr, err)
				continue
			}

			if len(blockData) < 32 {
				continue
			}

			// Check if this block contains a container superblock
			if cds.isSuperblock(blockData) {
				sb, err := cds.parseSuperblock(blockData)
				if err != nil {
					fmt.Printf("DEBUG: Failed to parse superblock at block %d: %v\n", blockAddr, err)
					candidate := &CheckpointCandidate{
						Superblock:    nil,
						TransactionID: 0,
						BlockAddress:  blockAddr,
						IsValid:       false,
						ErrorMsg:      fmt.Sprintf("parse error: %v", err),
					}
					candidates = append(candidates, candidate)
					continue
				}

				isValid := cds.validateSuperblock(sb)
				candidate := &CheckpointCandidate{
					Superblock:    sb,
					TransactionID: sb.NxNextXid,
					BlockAddress:  blockAddr,
					IsValid:       isValid,
					ErrorMsg:      "",
				}
				if !isValid {
					candidate.ErrorMsg = "superblock validation failed"
				}

				fmt.Printf("DEBUG: Found superblock at block %d - XID=%d, Valid=%t\n",
					blockAddr, sb.NxNextXid, isValid)
				candidates = append(candidates, candidate)
			} else {
				// This might be a checkpoint_map_phys_t - we could parse these too
				// but for now we're focused on finding superblocks
				fmt.Printf("DEBUG: Block %d is not a superblock (type=0x%08X)\n",
					blockAddr, binary.LittleEndian.Uint32(blockData[24:28]))
			}
		}
	}

	// Step 4: Find the superblock with the largest XID that is valid
	var latestValid *CheckpointCandidate
	for _, candidate := range candidates {
		if candidate.IsValid && candidate.Superblock != nil {
			if latestValid == nil || candidate.TransactionID > latestValid.TransactionID {
				latestValid = candidate
			}
		}
	}

	if latestValid == nil {
		return nil, fmt.Errorf("no valid superblocks found in checkpoint area")
	}

	fmt.Printf("DEBUG: Selected superblock at block %d with XID=%d\n", 
		latestValid.BlockAddress, latestValid.TransactionID)

	return latestValid, nil
}

// isSuperblock checks if the data contains a container superblock
func (cds *CheckpointDiscoveryService) isSuperblock(data []byte) bool {
	if len(data) < 32 {
		return false
	}

	// Check magic number at offset 24 (in obj_phys_t header)
	objType := binary.LittleEndian.Uint32(data[24:28])
	
	// Container superblock type
	return objType == types.ObjectTypeNxSuperblock
}

// parseSuperblock parses a container superblock from raw data
func (cds *CheckpointDiscoveryService) parseSuperblock(data []byte) (*types.NxSuperblockT, error) {
	// Parse the superblock directly from raw data
	sb := &types.NxSuperblockT{}
	
	// Parse essential fields directly from the data
	if len(data) < 200 { // Need enough data for all fields
		return nil, fmt.Errorf("insufficient data for container superblock")
	}
	
	// Parse object header (32 bytes)
	copy(sb.NxO.OChecksum[:], data[0:8])
	sb.NxO.OOid = types.OidT(binary.LittleEndian.Uint64(data[8:16]))
	sb.NxO.OXid = types.XidT(binary.LittleEndian.Uint64(data[16:24]))
	sb.NxO.OType = binary.LittleEndian.Uint32(data[24:28])
	sb.NxO.OSubtype = binary.LittleEndian.Uint32(data[28:32])

	// Parse superblock specific fields (starting at offset 32)
	sb.NxMagic = binary.LittleEndian.Uint32(data[32:36])
	sb.NxBlockSize = binary.LittleEndian.Uint32(data[36:40])
	sb.NxBlockCount = binary.LittleEndian.Uint64(data[40:48])
	sb.NxFeatures = binary.LittleEndian.Uint64(data[48:56])
	sb.NxReadonlyCompatibleFeatures = binary.LittleEndian.Uint64(data[56:64])
	sb.NxIncompatibleFeatures = binary.LittleEndian.Uint64(data[64:72])
	copy(sb.NxUuid[:], data[72:88])
	sb.NxNextOid = types.OidT(binary.LittleEndian.Uint64(data[88:96]))
	sb.NxNextXid = types.XidT(binary.LittleEndian.Uint64(data[96:104]))
	sb.NxXpDescBlocks = binary.LittleEndian.Uint32(data[104:108])
	sb.NxXpDataBlocks = binary.LittleEndian.Uint32(data[108:112])
	sb.NxXpDescBase = types.Paddr(binary.LittleEndian.Uint64(data[112:120]))
	sb.NxXpDataBase = types.Paddr(binary.LittleEndian.Uint64(data[120:128]))
	sb.NxXpDescNext = binary.LittleEndian.Uint32(data[128:132])
	sb.NxXpDataNext = binary.LittleEndian.Uint32(data[132:136])
	sb.NxXpDescIndex = binary.LittleEndian.Uint32(data[136:140])
	sb.NxXpDescLen = binary.LittleEndian.Uint32(data[140:144])
	sb.NxXpDataIndex = binary.LittleEndian.Uint32(data[144:148])
	sb.NxXpDataLen = binary.LittleEndian.Uint32(data[148:152])
	sb.NxSpacemanOid = types.OidT(binary.LittleEndian.Uint64(data[152:160]))
	sb.NxOmapOid = types.OidT(binary.LittleEndian.Uint64(data[160:168]))
	sb.NxReaperOid = types.OidT(binary.LittleEndian.Uint64(data[168:176]))
	sb.NxTestType = binary.LittleEndian.Uint32(data[176:180])
	sb.NxMaxFileSystems = binary.LittleEndian.Uint32(data[180:184])
	
	// Parse filesystem OIDs array (array of 100 OIDs starting at offset 184)
	for i := 0; i < 100 && 184+i*8+8 <= len(data); i++ {
		offset := 184 + i*8
		sb.NxFsOid[i] = types.OidT(binary.LittleEndian.Uint64(data[offset:offset+8]))
	}

	return sb, nil
}

// validateSuperblock performs basic validation of a container superblock
func (cds *CheckpointDiscoveryService) validateSuperblock(sb *types.NxSuperblockT) bool {
	if sb == nil {
		return false
	}

	// Check magic number
	if sb.NxMagic != types.NxMagic {
		return false
	}

	// Check that essential OIDs are not zero
	if sb.NxOmapOid == 0 {
		return false
	}

	// Check that block size is reasonable (4KB is typical)
	if sb.NxBlockSize == 0 || sb.NxBlockSize > 65536 {
		return false
	}

	// Check that transaction ID is reasonable (not zero)
	if sb.NxNextXid == 0 {
		return false
	}

	return true
}

// GetCandidates returns all checkpoint candidates found during discovery
func (cds *CheckpointDiscoveryService) GetCandidates() ([]*CheckpointCandidate, error) {
	// This is a simplified version that just calls FindLatestValidSuperblock
	// In a full implementation, we might want to return all candidates
	latest, err := cds.FindLatestValidSuperblock()
	if err != nil {
		return nil, err
	}

	return []*CheckpointCandidate{latest}, nil
}

// blockRange represents a contiguous range of blocks
type blockRange struct {
	start uint64 // Starting block address
	count uint64 // Number of blocks in range
}

// parseDescriptorAreaBTree parses the B-tree that maps logical offsets to physical block ranges
// This B-tree is used when the checkpoint descriptor area is non-contiguous (MSB of nx_xp_desc_base is set)
// The B-tree maps uint64_t logical offsets to prange_t physical block ranges
func (cds *CheckpointDiscoveryService) parseDescriptorAreaBTree(btreeRootAddr uint64) ([]blockRange, error) {
	// Read the B-tree root node
	rootData, err := cds.container.ReadBlock(btreeRootAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to read B-tree root node at block %d: %w", btreeRootAddr, err)
	}

	// For a complete implementation, we would need to:
	// 1. Parse the btree_node_phys_t structure from rootData
	// 2. Read the btree_info_t from the end of the root node
	// 3. Traverse the B-tree to enumerate all key/value pairs
	// 4. Each key is a uint64_t logical offset
	// 5. Each value is a prange_t (pr_start_paddr, pr_block_count)
	// 6. Sort by logical offset and return the ranges in order

	// This is a simplified implementation that assumes common cases
	// A full implementation would require using the btree service
	if len(rootData) < 56 {
		return nil, fmt.Errorf("B-tree root node too small: %d bytes", len(rootData))
	}

	// Check object type (should be BTREE or BTREE_NODE)
	objType := binary.LittleEndian.Uint32(rootData[24:28])
	if objType != types.ObjectTypeBtree && objType != types.ObjectTypeBtreeNode {
		return nil, fmt.Errorf("invalid B-tree root node type: 0x%08X", objType)
	}

	// For now, return an error indicating this feature needs full implementation
	// In production, this would parse the B-tree using the existing btree service
	return nil, fmt.Errorf("non-contiguous checkpoint descriptor areas require full B-tree parsing (not yet implemented)")

	// TODO: Implement full B-tree traversal:
	// 1. Parse btree_node_phys_t header
	// 2. Read table of contents (kvloc_t or kvoff_t based on BTNODE_FIXED_KV_SIZE flag)
	// 3. For each entry:
	//    - Read key (uint64_t logical offset)
	//    - Read value (prange_t: start address + block count)
	// 4. If non-leaf node (btn_level > 0), recursively traverse children
	// 5. Collect all ranges and return in sorted order
}