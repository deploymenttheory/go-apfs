package container

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// evictMappingReader implements the EvictMappingReader interface
type evictMappingReader struct {
	mapping *types.EvictMappingValT
	data    []byte
	endian  binary.ByteOrder
}

// NewEvictMappingReader creates a new EvictMappingReader implementation
func NewEvictMappingReader(data []byte, endian binary.ByteOrder) (interfaces.EvictMappingReader, error) {
	if len(data) < 16 { // EvictMappingValT is 16 bytes (8 + 8)
		return nil, fmt.Errorf("data too small for evict mapping: %d bytes", len(data))
	}

	mapping, err := parseEvictMapping(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse evict mapping: %w", err)
	}

	return &evictMappingReader{
		mapping: mapping,
		data:    data,
		endian:  endian,
	}, nil
}

// parseEvictMapping parses raw bytes into an EvictMappingValT structure
func parseEvictMapping(data []byte, endian binary.ByteOrder) (*types.EvictMappingValT, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("insufficient data for evict mapping")
	}

	mapping := &types.EvictMappingValT{}
	mapping.DstPaddr = types.Paddr(endian.Uint64(data[0:8]))
	mapping.Len = endian.Uint64(data[8:16])

	return mapping, nil
}

// DestinationAddress returns the address where the destination starts
func (emr *evictMappingReader) DestinationAddress() types.Paddr {
	return emr.mapping.DstPaddr
}

// Length returns the number of blocks being moved
func (emr *evictMappingReader) Length() uint64 {
	return emr.mapping.Len
}
