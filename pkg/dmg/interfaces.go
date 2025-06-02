package dmg

import (
	"context"
	"io"
	"time"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// DMGService defines the interface for DMG file operations
type DMGService interface {
	// OpenDMG opens and parses a DMG file
	OpenDMG(ctx context.Context, dmgPath string) (*DMGInfo, error)

	// ListVolumes lists APFS volumes within the DMG
	ListVolumes(ctx context.Context, dmgPath string) ([]VolumeInfo, error)

	// ExtractVolume extracts an APFS volume from DMG to a file
	ExtractVolume(ctx context.Context, dmgPath string, volumeID uint64, outputPath string) error

	// StreamVolume provides streaming access to volume data
	StreamVolume(ctx context.Context, dmgPath string, volumeID uint64) (io.ReadCloser, error)

	// ValidateDMG performs DMG integrity validation
	ValidateDMG(ctx context.Context, dmgPath string) (*ValidationResult, error)

	// GetDMGInfo returns detailed DMG information
	GetDMGInfo(ctx context.Context, dmgPath string) (*DMGInfo, error)

	// Close closes the DMG and releases resources
	Close() error
}

// Parser defines the interface for DMG format parsers
type Parser interface {
	// CanHandle determines if this parser can handle the DMG format
	CanHandle(data []byte) bool

	// ParseHeader parses the DMG header
	ParseHeader(data []byte) (*DMGHeader, error)

	// ParsePartitionMap parses the partition map
	ParsePartitionMap(reader io.ReadSeeker) (*PartitionMap, error)

	// GetFormatName returns the format name
	GetFormatName() string

	// GetMinHeaderSize returns minimum header size for detection
	GetMinHeaderSize() int
}

// Extractor defines the interface for volume extraction
type Extractor interface {
	// ExtractVolume extracts a volume to the specified path
	ExtractVolume(ctx context.Context, dmgPath string, volume *VolumeInfo, outputPath string) error

	// StreamVolume provides streaming access to volume data
	StreamVolume(ctx context.Context, dmgPath string, volume *VolumeInfo) (io.ReadCloser, error)

	// ValidateExtraction validates the extracted volume
	ValidateExtraction(ctx context.Context, originalPath, extractedPath string) error
}

// DMGInfo contains comprehensive DMG file information
type DMGInfo struct {
	FilePath        string
	Format          string
	Size            uint64
	BlockSize       uint32
	Encrypted       bool
	Compressed      bool
	Volumes         []VolumeInfo
	PartitionMap    *PartitionMap
	CreatedAt       time.Time
	ModifiedAt      time.Time
	Checksum        string
	CompressionType string
	EncryptionType  string
	Metadata        map[string]interface{}
}

// VolumeInfo represents an APFS volume within the DMG
type VolumeInfo struct {
	ID            uint64
	Name          string
	Type          string
	FileSystem    string
	StartOffset   uint64
	Size          uint64
	BlockCount    uint64
	Encrypted     bool
	Compressed    bool
	ContainerUUID types.UUID
	VolumeUUID    types.UUID
	LastModified  time.Time
	Attributes    map[string]interface{}
}

// PartitionMap represents the DMG partition structure
type PartitionMap struct {
	Scheme     string
	BlockSize  uint32
	BlockCount uint64
	Partitions []Partition
}

// Partition represents a single partition in the DMG
type Partition struct {
	ID          uint32
	Type        string
	Name        string
	StartBlock  uint64
	BlockCount  uint64
	Attributes  uint32
	ProcessorID string
	BootArgs    string
	BootCode    []byte
}

// DMGHeader represents the common DMG header structure
type DMGHeader struct {
	Signature      uint32
	Version        uint32
	Format         string
	HeaderSize     uint32
	Flags          uint32
	BlockSize      uint32
	BlockCount     uint64
	DataOffset     uint64
	DataSize       uint64
	ChecksumType   string
	ChecksumOffset uint64
	ChecksumSize   uint32
	XMLOffset      uint64
	XMLSize        uint32
	Reserved       []byte
}

// ValidationResult represents DMG validation results
type ValidationResult struct {
	Valid             bool
	Format            string
	Errors            []string
	Warnings          []string
	ChecksumValid     bool
	StructureValid    bool
	VolumesAccessible []bool
	RecommendedAction string
	ValidationTime    time.Duration
}

// ExtractionOptions configures volume extraction behavior
type ExtractionOptions struct {
	PreserveMetadata   bool
	VerifyIntegrity    bool
	CreateSparseFile   bool
	CompressionEnabled bool
	ProgressCallback   func(bytesProcessed, totalBytes uint64)
	BufferSize         int
}
