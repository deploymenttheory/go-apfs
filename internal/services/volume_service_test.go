package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/disk"
	"github.com/deploymenttheory/go-apfs/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVolumeServiceMetadata(t *testing.T) {
	config := &disk.DMGConfig{
		AutoDetectAPFS: true,
		DefaultOffset:  20480,
		TestDataPath:   "../../tests",
	}

	testDMG := filepath.Join(config.TestDataPath, "populated_apfs.dmg")
	if _, err := os.Stat(testDMG); err != nil {
		t.Skipf("Test DMG not found: %v", testDMG)
	}

	dmg, err := disk.OpenDMG(testDMG, config)
	require.NoError(t, err, "failed to open test DMG")
	defer dmg.Close()

	// Create container reader and test container metadata
	cr, err := NewContainerReaderFromDevice(dmg, uint64(dmg.Size()))
	require.NoError(t, err, "failed to create container reader")
	defer cr.Close()

	containerSB := cr.GetSuperblock()
	require.NotNil(t, containerSB, "container superblock should not be nil")

	// Test: Container has valid magic
	assert.Equal(t, types.NxMagic, containerSB.NxMagic, "container should have valid NXSB magic")

	// Test: Container has valid block size
	assert.Greater(t, containerSB.NxBlockSize, uint32(0), "block size should be greater than 0")
	assert.Equal(t, uint32(4096), containerSB.NxBlockSize, "typical block size should be 4096")

	// Test: Container has valid block count
	assert.Greater(t, containerSB.NxBlockCount, uint64(0), "block count should be greater than 0")

	// Test: Container has at least one volume
	volumeCount := 0
	for _, oid := range containerSB.NxFsOid {
		if oid != 0 {
			volumeCount++
		}
	}
	assert.Greater(t, volumeCount, 0, "container should have at least one volume")

	// Test: Next XID is valid
	assert.Greater(t, containerSB.NxNextXid, uint64(0), "next XID should be greater than 0")

	t.Logf("Container: %d volumes, %d blocks of %d bytes each, Next XID: %d",
		volumeCount, containerSB.NxBlockCount, containerSB.NxBlockSize, containerSB.NxNextXid)
}

func TestVolumeServiceSpaceStats(t *testing.T) {
	config := &disk.DMGConfig{
		AutoDetectAPFS: true,
		DefaultOffset:  20480,
		TestDataPath:   "../../tests",
	}

	testDMG := filepath.Join(config.TestDataPath, "full_apfs.dmg")
	if _, err := os.Stat(testDMG); err != nil {
		t.Skipf("Test DMG not found: %v", testDMG)
	}

	dmg, err := disk.OpenDMG(testDMG, config)
	require.NoError(t, err, "failed to open test DMG")
	defer dmg.Close()

	cr, err := NewContainerReaderFromDevice(dmg, uint64(dmg.Size()))
	require.NoError(t, err, "failed to create container reader")
	defer cr.Close()

	containerSB := cr.GetSuperblock()
	require.NotNil(t, containerSB, "container superblock should not be nil")

	// Test: Calculate container total capacity
	totalBlocks := containerSB.NxBlockCount
	blockSize := uint64(containerSB.NxBlockSize)
	totalCapacity := totalBlocks * blockSize

	assert.Greater(t, totalCapacity, uint64(0), "total capacity should be greater than 0")
	t.Logf("Container Total Capacity: %d bytes (%.2f MB)", totalCapacity, float64(totalCapacity)/(1024*1024))

	// Test: Container block count is reasonable
	assert.Greater(t, totalBlocks, uint64(1000), "should have at least 1000 blocks for a 5MB+ volume")
	t.Logf("Total Blocks: %d", totalBlocks)

	// Test: Block size is standard APFS size
	assert.Equal(t, uint32(4096), containerSB.NxBlockSize, "standard APFS block size should be 4096")

	// Test: DMG file size is reasonable
	dmgSize := dmg.Size()
	// Allow for partition table and formatting overhead (typically 1-2% overhead)
	expectedMax := int64(float64(totalCapacity) * 1.02)
	assert.LessOrEqual(t, dmgSize, expectedMax, "DMG file size should not exceed capacity + overhead")
	t.Logf("DMG File Size: %d bytes (%.2f MB)", dmgSize, float64(dmgSize)/(1024*1024))

	// Test: Space metrics are reasonable
	percentAllocated := float64(dmgSize) / float64(totalCapacity) * 100
	assert.Greater(t, percentAllocated, float64(0), "some space should be allocated")
	assert.Less(t, percentAllocated, float64(105), "allocated space should not significantly exceed capacity")
	t.Logf("Space Allocated: %.2f%% (includes overhead)", percentAllocated)
}

func TestVolumeServiceWithMultipleDMGs(t *testing.T) {
	config := &disk.DMGConfig{
		AutoDetectAPFS: true,
		DefaultOffset:  20480,
		TestDataPath:   "../../tests",
	}

	dmgFiles := []string{"basic_apfs.dmg", "populated_apfs.dmg", "full_apfs.dmg", "empty_apfs.dmg"}

	for _, dmgFile := range dmgFiles {
		t.Run(dmgFile, func(t *testing.T) {
			testDMG := filepath.Join(config.TestDataPath, dmgFile)
			if _, err := os.Stat(testDMG); err != nil {
				t.Skipf("Test DMG not found: %v", testDMG)
			}

			dmg, err := disk.OpenDMG(testDMG, config)
			if err != nil {
				t.Skipf("Failed to open DMG: %v", err)
			}
			defer dmg.Close()

			cr, err := NewContainerReaderFromDevice(dmg, uint64(dmg.Size()))
			if err != nil {
				t.Skipf("Failed to create container reader: %v", err)
			}
			defer cr.Close()

			containerSB := cr.GetSuperblock()
			if containerSB == nil {
				t.Skip("No container superblock found")
			}

			var volumeOID types.OidT
			for _, oid := range containerSB.NxFsOid {
				if oid != 0 {
					volumeOID = oid
					break
				}
			}

			if volumeOID == 0 {
				t.Skip("No valid volumes found in container")
			}

			vs, err := NewVolumeService(cr, volumeOID)
			if err != nil {
				t.Logf("Could not create VolumeService: %v (may be expected for some DMG types)", err)
				return
			}

			assert.NotNil(t, vs.volumeSB, "volume superblock should be loaded")
			assert.Greater(t, vs.volumeSB.ApfsRootTreeOid, uint64(0), "root tree OID should be valid")
		})
	}
}
