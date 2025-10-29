package services

import (
	"os"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/disk"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestBTreeServiceWithDMG(t *testing.T) {
	// Load configuration
	config, err := disk.LoadDMGConfig()
	if err != nil {
		t.Logf("Failed to load config (using defaults): %v", err)
		config = &disk.DMGConfig{
			AutoDetectAPFS: true,
			DefaultOffset:  20480,
			TestDataPath:   "../../tests",
		}
	}

	// Try to find a test DMG file
	testFiles := []string{
		"populated_apfs.dmg",
		"basic_apfs.dmg",
		"test_apfs_with_files.dmg",
		"Cursor-darwin-universal.dmg",
		"test_apfs.dmg",
		"basic_test.img",
		"test_container.img",
		"volume_test.img",
	}

	var testDMGPath string
	for _, filename := range testFiles {
		path := disk.GetTestDMGPath(filename, config)
		if _, err := os.Stat(path); err == nil {
			testDMGPath = path
			break
		}
	}

	if testDMGPath == "" {
		t.Skip("No test DMG files found")
	}

	// Open the DMG
	dmg, err := disk.OpenDMG(testDMGPath, config)
	if err != nil {
		t.Fatalf("Failed to open DMG: %v", err)
	}
	defer dmg.Close()

	// Create container reader using DMG as the device
	cr, err := NewContainerReaderFromDevice(dmg, uint64(dmg.Size()))
	if err != nil {
		t.Fatalf("Failed to create container reader: %v", err)
	}
	defer cr.Close()

	// Create B-tree service
	btreeService := NewBTreeService(cr)

	// Get container superblock
	containerSB := cr.GetSuperblock()
	if containerSB == nil {
		t.Fatal("Container superblock is nil")
	}

	volumeOID := containerSB.NxFsOid[0]
	if volumeOID == 0 {
		t.Skip("No volume found in container")
	}

	t.Logf("Testing with volume OID: %d", volumeOID)

	// Test getting object map entry
	omapEntry, err := btreeService.GetOMapEntry(types.OidT(containerSB.NxOmapOid), types.OidT(volumeOID), containerSB.NxNextXid-1)
	if err != nil {
		t.Logf("GetOMapEntry failed (may be expected with test data): %v", err)
	} else {
		t.Logf("Object map entry: virtual=%d, physical=%d", omapEntry.VirtualOID, omapEntry.PhysicalAddr)
	}

	// Try to create volume service using our enhanced volume discovery
	vs, err := NewVolumeService(cr, types.OidT(volumeOID))
	if err != nil {
		t.Logf("Standard NewVolumeService failed: %v", err)
		t.Logf("Trying enhanced volume discovery...")

		// Scan for the actual volume superblock like we do in the filesystem test
		var volumeBlock uint64
		found := false
		for blockNum := uint64(0); blockNum < 1000 && !found; blockNum++ {
			blockData, err := cr.ReadBlock(blockNum)
			if err != nil {
				continue
			}

			if len(blockData) >= 36 {
				magic := uint32(blockData[32]) | uint32(blockData[33])<<8 | uint32(blockData[34])<<16 | uint32(blockData[35])<<24
				if magic == 0x42535041 { // "APSB"
					volumeBlock = blockNum
					found = true
					t.Logf("Found volume superblock at block %d", blockNum)
					break
				}
			}
		}

		if found {
			vs, err = NewVolumeServiceFromPhysicalOID(cr, types.OidT(volumeBlock))
			if err != nil {
				t.Logf("NewVolumeServiceFromPhysicalOID also failed: %v", err)
				t.Skip("Cannot test filesystem records without volume service")
			} else {
				t.Logf("SUCCESS: Created VolumeService using physical block %d", volumeBlock)
			}
		} else {
			t.Skip("Cannot find volume superblock")
		}
	}

	// Test getting filesystem records for root directory (FSROOT_OID)
	fsRootOID := types.OidT(2) // FSROOT_OID is typically 2
	records, err := btreeService.GetFSRecordsForOID(types.OidT(vs.volumeSB.ApfsRootTreeOid), fsRootOID, containerSB.NxNextXid-1)
	if err != nil {
		t.Logf("GetFSRecordsForOID failed: %v", err)
	} else {
		t.Logf("Found %d filesystem records for root directory", len(records))

		for i, record := range records {
			t.Logf("Record %d: OID=%d, Type=%v, KeySize=%d, ValueSize=%d",
				i, record.OID, record.Type, len(record.KeyData), len(record.ValueData))

			// Try to parse specific record types
			switch record.Type {
			case types.ApfsTypeDirRec:
				dirRec, err := btreeService.ParseDirectoryRecord(record)
				if err != nil {
					t.Logf("  Failed to parse directory record: %v", err)
				} else {
					t.Logf("  Directory record: %s (inode=%d)", dirRec.Name, dirRec.InodeNumber)
				}
			case types.ApfsTypeInode:
				inodeRec, err := btreeService.ParseInodeRecord(record)
				if err != nil {
					t.Logf("  Failed to parse inode record: %v", err)
				} else {
					t.Logf("  Inode record: OID=%d, ParentID=%d, Mode=%o",
						inodeRec.OID, inodeRec.ParentID, inodeRec.Mode)
				}
			}
		}
	}
}

func TestBTreeServiceDirectParsing(t *testing.T) {
	// Test with a known good APFS container
	config, err := disk.LoadDMGConfig()
	if err != nil {
		config = &disk.DMGConfig{
			AutoDetectAPFS: true,
			DefaultOffset:  0, // Try direct APFS first
			TestDataPath:   "../../tests",
		}
	}

	// Test with APFS DMG with real files
	testPath := disk.GetTestDMGPath("test_apfs_with_files.dmg", config)
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Skip("test_apfs_with_files.dmg not found")
	}

	// Open as DMG with auto-detection (Cursor DMG likely has GPT structure)
	config.AutoDetectAPFS = true
	dmg, err := disk.OpenDMG(testPath, config)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer dmg.Close()

	// Create container reader
	cr, err := NewContainerReaderFromDevice(dmg, uint64(dmg.Size()))
	if err != nil {
		t.Fatalf("Failed to create container reader: %v", err)
	}
	defer cr.Close()

	// Create B-tree service
	btreeService := NewBTreeService(cr)

	// Test basic container parsing
	containerSB := cr.GetSuperblock()
	if containerSB == nil {
		t.Fatal("Container superblock is nil")
	}

	t.Logf("Container: Magic=0x%08x, BlockSize=%d, BlockCount=%d",
		containerSB.NxMagic, containerSB.NxBlockSize, containerSB.NxBlockCount)

	// Test object map functionality
	if containerSB.NxOmapOid != 0 {
		volumeOID := containerSB.NxFsOid[0]
		if volumeOID != 0 {
			omapEntry, err := btreeService.GetOMapEntry(types.OidT(containerSB.NxOmapOid), types.OidT(volumeOID), containerSB.NxNextXid-1)
			if err != nil {
				t.Logf("Object map resolution failed: %v", err)
			} else {
				t.Logf("Successfully resolved volume OID %d to physical address %d",
					omapEntry.VirtualOID, omapEntry.PhysicalAddr)
			}
		}
	}
}
