package encryptionrolling

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRecoveryBlockReader(t *testing.T) {
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
			errorMsg:    "data too small for recovery block",
		},
		{
			name:        "valid data",
			data:        createValidRecoveryBlockData(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewRecoveryBlockReader(tt.data, binary.LittleEndian)

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

func TestRecoveryBlockReader_Methods(t *testing.T) {
	data := createValidRecoveryBlockData()
	reader, err := NewRecoveryBlockReader(data, binary.LittleEndian)
	require.NoError(t, err)
	require.NotNil(t, reader)

	// Test interface methods
	assert.Equal(t, uint64(1024), reader.Offset())
	assert.Equal(t, types.OidT(456), reader.NextObjectID())

	recoveryData := reader.Data()
	assert.NotNil(t, recoveryData)
	assert.Equal(t, 8, len(recoveryData))

	// Verify it's a copy (modifying returned data shouldn't affect original)
	originalData := reader.Data()
	recoveryData[0] = 0xFF
	newData := reader.Data()
	assert.Equal(t, originalData[0], newData[0])
	assert.NotEqual(t, recoveryData[0], newData[0])
}

func TestRecoveryBlockReader_EmptyData(t *testing.T) {
	data := createRecoveryBlockDataWithoutRecoveryData()
	reader, err := NewRecoveryBlockReader(data, binary.LittleEndian)
	require.NoError(t, err)
	require.NotNil(t, reader)

	// Test interface methods
	assert.Equal(t, uint64(2048), reader.Offset())
	assert.Equal(t, types.OidT(789), reader.NextObjectID())

	recoveryData := reader.Data()
	assert.Nil(t, recoveryData)
}

func TestParseRecoveryBlock_EdgeCases(t *testing.T) {
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
			errorMsg:    "insufficient data for recovery block",
		},
		{
			name:        "empty data",
			data:        []byte{},
			expectError: true,
			errorMsg:    "insufficient data for recovery block",
		},
		{
			name:        "too small data",
			data:        make([]byte, 47), // One less than required 48
			expectError: true,
			errorMsg:    "insufficient data for recovery block",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseRecoveryBlock(tt.data, binary.LittleEndian)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

// Helper functions to create test data

func createValidRecoveryBlockData() []byte {
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
	endian.PutUint32(data[offset:offset+4], uint32(types.ObjectTypeErRecoveryBlock)) // type
	offset += 4
	endian.PutUint32(data[offset:offset+4], uint32(0)) // subtype
	offset += 4

	// Recovery block specific fields
	endian.PutUint64(data[offset:offset+8], uint64(1024)) // offset
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(456)) // next OID
	offset += 8

	// Recovery data (8 bytes)
	for i := 0; i < 8; i++ {
		data[offset+i] = byte(i + 1)
	}

	return data
}

func createRecoveryBlockDataWithoutRecoveryData() []byte {
	data := make([]byte, 48)
	endian := binary.LittleEndian

	offset := 0

	// Object header (32 bytes)
	copy(data[offset:offset+8], []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) // checksum
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(456)) // OID
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(123)) // XID
	offset += 8
	endian.PutUint32(data[offset:offset+4], uint32(types.ObjectTypeErRecoveryBlock)) // type
	offset += 4
	endian.PutUint32(data[offset:offset+4], uint32(0)) // subtype
	offset += 4

	// Recovery block specific fields
	endian.PutUint64(data[offset:offset+8], uint64(2048)) // offset
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(789)) // next OID

	// No recovery data

	return data
}
