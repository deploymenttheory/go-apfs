package services

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/deploymenttheory/go-apfs/internal/parsers/volumes"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// VolumeServiceImpl provides volume-level operations
type VolumeServiceImpl struct {
	container *ContainerReader
	volumeOID types.OidT
	volumeSB  *types.ApfsSuperblockT
	mu        sync.RWMutex
}

// NewVolumeService creates a new VolumeService instance
func NewVolumeService(container *ContainerReader, volumeOID types.OidT) (*VolumeServiceImpl, error) {
	if container == nil {
		return nil, fmt.Errorf("container reader cannot be nil")
	}

	if volumeOID == 0 {
		return nil, fmt.Errorf("invalid volume OID: 0")
	}

	// APFS volume OIDs from container superblock are typically VIRTUAL object identifiers
	// They need to be resolved through the object map to get the physical address
	// However, in freshly created volumes, the object map may be empty, so we try different approaches
	resolver := NewBTreeObjectResolver(container)
	physicalAddr, err := resolver.ResolveVirtualObject(volumeOID, container.GetSuperblock().NxNextXid-1)
	if err != nil {
		// Fallback 1: try treating volumeOID as a direct physical address
		physicalAddr = types.Paddr(volumeOID)

		// Fallback 2: if that fails, scan for volume superblock near container structures
		volSBData, readErr := container.ReadBlock(uint64(physicalAddr))
		if readErr != nil || !isValidVolumeSuperblock(volSBData) {
			// Scan more blocks for the volume superblock (volumes can be stored anywhere in container)
			found := false
			for scanBlock := uint64(1); scanBlock < 500; scanBlock++ {
				scanData, scanErr := container.ReadBlock(scanBlock)
				if scanErr == nil && isValidVolumeSuperblock(scanData) {
					physicalAddr = types.Paddr(scanBlock)
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("failed to locate volume superblock for OID %d", volumeOID)
			}
		}
	}

	// Read volume superblock from resolved physical address
	volSBData, err := container.ReadBlock(uint64(physicalAddr))
	if err != nil {
		return nil, fmt.Errorf("failed to read volume superblock at physical address %d: %w", physicalAddr, err)
	}

	// Parse volume superblock using the volume superblock reader
	volSBReader, err := volumes.NewVolumeSuperblockReader(volSBData, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume superblock: %w", err)
	}

	volSB := volSBReader.GetSuperblock()
	if volSB == nil {
		return nil, fmt.Errorf("failed to extract superblock")
	}

	vs := &VolumeServiceImpl{
		container: container,
		volumeOID: volumeOID,
		volumeSB:  volSB,
	}

	return vs, nil
}

// isValidVolumeSuperblock checks if a block contains a valid APFS volume superblock
func isValidVolumeSuperblock(data []byte) bool {
	if len(data) < 36 {
		return false
	}
	// Check for APFS volume superblock magic "APSB" (0x42535041)
	magic := uint32(data[32]) | uint32(data[33])<<8 | uint32(data[34])<<16 | uint32(data[35])<<24
	return magic == 0x42535041
}

// NewVolumeServiceFromPhysicalOID creates a VolumeService using a direct physical OID (bypassing object map)
func NewVolumeServiceFromPhysicalOID(container *ContainerReader, physicalOID types.OidT) (*VolumeServiceImpl, error) {
	if container == nil {
		return nil, fmt.Errorf("container reader cannot be nil")
	}

	if physicalOID == 0 {
		return nil, fmt.Errorf("invalid physical OID: 0")
	}

	// Read volume superblock directly from physical address
	volSBData, err := container.ReadBlock(uint64(physicalOID))
	if err != nil {
		return nil, fmt.Errorf("failed to read volume superblock at physical address %d: %w", physicalOID, err)
	}

	// Parse volume superblock using the volume superblock reader
	volSBReader, err := volumes.NewVolumeSuperblockReader(volSBData, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume superblock: %w", err)
	}

	volSB := volSBReader.GetSuperblock()
	if volSB == nil {
		return nil, fmt.Errorf("failed to extract superblock")
	}

	vs := &VolumeServiceImpl{
		container: container,
		volumeOID: physicalOID,
		volumeSB:  volSB,
	}

	return vs, nil
}

// GetVolumeMetadata returns comprehensive volume metadata
func (vs *VolumeServiceImpl) GetVolumeMetadata() (*VolumeReport, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	if vs.volumeSB == nil {
		return nil, fmt.Errorf("volume superblock not loaded")
	}

	// Create space management reader
	volSpace := volumes.NewVolumeSpaceManagement(vs.volumeSB)
	if volSpace == nil {
		return nil, fmt.Errorf("failed to create volume space management reader")
	}

	// Build space stats
	blockSize := vs.container.GetBlockSize()
	quotaBlocks := volSpace.QuotaBlockCount()
	allocBlocks := volSpace.AllocatedBlockCount()
	freeBlocks := quotaBlocks - allocBlocks

	spaceStats := SpaceStats{
		TotalCapacity:       quotaBlocks * uint64(blockSize),
		UsedSpace:           allocBlocks * uint64(blockSize),
		FreeSpace:           freeBlocks * uint64(blockSize),
		AllocationBlockSize: blockSize,
		UsagePercentage:     volSpace.SpaceUtilization(),
		FreeBlocks:          freeBlocks,
		AllocatedBlocks:     allocBlocks,
		FragmentationRatio:  0.0,
	}

	report := &VolumeReport{
		VolumeOID:      uint64(vs.volumeOID),
		SpaceStats:     spaceStats,
		GeneratedAt:    time.Now(),
		FileCount:      vs.volumeSB.ApfsNumFiles,
		DirectoryCount: vs.volumeSB.ApfsNumDirectories,
		SymlinkCount:   vs.volumeSB.ApfsNumSymlinks,
		SnapshotCount:  vs.volumeSB.ApfsNumSnapshots,
	}

	return report, nil
}

// GetSpaceUsageStats returns detailed space usage statistics
func (vs *VolumeServiceImpl) GetSpaceUsageStats() (*SpaceStats, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	if vs.volumeSB == nil {
		return nil, fmt.Errorf("volume superblock not loaded")
	}

	// Create space management reader
	volSpace := volumes.NewVolumeSpaceManagement(vs.volumeSB)
	if volSpace == nil {
		return nil, fmt.Errorf("failed to create volume space management reader")
	}

	blockSize := vs.container.GetBlockSize()
	quotaBlocks := volSpace.QuotaBlockCount()
	allocBlocks := volSpace.AllocatedBlockCount()
	freeBlocks := quotaBlocks - allocBlocks

	stats := &SpaceStats{
		TotalCapacity:       quotaBlocks * uint64(blockSize),
		UsedSpace:           allocBlocks * uint64(blockSize),
		FreeSpace:           freeBlocks * uint64(blockSize),
		AllocationBlockSize: blockSize,
		UsagePercentage:     volSpace.SpaceUtilization(),
		FreeBlocks:          freeBlocks,
		AllocatedBlocks:     allocBlocks,
		FragmentationRatio:  0.0,
	}

	return stats, nil
}

// AnalyzeVolumeFragmentation analyzes the filesystem fragmentation
func (vs *VolumeServiceImpl) AnalyzeVolumeFragmentation() (map[string]interface{}, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	if vs.volumeSB == nil {
		return nil, fmt.Errorf("volume superblock not loaded")
	}

	// TODO: Implement fragmentation analysis using space manager
	result := map[string]interface{}{
		"status":                "not_implemented",
		"fragmentation_ratio":   0.0,
		"contiguous_extents":    0,
		"fragmented_extents":    0,
		"largest_contiguous":    0,
		"smallest_fragment":     0,
		"average_fragment_size": 0,
	}

	return result, nil
}

// DetectCorruption scans for corruption anomalies
func (vs *VolumeServiceImpl) DetectCorruption() ([]VolumeCorruptionAnomaly, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	if vs.volumeSB == nil {
		return nil, fmt.Errorf("volume superblock not loaded")
	}

	anomalies := []VolumeCorruptionAnomaly{}

	// TODO: Implement corruption detection
	// - Check superblock consistency
	// - Verify space manager integrity
	// - Check object map consistency

	return anomalies, nil
}

// GenerateVolumeReport generates a comprehensive volume report
func (vs *VolumeServiceImpl) GenerateVolumeReport() (*VolumeReport, error) {
	return vs.GetVolumeMetadata()
}

// GetFileCount returns the total number of files
func (vs *VolumeServiceImpl) GetFileCount() (uint64, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	if vs.volumeSB == nil {
		return 0, fmt.Errorf("volume superblock not loaded")
	}

	return vs.volumeSB.ApfsNumFiles, nil
}

// GetDirectoryCount returns the total number of directories
func (vs *VolumeServiceImpl) GetDirectoryCount() (uint64, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	if vs.volumeSB == nil {
		return 0, fmt.Errorf("volume superblock not loaded")
	}

	return vs.volumeSB.ApfsNumDirectories, nil
}

// GetSymlinkCount returns the total number of symlinks
func (vs *VolumeServiceImpl) GetSymlinkCount() (uint64, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	if vs.volumeSB == nil {
		return 0, fmt.Errorf("volume superblock not loaded")
	}

	return vs.volumeSB.ApfsNumSymlinks, nil
}
