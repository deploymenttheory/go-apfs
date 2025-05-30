package btrees

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestKVLocation creates a test KvlocT structure
func createTestKVLocation(keyOffset, keyLength, valueOffset, valueLength uint16) types.KvlocT {
	return types.KvlocT{
		K: types.NlocT{
			Off: keyOffset,
			Len: keyLength,
		},
		V: types.NlocT{
			Off: valueOffset,
			Len: valueLength,
		},
	}
}

// TestBTreeKVLocationReader tests all KV location reader method implementations
func TestBTreeKVLocationReader(t *testing.T) {
	testCases := []struct {
		name             string
		keyOffset        uint16
		keyLength        uint16
		valueOffset      uint16
		valueLength      uint16
		expectedKeyOff   uint16
		expectedKeyLen   uint16
		expectedValueOff uint16
		expectedValueLen uint16
	}{
		{
			name:             "Valid KV Location",
			keyOffset:        100,
			keyLength:        32,
			valueOffset:      200,
			valueLength:      64,
			expectedKeyOff:   100,
			expectedKeyLen:   32,
			expectedValueOff: 200,
			expectedValueLen: 64,
		},
		{
			name:             "Zero Offset Locations",
			keyOffset:        0,
			keyLength:        16,
			valueOffset:      0,
			valueLength:      32,
			expectedKeyOff:   0,
			expectedKeyLen:   16,
			expectedValueOff: 0,
			expectedValueLen: 32,
		},
		{
			name:             "Zero Length Values",
			keyOffset:        50,
			keyLength:        0,
			valueOffset:      100,
			valueLength:      0,
			expectedKeyOff:   50,
			expectedKeyLen:   0,
			expectedValueOff: 100,
			expectedValueLen: 0,
		},
		{
			name:             "Maximum Values",
			keyOffset:        ^uint16(0) - 1,
			keyLength:        ^uint16(0),
			valueOffset:      ^uint16(0) - 2,
			valueLength:      ^uint16(0) - 1,
			expectedKeyOff:   ^uint16(0) - 1,
			expectedKeyLen:   ^uint16(0),
			expectedValueOff: ^uint16(0) - 2,
			expectedValueLen: ^uint16(0) - 1,
		},
		{
			name:             "Adjacent Locations",
			keyOffset:        100,
			keyLength:        20,
			valueOffset:      120, // Right after key
			valueLength:      40,
			expectedKeyOff:   100,
			expectedKeyLen:   20,
			expectedValueOff: 120,
			expectedValueLen: 40,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			kvLocation := createTestKVLocation(tc.keyOffset, tc.keyLength, tc.valueOffset, tc.valueLength)
			kvlr := NewBTreeKVLocationReader(kvLocation)

			// Test KeyLocation
			keyLoc := kvlr.KeyLocation()
			if keyLoc.Off != tc.expectedKeyOff {
				t.Errorf("KeyLocation().Off = %d, want %d", keyLoc.Off, tc.expectedKeyOff)
			}
			if keyLoc.Len != tc.expectedKeyLen {
				t.Errorf("KeyLocation().Len = %d, want %d", keyLoc.Len, tc.expectedKeyLen)
			}

			// Test ValueLocation
			valueLoc := kvlr.ValueLocation()
			if valueLoc.Off != tc.expectedValueOff {
				t.Errorf("ValueLocation().Off = %d, want %d", valueLoc.Off, tc.expectedValueOff)
			}
			if valueLoc.Len != tc.expectedValueLen {
				t.Errorf("ValueLocation().Len = %d, want %d", valueLoc.Len, tc.expectedValueLen)
			}
		})
	}
}

// TestBTreeKVLocationReader_NewConstructor tests the constructor
func TestBTreeKVLocationReader_NewConstructor(t *testing.T) {
	kvLocation := createTestKVLocation(10, 20, 30, 40)
	kvlr := NewBTreeKVLocationReader(kvLocation)

	if kvlr == nil {
		t.Error("NewBTreeKVLocationReader() returned nil")
	}
}

// TestBTreeKVLocationReader_InvalidOffsets tests with invalid offsets
func TestBTreeKVLocationReader_InvalidOffsets(t *testing.T) {
	kvLocation := createTestKVLocation(types.BtoffInvalid, 10, types.BtoffInvalid, 20)
	kvlr := NewBTreeKVLocationReader(kvLocation)

	keyLoc := kvlr.KeyLocation()
	if keyLoc.Off != types.BtoffInvalid {
		t.Errorf("KeyLocation().Off = %d, want %d (BtoffInvalid)", keyLoc.Off, types.BtoffInvalid)
	}

	valueLoc := kvlr.ValueLocation()
	if valueLoc.Off != types.BtoffInvalid {
		t.Errorf("ValueLocation().Off = %d, want %d (BtoffInvalid)", valueLoc.Off, types.BtoffInvalid)
	}
}

// TestBTreeKVLocationReader_StructureIndependence tests that modifications to returned structs don't affect the reader
func TestBTreeKVLocationReader_StructureIndependence(t *testing.T) {
	kvLocation := createTestKVLocation(100, 50, 200, 75)
	kvlr := NewBTreeKVLocationReader(kvLocation)

	// Get locations and modify them
	keyLoc := kvlr.KeyLocation()
	keyLoc.Off = 999
	keyLoc.Len = 888

	valueLoc := kvlr.ValueLocation()
	valueLoc.Off = 777
	valueLoc.Len = 666

	// Original values should be preserved
	newKeyLoc := kvlr.KeyLocation()
	if newKeyLoc.Off != 100 || newKeyLoc.Len != 50 {
		t.Errorf("KeyLocation() modified: got {Off: %d, Len: %d}, want {Off: 100, Len: 50}", newKeyLoc.Off, newKeyLoc.Len)
	}

	newValueLoc := kvlr.ValueLocation()
	if newValueLoc.Off != 200 || newValueLoc.Len != 75 {
		t.Errorf("ValueLocation() modified: got {Off: %d, Len: %d}, want {Off: 200, Len: 75}", newValueLoc.Off, newValueLoc.Len)
	}
}

// Benchmark KV location reader methods
func BenchmarkBTreeKVLocationReader(b *testing.B) {
	kvLocation := createTestKVLocation(100, 50, 200, 75)
	kvlr := NewBTreeKVLocationReader(kvLocation)

	b.Run("KeyLocation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = kvlr.KeyLocation()
		}
	})

	b.Run("ValueLocation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = kvlr.ValueLocation()
		}
	})
}
