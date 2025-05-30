package volumes

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestSuperblockWithMagic creates a test superblock with a specific magic number
func createTestSuperblockWithMagic(magic uint32) *types.ApfsSuperblockT {
	return &types.ApfsSuperblockT{
		ApfsMagic: magic,
	}
}

// TestVolumeIntegrityCheck tests all integrity check method implementations
func TestVolumeIntegrityCheck(t *testing.T) {
	testCases := []struct {
		name            string
		magic           uint32
		expectedMagic   uint32
		expectedIsValid bool
	}{
		{
			name:            "Valid Magic Number",
			magic:           types.ApfsMagic,
			expectedMagic:   types.ApfsMagic,
			expectedIsValid: true,
		},
		{
			name:            "Invalid Magic Number",
			magic:           0x12345678,
			expectedMagic:   0x12345678,
			expectedIsValid: false,
		},
		{
			name:            "Zero Magic Number",
			magic:           0,
			expectedMagic:   0,
			expectedIsValid: false,
		},
		{
			name:            "Maximum Magic Number",
			magic:           ^uint32(0), // Maximum uint32
			expectedMagic:   ^uint32(0),
			expectedIsValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sb := createTestSuperblockWithMagic(tc.magic)
			vic := NewVolumeIntegrityCheck(sb)

			// Test MagicNumber
			if magic := vic.MagicNumber(); magic != tc.expectedMagic {
				t.Errorf("MagicNumber() = 0x%08X, want 0x%08X", magic, tc.expectedMagic)
			}

			// Test ValidateMagicNumber
			if isValid := vic.ValidateMagicNumber(); isValid != tc.expectedIsValid {
				t.Errorf("ValidateMagicNumber() = %v, want %v", isValid, tc.expectedIsValid)
			}
		})
	}
}

// TestVolumeIntegrityCheck_NewConstructor tests the constructor
func TestVolumeIntegrityCheck_NewConstructor(t *testing.T) {
	sb := createTestSuperblockWithMagic(types.ApfsMagic)
	vic := NewVolumeIntegrityCheck(sb)

	if vic == nil {
		t.Error("NewVolumeIntegrityCheck() returned nil")
	}
}

// TestVolumeIntegrityCheck_ValidMagicConstant tests against the actual APFS magic constant
func TestVolumeIntegrityCheck_ValidMagicConstant(t *testing.T) {
	sb := createTestSuperblockWithMagic(types.ApfsMagic)
	vic := NewVolumeIntegrityCheck(sb)

	if !vic.ValidateMagicNumber() {
		t.Error("ValidateMagicNumber() should return true for types.ApfsMagic")
	}

	if magic := vic.MagicNumber(); magic != types.ApfsMagic {
		t.Errorf("MagicNumber() = 0x%08X, want 0x%08X (types.ApfsMagic)", magic, types.ApfsMagic)
	}
}

// TestVolumeIntegrityCheck_IndividualMethods tests each method independently
func TestVolumeIntegrityCheck_IndividualMethods(t *testing.T) {
	testMagic := uint32(0xABCDEF12)
	sb := createTestSuperblockWithMagic(testMagic)
	vic := NewVolumeIntegrityCheck(sb)

	t.Run("MagicNumber", func(t *testing.T) {
		if got := vic.MagicNumber(); got != testMagic {
			t.Errorf("MagicNumber() = 0x%08X, want 0x%08X", got, testMagic)
		}
	})

	t.Run("ValidateMagicNumber", func(t *testing.T) {
		// Should be false since testMagic is not types.ApfsMagic
		if got := vic.ValidateMagicNumber(); got != false {
			t.Errorf("ValidateMagicNumber() = %v, want %v", got, false)
		}
	})
}

// TestVolumeIntegrityCheck_EdgeCases tests edge cases for magic number validation
func TestVolumeIntegrityCheck_EdgeCases(t *testing.T) {
	edgeCases := []struct {
		name        string
		magic       uint32
		shouldMatch bool
	}{
		{"Magic - 1", types.ApfsMagic - 1, false},
		{"Magic + 1", types.ApfsMagic + 1, false},
		{"Byte swapped magic (little endian)", 0x53504641, false}, // "APFS" but byte swapped
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			sb := createTestSuperblockWithMagic(tc.magic)
			vic := NewVolumeIntegrityCheck(sb)

			if got := vic.ValidateMagicNumber(); got != tc.shouldMatch {
				t.Errorf("ValidateMagicNumber() with magic 0x%08X = %v, want %v", tc.magic, got, tc.shouldMatch)
			}
		})
	}
}

// Benchmark integrity check methods
func BenchmarkVolumeIntegrityCheck(b *testing.B) {
	sb := createTestSuperblockWithMagic(types.ApfsMagic)
	vic := NewVolumeIntegrityCheck(sb)

	b.Run("MagicNumber", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = vic.MagicNumber()
		}
	})

	b.Run("ValidateMagicNumber", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = vic.ValidateMagicNumber()
		}
	})
}
