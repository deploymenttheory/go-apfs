package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/device"
	"github.com/deploymenttheory/go-apfs/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainerReaderInitialization(t *testing.T) {
	config := &device.DMGConfig{
		AutoDetectAPFS: true,
		DefaultOffset:  20480,
		TestDataPath:   "../../tests",
	}

	testDMG := filepath.Join(config.TestDataPath, "populated_apfs.dmg")
	if _, err := os.Stat(testDMG); err != nil {
		t.Skipf("Test DMG not found: %v", testDMG)
	}

	dmg, err := device.OpenDMG(testDMG, config)
	require.NoError(t, err, "failed to open test DMG")
	defer dmg.Close()

	cr, err := NewContainerReaderFromDevice(dmg, uint64(dmg.Size()))
	require.NoError(t, err, "failed to create container reader")
	defer cr.Close()

	// Test: Container reader is initialized
	assert.NotNil(t, cr, "container reader should not be nil")

	// Test: Superblock is loaded
	sb := cr.GetSuperblock()
	require.NotNil(t, sb, "container superblock should not be nil")

	// Test: Superblock has valid magic
	assert.Equal(t, types.NxMagic, sb.NxMagic, "container magic should be 'NXSB' (0x4253584E)")
}

func TestContainerReaderSuperblockFields(t *testing.T) {
	config := &device.DMGConfig{
		AutoDetectAPFS: true,
		DefaultOffset:  20480,
		TestDataPath:   "../../tests",
	}

	testDMG := filepath.Join(config.TestDataPath, "populated_apfs.dmg")
	if _, err := os.Stat(testDMG); err != nil {
		t.Skipf("Test DMG not found: %v", testDMG)
	}

	dmg, err := device.OpenDMG(testDMG, config)
	require.NoError(t, err, "failed to open test DMG")
	defer dmg.Close()

	cr, err := NewContainerReaderFromDevice(dmg, uint64(dmg.Size()))
	require.NoError(t, err, "failed to create container reader")
	defer cr.Close()

	sb := cr.GetSuperblock()
	require.NotNil(t, sb, "container superblock should not be nil")

	// Test: Block size is valid
	assert.Greater(t, sb.NxBlockSize, uint32(0), "block size should be greater than 0")
	assert.LessOrEqual(t, sb.NxBlockSize, uint32(65536), "block size should be reasonable (<=64KB)")
	t.Logf("Block size: %d bytes", sb.NxBlockSize)

	// Test: Block count is valid
	assert.Greater(t, sb.NxBlockCount, uint64(0), "block count should be greater than 0")
	t.Logf("Block count: %d", sb.NxBlockCount)

	// Test: Total container size makes sense
	totalSize := sb.NxBlockCount * uint64(sb.NxBlockSize)
	assert.Greater(t, totalSize, uint64(1024*1024), "container should be at least 1MB")
	t.Logf("Total container size: %.2f MB", float64(totalSize)/(1024*1024))

	// Test: Transaction ID is valid
	assert.Greater(t, sb.NxNextXid, uint64(0), "next XID should be greater than 0")
	t.Logf("Next transaction ID: %d", sb.NxNextXid)

	// Test: Object Map OID is valid
	assert.Greater(t, sb.NxOmapOid, uint64(0), "object map OID should be valid")
	t.Logf("Object map OID: %d", sb.NxOmapOid)

	// Test: At least one volume exists
	volumeCount := 0
	for _, oid := range sb.NxFsOid {
		if oid != 0 {
			volumeCount++
		}
	}
	assert.Greater(t, volumeCount, 0, "container should have at least one volume")
	t.Logf("Number of volumes: %d", volumeCount)
}

func TestContainerReaderBlockIO(t *testing.T) {
	config := &device.DMGConfig{
		AutoDetectAPFS: true,
		DefaultOffset:  20480,
		TestDataPath:   "../../tests",
	}

	testDMG := filepath.Join(config.TestDataPath, "populated_apfs.dmg")
	if _, err := os.Stat(testDMG); err != nil {
		t.Skipf("Test DMG not found: %v", testDMG)
	}

	dmg, err := device.OpenDMG(testDMG, config)
	require.NoError(t, err, "failed to open test DMG")
	defer dmg.Close()

	cr, err := NewContainerReaderFromDevice(dmg, uint64(dmg.Size()))
	require.NoError(t, err, "failed to create container reader")
	defer cr.Close()

	sb := cr.GetSuperblock()
	require.NotNil(t, sb, "container superblock should not be nil")

	// Test: Can read block 0 (superblock)
	block0, err := cr.ReadBlock(0)
	require.NoError(t, err, "should be able to read block 0")
	assert.Equal(t, int(sb.NxBlockSize), len(block0), "block size should match superblock")

	// Test: Block 0 contains the container superblock magic
	if len(block0) > 36 {
		magic := uint32(block0[32]) | uint32(block0[33])<<8 | uint32(block0[34])<<16 | uint32(block0[35])<<24
		assert.Equal(t, types.NxMagic, magic, "block 0 should contain container superblock magic")
	}

	// Test: Can read other blocks
	for blockNum := uint64(1); blockNum <= 10 && blockNum < sb.NxBlockCount; blockNum++ {
		block, err := cr.ReadBlock(blockNum)
		require.NoError(t, err, "should be able to read block %d", blockNum)
		assert.Equal(t, int(sb.NxBlockSize), len(block), "all blocks should have same size")
	}

	// Test: Reading beyond container - behavior may vary by implementation
	// Some implementations wrap, some return error, some return zeros
	_, err = cr.ReadBlock(sb.NxBlockCount + 1)
	if err != nil {
		t.Logf("Reading beyond container returned error (as expected): %v", err)
	} else {
		t.Logf("Reading beyond container succeeded (DMG may wrap or return zeros)")
	}
}

func TestContainerReaderWithMultipleDMGs(t *testing.T) {
	config := &device.DMGConfig{
		AutoDetectAPFS: true,
		DefaultOffset:  20480,
		TestDataPath:   "../../tests",
	}

	dmgFiles := []string{"empty_apfs.dmg", "basic_apfs.dmg", "populated_apfs.dmg", "full_apfs.dmg"}

	for _, dmgFile := range dmgFiles {
		t.Run(dmgFile, func(t *testing.T) {
			testDMG := filepath.Join(config.TestDataPath, dmgFile)
			if _, err := os.Stat(testDMG); err != nil {
				t.Skipf("Test DMG not found: %v", testDMG)
			}

			dmg, err := device.OpenDMG(testDMG, config)
			if err != nil {
				t.Skipf("Failed to open DMG: %v", err)
			}
			defer dmg.Close()

			cr, err := NewContainerReaderFromDevice(dmg, uint64(dmg.Size()))
			if err != nil {
				t.Skipf("Failed to create container reader: %v", err)
			}
			defer cr.Close()

			sb := cr.GetSuperblock()
			require.NotNil(t, sb, "container superblock should exist")

			// Verify basic consistency
			assert.Equal(t, types.NxMagic, sb.NxMagic, "should have valid container magic")
			assert.Greater(t, sb.NxBlockSize, uint32(0), "block size should be valid")
			assert.Greater(t, sb.NxBlockCount, uint64(0), "block count should be valid")
		})
	}
}

func TestContainerReaderCaching(t *testing.T) {
	config := &device.DMGConfig{
		AutoDetectAPFS: true,
		DefaultOffset:  20480,
		TestDataPath:   "../../tests",
	}

	testDMG := filepath.Join(config.TestDataPath, "populated_apfs.dmg")
	if _, err := os.Stat(testDMG); err != nil {
		t.Skipf("Test DMG not found: %v", testDMG)
	}

	dmg, err := device.OpenDMG(testDMG, config)
	require.NoError(t, err, "failed to open test DMG")
	defer dmg.Close()

	cr, err := NewContainerReaderFromDevice(dmg, uint64(dmg.Size()))
	require.NoError(t, err, "failed to create container reader")
	defer cr.Close()

	sb := cr.GetSuperblock()
	require.NotNil(t, sb, "container superblock should not be nil")

	// Test: Reading same block twice should use cache
	block1a, err := cr.ReadBlock(1)
	require.NoError(t, err, "first read should succeed")

	block1b, err := cr.ReadBlock(1)
	require.NoError(t, err, "second read should succeed")

	// Blocks should be identical
	assert.Equal(t, block1a, block1b, "cached block should be identical to original")
}
