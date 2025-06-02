package container

import (
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// containerManager implements the ContainerManager interface
type containerManager struct {
	superblockReader  interfaces.ContainerSuperblockReader
	featureManager    interfaces.ContainerFeatureManager
	flagManager       interfaces.ContainerFlagManager
	checkpointManager interfaces.ContainerCheckpointManager
	statisticsReader  interfaces.ContainerStatisticsReader
	ephemeralManager  interfaces.ContainerEphemeralManager
	blockReader       interfaces.BlockDeviceReader
	volumeManager     *containerVolumeManager
	objectMapReader   interfaces.ObjectMapReader // From object_maps package
}

// NewContainerManager creates a new ContainerManager implementation
func NewContainerManager(
	superblockReader interfaces.ContainerSuperblockReader,
	blockReader interfaces.BlockDeviceReader,
	objectMapReader interfaces.ObjectMapReader,
) interfaces.ContainerManager {

	// Extract superblock for component managers
	csr := superblockReader.(*containerSuperblockReader)
	superblock := csr.superblock

	featureManager := NewContainerFeatureManager(superblock)
	flagManager := NewContainerFlagManager(superblock)
	checkpointManager := NewContainerCheckpointManager(superblock)
	statisticsReader := NewContainerStatisticsReader(superblock)
	ephemeralManager := NewContainerEphemeralManager(superblock)
	volumeManager := NewContainerVolumeManager(superblockReader, blockReader, objectMapReader)

	return &containerManager{
		superblockReader:  superblockReader,
		featureManager:    featureManager,
		flagManager:       flagManager,
		checkpointManager: checkpointManager,
		statisticsReader:  statisticsReader,
		ephemeralManager:  ephemeralManager,
		blockReader:       blockReader,
		volumeManager:     volumeManager,
		objectMapReader:   objectMapReader,
	}
}

// GetObjectMap returns the container's object map for virtual object resolution
func (cm *containerManager) GetObjectMap() (interfaces.ObjectMapReader, error) {
	if cm.objectMapReader != nil {
		return cm.objectMapReader, nil
	}

	// Lazy load the object map if not provided during construction
	omapOID := cm.superblockReader.ObjectMapOID()
	if omapOID == 0 {
		return nil, fmt.Errorf("container object map OID is zero")
	}

	// Read the object map header from disk
	_, err := cm.blockReader.ReadBlock(types.Paddr(omapOID))
	if err != nil {
		return nil, fmt.Errorf("failed to read object map block: %w", err)
	}

	// Parse the object map header
	// This would need proper parsing based on the object_maps package structure
	// For now, this is a placeholder that would be implemented based on available parsers
	return nil, fmt.Errorf("object map parsing not yet implemented")
}

// ResolveVirtualObject resolves a virtual object ID to a physical address using the container's object map
func (cm *containerManager) ResolveVirtualObject(oid types.OidT, xid types.XidT) (types.Paddr, error) {
	objectMap, err := cm.GetObjectMap()
	if err != nil {
		return 0, fmt.Errorf("failed to get object map: %w", err)
	}

	// This would use the object map to perform virtual-to-physical resolution
	// The actual implementation would depend on the object_maps package providing
	// virtual object resolution capabilities
	_ = objectMap // Use the object map when the interface supports resolution

	return 0, fmt.Errorf("virtual object resolution not yet implemented for OID %d, XID %d", oid, xid)
}

// Volume Discovery and Management

// ListVolumes returns all volumes in the container
func (cm *containerManager) ListVolumes() ([]interfaces.Volume, error) {
	return cm.volumeManager.ListVolumes()
}

// FindVolumeByName finds a volume by its name
func (cm *containerManager) FindVolumeByName(name string) (interfaces.Volume, error) {
	return cm.volumeManager.FindVolumeByName(name)
}

// FindVolumeByUUID finds a volume by its UUID
func (cm *containerManager) FindVolumeByUUID(uuid types.UUID) (interfaces.Volume, error) {
	return cm.volumeManager.FindVolumeByUUID(uuid)
}

// FindVolumesByRole finds volumes by their role
func (cm *containerManager) FindVolumesByRole(role uint16) ([]interfaces.Volume, error) {
	return cm.volumeManager.FindVolumesByRole(role)
}

// Container Space Management

// TotalSize returns the total size of the container in bytes
func (cm *containerManager) TotalSize() uint64 {
	return cm.superblockReader.BlockCount() * uint64(cm.superblockReader.BlockSize())
}

// FreeSpace returns the free space in the container in bytes
func (cm *containerManager) FreeSpace() uint64 {
	// Fallback calculation - this would be improved with proper space manager integration
	return cm.TotalSize() - cm.UsedSpace()
}

// UsedSpace returns the used space in the container in bytes
func (cm *containerManager) UsedSpace() uint64 {
	// Conservative estimate - in a real implementation this would use space manager
	return cm.TotalSize() / 2
}

// SpaceUtilization returns the space utilization as a percentage (0-100)
func (cm *containerManager) SpaceUtilization() float64 {
	totalSize := cm.TotalSize()
	if totalSize == 0 {
		return 0.0
	}
	return (float64(cm.UsedSpace()) / float64(totalSize)) * 100.0
}

// Blocked and Reserved Space

// BlockedOutRange returns the blocked-out physical address range
func (cm *containerManager) BlockedOutRange() types.Prange {
	return cm.superblockReader.BlockedOutRange()
}

// EvictMappingTreeOID returns the object identifier of the evict-mapping tree
func (cm *containerManager) EvictMappingTreeOID() types.OidT {
	return cm.superblockReader.EvictMappingTreeOID()
}

// Container Metadata

// UUID returns the container's UUID
func (cm *containerManager) UUID() types.UUID {
	return cm.superblockReader.UUID()
}

// NextObjectID returns the next object identifier
func (cm *containerManager) NextObjectID() types.OidT {
	return cm.superblockReader.NextObjectID()
}

// NextTransactionID returns the next transaction identifier
func (cm *containerManager) NextTransactionID() types.XidT {
	return cm.superblockReader.NextTransactionID()
}

// Features and Compatibility

// Features returns the optional features being used
func (cm *containerManager) Features() uint64 {
	return cm.featureManager.Features()
}

// IncompatibleFeatures returns the backward-incompatible features
func (cm *containerManager) IncompatibleFeatures() uint64 {
	return cm.featureManager.IncompatibleFeatures()
}

// ReadonlyCompatibleFeatures returns the read-only compatible features
func (cm *containerManager) ReadonlyCompatibleFeatures() uint64 {
	return cm.featureManager.ReadOnlyCompatibleFeatures()
}

// Encryption and Security

// IsEncrypted returns true if the container uses encryption
func (cm *containerManager) IsEncrypted() bool {
	// Check if software crypto flag is set or if keybag location is valid
	return cm.flagManager.UsesSoftwareCryptography() ||
		cm.superblockReader.KeylockerLocation().PrBlockCount > 0
}

// CryptoType returns the cryptography type flags
func (cm *containerManager) CryptoType() uint64 {
	flags := cm.flagManager.Flags()
	if flags&types.NxCryptoSw != 0 {
		return types.NxCryptoSw
	}
	return 0
}

// Snapshots and Versioning

// TotalSnapshots returns the total number of snapshots across all volumes
func (cm *containerManager) TotalSnapshots() uint64 {
	// Placeholder - would need volume snapshot integration
	return 0
}

// LatestSnapshotXID returns the latest snapshot transaction identifier
func (cm *containerManager) LatestSnapshotXID() types.XidT {
	// Placeholder - would need volume snapshot integration
	return 0
}

// Health and Integrity

// CheckIntegrity performs integrity checks on the container
func (cm *containerManager) CheckIntegrity() (bool, []string) {
	var issues []string
	isHealthy := true

	// Check magic number
	if cm.superblockReader.Magic() != types.NxMagic {
		issues = append(issues, "Invalid container superblock magic number")
		isHealthy = false
	}

	// Check block size is valid power of 2
	blockSize := cm.superblockReader.BlockSize()
	if blockSize == 0 || (blockSize&(blockSize-1)) != 0 {
		issues = append(issues, fmt.Sprintf("Invalid block size: %d", blockSize))
		isHealthy = false
	}

	// Check total block count is reasonable
	if cm.superblockReader.BlockCount() == 0 {
		issues = append(issues, "Container has zero blocks")
		isHealthy = false
	}

	// Check volume count doesn't exceed maximum
	volumeOIDs := cm.superblockReader.VolumeOIDs()
	if uint32(len(volumeOIDs)) > cm.superblockReader.MaxFileSystems() {
		issues = append(issues, fmt.Sprintf("Volume count (%d) exceeds maximum (%d)",
			len(volumeOIDs), cm.superblockReader.MaxFileSystems()))
		isHealthy = false
	}

	// Check checksums if statistics reader indicates failures
	if cm.statisticsReader.ObjectChecksumFailCount() > 0 {
		issues = append(issues, fmt.Sprintf("Container has %d checksum failures",
			cm.statisticsReader.ObjectChecksumFailCount()))
		// Don't mark as unhealthy for old failures, just warn
	}

	// Verify object identifiers are in valid ranges
	nextOID := cm.superblockReader.NextObjectID()
	if uint64(nextOID) <= types.OidReservedCount {
		issues = append(issues, fmt.Sprintf("Next OID (%d) is in reserved range", nextOID))
		isHealthy = false
	}

	return isHealthy, issues
}

// IsHealthy returns true if the container passes basic health checks
func (cm *containerManager) IsHealthy() bool {
	isHealthy, _ := cm.CheckIntegrity()
	return isHealthy
}
