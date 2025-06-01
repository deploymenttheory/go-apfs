package encryptionrolling

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// generalBitmapBlockReader implements the GeneralBitmapBlockReader interface
type generalBitmapBlockReader struct {
	bitmapBlock *types.GbitmapBlockPhysT
	endian      binary.ByteOrder
}

// Ensure implementation matches interface
var _ interfaces.GeneralBitmapBlockReader = (*generalBitmapBlockReader)(nil)

// NewGeneralBitmapBlockReader creates a new GeneralBitmapBlockReader from raw data
func NewGeneralBitmapBlockReader(data []byte, endian binary.ByteOrder) (interfaces.GeneralBitmapBlockReader, error) {
	if endian == nil {
		endian = binary.LittleEndian
	}

	bitmapBlock, err := parseGeneralBitmapBlock(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse general bitmap block: %w", err)
	}

	return &generalBitmapBlockReader{
		bitmapBlock: bitmapBlock,
		endian:      endian,
	}, nil
}

// parseGeneralBitmapBlock parses raw bytes into a GbitmapBlockPhysT structure
func parseGeneralBitmapBlock(data []byte, endian binary.ByteOrder) (*types.GbitmapBlockPhysT, error) {
	// Minimum size: ObjPhysT(32) bytes
	if len(data) < 32 {
		return nil, fmt.Errorf("insufficient data for general bitmap block: need at least 32 bytes, got %d", len(data))
	}

	bitmapBlock := &types.GbitmapBlockPhysT{}
	offset := 0

	// Parse object header (32 bytes)
	copy(bitmapBlock.BmbO.OChecksum[:], data[offset:offset+8])
	offset += 8
	bitmapBlock.BmbO.OOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	bitmapBlock.BmbO.OXid = types.XidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	bitmapBlock.BmbO.OType = endian.Uint32(data[offset : offset+4])
	offset += 4
	bitmapBlock.BmbO.OSubtype = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse bitmap field data (remaining bytes as uint64 array)
	remainingBytes := len(data) - offset
	if remainingBytes%8 != 0 {
		return nil, fmt.Errorf("bitmap field data size must be multiple of 8 bytes, got %d", remainingBytes)
	}

	fieldCount := remainingBytes / 8
	bitmapBlock.BmbField = make([]uint64, fieldCount)
	for i := 0; i < fieldCount; i++ {
		bitmapBlock.BmbField[i] = endian.Uint64(data[offset : offset+8])
		offset += 8
	}

	return bitmapBlock, nil
}

// BitmapField returns the bitmap data as an array of uint64
func (gbbr *generalBitmapBlockReader) BitmapField() []uint64 {
	result := make([]uint64, len(gbbr.bitmapBlock.BmbField))
	copy(result, gbbr.bitmapBlock.BmbField)
	return result
}

// IsBitSet checks if a specific bit is set in the bitmap
func (gbbr *generalBitmapBlockReader) IsBitSet(bitIndex uint64) bool {
	wordIndex := bitIndex / 64
	bitOffset := bitIndex % 64

	if wordIndex >= uint64(len(gbbr.bitmapBlock.BmbField)) {
		return false
	}

	return (gbbr.bitmapBlock.BmbField[wordIndex] & (1 << bitOffset)) != 0
}

// SetBit sets a specific bit in the bitmap
func (gbbr *generalBitmapBlockReader) SetBit(bitIndex uint64) {
	wordIndex := bitIndex / 64
	bitOffset := bitIndex % 64

	if wordIndex >= uint64(len(gbbr.bitmapBlock.BmbField)) {
		return
	}

	gbbr.bitmapBlock.BmbField[wordIndex] |= (1 << bitOffset)
}

// ClearBit clears a specific bit in the bitmap
func (gbbr *generalBitmapBlockReader) ClearBit(bitIndex uint64) {
	wordIndex := bitIndex / 64
	bitOffset := bitIndex % 64

	if wordIndex >= uint64(len(gbbr.bitmapBlock.BmbField)) {
		return
	}

	gbbr.bitmapBlock.BmbField[wordIndex] &^= (1 << bitOffset)
}
