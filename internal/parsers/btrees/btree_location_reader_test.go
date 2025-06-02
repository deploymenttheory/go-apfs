package btrees

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestLocation creates a test NlocT structure
func createTestLocation(offset, length uint16) types.NlocT {
	return types.NlocT{
		Off: offset,
		Len: length,
	}
}

// TestBTreeLocationReader tests all location reader method implementations
func TestBTreeLocationReader(t *testing.T) {
	testCases := []struct {
		name           string
		offset         uint16
		length         uint16
		expectedOffset uint16
		expectedLength uint16
		expectedValid  bool
	}{
		{
			name:           "Valid Location",
			offset:         100,
			length:         50,
			expectedOffset: 100,
			expectedLength: 50,
			expectedValid:  true,
		},
		{
			name:           "Zero Offset Location",
			offset:         0,
			length:         25,
			expectedOffset: 0,
			expectedLength: 25,
			expectedValid:  true,
		},
		{
			name:           "Invalid Offset Location",
			offset:         types.BtoffInvalid,
			length:         10,
			expectedOffset: types.BtoffInvalid,
			expectedLength: 10,
			expectedValid:  false,
		},
		{
			name:           "Zero Length Location",
			offset:         200,
			length:         0,
			expectedOffset: 200,
			expectedLength: 0,
			expectedValid:  true,
		},
		{
			name:           "Maximum Values",
			offset:         ^uint16(0) - 1, // Max valid offset
			length:         ^uint16(0),     // Max length
			expectedOffset: ^uint16(0) - 1,
			expectedLength: ^uint16(0),
			expectedValid:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			location := createTestLocation(tc.offset, tc.length)
			lr := NewBTreeLocationReader(location)

			// Test Offset
			if offset := lr.Offset(); offset != tc.expectedOffset {
				t.Errorf("Offset() = %d, want %d", offset, tc.expectedOffset)
			}

			// Test Length
			if length := lr.Length(); length != tc.expectedLength {
				t.Errorf("Length() = %d, want %d", length, tc.expectedLength)
			}

			// Test IsValid
			if valid := lr.IsValid(); valid != tc.expectedValid {
				t.Errorf("IsValid() = %v, want %v", valid, tc.expectedValid)
			}
		})
	}
}

// TestBTreeLocationReader_NewConstructor tests the constructor
func TestBTreeLocationReader_NewConstructor(t *testing.T) {
	location := createTestLocation(42, 84)
	lr := NewBTreeLocationReader(location)

	if lr == nil {
		t.Error("NewBTreeLocationReader() returned nil")
	}
}

// TestBTreeLocationReader_InvalidConstants tests the invalid offset constant
func TestBTreeLocationReader_InvalidConstants(t *testing.T) {
	location := createTestLocation(types.BtoffInvalid, 100)
	lr := NewBTreeLocationReader(location)

	if lr.IsValid() {
		t.Error("IsValid() should return false for BtoffInvalid offset")
	}

	if offset := lr.Offset(); offset != types.BtoffInvalid {
		t.Errorf("Offset() = %d, want %d (BtoffInvalid)", offset, types.BtoffInvalid)
	}
}

// Benchmark location reader methods
func BenchmarkBTreeLocationReader(b *testing.B) {
	location := createTestLocation(100, 50)
	lr := NewBTreeLocationReader(location)

	b.Run("Offset", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = lr.Offset()
		}
	})

	b.Run("Length", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = lr.Length()
		}
	})

	b.Run("IsValid", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = lr.IsValid()
		}
	})
}
