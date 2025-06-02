package volumes

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestSuperblockWithEncryptionRollingState creates a test superblock with specific encryption rolling state details
func createTestSuperblockWithEncryptionRollingState(erStateOid types.OidT) *types.ApfsSuperblockT {
	return &types.ApfsSuperblockT{
		ApfsErStateOid: erStateOid,
	}
}

// TestVolumeEncryptionRollingState tests all encryption rolling state method implementations
func TestVolumeEncryptionRollingState(t *testing.T) {
	testCases := []struct {
		name                        string
		erStateOid                  types.OidT
		expectedOid                 types.OidT
		expectedIsRollingInProgress bool
	}{
		{
			name:                        "Valid Encryption Rolling State",
			erStateOid:                  12345,
			expectedOid:                 12345,
			expectedIsRollingInProgress: true,
		},
		{
			name:                        "Invalid OID (OidInvalid)",
			erStateOid:                  types.OidInvalid,
			expectedOid:                 types.OidInvalid,
			expectedIsRollingInProgress: false,
		},
		{
			name:                        "Zero OID",
			erStateOid:                  0,
			expectedOid:                 0,
			expectedIsRollingInProgress: false,
		},
		{
			name:                        "Maximum OID Value",
			erStateOid:                  ^types.OidT(0), // Maximum OidT
			expectedOid:                 ^types.OidT(0),
			expectedIsRollingInProgress: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sb := createTestSuperblockWithEncryptionRollingState(tc.erStateOid)
			vers := NewVolumeEncryptionRollingState(sb)

			// Test EncryptionRollingStateOID
			if oid := vers.EncryptionRollingStateOID(); oid != tc.expectedOid {
				t.Errorf("EncryptionRollingStateOID() = %d, want %d", oid, tc.expectedOid)
			}

			// Test IsEncryptionRollingInProgress
			if isRolling := vers.IsEncryptionRollingInProgress(); isRolling != tc.expectedIsRollingInProgress {
				t.Errorf("IsEncryptionRollingInProgress() = %v, want %v", isRolling, tc.expectedIsRollingInProgress)
			}
		})
	}
}

// TestVolumeEncryptionRollingState_NewConstructor tests the constructor
func TestVolumeEncryptionRollingState_NewConstructor(t *testing.T) {
	sb := createTestSuperblockWithEncryptionRollingState(42)
	vers := NewVolumeEncryptionRollingState(sb)

	if vers == nil {
		t.Error("NewVolumeEncryptionRollingState() returned nil")
	}
}

// TestVolumeEncryptionRollingState_EdgeCases tests edge cases for IsEncryptionRollingInProgress
func TestVolumeEncryptionRollingState_EdgeCases(t *testing.T) {
	edgeCases := []struct {
		name     string
		oid      types.OidT
		expected bool
	}{
		{"OID 1", 1, true},
		{"OID -1 (if signed)", ^types.OidT(0), true},
		{"Small positive OID", 100, true},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			sb := createTestSuperblockWithEncryptionRollingState(tc.oid)
			vers := NewVolumeEncryptionRollingState(sb)

			if got := vers.IsEncryptionRollingInProgress(); got != tc.expected {
				t.Errorf("IsEncryptionRollingInProgress() with OID %d = %v, want %v", tc.oid, got, tc.expected)
			}
		})
	}
}

// Benchmark encryption rolling state methods
func BenchmarkVolumeEncryptionRollingState(b *testing.B) {
	sb := createTestSuperblockWithEncryptionRollingState(12345)
	vers := NewVolumeEncryptionRollingState(sb)

	b.Run("EncryptionRollingStateOID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = vers.EncryptionRollingStateOID()
		}
	})

	b.Run("IsEncryptionRollingInProgress", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = vers.IsEncryptionRollingInProgress()
		}
	})
}
