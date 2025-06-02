package parser

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/deploymenttheory/go-apfs/pkg/dmg"
)

// Constants for DMG format detection
const (
	// DMG signatures
	UDIFSignature uint32 = 0x6B6F6C79 // 'koly' - UDIF trailer signature
	UDCOSignature uint32 = 0x636F7565 // 'coue' - Compressed signature

	// Block sizes
	DefaultBlockSize = 512
	MinBlockSize     = 512
	MaxBlockSize     = 4096

	// Header sizes
	UDIFHeaderSize = 512
	MinHeaderSize  = 64
)

// HeaderParser handles DMG header parsing
type HeaderParser struct {
	blockSize uint32
}

// NewHeaderParser creates a new header parser
func NewHeaderParser() *HeaderParser {
	return &HeaderParser{
		blockSize: DefaultBlockSize,
	}
}

// ParseDMGHeader parses a DMG header from the given data
func (p *HeaderParser) ParseDMGHeader(reader io.ReadSeeker) (*dmg.DMGHeader, error) {
	// Seek to end to find UDIF trailer
	size, err := reader.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to end: %w", err)
	}

	if size < UDIFHeaderSize {
		return nil, fmt.Errorf("file too small for DMG: %d bytes", size)
	}

	// Read UDIF trailer (last 512 bytes)
	trailerOffset := size - UDIFHeaderSize
	_, err = reader.Seek(trailerOffset, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to trailer: %w", err)
	}

	trailerData := make([]byte, UDIFHeaderSize)
	_, err = io.ReadFull(reader, trailerData)
	if err != nil {
		return nil, fmt.Errorf("failed to read trailer: %w", err)
	}

	return p.parseUDIFTrailer(trailerData)
}

// parseUDIFTrailer parses the UDIF trailer structure
func (p *HeaderParser) parseUDIFTrailer(data []byte) (*dmg.DMGHeader, error) {
	if len(data) < UDIFHeaderSize {
		return nil, fmt.Errorf("insufficient data for UDIF trailer")
	}

	// Check signature at the beginning of trailer
	signature := binary.BigEndian.Uint32(data[0:4])
	if signature != UDIFSignature {
		return nil, fmt.Errorf("invalid UDIF signature: 0x%08X", signature)
	}

	header := &dmg.DMGHeader{
		Signature:  signature,
		Format:     "UDIF",
		HeaderSize: UDIFHeaderSize,
	}

	// Parse key fields (simplified UDIF structure)
	header.Version = binary.BigEndian.Uint32(data[4:8])
	header.HeaderSize = binary.BigEndian.Uint32(data[8:12])
	header.Flags = binary.BigEndian.Uint32(data[12:16])

	// Data fork information
	header.DataOffset = binary.BigEndian.Uint64(data[16:24])
	header.DataSize = binary.BigEndian.Uint64(data[24:32])

	// Block information
	header.BlockSize = DefaultBlockSize
	if header.DataSize > 0 {
		header.BlockCount = (header.DataSize + uint64(header.BlockSize) - 1) / uint64(header.BlockSize)
	}

	// Checksum information
	header.ChecksumType = "CRC32"
	header.ChecksumOffset = binary.BigEndian.Uint64(data[32:40])
	header.ChecksumSize = binary.BigEndian.Uint32(data[40:44])

	// XML plist information (contains partition data)
	header.XMLOffset = binary.BigEndian.Uint64(data[44:52])
	header.XMLSize = binary.BigEndian.Uint32(data[52:56])

	// Reserved space
	header.Reserved = make([]byte, 56)
	copy(header.Reserved, data[56:112])

	return header, nil
}

// ValidateHeader validates the DMG header structure
func (p *HeaderParser) ValidateHeader(header *dmg.DMGHeader) error {
	if header.Signature != UDIFSignature {
		return fmt.Errorf("invalid signature: 0x%08X", header.Signature)
	}

	if header.HeaderSize < MinHeaderSize {
		return fmt.Errorf("header size too small: %d", header.HeaderSize)
	}

	if header.BlockSize < MinBlockSize || header.BlockSize > MaxBlockSize {
		return fmt.Errorf("invalid block size: %d", header.BlockSize)
	}

	if header.DataSize == 0 {
		return fmt.Errorf("data size cannot be zero")
	}

	return nil
}

// DetectFormat attempts to detect the DMG format from header data
func (p *HeaderParser) DetectFormat(data []byte) (string, error) {
	if len(data) < 4 {
		return "", fmt.Errorf("insufficient data for format detection")
	}

	// Check for various DMG signatures
	signature := binary.BigEndian.Uint32(data[0:4])

	switch signature {
	case UDIFSignature:
		return "UDIF", nil
	case UDCOSignature:
		return "UDCO", nil
	default:
		// Check if it might be a different format or raw disk image
		if p.looksLikeRawDisk(data) {
			return "RAW", nil
		}
		return "", fmt.Errorf("unknown DMG format: signature 0x%08X", signature)
	}
}

// looksLikeRawDisk performs heuristic detection for raw disk images
func (p *HeaderParser) looksLikeRawDisk(data []byte) bool {
	if len(data) < 512 {
		return false
	}

	// Look for partition table signatures or filesystem signatures
	// Check for MBR signature
	if len(data) >= 510 && data[510] == 0x55 && data[511] == 0xAA {
		return true
	}

	// Check for GPT signature
	if len(data) >= 512 {
		gptSig := string(data[0:8])
		if gptSig == "EFI PART" {
			return true
		}
	}

	// Check for APFS container signature
	if len(data) >= 32 {
		apfsSig := binary.LittleEndian.Uint32(data[32:36])
		if apfsSig == 0x42535058 { // 'NXSB' in little endian
			return true
		}
	}

	return false
}

// SetBlockSize sets the block size for parsing
func (p *HeaderParser) SetBlockSize(size uint32) error {
	if size < MinBlockSize || size > MaxBlockSize {
		return fmt.Errorf("invalid block size: %d", size)
	}
	p.blockSize = size
	return nil
}

// GetBlockSize returns the current block size
func (p *HeaderParser) GetBlockSize() uint32 {
	return p.blockSize
}
