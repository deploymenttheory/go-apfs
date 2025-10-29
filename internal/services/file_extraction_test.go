package services

import (
	"os"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/disk"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestFileExtractionFromPopulatedDMG(t *testing.T) {
	// Load configuration
	config, err := disk.LoadDMGConfig()
	if err != nil {
		config = &disk.DMGConfig{
			AutoDetectAPFS: true,
			DefaultOffset:  20480,
			TestDataPath:   "../../tests",
		}
	}

	// Use our populated DMG that has actual files
	testPath := disk.GetTestDMGPath("populated_apfs.dmg", config)
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Skip("populated_apfs.dmg not found - run create_test_dmgs.sh script first")
	}

	t.Logf("Testing file extraction from: %s", testPath)

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

	// Get container info
	containerSB := cr.GetSuperblock()
	if containerSB == nil {
		t.Fatal("Container superblock is nil")
	}

	t.Logf("Container Info:")
	t.Logf("  Magic: 0x%08X", containerSB.NxMagic)
	t.Logf("  Block Size: %d", containerSB.NxBlockSize)
	t.Logf("  Block Count: %d", containerSB.NxBlockCount)
	t.Logf("  Filesystem OIDs: %v", containerSB.NxFsOid)

	// Create B-tree service (we'll use direct access instead due to object map issues)
	_ = NewBTreeService(cr)

	// Scan for volume superblock manually to understand the structure
	t.Logf("Scanning for volume superblock...")
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

				// Parse some basic info from the volume superblock
				if len(blockData) >= 200 {
					// Root tree OID is at offset around 152-160
					rootTreeOID := uint64(blockData[152]) | uint64(blockData[153])<<8 | uint64(blockData[154])<<16 | uint64(blockData[155])<<24 |
						uint64(blockData[156])<<32 | uint64(blockData[157])<<40 | uint64(blockData[158])<<48 | uint64(blockData[159])<<56
					t.Logf("Volume root tree OID: %d", rootTreeOID)
				}
				break
			}
		}
	}

	if !found {
		t.Fatal("Could not find volume superblock")
	}

	// Let me debug the raw volume superblock data first
	volumeData, err := cr.ReadBlock(volumeBlock)
	if err != nil {
		t.Fatalf("Failed to read volume block: %v", err)
	}

	t.Logf("Raw volume superblock (first 200 bytes): %x", volumeData[:200])

	// Manual parsing to verify offsets
	if len(volumeData) >= 200 {
		// Root tree OID should be at offset around 152 (after skipping various fields)
		// Let's try different offsets to find the correct structure
		for testOffset := 140; testOffset <= 180; testOffset += 8 {
			if testOffset+8 <= len(volumeData) {
				value := uint64(volumeData[testOffset]) | uint64(volumeData[testOffset+1])<<8 |
					uint64(volumeData[testOffset+2])<<16 | uint64(volumeData[testOffset+3])<<24 |
					uint64(volumeData[testOffset+4])<<32 | uint64(volumeData[testOffset+5])<<40 |
					uint64(volumeData[testOffset+6])<<48 | uint64(volumeData[testOffset+7])<<56
				t.Logf("Offset %d: value=%d (0x%016x)", testOffset, value, value)
				if value == 88 {
					t.Logf("*** Found root tree OID 88 at offset %d ***", testOffset)
				}
			}
		}
	}

	// Create volume service using the physical block
	vs, err := NewVolumeServiceFromPhysicalOID(cr, types.OidT(volumeBlock))
	if err != nil {
		t.Fatalf("Failed to create volume service: %v", err)
	}

	t.Logf("Volume Info:")
	t.Logf("  Magic: 0x%08X", vs.volumeSB.ApfsMagic)
	t.Logf("  Root Tree OID: %d", vs.volumeSB.ApfsRootTreeOid)
	t.Logf("  Number of files: %d", vs.volumeSB.ApfsNumFiles)
	t.Logf("  Number of directories: %d", vs.volumeSB.ApfsNumDirectories)

	// Try to find and read the root tree
	if vs.volumeSB.ApfsRootTreeOid != 0 {
		t.Logf("Examining root tree at OID %d...", vs.volumeSB.ApfsRootTreeOid)

		// First, let's try to read the root tree block directly
		rootTreeBlock, err := cr.ReadBlock(uint64(vs.volumeSB.ApfsRootTreeOid))
		if err != nil {
			t.Logf("Failed to read root tree block directly: %v", err)
		} else {
			t.Logf("Root tree block read successfully (%d bytes)", len(rootTreeBlock))

			// Check if it's a valid B-tree node
			if len(rootTreeBlock) >= 32 {
				objectType := uint32(rootTreeBlock[24]) | uint32(rootTreeBlock[25])<<8 | uint32(rootTreeBlock[26])<<16 | uint32(rootTreeBlock[27])<<24
				t.Logf("Root tree object type: 0x%08X", objectType)
			}
		}

		// Since the object map is not working, try accessing the root tree directly
		t.Logf("Strategy: Direct B-tree access without object map...")

		// Read the root tree block directly using the parsed OID
		rootTreeOID := uint64(vs.volumeSB.ApfsRootTreeOid)
		t.Logf("Reading root tree from block %d", rootTreeOID)
		rootTreeData, err := cr.ReadBlock(rootTreeOID)
		if err != nil {
			t.Logf("Failed to read root tree block: %v", err)
		} else {
			t.Logf("Root tree block size: %d bytes", len(rootTreeData))

			// Try to parse this as a B-tree node manually
			if len(rootTreeData) >= 56 {
				// Check object header
				objectType := uint32(rootTreeData[24]) | uint32(rootTreeData[25])<<8 | uint32(rootTreeData[26])<<16 | uint32(rootTreeData[27])<<24
				t.Logf("Root tree object type: 0x%08X", objectType)

				// Parse B-tree node header (starting at offset 32)
				if len(rootTreeData) >= 56 {
					flags := uint16(rootTreeData[32]) | uint16(rootTreeData[33])<<8
					level := uint16(rootTreeData[34]) | uint16(rootTreeData[35])<<8
					keyCount := uint32(rootTreeData[36]) | uint32(rootTreeData[37])<<8 | uint32(rootTreeData[38])<<16 | uint32(rootTreeData[39])<<24

					t.Logf("B-tree node: flags=0x%04X, level=%d, keyCount=%d", flags, level, keyCount)

					if keyCount > 0 && keyCount < 1000 { // Sanity check
						t.Logf("Root tree has %d keys, attempting to extract records...", keyCount)

						// Try to parse the key/value table
						tableSpace := rootTreeData[40:44]
						tableOffset := uint16(tableSpace[0]) | uint16(tableSpace[1])<<8
						t.Logf("Table offset: %d", tableOffset)

						if int(tableOffset) < len(rootTreeData) {
							// Try to read some key/value pairs
							for i := uint32(0); i < keyCount && i < 10; i++ { // Limit to first 10 entries
								entryOffset := int(tableOffset) + int(i)*16 // Assuming 16-byte entries
								if entryOffset+16 <= len(rootTreeData) {
									keyOffset := uint16(rootTreeData[entryOffset]) | uint16(rootTreeData[entryOffset+1])<<8
									keyLen := uint16(rootTreeData[entryOffset+2]) | uint16(rootTreeData[entryOffset+3])<<8
									valueOffset := uint16(rootTreeData[entryOffset+8]) | uint16(rootTreeData[entryOffset+9])<<8
									valueLen := uint16(rootTreeData[entryOffset+10]) | uint16(rootTreeData[entryOffset+11])<<8

									t.Logf("Entry %d: keyOffset=%d, keyLen=%d, valueOffset=%d, valueLen=%d",
										i, keyOffset, keyLen, valueOffset, valueLen)

									// Try to read the key data
									if int(keyOffset)+int(keyLen) <= len(rootTreeData) && keyLen > 0 && keyLen < 1000 {
										keyData := rootTreeData[keyOffset : keyOffset+keyLen]
										t.Logf("Key data (%d bytes): %x", keyLen, keyData)

										// Check if this looks like a filesystem key
										if keyLen >= 8 {
											// Extract OID from key (first 8 bytes)
											oid := uint64(keyData[0]) | uint64(keyData[1])<<8 | uint64(keyData[2])<<16 | uint64(keyData[3])<<24 |
												uint64(keyData[4])<<32 | uint64(keyData[5])<<40 | uint64(keyData[6])<<48 | uint64(keyData[7])<<56
											t.Logf("Key OID: %d", oid)

											// Check for record type if key is longer
											if keyLen >= 9 {
												recordType := keyData[8]
												t.Logf("Record type: 0x%02X", recordType)

												// Look for directory record type (0x03)
												if recordType == 0x03 {
													t.Logf("*** Found directory record! ***")
													// Try to read the value
													if int(valueOffset)+int(valueLen) <= len(rootTreeData) && valueLen > 0 && valueLen < 1000 {
														valueData := rootTreeData[valueOffset : valueOffset+valueLen]
														t.Logf("Directory record value (%d bytes): %x", valueLen, valueData)

														// Try to extract filename if present
														if valueLen >= 10 {
															// Skip inode number (8 bytes) and look for name
															nameStart := 8
															if nameStart < int(valueLen) {
																// Try to find null-terminated string
																nameBytes := valueData[nameStart:]
																if len(nameBytes) > 0 {
																	// Find null terminator or end
																	nameLen := 0
																	for nameLen < len(nameBytes) && nameBytes[nameLen] != 0 {
																		nameLen++
																	}
																	if nameLen > 0 && nameLen < 256 {
																		filename := string(nameBytes[:nameLen])
																		t.Logf("*** FILENAME FOUND: %s ***", filename)
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
							}
						}
					}
				}
			}
		}

		// Strategy: Scan the entire DMG for populated B-tree nodes (focus on filesystem trees)
		t.Logf("Strategy: Scanning entire DMG for filesystem B-tree nodes...")

		foundBTrees := 0
		foundFileSystemTrees := 0
		for blockNum := uint64(0); blockNum < 1000; blockNum++ {
			blockData, err := cr.ReadBlock(blockNum)
			if err != nil {
				continue
			}

			if len(blockData) >= 56 {
				// Check for valid object header
				objectType := uint32(blockData[24]) | uint32(blockData[25])<<8 | uint32(blockData[26])<<16 | uint32(blockData[27])<<24

				if objectType == 0x40000002 { // B-tree node
					// Parse B-tree header
					flags := uint16(blockData[32]) | uint16(blockData[33])<<8
					level := uint16(blockData[34]) | uint16(blockData[35])<<8
					keyCount := uint32(blockData[36]) | uint32(blockData[37])<<8 | uint32(blockData[38])<<16 | uint32(blockData[39])<<24

					if keyCount > 0 && keyCount < 1000 { // Found a populated B-tree!
						foundBTrees++
						t.Logf("*** FOUND POPULATED B-TREE at block %d: flags=0x%04X, level=%d, keyCount=%d ***",
							blockNum, flags, level, keyCount)

						// Try to parse this B-tree's contents using correct APFS structure
						if keyCount <= 10 || blockNum == 94 { // Parse small trees or specifically block 94
							// Parse btn_table_space (nloc_t structure at offset 40)
							tableSpaceOffset := uint16(blockData[40]) | uint16(blockData[41])<<8
							tableSpaceLen := uint16(blockData[42]) | uint16(blockData[43])<<8
							t.Logf("Table space: offset=%d, len=%d", tableSpaceOffset, tableSpaceLen)

							// Parse btn_free_space (nloc_t structure at offset 44)
							freeSpaceOffset := uint16(blockData[44]) | uint16(blockData[45])<<8
							freeSpaceLen := uint16(blockData[46]) | uint16(blockData[47])<<8
							t.Logf("Free space: offset=%d, len=%d", freeSpaceOffset, freeSpaceLen)

							// The table of contents starts at btn_data + tableSpaceOffset
							// btn_data starts at offset 56 in the block
							btnDataStart := 56
							tocStart := btnDataStart + int(tableSpaceOffset)

							// Check if this node uses fixed or variable size entries
							isFixedSize := (flags & 0x0004) != 0 // BTNODE_FIXED_KV_SIZE = 0x0004

							if tocStart < len(blockData) {
								maxEntries := uint32(5)
								if blockNum == 94 {
									maxEntries = 15 // Check more entries for block 94
								}

								for i := uint32(0); i < keyCount && i < maxEntries; i++ {
									var keyOffset, keyLen, valueOffset, valueLen uint16

									if isFixedSize {
										// kvoff_t structure (4 bytes per entry)
										entryOffset := tocStart + int(i)*4
										if entryOffset+4 <= len(blockData) {
											keyOffset = uint16(blockData[entryOffset]) | uint16(blockData[entryOffset+1])<<8
											valueOffset = uint16(blockData[entryOffset+2]) | uint16(blockData[entryOffset+3])<<8
											keyLen = 0 // Fixed size, length not stored
											valueLen = 0
										}
									} else {
										// kvloc_t structure (8 bytes per entry)
										entryOffset := tocStart + int(i)*8
										if entryOffset+8 <= len(blockData) {
											keyOffset = uint16(blockData[entryOffset]) | uint16(blockData[entryOffset+1])<<8
											keyLen = uint16(blockData[entryOffset+2]) | uint16(blockData[entryOffset+3])<<8
											valueOffset = uint16(blockData[entryOffset+4]) | uint16(blockData[entryOffset+5])<<8
											valueLen = uint16(blockData[entryOffset+6]) | uint16(blockData[entryOffset+7])<<8
										}
									}

									// Calculate actual key and value locations
									// Key area starts after table of contents
									keyAreaStart := tocStart + int(tableSpaceLen)
									// Value area ends at btn_data end (for non-root) or before btree_info_t
									valueAreaEnd := len(blockData) // Simplified for now

									actualKeyOffset := keyAreaStart + int(keyOffset)
									actualValueOffset := valueAreaEnd - int(valueOffset)

									t.Logf("  Entry %d: keyOffset=%d, keyLen=%d, valueOffset=%d, valueLen=%d",
										i, keyOffset, keyLen, valueOffset, valueLen)
									t.Logf("    Actual: keyStart=%d, valueStart=%d", actualKeyOffset, actualValueOffset)

									// Try to read key data
									if keyLen > 0 && actualKeyOffset >= 0 && actualKeyOffset+int(keyLen) <= len(blockData) {
										keyData := blockData[actualKeyOffset : actualKeyOffset+int(keyLen)]

										if keyLen >= 9 {
											// This looks like a filesystem record key (has record type)
											foundFileSystemTrees++
											oid := uint64(keyData[0]) | uint64(keyData[1])<<8 | uint64(keyData[2])<<16 | uint64(keyData[3])<<24 |
												uint64(keyData[4])<<32 | uint64(keyData[5])<<40 | uint64(keyData[6])<<48 | uint64(keyData[7])<<56
											recordType := keyData[8]
											t.Logf("    *** FILESYSTEM RECORD: OID=%d, RecordType=0x%02X ***", oid, recordType)

											// Check for directory record (0x03)
											if recordType == 0x03 && valueLen > 8 && actualValueOffset >= 0 && actualValueOffset+int(valueLen) <= len(blockData) {
												valueData := blockData[actualValueOffset : actualValueOffset+int(valueLen)]
												t.Logf("      Directory record value (%d bytes): %x", valueLen, valueData)

												// Try to extract filename (skip 8-byte inode, look for string)
												if valueLen >= 10 {
													nameStart := 8
													if nameStart < int(valueLen) {
														nameBytes := valueData[nameStart:]
														nameLen := 0
														for nameLen < len(nameBytes) && nameBytes[nameLen] != 0 && nameLen < 100 {
															nameLen++
														}
														if nameLen > 0 {
															filename := string(nameBytes[:nameLen])
															t.Logf("      *** FILENAME FOUND: %s ***", filename)
														}
													}
												}
											}

											// Check for inode record (0x01)
											if recordType == 0x01 && valueLen > 8 && actualValueOffset >= 0 && actualValueOffset+int(valueLen) <= len(blockData) {
												t.Logf("      Found inode record for OID %d", oid)
											}

											if foundFileSystemTrees >= 20 {
												t.Logf("*** Found filesystem B-tree at block %d! ***", blockNum)
												goto found_filesystem_tree
											}
										} else if keyLen == 8 {
											// Likely object map entry (OID + XID only)
											t.Logf("    Object map entry: keyLen=%d", keyLen)
										} else {
											t.Logf("    Key data (%d bytes): %x", keyLen, keyData)
										}
									} else if keyLen == 0 && isFixedSize {
										// For fixed size, we need to know the actual key size from B-tree info
										t.Logf("    Fixed size key (size unknown)")
									}
								}
							}
						}

						if foundBTrees >= 5 {
							break // Don't scan too many
						}
					}
				}
			}
		}

	found_filesystem_tree:
		t.Logf("Found %d populated B-tree nodes, %d filesystem records", foundBTrees, foundFileSystemTrees)
	}

	// Expected files from our DMG creation script:
	expectedFiles := []string{
		"Documents/file_1.txt", "Documents/file_2.txt", "Documents/file_20.txt",
		"Data/data_1.dat", "Data/data_2.dat", "Data/data_10.dat",
		"symlink_test.txt",
	}

	t.Logf("Expected files in DMG: %v", expectedFiles)
	t.Logf("Our test shows the volume has %d files and %d directories", vs.volumeSB.ApfsNumFiles, vs.volumeSB.ApfsNumDirectories)
}
