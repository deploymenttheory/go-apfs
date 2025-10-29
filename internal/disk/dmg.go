package device

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// DMGDevice provides access to APFS containers within DMG files
type DMGDevice struct {
	file   *os.File
	size   int64
	offset int64 // Offset to APFS container within DMG
}

// DMGConfig holds configuration for DMG handling
type DMGConfig struct {
	AutoDetectAPFS bool   `mapstructure:"auto_detect_apfs"`
	DefaultOffset  int64  `mapstructure:"default_offset"`
	CacheEnabled   bool   `mapstructure:"cache_enabled"`
	CacheSize      int    `mapstructure:"cache_size"`
	TestDataPath   string `mapstructure:"test_data_path"`
}

// LoadDMGConfig loads DMG configuration using Viper
func LoadDMGConfig() (*DMGConfig, error) {
	viper.SetConfigName("apfs-config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../..") // For tests running from subdirectories
	viper.AddConfigPath("$HOME/.apfs")
	viper.AddConfigPath("/etc/apfs")

	// Set defaults
	viper.SetDefault("auto_detect_apfs", true)
	viper.SetDefault("default_offset", 20480) // Common GPT offset for APFS
	viper.SetDefault("cache_enabled", true)
	viper.SetDefault("cache_size", 100)
	viper.SetDefault("test_data_path", "./tests")

	// Allow environment variables
	viper.SetEnvPrefix("APFS")
	viper.AutomaticEnv()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults
	}

	var config DMGConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

// OpenDMG opens a DMG file and detects the APFS container within it
func OpenDMG(path string, config *DMGConfig) (*DMGDevice, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open DMG file: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to stat DMG file: %w", err)
	}

	device := &DMGDevice{
		file: file,
		size: stat.Size(),
	}

	// Try to detect APFS container automatically
	if config.AutoDetectAPFS {
		offset, err := device.detectAPFSOffset()
		if err != nil {
			// Fall back to default offset
			device.offset = config.DefaultOffset
		} else {
			device.offset = offset
		}
	} else {
		device.offset = config.DefaultOffset
	}

	return device, nil
}

// detectAPFSOffset tries to find the APFS container within the DMG
// First attempts to parse GPT partition table to locate APFS partition,
// then falls back to signature scanning if GPT parsing fails
func (d *DMGDevice) detectAPFSOffset() (int64, error) {
	// Read file in chunks for scanning
	buf := make([]byte, 2*1024*1024) // 2MB buffer for scanning
	n, err := d.file.ReadAt(buf, 0)
	if err != nil && err != io.EOF {
		return 0, fmt.Errorf("failed to read DMG: %w", err)
	}

	fmt.Printf("[DMG] Starting APFS offset detection (file size: %d bytes)\n", d.size)

	// Method 1: Try to parse GPT partition table
	offset, err := d.parseGPTPartitionTable(buf[:n])
	if err == nil {
		fmt.Printf("[DMG] ✓ APFS found via GPT partition table at offset: %d (0x%x)\n", offset, offset)
		return offset, nil
	}
	fmt.Printf("[DMG] GPT parsing failed: %v\n", err)

	// Method 2: Scan common offsets for APFS magic
	fmt.Printf("[DMG] Attempting signature scan at common offsets...\n")
	commonOffsets := []struct {
		offset      int64
		description string
	}{
		{0, "Raw/unpartitioned APFS (no GPT)"},
		{types.GPTAPFSOffset, "After GPT header and partition entries (LBA 40, standard location)"},
		{types.NxMinimumContainerSize, "1MB boundary (some alternative formats)"},
	}

	for _, od := range commonOffsets {
		if od.offset+int64(types.APFSMagicOffset)+4 > int64(n) {
			continue
		}

		magicBytes := buf[od.offset+int64(types.APFSMagicOffset) : od.offset+int64(types.APFSMagicOffset)+4]
		magic := uint32(magicBytes[0]) |
			uint32(magicBytes[1])<<8 |
			uint32(magicBytes[2])<<16 |
			uint32(magicBytes[3])<<24

		if magic == types.NxMagic {
			fmt.Printf("[DMG] ✓ APFS found via signature scan at offset: %d (0x%x) - %s\n",
				od.offset, od.offset, od.description)
			return od.offset, nil
		}
	}

	// Method 3: Full scan at 4096-byte boundaries (APFS block size)
	fmt.Printf("[DMG] Attempting full signature scan at 4096-byte boundaries...\n")
	for i := int64(0); i < int64(n)-int64(types.APFSMagicOffset)-4; i += int64(types.NxDefaultBlockSize) {
		magicBytes := buf[i+int64(types.APFSMagicOffset) : i+int64(types.APFSMagicOffset)+4]
		magic := uint32(magicBytes[0]) |
			uint32(magicBytes[1])<<8 |
			uint32(magicBytes[2])<<16 |
			uint32(magicBytes[3])<<24

		if magic == types.NxMagic {
			fmt.Printf("[DMG] ✓ APFS found via full scan at offset: %d (0x%x)\n", i, i)
			return i, nil
		}
	}

	fmt.Printf("[DMG] ✗ APFS container not found\n")
	return 0, fmt.Errorf("APFS container not found in DMG file")
}

// parseGPTPartitionTable parses the GPT header to find APFS partition
// GPT structure: LBA 0-1 are reserved, LBA 1 contains primary GPT header,
// LBA 2+ contains partition entries
func (d *DMGDevice) parseGPTPartitionTable(buf []byte) (int64, error) {
	// Verify we have enough data
	if len(buf) < types.GPTEntriesStartOffset+types.GPTEntrySize {
		return 0, fmt.Errorf("buffer too small for GPT parsing")
	}

	// Check GPT header signature
	if len(buf) < types.GPTHeaderOffset+8 {
		return 0, fmt.Errorf("insufficient data for GPT header signature")
	}

	gptSignature := string(buf[types.GPTHeaderOffset : types.GPTHeaderOffset+8])
	if gptSignature != "EFI PART" {
		return 0, fmt.Errorf("no valid GPT signature found")
	}

	fmt.Printf("[DMG] Found GPT header signature\n")

	// Parse partition entries looking for APFS
	// Convert the APFS partition UUID string to byte array (in little-endian format for comparison)
	// UUID: 7C3457EF-0000-11AA-AA11-00306543ECAC
	apfsUUID := []byte{0xEF, 0x57, 0x34, 0x7C, 0x00, 0x00, 0xAA, 0x11,
		0xAA, 0x11, 0x00, 0x30, 0x65, 0x43, 0xEC, 0xAC}
	// Reference: types.ApfsGptPartitionUUID in efi_jumpstart.go

	// Read up to 128 partition entries
	for entryIdx := 0; entryIdx < 128; entryIdx++ {
		entryOffset := types.GPTEntriesStartOffset + (entryIdx * types.GPTEntrySize)
		if entryOffset+types.GPTEntrySize > len(buf) {
			break
		}

		entry := buf[entryOffset : entryOffset+types.GPTEntrySize]

		// First 16 bytes are partition type UUID
		partTypeUUID := entry[0:16]

		// Check if this is an APFS partition
		if bytes.Equal(partTypeUUID, apfsUUID) {
			// Bytes 32-39 contain start LBA (little-endian)
			startLBA := binary.LittleEndian.Uint64(entry[32:40])
			// Bytes 40-47 contain end LBA (little-endian)
			endLBA := binary.LittleEndian.Uint64(entry[40:48])

			startOffset := int64(startLBA) * 512 // Convert LBA to byte offset
			endOffset := int64(endLBA) * 512

			fmt.Printf("[DMG] Found APFS partition #%d: LBA %d-%d (offset 0x%x-0x%x, size %d MB)\n",
				entryIdx+1, startLBA, endLBA, startOffset, endOffset,
				(endOffset-startOffset)/(1024*1024))

			return startOffset, nil
		}
	}

	return 0, fmt.Errorf("no APFS partition found in GPT table")
}

// ReadAt implements io.ReaderAt for the APFS container within the DMG
func (d *DMGDevice) ReadAt(p []byte, off int64) (n int, err error) {
	// Adjust offset to account for APFS container position
	adjustedOff := d.offset + off
	return d.file.ReadAt(p, adjustedOff)
}

// Size returns the size of the APFS container
func (d *DMGDevice) Size() int64 {
	return d.size - d.offset
}

// Close closes the DMG file
func (d *DMGDevice) Close() error {
	if d.file != nil {
		return d.file.Close()
	}
	return nil
}

// BlockSize returns the block size (typically 4096 for APFS)
func (d *DMGDevice) BlockSize() int64 {
	return 4096
}

// GetTestDMGPath returns a path to test DMG files based on configuration
func GetTestDMGPath(filename string, config *DMGConfig) string {
	return filepath.Join(config.TestDataPath, filename)
}

// CreateDMGFromAPFS creates a test DMG file from an APFS container
func CreateDMGFromAPFS(apfsPath, dmgPath string) error {
	// Read the APFS container
	apfsFile, err := os.Open(apfsPath)
	if err != nil {
		return fmt.Errorf("failed to open APFS file: %w", err)
	}
	defer apfsFile.Close()

	// Create the DMG with GPT header
	dmgFile, err := os.Create(dmgPath)
	if err != nil {
		return fmt.Errorf("failed to create DMG file: %w", err)
	}
	defer dmgFile.Close()

	// Write placeholder GPT header (20480 bytes)
	gptHeader := make([]byte, 20480)
	// Add minimal GPT structure
	copy(gptHeader[0:8], []byte("EFI PART")) // GPT header signature

	if _, err := dmgFile.Write(gptHeader); err != nil {
		return fmt.Errorf("failed to write GPT header: %w", err)
	}

	// Copy APFS container
	if _, err := io.Copy(dmgFile, apfsFile); err != nil {
		return fmt.Errorf("failed to copy APFS container: %w", err)
	}

	return nil
}
