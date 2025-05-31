package datastreams

import (
	"encoding/binary"
	"testing"
)

// createTestDataStreamData creates test data for data stream
func createTestDataStreamData(size, allocedSize, defaultCryptoID, totalBytesWritten, totalBytesRead uint64, endian binary.ByteOrder) []byte {
	// Create data stream data (40 bytes)
	data := make([]byte, 40)
	endian.PutUint64(data[0:8], size)
	endian.PutUint64(data[8:16], allocedSize)
	endian.PutUint64(data[16:24], defaultCryptoID)
	endian.PutUint64(data[24:32], totalBytesWritten)
	endian.PutUint64(data[32:40], totalBytesRead)

	return data
}

func TestDataStreamReader(t *testing.T) {
	tests := []struct {
		name              string
		size              uint64
		allocedSize       uint64
		defaultCryptoID   uint64
		totalBytesWritten uint64
		totalBytesRead    uint64
	}{
		{
			name:              "Empty data stream",
			size:              0,
			allocedSize:       0,
			defaultCryptoID:   0,
			totalBytesWritten: 0,
			totalBytesRead:    0,
		},
		{
			name:              "Small file data stream",
			size:              1024,
			allocedSize:       4096,
			defaultCryptoID:   0x123456789ABCDEF0,
			totalBytesWritten: 1024,
			totalBytesRead:    512,
		},
		{
			name:              "Large file data stream",
			size:              0x100000000, // 4GB
			allocedSize:       0x100001000, // 4GB + 4KB
			defaultCryptoID:   0xFEDCBA9876543210,
			totalBytesWritten: 0x80000000, // 2GB
			totalBytesRead:    0x40000000, // 1GB
		},
		{
			name:              "Active data stream",
			size:              0x10000, // 64KB
			allocedSize:       0x20000, // 128KB
			defaultCryptoID:   0x1111111111111111,
			totalBytesWritten: 0x15000, // 84KB
			totalBytesRead:    0x12000, // 72KB
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endian := binary.LittleEndian

			data := createTestDataStreamData(
				tt.size, tt.allocedSize, tt.defaultCryptoID,
				tt.totalBytesWritten, tt.totalBytesRead, endian)

			reader, err := NewDataStreamReader(data, endian)
			if err != nil {
				t.Fatalf("NewDataStreamReader() error = %v", err)
			}

			// Test Size
			if size := reader.Size(); size != tt.size {
				t.Errorf("Size() = %d, want %d", size, tt.size)
			}

			// Test AllocatedSize
			if allocedSize := reader.AllocatedSize(); allocedSize != tt.allocedSize {
				t.Errorf("AllocatedSize() = %d, want %d", allocedSize, tt.allocedSize)
			}

			// Test DefaultCryptoID
			if cryptoID := reader.DefaultCryptoID(); cryptoID != tt.defaultCryptoID {
				t.Errorf("DefaultCryptoID() = 0x%X, want 0x%X", cryptoID, tt.defaultCryptoID)
			}

			// Test TotalBytesWritten
			if bytesWritten := reader.TotalBytesWritten(); bytesWritten != tt.totalBytesWritten {
				t.Errorf("TotalBytesWritten() = %d, want %d", bytesWritten, tt.totalBytesWritten)
			}

			// Test TotalBytesRead
			if bytesRead := reader.TotalBytesRead(); bytesRead != tt.totalBytesRead {
				t.Errorf("TotalBytesRead() = %d, want %d", bytesRead, tt.totalBytesRead)
			}
		})
	}
}

func TestDataStreamReader_ErrorCases(t *testing.T) {
	endian := binary.LittleEndian

	tests := []struct {
		name     string
		dataSize int
	}{
		{
			name:     "Empty data",
			dataSize: 0,
		},
		{
			name:     "Too small data - 39 bytes",
			dataSize: 39,
		},
		{
			name:     "Too small data - 32 bytes",
			dataSize: 32,
		},
		{
			name:     "Too small data - 16 bytes",
			dataSize: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.dataSize)

			_, err := NewDataStreamReader(data, endian)
			if err == nil {
				t.Error("NewDataStreamReader() expected error, got nil")
			}
		})
	}
}
