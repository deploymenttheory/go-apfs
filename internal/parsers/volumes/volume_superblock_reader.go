package volumes

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// volumeSuperblockReader parses and provides access to volume superblock data
type volumeSuperblockReader struct {
	superblock *types.ApfsSuperblockT
	endian     binary.ByteOrder
}

// NewVolumeSuperblockReader creates a new VolumeSuperblockReader from raw bytes
func NewVolumeSuperblockReader(data []byte, endian binary.ByteOrder) (*volumeSuperblockReader, error) {
	if endian == nil {
		endian = binary.LittleEndian
	}

	sb, err := parseVolumeSuperblock(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume superblock: %w", err)
	}

	return &volumeSuperblockReader{
		superblock: sb,
		endian:     endian,
	}, nil
}

// parseVolumeSuperblock parses raw bytes into an ApfsSuperblockT structure
func parseVolumeSuperblock(data []byte, endian binary.ByteOrder) (*types.ApfsSuperblockT, error) {
	if len(data) < 1024 {
		return nil, fmt.Errorf("insufficient data for volume superblock: need at least 1024 bytes, got %d", len(data))
	}

	sb := &types.ApfsSuperblockT{}

	// Parse object header (first 32 bytes)
	copy(sb.ApfsO.OChecksum[:], data[0:8])
	sb.ApfsO.OOid = types.OidT(endian.Uint64(data[8:16]))
	sb.ApfsO.OXid = types.XidT(endian.Uint64(data[16:24]))
	sb.ApfsO.OType = endian.Uint32(data[24:28])
	sb.ApfsO.OSubtype = endian.Uint32(data[28:32])

	offset := 32

	// Parse magic (validation) - magic is always stored in big-endian format as ASCII bytes
	sb.ApfsMagic = binary.BigEndian.Uint32(data[offset : offset+4])
	if sb.ApfsMagic != types.ApfsMagic {
		return nil, fmt.Errorf("invalid volume superblock magic: got 0x%08X, want 0x%08X", sb.ApfsMagic, types.ApfsMagic)
	}
	offset += 4

	// Parse filesystem index
	sb.ApfsFsIndex = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse feature flags
	sb.ApfsFeatures = endian.Uint64(data[offset : offset+8])
	offset += 8

	sb.ApfsReadonlyCompatibleFeatures = endian.Uint64(data[offset : offset+8])
	offset += 8

	sb.ApfsIncompatibleFeatures = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse unmount time
	sb.ApfsUnmountTime = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse space management fields
	sb.ApfsFsReserveBlockCount = endian.Uint64(data[offset : offset+8])
	offset += 8

	sb.ApfsFsQuotaBlockCount = endian.Uint64(data[offset : offset+8])
	offset += 8

	sb.ApfsFsAllocCount = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse metadata crypto structure (20 bytes, not 112!)
	// struct wrapped_meta_crypto_state { 
	//   uint16_t major_version; uint16_t minor_version; uint32_t cpflags;
	//   uint32_t persistent_class; uint32_t key_os_version; uint16_t key_revision; uint16_t unused;
	// }
	fmt.Printf("DEBUG: Parsing crypto structure at offset %d\n", offset)
	offset += 20 // Skip wrapped_meta_crypto_state_t (20 bytes, not 112)

	// Parse tree types
	sb.ApfsRootTreeType = endian.Uint32(data[offset : offset+4])
	offset += 4

	sb.ApfsExtentreftreeType = endian.Uint32(data[offset : offset+4])
	offset += 4

	sb.ApfsSnapMetatreeType = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse OIDs at their correct locations according to APFS spec
	fmt.Printf("DEBUG: Parsing OIDs starting from offset %d\n", offset)
	
	sb.ApfsOmapOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	fmt.Printf("DEBUG: ApfsOmapOid at offset %d = %d\n", offset, sb.ApfsOmapOid)
	offset += 8

	sb.ApfsRootTreeOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	fmt.Printf("DEBUG: ApfsRootTreeOid at offset %d = %d\n", offset, sb.ApfsRootTreeOid)
	offset += 8

	sb.ApfsExtentrefTreeOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	fmt.Printf("DEBUG: ApfsExtentrefTreeOid at offset %d = %d\n", offset, sb.ApfsExtentrefTreeOid)
	offset += 8

	sb.ApfsSnapMetaTreeOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	fmt.Printf("DEBUG: ApfsSnapMetaTreeOid at offset %d = %d\n", offset, sb.ApfsSnapMetaTreeOid)
	offset += 8

	// Parse revert fields
	sb.ApfsRevertToXid = types.XidT(endian.Uint64(data[offset : offset+8]))
	offset += 8

	sb.ApfsRevertToSblockOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8

	// Parse next object ID
	sb.ApfsNextObjId = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse file/directory/symlink counts
	sb.ApfsNumFiles = endian.Uint64(data[offset : offset+8])
	fmt.Printf("DEBUG: ApfsNumFiles at offset %d = %d\n", offset, sb.ApfsNumFiles)
	offset += 8

	sb.ApfsNumDirectories = endian.Uint64(data[offset : offset+8])
	fmt.Printf("DEBUG: ApfsNumDirectories at offset %d = %d\n", offset, sb.ApfsNumDirectories)
	offset += 8

	sb.ApfsNumSymlinks = endian.Uint64(data[offset : offset+8])
	offset += 8

	sb.ApfsNumOtherFsobjects = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse snapshot count
	sb.ApfsNumSnapshots = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse total blocks allocated/freed
	sb.ApfsTotalBlocksAlloced = endian.Uint64(data[offset : offset+8])
	offset += 8

	sb.ApfsTotalBlocksFreed = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse UUID
	copy(sb.ApfsVolUuid[:], data[offset:offset+16])
	offset += 16

	// Parse last modification time
	sb.ApfsLastModTime = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse filesystem flags
	sb.ApfsFsFlags = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse formatted by timestamp (ApfsModifiedByT structure - 8 bytes)
	offset += 8

	// Parse modification history - array of ApfsModifiedByT (APFS_MAX_HIST = 8 entries, 8 bytes each)
	offset += types.ApfsMaxHist * 8

	// Parse volume name
	if offset+types.ApfsVolnameLen <= len(data) {
		copy(sb.ApfsVolname[:], data[offset:offset+types.ApfsVolnameLen])
		offset += types.ApfsVolnameLen
	}

	// Parse next document ID
	if offset+4 <= len(data) {
		sb.ApfsNextDocId = endian.Uint32(data[offset : offset+4])
		offset += 4
	}

	// Parse additional fields if present
	if offset+2 <= len(data) {
		sb.ApfsRole = endian.Uint16(data[offset : offset+2])
		offset += 2
	}

	return sb, nil
}

// GetSuperblock returns the parsed superblock
func (vsr *volumeSuperblockReader) GetSuperblock() *types.ApfsSuperblockT {
	return vsr.superblock
}

// BlockSize returns the block size (default APFS block size)
func (vsr *volumeSuperblockReader) BlockSize() uint32 {
	return 4096
}

// Validate checks if the superblock is valid
func (vsr *volumeSuperblockReader) Validate() (bool, []string) {
	issues := []string{}

	// Check magic number
	if vsr.superblock.ApfsMagic != types.ApfsMagic {
		issues = append(issues, fmt.Sprintf("invalid magic number: 0x%08X", vsr.superblock.ApfsMagic))
	}

	// Check filesystem index is reasonable
	if vsr.superblock.ApfsFsIndex > 100 {
		issues = append(issues, fmt.Sprintf("filesystem index unreasonably high: %d", vsr.superblock.ApfsFsIndex))
	}

	// Check quota is not less than allocated
	if vsr.superblock.ApfsFsQuotaBlockCount > 0 && vsr.superblock.ApfsFsAllocCount > vsr.superblock.ApfsFsQuotaBlockCount {
		issues = append(issues, fmt.Sprintf("allocated blocks (%d) exceed quota (%d)",
			vsr.superblock.ApfsFsAllocCount, vsr.superblock.ApfsFsQuotaBlockCount))
	}

	return len(issues) == 0, issues
}
