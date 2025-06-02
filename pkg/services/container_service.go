package services

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/parsers/container"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// containerService implements the ContainerService interface
type containerService struct {
	openContainers map[string]*containerHandle
}

// containerHandle represents an open container
type containerHandle struct {
	devicePath string
	file       *os.File
	superblock *types.NxSuperblockT
	reader     interfaces.ContainerSuperblockReader
	openedAt   time.Time
}

// NewContainerService creates a new container service instance
func NewContainerService() ContainerService {
	return &containerService{
		openContainers: make(map[string]*containerHandle),
	}
}

// DiscoverContainers finds APFS containers on accessible devices
func (cs *containerService) DiscoverContainers(ctx context.Context) ([]ContainerInfo, error) {
	var containers []ContainerInfo

	// Look for APFS containers in common locations
	searchPaths := []string{
		"/dev/disk*", // macOS disk devices
		"/Volumes/*", // Mounted volumes
		"*.dmg",      // DMG files in current directory
		"*.img",      // IMG files in current directory
	}

	for _, pattern := range searchPaths {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, path := range matches {
			// Skip if context is cancelled
			select {
			case <-ctx.Done():
				return containers, ctx.Err()
			default:
			}

			if info, err := cs.tryOpenContainer(ctx, path); err == nil {
				containers = append(containers, info)
			}
		}
	}

	// Ensure we always return a non-nil slice
	if containers == nil {
		containers = []ContainerInfo{}
	}

	return containers, nil
}

// OpenContainer opens a container at the specified path
func (cs *containerService) OpenContainer(ctx context.Context, devicePath string) (ContainerInfo, error) {
	// Check if already open
	if handle, exists := cs.openContainers[devicePath]; exists {
		return cs.buildContainerInfo(handle)
	}

	// Open the device/file
	file, err := os.Open(devicePath)
	if err != nil {
		return ContainerInfo{}, fmt.Errorf("failed to open device %s: %w", devicePath, err)
	}

	// Read the container superblock at block 0
	superblock, reader, err := cs.readContainerSuperblock(file)
	if err != nil {
		file.Close()
		return ContainerInfo{}, fmt.Errorf("failed to read superblock: %w", err)
	}

	// Create container handle
	handle := &containerHandle{
		devicePath: devicePath,
		file:       file,
		superblock: superblock,
		reader:     reader,
		openedAt:   time.Now(),
	}

	cs.openContainers[devicePath] = handle
	return cs.buildContainerInfo(handle)
}

// ReadSuperblock reads and parses the container superblock
func (cs *containerService) ReadSuperblock(ctx context.Context, devicePath string) (*types.NxSuperblockT, error) {
	_, err := cs.OpenContainer(ctx, devicePath)
	if err != nil {
		return nil, err
	}

	handle := cs.openContainers[devicePath]
	return handle.superblock, nil
}

// ListVolumes enumerates all volumes in the container
func (cs *containerService) ListVolumes(ctx context.Context, devicePath string) ([]VolumeInfo, error) {
	info, err := cs.OpenContainer(ctx, devicePath)
	if err != nil {
		return nil, err
	}

	return info.Volumes, nil
}

// GetSpaceManagerInfo retrieves space management information
func (cs *containerService) GetSpaceManagerInfo(ctx context.Context, devicePath string) (SpaceManagerInfo, error) {
	info, err := cs.OpenContainer(ctx, devicePath)
	if err != nil {
		return SpaceManagerInfo{}, err
	}

	return info.SpaceManager, nil
}

// VerifyCheckpoints validates container checkpoints
func (cs *containerService) VerifyCheckpoints(ctx context.Context, devicePath string) error {
	handle, exists := cs.openContainers[devicePath]
	if !exists {
		_, err := cs.OpenContainer(ctx, devicePath)
		if err != nil {
			return err
		}
		handle = cs.openContainers[devicePath]
	}

	// Read checkpoint area and validate
	// This is a simplified implementation - real verification would be more comprehensive
	if handle.reader.NextTransactionID() == 0 {
		return fmt.Errorf("invalid checkpoint data")
	}

	return nil
}

// Close closes the container and releases resources
func (cs *containerService) Close() error {
	for path, handle := range cs.openContainers {
		if err := handle.file.Close(); err != nil {
			// Log error but continue closing others
			fmt.Printf("Error closing container %s: %v\n", path, err)
		}
	}
	cs.openContainers = make(map[string]*containerHandle)
	return nil
}

// tryOpenContainer attempts to open a container and return basic info
func (cs *containerService) tryOpenContainer(ctx context.Context, path string) (ContainerInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return ContainerInfo{}, err
	}
	defer file.Close()

	// Try to read APFS superblock
	_, reader, err := cs.readContainerSuperblock(file)
	if err != nil {
		return ContainerInfo{}, err
	}

	// Basic container info without full parsing
	info := ContainerInfo{
		DevicePath:   path,
		BlockSize:    reader.BlockSize(),
		BlockCount:   reader.BlockCount(),
		VolumeCount:  uint32(len(reader.VolumeOIDs())),
		CheckpointID: uint64(reader.NextTransactionID()),
		Features:     cs.extractFeatures(reader),
		Encrypted:    cs.isContainerEncrypted(reader),
	}

	return info, nil
}

// readContainerSuperblock reads and parses the container superblock
func (cs *containerService) readContainerSuperblock(file *os.File) (*types.NxSuperblockT, interfaces.ContainerSuperblockReader, error) {
	// Read block 0 (4096 bytes - typical APFS block size)
	blockSize := 4096
	data := make([]byte, blockSize)

	_, err := file.ReadAt(data, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read superblock: %w", err)
	}

	// Check for APFS magic
	magic := binary.LittleEndian.Uint32(data[32:36]) // Magic is at offset 32 in obj_phys_t
	if magic != 0x4253584e {                         // "NXSB" in little-endian
		return nil, nil, fmt.Errorf("invalid APFS magic: expected 0x4253584e, got 0x%x", magic)
	}

	// Parse superblock using container parser
	reader, err := container.NewContainerSuperblockReader(data, binary.LittleEndian)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse superblock: %w", err)
	}

	// Create types structure (simplified - real implementation would parse all fields)
	superblock := &types.NxSuperblockT{
		NxMagic:       reader.Magic(),
		NxBlockSize:   reader.BlockSize(),
		NxBlockCount:  reader.BlockCount(),
		NxNextXid:     reader.NextTransactionID(),
		NxSpacemanOid: reader.SpaceManagerOID(),
		NxOmapOid:     reader.ObjectMapOID(),
		NxReaperOid:   reader.ReaperOID(),
	}

	return superblock, reader, nil
}

// buildContainerInfo creates a complete ContainerInfo from a container handle
func (cs *containerService) buildContainerInfo(handle *containerHandle) (ContainerInfo, error) {
	info := ContainerInfo{
		DevicePath:      handle.devicePath,
		BlockSize:       handle.reader.BlockSize(),
		BlockCount:      handle.reader.BlockCount(),
		VolumeCount:     uint32(len(handle.reader.VolumeOIDs())),
		CheckpointID:    uint64(handle.reader.NextTransactionID()),
		Features:        cs.extractFeatures(handle.reader),
		Encrypted:       cs.isContainerEncrypted(handle.reader),
		CaseInsensitive: cs.isCaseInsensitive(handle.reader),
	}

	// Get volume information
	volumes, err := cs.buildVolumeList(handle)
	if err != nil {
		return info, fmt.Errorf("failed to build volume list: %w", err)
	}
	info.Volumes = volumes

	// Get space manager information
	spaceInfo, err := cs.buildSpaceManagerInfo(handle)
	if err != nil {
		return info, fmt.Errorf("failed to build space manager info: %w", err)
	}
	info.SpaceManager = spaceInfo

	return info, nil
}

// buildVolumeList creates volume information list
func (cs *containerService) buildVolumeList(handle *containerHandle) ([]VolumeInfo, error) {
	var volumes []VolumeInfo

	volumeOIDs := handle.reader.VolumeOIDs()
	for i, oid := range volumeOIDs {
		volume := VolumeInfo{
			ObjectID:      uint64(oid),
			Name:          fmt.Sprintf("Volume_%d", i),            // Simplified - real implementation would read volume names
			Role:          "Unknown",                              // Would need volume superblock to determine role
			Encrypted:     cs.isContainerEncrypted(handle.reader), // Simplified
			CaseSensitive: !cs.isCaseInsensitive(handle.reader),
			LastModified:  time.Now(), // Would need to read from volume metadata
		}
		volumes = append(volumes, volume)
	}

	return volumes, nil
}

// buildSpaceManagerInfo creates space manager information
func (cs *containerService) buildSpaceManagerInfo(handle *containerHandle) (SpaceManagerInfo, error) {
	return SpaceManagerInfo{
		BlockSize:          handle.reader.BlockSize(),
		ChunkCount:         0,   // Would need to read space manager structure
		FreeBlocks:         0,   // Would need space manager data
		UsedBlocks:         0,   // Would need space manager data
		ReservedBlocks:     0,   // Would need space manager data
		FragmentationRatio: 0.0, // Would need analysis
	}, nil
}

// extractFeatures extracts feature flags from the container
func (cs *containerService) extractFeatures(reader interfaces.ContainerSuperblockReader) []string {
	var features []string

	// Check various feature flags (simplified)
	features = append(features, "APFS")

	// Would check actual feature flags from superblock
	// flags := reader.FeatureFlags()
	// if flags & SOME_FLAG != 0 {
	//     features = append(features, "SomeFeature")
	// }

	return features
}

// isContainerEncrypted checks if the container has encryption enabled
func (cs *containerService) isContainerEncrypted(reader interfaces.ContainerSuperblockReader) bool {
	// Simplified check - real implementation would check keybag presence and encryption flags
	// For now, check if media key location is set
	mediaKeyLocation := reader.MediaKeyLocation()
	return mediaKeyLocation.PrStartPaddr != 0
}

// isCaseInsensitive checks if the container uses case-insensitive filenames
func (cs *containerService) isCaseInsensitive(reader interfaces.ContainerSuperblockReader) bool {
	// Simplified - would check actual feature flags
	return false
}
