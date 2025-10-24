package sealed_volumes

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// integrityMetaReader implements the IntegrityReader interface
type integrityMetaReader struct {
	metadata *types.IntegrityMetaPhysT
	endian   binary.ByteOrder
}

// NewIntegrityMetaReader creates a new IntegrityReader implementation
func NewIntegrityMetaReader(data []byte, endian binary.ByteOrder) (interfaces.IntegrityReader, error) {
	if len(data) < 56 { // ObjPhysT (56) + ImVersion (4) + ImFlags (4) + ImHashType (4) + ImRootHashOffset (4) + ImBrokenXid (8) + ImReserved (72) = 152 bytes
		return nil, fmt.Errorf("data too small for integrity metadata: %d bytes", len(data))
	}

	metadata, err := parseIntegrityMeta(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse integrity metadata: %w", err)
	}

	return &integrityMetaReader{
		metadata: metadata,
		endian:   endian,
	}, nil
}

// parseIntegrityMeta parses raw bytes into an IntegrityMetaPhysT structure
func parseIntegrityMeta(data []byte, endian binary.ByteOrder) (*types.IntegrityMetaPhysT, error) {
	if len(data) < 152 {
		return nil, fmt.Errorf("insufficient data for integrity metadata: need 152 bytes, got %d", len(data))
	}

	metadata := &types.IntegrityMetaPhysT{}

	// Parse ObjPhysT header (40 bytes)
	// OChecksum (32 bytes)
	copy(metadata.ImO.OChecksum[:], data[0:32])
	// OOid (8 bytes)
	metadata.ImO.OOid = types.OidT(endian.Uint64(data[32:40]))
	// OXid (8 bytes)
	metadata.ImO.OXid = types.XidT(endian.Uint64(data[40:48]))
	// OType (4 bytes)
	metadata.ImO.OType = endian.Uint32(data[48:52])
	// OSubtype (4 bytes)
	metadata.ImO.OSubtype = endian.Uint32(data[52:56])

	// Parse ImVersion
	metadata.ImVersion = endian.Uint32(data[56:60])

	// Parse ImFlags
	metadata.ImFlags = endian.Uint32(data[60:64])

	// Parse ImHashType
	metadata.ImHashType = types.ApfsHashTypeT(endian.Uint32(data[64:68]))

	// Parse ImRootHashOffset
	metadata.ImRootHashOffset = endian.Uint32(data[68:72])

	// Parse ImBrokenXid
	metadata.ImBrokenXid = types.XidT(endian.Uint64(data[72:80]))

	// Parse ImReserved (9 * 8 = 72 bytes)
	for i := 0; i < 9; i++ {
		offset := 80 + (i * 8)
		metadata.ImReserved[i] = endian.Uint64(data[offset : offset+8])
	}

	return metadata, nil
}

// Version returns the version of the integrity metadata structure
func (imr *integrityMetaReader) Version() uint32 {
	return imr.metadata.ImVersion
}

// Flags returns the integrity metadata flags
func (imr *integrityMetaReader) Flags() uint32 {
	return imr.metadata.ImFlags
}

// HashType returns the hash algorithm being used
func (imr *integrityMetaReader) HashType() types.ApfsHashTypeT {
	return imr.metadata.ImHashType
}

// RootHashOffset returns the offset of the root hash
func (imr *integrityMetaReader) RootHashOffset() uint32 {
	return imr.metadata.ImRootHashOffset
}

// IsSealBroken checks if the seal has been broken
func (imr *integrityMetaReader) IsSealBroken() bool {
	return (imr.metadata.ImFlags & types.ApfsSealBroken) != 0
}

// BrokenTransactionID returns the transaction ID that broke the seal
func (imr *integrityMetaReader) BrokenTransactionID() types.XidT {
	return imr.metadata.ImBrokenXid
}

// IsVersionValid checks if the version is valid
func (imr *integrityMetaReader) IsVersionValid() bool {
	return imr.metadata.ImVersion >= types.IntegrityMetaVersion1 && imr.metadata.ImVersion <= types.IntegrityMetaVersionHighest
}

// IsHashTypeValid checks if the hash type is valid
func (imr *integrityMetaReader) IsHashTypeValid() bool {
	return imr.metadata.ImHashType >= types.ApfsHashMin && imr.metadata.ImHashType <= types.ApfsHashMax
}
