// File: pkg/util/io.go
package util

import (
	"encoding/binary"

	"fmt"
	"io"
	"os"

	"github.com/deploymenttheory/go-apfs/pkg/types"
)

// FileDevice implements BlockDevice backed by a file (e.g., .dmg or raw image)
type FileDevice struct {
	f          *os.File
	blockSize  uint32
	blockCount uint64
}

// OpenFileDevice opens a file and wraps it as a BlockDevice with the given block size
func OpenFileDevice(path string, blockSize uint32) (*FileDevice, error) {
	f, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		return nil, types.NewAPFSError(types.ErrIOError, "OpenFileDevice", path, err.Error())
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, types.NewAPFSError(types.ErrIOError, "OpenFileDevice", path, err.Error())
	}

	size := stat.Size()
	if size < int64(blockSize) {
		return nil, types.NewAPFSError(types.ErrInvalidBlockSize, "OpenFileDevice", path, fmt.Sprintf("file size (%d) smaller than block size (%d)", size, blockSize))
	}

	blockCount := uint64(size) / uint64(blockSize)
	return &FileDevice{f: f, blockSize: blockSize, blockCount: blockCount}, nil
}

// ReadBlock reads a single block at a given physical address
func (d *FileDevice) ReadBlock(addr types.PAddr) ([]byte, error) {
	offset := int64(addr) * int64(d.blockSize)
	buf := make([]byte, d.blockSize)
	_, err := d.f.ReadAt(buf, offset)
	if err != nil && err != io.EOF {
		return nil, types.NewAPFSError(types.ErrIOError, "ReadBlock", fmt.Sprintf("PAddr(%d)", addr), err.Error())
	}
	return buf, nil
}

// WriteBlock writes a single block to a given physical address
func (d *FileDevice) WriteBlock(addr types.PAddr, data []byte) error {
	if len(data) != int(d.blockSize) {
		return types.NewAPFSError(types.ErrInvalidBlockSize, "WriteBlock", fmt.Sprintf("PAddr(%d)", addr), fmt.Sprintf("data size %d does not match block size %d", len(data), d.blockSize))
	}
	offset := int64(addr) * int64(d.blockSize)
	_, err := d.f.WriteAt(data, offset)
	if err != nil {
		return types.NewAPFSError(types.ErrIOError, "WriteBlock", fmt.Sprintf("PAddr(%d)", addr), err.Error())
	}
	return nil
}

// ReadAt reads raw data at a given byte offset
func (d *FileDevice) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 || off >= int64(d.blockCount)*int64(d.blockSize) {
		return 0, types.NewAPFSError(types.ErrInvalidBlockAddr, "ReadAt", fmt.Sprintf("offset=%d", off), "read offset out of range")
	}
	n, err := d.f.ReadAt(p, off)
	if err != nil && err != io.EOF {
		return n, types.NewAPFSError(types.ErrIOError, "ReadAt", fmt.Sprintf("offset=%d", off), err.Error())
	}
	return n, nil
}

// GetBlockSize returns the block size used by the device
func (d *FileDevice) GetBlockSize() uint32 {
	return d.blockSize
}

// GetBlockCount returns the total number of blocks available in the device
func (d *FileDevice) GetBlockCount() uint64 {
	return d.blockCount
}

// Close closes the underlying file
func (d *FileDevice) Close() error {
	return d.f.Close()
}

// SeekBlock returns the byte offset for a given block number and block size
func SeekBlock(blockNum uint64, blockSize uint32) int64 {
	return int64(blockNum) * int64(blockSize)
}

// IsAligned checks whether a given offset is aligned to blockSize
func IsAligned(offset int64, blockSize uint32) bool {
	return offset%int64(blockSize) == 0
}

// ------------------ Endian Helpers ------------------

// ReadUint16LE reads a little-endian uint16 from 2 bytes
func ReadUint16LE(b []byte) uint16 {
	return binary.LittleEndian.Uint16(b)
}

// ReadUint32LE reads a little-endian uint32 from 4 bytes
func ReadUint32LE(b []byte) uint32 {
	return binary.LittleEndian.Uint32(b)
}

// ReadUint64LE reads a little-endian uint64 from 8 bytes
func ReadUint64LE(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}

// ReadBytes copies a slice from a fixed offset and length safely
func ReadBytes(b []byte, offset, length int) ([]byte, error) {
	if offset+length > len(b) {
		return nil, types.NewAPFSError(types.ErrInvalidArgument, "ReadBytes", "buffer", fmt.Sprintf("offset=%d length=%d size=%d", offset, length, len(b)))
	}
	return b[offset : offset+length], nil
}
