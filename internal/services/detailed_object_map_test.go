package services

import (
	"os"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/disk"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestDetailedObjectMapAnalysis(t *testing.T) {
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

	t.Logf("Testing detailed object map analysis with: %s", testPath)

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

	t.Logf("Container Analysis:")
	t.Logf("  Magic: 0x%08X", containerSB.NxMagic)
	t.Logf("  Container Object Map OID: %d", containerSB.NxOmapOid)
	t.Logf("  Next XID: %d", containerSB.NxNextXid)
	t.Logf("  Volume OID: %d", containerSB.NxFsOid[0])
	t.Logf("  Block Size: %d bytes", containerSB.NxBlockSize)
	t.Logf("  Block Count: %d blocks", containerSB.NxBlockCount)

	// Check all OIDs from the superblock and see what they resolve to
	oidsToCheck := []struct {
		name string
		oid  types.OidT
	}{
		{"Container Object Map", containerSB.NxOmapOid},
		{"Spaceman", containerSB.NxSpacemanOid},
		{"Reaper", containerSB.NxReaperOid},
		{"Volume", containerSB.NxFsOid[0]},
	}

	// Try to read these objects directly at their physical addresses
	for _, entry := range oidsToCheck {
		if entry.oid == 0 {
			t.Logf("%s: OID=0 (not used)", entry.name)
			continue
		}

		t.Logf("*** Analyzing %s (OID=%d) ***", entry.name, entry.oid)

		// Try to read the object at its physical block (assuming OID = physical address)
		blockData, err := cr.ReadBlock(uint64(entry.oid))
		if err != nil {
			t.Logf("  Failed to read block %d: %v", entry.oid, err)
			continue
		}

		if len(blockData) < 32 {
			t.Logf("  Block too small: %d bytes", len(blockData))
			continue
		}

		// Parse object header
		magic := blockData[32:36] // Magic is after the obj_phys_t header
		objType := uint32(blockData[24]) | uint32(blockData[25])<<8 | uint32(blockData[26])<<16 | uint32(blockData[27])<<24
		objOID := uint64(blockData[8]) | uint64(blockData[9])<<8 | uint64(blockData[10])<<16 | uint64(blockData[11])<<24 |
			uint64(blockData[12])<<32 | uint64(blockData[13])<<40 | uint64(blockData[14])<<48 | uint64(blockData[15])<<56
		objXID := uint64(blockData[16]) | uint64(blockData[17])<<8 | uint64(blockData[18])<<16 | uint64(blockData[19])<<24 |
			uint64(blockData[20])<<32 | uint64(blockData[21])<<40 | uint64(blockData[22])<<48 | uint64(blockData[23])<<56

		t.Logf("  Block contains object:")
		t.Logf("    Type: 0x%08X", objType)
		t.Logf("    OID: %d", objOID)
		t.Logf("    XID: %d", objXID)
		t.Logf("    Magic bytes: %02X %02X %02X %02X", magic[0], magic[1], magic[2], magic[3])

		// Check if this is the object map
		if entry.name == "Container Object Map" {
			// Extract the base type (remove flags)
			baseType := objType & types.ObjectTypeMask
			flags := objType & types.ObjectTypeFlagsMask

			t.Logf("    Base Type: 0x%08X, Flags: 0x%08X", baseType, flags)

			if flags&types.ObjPhysical != 0 {
				t.Logf("    ✓ Object is marked as PHYSICAL")
			}
			if flags&types.ObjVirtual == 0 && flags&types.ObjEphemeral == 0 && flags&types.ObjPhysical != 0 {
				t.Logf("    ✓ Physical object storage type confirmed")
			}

			// Try to parse as object map
			if baseType == types.ObjectTypeOmap {
				t.Logf("  ✓ Confirmed: This is an object map structure")

				// Parse additional object map fields
				if len(blockData) >= 72 {
					flags := uint32(blockData[36]) | uint32(blockData[37])<<8 | uint32(blockData[38])<<16 | uint32(blockData[39])<<24
					snapCount := uint32(blockData[40]) | uint32(blockData[41])<<8 | uint32(blockData[42])<<16 | uint32(blockData[43])<<24
					treeOID := uint64(blockData[56]) | uint64(blockData[57])<<8 | uint64(blockData[58])<<16 | uint64(blockData[59])<<24 |
						uint64(blockData[60])<<32 | uint64(blockData[61])<<40 | uint64(blockData[62])<<48 | uint64(blockData[63])<<56

					t.Logf("    Object Map Details:")
					t.Logf("      Flags: 0x%08X", flags)
					t.Logf("      Snapshot Count: %d", snapCount)
					t.Logf("      B-tree OID: %d", treeOID)

					// Try to read the B-tree
					if treeOID != 0 {
						t.Logf("  *** Examining Object Map B-tree at OID %d ***", treeOID)
						treeData, err := cr.ReadBlock(uint64(treeOID))
						if err != nil {
							t.Logf("    Failed to read B-tree block: %v", err)
						} else {
							if len(treeData) >= 32 {
								treeType := uint32(treeData[24]) | uint32(treeData[25])<<8 | uint32(treeData[26])<<16 | uint32(treeData[27])<<24
								t.Logf("    B-tree Type: 0x%08X", treeType)

								if treeType == types.ObjectTypeBtreeNode {
									t.Logf("    ✓ This is a B-tree node - analyzing...")

									// Parse B-tree node header to see entry count
									if len(treeData) >= 56 {
										flags := uint16(treeData[32]) | uint16(treeData[33])<<8
										level := uint16(treeData[34]) | uint16(treeData[35])<<8
										nkeys := uint32(treeData[36]) | uint32(treeData[37])<<8 | uint32(treeData[38])<<16 | uint32(treeData[39])<<24

										t.Logf("      Node Flags: 0x%04X", flags)
										t.Logf("      Node Level: %d", level)
										t.Logf("      Key Count: %d", nkeys)

										if nkeys == 0 {
											t.Logf("      ⚠ B-tree node is empty - this explains missing object mappings!")
										} else {
											t.Logf("      ✓ B-tree has %d entries", nkeys)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Let's also try scanning for ANY blocks that might contain object map entries
	t.Logf("*** Scanning for any B-tree nodes that might contain object mappings ***")
	totalBlocks := min(uint64(containerSB.NxBlockCount), 1000) // Limit scan to first 1000 blocks

	objectMapNodes := 0
	for blockNum := uint64(0); blockNum < totalBlocks; blockNum++ {
		blockData, err := cr.ReadBlock(blockNum)
		if err != nil {
			continue
		}

		if len(blockData) < 56 {
			continue
		}

		// Check if this looks like a B-tree node
		objType := uint32(blockData[24]) | uint32(blockData[25])<<8 | uint32(blockData[26])<<16 | uint32(blockData[27])<<24
		if objType == types.ObjectTypeBtreeNode {
			// Check if it's specifically an object map B-tree (subtype should be OBJECT_TYPE_OMAP)
			subtype := uint32(blockData[28]) | uint32(blockData[29])<<8 | uint32(blockData[30])<<16 | uint32(blockData[31])<<24

			if subtype == types.ObjectTypeOmap {
				objectMapNodes++
				nkeys := uint32(blockData[36]) | uint32(blockData[37])<<8 | uint32(blockData[38])<<16 | uint32(blockData[39])<<24
				t.Logf("  Found object map B-tree node at block %d with %d keys", blockNum, nkeys)
			}
		}
	}

	if objectMapNodes == 0 {
		t.Logf("  ⚠ No object map B-tree nodes found in first %d blocks", totalBlocks)
		t.Logf("  This confirms the object mappings are missing/empty")
	} else {
		t.Logf("  Found %d object map B-tree nodes", objectMapNodes)
	}
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
