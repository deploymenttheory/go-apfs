package container

import (
	"fmt"

	"github.com/deploymenttheory/go-apfs/apfs/pkg/types"
)

// WriteVolumeSuperblock writes a volume superblock to the given physical address on disk.
func WriteVolumeSuperblock(device types.BlockDevice, addr types.PAddr, volume *types.APFSSuperblock) error {
	if device == nil || volume == nil {
		return fmt.Errorf("device and volume must not be nil")
	}

	// Serialize the volume superblock
	data, err := serializeAPFSSuperblock(volume)
	if err != nil {
		return fmt.Errorf("failed to serialize volume superblock: %w", err)
	}

	// Write to the specified block address
	return device.WriteBlock(addr, data)
}
