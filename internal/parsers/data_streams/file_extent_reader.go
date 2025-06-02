package datastreams

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// fileExtentReader implements the FileExtentReader interface
type fileExtentReader struct {
	key    *types.JFileExtentKeyT
	value  *types.JFileExtentValT
	data   []byte
	endian binary.ByteOrder
}

// NewFileExtentReader creates a new FileExtentReader implementation
func NewFileExtentReader(keyData, valueData []byte, endian binary.ByteOrder) (interfaces.FileExtentReader, error) {
	if len(keyData) < 16 { // JKeyT (8 bytes) + LogicalAddr (8 bytes)
		return nil, fmt.Errorf("key data too small for file extent key: %d bytes", len(keyData))
	}

	if len(valueData) < 24 { // JFileExtentValT is 24 bytes (8 + 8 + 8)
		return nil, fmt.Errorf("value data too small for file extent value: %d bytes", len(valueData))
	}

	key, err := parseFileExtentKey(keyData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file extent key: %w", err)
	}

	value, err := parseFileExtentValue(valueData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file extent value: %w", err)
	}

	return &fileExtentReader{
		key:    key,
		value:  value,
		data:   append(keyData, valueData...),
		endian: endian,
	}, nil
}

// parseFileExtentKey parses raw bytes into a JFileExtentKeyT structure
func parseFileExtentKey(data []byte, endian binary.ByteOrder) (*types.JFileExtentKeyT, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("insufficient data for file extent key")
	}

	key := &types.JFileExtentKeyT{}
	key.Hdr.ObjIdAndType = endian.Uint64(data[0:8])
	key.LogicalAddr = endian.Uint64(data[8:16])

	return key, nil
}

// parseFileExtentValue parses raw bytes into a JFileExtentValT structure
func parseFileExtentValue(data []byte, endian binary.ByteOrder) (*types.JFileExtentValT, error) {
	if len(data) < 24 {
		return nil, fmt.Errorf("insufficient data for file extent value")
	}

	value := &types.JFileExtentValT{}
	value.LenAndFlags = endian.Uint64(data[0:8])
	value.PhysBlockNum = endian.Uint64(data[8:16])
	value.CryptoId = endian.Uint64(data[16:24])

	return value, nil
}

// Length returns the length of the extent in bytes
func (fer *fileExtentReader) Length() uint64 {
	// The extent's length is accessed as len_and_flags & J_FILE_EXTENT_LEN_MASK
	return fer.value.LenAndFlags & types.JFileExtentLenMask
}

// Flags returns the file extent flags
func (fer *fileExtentReader) Flags() uint64 {
	// The extent's flags are accessed as (len_and_flags & J_FILE_EXTENT_FLAG_MASK) >> J_FILE_EXTENT_FLAG_SHIFT
	return (fer.value.LenAndFlags & types.JFileExtentFlagMask) >> types.JFileExtentFlagShift
}

// PhysicalBlockNumber returns the physical block address that the extent starts at
func (fer *fileExtentReader) PhysicalBlockNumber() uint64 {
	return fer.value.PhysBlockNum
}

// CryptoID returns the encryption key or encryption tweak used in this extent
func (fer *fileExtentReader) CryptoID() uint64 {
	return fer.value.CryptoId
}

// LogicalAddress returns the offset within the file's data where the data is stored
func (fer *fileExtentReader) LogicalAddress() uint64 {
	return fer.key.LogicalAddr
}

// IsCryptoIDTweak checks if the crypto_id field contains an encryption tweak value
func (fer *fileExtentReader) IsCryptoIDTweak() bool {
	// Check if the FEXT_CRYPTO_ID_IS_TWEAK flag is set
	return fer.Flags()&uint64(types.FextCryptoIdIsTweak) != 0
}
