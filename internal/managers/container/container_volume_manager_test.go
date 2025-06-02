package container

import (
	"encoding/binary"
	"testing"

	parser "github.com/deploymenttheory/go-apfs/internal/parsers/container"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestNewContainerVolumeManager(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	uuid := types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	endian := binary.LittleEndian

	data := CreateTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := parser.NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	objectMapReader := NewMockObjectMapReader()

	volumeManager := NewContainerVolumeManager(superblockReader, blockReader, objectMapReader)
	if volumeManager == nil {
		t.Fatal("NewContainerVolumeManager() returned nil")
	}
}

func TestContainerVolumeManager_ListVolumes(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	uuid := types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	endian := binary.LittleEndian

	data := CreateTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := parser.NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	objectMapReader := NewMockObjectMapReader()
	volumeManager := NewContainerVolumeManager(superblockReader, blockReader, objectMapReader)

	volumes, err := volumeManager.ListVolumes()
	if err != nil {
		t.Fatalf("ListVolumes() failed: %v", err)
	}

	// Based on our test data which creates 3 valid volumes (limited by the test)
	expectedVolumes := 3
	if len(volumes) != expectedVolumes {
		t.Errorf("ListVolumes() returned %d volumes, want %d", len(volumes), expectedVolumes)
	}

	// Verify each volume is valid
	for i, volume := range volumes {
		if volume == nil {
			t.Errorf("Volume %d is nil", i)
		}
	}
}

func TestContainerVolumeManager_FindVolumeByName(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	uuid := types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	endian := binary.LittleEndian

	data := CreateTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := parser.NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	objectMapReader := NewMockObjectMapReader()
	volumeManager := NewContainerVolumeManager(superblockReader, blockReader, objectMapReader)

	// Test finding a non-existent volume
	_, err = volumeManager.FindVolumeByName("NonExistentVolume")
	if err == nil {
		t.Error("FindVolumeByName() should have failed for non-existent volume")
	}

	// Test finding an existing volume would require mocking volume names
	// Since our current implementation creates placeholder volumes, this is limited
}

func TestContainerVolumeManager_FindVolumeByUUID(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	uuid := types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	endian := binary.LittleEndian

	data := CreateTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := parser.NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	objectMapReader := NewMockObjectMapReader()
	volumeManager := NewContainerVolumeManager(superblockReader, blockReader, objectMapReader)

	// Test finding a non-existent volume
	nonExistentUUID := types.UUID{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	_, err = volumeManager.FindVolumeByUUID(nonExistentUUID)
	if err == nil {
		t.Error("FindVolumeByUUID() should have failed for non-existent volume")
	}
}

func TestContainerVolumeManager_FindVolumesByRole(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	uuid := types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	endian := binary.LittleEndian

	data := CreateTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := parser.NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	objectMapReader := NewMockObjectMapReader()
	volumeManager := NewContainerVolumeManager(superblockReader, blockReader, objectMapReader)

	// Test finding volumes by role (0 = no specific role)
	volumes, err := volumeManager.FindVolumesByRole(0)
	if err != nil {
		t.Fatalf("FindVolumesByRole() failed: %v", err)
	}

	// Since our placeholder volumes have default role 0, we should find all of them
	expectedCount := 3 // Based on test data
	if len(volumes) != expectedCount {
		t.Errorf("FindVolumesByRole() returned %d volumes, want %d", len(volumes), expectedCount)
	}

	// Test finding volumes with a specific role that doesn't exist
	volumes, err = volumeManager.FindVolumesByRole(999)
	if err != nil {
		t.Fatalf("FindVolumesByRole() failed: %v", err)
	}

	if len(volumes) != 0 {
		t.Errorf("FindVolumesByRole() returned %d volumes for non-existent role, want 0", len(volumes))
	}
}

func TestContainerVolumeManager_LoadVolume(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	uuid := types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	endian := binary.LittleEndian

	data := CreateTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := parser.NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	objectMapReader := NewMockObjectMapReader()
	volumeManager := NewContainerVolumeManager(superblockReader, blockReader, objectMapReader)

	// Test loading a volume (this tests the private loadVolume method indirectly)
	testOID := types.OidT(6000)
	volume, err := volumeManager.loadVolume(testOID)
	if err != nil {
		t.Fatalf("loadVolume() failed: %v", err)
	}

	if volume == nil {
		t.Fatal("loadVolume() returned nil volume")
	}

	// Verify the volume has the correct magic number
	if !volume.ValidateMagicNumber() {
		t.Error("Volume magic number validation failed")
	}

	// Verify basic volume properties
	if volume.MagicNumber() != types.ApfsMagic {
		t.Errorf("Volume magic = 0x%08X, want 0x%08X", volume.MagicNumber(), types.ApfsMagic)
	}
}

func TestContainerVolumeManager_EmptyContainer(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	uuid := types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	endian := binary.LittleEndian

	// Create container with no volumes (maxFileSystems = 0)
	data := CreateTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 0, endian)
	superblockReader, err := parser.NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	objectMapReader := NewMockObjectMapReader()
	volumeManager := NewContainerVolumeManager(superblockReader, blockReader, objectMapReader)

	volumes, err := volumeManager.ListVolumes()
	if err != nil {
		t.Fatalf("ListVolumes() failed: %v", err)
	}

	if len(volumes) != 0 {
		t.Errorf("ListVolumes() returned %d volumes for empty container, want 0", len(volumes))
	}
}

func TestContainerVolumeManager_ErrorHandling(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	uuid := types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	endian := binary.LittleEndian

	data := CreateTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := parser.NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	objectMapReader := NewMockObjectMapReader()
	volumeManager := NewContainerVolumeManager(superblockReader, blockReader, objectMapReader)

	// Test error propagation when finding non-existent volumes
	testCases := []struct {
		name     string
		testFunc func() error
	}{
		{
			name: "FindVolumeByName non-existent",
			testFunc: func() error {
				_, err := volumeManager.FindVolumeByName("DoesNotExist")
				return err
			},
		},
		{
			name: "FindVolumeByUUID non-existent",
			testFunc: func() error {
				nonExistentUUID := types.UUID{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
				_, err := volumeManager.FindVolumeByUUID(nonExistentUUID)
				return err
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.testFunc()
			if err == nil {
				t.Errorf("Expected error for %s, but got nil", tc.name)
			}
		})
	}
}
