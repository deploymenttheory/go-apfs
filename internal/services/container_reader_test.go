package services

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func getTestContainerPath() string {
	// Walk up directory tree to find go-apfs root
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Keep walking up until we find the go-apfs directory
	for {
		if filepath.Base(cwd) == "go-apfs" {
			// Found it! Now navigate to tests folder
			return filepath.Join(cwd, "tests", "test_container.img")
		}

		parent := filepath.Dir(cwd)
		if parent == cwd {
			// Reached root directory without finding go-apfs
			break
		}
		cwd = parent
	}

	// If we can't find go-apfs, return empty string and let test fail
	return ""
}

// extractAPFSContainerFromDiskImage extracts the APFS container from a raw disk image
// The raw disk image has a GPT partition table at the start, with APFS container at block 5
func extractAPFSContainerFromDiskImage(diskImagePath string) (string, error) {
	tempFile, err := os.CreateTemp("", "apfs-test-*.img")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	originalFile, err := os.Open(diskImagePath)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to open disk image: %w", err)
	}
	defer originalFile.Close()

	// Skip GPT partition table (20480 bytes = block 5)
	if _, err := originalFile.Seek(20480, io.SeekStart); err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to seek: %w", err)
	}

	// Copy APFS container data to temp file
	if _, err := io.Copy(tempFile, originalFile); err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to copy container data: %w", err)
	}

	return tempFile.Name(), nil
}

func TestNewContainerReader(t *testing.T) {
	containerPath := getTestContainerPath()

	// The test_container.img has GPT partition table at the start
	// APFS container starts at block 5 (offset 20480 bytes)
	// We need to read from the actual APFS container, not the raw disk image
	// For now, extract just the APFS container portion
	tempFile, err := os.CreateTemp("", "apfs-test-*.img")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy APFS container data (starting at block 5) to temp file
	originalFile, err := os.Open(containerPath)
	if err != nil {
		t.Fatalf("Failed to open test container: %v", err)
	}
	defer originalFile.Close()

	// Skip GPT and read from offset 20480 (block 5)
	if _, err := originalFile.Seek(20480, io.SeekStart); err != nil {
		t.Fatalf("Failed to seek: %v", err)
	}

	if _, err := io.Copy(tempFile, originalFile); err != nil {
		t.Fatalf("Failed to copy container data: %v", err)
	}
	tempFile.Close()

	cr, err := NewContainerReader(tempFile.Name())
	if err != nil {
		t.Fatalf("NewContainerReader failed: %v", err)
	}
	defer cr.Close()

	if cr == nil {
		t.Fatal("ContainerReader is nil")
	}

	if cr.GetBlockSize() == 0 {
		t.Error("Block size is 0")
	}

	if cr.GetContainerSize() == 0 {
		t.Error("Container size is 0")
	}
}

func TestNewContainerReaderInvalidPath(t *testing.T) {
	_, err := NewContainerReader("")
	if err == nil {
		t.Error("Expected error for empty path")
	}
}

func TestNewContainerReaderNonexistent(t *testing.T) {
	_, err := NewContainerReader("/nonexistent/file.img")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestReadBlock(t *testing.T) {
	containerPath := getTestContainerPath()
	if containerPath == "" {
		t.Skip("test_container.img not found")
	}

	tempPath, err := extractAPFSContainerFromDiskImage(containerPath)
	if err != nil {
		t.Fatalf("Failed to extract APFS container: %v", err)
	}
	defer os.Remove(tempPath)

	cr, err := NewContainerReader(tempPath)
	if err != nil {
		t.Fatalf("NewContainerReader failed: %v", err)
	}
	defer cr.Close()

	block, err := cr.ReadBlock(0)
	if err != nil {
		t.Fatalf("ReadBlock failed: %v", err)
	}

	if len(block) != int(cr.GetBlockSize()) {
		t.Errorf("Block size mismatch: got %d, want %d", len(block), cr.GetBlockSize())
	}

	// Verify magic number is present
	magic := binary.LittleEndian.Uint32(block[32:36])
	if magic != types.NxMagic {
		t.Errorf("Magic number mismatch: got 0x%08x, want 0x%08x", magic, types.NxMagic)
	}
}

func TestReadBlockCaching(t *testing.T) {
	containerPath := getTestContainerPath()
	if containerPath == "" {
		t.Skip("test_container.img not found")
	}

	tempPath, err := extractAPFSContainerFromDiskImage(containerPath)
	if err != nil {
		t.Fatalf("Failed to extract APFS container: %v", err)
	}
	defer os.Remove(tempPath)

	cr, err := NewContainerReader(tempPath)
	if err != nil {
		t.Fatalf("NewContainerReader failed: %v", err)
	}
	defer cr.Close()

	_, err = cr.ReadBlock(1)
	if err != nil {
		t.Fatalf("ReadBlock failed: %v", err)
	}

	if !cr.IsCached(1) {
		t.Error("Block should be cached")
	}

	if cr.IsCached(999) {
		t.Error("Block 999 should not be cached")
	}
}

func TestReadBlockOutOfBounds(t *testing.T) {
	containerPath := getTestContainerPath()
	if containerPath == "" {
		t.Skip("test_container.img not found")
	}

	tempPath, err := extractAPFSContainerFromDiskImage(containerPath)
	if err != nil {
		t.Fatalf("Failed to extract APFS container: %v", err)
	}
	defer os.Remove(tempPath)

	cr, err := NewContainerReader(tempPath)
	if err != nil {
		t.Fatalf("NewContainerReader failed: %v", err)
	}
	defer cr.Close()

	_, err = cr.ReadBlock(10000000)
	if err == nil {
		t.Error("Expected error for out-of-bounds block")
	}
}

func TestReadBlocks(t *testing.T) {
	containerPath := getTestContainerPath()
	if containerPath == "" {
		t.Skip("test_container.img not found")
	}

	tempPath, err := extractAPFSContainerFromDiskImage(containerPath)
	if err != nil {
		t.Fatalf("Failed to extract APFS container: %v", err)
	}
	defer os.Remove(tempPath)

	cr, err := NewContainerReader(tempPath)
	if err != nil {
		t.Fatalf("NewContainerReader failed: %v", err)
	}
	defer cr.Close()

	data, err := cr.ReadBlocks(1, 3)
	if err != nil {
		t.Fatalf("ReadBlocks failed: %v", err)
	}

	expectedSize := 3 * cr.GetBlockSize()
	if len(data) != int(expectedSize) {
		t.Errorf("Data size mismatch: got %d, want %d", len(data), expectedSize)
	}
}

func TestReadBlocksZeroCount(t *testing.T) {
	containerPath := getTestContainerPath()
	if containerPath == "" {
		t.Skip("test_container.img not found")
	}

	tempPath, err := extractAPFSContainerFromDiskImage(containerPath)
	if err != nil {
		t.Fatalf("Failed to extract APFS container: %v", err)
	}
	defer os.Remove(tempPath)

	cr, err := NewContainerReader(tempPath)
	if err != nil {
		t.Fatalf("NewContainerReader failed: %v", err)
	}
	defer cr.Close()

	data, err := cr.ReadBlocks(0, 0)
	if err != nil {
		t.Fatalf("ReadBlocks failed: %v", err)
	}

	if len(data) != 0 {
		t.Errorf("Expected empty data, got %d bytes", len(data))
	}
}

func TestGetBlockSize(t *testing.T) {
	containerPath := getTestContainerPath()
	if containerPath == "" {
		t.Skip("test_container.img not found")
	}

	tempPath, err := extractAPFSContainerFromDiskImage(containerPath)
	if err != nil {
		t.Fatalf("Failed to extract APFS container: %v", err)
	}
	defer os.Remove(tempPath)

	cr, err := NewContainerReader(tempPath)
	if err != nil {
		t.Fatalf("NewContainerReader failed: %v", err)
	}
	defer cr.Close()

	blockSize := cr.GetBlockSize()
	if blockSize != 4096 {
		t.Errorf("Block size mismatch: got %d, want 4096", blockSize)
	}
}

func TestGetContainerSize(t *testing.T) {
	containerPath := getTestContainerPath()
	if containerPath == "" {
		t.Skip("test_container.img not found")
	}

	tempPath, err := extractAPFSContainerFromDiskImage(containerPath)
	if err != nil {
		t.Fatalf("Failed to extract APFS container: %v", err)
	}
	defer os.Remove(tempPath)

	cr, err := NewContainerReader(tempPath)
	if err != nil {
		t.Fatalf("NewContainerReader failed: %v", err)
	}
	defer cr.Close()

	size := cr.GetContainerSize()
	if size == 0 {
		t.Error("Container size should not be 0")
	}

	// Size should be ~100MB (104857600 bytes total - 20480 bytes skipped for GPT = ~104837120)
	if size < 99*1024*1024 {
		t.Errorf("Container size too small: got %d bytes", size)
	}
}

func TestGetSuperblock(t *testing.T) {
	containerPath := getTestContainerPath()
	if containerPath == "" {
		t.Skip("test_container.img not found")
	}

	tempPath, err := extractAPFSContainerFromDiskImage(containerPath)
	if err != nil {
		t.Fatalf("Failed to extract APFS container: %v", err)
	}
	defer os.Remove(tempPath)

	cr, err := NewContainerReader(tempPath)
	if err != nil {
		t.Fatalf("NewContainerReader failed: %v", err)
	}
	defer cr.Close()

	sb := cr.GetSuperblock()
	if sb == nil {
		t.Fatal("Superblock is nil")
	}

	if sb.NxMagic != types.NxMagic {
		t.Errorf("Magic number mismatch: got 0x%08x, want 0x%08x", sb.NxMagic, types.NxMagic)
	}
}

func TestClearCache(t *testing.T) {
	containerPath := getTestContainerPath()
	if containerPath == "" {
		t.Skip("test_container.img not found")
	}

	tempPath, err := extractAPFSContainerFromDiskImage(containerPath)
	if err != nil {
		t.Fatalf("Failed to extract APFS container: %v", err)
	}
	defer os.Remove(tempPath)

	cr, err := NewContainerReader(tempPath)
	if err != nil {
		t.Fatalf("NewContainerReader failed: %v", err)
	}
	defer cr.Close()

	cr.ReadBlock(1)
	cr.ReadBlock(2)
	cr.ReadBlock(3)

	cr.ClearCache()

	if cr.IsCached(1) || cr.IsCached(2) || cr.IsCached(3) {
		t.Error("Cache should be empty after ClearCache")
	}
}

func TestGetCacheStats(t *testing.T) {
	containerPath := getTestContainerPath()
	if containerPath == "" {
		t.Skip("test_container.img not found")
	}

	tempPath, err := extractAPFSContainerFromDiskImage(containerPath)
	if err != nil {
		t.Fatalf("Failed to extract APFS container: %v", err)
	}
	defer os.Remove(tempPath)

	cr, err := NewContainerReader(tempPath)
	if err != nil {
		t.Fatalf("NewContainerReader failed: %v", err)
	}
	defer cr.Close()

	cr.ReadBlock(1)
	cr.ReadBlock(2)

	stats := cr.GetCacheStats()

	if stats == nil {
		t.Fatal("Cache stats is nil")
	}

	cachedBlocks, ok := stats["cached_blocks"].(int)
	if !ok || cachedBlocks != 2 {
		t.Errorf("Expected 2 cached blocks, got %v", stats["cached_blocks"])
	}

	cacheSize, ok := stats["cache_size_bytes"].(int)
	if !ok || cacheSize <= 0 {
		t.Errorf("Expected positive cache size, got %v", stats["cache_size_bytes"])
	}
}

func TestSeekToBlock(t *testing.T) {
	containerPath := getTestContainerPath()
	if containerPath == "" {
		t.Skip("test_container.img not found")
	}

	tempPath, err := extractAPFSContainerFromDiskImage(containerPath)
	if err != nil {
		t.Fatalf("Failed to extract APFS container: %v", err)
	}
	defer os.Remove(tempPath)

	cr, err := NewContainerReader(tempPath)
	if err != nil {
		t.Fatalf("NewContainerReader failed: %v", err)
	}
	defer cr.Close()

	err = cr.SeekToBlock(5)
	if err != nil {
		t.Fatalf("SeekToBlock failed: %v", err)
	}
}

func TestSeekToBlockOutOfBounds(t *testing.T) {
	containerPath := getTestContainerPath()
	if containerPath == "" {
		t.Skip("test_container.img not found")
	}

	tempPath, err := extractAPFSContainerFromDiskImage(containerPath)
	if err != nil {
		t.Fatalf("Failed to extract APFS container: %v", err)
	}
	defer os.Remove(tempPath)

	cr, err := NewContainerReader(tempPath)
	if err != nil {
		t.Fatalf("NewContainerReader failed: %v", err)
	}
	defer cr.Close()

	err = cr.SeekToBlock(10000000)
	if err == nil {
		t.Error("Expected error for out-of-bounds seek")
	}
}

func TestGetEndianness(t *testing.T) {
	containerPath := getTestContainerPath()
	if containerPath == "" {
		t.Skip("test_container.img not found")
	}

	tempPath, err := extractAPFSContainerFromDiskImage(containerPath)
	if err != nil {
		t.Fatalf("Failed to extract APFS container: %v", err)
	}
	defer os.Remove(tempPath)

	cr, err := NewContainerReader(tempPath)
	if err != nil {
		t.Fatalf("NewContainerReader failed: %v", err)
	}
	defer cr.Close()

	endian := cr.GetEndianness()
	if endian != binary.LittleEndian {
		t.Error("Expected little-endian")
	}
}

func TestClose(t *testing.T) {
	containerPath := getTestContainerPath()
	if containerPath == "" {
		t.Skip("test_container.img not found")
	}

	tempPath, err := extractAPFSContainerFromDiskImage(containerPath)
	if err != nil {
		t.Fatalf("Failed to extract APFS container: %v", err)
	}
	defer os.Remove(tempPath)

	cr, err := NewContainerReader(tempPath)
	if err != nil {
		t.Fatalf("NewContainerReader failed: %v", err)
	}

	err = cr.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}
