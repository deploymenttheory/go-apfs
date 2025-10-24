package sealed_volumes

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// fileInfoReader parses file info records
type fileInfoReader struct {
	key    *types.JFileInfoKeyT
	value  *types.JFileInfoValT
	endian binary.ByteOrder
}

// NewFileInfoReader creates a new reader for file info records
func NewFileInfoReader(keyData, valueData []byte, endian binary.ByteOrder) (interfaces.FileIntegrityReader, error) {
	if len(keyData) < 16 { // JKeyT (8) + InfoAndLba (8) = 16 bytes
		return nil, fmt.Errorf("key data too small for file info: %d bytes", len(keyData))
	}

	if len(valueData) < 4 { // HashedLen (2) + HashSize (1) + minimum 1 byte for hash
		return nil, fmt.Errorf("value data too small for file info: %d bytes", len(valueData))
	}

	key, err := parseFileInfoKey(keyData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file info key: %w", err)
	}

	value, err := parseFileInfoValue(valueData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file info value: %w", err)
	}

	return &fileInfoReader{
		key:    key,
		value:  value,
		endian: endian,
	}, nil
}

// parseFileInfoKey parses the key portion
func parseFileInfoKey(data []byte, endian binary.ByteOrder) (*types.JFileInfoKeyT, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("insufficient data for file info key: need 16 bytes, got %d", len(data))
	}

	key := &types.JFileInfoKeyT{}
	key.Hdr.ObjIdAndType = endian.Uint64(data[0:8])
	key.InfoAndLba = endian.Uint64(data[8:16])

	return key, nil
}

// parseFileInfoValue parses the value portion
func parseFileInfoValue(data []byte, endian binary.ByteOrder) (*types.JFileInfoValT, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("insufficient data for file info value: need at least 4 bytes, got %d", len(data))
	}

	value := &types.JFileInfoValT{}

	// Parse HashedLen
	value.Dhash.HashedLen = endian.Uint16(data[0:2])

	// Parse HashSize
	value.Dhash.HashSize = data[2]

	// Validate hash size
	if int(value.Dhash.HashSize) > int(types.ApfsHashMaxSize) {
		return nil, fmt.Errorf("hash size exceeds maximum: %d > %d", value.Dhash.HashSize, types.ApfsHashMaxSize)
	}

	// Verify sufficient data for hash
	if len(data) < 3+int(value.Dhash.HashSize) {
		return nil, fmt.Errorf("insufficient data for hash: need %d bytes, got %d", 3+value.Dhash.HashSize, len(data))
	}

	// Copy hash data
	copy(value.Dhash.Hash[:value.Dhash.HashSize], data[3:3+value.Dhash.HashSize])

	return value, nil
}

// DataHash returns the hash of the file's data
func (fir *fileInfoReader) DataHash() []byte {
	return fir.value.Dhash.Hash[:fir.value.Dhash.HashSize]
}

// HashType returns the type of hash used
func (fir *fileInfoReader) HashType() types.ApfsHashTypeT {
	// The hash type is not stored in the file info record itself
	// It should be obtained from the integrity metadata
	return types.ApfsHashInvalid
}

// HashedLength returns the length of the data segment that was hashed (in blocks)
func (fir *fileInfoReader) HashedLength() uint16 {
	return fir.value.Dhash.HashedLen
}

// GetInfoType returns the type of file info from the key
func (fir *fileInfoReader) GetInfoType() types.JObjFileInfoType {
	return types.JObjFileInfoType((fir.key.InfoAndLba & types.JFileInfoTypeMask) >> types.JFileInfoTypeShift)
}

// GetLogicalBlockAddress returns the logical block address from the key
func (fir *fileInfoReader) GetLogicalBlockAddress() uint64 {
	return fir.key.InfoAndLba & types.JFileInfoLbaMask
}
