package encryptionrolling

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
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
		magic       uint32
		version     uint32
		flags       uint64
		endian      binary.ByteOrder
		expectError bool
	}{
		{
			name:        "valid little endian",
			magic:       types.ErMagic,
			version:     1,
			flags:       types.ErsbFlagEncrypting,
			endian:      binary.LittleEndian,
			expectError: false,
		},
		{
			name:        "valid big endian",
			magic:       types.ErMagic,
			version:     1,
			flags:       types.ErsbFlagDecrypting,
			endian:      binary.BigEndian,
			expectError: false,
		},
		{
			name:        "invalid magic",
			magic:       0x12345678,
			version:     1,
			flags:       types.ErsbFlagEncrypting,
			endian:      binary.LittleEndian,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createTestEncryptionRollingStateData(tt.magic, tt.version, tt.flags, tt.endian)

			reader, err := NewEncryptionRollingStateReader(data, tt.endian)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify basic properties
			if reader.Magic() != tt.magic {
				t.Errorf("expected magic 0x%08X, got 0x%08X", tt.magic, reader.Magic())
			}

			if reader.Version() != tt.version {
				t.Errorf("expected version %d, got %d", tt.version, reader.Version())
			}

			if reader.Flags() != tt.flags {
				t.Errorf("expected flags 0x%016X, got 0x%016X", tt.flags, reader.Flags())
			}
		})
	}
}

func TestNewEncryptionRollingV1StateReader(t *testing.T) {
	tests := []struct {
		name          string
		magic         uint32
		version       uint32
		flags         uint64
		checksumCount uint32
		endian        binary.ByteOrder
		expectError   bool
	}{
		{
			name:          "valid v1 with checksums",
			magic:         types.ErMagic,
			version:       1,
			flags:         types.ErsbFlagKeyrolling,
			checksumCount: 3,
			endian:        binary.LittleEndian,
			expectError:   false,
		},
		{
			name:          "valid v1 no checksums",
			magic:         types.ErMagic,
			version:       1,
			flags:         types.ErsbFlagPaused,
			checksumCount: 0,
			endian:        binary.BigEndian,
			expectError:   false,
		},
		{
			name:          "invalid magic",
			magic:         0x87654321,
			version:       1,
			flags:         types.ErsbFlagEncrypting,
			checksumCount: 1,
			endian:        binary.LittleEndian,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createTestEncryptionRollingV1StateData(tt.magic, tt.version, tt.flags, tt.checksumCount, tt.endian)

			reader, err := NewEncryptionRollingV1StateReader(data, tt.endian)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify basic properties
			if reader.Magic() != tt.magic {
				t.Errorf("expected magic 0x%08X, got 0x%08X", tt.magic, reader.Magic())
			}

			if reader.Version() != tt.version {
				t.Errorf("expected version %d, got %d", tt.version, reader.Version())
			}

			if reader.ChecksumCount() != tt.checksumCount {
				t.Errorf("expected checksum count %d, got %d", tt.checksumCount, reader.ChecksumCount())
			}

			// Verify checksum data length
			expectedChecksumDataSize := int(tt.checksumCount * types.ErChecksumLength)
			actualChecksumData := reader.Checksums()
			if len(actualChecksumData) != expectedChecksumDataSize {
				t.Errorf("expected checksum data size %d, got %d", expectedChecksumDataSize, len(actualChecksumData))
			}
		})
	}
}

func TestEncryptionRollingStateReaderMethods(t *testing.T) {
	flags := types.ErsbFlagEncrypting | types.ErsbFlagCidIsTweak
	data := createTestEncryptionRollingStateData(types.ErMagic, 1, flags, binary.LittleEndian)

	reader, err := NewEncryptionRollingStateReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}

	// Test specific method values
	if reader.SnapshotXID() != 12345 {
		t.Errorf("expected snapshot XID 12345, got %d", reader.SnapshotXID())
	}

	if reader.CurrentFileExtentObjectID() != 67890 {
		t.Errorf("expected file extent object ID 67890, got %d", reader.CurrentFileExtentObjectID())
	}

	if reader.FileOffset() != 1024 {
		t.Errorf("expected file offset 1024, got %d", reader.FileOffset())
	}

	if reader.Progress() != 500 {
		t.Errorf("expected progress 500, got %d", reader.Progress())
	}

	if reader.TotalBlocksToEncrypt() != 1000 {
		t.Errorf("expected total blocks 1000, got %d", reader.TotalBlocksToEncrypt())
	}

	if reader.BlockmapOID() != types.OidT(0x987654321) {
		t.Errorf("expected blockmap OID 0x987654321, got 0x%X", reader.BlockmapOID())
	}

	if reader.TidemarkObjectID() != 0x111222333 {
		t.Errorf("expected tidemark object ID 0x111222333, got 0x%X", reader.TidemarkObjectID())
	}

	if reader.RecoveryExtentsCount() != 5 {
		t.Errorf("expected recovery extents count 5, got %d", reader.RecoveryExtentsCount())
	}

	if reader.RecoveryListOID() != types.OidT(0x444555666) {
		t.Errorf("expected recovery list OID 0x444555666, got 0x%X", reader.RecoveryListOID())
	}
}

func TestEncryptionRollingV1StateReaderMethods(t *testing.T) {
	flags := types.ErsbFlagDecrypting | types.ErsbFlagPaused
	checksumCount := uint32(2)
	data := createTestEncryptionRollingV1StateData(types.ErMagic, 1, flags, checksumCount, binary.LittleEndian)

	reader, err := NewEncryptionRollingV1StateReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("failed to create V1 reader: %v", err)
	}

	// Test V1-specific methods
	if reader.FileExtentPhysicalBlockNumber() != 2048 {
		t.Errorf("expected file extent PBN 2048, got %d", reader.FileExtentPhysicalBlockNumber())
	}

	if reader.PhysicalAddress() != 4096 {
		t.Errorf("expected physical address 4096, got %d", reader.PhysicalAddress())
	}

	if reader.BlockmapOID() != 0x987654321 {
		t.Errorf("expected blockmap OID 0x987654321, got 0x%X", reader.BlockmapOID())
	}

	if reader.FileExtentCryptoID() != 0x777888999 {
		t.Errorf("expected file extent crypto ID 0x777888999, got 0x%X", reader.FileExtentCryptoID())
	}

	// Verify checksum data pattern
	checksumData := reader.Checksums()
	expectedSize := int(checksumCount * types.ErChecksumLength)
	if len(checksumData) != expectedSize {
		t.Errorf("expected checksum data size %d, got %d", expectedSize, len(checksumData))
	}

	// Verify the pattern we set in createTestEncryptionRollingV1StateData
	for i := uint32(0); i < checksumCount; i++ {
		for j := uint32(0); j < types.ErChecksumLength; j++ {
			expectedByte := byte(i*10 + j)
			actualByte := checksumData[i*types.ErChecksumLength+j]
			if actualByte != expectedByte {
				t.Errorf("checksum data mismatch at position %d: expected %d, got %d",
					i*types.ErChecksumLength+j, expectedByte, actualByte)
			}
		}
	}
}

func TestParseEncryptionRollingStateErrors(t *testing.T) {
	tests := []struct {
		name     string
		dataSize int
	}{
		{"insufficient data - empty", 0},
		{"insufficient data - too small", 50},
		{"insufficient data - just header", 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.dataSize)
			_, err := NewEncryptionRollingStateReader(data, binary.LittleEndian)
			if err == nil {
				t.Errorf("expected error for data size %d but got none", tt.dataSize)
			}
		})
	}
}

func TestParseEncryptionRollingV1StateErrors(t *testing.T) {
	tests := []struct {
		name     string
		dataSize int
	}{
		{"insufficient data - empty", 0},
		{"insufficient data - too small", 100},
		{"insufficient data - just header", 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.dataSize)
			_, err := NewEncryptionRollingV1StateReader(data, binary.LittleEndian)
			if err == nil {
				t.Errorf("expected error for data size %d but got none", tt.dataSize)
			}
		})
	}
}

func TestEncryptionRollingStateReaderNilEndian(t *testing.T) {
	data := createTestEncryptionRollingStateData(types.ErMagic, 1, types.ErsbFlagEncrypting, binary.LittleEndian)

	// Test with nil endian (should default to little endian)
	reader, err := NewEncryptionRollingStateReader(data, nil)
	if err != nil {
		t.Fatalf("failed to create reader with nil endian: %v", err)
	}

	if reader.Magic() != types.ErMagic {
		t.Errorf("expected magic 0x%08X, got 0x%08X", types.ErMagic, reader.Magic())
	}
}

func TestEncryptionRollingV1StateReaderNilEndian(t *testing.T) {
	data := createTestEncryptionRollingV1StateData(types.ErMagic, 1, types.ErsbFlagEncrypting, 1, binary.LittleEndian)

	// Test with nil endian (should default to little endian)
	reader, err := NewEncryptionRollingV1StateReader(data, nil)
	if err != nil {
		t.Fatalf("failed to create V1 reader with nil endian: %v", err)
	}

	if reader.Magic() != types.ErMagic {
		t.Errorf("expected magic 0x%08X, got 0x%08X", types.ErMagic, reader.Magic())
	}
}
