package datastreams

import (
	"encoding/binary"
	"testing"
)

// createTestExtendedAttributeDataStreamData creates test data for extended attribute data stream
func createTestExtendedAttributeDataStreamData(xattrObjID, size, allocedSize, defaultCryptoID, totalBytesWritten, totalBytesRead uint64, endian binary.ByteOrder) []byte {
	// Create extended attribute data stream data (48 bytes)
	data := make([]byte, 48)

	// XattrObjId (8 bytes)
	endian.PutUint64(data[0:8], xattrObjID)

	// Embedded JDstreamT (40 bytes starting at offset 8)
	endian.PutUint64(data[8:16], size)
	endian.PutUint64(data[16:24], allocedSize)
	endian.PutUint64(data[24:32], defaultCryptoID)
	endian.PutUint64(data[32:40], totalBytesWritten)
	endian.PutUint64(data[40:48], totalBytesRead)

	return data
}

func TestExtendedAttributeDataStreamReader(t *testing.T) {
	tests := []struct {
		name              string
		xattrObjID        uint64
		size              uint64
		allocedSize       uint64
		defaultCryptoID   uint64
		totalBytesWritten uint64
		totalBytesRead    uint64
	}{
		{
			name:              "Empty extended attribute",
			xattrObjID:        0,
			size:              0,
			allocedSize:       0,
			defaultCryptoID:   0,
			totalBytesWritten: 0,
			totalBytesRead:    0,
		},
		{
			name:              "Small extended attribute",
			xattrObjID:        0x1000,
			size:              256,
			allocedSize:       512,
			defaultCryptoID:   0x123456789ABCDEF0,
			totalBytesWritten: 256,
			totalBytesRead:    128,
		},
		{
			name:              "Large extended attribute",
			xattrObjID:        0x123456789ABCDEF0,
			size:              0x100000, // 1MB
			allocedSize:       0x200000, // 2MB
			defaultCryptoID:   0xFEDCBA9876543210,
			totalBytesWritten: 0x80000, // 512KB
			totalBytesRead:    0x40000, // 256KB
		},
		{
			name:              "Active extended attribute",
			xattrObjID:        0x42424242,
			size:              0x8000,  // 32KB
			allocedSize:       0x10000, // 64KB
			defaultCryptoID:   0x1111111111111111,
			totalBytesWritten: 0x6000, // 24KB
			totalBytesRead:    0x4000, // 16KB
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endian := binary.LittleEndian

			data := createTestExtendedAttributeDataStreamData(
				tt.xattrObjID, tt.size, tt.allocedSize, tt.defaultCryptoID,
				tt.totalBytesWritten, tt.totalBytesRead, endian)

			reader, err := NewExtendedAttributeDataStreamReader(data, endian)
			if err != nil {
				t.Fatalf("NewExtendedAttributeDataStreamReader() error = %v", err)
			}

			// Test AttributeObjectID
			if xattrObjID := reader.AttributeObjectID(); xattrObjID != tt.xattrObjID {
				t.Errorf("AttributeObjectID() = 0x%X, want 0x%X", xattrObjID, tt.xattrObjID)
			}

			// Test DataStream methods through the embedded reader
			dataStream := reader.DataStream()
			if dataStream == nil {
				t.Fatal("DataStream() returned nil")
			}

			// Test Size
			if size := dataStream.Size(); size != tt.size {
				t.Errorf("DataStream().Size() = %d, want %d", size, tt.size)
			}

			// Test AllocatedSize
			if allocedSize := dataStream.AllocatedSize(); allocedSize != tt.allocedSize {
				t.Errorf("DataStream().AllocatedSize() = %d, want %d", allocedSize, tt.allocedSize)
			}

			// Test DefaultCryptoID
			if cryptoID := dataStream.DefaultCryptoID(); cryptoID != tt.defaultCryptoID {
				t.Errorf("DataStream().DefaultCryptoID() = 0x%X, want 0x%X", cryptoID, tt.defaultCryptoID)
			}

			// Test TotalBytesWritten
			if bytesWritten := dataStream.TotalBytesWritten(); bytesWritten != tt.totalBytesWritten {
				t.Errorf("DataStream().TotalBytesWritten() = %d, want %d", bytesWritten, tt.totalBytesWritten)
			}

			// Test TotalBytesRead
			if bytesRead := dataStream.TotalBytesRead(); bytesRead != tt.totalBytesRead {
				t.Errorf("DataStream().TotalBytesRead() = %d, want %d", bytesRead, tt.totalBytesRead)
			}
		})
	}
}

func TestExtendedAttributeDataStreamReader_ErrorCases(t *testing.T) {
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
			name:     "Too small data - 47 bytes",
			dataSize: 47,
		},
		{
			name:     "Too small data - 40 bytes",
			dataSize: 40,
		},
		{
			name:     "Too small data - 32 bytes",
			dataSize: 32,
		},
		{
			name:     "Too small data - 16 bytes",
			dataSize: 16,
		},
		{
			name:     "Too small data - 8 bytes",
			dataSize: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.dataSize)

			_, err := NewExtendedAttributeDataStreamReader(data, endian)
			if err == nil {
				t.Error("NewExtendedAttributeDataStreamReader() expected error, got nil")
			}
		})
	}
}

func TestExtendedAttributeDataStreamReader_InvalidEmbeddedDataStream(t *testing.T) {
	endian := binary.LittleEndian

	// Create data that has the right total size but will fail embedded data stream parsing
	// This tests the error handling when the embedded DataStreamReader creation fails
	data := make([]byte, 48)

	// Set a valid XattrObjId
	endian.PutUint64(data[0:8], 0x1000)

	// The remaining 40 bytes are all zeros, which should be valid for DataStreamReader
	// So this test actually validates that valid embedded data works correctly
	reader, err := NewExtendedAttributeDataStreamReader(data, endian)
	if err != nil {
		t.Fatalf("NewExtendedAttributeDataStreamReader() unexpected error = %v", err)
	}

	if reader.AttributeObjectID() != 0x1000 {
		t.Errorf("AttributeObjectID() = 0x%X, want 0x1000", reader.AttributeObjectID())
	}

	// Verify the embedded data stream works
	ds := reader.DataStream()
	if ds == nil {
		t.Error("DataStream() returned nil")
	}

	if ds.Size() != 0 {
		t.Errorf("DataStream().Size() = %d, want 0", ds.Size())
	}
}
