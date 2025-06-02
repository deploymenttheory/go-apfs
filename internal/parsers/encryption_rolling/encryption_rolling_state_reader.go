package encryptionrolling

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// encryptionRollingStateReader implements the EncryptionRollingStateReader interface
type encryptionRollingStateReader struct {
	state *types.ErStatePhysT
}

// encryptionRollingV1StateReader implements the EncryptionRollingV1StateReader interface
type encryptionRollingV1StateReader struct {
	state *types.ErStatePhysV1T
}

// Ensure interface compliance
var _ interfaces.EncryptionRollingStateReader = (*encryptionRollingStateReader)(nil)
var _ interfaces.EncryptionRollingV1StateReader = (*encryptionRollingV1StateReader)(nil)

// NewEncryptionRollingStateReader creates a new EncryptionRollingStateReader from raw data
func NewEncryptionRollingStateReader(data []byte, endian binary.ByteOrder) (interfaces.EncryptionRollingStateReader, error) {
	if len(data) < 128 { // Minimum size check
		return nil, fmt.Errorf("data too small for encryption rolling state: %d bytes", len(data))
	}

	state, err := parseEncryptionRollingState(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encryption rolling state: %w", err)
	}

	return &encryptionRollingStateReader{
		state: state,
	}, nil
}

// NewEncryptionRollingV1StateReader creates a new EncryptionRollingV1StateReader from raw data
func NewEncryptionRollingV1StateReader(data []byte, endian binary.ByteOrder) (interfaces.EncryptionRollingV1StateReader, error) {
	if len(data) < 128 { // Minimum size check
		return nil, fmt.Errorf("data too small for encryption rolling v1 state: %d bytes", len(data))
	}

	state, err := parseEncryptionRollingV1State(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encryption rolling v1 state: %w", err)
	}

	return &encryptionRollingV1StateReader{
		state: state,
	}, nil
}

// parseEncryptionRollingState parses raw bytes into ErStatePhysT
func parseEncryptionRollingState(data []byte, endian binary.ByteOrder) (*types.ErStatePhysT, error) {
	if len(data) < 136 { // Changed from 128 to 136 to match actual field requirements
		return nil, fmt.Errorf("insufficient data for encryption rolling state")
	}

	state := &types.ErStatePhysT{}
	offset := 0

	// Parse header (48 bytes total)
	// Object header (32 bytes)
	copy(state.ErsbHeader.ErsbO.OChecksum[:], data[offset:offset+8])
	offset += 8
	state.ErsbHeader.ErsbO.OOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	state.ErsbHeader.ErsbO.OXid = types.XidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	state.ErsbHeader.ErsbO.OType = endian.Uint32(data[offset : offset+4])
	offset += 4
	state.ErsbHeader.ErsbO.OSubtype = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Header specific fields (16 bytes)
	state.ErsbHeader.ErsbMagic = endian.Uint32(data[offset : offset+4])
	offset += 4
	state.ErsbHeader.ErsbVersion = endian.Uint32(data[offset : offset+4])
	offset += 4
	offset += 8 // Reserved/padding

	// Parse state fields
	state.ErsbFlags = endian.Uint64(data[offset : offset+8])
	offset += 8
	state.ErsbSnapXid = endian.Uint64(data[offset : offset+8])
	offset += 8
	state.ErsbCurrentFextObjId = endian.Uint64(data[offset : offset+8])
	offset += 8
	state.ErsbFileOffset = endian.Uint64(data[offset : offset+8])
	offset += 8
	state.ErsbProgress = endian.Uint64(data[offset : offset+8])
	offset += 8
	state.ErsbTotalBlkToEncrypt = endian.Uint64(data[offset : offset+8])
	offset += 8
	state.ErsbBlockmapOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	state.ErsbTidemarkObjId = endian.Uint64(data[offset : offset+8])
	offset += 8
	state.ErsbRecoveryExtentsCount = endian.Uint64(data[offset : offset+8])
	offset += 8
	state.ErsbRecoveryListOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	state.ErsbRecoveryLength = endian.Uint64(data[offset : offset+8])

	return state, nil
}

// parseEncryptionRollingV1State parses raw bytes into ErStatePhysV1T
func parseEncryptionRollingV1State(data []byte, endian binary.ByteOrder) (*types.ErStatePhysV1T, error) {
	if len(data) < 136 { // Changed from 128 to 136 to include all required fields
		return nil, fmt.Errorf("insufficient data for encryption rolling v1 state")
	}

	state := &types.ErStatePhysV1T{}
	offset := 0

	// Parse header (48 bytes total)
	// Object header (32 bytes)
	copy(state.ErsbHeader.ErsbO.OChecksum[:], data[offset:offset+8])
	offset += 8
	state.ErsbHeader.ErsbO.OOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	state.ErsbHeader.ErsbO.OXid = types.XidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	state.ErsbHeader.ErsbO.OType = endian.Uint32(data[offset : offset+4])
	offset += 4
	state.ErsbHeader.ErsbO.OSubtype = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Header specific fields (16 bytes)
	state.ErsbHeader.ErsbMagic = endian.Uint32(data[offset : offset+4])
	offset += 4
	state.ErsbHeader.ErsbVersion = endian.Uint32(data[offset : offset+4])
	offset += 4
	offset += 8 // Reserved/padding

	// Parse v1 specific fields
	state.ErsbFlags = endian.Uint64(data[offset : offset+8])
	offset += 8
	state.ErsbSnapXid = endian.Uint64(data[offset : offset+8])
	offset += 8
	state.ErsbCurrentFextObjId = endian.Uint64(data[offset : offset+8])
	offset += 8
	state.ErsbFileOffset = endian.Uint64(data[offset : offset+8])
	offset += 8
	state.ErsbFextPbn = endian.Uint64(data[offset : offset+8])
	offset += 8
	state.ErsbPaddr = endian.Uint64(data[offset : offset+8])
	offset += 8
	state.ErsbProgress = endian.Uint64(data[offset : offset+8])
	offset += 8
	state.ErsbTotalBlkToEncrypt = endian.Uint64(data[offset : offset+8])
	offset += 8
	state.ErsbBlockmapOid = endian.Uint64(data[offset : offset+8])
	offset += 8
	state.ErsbChecksumCount = endian.Uint32(data[offset : offset+4])
	offset += 4
	state.ErsbReserved = endian.Uint32(data[offset : offset+4])
	offset += 4
	state.ErsbFextCid = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse checksum data if present
	if int(state.ErsbChecksumCount) > 0 && offset < len(data) {
		checksumSize := int(state.ErsbChecksumCount * 8) // 8 bytes per checksum
		if offset+checksumSize <= len(data) {
			state.ErsbChecksum = make([]byte, checksumSize)
			copy(state.ErsbChecksum, data[offset:offset+checksumSize])
		}
	}

	return state, nil
}

// Implementation of EncryptionRollingStateReader interface

func (e *encryptionRollingStateReader) Version() uint32 {
	return e.state.ErsbHeader.ErsbVersion
}

func (e *encryptionRollingStateReader) Magic() uint32 {
	return e.state.ErsbHeader.ErsbMagic
}

func (e *encryptionRollingStateReader) Flags() uint64 {
	return e.state.ErsbFlags
}

func (e *encryptionRollingStateReader) SnapshotXID() types.XidT {
	return types.XidT(e.state.ErsbSnapXid)
}

func (e *encryptionRollingStateReader) CurrentFileExtentObjectID() uint64 {
	return e.state.ErsbCurrentFextObjId
}

func (e *encryptionRollingStateReader) FileOffset() uint64 {
	return e.state.ErsbFileOffset
}

func (e *encryptionRollingStateReader) Progress() uint64 {
	return e.state.ErsbProgress
}

func (e *encryptionRollingStateReader) TotalBlocksToEncrypt() uint64 {
	return e.state.ErsbTotalBlkToEncrypt
}

func (e *encryptionRollingStateReader) BlockmapOID() types.OidT {
	return e.state.ErsbBlockmapOid
}

func (e *encryptionRollingStateReader) TidemarkObjectID() uint64 {
	return e.state.ErsbTidemarkObjId
}

func (e *encryptionRollingStateReader) RecoveryExtentsCount() uint64 {
	return e.state.ErsbRecoveryExtentsCount
}

func (e *encryptionRollingStateReader) RecoveryListOID() types.OidT {
	return e.state.ErsbRecoveryListOid
}

func (e *encryptionRollingStateReader) RecoveryLength() uint64 {
	return e.state.ErsbRecoveryLength
}

// Implementation of EncryptionRollingV1StateReader interface

func (e *encryptionRollingV1StateReader) Version() uint32 {
	return e.state.ErsbHeader.ErsbVersion
}

func (e *encryptionRollingV1StateReader) Magic() uint32 {
	return e.state.ErsbHeader.ErsbMagic
}

func (e *encryptionRollingV1StateReader) Flags() uint64 {
	return e.state.ErsbFlags
}

func (e *encryptionRollingV1StateReader) SnapshotXID() types.XidT {
	return types.XidT(e.state.ErsbSnapXid)
}

func (e *encryptionRollingV1StateReader) CurrentFileExtentObjectID() uint64 {
	return e.state.ErsbCurrentFextObjId
}

func (e *encryptionRollingV1StateReader) FileOffset() uint64 {
	return e.state.ErsbFileOffset
}

func (e *encryptionRollingV1StateReader) FileExtentPhysicalBlockNumber() uint64 {
	return e.state.ErsbFextPbn
}

func (e *encryptionRollingV1StateReader) PhysicalAddress() uint64 {
	return e.state.ErsbPaddr
}

func (e *encryptionRollingV1StateReader) Progress() uint64 {
	return e.state.ErsbProgress
}

func (e *encryptionRollingV1StateReader) TotalBlocksToEncrypt() uint64 {
	return e.state.ErsbTotalBlkToEncrypt
}

func (e *encryptionRollingV1StateReader) BlockmapOID() uint64 {
	return e.state.ErsbBlockmapOid
}

func (e *encryptionRollingV1StateReader) ChecksumCount() uint32 {
	return e.state.ErsbChecksumCount
}

func (e *encryptionRollingV1StateReader) FileExtentCryptoID() uint64 {
	return e.state.ErsbFextCid
}

func (e *encryptionRollingV1StateReader) Checksums() []byte {
	return e.state.ErsbChecksum
}
