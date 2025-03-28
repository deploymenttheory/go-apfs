// File: pkg/util/io_test.go
package util

import (
	"errors"
	"os"
	"testing"

	"github.com/deploymenttheory/go-apfs/pkg/types"
)

func TestFileDeviceReadWriteBlock(t *testing.T) {
	blockSize := uint32(4096)

	// Create and open a temp file for read/write
	tempFile, err := os.CreateTemp("", "apfs-dev-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	// Pre-fill the file with one block of data before opening it with FileDevice
	data := make([]byte, blockSize)
	for i := range data {
		data[i] = byte(i % 256)
	}
	if _, err := tempFile.WriteAt(data, 0); err != nil {
		t.Fatalf("failed to write initial block to temp file: %v", err)
	}
	tempFile.Close()

	device, err := OpenFileDevice(tempFile.Name(), blockSize)
	if err != nil {
		t.Fatalf("failed to open file device: %v", err)
	}
	defer device.Close()

	err = device.WriteBlock(0, data)
	if err != nil {
		t.Errorf("WriteBlock failed: %v", err)
	}

	readBack, err := device.ReadBlock(0)
	if err != nil {
		t.Errorf("ReadBlock failed: %v", err)
	}
	if len(readBack) != int(blockSize) {
		t.Errorf("unexpected read length: got %d, want %d", len(readBack), blockSize)
	}
	for i := range data {
		if data[i] != readBack[i] {
			t.Errorf("mismatch at byte %d: got %x, want %x", i, readBack[i], data[i])
			break
		}
	}
}

func TestReadAtInvalidOffset(t *testing.T) {
	tempFile, _ := os.CreateTemp("", "apfs-dev-test")
	defer os.Remove(tempFile.Name())
	tempFile.Write(make([]byte, 4096))
	tempFile.Close()

	blockSize := uint32(4096)
	device, _ := OpenFileDevice(tempFile.Name(), blockSize)
	defer device.Close()

	buf := make([]byte, 16)
	_, err := device.ReadAt(buf, int64(999999999))
	if err == nil {
		t.Error("expected error for out-of-range read, got nil")
	}
	if !types.IsIOError(err) && !errors.Is(err, types.ErrInvalidBlockAddr) {
		t.Errorf("expected IO or invalid block address error, got: %v", err)
	}
}

func TestInvalidBlockSizeWrite(t *testing.T) {
	tempFile, _ := os.CreateTemp("", "apfs-dev-test")
	defer os.Remove(tempFile.Name())
	tempFile.Write(make([]byte, 4096))
	tempFile.Close()

	device, _ := OpenFileDevice(tempFile.Name(), 4096)
	err := device.WriteBlock(0, []byte{0x00})
	if err == nil {
		t.Error("expected error for invalid block size write")
	}
	if !errors.Is(err, types.ErrInvalidBlockSize) && err.Error() == "" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSeekBlock(t *testing.T) {
	tests := []struct {
		blockNum  uint64
		blockSize uint32
		expected  int64
	}{
		{0, 4096, 0},
		{1, 4096, 4096},
		{2, 4096, 8192},
		{100, 512, 51200},
	}
	for _, tt := range tests {
		offset := SeekBlock(tt.blockNum, tt.blockSize)
		if offset != tt.expected {
			t.Errorf("SeekBlock(%d, %d) = %d; want %d", tt.blockNum, tt.blockSize, offset, tt.expected)
		}
	}
}

func TestIsAligned(t *testing.T) {
	tests := []struct {
		offset    int64
		blockSize uint32
		expected  bool
	}{
		{0, 4096, true},
		{4096, 4096, true},
		{8192, 4096, true},
		{4100, 4096, false},
		{12345, 512, false},
	}
	for _, tt := range tests {
		result := IsAligned(tt.offset, tt.blockSize)
		if result != tt.expected {
			t.Errorf("IsAligned(%d, %d) = %v; want %v", tt.offset, tt.blockSize, result, tt.expected)
		}
	}
}
