package encryptionrolling

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// encryptionRollingRecoveryBlockReader implements the EncryptionRollingRecoveryBlockReader interface
type encryptionRollingRecoveryBlockReader struct {
	recoveryBlock *types.ErRecoveryBlockPhysT
	endian        binary.ByteOrder
}

// Ensure implementation matches interface
var _ interfaces.EncryptionRollingRecoveryBlockReader = (*encryptionRollingRecoveryBlockReader)(nil)

// NewEncryptionRollingRecoveryBlockReader creates a new EncryptionRollingRecoveryBlockReader from raw data
func NewEncryptionRollingRecoveryBlockReader(data []byte, endian binary.ByteOrder) (interfaces.EncryptionRollingRecoveryBlockReader, error) {
	if endian == nil {
		endian = binary.LittleEndian
	}

	recoveryBlock, err := parseEncryptionRollingRecoveryBlock(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encryption rolling recovery block: %w", err)
	}

	return &encryptionRollingRecoveryBlockReader{
		recoveryBlock: recoveryBlock,
		endian:        endian,
	}, nil
}

// parseEncryptionRollingRecoveryBlock parses raw bytes into an ErRecoveryBlockPhysT structure
func parseEncryptionRollingRecoveryBlock(data []byte, endian binary.ByteOrder) (*types.ErRecoveryBlockPhysT, error) {
	// Minimum size: ObjPhysT(32) + offset(8) + next_oid(8) = 48 bytes
	if len(data) < 48 {
		return nil, fmt.Errorf("insufficient data for recovery block: need at least 48 bytes, got %d", len(data))
	}

	recoveryBlock := &types.ErRecoveryBlockPhysT{}
	offset := 0

	// Parse object header (32 bytes)
	copy(recoveryBlock.ErbO.OChecksum[:], data[offset:offset+8])
	offset += 8
	recoveryBlock.ErbO.OOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	recoveryBlock.ErbO.OXid = types.XidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	recoveryBlock.ErbO.OType = endian.Uint32(data[offset : offset+4])
	offset += 4
	recoveryBlock.ErbO.OSubtype = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse recovery block fields
	recoveryBlock.ErbOffset = endian.Uint64(data[offset : offset+8])
	offset += 8
	recoveryBlock.ErbNextOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8

	// Parse data (remaining bytes)
	if len(data) > offset {
		recoveryBlock.ErbData = make([]byte, len(data)-offset)
		copy(recoveryBlock.ErbData, data[offset:])
	}

	return recoveryBlock, nil
}

// Offset returns the offset of the recovery block
func (errbr *encryptionRollingRecoveryBlockReader) Offset() uint64 {
	return errbr.recoveryBlock.ErbOffset
}

// NextObjectID returns the object identifier of the next recovery block
func (errbr *encryptionRollingRecoveryBlockReader) NextObjectID() types.OidT {
	return errbr.recoveryBlock.ErbNextOid
}

// Data returns the data in the recovery block
func (errbr *encryptionRollingRecoveryBlockReader) Data() []byte {
	result := make([]byte, len(errbr.recoveryBlock.ErbData))
	copy(result, errbr.recoveryBlock.ErbData)
	return result
}

// generalBitmapReader implements the GeneralBitmapReader interface
type generalBitmapReader struct {
	bitmap *types.GbitmapPhysT
	endian binary.ByteOrder
}

// Ensure implementation matches interface
var _ interfaces.GeneralBitmapReader = (*generalBitmapReader)(nil)

// NewGeneralBitmapReader creates a new GeneralBitmapReader from raw data
func NewGeneralBitmapReader(data []byte, endian binary.ByteOrder) (interfaces.GeneralBitmapReader, error) {
	if endian == nil {
		endian = binary.LittleEndian
	}

	bitmap, err := parseGeneralBitmap(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse general bitmap: %w", err)
	}

	return &generalBitmapReader{
		bitmap: bitmap,
		endian: endian,
	}, nil
}

// parseGeneralBitmap parses raw bytes into a GbitmapPhysT structure
func parseGeneralBitmap(data []byte, endian binary.ByteOrder) (*types.GbitmapPhysT, error) {
	// Minimum size: ObjPhysT(32) + tree_oid(8) + bit_count(8) + flags(8) = 56 bytes
	if len(data) < 56 {
		return nil, fmt.Errorf("insufficient data for general bitmap: need at least 56 bytes, got %d", len(data))
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

	// Parse bitmap fields
	bitmap.BmTreeOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	bitmap.BmBitCount = endian.Uint64(data[offset : offset+8])
	offset += 8
	bitmap.BmFlags = endian.Uint64(data[offset : offset+8])
	offset += 8

	return bitmap, nil
}

// TreeObjectID returns the object identifier of the bitmap tree
func (gbr *generalBitmapReader) TreeObjectID() types.OidT {
	return gbr.bitmap.BmTreeOid
}

// BitCount returns the number of bits in the bitmap
func (gbr *generalBitmapReader) BitCount() uint64 {
	return gbr.bitmap.BmBitCount
}

// Flags returns the flags for the bitmap
func (gbr *generalBitmapReader) Flags() uint64 {
	return gbr.bitmap.BmFlags
}
