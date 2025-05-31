package datastreams

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestFileExtentData creates test data for file extent key and value
func createTestFileExtentData(fileOID uint64, logicalAddr uint64, length uint64, flags uint64, physBlockNum uint64, cryptoID uint64, endian binary.ByteOrder) ([]byte, []byte) {
	// Create key data (16 bytes)
	keyData := make([]byte, 16)
	objIdAndType := fileOID & types.ObjIdMask // No type bits set for simplicity
	endian.PutUint64(keyData[0:8], objIdAndType)
	endian.PutUint64(keyData[8:16], logicalAddr)

	// Create value data (24 bytes)
	valueData := make([]byte, 24)
	lenAndFlags := (length & types.JFileExtentLenMask) | ((flags << types.JFileExtentFlagShift) & types.JFileExtentFlagMask)
	endian.PutUint64(valueData[0:8], lenAndFlags)
	endian.PutUint64(valueData[8:16], physBlockNum)
	endian.PutUint64(valueData[16:24], cryptoID)

	return keyData, valueData
}

func TestFileExtentReader(t *testing.T) {
	tests := []struct {
		name             string
		fileOID          uint64
		logicalAddr      uint64
		length           uint64
		flags            uint64
		physBlockNum     uint64
		cryptoID         uint64
		expectedCryptoID uint64
		expectTweak      bool
	}{
		{
			name:             "Basic file extent",
			fileOID:          0x1000,
			logicalAddr:      0,
			length:           0x1000, // 4KB
			flags:            0,
			physBlockNum:     0x2000,
			cryptoID:         0x123456789ABCDEF0,
			expectedCryptoID: 0x123456789ABCDEF0,
			expectTweak:      false,
		},
		{
			name:             "File extent with tweak flag",
			fileOID:          0x2000,
			logicalAddr:      0x1000,
			length:           0x2000, // 8KB
			flags:            uint64(types.FextCryptoIdIsTweak),
			physBlockNum:     0x3000,
			cryptoID:         0xFEDCBA9876543210,
			expectedCryptoID: 0xFEDCBA9876543210,
			expectTweak:      true,
		},
		{
			name:             "Large file extent",
			fileOID:          0x3000,
			logicalAddr:      0x100000, // 1MB offset
			length:           0x100000, // 1MB length
			flags:            0,
			physBlockNum:     0x40000,
			cryptoID:         types.CryptoSwId,
			expectedCryptoID: types.CryptoSwId,
			expectTweak:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endian := binary.LittleEndian

			keyData, valueData := createTestFileExtentData(
				tt.fileOID, tt.logicalAddr, tt.length, tt.flags,
				tt.physBlockNum, tt.cryptoID, endian)

			reader, err := NewFileExtentReader(keyData, valueData, endian)
			if err != nil {
				t.Fatalf("NewFileExtentReader() error = %v", err)
			}

			// Test LogicalAddress
			if addr := reader.LogicalAddress(); addr != tt.logicalAddr {
				t.Errorf("LogicalAddress() = %d, want %d", addr, tt.logicalAddr)
			}

			// Test Length
			if length := reader.Length(); length != tt.length {
				t.Errorf("Length() = %d, want %d", length, tt.length)
			}

			// Test Flags
			if flags := reader.Flags(); flags != tt.flags {
				t.Errorf("Flags() = %d, want %d", flags, tt.flags)
			}

			// Test PhysicalBlockNumber
			if physBlock := reader.PhysicalBlockNumber(); physBlock != tt.physBlockNum {
				t.Errorf("PhysicalBlockNumber() = %d, want %d", physBlock, tt.physBlockNum)
			}

			// Test CryptoID
			if cryptoID := reader.CryptoID(); cryptoID != tt.expectedCryptoID {
				t.Errorf("CryptoID() = 0x%X, want 0x%X", cryptoID, tt.expectedCryptoID)
			}

			// Test IsCryptoIDTweak
			if isTweak := reader.IsCryptoIDTweak(); isTweak != tt.expectTweak {
				t.Errorf("IsCryptoIDTweak() = %v, want %v", isTweak, tt.expectTweak)
			}
		})
	}
}

func TestFileExtentReader_ErrorCases(t *testing.T) {
	endian := binary.LittleEndian

	tests := []struct {
		name      string
		keySize   int
		valueSize int
	}{
		{
			name:      "Empty key data",
			keySize:   0,
			valueSize: 24,
		},
		{
			name:      "Too small key data",
			keySize:   15,
			valueSize: 24,
		},
		{
			name:      "Empty value data",
			keySize:   16,
			valueSize: 0,
		},
		{
			name:      "Too small value data",
			keySize:   16,
			valueSize: 23,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyData := make([]byte, tt.keySize)
			valueData := make([]byte, tt.valueSize)

			_, err := NewFileExtentReader(keyData, valueData, endian)
			if err == nil {
				t.Error("NewFileExtentReader() expected error, got nil")
			}
		})
	}
}
