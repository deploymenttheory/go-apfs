package services

import (
	"encoding/binary"
	"os"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/device"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestPhysicalVolumeAccess(t *testing.T) {
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

	t.Logf("Testing physical volume access with: %s", testPath)

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

	// Get container superblock
	containerSB := cr.GetSuperblock()
	if containerSB == nil {
		t.Fatal("Container superblock is nil")
	}

	volumeOID := containerSB.NxFsOid[0]
	t.Logf("Container reports volume OID: %d", volumeOID)

	if volumeOID == 0 {
		t.Fatal("No volume OID in container superblock")
	}

	// Test 1: Try to read the volume as a PHYSICAL object
	t.Logf("*** Attempting to read volume as PHYSICAL object at block %d ***", volumeOID)
	volumeData, err := cr.ReadBlock(uint64(volumeOID))
	if err != nil {
		t.Fatalf("Failed to read volume block %d: %v", volumeOID, err)
	}

	if len(volumeData) < 32 {
		t.Fatalf("Volume block too small: %d bytes", len(volumeData))
	}

	// Parse the object header
	objOID := binary.LittleEndian.Uint64(volumeData[8:16])
	objXID := binary.LittleEndian.Uint64(volumeData[16:24])
	objType := binary.LittleEndian.Uint32(volumeData[24:28])
	objSubtype := binary.LittleEndian.Uint32(volumeData[28:32])

	t.Logf("Object at block %d:", volumeOID)
	t.Logf("  OID: %d", objOID)
	t.Logf("  XID: %d", objXID)
	t.Logf("  Type: 0x%08X", objType)
	t.Logf("  Subtype: 0x%08X", objSubtype)

	// Check if this is a volume superblock
	baseType := objType & types.ObjectTypeMask
	flags := objType & types.ObjectTypeFlagsMask

	t.Logf("  Base Type: 0x%08X", baseType)
	t.Logf("  Flags: 0x%08X", flags)

	if flags&types.ObjPhysical != 0 {
		t.Logf("  ✓ Object is marked as PHYSICAL")
	}
	if flags&types.ObjVirtual == 0 && flags&types.ObjEphemeral == 0 {
		t.Logf("  ✓ Object storage type is not virtual or ephemeral")
	}

	if baseType == types.ObjectTypeFs {
		t.Logf("*** SUCCESS: Found APFS volume superblock as PHYSICAL object! ***")
		
		// Parse volume superblock magic
		if len(volumeData) >= 40 {
			// Volume superblock specific magic starts at offset 32
			volumeMagic := binary.BigEndian.Uint32(volumeData[32:36]) // APFS uses big-endian for magic
			t.Logf("Volume Magic: 0x%08X (expected: 0x42535041 for 'APSB')", volumeMagic)
			
			if volumeMagic == 0x42535041 { // 'APSB' in big-endian
				t.Logf("✓✓✓ CONFIRMED: Valid APFS volume superblock found!")
				
				// Parse key volume fields
				if len(volumeData) >= 200 {
					// Parse important volume fields (based on apple_apfs_subr structure)
					blockSize := binary.LittleEndian.Uint32(volumeData[36:40])
					blockCount := binary.LittleEndian.Uint64(volumeData[40:48])
					
					t.Logf("Volume Details:")
					t.Logf("  Block Size: %d bytes", blockSize)
					t.Logf("  Block Count: %d blocks", blockCount)
					
					// Look for the filesystem root tree OID (this would contain file/directory records)
					// This is typically around offset 120-140 in the volume superblock
					for offset := 120; offset <= 200; offset += 8 {
						if offset+8 <= len(volumeData) {
							oid := binary.LittleEndian.Uint64(volumeData[offset:offset+8])
							if oid != 0 && oid < uint64(containerSB.NxBlockCount) {
								t.Logf("  Potential FS Root Tree OID at offset %d: %d", offset, oid)
								
								// Try to read this as a B-tree
								if rootData, err := cr.ReadBlock(oid); err == nil && len(rootData) >= 32 {
									rootType := binary.LittleEndian.Uint32(rootData[24:28])
									rootBaseType := rootType & types.ObjectTypeMask
									if rootBaseType == types.ObjectTypeBtreeNode {
										t.Logf("    ✓ OID %d is a B-tree node - likely filesystem root!", oid)
									}
								}
							}
						}
					}
				}
			} else {
				t.Logf("⚠ Volume magic mismatch - may be corrupted or different format")
			}
		}
	} else {
		t.Logf("⚠ Object type 0x%08X is not a volume superblock (expected 0x%08X)", 
			baseType, types.ObjectTypeFs)
		
		// Let's see what other types we might find
		switch baseType {
		case types.ObjectTypeBtreeNode:
			t.Logf("  This is a B-tree node")
		case types.ObjectTypeOmap:
			t.Logf("  This is an object map")
		case types.ObjectTypeSpaceman:
			t.Logf("  This is a space manager")
		default:
			t.Logf("  Unknown object type")
		}
	}

	// Test 2: Try the next few blocks in case the volume superblock is nearby
	t.Logf("*** Scanning nearby blocks for volume superblock ***")
	for offset := int64(-5); offset <= 5; offset++ {
		blockAddr := int64(volumeOID) + offset
		if blockAddr < 0 || blockAddr >= int64(containerSB.NxBlockCount) {
			continue
		}
		
		if blockAddr == int64(volumeOID) {
			continue // Already tested this one
		}
		
		t.Logf("Checking block %d (offset %+d)...", blockAddr, offset)
		blockData, err := cr.ReadBlock(uint64(blockAddr))
		if err != nil {
			continue
		}
		
		if len(blockData) >= 36 {
			objType := binary.LittleEndian.Uint32(blockData[24:28])
			baseType := objType & types.ObjectTypeMask
			
			if baseType == types.ObjectTypeFs {
				t.Logf("*** FOUND volume superblock at block %d! ***", blockAddr)
				break
			}
		}
	}
}