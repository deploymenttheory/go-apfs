package services

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"fmt"
	"hash/adler32"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// CompressionService handles decompression of APFS compressed data
type CompressionService struct{}

// NewCompressionService creates a new compression service
func NewCompressionService() *CompressionService {
	return &CompressionService{}
}

// Decompress decompresses data based on the compression method
func (cs *CompressionService) Decompress(compressedData []byte, method types.CompressionMethodType) ([]byte, error) {
	switch method {
	case types.CompressionMethodDeflate:
		return cs.DecompressDeflate(compressedData)
	case types.CompressionMethodLzfse:
		return nil, fmt.Errorf("lzfse decompression not yet implemented")
	case types.CompressionMethodLzvn:
		return nil, fmt.Errorf("lzvn decompression not yet implemented")
	case types.CompressionMethodLz4:
		return nil, fmt.Errorf("lz4 decompression requires external package")
	case types.CompressionMethodZstd:
		return nil, fmt.Errorf("zstandard decompression requires external package")
	default:
		return nil, fmt.Errorf("unknown compression method: %d", method)
	}
}

// DecompressDeflate decompresses DEFLATE-compressed data
func (cs *CompressionService) DecompressDeflate(compressedData []byte) ([]byte, error) {
	if len(compressedData) < 4 {
		return nil, fmt.Errorf("insufficient data for deflate decompression")
	}

	// Use Go's built-in DEFLATE decompressor
	reader := flate.NewReader(bytes.NewReader(compressedData))
	defer reader.Close()

	var result bytes.Buffer
	if _, err := result.ReadFrom(reader); err != nil {
		return nil, fmt.Errorf("deflate decompression failed: %w", err)
	}

	return result.Bytes(), nil
}

// DecompressDeflateZlib decompresses zlib-wrapped DEFLATE data (RFC 1950)
func (cs *CompressionService) DecompressDeflateZlib(compressedData []byte) ([]byte, error) {
	if len(compressedData) < 2 {
		return nil, fmt.Errorf("insufficient data for zlib header")
	}

	// Verify zlib header
	header := binary.BigEndian.Uint16(compressedData[:2])
	if (header % 31) != 0 {
		return nil, fmt.Errorf("invalid zlib header checksum")
	}

	// Skip the 2-byte header and last 4-byte checksum
	if len(compressedData) < 6 {
		return nil, fmt.Errorf("insufficient data for zlib format")
	}

	deflateData := compressedData[2 : len(compressedData)-4]

	// Decompress the DEFLATE content
	reader := flate.NewReader(bytes.NewReader(deflateData))
	defer reader.Close()

	var result bytes.Buffer
	if _, err := result.ReadFrom(reader); err != nil {
		return nil, fmt.Errorf("deflate decompression failed: %w", err)
	}

	decompressed := result.Bytes()

	// Verify Adler-32 checksum
	expectedChecksum := binary.BigEndian.Uint32(compressedData[len(compressedData)-4:])
	actualChecksum := uint32(adler32.Checksum(decompressed))

	if expectedChecksum != actualChecksum {
		return nil, fmt.Errorf("adler32 checksum mismatch: expected %08x, got %08x", expectedChecksum, actualChecksum)
	}

	return decompressed, nil
}

// VerifyDeflateChecksum verifies the Adler-32 checksum of deflate data
func (cs *CompressionService) VerifyDeflateChecksum(data []byte, expectedChecksum uint32) bool {
	return uint32(adler32.Checksum(data)) == expectedChecksum
}

// ComputeDeflateChecksum computes the Adler-32 checksum for data
func (cs *CompressionService) ComputeDeflateChecksum(data []byte) uint32 {
	return uint32(adler32.Checksum(data))
}
