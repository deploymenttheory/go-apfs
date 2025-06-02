package encryptionrolling

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestEncryptionRollingStateData creates test encryption rolling state data
func createTestEncryptionRollingStateData(magic uint32, version uint32, flags uint64, endian binary.ByteOrder) []byte {
	data := make([]byte, 120) // Minimum size for encryption rolling state
	offset := 0

	// Object header (32 bytes)
	// Checksum (8 bytes) - zeros
	offset += 8

	// Object ID
	endian.PutUint64(data[offset:], uint64(0x123456789ABCDEF0))
	offset += 8

	// Transaction ID
	endian.PutUint64(data[offset:], uint64(0xFEDCBA9876543210))
	offset += 8

	// Object Type
	endian.PutUint32(data[offset:], types.ObjectTypeErState)
	offset += 4

	// Object Subtype
	endian.PutUint32(data[offset:], 0)
	offset += 4

	// Header fields
	endian.PutUint32(data[offset:], magic)
	offset += 4
	endian.PutUint32(data[offset:], version)
	offset += 4

	// State fields
	endian.PutUint64(data[offset:], flags) // ErsbFlags
	offset += 8
	endian.PutUint64(data[offset:], 12345) // ErsbSnapXid
	offset += 8
	endian.PutUint64(data[offset:], 67890) // ErsbCurrentFextObjId
	offset += 8
	endian.PutUint64(data[offset:], 1024) // ErsbFileOffset
	offset += 8
	endian.PutUint64(data[offset:], 500) // ErsbProgress
	offset += 8
	endian.PutUint64(data[offset:], 1000) // ErsbTotalBlkToEncrypt
	offset += 8
	endian.PutUint64(data[offset:], 0x987654321) // ErsbBlockmapOid
	offset += 8
	endian.PutUint64(data[offset:], 0x111222333) // ErsbTidemarkObjId
	offset += 8
	endian.PutUint64(data[offset:], 5) // ErsbRecoveryExtentsCount
	offset += 8
	endian.PutUint64(data[offset:], 0x444555666) // ErsbRecoveryListOid
	offset += 8

	return data
}

// createTestEncryptionRollingV1StateData creates test encryption rolling V1 state data
func createTestEncryptionRollingV1StateData(magic uint32, version uint32, flags uint64, checksumCount uint32, endian binary.ByteOrder) []byte {
	baseSize := 128 // Minimum size for V1 state
	checksumDataSize := int(checksumCount * types.ErChecksumLength)
	data := make([]byte, baseSize+checksumDataSize)
	offset := 0

	// Object header (32 bytes)
	// Checksum (8 bytes) - zeros
	offset += 8

	// Object ID
	endian.PutUint64(data[offset:], uint64(0x123456789ABCDEF0))
	offset += 8

	// Transaction ID
	endian.PutUint64(data[offset:], uint64(0xFEDCBA9876543210))
	offset += 8

	// Object Type
	endian.PutUint32(data[offset:], types.ObjectTypeErState)
	offset += 4

	// Object Subtype
	endian.PutUint32(data[offset:], 0)
	offset += 4

	// Header fields
	endian.PutUint32(data[offset:], magic)
	offset += 4
	endian.PutUint32(data[offset:], version)
	offset += 4

	// V1 State fields
	endian.PutUint64(data[offset:], flags) // ErsbFlags
	offset += 8
	endian.PutUint64(data[offset:], 12345) // ErsbSnapXid
	offset += 8
	endian.PutUint64(data[offset:], 67890) // ErsbCurrentFextObjId
	offset += 8
	endian.PutUint64(data[offset:], 1024) // ErsbFileOffset
	offset += 8
	endian.PutUint64(data[offset:], 2048) // ErsbFextPbn
	offset += 8
	endian.PutUint64(data[offset:], 4096) // ErsbPaddr
	offset += 8
	endian.PutUint64(data[offset:], 500) // ErsbProgress
	offset += 8
	endian.PutUint64(data[offset:], 1000) // ErsbTotalBlkToEncrypt
	offset += 8
	endian.PutUint64(data[offset:], 0x987654321) // ErsbBlockmapOid
	offset += 8
	endian.PutUint32(data[offset:], checksumCount) // ErsbChecksumCount
	offset += 4
	endian.PutUint32(data[offset:], 0) // ErsbReserved
	offset += 4
	endian.PutUint64(data[offset:], 0x777888999) // ErsbFextCid
	offset += 8

	// Checksum data
	for i := uint32(0); i < checksumCount; i++ {
		for j := uint32(0); j < types.ErChecksumLength; j++ {
			data[offset] = byte(i*10 + j)
			offset++
		}
	}

	return data
}

func TestNewEncryptionRollingStateReader(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
		errorMsg    string
	}{
		{
			name:        "insufficient data",
			data:        make([]byte, 64),
			expectError: true,
			errorMsg:    "data too small for encryption rolling state",
		},
		{
			name:        "valid data",
			data:        createValidEncryptionRollingStateData(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewEncryptionRollingStateReader(tt.data, binary.LittleEndian)

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

func TestNewEncryptionRollingV1StateReader(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
		errorMsg    string
	}{
		{
			name:        "insufficient data",
			data:        make([]byte, 64),
			expectError: true,
			errorMsg:    "data too small for encryption rolling v1 state",
		},
		{
			name:        "valid data",
			data:        createValidEncryptionRollingV1StateData(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewEncryptionRollingV1StateReader(tt.data, binary.LittleEndian)

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

func TestEncryptionRollingStateReader_Methods(t *testing.T) {
	data := createValidEncryptionRollingStateData()
	reader, err := NewEncryptionRollingStateReader(data, binary.LittleEndian)
	require.NoError(t, err)
	require.NotNil(t, reader)

	// Test all interface methods
	assert.Equal(t, uint32(0x12345678), reader.Magic())
	assert.Equal(t, uint32(1), reader.Version())
	assert.Equal(t, uint64(0x0000000000000001), reader.Flags())
	assert.Equal(t, types.XidT(100), reader.SnapshotXID())
	assert.Equal(t, uint64(200), reader.CurrentFileExtentObjectID())
	assert.Equal(t, uint64(300), reader.FileOffset())
	assert.Equal(t, uint64(400), reader.Progress())
	assert.Equal(t, uint64(500), reader.TotalBlocksToEncrypt())
	assert.Equal(t, types.OidT(600), reader.BlockmapOID())
	assert.Equal(t, uint64(700), reader.TidemarkObjectID())
	assert.Equal(t, uint64(800), reader.RecoveryExtentsCount())
	assert.Equal(t, types.OidT(900), reader.RecoveryListOID())
	assert.Equal(t, uint64(1000), reader.RecoveryLength())
}

func TestEncryptionRollingV1StateReader_Methods(t *testing.T) {
	data := createValidEncryptionRollingV1StateData()
	reader, err := NewEncryptionRollingV1StateReader(data, binary.LittleEndian)
	require.NoError(t, err)
	require.NotNil(t, reader)

	// Test all interface methods
	assert.Equal(t, uint32(0x12345678), reader.Magic())
	assert.Equal(t, uint32(1), reader.Version())
	assert.Equal(t, uint64(0x0000000000000001), reader.Flags())
	assert.Equal(t, types.XidT(100), reader.SnapshotXID())
	assert.Equal(t, uint64(200), reader.CurrentFileExtentObjectID())
	assert.Equal(t, uint64(300), reader.FileOffset())
	assert.Equal(t, uint64(400), reader.FileExtentPhysicalBlockNumber())
	assert.Equal(t, uint64(500), reader.PhysicalAddress())
	assert.Equal(t, uint64(600), reader.Progress())
	assert.Equal(t, uint64(700), reader.TotalBlocksToEncrypt())
	assert.Equal(t, uint64(800), reader.BlockmapOID())
	assert.Equal(t, uint32(2), reader.ChecksumCount())
	assert.Equal(t, uint64(900), reader.FileExtentCryptoID())

	checksums := reader.Checksums()
	assert.NotNil(t, checksums)
	assert.Equal(t, 16, len(checksums)) // 2 checksums * 8 bytes each
}

func TestParseEncryptionRollingState_EdgeCases(t *testing.T) {
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
			errorMsg:    "insufficient data",
		},
		{
			name:        "empty data",
			data:        []byte{},
			expectError: true,
			errorMsg:    "insufficient data",
		},
		{
			name:        "too small data",
			data:        make([]byte, 127), // Changed to 127 (one less than required 128)
			expectError: true,
			errorMsg:    "insufficient data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseEncryptionRollingState(tt.data, binary.LittleEndian)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestParseEncryptionRollingV1State_EdgeCases(t *testing.T) {
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
			errorMsg:    "insufficient data",
		},
		{
			name:        "empty data",
			data:        []byte{},
			expectError: true,
			errorMsg:    "insufficient data",
		},
		{
			name:        "too small data",
			data:        make([]byte, 135), // One less than required 136
			expectError: true,
			errorMsg:    "insufficient data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseEncryptionRollingV1State(tt.data, binary.LittleEndian)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestEncryptionRollingV1StateReader_NoChecksums(t *testing.T) {
	data := createValidEncryptionRollingV1StateDataNoChecksums()
	reader, err := NewEncryptionRollingV1StateReader(data, binary.LittleEndian)
	require.NoError(t, err)
	require.NotNil(t, reader)

	assert.Equal(t, uint32(0), reader.ChecksumCount())
	checksums := reader.Checksums()
	assert.Nil(t, checksums)
}

// Helper functions to create test data

func createValidEncryptionRollingStateData() []byte {
	data := make([]byte, 200)
	endian := binary.LittleEndian

	offset := 0

	// Object header (32 bytes)
	copy(data[offset:offset+8], []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) // checksum
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(123)) // OID
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(456)) // XID
	offset += 8
	endian.PutUint32(data[offset:offset+4], uint32(types.ObjectTypeErState)) // type
	offset += 4
	endian.PutUint32(data[offset:offset+4], uint32(0)) // subtype
	offset += 4

	// Header specific fields
	endian.PutUint32(data[offset:offset+4], uint32(0x12345678)) // magic
	offset += 4
	endian.PutUint32(data[offset:offset+4], uint32(1)) // version
	offset += 4
	offset += 8 // padding

	// State fields
	endian.PutUint64(data[offset:offset+8], uint64(0x0000000000000001)) // flags
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(100)) // snap XID
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(200)) // current fext obj ID
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(300)) // file offset
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(400)) // progress
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(500)) // total blocks to encrypt
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(600)) // blockmap OID
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(700)) // tidemark obj ID
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(800)) // recovery extents count
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(900)) // recovery list OID
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(1000)) // recovery length

	return data
}

func createValidEncryptionRollingV1StateData() []byte {
	data := make([]byte, 200)
	endian := binary.LittleEndian

	offset := 0

	// Object header (32 bytes)
	copy(data[offset:offset+8], []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) // checksum
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(123)) // OID
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(456)) // XID
	offset += 8
	endian.PutUint32(data[offset:offset+4], uint32(types.ObjectTypeErState)) // type
	offset += 4
	endian.PutUint32(data[offset:offset+4], uint32(0)) // subtype
	offset += 4

	// Header specific fields
	endian.PutUint32(data[offset:offset+4], uint32(0x12345678)) // magic
	offset += 4
	endian.PutUint32(data[offset:offset+4], uint32(1)) // version
	offset += 4
	offset += 8 // padding

	// V1 specific fields
	endian.PutUint64(data[offset:offset+8], uint64(0x0000000000000001)) // flags
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(100)) // snap XID
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(200)) // current fext obj ID
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(300)) // file offset
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(400)) // fext PBN
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(500)) // paddr
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(600)) // progress
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(700)) // total blocks to encrypt
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(800)) // blockmap OID
	offset += 8
	endian.PutUint32(data[offset:offset+4], uint32(2)) // checksum count
	offset += 4
	endian.PutUint32(data[offset:offset+4], uint32(0)) // reserved
	offset += 4
	endian.PutUint64(data[offset:offset+8], uint64(900)) // fext CID

	// Add checksum data (2 * 8 bytes = 16 bytes)
	offset += 8
	for i := 0; i < 16; i++ {
		data[offset+i] = byte(i + 1)
	}

	return data
}

func createValidEncryptionRollingV1StateDataNoChecksums() []byte {
	data := make([]byte, 136) // Changed from 128 to 136 bytes to include all fields
	endian := binary.LittleEndian

	offset := 0

	// Object header (32 bytes)
	copy(data[offset:offset+8], []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) // checksum
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(123)) // OID
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(456)) // XID
	offset += 8
	endian.PutUint32(data[offset:offset+4], uint32(types.ObjectTypeErState)) // type
	offset += 4
	endian.PutUint32(data[offset:offset+4], uint32(0)) // subtype
	offset += 4

	// Header specific fields (16 bytes)
	endian.PutUint32(data[offset:offset+4], uint32(0x12345678)) // magic
	offset += 4
	endian.PutUint32(data[offset:offset+4], uint32(1)) // version
	offset += 4
	offset += 8 // padding

	// V1 specific fields (88 bytes total)
	endian.PutUint64(data[offset:offset+8], uint64(0x0000000000000001)) // flags
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(100)) // snap XID
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(200)) // current fext obj ID
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(300)) // file offset
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(400)) // fext PBN
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(500)) // paddr
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(600)) // progress
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(700)) // total blocks to encrypt
	offset += 8
	endian.PutUint64(data[offset:offset+8], uint64(800)) // blockmap OID
	offset += 8
	endian.PutUint32(data[offset:offset+4], uint32(0)) // checksum count = 0
	offset += 4
	endian.PutUint32(data[offset:offset+4], uint32(0)) // reserved
	offset += 4
	endian.PutUint64(data[offset:offset+8], uint64(900)) // fext CID

	return data
}
