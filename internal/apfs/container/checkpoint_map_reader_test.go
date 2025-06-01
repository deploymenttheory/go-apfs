package container

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestCheckpointMapData creates test checkpoint map data
func createTestCheckpointMapData(flags uint32, count uint32, endian binary.ByteOrder) []byte {
	// Calculate size: ObjPhysT (32) + flags (4) + count (4) + mappings (count * 40)
	size := 40 + int(count)*40
	data := make([]byte, size)

	// Object header (32 bytes)
	copy(data[0:8], []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}) // Checksum
	endian.PutUint64(data[8:16], 2000)                                      // OID
	endian.PutUint64(data[16:24], 3000)                                     // XID
	endian.PutUint32(data[24:28], types.ObjectTypeCheckpointMap)            // Type
	endian.PutUint32(data[28:32], 0)                                        // Subtype

	// Checkpoint map fields
	endian.PutUint32(data[32:36], flags)
	endian.PutUint32(data[36:40], count)

	// Create test mappings
	offset := 40
	for i := uint32(0); i < count; i++ {
		// Each mapping is 40 bytes
		endian.PutUint32(data[offset:offset+4], types.ObjectTypeFs) // Type
		endian.PutUint32(data[offset+4:offset+8], 0)                // Subtype
		endian.PutUint32(data[offset+8:offset+12], 4096)            // Size
		endian.PutUint32(data[offset+12:offset+16], 0)              // Pad
		endian.PutUint64(data[offset+16:offset+24], uint64(5000+i)) // FS OID
		endian.PutUint64(data[offset+24:offset+32], uint64(6000+i)) // OID
		endian.PutUint64(data[offset+32:offset+40], uint64(7000+i)) // Physical address
		offset += 40
	}

	return data
}

func TestCheckpointMapReader(t *testing.T) {
	testCases := []struct {
		name             string
		flags            uint32
		count            uint32
		expectIsLast     bool
		expectedMappings int
	}{
		{
			name:             "Empty checkpoint map",
			flags:            0,
			count:            0,
			expectIsLast:     false,
			expectedMappings: 0,
		},
		{
			name:             "Single mapping, not last",
			flags:            0,
			count:            1,
			expectIsLast:     false,
			expectedMappings: 1,
		},
		{
			name:             "Multiple mappings, last block",
			flags:            types.CheckpointMapLast,
			count:            3,
			expectIsLast:     true,
			expectedMappings: 3,
		},
	}

	endian := binary.LittleEndian

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := createTestCheckpointMapData(tc.flags, tc.count, endian)

			reader, err := NewCheckpointMapReader(data, endian)
			if err != nil {
				t.Fatalf("NewCheckpointMapReader() failed: %v", err)
			}

			// Test Flags() method
			if flags := reader.Flags(); flags != tc.flags {
				t.Errorf("Flags() = 0x%X, want 0x%X", flags, tc.flags)
			}

			// Test Count() method
			if count := reader.Count(); count != tc.count {
				t.Errorf("Count() = %d, want %d", count, tc.count)
			}

			// Test IsLast() method
			if isLast := reader.IsLast(); isLast != tc.expectIsLast {
				t.Errorf("IsLast() = %v, want %v", isLast, tc.expectIsLast)
			}

			// Test Mappings() method
			mappings := reader.Mappings()
			if len(mappings) != tc.expectedMappings {
				t.Errorf("Mappings() length = %d, want %d", len(mappings), tc.expectedMappings)
			}
		})
	}
}

func TestCheckpointMapReader_ErrorCases(t *testing.T) {
	endian := binary.LittleEndian

	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name:        "Data too small",
			data:        make([]byte, 30),
			expectError: true,
		},
		{
			name:        "Insufficient data for mappings",
			data:        createTestCheckpointMapData(0, 5, endian)[:100], // Truncate data
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCheckpointMapReader(tt.data, endian)
			if tt.expectError && err == nil {
				t.Error("NewCheckpointMapReader() should have failed")
			}
			if !tt.expectError && err != nil {
				t.Errorf("NewCheckpointMapReader() should not have failed: %v", err)
			}
		})
	}
}
