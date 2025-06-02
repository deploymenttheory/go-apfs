// File: internal/efijumpstart/efi_driver_extractor_test.go
package efijumpstart

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// EFI Driver content parts
	testDriverPart1 = bytes.Repeat([]byte{0xAA}, testBlockSize)
	testDriverPart2 = bytes.Repeat([]byte{0xBB}, testBlockSize/2) // Partial block
	testDriverPart3 = bytes.Repeat([]byte{0xCC}, testBlockSize)

	// Correctly calculated expected data based on mock disk layout, extents, and file length
	testCorrectExpectedDriverData = func() []byte {
		expectedLen := 10240 // As defined in testJumpData.NejEfiFileLen
		buf := &bytes.Buffer{}
		// Add Block 0 content
		buf.Write(testDriverPart1) // 4096 bytes (0xAA)
		// Add Block 1 content (part2 + zeroes)
		block1Content := append(testDriverPart2, make([]byte, testBlockSize/2)...)
		buf.Write(block1Content) // 4096 bytes (0xBB... + 0x00...)
		// Add required part of Block 2 content (first half of part3)
		bytesNeededFromBlock2 := expectedLen - buf.Len() // 10240 - 8192 = 2048
		if bytesNeededFromBlock2 > 0 {
			buf.Write(testDriverPart3[:bytesNeededFromBlock2]) // 2048 bytes (0xCC...)
		}
		// Ensure final buffer size matches expected length exactly
		finalBytes := buf.Bytes()
		if len(finalBytes) > expectedLen {
			return finalBytes[:expectedLen]
		}
		return finalBytes
	}()
	testExpectedDriverLen = uint32(len(testCorrectExpectedDriverData)) // Should still be 10240

	// Mock disk layout containing the parts
	testMockDisk = func() []byte {
		disk := make([]byte, 3*testBlockSize) // 12288 bytes total
		copy(disk[testBlock0Offset:], testDriverPart1)
		copy(disk[testBlock1Offset:], testDriverPart2) // Only first half of block 1 used (2048 bytes), rest is 0x00
		copy(disk[testBlock2Offset:], testDriverPart3)
		return disk
	}()

	// Corresponding EFI Jumpstart structure
	testJumpData = types.NxEfiJumpstartT{
		NejMagic:      types.NxEfiJumpstartMagic,
		NejVersion:    types.NxEfiJumpstartVersion,
		NejEfiFileLen: testExpectedDriverLen, // Use the length calculated from the correct data (10240)
		NejNumExtents: 3,
		NejRecExtents: []types.Prange{
			{PrStartPaddr: 0, PrBlockCount: 1}, // Block 0 (reads 4096)
			{PrStartPaddr: 1, PrBlockCount: 1}, // Block 1 (reads 4096)
			{PrStartPaddr: 2, PrBlockCount: 1}, // Block 2 (reads 2048 needed)
		},
		// NejO and NejReserved omitted for brevity
	}
)

// --- Tests ---

func TestNewEFIDriverExtractor(t *testing.T) {
	reader := bytes.NewReader(testMockDisk)

	t.Run("Success", func(t *testing.T) {
		extractor, err := NewEFIDriverExtractor(testJumpData, reader, testBlockSize)
		require.NoError(t, err)
		require.NotNil(t, extractor)
	})

	t.Run("NilReader", func(t *testing.T) {
		extractor, err := NewEFIDriverExtractor(testJumpData, nil, testBlockSize)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reader cannot be nil")
		assert.Nil(t, extractor)
	})

	t.Run("ZeroBlockSize", func(t *testing.T) {
		extractor, err := NewEFIDriverExtractor(testJumpData, reader, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "block size cannot be zero")
		assert.Nil(t, extractor)
	})

	t.Run("ExtentCountMismatch", func(t *testing.T) {
		mismatchJump := testJumpData   // copy
		mismatchJump.NejNumExtents = 2 // Incorrect count (has 3 extents)
		extractor, err := NewEFIDriverExtractor(mismatchJump, reader, testBlockSize)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "NejNumExtents (2) does not match NejRecExtents length (3)")
		assert.Nil(t, extractor)
	})
}

func TestEFIDriverExtractorImpl_GetEFIDriverData(t *testing.T) {
	reader := bytes.NewReader(testMockDisk)

	t.Run("Success", func(t *testing.T) {
		extractor, err := NewEFIDriverExtractor(testJumpData, reader, testBlockSize)
		require.NoError(t, err)
		data, err := extractor.GetEFIDriverData()
		require.NoError(t, err)
		require.NotNil(t, data)
		assert.Equal(t, int(testExpectedDriverLen), len(data), "Data length mismatch")
		assert.Equal(t, testCorrectExpectedDriverData, data, "Data content mismatch")
	})

	t.Run("Success_ZeroLength", func(t *testing.T) {
		zeroLenJumpData := testJumpData // copy struct
		zeroLenJumpData.NejEfiFileLen = 0
		zeroLenJumpData.NejNumExtents = 0 // Consistent
		zeroLenJumpData.NejRecExtents = nil
		extractor, err := NewEFIDriverExtractor(zeroLenJumpData, reader, testBlockSize)
		require.NoError(t, err)

		data, err := extractor.GetEFIDriverData()
		require.NoError(t, err)
		assert.Empty(t, data, "Expected empty data for zero length driver")
	})

	t.Run("PartialLastExtentUsed", func(t *testing.T) {
		partialLen := uint32(testBlockSize + testBlockSize/4) // Expect 4096 + 1024 = 5120 bytes
		partialJumpData := testJumpData                       // copy
		partialJumpData.NejEfiFileLen = partialLen
		// Use the same extents, let the length limit do the work

		// Calculate the correct expected data for this partial length
		expectedPartialData := func() []byte {
			buf := &bytes.Buffer{}
			buf.Write(testDriverPart1)                           // 4096 AA
			bytesNeededFromBlock1 := int(partialLen) - buf.Len() // 5120 - 4096 = 1024
			if bytesNeededFromBlock1 > 0 {
				// Read from Block 1 content (part2 + zeroes)
				block1Content := append(testDriverPart2, make([]byte, testBlockSize/2)...)
				buf.Write(block1Content[:bytesNeededFromBlock1]) // Write first 1024 bytes of block 1
			}
			finalBytes := buf.Bytes() // Should be exactly 5120 bytes
			if len(finalBytes) > int(partialLen) {
				return finalBytes[:partialLen]
			}
			return finalBytes
		}()

		extractor, err := NewEFIDriverExtractor(partialJumpData, reader, testBlockSize)
		require.NoError(t, err)

		data, err := extractor.GetEFIDriverData()
		require.NoError(t, err)
		require.NotNil(t, data)
		assert.Equal(t, int(partialLen), len(data), "Data length mismatch")
		assert.Equal(t, expectedPartialData, data, "Data content mismatch")
	})

	t.Run("InvalidJumpstartData", func(t *testing.T) {
		invalidJumpData := testJumpData       // copy
		invalidJumpData.NejMagic = 0xDEADBEEF // Use a valid hex literal
		extractor, err := NewEFIDriverExtractor(invalidJumpData, reader, testBlockSize)
		require.NoError(t, err) // Constructor doesn't validate magic, Get does

		_, err = extractor.GetEFIDriverData()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "EFI jumpstart data is invalid")
	})

	t.Run("ReadError", func(t *testing.T) {
		mockErrReader := &mockReaderAt{
			data:        testMockDisk,
			failOnRead:  true,
			failAfterN:  1, // Fail on the second ReadAt call (extent index 1)
			errToReturn: errors.New("disk read failed"),
		}

		extractor, err := NewEFIDriverExtractor(testJumpData, mockErrReader, testBlockSize)
		require.NoError(t, err)

		_, err = extractor.GetEFIDriverData()
		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to read extent 1")
		assert.ErrorContains(t, err, "disk read failed")
	})

	t.Run("ShortReadError_DiskTooSmall", func(t *testing.T) {
		// Disk only contains block 0 and 1 (8192 bytes), but jumpstart expects block 2 as well
		shortDisk := testMockDisk[:2*testBlockSize]
		shortReader := bytes.NewReader(shortDisk)

		extractor, err := NewEFIDriverExtractor(testJumpData, shortReader, testBlockSize)
		require.NoError(t, err)

		_, err = extractor.GetEFIDriverData()
		require.Error(t, err)
		// Expect error when trying to read the 3rd extent (index 2) from missing data
		// mockReaderAt/bytes.Reader might return io.EOF or io.ErrUnexpectedEOF when reading past end.
		// Our GetEFIDriverData checks for short reads.
		assert.ErrorContains(t, err, "short read on extent 2")
	})

	t.Run("DeclaredLengthExceedsExtentData", func(t *testing.T) {
		longLenJump := testJumpData                                 // Copy
		longLenJump.NejEfiFileLen = uint32(len(testMockDisk) + 100) // Ask for more data than extents provide (12288 + 100)

		extractor, err := NewEFIDriverExtractor(longLenJump, reader, testBlockSize)
		require.NoError(t, err)

		_, err = extractor.GetEFIDriverData()
		require.Error(t, err)
		// This should fail at the *end* check, after reading all available extent data (12288 bytes)
		// The code reads 3 full blocks = 12288 bytes.
		// The final check compares totalBytesRead (12288) with expectedSize (12288+100).
		assert.ErrorContains(t, err, "failed to collect complete EFI driver")
		assert.ErrorContains(t, err, fmt.Sprintf("got %d bytes, expected %d bytes", len(testMockDisk), longLenJump.NejEfiFileLen))
	})
}

func TestEFIDriverExtractorImpl_ExtractEFIDriver(t *testing.T) {
	reader := bytes.NewReader(testMockDisk)

	t.Run("Success", func(t *testing.T) {
		extractor, err := NewEFIDriverExtractor(testJumpData, reader, testBlockSize)
		require.NoError(t, err)

		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "efi_driver.bin")

		err = extractor.ExtractEFIDriver(outputPath)
		require.NoError(t, err)

		// Verify file content
		extractedData, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Equal(t, testCorrectExpectedDriverData, extractedData)
	})

	t.Run("GetEFIDriverData_Fails", func(t *testing.T) {
		invalidJumpData := testJumpData       // copy
		invalidJumpData.NejMagic = 0xDEADBEEF // Use a valid hex literal
		extractor, err := NewEFIDriverExtractor(invalidJumpData, reader, testBlockSize)
		require.NoError(t, err) // Constructor is okay

		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "efi_driver_fail.bin")

		err = extractor.ExtractEFIDriver(outputPath) // This internally calls GetEFIDriverData which will fail
		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to get EFI driver data")
		assert.ErrorContains(t, err, "EFI jumpstart data is invalid") // Check for the specific underlying error

		_, statErr := os.Stat(outputPath)
		assert.True(t, os.IsNotExist(statErr), "Output file should not exist after failed get")
	})

	t.Run("FileCreateError", func(t *testing.T) {
		extractor, err := NewEFIDriverExtractor(testJumpData, reader, testBlockSize)
		require.NoError(t, err)

		// Use an invalid path (e.g., path is a directory)
		tempDir := t.TempDir()
		invalidOutputPath := tempDir // Cannot create file with name of existing dir

		err = extractor.ExtractEFIDriver(invalidOutputPath)
		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to create output file")
		// Check the specific OS error if needed (e.g., EISDIR on Unix)
		// require.ErrorIs(t, err, os.ErrExist) // Or similar depending on OS/Go version behavior
	})

	t.Run("FileWriteError", func(t *testing.T) {
		// Simulating a write error reliably in a unit test without complex mocks is hard.
		// Common causes: disk full, permissions error, device error.
		// This test case is often skipped or requires OS-level mocking/setup.
		t.Skip("Skipping direct file write error simulation in unit test")
	})
}

func TestEFIDriverExtractorImpl_ValidateEFIDriver(t *testing.T) {
	reader := bytes.NewReader(testMockDisk)

	t.Run("Success", func(t *testing.T) {
		extractor, err := NewEFIDriverExtractor(testJumpData, reader, testBlockSize)
		require.NoError(t, err)

		err = extractor.ValidateEFIDriver()
		assert.NoError(t, err)
	})

	t.Run("Failure_DueTo_GetEFIDriverData_Error", func(t *testing.T) {
		// Use a setup known to fail GetEFIDriverData (e.g., read error)
		mockErrReader := &mockReaderAt{
			data:        testMockDisk,
			failOnRead:  true,
			failAfterN:  0, // Fail immediately on first read attempt
			errToReturn: errors.New("validate read failed"),
		}
		extractor, err := NewEFIDriverExtractor(testJumpData, mockErrReader, testBlockSize)
		require.NoError(t, err)

		err = extractor.ValidateEFIDriver()
		require.Error(t, err)
		assert.ErrorContains(t, err, "EFI driver validation failed")
		assert.ErrorContains(t, err, "validate read failed") // Check underlying error is propagated
	})
}
