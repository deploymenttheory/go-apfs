package container

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestEvictMappingData creates test evict mapping data
func createTestEvictMappingData(dstPaddr types.Paddr, length uint64, endian binary.ByteOrder) []byte {
	data := make([]byte, 16) // EvictMappingValT is 16 bytes

	endian.PutUint64(data[0:8], uint64(dstPaddr))
	endian.PutUint64(data[8:16], length)

	return data
}

func TestEvictMappingReader(t *testing.T) {
	testCases := []struct {
		name     string
		dstPaddr types.Paddr
		length   uint64
	}{
		{
			name:     "Zero values",
			dstPaddr: types.Paddr(0),
			length:   0,
		},
		{
			name:     "Small mapping",
			dstPaddr: types.Paddr(1000),
			length:   10,
		},
		{
			name:     "Large mapping",
			dstPaddr: types.Paddr(1000000),
			length:   50000,
		},
	}

	endian := binary.LittleEndian

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := createTestEvictMappingData(tc.dstPaddr, tc.length, endian)

			reader, err := NewEvictMappingReader(data, endian)
			if err != nil {
				t.Fatalf("NewEvictMappingReader() failed: %v", err)
			}

			// Test DestinationAddress() method
			if dstAddr := reader.DestinationAddress(); dstAddr != tc.dstPaddr {
				t.Errorf("DestinationAddress() = %d, want %d", dstAddr, tc.dstPaddr)
			}

			// Test Length() method
			if length := reader.Length(); length != tc.length {
				t.Errorf("Length() = %d, want %d", length, tc.length)
			}
		})
	}
}

func TestEvictMappingReader_ErrorCases(t *testing.T) {
	endian := binary.LittleEndian

	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name:        "Empty data",
			data:        []byte{},
			expectError: true,
		},
		{
			name:        "Data too small",
			data:        make([]byte, 8),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEvictMappingReader(tt.data, endian)
			if tt.expectError && err == nil {
				t.Error("NewEvictMappingReader() should have failed")
			}
			if !tt.expectError && err != nil {
				t.Errorf("NewEvictMappingReader() should not have failed: %v", err)
			}
		})
	}
}
