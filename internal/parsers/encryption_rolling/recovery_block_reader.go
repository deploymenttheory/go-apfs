package encryptionrolling

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// recoveryBlockReader implements the EncryptionRollingRecoveryBlockReader interface
type recoveryBlockReader struct {
	block *types.ErRecoveryBlockPhysT
}

// Ensure interface compliance
var _ interfaces.EncryptionRollingRecoveryBlockReader = (*recoveryBlockReader)(nil)

// NewRecoveryBlockReader creates a new EncryptionRollingRecoveryBlockReader from raw data
func NewRecoveryBlockReader(data []byte, endian binary.ByteOrder) (interfaces.EncryptionRollingRecoveryBlockReader, error) {
	if len(data) < 48 { // Minimum size check for header + basic fields (32 + 8 + 8)
		return nil, fmt.Errorf("data too small for recovery block: %d bytes", len(data))
	}

	block, err := parseRecoveryBlock(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse recovery block: %w", err)
	}

	return &recoveryBlockReader{
		block: block,
	}, nil
}

// parseRecoveryBlock parses raw bytes into ErRecoveryBlockPhysT
func parseRecoveryBlock(data []byte, endian binary.ByteOrder) (*types.ErRecoveryBlockPhysT, error) {
	if len(data) < 48 { // Object header (32) + offset (8) + next OID (8)
		return nil, fmt.Errorf("insufficient data for recovery block")
	}

	block := &types.ErRecoveryBlockPhysT{}
	offset := 0

	// Parse object header (32 bytes)
	copy(block.ErbO.OChecksum[:], data[offset:offset+8])
	offset += 8
	block.ErbO.OOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	block.ErbO.OXid = types.XidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	block.ErbO.OType = endian.Uint32(data[offset : offset+4])
	offset += 4
	block.ErbO.OSubtype = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse recovery block specific fields
	block.ErbOffset = endian.Uint64(data[offset : offset+8])
	offset += 8
	block.ErbNextOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8

	// Parse the remaining data as recovery data
	if offset < len(data) {
		dataSize := len(data) - offset
		block.ErbData = make([]byte, dataSize)
		copy(block.ErbData, data[offset:])
	}

	return block, nil
}

// Implementation of EncryptionRollingRecoveryBlockReader interface

func (r *recoveryBlockReader) Offset() uint64 {
	return r.block.ErbOffset
}

func (r *recoveryBlockReader) NextObjectID() types.OidT {
	return r.block.ErbNextOid
}

func (r *recoveryBlockReader) Data() []byte {
	if r.block.ErbData == nil {
		return nil
	}
	// Return a copy to prevent external modification
	result := make([]byte, len(r.block.ErbData))
	copy(result, r.block.ErbData)
	return result
}
