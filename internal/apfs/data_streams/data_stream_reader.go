package datastreams

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// dataStreamReader implements the DataStreamReader interface
type dataStreamReader struct {
	stream *types.JDstreamT
	data   []byte
	endian binary.ByteOrder
}

// NewDataStreamReader creates a new DataStreamReader implementation
func NewDataStreamReader(data []byte, endian binary.ByteOrder) (interfaces.DataStreamReader, error) {
	if len(data) < 40 { // JDstreamT is 40 bytes (8 + 8 + 8 + 8 + 8)
		return nil, fmt.Errorf("data too small for data stream: %d bytes", len(data))
	}

	stream, err := parseDataStream(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse data stream: %w", err)
	}

	return &dataStreamReader{
		stream: stream,
		data:   data,
		endian: endian,
	}, nil
}

// parseDataStream parses raw bytes into a JDstreamT structure
func parseDataStream(data []byte, endian binary.ByteOrder) (*types.JDstreamT, error) {
	if len(data) < 40 {
		return nil, fmt.Errorf("insufficient data for data stream")
	}

	stream := &types.JDstreamT{}
	stream.Size = endian.Uint64(data[0:8])
	stream.AllocedSize = endian.Uint64(data[8:16])
	stream.DefaultCryptoId = endian.Uint64(data[16:24])
	stream.TotalBytesWritten = endian.Uint64(data[24:32])
	stream.TotalBytesRead = endian.Uint64(data[32:40])

	return stream, nil
}

// Size returns the size of the data in bytes
func (dsr *dataStreamReader) Size() uint64 {
	return dsr.stream.Size
}

// AllocatedSize returns the total space allocated for the data stream
func (dsr *dataStreamReader) AllocatedSize() uint64 {
	return dsr.stream.AllocedSize
}

// DefaultCryptoID returns the default encryption key or tweak used in this data stream
func (dsr *dataStreamReader) DefaultCryptoID() uint64 {
	return dsr.stream.DefaultCryptoId
}

// TotalBytesWritten returns the total bytes written to this data stream
func (dsr *dataStreamReader) TotalBytesWritten() uint64 {
	return dsr.stream.TotalBytesWritten
}

// TotalBytesRead returns the total bytes read from this data stream
func (dsr *dataStreamReader) TotalBytesRead() uint64 {
	return dsr.stream.TotalBytesRead
}
