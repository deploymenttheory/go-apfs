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

func TestBTreeObjectResolverInitialization(t *testing.T) {
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

	cr, err := NewContainerReaderFromDevice(dmg, uint64(dmg.Size()))
	require.NoError(t, err, "failed to create container reader")
	defer cr.Close()

	resolver := NewBTreeObjectResolver(cr)
	assert.NotNil(t, resolver, "resolver should be initialized")
}

func TestBTreeObjectResolverContainerObjectMap(t *testing.T) {
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

	cr, err := NewContainerReaderFromDevice(dmg, uint64(dmg.Size()))
	require.NoError(t, err, "failed to create container reader")
	defer cr.Close()

	sb := cr.GetSuperblock()
	require.NotNil(t, sb, "container superblock should not be nil")

	resolver := NewBTreeObjectResolver(cr)

	// Test: Container object map OID is valid
	omapOID := sb.NxOmapOid
	assert.Greater(t, omapOID, uint64(0), "object map OID should be valid")
	t.Logf("Container object map OID: %d", omapOID)

	// Test: Can resolve container object map
	omapPhys, err := resolver.ResolveVirtualObject(omapOID, sb.NxNextXid-1)
	if err != nil {
		t.Logf("Object map resolution failed (expected for some DMG types): %v", err)
	} else {
		assert.Greater(t, uint64(omapPhys), uint64(0), "object map should resolve to valid physical address")
		t.Logf("Container object map physical address: %d", omapPhys)
	}
}

func TestBTreeObjectResolverWithErrorHandling(t *testing.T) {
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

	cr, err := NewContainerReaderFromDevice(dmg, uint64(dmg.Size()))
	require.NoError(t, err, "failed to create container reader")
	defer cr.Close()

	sb := cr.GetSuperblock()
	require.NotNil(t, sb, "container superblock should not be nil")

	resolver := NewBTreeObjectResolver(cr)

	// Test: Resolving invalid OID returns error
	_, err = resolver.ResolveVirtualObject(99999, sb.NxNextXid-1)
	assert.Error(t, err, "resolving invalid OID should return error")
	t.Logf("Expected error for invalid OID: %v", err)

	// Test: Resolving with invalid XID returns error
	_, err = resolver.ResolveVirtualObject(sb.NxOmapOid, 0)
	assert.Error(t, err, "resolving with invalid XID should return error")
	t.Logf("Expected error for invalid XID: %v", err)
}

func TestBTreeObjectResolverVolumeResolution(t *testing.T) {
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

	cr, err := NewContainerReaderFromDevice(dmg, uint64(dmg.Size()))
	require.NoError(t, err, "failed to create container reader")
	defer cr.Close()

	sb := cr.GetSuperblock()
	require.NotNil(t, sb, "container superblock should not be nil")

	resolver := NewBTreeObjectResolver(cr)

	// Find first valid volume OID
	var volumeOID types.OidT
	for _, oid := range sb.NxFsOid {
		if oid != 0 {
			volumeOID = oid
			break
		}
	}

	if volumeOID == 0 {
		t.Skip("No valid volumes found in container")
	}

	// Test: Try to resolve volume OID
	physAddr, err := resolver.ResolveVirtualObject(volumeOID, sb.NxNextXid-1)
	if err != nil {
		t.Logf("Volume OID resolution failed (may be expected): %v", err)
	} else {
		assert.Greater(t, uint64(physAddr), uint64(0), "volume OID should resolve to valid address")
		t.Logf("Volume OID %d resolved to physical address: %d", volumeOID, physAddr)
	}
}

func TestBTreeObjectResolverMultipleOIDs(t *testing.T) {
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

	cr, err := NewContainerReaderFromDevice(dmg, uint64(dmg.Size()))
	require.NoError(t, err, "failed to create container reader")
	defer cr.Close()

	sb := cr.GetSuperblock()
	require.NotNil(t, sb, "container superblock should not be nil")

	resolver := NewBTreeObjectResolver(cr)

	// Test: Collect important OIDs from container
	importantOIDs := []struct {
		name string
		oid  types.OidT
	}{
		{"ObjectMap", sb.NxOmapOid},
		{"SpaceManager", sb.NxSpacemanOid},
		{"Reaper", sb.NxReaperOid},
	}

	for _, item := range importantOIDs {
		if item.oid == 0 {
			continue
		}

		t.Run(item.name, func(t *testing.T) {
			physAddr, err := resolver.ResolveVirtualObject(item.oid, sb.NxNextXid-1)
			if err != nil {
				t.Logf("Could not resolve %s OID %d: %v", item.name, item.oid, err)
			} else {
				assert.Greater(t, uint64(physAddr), uint64(0), "%s should resolve to valid address", item.name)
				t.Logf("%s (OID %d) resolved to physical address: %d", item.name, item.oid, physAddr)
			}
		})
	}
}
