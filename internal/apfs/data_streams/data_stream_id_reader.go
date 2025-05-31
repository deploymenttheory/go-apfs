package datastreams

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// dataStreamIDReader implements the DataStreamIDReader interface
type dataStreamIDReader struct {
	key    *types.JDstreamIdKeyT
	value  *types.JDstreamIdValT
	data   []byte
	endian binary.ByteOrder
}

// NewDataStreamIDReader creates a new DataStreamIDReader implementation
func NewDataStreamIDReader(keyData, valueData []byte, endian binary.ByteOrder) (interfaces.DataStreamIDReader, error) {
	if len(keyData) < 8 { // JKeyT is 8 bytes
		return nil, fmt.Errorf("key data too small for data stream ID key: %d bytes", len(keyData))
	}

	if len(valueData) < 4 { // JDstreamIdValT is 4 bytes (uint32)
		return nil, fmt.Errorf("value data too small for data stream ID value: %d bytes", len(valueData))
	}

	key, err := parseDataStreamIDKey(keyData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse data stream ID key: %w", err)
	}

	value, err := parseDataStreamIDValue(valueData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse data stream ID value: %w", err)
	}

	return &dataStreamIDReader{
		key:    key,
		value:  value,
		data:   append(keyData, valueData...),
		endian: endian,
	}, nil
}

// parseDataStreamIDKey parses raw bytes into a JDstreamIdKeyT structure
func parseDataStreamIDKey(data []byte, endian binary.ByteOrder) (*types.JDstreamIdKeyT, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("insufficient data for data stream ID key")
	}

	key := &types.JDstreamIdKeyT{}
	key.Hdr.ObjIdAndType = endian.Uint64(data[0:8])

	return key, nil
}

// parseDataStreamIDValue parses raw bytes into a JDstreamIdValT structure
func parseDataStreamIDValue(data []byte, endian binary.ByteOrder) (*types.JDstreamIdValT, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("insufficient data for data stream ID value")
	}

	value := &types.JDstreamIdValT{}
	value.Refcnt = endian.Uint32(data[0:4])

	return value, nil
}

// ReferenceCount returns the reference count for the data stream record
func (dsir *dataStreamIDReader) ReferenceCount() uint32 {
	return dsir.value.Refcnt
}

// ObjectID returns the object identifier for the data stream
func (dsir *dataStreamIDReader) ObjectID() uint64 {
	// The object identifier is stored in the header
	return dsir.key.Hdr.ObjIdAndType & types.ObjIdMask
}
