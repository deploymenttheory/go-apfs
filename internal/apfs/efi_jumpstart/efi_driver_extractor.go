// File: internal/efijumpstart/efi_driver_extractor.go
package efijumpstart

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// EFIDriverExtractor implements the EFIDriverExtractor interface.
// It holds the parsed EFI jumpstart data, a reader for the underlying
// storage, and the filesystem block size.
type EFIDriverExtractor struct {
	data      types.NxEfiJumpstartT
	reader    io.ReaderAt // Used to read raw data from the underlying storage
	blockSize uint32      // Block size of the APFS container filesystem
}

// Compile-time check to ensure EFIDriverExtractor implements EFIDriverExtractor
var _ interfaces.EFIDriverExtractor = (*EFIDriverExtractor)(nil)

// NewEFIDriverExtractor creates a new extractor instance.
// It requires the jumpstart data, an io.ReaderAt for the underlying storage,
// and the container's block size. Returns an error if blockSize is zero.
func NewEFIDriverExtractor(data types.NxEfiJumpstartT, reader io.ReaderAt, blockSize uint32) (interfaces.EFIDriverExtractor, error) {
	if reader == nil {
		return nil, fmt.Errorf("reader cannot be nil")
	}
	if blockSize == 0 {
		return nil, fmt.Errorf("block size cannot be zero")
	}
	if data.NejNumExtents != uint32(len(data.NejRecExtents)) {
		return nil, fmt.Errorf("NejNumExtents (%d) does not match NejRecExtents length (%d)", data.NejNumExtents, len(data.NejRecExtents))
	}

	return &EFIDriverExtractor{
		data:      data,
		reader:    reader,
		blockSize: blockSize,
	}, nil
}

// isValid checks the magic and version of the jumpstart data.
func (e *EFIDriverExtractor) isValid() bool {
	return e.data.NejMagic == types.NxEfiJumpstartMagic && e.data.NejVersion == types.NxEfiJumpstartVersion
}

// GetEFIDriverData reads the EFI driver content from the underlying storage based on extents.
func (e *EFIDriverExtractor) GetEFIDriverData() ([]byte, error) {
	if !e.isValid() {
		return nil, fmt.Errorf("EFI jumpstart data is invalid (magic: %x, version: %d)", e.data.NejMagic, e.data.NejVersion)
	}

	expectedSize := int64(e.data.NejEfiFileLen)
	if expectedSize == 0 {
		return []byte{}, nil // No driver data expected
	}

	driverData := bytes.NewBuffer(make([]byte, 0, expectedSize)) // Pre-allocate buffer
	var totalBytesRead int64 = 0

	// Use a copy of extents to avoid potential modification issues if data comes from elsewhere
	extents := make([]types.Prange, len(e.data.NejRecExtents))
	copy(extents, e.data.NejRecExtents)

	for i, extent := range extents {
		if extent.PrBlockCount == 0 {
			continue // Skip empty extents
		}

		offset := int64(extent.PrStartPaddr) * int64(e.blockSize)
		bytesInExtent := int64(extent.PrBlockCount) * int64(e.blockSize)
		bytesToRead := bytesInExtent

		// Avoid reading past the end of the expected file size
		remainingExpected := expectedSize - totalBytesRead
		if remainingExpected <= 0 {
			// Already read the expected amount
			break
		}
		if bytesToRead > remainingExpected {
			bytesToRead = remainingExpected
		}

		if bytesToRead <= 0 {
			// Should not happen if remainingExpected > 0, but defensive check
			continue
		}

		extentBuffer := make([]byte, bytesToRead)
		n, err := e.reader.ReadAt(extentBuffer, offset)

		// Handle ReadAt errors carefully
		if err != nil && err != io.EOF { // EOF might be okay if we requested exactly up to the end
			return nil, fmt.Errorf("failed to read extent %d (offset %d, size %d): %w", i, offset, bytesToRead, err)
		}
		// ReadAt should return an error if n < len(p), unless EOF was hit before reading any data.
		// If n < bytesToRead, it's an unexpected short read.
		if int64(n) < bytesToRead {
			// Check if EOF was the *only* error and we read *some* data. This might happen
			// if the underlying storage ends exactly where we tried to read.
			// However, for ReadAt, reading less than requested without error is unusual.
			// A robust check might be more complex, but ErrUnexpectedEOF is often used.
			// Let's consider any read less than requested as an error here.
			return nil, fmt.Errorf("short read on extent %d: read %d bytes, expected %d (offset %d, error: %v)", i, n, bytesToRead, offset, err)
		}

		// Write the buffer read from the extent into our main driver buffer
		written, writeErr := driverData.Write(extentBuffer[:n]) // Use [:n] for correctness
		if writeErr != nil {
			return nil, fmt.Errorf("failed to write extent %d data to buffer: %w", i, writeErr)
		}
		if int64(written) != int64(n) {
			return nil, fmt.Errorf("short write to internal buffer for extent %d", i)
		}

		totalBytesRead += int64(n)

		// Optimization: if we've read the expected total size, we can stop early.
		if totalBytesRead >= expectedSize {
			break
		}
	}

	// Final check: Did we collect the expected number of bytes?
	if totalBytesRead != expectedSize {
		// This might happen if extents cover less space than NejEfiFileLen declares
		return nil, fmt.Errorf("failed to collect complete EFI driver: got %d bytes, expected %d bytes", totalBytesRead, expectedSize)
	}

	return driverData.Bytes(), nil
}

// ExtractEFIDriver retrieves the EFI driver data and saves it to the specified file path.
func (e *EFIDriverExtractor) ExtractEFIDriver(outputPath string) error {
	driverData, err := e.GetEFIDriverData()
	if err != nil {
		return fmt.Errorf("failed to get EFI driver data: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file '%s': %w", outputPath, err)
	}
	defer file.Close() // Ensure file is closed

	n, err := file.Write(driverData)
	if err != nil {
		// Attempt to remove partially written file on error? Optional.
		// os.Remove(outputPath)
		return fmt.Errorf("failed to write driver data to '%s': %w", outputPath, err)
	}
	if n != len(driverData) {
		// os.Remove(outputPath)
		return fmt.Errorf("incomplete write to '%s': wrote %d bytes, expected %d", outputPath, n, len(driverData))
	}

	// Ensure data is flushed to disk
	if err := file.Sync(); err != nil {
		// Log warning maybe, but don't fail the operation for sync error usually
		log.Printf("Warning: failed to sync output file '%s': %v", outputPath, err)
	} //

	return nil // Success
}

// ValidateEFIDriver checks if the EFI driver data can be fully read according to the extents and expected size.
// It calls GetEFIDriverData internally to perform the checks.
func (e *EFIDriverExtractor) ValidateEFIDriver() error {
	_, err := e.GetEFIDriverData()
	if err != nil {
		return fmt.Errorf("EFI driver validation failed: %w", err)
	}
	// If GetEFIDriverData succeeded, the driver is considered valid in terms of readability and size match.
	return nil
}
