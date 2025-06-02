package encryptionrolling

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGeneralBitmapReader(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
		errorMsg    string
	}{
		{
			name:        "insufficient data",
			data:        make([]byte, 32),
			expectError: true,
			errorMsg:    "data too small for general bitmap",
		},
		{
			name:        "valid data",
			data:        createValidGeneralBitmapData(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewGeneralBitmapReader(tt.data, binary.LittleEndian)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, reader)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reader)
			}
		})
	}
}

func TestNewGeneralBitmapBlockReader(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
		errorMsg    string
	}{
		{
			name:        "insufficient data",
			data:        make([]byte, 32),
			expectError: true,
			errorMsg:    "data too small for bitmap block",
		},
		{
			name:        "valid data",
			data:        createValidGeneralBitmapBlockData(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewGeneralBitmapBlockReader(tt.data, binary.LittleEndian)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, reader)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reader)
			}
		})
	}
}

func TestGeneralBitmapReader_Methods(t *testing.T) {
	data := createValidGeneralBitmapData()
	reader, err := NewGeneralBitmapReader(data, binary.LittleEndian)
	require.NoError(t, err)
	require.NotNil(t, reader)

	// Test interface methods
	assert.Equal(t, types.OidT(456), reader.TreeObjectID())
	assert.Equal(t, uint64(1024), reader.BitCount())
	assert.Equal(t, uint64(0x0000000000000001), reader.Flags())
}

func TestGeneralBitmapBlockReader_Methods(t *testing.T) {
	data := createValidGeneralBitmapBlockData()
	reader, err := NewGeneralBitmapBlockReader(data, binary.LittleEndian)
	require.NoError(t, err)
	require.NotNil(t, reader)

	// Test BitmapField returns a copy
	field1 := reader.BitmapField()
	field2 := reader.BitmapField()
	assert.Equal(t, field1, field2)
	assert.NotSame(t, &field1[0], &field2[0]) // Different memory addresses

	// Test specific values
	field := reader.BitmapField()
	require.Len(t, field, 1)
	assert.Equal(t, uint64(0x0F0F0F0F0F0F0F0F), field[0])
}

func TestGeneralBitmapBlockReader_BitOperations(t *testing.T) {
	data := createValidGeneralBitmapBlockData()
	reader, err := NewGeneralBitmapBlockReader(data, binary.LittleEndian)
	require.NoError(t, err)
	require.NotNil(t, reader)

	// Test bit operations on the known pattern 0x0F0F0F0F0F0F0F0F
	// This pattern has bits set at positions: 0,1,2,3, 8,9,10,11, 16,17,18,19, etc.

	// Test IsBitSet for bits that should be set
	assert.True(t, reader.IsBitSet(0))
	assert.True(t, reader.IsBitSet(1))
	assert.True(t, reader.IsBitSet(2))
	assert.True(t, reader.IsBitSet(3))
	assert.True(t, reader.IsBitSet(8))
	assert.True(t, reader.IsBitSet(9))
	assert.True(t, reader.IsBitSet(10))
	assert.True(t, reader.IsBitSet(11))

	// Test IsBitSet for bits that should NOT be set
	assert.False(t, reader.IsBitSet(4))
	assert.False(t, reader.IsBitSet(5))
	assert.False(t, reader.IsBitSet(6))
	assert.False(t, reader.IsBitSet(7))
	assert.False(t, reader.IsBitSet(12))
	assert.False(t, reader.IsBitSet(13))
	assert.False(t, reader.IsBitSet(14))
	assert.False(t, reader.IsBitSet(15))

	// Test SetBit
	reader.SetBit(4) // Set a bit that was previously clear
	assert.True(t, reader.IsBitSet(4))

	// Test ClearBit
	reader.ClearBit(0) // Clear a bit that was previously set
	assert.False(t, reader.IsBitSet(0))

	// Test operations on out-of-bounds bits
	assert.False(t, reader.IsBitSet(999)) // Should return false for out-of-bounds
	reader.SetBit(999)                    // Should not crash
	reader.ClearBit(999)                  // Should not crash
}

func TestGeneralBitmapBlockReader_EmptyBitmap(t *testing.T) {
	data := createEmptyGeneralBitmapBlockData()
	reader, err := NewGeneralBitmapBlockReader(data, binary.LittleEndian)
	require.NoError(t, err)
	require.NotNil(t, reader)

	// Test methods with empty bitmap field (all zeros)
	field := reader.BitmapField()
	assert.Len(t, field, 1)              // Should have one uint64 with value 0
	assert.Equal(t, uint64(0), field[0]) // Should be zero

	assert.False(t, reader.IsBitSet(0))
	reader.SetBit(0)   // Should not crash
	reader.ClearBit(0) // Should not crash
}

func TestParseGeneralBitmap_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil data",
			data:        nil,
			expectError: true,
			errorMsg:    "insufficient data for general bitmap",
		},
		{
			name:        "empty data",
			data:        []byte{},
			expectError: true,
			errorMsg:    "insufficient data for general bitmap",
		},
		{
			name:        "too small data",
			data:        make([]byte, 48),
			expectError: true,
			errorMsg:    "insufficient data for general bitmap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseGeneralBitmap(tt.data, binary.LittleEndian)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestParseGeneralBitmapBlock_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil data",
			data:        nil,
			expectError: true,
			errorMsg:    "insufficient data for bitmap block",
		},
		{
			name:        "empty data",
			data:        []byte{},
			expectError: true,
			errorMsg:    "insufficient data for bitmap block",
		},
		{
			name:        "too small data",
			data:        make([]byte, 32),
			expectError: true,
			errorMsg:    "insufficient data for bitmap block",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseGeneralBitmapBlock(tt.data, binary.LittleEndian)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestParseGeneralBitmapBlock_PartialData(t *testing.T) {
	// Test with data that has one complete uint64 field
	data := createGeneralBitmapBlockDataWithPartialField()
	reader, err := NewGeneralBitmapBlockReader(data, binary.LittleEndian)
	require.NoError(t, err)
	require.NotNil(t, reader)

	field := reader.BitmapField()
	require.Len(t, field, 1)                              // Should have one complete uint64
	assert.Equal(t, uint64(0x0F0F0F0F0F0F0F0F), field[0]) // Verify the expected pattern
}

// Helper functions to create test data

func createValidGeneralBitmapData() []byte {
	data := make([]byte, 56)
	endian := binary.LittleEndian

	offset := 0

	// Object header (32 bytes)
	copy(data[offset:offset+8], []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) // checksum
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(123)) // OID
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(789)) // XID
	offset += 8
	endian.PutUint32(data[offset:offset+4], uint32(types.ObjectTypeGbitmap)) // type
	offset += 4
	endian.PutUint32(data[offset:offset+4], uint32(0)) // subtype
	offset += 4

	// Bitmap specific fields
	endian.PutUint64(data[offset:offset+8], uint64(456)) // tree OID
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(1024)) // bit count
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(0x0000000000000001)) // flags

	return data
}

func createValidGeneralBitmapBlockData() []byte {
	data := make([]byte, 40)
	endian := binary.LittleEndian

	offset := 0

	// Object header (32 bytes)
	copy(data[offset:offset+8], []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) // checksum
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(456)) // OID
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(123)) // XID
	offset += 8
	endian.PutUint32(data[offset:offset+4], uint32(types.ObjectTypeGbitmapBlock)) // type
	offset += 4
	endian.PutUint32(data[offset:offset+4], uint32(0)) // subtype
	offset += 4

	// Bitmap field data (8 bytes = 1 uint64)
	endian.PutUint64(data[offset:offset+8], uint64(0x0F0F0F0F0F0F0F0F)) // bitmap pattern

	return data
}

func createEmptyGeneralBitmapBlockData() []byte {
	data := make([]byte, 40) // Changed from 32 to 40 bytes to include minimum bitmap data
	endian := binary.LittleEndian

	offset := 0

	// Object header (32 bytes)
	copy(data[offset:offset+8], []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) // checksum
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(789)) // OID
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(456)) // XID
	offset += 8
	endian.PutUint32(data[offset:offset+4], uint32(types.ObjectTypeGbitmapBlock)) // type
	offset += 4
	endian.PutUint32(data[offset:offset+4], uint32(0)) // subtype
	offset += 4

	// Empty bitmap field data (8 bytes of zeros)
	endian.PutUint64(data[offset:offset+8], uint64(0)) // empty bitmap

	return data
}

func createGeneralBitmapBlockDataWithPartialField() []byte {
	data := make([]byte, 40) // Changed from 36 to 40 bytes to meet minimum requirement
	endian := binary.LittleEndian

	offset := 0

	// Object header (32 bytes)
	copy(data[offset:offset+8], []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) // checksum
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(321)) // OID
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(654)) // XID
	offset += 8
	endian.PutUint32(data[offset:offset+4], uint32(types.ObjectTypeGbitmapBlock)) // type
	offset += 4
	endian.PutUint32(data[offset:offset+4], uint32(0)) // subtype
	offset += 4

	// Bitmap field data (8 bytes, properly aligned)
	endian.PutUint64(data[offset:offset+8], uint64(0x0F0F0F0F0F0F0F0F)) // bitmap pattern

	return data
}
