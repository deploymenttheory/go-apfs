package btrees

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestBTreeInfoData creates test B-tree info data
func createTestBTreeInfoData(flags, nodeSize, keySize, valSize uint32, longestKey, longestVal uint32, keyCount, nodeCount uint64, endian binary.ByteOrder) []byte {
	data := make([]byte, 48)

	// Fixed B-tree info (16 bytes)
	endian.PutUint32(data[0:4], flags)
	endian.PutUint32(data[4:8], nodeSize)
	endian.PutUint32(data[8:12], keySize)
	endian.PutUint32(data[12:16], valSize)

	// Variable B-tree info (32 bytes)
	endian.PutUint32(data[16:20], longestKey)
	endian.PutUint32(data[20:24], longestVal)
	endian.PutUint64(data[24:32], keyCount)
	endian.PutUint64(data[32:40], nodeCount)

	return data
}

// TestBTreeInfoReader tests all B-tree info reader method implementations
func TestBTreeInfoReader(t *testing.T) {
	testCases := []struct {
		name             string
		flags            uint32
		nodeSize         uint32
		keySize          uint32
		valSize          uint32
		longestKey       uint32
		longestVal       uint32
		keyCount         uint64
		nodeCount        uint64
		expectedHasU64   bool
		expectedSeqIns   bool
		expectedGhosts   bool
		expectedEphem    bool
		expectedPhys     bool
		expectedPersist  bool
		expectedAligned  bool
		expectedHashed   bool
		expectedNoHeader bool
	}{
		{
			name:             "Basic B-tree",
			flags:            0,
			nodeSize:         4096,
			keySize:          0, // Variable size
			valSize:          0, // Variable size
			longestKey:       256,
			longestVal:       1024,
			keyCount:         1000,
			nodeCount:        10,
			expectedHasU64:   false,
			expectedSeqIns:   false,
			expectedGhosts:   false,
			expectedEphem:    false,
			expectedPhys:     false,
			expectedPersist:  true, // Not non-persistent
			expectedAligned:  true, // Not non-aligned
			expectedHashed:   false,
			expectedNoHeader: false,
		},
		{
			name:             "Hashed B-tree with uint64 keys",
			flags:            types.BtreeUint64Keys | types.BtreeHashed | types.BtreePhysical,
			nodeSize:         8192,
			keySize:          8, // Fixed size uint64
			valSize:          16,
			longestKey:       8,
			longestVal:       16,
			keyCount:         5000,
			nodeCount:        50,
			expectedHasU64:   true,
			expectedSeqIns:   false,
			expectedGhosts:   false,
			expectedEphem:    false,
			expectedPhys:     true,
			expectedPersist:  true,
			expectedAligned:  true,
			expectedHashed:   true,
			expectedNoHeader: false,
		},
		{
			name:             "Ephemeral non-persistent B-tree",
			flags:            types.BtreeEphemeral | types.BtreeNonpersistent | types.BtreeKvNonaligned,
			nodeSize:         2048,
			keySize:          0,
			valSize:          0,
			longestKey:       512,
			longestVal:       2048,
			keyCount:         100,
			nodeCount:        5,
			expectedHasU64:   false,
			expectedSeqIns:   false,
			expectedGhosts:   false,
			expectedEphem:    true,
			expectedPhys:     false,
			expectedPersist:  false,
			expectedAligned:  false,
			expectedHashed:   false,
			expectedNoHeader: false,
		},
		{
			name:             "All flags enabled",
			flags:            types.BtreeUint64Keys | types.BtreeSequentialInsert | types.BtreeAllowGhosts | types.BtreeEphemeral | types.BtreeNonpersistent | types.BtreeKvNonaligned | types.BtreeHashed | types.BtreeNoheader,
			nodeSize:         4096,
			keySize:          8,
			valSize:          8,
			longestKey:       8,
			longestVal:       8,
			keyCount:         10000,
			nodeCount:        100,
			expectedHasU64:   true,
			expectedSeqIns:   true,
			expectedGhosts:   true,
			expectedEphem:    true,
			expectedPhys:     false,
			expectedPersist:  false,
			expectedAligned:  false,
			expectedHashed:   true,
			expectedNoHeader: true,
		},
	}

	endian := binary.LittleEndian

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := createTestBTreeInfoData(tc.flags, tc.nodeSize, tc.keySize, tc.valSize, tc.longestKey, tc.longestVal, tc.keyCount, tc.nodeCount, endian)
			bir, err := NewBTreeInfoReader(data, endian)
			if err != nil {
				t.Fatalf("NewBTreeInfoReader() failed: %v", err)
			}

			// Test basic properties
			if flags := bir.Flags(); flags != tc.flags {
				t.Errorf("Flags() = 0x%08X, want 0x%08X", flags, tc.flags)
			}

			if nodeSize := bir.NodeSize(); nodeSize != tc.nodeSize {
				t.Errorf("NodeSize() = %d, want %d", nodeSize, tc.nodeSize)
			}

			if keySize := bir.KeySize(); keySize != tc.keySize {
				t.Errorf("KeySize() = %d, want %d", keySize, tc.keySize)
			}

			if valSize := bir.ValueSize(); valSize != tc.valSize {
				t.Errorf("ValueSize() = %d, want %d", valSize, tc.valSize)
			}

			if longestKey := bir.LongestKey(); longestKey != tc.longestKey {
				t.Errorf("LongestKey() = %d, want %d", longestKey, tc.longestKey)
			}

			if longestVal := bir.LongestValue(); longestVal != tc.longestVal {
				t.Errorf("LongestValue() = %d, want %d", longestVal, tc.longestVal)
			}

			if keyCount := bir.KeyCount(); keyCount != tc.keyCount {
				t.Errorf("KeyCount() = %d, want %d", keyCount, tc.keyCount)
			}

			if nodeCount := bir.NodeCount(); nodeCount != tc.nodeCount {
				t.Errorf("NodeCount() = %d, want %d", nodeCount, tc.nodeCount)
			}

			// Test flag-based properties
			if hasU64 := bir.HasUint64Keys(); hasU64 != tc.expectedHasU64 {
				t.Errorf("HasUint64Keys() = %v, want %v", hasU64, tc.expectedHasU64)
			}

			if seqIns := bir.SupportsSequentialInsert(); seqIns != tc.expectedSeqIns {
				t.Errorf("SupportsSequentialInsert() = %v, want %v", seqIns, tc.expectedSeqIns)
			}

			if ghosts := bir.AllowsGhosts(); ghosts != tc.expectedGhosts {
				t.Errorf("AllowsGhosts() = %v, want %v", ghosts, tc.expectedGhosts)
			}

			if ephem := bir.IsEphemeral(); ephem != tc.expectedEphem {
				t.Errorf("IsEphemeral() = %v, want %v", ephem, tc.expectedEphem)
			}

			if phys := bir.IsPhysical(); phys != tc.expectedPhys {
				t.Errorf("IsPhysical() = %v, want %v", phys, tc.expectedPhys)
			}

			if persist := bir.IsPersistent(); persist != tc.expectedPersist {
				t.Errorf("IsPersistent() = %v, want %v", persist, tc.expectedPersist)
			}

			if aligned := bir.HasAlignedKV(); aligned != tc.expectedAligned {
				t.Errorf("HasAlignedKV() = %v, want %v", aligned, tc.expectedAligned)
			}

			if hashed := bir.IsHashed(); hashed != tc.expectedHashed {
				t.Errorf("IsHashed() = %v, want %v", hashed, tc.expectedHashed)
			}

			if noHeader := bir.HasHeaderlessNodes(); noHeader != tc.expectedNoHeader {
				t.Errorf("HasHeaderlessNodes() = %v, want %v", noHeader, tc.expectedNoHeader)
			}
		})
	}
}

// TestBTreeInfoReader_ErrorCases tests error conditions
func TestBTreeInfoReader_ErrorCases(t *testing.T) {
	endian := binary.LittleEndian

	testCases := []struct {
		name     string
		dataSize int
	}{
		{"Empty data", 0},
		{"Too small data", 20},
		{"Minimum size - 1", 47},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := make([]byte, tc.dataSize)
			_, err := NewBTreeInfoReader(data, endian)
			if err == nil {
				t.Error("NewBTreeInfoReader() should have failed with insufficient data")
			}
		})
	}
}

// TestBTreeInfoReader_MinimumSize tests with minimum valid size
func TestBTreeInfoReader_MinimumSize(t *testing.T) {
	endian := binary.LittleEndian
	data := createTestBTreeInfoData(0, 4096, 0, 0, 100, 200, 500, 5, endian)

	bir, err := NewBTreeInfoReader(data, endian)
	if err != nil {
		t.Fatalf("NewBTreeInfoReader() failed with minimum size data: %v", err)
	}

	if bir == nil {
		t.Error("NewBTreeInfoReader() returned nil with valid data")
	}
}

// Benchmark B-tree info reader methods
func BenchmarkBTreeInfoReader(b *testing.B) {
	endian := binary.LittleEndian
	data := createTestBTreeInfoData(types.BtreeUint64Keys|types.BtreeHashed, 4096, 8, 8, 8, 8, 1000, 10, endian)
	bir, _ := NewBTreeInfoReader(data, endian)

	b.Run("Flags", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bir.Flags()
		}
	})

	b.Run("HasUint64Keys", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bir.HasUint64Keys()
		}
	})

	b.Run("KeyCount", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bir.KeyCount()
		}
	})
}
