package datastreams

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestPhysicalExtentData creates test physical extent key and value data
func createTestPhysicalExtentData(blockAddress uint64, length uint64, kind uint8, owningObjectID uint64, refCount int32, endian binary.ByteOrder) ([]byte, []byte) {
	// Create key data (8 bytes)
	keyData := make([]byte, 8)
	// The object identifier in the header is the physical block address
	objIdAndType := blockAddress & types.ObjIdMask // No type bits set for simplicity
	endian.PutUint64(keyData[0:8], objIdAndType)

	// Create value data (20 bytes)
	valueData := make([]byte, 20)
	// Pack length and kind into LenAndKind field
	lenAndKind := (length & types.PextLenMask) | ((uint64(kind) << types.PextKindShift) & types.PextKindMask)
	endian.PutUint64(valueData[0:8], lenAndKind)
	endian.PutUint64(valueData[8:16], owningObjectID)
	endian.PutUint32(valueData[16:20], uint32(refCount))

	return keyData, valueData
}

// TestPhysicalExtentReader tests all physical extent reader method implementations
func TestPhysicalExtentReader(t *testing.T) {
	testCases := []struct {
		name              string
		blockAddress      uint64
		length            uint64
		kind              uint8
		owningObjectID    uint64
		refCount          int32
		expectedValid     bool
		expectedShared    bool
		expectedDeletable bool
	}{
		{
			name:              "Valid New Extent",
			blockAddress:      1000,
			length:            100,
			kind:              uint8(types.ApfsKindNew),
			owningObjectID:    500,
			refCount:          1,
			expectedValid:     true,
			expectedShared:    false,
			expectedDeletable: false,
		},
		{
			name:              "Valid Update Extent",
			blockAddress:      2000,
			length:            50,
			kind:              uint8(types.ApfsKindUpdate),
			owningObjectID:    600,
			refCount:          1,
			expectedValid:     true,
			expectedShared:    false,
			expectedDeletable: false,
		},
		{
			name:              "Shared Extent",
			blockAddress:      3000,
			length:            200,
			kind:              uint8(types.ApfsKindNew),
			owningObjectID:    700,
			refCount:          3,
			expectedValid:     true,
			expectedShared:    true,
			expectedDeletable: false,
		},
		{
			name:              "Zero Reference Count (Deletable)",
			blockAddress:      4000,
			length:            25,
			kind:              uint8(types.ApfsKindNew),
			owningObjectID:    800,
			refCount:          0,
			expectedValid:     true,
			expectedShared:    false,
			expectedDeletable: true,
		},
		{
			name:              "Dead Kind (Invalid on disk)",
			blockAddress:      5000,
			length:            75,
			kind:              uint8(types.ApfsKindDead),
			owningObjectID:    900,
			refCount:          1,
			expectedValid:     false,
			expectedShared:    false,
			expectedDeletable: false,
		},
		{
			name:              "Large Extent",
			blockAddress:      0x123456789ABC,
			length:            0xFFFFFFFFFFFFF, // Maximum length (60 bits)
			kind:              uint8(types.ApfsKindNew),
			owningObjectID:    0xDEADBEEFCAFE,
			refCount:          1,
			expectedValid:     true,
			expectedShared:    false,
			expectedDeletable: false,
		},
	}

	endian := binary.LittleEndian

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keyData, valueData := createTestPhysicalExtentData(tc.blockAddress, tc.length, tc.kind, tc.owningObjectID, tc.refCount, endian)

			per, err := NewPhysicalExtentReader(keyData, valueData, endian)
			if err != nil {
				t.Fatalf("NewPhysicalExtentReader() failed: %v", err)
			}

			// Test basic properties
			if addr := per.PhysicalBlockAddress(); addr != tc.blockAddress {
				t.Errorf("PhysicalBlockAddress() = %d, want %d", addr, tc.blockAddress)
			}

			if length := per.Length(); length != tc.length {
				t.Errorf("Length() = %d, want %d", length, tc.length)
			}

			if kind := per.Kind(); kind != tc.kind {
				t.Errorf("Kind() = %d, want %d", kind, tc.kind)
			}

			if ownerID := per.OwningObjectID(); ownerID != tc.owningObjectID {
				t.Errorf("OwningObjectID() = %d, want %d", ownerID, tc.owningObjectID)
			}

			if localRefCount := per.ReferenceCount(); localRefCount != tc.refCount {
				t.Errorf("ReferenceCount() = %d, want %d", localRefCount, tc.refCount)
			}

			// Test kind checking methods
			switch tc.kind {
			case uint8(types.ApfsKindNew):
				if !per.IsKindNew() {
					t.Error("IsKindNew() should be true for APFS_KIND_NEW")
				}
				if per.IsKindUpdate() || per.IsKindDead() {
					t.Error("Only IsKindNew() should be true")
				}
			case uint8(types.ApfsKindUpdate):
				if !per.IsKindUpdate() {
					t.Error("IsKindUpdate() should be true for APFS_KIND_UPDATE")
				}
				if per.IsKindNew() || per.IsKindDead() {
					t.Error("Only IsKindUpdate() should be true")
				}
			case uint8(types.ApfsKindDead):
				if !per.IsKindDead() {
					t.Error("IsKindDead() should be true for APFS_KIND_DEAD")
				}
				if per.IsKindNew() || per.IsKindUpdate() {
					t.Error("Only IsKindDead() should be true")
				}
			}

			// Test validation
			if per.IsValidKind() != tc.expectedValid {
				t.Errorf("IsValidKind() = %v, want %v", per.IsValidKind(), tc.expectedValid)
			}

			// Test sharing and deletion status
			if per.IsShared() != tc.expectedShared {
				t.Errorf("IsShared() = %v, want %v", per.IsShared(), tc.expectedShared)
			}

			if per.CanBeDeleted() != tc.expectedDeletable {
				t.Errorf("CanBeDeleted() = %v, want %v", per.CanBeDeleted(), tc.expectedDeletable)
			}

			// Test size calculations
			blockSize := uint32(4096)
			expectedSizeBytes := tc.length * uint64(blockSize)
			if sizeBytes := per.SizeInBytes(blockSize); sizeBytes != expectedSizeBytes {
				t.Errorf("SizeInBytes(%d) = %d, want %d", blockSize, sizeBytes, expectedSizeBytes)
			}

			// Test end address calculation
			expectedEndAddr := tc.blockAddress + tc.length
			if endAddr := per.EndBlockAddress(); endAddr != expectedEndAddr {
				t.Errorf("EndBlockAddress() = %d, want %d", endAddr, expectedEndAddr)
			}

			// Test block containment
			midBlock := tc.blockAddress + tc.length/2
			if !per.ContainsBlock(midBlock) {
				t.Errorf("ContainsBlock(%d) should be true for block in middle of extent", midBlock)
			}

			if per.ContainsBlock(tc.blockAddress + tc.length) {
				t.Errorf("ContainsBlock(%d) should be false for block at end (exclusive)", tc.blockAddress+tc.length)
			}

			// Test kind description
			description := per.GetKindDescription()
			if description == "" {
				t.Error("GetKindDescription() should not return empty string")
			}
		})
	}
}

// TestPhysicalExtentReader_ErrorCases tests error conditions
func TestPhysicalExtentReader_ErrorCases(t *testing.T) {
	endian := binary.LittleEndian

	testCases := []struct {
		name        string
		keyDataSize int
		valDataSize int
	}{
		{"Empty key data", 0, 20},
		{"Too small key data", 7, 20},
		{"Empty value data", 8, 0},
		{"Too small value data", 8, 19},
		{"Both too small", 4, 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keyData := make([]byte, tc.keyDataSize)
			valueData := make([]byte, tc.valDataSize)

			_, err := NewPhysicalExtentReader(keyData, valueData, endian)
			if err == nil {
				t.Error("NewPhysicalExtentReader() should have failed with insufficient data")
			}
		})
	}
}

// TestPhysicalExtentReader_Validation tests the Validate method
func TestPhysicalExtentReader_Validation(t *testing.T) {
	endian := binary.LittleEndian

	testCases := []struct {
		name         string
		blockAddress uint64
		length       uint64
		kind         uint8
		refCount     int32
		shouldError  bool
	}{
		{
			name:         "Valid extent",
			blockAddress: 1000,
			length:       100,
			kind:         uint8(types.ApfsKindNew),
			refCount:     1,
			shouldError:  false,
		},
		{
			name:         "Zero length",
			blockAddress: 1000,
			length:       0,
			kind:         uint8(types.ApfsKindNew),
			refCount:     1,
			shouldError:  true,
		},
		{
			name:         "Invalid kind (Dead)",
			blockAddress: 1000,
			length:       100,
			kind:         uint8(types.ApfsKindDead),
			refCount:     1,
			shouldError:  true,
		},
		{
			name:         "Negative reference count",
			blockAddress: 1000,
			length:       100,
			kind:         uint8(types.ApfsKindNew),
			refCount:     -1,
			shouldError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keyData, valueData := createTestPhysicalExtentData(tc.blockAddress, tc.length, tc.kind, 1000, tc.refCount, endian)

			per, err := NewPhysicalExtentReader(keyData, valueData, endian)
			if err != nil {
				t.Fatalf("NewPhysicalExtentReader() failed: %v", err)
			}

			err = per.Validate()
			if tc.shouldError && err == nil {
				t.Error("Validate() should have failed but didn't")
			} else if !tc.shouldError && err != nil {
				t.Errorf("Validate() failed unexpectedly: %v", err)
			}
		})
	}
}

// TestPhysicalExtentReader_EndianHandling tests both little and big endian
func TestPhysicalExtentReader_EndianHandling(t *testing.T) {
	testCases := []struct {
		name   string
		endian binary.ByteOrder
	}{
		{"Little Endian", binary.LittleEndian},
		{"Big Endian", binary.BigEndian},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			blockAddress := uint64(0x023456789ABCDEF0) // Fits within ObjIdMask (60 bits)
			extentLength := uint64(0x87654321)
			extentKind := uint8(types.ApfsKindNew)
			owningObjectID := uint64(0xFEDCBA0987654321)
			refCount := int32(42)

			keyData, valueData := createTestPhysicalExtentData(blockAddress, extentLength, extentKind, owningObjectID, refCount, tc.endian)

			per, err := NewPhysicalExtentReader(keyData, valueData, tc.endian)
			if err != nil {
				t.Fatalf("NewPhysicalExtentReader() failed: %v", err)
			}

			// Values should be parsed correctly regardless of endianness
			if addr := per.PhysicalBlockAddress(); addr != blockAddress {
				t.Errorf("PhysicalBlockAddress() = 0x%016X, want 0x%016X", addr, blockAddress)
			}

			if length := per.Length(); length != extentLength {
				t.Errorf("Length() = %d, want %d", length, extentLength)
			}

			if kind := per.Kind(); kind != extentKind {
				t.Errorf("Kind() = %d, want %d", kind, extentKind)
			}

			if ownerID := per.OwningObjectID(); ownerID != owningObjectID {
				t.Errorf("OwningObjectID() = 0x%016X, want 0x%016X", ownerID, owningObjectID)
			}

			if localRefCount := per.ReferenceCount(); localRefCount != refCount {
				t.Errorf("ReferenceCount() = %d, want %d", localRefCount, refCount)
			}
		})
	}
}

// TestPhysicalExtentReader_BitMasking tests bit masking operations
func TestPhysicalExtentReader_BitMasking(t *testing.T) {
	endian := binary.LittleEndian

	testCases := []struct {
		name              string
		blockAddress      uint64
		length            uint64
		kind              uint8
		expectedObjIdMask uint64
		expectedLenMask   uint64
		expectedKindMask  uint8
	}{
		{
			name:              "Maximum values",
			blockAddress:      types.ObjIdMask,   // All 60 bits set
			length:            types.PextLenMask, // All 60 bits set
			kind:              0xF,               // All 4 bits set
			expectedObjIdMask: types.ObjIdMask,
			expectedLenMask:   types.PextLenMask,
			expectedKindMask:  0xF,
		},
		{
			name:              "Minimum values",
			blockAddress:      1,
			length:            1,
			kind:              0,
			expectedObjIdMask: 1,
			expectedLenMask:   1,
			expectedKindMask:  0,
		},
		{
			name:              "Mixed values",
			blockAddress:      0x123456789ABCDEF,
			length:            0x456789ABCDEF123,
			kind:              5,
			expectedObjIdMask: 0x123456789ABCDEF,
			expectedLenMask:   0x456789ABCDEF123,
			expectedKindMask:  5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keyData, valueData := createTestPhysicalExtentData(tc.blockAddress, tc.length, tc.kind, 1000, 1, endian)

			per, err := NewPhysicalExtentReader(keyData, valueData, endian)
			if err != nil {
				t.Fatalf("NewPhysicalExtentReader() failed: %v", err)
			}

			// Test that bit masking works correctly
			if addr := per.PhysicalBlockAddress(); addr != tc.expectedObjIdMask {
				t.Errorf("PhysicalBlockAddress() = 0x%016X, want 0x%016X", addr, tc.expectedObjIdMask)
			}

			if length := per.Length(); length != tc.expectedLenMask {
				t.Errorf("Length() = 0x%016X, want 0x%016X", length, tc.expectedLenMask)
			}

			if kind := per.Kind(); kind != tc.expectedKindMask {
				t.Errorf("Kind() = 0x%02X, want 0x%02X", kind, tc.expectedKindMask)
			}
		})
	}
}

// Benchmark physical extent reader methods
func BenchmarkPhysicalExtentReader(b *testing.B) {
	endian := binary.LittleEndian
	keyData, valueData := createTestPhysicalExtentData(1000, 100, uint8(types.ApfsKindNew), 500, 1, endian)
	per, _ := NewPhysicalExtentReader(keyData, valueData, endian)

	b.Run("PhysicalBlockAddress", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = per.PhysicalBlockAddress()
		}
	})

	b.Run("Length", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = per.Length()
		}
	})

	b.Run("Kind", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = per.Kind()
		}
	})

	b.Run("IsValidKind", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = per.IsValidKind()
		}
	})

	b.Run("ContainsBlock", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = per.ContainsBlock(1050)
		}
	})
}
