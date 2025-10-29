package services

import (
	"os"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/disk"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestObjectMapResolution(t *testing.T) {
	// Load configuration
	config, err := disk.LoadDMGConfig()
	if err != nil {
		config = &disk.DMGConfig{
			AutoDetectAPFS: true,
			DefaultOffset:  20480,
			TestDataPath:   "../../tests",
		}
	}

	// Use our populated DMG
	testPath := disk.GetTestDMGPath("populated_apfs.dmg", config)
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Skip("populated_apfs.dmg not found")
	}

	t.Logf("Testing object map resolution with: %s", testPath)

	// Open the DMG
	dmg, err := disk.OpenDMG(testPath, config)
	if err != nil {
		t.Fatalf("Failed to open DMG: %v", err)
	}
	defer dmg.Close()

	// Create container reader
	cr, err := NewContainerReaderFromDevice(dmg, uint64(dmg.Size()))
	if err != nil {
		t.Fatalf("Failed to create container reader: %v", err)
	}
	defer cr.Close()

	// Get container superblock
	containerSB := cr.GetSuperblock()
	if containerSB == nil {
		t.Fatal("Container superblock is nil")
	}

	t.Logf("Container Info:")
	t.Logf("  Magic: 0x%08X", containerSB.NxMagic)
	t.Logf("  Container Object Map OID: %d", containerSB.NxOmapOid)
	t.Logf("  Next XID: %d", containerSB.NxNextXid)
	t.Logf("  Volume OID: %d", containerSB.NxFsOid[0])

	// Create B-tree object resolver
	resolver := NewBTreeObjectResolver(cr)

	// Test 1: Try to resolve the volume OID from container
	volumeOID := containerSB.NxFsOid[0]
	if volumeOID != 0 {
		t.Logf("*** Testing resolution of volume OID %d ***", volumeOID)
		physAddr, err := resolver.ResolveVirtualObject(types.OidT(volumeOID), containerSB.NxNextXid-1)
		if err != nil {
			t.Logf("Failed to resolve volume OID %d: %v", volumeOID, err)
		} else {
			t.Logf("SUCCESS: Volume OID %d resolved to physical address %d", volumeOID, physAddr)

			// Try to read the volume superblock at this physical address
			volumeData, err := cr.ReadBlock(uint64(physAddr))
			if err != nil {
				t.Logf("Failed to read volume data at address %d: %v", physAddr, err)
			} else {
				t.Logf("Volume data read successfully (%d bytes)", len(volumeData))

				// Check if it has valid volume superblock magic
				if len(volumeData) >= 36 {
					magic := uint32(volumeData[32]) | uint32(volumeData[33])<<8 | uint32(volumeData[34])<<16 | uint32(volumeData[35])<<24
					t.Logf("Volume superblock magic: 0x%08X (expected: 0x42535041)", magic)
					if magic == 0x42535041 {
						t.Logf("*** SUCCESS: Found valid volume superblock via object map! ***")
					}
				}
			}
		}
	}

	// Test 2: Try some other OIDs from our scan
	testOIDs := []types.OidT{85, 87, 88, 90, 94} // These were found during our B-tree scan
	for _, oid := range testOIDs {
		t.Logf("*** Testing resolution of OID %d ***", oid)
		physAddr, err := resolver.ResolveVirtualObject(oid, containerSB.NxNextXid-1)
		if err != nil {
			t.Logf("Failed to resolve OID %d: %v", oid, err)
		} else {
			t.Logf("SUCCESS: OID %d resolved to physical address %d", oid, physAddr)
		}
	}

	// Test 3: Test resolution with different transaction IDs
	if volumeOID != 0 {
		t.Logf("*** Testing resolution with different transaction IDs ***")
		for xid := types.XidT(1); xid <= containerSB.NxNextXid; xid++ {
			physAddr, err := resolver.ResolveVirtualObject(types.OidT(volumeOID), xid)
			if err == nil {
				t.Logf("XID %d: Volume OID %d -> Physical address %d", xid, volumeOID, physAddr)
			}
		}
	}
}
