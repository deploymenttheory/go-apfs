package datastreams

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// extendedAttributeDataStreamReader implements the ExtendedAttributeDataStreamReader interface
type extendedAttributeDataStreamReader struct {
	xattrStream *types.JXattrDstreamT
	dataStream  interfaces.DataStreamReader
	data        []byte
	endian      binary.ByteOrder
}

// NewExtendedAttributeDataStreamReader creates a new ExtendedAttributeDataStreamReader implementation
func NewExtendedAttributeDataStreamReader(data []byte, endian binary.ByteOrder) (interfaces.ExtendedAttributeDataStreamReader, error) {
	if len(data) < 48 { // JXattrDstreamT is 48 bytes (8 + 40 for JDstreamT)
		return nil, fmt.Errorf("data too small for extended attribute data stream: %d bytes", len(data))
	}

	xattrStream, err := parseExtendedAttributeDataStream(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extended attribute data stream: %w", err)
	}

	// Create a data stream reader for the embedded JDstreamT
	dataStreamReader, err := NewDataStreamReader(data[8:], endian) // Skip the first 8 bytes (XattrObjId)
	if err != nil {
		return nil, fmt.Errorf("failed to create data stream reader: %w", err)
	}

	return &extendedAttributeDataStreamReader{
		xattrStream: xattrStream,
		dataStream:  dataStreamReader,
		data:        data,
		endian:      endian,
	}, nil
}

// parseExtendedAttributeDataStream parses raw bytes into a JXattrDstreamT structure
func parseExtendedAttributeDataStream(data []byte, endian binary.ByteOrder) (*types.JXattrDstreamT, error) {
	if len(data) < 48 {
		return nil, fmt.Errorf("insufficient data for extended attribute data stream")
	}

	xattrStream := &types.JXattrDstreamT{}
	xattrStream.XattrObjId = endian.Uint64(data[0:8])

	// Parse the embedded JDstreamT (starting at offset 8)
	xattrStream.Dstream.Size = endian.Uint64(data[8:16])
	xattrStream.Dstream.AllocedSize = endian.Uint64(data[16:24])
	xattrStream.Dstream.DefaultCryptoId = endian.Uint64(data[24:32])
	xattrStream.Dstream.TotalBytesWritten = endian.Uint64(data[32:40])
	xattrStream.Dstream.TotalBytesRead = endian.Uint64(data[40:48])

	return xattrStream, nil
}

// AttributeObjectID returns the identifier for the data stream
func (eadsr *extendedAttributeDataStreamReader) AttributeObjectID() uint64 {
	return eadsr.xattrStream.XattrObjId
}

// DataStream returns the data stream information
func (eadsr *extendedAttributeDataStreamReader) DataStream() interfaces.DataStreamReader {
	return eadsr.dataStream
}
