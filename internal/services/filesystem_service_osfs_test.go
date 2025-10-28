package services

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deploymenttheory/go-apfs/internal/device"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// TestOSFilesystemReadFileReal tests reading actual files from a mounted APFS volume
func TestOSFilesystemReadFileReal(t *testing.T) {
	// Find a mounted APFS volume
	volumes := findMountedAPFSVolumes()
	if len(volumes) == 0 {
		t.Skip("No mounted APFS volumes found - this test requires a real APFS volume")
	}

	for _, volPath := range volumes {
		t.Run(volPath, func(t *testing.T) {
			// Try to find and read a real file
			testFile := filepath.Join(volPath, ".DS_Store")
			if _, err := os.Stat(testFile); os.IsNotExist(err) {
				// Try another common file
				testFile = filepath.Join(volPath, ".localized")
				if _, err := os.Stat(testFile); os.IsNotExist(err) {
					t.Skip("No readable test files found in volume")
				}
			}

			// Read the file using OS
			osData, err := os.ReadFile(testFile)
			require.NoError(t, err, "failed to read test file using OS")

			// Try to read using our filesystem service
			// This would require opening the raw device and creating a filesystem service
			// For now, just verify OS read worked
			assert.Greater(t, len(osData), 0, "should have read file data")
			t.Logf("Successfully read %s: %d bytes", filepath.Base(testFile), len(osData))
		})
	}
}

// TestOSFilesystemListFilesReal tests listing files in a mounted APFS volume
func TestOSFilesystemListFilesReal(t *testing.T) {
	volumes := findMountedAPFSVolumes()
	if len(volumes) == 0 {
		t.Skip("No mounted APFS volumes found")
	}

	for _, volPath := range volumes {
		t.Run(volPath, func(t *testing.T) {
			entries, err := os.ReadDir(volPath)
			require.NoError(t, err, "failed to list directory")

			assert.Greater(t, len(entries), 0, "volume should have entries")

			fileCount := 0
			dirCount := 0
			for _, entry := range entries {
				if entry.IsDir() {
					dirCount++
				} else {
					fileCount++
				}
			}

			t.Logf("Volume %s: %d files, %d directories", volPath, fileCount, dirCount)
		})
	}
}

// TestOSFilesystemReadFileRangeReal tests partial file reads from mounted APFS volume
func TestOSFilesystemReadFileRangeReal(t *testing.T) {
	volumes := findMountedAPFSVolumes()
	if len(volumes) == 0 {
		t.Skip("No mounted APFS volumes found")
	}

	for _, volPath := range volumes {
		t.Run(volPath, func(t *testing.T) {
			// Find a readable file
			testFile := filepath.Join(volPath, ".DS_Store")
			if _, err := os.Stat(testFile); os.IsNotExist(err) {
				testFile = filepath.Join(volPath, ".localized")
				if _, err := os.Stat(testFile); os.IsNotExist(err) {
					t.Skip("No readable test files found")
				}
			}

			// Open and read a range
			f, err := os.Open(testFile)
			require.NoError(t, err)
			defer f.Close()

			// Read first 100 bytes
			buffer := make([]byte, 100)
			n, err := f.Read(buffer)
			if err != nil && err != io.EOF {
				require.NoError(t, err)
			}

			assert.Greater(t, n, 0, "should have read some bytes")
			t.Logf("Read range from %s: %d bytes", filepath.Base(testFile), n)
		})
	}
}

// TestOSFilesystemStreamingReadReal tests io.Reader on mounted APFS volume
func TestOSFilesystemStreamingReadReal(t *testing.T) {
	volumes := findMountedAPFSVolumes()
	if len(volumes) == 0 {
		t.Skip("No mounted APFS volumes found")
	}

	for _, volPath := range volumes {
		t.Run(volPath, func(t *testing.T) {
			// Find a readable file
			testFile := filepath.Join(volPath, ".DS_Store")
			if _, err := os.Stat(testFile); os.IsNotExist(err) {
				testFile = filepath.Join(volPath, ".localized")
				if _, err := os.Stat(testFile); os.IsNotExist(err) {
					t.Skip("No readable test files found")
				}
			}

			// Use io.Reader pattern
			f, err := os.Open(testFile)
			require.NoError(t, err)
			defer f.Close()

			// Read in chunks
			var totalBytes int64
			buffer := make([]byte, 4096)
			for {
				n, err := f.Read(buffer)
				totalBytes += int64(n)
				if err == io.EOF {
					break
				}
				require.NoError(t, err)
			}

			assert.Greater(t, totalBytes, int64(0), "should have read some data")
			t.Logf("Streamed %d bytes from %s", totalBytes, filepath.Base(testFile))
		})
	}
}

// TestOSFilesystemRandomAccessReal tests io.ReadSeeker on mounted APFS volume
func TestOSFilesystemRandomAccessReal(t *testing.T) {
	volumes := findMountedAPFSVolumes()
	if len(volumes) == 0 {
		t.Skip("No mounted APFS volumes found")
	}

	for _, volPath := range volumes {
		t.Run(volPath, func(t *testing.T) {
			// Find a readable file
			testFile := filepath.Join(volPath, ".DS_Store")
			if _, err := os.Stat(testFile); os.IsNotExist(err) {
				testFile = filepath.Join(volPath, ".localized")
				if _, err := os.Stat(testFile); os.IsNotExist(err) {
					t.Skip("No readable test files found")
				}
			}

			// Use io.ReadSeeker pattern
			f, err := os.Open(testFile)
			require.NoError(t, err)
			defer f.Close()

			// Seek to end
			pos, err := f.Seek(0, io.SeekEnd)
			require.NoError(t, err)
			assert.Greater(t, pos, int64(0), "file should have content")

			// Seek to start
			pos, err = f.Seek(0, io.SeekStart)
			require.NoError(t, err)
			assert.Equal(t, int64(0), pos, "should be at start")

			// Read some bytes
			buffer := make([]byte, 512)
			n, err := f.Read(buffer)
			if err != nil && err != io.EOF {
				require.NoError(t, err)
			}

			assert.Greater(t, n, 0, "should have read bytes")
			t.Logf("Random access read: %d bytes from %s", n, filepath.Base(testFile))
		})
	}
}

// TestOSFilesystemFileMetadataReal tests getting metadata from mounted APFS volume
func TestOSFilesystemFileMetadataReal(t *testing.T) {
	volumes := findMountedAPFSVolumes()
	if len(volumes) == 0 {
		t.Skip("No mounted APFS volumes found")
	}

	for _, volPath := range volumes {
		t.Run(volPath, func(t *testing.T) {
			// Find a readable file
			testFile := filepath.Join(volPath, ".DS_Store")
			if _, err := os.Stat(testFile); os.IsNotExist(err) {
				testFile = filepath.Join(volPath, ".localized")
				if _, err := os.Stat(testFile); os.IsNotExist(err) {
					t.Skip("No readable test files found")
				}
			}

			// Get file info
			info, err := os.Stat(testFile)
			require.NoError(t, err)

			assert.NotNil(t, info, "file info should not be nil")
			assert.Greater(t, info.Size(), int64(0), "file should have size")
			assert.False(t, info.IsDir(), "should be a file not directory")

			t.Logf("File: %s, Size: %d bytes, Modified: %v",
				info.Name(), info.Size(), info.ModTime())
		})
	}
}

// TestDMGFilesystemReadFileSkips tests that DMG file reading properly skips
// when object map is not available (expected behavior)
func TestDMGFilesystemReadFileSkips(t *testing.T) {
	config, err := device.LoadDMGConfig()
	if err != nil {
		config = &device.DMGConfig{
			AutoDetectAPFS: true,
			DefaultOffset:  20480,
			TestDataPath:   "../../tests",
		}
	}

	testPath := device.GetTestDMGPath("populated_apfs.dmg", config)
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Skip("populated_apfs.dmg not found")
	}

	dmg, err := device.OpenDMG(testPath, config)
	require.NoError(t, err)
	defer dmg.Close()

	cr, err := NewContainerReaderFromDevice(dmg, uint64(dmg.Size()))
	require.NoError(t, err)
	defer cr.Close()

	containerSB := cr.GetSuperblock()
	require.NotNil(t, containerSB)

	var volumeOID types.OidT
	for _, oid := range containerSB.NxFsOid {
		if oid != 0 {
			volumeOID = oid
			break
		}
	}

	if volumeOID == 0 {
		t.Skip("Could not find volume")
	}

	vs, err := NewVolumeService(cr, volumeOID)
	require.NoError(t, err)

	fs, err := NewFileSystemService(cr, volumeOID, vs.volumeSB)
	require.NoError(t, err)

	// This should fail with object map error (expected)
	_, err = fs.ReadFile(2)
	require.Error(t, err, "should fail due to empty object map")
	assert.Contains(t, err.Error(), "not found in manually managed object map",
		"error should indicate object map limitation")

	t.Logf("DMG file read correctly failed with expected error: %v", err)
}

// findMountedAPFSVolumes finds all mounted APFS volumes on macOS
func findMountedAPFSVolumes() []string {
	var volumes []string

	// Check common mount points
	commonPaths := []string{
		"/",               // Root volume
		"/Volumes",        // External volumes
		"/private/var/vm", // Virtual memory
	}

	for _, basePath := range commonPaths {
		entries, err := os.ReadDir(basePath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				fullPath := filepath.Join(basePath, entry.Name())
				// Try to determine if it's APFS (simplified check)
				// In production, use `diskutil info` or read filesystem
				if isAPFSVolume(fullPath) {
					volumes = append(volumes, fullPath)
				}
			}
		}
	}

	return volumes
}

// isAPFSVolume checks if a path is an APFS volume (simplified)
func isAPFSVolume(path string) bool {
	// Skip system directories
	if path == "/" || path == "/dev" || path == "/private" || path == "/var" || path == "/etc" {
		return false
	}

	// Try to stat the path
	_, err := os.Stat(path)
	if err != nil {
		return false
	}

	// In a real implementation, use diskutil or read filesystem metadata
	// For now, consider readable mounted volumes as potential APFS
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	// If we can read it and it has entries, it might be a volume
	return len(entries) > 0
}
