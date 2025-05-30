package btrees

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestIndexNodeValue creates a test BtnIndexNodeValT structure
func createTestIndexNodeValue(childOID types.OidT, childHash [types.BtreeNodeHashSizeMax]byte) *types.BtnIndexNodeValT {
	return &types.BtnIndexNodeValT{
		BinvChildOid:  childOID,
		BinvChildHash: childHash,
	}
}

// createTestHash creates a test hash array
func createTestHash(pattern byte) [types.BtreeNodeHashSizeMax]byte {
	var hash [types.BtreeNodeHashSizeMax]byte
	for i := range hash {
		hash[i] = pattern + byte(i)
	}
	return hash
}

// TestBTreeIndexNodeValueReader tests all index node value reader method implementations
func TestBTreeIndexNodeValueReader(t *testing.T) {
	testCases := []struct {
		name             string
		childOID         types.OidT
		childHashPattern byte
		expectedOID      types.OidT
	}{
		{
			name:             "Valid Index Node Value",
			childOID:         12345,
			childHashPattern: 0x01,
			expectedOID:      12345,
		},
		{
			name:             "Zero OID",
			childOID:         0,
			childHashPattern: 0x00,
			expectedOID:      0,
		},
		{
			name:             "Maximum OID",
			childOID:         types.OidT(^uint64(0)),
			childHashPattern: 0xFF,
			expectedOID:      types.OidT(^uint64(0)),
		},
		{
			name:             "Random Pattern Hash",
			childOID:         987654321,
			childHashPattern: 0xAB,
			expectedOID:      987654321,
		},
		{
			name:             "Sequential Hash Pattern",
			childOID:         555666777,
			childHashPattern: 0x10,
			expectedOID:      555666777,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			childHash := createTestHash(tc.childHashPattern)
			indexValue := createTestIndexNodeValue(tc.childOID, childHash)
			inr := NewBTreeIndexNodeValueReader(indexValue)

			// Test ChildObjectID
			if oid := inr.ChildObjectID(); oid != tc.expectedOID {
				t.Errorf("ChildObjectID() = %d, want %d", oid, tc.expectedOID)
			}

			// Test ChildHash
			hash := inr.ChildHash()
			for i := 0; i < types.BtreeNodeHashSizeMax; i++ {
				expectedByte := tc.childHashPattern + byte(i)
				if hash[i] != expectedByte {
					t.Errorf("ChildHash()[%d] = 0x%02X, want 0x%02X", i, hash[i], expectedByte)
				}
			}
		})
	}
}

// TestBTreeIndexNodeValueReader_NewConstructor tests the constructor
func TestBTreeIndexNodeValueReader_NewConstructor(t *testing.T) {
	hash := createTestHash(0x42)
	indexValue := createTestIndexNodeValue(12345, hash)
	inr := NewBTreeIndexNodeValueReader(indexValue)

	if inr == nil {
		t.Error("NewBTreeIndexNodeValueReader() returned nil")
	}
}

// TestBTreeIndexNodeValueReader_HashIntegrity tests that hash values are preserved
func TestBTreeIndexNodeValueReader_HashIntegrity(t *testing.T) {
	// Create a hash with specific pattern
	var originalHash [types.BtreeNodeHashSizeMax]byte
	for i := range originalHash {
		originalHash[i] = byte(i % 256)
	}

	indexValue := createTestIndexNodeValue(123456, originalHash)
	inr := NewBTreeIndexNodeValueReader(indexValue)

	// Get hash multiple times and verify consistency
	for call := 0; call < 3; call++ {
		hash := inr.ChildHash()
		for i := 0; i < types.BtreeNodeHashSizeMax; i++ {
			expected := byte(i % 256)
			if hash[i] != expected {
				t.Errorf("Call %d: ChildHash()[%d] = 0x%02X, want 0x%02X", call+1, i, hash[i], expected)
			}
		}
	}
}

// TestBTreeIndexNodeValueReader_HashIndependence tests that returned hash can be modified without affecting the reader
func TestBTreeIndexNodeValueReader_HashIndependence(t *testing.T) {
	originalHash := createTestHash(0x55)
	indexValue := createTestIndexNodeValue(789, originalHash)
	inr := NewBTreeIndexNodeValueReader(indexValue)

	// Get hash and modify it
	modifiedHash := inr.ChildHash()
	for i := range modifiedHash {
		modifiedHash[i] = 0xFF
	}

	// Original should be preserved
	preservedHash := inr.ChildHash()
	for i := 0; i < types.BtreeNodeHashSizeMax; i++ {
		expected := 0x55 + byte(i)
		if preservedHash[i] != expected {
			t.Errorf("Hash was modified: ChildHash()[%d] = 0x%02X, want 0x%02X", i, preservedHash[i], expected)
		}
	}
}

// TestBTreeIndexNodeValueReader_ZeroHash tests handling of zero hash
func TestBTreeIndexNodeValueReader_ZeroHash(t *testing.T) {
	var zeroHash [types.BtreeNodeHashSizeMax]byte
	indexValue := createTestIndexNodeValue(42, zeroHash)
	inr := NewBTreeIndexNodeValueReader(indexValue)

	hash := inr.ChildHash()
	for i := 0; i < types.BtreeNodeHashSizeMax; i++ {
		if hash[i] != 0 {
			t.Errorf("ZeroHash[%d] = 0x%02X, want 0x00", i, hash[i])
		}
	}
}

// TestBTreeIndexNodeValueReader_ConsistentCalls tests that multiple calls return consistent values
func TestBTreeIndexNodeValueReader_ConsistentCalls(t *testing.T) {
	hash := createTestHash(0x77)
	indexValue := createTestIndexNodeValue(999888777, hash)
	inr := NewBTreeIndexNodeValueReader(indexValue)

	// Test OID consistency
	for i := 0; i < 5; i++ {
		if oid := inr.ChildObjectID(); oid != 999888777 {
			t.Errorf("Call %d: ChildObjectID() = %d, want 999888777", i+1, oid)
		}
	}

	// Test hash consistency
	firstHash := inr.ChildHash()
	for i := 0; i < 3; i++ {
		currentHash := inr.ChildHash()
		for j := 0; j < types.BtreeNodeHashSizeMax; j++ {
			if currentHash[j] != firstHash[j] {
				t.Errorf("Call %d: Hash inconsistency at byte %d: got 0x%02X, want 0x%02X", i+2, j, currentHash[j], firstHash[j])
			}
		}
	}
}

// TestBTreeIndexNodeValueReader_EdgeCases tests edge cases for OID values
func TestBTreeIndexNodeValueReader_EdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		childOID types.OidT
	}{
		{"Minimum OID", 0},
		{"Small OID", 1},
		{"Medium OID", 0x1234567890ABCDEF},
		{"Large OID", 0xFEDCBA0987654321},
		{"Maximum OID", types.OidT(^uint64(0))},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash := createTestHash(0x33)
			indexValue := createTestIndexNodeValue(tc.childOID, hash)
			inr := NewBTreeIndexNodeValueReader(indexValue)

			if oid := inr.ChildObjectID(); oid != tc.childOID {
				t.Errorf("ChildObjectID() = %d, want %d", oid, tc.childOID)
			}
		})
	}
}

// Benchmark index node value reader methods
func BenchmarkBTreeIndexNodeValueReader(b *testing.B) {
	hash := createTestHash(0x88)
	indexValue := createTestIndexNodeValue(123456789, hash)
	inr := NewBTreeIndexNodeValueReader(indexValue)

	b.Run("ChildObjectID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = inr.ChildObjectID()
		}
	})

	b.Run("ChildHash", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = inr.ChildHash()
		}
	})
}
