package container

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// checkpointMappingReader implements the CheckpointMappingReader interface
type checkpointMappingReader struct {
	mapping *types.CheckpointMappingT
	data    []byte
	endian  binary.ByteOrder
}

// NewCheckpointMappingReader creates a new CheckpointMappingReader implementation
func NewCheckpointMappingReader(data []byte, endian binary.ByteOrder) (interfaces.CheckpointMappingReader, error) {
	if len(data) < 40 { // CheckpointMappingT is 40 bytes (4+4+4+4+8+8+8)
		return nil, fmt.Errorf("data too small for checkpoint mapping: %d bytes", len(data))
	}

	mapping, err := parseCheckpointMapping(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse checkpoint mapping: %w", err)
	}

	return &checkpointMappingReader{
		mapping: mapping,
		data:    data,
		endian:  endian,
	}, nil
}

// parseCheckpointMapping parses raw bytes into a CheckpointMappingT structure
func parseCheckpointMapping(data []byte, endian binary.ByteOrder) (*types.CheckpointMappingT, error) {
	if len(data) < 40 {
		return nil, fmt.Errorf("insufficient data for checkpoint mapping")
	}

	mapping := &types.CheckpointMappingT{}
	mapping.CpmType = endian.Uint32(data[0:4])
	mapping.CpmSubtype = endian.Uint32(data[4:8])
	mapping.CpmSize = endian.Uint32(data[8:12])
	mapping.CpmPad = endian.Uint32(data[12:16])
	mapping.CpmFsOid = types.OidT(endian.Uint64(data[16:24]))
	mapping.CpmOid = types.OidT(endian.Uint64(data[24:32]))
	mapping.CpmPaddr = types.Paddr(endian.Uint64(data[32:40]))

	return mapping, nil
}

// Type returns the object's type
func (cmr *checkpointMappingReader) Type() uint32 {
	return cmr.mapping.CpmType
}

// Subtype returns the object's subtype
func (cmr *checkpointMappingReader) Subtype() uint32 {
	return cmr.mapping.CpmSubtype
}

// Size returns the size of the object in bytes
func (cmr *checkpointMappingReader) Size() uint32 {
	return cmr.mapping.CpmSize
}

// FilesystemOID returns the virtual object identifier of the volume
func (cmr *checkpointMappingReader) FilesystemOID() types.OidT {
	return cmr.mapping.CpmFsOid
}

// ObjectID returns the ephemeral object identifier
func (cmr *checkpointMappingReader) ObjectID() types.OidT {
	return cmr.mapping.CpmOid
}

// PhysicalAddress returns the address in the checkpoint data area where the object is stored
func (cmr *checkpointMappingReader) PhysicalAddress() types.Paddr {
	return cmr.mapping.CpmPaddr
}
