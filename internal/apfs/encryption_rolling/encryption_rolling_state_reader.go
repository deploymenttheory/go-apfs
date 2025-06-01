package encryptionrolling

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// encryptionRollingStateReader implements the EncryptionRollingStateReader interface
type encryptionRollingStateReader struct {
	state  *types.ErStatePhysT
	endian binary.ByteOrder
}

// encryptionRollingV1StateReader implements the EncryptionRollingV1StateReader interface
type encryptionRollingV1StateReader struct {
	state  *types.ErStatePhysV1T
	endian binary.ByteOrder
}

// Ensure implementations match interfaces
var _ interfaces.EncryptionRollingStateReader = (*encryptionRollingStateReader)(nil)
var _ interfaces.EncryptionRollingV1StateReader = (*encryptionRollingV1StateReader)(nil)

// NewEncryptionRollingStateReader creates a new EncryptionRollingStateReader from raw data
func NewEncryptionRollingStateReader(data []byte, endian binary.ByteOrder) (interfaces.EncryptionRollingStateReader, error) {
	if endian == nil {
		endian = binary.LittleEndian
	}

	state, err := parseEncryptionRollingState(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encryption rolling state: %w", err)
	}

	return &encryptionRollingStateReader{
		state:  state,
		endian: endian,
	}, nil
}

// NewEncryptionRollingV1StateReader creates a new EncryptionRollingV1StateReader from raw data
func NewEncryptionRollingV1StateReader(data []byte, endian binary.ByteOrder) (interfaces.EncryptionRollingV1StateReader, error) {
	if endian == nil {
		endian = binary.LittleEndian
	}

	state, err := parseEncryptionRollingV1State(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encryption rolling v1 state: %w", err)
	}

	return &encryptionRollingV1StateReader{
		state:  state,
		endian: endian,
	}, nil
}

// parseEncryptionRollingState parses raw bytes into an ErStatePhysT structure
func parseEncryptionRollingState(data []byte, endian binary.ByteOrder) (*types.ErStatePhysT, error) {
	// Minimum size: header(40) + base fields(80) = 120 bytes
	if len(data) < 120 {
		return nil, fmt.Errorf("insufficient data for encryption rolling state: need at least 120 bytes, got %d", len(data))
	}

	state := &types.ErStatePhysT{}
	offset := 0

	// Parse header - ObjPhysT (32 bytes) + magic(4) + version(4) = 40 bytes
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
	state.ErsbHeader.ErsbMagic = endian.Uint32(data[offset : offset+4])
	offset += 4
	state.ErsbHeader.ErsbVersion = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Validate magic number
	if state.ErsbHeader.ErsbMagic != types.ErMagic {
		return nil, fmt.Errorf("invalid magic number: got 0x%08X, expected 0x%08X", state.ErsbHeader.ErsbMagic, types.ErMagic)
	}

	// Parse state fields (80 bytes)
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

	return state, nil
}

// parseEncryptionRollingV1State parses raw bytes into an ErStatePhysV1T structure
func parseEncryptionRollingV1State(data []byte, endian binary.ByteOrder) (*types.ErStatePhysV1T, error) {
	// Minimum size: header(40) + base fields(88) + optional checksum data
	if len(data) < 128 {
		return nil, fmt.Errorf("insufficient data for encryption rolling v1 state: need at least 128 bytes, got %d", len(data))
	}

	state := &types.ErStatePhysV1T{}
	offset := 0

	// Parse header (same as regular state)
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
	state.ErsbHeader.ErsbMagic = endian.Uint32(data[offset : offset+4])
	offset += 4
	state.ErsbHeader.ErsbVersion = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Validate magic number
	if state.ErsbHeader.ErsbMagic != types.ErMagic {
		return nil, fmt.Errorf("invalid magic number: got 0x%08X, expected 0x%08X", state.ErsbHeader.ErsbMagic, types.ErMagic)
	}

	// Parse v1-specific fields
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

	// Parse checksum data if available
	checksumDataSize := int(state.ErsbChecksumCount * types.ErChecksumLength)
	if len(data) >= offset+checksumDataSize {
		state.ErsbChecksum = make([]byte, checksumDataSize)
		copy(state.ErsbChecksum, data[offset:offset+checksumDataSize])
	}

	return state, nil
}

// EncryptionRollingStateReader interface implementation

// Version returns the version number of the encryption rolling state
func (ersr *encryptionRollingStateReader) Version() uint32 {
	return ersr.state.ErsbHeader.ErsbVersion
}

// Magic returns the magic number for validating the encryption rolling state
func (ersr *encryptionRollingStateReader) Magic() uint32 {
	return ersr.state.ErsbHeader.ErsbMagic
}

// Flags returns the encryption rolling state flags
func (ersr *encryptionRollingStateReader) Flags() uint64 {
	return ersr.state.ErsbFlags
}

// SnapshotXID returns the snapshot transaction identifier
func (ersr *encryptionRollingStateReader) SnapshotXID() types.XidT {
	return types.XidT(ersr.state.ErsbSnapXid)
}

// CurrentFileExtentObjectID returns the current file extent object identifier
func (ersr *encryptionRollingStateReader) CurrentFileExtentObjectID() uint64 {
	return ersr.state.ErsbCurrentFextObjId
}

// FileOffset returns the file offset where encryption rolling is currently at
func (ersr *encryptionRollingStateReader) FileOffset() uint64 {
	return ersr.state.ErsbFileOffset
}

// Progress returns the current progress of encryption rolling
func (ersr *encryptionRollingStateReader) Progress() uint64 {
	return ersr.state.ErsbProgress
}

// TotalBlocksToEncrypt returns the total number of blocks to encrypt
func (ersr *encryptionRollingStateReader) TotalBlocksToEncrypt() uint64 {
	return ersr.state.ErsbTotalBlkToEncrypt
}

// BlockmapOID returns the object identifier of the block map
func (ersr *encryptionRollingStateReader) BlockmapOID() types.OidT {
	return ersr.state.ErsbBlockmapOid
}

// TidemarkObjectID returns the tidemark object identifier
func (ersr *encryptionRollingStateReader) TidemarkObjectID() uint64 {
	return ersr.state.ErsbTidemarkObjId
}

// RecoveryExtentsCount returns the count of recovery extents
func (ersr *encryptionRollingStateReader) RecoveryExtentsCount() uint64 {
	return ersr.state.ErsbRecoveryExtentsCount
}

// RecoveryListOID returns the object identifier of the recovery list
func (ersr *encryptionRollingStateReader) RecoveryListOID() types.OidT {
	return ersr.state.ErsbRecoveryListOid
}

// RecoveryLength returns the length of the recovery
func (ersr *encryptionRollingStateReader) RecoveryLength() uint64 {
	return ersr.state.ErsbRecoveryLength
}

// EncryptionRollingV1StateReader interface implementation

// Version returns the version number of the encryption rolling state
func (erv1sr *encryptionRollingV1StateReader) Version() uint32 {
	return erv1sr.state.ErsbHeader.ErsbVersion
}

// Magic returns the magic number for validating the encryption rolling state
func (erv1sr *encryptionRollingV1StateReader) Magic() uint32 {
	return erv1sr.state.ErsbHeader.ErsbMagic
}

// Flags returns the encryption rolling state flags
func (erv1sr *encryptionRollingV1StateReader) Flags() uint64 {
	return erv1sr.state.ErsbFlags
}

// SnapshotXID returns the snapshot transaction identifier
func (erv1sr *encryptionRollingV1StateReader) SnapshotXID() types.XidT {
	return types.XidT(erv1sr.state.ErsbSnapXid)
}

// CurrentFileExtentObjectID returns the current file extent object identifier
func (erv1sr *encryptionRollingV1StateReader) CurrentFileExtentObjectID() uint64 {
	return erv1sr.state.ErsbCurrentFextObjId
}

// FileOffset returns the file offset where encryption rolling is currently at
func (erv1sr *encryptionRollingV1StateReader) FileOffset() uint64 {
	return erv1sr.state.ErsbFileOffset
}

// FileExtentPhysicalBlockNumber returns the file extent physical block number
func (erv1sr *encryptionRollingV1StateReader) FileExtentPhysicalBlockNumber() uint64 {
	return erv1sr.state.ErsbFextPbn
}

// PhysicalAddress returns the physical address
func (erv1sr *encryptionRollingV1StateReader) PhysicalAddress() uint64 {
	return erv1sr.state.ErsbPaddr
}

// Progress returns the current progress of encryption rolling
func (erv1sr *encryptionRollingV1StateReader) Progress() uint64 {
	return erv1sr.state.ErsbProgress
}

// TotalBlocksToEncrypt returns the total number of blocks to encrypt
func (erv1sr *encryptionRollingV1StateReader) TotalBlocksToEncrypt() uint64 {
	return erv1sr.state.ErsbTotalBlkToEncrypt
}

// BlockmapOID returns the object identifier of the block map
func (erv1sr *encryptionRollingV1StateReader) BlockmapOID() uint64 {
	return erv1sr.state.ErsbBlockmapOid
}

// ChecksumCount returns the count of checksums
func (erv1sr *encryptionRollingV1StateReader) ChecksumCount() uint32 {
	return erv1sr.state.ErsbChecksumCount
}

// FileExtentCryptoID returns the file extent crypto identifier
func (erv1sr *encryptionRollingV1StateReader) FileExtentCryptoID() uint64 {
	return erv1sr.state.ErsbFextCid
}

// Checksums returns the checksums for the file extents
func (erv1sr *encryptionRollingV1StateReader) Checksums() []byte {
	result := make([]byte, len(erv1sr.state.ErsbChecksum))
	copy(result, erv1sr.state.ErsbChecksum)
	return result
}
