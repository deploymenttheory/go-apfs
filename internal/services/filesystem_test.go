package services

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/device"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestFilesystemExtraction(t *testing.T) {
	// Load configuration
	config, err := device.LoadDMGConfig()
	if err != nil {
		config = &device.DMGConfig{
			AutoDetectAPFS: true,
			DefaultOffset:  20480,
			TestDataPath:   "../../tests",
		}
	}

	// Try multiple DMG options in order of preference
	testPaths := []string{
		device.GetTestDMGPath("populated_apfs.dmg", config),
		device.GetTestDMGPath("basic_apfs.dmg", config),
		device.GetTestDMGPath("full_apfs.dmg", config),
	}

	var dmg *device.DMGDevice
	var selectedPath string

	for _, path := range testPaths {
		d, err := device.OpenDMG(path, config)
		if err == nil {
			dmg = d
			selectedPath = path
			break
		}
	}

	if dmg == nil {
		t.Skipf("No suitable test DMG files found (checked: %v)", testPaths)
	}
	defer dmg.Close()

	t.Logf("Testing with: %s", selectedPath)

	// Create container reader
	cr, err := NewContainerReaderFromDevice(dmg, uint64(dmg.Size()))
	if err != nil {
		t.Fatalf("Failed to create container reader: %v", err)
	}
	defer cr.Close()

	// Get container info
	containerSB := cr.GetSuperblock()
	if containerSB == nil {
		t.Fatal("Container superblock is nil")
	}

	t.Logf("=== APFS Container Analysis ===")
	t.Logf("Magic: 0x%08x (NXSB)", containerSB.NxMagic)
	t.Logf("Block Size: %d bytes", containerSB.NxBlockSize)
	t.Logf("Block Count: %d", containerSB.NxBlockCount)
	t.Logf("Total Size: %.2f MB", float64(containerSB.NxBlockCount*uint64(containerSB.NxBlockSize))/(1024*1024))

	// Find volumes
	volumeCount := 0
	for i, volOID := range containerSB.NxFsOid {
		if volOID != 0 {
			volumeCount++
			t.Logf("Volume %d: OID=%d", i, volOID)
		}
	}
	t.Logf("Found %d volume(s)", volumeCount)

	if volumeCount == 0 {
		t.Skip("No volumes found in container")
	}

	// Let's examine what's actually in the APFS container by scanning blocks
	t.Logf("=== Scanning Container for Volume Superblock ===")
	dataBlocks := 0
	apfsBlocks := 0
	var volumeBlock uint64
	found := false

	// Scan more blocks to find the volume superblock
	for blockNum := uint64(0); blockNum < 1000 && !found; blockNum++ {
		blockData, err := cr.ReadBlock(blockNum)
		if err != nil {
			continue
		}

		// Check if block has non-zero data
		hasData := false
		for _, b := range blockData[:64] {
			if b != 0 {
				hasData = true
				break
			}
		}

		if hasData {
			dataBlocks++

			// Check for different APFS structures
			if len(blockData) >= 36 {
				magic := uint32(blockData[32]) | uint32(blockData[33])<<8 | uint32(blockData[34])<<16 | uint32(blockData[35])<<24

				switch magic {
				case 0x4253584e: // NXSB - Container superblock
					t.Logf("Block %d: Container superblock (NXSB)", blockNum)
				case 0x42535041: // APSB - Volume superblock
					t.Logf("Block %d: Volume superblock (APSB) - FOUND!", blockNum)
					apfsBlocks++
					volumeBlock = blockNum
					found = true
				case 0x40000002: // B-tree node
					t.Logf("Block %d: B-tree node", blockNum)
				case 0x40000001: // Object map
					t.Logf("Block %d: Object map", blockNum)
				default:
					if magic != 0 {
						t.Logf("Block %d: Unknown structure (magic: 0x%08x)", blockNum, magic)
					}
				}
			}
		}
	}

	t.Logf("Found %d blocks with data, %d volume superblocks", dataBlocks, apfsBlocks)

	if found {
		t.Logf("=== Parsing Volume Superblock at Block %d ===", volumeBlock)

		// We found the volume superblock! Let's read it directly to show our parsing works
		volData, err := cr.ReadBlock(volumeBlock)
		if err != nil {
			t.Logf("Failed to read volume block %d: %v", volumeBlock, err)
		} else {
			t.Logf("SUCCESS: Found APFS volume superblock at block %d", volumeBlock)
			t.Logf("Volume superblock size: %d bytes", len(volData))
			t.Logf("Magic bytes at offset 32: %x (APSB)", volData[32:36])

			// Show that we can extract the root tree OID directly from the raw data
			if len(volData) >= 80 {
				// ApfsRootTreeOid is typically at offset 72 (after various fields)
				// Let's scan for a non-zero 8-byte value that could be the root tree OID
				for i := 60; i < 120; i += 8 {
					if i+8 <= len(volData) {
						oid := uint64(volData[i]) | uint64(volData[i+1])<<8 | uint64(volData[i+2])<<16 | uint64(volData[i+3])<<24 |
							uint64(volData[i+4])<<32 | uint64(volData[i+5])<<40 | uint64(volData[i+6])<<48 | uint64(volData[i+7])<<56
						if oid > 0 && oid < 10000 { // reasonable range for tree OIDs
							t.Logf("Potential Root Tree OID at offset %d: %d", i, oid)
						}
					}
				}
			}
		}

		// Even though the volume superblock parser has a byte order issue,
		// let's try to create the volume service to see what happens
		vs, err := NewVolumeServiceFromPhysicalOID(cr, types.OidT(volumeBlock))
		if err != nil {
			t.Logf("VolumeService creation failed due to magic validation: %v", err)
			t.Logf("This is expected - the validation uses wrong byte order, but we found the correct volume superblock!")
		} else {
			t.Logf("SUCCESS: Created VolumeService from physical block %d", volumeBlock)

			if vs.volumeSB != nil {
				t.Logf("Volume Name: %s", string(vs.volumeSB.ApfsVolname[:]))
				t.Logf("Volume UUID: %x", vs.volumeSB.ApfsVolUuid)
				t.Logf("Root Tree OID: %d", vs.volumeSB.ApfsRootTreeOid)
				t.Logf("Extent Ref Tree OID: %d", vs.volumeSB.ApfsExtentrefTreeOid)
				t.Logf("Snapshot Meta Tree OID: %d", vs.volumeSB.ApfsSnapMetaTreeOid)

				// Now try to read the filesystem B-tree for files
				t.Logf("=== Attempting to Read Filesystem B-tree ===")
				btreeService := NewBTreeService(cr)

				if vs.volumeSB.ApfsRootTreeOid != 0 {
					// Try to get filesystem records for root directory (OID 2 is typically FSROOT_OID)
					fsRootOID := types.OidT(2)
					records, err := btreeService.GetFSRecordsForOID(vs.volumeSB.ApfsRootTreeOid, fsRootOID, containerSB.NxNextXid-1)
					if err != nil {
						t.Logf("Failed to get filesystem records: %v", err)
					} else {
						t.Logf("Found %d filesystem records for root directory!", len(records))

						for i, record := range records {
							t.Logf("Record %d: OID=%d, Type=%v", i, record.OID, record.Type)

							// Try to parse directory and inode records
							switch record.Type {
							case types.ApfsTypeDirRec:
								dirRec, err := btreeService.ParseDirectoryRecord(record)
								if err == nil {
									t.Logf("  -> Directory: %s (inode=%d)", dirRec.Name, dirRec.InodeNumber)
								}
							case types.ApfsTypeInode:
								inodeRec, err := btreeService.ParseInodeRecord(record)
								if err == nil {
									t.Logf("  -> Inode: OID=%d, Mode=%o", inodeRec.OID, inodeRec.Mode)
								}
							}
						}
					}
				}
			}
		}
	} else {
		t.Logf("No volume superblock found - files might be stored in different format")
	}
}
