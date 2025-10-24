package sealed_volumes

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// fextTreeKeyReader parses file extent tree keys
type fextTreeKeyReader struct {
	key    *types.FextTreeKeyT
	endian binary.ByteOrder
}

// fextTreeValReader parses file extent tree values
type fextTreeValReader struct {
	value  *types.FextTreeValT
	endian binary.ByteOrder
}

// NewFextTreeKeyReader creates a new reader for file extent tree keys
func NewFextTreeKeyReader(data []byte, endian binary.ByteOrder) (*fextTreeKeyReader, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("key data too small for fext tree key: %d bytes", len(data))
	}

	key, err := parseFextTreeKey(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse fext tree key: %w", err)
	}

	return &fextTreeKeyReader{
		key:    key,
		endian: endian,
	}, nil
}

// NewFextTreeValReader creates a new reader for file extent tree values
func NewFextTreeValReader(data []byte, endian binary.ByteOrder) (*fextTreeValReader, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("value data too small for fext tree value: %d bytes", len(data))
	}

	value, err := parseFextTreeVal(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse fext tree value: %w", err)
	}

	return &fextTreeValReader{
		value:  value,
		endian: endian,
	}, nil
}

// parseFextTreeKey parses file extent tree key
func parseFextTreeKey(data []byte, endian binary.ByteOrder) (*types.FextTreeKeyT, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("insufficient data for fext tree key: need 16 bytes, got %d", len(data))
	}

	key := &types.FextTreeKeyT{}
	key.PrivateId = endian.Uint64(data[0:8])
	key.LogicalAddr = endian.Uint64(data[8:16])

	return key, nil
}

// parseFextTreeVal parses file extent tree value
func parseFextTreeVal(data []byte, endian binary.ByteOrder) (*types.FextTreeValT, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("insufficient data for fext tree value: need 16 bytes, got %d", len(data))
	}

	value := &types.FextTreeValT{}
	value.LenAndFlags = endian.Uint64(data[0:8])
	value.PhysBlockNum = endian.Uint64(data[8:16])

	return value, nil
}

// FileID returns the object identifier of the file
func (ftr *fextTreeKeyReader) FileID() uint64 {
	return ftr.key.PrivateId
}

// LogicalAddress returns the logical address within the file
func (ftr *fextTreeKeyReader) LogicalAddress() uint64 {
	return ftr.key.LogicalAddr
}

// Length returns the length of the extent in bytes
func (ftvr *fextTreeValReader) Length() uint64 {
	// Use the same mask as file extent records
	return ftvr.value.LenAndFlags & types.JFileExtentLenMask
}

// Flags returns the flags for this extent
func (ftvr *fextTreeValReader) Flags() uint64 {
	// Use the same mask and shift as file extent records
	return (ftvr.value.LenAndFlags & types.JFileExtentFlagMask) >> types.JFileExtentFlagShift
}

// PhysicalBlockNumber returns the physical block address
func (ftvr *fextTreeValReader) PhysicalBlockNumber() uint64 {
	return ftvr.value.PhysBlockNum
}
