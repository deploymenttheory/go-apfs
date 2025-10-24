package device

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
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
	viper.AddConfigPath("../..")  // For tests running from subdirectories
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
func (d *DMGDevice) detectAPFSOffset() (int64, error) {
	// Read first few KB to look for APFS magic
	buf := make([]byte, 65536) // 64KB should be enough
	n, err := d.file.ReadAt(buf, 0)
	if err != nil && err != io.EOF {
		return 0, err
	}

	// Look for APFS magic number (NXSB - 0x4253584E in little endian)
	apfsMagic := []byte{0x4E, 0x58, 0x53, 0x42}
	
	// Common offsets to check
	offsets := []int64{
		0,     // DMG might be raw APFS
		20480, // Common GPT partition start (block 40 * 512)
		32768, // Block 64 * 512
		65536, // Block 128 * 512
	}

	for _, offset := range offsets {
		if offset+4 > int64(n) {
			continue
		}
		
		// Check for APFS magic at this offset + 32 bytes (where magic is in nx_superblock_t)
		magicOffset := offset + 32
		if magicOffset+4 <= int64(n) {
			if buf[magicOffset] == apfsMagic[0] &&
				buf[magicOffset+1] == apfsMagic[1] &&
				buf[magicOffset+2] == apfsMagic[2] &&
				buf[magicOffset+3] == apfsMagic[3] {
				return offset, nil
			}
		}
	}

	return 0, fmt.Errorf("APFS container not found in DMG")
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