package encryptionrolling

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// generalBitmapReader implements the GeneralBitmapReader interface
type generalBitmapReader struct {
	bitmap *types.GbitmapPhysT
}

// generalBitmapBlockReader implements the GeneralBitmapBlockReader interface
type generalBitmapBlockReader struct {
	block *types.GbitmapBlockPhysT
}

// Ensure interface compliance
var _ interfaces.GeneralBitmapReader = (*generalBitmapReader)(nil)
var _ interfaces.GeneralBitmapBlockReader = (*generalBitmapBlockReader)(nil)

// NewGeneralBitmapReader creates a new GeneralBitmapReader from raw data
func NewGeneralBitmapReader(data []byte, endian binary.ByteOrder) (interfaces.GeneralBitmapReader, error) {
	if len(data) < 56 { // Object header (32) + tree OID (8) + bit count (8) + flags (8)
		return nil, fmt.Errorf("data too small for general bitmap: %d bytes", len(data))
	}

	bitmap, err := parseGeneralBitmap(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse general bitmap: %w", err)
	}

	return &generalBitmapReader{
		bitmap: bitmap,
	}, nil
}

// NewGeneralBitmapBlockReader creates a new GeneralBitmapBlockReader from raw data
func NewGeneralBitmapBlockReader(data []byte, endian binary.ByteOrder) (interfaces.GeneralBitmapBlockReader, error) {
	if len(data) < 40 { // Object header (32) + minimum bitmap data (8)
		return nil, fmt.Errorf("data too small for bitmap block: %d bytes", len(data))
	}

	block, err := parseGeneralBitmapBlock(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bitmap block: %w", err)
	}

	return &generalBitmapBlockReader{
		block: block,
	}, nil
}

// parseGeneralBitmap parses raw bytes into GbitmapPhysT
func parseGeneralBitmap(data []byte, endian binary.ByteOrder) (*types.GbitmapPhysT, error) {
	if len(data) < 56 {
		return nil, fmt.Errorf("insufficient data for general bitmap")
	}

	bitmap := &types.GbitmapPhysT{}
	offset := 0

	// Parse object header (32 bytes)
	copy(bitmap.BmO.OChecksum[:], data[offset:offset+8])
	offset += 8
	bitmap.BmO.OOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	bitmap.BmO.OXid = types.XidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	bitmap.BmO.OType = endian.Uint32(data[offset : offset+4])
	offset += 4
	bitmap.BmO.OSubtype = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse bitmap specific fields
	bitmap.BmTreeOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	bitmap.BmBitCount = endian.Uint64(data[offset : offset+8])
	offset += 8
	bitmap.BmFlags = endian.Uint64(data[offset : offset+8])

	return bitmap, nil
}

// parseGeneralBitmapBlock parses raw bytes into GbitmapBlockPhysT
func parseGeneralBitmapBlock(data []byte, endian binary.ByteOrder) (*types.GbitmapBlockPhysT, error) {
	if len(data) < 40 {
		return nil, fmt.Errorf("insufficient data for bitmap block")
	}

	block := &types.GbitmapBlockPhysT{}
	offset := 0

	// Parse object header (32 bytes)
	copy(block.BmbO.OChecksum[:], data[offset:offset+8])
	offset += 8
	block.BmbO.OOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	block.BmbO.OXid = types.XidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	block.BmbO.OType = endian.Uint32(data[offset : offset+4])
	offset += 4
	block.BmbO.OSubtype = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse bitmap field data
	remainingBytes := len(data) - offset
	if remainingBytes > 0 {
		// Each uint64 is 8 bytes
		fieldCount := remainingBytes / 8
		if remainingBytes%8 != 0 {
			fieldCount++ // Account for partial last uint64
		}

		block.BmbField = make([]uint64, fieldCount)
		for i := 0; i < fieldCount && offset+8 <= len(data); i++ {
			block.BmbField[i] = endian.Uint64(data[offset : offset+8])
			offset += 8
		}

		// Handle partial last uint64 if necessary
		if remainingBytes%8 != 0 && fieldCount > 0 {
			lastBytes := remainingBytes % 8
			lastData := make([]byte, 8)
			copy(lastData, data[len(data)-lastBytes:])
			block.BmbField[fieldCount-1] = endian.Uint64(lastData)
		}
	}

	return block, nil
}

// Implementation of GeneralBitmapReader interface

func (g *generalBitmapReader) TreeObjectID() types.OidT {
	return g.bitmap.BmTreeOid
}

func (g *generalBitmapReader) BitCount() uint64 {
	return g.bitmap.BmBitCount
}

func (g *generalBitmapReader) Flags() uint64 {
	return g.bitmap.BmFlags
}

// Implementation of GeneralBitmapBlockReader interface

func (g *generalBitmapBlockReader) BitmapField() []uint64 {
	if g.block.BmbField == nil {
		return nil
	}
	// Return a copy to prevent external modification
	result := make([]uint64, len(g.block.BmbField))
	copy(result, g.block.BmbField)
	return result
}

func (g *generalBitmapBlockReader) IsBitSet(bitIndex uint64) bool {
	if g.block.BmbField == nil {
		return false
	}

	fieldIndex := bitIndex / 64 // 64 bits per uint64
	if fieldIndex >= uint64(len(g.block.BmbField)) {
		return false
	}

	bitOffset := bitIndex % 64
	return (g.block.BmbField[fieldIndex] & (1 << bitOffset)) != 0
}

func (g *generalBitmapBlockReader) SetBit(bitIndex uint64) {
	if g.block.BmbField == nil {
		return
	}

	fieldIndex := bitIndex / 64
	if fieldIndex >= uint64(len(g.block.BmbField)) {
		return
	}

	bitOffset := bitIndex % 64
	g.block.BmbField[fieldIndex] |= (1 << bitOffset)
}

func (g *generalBitmapBlockReader) ClearBit(bitIndex uint64) {
	if g.block.BmbField == nil {
		return
	}

	fieldIndex := bitIndex / 64
	if fieldIndex >= uint64(len(g.block.BmbField)) {
		return
	}

	bitOffset := bitIndex % 64
	g.block.BmbField[fieldIndex] &^= (1 << bitOffset)
}
