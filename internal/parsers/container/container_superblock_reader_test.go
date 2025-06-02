package container

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestContainerSuperblockData creates test container superblock data
func createTestContainerSuperblockData(magic uint32, blockSize uint32, blockCount uint64, features uint64, uuid types.UUID, nextOID types.OidT, nextXID types.XidT, maxFileSystems uint32, endian binary.ByteOrder) []byte {
	// Calculate required size:
	// Object header: 32 bytes
	// Container superblock fixed fields: ~156 bytes (up to volume OIDs)
	// Volume OIDs: 100 * 8 = 800 bytes
	// Counters: 32 * 8 = 256 bytes
	// Blocked out range: 16 bytes
	// Various fields: ~64 bytes
	// Ephemeral info: 4 * 8 = 32 bytes
	// Final fields: ~48 bytes
	// Total: ~1400 bytes
	data := make([]byte, 1500) // Conservative size for container superblock

	// Object header (32 bytes)
	copy(data[0:8], []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) // Checksum
	endian.PutUint64(data[8:16], uint64(1000))                              // OID
	endian.PutUint64(data[16:24], uint64(2000))                             // XID
	endian.PutUint32(data[24:28], types.ObjectTypeNxSuperblock)             // Type
	endian.PutUint32(data[28:32], 0)                                        // Subtype

	// Container superblock fields
	endian.PutUint32(data[32:36], magic)
	endian.PutUint32(data[36:40], blockSize)
	endian.PutUint64(data[40:48], blockCount)
	endian.PutUint64(data[48:56], features)
	endian.PutUint64(data[56:64], 0) // readonly compatible features
	endian.PutUint64(data[64:72], 0) // incompatible features

	// UUID
	copy(data[72:88], uuid[:])

	// Object and transaction IDs
	endian.PutUint64(data[88:96], uint64(nextOID))
	endian.PutUint64(data[96:104], uint64(nextXID))

	// Checkpoint fields
	endian.PutUint32(data[104:108], 100)  // desc blocks
	endian.PutUint32(data[108:112], 200)  // data blocks
	endian.PutUint64(data[112:120], 1000) // desc base
	endian.PutUint64(data[120:128], 2000) // data base
	endian.PutUint32(data[128:132], 10)   // desc next
	endian.PutUint32(data[132:136], 20)   // data next
	endian.PutUint32(data[136:140], 5)    // desc index
	endian.PutUint32(data[140:144], 50)   // desc len
	endian.PutUint32(data[144:148], 15)   // data index
	endian.PutUint32(data[148:152], 150)  // data len

	// Critical OIDs
	endian.PutUint64(data[152:160], 3000) // spaceman OID
	endian.PutUint64(data[160:168], 4000) // omap OID
	endian.PutUint64(data[168:176], 5000) // reaper OID

	// Testing and filesystem management
	endian.PutUint32(data[176:180], 0)              // test type
	endian.PutUint32(data[180:184], maxFileSystems) // max filesystems

	// Volume OIDs array (NxMaxFileSystems * 8 bytes)
	offset := 184
	for i := 0; i < types.NxMaxFileSystems; i++ {
		if i < int(maxFileSystems) && i < 3 { // Respect maxFileSystems parameter and limit to 3 for testing
			endian.PutUint64(data[offset:offset+8], uint64(6000+i))
		} else {
			endian.PutUint64(data[offset:offset+8], 0) // Invalid/empty volume
		}
		offset += 8
	}

	// Counters array (NxNumCounters * 8 bytes)
	for i := 0; i < types.NxNumCounters; i++ {
		endian.PutUint64(data[offset:offset+8], uint64(i*100))
		offset += 8
	}

	// Blocked out range
	endian.PutUint64(data[offset:offset+8], 7000)   // start paddr
	endian.PutUint64(data[offset+8:offset+16], 100) // block count
	offset += 16

	// Remaining fields with test values
	endian.PutUint64(data[offset:offset+8], 8000) // evict mapping tree OID
	offset += 8
	endian.PutUint64(data[offset:offset+8], 0x04) // flags (NxCryptoSw)
	offset += 8
	endian.PutUint64(data[offset:offset+8], 9000) // EFI jumpstart
	offset += 8

	// Fusion UUID (16 bytes)
	for i := 0; i < 16; i++ {
		data[offset+i] = byte(0x10 + i)
	}
	offset += 16

	// Keybag location
	endian.PutUint64(data[offset:offset+8], 10000) // start paddr
	endian.PutUint64(data[offset+8:offset+16], 50) // block count
	offset += 16

	// Ephemeral info array
	for i := 0; i < types.NxEphInfoCount; i++ {
		endian.PutUint64(data[offset:offset+8], uint64(11000+i*100))
		offset += 8
	}

	// Test OID, Fusion OIDs, and cache range
	endian.PutUint64(data[offset:offset+8], 12000) // test OID
	offset += 8
	endian.PutUint64(data[offset:offset+8], 13000) // fusion MT OID
	offset += 8
	endian.PutUint64(data[offset:offset+8], 14000) // fusion WBC OID
	offset += 8

	// Fusion WBC range
	endian.PutUint64(data[offset:offset+8], 15000) // start paddr
	endian.PutUint64(data[offset+8:offset+16], 25) // block count
	offset += 16

	// Final fields
	endian.PutUint64(data[offset:offset+8], 2) // newest mounted version
	offset += 8

	// Media key locker location
	endian.PutUint64(data[offset:offset+8], 16000) // start paddr
	endian.PutUint64(data[offset+8:offset+16], 75) // block count

	return data
}

// TestContainerSuperblockReader tests all container superblock reader method implementations
func TestContainerSuperblockReader(t *testing.T) {
	testCases := []struct {
		name                string
		magic               uint32
		blockSize           uint32
		blockCount          uint64
		features            uint64
		maxFileSystems      uint32
		expectValidMagic    bool
		expectedVolumeCount int
	}{
		{
			name:                "Valid Container Superblock",
			magic:               types.NxMagic,
			blockSize:           4096,
			blockCount:          1000000,
			features:            types.NxFeatureDefrag,
			maxFileSystems:      10,
			expectValidMagic:    true,
			expectedVolumeCount: 3, // Based on our test data
		},
		{
			name:                "Large Container",
			magic:               types.NxMagic,
			blockSize:           8192,
			blockCount:          50000000,
			features:            types.NxFeatureDefrag | types.NxFeatureLcfd,
			maxFileSystems:      100,
			expectValidMagic:    true,
			expectedVolumeCount: 3,
		},
		{
			name:                "Minimum Size Container",
			magic:               types.NxMagic,
			blockSize:           types.NxMinimumBlockSize,
			blockCount:          types.NxMinimumContainerSize / types.NxMinimumBlockSize,
			features:            0,
			maxFileSystems:      1,
			expectValidMagic:    true,
			expectedVolumeCount: 1, // Only 1 volume since maxFileSystems is 1
		},
	}

	endian := binary.LittleEndian

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uuid := types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
			data := createTestContainerSuperblockData(tc.magic, tc.blockSize, tc.blockCount, tc.features, uuid, 5000, 6000, tc.maxFileSystems, endian)

			csr, err := NewContainerSuperblockReader(data, endian)
			if !tc.expectValidMagic {
				if err == nil {
					t.Error("NewContainerSuperblockReader() should have failed with invalid magic")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
			}

			// Test basic properties
			if magic := csr.Magic(); magic != tc.magic {
				t.Errorf("Magic() = 0x%08X, want 0x%08X", magic, tc.magic)
			}

			if blockSize := csr.BlockSize(); blockSize != tc.blockSize {
				t.Errorf("BlockSize() = %d, want %d", blockSize, tc.blockSize)
			}

			if blockCount := csr.BlockCount(); blockCount != tc.blockCount {
				t.Errorf("BlockCount() = %d, want %d", blockCount, tc.blockCount)
			}

			if maxFS := csr.MaxFileSystems(); maxFS != tc.maxFileSystems {
				t.Errorf("MaxFileSystems() = %d, want %d", maxFS, tc.maxFileSystems)
			}

			// Test UUID
			containerUUID := csr.UUID()
			if containerUUID != uuid {
				t.Errorf("UUID() = %v, want %v", containerUUID, uuid)
			}

			// Test object identifiers
			if nextOID := csr.NextObjectID(); nextOID != 5000 {
				t.Errorf("NextObjectID() = %d, want 5000", nextOID)
			}

			if nextXID := csr.NextTransactionID(); nextXID != 6000 {
				t.Errorf("NextTransactionID() = %d, want 6000", nextXID)
			}

			if spacemanOID := csr.SpaceManagerOID(); spacemanOID != 3000 {
				t.Errorf("SpaceManagerOID() = %d, want 3000", spacemanOID)
			}

			if omapOID := csr.ObjectMapOID(); omapOID != 4000 {
				t.Errorf("ObjectMapOID() = %d, want 4000", omapOID)
			}

			if reaperOID := csr.ReaperOID(); reaperOID != 5000 {
				t.Errorf("ReaperOID() = %d, want 5000", reaperOID)
			}

			// Test volume OIDs
			volumeOIDs := csr.VolumeOIDs()
			if len(volumeOIDs) != tc.expectedVolumeCount {
				t.Errorf("VolumeOIDs() count = %d, want %d", len(volumeOIDs), tc.expectedVolumeCount)
			}

			for i, oid := range volumeOIDs {
				expectedOID := types.OidT(6000 + i)
				if oid != expectedOID {
					t.Errorf("VolumeOIDs()[%d] = %d, want %d", i, oid, expectedOID)
				}
			}

			// Test EFI jumpstart
			if efiJump := csr.EFIJumpstart(); efiJump != 9000 {
				t.Errorf("EFIJumpstart() = %d, want 9000", efiJump)
			}

			// Test Fusion UUID
			fusionUUID := csr.FusionUUID()
			expectedFusionUUID := types.UUID{0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F}
			if fusionUUID != expectedFusionUUID {
				t.Errorf("FusionUUID() = %v, want %v", fusionUUID, expectedFusionUUID)
			}

			// Test Keybag location
			keylocker := csr.KeylockerLocation()
			if keylocker.PrStartPaddr != 10000 || keylocker.PrBlockCount != 50 {
				t.Errorf("KeylockerLocation() = {PrStartPaddr: %d, PrBlockCount: %d}, want {PrStartPaddr: 10000, PrBlockCount: 50}", keylocker.PrStartPaddr, keylocker.PrBlockCount)
			}

			// Test Media key location
			mediaKey := csr.MediaKeyLocation()
			if mediaKey.PrStartPaddr != 16000 || mediaKey.PrBlockCount != 75 {
				t.Errorf("MediaKeyLocation() = {PrStartPaddr: %d, PrBlockCount: %d}, want {PrStartPaddr: 16000, PrBlockCount: 75}", mediaKey.PrStartPaddr, mediaKey.PrBlockCount)
			}
		})
	}
}

// TestContainerSuperblockReader_ErrorCases tests error conditions
func TestContainerSuperblockReader_ErrorCases(t *testing.T) {
	endian := binary.LittleEndian

	testCases := []struct {
		name     string
		dataSize int
		magic    uint32
	}{
		{"Empty data", 0, types.NxMagic},
		{"Too small data", 500, types.NxMagic},
		{"Minimum size - 1", 1023, types.NxMagic},
		{"Invalid magic", 1024, 0x12345678},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var data []byte
			if tc.dataSize > 0 {
				uuid := types.UUID{}
				data = createTestContainerSuperblockData(tc.magic, 4096, 1000000, 0, uuid, 1000, 2000, 10, endian)
				if tc.dataSize < len(data) {
					data = data[:tc.dataSize]
				}
			}

			_, err := NewContainerSuperblockReader(data, endian)
			if err == nil {
				t.Error("NewContainerSuperblockReader() should have failed")
			}
		})
	}
}

// TestContainerSuperblockReader_EndianHandling tests both little and big endian
func TestContainerSuperblockReader_EndianHandling(t *testing.T) {
	testCases := []struct {
		name   string
		endian binary.ByteOrder
	}{
		{"Little Endian", binary.LittleEndian},
		{"Big Endian", binary.BigEndian},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uuid := types.UUID{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00}
			data := createTestContainerSuperblockData(types.NxMagic, 8192, 5000000, types.NxFeatureDefrag, uuid, 0x1234567890ABCDEF, 0xFEDCBA0987654321, 25, tc.endian)

			csr, err := NewContainerSuperblockReader(data, tc.endian)
			if err != nil {
				t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
			}

			// Values should be parsed correctly regardless of endianness
			if magic := csr.Magic(); magic != types.NxMagic {
				t.Errorf("Magic() = 0x%08X, want 0x%08X", magic, types.NxMagic)
			}

			if blockSize := csr.BlockSize(); blockSize != 8192 {
				t.Errorf("BlockSize() = %d, want 8192", blockSize)
			}

			if blockCount := csr.BlockCount(); blockCount != 5000000 {
				t.Errorf("BlockCount() = %d, want 5000000", blockCount)
			}

			if nextOID := csr.NextObjectID(); nextOID != 0x1234567890ABCDEF {
				t.Errorf("NextObjectID() = 0x%016X, want 0x1234567890ABCDEF", nextOID)
			}

			if nextXID := csr.NextTransactionID(); nextXID != 0xFEDCBA0987654321 {
				t.Errorf("NextTransactionID() = 0x%016X, want 0xFEDCBA0987654321", nextXID)
			}
		})
	}
}

// TestContainerSuperblockReader_MagicValidation tests magic number validation
func TestContainerSuperblockReader_MagicValidation(t *testing.T) {
	endian := binary.LittleEndian
	uuid := types.UUID{}

	testCases := []struct {
		name        string
		magic       uint32
		shouldError bool
	}{
		{"Valid Magic", types.NxMagic, false},
		{"Invalid Magic 1", 0x00000000, true},
		{"Invalid Magic 2", 0xFFFFFFFF, true},
		{"Invalid Magic 3", 0x12345678, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := createTestContainerSuperblockData(tc.magic, 4096, 1000000, 0, uuid, 1000, 2000, 10, endian)

			_, err := NewContainerSuperblockReader(data, endian)
			if tc.shouldError && err == nil {
				t.Error("NewContainerSuperblockReader() should have failed with invalid magic")
			} else if !tc.shouldError && err != nil {
				t.Errorf("NewContainerSuperblockReader() failed unexpectedly: %v", err)
			}
		})
	}
}

// TestContainerSuperblockReader_VolumeOIDFiltering tests that only valid volume OIDs are returned
func TestContainerSuperblockReader_VolumeOIDFiltering(t *testing.T) {
	endian := binary.LittleEndian
	uuid := types.UUID{}
	data := createTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 1000, 2000, 10, endian)

	csr, err := NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	volumeOIDs := csr.VolumeOIDs()

	// Should only return non-zero volume OIDs
	for _, oid := range volumeOIDs {
		if oid == 0 {
			t.Error("VolumeOIDs() returned zero OID, should filter out invalid OIDs")
		}
	}

	// Based on our test data, we expect 3 valid volume OIDs
	if len(volumeOIDs) != 3 {
		t.Errorf("VolumeOIDs() count = %d, want 3", len(volumeOIDs))
	}
}

// Benchmark container superblock reader methods
func BenchmarkContainerSuperblockReader(b *testing.B) {
	endian := binary.LittleEndian
	uuid := types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	data := createTestContainerSuperblockData(types.NxMagic, 4096, 1000000, types.NxFeatureDefrag, uuid, 1000, 2000, 10, endian)
	csr, _ := NewContainerSuperblockReader(data, endian)

	b.Run("Magic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = csr.Magic()
		}
	})

	b.Run("BlockSize", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = csr.BlockSize()
		}
	})

	b.Run("UUID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = csr.UUID()
		}
	})

	b.Run("VolumeOIDs", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = csr.VolumeOIDs()
		}
	})
}
