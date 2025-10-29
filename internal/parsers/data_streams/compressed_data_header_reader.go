package datastreams

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// CompressedDataHeaderReader reads and parses compressed data headers
type CompressedDataHeaderReader struct {
	CompressionMethod types.CompressionMethodType
	UncompressedSize  uint64
	RawData           []byte
	Endian            binary.ByteOrder
}

// NewCompressedDataHeaderReader creates a new compressed data header reader
func NewCompressedDataHeaderReader(data []byte, endian binary.ByteOrder) (*CompressedDataHeaderReader, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("insufficient data for compressed data header: %d bytes (need 16)", len(data))
	}

	// Verify the signature
	signature := endian.Uint32(data[0:4])
	if signature != types.CompressionSignature {
		return nil, fmt.Errorf("invalid compression signature: 0x%08x (expected 0x%08x)", signature, types.CompressionSignature)
	}

	reader := &CompressedDataHeaderReader{
		CompressionMethod: types.CompressionMethodType(endian.Uint32(data[4:8])),
		UncompressedSize:  endian.Uint64(data[8:16]),
		RawData:           data,
		Endian:            endian,
	}

	return reader, nil
}

// GetCompressionMethod returns the compression method used
func (cdhr *CompressedDataHeaderReader) GetCompressionMethod() types.CompressionMethodType {
	return cdhr.CompressionMethod
}

// GetUncompressedSize returns the uncompressed data size in bytes
func (cdhr *CompressedDataHeaderReader) GetUncompressedSize() uint64 {
	return cdhr.UncompressedSize
}

// IsValidCompressionMethod checks if the compression method is recognized
func (cdhr *CompressedDataHeaderReader) IsValidCompressionMethod() bool {
	switch cdhr.CompressionMethod {
	case types.CompressionMethodDeflate,
		types.CompressionMethodLzfse,
		types.CompressionMethodLzvn,
		types.CompressionMethodLz4,
		types.CompressionMethodZstd:
		return true
	default:
		return false
	}
}

// GetCompressionMethodName returns a human-readable name for the compression method
func (cdhr *CompressedDataHeaderReader) GetCompressionMethodName() string {
	switch cdhr.CompressionMethod {
	case types.CompressionMethodDeflate:
		return "DEFLATE"
	case types.CompressionMethodLzfse:
		return "LZFSE"
	case types.CompressionMethodLzvn:
		return "LZVN"
	case types.CompressionMethodLz4:
		return "LZ4"
	case types.CompressionMethodZstd:
		return "Zstandard"
	default:
		return fmt.Sprintf("Unknown (0x%x)", cdhr.CompressionMethod)
	}
}
