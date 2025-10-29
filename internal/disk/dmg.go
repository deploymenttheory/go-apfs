package disk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/viper"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// DMGDevice provides access to APFS containers within DMG files
type DMGDevice struct {
	file             *os.File
	size             int64
	offset           int64 // Offset to APFS container within DMG
	blockCache       map[uint64][]byte
	cacheMutex       sync.RWMutex
	maxCacheSize     int64
	currentCacheSize int64
	stats            *DMGStatistics
}

// DMGStatistics tracks DMG access statistics
type DMGStatistics struct {
	offsetDetectionTime time.Duration
	offsetMethod        string
	blocksRead          int64
	bytesRead           int64
	cacheHits           int64
	cacheMisses         int64
	mu                  sync.RWMutex
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
		file:             file,
		size:             stat.Size(),
		blockCache:       make(map[uint64][]byte),
		maxCacheSize:     int64(config.CacheSize) * 1024 * 1024,
		currentCacheSize: 0,
		stats: &DMGStatistics{
			offsetMethod: "unknown",
		},
	}

	// Try to detect APFS container automatically
	if config.AutoDetectAPFS {
		startTime := time.Now()
		offset, method, err := device.detectAPFSOffsetWithMethod()
		device.stats.offsetDetectionTime = time.Since(startTime)
		device.stats.offsetMethod = method
		if err != nil {
			// Fall back to default offset
			device.offset = config.DefaultOffset
			device.stats.offsetMethod = "fallback"
			fmt.Printf("[DMG] Using fallback offset: %d\n", device.offset)
		} else {
			device.offset = offset
			fmt.Printf("[DMG] Detection complete via %s in %v\n", method, device.stats.offsetDetectionTime)
		}
	} else {
		device.offset = config.DefaultOffset
		device.stats.offsetMethod = "configured"
	}

	return device, nil
}

// detectAPFSOffsetWithMethod tries to find the APFS container and returns the detection method
func (d *DMGDevice) detectAPFSOffsetWithMethod() (int64, string, error) {
	// Read file in chunks for scanning
	buf := make([]byte, 2*1024*1024) // 2MB buffer for scanning
	n, err := d.file.ReadAt(buf, 0)
	if err != nil && err != io.EOF {
		return 0, "", fmt.Errorf("failed to read DMG: %w", err)
	}

	fmt.Printf("[DMG] Starting APFS offset detection (file size: %d bytes)\n", d.size)

	// Method 1: Try to parse GPT partition table
	offset, err := d.parseGPTPartitionTable(buf[:n])
	if err == nil {
		fmt.Printf("[DMG] ✓ APFS found via GPT partition table at offset: %d (0x%x)\n", offset, offset)
		return offset, "gpt", nil
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
			return od.offset, "common_offsets", nil
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
			return i, "full_scan", nil
		}
	}

	fmt.Printf("[DMG] ✗ APFS container not found\n")
	return 0, "", fmt.Errorf("APFS container not found in DMG file")
}

// detectAPFSOffset tries to find the APFS container within the DMG (legacy wrapper)
func (d *DMGDevice) detectAPFSOffset() (int64, error) {
	offset, _, err := d.detectAPFSOffsetWithMethod()
	return offset, err
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

	// Check cache for this block
	blockNum := off / 4096
	if blockNum >= 0 && blockNum < int64(len(p)/4096) {
		d.cacheMutex.RLock()
		if cached, exists := d.blockCache[uint64(blockNum)]; exists && len(cached) == len(p) {
			copy(p, cached)
			d.stats.mu.Lock()
			d.stats.cacheHits++
			d.stats.mu.Unlock()
			d.cacheMutex.RUnlock()
			return len(p), nil
		}
		d.cacheMutex.RUnlock()
	}

	// Cache miss - read from file
	n, err = d.file.ReadAt(p, adjustedOff)
	if err == nil && n > 0 {
		// Update statistics
		d.stats.mu.Lock()
		d.stats.blocksRead++
		d.stats.bytesRead += int64(n)
		d.stats.cacheMisses++
		d.stats.mu.Unlock()

		// Try to cache the block if it's reasonable size
		if n == 4096 && blockNum >= 0 {
			d.cacheMutex.Lock()
			// Check if we have space
			if d.currentCacheSize+int64(n) <= d.maxCacheSize {
				blockData := make([]byte, n)
				copy(blockData, p)
				d.blockCache[uint64(blockNum)] = blockData
				d.currentCacheSize += int64(n)
			}
			d.cacheMutex.Unlock()
		}
	}

	return n, err
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

// GetStats returns current DMG access statistics
func (d *DMGDevice) GetStats() *DMGStatistics {
	d.stats.mu.RLock()
	defer d.stats.mu.RUnlock()
	return d.stats
}

// GetOffsetInfo returns information about the detected offset
func (d *DMGDevice) GetOffsetInfo() (int64, string, time.Duration) {
	d.stats.mu.RLock()
	defer d.stats.mu.RUnlock()
	return d.offset, d.stats.offsetMethod, d.stats.offsetDetectionTime
}

// ClearCache clears the block cache
func (d *DMGDevice) ClearCache() {
	d.cacheMutex.Lock()
	defer d.cacheMutex.Unlock()
	d.blockCache = make(map[uint64][]byte)
	d.currentCacheSize = 0
}

// CacheHitRate returns the cache hit rate as a percentage
func (d *DMGDevice) CacheHitRate() float64 {
	d.stats.mu.RLock()
	defer d.stats.mu.RUnlock()
	total := d.stats.cacheHits + d.stats.cacheMisses
	if total == 0 {
		return 0.0
	}
	return float64(d.stats.cacheHits) / float64(total) * 100.0
}

// PrintStats prints detailed statistics about DMG access
func (d *DMGDevice) PrintStats() {
	d.stats.mu.RLock()
	defer d.stats.mu.RUnlock()

	fmt.Println("=== DMG Device Statistics ===")
	fmt.Printf("Offset: %d bytes (0x%x)\n", d.offset, d.offset)
	fmt.Printf("Detection method: %s\n", d.stats.offsetMethod)
	fmt.Printf("Detection time: %v\n", d.stats.offsetDetectionTime)
	fmt.Printf("Blocks read: %d\n", d.stats.blocksRead)
	fmt.Printf("Bytes read: %d (%d MB)\n", d.stats.bytesRead, d.stats.bytesRead/(1024*1024))
	fmt.Printf("Cache hits: %d\n", d.stats.cacheHits)
	fmt.Printf("Cache misses: %d\n", d.stats.cacheMisses)
	total := d.stats.cacheHits + d.stats.cacheMisses
	if total > 0 {
		hitRate := float64(d.stats.cacheHits) / float64(total) * 100.0
		fmt.Printf("Hit rate: %.2f%%\n", hitRate)
	}
	fmt.Printf("Cache size: %d / %d bytes\n", d.currentCacheSize, d.maxCacheSize)
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

// ValidateAPFSMagic verifies the APFS magic signature at the detected offset
func (d *DMGDevice) ValidateAPFSMagic() (bool, error) {
	// Read the superblock area
	buf := make([]byte, 256)
	n, err := d.file.ReadAt(buf, d.offset)
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("failed to read for validation: %w", err)
	}
	if n < int(types.APFSMagicOffset)+4 {
		return false, fmt.Errorf("insufficient data for validation")
	}

	// Check for NXSB magic
	magicBytes := buf[types.APFSMagicOffset : types.APFSMagicOffset+4]
	magic := uint32(magicBytes[0]) |
		uint32(magicBytes[1])<<8 |
		uint32(magicBytes[2])<<16 |
		uint32(magicBytes[3])<<24

	if magic != types.NxMagic {
		return false, fmt.Errorf("invalid APFS magic: expected 0x%x, got 0x%x", types.NxMagic, magic)
	}

	return true, nil
}

// VerifyGPTStructure validates the GPT partition structure if present
func (d *DMGDevice) VerifyGPTStructure() (bool, error) {
	buf := make([]byte, 512+128*4) // GPT header + some entries
	n, err := d.file.ReadAt(buf, 0)
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("failed to read GPT: %w", err)
	}

	// Check GPT signature
	if n < types.GPTHeaderOffset+8 {
		return false, nil // No GPT present
	}

	gptSignature := string(buf[types.GPTHeaderOffset : types.GPTHeaderOffset+8])
	if gptSignature != "EFI PART" {
		return false, nil // No GPT present
	}

	return true, nil
}
