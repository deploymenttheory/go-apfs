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
// 3. Read entries in checkpoint descriptor area (checkpoint_map_phys_t or nx_superblock_t)
// 4. Find container superblock with largest XID that isn't malformed
func (cds *CheckpointDiscoveryService) FindLatestValidSuperblock() (*CheckpointCandidate, error) {
	// Step 1: Read block zero superblock (this might be stale)
	blockZeroSB := cds.container.GetSuperblock()
	if blockZeroSB == nil {
		return nil, fmt.Errorf("no container superblock available at block zero")
	}

	fmt.Printf("DEBUG: Block zero superblock - XID=%d, DescBase=%d, DescBlocks=%d\n", 
		blockZeroSB.NxNextXid, blockZeroSB.NxXpDescBase, blockZeroSB.NxXpDescBlocks)

	// Step 2: Use nx_xp_desc_base to locate checkpoint descriptor area
	descBase := uint64(blockZeroSB.NxXpDescBase)
	descBlocks := blockZeroSB.NxXpDescBlocks & 0x7FFFFFFF // Clear high bit flag
	
	fmt.Printf("DEBUG: Scanning checkpoint descriptor area at block %d, %d blocks\n", descBase, descBlocks)

	// Step 3: Read entries in checkpoint descriptor area
	candidates := []*CheckpointCandidate{}
	
	// Add block zero candidate first
	blockZeroCandidate := &CheckpointCandidate{
		Superblock:   blockZeroSB,
		TransactionID: blockZeroSB.NxNextXid,
		BlockAddress:  0,
		IsValid:      cds.validateSuperblock(blockZeroSB),
		ErrorMsg:     "",
	}
	if !blockZeroCandidate.IsValid {
		blockZeroCandidate.ErrorMsg = "block zero superblock validation failed"
	}
	candidates = append(candidates, blockZeroCandidate)

	// Scan checkpoint descriptor area for additional superblocks
	for blockOffset := uint64(0); blockOffset < uint64(descBlocks); blockOffset++ {
		blockAddr := descBase + blockOffset
		
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
					Superblock:   nil,
					TransactionID: 0,
					BlockAddress:  blockAddr,
					IsValid:      false,
					ErrorMsg:     fmt.Sprintf("parse error: %v", err),
				}
				candidates = append(candidates, candidate)
				continue
			}

			isValid := cds.validateSuperblock(sb)
			candidate := &CheckpointCandidate{
				Superblock:   sb,
				TransactionID: sb.NxNextXid,
				BlockAddress:  blockAddr,
				IsValid:      isValid,
				ErrorMsg:     "",
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