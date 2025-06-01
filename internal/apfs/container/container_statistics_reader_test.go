package container

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestSuperblockForStatistics creates a test superblock with specific counter values
func createTestSuperblockForStatistics(counters [types.NxNumCounters]uint64) *types.NxSuperblockT {
	return &types.NxSuperblockT{
		NxCounters: counters,
	}
}

func TestContainerStatisticsReader(t *testing.T) {
	// Create test counters array
	var testCounters [types.NxNumCounters]uint64
	for i := 0; i < types.NxNumCounters; i++ {
		testCounters[i] = uint64(i * 100) // Test values: 0, 100, 200, ...
	}

	superblock := createTestSuperblockForStatistics(testCounters)
	reader := NewContainerStatisticsReader(superblock)

	// Test Counters() method
	t.Run("Counters", func(t *testing.T) {
		counters := reader.Counters()

		if len(counters) != types.NxNumCounters {
			t.Errorf("Counters() length = %d, want %d", len(counters), types.NxNumCounters)
		}

		for i, counter := range counters {
			expected := uint64(i * 100)
			if counter != expected {
				t.Errorf("Counters()[%d] = %d, want %d", i, counter, expected)
			}
		}
	})

	// Test ObjectChecksumSetCount() method
	t.Run("ObjectChecksumSetCount", func(t *testing.T) {
		count := reader.ObjectChecksumSetCount()
		expected := testCounters[types.NxCntrObjCksumSet]

		if count != expected {
			t.Errorf("ObjectChecksumSetCount() = %d, want %d", count, expected)
		}
	})

	// Test ObjectChecksumFailCount() method
	t.Run("ObjectChecksumFailCount", func(t *testing.T) {
		count := reader.ObjectChecksumFailCount()
		expected := testCounters[types.NxCntrObjCksumFail]

		if count != expected {
			t.Errorf("ObjectChecksumFailCount() = %d, want %d", count, expected)
		}
	})
}

func TestContainerStatisticsReader_EmptyCounters(t *testing.T) {
	var emptyCounters [types.NxNumCounters]uint64
	superblock := createTestSuperblockForStatistics(emptyCounters)
	reader := NewContainerStatisticsReader(superblock)

	// Test with all zero counters
	counters := reader.Counters()
	for i, counter := range counters {
		if counter != 0 {
			t.Errorf("Expected zero counter at index %d, got %d", i, counter)
		}
	}

	if reader.ObjectChecksumSetCount() != 0 {
		t.Errorf("ObjectChecksumSetCount() = %d, want 0", reader.ObjectChecksumSetCount())
	}

	if reader.ObjectChecksumFailCount() != 0 {
		t.Errorf("ObjectChecksumFailCount() = %d, want 0", reader.ObjectChecksumFailCount())
	}
}

func TestContainerStatisticsReader_SpecificCounterValues(t *testing.T) {
	var testCounters [types.NxNumCounters]uint64
	testCounters[types.NxCntrObjCksumSet] = 12345
	testCounters[types.NxCntrObjCksumFail] = 67890

	superblock := createTestSuperblockForStatistics(testCounters)
	reader := NewContainerStatisticsReader(superblock)

	if reader.ObjectChecksumSetCount() != 12345 {
		t.Errorf("ObjectChecksumSetCount() = %d, want 12345", reader.ObjectChecksumSetCount())
	}

	if reader.ObjectChecksumFailCount() != 67890 {
		t.Errorf("ObjectChecksumFailCount() = %d, want 67890", reader.ObjectChecksumFailCount())
	}
}

func TestContainerStatisticsReader_CountersSliceIndependence(t *testing.T) {
	var testCounters [types.NxNumCounters]uint64
	for i := 0; i < types.NxNumCounters; i++ {
		testCounters[i] = uint64(i)
	}

	superblock := createTestSuperblockForStatistics(testCounters)
	reader := NewContainerStatisticsReader(superblock)

	// Get counters slice and modify it
	counters := reader.Counters()
	counters[0] = 9999

	// Get counters again and verify original data is unchanged
	counters2 := reader.Counters()
	if counters2[0] != 0 {
		t.Errorf("Modifying returned slice affected internal data: got %d, want 0", counters2[0])
	}
}
