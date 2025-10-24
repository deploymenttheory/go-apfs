package snapshot

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// snapMetaExtReader implements the SnapMetaExtReader interface
type snapMetaExtReader struct {
	value  *types.SnapMetaExtT
	endian binary.ByteOrder
}

// NewSnapMetaExtReader creates a new SnapMetaExtReader implementation
func NewSnapMetaExtReader(data []byte, endian binary.ByteOrder) (interfaces.SnapMetaExtReader, error) {
	if len(data) < 40 { // SmeVersion (4) + SmeFlags (4) + SmeSnapXid (8) + SmeUuid (16) + SmeToken (8)
		return nil, fmt.Errorf("data too small for snapshot metadata extension: %d bytes", len(data))
	}

	value, err := parseSnapMetaExt(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse snapshot metadata extension: %w", err)
	}

	return &snapMetaExtReader{
		value:  value,
		endian: endian,
	}, nil
}

// parseSnapMetaExt parses raw bytes into a SnapMetaExtT structure
func parseSnapMetaExt(data []byte, endian binary.ByteOrder) (*types.SnapMetaExtT, error) {
	if len(data) < 40 {
		return nil, fmt.Errorf("insufficient data for snapshot metadata extension")
	}

	value := &types.SnapMetaExtT{}
	value.SmeVersion = endian.Uint32(data[0:4])
	value.SmeFlags = endian.Uint32(data[4:8])
	value.SmeSnapXid = types.XidT(endian.Uint64(data[8:16]))

	// Extract UUID (16 bytes)
	copy(value.SmeUuid[:], data[16:32])

	value.SmeToken = endian.Uint64(data[32:40])

	return value, nil
}

// Version returns the version of the extended metadata structure.
func (smer *snapMetaExtReader) Version() uint32 {
	return smer.value.SmeVersion
}

// Flags returns the extended metadata flags.
func (smer *snapMetaExtReader) Flags() uint32 {
	return smer.value.SmeFlags
}

// SnapXID returns the snapshot's transaction identifier.
func (smer *snapMetaExtReader) SnapXID() uint64 {
	return uint64(smer.value.SmeSnapXid)
}

// UUID returns the snapshot's UUID.
func (smer *snapMetaExtReader) UUID() [16]byte {
	return smer.value.SmeUuid
}

// Token returns the opaque metadata token.
func (smer *snapMetaExtReader) Token() uint64 {
	return smer.value.SmeToken
}
