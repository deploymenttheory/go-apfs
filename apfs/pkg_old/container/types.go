// Package container provides functionality for interacting with APFS containers.
//
// An APFS container represents a physical storage device divided into multiple
// volumes. This package provides the necessary structures and interfaces for
// working with APFS container-level objects, including:
//
// - Container superblocks (NX superblocks)
// - Object maps, which map object IDs to physical locations
// - Space managers, which track allocation of physical blocks
// - Checkpoints, which represent consistent states of the container
// - Reapers, which handle garbage collection of deleted objects
// - Encryption and key management
// - Fusion drive support
//
// The container layer is the foundation of the APFS file system and manages
// the physical storage resources that are shared between volumes. Each volume
// within a container has its own namespace of object IDs, but they all share
// the same physical device for storage.
//
// Basic usage:
//
//	import (
//		"github.com/yourusername/apfs/common"
//		"github.com/yourusername/apfs/container"
//	)
//
//	// Open a device
//	device, err := common.OpenDevice("/dev/disk0s2")
//	if err != nil {
//		panic(err)
//	}
//	defer device.Close()
//
//	// Create a factory
//	factory := container.NewFactory()
//
//	// Open the container
//	cont, err := factory.OpenContainer(device)
//	if err != nil {
//		panic(err)
//	}
//	defer cont.Close()
//
//	// List volumes
//	volumes, err := cont.ListVolumes()
//	if err != nil {
//		panic(err)
//	}
//
//	for _, vol := range volumes {
//		fmt.Printf("Volume: %s (UUID: %s)\n", vol.Name, vol.UUID)
//	}
//
// The container package is designed to work alongside the filesystem package,
// which provides higher-level access to the actual file systems within volumes.
package container
