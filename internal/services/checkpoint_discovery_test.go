package services

import (
	"os"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/device"
)

func TestCheckpointDiscovery(t *testing.T) {
	// Load configuration
	config, err := device.LoadDMGConfig()
	if err != nil {
		config = &device.DMGConfig{
			AutoDetectAPFS: true,
			DefaultOffset:  20480,
			TestDataPath:   "../../tests",
		}
	}

	// Use our populated DMG
	testPath := device.GetTestDMGPath("populated_apfs.dmg", config)
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Skip("populated_apfs.dmg not found")
	}

	t.Logf("Testing checkpoint discovery with: %s", testPath)

	// Open the DMG
	dmg, err := device.OpenDMG(testPath, config)
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

	// Get the block zero superblock for comparison
	blockZeroSB := cr.GetSuperblock()
	if blockZeroSB == nil {
		t.Fatal("Container superblock is nil")
	}

	t.Logf("Block Zero Superblock:")
	t.Logf("  Magic: 0x%08X", blockZeroSB.NxMagic)
	t.Logf("  XID: %d", blockZeroSB.NxNextXid)
	t.Logf("  Object Map OID: %d", blockZeroSB.NxOmapOid)
	t.Logf("  Checkpoint Desc Base: %d", blockZeroSB.NxXpDescBase)
	t.Logf("  Checkpoint Desc Blocks: %d", blockZeroSB.NxXpDescBlocks)

	// Create checkpoint discovery service
	checkpointService := NewCheckpointDiscoveryService(cr)

	// Find the latest valid superblock
	t.Logf("*** Starting checkpoint discovery ***")
	latestCandidate, err := checkpointService.FindLatestValidSuperblock()
	if err != nil {
		t.Fatalf("Failed to find latest valid superblock: %v", err)
	}

	if latestCandidate == nil {
		t.Fatal("No checkpoint candidate returned")
	}

	t.Logf("*** Latest Valid Superblock Found ***")
	t.Logf("  Block Address: %d", latestCandidate.BlockAddress)
	t.Logf("  Transaction ID: %d", latestCandidate.TransactionID)
	t.Logf("  Is Valid: %t", latestCandidate.IsValid)
	if latestCandidate.ErrorMsg != "" {
		t.Logf("  Error: %s", latestCandidate.ErrorMsg)
	}

	if latestCandidate.Superblock != nil {
		sb := latestCandidate.Superblock
		t.Logf("  Magic: 0x%08X", sb.NxMagic)
		t.Logf("  Object Map OID: %d", sb.NxOmapOid)
		t.Logf("  Volume OID: %d", sb.NxFsOid[0])
		
		// Compare with block zero
		if sb.NxNextXid != blockZeroSB.NxNextXid {
			t.Logf("*** DIFFERENT XID: Checkpoint=%d vs BlockZero=%d ***", 
				sb.NxNextXid, blockZeroSB.NxNextXid)
		} else {
			t.Logf("*** Same XID as block zero - no checkpoint differences ***")
		}
		
		if sb.NxOmapOid != blockZeroSB.NxOmapOid {
			t.Logf("*** DIFFERENT OMAP: Checkpoint=%d vs BlockZero=%d ***", 
				sb.NxOmapOid, blockZeroSB.NxOmapOid)
		}
	}

	// Test if the new superblock gives us better object map resolution
	if latestCandidate.Superblock != nil && latestCandidate.Superblock.NxOmapOid != blockZeroSB.NxOmapOid {
		t.Logf("*** Testing object map with checkpoint-discovered superblock ***")
		
		// Test the new object map
		resolver := NewBTreeObjectResolver(cr)
		volumeOID := latestCandidate.Superblock.NxFsOid[0]
		if volumeOID != 0 {
			t.Logf("Trying to resolve volume OID %d with checkpoint superblock", volumeOID)
			physAddr, err := resolver.ResolveVirtualObject(volumeOID, latestCandidate.Superblock.NxNextXid-1)
			if err != nil {
				t.Logf("Failed to resolve with checkpoint superblock: %v", err)
			} else {
				t.Logf("SUCCESS: Resolved to physical address %d using checkpoint!", physAddr)
			}
		}
	}
}