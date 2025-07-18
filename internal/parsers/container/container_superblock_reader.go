package container

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// ContainerSuperblockReader implements the ContainerSuperblockReader interface
type ContainerSuperblockReader struct {
	Superblock *types.NxSuperblockT
	data       []byte
	endian     binary.ByteOrder
}

// NewContainerSuperblockReader creates a new ContainerSuperblockReader implementation
func NewContainerSuperblockReader(data []byte, endian binary.ByteOrder) (interfaces.ContainerSuperblockReader, error) {
	if len(data) < 1024 { // Minimum size for container superblock
		return nil, fmt.Errorf("data too small for container superblock: %d bytes", len(data))
	}

	superblock, err := parseContainerSuperblock(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse container superblock: %w", err)
	}

	// Validate magic number
	if superblock.NxMagic != types.NxMagic {
		return nil, fmt.Errorf("invalid container superblock magic: got 0x%08X, want 0x%08X", superblock.NxMagic, types.NxMagic)
	}

	return &ContainerSuperblockReader{
		Superblock: superblock,
		data:       data,
		endian:     endian,
	}, nil
}

// parseContainerSuperblock parses raw bytes into a NxSuperblockT structure
func parseContainerSuperblock(data []byte, endian binary.ByteOrder) (*types.NxSuperblockT, error) {
	if len(data) < 1024 { // Conservative minimum size
		return nil, fmt.Errorf("insufficient data for container superblock")
	}

	sb := &types.NxSuperblockT{}

	// Parse object header (first 32 bytes based on obj_phys_t)
	copy(sb.NxO.OChecksum[:], data[0:8])
	sb.NxO.OOid = types.OidT(endian.Uint64(data[8:16]))
	sb.NxO.OXid = types.XidT(endian.Uint64(data[16:24]))
	sb.NxO.OType = endian.Uint32(data[24:28])
	sb.NxO.OSubtype = endian.Uint32(data[28:32])

	// Parse container superblock specific fields
	sb.NxMagic = endian.Uint32(data[32:36])
	sb.NxBlockSize = endian.Uint32(data[36:40])
	sb.NxBlockCount = endian.Uint64(data[40:48])
	sb.NxFeatures = endian.Uint64(data[48:56])
	sb.NxReadonlyCompatibleFeatures = endian.Uint64(data[56:64])
	sb.NxIncompatibleFeatures = endian.Uint64(data[64:72])

	// Parse UUID (16 bytes)
	copy(sb.NxUuid[:], data[72:88])

	// Continue parsing critical fields
	sb.NxNextOid = types.OidT(endian.Uint64(data[88:96]))
	sb.NxNextXid = types.XidT(endian.Uint64(data[96:104]))

	// Checkpoint fields
	sb.NxXpDescBlocks = endian.Uint32(data[104:108])
	sb.NxXpDataBlocks = endian.Uint32(data[108:112])
	sb.NxXpDescBase = types.Paddr(endian.Uint64(data[112:120]))
	sb.NxXpDataBase = types.Paddr(endian.Uint64(data[120:128]))
	sb.NxXpDescNext = endian.Uint32(data[128:132])
	sb.NxXpDataNext = endian.Uint32(data[132:136])
	sb.NxXpDescIndex = endian.Uint32(data[136:140])
	sb.NxXpDescLen = endian.Uint32(data[140:144])
	sb.NxXpDataIndex = endian.Uint32(data[144:148])
	sb.NxXpDataLen = endian.Uint32(data[148:152])

	// Critical object identifiers
	sb.NxSpacemanOid = types.OidT(endian.Uint64(data[152:160]))
	sb.NxOmapOid = types.OidT(endian.Uint64(data[160:168]))
	sb.NxReaperOid = types.OidT(endian.Uint64(data[168:176]))

	// Testing and filesystem management
	sb.NxTestType = endian.Uint32(data[176:180])
	sb.NxMaxFileSystems = endian.Uint32(data[180:184])

	// Parse volume OIDs array (NxMaxFileSystems * 8 bytes)
	offset := 184
	for i := 0; i < types.NxMaxFileSystems; i++ {
		sb.NxFsOid[i] = types.OidT(endian.Uint64(data[offset : offset+8]))
		offset += 8
	}

	// Parse counters array (NxNumCounters * 8 bytes)
	for i := 0; i < types.NxNumCounters; i++ {
		sb.NxCounters[i] = endian.Uint64(data[offset : offset+8])
		offset += 8
	}

	// Parse blocked out range
	sb.NxBlockedOutPrange.PrStartPaddr = types.Paddr(endian.Uint64(data[offset : offset+8]))
	sb.NxBlockedOutPrange.PrBlockCount = endian.Uint64(data[offset+8 : offset+16])
	offset += 16

	// Continue with remaining fields
	sb.NxEvictMappingTreeOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	sb.NxFlags = endian.Uint64(data[offset : offset+8])
	offset += 8
	sb.NxEfiJumpstart = types.Paddr(endian.Uint64(data[offset : offset+8]))
	offset += 8

	// Parse Fusion UUID
	copy(sb.NxFusionUuid[:], data[offset:offset+16])
	offset += 16

	// Parse keybag location
	sb.NxKeylocker.PrStartPaddr = types.Paddr(endian.Uint64(data[offset : offset+8]))
	sb.NxKeylocker.PrBlockCount = endian.Uint64(data[offset+8 : offset+16])
	offset += 16

	// Parse ephemeral info array
	for i := 0; i < types.NxEphInfoCount; i++ {
		sb.NxEphemeralInfo[i] = endian.Uint64(data[offset : offset+8])
		offset += 8
	}

	// Parse remaining fields
	sb.NxTestOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	sb.NxFusionMtOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	sb.NxFusionWbcOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8

	// Parse Fusion write-back cache range
	sb.NxFusionWbc.PrStartPaddr = types.Paddr(endian.Uint64(data[offset : offset+8]))
	sb.NxFusionWbc.PrBlockCount = endian.Uint64(data[offset+8 : offset+16])
	offset += 16

	// Parse final fields
	sb.NxNewestMountedVersion = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse media key locker location
	if offset+16 <= len(data) {
		sb.NxMkbLocker.PrStartPaddr = types.Paddr(endian.Uint64(data[offset : offset+8]))
		sb.NxMkbLocker.PrBlockCount = endian.Uint64(data[offset+8 : offset+16])
	}

	return sb, nil
}

// Magic returns the magic number for validating the container superblock
func (csr *ContainerSuperblockReader) Magic() uint32 {
	return csr.Superblock.NxMagic
}

// BlockSize returns the logical block size used in the container
func (csr *ContainerSuperblockReader) BlockSize() uint32 {
	return csr.Superblock.NxBlockSize
}

// BlockCount returns the total number of logical blocks available in the container
func (csr *ContainerSuperblockReader) BlockCount() uint64 {
	return csr.Superblock.NxBlockCount
}

// UUID returns the universally unique identifier of the container
func (csr *ContainerSuperblockReader) UUID() types.UUID {
	return csr.Superblock.NxUuid
}

// NextObjectID returns the next object identifier to be used for new ephemeral or virtual objects
func (csr *ContainerSuperblockReader) NextObjectID() types.OidT {
	return csr.Superblock.NxNextOid
}

// NextTransactionID returns the next transaction to be used
func (csr *ContainerSuperblockReader) NextTransactionID() types.XidT {
	return csr.Superblock.NxNextXid
}

// SpaceManagerOID returns the ephemeral object identifier for the space manager
func (csr *ContainerSuperblockReader) SpaceManagerOID() types.OidT {
	return csr.Superblock.NxSpacemanOid
}

// ObjectMapOID returns the physical object identifier for the container's object map
func (csr *ContainerSuperblockReader) ObjectMapOID() types.OidT {
	return csr.Superblock.NxOmapOid
}

// ReaperOID returns the ephemeral object identifier for the reaper
func (csr *ContainerSuperblockReader) ReaperOID() types.OidT {
	return csr.Superblock.NxReaperOid
}

// MaxFileSystems returns the maximum number of volumes that can be stored in this container
func (csr *ContainerSuperblockReader) MaxFileSystems() uint32 {
	return csr.Superblock.NxMaxFileSystems
}

// VolumeOIDs returns the array of virtual object identifiers for volumes
func (csr *ContainerSuperblockReader) VolumeOIDs() []types.OidT {
	// Return only the valid (non-zero) volume OIDs
	var validOIDs []types.OidT
	for i := uint32(0); i < csr.Superblock.NxMaxFileSystems; i++ {
		if csr.Superblock.NxFsOid[i] != 0 {
			validOIDs = append(validOIDs, csr.Superblock.NxFsOid[i])
		}
	}
	return validOIDs
}

// EFIJumpstart returns the physical object identifier of the object that contains EFI driver data
func (csr *ContainerSuperblockReader) EFIJumpstart() types.Paddr {
	return csr.Superblock.NxEfiJumpstart
}

// FusionUUID returns the UUID of the container's Fusion set
func (csr *ContainerSuperblockReader) FusionUUID() types.UUID {
	return csr.Superblock.NxFusionUuid
}

// KeylockerLocation returns the location of the container's keybag
func (csr *ContainerSuperblockReader) KeylockerLocation() types.Prange {
	return csr.Superblock.NxKeylocker
}

// MediaKeyLocation returns the wrapped media key location
func (csr *ContainerSuperblockReader) MediaKeyLocation() types.Prange {
	return csr.Superblock.NxMkbLocker
}

// BlockedOutRange returns the blocked-out physical address range
func (csr *ContainerSuperblockReader) BlockedOutRange() types.Prange {
	return csr.Superblock.NxBlockedOutPrange
}

// EvictMappingTreeOID returns the object identifier of the evict-mapping tree
func (csr *ContainerSuperblockReader) EvictMappingTreeOID() types.OidT {
	return csr.Superblock.NxEvictMappingTreeOid
}

// TestType returns the container's test type for debugging
func (csr *ContainerSuperblockReader) TestType() uint32 {
	return csr.Superblock.NxTestType
}

// TestOID returns the test object identifier for debugging
func (csr *ContainerSuperblockReader) TestOID() types.OidT {
	return csr.Superblock.NxTestOid
}

// NewestMountedVersion returns the newest version of APFS that has mounted this container
func (csr *ContainerSuperblockReader) NewestMountedVersion() uint64 {
	return csr.Superblock.NxNewestMountedVersion
}
