package btrees

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestKVOffset creates a test KvoffT structure
func createTestKVOffset(keyOffset, valueOffset uint16) types.KvoffT {
	return types.KvoffT{
		K: keyOffset,
		V: valueOffset,
	}
}

// TestBTreeKVOffsetReader tests all KV offset reader method implementations
func TestBTreeKVOffsetReader(t *testing.T) {
	testCases := []struct {
		name                string
		keyOffset           uint16
		valueOffset         uint16
		expectedKeyOffset   uint16
		expectedValueOffset uint16
	}{
		{
			name:                "Valid KV Offset",
			keyOffset:           100,
			valueOffset:         200,
			expectedKeyOffset:   100,
			expectedValueOffset: 200,
		},
		{
			name:                "Zero Offsets",
			keyOffset:           0,
			valueOffset:         0,
			expectedKeyOffset:   0,
			expectedValueOffset: 0,
		},
		{
			name:                "Adjacent Offsets",
			keyOffset:           100,
			valueOffset:         108, // Assuming 8-byte key
			expectedKeyOffset:   100,
			expectedValueOffset: 108,
		},
		{
			name:                "Maximum Values",
			keyOffset:           ^uint16(0),     // Maximum uint16
			valueOffset:         ^uint16(0) - 1, // Maximum uint16 - 1
			expectedKeyOffset:   ^uint16(0),
			expectedValueOffset: ^uint16(0) - 1,
		},
		{
			name:                "Reverse Order",
			keyOffset:           500,
			valueOffset:         100, // Value before key in storage
			expectedKeyOffset:   500,
			expectedValueOffset: 100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			kvOffset := createTestKVOffset(tc.keyOffset, tc.valueOffset)
			kvor := NewBTreeKVOffsetReader(kvOffset)

			// Test KeyOffset
			if keyOffset := kvor.KeyOffset(); keyOffset != tc.expectedKeyOffset {
				t.Errorf("KeyOffset() = %d, want %d", keyOffset, tc.expectedKeyOffset)
			}

			// Test ValueOffset
			if valueOffset := kvor.ValueOffset(); valueOffset != tc.expectedValueOffset {
				t.Errorf("ValueOffset() = %d, want %d", valueOffset, tc.expectedValueOffset)
			}
		})
	}
}

// TestBTreeKVOffsetReader_NewConstructor tests the constructor
func TestBTreeKVOffsetReader_NewConstructor(t *testing.T) {
	kvOffset := createTestKVOffset(42, 84)
	kvor := NewBTreeKVOffsetReader(kvOffset)

	if kvor == nil {
		t.Error("NewBTreeKVOffsetReader() returned nil")
	}
}

// TestBTreeKVOffsetReader_Consistency tests that multiple calls return the same values
func TestBTreeKVOffsetReader_Consistency(t *testing.T) {
	kvOffset := createTestKVOffset(123, 456)
	kvor := NewBTreeKVOffsetReader(kvOffset)

	// Call methods multiple times and ensure consistency
	for i := 0; i < 5; i++ {
		if keyOffset := kvor.KeyOffset(); keyOffset != 123 {
			t.Errorf("KeyOffset() call %d = %d, want 123", i+1, keyOffset)
		}

		if valueOffset := kvor.ValueOffset(); valueOffset != 456 {
			t.Errorf("ValueOffset() call %d = %d, want 456", i+1, valueOffset)
		}
	}
}

// TestBTreeKVOffsetReader_BoundaryValues tests boundary values
func TestBTreeKVOffsetReader_BoundaryValues(t *testing.T) {
	testCases := []struct {
		name        string
		keyOffset   uint16
		valueOffset uint16
	}{
		{"Zero values", 0, 0},
		{"One value", 1, 1},
		{"Max minus one", ^uint16(0) - 1, ^uint16(0) - 1},
		{"Max values", ^uint16(0), ^uint16(0)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			kvOffset := createTestKVOffset(tc.keyOffset, tc.valueOffset)
			kvor := NewBTreeKVOffsetReader(kvOffset)

			if keyOffset := kvor.KeyOffset(); keyOffset != tc.keyOffset {
				t.Errorf("KeyOffset() = %d, want %d", keyOffset, tc.keyOffset)
			}

			if valueOffset := kvor.ValueOffset(); valueOffset != tc.valueOffset {
				t.Errorf("ValueOffset() = %d, want %d", valueOffset, tc.valueOffset)
			}
		})
	}
}

// TestBTreeKVOffsetReader_FixedSizeUseCase tests typical fixed-size key-value scenarios
func TestBTreeKVOffsetReader_FixedSizeUseCase(t *testing.T) {
	// Simulate fixed-size key-value pairs (8-byte keys, 16-byte values)
	testCases := []struct {
		name        string
		keyOffset   uint16
		valueOffset uint16
		description string
	}{
		{
			name:        "First entry",
			keyOffset:   0,
			valueOffset: 8,
			description: "Key at start, value right after",
		},
		{
			name:        "Second entry",
			keyOffset:   24, // After first key+value (8+16)
			valueOffset: 32, // 24 + 8
			description: "Second key-value pair",
		},
		{
			name:        "Third entry",
			keyOffset:   48, // After two key+value pairs (24*2)
			valueOffset: 56, // 48 + 8
			description: "Third key-value pair",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			kvOffset := createTestKVOffset(tc.keyOffset, tc.valueOffset)
			kvor := NewBTreeKVOffsetReader(kvOffset)

			if keyOffset := kvor.KeyOffset(); keyOffset != tc.keyOffset {
				t.Errorf("KeyOffset() = %d, want %d (%s)", keyOffset, tc.keyOffset, tc.description)
			}

			if valueOffset := kvor.ValueOffset(); valueOffset != tc.valueOffset {
				t.Errorf("ValueOffset() = %d, want %d (%s)", valueOffset, tc.valueOffset, tc.description)
			}
		})
	}
}

// Benchmark KV offset reader methods
func BenchmarkBTreeKVOffsetReader(b *testing.B) {
	kvOffset := createTestKVOffset(100, 200)
	kvor := NewBTreeKVOffsetReader(kvOffset)

	b.Run("KeyOffset", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = kvor.KeyOffset()
		}
	})

	b.Run("ValueOffset", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = kvor.ValueOffset()
		}
	})
}
