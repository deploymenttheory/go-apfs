package volumes

import (
	"math"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestSuperblockWithSpaceInfo creates a test superblock with specific space management details
func createTestSuperblockWithSpaceInfo(
	reservedBlocks,
	quotaBlocks,
	allocatedBlocks,
	totalAllocated,
	totalFreed uint64,
) *types.ApfsSuperblockT {
	return &types.ApfsSuperblockT{
		ApfsFsReserveBlockCount: reservedBlocks,
		ApfsFsQuotaBlockCount:   quotaBlocks,
		ApfsFsAllocCount:        allocatedBlocks,
		ApfsTotalBlocksAlloced:  totalAllocated,
		ApfsTotalBlocksFreed:    totalFreed,
	}
}

// TestVolumeSpaceManagement tests all space management method implementations
func TestVolumeSpaceManagement(t *testing.T) {
	testCases := []struct {
		name                    string
		reservedBlocks          uint64
		quotaBlocks             uint64
		allocatedBlocks         uint64
		totalAllocated          uint64
		totalFreed              uint64
		expectedReservedBlocks  uint64
		expectedQuotaBlocks     uint64
		expectedAllocatedBlocks uint64
		expectedTotalAllocated  uint64
		expectedTotalFreed      uint64
		expectedUtilization     float64
	}{
		{
			name:                    "Normal Allocation",
			reservedBlocks:          1000,
			quotaBlocks:             10000,
			allocatedBlocks:         5000,
			totalAllocated:          5000,
			totalFreed:              2000,
			expectedReservedBlocks:  1000,
			expectedQuotaBlocks:     10000,
			expectedAllocatedBlocks: 5000,
			expectedTotalAllocated:  5000,
			expectedTotalFreed:      2000,
			expectedUtilization:     50.0,
		},
		{
			name:                    "Full Allocation",
			reservedBlocks:          500,
			quotaBlocks:             10000,
			allocatedBlocks:         10000,
			totalAllocated:          10000,
			totalFreed:              0,
			expectedReservedBlocks:  500,
			expectedQuotaBlocks:     10000,
			expectedAllocatedBlocks: 10000,
			expectedTotalAllocated:  10000,
			expectedTotalFreed:      0,
			expectedUtilization:     100.0,
		},
		{
			name:                    "Zero Quota",
			reservedBlocks:          0,
			quotaBlocks:             0,
			allocatedBlocks:         100,
			totalAllocated:          100,
			totalFreed:              50,
			expectedReservedBlocks:  0,
			expectedQuotaBlocks:     0,
			expectedAllocatedBlocks: 100,
			expectedTotalAllocated:  100,
			expectedTotalFreed:      50,
			expectedUtilization:     0.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sb := createTestSuperblockWithSpaceInfo(
				tc.reservedBlocks,
				tc.quotaBlocks,
				tc.allocatedBlocks,
				tc.totalAllocated,
				tc.totalFreed,
			)
			vsm := NewVolumeSpaceManagement(sb)

			// Test each method
			assertUint64Equal(t, "ReservedBlockCount", vsm.ReservedBlockCount(), tc.expectedReservedBlocks)
			assertUint64Equal(t, "QuotaBlockCount", vsm.QuotaBlockCount(), tc.expectedQuotaBlocks)
			assertUint64Equal(t, "AllocatedBlockCount", vsm.AllocatedBlockCount(), tc.expectedAllocatedBlocks)
			assertUint64Equal(t, "TotalBlocksAllocated", vsm.TotalBlocksAllocated(), tc.expectedTotalAllocated)
			assertUint64Equal(t, "TotalBlocksFreed", vsm.TotalBlocksFreed(), tc.expectedTotalFreed)

			// Test space utilization with floating-point comparison
			utilization := vsm.SpaceUtilization()
			if !floatEquals(utilization, tc.expectedUtilization) {
				t.Errorf("SpaceUtilization(): expected %f, got %f", tc.expectedUtilization, utilization)
			}
		})
	}
}

// Benchmark space management methods
func BenchmarkVolumeSpaceManagement(b *testing.B) {
	sb := createTestSuperblockWithSpaceInfo(1000, 10000, 5000, 5000, 2000)
	vsm := NewVolumeSpaceManagement(sb)

	// Benchmark individual method calls
	b.Run("ReservedBlockCount", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = vsm.ReservedBlockCount()
		}
	})

	b.Run("SpaceUtilization", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = vsm.SpaceUtilization()
		}
	})
}

// Helper function to assert uint64 equality
func assertUint64Equal(t *testing.T, name string, actual, expected uint64) {
	if actual != expected {
		t.Errorf("%s: expected %d, got %d", name, expected, actual)
	}
}

// Helper function for floating-point comparison with small tolerance
func floatEquals(a, b float64) bool {
	const tolerance = 1e-6
	return math.Abs(a-b) < tolerance
}
